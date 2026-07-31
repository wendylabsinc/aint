# swift-implicitly-unwrapped-optional

**Flags:** a type annotation ending in `!` (e.g. `var window: UIWindow!`).

**Why it matters:** an implicitly unwrapped optional behaves like a regular
optional but is force-unwrapped automatically every time it's accessed. If
it's accessed while still `nil` — before it's been assigned, after being
reset, or on a code path that never assigns it — the app crashes at that
access site, with no compiler-enforced check reminding you it could be
`nil`.

**Fix:** use a regular optional (`UIWindow?`) and unwrap it explicitly
(`if let`, `guard let`), or restructure the code so the value is available
at initialization time and doesn't need to be optional at all.

```swift
// Flags this:
var window: UIWindow!

// Prefer this:
var window: UIWindow?

// ...and unwrap explicitly at the point of use:
guard let window else { return }
```
