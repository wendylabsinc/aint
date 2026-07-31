# secret-jwt-token

**Flags:** a hardcoded JWT-shaped string in source (three base64url segments
separated by dots, e.g. `eyJhbGciOi....eyJzdWIiOi....dozjgNr...`).

**Why it matters:** a JWT committed to source is a real, usable credential
until it expires — unlike a password, it can't be rotated by changing a
config value, since anyone who reads the repo (or its git history) can
replay the token as-is. It often ends up there from copy-pasting a
debugging session or a "temporary" test fixture that never got removed.

**Fix:** don't hardcode tokens. Generate them at runtime, load them from a
secrets manager or environment variable for tests, and use short-lived
tokens with proper rotation instead of long-lived ones baked into code.

```go
// Flags this:
token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"

// Prefer this:
token := os.Getenv("TEST_JWT")  // or mint a fresh one in test setup
```
