package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
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

	cacheKey := parsed.Scheme + "://" + domain
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
	s.siteLogoCache[key] = siteLogoCacheEntry{
		Payload:   payload,
		ExpiresAt: time.Now().Add(12 * time.Hour),
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
	return parsed, nil
}

func siteLogoHostAllowed(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" || host == "localhost" || strings.HasSuffix(host, ".local") {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
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
	client := &http.Client{Timeout: 8 * time.Second}
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
	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
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
	if target == nil || !siteLogoHostAllowed(target.Hostname()) {
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

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil || len(body) == 0 {
		return "", false
	}

	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = http.DetectContentType(body)
	}
	return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(body), true
}
