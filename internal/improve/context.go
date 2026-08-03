// internal/improve/context.go
package improve

import (
	"fmt"
	"strings"
)

// MaxContextChars bounds how much preceding-turn context BuildContext will
// include, so the claude -p prompt stays a fixed size regardless of how
// verbose the surrounding conversation turn was.
const MaxContextChars = 4000

// BuildContext returns a bounded summary of assistant activity between the
// human message immediately preceding c (if any) and c itself, for use as
// analysis context passed to claude. humans and assistants need not be
// sorted; every entry is checked against c.Line.
func BuildContext(humans []HumanMessage, assistants []AssistantTurn, c Candidate) string {
	prevLine := 0
	for _, h := range humans {
		if h.Line < c.Line && h.Line > prevLine {
			prevLine = h.Line
		}
	}

	var b strings.Builder
	for _, a := range assistants {
		if a.Line <= prevLine || a.Line >= c.Line || a.Rendered == "" {
			continue
		}
		fmt.Fprintf(&b, "[assistant] %s\n", a.Rendered)
	}

	out := strings.TrimSpace(b.String())
	if len(out) > MaxContextChars {
		out = out[:MaxContextChars] + "... [truncated]"
	}
	return out
}
