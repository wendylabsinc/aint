// internal/checks/yaml/yaml.go
package yaml

import (
	"regexp"

	"aint/internal/check"
)

var ConditionalRunnerSizing = check.Check{
	ID:       "yaml-ci-conditional-runner-size",
	Title:    "Conditional CI runner sizing",
	Severity: check.SeverityWarning,
	Langs:    []string{"yaml"},
	Pattern:  regexp.MustCompile(`runs-on:\s*\$\{\{[^\n}]*&&[^\n}]*\|\|[^\n}]*\}\}`),
	Message:  "conditional runner sizing (ternary on event type) was explicitly rejected before - use one larger instance unconditionally for all event types",
	DocsPath: "yaml-ci-conditional-runner-size.md",
}

func init() {
	check.Register(ConditionalRunnerSizing)
}
