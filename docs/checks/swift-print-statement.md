# swift-print-statement

**Flags:** `print(`, `debugPrint(`, `dump(`, or `NSLog(`.

**Why it matters:** these functions write straight to stdout/the system
log with no structure, no log levels, and no way to route, filter, or
redact output in production. Debug statements added during development
routinely get left in, and `NSLog`/`print` output can end up in device
logs or crash reports where it's harder to control what leaks (including
values that shouldn't be logged at all).

**Fix:** use `swift-log`'s `Logger` instead, with an appropriate log level,
so output can be filtered, redacted, and routed consistently in
production.

```swift
// Flags this:
print("debug: \(value)")
NSLog("debug: %@", value)

// Prefer this:
logger.debug("debug: \(value)")
```
