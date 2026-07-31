# node-tls-reject-unauthorized-false

**Flags:** `rejectUnauthorized: false` in a TLS/HTTPS options object, or
setting `NODE_TLS_REJECT_UNAUTHORIZED` to `'0'`.

**Why it matters:** both disable TLS certificate verification — for a
single request in the first case, or process-wide (every TLS connection
the process makes) in the second. Either way, the client will accept a
certificate from any server, including an attacker performing a
man-in-the-middle attack. The environment-variable form is especially
dangerous since it's easy to leave set globally after using it to debug a
certificate issue once.

**Fix:** don't disable verification. If the issue is a self-signed or
internal CA certificate, supply that CA explicitly via the `ca` option
instead.

```javascript
// Flags this:
https.request({ rejectUnauthorized: false });

// Prefer this:
https.request({ ca: fs.readFileSync("internal-ca.pem") });
```
