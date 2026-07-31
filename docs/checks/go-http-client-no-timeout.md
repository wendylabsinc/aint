# go-http-client-no-timeout

**Flags:** `http.Client{}` constructed with no fields set (an empty struct
literal on the same line).

**Why it matters:** the default `http.Client` has no request timeout at
all. If the remote server hangs — never sends a response, or sends one
byte at a time forever — the goroutine making the request blocks
indefinitely. In a server handling many requests, this is an easy way to
leak goroutines and exhaust resources under a slow or unresponsive
dependency.

**Fix:** always set a `Timeout` when constructing an `http.Client`, sized
to what's reasonable for the calls it will make.

```go
// Flags this:
client := &http.Client{}

// Prefer this:
client := &http.Client{Timeout: 10 * time.Second}
```
