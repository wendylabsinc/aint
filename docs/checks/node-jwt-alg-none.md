# node-jwt-alg-none

**Flags:** an `algorithms`/`algorithm` option containing `'none'` (e.g.
`{ algorithms: ['none'] }`).

**Why it matters:** the JWT `none` algorithm means the token has no
signature at all. If a verifier accepts it, an attacker can hand-craft a
token with an arbitrary payload — claiming to be any user, with any role —
and it will pass verification since there's nothing to check. This is one
of the most well-known JWT library footguns.

**Fix:** always specify an explicit allowlist of real signing algorithms
(e.g. `HS256`, `RS256`) and never include `none` in it.

```javascript
// Flags this:
jwt.verify(token, secret, { algorithms: ["none"] });

// Prefer this:
jwt.verify(token, secret, { algorithms: ["HS256"] });
```
