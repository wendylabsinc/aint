# sql-powersync-column-allowlist-reminder

**Flags:** `ALTER TABLE projects ADD COLUMN ...` / `ALTER TABLE threads ADD
COLUMN ...` - adding a column to a table that syncs through PowerSync.

**Why it matters:** a Postgres migration and a `schema.ts` update are not
enough on their own. The backend's write-permission handler filters every
incoming column through a per-table allow-list; a column missing from that
list is silently dropped from the write before the UPDATE/INSERT is built -
no error, no rejection. If the dropped column was the only field in that
write (e.g. a soft-delete's `UPDATE threads SET deleted_at = ?`), the whole
write becomes a server-side no-op: the local optimistic change looks right in
the UI, then the next sync-down pulls the untouched row back over it and the
change silently reverts.

**Fix:** before finishing a change that adds a column to a synced table,
checklist: (1) the Postgres migration, (2) the shared `schema.ts`, (3) the
backend's per-table column allow-list for that table, (4) every read query
that should honor the new column.

```sql
-- Flags this (and should also touch the backend allow-list + schema.ts):
ALTER TABLE projects ADD COLUMN deleted_at timestamptz;
```
