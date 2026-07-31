# swift-userdefaults-sensitive-key

**Flags:** a `UserDefaults.<suite>.set(..., forKey: "...")` call where the
key name contains `password`, `token`, `secret`, or `api_key`/`apikey`
(case-insensitive).

**Why it matters:** `UserDefaults` is backed by an unencrypted plist on
disk — it's meant for user preferences, not credentials. Anything stored
there is readable by anyone with filesystem access to the device (or, on
jailbroken/rooted devices, considerably easier access). The Keychain exists
specifically to store this class of sensitive data securely.

Note: this is a name-based heuristic — it flags based on the *key name*
looking sensitive, not the actual content, so it can both miss sensitive
data stored under an innocuous-looking key and flag a key whose name
happens to contain one of these words but isn't actually a credential.

**Fix:** store credentials, tokens, and secrets in the Keychain instead of
`UserDefaults`.

```swift
// Flags this:
UserDefaults.standard.set(authToken, forKey: "userAuthToken")

// Prefer this:
KeychainHelper.save(authToken, forKey: "userAuthToken")
```
