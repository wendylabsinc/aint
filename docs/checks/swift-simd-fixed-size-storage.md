# swift-simd-fixed-size-storage

**Flags:** use of a `SIMDn` type (`SIMD2`, `SIMD4`, `SIMD16`, `SIMD64`, …).

**Why it matters:** `SIMDn` is a lane-parallel arithmetic type, not a
general-purpose fixed-size container. Reaching for `SIMD16<UInt8>` because
you want "16 bytes inline, no heap allocation" gets you an API shaped
around vector math — element access via lane subscripts, a `Scalar` type
constrained to `SIMDScalar`, no `Collection` conformance, and sizes limited
to powers of two. `InlineArray` gives the same inline, allocation-free
storage with ordinary element access and any element type.

**Fix:** use `InlineArray<Count, Element>` for fixed-capacity inline
storage. Keep `SIMDn` only where you are actually doing lane-parallel
arithmetic.

```swift
// Flags this (SIMD used purely as a 4-slot buffer):
var channels = SIMD4<Float>()
channels[0] = sample

// Prefer this:
var channels: InlineArray<4, Float> = [0, 0, 0, 0]
channels[0] = sample
```

**Known exception — real vector math:** genuine SIMD arithmetic (geometry,
DSP, colour transforms) is exactly what `SIMDn` is for, and this check will
flag it too. That's why it ships at `info` severity. If a project is
graphics- or DSP-heavy, turn it off in `.aint.yaml`:

```yaml
checks:
  swift-simd-fixed-size-storage: off
```

```swift
// Legitimate — lane-parallel arithmetic, not storage:
let scaled = position * SIMD3<Float>(repeating: scale)
```
