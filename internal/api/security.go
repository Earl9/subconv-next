package api

import (
	"crypto/subtle"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	managementRateWindow           = time.Minute
	managementRateLimit            = 240
	expensiveRateLimit             = 30
	authFailureRateLimit           = 20
	anonymousRateLimit             = 120
	anonymousExpensiveRateLimit    = 20
	anonymousWorkspaceRateLimit    = 6
	publishedSubscriptionRateLimit = 120
	requestRateLimiterMaxKeys      = 10000
)

// requestRateLimiter is deliberately small and in-memory. It protects the
// process from accidental bursts; deployments that need distributed limits
// should enforce them at the reverse proxy as well.
type requestRateLimiter struct {
	mu          sync.Mutex
	entries     map[string][]time.Time
	lastCleanup time.Time
}

func newRequestRateLimiter() *requestRateLimiter {
	return &requestRateLimiter{entries: make(map[string][]time.Time)}
}

func (l *requestRateLimiter) allow(key string, limit int, now time.Time) bool {
	if l == nil || limit <= 0 {
		return true
	}
	cutoff := now.Add(-managementRateWindow)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lastCleanup.IsZero() || now.Sub(l.lastCleanup) >= managementRateWindow {
		for entryKey, entryRequests := range l.entries {
			last := len(entryRequests) - 1
			if last < 0 || entryRequests[last].Before(cutoff) {
				delete(l.entries, entryKey)
			}
		}
		l.lastCleanup = now
	}
	if _, exists := l.entries[key]; !exists && len(l.entries) >= requestRateLimiterMaxKeys {
		return false
	}
	requests := l.entries[key]
	first := 0
	for first < len(requests) && requests[first].Before(cutoff) {
		first++
	}
	requests = requests[first:]
	if len(requests) >= limit {
		l.entries[key] = requests
		return false
	}
	l.entries[key] = append(requests, now)
	return true
}

func (s *Server) securityMiddleware(next http.Handler) http.Handler {
	limiter := newRequestRateLimiter()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		service := s.snapshotConfig().Service
		setSecurityHeaders(w, r, service.TrustProxyHeaders)
		if strings.HasPrefix(r.URL.Path, "/api/") && (service.PublicConverter || isUnsafeMethod(r.Method)) && !sameOriginRequest(r, service.PublicBaseURL, service.TrustProxyHeaders) {
			writeAPIError(w, http.StatusForbidden, "CROSS_ORIGIN_REQUEST", "cross-origin API request rejected")
			return
		}
		if r.URL.Path == "/api/auth/login" {
			if !limiter.allow(s.clientHost(r)+"\x00login", authFailureRateLimit, time.Now()) {
				w.Header().Set("Retry-After", "60")
				writeAPIError(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many login attempts; retry later")
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/s/") {
			if !limiter.allow(s.clientHost(r)+"\x00published-subscription", publishedSubscriptionRateLimit, time.Now()) {
				w.Header().Set("Retry-After", "60")
				writeAPIError(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many subscription requests; retry later")
				return
			}
		}
		if isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		if service.PublicConverter && isPublicConverterPath(r.URL.Path) {
			if strings.HasPrefix(r.URL.Path, "/api/") && !s.allowAnonymousAPIRequest(limiter, r) {
				w.Header().Set("Retry-After", "60")
				writeAPIError(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many requests; retry later")
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		authKind := s.managementAuthorization(r)
		if authKind == managementAuthNone {
			if !limiter.allow(s.clientHost(r)+"\x00auth", authFailureRateLimit, time.Now()) {
				w.Header().Set("Retry-After", "60")
				writeAPIError(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many authentication attempts; retry later")
				return
			}
			writeManagementUnauthorized(w, r)
			return
		}
		if authKind == managementAuthSession && strings.HasPrefix(r.URL.Path, "/api/") && isUnsafeMethod(r.Method) && !s.validManagementCSRF(r) {
			writeAPIError(w, http.StatusForbidden, "CSRF_TOKEN_INVALID", "management CSRF token is missing or invalid")
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/") && s.publiclyBound() {
			limit := managementRateLimit
			if isExpensiveAPIPath(r.URL.Path) {
				limit = expensiveRateLimit
			}
			key := s.clientHost(r) + "\x00" + rateLimitClass(r.URL.Path)
			if !limiter.allow(key, limit, time.Now()) {
				w.Header().Set("Retry-After", "60")
				writeAPIError(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many requests; retry later")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func writeManagementUnauthorized(w http.ResponseWriter, r *http.Request) {
	if (r.Method == http.MethodGet || r.Method == http.MethodHead) && r.URL.Path == "/" {
		next := r.URL.RequestURI()
		http.Redirect(w, r, "/login?next="+url.QueryEscape(next), http.StatusFound)
		return
	}
	w.Header().Set("WWW-Authenticate", `Bearer realm="subconv-api"`)
	writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "management login is required")
}

func setSecurityHeaders(w http.ResponseWriter, r *http.Request, trustProxyHeaders bool) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Permissions-Policy", "camera=(), geolocation=(), microphone=(), payment=(), usb=()")
	w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
	w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'; object-src 'none'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self'; font-src 'self' data:")
	if requestTransportUsesHTTPS(r, trustProxyHeaders) {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}
	if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/s/") || r.URL.Path == "/sub/mihomo.yaml" {
		w.Header().Set("Cache-Control", "no-store")
	}
	if r.URL.Path == "/" || r.URL.Path == "/login" || r.URL.Path == "/login.html" {
		w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
	}
}

func requestAccessToken(r *http.Request) string {
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(authorization) >= len("Bearer ") && strings.EqualFold(authorization[:len("Bearer ")], "Bearer ") {
		return strings.TrimSpace(authorization[len("Bearer "):])
	}
	if _, password, ok := r.BasicAuth(); ok {
		return strings.TrimSpace(password)
	}
	return strings.TrimSpace(r.Header.Get("X-SubConv-Access-Token"))
}

func constantTimeEqual(provided, expected string) bool {
	if len(provided) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

func (s *Server) publiclyBound() bool {
	cfg := s.snapshotConfig()
	if cfg.Service.AllowLAN {
		return true
	}
	host := strings.TrimSpace(cfg.Service.ListenAddr)
	if host == "" {
		return true
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	if ip != nil {
		return !ip.IsLoopback()
	}
	return !strings.EqualFold(host, "localhost")
}

func (s *Server) clientHost(r *http.Request) string {
	if s.snapshotConfig().Service.TrustProxyHeaders {
		if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
			candidate := strings.TrimSpace(strings.Split(forwarded, ",")[0])
			if ip := net.ParseIP(strings.Trim(candidate, "[]")); ip != nil {
				return ip.String()
			}
		}
		if candidate := strings.TrimSpace(r.Header.Get("X-Real-IP")); candidate != "" {
			if ip := net.ParseIP(strings.Trim(candidate, "[]")); ip != nil {
				return ip.String()
			}
		}
	}
	return remoteHost(r)
}

func remoteHost(r *http.Request) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func isExpensiveAPIPath(path string) bool {
	switch path {
	case "/api/workspaces", "/api/parse", "/api/generate", "/api/refresh", "/api/site-logo", "/api/update-check":
		return true
	default:
		return false
	}
}

func rateLimitClass(path string) string {
	if isExpensiveAPIPath(path) {
		return "expensive"
	}
	return "management"
}

func (s *Server) allowAnonymousAPIRequest(limiter *requestRateLimiter, r *http.Request) bool {
	limit := anonymousRateLimit
	class := "anonymous"
	if r.URL.Path == "/api/workspaces" && r.Method == http.MethodPost {
		limit = anonymousWorkspaceRateLimit
		class = "anonymous-workspace"
	} else if isExpensiveAPIPath(r.URL.Path) {
		limit = anonymousExpensiveRateLimit
		class = "anonymous-expensive"
	}
	return limiter.allow(s.clientHost(r)+"\x00"+class, limit, time.Now())
}

func isPublicConverterPath(path string) bool {
	switch path {
	case "/", "/style.css", "/app.js", "/favicon.svg", "/favicon.ico",
		"/api/status", "/api/workspaces", "/api/published", "/api/config", "/api/custom-rules/validate",
		"/api/site-logo", "/api/subscription-meta", "/api/audit",
		"/api/preview-yaml", "/api/validate-output", "/api/nodes",
		"/api/parse", "/api/generate", "/api/refresh", "/api/logs",
		"/api/update-check":
		return true
	default:
		return strings.HasPrefix(path, "/api/workspaces/") ||
			strings.HasPrefix(path, "/api/published/") ||
			strings.HasPrefix(path, "/api/nodes/")
	}
}

func isPublicPath(path string) bool {
	switch path {
	case "/healthz", "/sub/mihomo.yaml", "/login", "/login.html", "/login.css", "/login.js", "/favicon.svg", "/favicon.ico", "/api/auth/session":
		return true
	default:
		return strings.HasPrefix(path, "/s/")
	}
}

func isUnsafeMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func sameOriginRequest(r *http.Request, publicBaseURL string, trustProxyHeaders bool) bool {
	if strings.EqualFold(strings.TrimSpace(r.Header.Get("Sec-Fetch-Site")), "cross-site") {
		return false
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil {
		return false
	}

	expected, err := url.Parse(strings.TrimSpace(publicBaseURL))
	if err != nil || expected.Scheme == "" || expected.Host == "" {
		scheme := "http"
		if requestTransportUsesHTTPS(r, trustProxyHeaders) {
			scheme = "https"
		}
		expected = &url.URL{Scheme: scheme, Host: r.Host}
	}
	return strings.EqualFold(parsed.Scheme, expected.Scheme) && sameOriginAuthority(parsed, expected)
}

func requestTransportUsesHTTPS(r *http.Request, trustProxyHeaders bool) bool {
	if r.TLS != nil {
		return true
	}
	if !trustProxyHeaders {
		return false
	}
	forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	proto := strings.TrimSpace(strings.Split(forwarded, ",")[0])
	return strings.EqualFold(proto, "https")
}

func sameOriginAuthority(left, right *url.URL) bool {
	if !strings.EqualFold(left.Hostname(), right.Hostname()) {
		return false
	}
	leftPort := left.Port()
	rightPort := right.Port()
	if leftPort == "" {
		leftPort = defaultHTTPPort(left.Scheme)
	}
	if rightPort == "" {
		rightPort = defaultHTTPPort(right.Scheme)
	}
	return leftPort == rightPort
}

func defaultHTTPPort(scheme string) string {
	if strings.EqualFold(scheme, "https") {
		return "443"
	}
	return "80"
}
