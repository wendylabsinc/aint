# swift-jsonserialization

**Flags:** use of Foundation's `JSONSerialization` (the untyped `[String:
Any]` / `jsonObject(with:)` API).

**Why it matters:** `JSONSerialization`'s untyped casts push type-safety
checks to runtime and invite sloppy `as?`/`as!` chains where a typed decoder
would catch shape mismatches at compile time or with a clear decode error.

**Fix:** choose the replacement that matches the payload:

- For a stable, typed payload, use
  [wendylabsinc/swift-json-schema](https://github.com/wendylabsinc/swift-json-schema).
  Add a JSON Schema Draft 2020-12 `.schema.json` file to the target and apply
  its `JSONSchemaPlugin`; the plugin generates concrete Swift models and
  decoding code at build time. Generated models also conform to `Codable`,
  and the optional `SwiftJSON` trait adds direct `Span<UInt8>`,
  `JSONObjectView`, and `JSONObject` initializers.
- For raw, dynamic, or ad-hoc JSON, use
  [orlandos-nl/swift-json](https://github.com/orlandos-nl/swift-json)
  (IkigaJSON). Its `JSONObject` and `JSONArray` types provide structured raw
  JSON access without falling back to `[String: Any]`; it also provides
  `IkigaJSONDecoder` for `Codable` payloads.

```swift
// Flags this:
let obj = try JSONSerialization.jsonObject(with: data) as? [String: Any]

// Prefer a generated model for a stable payload:
let user = try JSONDecoder().decode(User.self, from: data)

// Or use IkigaJSON for a dynamic payload:
import IkigaJSON

let object = try JSONObject(data: data)
let name = object["name"].string
```

For `swift-json-schema`, add the package and plugin to `Package.swift`:

```swift
.package(
    url: "https://github.com/wendylabsinc/swift-json-schema.git",
    from: "0.1.0"
)

.target(
    name: "MyTarget",
    plugins: [
        .plugin(name: "JSONSchemaPlugin", package: "swift-json-schema")
    ]
)
```

Use both packages when useful: `swift-json-schema` can generate the typed
model while its `SwiftJSON` trait uses IkigaJSON for direct, lower-allocation
decoding.
