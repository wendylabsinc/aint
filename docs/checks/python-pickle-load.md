# python-pickle-load

**Flags:** a call to `pickle.load(` or `pickle.loads(`.

**Why it matters:** unpickling data can execute arbitrary code — a crafted
pickle stream can call any callable during deserialization via its
`__reduce__` protocol. If the data being unpickled comes from anywhere
outside your own trusted, tightly-controlled process (a network request,
a file uploaded by a user, a cache shared with less-trusted code), loading
it is equivalent to running arbitrary code from that source.

**Fix:** use a safe serialization format like `json` for data that needs to
cross a trust boundary. If pickle is genuinely required (e.g. for internal
caching of Python-specific objects), make sure the data source is fully
trusted and never externally writable.

```python
# Flags this:
data = pickle.loads(payload)

# Prefer this:
data = json.loads(payload)
```
