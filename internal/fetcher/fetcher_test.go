package fetcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"subconv-next/internal/model"
)

func TestFetchSuccessAndCacheFallback(t *testing.T) {
	callCount := 0

	f := New(Options{
		CacheDir:          t.TempDir(),
		Timeout:           5 * time.Second,
		MaxBodyBytes:      1024,
		MaxRedirects:      3,
		AllowPrivateHosts: true,
		Resolver: staticResolver{
			ips: []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}},
		},
		RequestDoer: func(ctx context.Context, target *url.URL, resolvedIP net.IP, source Source) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"text/plain"},
					},
					Body: io.NopCloser(strings.NewReader("ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#cached")),
				}, nil
			}
			return nil, errors.New("network down")
		},
	})

	source := Source{
		Name:      "demo",
		URL:       "https://example.com/subscription",
		UserAgent: "SubConvNext/0.1",
		Enabled:   true,
	}

	first, warnings, err := f.Fetch(context.Background(), source)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %#v, want none", warnings)
	}
	if first.FromCache {
		t.Fatalf("first.FromCache = true, want false")
	}

	second, warnings, err := f.Fetch(context.Background(), source)
	if err != nil {
		t.Fatalf("Fetch() cache fallback error = %v", err)
	}
	if !second.FromCache {
		t.Fatalf("second.FromCache = false, want true")
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings = %#v, want one cache warning", warnings)
	}
}

func TestFetchClosesPerRequestTransportConnections(t *testing.T) {
	var openConnections atomic.Int64
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("subscription"))
	}))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		switch state {
		case http.StateNew:
			openConnections.Add(1)
		case http.StateClosed:
			openConnections.Add(-1)
		}
	}
	server.Start()
	defer server.Close()

	f := New(Options{
		CacheDir:          t.TempDir(),
		Timeout:           2 * time.Second,
		MaxBodyBytes:      1024,
		AllowPrivateHosts: true,
		Resolver: staticResolver{
			ips: []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}},
		},
	})
	source := Source{Name: "transport-lifecycle", URL: server.URL, Enabled: true, AllowPrivateHosts: true}
	for i := 0; i < 20; i++ {
		if _, _, err := f.Fetch(context.Background(), source); err != nil {
			t.Fatalf("Fetch() error on request %d: %v", i, err)
		}
	}

	deadline := time.Now().Add(time.Second)
	for openConnections.Load() != 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if got := openConnections.Load(); got != 0 {
		t.Fatalf("open connections after fetches = %d, want 0", got)
	}
}

func TestFetchUsesFreshCacheBeforeNetwork(t *testing.T) {
	callCount := 0
	f := New(Options{
		CacheDir:     t.TempDir(),
		MaxBodyBytes: 1024,
		Resolver: staticResolver{
			ips: []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}},
		},
		RequestDoer: func(context.Context, *url.URL, net.IP, Source) (*http.Response, error) {
			callCount++
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("cached body")),
			}, nil
		},
	})
	source := Source{
		Name:     "rules",
		URL:      "https://example.com/rules.txt",
		Enabled:  true,
		CacheTTL: time.Hour,
	}

	first, _, err := f.Fetch(context.Background(), source)
	if err != nil {
		t.Fatalf("first Fetch() error = %v", err)
	}
	second, warnings, err := f.Fetch(context.Background(), source)
	if err != nil {
		t.Fatalf("second Fetch() error = %v", err)
	}
	if callCount != 1 {
		t.Fatalf("network calls = %d, want 1", callCount)
	}
	if first.FromCache || !second.FromCache || string(second.Content) != "cached body" {
		t.Fatalf("cache results = first:%+v second:%+v", first, second)
	}
	if !strings.Contains(strings.Join(warnings, "\n"), "fresh cached content") {
		t.Fatalf("warnings = %#v, want fresh cache diagnostic", warnings)
	}
}

func TestFetchRejectsCacheBodyMetadataMismatch(t *testing.T) {
	callCount := 0
	cacheDir := t.TempDir()
	f := New(Options{
		CacheDir:     cacheDir,
		MaxBodyBytes: 1024,
		Resolver:     staticResolver{ips: []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}},
		RequestDoer: func(context.Context, *url.URL, net.IP, Source) (*http.Response, error) {
			callCount++
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("network body")),
			}, nil
		},
	})
	source := Source{Name: "rules", URL: "https://example.com/rules.txt", Enabled: true, CacheTTL: time.Hour}
	if _, _, err := f.Fetch(context.Background(), source); err != nil {
		t.Fatalf("first Fetch() error = %v", err)
	}
	bodyPath, _ := cachePaths(cacheDir, source.URL)
	if err := os.WriteFile(bodyPath, []byte("tampered body with a different size"), 0600); err != nil {
		t.Fatal(err)
	}
	second, _, err := f.Fetch(context.Background(), source)
	if err != nil {
		t.Fatalf("second Fetch() error = %v", err)
	}
	if callCount != 2 || second.FromCache || string(second.Content) != "network body" {
		t.Fatalf("cache mismatch was reused: calls=%d result=%+v", callCount, second)
	}
}

func TestFetchPrefersExpiredCacheAndStillAllowsRefresh(t *testing.T) {
	calls := 0
	f := New(Options{
		CacheDir: t.TempDir(),
		Resolver: staticResolver{ips: []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}},
		RequestDoer: func(context.Context, *url.URL, net.IP, Source) (*http.Response, error) {
			calls++
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("updated body")),
			}, nil
		},
	})
	source := Source{Name: "subscription", URL: "https://example.com/sub", Enabled: true, PreferCache: true}
	if err := f.writeCache(source.URL, FetchedSubscription{
		Content: []byte("cached body"), FetchedAt: time.Now().Add(-24 * time.Hour),
	}); err != nil {
		t.Fatal(err)
	}
	result, _, err := f.Fetch(context.Background(), source)
	if err != nil || calls != 0 || !result.FromCache || string(result.Content) != "cached body" {
		t.Fatalf("cached read = %+v, calls = %d, err = %v", result, calls, err)
	}

	source.PreferCache = false
	result, _, err = f.Fetch(context.Background(), source)
	if err != nil || calls != 1 || result.FromCache || string(result.Content) != "updated body" {
		t.Fatalf("refresh = %+v, calls = %d, err = %v", result, calls, err)
	}

	source.PreferCache = true
	source.URL = "https://example.com/new-sub"
	if _, _, err := f.Fetch(context.Background(), source); !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("cache miss error = %v, want ErrCacheMiss", err)
	}
	if calls != 1 {
		t.Fatalf("network calls after cache miss = %d, want 1: offline views must not fetch", calls)
	}
}

func TestFetchPrefersCacheFallsBackToSharedCopy(t *testing.T) {
	workspaceCache := t.TempDir()
	sharedCache := t.TempDir()
	f := New(Options{
		CacheDir:       workspaceCache,
		SharedCacheDir: sharedCache,
		MaxBodyBytes:   1024,
		RequestDoer: func(context.Context, *url.URL, net.IP, Source) (*http.Response, error) {
			t.Error("network fetch is not allowed when the shared cache has the content")
			return nil, fmt.Errorf("network must not be used")
		},
	})
	source := Source{Name: "rules", URL: "https://example.com/rules.txt", Enabled: true, PreferCache: true}
	if err := writeCacheDir(sharedCache, source.URL, FetchedSubscription{
		Content:   []byte("shared body"),
		FetchedAt: time.Now().Add(-time.Hour),
	}); err != nil {
		t.Fatal(err)
	}

	cached, _, err := f.Fetch(context.Background(), source)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if !cached.FromCache || string(cached.Content) != "shared body" {
		t.Fatalf("Fetch() = %+v, want shared cache content", cached)
	}

	if _, _, err := f.Fetch(context.Background(), Source{
		Name: source.Name, URL: "https://example.com/missing.txt", Enabled: true, PreferCache: true,
	}); !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("missing everywhere error = %v, want ErrCacheMiss", err)
	}
}

func TestFetchWritesThroughToSharedCache(t *testing.T) {
	workspaceCache := t.TempDir()
	sharedCache := t.TempDir()
	f := New(Options{
		CacheDir:       workspaceCache,
		SharedCacheDir: sharedCache,
		Resolver:       staticResolver{ips: []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}},
		RequestDoer: func(context.Context, *url.URL, net.IP, Source) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("fresh body")),
			}, nil
		},
	})
	source := Source{Name: "rules", URL: "https://example.com/rules.txt", Enabled: true}
	if _, _, err := f.Fetch(context.Background(), source); err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	sharedPath, _ := cachePaths(sharedCache, source.URL)
	body, err := os.ReadFile(sharedPath)
	if err != nil {
		t.Fatalf("shared cache was not written: %v", err)
	}
	if string(body) != "fresh body" {
		t.Fatalf("shared cache body = %q, want fresh body", string(body))
	}

	// A second fetcher simulating a fresh draft workspace reads the shared copy.
	draft := New(Options{
		CacheDir:       t.TempDir(),
		SharedCacheDir: sharedCache,
		MaxBodyBytes:   1024,
	})
	draftSource := Source{Name: "rules", URL: source.URL, Enabled: true, PreferCache: true}
	cached, _, err := draft.Fetch(context.Background(), draftSource)
	if err != nil || !cached.FromCache || string(cached.Content) != "fresh body" {
		t.Fatalf("draft Fetch() = %+v, err = %v, want shared content", cached, err)
	}
}

func TestFetchCapturesSubscriptionUserinfoAndDefaultUserAgent(t *testing.T) {
	var gotUserAgent string
	f := New(Options{
		CacheDir:          t.TempDir(),
		Timeout:           5 * time.Second,
		MaxBodyBytes:      1024,
		MaxRedirects:      3,
		AllowPrivateHosts: true,
		Resolver: staticResolver{
			ips: []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}},
		},
		RequestDoer: func(ctx context.Context, target *url.URL, resolvedIP net.IP, source Source) (*http.Response, error) {
			gotUserAgent = userAgentOrDefault(source.UserAgent)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type":          []string{"text/plain"},
					"Subscription-Userinfo": []string{"upload=100; download=200; total=1000; expire=1745942400"},
				},
				Body: io.NopCloser(strings.NewReader("ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#userinfo")),
			}, nil
		},
	})

	fetched, warnings, err := f.Fetch(context.Background(), Source{
		Name:    "demo",
		URL:     "https://example.com/subscription",
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %#v, want none", warnings)
	}
	if gotUserAgent != model.DefaultUserAgent {
		t.Fatalf("User-Agent = %q, want %q", gotUserAgent, model.DefaultUserAgent)
	}
	if fetched.SubscriptionUserinfo != "upload=100; download=200; total=1000; expire=1745942400" {
		t.Fatalf("SubscriptionUserinfo = %q, want upstream header", fetched.SubscriptionUserinfo)
	}
}

func TestFetchBlocksPrivateHostsByDefault(t *testing.T) {
	f := New(Options{
		CacheDir:     t.TempDir(),
		Timeout:      5 * time.Second,
		MaxBodyBytes: 1024,
		MaxRedirects: 3,
	})

	_, _, err := f.Fetch(context.Background(), Source{
		Name:      "blocked",
		URL:       "http://localhost/private",
		UserAgent: "SubConvNext/0.1",
		Enabled:   true,
	})
	if err == nil {
		t.Fatalf("Fetch() error = nil, want blocked host error")
	}
}

func TestFetchRejectsURLCredentials(t *testing.T) {
	f := New(Options{CacheDir: t.TempDir()})

	_, _, err := f.Fetch(context.Background(), Source{
		Name:    "credentials",
		URL:     "https://user:password@example.com/subscription",
		Enabled: true,
	})
	if err == nil || !strings.Contains(err.Error(), "credentials") {
		t.Fatalf("Fetch() error = %v, want URL credentials error", err)
	}
}

func TestPublicModeRejectsNonWebPortsBeforeDNS(t *testing.T) {
	f := New(Options{
		CacheDir:   t.TempDir(),
		PublicMode: true,
		Resolver: staticResolver{
			ips: []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}},
		},
	})
	_, _, err := f.Fetch(context.Background(), Source{
		Name:    "blocked-port",
		URL:     "https://example.com:22/subscription",
		Enabled: true,
	})
	if err == nil || !strings.Contains(err.Error(), "outbound port 22") {
		t.Fatalf("Fetch() error = %v, want public port rejection", err)
	}
}

func TestValidatePublicURLAllowsCommonAlternateHTTPSPort(t *testing.T) {
	if err := ValidatePublicURL("https://example.com:8443/subscription"); err != nil {
		t.Fatalf("ValidatePublicURL(8443) error = %v", err)
	}
	if err := ValidatePublicURL("http://example.com:6379/subscription"); err == nil {
		t.Fatal("ValidatePublicURL(6379) error = nil, want rejection")
	}
}

func TestFetchAllowsPrivateHostWhenSourceAllowsLAN(t *testing.T) {
	f := New(Options{
		CacheDir:     t.TempDir(),
		Timeout:      5 * time.Second,
		MaxBodyBytes: 1024,
		MaxRedirects: 3,
		Resolver: staticResolver{
			ips: []net.IPAddr{{IP: net.ParseIP("10.0.0.3")}},
		},
		RequestDoer: func(ctx context.Context, target *url.URL, resolvedIP net.IP, source Source) (*http.Response, error) {
			if got := resolvedIP.String(); got != "10.0.0.3" {
				t.Fatalf("resolvedIP = %q, want 10.0.0.3", got)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type": []string{"text/plain"},
				},
				Body: io.NopCloser(strings.NewReader("ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#lan")),
			}, nil
		},
	})

	fetched, _, err := f.Fetch(context.Background(), Source{
		Name:              "lan",
		URL:               "http://10.0.0.3/sub",
		Enabled:           true,
		AllowPrivateHosts: true,
	})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if got := string(fetched.Content); !strings.Contains(got, "#lan") {
		t.Fatalf("fetched.Content = %q, want lan node", got)
	}
}

func TestSharedCacheDirLayout(t *testing.T) {
	cases := map[string]string{
		"/data/cache":                                "/data/cache/shared",
		"/data/cache/workspaces/abc123":              "/data/cache/shared",
		"/data/cache/workspaces/abc123/custom-rules": "/data/cache/workspaces/abc123/custom-rules/shared",
		"": "",
	}
	for cacheDir, want := range cases {
		if got := SharedCacheDir(cacheDir); got != want {
			t.Fatalf("SharedCacheDir(%q) = %q, want %q", cacheDir, got, want)
		}
	}
}

func TestOptionsFromConfigUsesServiceAllowLAN(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Service.AllowLAN = true

	opts := OptionsFromConfig(cfg)

	if !opts.AllowPrivateHosts {
		t.Fatalf("AllowPrivateHosts = false, want true when service.allow_lan is enabled")
	}
	if opts.SharedCacheDir != SharedCacheDir(cfg.Service.CacheDir) || opts.SharedCacheDir == "" {
		t.Fatalf("SharedCacheDir = %q, want derived from cache dir %q", opts.SharedCacheDir, cfg.Service.CacheDir)
	}
}

func TestFetchFallsBackFromEmptyClashYAML(t *testing.T) {
	callUserAgents := []string{}

	f := New(Options{
		CacheDir:          t.TempDir(),
		Timeout:           5 * time.Second,
		MaxBodyBytes:      4096,
		MaxRedirects:      3,
		AllowPrivateHosts: true,
		Resolver: staticResolver{
			ips: []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}},
		},
		RequestDoer: func(ctx context.Context, target *url.URL, resolvedIP net.IP, source Source) (*http.Response, error) {
			callUserAgents = append(callUserAgents, source.UserAgent)
			if strings.EqualFold(strings.TrimSpace(source.UserAgent), "clash") {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"text/plain"},
					},
					Body: io.NopCloser(strings.NewReader("proxies: []\nproxy-groups:\n  - { name: demo, type: select, proxies: [DIRECT] }\nrules:\n  - MATCH,DIRECT\n")),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Content-Type": []string{"text/plain"},
				},
				Body: io.NopCloser(strings.NewReader("ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#fallback")),
			}, nil
		},
	})

	source := Source{
		Name:      "demo",
		URL:       "https://example.com/subscription",
		UserAgent: "clash",
		Enabled:   true,
	}

	fetched, warnings, err := f.Fetch(context.Background(), source)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(callUserAgents) != 2 {
		t.Fatalf("len(callUserAgents) = %d, want 2", len(callUserAgents))
	}
	if callUserAgents[1] != model.DefaultUserAgent {
		t.Fatalf("fallback user-agent = %q, want %q", callUserAgents[1], model.DefaultUserAgent)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "generic user-agent") {
		t.Fatalf("warnings = %#v, want generic user-agent retry warning", warnings)
	}
	if got := string(fetched.Content); !strings.Contains(got, "ss://") {
		t.Fatalf("fetched.Content = %q, want fallback URI content", got)
	}
}

func TestBlockedIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{ip: "127.0.0.1", want: true},
		{ip: "10.0.0.1", want: true},
		{ip: "192.168.1.10", want: true},
		{ip: "100.64.0.1", want: true},
		{ip: "192.0.2.1", want: true},
		{ip: "198.18.0.1", want: true},
		{ip: "203.0.113.1", want: true},
		{ip: "64:ff9b::a00:1", want: true},
		{ip: "2002:0a00:0001::", want: true},
		{ip: "8.8.8.8", want: false},
		{ip: "2606:4700:4700::1111", want: false},
	}

	for _, tt := range tests {
		if got := isBlockedIP(netParseIP(tt.ip)); got != tt.want {
			t.Fatalf("isBlockedIP(%q) = %v, want %v", tt.ip, got, tt.want)
		}
	}
}

func netParseIP(value string) net.IP {
	return net.ParseIP(value)
}

func TestPublicHostResolverBypassesFakeIPDNS(t *testing.T) {
	resolver := publicHostResolver{
		primary: staticResolver{
			ips: []net.IPAddr{{IP: net.ParseIP("198.18.0.125")}},
		},
		fallbacks: []HostResolver{
			staticResolver{ips: []net.IPAddr{{IP: net.ParseIP("104.21.73.232")}}},
		},
	}

	ips, err := resolver.LookupIPAddr(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("LookupIPAddr() error = %v", err)
	}
	if len(ips) != 1 || !ips[0].IP.Equal(net.ParseIP("104.21.73.232")) {
		t.Fatalf("LookupIPAddr() = %#v, want public fallback address", ips)
	}
}

func TestNewUsesPublicDNSFallbackForNormalService(t *testing.T) {
	fetcher := New(Options{})
	if _, ok := fetcher.opts.Resolver.(publicHostResolver); !ok {
		t.Fatalf("default resolver type = %T, want publicHostResolver", fetcher.opts.Resolver)
	}
}

func TestPublicHostResolverKeepsSafePrimaryResult(t *testing.T) {
	resolver := publicHostResolver{
		primary: staticResolver{
			ips: []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}},
		},
		fallbacks: []HostResolver{
			staticResolver{ips: []net.IPAddr{{IP: net.ParseIP("1.1.1.1")}}},
		},
	}

	ips, err := resolver.LookupIPAddr(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("LookupIPAddr() error = %v", err)
	}
	if len(ips) != 1 || !ips[0].IP.Equal(net.ParseIP("8.8.8.8")) {
		t.Fatalf("LookupIPAddr() = %#v, want primary address", ips)
	}
}

type staticResolver struct {
	ips []net.IPAddr
}

func (r staticResolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	return append([]net.IPAddr(nil), r.ips...), nil
}
