# python-eval-exec

**Flags:** a call to `eval(` or `exec(`.

**Why it matters:** `eval`/`exec` compile and run a string as arbitrary
Python code. If any part of that string can be influenced by external
input — a request parameter, a config value from an untrusted source, user
input of any kind — an attacker can run arbitrary code with the
permissions of the process.

**Fix:** avoid `eval`/`exec` entirely where possible. For parsing data
(not code), use `ast.literal_eval` for literal Python values or `json`
for structured data; for building small DSLs, write an explicit parser
instead of executing arbitrary source.

```python
# Flags this:
result = eval(user_input)

# Prefer this, for literal values:
import ast
result = ast.literal_eval(user_input)

# Prefer this, for structured data:
import json
result = json.loads(user_input)
```
