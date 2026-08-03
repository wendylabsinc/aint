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
