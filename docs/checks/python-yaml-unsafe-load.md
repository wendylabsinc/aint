# python-yaml-unsafe-load

**Flags:** a call to `yaml.load(...)`.

**Why it matters:** `yaml.load()` can deserialize arbitrary Python objects via
`!!python/object` tags, which can execute arbitrary code when the loaded
document isn't trusted. Note: this check flags every `yaml.load(...)` call,
including ones that already pass `Loader=yaml.SafeLoader` — aint's regex
engine can't see keyword arguments reliably, so a call that's actually safe
may still be flagged. Prefer `yaml.safe_load()` outright, which sidesteps
the ambiguity entirely.

**Fix:** use `yaml.safe_load()`, which only ever constructs plain Python
types (dicts, lists, strings, numbers).

```python
# Flags this:
config = yaml.load(f)

# Prefer this:
config = yaml.safe_load(f)
```
