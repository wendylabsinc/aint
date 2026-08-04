# node-pg-client-raw-connect

**Flags:** `new pg.Client(...)` / `new Client(...)` - a raw node-postgres
client instantiation.

**Why it matters:** node-postgres emits an `'error'` event on the client
instance itself on any abrupt disconnect (for example, a single
self-permitted `SELECT pg_terminate_backend(pg_backend_pid())`). With zero
listeners attached, Node treats that as an uncaught exception and kills the
**whole process**, not just the one query.

**Fix:** attach `client.on('error', () => {})` (or real handling) before
calling `.connect()`, every time a raw `pg.Client` is used for a per-query
connection. If you don't need per-connection lifecycle control, prefer a
`pg.Pool`, which handles this internally.

```js
// Flags this:
const client = new pg.Client(config)
await client.connect()

// Prefer this:
const client = new pg.Client(config)
client.on('error', (err) => log.warn('pg client error', err))
await client.connect()
```
