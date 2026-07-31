# go-exec-command-injection

**Flags:** `exec.Command("sh", "-c", ...)` (or `bash`/`/bin/sh`/`/bin/bash`
in place of `sh`) — spawning a shell explicitly via `-c`.

**Why it matters:** spawning a shell via `-c` hands the shell a single
string to interpret, which means shell metacharacters (`;`, `|`, `` ` ``,
`$()`) in that string are meaningful. If any part of the command string is
built from external input rather than a fixed literal, an attacker can
inject additional commands. `exec.Command` normally avoids this entirely by
passing arguments directly to the program (no shell involved) — using `-c`
opts back into shell parsing.

**Fix:** call the target program directly with its arguments as separate
strings, without going through a shell. If shell features (pipes, globs,
redirection) are genuinely required, ensure every part of the command
string is a fixed literal, never user- or externally-controlled input.

```go
// Flags this:
cmd := exec.Command("sh", "-c", cmd)

// Prefer this:
cmd := exec.Command("ls", "-la")
```
