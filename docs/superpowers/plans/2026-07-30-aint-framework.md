# aint Framework Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the `aint` static-analysis CLI framework — check engine, config, reporting, a seed set of checks across secrets/shell-IaC/Go/Swift/Python/Node, and a `aint install` command that wires it into Claude Code's hooks.

**Architecture:** Single static Go binary. Checks are regex matchers registered into a global registry by category package (`internal/checks/{secrets,shell,golang,swift,python,nodejs}`). A scanner (`internal/scan`) walks files or wraps a raw shell command string, classifies language, and runs registered checks through `.aint.yaml` config (`internal/config`) to produce findings, rendered by `internal/report`. Two hook subcommands (`aint hook pre-bash` / `aint hook post-edit`) adapt Claude Code's hook JSON stdin payloads onto the same engine; `aint install` merges the hook wiring into `.claude/settings.json`.

**Tech Stack:** Go 1.22+, `gopkg.in/yaml.v3` for config parsing (only external dependency), stdlib `regexp`/`flag`/`encoding/json` for everything else.

## Global Constraints

- Module name: `aint` (import paths are `aint/internal/...`).
- No AST/tree-sitter parsing anywhere — all checks are line-based regex matches (per spec).
- No auto-fix — findings are reported, never applied.
- No `aint uninstall` command in this milestone.
- `.aint.yaml` is optional; built-in defaults apply when absent.
- `aint install` must be idempotent — running it twice must not duplicate hook entries, and must not disturb unrelated existing `settings.json` content.
- Exit codes: `aint check` returns `0` clean / `1` if findings at/above `fail_on` exist. `aint hook pre-bash` / `aint hook post-edit` return `0` clean / `2` if findings at/above `fail_on` exist (so Claude Code blocks or surfaces feedback).
- Spec reference: `docs/superpowers/specs/2026-07-30-aint-design.md`.

---

### Task 1: Project scaffold

**Files:**
- Create: `go.mod`
- Create: `cmd/aint/main.go`

**Interfaces:**
- Produces: a buildable `aint` binary with no subcommands yet, printing usage and exiting 2 on any invocation.

- [ ] **Step 1: Initialize the Go module**

Run: `go mod init aint`
Expected: creates `go.mod` with `module aint` and a `go 1.22`-or-later directive.

- [ ] **Step 2: Write the scaffold main.go**

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	os.Exit(dispatch(os.Args))
}

func dispatch(args []string) int {
	if len(args) < 2 {
		printUsage()
		return 2
	}
	switch args[1] {
	default:
		printUsage()
		return 2
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `usage: aint <command> [args]

commands:
  check [paths...]            scan files/dirs for issues
  list                        list all registered checks
  install [--global]          wire aint into Claude Code hooks
  hook <pre-bash|post-edit>   internal: used by installed hooks`)
}
```

- [ ] **Step 3: Build and smoke-test**

Run: `go build ./... && go run ./cmd/aint`
Expected: usage text printed to stderr, exit code 2 (check with `echo $?`).

- [ ] **Step 4: Commit**

```bash
git add go.mod cmd/aint/main.go
git commit -m "scaffold aint Go module and CLI entrypoint"
```

---

### Task 2: Core check/finding types and registry

**Files:**
- Create: `internal/check/types.go`
- Create: `internal/check/registry.go`
- Test: `internal/check/check_test.go`

**Interfaces:**
- Produces:
  - `type Severity string` with constants `SeverityInfo`, `SeverityWarning`, `SeverityError`, and method `(s Severity) AtLeast(min Severity) bool`.
  - `type Check struct { ID, Title string; Severity Severity; Langs []string; Pattern *regexp.Regexp; Message string; DocsPath string }`.
  - `func (c Check) Run(source string, content []byte, docsURL string) []Finding`.
  - `type Finding struct { CheckID string; Severity Severity; Source string; Line, Column int; Message, DocsURL string }`.
  - `func Register(c Check)` and `func All() []Check` (package-level registry).

- [ ] **Step 1: Write the failing tests**

```go
// internal/check/check_test.go
package check_test

import (
	"regexp"
	"testing"

	"aint/internal/check"
)

func TestCheckRunFindsMatch(t *testing.T) {
	c := check.Check{
		ID:       "test-check",
		Severity: check.SeverityError,
		Pattern:  regexp.MustCompile(`TODO`),
		Message:  "found a TODO",
	}
	findings := c.Run("file.go", []byte("line one\nTODO: fix this\nline three"), "docs/checks/test-check.md")

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.Line != 2 {
		t.Errorf("expected line 2, got %d", f.Line)
	}
	if f.Column != 1 {
		t.Errorf("expected column 1, got %d", f.Column)
	}
	if f.CheckID != "test-check" {
		t.Errorf("expected check ID to be passed through, got %q", f.CheckID)
	}
	if f.DocsURL != "docs/checks/test-check.md" {
		t.Errorf("expected docs URL to be passed through, got %q", f.DocsURL)
	}
}

func TestCheckRunNoMatch(t *testing.T) {
	c := check.Check{ID: "test-check", Pattern: regexp.MustCompile(`TODO`)}
	findings := c.Run("file.go", []byte("nothing to see here"), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestSeverityAtLeast(t *testing.T) {
	if !check.SeverityError.AtLeast(check.SeverityWarning) {
		t.Error("error should be at least warning")
	}
	if check.SeverityInfo.AtLeast(check.SeverityError) {
		t.Error("info should not be at least error")
	}
	if !check.SeverityWarning.AtLeast(check.SeverityWarning) {
		t.Error("a severity should be at least itself")
	}
}

func TestRegisterAndAll(t *testing.T) {
	before := len(check.All())
	check.Register(check.Check{ID: "temp-check-for-test"})
	after := check.All()
	if len(after) != before+1 {
		t.Fatalf("expected registry to grow by 1, got %d -> %d", before, len(after))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/check/... -v`
Expected: FAIL — package `aint/internal/check` does not exist yet.

- [ ] **Step 3: Implement the types**

```go
// internal/check/types.go
package check

import (
	"regexp"
	"strings"
)

type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

var severityRank = map[Severity]int{
	SeverityInfo:    0,
	SeverityWarning: 1,
	SeverityError:   2,
}

// AtLeast reports whether s is at least as severe as min.
func (s Severity) AtLeast(min Severity) bool {
	return severityRank[s] >= severityRank[min]
}

// Check is a single regex-based rule. Langs is a list of language tags
// (e.g. "go", "swift", "shell") the check applies to; an empty/nil Langs
// means the check applies to every language, including the synthetic
// "shell" pseudo-file used for raw shell command strings.
type Check struct {
	ID       string
	Title    string
	Severity Severity
	Langs    []string
	Pattern  *regexp.Regexp
	Message  string
	DocsPath string
}

// Finding is a single reported match.
type Finding struct {
	CheckID  string
	Severity Severity
	Source   string
	Line     int
	Column   int
	Message  string
	DocsURL  string
}

// Run scans content line by line for the check's pattern, returning one
// Finding per matching line. docsURL is passed in by the caller (resolved
// from config) rather than read from c.DocsPath directly.
func (c Check) Run(source string, content []byte, docsURL string) []Finding {
	var findings []Finding
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		loc := c.Pattern.FindStringIndex(line)
		if loc == nil {
			continue
		}
		findings = append(findings, Finding{
			CheckID:  c.ID,
			Severity: c.Severity,
			Source:   source,
			Line:     i + 1,
			Column:   loc[0] + 1,
			Message:  c.Message,
			DocsURL:  docsURL,
		})
	}
	return findings
}
```

```go
// internal/check/registry.go
package check

var registry []Check

// Register adds a check to the global registry. Called from each
// checks/* package's init().
func Register(c Check) {
	registry = append(registry, c)
}

// All returns a copy of every registered check.
func All() []Check {
	out := make([]Check, len(registry))
	copy(out, registry)
	return out
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/check/... -v`
Expected: PASS (all 4 tests).

- [ ] **Step 5: Commit**

```bash
git add internal/check
git commit -m "add check/finding types and global check registry"
```

---

### Task 3: Language classification

**Files:**
- Create: `internal/scan/lang.go`
- Test: `internal/scan/lang_test.go`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: `func LangForFile(path string) string` and `func MatchesLangs(langs []string, fileLang string) bool`, both used by Task 6.

- [ ] **Step 1: Write the failing tests**

```go
// internal/scan/lang_test.go
package scan_test

import (
	"testing"

	"aint/internal/scan"
)

func TestLangForFile(t *testing.T) {
	cases := map[string]string{
		"main.go":       "go",
		"App.swift":     "swift",
		"script.py":     "python",
		"index.js":      "nodejs",
		"index.ts":      "nodejs",
		"deploy.sh":     "shell",
		"README.md":     "",
		"path/to/a.go":  "go",
	}
	for path, want := range cases {
		if got := scan.LangForFile(path); got != want {
			t.Errorf("LangForFile(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestMatchesLangs(t *testing.T) {
	if !scan.MatchesLangs(nil, "go") {
		t.Error("nil Langs should match any file language")
	}
	if !scan.MatchesLangs([]string{}, "shell") {
		t.Error("empty Langs should match any file language, including shell")
	}
	if !scan.MatchesLangs([]string{"go", "swift"}, "swift") {
		t.Error("swift should match when listed")
	}
	if scan.MatchesLangs([]string{"go"}, "python") {
		t.Error("python should not match when only go is listed")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/scan/... -v`
Expected: FAIL — package `aint/internal/scan` does not exist yet.

- [ ] **Step 3: Implement**

```go
// internal/scan/lang.go
package scan

import "strings"

// LangForFile classifies a file path into a language tag by extension.
// Returns "" if the extension isn't recognized.
func LangForFile(path string) string {
	switch {
	case strings.HasSuffix(path, ".go"):
		return "go"
	case strings.HasSuffix(path, ".swift"):
		return "swift"
	case strings.HasSuffix(path, ".py"):
		return "python"
	case strings.HasSuffix(path, ".js"), strings.HasSuffix(path, ".ts"):
		return "nodejs"
	case strings.HasSuffix(path, ".sh"):
		return "shell"
	default:
		return ""
	}
}

// MatchesLangs reports whether a check whose Langs field is langs applies
// to a file classified as fileLang. An empty/nil langs matches everything.
func MatchesLangs(langs []string, fileLang string) bool {
	if len(langs) == 0 {
		return true
	}
	for _, l := range langs {
		if l == fileLang {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/scan/... -v`
Expected: PASS (2 tests).

- [ ] **Step 5: Commit**

```bash
git add internal/scan/lang.go internal/scan/lang_test.go
git commit -m "add file language classification for scan targets"
```

---

### Task 4: Config loading

**Files:**
- Create: `internal/config/config.go`
- Test: `internal/config/config_test.go`

**Interfaces:**
- Consumes: `check.Check`, `check.Severity`, `check.SeverityError` from Task 2.
- Produces:
  - `type Config struct { FailOn check.Severity; Ignore []string; Checks map[string]string; DocsBaseURL string }`.
  - `func Default() Config`.
  - `func Load(path string) (Config, error)` — returns `Default()` unmodified if the file doesn't exist.
  - `func (c Config) IsIgnored(path string) bool`.
  - `func (c Config) SeverityFor(chk check.Check) (sev check.Severity, enabled bool)`.
  - `func ResolveDocsURL(c Config, docsFile string) string`.

- [ ] **Step 1: Add the yaml dependency**

Run: `go get gopkg.in/yaml.v3`
Expected: `go.mod`/`go.sum` updated with the dependency.

- [ ] **Step 2: Write the failing tests**

```go
// internal/config/config_test.go
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
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./internal/config/... -v`
Expected: FAIL — package `aint/internal/config` does not exist yet.

- [ ] **Step 4: Implement**

```go
// internal/config/config.go
package config

import (
	"os"
	"path/filepath"
	"strings"

	"aint/internal/check"
	"gopkg.in/yaml.v3"
)

type Config struct {
	FailOn      check.Severity    `yaml:"fail_on"`
	Ignore      []string          `yaml:"ignore"`
	Checks      map[string]string `yaml:"checks"`
	DocsBaseURL string            `yaml:"docs_base_url"`
}

func Default() Config {
	return Config{FailOn: check.SeverityError}
}

// Load reads and parses path as YAML, returning Default() unmodified if
// the file does not exist.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Default(), nil
	}
	if err != nil {
		return Config{}, err
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.FailOn == "" {
		cfg.FailOn = check.SeverityError
	}
	return cfg, nil
}

// IsIgnored reports whether path matches any of the configured ignore
// globs. Supports a "dir/**" prefix form and single-segment "*.ext" globs
// via filepath.Match against the basename or full path.
func (c Config) IsIgnored(path string) bool {
	path = filepath.ToSlash(path)
	for _, pattern := range c.Ignore {
		pattern = filepath.ToSlash(pattern)
		if strings.HasSuffix(pattern, "/**") {
			prefix := strings.TrimSuffix(pattern, "/**")
			if path == prefix || strings.HasPrefix(path, prefix+"/") {
				return true
			}
			continue
		}
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
	}
	return false
}

// SeverityFor returns the effective severity for chk given config
// overrides, and whether the check is enabled at all ("off" disables it).
func (c Config) SeverityFor(chk check.Check) (check.Severity, bool) {
	override, ok := c.Checks[chk.ID]
	if !ok {
		return chk.Severity, true
	}
	if override == "off" {
		return "", false
	}
	return check.Severity(override), true
}

// ResolveDocsURL builds the docs link for a finding: docsBaseURL + file if
// configured, else a relative docs/checks/<file> path.
func ResolveDocsURL(c Config, docsFile string) string {
	if c.DocsBaseURL != "" {
		return strings.TrimSuffix(c.DocsBaseURL, "/") + "/" + docsFile
	}
	return "docs/checks/" + docsFile
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/config/... -v`
Expected: PASS (6 tests).

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum internal/config
git commit -m "add .aint.yaml config loading, ignore globs, and severity overrides"
```

---

### Task 5: Report formatters

**Files:**
- Create: `internal/report/report.go`
- Test: `internal/report/report_test.go`

**Interfaces:**
- Consumes: `check.Finding` from Task 2.
- Produces: `func WriteText(w io.Writer, findings []check.Finding)` and `func WriteJSON(w io.Writer, findings []check.Finding) error`.

- [ ] **Step 1: Write the failing tests**

```go
// internal/report/report_test.go
package report_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"aint/internal/check"
	"aint/internal/report"
)

func sampleFindings() []check.Finding {
	return []check.Finding{
		{
			CheckID:  "go-ignored-error",
			Severity: check.SeverityWarning,
			Source:   "main.go",
			Line:     42,
			Column:   5,
			Message:  "error return value is discarded",
			DocsURL:  "docs/checks/go-ignored-error.md",
		},
	}
}

func TestWriteText(t *testing.T) {
	var buf bytes.Buffer
	report.WriteText(&buf, sampleFindings())
	out := buf.String()
	want := "main.go:42:5: warning [go-ignored-error] error return value is discarded — docs/checks/go-ignored-error.md\n"
	if out != want {
		t.Errorf("WriteText output mismatch:\ngot:  %q\nwant: %q", out, want)
	}
}

func TestWriteTextEmpty(t *testing.T) {
	var buf bytes.Buffer
	report.WriteText(&buf, nil)
	if buf.Len() != 0 {
		t.Errorf("expected no output for zero findings, got %q", buf.String())
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := report.WriteJSON(&buf, sampleFindings()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var decoded []check.Finding
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if len(decoded) != 1 || decoded[0].CheckID != "go-ignored-error" {
		t.Errorf("unexpected decoded findings: %+v", decoded)
	}
	if !strings.Contains(buf.String(), "\"check_id\"") {
		t.Errorf("expected snake_case check_id field in JSON output, got: %s", buf.String())
	}
}
```

This requires `Finding` fields to carry JSON tags — add them in this task since Task 2 didn't need JSON output yet.

- [ ] **Step 2: Add JSON tags to Finding**

```go
// internal/check/types.go — update the Finding struct to:
type Finding struct {
	CheckID  string   `json:"check_id"`
	Severity Severity `json:"severity"`
	Source   string   `json:"source"`
	Line     int      `json:"line"`
	Column   int      `json:"column"`
	Message  string   `json:"message"`
	DocsURL  string   `json:"docs_url"`
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./internal/report/... -v`
Expected: FAIL — package `aint/internal/report` does not exist yet.

- [ ] **Step 4: Implement**

```go
// internal/report/report.go
package report

import (
	"encoding/json"
	"fmt"
	"io"

	"aint/internal/check"
)

// WriteText writes one human-readable line per finding:
// source:line:col: severity [check-id] message — docs-url
func WriteText(w io.Writer, findings []check.Finding) {
	for _, f := range findings {
		fmt.Fprintf(w, "%s:%d:%d: %s [%s] %s — %s\n",
			f.Source, f.Line, f.Column, f.Severity, f.CheckID, f.Message, f.DocsURL)
	}
}

// WriteJSON writes findings as an indented JSON array.
func WriteJSON(w io.Writer, findings []check.Finding) error {
	if findings == nil {
		findings = []check.Finding{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(findings)
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/report/... ./internal/check/... -v`
Expected: PASS (all report tests, and check tests still pass after the Finding struct change).

- [ ] **Step 6: Commit**

```bash
git add internal/report internal/check/types.go
git commit -m "add text and JSON finding formatters"
```

---

### Task 6: Scanner (file walk + shell command target + check dispatch)

**Files:**
- Create: `internal/scan/scan.go`
- Test: `internal/scan/scan_test.go`

**Interfaces:**
- Consumes: `check.Check`, `check.Finding` (Task 2); `config.Config`, `config.ResolveDocsURL` (Task 4); `LangForFile`, `MatchesLangs` (Task 3).
- Produces:
  - `type Target struct { Source, Lang string; Content []byte }`.
  - `func CommandTarget(command string) Target` — `Source: "<command>"`, `Lang: "shell"`.
  - `func Walk(paths []string, cfg config.Config) ([]Target, error)`.
  - `func Run(targets []Target, checks []check.Check, cfg config.Config) []check.Finding`.

- [ ] **Step 1: Write the failing tests**

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/scan/... -v`
Expected: FAIL — `Target`, `CommandTarget`, `Walk`, `Run` undefined.

- [ ] **Step 3: Implement**

```go
// internal/scan/scan.go
package scan

import (
	"os"
	"path/filepath"

	"aint/internal/check"
	"aint/internal/config"
)

// Target is a single file or synthetic shell-command blob to run checks
// against.
type Target struct {
	Source  string
	Lang    string
	Content []byte
}

// CommandTarget wraps a raw shell command string (from a Claude Code
// PreToolUse Bash hook payload) as a Target classified as "shell".
func CommandTarget(command string) Target {
	return Target{Source: "<command>", Lang: "shell", Content: []byte(command)}
}

// Walk reads every non-ignored file under the given paths into Targets.
func Walk(paths []string, cfg config.Config) ([]Target, error) {
	var targets []Target
	for _, root := range paths {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if cfg.IsIgnored(path) {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			targets = append(targets, Target{
				Source:  path,
				Lang:    LangForFile(path),
				Content: content,
			})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return targets, nil
}

// Run runs every check against every matching target, applying config
// severity overrides/disables, and returns all findings.
func Run(targets []Target, checks []check.Check, cfg config.Config) []check.Finding {
	var findings []check.Finding
	for _, c := range checks {
		sev, enabled := cfg.SeverityFor(c)
		if !enabled {
			continue
		}
		effective := c
		effective.Severity = sev
		docsURL := config.ResolveDocsURL(cfg, c.DocsPath)

		for _, t := range targets {
			if !MatchesLangs(c.Langs, t.Lang) {
				continue
			}
			findings = append(findings, effective.Run(t.Source, t.Content, docsURL)...)
		}
	}
	return findings
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/scan/... -v`
Expected: PASS (all scan tests).

- [ ] **Step 5: Commit**

```bash
git add internal/scan/scan.go internal/scan/scan_test.go
git commit -m "add file walker, shell command target, and check dispatch"
```

---

### Task 7: Seed checks — secrets

**Files:**
- Create: `internal/checks/secrets/secrets.go`
- Test: `internal/checks/secrets/secrets_test.go`

**Interfaces:**
- Consumes: `check.Check`, `check.Register`, `check.SeverityError` (Task 2).
- Produces: package-level vars `secrets.HardcodedKey` and `secrets.PrivateKeyBlock` (exported so tests and later docs work can reference them directly), registered via `init()`.

- [ ] **Step 1: Write the failing tests**

```go
// internal/checks/secrets/secrets_test.go
package secrets_test

import (
	"testing"

	"aint/internal/checks/secrets"
)

func TestHardcodedKeyDetectsAWSKey(t *testing.T) {
	findings := secrets.HardcodedKey.Run("test.go", []byte(`key := "AKIAABCDEFGHIJKLMNOP"`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestHardcodedKeyDetectsGenericSecretAssignment(t *testing.T) {
	findings := secrets.HardcodedKey.Run("config.py", []byte(`api_key = "sk-abcdefghijklmnopqrstuvwx"`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestHardcodedKeyIgnoresShortStrings(t *testing.T) {
	findings := secrets.HardcodedKey.Run("test.go", []byte(`name := "hello"`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestPrivateKeyBlockDetectsCommittedKey(t *testing.T) {
	content := "-----BEGIN RSA PRIVATE KEY-----\nMIIB...\n-----END RSA PRIVATE KEY-----"
	findings := secrets.PrivateKeyBlock.Run("id_rsa", []byte(content), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestPrivateKeyBlockIgnoresUnrelatedFiles(t *testing.T) {
	findings := secrets.PrivateKeyBlock.Run("readme.md", []byte("# hello world"), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/checks/secrets/... -v`
Expected: FAIL — package does not exist yet.

- [ ] **Step 3: Implement**

```go
// internal/checks/secrets/secrets.go
package secrets

import (
	"regexp"

	"aint/internal/check"
)

var HardcodedKey = check.Check{
	ID:       "secret-hardcoded-key",
	Title:    "Hardcoded API key or token",
	Severity: check.SeverityError,
	Pattern: regexp.MustCompile(
		`AKIA[0-9A-Z]{16}|sk-[a-zA-Z0-9]{20,}|(?i)(api_key|apikey|secret|token)\s*[:=]\s*["'][^"']{16,}["']`,
	),
	Message:  "possible hardcoded API key or token",
	DocsPath: "secret-hardcoded-key.md",
}

var PrivateKeyBlock = check.Check{
	ID:       "secret-private-key-block",
	Title:    "Committed private key material",
	Severity: check.SeverityError,
	Pattern:  regexp.MustCompile(`-----BEGIN (RSA |EC )?PRIVATE KEY-----`),
	Message:  "private key material committed to file",
	DocsPath: "secret-private-key-block.md",
}

func init() {
	check.Register(HardcodedKey)
	check.Register(PrivateKeyBlock)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/checks/secrets/... -v`
Expected: PASS (5 tests).

- [ ] **Step 5: Commit**

```bash
git add internal/checks/secrets
git commit -m "add secret-hardcoded-key and secret-private-key-block checks"
```

---

### Task 8: Seed checks — shell / IaC overscoping

**Files:**
- Create: `internal/checks/shell/shell.go`
- Test: `internal/checks/shell/shell_test.go`

**Interfaces:**
- Consumes: `check.Check`, `check.Register`, `check.SeverityError`, `check.SeverityWarning` (Task 2).
- Produces: package-level vars `shell.GCPRoleWildcard`, `shell.ChmodPermissive`, registered via `init()`.

- [ ] **Step 1: Write the failing tests**

```go
// internal/checks/shell/shell_test.go
package shell_test

import (
	"testing"

	"aint/internal/checks/shell"
)

func TestGCPRoleWildcardDetectsOwnerGrant(t *testing.T) {
	cmd := "gcloud projects add-iam-policy-binding my-project --member=user:x@example.com --role=roles/owner"
	findings := shell.GCPRoleWildcard.Run("<command>", []byte(cmd), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestGCPRoleWildcardDetectsEditorGrant(t *testing.T) {
	cmd := "gcloud projects add-iam-policy-binding my-project --role=roles/editor"
	findings := shell.GCPRoleWildcard.Run("<command>", []byte(cmd), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestGCPRoleWildcardIgnoresScopedRole(t *testing.T) {
	cmd := "gcloud projects add-iam-policy-binding my-project --role=roles/logging.viewer"
	findings := shell.GCPRoleWildcard.Run("<command>", []byte(cmd), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestChmodPermissiveDetects777(t *testing.T) {
	findings := shell.ChmodPermissive.Run("<command>", []byte("chmod 777 script.sh"), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestChmodPermissiveIgnoresNarrowMode(t *testing.T) {
	findings := shell.ChmodPermissive.Run("<command>", []byte("chmod 755 script.sh"), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/checks/shell/... -v`
Expected: FAIL — package does not exist yet.

- [ ] **Step 3: Implement**

```go
// internal/checks/shell/shell.go
package shell

import (
	"regexp"

	"aint/internal/check"
)

var GCPRoleWildcard = check.Check{
	ID:       "shell-gcp-role-wildcard",
	Title:    "Overscoped GCP IAM role grant",
	Severity: check.SeverityError,
	Langs:    []string{"shell"},
	Pattern:  regexp.MustCompile(`add-iam-policy-binding.*--role[= ]roles/(owner|editor)\b`),
	Message:  "granting roles/owner or roles/editor is overscoped; use a narrower predefined or custom role",
	DocsPath: "shell-gcp-role-wildcard.md",
}

var ChmodPermissive = check.Check{
	ID:       "shell-chmod-permissive",
	Title:    "World-writable chmod",
	Severity: check.SeverityWarning,
	Langs:    []string{"shell"},
	Pattern:  regexp.MustCompile(`chmod\s+(-R\s+)?(777|a\+rwx|ugo\+rwx)\b`),
	Message:  "world-writable permissions granted via chmod",
	DocsPath: "shell-chmod-permissive.md",
}

func init() {
	check.Register(GCPRoleWildcard)
	check.Register(ChmodPermissive)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/checks/shell/... -v`
Expected: PASS (5 tests).

- [ ] **Step 5: Commit**

```bash
git add internal/checks/shell
git commit -m "add shell-gcp-role-wildcard and shell-chmod-permissive checks"
```

---

### Task 9: Seed checks — language footguns (Go, Swift, Python, Node)

**Files:**
- Create: `internal/checks/golang/golang.go`, `internal/checks/golang/golang_test.go`
- Create: `internal/checks/swift/swift.go`, `internal/checks/swift/swift_test.go`
- Create: `internal/checks/python/python.go`, `internal/checks/python/python_test.go`
- Create: `internal/checks/nodejs/nodejs.go`, `internal/checks/nodejs/nodejs_test.go`

**Interfaces:**
- Consumes: `check.Check`, `check.Register`, `check.SeverityWarning` (Task 2).
- Produces: `golang.IgnoredError`, `swift.ForceUnwrap`, `python.ShellTrue`, `nodejs.Eval`, each registered via its package's `init()`.

- [ ] **Step 1: Write the failing test for Go**

```go
// internal/checks/golang/golang_test.go
package golang_test

import (
	"testing"

	"aint/internal/checks/golang"
)

func TestIgnoredErrorDetectsDiscard(t *testing.T) {
	findings := golang.IgnoredError.Run("main.go", []byte("_ = err"), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestIgnoredErrorIgnoresHandledError(t *testing.T) {
	findings := golang.IgnoredError.Run("main.go", []byte("result, err := doSomething()"), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./internal/checks/golang/... -v`
Expected: FAIL — package does not exist yet.

- [ ] **Step 3: Implement the Go check**

```go
// internal/checks/golang/golang.go
package golang

import (
	"regexp"

	"aint/internal/check"
)

var IgnoredError = check.Check{
	ID:       "go-ignored-error",
	Title:    "Ignored error return value",
	Severity: check.SeverityWarning,
	Langs:    []string{"go"},
	Pattern:  regexp.MustCompile(`\b_\s*=\s*err\b`),
	Message:  "error value discarded via `_ = err`",
	DocsPath: "go-ignored-error.md",
}

func init() {
	check.Register(IgnoredError)
}
```

- [ ] **Step 4: Run it to verify it passes**

Run: `go test ./internal/checks/golang/... -v`
Expected: PASS (2 tests).

- [ ] **Step 5: Write the failing test for Swift**

```go
// internal/checks/swift/swift_test.go
package swift_test

import (
	"testing"

	"aint/internal/checks/swift"
)

func TestForceUnwrapDetectsTryBang(t *testing.T) {
	findings := swift.ForceUnwrap.Run("main.swift", []byte("let x = try! risky()"), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestForceUnwrapDetectsAsBang(t *testing.T) {
	findings := swift.ForceUnwrap.Run("main.swift", []byte("let s = value as! String"), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestForceUnwrapIgnoresSafeTry(t *testing.T) {
	findings := swift.ForceUnwrap.Run("main.swift", []byte("let x = try? risky()"), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
```

- [ ] **Step 6: Run it to verify it fails**

Run: `go test ./internal/checks/swift/... -v`
Expected: FAIL — package does not exist yet.

- [ ] **Step 7: Implement the Swift check**

```go
// internal/checks/swift/swift.go
package swift

import (
	"regexp"

	"aint/internal/check"
)

var ForceUnwrap = check.Check{
	ID:       "swift-force-unwrap",
	Title:    "Force unwrap or force cast",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\b(try|as)!`),
	Message:  "force unwrap/cast can crash at runtime; prefer try?/as? with explicit handling",
	DocsPath: "swift-force-unwrap.md",
}

func init() {
	check.Register(ForceUnwrap)
}
```

- [ ] **Step 8: Run it to verify it passes**

Run: `go test ./internal/checks/swift/... -v`
Expected: PASS (3 tests).

- [ ] **Step 9: Write the failing test for Python**

```go
// internal/checks/python/python_test.go
package python_test

import (
	"testing"

	"aint/internal/checks/python"
)

func TestShellTrueDetectsShellInjectionRisk(t *testing.T) {
	findings := python.ShellTrue.Run("script.py", []byte(`subprocess.run(cmd, shell=True)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestShellTrueIgnoresSafeCall(t *testing.T) {
	findings := python.ShellTrue.Run("script.py", []byte(`subprocess.run(cmd)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
```

- [ ] **Step 10: Run it to verify it fails**

Run: `go test ./internal/checks/python/... -v`
Expected: FAIL — package does not exist yet.

- [ ] **Step 11: Implement the Python check**

```go
// internal/checks/python/python.go
package python

import (
	"regexp"

	"aint/internal/check"
)

var ShellTrue = check.Check{
	ID:       "python-shell-true",
	Title:    "subprocess call with shell=True",
	Severity: check.SeverityWarning,
	Langs:    []string{"python"},
	Pattern:  regexp.MustCompile(`subprocess\.\w+\([^)]*shell\s*=\s*True`),
	Message:  "shell=True risks shell injection if any part of the command is untrusted input",
	DocsPath: "python-shell-true.md",
}

func init() {
	check.Register(ShellTrue)
}
```

- [ ] **Step 12: Run it to verify it passes**

Run: `go test ./internal/checks/python/... -v`
Expected: PASS (2 tests).

- [ ] **Step 13: Write the failing test for Node**

```go
// internal/checks/nodejs/nodejs_test.go
package nodejs_test

import (
	"testing"

	"aint/internal/checks/nodejs"
)

func TestEvalDetectsCall(t *testing.T) {
	findings := nodejs.Eval.Run("index.js", []byte(`eval(userInput)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestEvalIgnoresSimilarIdentifier(t *testing.T) {
	findings := nodejs.Eval.Run("index.js", []byte(`evaluate(userInput)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
```

- [ ] **Step 14: Run it to verify it fails**

Run: `go test ./internal/checks/nodejs/... -v`
Expected: FAIL — package does not exist yet.

- [ ] **Step 15: Implement the Node check**

```go
// internal/checks/nodejs/nodejs.go
package nodejs

import (
	"regexp"

	"aint/internal/check"
)

var Eval = check.Check{
	ID:       "node-eval",
	Title:    "Use of eval()",
	Severity: check.SeverityWarning,
	Langs:    []string{"nodejs"},
	Pattern:  regexp.MustCompile(`\beval\s*\(`),
	Message:  "eval() executes arbitrary strings as code; avoid it or use a safer parser",
	DocsPath: "node-eval.md",
}

func init() {
	check.Register(Eval)
}
```

- [ ] **Step 16: Run it to verify it passes**

Run: `go test ./internal/checks/nodejs/... -v`
Expected: PASS (2 tests).

- [ ] **Step 17: Run all four packages together and commit**

Run: `go test ./internal/checks/... -v`
Expected: PASS (secrets, shell, golang, swift, python, nodejs all pass).

```bash
git add internal/checks/golang internal/checks/swift internal/checks/python internal/checks/nodejs
git commit -m "add go-ignored-error, swift-force-unwrap, python-shell-true, node-eval checks"
```

---

### Task 10: `aint check` and `aint list` CLI commands

**Files:**
- Create: `internal/checks/register.go`
- Modify: `cmd/aint/main.go`
- Create: `cmd/aint/check.go`
- Create: `cmd/aint/list.go`
- Test: `cmd/aint/check_test.go`
- Test fixtures: `cmd/aint/testdata/checkfixture/main.go`, `cmd/aint/testdata/checkfixture/clean.go`

**Interfaces:**
- Consumes: `check.All()` (Task 2), `config.Load` (Task 4), `scan.Walk`/`scan.Run` (Task 6), `report.WriteText`/`report.WriteJSON` (Task 5), and every seed check package (Tasks 7-9) via the new blank-import aggregator.
- Produces: `func runCheck(args []string) int`, `func runList(args []string) int`, wired into `dispatch` in `main.go`.

- [ ] **Step 1: Create the checks aggregator package**

```go
// internal/checks/register.go
package checks

import (
	_ "aint/internal/checks/golang"
	_ "aint/internal/checks/nodejs"
	_ "aint/internal/checks/python"
	_ "aint/internal/checks/secrets"
	_ "aint/internal/checks/shell"
	_ "aint/internal/checks/swift"
)
```

- [ ] **Step 2: Write the failing end-to-end test**

```go
// cmd/aint/check_test.go
package main

import (
	"bytes"
	"testing"
)

func TestRunCheckFindsSeedViolationAndReturnsNonZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCheckWithIO([]string{"testdata/checkfixture/main.go"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d (stderr: %s)", code, stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte("go-ignored-error")) {
		t.Errorf("expected output to mention go-ignored-error, got: %s", stdout.String())
	}
}

func TestRunCheckOnCleanFileReturnsZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCheckWithIO([]string{"testdata/checkfixture/clean.go"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
	}
}
```

- [ ] **Step 3: Add the test fixtures**

```go
// cmd/aint/testdata/checkfixture/main.go
package checkfixture

func doWork() {
	_ = err
}
```

```go
// cmd/aint/testdata/checkfixture/clean.go
package checkfixture

func doWorkProperly() error {
	return nil
}
```

- [ ] **Step 4: Run the test to verify it fails**

Run: `go test ./cmd/aint/... -v -run TestRunCheck`
Expected: FAIL — `runCheckWithIO` undefined.

- [ ] **Step 5: Implement `runCheck` / `runCheckWithIO` and `runList`**

```go
// cmd/aint/check.go
package main

import (
	"io"
	"os"
	"strings"

	"aint/internal/check"
	"aint/internal/config"
	"aint/internal/report"
	"aint/internal/scan"
)

func runCheck(args []string) int {
	return runCheckWithIO(args, os.Stdout, os.Stderr)
}

func runCheckWithIO(args []string, stdout, stderr io.Writer) int {
	format := "text"
	var paths []string
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--format" && i+1 < len(args):
			i++
			format = args[i]
		case strings.HasPrefix(args[i], "--format="):
			format = strings.TrimPrefix(args[i], "--format=")
		default:
			paths = append(paths, args[i])
		}
	}
	if len(paths) == 0 {
		paths = []string{"."}
	}

	cfg, err := config.Load(".aint.yaml")
	if err != nil {
		writeErr(stderr, "loading config", err)
		return 1
	}

	targets, err := scan.Walk(paths, cfg)
	if err != nil {
		writeErr(stderr, "scanning", err)
		return 1
	}

	findings := scan.Run(targets, check.All(), cfg)

	if format == "json" {
		if err := report.WriteJSON(stdout, findings); err != nil {
			writeErr(stderr, "writing JSON output", err)
			return 1
		}
	} else {
		report.WriteText(stdout, findings)
	}

	for _, f := range findings {
		if f.Severity.AtLeast(cfg.FailOn) {
			return 1
		}
	}
	return 0
}

func writeErr(w io.Writer, action string, err error) {
	io.WriteString(w, "aint: "+action+": "+err.Error()+"\n")
}
```

```go
// cmd/aint/list.go
package main

import (
	"fmt"
	"os"

	"aint/internal/check"
)

func runList(args []string) int {
	for _, c := range check.All() {
		fmt.Fprintf(os.Stdout, "%-28s %-8s %-20v %s\n", c.ID, c.Severity, c.Langs, c.Title)
	}
	return 0
}
```

- [ ] **Step 6: Wire the aggregator import and subcommands into main.go**

```go
// cmd/aint/main.go
package main

import (
	"fmt"
	"os"

	_ "aint/internal/checks"
)

func main() {
	os.Exit(dispatch(os.Args))
}

func dispatch(args []string) int {
	if len(args) < 2 {
		printUsage()
		return 2
	}
	switch args[1] {
	case "check":
		return runCheck(args[2:])
	case "list":
		return runList(args[2:])
	default:
		printUsage()
		return 2
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `usage: aint <command> [args]

commands:
  check [paths...]            scan files/dirs for issues
  list                        list all registered checks
  install [--global]          wire aint into Claude Code hooks
  hook <pre-bash|post-edit>   internal: used by installed hooks`)
}
```

- [ ] **Step 7: Run the tests to verify they pass**

Run: `go test ./... -v`
Expected: PASS across every package.

- [ ] **Step 8: Manual smoke test**

Run: `go run ./cmd/aint check cmd/aint/testdata/checkfixture/main.go`
Expected: prints a line mentioning `go-ignored-error`; `echo $?` shows `1`.

Run: `go run ./cmd/aint list`
Expected: prints all 8 registered checks.

- [ ] **Step 9: Commit**

```bash
git add internal/checks/register.go cmd/aint
git commit -m "wire aint check and aint list CLI commands to the scan engine"
```

---

### Task 11: Claude Code hook subcommands

**Files:**
- Create: `internal/hook/payload.go`
- Test: `internal/hook/payload_test.go`
- Create: `cmd/aint/hook.go`
- Test: `cmd/aint/hook_test.go`

**Interfaces:**
- Consumes: `check.All()`, `check.Finding` (Task 2); `config.Config`, `config.Load` (Task 4); `scan.CommandTarget`, `scan.Target`, `scan.Run`, `LangForFile` (Task 6); `report.WriteText` (Task 5).
- Produces:
  - `type PreToolUsePayload struct { ToolName string; ToolInput struct{ Command string } }` and `func ParsePreToolUse(data []byte) (PreToolUsePayload, error)`.
  - `type PostToolUsePayload struct { ToolName string; ToolInput struct{ FilePath string } }` and `func ParsePostToolUse(data []byte) (PostToolUsePayload, error)`.
  - `func runHookWithIO(args []string, stdin io.Reader, stdout, stderr io.Writer) int`, wired into `dispatch` as `runHook`.

- [ ] **Step 1: Write the failing tests for payload parsing**

```go
// internal/hook/payload_test.go
package hook_test

import (
	"testing"

	"aint/internal/hook"
)

func TestParsePreToolUse(t *testing.T) {
	data := []byte(`{"tool_name":"Bash","tool_input":{"command":"chmod 777 x.sh"}}`)
	p, err := hook.ParsePreToolUse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ToolInput.Command != "chmod 777 x.sh" {
		t.Errorf("unexpected command: %q", p.ToolInput.Command)
	}
}

func TestParsePostToolUse(t *testing.T) {
	data := []byte(`{"tool_name":"Write","tool_input":{"file_path":"/tmp/main.go"}}`)
	p, err := hook.ParsePostToolUse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ToolInput.FilePath != "/tmp/main.go" {
		t.Errorf("unexpected file path: %q", p.ToolInput.FilePath)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./internal/hook/... -v`
Expected: FAIL — package does not exist yet.

- [ ] **Step 3: Implement payload parsing**

```go
// internal/hook/payload.go
package hook

import "encoding/json"

type PreToolUsePayload struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
}

type PostToolUsePayload struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
}

func ParsePreToolUse(data []byte) (PreToolUsePayload, error) {
	var p PreToolUsePayload
	err := json.Unmarshal(data, &p)
	return p, err
}

func ParsePostToolUse(data []byte) (PostToolUsePayload, error) {
	var p PostToolUsePayload
	err := json.Unmarshal(data, &p)
	return p, err
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./internal/hook/... -v`
Expected: PASS (2 tests).

- [ ] **Step 5: Write the failing tests for the hook CLI subcommand**

```go
// cmd/aint/hook_test.go
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunHookPreBashBlocksOverscopedGrant(t *testing.T) {
	stdin := bytes.NewBufferString(`{"tool_name":"Bash","tool_input":{"command":"gcloud projects add-iam-policy-binding p --role=roles/owner"}}`)
	var stdout, stderr bytes.Buffer
	code := runHookWithIO([]string{"pre-bash"}, stdin, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d (stderr: %s)", code, stderr.String())
	}
	if !bytes.Contains(stderr.Bytes(), []byte("shell-gcp-role-wildcard")) {
		t.Errorf("expected stderr to mention shell-gcp-role-wildcard, got: %s", stderr.String())
	}
}

func TestRunHookPreBashAllowsScopedGrant(t *testing.T) {
	stdin := bytes.NewBufferString(`{"tool_name":"Bash","tool_input":{"command":"gcloud projects add-iam-policy-binding p --role=roles/logging.viewer"}}`)
	var stdout, stderr bytes.Buffer
	code := runHookWithIO([]string{"pre-bash"}, stdin, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", code, stderr.String())
	}
}

func TestRunHookPostEditReportsFindingsInEditedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte("_ = err"), 0644); err != nil {
		t.Fatal(err)
	}

	stdin := bytes.NewBufferString(`{"tool_name":"Write","tool_input":{"file_path":"` + path + `"}}`)
	var stdout, stderr bytes.Buffer
	code := runHookWithIO([]string{"post-edit"}, stdin, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d (stderr: %s)", code, stderr.String())
	}
	if !bytes.Contains(stderr.Bytes(), []byte("go-ignored-error")) {
		t.Errorf("expected stderr to mention go-ignored-error, got: %s", stderr.String())
	}
}
```

- [ ] **Step 6: Run to verify it fails**

Run: `go test ./cmd/aint/... -v -run TestRunHook`
Expected: FAIL — `runHookWithIO` undefined.

- [ ] **Step 7: Implement the hook subcommand**

```go
// cmd/aint/hook.go
package main

import (
	"io"
	"os"

	"aint/internal/check"
	"aint/internal/config"
	"aint/internal/hook"
	"aint/internal/report"
	"aint/internal/scan"
)

func runHook(args []string) int {
	return runHookWithIO(args, os.Stdin, os.Stdout, os.Stderr)
}

func runHookWithIO(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		io.WriteString(stderr, "usage: aint hook <pre-bash|post-edit>\n")
		return 2
	}

	data, err := io.ReadAll(stdin)
	if err != nil {
		writeErr(stderr, "reading stdin", err)
		return 1
	}

	cfg, err := config.Load(".aint.yaml")
	if err != nil {
		writeErr(stderr, "loading config", err)
		return 1
	}

	switch args[0] {
	case "pre-bash":
		return runPreBash(data, cfg, stderr)
	case "post-edit":
		return runPostEdit(data, cfg, stderr)
	default:
		io.WriteString(stderr, "usage: aint hook <pre-bash|post-edit>\n")
		return 2
	}
}

func runPreBash(data []byte, cfg config.Config, stderr io.Writer) int {
	payload, err := hook.ParsePreToolUse(data)
	if err != nil {
		writeErr(stderr, "parsing hook payload", err)
		return 1
	}
	target := scan.CommandTarget(payload.ToolInput.Command)
	findings := scan.Run([]scan.Target{target}, check.All(), cfg)
	return reportHookFindings(findings, cfg, stderr)
}

func runPostEdit(data []byte, cfg config.Config, stderr io.Writer) int {
	payload, err := hook.ParsePostToolUse(data)
	if err != nil {
		writeErr(stderr, "parsing hook payload", err)
		return 1
	}
	content, err := os.ReadFile(payload.ToolInput.FilePath)
	if err != nil {
		writeErr(stderr, "reading file", err)
		return 1
	}
	target := scan.Target{
		Source:  payload.ToolInput.FilePath,
		Lang:    scan.LangForFile(payload.ToolInput.FilePath),
		Content: content,
	}
	findings := scan.Run([]scan.Target{target}, check.All(), cfg)
	return reportHookFindings(findings, cfg, stderr)
}

func reportHookFindings(findings []check.Finding, cfg config.Config, stderr io.Writer) int {
	blocking := false
	for _, f := range findings {
		if f.Severity.AtLeast(cfg.FailOn) {
			blocking = true
		}
	}
	if !blocking {
		return 0
	}
	report.WriteText(stderr, findings)
	return 2
}
```

- [ ] **Step 8: Wire `hook` into dispatch**

```go
// cmd/aint/main.go — update dispatch's switch statement to add:
	case "hook":
		return runHook(args[2:])
```

- [ ] **Step 9: Run the tests to verify they pass**

Run: `go test ./... -v`
Expected: PASS across every package.

- [ ] **Step 10: Commit**

```bash
git add internal/hook/payload.go internal/hook/payload_test.go cmd/aint/hook.go cmd/aint/hook_test.go cmd/aint/main.go
git commit -m "add aint hook pre-bash/post-edit subcommands for Claude Code integration"
```

---

### Task 12: `aint install` command

**Files:**
- Create: `internal/hook/install.go`
- Test: `internal/hook/install_test.go`
- Create: `cmd/aint/install.go`
- Test: `cmd/aint/install_test.go`

**Interfaces:**
- Consumes: nothing from other internal packages (standalone JSON manipulation).
- Produces:
  - `func LoadSettings(path string) (map[string]interface{}, error)`.
  - `func Install(settings map[string]interface{}) (map[string]interface{}, []string)` — returns the updated settings and a list of human-readable "what was added" strings (empty if everything was already present).
  - `func WriteSettings(path string, settings map[string]interface{}) error`.
  - `func runInstallWithIO(args []string, stdout, stderr io.Writer) int`, wired into `dispatch` as `runInstall`.

- [ ] **Step 1: Write the failing tests for the install merge logic**

```go
// internal/hook/install_test.go
package hook_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"aint/internal/hook"
)

func TestInstallOnEmptySettingsAddsBothHooks(t *testing.T) {
	settings := map[string]interface{}{}
	updated, added := hook.Install(settings)

	if len(added) != 2 {
		t.Fatalf("expected 2 additions, got %d: %v", len(added), added)
	}

	hooks := updated["hooks"].(map[string]interface{})
	if _, ok := hooks["PreToolUse"]; !ok {
		t.Error("expected PreToolUse to be set")
	}
	if _, ok := hooks["PostToolUse"]; !ok {
		t.Error("expected PostToolUse to be set")
	}
}

func TestInstallIsIdempotent(t *testing.T) {
	settings := map[string]interface{}{}
	updated, _ := hook.Install(settings)
	_, addedSecondRun := hook.Install(updated)

	if len(addedSecondRun) != 0 {
		t.Fatalf("expected no additions on second run, got %v", addedSecondRun)
	}
}

func TestInstallPreservesUnrelatedSettings(t *testing.T) {
	settings := map[string]interface{}{
		"model": "claude-sonnet-5",
		"hooks": map[string]interface{}{
			"PreToolUse": []interface{}{
				map[string]interface{}{
					"matcher": "Read",
					"hooks": []interface{}{
						map[string]interface{}{"type": "command", "command": "some-other-tool check"},
					},
				},
			},
		},
	}
	updated, added := hook.Install(settings)

	if updated["model"] != "claude-sonnet-5" {
		t.Error("expected unrelated top-level key to survive")
	}
	pre := updated["hooks"].(map[string]interface{})["PreToolUse"].([]interface{})
	if len(pre) != 2 {
		t.Fatalf("expected existing PreToolUse entry preserved plus aint's own, got %d entries", len(pre))
	}
	if len(added) != 2 {
		t.Fatalf("expected 2 additions (Bash pre and edit post), got %d: %v", len(added), added)
	}
}

func TestLoadSettingsReturnsEmptyMapWhenFileMissing(t *testing.T) {
	settings, err := hook.LoadSettings(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(settings) != 0 {
		t.Errorf("expected empty settings, got %v", settings)
	}
}

func TestWriteSettingsThenLoadRoundTrips(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	settings := map[string]interface{}{"model": "claude-sonnet-5"}

	if err := hook.WriteSettings(path, settings); err != nil {
		t.Fatalf("unexpected error writing: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("unexpected error reading back: %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("written file is not valid JSON: %v", err)
	}
	if decoded["model"] != "claude-sonnet-5" {
		t.Errorf("unexpected round-tripped content: %v", decoded)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./internal/hook/... -v -run 'TestInstall|TestLoadSettings|TestWriteSettings'`
Expected: FAIL — `Install`, `LoadSettings`, `WriteSettings` undefined.

- [ ] **Step 3: Implement the install merge logic**

```go
// internal/hook/install.go
package hook

import (
	"encoding/json"
	"os"
	"strings"
)

const preBashCommand = "aint hook pre-bash"
const postEditCommand = "aint hook post-edit"

// LoadSettings reads a Claude Code settings.json file into a generic map,
// returning an empty map (not an error) if the file doesn't exist yet.
func LoadSettings(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]interface{}{}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return map[string]interface{}{}, nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// Install merges aint's PreToolUse/Bash and PostToolUse/Write|Edit|MultiEdit
// hook entries into settings, leaving everything else untouched. It is
// idempotent: re-running it against its own output adds nothing. Returns
// the updated settings and a human-readable list of what was added.
func Install(settings map[string]interface{}) (map[string]interface{}, []string) {
	var added []string

	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = map[string]interface{}{}
	}

	if ensureHook(hooks, "PreToolUse", "Bash", preBashCommand) {
		added = append(added, "PreToolUse/Bash -> "+preBashCommand)
	}
	if ensureHook(hooks, "PostToolUse", "Write|Edit|MultiEdit", postEditCommand) {
		added = append(added, "PostToolUse/Write|Edit|MultiEdit -> "+postEditCommand)
	}

	settings["hooks"] = hooks
	return settings, added
}

func ensureHook(hooks map[string]interface{}, event, matcher, command string) bool {
	list, _ := hooks[event].([]interface{})

	for _, entryRaw := range list {
		entry, _ := entryRaw.(map[string]interface{})
		if entry == nil {
			continue
		}
		hookList, _ := entry["hooks"].([]interface{})
		for _, hRaw := range hookList {
			h, _ := hRaw.(map[string]interface{})
			if h == nil {
				continue
			}
			if cmd, _ := h["command"].(string); strings.Contains(cmd, command) {
				return false
			}
		}
	}

	list = append(list, map[string]interface{}{
		"matcher": matcher,
		"hooks": []interface{}{
			map[string]interface{}{"type": "command", "command": command},
		},
	})
	hooks[event] = list
	return true
}

// WriteSettings writes settings back out as indented JSON.
func WriteSettings(path string, settings map[string]interface{}) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `go test ./internal/hook/... -v`
Expected: PASS (payload tests from Task 11 plus all install tests here).

- [ ] **Step 5: Write the failing test for the install CLI subcommand**

```go
// cmd/aint/install_test.go
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunInstallCreatesSettingsFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude", "settings.json")

	var stdout, stderr bytes.Buffer
	code := runInstallWithIO([]string{}, path, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", code, stderr.String())
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected settings file to be created: %v", err)
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("written settings is not valid JSON: %v", err)
	}
	if _, ok := settings["hooks"]; !ok {
		t.Error("expected hooks key in written settings")
	}
}

func TestRunInstallTwiceIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude", "settings.json")

	var stdout1, stderr1 bytes.Buffer
	runInstallWithIO([]string{}, path, &stdout1, &stderr1)

	var stdout2, stderr2 bytes.Buffer
	code := runInstallWithIO([]string{}, path, &stdout2, &stderr2)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	first, _ := os.ReadFile(path)
	var stdout3, stderr3 bytes.Buffer
	runInstallWithIO([]string{}, path, &stdout3, &stderr3)
	second, _ := os.ReadFile(path)
	if string(first) != string(second) {
		t.Error("expected settings file to be byte-identical after a third install run")
	}
}
```

Note: `runInstallWithIO` takes the resolved settings path directly (rather than
resolving `--global` internally) so tests can point it at a temp directory;
the thin `runInstall` wrapper (Step 7) does the `--global` → path resolution
before calling it.

- [ ] **Step 6: Run to verify it fails**

Run: `go test ./cmd/aint/... -v -run TestRunInstall`
Expected: FAIL — `runInstallWithIO` undefined.

- [ ] **Step 7: Implement the install subcommand**

```go
// cmd/aint/install.go
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"aint/internal/hook"
)

func runInstall(args []string) int {
	global := false
	for _, a := range args {
		if a == "--global" {
			global = true
		}
	}

	path := ".claude/settings.json"
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			writeErr(os.Stderr, "resolving home dir", err)
			return 1
		}
		path = filepath.Join(home, ".claude", "settings.json")
	}

	return runInstallWithIO(args, path, os.Stdout, os.Stderr)
}

func runInstallWithIO(args []string, path string, stdout, stderr io.Writer) int {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		writeErr(stderr, "creating settings directory", err)
		return 1
	}

	settings, err := hook.LoadSettings(path)
	if err != nil {
		writeErr(stderr, "loading settings", err)
		return 1
	}

	settings, added := hook.Install(settings)

	if err := hook.WriteSettings(path, settings); err != nil {
		writeErr(stderr, "writing settings", err)
		return 1
	}

	if len(added) == 0 {
		fmt.Fprintln(stdout, "aint: hooks already installed in", path)
	} else {
		fmt.Fprintln(stdout, "aint: installed into", path)
		for _, a := range added {
			fmt.Fprintln(stdout, "  +", a)
		}
	}
	return 0
}
```

- [ ] **Step 8: Wire `install` into dispatch**

```go
// cmd/aint/main.go — update dispatch's switch statement to add:
	case "install":
		return runInstall(args[2:])
```

- [ ] **Step 9: Run the tests to verify they pass**

Run: `go test ./... -v`
Expected: PASS across every package.

- [ ] **Step 10: Manual smoke test**

Run: `go run ./cmd/aint install` from a scratch temp directory (e.g. `cd $(mktemp -d) && go run /path/to/aint/cmd/aint install`)
Expected: creates `.claude/settings.json` with `hooks.PreToolUse`/`hooks.PostToolUse` entries; running it again prints "already installed".

- [ ] **Step 11: Commit**

```bash
git add internal/hook/install.go internal/hook/install_test.go cmd/aint/install.go cmd/aint/install_test.go cmd/aint/main.go
git commit -m "add aint install command to wire hooks into Claude Code settings.json"
```

---

### Task 13: Check docs pages and example config

**Files:**
- Create: `docs/checks/secret-hardcoded-key.md`
- Create: `docs/checks/secret-private-key-block.md`
- Create: `docs/checks/shell-gcp-role-wildcard.md`
- Create: `docs/checks/shell-chmod-permissive.md`
- Create: `docs/checks/go-ignored-error.md`
- Create: `docs/checks/swift-force-unwrap.md`
- Create: `docs/checks/python-shell-true.md`
- Create: `docs/checks/node-eval.md`
- Create: `.aint.yaml` (example, at repo root)
- Create: `README.md`

**Interfaces:**
- None — pure documentation, no code.

- [ ] **Step 1: Write each check's docs page**

Each page follows the same shape: what it flags, why it matters, how to fix
it, one safe vs. unsafe example. Example for one (repeat the same shape for
the other seven, substituting the relevant detection/fix/example):

```markdown
<!-- docs/checks/go-ignored-error.md -->
# go-ignored-error

**Flags:** an error return value discarded via `_ = err`.

**Why it matters:** silently discarding an error hides failures that should
either be handled or explicitly and visibly ignored with a comment
explaining why it's safe to do so.

**Fix:** handle the error, return it up the call stack, or log it. If
discarding really is correct (e.g. a `Close()` call whose error truly
doesn't matter here), leave a comment explaining why instead of a bare
discard.

```go
// Flags this:
_ = err

// Prefer this:
if err != nil {
    return fmt.Errorf("doing thing: %w", err)
}
```
```

Write the remaining seven pages (`secret-hardcoded-key.md`,
`secret-private-key-block.md`, `shell-gcp-role-wildcard.md`,
`shell-chmod-permissive.md`, `swift-force-unwrap.md`,
`python-shell-true.md`, `node-eval.md`) using the same four-section
structure, drawing the "flags"/"fix" content from each check's `Message`
and `Pattern` defined in Tasks 7-9.

- [ ] **Step 2: Write the example `.aint.yaml`**

```yaml
# .aint.yaml — example configuration for aint.
# Delete or edit freely; aint runs with sane defaults if this file is absent.

fail_on: error   # info|warning|error — findings at/above this exit non-zero
                 # and block/report in Claude Code hooks

ignore:
  - vendor/**
  - node_modules/**
  - "*.pb.go"

checks:
  {}
  # Example overrides:
  # go-ignored-error: error
  # node-eval: off

docs_base_url: ""  # set to a hosted docs URL once this repo has one;
                    # leave blank to use relative docs/checks/<id>.md links
```

- [ ] **Step 3: Write the README**

```markdown
<!-- README.md -->
# aint

Static analysis for code and shell commands, with human-readable findings
and a link to docs for each one.

## Usage

```
aint check [paths...]           # scan files/dirs (default: .)
aint check --format=json ...    # machine-readable output
aint list                       # list all registered checks
aint install [--global]         # wire aint into Claude Code hooks
```

## Claude Code integration

Run `aint install` from a project root to wire `aint` into
`.claude/settings.json`: shell commands are checked (and blocked on
error-level findings) before Claude runs them, and files Claude writes or
edits are checked afterward, with findings reported back as feedback.

Use `--global` to install into `~/.claude/settings.json` instead of the
current project.

## Configuration

Optional `.aint.yaml` at the repo root — see the example file in this repo
for the full set of options (`fail_on`, `ignore`, `checks`, `docs_base_url`).

## Checks

Run `aint list` for the full, current set. See `docs/checks/` for what each
one flags and how to fix it.
```

- [ ] **Step 4: Verify the docs build no errors and commit**

Run: `go build ./... && go test ./...`
Expected: PASS (docs changes don't affect Go compilation, but this confirms
nothing else broke).

```bash
git add docs/checks .aint.yaml README.md
git commit -m "add per-check docs pages, example .aint.yaml, and README"
```

---

## Plan Self-Review Notes

- **Spec coverage:** every spec section maps to a task — architecture (Tasks 1-2, 6, 10), CLI (Task 10), config (Task 4), seed checks across all 5 categories (Tasks 7-9), hook integration + install (Tasks 11-12), docs links (Task 13), testing approach (table tests throughout, end-to-end smoke tests in Tasks 10 and 11).
- **Type consistency verified:** `Check`, `Finding`, `Config`, `Target` fields and method signatures are used identically across every later task that consumes them (`check.Run(source, content, docsURL)`, `scan.Run(targets, checks, cfg)`, `config.SeverityFor(chk)`, `config.ResolveDocsURL(cfg, docsFile)`).
- **No placeholders:** every step has complete, runnable code — no "TBD" or "add validation" stand-ins.
