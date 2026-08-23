# go-error-direct-comparison

**Flags:** equality or inequality comparisons between an error-like value and a non-`nil` value.

**Why it matters:** direct comparison fails when an error has been wrapped with `%w`. Callers then handle the same underlying error differently depending on whether an intermediate layer added context.

**Fix:** use `errors.Is` for sentinel errors. Disable or override this warning locally if exact identity is deliberately part of the design.

```go
// Flags this:
if err == os.ErrNotExist { ... }

// Prefer this:
if errors.Is(err, os.ErrNotExist) { ... }
```
