// internal/checks/sql/sql_test.go
package sql_test

import (
	"testing"

	"aint/internal/checks/sql"
)

func TestPowerSyncColumnAllowlistReminderDetectsAlterProjects(t *testing.T) {
	findings := sql.PowerSyncColumnAllowlistReminder.Run("migration.sql", []byte(`ALTER TABLE projects ADD COLUMN deleted_at timestamptz;`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestPowerSyncColumnAllowlistReminderDetectsAlterThreads(t *testing.T) {
	findings := sql.PowerSyncColumnAllowlistReminder.Run("migration.sql", []byte(`alter table threads add column archived boolean;`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestPowerSyncColumnAllowlistReminderIgnoresUnrelatedTable(t *testing.T) {
	findings := sql.PowerSyncColumnAllowlistReminder.Run("migration.sql", []byte(`ALTER TABLE audit_log ADD COLUMN note text;`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestPowerSyncColumnAllowlistReminderIgnoresDropColumn(t *testing.T) {
	findings := sql.PowerSyncColumnAllowlistReminder.Run("migration.sql", []byte(`ALTER TABLE projects DROP COLUMN legacy_flag;`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
