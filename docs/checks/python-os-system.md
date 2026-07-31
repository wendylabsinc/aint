# python-os-system

**Flags:** a call to `os.system(`.

**Why it matters:** `os.system` runs its argument through `/bin/sh`. If any
part of that command string is built from external input, shell
metacharacters in it (`;`, `|`, `` ` ``, `$()`) let an attacker inject and
run additional commands with whatever privileges the Python process has —
the same class of risk as `subprocess.run(..., shell=True)`.

**Fix:** use `subprocess.run()` with a list of arguments (no shell
involved) instead of building a single command string.

```python
# Flags this:
os.system(f"cat {user_input}")

# Prefer this:
subprocess.run(["cat", user_input])
```
