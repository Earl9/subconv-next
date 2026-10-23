package fetcher

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"subconv-next/internal/model"
	"subconv-next/internal/storage"
)

type Source struct {
	Name               string
	URL                string
	UserAgent          string
	Enabled            bool
	InsecureSkipVerify bool
	AllowPrivateHosts  bool
	CacheTTL           time.Duration
	PreferCache        bool
}

// ErrCacheMiss is returned when PreferCache is set but no cached content
// exists yet. Callers that prefer cache deliberately avoid the network, so a
// missing cache entry must fail fast instead of blocking on an upstream fetch.
var ErrCacheMiss = errors.New("no cached subscription content")

type FetchedSubscription struct {
	Name                 string
	URL                  string
	Content              []byte
	ContentType          string
	SubscriptionUserinfo string
	FetchedAt            time.Time
	FromCache            bool
}

type CacheMeta struct {
	URLHash              string    `json:"url_hash"`
	FetchedAt            time.Time `json:"fetched_at"`
	StatusCode           int       `json:"status_code"`
	ETag                 string    `json:"etag"`
	LastModified         string    `json:"last_modified"`
	Size                 int       `json:"size"`
	ContentType          string    `json:"content_type"`
	SubscriptionUserinfo string    `json:"subscription_userinfo,omitempty"`
	OriginalURL          string    `json:"original_url"`
}

type Options struct {
	CacheDir          string
	SharedCacheDir    string
	Timeout           time.Duration
	MaxBodyBytes      int64
	MaxRedirects      int
	CacheTTL          time.Duration
	AllowPrivateHosts bool
	PublicMode        bool
	Resolver          HostResolver
	RequestDoer       RequestDoer
}

type HostResolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

type RequestDoer func(ctx context.Context, target *url.URL, resolvedIP net.IP, source Source) (*http.Response, error)

type Fetcher struct {
	opts Options
}

func OptionsFromConfig(cfg model.Config) Options {
	return Options{
		CacheDir:          cfg.Service.CacheDir,
		SharedCacheDir:    SharedCacheDir(cfg.Service.CacheDir),
		Timeout:           time.Duration(cfg.Service.FetchTimeoutSeconds) * time.Second,
		MaxBodyBytes:      int64(cfg.Service.MaxSubscriptionBytes),
		MaxRedirects:      3,
		CacheTTL:          time.Duration(cfg.Service.RefreshInterval) * time.Second,
		AllowPrivateHosts: cfg.Service.AllowLAN,
		PublicMode:        cfg.Service.PublicConverter,
	}
}

// SharedCacheDir is the cross-workspace cache location derived from a cache
// directory: subscriptions with the same URL share the most recently
// downloaded content across workspaces (draft sessions create workspaces
// constantly). Workspace cache dirs live under a shared ".../workspaces"
// root, so their fallback location is a sibling "shared" dir of that root.
func SharedCacheDir(cacheDir string) string {
	cacheDir = strings.TrimSpace(cacheDir)
	if cacheDir == "" {
		return ""
	}
	if filepath.Base(filepath.Dir(cacheDir)) == "workspaces" {
		return filepath.Join(filepath.Dir(filepath.Dir(cacheDir)), "shared")
	}
	return filepath.Join(cacheDir, "shared")
}

func New(opts Options) *Fetcher {
	if opts.Timeout <= 0 {
		opts.Timeout = 15 * time.Second
	}
	if opts.MaxBodyBytes <= 0 {
		opts.MaxBodyBytes = 5 * 1024 * 1024
	}
	if opts.MaxRedirects <= 0 {
		opts.MaxRedirects = 3
	}
	if opts.Resolver == nil {
		// Router proxy DNS often returns Fake-IP addresses (198.18.0.0/15).
		// Keep the router resolver for split-DNS environments, then retry with
		// public DNS only when it returns no usable public address.
		opts.Resolver = newPublicHostResolver(net.DefaultResolver)
	}
	return &Fetcher{opts: opts}
}

type publicHostResolver struct {
	primary   HostResolver
	fallbacks []HostResolver
}

func newPublicHostResolver(primary HostResolver) HostResolver {
	if primary == nil {
		primary = net.DefaultResolver
	}
	return publicHostResolver{
		primary: primary,
		fallbacks: []HostResolver{
			newDirectDNSResolver("1.1.1.1:53"),
			newDirectDNSResolver("8.8.8.8:53"),
		},
	}
}

func newDirectDNSResolver(address string) HostResolver {
	return &net.Resolver{
		PreferGo:     true,
		StrictErrors: false,
		Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
			dialer := &net.Dialer{Timeout: 2 * time.Second}
			return dialer.DialContext(ctx, "udp", address)
		},
	}
}

func (resolver publicHostResolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	primaryIPs, primaryErr := resolver.primary.LookupIPAddr(ctx, host)
	if containsPublicIP(primaryIPs) {
		return primaryIPs, nil
	}

	for _, fallback := range resolver.fallbacks {
		lookupCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		fallbackIPs, err := fallback.LookupIPAddr(lookupCtx, host)
		cancel()
		if err == nil && containsPublicIP(fallbackIPs) {
			return fallbackIPs, nil
		}
	}

	if primaryErr != nil {
		return nil, primaryErr
	}
	return primaryIPs, nil
}

func containsPublicIP(ips []net.IPAddr) bool {
	for _, ipAddr := range ips {
		if !isBlockedIP(ipAddr.IP) {
			return true
		}
	}
	return false
}

func (f *Fetcher) Fetch(ctx context.Context, source Source) (FetchedSubscription, []string, error) {
	if !source.Enabled {
		return FetchedSubscription{}, nil, errors.New("source is disabled")
	}

	currentURL, err := validateFetchURL(source.URL)
	if err != nil {
		return FetchedSubscription{}, nil, err
	}
	if source.PreferCache {
		if cached, cacheErr := f.readCache(source.URL); cacheErr == nil {
			cached.FromCache = true
			return cached, nil, nil
		}
		if cached, ok := f.readSharedCache(source.URL); ok {
			cached.FromCache = true
			return cached, nil, nil
		}
		return FetchedSubscription{}, nil, fmt.Errorf("%w for %s (offline view; run a refresh to fetch it)", ErrCacheMiss, source.Name)
	}
	if source.CacheTTL > 0 {
		if cached, cacheErr := f.readCache(source.URL); cacheErr == nil {
			age := time.Since(cached.FetchedAt)
			if age >= 0 && age < source.CacheTTL {
				cached.FromCache = true
				return cached, []string{fmt.Sprintf("using fresh cached content for %s", source.Name)}, nil
			}
		}
	}

	fetched, err := f.fetchNetwork(ctx, currentURL, source)
	if err == nil {
		var warnings []string
		if shouldRetryWithGenericUserAgent(source, fetched.Content) {
			fallback := source
			fallback.UserAgent = model.DefaultUserAgent
			refetched, retryErr := f.fetchNetwork(ctx, currentURL, fallback)
			if retryErr == nil {
				fetched = refetched
				warnings = append(warnings, fmt.Sprintf("retried %s with generic user-agent after receiving empty Clash/Mihomo YAML", source.Name))
			} else {
				warnings = append(warnings, fmt.Sprintf("generic user-agent retry failed for %s: %v", source.Name, retryErr))
			}
		}
		if cacheErr := f.writeCache(source.URL, fetched); cacheErr != nil {
			warnings = append(warnings, fmt.Sprintf("cache write failed for %s: %v", source.Name, cacheErr))
			return fetched, warnings, nil
		}
		f.writeSharedCache(source.URL, fetched)
		return fetched, warnings, nil
	}

	cached, cacheErr := f.readCache(source.URL)
	if cacheErr == nil {
		cached.FromCache = true
		return cached, []string{fmt.Sprintf("using cached subscription for %s after fetch failure", source.Name)}, nil
	}
	if shared, ok := f.readSharedCache(source.URL); ok {
		shared.FromCache = true
		return shared, []string{fmt.Sprintf("using shared cached subscription for %s after fetch failure", source.Name)}, nil
	}

	return FetchedSubscription{}, nil, err
}

func (f *Fetcher) fetchNetwork(ctx context.Context, currentURL *url.URL, source Source) (FetchedSubscription, error) {
	redirects := 0

	for {
		if f.opts.PublicMode && !publicFetchPortAllowed(currentURL) {
			return FetchedSubscription{}, fmt.Errorf("outbound port %s is not allowed in public converter mode", effectivePort(currentURL))
		}
		resolvedIP, err := f.resolveHost(ctx, currentURL.Hostname(), source.AllowPrivateHosts)
		if err != nil {
			return FetchedSubscription{}, err
		}

		resp, err := f.doRequest(ctx, currentURL, resolvedIP, source)
		if err != nil {
			return FetchedSubscription{}, err
		}

		if isRedirectStatus(resp.StatusCode) {
			_ = resp.Body.Close()
			if redirects >= f.opts.MaxRedirects {
				return FetchedSubscription{}, fmt.Errorf("too many redirects")
			}

			location, err := resp.Location()
			if err != nil {
				return FetchedSubscription{}, fmt.Errorf("invalid redirect location: %w", err)
			}

			currentURL = currentURL.ResolveReference(location)
			if _, err := validateFetchURL(currentURL.String()); err != nil {
				return FetchedSubscription{}, err
			}
			redirects++
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			_ = resp.Body.Close()
			return FetchedSubscription{}, fmt.Errorf("unexpected status code %d", resp.StatusCode)
		}

		body, err := readLimited(resp.Body, f.opts.MaxBodyBytes)
		if err != nil {
			_ = resp.Body.Close()
			return FetchedSubscription{}, err
		}
		_ = resp.Body.Close()

		return FetchedSubscription{
			Name:                 source.Name,
			URL:                  source.URL,
			Content:              body,
			ContentType:          resp.Header.Get("Content-Type"),
			SubscriptionUserinfo: resp.Header.Get("Subscription-Userinfo"),
			FetchedAt:            time.Now().UTC(),
		}, nil
	}
}

func (f *Fetcher) doRequest(ctx context.Context, target *url.URL, resolvedIP net.IP, source Source) (*http.Response, error) {
	if f.opts.RequestDoer != nil {
		return f.opts.RequestDoer(ctx, target, resolvedIP, source)
	}

	dialTarget := net.JoinHostPort(resolvedIP.String(), effectivePort(target))
	transport := &http.Transport{
		Proxy: nil,
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			dialer := &net.Dialer{Timeout: f.opts.Timeout}
			return dialer.DialContext(ctx, network, dialTarget)
		},
		TLSClientConfig: &tls.Config{
			ServerName:         target.Hostname(),
			InsecureSkipVerify: source.InsecureSkipVerify,
			MinVersion:         tls.VersionTLS12,
		},
		ForceAttemptHTTP2: true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   f.opts.Timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgentOrDefault(source.UserAgent))
	resp, err := client.Do(req)
	if err != nil {
		transport.CloseIdleConnections()
		return nil, err
	}
	// Each request uses a transport with a destination-specific dialer. Close
	// its idle pool when the caller finishes consuming the response so refresh
	// cycles do not retain one idle connection pool per source request.
	resp.Body = &closeIdleBody{
		ReadCloser: resp.Body,
		closeIdle:  transport.CloseIdleConnections,
	}
	return resp, nil
}

type closeIdleBody struct {
	io.ReadCloser
	closeIdle func()
}

func (body *closeIdleBody) Close() error {
	err := body.ReadCloser.Close()
	if body.closeIdle != nil {
		body.closeIdle()
	}
	return err
}

func (f *Fetcher) resolveHost(ctx context.Context, host string, allowPrivateHosts bool) (net.IP, error) {
	allowPrivateHosts = allowPrivateHosts || f.opts.AllowPrivateHosts
	if isBlockedHostname(host) && !allowPrivateHosts {
		return nil, fmt.Errorf("blocked host %q", host)
	}

	ips, err := f.opts.Resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("resolve host %q: %w", host, err)
	}

	for _, ipAddr := range ips {
		if allowPrivateHosts || !isBlockedIP(ipAddr.IP) {
			return ipAddr.IP, nil
		}
	}

	return nil, fmt.Errorf("no allowed IPs for host %q", host)
}

func (f *Fetcher) readCache(rawURL string) (FetchedSubscription, error) {
	return f.readCacheFrom(f.opts.CacheDir, rawURL)
}

// readSharedCache falls back to the cross-workspace shared copy of a
// subscription. It reports ok=false when sharing is disabled or no copy
// exists, which callers treat the same as a normal cache miss.
func (f *Fetcher) readSharedCache(rawURL string) (FetchedSubscription, bool) {
	if strings.TrimSpace(f.opts.SharedCacheDir) == "" || f.opts.SharedCacheDir == f.opts.CacheDir {
		return FetchedSubscription{}, false
	}
	cached, err := f.readCacheFrom(f.opts.SharedCacheDir, rawURL)
	if err != nil {
		return FetchedSubscription{}, false
	}
	return cached, true
}

func (f *Fetcher) readCacheFrom(dir, rawURL string) (FetchedSubscription, error) {
	bodyPath, metaPath := cachePaths(dir, rawURL)

	body, err := os.ReadFile(bodyPath)
	if err != nil {
		return FetchedSubscription{}, err
	}

	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		return FetchedSubscription{}, err
	}

	var meta CacheMeta
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return FetchedSubscription{}, fmt.Errorf("decode cache meta: %w", err)
	}
	expectedHash := urlHash(rawURL)
	if meta.URLHash != expectedHash || strings.TrimSpace(meta.OriginalURL) != strings.TrimSpace(rawURL) {
		return FetchedSubscription{}, fmt.Errorf("cache metadata does not match source URL")
	}
	if meta.Size < 0 || meta.Size != len(body) || int64(len(body)) > f.opts.MaxBodyBytes {
		return FetchedSubscription{}, fmt.Errorf("cache body size does not match metadata")
	}
	if meta.FetchedAt.IsZero() {
		return FetchedSubscription{}, fmt.Errorf("cache metadata is missing fetch time")
	}

	return FetchedSubscription{
		URL:                  rawURL,
		Content:              body,
		ContentType:          meta.ContentType,
		SubscriptionUserinfo: meta.SubscriptionUserinfo,
		FetchedAt:            meta.FetchedAt,
	}, nil
}

func (f *Fetcher) writeCache(rawURL string, fetched FetchedSubscription) error {
	return writeCacheDir(f.opts.CacheDir, rawURL, fetched)
}

// writeSharedCache mirrors a successful fetch into the cross-workspace shared
// directory. It is best effort: failures are ignored because the workspace
// cache remains the authoritative copy.
func (f *Fetcher) writeSharedCache(rawURL string, fetched FetchedSubscription) {
	if strings.TrimSpace(f.opts.SharedCacheDir) == "" || f.opts.SharedCacheDir == f.opts.CacheDir {
		return
	}
	_ = writeCacheDir(f.opts.SharedCacheDir, rawURL, fetched)
}

func writeCacheDir(dir, rawURL string, fetched FetchedSubscription) error {
	if strings.TrimSpace(dir) == "" {
		return nil
	}

	bodyPath, metaPath := cachePaths(dir, rawURL)
	if err := storage.AtomicWriteFile(bodyPath, fetched.Content, 0o600); err != nil {
		return err
	}

	meta := CacheMeta{
		URLHash:              urlHash(rawURL),
		FetchedAt:            fetched.FetchedAt,
		StatusCode:           http.StatusOK,
		Size:                 len(fetched.Content),
		ContentType:          fetched.ContentType,
		SubscriptionUserinfo: fetched.SubscriptionUserinfo,
		OriginalURL:          rawURL,
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache meta: %w", err)
	}
	data = append(data, '\n')

	return storage.AtomicWriteFile(metaPath, data, 0o600)
}

func cachePaths(cacheDir, rawURL string) (string, string) {
	hash := urlHash(rawURL)
	return filepath.Join(cacheDir, hash+".body"), filepath.Join(cacheDir, hash+".meta.json")
}

func urlHash(rawURL string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(rawURL)))
	return hex.EncodeToString(sum[:])
}

func validateFetchURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("unsupported url scheme %q", parsed.Scheme)
	}
	if parsed.Hostname() == "" {
		return nil, fmt.Errorf("missing host")
	}
	if parsed.User != nil {
		return nil, fmt.Errorf("url credentials are not allowed")
	}
	return parsed, nil
}

// ValidatePublicURL applies the public converter's outbound Web-port policy.
// DNS and IP safety are checked immediately before each network dial.
func ValidatePublicURL(raw string) error {
	parsed, err := validateFetchURL(raw)
	if err != nil {
		return err
	}
	if !publicFetchPortAllowed(parsed) {
		return fmt.Errorf("outbound port %s is not allowed in public converter mode", effectivePort(parsed))
	}
	return nil
}

func publicFetchPortAllowed(target *url.URL) bool {
	if target == nil {
		return false
	}
	port := effectivePort(target)
	switch strings.ToLower(strings.TrimSpace(target.Scheme)) {
	case "http":
		switch port {
		case "80", "8080", "8880", "2052", "2082", "2086", "2095":
			return true
		}
	case "https":
		switch port {
		case "443", "8443", "2053", "2083", "2087", "2096":
			return true
		}
	}
	return false
}

func effectivePort(target *url.URL) string {
	if port := target.Port(); port != "" {
		return port
	}
	if target.Scheme == "https" {
		return "443"
	}
	return "80"
}

func isRedirectStatus(statusCode int) bool {
	return statusCode == http.StatusMovedPermanently ||
		statusCode == http.StatusFound ||
		statusCode == http.StatusSeeOther ||
		statusCode == http.StatusTemporaryRedirect ||
		statusCode == http.StatusPermanentRedirect
}

func readLimited(body io.Reader, maxBytes int64) ([]byte, error) {
	limited := io.LimitReader(body, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("response body exceeded limit of %d bytes", maxBytes)
	}
	return data, nil
}

func userAgentOrDefault(value string) string {
	if strings.TrimSpace(value) == "" {
		return model.DefaultUserAgent
	}
	return value
}

var emptyClashYAMLPattern = regexp.MustCompile(`(?mi)^proxies:\s*\[\s*\]\s*$`)

func shouldRetryWithGenericUserAgent(source Source, content []byte) bool {
	ua := strings.ToLower(strings.TrimSpace(source.UserAgent))
	if ua == "" || ua == strings.ToLower(model.DefaultUserAgent) {
		return false
	}
	if !strings.Contains(ua, "clash") && !strings.Contains(ua, "mihomo") && !strings.Contains(ua, "meta") && !strings.Contains(ua, "stash") && !strings.Contains(ua, "sing-box") {
		return false
	}

	body := string(content)
	lower := strings.ToLower(body)
	if !strings.Contains(lower, "proxy-groups:") && !strings.Contains(lower, "rules:") {
		return false
	}
	return emptyClashYAMLPattern.MatchString(body)
}
