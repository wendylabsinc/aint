package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aint/internal/improve"
)

type stubRunner struct {
	calls int
}

func (s *stubRunner) Run(ctx context.Context, prompt string) (string, error) {
	s.calls++
	return `{"is_incident": true, "summary": "Claude edited the wrong file", "root_cause": "misread the path", "aint_rule_suggestion": null, "lint_rule_suggestion": null, "doc_memory_suggestion": null}`, nil
}

func writeSessionFile(t *testing.T, dir, name string, lines []string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func humanLine(text string) string {
	return `{"type":"user","message":{"role":"user","content":"` + text + `"},"timestamp":"2026-08-01T03:50:01.199Z","sessionId":"sess-1","cwd":"/repo","origin":{"kind":"human"}}`
}

func assistantLine() string {
	return `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"working on it"}]},"timestamp":"2026-08-01T03:50:05.000Z"}`
}

func TestRunImproveColdRunFindsAndReportsIncident(t *testing.T) {
	sessionDir := t.TempDir()
	writeSessionFile(t, sessionDir, "session.jsonl", []string{
		assistantLine(),
		humanLine(`that's wrong, try again`),
	})

	statePath := filepath.Join(t.TempDir(), "state.json")
	outPath := filepath.Join(t.TempDir(), "report.md")
	runner := &stubRunner{}

	var stdout, stderr bytes.Buffer
	code := runImproveWithIO(
		[]string{"--dir", sessionDir, "--state", statePath, "--out", outPath, "--limit", "10"},
		&stdout, &stderr, time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC), runner)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", code, stderr.String())
	}
	if runner.calls != 1 {
		t.Errorf("expected 1 claude call, got %d", runner.calls)
	}
	if !strings.Contains(stdout.String(), "Found 1 new incidents") {
		t.Errorf("expected stdout to report 1 incident, got %q", stdout.String())
	}

	report, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("expected report file to exist: %v", err)
	}
	if !strings.Contains(string(report), "Claude edited the wrong file") {
		t.Errorf("expected report to contain the incident summary, got:\n%s", report)
	}

	state, _, err := improve.LoadState(statePath)
	if err != nil {
		t.Fatalf("unexpected error loading state: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, "session.jsonl")
	if state.Offsets[sessionFile] != 2 {
		t.Errorf("expected offset 2 (both lines processed), got %d", state.Offsets[sessionFile])
	}
}

func TestRunImproveSecondRunFindsNothingNew(t *testing.T) {
	sessionDir := t.TempDir()
	writeSessionFile(t, sessionDir, "session.jsonl", []string{
		humanLine(`that's wrong, try again`),
	})
	statePath := filepath.Join(t.TempDir(), "state.json")
	runner := &stubRunner{}
	now := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)

	firstOut := filepath.Join(t.TempDir(), "report1.md")
	code := runImproveWithIO([]string{"--dir", sessionDir, "--state", statePath, "--out", firstOut, "--limit", "10"}, &bytes.Buffer{}, &bytes.Buffer{}, now, runner)
	if code != 0 {
		t.Fatalf("expected first run to exit 0, got %d", code)
	}

	secondOut := filepath.Join(t.TempDir(), "report2.md")
	var stdout bytes.Buffer
	code = runImproveWithIO([]string{"--dir", sessionDir, "--state", statePath, "--out", secondOut, "--limit", "10"}, &stdout, &bytes.Buffer{}, now, runner)
	if code != 0 {
		t.Fatalf("expected second run to exit 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "No new incidents") {
		t.Errorf("expected second run to report nothing new, got %q", stdout.String())
	}
	if _, err := os.Stat(secondOut); err == nil {
		t.Error("expected no report file to be written on a run with nothing new")
	}
	if runner.calls != 1 {
		t.Errorf("expected claude to be called only once total (not re-invoked on the cached candidate), got %d", runner.calls)
	}
}

func TestRunImproveLimitStopsPartwayAndResumesNextRun(t *testing.T) {
	sessionDir := t.TempDir()
	writeSessionFile(t, sessionDir, "session.jsonl", []string{
		humanLine(`that's wrong, try again`),
		assistantLine(),
		humanLine(`no, that's wrong too`),
	})
	statePath := filepath.Join(t.TempDir(), "state.json")
	runner := &stubRunner{}
	now := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)

	out1 := filepath.Join(t.TempDir(), "report1.md")
	code := runImproveWithIO([]string{"--dir", sessionDir, "--state", statePath, "--out", out1, "--limit", "1"}, &bytes.Buffer{}, &bytes.Buffer{}, now, runner)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if runner.calls != 1 {
		t.Fatalf("expected exactly 1 claude call with --limit 1, got %d", runner.calls)
	}

	out2 := filepath.Join(t.TempDir(), "report2.md")
	code = runImproveWithIO([]string{"--dir", sessionDir, "--state", statePath, "--out", out2, "--limit", "10"}, &bytes.Buffer{}, &bytes.Buffer{}, now, runner)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if runner.calls != 2 {
		t.Errorf("expected the second candidate to be picked up on the next run, total calls = %d", runner.calls)
	}
}

func TestRunImproveClaudeUnavailableMarksAnalysisFailed(t *testing.T) {
	sessionDir := t.TempDir()
	writeSessionFile(t, sessionDir, "session.jsonl", []string{
		humanLine(`that's wrong, try again`),
	})
	statePath := filepath.Join(t.TempDir(), "state.json")
	outPath := filepath.Join(t.TempDir(), "report.md")

	var stdout, stderr bytes.Buffer
	code := runImproveWithIO(
		[]string{"--dir", sessionDir, "--state", statePath, "--out", outPath, "--claude-bin", "definitely-not-a-real-claude-binary"},
		&stdout, &stderr, time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC), nil)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", code, stderr.String())
	}
	report, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("expected report file to exist: %v", err)
	}
	if !strings.Contains(string(report), "claude analysis unavailable") {
		t.Errorf("expected report to flag claude as unavailable, got:\n%s", report)
	}
}
