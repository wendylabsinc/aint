package improve_test

import (
	"os"
	"path/filepath"
	"testing"

	"aint/internal/improve"
)

func writeLines(t *testing.T, path string, lines []string) {
	t.Helper()
	content := ""
	for _, l := range lines {
		content += l + "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func fixtureLines() []string {
	return []string{
		`{"type":"user","message":{"role":"user","content":"no that's wrong, revert that"},"timestamp":"2026-08-01T03:50:01.199Z","sessionId":"sess-1","cwd":"/repo","origin":{"kind":"human"}}`,
		`{"type":"user","message":{"role":"user","content":[{"type":"tool_result","content":"ok","tool_use_id":"t1"}]},"timestamp":"2026-08-01T03:50:05.000Z","sessionId":"sess-1","cwd":"/repo"}`,
		`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Editing the file now."},{"type":"tool_use","id":"t2","name":"Bash","input":{"command":"rm -rf /tmp/x"}}]},"timestamp":"2026-08-01T03:50:10.000Z","sessionId":"sess-1","cwd":"/repo"}`,
		`{"type":"system","timestamp":"2026-08-01T03:50:11.000Z"}`,
		`{not valid json`,
	}
}

func TestParseSessionFileExtractsHumanAndAssistantTurns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl")
	writeLines(t, path, fixtureLines())

	humans, assistants, totalLines, err := improve.ParseSessionFile(path, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if totalLines != 5 {
		t.Errorf("expected 5 total lines, got %d", totalLines)
	}

	if len(humans) != 1 {
		t.Fatalf("expected 1 human message, got %d: %+v", len(humans), humans)
	}
	h := humans[0]
	if h.Line != 1 || h.Text != "no that's wrong, revert that" || h.Project != "/repo" || h.SessionID != "sess-1" {
		t.Errorf("unexpected human message: %+v", h)
	}

	if len(assistants) != 1 {
		t.Fatalf("expected 1 assistant turn, got %d: %+v", len(assistants), assistants)
	}
	a := assistants[0]
	if a.Line != 3 {
		t.Errorf("expected assistant turn on line 3, got %d", a.Line)
	}
	if !contains(a.Rendered, "Editing the file now.") || !contains(a.Rendered, "Bash(") || !contains(a.Rendered, "rm -rf /tmp/x") {
		t.Errorf("unexpected rendered assistant turn: %q", a.Rendered)
	}
}

func TestParseSessionFileResumesFromStartLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl")
	writeLines(t, path, fixtureLines())

	humans, assistants, totalLines, err := improve.ParseSessionFile(path, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if totalLines != 5 {
		t.Errorf("expected 5 total lines, got %d", totalLines)
	}
	if len(humans) != 0 {
		t.Errorf("expected 0 human messages after resuming past line 3, got %d", len(humans))
	}
	if len(assistants) != 0 {
		t.Errorf("expected 0 assistant turns after resuming past line 3, got %d", len(assistants))
	}
}

func TestFindSessionFilesFindsJSONLRecursivelyAndSorts(t *testing.T) {
	dir := t.TempDir()
	mustWriteEmpty(t, filepath.Join(dir, "b", "2.jsonl"))
	mustWriteEmpty(t, filepath.Join(dir, "a", "1.jsonl"))
	mustWriteEmpty(t, filepath.Join(dir, "3.jsonl"))
	mustWriteEmpty(t, filepath.Join(dir, "readme.txt"))

	files, err := improve.FindSessionFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 3 {
		t.Fatalf("expected 3 .jsonl files, got %d: %v", len(files), files)
	}
	for i := 1; i < len(files); i++ {
		if files[i-1] > files[i] {
			t.Errorf("expected sorted output, got %v", files)
		}
	}
}

func mustWriteEmpty(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (func() bool {
		for i := 0; i+len(substr) <= len(s); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	})()
}
