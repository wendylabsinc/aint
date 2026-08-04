# shell-git-add-broad

**Flags:** `git add -A`, `git add --all`, `git add -u`, or a bare `git add .`.

**Why it matters:** a broad `git add` stages the whole working tree, not just
the files a task actually touched. If the tree already has unrelated
uncommitted WIP (common when a shared main tree or a subagent's `git add
<path>` sweeps a whole file, not just its own hunks), that WIP gets bundled
into the task's commit - splitting a feature across commits and sometimes
leaving the committed HEAD unable to build, while the working tree still
looks fine because the rest of the WIP is still sitting there uncommitted.

**Fix:** run `git status` first and stage only the specific files the current
task changed. If a file the task needs to touch already has unrelated
uncommitted changes, have that WIP committed on its own first (or stage with
`git add -p` / explicit paths) rather than sweeping it in.

```bash
# Flags this:
git add -A
git add .

# Prefer this:
git status
git add internal/apps_install.go internal/apps_install_test.go
```
