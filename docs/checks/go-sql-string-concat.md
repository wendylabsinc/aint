# go-sql-string-concat

**Flags:** a call to `.Query(`, `.Exec(`, or `.QueryRow(` (or their
variants, e.g. `QueryContext`) whose arguments contain `fmt.Sprintf` or
string concatenation (`+`) with a quoted string.

**Why it matters:** building a SQL query by formatting or concatenating
values directly into the query string means any value that traces back to
external input can inject SQL — altering the query's logic, exfiltrating
data, or modifying data outside the intended scope. Parameterized queries
sidestep this entirely by keeping data separate from the query structure.

**Fix:** use placeholder parameters (`?`, `$1`, depending on the driver)
and pass values as separate arguments to `Query`/`Exec`, letting the driver
handle safe escaping.

```go
// Flags this:
db.Query(fmt.Sprintf("SELECT * FROM users WHERE id = %s", id))

// Prefer this:
db.Query("SELECT * FROM users WHERE id = $1", id)
```
