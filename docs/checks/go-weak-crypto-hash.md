# go-weak-crypto-hash

**Flags:** `md5.New(`, `md5.Sum(`, `sha1.New(`, or `sha1.Sum(`.

**Why it matters:** MD5 and SHA-1 are both cryptographically broken —
practical collision attacks exist for both. They're unsafe for password
hashing (they're also just fast general-purpose hashes, not
password-hashing functions, so they're trivially brute-forceable) and
shouldn't be relied on anywhere collision resistance matters, such as
integrity checksums for security-sensitive data.

**Fix:** for password hashing, use a dedicated password-hashing algorithm
(bcrypt, scrypt, or argon2) that's deliberately slow and salted. For
general-purpose hashing/checksums, use SHA-256 or better.

```go
// Flags this:
h := md5.New()

// Prefer this, for general hashing:
h := sha256.New()

// Prefer this, for password hashing:
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```
