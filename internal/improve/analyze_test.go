package improve_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"aint/internal/improve"
)

type fakeRunner struct {
	output     string
	err        error
	lastPrompt string
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

func TestAnalyzeTruncatesVeryLongCandidateText(t *testing.T) {
	runner := &fakeRunner{output: `{"is_incident": false, "summary": "", "root_cause": "", "aint_rule_suggestion": null, "lint_rule_suggestion": null, "doc_memory_suggestion": null}`}

	longText := strings.Repeat("a", 20000)
	c := testCandidate()
	c.Text = longText

	improve.Analyze(runner, c)

	if !strings.Contains(runner.lastPrompt, "... [truncated]") {
		t.Error("expected prompt to contain a truncation marker for very long candidate text")
	}
	if strings.Contains(runner.lastPrompt, longText) {
		t.Error("expected prompt to NOT contain the full untruncated candidate text")
	}
}
