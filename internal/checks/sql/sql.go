// internal/checks/sql/sql.go
package sql

import (
	"regexp"

	"aint/internal/check"
)

var PowerSyncColumnAllowlistReminder = check.Check{
	ID:       "sql-powersync-column-allowlist-reminder",
	Title:    "New column on a PowerSync-synced table",
	Severity: check.SeverityWarning,
	Langs:    []string{"sql"},
	Pattern:  regexp.MustCompile(`(?i)ALTER\s+TABLE\s+(projects|threads)\b[^\n]*ADD\s+COLUMN`),
	Message:  "a new column on a PowerSync-synced table also needs frontend/packages/shared/src/schema.ts AND backend/server.mjs's TABLES allow-list updated, or writes to it silently no-op",
	DocsPath: "sql-powersync-column-allowlist-reminder.md",
}

func init() {
	check.Register(PowerSyncColumnAllowlistReminder)
}
