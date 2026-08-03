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
