# swift-force-unwrap

**Flags:** force try (`try!`) and force cast (`as!`) operations, which crash at runtime if the operation fails.

**Why it matters:** a force try on an error or a failed force cast crashes your app immediately. This is worse than a graceful error or a safe fallback — it leaves no room for recovery. Swift's optional system is designed to make these situations explicit and safer.

**Fix:** use the safe operators `try?` and `as?`, which return nil on failure, allowing you to handle the error gracefully or use a default value.

```swift
// Flags this:
try! doSomethingRisky()
let castedValue = object as! MyType

// Prefer this:
if let result = try? doSomethingRisky() {
    // handle success
} else {
    // handle failure
}

if let castedValue = object as? MyType {
    // use castedValue safely
} else {
    // handle cast failure
}
```
