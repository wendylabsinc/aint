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
	cats := []string{}
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
