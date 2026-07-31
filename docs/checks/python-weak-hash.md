# python-weak-hash

**Flags:** `hashlib.md5(` or `hashlib.sha1(`.

**Why it matters:** MD5 and SHA-1 are cryptographically broken and, more
importantly for this context, are fast general-purpose hashes — not
password-hashing functions. Used directly on a password, they're
trivially brute-forceable with commodity hardware. They're also unsafe
anywhere collision resistance matters.

**Fix:** for password hashing, use a dedicated password-hashing library
(`bcrypt`, `argon2-cffi`) that's deliberately slow and salted. For
general-purpose hashing where collisions matter, use SHA-256 or better.

```python
# Flags this:
hashlib.md5(password.encode())

# Prefer this, for general hashing:
hashlib.sha256(password.encode())

# Prefer this, for password hashing:
import bcrypt
hashed = bcrypt.hashpw(password.encode(), bcrypt.gensalt())
```
