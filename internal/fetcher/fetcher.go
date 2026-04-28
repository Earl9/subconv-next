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
}

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
	Timeout           time.Duration
	MaxBodyBytes      int64
	MaxRedirects      int
	CacheTTL          time.Duration
	AllowPrivateHosts bool
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
		CacheDir:     cfg.Service.CacheDir,
		Timeout:      time.Duration(cfg.Service.FetchTimeoutSeconds) * time.Second,
		MaxBodyBytes: int64(cfg.Service.MaxSubscriptionBytes),
		MaxRedirects: 3,
		CacheTTL:     time.Duration(cfg.Service.RefreshInterval) * time.Second,
	}
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
		opts.Resolver = net.DefaultResolver
	}
	return &Fetcher{opts: opts}
}

func (f *Fetcher) Fetch(ctx context.Context, source Source) (FetchedSubscription, []string, error) {
	if !source.Enabled {
		return FetchedSubscription{}, nil, errors.New("source is disabled")
	}

	currentURL, err := validateFetchURL(source.URL)
	if err != nil {
		return FetchedSubscription{}, nil, err
	}

	fetched, err := f.fetchNetwork(ctx, currentURL, source)
	if err == nil {
		if cacheErr := f.writeCache(source.URL, fetched); cacheErr != nil {
			return fetched, []string{fmt.Sprintf("cache write failed for %s: %v", source.Name, cacheErr)}, nil
		}
		return fetched, nil, nil
	}

	cached, cacheErr := f.readCache(source.URL)
	if cacheErr == nil {
		cached.FromCache = true
		return cached, []string{fmt.Sprintf("using cached subscription for %s after fetch failure", source.Name)}, nil
	}

	return FetchedSubscription{}, nil, err
}

func (f *Fetcher) fetchNetwork(ctx context.Context, currentURL *url.URL, source Source) (FetchedSubscription, error) {
	redirects := 0

	for {
		resolvedIP, err := f.resolveHost(ctx, currentURL.Hostname())
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
	return client.Do(req)
}

func (f *Fetcher) resolveHost(ctx context.Context, host string) (net.IP, error) {
	if isBlockedHostname(host) && !f.opts.AllowPrivateHosts {
		return nil, fmt.Errorf("blocked host %q", host)
	}

	ips, err := f.opts.Resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("resolve host %q: %w", host, err)
	}

	for _, ipAddr := range ips {
		if f.opts.AllowPrivateHosts || !isBlockedIP(ipAddr.IP) {
			return ipAddr.IP, nil
		}
	}

	return nil, fmt.Errorf("no allowed IPs for host %q", host)
}

func (f *Fetcher) readCache(rawURL string) (FetchedSubscription, error) {
	bodyPath, metaPath := cachePaths(f.opts.CacheDir, rawURL)

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

	return FetchedSubscription{
		URL:                  rawURL,
		Content:              body,
		ContentType:          meta.ContentType,
		SubscriptionUserinfo: meta.SubscriptionUserinfo,
		FetchedAt:            meta.FetchedAt,
	}, nil
}

func (f *Fetcher) writeCache(rawURL string, fetched FetchedSubscription) error {
	if strings.TrimSpace(f.opts.CacheDir) == "" {
		return nil
	}

	bodyPath, metaPath := cachePaths(f.opts.CacheDir, rawURL)
	if err := storage.AtomicWriteFile(bodyPath, fetched.Content, 0o644); err != nil {
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

	return storage.AtomicWriteFile(metaPath, data, 0o644)
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
	return parsed, nil
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
