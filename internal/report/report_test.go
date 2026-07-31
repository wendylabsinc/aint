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
