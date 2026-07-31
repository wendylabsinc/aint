# swift-fatal-error

**Flags:** a call to `fatalError(`.

**Why it matters:** `fatalError` immediately terminates the process — there
is no way to recover, log context gracefully, or degrade functionality.
It's appropriate for truly unreachable states during development, but
reaching one in production means an unhandled situation crashes the entire
app for the user, rather than failing a single operation or returning an
error the caller can act on.

**Fix:** for recoverable conditions, throw an error the caller can catch
and handle. For genuine programmer invariants that should never happen at
runtime, `precondition`/`preconditionFailure` document intent similarly
but consider whether the condition can be handled instead of crashed on.

```swift
// Flags this:
fatalError("unreachable")

// Prefer this, when the caller can recover:
throw MyError.unexpectedState

// Or, for a true programmer invariant:
preconditionFailure("unreachable")
```
