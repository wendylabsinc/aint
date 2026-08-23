# go-defer-in-loop

**Flags:** a `defer` statement lexically inside a `for` or `range` loop.

**Why it matters:** deferred calls run when the surrounding function returns, not at the end of the iteration. Files, response bodies, locks, or other resources can accumulate throughout the loop.

**Fix:** move one iteration into a helper or immediately invoked function so its deferred cleanup runs per iteration.

```go
for _, path := range paths {
    func() {
        f, err := os.Open(path)
        if err != nil { return }
        defer f.Close()
        // ...
    }()
}
```
