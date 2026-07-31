# swift-force-unwrap

**Flags:** force unwrap operations using `!` and force casts using `as!`, which crash at runtime if the value is nil or the cast fails.

**Why it matters:** a force unwrap on a nil value or a failed force cast crashes your app immediately. This is worse than a graceful error or a safe fallback — it leaves no room for recovery. Swift's optional system is designed to make these situations explicit and safer.

**Fix:** use the safe unwrap operators `try?` and `as?`, which return nil on failure, or add an explicit guard or if-let binding to check the value before using it.

```swift
// Flags this:
let number = Int(stringValue)!
let castedValue = object as! MyType
try! doSomethingRisky()

// Prefer this:
if let number = Int(stringValue) {
    // use number safely
}

if let castedValue = object as? MyType {
    // use castedValue safely
}

let result = try? doSomethingRisky()  // returns nil on error
```
