# go-panic-with-error

**Flags:** `panic(err)`, `panic(fmt.Errorf(...))`, and `panic(errors.New(...))`.

**Why it matters:** expected failures such as invalid input, missing files, and network errors should remain recoverable. Panicking moves that decision to a distant recovery boundary and may crash the process.

**Fix:** return or handle the error. Panics for genuinely broken invariants can be allowed by disabling or overriding this warning for the relevant project.

```go
// Flags this:
if err != nil { panic(err) }

// Prefer this:
if err != nil { return fmt.Errorf("load config: %w", err) }
```
