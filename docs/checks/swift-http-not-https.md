# swift-http-not-https

**Flags:** a string literal containing a hardcoded `http://` URL.

**Why it matters:** traffic over plain HTTP is unencrypted — anything sent
or received (including auth tokens, session cookies, or personal data) can
be read or modified in transit by anyone positioned on the network path.
A hardcoded `http://` endpoint also can't be upgraded to HTTPS later
without a code change and redeploy.

**Fix:** use `https://` for all hardcoded endpoints. If a service genuinely
doesn't support HTTPS, that's worth flagging as its own problem rather than
hardcoding a fallback to plaintext.

```swift
// Flags this:
let url = "http://api.example.com/data"

// Prefer this:
let url = "https://api.example.com/data"
```
