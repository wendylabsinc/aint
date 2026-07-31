# python-requests-no-verify

**Flags:** a `requests.<method>(...)` call with `verify=False`.

**Why it matters:** `verify=False` disables TLS certificate verification
for that request, meaning `requests` will accept a certificate from any
server — expired, self-signed, or attacker-controlled — without complaint.
This is exactly the opening a man-in-the-middle attack needs. It's often
added to silence a certificate error during local development and then
left in place.

**Fix:** don't disable verification. If the issue is a self-signed or
internal CA certificate, point `verify` at that CA bundle explicitly
instead of turning verification off.

```python
# Flags this:
requests.get(url, verify=False)

# Prefer this:
requests.get(url)  # verify=True is the default

# Or, for an internal CA:
requests.get(url, verify="/path/to/internal-ca-bundle.pem")
```
