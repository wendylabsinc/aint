# node-child-process-exec

**Flags:** a call to `child_process.exec(`.

**Why it matters:** `child_process.exec` runs its argument through a shell,
so shell metacharacters in the command string (`;`, `|`, `` ` ``, `$()`)
are meaningful. If any part of that string is built from external input,
an attacker can inject additional commands. `execFile`/`spawn`, by
contrast, take the program and its arguments as a separate argv array and
never invoke a shell, so there's no metacharacter parsing to exploit.

**Fix:** use `execFile` or `spawn` with the command and its arguments
passed as a list, instead of building a single shell command string.

```javascript
// Flags this:
child_process.exec(cmd);

// Prefer this:
child_process.execFile(cmd, args);
```
