package api

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"subconv-next/internal/model"
)

type siteLogoResolverStub struct {
	ips []net.IPAddr
	err error
}

func (stub siteLogoResolverStub) LookupIPAddr(context.Context, string) ([]net.IPAddr, error) {
	return stub.ips, stub.err
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestSiteLogoTargetRejectsPrivateAndCredentialedURLs(t *testing.T) {
	private := siteLogoResolverStub{ips: []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}}}
	public := siteLogoResolverStub{ips: []net.IPAddr{{IP: net.ParseIP("203.0.113.10")}}}

	if err := validateSiteLogoTarget(context.Background(), private, mustParseURL(t, "https://example.com/icon.png")); err == nil {
		t.Fatal("validateSiteLogoTarget() accepted a private redirect target")
	}
	if err := validateSiteLogoTarget(context.Background(), public, mustParseURL(t, "https://user:pass@example.com/icon.png")); err == nil {
		t.Fatal("validateSiteLogoTarget() accepted URL credentials")
	}
	if err := validateSiteLogoTarget(context.Background(), public, mustParseURL(t, "https://example.com/icon.png")); err != nil {
		t.Fatalf("validateSiteLogoTarget() rejected public target: %v", err)
	}
}

func TestParseSiteLogoURLRejectsCredentials(t *testing.T) {
	if _, err := parseSiteLogoURL("https://user:pass@example.com/"); err == nil {
		t.Fatal("parseSiteLogoURL() accepted URL credentials")
	}
}

func TestResolveSiteLogoDialTargetUsesValidatedIP(t *testing.T) {
	resolver := siteLogoResolverStub{ips: []net.IPAddr{{IP: net.ParseIP("203.0.113.10")}}}
	got, err := resolveSiteLogoDialTarget(context.Background(), resolver, "example.com:443")
	if err != nil {
		t.Fatalf("resolveSiteLogoDialTarget() error = %v", err)
	}
	if got != "203.0.113.10:443" {
		t.Fatalf("dial target = %q, want public resolved address", got)
	}
}

func TestFetchLogoRejectsNonImageAndOversizedBodies(t *testing.T) {
	target := mustParseURL(t, "https://example.com/icon")
	response := func(contentType string, body []byte) *http.Client {
		return &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{contentType}},
				Body:       io.NopCloser(bytes.NewReader(body)),
			}, nil
		})}
	}

	if _, ok := fetchLogoAsDataURL(response("text/html", []byte("<html></html>")), target); ok {
		t.Fatal("fetchLogoAsDataURL() accepted non-image content")
	}
	if _, ok := fetchLogoAsDataURL(response("image/png", bytes.Repeat([]byte{'x'}, 256*1024+1)), target); ok {
		t.Fatal("fetchLogoAsDataURL() accepted oversized image")
	}
	dataURL, ok := fetchLogoAsDataURL(response("image/png", []byte("png")), target)
	if !ok || !strings.HasPrefix(dataURL, "data:image/png;base64,") {
		t.Fatalf("fetchLogoAsDataURL() = %q, %v", dataURL, ok)
	}
}

func TestSiteLogoCacheIsBounded(t *testing.T) {
	server := NewServer("test", model.DefaultConfig())
	for index := 0; index < maxSiteLogoCacheEntries+20; index++ {
		server.storeSiteLogoCache(string(rune(index+1)), siteLogoResponse{OK: true})
	}
	server.siteLogoMu.RLock()
	count := len(server.siteLogoCache)
	server.siteLogoMu.RUnlock()
	if count > maxSiteLogoCacheEntries {
		t.Fatalf("site logo cache size = %d, want <= %d", count, maxSiteLogoCacheEntries)
	}

	server.siteLogoMu.Lock()
	server.siteLogoCache["expired"] = siteLogoCacheEntry{ExpiresAt: time.Now().Add(-time.Minute)}
	server.siteLogoMu.Unlock()
	server.storeSiteLogoCache("replacement", siteLogoResponse{OK: true})
	if _, ok := server.lookupSiteLogoCache("expired"); ok {
		t.Fatal("expired site logo cache entry was retained")
	}
}

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
