# swift-todo-comment

**Flags:** a line comment containing `TODO`, `FIXME`, or `HACK`
(case-insensitive).

**Why it matters:** these markers are useful during active development, but
left unresolved they represent known gaps, workarounds, or incomplete work
that's easy to forget about once it's buried in the codebase. Surfacing
them as a check keeps them visible instead of letting them silently ship
to production and get forgotten.

**Fix:** track the underlying work in an issue tracker (with a reference
back in the comment if useful) and resolve it before release, or remove
the comment if it no longer applies.

```swift
// Flags this:
// TODO: handle this edge case

// Prefer this:
// See PROJ-1234 for the edge case where the session has already expired.
```
