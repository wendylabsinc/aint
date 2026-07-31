package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"aint/internal/check"
	"aint/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.Default()
	if cfg.FailOn != check.SeverityError {
		t.Errorf("expected default fail_on to be error, got %q", cfg.FailOn)
	}
}

func TestLoadMissingFileReturnsDefault(t *testing.T) {
	cfg, err := config.Load(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.FailOn != check.SeverityError {
		t.Errorf("expected default fail_on, got %q", cfg.FailOn)
	}
}

func TestLoadParsesYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".aint.yaml")
	yaml := `
fail_on: warning
ignore:
  - vendor/**
  - "*.pb.go"
checks:
  go-ignored-error: error
  node-console-log: off
docs_base_url: https://example.com/docs
`
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.FailOn != check.SeverityWarning {
		t.Errorf("expected fail_on warning, got %q", cfg.FailOn)
	}
	if cfg.DocsBaseURL != "https://example.com/docs" {
		t.Errorf("unexpected docs_base_url: %q", cfg.DocsBaseURL)
	}
	if cfg.Checks["go-ignored-error"] != "error" {
		t.Errorf("expected go-ignored-error override to be error, got %q", cfg.Checks["go-ignored-error"])
	}
	if cfg.Checks["node-console-log"] != "off" {
		t.Errorf("expected node-console-log override to be off, got %q", cfg.Checks["node-console-log"])
	}
}

func TestIsIgnored(t *testing.T) {
	cfg := config.Config{Ignore: []string{"vendor/**", "*.pb.go"}}
	cases := map[string]bool{
		"vendor/foo/bar.go":  true,
		"vendor":             true,
		"pkg/api.pb.go":      true,
		"pkg/main.go":        false,
	}
	for path, want := range cases {
		if got := cfg.IsIgnored(path); got != want {
			t.Errorf("IsIgnored(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestSeverityForOverridesAndDisables(t *testing.T) {
	cfg := config.Config{Checks: map[string]string{
		"go-ignored-error": "error",
		"node-eval":        "off",
	}}
	c := check.Check{ID: "go-ignored-error", Severity: check.SeverityWarning}
	sev, enabled := cfg.SeverityFor(c)
	if !enabled || sev != check.SeverityError {
		t.Errorf("expected override to error, got sev=%q enabled=%v", sev, enabled)
	}

	off := check.Check{ID: "node-eval", Severity: check.SeverityError}
	_, enabled = cfg.SeverityFor(off)
	if enabled {
		t.Error("expected node-eval to be disabled")
	}

	untouched := check.Check{ID: "some-other-check", Severity: check.SeverityInfo}
	sev, enabled = cfg.SeverityFor(untouched)
	if !enabled || sev != check.SeverityInfo {
		t.Errorf("expected untouched check to keep its own severity, got sev=%q enabled=%v", sev, enabled)
	}
}

func TestResolveDocsURL(t *testing.T) {
	withBase := config.Config{DocsBaseURL: "https://example.com/docs/"}
	if got := config.ResolveDocsURL(withBase, "go-ignored-error.md"); got != "https://example.com/docs/go-ignored-error.md" {
		t.Errorf("unexpected docs URL: %q", got)
	}

	noBase := config.Config{}
	if got := config.ResolveDocsURL(noBase, "go-ignored-error.md"); got != "docs/checks/go-ignored-error.md" {
		t.Errorf("unexpected relative docs URL: %q", got)
	}
}
