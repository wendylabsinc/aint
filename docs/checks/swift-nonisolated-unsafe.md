# swift-nonisolated-unsafe

**Flags:** use of the `nonisolated(unsafe)` attribute.

**Why it matters:** `nonisolated(unsafe)` tells the compiler to skip
data-race checking for that specific declaration entirely — it's an escape
hatch from Swift's actor isolation model, not a fix. Every use is a spot
where the compiler can no longer prove the code is safe; the safety
argument now lives in the programmer's head instead of in the type system.

**Fix:** first try to make the isolation checker happy the normal way —
move the state behind an actor, make the type `Sendable`, or restructure so
the value doesn't cross an isolation boundary. If `nonisolated(unsafe)` is
genuinely required (a common legitimate case: a C global or other
external resource that is provably safe to touch from any isolation
domain), keep it narrowly scoped, document *why* it's safe in a comment,
and leave a follow-up ticket to remove it if the underlying constraint
ever goes away.

```swift
// Flags this:
nonisolated(unsafe) var sharedCache: [String: Data] = [:]

// Prefer this:
actor Cache {
    private var storage: [String: Data] = [:]
    func get(_ key: String) -> Data? { storage[key] }
    func set(_ key: String, _ value: Data) { storage[key] = value }
}

// Acceptable exception, narrowly scoped and documented:
@inline(__always)
private func flushStdout() {
    // Safe: `stdout` is a process-wide C global with no Swift isolation
    // domain of its own; fflush is thread-safe per C stdlib guarantees.
    nonisolated(unsafe) let out = stdout
    fflush(out)
}
```
