# go-tls-insecure-skip-verify

**Flags:** `InsecureSkipVerify: true` in a `tls.Config`.

**Why it matters:** `InsecureSkipVerify` disables verification of the
server's TLS certificate, meaning the client will happily connect to any
server presenting any certificate — expired, self-signed, or one belonging
to an attacker performing a man-in-the-middle attack. It's sometimes added
temporarily to work around a certificate error in development and
accidentally left in place.

**Fix:** don't disable verification. If the underlying issue is a
self-signed or internal CA certificate, add that CA to the `RootCAs` pool
instead of turning off verification entirely.

```go
// Flags this:
tlsConfig := &tls.Config{InsecureSkipVerify: true}

// Prefer this:
tlsConfig := &tls.Config{
    RootCAs: certPool, // pool containing the trusted internal CA
}
```
