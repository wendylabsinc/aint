# node-cors-wildcard-credentials

**Flags:** a wildcard CORS origin (`origin: '*'`) combined with
`credentials: true` in the same configuration object (in either order).

**Why it matters:** `credentials: true` tells the browser it's allowed to
send cookies/auth headers with the cross-origin request and expose the
response to the requesting page. Combined with a wildcard origin, this
means *any* website can make an authenticated request to your API using
the victim's browser session and read the response — effectively
defeating the same-origin policy for authenticated data. (Browsers
actually reject this exact combination at the spec level for `fetch`
credentials, but plenty of server-side CORS middleware will still
construct and send it, and some clients don't enforce the restriction.)

**Fix:** use an explicit allowlist of trusted origins instead of a
wildcard when `credentials: true` is set.

```javascript
// Flags this:
cors({ origin: "*", credentials: true });

// Prefer this:
cors({ origin: "https://app.example.com", credentials: true });
```
