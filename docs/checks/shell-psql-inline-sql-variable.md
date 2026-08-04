# shell-psql-inline-sql-variable

**Flags:** `psql -c "<...>$VAR<...>"` - a `psql -c` invocation whose SQL text
contains a shell-interpolated variable.

**Why it matters:** any text-based defense built around a `psql -c` string
concatenated with untrusted/dynamic SQL is defeated by comment/quote
smuggling. `SET default_transaction_read_only = on` inside the same `-c`
string doesn't apply retroactively; wrapping in `BEGIN READ ONLY; ...;
COMMIT;` is defeated by SQL containing its own `COMMIT;`/`SET ...;` (psql
treats embedded `BEGIN`/`COMMIT` as dividing the string into multiple
transactions); and a regex pre-filter rejecting `SET`/`COMMIT`/multi-statement
text is defeated by a quote hidden inside a `/* ... */` or `-- ...` comment,
which blanks the lexer's view of everything up to the next real quote.

**Fix:** enforce any guarantee (read-only, single-statement, etc.) at the
protocol level with a real client library instead of string concatenation.
For Postgres specifically: a `pg.Client` running the query via the extended
query protocol (`queryMode: 'extended'`) makes Postgres itself reject
multi-statement payloads structurally - there's no string concatenation left
for a comment/quote to hide a second statement in.

```bash
# Flags this:
psql -c "BEGIN READ ONLY; $SQL; COMMIT;"

# Prefer this (Node.js example):
# client.query({ text: sql, queryMode: 'extended' })
```
