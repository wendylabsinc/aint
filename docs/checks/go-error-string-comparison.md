# go-error-string-comparison

**Flags:** comparing `err.Error()` with a string or searching it with common `strings` functions.

**Why it matters:** error messages are for people. Wording changes, wrapping, and added context can silently break control flow based on message text.

**Fix:** expose a sentinel or typed error and inspect it with `errors.Is` or `errors.As`.

```go
// Flags this:
if strings.Contains(err.Error(), "not found") { ... }

// Prefer this:
if errors.Is(err, fs.ErrNotExist) { ... }
```
