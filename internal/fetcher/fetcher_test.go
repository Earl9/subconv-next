package fetcher

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
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
		{ip: "8.8.8.8", want: false},
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

type staticResolver struct {
	ips []net.IPAddr
}

func (r staticResolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	return append([]net.IPAddr(nil), r.ips...), nil
}
