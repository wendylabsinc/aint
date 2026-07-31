// internal/checks/shell/shell.go
package shell

import (
	"regexp"

	"aint/internal/check"
)

var GCPRoleWildcard = check.Check{
	ID:       "shell-gcp-role-wildcard",
	Title:    "Overscoped GCP IAM role grant",
	Severity: check.SeverityError,
	Langs:    []string{"shell"},
	Pattern:  regexp.MustCompile(`add-iam-policy-binding.*--role[= ]roles/(owner|editor)\b`),
	Message:  "granting roles/owner or roles/editor is overscoped; use a narrower predefined or custom role",
	DocsPath: "shell-gcp-role-wildcard.md",
}

var ChmodPermissive = check.Check{
	ID:       "shell-chmod-permissive",
	Title:    "World-writable chmod",
	Severity: check.SeverityWarning,
	Langs:    []string{"shell"},
	Pattern:  regexp.MustCompile(`chmod\s+(-R\s+)?(777|a\+rwx|ugo\+rwx)\b`),
	Message:  "world-writable permissions granted via chmod",
	DocsPath: "shell-chmod-permissive.md",
}

func init() {
	check.Register(GCPRoleWildcard)
	check.Register(ChmodPermissive)
}
