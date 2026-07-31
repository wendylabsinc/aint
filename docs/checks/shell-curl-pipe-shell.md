# shell-curl-pipe-shell

**Flags:** `curl ... | bash` or `curl ... | sh` (with an optional `sudo`
before the shell).

**Why it matters:** piping a remote script straight into a shell executes
it with zero verification — no checksum, no signature, no review of what
actually ran. If the remote host is compromised, or the connection is
intercepted, this runs arbitrary code with whatever privileges the shell
has.

**Fix:** download the script first, review it (or verify a checksum/
signature), then execute it explicitly.

```bash
# Flags this:
curl -sSL https://example.com/install.sh | bash

# Prefer this:
curl -sSL https://example.com/install.sh -o install.sh
sha256sum install.sh   # verify against a published checksum
bash install.sh
```
