// internal/checks/golang/golang.go
package golang

import (
	"regexp"

	"aint/internal/check"
)

var IgnoredError = check.Check{
	ID:       "go-ignored-error",
	Title:    "Ignored error return value",
	Severity: check.SeverityWarning,
	Langs:    []string{"go"},
	Pattern:  regexp.MustCompile(`\b_\s*=\s*err\b`),
	Message:  "error value discarded via `_ = err`",
	DocsPath: "go-ignored-error.md",
}

func init() {
	check.Register(IgnoredError)
}
