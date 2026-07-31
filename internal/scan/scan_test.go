// internal/scan/scan_test.go
package scan_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"aint/internal/check"
	"aint/internal/config"
	"aint/internal/scan"
)

func TestCommandTarget(t *testing.T) {
	target := scan.CommandTarget("gcloud projects add-iam-policy-binding x --role=roles/owner")
	if target.Source != "<command>" {
		t.Errorf("expected source <command>, got %q", target.Source)
	}
	if target.Lang != "shell" {
		t.Errorf("expected lang shell, got %q", target.Lang)
	}
}

func TestWalkRespectsIgnoreAndClassifiesLang(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "main.go"), "package main")
	mustWrite(t, filepath.Join(dir, "vendor", "dep.go"), "package dep")

	cfg := config.Config{Ignore: []string{"vendor/**"}}
	targets, err := scan.Walk([]string{dir}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var gotMain, gotVendor bool
	for _, tg := range targets {
		if filepath.Base(tg.Source) == "main.go" {
			gotMain = true
			if tg.Lang != "go" {
				t.Errorf("expected main.go lang go, got %q", tg.Lang)
			}
		}
		if filepath.Base(tg.Source) == "dep.go" {
			gotVendor = true
		}
	}
	if !gotMain {
		t.Error("expected main.go to be walked")
	}
	if gotVendor {
		t.Error("expected vendor/dep.go to be ignored")
	}
}

func TestWalkIgnoresSingleFileRootByBasenamePattern(t *testing.T) {
	dir := t.TempDir()
	ignoredFile := filepath.Join(dir, "dep.pb.go")
	keptFile := filepath.Join(dir, "main.go")
	mustWrite(t, ignoredFile, "package dep")
	mustWrite(t, keptFile, "package main")

	cfg := config.Config{Ignore: []string{"*.pb.go"}}

	// Passing a single FILE (not a directory) as the walk root, as a later
	// CLI command like `aint check dep.pb.go` would. filepath.Walk invokes
	// the callback once with path == root, so relative-to-root logic must
	// not defeat basename matching in this case.
	targets, err := scan.Walk([]string{ignoredFile}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 0 {
		t.Fatalf("expected ignored file root to produce no targets, got %+v", targets)
	}

	keptTargets, err := scan.Walk([]string{keptFile}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keptTargets) != 1 {
		t.Fatalf("expected non-ignored file root to produce 1 target, got %d", len(keptTargets))
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestRunAppliesLangFilterAndSeverityOverride(t *testing.T) {
	goCheck := check.Check{
		ID:       "sample-go-check",
		Severity: check.SeverityWarning,
		Langs:    []string{"go"},
		Pattern:  regexp.MustCompile(`TODO`),
		Message:  "found a TODO",
		DocsPath: "sample-go-check.md",
	}
	targets := []scan.Target{
		{Source: "main.go", Lang: "go", Content: []byte("// TODO fix")},
		{Source: "script.py", Lang: "python", Content: []byte("# TODO fix")},
	}

	cfg := config.Config{FailOn: check.SeverityError}
	findings := scan.Run(targets, []check.Check{goCheck}, cfg)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding (go file only), got %d", len(findings))
	}
	if findings[0].Source != "main.go" {
		t.Errorf("expected finding on main.go, got %q", findings[0].Source)
	}
	if findings[0].Severity != check.SeverityWarning {
		t.Errorf("expected default severity warning, got %q", findings[0].Severity)
	}

	cfg.Checks = map[string]string{"sample-go-check": "error"}
	overridden := scan.Run(targets, []check.Check{goCheck}, cfg)
	if len(overridden) != 1 || overridden[0].Severity != check.SeverityError {
		t.Fatalf("expected overridden severity error, got %+v", overridden)
	}

	cfg.Checks = map[string]string{"sample-go-check": "off"}
	disabled := scan.Run(targets, []check.Check{goCheck}, cfg)
	if len(disabled) != 0 {
		t.Fatalf("expected 0 findings when check disabled, got %d", len(disabled))
	}
}
