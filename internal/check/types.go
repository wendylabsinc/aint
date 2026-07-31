// internal/check/types.go
package check

import (
	"regexp"
	"strings"
)

type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

var severityRank = map[Severity]int{
	SeverityInfo:    0,
	SeverityWarning: 1,
	SeverityError:   2,
}

// AtLeast reports whether s is at least as severe as min.
func (s Severity) AtLeast(min Severity) bool {
	return severityRank[s] >= severityRank[min]
}

// Check is a single regex-based rule. Langs is a list of language tags
// (e.g. "go", "swift", "shell") the check applies to; an empty/nil Langs
// means the check applies to every language, including the synthetic
// "shell" pseudo-file used for raw shell command strings.
type Check struct {
	ID       string
	Title    string
	Severity Severity
	Langs    []string
	Pattern  *regexp.Regexp
	Message  string
	DocsPath string
}

// Finding is a single reported match.
type Finding struct {
	CheckID  string
	Severity Severity
	Source   string
	Line     int
	Column   int
	Message  string
	DocsURL  string
}

// Run scans content line by line for the check's pattern, returning one
// Finding per matching line. docsURL is passed in by the caller (resolved
// from config) rather than read from c.DocsPath directly.
func (c Check) Run(source string, content []byte, docsURL string) []Finding {
	var findings []Finding
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		loc := c.Pattern.FindStringIndex(line)
		if loc == nil {
			continue
		}
		findings = append(findings, Finding{
			CheckID:  c.ID,
			Severity: c.Severity,
			Source:   source,
			Line:     i + 1,
			Column:   loc[0] + 1,
			Message:  c.Message,
			DocsURL:  docsURL,
		})
	}
	return findings
}
