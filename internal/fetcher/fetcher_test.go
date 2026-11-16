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
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"subconv-next/internal/model"
)

func TestEmptySuccessResponsePreservesValidCache(t *testing.T) {
	for _, empty := range []string{"", " \n\t"} {
		t.Run(fmt.Sprintf("bytes_%d", len(empty)), func(t *testing.T) {
			cacheRoot := t.TempDir()
			f := New(Options{
				CacheDir: filepath.Join(cacheRoot, "workspace"), SharedCacheDir: filepath.Join(cacheRoot, "shared"),
				Resolver: staticResolver{ips: []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}},
				RequestDoer: func(context.Context, *url.URL, net.IP, Source) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(empty))}, nil
				},
			})
			source := Source{Name: "subscription", URL: "https://example.com/sub", Enabled: true}
			if _, _, err := f.Fetch(context.Background(), source); err == nil {
				t.Fatal("empty response without a cache was accepted")
			}
			good := FetchedSubscription{Content: []byte("trojan://test@keep.example:443#keep"), FetchedAt: time.Now()}
			if err := f.writeCache(source.URL, good); err != nil {
				t.Fatal(err)
			}
			f.writeSharedCache(source.URL, good)
			result, warnings, err := f.Fetch(context.Background(), source)
			if err != nil || !result.FromCache || string(result.Content) != string(good.Content) || len(warnings) == 0 {
				t.Fatal("empty response did not fall back to valid cached nodes")
			}
			cached, err := f.readCache(source.URL)
			if err != nil || string(cached.Content) != string(good.Content) {
				t.Fatal("empty response overwrote the valid cache")
			}
			// Ignore empty caches written by older versions and use shared content.
			if err := f.writeCache(source.URL, FetchedSubscription{Content: []byte(empty), FetchedAt: time.Now()}); err != nil {
				t.Fatal(err)
			}
			source.PreferCache = true
			result, _, err = f.Fetch(context.Background(), source)
			if err != nil || string(result.Content) != string(good.Content) {
				t.Fatal("legacy empty cache prevented shared-cache recovery")
			}
		})
	}
}

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
	const fallbackContent = "ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#fallback"
	const originalUserinfo = "upload=100; download=200; total=1000"
	const fallbackUserinfo = "upload=100; download=300; total=1000"
	for _, tt := range []struct {
		name, userAgent, content string
		fallbackHasUserinfo      bool
	}{
		{"clash", "clash", "proxies: []\nproxy-groups: []\nrules: []\n", true},
		{"default", model.DefaultUserAgent, "proxies: [] # no compatible nodes\nrules: []\n", true},
		{"implicit_default", "", "proxies: []\n", true},
		{"blank_default", " \t", "proxies: null\nrules: []\n", true},
		{"mihomo", "Mihomo/1.19.0", "proxies:\nproxy-groups: []\n", true},
		{"stash", "Stash", "proxies: [\n]\nrules: []\n", true},
		{"sing_box", "sing-box", "proxies: []\n", true},
		{"empty_yaml_alias", model.DefaultUserAgent, "empty: &empty []\nproxies: *empty\n", true},
		{"retain_original_metadata", model.DefaultUserAgent, "proxies: []\n", false},
		{"provider_only_without_proxies_keyword", model.DefaultUserAgent, "proxy-providers:\n  remote: {type: http, url: https://example.com/provider.yaml, path: ./providers/remote.yaml, interval: 3600}\n", false},
		{"provider_only_with_group_use", model.DefaultUserAgent, "proxy-providers:\n  remote: {type: http, url: https://example.com/provider.yaml, path: ./providers/remote.yaml, interval: 3600}\nproxy-groups:\n  - {name: demo, type: select, use: [remote]}\nrules: [MATCH,demo]\n", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cacheRoot := t.TempDir()
			var callUserAgents []string
			f := New(Options{
				CacheDir:       filepath.Join(cacheRoot, "workspace"),
				SharedCacheDir: filepath.Join(cacheRoot, "shared"),
				Resolver:       staticResolver{ips: []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}},
				RequestDoer: func(_ context.Context, _ *url.URL, _ net.IP, source Source) (*http.Response, error) {
					callUserAgents = append(callUserAgents, userAgentOrDefault(source.UserAgent))
					if len(callUserAgents) == 1 {
						return &http.Response{
							StatusCode: http.StatusOK,
							Header: http.Header{
								"Content-Type":          []string{"application/yaml"},
								"Subscription-Userinfo": []string{originalUserinfo},
							},
							Body: io.NopCloser(strings.NewReader(tt.content)),
						}, nil
					}
					header := http.Header{"Content-Type": []string{"text/plain"}}
					if tt.fallbackHasUserinfo {
						header.Set("Subscription-Userinfo", fallbackUserinfo)
					}
					return &http.Response{StatusCode: http.StatusOK, Header: header, Body: io.NopCloser(strings.NewReader(fallbackContent))}, nil
				},
			})
			source := Source{Name: "demo", URL: "https://example.com/subscription", UserAgent: tt.userAgent, Enabled: true}
			fetched, warnings, err := f.Fetch(context.Background(), source)
			if err != nil {
				t.Fatalf("Fetch() error = %v", err)
			}
			if len(callUserAgents) != 2 || callUserAgents[0] != userAgentOrDefault(tt.userAgent) || callUserAgents[1] != "SubConvNext/0.1 OpenWrt" {
				t.Fatalf("request user-agents = %#v, want configured/default followed by generic user-agent", callUserAgents)
			}
			if len(warnings) != 1 || !strings.Contains(warnings[0], "generic user-agent") {
				t.Fatalf("warnings = %#v, want generic user-agent retry warning", warnings)
			}
			wantUserinfo := originalUserinfo
			if tt.fallbackHasUserinfo {
				wantUserinfo = fallbackUserinfo
			}
			if string(fetched.Content) != fallbackContent || fetched.ContentType != "text/plain" || fetched.SubscriptionUserinfo != wantUserinfo || fetched.FetchedAt.IsZero() || fetched.FromCache {
				t.Fatalf("unexpected fallback result: %+v", fetched)
			}
			for _, cacheDir := range []string{f.opts.CacheDir, f.opts.SharedCacheDir} {
				cached, err := f.readCacheFrom(cacheDir, source.URL)
				if err != nil || string(cached.Content) != fallbackContent || cached.ContentType != fetched.ContentType || cached.SubscriptionUserinfo != wantUserinfo || !cached.FetchedAt.Equal(fetched.FetchedAt) {
					t.Fatalf("cache %s did not retain fallback content and metadata: %+v, err = %v", cacheDir, cached, err)
				}
			}
			source.PreferCache = true
			cached, _, err := f.Fetch(context.Background(), source)
			if err != nil || !cached.FromCache || string(cached.Content) != fallbackContent || len(callUserAgents) != 2 {
				t.Fatalf("offline fetch did not reuse fallback cache: %+v, err = %v, requests = %d", cached, err, len(callUserAgents))
			}
		})
	}
}

func TestFetchDoesNotRetryOtherSubscriptionContent(t *testing.T) {
	for _, tt := range []struct {
		name, userAgent, content string
	}{
		{"populated_yaml", model.DefaultUserAgent, "proxies:\n  - {name: demo, type: ss, server: example.com, port: 443, cipher: aes-256-gcm, password: pass}\nrules: []\n"},
		{"uri", model.DefaultUserAgent, "trojan://test@example.com:443#demo"},
		{"base64", model.DefaultUserAgent, "dHJvamFuOi8vdGVzdEBleGFtcGxlLmNvbTo0NDMjZGVtbw=="},
		{"populated_yaml_alias", model.DefaultUserAgent, "nodes: &nodes\n  - {name: demo, type: ss, server: example.com, port: 443, cipher: aes-256-gcm, password: pass}\nproxies: *nodes\n"},
		{"generic_user_agent", "SubConvNext/0.1 OpenWrt", "proxies: []\nrules: []\n"},
		{"custom_user_agent", "CustomClient", "proxies: []\nrules: []\n"},
		{"no_top_level_proxies", model.DefaultUserAgent, "proxy-groups:\n  - name: demo\n    proxies: []\nrules: []\n"},
		{"group_use_without_providers", model.DefaultUserAgent, "proxy-groups:\n  - {name: demo, type: select, use: [remote]}\nrules: [MATCH,demo]\n"},
		{"empty_proxy_providers", model.DefaultUserAgent, "proxy-providers: {}\nproxy-groups: []\nrules: []\n"},
		{"null_proxy_providers", model.DefaultUserAgent, "proxy-providers: null\nproxy-groups: []\nrules: []\n"},
		{"populated_yaml_with_providers", model.DefaultUserAgent, "proxies:\n  - {name: demo, type: ss, server: example.com, port: 443, cipher: aes-256-gcm, password: pass}\nproxy-providers:\n  remote: {type: http, url: https://example.com/provider.yaml}\nrules: []\n"},
		{"malformed_yaml", model.DefaultUserAgent, "proxies: []\nproxies: []\nrules: []\n"},
		{"invalid_proxies_type", model.DefaultUserAgent, "proxies: {}\nrules: []\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			f := New(Options{
				CacheDir: t.TempDir(),
				Resolver: staticResolver{ips: []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}},
				RequestDoer: func(context.Context, *url.URL, net.IP, Source) (*http.Response, error) {
					calls++
					return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(tt.content))}, nil
				},
			})
			fetched, warnings, err := f.Fetch(context.Background(), Source{Name: "demo", URL: "https://example.com/subscription", UserAgent: tt.userAgent, Enabled: true})
			if err != nil || calls != 1 || len(warnings) != 0 || string(fetched.Content) != tt.content {
				t.Fatalf("Fetch() made an unexpected retry: requests = %d, warnings = %#v, err = %v", calls, warnings, err)
			}
		})
	}
}

func TestFetchGenericUserAgentFailureRetainsOriginalResponse(t *testing.T) {
	const originalContent = "proxies: []\nrules: []\n"
	const originalUserinfo = "upload=100; download=200; total=1000"
	for _, failure := range []string{"network", "http_status", "empty_body", "empty_yaml", "provider_only"} {
		t.Run(failure, func(t *testing.T) {
			calls := 0
			f := New(Options{
				CacheDir: t.TempDir(),
				Resolver: staticResolver{ips: []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}},
				RequestDoer: func(context.Context, *url.URL, net.IP, Source) (*http.Response, error) {
					calls++
					if calls == 1 {
						return &http.Response{
							StatusCode: http.StatusOK,
							Header: http.Header{
								"Content-Type":          []string{"application/yaml"},
								"Subscription-Userinfo": []string{originalUserinfo},
							},
							Body: io.NopCloser(strings.NewReader(originalContent)),
						}, nil
					}
					switch failure {
					case "network":
						return nil, errors.New("network down")
					case "http_status":
						return &http.Response{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(strings.NewReader("unavailable"))}, nil
					case "empty_yaml":
						return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("proxies: []\n"))}, nil
					case "provider_only":
						return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("proxy-providers:\n  remote: {type: http, url: https://example.com/nodes}\n"))}, nil
					default:
						return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(" \n"))}, nil
					}
				},
			})
			source := Source{Name: "demo", URL: "https://example.com/subscription", Enabled: true}
			fetched, warnings, err := f.Fetch(context.Background(), source)
			if err != nil || calls != 2 || len(warnings) != 1 || !strings.Contains(warnings[0], "generic user-agent retry failed") {
				t.Fatalf("Fetch() requests = %d, warnings = %#v, err = %v", calls, warnings, err)
			}
			if string(fetched.Content) != originalContent || fetched.ContentType != "application/yaml" || fetched.SubscriptionUserinfo != originalUserinfo {
				t.Fatalf("failed fallback replaced original response: %+v", fetched)
			}
			cached, err := f.readCache(source.URL)
			if err != nil || string(cached.Content) != originalContent || cached.SubscriptionUserinfo != originalUserinfo {
				t.Fatalf("failed fallback did not retain original cached response: %+v, err = %v", cached, err)
			}
		})
	}
}

func TestEmptyClashResponsePreservesUsableSubscriptionCache(t *testing.T) {
	const goodContent = "dHJvamFuOi8vdGVzdEBleGFtcGxlLmNvbTo0NDMjZGVtbw=="
	const emptyTemplate = "proxy-providers:\n  remote: {type: http, url: https://example.com/nodes}\n"
	const cachedUserinfo = "upload=10; download=20; total=1000"
	for _, failure := range []string{"network", "http_status", "empty_body", "empty_yaml", "provider_only", "initial_network"} {
		for _, workspaceHasNodes := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/workspace_nodes_%t", failure, workspaceHasNodes), func(t *testing.T) {
				cacheRoot := t.TempDir()
				calls := 0
				f := New(Options{
					CacheDir:       filepath.Join(cacheRoot, "workspace"),
					SharedCacheDir: filepath.Join(cacheRoot, "shared"),
					Resolver:       staticResolver{ips: []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}},
					RequestDoer: func(context.Context, *url.URL, net.IP, Source) (*http.Response, error) {
						calls++
						if failure == "initial_network" {
							return nil, context.DeadlineExceeded
						}
						if calls == 1 {
							return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(emptyTemplate))}, nil
						}
						switch failure {
						case "network":
							return nil, context.DeadlineExceeded
						case "http_status":
							return &http.Response{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(strings.NewReader("unavailable"))}, nil
						case "empty_body":
							return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(" \n"))}, nil
						case "empty_yaml":
							return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("proxies: []\n"))}, nil
						default:
							return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(emptyTemplate))}, nil
						}
					},
				})
				source := Source{Name: "demo", URL: "https://example.com/subscription", Enabled: true}
				good := FetchedSubscription{Content: []byte(goodContent), ContentType: "text/plain", SubscriptionUserinfo: cachedUserinfo, FetchedAt: time.Now().UTC().Add(-time.Hour)}
				workspace := good
				if !workspaceHasNodes {
					workspace.Content = []byte(emptyTemplate)
				}
				if err := f.writeCache(source.URL, workspace); err != nil {
					t.Fatal(err)
				}
				if err := writeCacheDir(f.opts.SharedCacheDir, source.URL, good); err != nil {
					t.Fatal(err)
				}
				snapshots := map[string][]byte{}
				for _, dir := range []string{f.opts.CacheDir, f.opts.SharedCacheDir} {
					body, meta := cachePaths(dir, source.URL)
					for _, path := range []string{body, meta} {
						data, err := os.ReadFile(path)
						if err != nil {
							t.Fatal(err)
						}
						snapshots[path] = data
					}
				}
				fetched, warnings, err := f.Fetch(context.Background(), source)
				wantCalls := 2
				if failure == "initial_network" {
					wantCalls = 1
				}
				if err != nil || calls != wantCalls || !fetched.FromCache || string(fetched.Content) != goodContent || fetched.ContentType != good.ContentType || fetched.SubscriptionUserinfo != cachedUserinfo || !fetched.FetchedAt.Equal(good.FetchedAt) {
					t.Fatalf("Fetch() lost usable cached nodes or metadata: %+v, err = %v, requests = %d", fetched, err, calls)
				}
				if !strings.Contains(strings.Join(warnings, "\n"), "using cached subscription") {
					t.Fatalf("warnings = %#v, want cache fallback warning", warnings)
				}
				for path, before := range snapshots {
					if !workspaceHasNodes && filepath.Dir(path) == f.opts.CacheDir {
						continue // The old empty workspace copy is repaired from shared nodes.
					}
					after, err := os.ReadFile(path)
					if err != nil || string(after) != string(before) {
						t.Fatalf("usable cache changed at %s: err = %v", path, err)
					}
				}
				source.PreferCache = true
				offline, _, err := f.Fetch(context.Background(), source)
				if err != nil || !offline.FromCache || string(offline.Content) != goodContent || calls != wantCalls {
					t.Fatalf("offline node view lost restored nodes: %+v, err = %v, requests = %d", offline, err, calls)
				}
			})
		}
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
