# python-shell-true

**Flags:** subprocess calls (e.g. `subprocess.run()`, `subprocess.Popen()`) with `shell=True`.

**Why it matters:** when `shell=True`, the subprocess module spawns a shell and passes the command string to it. If any part of that command string comes from untrusted input — user input, environment variables, or external data — an attacker can inject shell metacharacters (e.g., `; rm -rf /`) and execute arbitrary commands.

**Fix:** avoid `shell=True` and pass the command as a list of arguments instead. If you need shell features like pipes or redirects, construct the command carefully or use the shell sparingly, validating all inputs.

```python
# Flags this:
user_input = request.args.get("file")
subprocess.run(f"cat {user_input}", shell=True)

# Prefer this:
user_input = request.args.get("file")
subprocess.run(["cat", user_input])  # shell=False (default)

# If pipes or redirects are truly needed, use shell-safe features:
result = subprocess.run(
    ["cat", user_input],
    shell=False,
    capture_output=True
)
```
