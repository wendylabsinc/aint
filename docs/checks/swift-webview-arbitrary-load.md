# swift-webview-arbitrary-load

**Flags:** `.load(URLRequest(url: ...))` called on a `WKWebView`.

**Why it matters:** loading a `URLRequest` into a `WKWebView` executes
whatever the target page returns as if it were part of the app's UI —
including any JavaScript the page runs (unless script execution is
explicitly disabled). If the URL being loaded isn't validated — comes from
a deep link, a server response, or another untrusted source — the app is
effectively giving an attacker-controlled page a foothold inside the app's
webview context.

**Fix:** verify the URL comes from a trusted, validated source (a fixed
allowlist of hosts, or a URL you constructed yourself) before loading it.
For static, trusted local content, `loadHTMLString`/`loadFileURL` avoid the
network entirely.

```swift
// Flags this (verify the URL's origin first):
webView.load(URLRequest(url: someURL))

// Safer for static content:
webView.loadHTMLString(staticHTML, baseURL: nil)

// If loading a remote URL, validate the host first:
guard allowedHosts.contains(someURL.host ?? "") else { return }
webView.load(URLRequest(url: someURL))
```
