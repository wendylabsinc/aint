// internal/checks/nodejs/nodejs.go
package nodejs

import (
	"regexp"

	"aint/internal/check"
)

var Eval = check.Check{
	ID:       "node-eval",
	Title:    "Use of eval()",
	Severity: check.SeverityWarning,
	Langs:    []string{"nodejs"},
	Pattern:  regexp.MustCompile(`\beval\s*\(`),
	Message:  "eval() executes arbitrary strings as code; avoid it or use a safer parser",
	DocsPath: "node-eval.md",
}

func init() {
	check.Register(Eval)
}
