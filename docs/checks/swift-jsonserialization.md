# swift-jsonserialization

**Flags:** use of Foundation's `JSONSerialization` (the untyped `[String:
Any]` / `jsonObject(with:)` API).

**Why it matters:** `JSONSerialization`'s untyped casts push type-safety
checks to runtime and invite sloppy `as?`/`as!` chains where a typed decoder
would catch shape mismatches at compile time or with a clear decode error.

**Fix:** use **swift-json-schema** instead:
- Typed payloads: define a `@Schemable struct` and decode with
  `Type.schema.decode(data)`.
- Ad-hoc/untyped parsing (e.g. in tests): `JSONValue.parse(data)`, then
  pattern-match `case .object(let obj)` and index `obj["key"]` (returns
  `JSONValue?`). Note `JSONValue.parse` returns an `OrderedDictionary`, not
  `[String: JSONValue]`.

```swift
// Flags this:
let obj = try JSONSerialization.jsonObject(with: data) as? [String: Any]

// Prefer this:
let envelope = try SimStateEnvelope.schema.decode(data)
```
