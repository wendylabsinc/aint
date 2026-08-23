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

// Check is a single regex-based or syntax-aware rule. Langs is a list of language tags
// (e.g. "go", "swift", "shell") the check applies to; an empty/nil Langs
// means the check applies to every language, including the synthetic
// "shell" pseudo-file used for raw shell command strings.
type Check struct {
	ID       string
	Title    string
	Severity Severity
	Langs    []string
	Pattern  *regexp.Regexp
	Analyze  func(source string, content []byte) []Match
	Message  string
	DocsPath string
}

// Match is a syntax-aware check's source location and optional message. When
// Message is empty, Check.Message is used. Analyze callbacks return Matches;
// Run fills in the check metadata and resolved documentation URL.
type Match struct {
	Line    int
	Column  int
	Message string
}

// Finding is a single reported match.
type Finding struct {
	CheckID  string   `json:"check_id"`
	Severity Severity `json:"severity"`
	Source   string   `json:"source"`
	Line     int      `json:"line"`
	Column   int      `json:"column"`
	Message  string   `json:"message"`
	DocsURL  string   `json:"docs_url"`
}

// Run analyzes content with Analyze when present, otherwise scans it line by
// line with Pattern. docsURL is passed in by the caller (resolved from config)
// rather than read from c.DocsPath directly.
func (c Check) Run(source string, content []byte, docsURL string) []Finding {
	if c.Analyze != nil {
		matches := c.Analyze(source, content)
		findings := make([]Finding, 0, len(matches))
		for _, match := range matches {
			message := match.Message
			if message == "" {
				message = c.Message
			}
			findings = append(findings, Finding{
				CheckID:  c.ID,
				Severity: c.Severity,
				Source:   source,
				Line:     match.Line,
				Column:   match.Column,
				Message:  message,
				DocsURL:  docsURL,
			})
		}
		return findings
	}

	if c.Pattern == nil {
		return nil
	}

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
