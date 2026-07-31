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
