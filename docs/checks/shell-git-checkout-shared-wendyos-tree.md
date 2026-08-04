# shell-git-checkout-shared-wendyos-tree

**Flags:** a `git checkout <branch>` or `git switch <branch>` targeting the
shared `wendyos` main working tree (e.g. `cd ~/git/wendy/wendyos && git
checkout ...` or `git -C ~/git/wendy/wendyos switch ...`). Sibling repos with
a similar name (`wendyos-builder`, `wendyos-update`, etc.) are separate trees
and are not flagged.

**Why it matters:** `~/git/wendy/wendyos`'s main tree is frequently shared by
concurrent Claude/agent sessions that create and check out branches in it
while another session is mid-task. A branch switch there can redirect the
other session's next commit onto the wrong branch, or race with it rewriting
refs (`fatal: Needed a single revision`).

**Fix:** for any multi-commit work in this repo, create an isolated `git
worktree` for the feature branch up front and do all edits/commits there
instead of switching branches in the shared main tree.

```bash
# Flags this:
cd ~/git/wendy/wendyos && git checkout jo/some-feature

# Prefer this:
git -C ~/git/wendy/wendyos worktree add ../wendyos-jo-some-feature jo/some-feature
cd ~/git/wendy/wendyos-jo-some-feature
```
