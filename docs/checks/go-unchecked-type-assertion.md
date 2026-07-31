# go-unchecked-type-assertion

**Flags:** a single-value type assertion of the form `x := y.(SomeType)` on
its own line, without the two-value "comma-ok" form.

**Why it matters:** a single-value type assertion panics immediately if the
underlying value isn't actually of the asserted type. That's fine when
you've independently guaranteed the type (e.g. immediately after a type
switch case), but when the value's type isn't fully controlled — it came
from an interface parameter, a map of `interface{}`, deserialized data —
an unexpected type crashes the program instead of returning a handleable
error.

Note: this is an approximate, regex-based check on the single-line pattern
`ident := expr.(Type)` — it can't see whether the assertion is already
guarded by context (like an exhaustive prior type switch), so it may flag
some assertions that are actually safe.

**Fix:** use the comma-ok form and handle the failure case explicitly.

```go
// Flags this:
val := raw.(string)

// Prefer this:
val, ok := raw.(string)
if !ok {
    return fmt.Errorf("expected string, got %T", raw)
}
```
