# shell-grep-search-codebase

**Flags:** shelling out to `grep -r`/`grep -l`/`grep --include`/`rg` to search a
codebase, instead of using the editor's dedicated search tool.

**Why it matters:** an agent that already has a file-search tool available
(Claude Code's `Grep`/`Read` tools) gets no benefit from re-implementing the
same search via a raw shell command, and it's disruptive to watch: it adds
noise, re-reads files that were already read, and burns time on searches a
direct `Read` would have skipped entirely once the path is known.

**Fix:** use the `Grep` tool for "where is X defined" style searches, and
`Read` directly once you know the path - don't grep for every type name or
method signature if you can read the relevant file once. This check does not
flag `grep`/`rg` used to filter another command's output (e.g. `ps aux | grep
node`), only recursive/multi-file source search.

```bash
# Flags this:
grep -rn "func HandleRequest" .
rg "TODO" src/

# Prefer this:
# (use the Grep tool, or Read the file directly if the path is already known)
```
