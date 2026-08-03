# aint improve Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `aint improve`, a command that mines Claude Code session transcripts under `~/.claude/projects` for user frustration/disagreement, and reports each as an incident with a suggested `aint`/linter rule and doc/memory fix.

**Architecture:** A new `internal/improve` package (transcript parsing → heuristic prefilter → context building → `claude -p` analysis → markdown report), each concern in its own file, wired together by a thin `cmd/aint/improve.go` following the existing `run<Cmd>WithIO` testable-CLI pattern used by `check.go`/`hook.go`/`install.go`.

**Tech Stack:** Go stdlib only (`encoding/json`, `bufio`, `regexp`, `os/exec`, `context`, `sort`) — no new `go.mod` dependencies. Module is `aint`, Go 1.26.1.

## Global Constraints

- No new dependencies in `go.mod` — stdlib only.
- Follow the existing `run<Cmd>` / `run<Cmd>WithIO(args, ...io.Writer, ...)` split so every command stays unit-testable without touching real stdin/stdout/subprocesses (see `check.go`, `hook.go`, `install.go`).
- Package-external test files (`package improve_test`, importing `aint/internal/improve`) for `internal/improve/*_test.go`, matching `golang_test.go`/`config_test.go` convention. `cmd/aint/improve_test.go` stays `package main` (same as `hook_test.go`), since it needs the unexported `runImproveWithIO`.
- No real subprocess execution or network access in any test — `analyze.go`'s `ClaudeRunner` interface must be fake-able.
- Every exported type/function needs a doc comment starting with its name (repo convention — see `internal/config`, `internal/check`).
- `aint check`/`list`/`hook`/`install` must remain fully offline; only `improve` shells out.

---

### Task 1: Transcript parsing

**Files:**
- Create: `internal/improve/transcript.go`
- Test: `internal/improve/transcript_test.go`

**Interfaces:**
- Produces: `type HumanMessage struct { SessionFile string; Line int; Timestamp time.Time; Project string; SessionID string; Text string }`, `type AssistantTurn struct { Line int; Timestamp time.Time; Rendered string }`, `func FindSessionFiles(dir string) ([]string, error)`, `func ParseSessionFile(path string, startLine int) (humans []HumanMessage, assistants []AssistantTurn, totalLines int, err error)`.

- [ ] **Step 1: Write the failing tests**

Create `internal/improve/transcript_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/improve/... -run TestParseSessionFile -v`
Expected: FAIL — `undefined: improve.ParseSessionFile` (package doesn't exist yet)

- [ ] **Step 3: Write the implementation**

Create `internal/improve/transcript.go`:

```go
// internal/improve/transcript.go
package improve

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// HumanMessage is a single human-typed chat line extracted from a Claude
// Code session transcript.
type HumanMessage struct {
	SessionFile string
	Line        int // 1-based line number within SessionFile
	Timestamp   time.Time
	Project     string // cwd recorded on the transcript line
	SessionID   string
	Text        string
}

// AssistantTurn is a single assistant-authored line extracted from a
// session transcript, with its text and tool_use blocks flattened into one
// summary string in original block order.
type AssistantTurn struct {
	Line      int
	Timestamp time.Time
	Rendered  string
}

type rawTranscriptLine struct {
	Type    string `json:"type"`
	Message struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
	Timestamp string `json:"timestamp"`
	SessionID string `json:"sessionId"`
	CWD       string `json:"cwd"`
	Origin    struct {
		Kind string `json:"kind"`
	} `json:"origin"`
}

type contentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// FindSessionFiles returns every *.jsonl file under dir, sorted
// lexicographically for deterministic processing order.
func FindSessionFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".jsonl") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

// ParseSessionFile reads path, skipping the first startLine lines (0 means
// read from the start), and returns every human-typed message and every
// assistant turn found from there on, plus the total number of lines in
// the file. Lines that fail to parse as JSON are skipped rather than
// causing an error, since a session file being actively appended to can
// have a truncated trailing line.
func ParseSessionFile(path string, startLine int) ([]HumanMessage, []AssistantTurn, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, 0, err
	}
	defer f.Close()

	var humans []HumanMessage
	var assistants []AssistantTurn
	lineNum := 0

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		lineNum++
		if lineNum <= startLine {
			continue
		}

		var raw rawTranscriptLine
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			continue
		}
		ts, _ := time.Parse(time.RFC3339, raw.Timestamp)

		switch raw.Type {
		case "user":
			var text string
			if err := json.Unmarshal(raw.Message.Content, &text); err != nil {
				continue // array content: a tool_result being echoed back, not human-typed
			}
			if raw.Origin.Kind != "human" {
				continue
			}
			humans = append(humans, HumanMessage{
				SessionFile: path,
				Line:        lineNum,
				Timestamp:   ts,
				Project:     raw.CWD,
				SessionID:   raw.SessionID,
				Text:        text,
			})
		case "assistant":
			var blocks []contentBlock
			if err := json.Unmarshal(raw.Message.Content, &blocks); err != nil {
				continue
			}
			rendered := renderBlocks(blocks)
			if rendered == "" {
				continue
			}
			assistants = append(assistants, AssistantTurn{
				Line:      lineNum,
				Timestamp: ts,
				Rendered:  rendered,
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return humans, assistants, lineNum, err
	}
	return humans, assistants, lineNum, nil
}

func renderBlocks(blocks []contentBlock) string {
	var parts []string
	for _, blk := range blocks {
		switch blk.Type {
		case "text":
			if t := strings.TrimSpace(blk.Text); t != "" {
				parts = append(parts, t)
			}
		case "tool_use":
			parts = append(parts, blk.Name+"("+summarizeInput(blk.Input)+")")
		}
	}
	return strings.Join(parts, "\n")
}

func summarizeInput(raw json.RawMessage) string {
	s := strings.Join(strings.Fields(string(raw)), " ")
	const max = 160
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/improve/... -v`
Expected: PASS (all three tests)

- [ ] **Step 5: Commit**

```bash
git add internal/improve/transcript.go internal/improve/transcript_test.go
git commit -m "$(cat <<'EOF'
Add session transcript parsing for aint improve

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01SjHVHdWrYF4U7fLbiZWX53
EOF
)"
```

---

### Task 2: Heuristic signal detection

**Files:**
- Create: `internal/improve/heuristics.go`
- Test: `internal/improve/heuristics_test.go`

**Interfaces:**
- Consumes: `HumanMessage` (Task 1).
- Produces: `func Detect(text string) []string`, `type Candidate struct { HumanMessage; Signals []string; Context string }`.

- [ ] **Step 1: Write the failing tests**

Create `internal/improve/heuristics_test.go`:

```go
package improve_test

import (
	"reflect"
	"sort"
	"testing"

	"aint/internal/improve"
)

func TestDetectSignalCategories(t *testing.T) {
	cases := []struct {
		text string
		want []string
	}{
		{"that's wrong, try again", []string{"correction"}},
		{"no, that's wrong too", []string{"correction"}},
		{"not what I asked for", []string{"correction"}},
		{"please undo that change", []string{"stop-undo"}},
		{"revert this now", []string{"stop-undo"}},
		{"why did you do that", []string{"stop-undo"}},
		{"ugh, seriously?", []string{"frustration-language"}},
		{"I already told you not to do that", []string{"frustration-language"}},
		{"WHY WOULD YOU DO THAT", []string{"shouting"}},
		{"stop!!", []string{"shouting"}},
		{"no.", []string{"terse-negative-reply"}},
		{"wrong.", []string{"terse-negative-reply"}},
		{"please add a test for this function", nil},
		{"This is a CLI tool", nil},
		{"no worries, that's totally fine, thanks!", nil},
	}

	for _, c := range cases {
		got := improve.Detect(c.text)
		sort.Strings(got)
		want := append([]string{}, c.want...)
		sort.Strings(want)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Detect(%q) = %v, want %v", c.text, got, want)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/improve/... -run TestDetectSignalCategories -v`
Expected: FAIL — `undefined: improve.Detect`

- [ ] **Step 3: Write the implementation**

Create `internal/improve/heuristics.go`:

```go
// internal/improve/heuristics.go
package improve

import (
	"regexp"
	"strings"
)

type signal struct {
	category string
	pattern  *regexp.Regexp
}

var signals = []signal{
	{"correction", regexp.MustCompile(`(?i)(that'?s (not|wrong)|not what i (asked|meant|wanted)|you (misunderstood|misread)|that'?s incorrect|^no+,)`)},
	{"stop-undo", regexp.MustCompile(`(?i)(stop doing that|don'?t do that|undo (that|this)|revert (this|that)|why did you)`)},
	{"frustration-language", regexp.MustCompile(`(?i)(ugh+\b|frustrat\w*|annoy\w*|seriously\?|come on\b|how many times|i already (said|told) you)`)},
	{"shouting", regexp.MustCompile(`!!+|\?\?+|\b[A-Z]{3,}(?:\s+[A-Z]{3,}){1,}\b`)},
}

var terseNegative = regexp.MustCompile(`(?i)^(no|nope|wrong)\.?!?$`)

// Detect returns the category names of every heuristic signal that matches
// text. A nil result means text isn't a candidate for analysis.
func Detect(text string) []string {
	var cats []string
	for _, s := range signals {
		if s.pattern.MatchString(text) {
			cats = append(cats, s.category)
		}
	}
	trimmed := strings.TrimSpace(text)
	if len(trimmed) < 15 && terseNegative.MatchString(trimmed) {
		cats = append(cats, "terse-negative-reply")
	}
	return cats
}

// Candidate is a human message flagged by Detect, carrying the matched
// signal categories and (once populated by BuildContext) the preceding
// conversation context to send to claude for analysis.
type Candidate struct {
	HumanMessage
	Signals []string
	Context string
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/improve/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/improve/heuristics.go internal/improve/heuristics_test.go
git commit -m "$(cat <<'EOF'
Add heuristic frustration/disagreement signal detection

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01SjHVHdWrYF4U7fLbiZWX53
EOF
)"
```

---

### Task 3: Context building

**Files:**
- Create: `internal/improve/context.go`
- Test: `internal/improve/context_test.go`

**Interfaces:**
- Consumes: `HumanMessage`, `AssistantTurn` (Task 1), `Candidate` (Task 2).
- Produces: `const MaxContextChars = 4000`, `func BuildContext(humans []HumanMessage, assistants []AssistantTurn, c Candidate) string`.

- [ ] **Step 1: Write the failing tests**

Create `internal/improve/context_test.go`:

```go
package improve_test

import (
	"strings"
	"testing"

	"aint/internal/improve"
)

func TestBuildContextIncludesOnlyTurnsSincePreviousHuman(t *testing.T) {
	humans := []improve.HumanMessage{
		{Line: 5},
		{Line: 10},
	}
	assistants := []improve.AssistantTurn{
		{Line: 6, Rendered: "a"},
		{Line: 8, Rendered: "b"},
		{Line: 12, Rendered: "c"},
	}
	candidate := improve.Candidate{HumanMessage: improve.HumanMessage{Line: 10}}

	got := improve.BuildContext(humans, assistants, candidate)
	if !strings.Contains(got, "a") || !strings.Contains(got, "b") {
		t.Errorf("expected context to include turns 6 and 8, got %q", got)
	}
	if strings.Contains(got, "c") {
		t.Errorf("expected context to exclude turn 12 (after candidate), got %q", got)
	}
}

func TestBuildContextWithNoPriorHumanMessage(t *testing.T) {
	humans := []improve.HumanMessage{{Line: 10}}
	assistants := []improve.AssistantTurn{{Line: 3, Rendered: "early turn"}}
	candidate := improve.Candidate{HumanMessage: improve.HumanMessage{Line: 10}}

	got := improve.BuildContext(humans, assistants, candidate)
	if !strings.Contains(got, "early turn") {
		t.Errorf("expected context to include turn before any human message, got %q", got)
	}
}

func TestBuildContextTruncatesAtMaxContextChars(t *testing.T) {
	long := strings.Repeat("x", improve.MaxContextChars+1000)
	assistants := []improve.AssistantTurn{{Line: 2, Rendered: long}}
	candidate := improve.Candidate{HumanMessage: improve.HumanMessage{Line: 5}}

	got := improve.BuildContext(nil, assistants, candidate)
	if !strings.HasSuffix(got, "... [truncated]") {
		t.Errorf("expected truncated context to end with marker, got suffix %q", got[len(got)-20:])
	}
	if len(got) > improve.MaxContextChars+len("... [truncated]")+20 {
		t.Errorf("truncated context too long: %d chars", len(got))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/improve/... -run TestBuildContext -v`
Expected: FAIL — `undefined: improve.BuildContext`

- [ ] **Step 3: Write the implementation**

Create `internal/improve/context.go`:

```go
// internal/improve/context.go
package improve

import (
	"fmt"
	"strings"
)

// MaxContextChars bounds how much preceding-turn context BuildContext will
// include, so the claude -p prompt stays a fixed size regardless of how
// verbose the surrounding conversation turn was.
const MaxContextChars = 4000

// BuildContext returns a bounded summary of assistant activity between the
// human message immediately preceding c (if any) and c itself, for use as
// analysis context passed to claude. humans and assistants need not be
// sorted; every entry is checked against c.Line.
func BuildContext(humans []HumanMessage, assistants []AssistantTurn, c Candidate) string {
	prevLine := 0
	for _, h := range humans {
		if h.Line < c.Line && h.Line > prevLine {
			prevLine = h.Line
		}
	}

	var b strings.Builder
	for _, a := range assistants {
		if a.Line <= prevLine || a.Line >= c.Line || a.Rendered == "" {
			continue
		}
		fmt.Fprintf(&b, "[assistant] %s\n", a.Rendered)
	}

	out := strings.TrimSpace(b.String())
	if len(out) > MaxContextChars {
		out = out[:MaxContextChars] + "... [truncated]"
	}
	return out
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/improve/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/improve/context.go internal/improve/context_test.go
git commit -m "$(cat <<'EOF'
Add preceding-turn context building for aint improve

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01SjHVHdWrYF4U7fLbiZWX53
EOF
)"
```

---

### Task 4: Offset cache (state)

**Files:**
- Create: `internal/improve/state.go`
- Test: `internal/improve/state_test.go`

**Interfaces:**
- Produces: `type State struct { Offsets map[string]int }`, `func LoadState(path string) (State, bool, error)`, `func SaveState(path string, state State) error`.

- [ ] **Step 1: Write the failing tests**

Create `internal/improve/state_test.go`:

```go
package improve_test

import (
	"os"
	"path/filepath"
	"testing"

	"aint/internal/improve"
)

func TestLoadStateMissingFileReturnsEmpty(t *testing.T) {
	state, corrupt, err := improve.LoadState(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if corrupt {
		t.Error("expected corrupt=false for a missing file")
	}
	if len(state.Offsets) != 0 {
		t.Errorf("expected empty offsets, got %v", state.Offsets)
	}
}

func TestLoadStateCorruptFileReturnsEmptyAndFlag(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := os.WriteFile(path, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	state, corrupt, err := improve.LoadState(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !corrupt {
		t.Error("expected corrupt=true for an unparsable file")
	}
	if len(state.Offsets) != 0 {
		t.Errorf("expected empty offsets, got %v", state.Offsets)
	}
}

func TestSaveAndLoadStateRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "state.json")
	want := improve.State{Offsets: map[string]int{"a.jsonl": 5, "b.jsonl": 12}}

	if err := improve.SaveState(path, want); err != nil {
		t.Fatalf("unexpected error saving: %v", err)
	}

	got, corrupt, err := improve.LoadState(path)
	if err != nil {
		t.Fatalf("unexpected error loading: %v", err)
	}
	if corrupt {
		t.Error("expected corrupt=false after a valid save")
	}
	if got.Offsets["a.jsonl"] != 5 || got.Offsets["b.jsonl"] != 12 {
		t.Errorf("unexpected offsets after round trip: %v", got.Offsets)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/improve/... -run TestLoadState -v`
Expected: FAIL — `undefined: improve.LoadState`

- [ ] **Step 3: Write the implementation**

Create `internal/improve/state.go`:

```go
// internal/improve/state.go
package improve

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// State is the offset cache: for each session file path, the last line
// number that has already been read and heuristic-checked. A file absent
// from Offsets has never been processed.
type State struct {
	Offsets map[string]int `json:"offsets"`
}

// LoadState reads path as JSON. A missing file returns an empty State,
// corrupt=false, err=nil. An unparsable file returns an empty State,
// corrupt=true, err=nil — callers should warn but keep going rather than
// fail the run over a cache-only file. Any other read error is returned as
// err.
func LoadState(path string) (State, bool, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return State{Offsets: map[string]int{}}, false, nil
	}
	if err != nil {
		return State{Offsets: map[string]int{}}, false, err
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return State{Offsets: map[string]int{}}, true, nil
	}
	if s.Offsets == nil {
		s.Offsets = map[string]int{}
	}
	return s, false, nil
}

// SaveState writes state to path as indented JSON, creating parent
// directories as needed.
func SaveState(path string, state State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/improve/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/improve/state.go internal/improve/state_test.go
git commit -m "$(cat <<'EOF'
Add offset cache for aint improve incremental scans

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01SjHVHdWrYF4U7fLbiZWX53
EOF
)"
```

---

### Task 5: Claude analysis

**Files:**
- Create: `internal/improve/analyze.go`
- Test: `internal/improve/analyze_test.go`

**Interfaces:**
- Consumes: `Candidate` (Task 2).
- Produces: `type ClaudeRunner interface { Run(ctx context.Context, prompt string) (string, error) }`, `func NewExecClaudeRunner(bin string) ClaudeRunner`, `type Incident struct { Candidate; Summary, RootCause, AintRuleSuggestion, LintRuleSuggestion, DocMemorySuggestion, AnalysisFailed string }`, `func Analyze(runner ClaudeRunner, c Candidate) (Incident, bool)`.

- [ ] **Step 1: Write the failing tests**

Create `internal/improve/analyze_test.go`:

```go
package improve_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"aint/internal/improve"
)

type fakeRunner struct {
	output      string
	err         error
	lastPrompt  string
}

func (f *fakeRunner) Run(ctx context.Context, prompt string) (string, error) {
	f.lastPrompt = prompt
	return f.output, f.err
}

func testCandidate() improve.Candidate {
	return improve.Candidate{
		HumanMessage: improve.HumanMessage{Text: "no that's wrong, revert that", Line: 10},
		Signals:      []string{"correction"},
		Context:      "[assistant] edited the wrong file",
	}
}

func TestAnalyzeConfirmedIncidentPopulatesFields(t *testing.T) {
	runner := &fakeRunner{output: `{
		"is_incident": true,
		"summary": "Claude edited the wrong file",
		"root_cause": "misread the file path",
		"aint_rule_suggestion": "flag edits outside the requested directory",
		"lint_rule_suggestion": null,
		"doc_memory_suggestion": "double-check file path before editing"
	}`}

	incident, include := improve.Analyze(runner, testCandidate())
	if !include {
		t.Fatal("expected include=true for a confirmed incident")
	}
	if incident.AnalysisFailed != "" {
		t.Errorf("expected no AnalysisFailed, got %q", incident.AnalysisFailed)
	}
	if incident.Summary != "Claude edited the wrong file" {
		t.Errorf("unexpected summary: %q", incident.Summary)
	}
	if incident.RootCause != "misread the file path" {
		t.Errorf("unexpected root cause: %q", incident.RootCause)
	}
	if incident.AintRuleSuggestion != "flag edits outside the requested directory" {
		t.Errorf("unexpected aint rule suggestion: %q", incident.AintRuleSuggestion)
	}
	if incident.LintRuleSuggestion != "" {
		t.Errorf("expected empty lint rule suggestion for null, got %q", incident.LintRuleSuggestion)
	}
	if incident.DocMemorySuggestion != "double-check file path before editing" {
		t.Errorf("unexpected doc/memory suggestion: %q", incident.DocMemorySuggestion)
	}

	if !strings.Contains(runner.lastPrompt, "no that's wrong, revert that") {
		t.Error("expected prompt to include the candidate text")
	}
	if !strings.Contains(runner.lastPrompt, "correction") {
		t.Error("expected prompt to include the matched signals")
	}
	if !strings.Contains(runner.lastPrompt, "edited the wrong file") {
		t.Error("expected prompt to include the candidate context")
	}
}

func TestAnalyzeNotAnIncidentIsDropped(t *testing.T) {
	runner := &fakeRunner{output: `{"is_incident": false, "summary": "", "root_cause": "", "aint_rule_suggestion": null, "lint_rule_suggestion": null, "doc_memory_suggestion": null}`}

	incident, include := improve.Analyze(runner, testCandidate())
	if include {
		t.Fatal("expected include=false when claude judges it not an incident")
	}
	if incident.Summary != "" || incident.AnalysisFailed != "" || len(incident.Signals) != 0 {
		t.Errorf("expected a zero-value incident, got %+v", incident)
	}
}

func TestAnalyzeParsesJSONWrappedInCommentary(t *testing.T) {
	runner := &fakeRunner{output: "Sure, here's my analysis:\n" + `{"is_incident": true, "summary": "s", "root_cause": "r", "aint_rule_suggestion": null, "lint_rule_suggestion": null, "doc_memory_suggestion": null}` + "\nHope that helps!"}

	incident, include := improve.Analyze(runner, testCandidate())
	if !include || incident.Summary != "s" || incident.RootCause != "r" {
		t.Errorf("expected parsed incident despite surrounding commentary, got include=%v incident=%+v", include, incident)
	}
}

func TestAnalyzeMalformedOutputSetsAnalysisFailed(t *testing.T) {
	runner := &fakeRunner{output: "not json at all"}

	incident, include := improve.Analyze(runner, testCandidate())
	if !include {
		t.Fatal("expected include=true so unanalyzed candidates still surface in the report")
	}
	if incident.AnalysisFailed == "" {
		t.Error("expected AnalysisFailed to be set for unparsable output")
	}
	if incident.Summary != "" {
		t.Errorf("expected empty summary on failure, got %q", incident.Summary)
	}
}

func TestAnalyzeRunnerErrorSetsAnalysisFailed(t *testing.T) {
	runner := &fakeRunner{err: errors.New("claude: command not found")}

	incident, include := improve.Analyze(runner, testCandidate())
	if !include {
		t.Fatal("expected include=true so a failed call still surfaces in the report")
	}
	if !strings.Contains(incident.AnalysisFailed, "command not found") {
		t.Errorf("expected AnalysisFailed to mention the runner error, got %q", incident.AnalysisFailed)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/improve/... -run TestAnalyze -v`
Expected: FAIL — `undefined: improve.Analyze`

- [ ] **Step 3: Write the implementation**

Create `internal/improve/analyze.go`:

```go
// internal/improve/analyze.go
package improve

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ClaudeRunner runs a single claude -p style prompt and returns its raw
// stdout. Tests supply a fake implementation; production code uses
// NewExecClaudeRunner.
type ClaudeRunner interface {
	Run(ctx context.Context, prompt string) (string, error)
}

type execClaudeRunner struct {
	bin string
}

// NewExecClaudeRunner returns a ClaudeRunner that shells out to bin
// (typically "claude") in print mode.
func NewExecClaudeRunner(bin string) ClaudeRunner {
	return execClaudeRunner{bin: bin}
}

func (r execClaudeRunner) Run(ctx context.Context, prompt string) (string, error) {
	cmd := exec.CommandContext(ctx, r.bin, "-p", prompt)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return "", err
	}
	return stdout.String(), nil
}

// Incident is a Candidate that claude confirmed as a genuine
// disagreement/frustration incident, or one whose analysis failed
// (AnalysisFailed set, all other fields empty) and is surfaced unanalyzed
// rather than silently dropped.
type Incident struct {
	Candidate
	Summary             string
	RootCause           string
	AintRuleSuggestion  string
	LintRuleSuggestion  string
	DocMemorySuggestion string
	AnalysisFailed      string
}

type claudeVerdict struct {
	IsIncident          bool    `json:"is_incident"`
	Summary             string  `json:"summary"`
	RootCause           string  `json:"root_cause"`
	AintRuleSuggestion  *string `json:"aint_rule_suggestion"`
	LintRuleSuggestion  *string `json:"lint_rule_suggestion"`
	DocMemorySuggestion *string `json:"doc_memory_suggestion"`
}

const analysisTimeout = 60 * time.Second

// Analyze asks runner to classify and analyze candidate c. It returns
// (incident, true) when the incident belongs in the report — either
// because claude confirmed it as real, or because the analysis attempt
// itself failed and c should be surfaced unanalyzed rather than dropped.
// It returns (Incident{}, false) when claude judged c not to be a genuine
// incident, in which case the caller should drop it silently.
func Analyze(runner ClaudeRunner, c Candidate) (Incident, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), analysisTimeout)
	defer cancel()

	out, err := runner.Run(ctx, buildPrompt(c))
	if err != nil {
		return Incident{Candidate: c, AnalysisFailed: err.Error()}, true
	}

	verdict, err := parseVerdict(out)
	if err != nil {
		return Incident{Candidate: c, AnalysisFailed: err.Error()}, true
	}

	if !verdict.IsIncident {
		return Incident{}, false
	}

	return Incident{
		Candidate:           c,
		Summary:             verdict.Summary,
		RootCause:           verdict.RootCause,
		AintRuleSuggestion:  derefOrEmpty(verdict.AintRuleSuggestion),
		LintRuleSuggestion:  derefOrEmpty(verdict.LintRuleSuggestion),
		DocMemorySuggestion: derefOrEmpty(verdict.DocMemorySuggestion),
	}, true
}

func buildPrompt(c Candidate) string {
	return fmt.Sprintf(`You are reviewing a Claude Code session transcript for an incident where the user showed disagreement or frustration with what Claude had just done.

User message (flagged by heuristic signals: %s):
%q

What Claude had done immediately before this message:
%s

Decide whether this is a genuine incident of the user disagreeing with or being frustrated by Claude's output (not just an unrelated negative word, or normal terse style). Respond with EXACTLY one JSON object and nothing else:

{
  "is_incident": true or false,
  "summary": "what happened, in a sentence or two",
  "root_cause": "why the user reacted negatively",
  "aint_rule_suggestion": "a concrete aint check idea (rough id, pattern, language, severity) that would have caught this mechanically, or null if not applicable",
  "lint_rule_suggestion": "a general linter rule for the relevant language if this is outside aint's coverage, or null if not applicable",
  "doc_memory_suggestion": "concrete text to add to CLAUDE.md or a memory/lessons-learned file to prevent this next time, or null if not applicable"
}

If is_incident is false, summary and root_cause should be empty strings and the three suggestion fields should be null.`, strings.Join(c.Signals, ", "), c.Text, c.Context)
}

func parseVerdict(output string) (claudeVerdict, error) {
	start := strings.IndexByte(output, '{')
	end := strings.LastIndexByte(output, '}')
	if start == -1 || end == -1 || end < start {
		return claudeVerdict{}, fmt.Errorf("no JSON object found in claude output: %q", output)
	}
	var v claudeVerdict
	if err := json.Unmarshal([]byte(output[start:end+1]), &v); err != nil {
		return claudeVerdict{}, fmt.Errorf("parsing claude JSON: %w", err)
	}
	return v, nil
}

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/improve/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/improve/analyze.go internal/improve/analyze_test.go
git commit -m "$(cat <<'EOF'
Add claude-backed incident analysis for aint improve

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01SjHVHdWrYF4U7fLbiZWX53
EOF
)"
```

---

### Task 6: Report rendering

**Files:**
- Create: `internal/improve/report.go`
- Test: `internal/improve/report_test.go`

**Interfaces:**
- Consumes: `Incident` (Task 5).
- Produces: `func WriteReport(w io.Writer, incidents []Incident, date string) error`.

- [ ] **Step 1: Write the failing tests**

Create `internal/improve/report_test.go`:

```go
package improve_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"aint/internal/improve"
)

func TestWriteReportEmptyIncidentsWritesNothing(t *testing.T) {
	var buf bytes.Buffer
	if err := improve.WriteReport(&buf, nil, "2026-08-03"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for zero incidents, got %q", buf.String())
	}
}

func TestWriteReportGroupsByProjectAndSession(t *testing.T) {
	ts := time.Date(2026, 8, 1, 3, 50, 0, 0, time.UTC)
	incidents := []improve.Incident{
		{
			Candidate: improve.Candidate{
				HumanMessage: improve.HumanMessage{Project: "/repo-a", SessionID: "sess-1", Line: 10, Timestamp: ts, Text: "that's wrong"},
				Signals:      []string{"correction"},
			},
			Summary:            "Claude edited the wrong file",
			RootCause:          "misread the path",
			AintRuleSuggestion: "flag edits outside the requested dir",
		},
		{
			Candidate: improve.Candidate{
				HumanMessage: improve.HumanMessage{Project: "/repo-b", SessionID: "sess-2", Line: 4, Timestamp: ts, Text: "stop doing that"},
				Signals:      []string{"stop-undo"},
			},
			AnalysisFailed: "claude: command not found",
		},
	}

	var buf bytes.Buffer
	if err := improve.WriteReport(&buf, incidents, "2026-08-03"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()

	for _, want := range []string{
		"aint improve report — 2026-08-03",
		"2 incidents across 2 sessions, 1 analysis failures",
		"/repo-a",
		"/repo-b",
		"sess-1",
		"sess-2",
		"correction",
		"that's wrong",
		"Claude edited the wrong file",
		"misread the path",
		"flag edits outside the requested dir",
		"⚠️ claude analysis unavailable: claude: command not found",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected report to contain %q, got:\n%s", want, out)
		}
	}

	if strings.Index(out, "/repo-a") > strings.Index(out, "/repo-b") {
		t.Error("expected /repo-a to be listed before /repo-b (sorted)")
	}
}

func TestWriteReportRendersNotApplicableForEmptySuggestions(t *testing.T) {
	incidents := []improve.Incident{
		{
			Candidate: improve.Candidate{
				HumanMessage: improve.HumanMessage{Project: "/repo", SessionID: "sess-1", Line: 1},
				Signals:      []string{"correction"},
			},
			Summary:   "s",
			RootCause: "r",
		},
	}

	var buf bytes.Buffer
	if err := improve.WriteReport(&buf, incidents, "2026-08-03"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Count(buf.String(), "Not applicable") != 3 {
		t.Errorf("expected 3 'Not applicable' fallbacks (aint/lint/doc), got:\n%s", buf.String())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/improve/... -run TestWriteReport -v`
Expected: FAIL — `undefined: improve.WriteReport`

- [ ] **Step 3: Write the implementation**

Create `internal/improve/report.go`:

```go
// internal/improve/report.go
package improve

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// WriteReport renders incidents as markdown to w, grouped by project then
// by session (chronological within a session). date is formatted into the
// report title (e.g. "2026-08-03"). Writing nothing for an empty incidents
// slice lets callers skip creating a report file entirely when there's
// nothing new.
func WriteReport(w io.Writer, incidents []Incident, date string) error {
	if len(incidents) == 0 {
		return nil
	}

	byProject := map[string][]Incident{}
	for _, inc := range incidents {
		byProject[inc.Project] = append(byProject[inc.Project], inc)
	}
	projects := make([]string, 0, len(byProject))
	for p := range byProject {
		projects = append(projects, p)
	}
	sort.Strings(projects)

	failed := 0
	sessions := map[string]bool{}
	for _, inc := range incidents {
		if inc.AnalysisFailed != "" {
			failed++
		}
		sessions[inc.SessionID] = true
	}

	if _, err := fmt.Fprintf(w, "# aint improve report — %s\n\n%d incidents across %d sessions, %d analysis failures.\n\n", date, len(incidents), len(sessions), failed); err != nil {
		return err
	}

	for _, project := range projects {
		if _, err := fmt.Fprintf(w, "## %s\n\n", project); err != nil {
			return err
		}
		if err := writeProjectSessions(w, byProject[project]); err != nil {
			return err
		}
	}
	return nil
}

func writeProjectSessions(w io.Writer, incidents []Incident) error {
	bySession := map[string][]Incident{}
	for _, inc := range incidents {
		bySession[inc.SessionID] = append(bySession[inc.SessionID], inc)
	}
	sessionIDs := make([]string, 0, len(bySession))
	for s := range bySession {
		sessionIDs = append(sessionIDs, s)
	}
	sort.Strings(sessionIDs)

	for _, sessionID := range sessionIDs {
		list := bySession[sessionID]
		sort.Slice(list, func(i, j int) bool { return list[i].Line < list[j].Line })

		if _, err := fmt.Fprintf(w, "### Session %s — %s\n\n", sessionID, list[0].Timestamp.Format("2006-01-02T15:04:05Z")); err != nil {
			return err
		}
		for _, inc := range list {
			if err := writeIncident(w, inc); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeIncident(w io.Writer, inc Incident) error {
	if _, err := fmt.Fprintf(w, "**Signals:** %s\n**User said:**\n> %q\n\n", joinOrNone(inc.Signals), inc.Text); err != nil {
		return err
	}
	if inc.AnalysisFailed != "" {
		_, err := fmt.Fprintf(w, "⚠️ claude analysis unavailable: %s\n\n---\n\n", inc.AnalysisFailed)
		return err
	}
	_, err := fmt.Fprintf(w,
		"**What happened:** %s\n**Root cause:** %s\n**Suggested aint rule:** %s\n**Suggested lint rule:** %s\n**Suggested doc/memory change:** %s\n\n---\n\n",
		inc.Summary, inc.RootCause,
		orNotApplicable(inc.AintRuleSuggestion),
		orNotApplicable(inc.LintRuleSuggestion),
		orNotApplicable(inc.DocMemorySuggestion))
	return err
}

func joinOrNone(signals []string) string {
	if len(signals) == 0 {
		return "none"
	}
	return strings.Join(signals, ", ")
}

func orNotApplicable(s string) string {
	if s == "" {
		return "Not applicable"
	}
	return s
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/improve/... -v`
Expected: PASS (every test in the package, including Tasks 1-5)

- [ ] **Step 5: Commit**

```bash
git add internal/improve/report.go internal/improve/report_test.go
git commit -m "$(cat <<'EOF'
Add markdown report rendering for aint improve

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01SjHVHdWrYF4U7fLbiZWX53
EOF
)"
```

---

### Task 7: CLI wiring

**Files:**
- Create: `cmd/aint/improve.go`
- Test: `cmd/aint/improve_test.go`
- Modify: `cmd/aint/main.go` (dispatch + usage)
- Modify: `README.md` (Commands section)

**Interfaces:**
- Consumes: everything from `internal/improve` (Tasks 1-6): `FindSessionFiles`, `ParseSessionFile`, `Detect`, `Candidate`, `BuildContext`, `ClaudeRunner`, `NewExecClaudeRunner`, `Analyze`, `Incident`, `LoadState`, `SaveState`, `State`, `WriteReport`.
- Produces: `func runImprove(args []string) int`, `func runImproveWithIO(args []string, stdout, stderr io.Writer, now time.Time, runner improve.ClaudeRunner) int`.

- [ ] **Step 1: Write the failing tests**

Create `cmd/aint/improve_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./cmd/aint/... -run TestRunImprove -v`
Expected: FAIL — `undefined: runImproveWithIO`

- [ ] **Step 3: Write the implementation**

Create `cmd/aint/improve.go`:

```go
// cmd/aint/improve.go
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"aint/internal/improve"
)

func runImprove(args []string) int {
	return runImproveWithIO(args, os.Stdout, os.Stderr, time.Now(), nil)
}

func runImproveWithIO(args []string, stdout, stderr io.Writer, now time.Time, runner improve.ClaudeRunner) int {
	home, err := os.UserHomeDir()
	if err != nil {
		writeErr(stderr, "resolving home directory", err)
		return 1
	}

	dir := filepath.Join(home, ".claude", "projects")
	statePath := filepath.Join(home, ".aint", "improve-state.json")
	outPath := ""
	claudeBin := "claude"
	limit := 50
	full := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--dir" && i+1 < len(args):
			i++
			dir = args[i]
		case strings.HasPrefix(arg, "--dir="):
			dir = strings.TrimPrefix(arg, "--dir=")
		case arg == "--out" && i+1 < len(args):
			i++
			outPath = args[i]
		case strings.HasPrefix(arg, "--out="):
			outPath = strings.TrimPrefix(arg, "--out=")
		case arg == "--state" && i+1 < len(args):
			i++
			statePath = args[i]
		case strings.HasPrefix(arg, "--state="):
			statePath = strings.TrimPrefix(arg, "--state=")
		case arg == "--claude-bin" && i+1 < len(args):
			i++
			claudeBin = args[i]
		case strings.HasPrefix(arg, "--claude-bin="):
			claudeBin = strings.TrimPrefix(arg, "--claude-bin=")
		case arg == "--limit" && i+1 < len(args):
			i++
			n, err := strconv.Atoi(args[i])
			if err != nil {
				writeErr(stderr, "parsing --limit", err)
				return 2
			}
			limit = n
		case strings.HasPrefix(arg, "--limit="):
			n, err := strconv.Atoi(strings.TrimPrefix(arg, "--limit="))
			if err != nil {
				writeErr(stderr, "parsing --limit", err)
				return 2
			}
			limit = n
		case arg == "--full":
			full = true
		default:
			writeErr(stderr, "unknown argument", fmt.Errorf("%s", arg))
			return 2
		}
	}

	if outPath == "" {
		outPath = filepath.Join(home, ".aint", "improve-reports", now.Format("2006-01-02")+".md")
	}

	state := improve.State{Offsets: map[string]int{}}
	if !full {
		loaded, corrupt, err := improve.LoadState(statePath)
		if err != nil {
			writeErr(stderr, "loading state", err)
			return 1
		}
		if corrupt {
			fmt.Fprintf(stderr, "aint: warning: state file %s was corrupt, starting fresh\n", statePath)
		}
		state = loaded
	}

	claudeUnavailable := ""
	if runner == nil {
		if _, err := exec.LookPath(claudeBin); err != nil {
			claudeUnavailable = fmt.Sprintf("claude CLI %q not found on PATH", claudeBin)
		} else {
			runner = improve.NewExecClaudeRunner(claudeBin)
		}
	}

	files, err := improve.FindSessionFiles(dir)
	if err != nil {
		writeErr(stderr, "scanning session directory", err)
		return 1
	}

	var incidents []improve.Incident
	callsUsed := 0

filesLoop:
	for _, file := range files {
		startLine := state.Offsets[file]
		humans, assistants, totalLines, err := improve.ParseSessionFile(file, startLine)
		if err != nil {
			writeErr(stderr, "parsing "+file, err)
			continue
		}

		newOffset := totalLines
		limitReached := false

		for _, h := range humans {
			signals := improve.Detect(h.Text)
			if len(signals) == 0 {
				continue
			}
			candidate := improve.Candidate{HumanMessage: h, Signals: signals}

			if claudeUnavailable != "" {
				incidents = append(incidents, improve.Incident{Candidate: candidate, AnalysisFailed: claudeUnavailable})
				continue
			}

			if callsUsed >= limit {
				newOffset = h.Line - 1
				limitReached = true
				break
			}

			candidate.Context = improve.BuildContext(humans, assistants, candidate)
			callsUsed++
			if incident, include := improve.Analyze(runner, candidate); include {
				incidents = append(incidents, incident)
			}
		}

		state.Offsets[file] = newOffset
		if err := improve.SaveState(statePath, state); err != nil {
			writeErr(stderr, "saving state", err)
			return 1
		}

		if limitReached {
			break filesLoop
		}
	}

	if len(incidents) == 0 {
		fmt.Fprintln(stdout, "No new incidents since last run.")
		return 0
	}

	failed := 0
	for _, inc := range incidents {
		if inc.AnalysisFailed != "" {
			failed++
		}
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		writeErr(stderr, "creating report directory", err)
		return 1
	}
	f, err := os.Create(outPath)
	if err != nil {
		writeErr(stderr, "creating report file", err)
		return 1
	}
	defer f.Close()

	if err := improve.WriteReport(f, incidents, now.Format("2006-01-02")); err != nil {
		writeErr(stderr, "writing report", err)
		return 1
	}

	fmt.Fprintf(stdout, "Found %d new incidents (%d analysis failures). Report: %s\n", len(incidents), failed, outPath)
	return 0
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./cmd/aint/... -v`
Expected: PASS for all `TestRunImprove*` tests. Some pre-existing `cmd/aint` tests may still be present from earlier tasks — all should pass.

- [ ] **Step 5: Wire into `main.go`**

Modify `cmd/aint/main.go`:

```go
	switch args[1] {
	case "check":
		return runCheck(args[2:])
	case "list":
		return runList(args[2:])
	case "hook":
		return runHook(args[2:])
	case "install":
		return runInstall(args[2:])
	case "improve":
		return runImprove(args[2:])
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
  improve                     mine Claude Code session history for incidents
  hook <pre-bash|post-edit>   internal: used by installed hooks`)
}
```

- [ ] **Step 6: Update `README.md`**

In the `## Commands` section at the bottom of `README.md`, add a line after `aint install`:

```
aint improve                    # mine ~/.claude/projects for incidents, report suggested aint/lint rules + doc fixes
```

- [ ] **Step 7: Run the full test suite**

Run: `go build ./... && go test ./...`
Expected: builds cleanly, all tests PASS (existing + new `internal/improve` and `cmd/aint` tests)

- [ ] **Step 8: Commit**

```bash
git add cmd/aint/improve.go cmd/aint/improve_test.go cmd/aint/main.go README.md
git commit -m "$(cat <<'EOF'
Wire up aint improve command

Mines ~/.claude/projects for user frustration/disagreement, analyzes
each candidate via the claude CLI, and reports incidents with
suggested aint/lint rules and doc/memory fixes.

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01SjHVHdWrYF4U7fLbiZWX53
EOF
)"
```
