# go-ignored-error

**Flags:** an error return value discarded via `_ = err`.

**Why it matters:** silently discarding an error hides failures that should either be handled or explicitly and visibly ignored with a comment explaining why it's safe to do so.

**Fix:** handle the error, return it up the call stack, or log it. If discarding really is correct (e.g. a `Close()` call whose error truly doesn't matter here), leave a comment explaining why instead of a bare discard.

```go
// Flags this:
_ = err

// Prefer this:
if err != nil {
    return fmt.Errorf("doing thing: %w", err)
}

// Or, if the error is truly ignorable:
_ = file.Close()  // error is benign; file will be reclaimed by GC
```
