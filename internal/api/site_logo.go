package api

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"subconv-next/internal/model"
)

type siteLogoResponse struct {
	OK      bool   `json:"ok"`
	LogoURL string `json:"logoUrl,omitempty"`
	Domain  string `json:"domain,omitempty"`
	Source  string `json:"source,omitempty"`
}

type siteLogoCacheEntry struct {
	Payload   siteLogoResponse
	ExpiresAt time.Time
}

type iconCandidate struct {
	URL    *url.URL
	Source string
}

var linkTagPattern = regexp.MustCompile(`(?is)<link\b[^>]*>`)

const maxSiteLogoCacheEntries = 256

type siteLogoResolver interface {
	LookupIPAddr(context.Context, string) ([]net.IPAddr, error)
}

func (s *Server) handleSiteLogo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	rawURL := strings.TrimSpace(r.URL.Query().Get("url"))
	if rawURL == "" {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "url is required")
		return
	}

	payload, err := s.resolveSiteLogo(rawURL)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) resolveSiteLogo(rawURL string) (siteLogoResponse, error) {
	parsed, err := parseSiteLogoURL(rawURL)
	if err != nil {
		return siteLogoResponse{}, err
	}

	domain := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if domain == "" {
		return siteLogoResponse{}, fmt.Errorf("missing host")
	}

	cacheKey := strings.ToLower(parsed.Scheme + "://" + parsed.Host)
	if cached, ok := s.lookupSiteLogoCache(cacheKey); ok {
		return cached, nil
	}

	payload := siteLogoResponse{
		OK:     true,
		Domain: domain,
		Source: "fallback",
	}

	if !siteLogoHostAllowed(domain) {
		s.storeSiteLogoCache(cacheKey, payload)
		return payload, nil
	}

	home := &url.URL{Scheme: parsed.Scheme, Host: parsed.Host}
	if logoURL, source, ok := fetchBestSiteLogo(home); ok {
		payload.LogoURL = logoURL
		payload.Source = source
	}

	s.storeSiteLogoCache(cacheKey, payload)
	return payload, nil
}

func (s *Server) lookupSiteLogoCache(key string) (siteLogoResponse, bool) {
	s.siteLogoMu.RLock()
	entry, ok := s.siteLogoCache[key]
	s.siteLogoMu.RUnlock()
	if !ok {
		return siteLogoResponse{}, false
	}
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		s.siteLogoMu.Lock()
		delete(s.siteLogoCache, key)
		s.siteLogoMu.Unlock()
		return siteLogoResponse{}, false
	}
	return entry.Payload, true
}

func (s *Server) storeSiteLogoCache(key string, payload siteLogoResponse) {
	s.siteLogoMu.Lock()
	now := time.Now()
	if len(s.siteLogoCache) >= maxSiteLogoCacheEntries {
		for cacheKey, entry := range s.siteLogoCache {
			if !entry.ExpiresAt.IsZero() && now.After(entry.ExpiresAt) {
				delete(s.siteLogoCache, cacheKey)
			}
		}
	}
	if len(s.siteLogoCache) >= maxSiteLogoCacheEntries {
		for cacheKey := range s.siteLogoCache {
			delete(s.siteLogoCache, cacheKey)
			break
		}
	}
	s.siteLogoCache[key] = siteLogoCacheEntry{
		Payload:   payload,
		ExpiresAt: now.Add(12 * time.Hour),
	}
	s.siteLogoMu.Unlock()
}

func parseSiteLogoURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(parsed.Scheme)) {
	case "http", "https":
	default:
		return nil, fmt.Errorf("unsupported url scheme")
	}
	if strings.TrimSpace(parsed.Hostname()) == "" {
		return nil, fmt.Errorf("missing host")
	}
	if parsed.User != nil {
		return nil, fmt.Errorf("url credentials are not allowed")
	}
	return parsed, nil
}

func siteLogoHostAllowed(host string) bool {
	return siteLogoHostAllowedWithResolver(context.Background(), net.DefaultResolver, host)
}

func siteLogoHostAllowedWithResolver(ctx context.Context, resolver siteLogoResolver, host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" || host == "localhost" || strings.HasSuffix(host, ".local") || resolver == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	ips, err := resolver.LookupIPAddr(ctx, host)
	if err != nil || len(ips) == 0 {
		return false
	}
	for _, ipAddr := range ips {
		ip := ipAddr.IP
		if ip == nil || ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() || ip.IsMulticast() || ip.IsUnspecified() {
			return false
		}
	}
	return true
}

func fetchBestSiteLogo(home *url.URL) (string, string, bool) {
	client := newSiteLogoHTTPClient(net.DefaultResolver)
	defer client.CloseIdleConnections()
	html, contentType, ok := fetchSiteDocument(client, home)
	if ok && strings.Contains(strings.ToLower(contentType), "html") {
		for _, candidate := range discoverSiteLogoCandidates(home, html) {
			if logo, ok := fetchLogoAsDataURL(client, candidate.URL); ok {
				return logo, candidate.Source, true
			}
		}
	}

	faviconURL := home.ResolveReference(&url.URL{Path: "/favicon.ico"})
	if logo, ok := fetchLogoAsDataURL(client, faviconURL); ok {
		return logo, "favicon", true
	}
	return "", "", false
}

func newSiteLogoHTTPClient(resolver siteLogoResolver) *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	transport := &http.Transport{
		Proxy:                 nil,
		ForceAttemptHTTP2:     true,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
	}
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		target, err := resolveSiteLogoDialTarget(ctx, resolver, address)
		if err != nil {
			return nil, err
		}
		return dialer.DialContext(ctx, network, target)
	}
	transport.DialTLSContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		target, err := resolveSiteLogoDialTarget(ctx, resolver, address)
		if err != nil {
			return nil, err
		}
		conn, err := dialer.DialContext(ctx, network, target)
		if err != nil {
			return nil, err
		}
		tlsConn := tls.Client(conn, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			_ = conn.Close()
			return nil, err
		}
		return tlsConn, nil
	}

	return &http.Client{
		Transport: transport,
		Timeout:   8 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("too many redirects")
			}
			if err := validateSiteLogoTarget(req.Context(), resolver, req.URL); err != nil {
				return err
			}
			return nil
		},
	}
}

func validateSiteLogoTarget(ctx context.Context, resolver siteLogoResolver, target *url.URL) error {
	if target == nil || (target.Scheme != "http" && target.Scheme != "https") || target.User != nil {
		return fmt.Errorf("unsafe site logo target")
	}
	if !siteLogoHostAllowedWithResolver(ctx, resolver, target.Hostname()) {
		return fmt.Errorf("site logo target is not publicly routable")
	}
	return nil
}

func resolveSiteLogoDialTarget(ctx context.Context, resolver siteLogoResolver, address string) (string, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", err
	}
	if !siteLogoHostAllowedWithResolver(ctx, resolver, host) {
		return "", fmt.Errorf("site logo target is not publicly routable")
	}
	ips, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return "", fmt.Errorf("resolve site logo host: %w", err)
	}
	if len(ips) == 0 {
		return "", fmt.Errorf("resolve site logo host: no addresses")
	}
	for _, item := range ips {
		if !siteLogoIPBlocked(item.IP) {
			return net.JoinHostPort(item.IP.String(), port), nil
		}
	}
	return "", fmt.Errorf("site logo target is not publicly routable")
}

func siteLogoIPBlocked(ip net.IP) bool {
	return ip == nil || ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalMulticast() ||
		ip.IsLinkLocalUnicast() || ip.IsMulticast() || ip.IsUnspecified()
}

func fetchSiteDocument(client *http.Client, target *url.URL) ([]byte, string, bool) {
	req, err := http.NewRequest(http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, "", false
	}
	req.Header.Set("User-Agent", model.DefaultUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", false
	}
	body, err := readSiteLogoBody(resp.Body, 512*1024)
	if err != nil {
		return nil, "", false
	}
	return body, resp.Header.Get("Content-Type"), true
}

func discoverSiteLogoCandidates(base *url.URL, html []byte) []iconCandidate {
	tags := linkTagPattern.FindAllString(string(html), -1)
	candidates := make([]iconCandidate, 0, len(tags))
	seen := map[string]struct{}{}

	for _, tag := range tags {
		rel := strings.ToLower(strings.TrimSpace(extractHTMLAttribute(tag, "rel")))
		href := strings.TrimSpace(extractHTMLAttribute(tag, "href"))
		if href == "" || rel == "" {
			continue
		}
		source := ""
		switch {
		case strings.Contains(rel, "apple-touch-icon"):
			source = "apple-touch-icon"
		case strings.Contains(rel, "shortcut icon"), strings.Contains(rel, "icon"):
			source = "favicon"
		default:
			continue
		}
		parsed, err := url.Parse(href)
		if err != nil {
			continue
		}
		resolved := base.ResolveReference(parsed)
		key := resolved.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		candidates = append(candidates, iconCandidate{URL: resolved, Source: source})
	}
	return candidates
}

func extractHTMLAttribute(tag, attr string) string {
	pattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(attr) + `\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+))`)
	match := pattern.FindStringSubmatch(tag)
	for i := 1; i < len(match); i++ {
		if strings.TrimSpace(match[i]) != "" {
			return match[i]
		}
	}
	return ""
}

func fetchLogoAsDataURL(client *http.Client, target *url.URL) (string, bool) {
	if target == nil {
		return "", false
	}
	req, err := http.NewRequest(http.MethodGet, target.String(), nil)
	if err != nil {
		return "", false
	}
	req.Header.Set("User-Agent", model.DefaultUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", false
	}

	body, err := readSiteLogoBody(resp.Body, 256*1024)
	if err != nil || len(body) == 0 {
		return "", false
	}

	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = http.DetectContentType(body)
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || !strings.HasPrefix(strings.ToLower(mediaType), "image/") {
		return "", false
	}
	return "data:" + mediaType + ";base64," + base64.StdEncoding.EncodeToString(body), true
}

func readSiteLogoBody(reader io.Reader, limit int64) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("site logo response exceeds size limit")
	}
	return body, nil
}
