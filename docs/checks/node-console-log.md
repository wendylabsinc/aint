# node-console-log

**Flags:** a call to `console.log(`.

**Why it matters:** `console.log` writes unstructured text straight to
stdout with no log level, no routing, and no way to filter or redact it in
production. Debug statements added during development are easy to forget
about, and leftover `console.log` calls can end up dumping sensitive
values (tokens, request bodies, internal state) into logs that weren't
meant to capture them.

**Fix:** use a structured logger (e.g. `pino`, `winston`) with an
appropriate log level instead, so output can be filtered and routed
consistently in production.

```javascript
// Flags this:
console.log(debugValue);

// Prefer this:
logger.info(debugValue);
```
