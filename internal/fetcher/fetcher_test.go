package fetcher

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchSuccessAndCacheFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#cached"))
	}))
	defer server.Close()

	f := New(Options{
		CacheDir:          t.TempDir(),
		Timeout:           5 * time.Second,
		MaxBodyBytes:      1024,
		MaxRedirects:      3,
		AllowPrivateHosts: true,
	})

	source := Source{
		Name:      "demo",
		URL:       server.URL,
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

	server.Close()

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

func TestFetchBlocksPrivateHostsByDefault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://127.0.0.1/private", http.StatusFound)
	}))
	defer server.Close()

	f := New(Options{
		CacheDir:     t.TempDir(),
		Timeout:      5 * time.Second,
		MaxBodyBytes: 1024,
		MaxRedirects: 3,
	})

	_, _, err := f.Fetch(context.Background(), Source{
		Name:      "blocked",
		URL:       server.URL,
		UserAgent: "SubConvNext/0.1",
		Enabled:   true,
	})
	if err == nil {
		t.Fatalf("Fetch() error = nil, want blocked host error")
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
