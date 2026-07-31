# python-assert-security-check

**Flags:** `assert` followed by something referencing `is_admin`,
`authenticated`, `authorized`, or `has_permission` (case-insensitive).

**Why it matters:** `assert` statements are stripped entirely when Python
runs with the `-O` (optimize) flag — the check silently vanishes, and
whatever it was gating executes unconditionally. Using `assert` to gate an
authorization or permission check means that check can disappear in an
optimized deployment with no error, no warning, and no obvious sign
anything changed.

**Fix:** use an explicit `if`/`raise` (or your framework's authorization
decorator/dependency) for anything security-relevant — never `assert`,
which is meant for internal invariants during development/testing, not
runtime security gates.

```python
# Flags this:
assert user.is_admin

# Prefer this:
if not user.is_admin:
    raise PermissionError("admin access required")
```
