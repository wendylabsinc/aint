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
