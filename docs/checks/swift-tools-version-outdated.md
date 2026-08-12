# swift-tools-version-outdated

**Flags:** a `// swift-tools-version:` header declaring 6.2 or earlier.

**Why it matters:** the tools-version header is not just a minimum
toolchain — it selects the manifest API and the default language mode SwiftPM
applies to the package. Pinning an old version silently opts the package out
of newer language features, defaults, and manifest capabilities, and that
choice then propagates to everyone depending on it. There is rarely a reason
for a new package to declare anything but the latest.

**Fix:** declare the latest tools version. As of this check, that is **6.3**:

```swift
// Flags this:
// swift-tools-version: 6.2

// Prefer this:
// swift-tools-version: 6.3
```

The header must be the very first line of `Package.swift`.

**Maintenance note:** this check hardcodes the current latest version. When
Swift 6.4 ships, bump both the regex (`6\.[0-2]` → `6\.[0-3]`) and the
message in `internal/checks/swift/swift.go`, and update the version named
above.

**Known exception:** a package deliberately supporting older toolchains for
downstream consumers. That's a real trade-off — make it explicitly, and
silence the check per-project in `.aint.yaml`:

```yaml
checks:
  swift-tools-version-outdated: off
```
