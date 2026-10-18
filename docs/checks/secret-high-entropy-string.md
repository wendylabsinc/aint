# secret-high-entropy-string

**Flags:** a long, high-entropy-looking string literal — 40+ characters of
base64-ish characters (`A-Za-z0-9+/`), optionally with trailing `=`
padding. Excludes unpadded hexadecimal literals of exactly 40 or 64
characters, the lengths of full SHA-1 and SHA-256 Git object IDs. Hexadecimal
letters may be lowercase or uppercase. The separate `secret-hardcoded-key`
check still flags these values in credential assignments such as
`api_key = "…"` or `token = "…"`.

**Why it matters:** this is a heuristic, not a precise detector: long
base64-shaped strings are exactly what encoded API keys, tokens, and
encryption keys look like, but they're also what plenty of legitimate
non-secret data looks like (encoded test fixtures, hashes, generated IDs).
This check ships as `info` severity by default because of that
false-positive risk. Findings need manual review.

**Fix:** if the flagged string is a real credential, move it to an
environment variable or secrets manager. For confirmed non-sensitive data,
configure `checks: {secret-high-entropy-string: off}` if needed. The default
severity is informational, but agent hooks block on every reported finding.

```go
// Flags this (verify manually):
blob := "aGVsbG8gd29ybGQgdGhpcyBpcyBhIGxvbmcgYmFzZTY0IHN0cmluZw=="

// Does not flag a full Git commit ID:
commit := "e28fdb1d3af7d4671d701f5cfef70493ffaf87fa"

// If it's a real secret, prefer this:
blob := os.Getenv("ENCODED_SECRET")
```
