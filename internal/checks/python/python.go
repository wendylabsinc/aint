// internal/checks/python/python.go
package python

import (
	"regexp"

	"aint/internal/check"
)

var ShellTrue = check.Check{
	ID:       "python-shell-true",
	Title:    "subprocess call with shell=True",
	Severity: check.SeverityWarning,
	Langs:    []string{"python"},
	Pattern:  regexp.MustCompile(`subprocess\.\w+\([^)]*shell\s*=\s*True`),
	Message:  "shell=True risks shell injection if any part of the command is untrusted input",
	DocsPath: "python-shell-true.md",
}

func init() {
	check.Register(ShellTrue)
}
