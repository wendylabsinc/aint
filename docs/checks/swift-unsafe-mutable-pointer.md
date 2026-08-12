# swift-unsafe-mutable-pointer

**Flags:** use of `UnsafeMutablePointer`, `UnsafeMutableRawPointer`,
`UnsafeMutableBufferPointer`, or `UnsafeMutableRawBufferPointer`.

**Why it matters:** these types bypass Swift's memory safety guarantees —
no bounds checking, no lifetime tracking, no exclusivity enforcement. A
pointer that outlives the buffer it points into, an off-by-one write, or a
type mismatch between what was allocated and what's read back are all
undefined behavior at runtime instead of compile-time or trap-time errors.
They're also unrelated to actor isolation, so a mutable pointer can be a
data-race vector if shared across concurrency domains without independent
synchronization.

**Fix:** prefer scoped, safe alternatives that give the compiler (or at
least a bounded closure) something to reason about:
- `Span`/`MutableSpan` (Swift 6.3+) for a bounds-checked, lifetime-tracked
  view over contiguous memory.
- `InlineArray`/`ContiguousArray`/`[T]` for owned storage.
- `withUnsafeMutableBufferPointer { buf in ... }` /
  `withUnsafeMutableBytes { ... }` when you must interoperate with a C API
  — the pointer is only valid inside the closure, so it can't accidentally
  escape or outlive its backing storage.

```swift
// Flags this:
let buf = UnsafeMutablePointer<UInt8>.allocate(capacity: count)
defer { buf.deallocate() }
fill(buf, count)

// Prefer this:
var storage = [UInt8](repeating: 0, count: count)
storage.withUnsafeMutableBufferPointer { buf in
    fill(buf.baseAddress!, buf.count)
}
```
