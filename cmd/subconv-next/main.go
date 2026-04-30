package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"subconv-next/internal/api"
	"subconv-next/internal/config"
	"subconv-next/internal/model"
	"subconv-next/internal/parser"
	"subconv-next/internal/pipeline"
)

var version = "dev"

const defaultConfigPath = "config/config.json"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	switch args[0] {
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	case "version":
		_, _ = fmt.Fprintln(stdout, version)
		return 0
	case "serve":
		return runServe(args[1:], stderr)
	case "parse":
		return runParse(args[1:], stdout, stderr)
	case "generate":
		return runGenerate(args[1:], stdout, stderr)
	default:
		if strings.HasPrefix(args[0], "-") {
			return runServe(args, stderr)
		}
		_, _ = fmt.Fprintf(stderr, "unknown command: %s\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage: subconv-next <command>")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Commands:")
	_, _ = fmt.Fprintln(w, "  serve       Run the local HTTP daemon")
	_, _ = fmt.Fprintln(w, "  generate    Generate Mihomo YAML from config")
	_, _ = fmt.Fprintln(w, "  parse       Parse subscription content into NodeIR")
	_, _ = fmt.Fprintln(w, "  version     Print the build version")
}

func runServe(args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(stderr)

	configPath := fs.String("config", defaultConfigPath, "Path to a JSON or UCI config file")
	host := fs.String("host", "", "Override service listen host")
	port := fs.Int("port", 0, "Override service listen port")
	dataDir := fs.String("data-dir", "", "Override runtime data directory")
	publicBaseURL := fs.String("public-base-url", "", "Override public base URL used in generated subscription links")
	logLevel := fs.String("log-level", "", "Override service log level")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		_, _ = fmt.Fprintf(stderr, "serve: unexpected arguments: %s\n", strings.Join(fs.Args(), " "))
		return 2
	}

	cfg, err := loadServeConfig(*configPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "serve: %v\n", err)
		return 1
	}
	explicitFlags := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		explicitFlags[f.Name] = true
	})
	overrides, err := serveOverridesFromEnvAndFlags(explicitFlags, serveOverrides{
		host:          *host,
		port:          *port,
		dataDir:       *dataDir,
		publicBaseURL: *publicBaseURL,
		logLevel:      *logLevel,
	})
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "serve: %v\n", err)
		return 2
	}
	if err := applyServeOverrides(&cfg, overrides); err != nil {
		_, _ = fmt.Fprintf(stderr, "serve: %v\n", err)
		return 2
	}

	server := api.NewServer(version, cfg)
	server.SetConfigPath(*configPath)
	httpServer := &http.Server{
		Addr:              api.ListenAddress(cfg),
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	server.StartScheduler(ctx.Done())

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return 0
		}
		_, _ = fmt.Fprintf(stderr, "serve: %v\n", err)
		return 1
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			_, _ = fmt.Fprintf(stderr, "serve shutdown: %v\n", err)
			return 1
		}

		err := <-errCh
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			_, _ = fmt.Fprintf(stderr, "serve: %v\n", err)
			return 1
		}
		return 0
	}
}

type serveOverrides struct {
	host          string
	port          int
	dataDir       string
	publicBaseURL string
	logLevel      string
}

func loadServeConfig(path string) (model.Config, error) {
	cfg, err := config.Load(path)
	if err == nil {
		return cfg, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return model.DefaultConfig(), nil
	}
	return model.Config{}, err
}

func serveOverridesFromEnvAndFlags(explicitFlags map[string]bool, flags serveOverrides) (serveOverrides, error) {
	envPort := 0
	if !explicitFlags["port"] {
		parsed, err := parseOptionalEnvPort("SUBCONV_PORT")
		if err != nil {
			return serveOverrides{}, err
		}
		envPort = parsed
	}
	return serveOverrides{
		host:          stringOverride("SUBCONV_HOST", flags.host, explicitFlags["host"]),
		port:          intOverride(envPort, flags.port, explicitFlags["port"]),
		dataDir:       stringOverride("SUBCONV_DATA_DIR", flags.dataDir, explicitFlags["data-dir"]),
		publicBaseURL: stringOverride("SUBCONV_PUBLIC_BASE_URL", flags.publicBaseURL, explicitFlags["public-base-url"]),
		logLevel:      stringOverride("SUBCONV_LOG_LEVEL", flags.logLevel, explicitFlags["log-level"]),
	}, nil
}

func parseOptionalEnvPort(key string) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}
	return parsed, nil
}

func stringOverride(envKey, flagValue string, flagExplicit bool) string {
	if flagExplicit {
		return flagValue
	}
	return os.Getenv(envKey)
}

func intOverride(envValue, flagValue int, flagExplicit bool) int {
	if flagExplicit {
		return flagValue
	}
	return envValue
}

func applyServeOverrides(cfg *model.Config, overrides serveOverrides) error {
	if cfg == nil {
		return nil
	}
	if value := strings.TrimSpace(overrides.host); value != "" {
		cfg.Service.ListenAddr = value
	}
	if overrides.port != 0 {
		if overrides.port < 1 || overrides.port > 65535 {
			return fmt.Errorf("--port must be between 1 and 65535")
		}
		cfg.Service.ListenPort = overrides.port
	}
	if value := strings.TrimSpace(overrides.logLevel); value != "" {
		cfg.Service.LogLevel = value
		cfg.Render.LogLevel = value
	}
	if value := strings.TrimRight(strings.TrimSpace(overrides.publicBaseURL), "/"); value != "" {
		cfg.Service.PublicBaseURL = value
	}
	if value := strings.TrimSpace(overrides.dataDir); value != "" {
		if !filepath.IsAbs(value) {
			return fmt.Errorf("--data-dir must be an absolute path")
		}
		cfg.Service.StatePath = filepath.Join(value, "state.json")
		cfg.Service.CacheDir = filepath.Join(value, "cache")
		cfg.Service.OutputPath = filepath.Join(value, "mihomo.yaml")
	}
	return nil
}

func runParse(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("parse", flag.ContinueOnError)
	fs.SetOutput(stderr)

	inputPath := fs.String("input", "", "Path to an input file")
	jsonOutput := fs.Bool("json", false, "Print the parse result as JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *inputPath == "" {
		_, _ = fmt.Fprintln(stderr, "parse: --input is required")
		return 2
	}
	if fs.NArg() != 0 {
		_, _ = fmt.Fprintf(stderr, "parse: unexpected arguments: %s\n", strings.Join(fs.Args(), " "))
		return 2
	}

	content, err := os.ReadFile(*inputPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "parse: read input %q: %v\n", *inputPath, err)
		return 1
	}

	result := parser.ParseContent(content, model.SourceInfo{
		Name: filepath.Base(*inputPath),
		Kind: "file",
	})

	if *jsonOutput {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			_, _ = fmt.Fprintf(stderr, "parse: write JSON: %v\n", err)
			return 1
		}
	} else {
		for _, node := range result.Nodes {
			_, _ = fmt.Fprintf(stdout, "%s\t%s\t%s:%d\n", node.Name, node.Type, node.Server, node.Port)
		}
	}

	for _, warning := range result.Warnings {
		_, _ = fmt.Fprintf(stderr, "warning: %s\n", warning)
	}
	for _, parseErr := range result.Errors {
		if parseErr.Line > 0 {
			_, _ = fmt.Fprintf(stderr, "error line %d [%s]: %s\n", parseErr.Line, parseErr.Kind, parseErr.Message)
			continue
		}
		_, _ = fmt.Fprintf(stderr, "error [%s]: %s\n", parseErr.Kind, parseErr.Message)
	}

	if len(result.Nodes) == 0 {
		return 1
	}
	return 0
}

func runGenerate(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(stderr)

	configPath := fs.String("config", defaultConfigPath, "Path to a JSON config file")
	outPath := fs.String("out", "", "Path to the output Mihomo YAML file")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		_, _ = fmt.Fprintf(stderr, "generate: unexpected arguments: %s\n", strings.Join(fs.Args(), " "))
		return 2
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "generate: %v\n", err)
		return 1
	}

	result, err := pipeline.RenderConfig(cfg)
	for _, warning := range result.Warnings {
		_, _ = fmt.Fprintf(stderr, "warning: %s\n", warning)
	}
	for _, parseErr := range result.Errors {
		if parseErr.Line > 0 {
			_, _ = fmt.Fprintf(stderr, "error line %d [%s]: %s\n", parseErr.Line, parseErr.Kind, parseErr.Message)
			continue
		}
		_, _ = fmt.Fprintf(stderr, "error [%s]: %s\n", parseErr.Kind, parseErr.Message)
	}
	if err != nil {
		if errors.Is(err, pipeline.ErrNoNodes) {
			_, _ = fmt.Fprintln(stderr, "generate: no nodes available for rendering")
		} else {
			_, _ = fmt.Fprintf(stderr, "generate: %v\n", err)
		}
		return 1
	}

	target := *outPath
	if strings.TrimSpace(target) == "" {
		target = result.OutputPath
	}

	if err := pipeline.WriteRendered(target, result.YAML); err != nil {
		_, _ = fmt.Fprintf(stderr, "generate: write output %q: %v\n", target, err)
		return 1
	}

	_, _ = fmt.Fprintln(stdout, target)
	return 0
}
