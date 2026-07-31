# secret-high-entropy-string

**Flags:** a long, high-entropy-looking string literal — 40+ characters of
base64-ish characters (`A-Za-z0-9+/`), optionally with trailing `=`
padding.

**Why it matters:** this is a heuristic, not a precise detector: long
base64-shaped strings are exactly what encoded API keys, tokens, and
encryption keys look like, but they're also what plenty of legitimate
non-secret data looks like (encoded test fixtures, hashes, generated IDs).
This check ships as `info` severity by default specifically because of that
false-positive risk — it's flagged as "look at this and confirm" rather
than "this is definitely wrong."

**Fix:** if the flagged string is a real credential, move it to an
environment variable or secrets manager. If it's confirmed to be
non-sensitive (a fixture, a hash, an encoded asset), no action is needed —
this check is informational.

```go
// Flags this (verify manually):
blob := "aGVsbG8gd29ybGQgdGhpcyBpcyBhIGxvbmcgYmFzZTY0IHN0cmluZw=="

// If it's a real secret, prefer this:
blob := os.Getenv("ENCODED_SECRET")
```
