package fetcher

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestContentValidationPreservesGoodCachesAndSkipsLegacyInvalidCopies(t *testing.T) {
	root := t.TempDir()
	calls := 0
	f := New(Options{
		CacheDir: filepath.Join(root, "workspace"), SharedCacheDir: filepath.Join(root, "shared"),
		Resolver: staticResolver{ips: []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}},
		RequestDoer: func(context.Context, *url.URL, net.IP, Source) (*http.Response, error) {
			calls++
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("invalid rules"))}, nil
		},
	})
	invalid := errors.New("invalid rule payload")
	source := Source{Name: "custom rule", URL: "https://example.com/rules", Enabled: true, ValidateContent: func(body []byte) error {
		if string(body) != "example.com" {
			return invalid
		}
		return nil
	}}
	if _, _, err := f.Fetch(context.Background(), source); !errors.Is(err, invalid) {
		t.Fatalf("error = %v, want content validation error", err)
	}
	for _, dir := range []string{f.opts.CacheDir, f.opts.SharedCacheDir} {
		if _, err := f.readCacheFrom(dir, source.URL); err == nil {
			t.Fatal("invalid response entered cache")
		}
	}
	good := FetchedSubscription{Content: []byte("example.com"), FetchedAt: time.Now()}
	if err := f.writeCache(source.URL, good); err != nil {
		t.Fatal(err)
	}
	f.writeSharedCache(source.URL, good)
	got, warnings, err := f.Fetch(context.Background(), source)
	if err != nil || !got.FromCache || string(got.Content) != string(good.Content) || len(warnings) == 0 {
		t.Fatalf("invalid update lost last good snapshot: result=%+v warnings=%v error=%v", got, warnings, err)
	}
	for _, dir := range []string{f.opts.CacheDir, f.opts.SharedCacheDir} {
		cached, err := f.readCacheFrom(dir, source.URL)
		if err != nil || string(cached.Content) != string(good.Content) {
			t.Fatal("invalid response overwrote valid cache")
		}
	}
	// Older versions may already have committed an invalid workspace copy.
	for _, offline := range []bool{false, true} {
		if err := f.writeCache(source.URL, FetchedSubscription{Content: []byte("invalid rules"), FetchedAt: time.Now()}); err != nil {
			t.Fatal(err)
		}
		source.PreferCache, source.CacheTTL = offline, time.Hour
		before := calls
		got, _, err := f.Fetch(context.Background(), source)
		if err != nil || !got.FromCache || string(got.Content) != string(good.Content) {
			t.Fatalf("legacy bad cache blocked shared recovery (offline=%v): %v", offline, err)
		}
		if offline && calls != before {
			t.Fatal("offline recovery contacted the upstream")
		}
	}
}
