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
