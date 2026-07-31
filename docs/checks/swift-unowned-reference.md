# swift-unowned-reference

**Flags:** `unowned` used in a capture list (e.g. `[unowned self]`) or a
property declaration (`unowned let`/`unowned var`).

**Why it matters:** an `unowned` reference doesn't zero out when the
referenced object is deallocated — it becomes a dangling pointer. Accessing
it after that point crashes the process immediately, with no chance to
handle the situation gracefully. `weak` references, by contrast, become
`nil` on deallocation, so the accessing code can check and handle the
missing-reference case explicitly instead of crashing.

**Fix:** prefer `weak` and explicitly unwrap/handle the `nil` case, unless
you can prove the referenced object always strictly outlives every place
that accesses it (in which case, document why).

```swift
// Flags this:
task = Task { [unowned self] in await self.run() }

// Prefer this:
task = Task { [weak self] in
    guard let self else { return }
    await self.run()
}
```
