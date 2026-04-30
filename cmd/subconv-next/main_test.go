package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"subconv-next/internal/model"
)

func TestRunVersion(t *testing.T) {
	oldVersion := version
	version = "0.1.0-test"
	t.Cleanup(func() {
		version = oldVersion
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run(version) exit code = %d, want 0", code)
	}

	if got := stdout.String(); got != "0.1.0-test\n" {
		t.Fatalf("stdout = %q, want %q", got, "0.1.0-test\n")
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"nope"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("run(nope) exit code = %d, want 2", code)
	}

	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}

	if got := stderr.String(); !strings.Contains(got, "unknown command: nope") {
		t.Fatalf("stderr = %q, want unknown command message", got)
	}
}

func TestRunRootServeFlagsUseServeCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"--bad-serve-flag"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("run(root serve flags) exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Fatalf("stderr = %q, want serve command handling", stderr.String())
	}
}

func TestApplyServeOverrides(t *testing.T) {
	cfg := model.DefaultConfig()
	dir := t.TempDir()

	err := applyServeOverrides(&cfg, serveOverrides{
		host:          "0.0.0.0",
		port:          19876,
		dataDir:       dir,
		publicBaseURL: "https://subconv.example.com/",
		logLevel:      "debug",
	})
	if err != nil {
		t.Fatalf("applyServeOverrides() error = %v", err)
	}

	if cfg.Service.ListenAddr != "0.0.0.0" || cfg.Service.ListenPort != 19876 {
		t.Fatalf("listen = %s:%d, want override", cfg.Service.ListenAddr, cfg.Service.ListenPort)
	}
	if cfg.Service.StatePath != filepath.Join(dir, "state.json") ||
		cfg.Service.CacheDir != filepath.Join(dir, "cache") ||
		cfg.Service.OutputPath != filepath.Join(dir, "mihomo.yaml") {
		t.Fatalf("data paths = %#v, want under %s", cfg.Service, dir)
	}
	if cfg.Service.PublicBaseURL != "https://subconv.example.com" {
		t.Fatalf("PublicBaseURL = %q, want trimmed override", cfg.Service.PublicBaseURL)
	}
	if cfg.Service.LogLevel != "debug" || cfg.Render.LogLevel != "debug" {
		t.Fatalf("log levels = %q/%q, want debug", cfg.Service.LogLevel, cfg.Render.LogLevel)
	}
}

func TestLoadServeConfigMissingUsesDefaults(t *testing.T) {
	cfg, err := loadServeConfig(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("loadServeConfig(missing) error = %v", err)
	}
	if cfg.Service.ListenPort != model.DefaultListenPort || cfg.Service.OutputPath != model.DefaultOutputPath {
		t.Fatalf("loadServeConfig(missing) = %+v, want defaults", cfg.Service)
	}
}

func TestServeOverridesFromEnvAndFlags(t *testing.T) {
	t.Setenv("SUBCONV_HOST", "0.0.0.0")
	t.Setenv("SUBCONV_PORT", "19876")
	t.Setenv("SUBCONV_DATA_DIR", "/tmp/subconv-data")
	t.Setenv("SUBCONV_PUBLIC_BASE_URL", "https://subconv.example.com")
	t.Setenv("SUBCONV_LOG_LEVEL", "debug")

	got, err := serveOverridesFromEnvAndFlags(map[string]bool{
		"port":      true,
		"log-level": true,
	}, serveOverrides{
		port:     9876,
		logLevel: "info",
	})
	if err != nil {
		t.Fatalf("serveOverridesFromEnvAndFlags() error = %v", err)
	}

	if got.host != "0.0.0.0" || got.port != 9876 || got.dataDir != "/tmp/subconv-data" ||
		got.publicBaseURL != "https://subconv.example.com" || got.logLevel != "info" {
		t.Fatalf("serveOverridesFromEnvAndFlags() = %#v", got)
	}
}

func TestServeOverridesRejectInvalidEnvPort(t *testing.T) {
	t.Setenv("SUBCONV_PORT", "bad")
	if _, err := serveOverridesFromEnvAndFlags(map[string]bool{}, serveOverrides{}); err == nil || !strings.Contains(err.Error(), "SUBCONV_PORT") {
		t.Fatalf("serveOverridesFromEnvAndFlags() error = %v, want SUBCONV_PORT failure", err)
	}
}

func TestRunParseJSON(t *testing.T) {
	inputPath := filepath.Join("..", "..", "testdata", "nodes", "vless-reality.txt")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"parse", "--input", inputPath, "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run(parse) exit code = %d, want 0; stderr=%q", code, stderr.String())
	}

	if got := stdout.String(); !strings.Contains(got, "\"type\": \"vless\"") {
		t.Fatalf("stdout = %q, want vless JSON output", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunGenerate(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	outputPath := filepath.Join(dir, "mihomo.yaml")

	configBody := `{
  "service": {
    "template": "lite"
  },
  "inline": [
    {
      "name": "manual",
      "content": "ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo0NDM=#ss-node"
    }
  ]
}`
	if err := os.WriteFile(configPath, []byte(configBody), 0o644); err != nil {
		t.Fatalf("WriteFile(config) error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"generate", "--config", configPath, "--out", outputPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run(generate) exit code = %d, want 0; stderr=%q", code, stderr.String())
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(output) error = %v", err)
	}
	if got := string(data); !strings.Contains(got, "type: ss") {
		t.Fatalf("output = %q, want ss proxy", got)
	}
	if strings.TrimSpace(stdout.String()) != outputPath {
		t.Fatalf("stdout = %q, want %q", stdout.String(), outputPath)
	}
}
