# secret-private-key-block

**Flags:** private key material (RSA, EC, or generic PRIVATE KEY blocks) committed to files.

**Why it matters:** a private key in version control means anyone with repo access can decrypt or sign on your behalf. Private keys should never be committed, even temporarily — they must be generated securely and managed outside the codebase.

**Fix:** remove the key file immediately and revoke the key pair. Regenerate a new key pair, keep the private key in a secure location (e.g., `~/.ssh/`, a secrets manager, or a hardware token), and only commit the public key (or its fingerprint) if necessary. If the key was exposed, rotate any credentials or access tokens that relied on it.

```
# Flags this:
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA3x4...
...
-----END RSA PRIVATE KEY-----

# Prefer this:
# Key file stored securely outside the repo (e.g., ~/.ssh/id_rsa)
# Only public key or certificate committed:
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A...
-----END PUBLIC KEY-----
```
