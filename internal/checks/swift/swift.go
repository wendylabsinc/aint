// internal/checks/swift/swift.go
package swift

import (
	"regexp"

	"aint/internal/check"
)

var ForceUnwrap = check.Check{
	ID:       "swift-force-unwrap",
	Title:    "Force unwrap or force cast",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\b(try|as)!`),
	Message:  "force unwrap/cast can crash at runtime; prefer try?/as? with explicit handling",
	DocsPath: "swift-force-unwrap.md",
}

func init() {
	check.Register(ForceUnwrap)
}
