# aint

Static analysis for code and shell commands, with human-readable findings and a link to docs for each one.

## Usage

```
aint check [paths...]           # scan files/dirs (default: .)
aint check --format=json ...    # machine-readable output
aint list                       # list all registered checks
aint install [--global]         # wire aint into Claude Code hooks
```

## Claude Code integration

Run `aint install` from a project root to wire `aint` into `.claude/settings.json`: shell commands are checked (and blocked on any finding) before Claude runs them, and files Claude writes or edits are checked afterward, with findings reported back as feedback.

**Important distinction:** `aint hook` blocks on ALL findings regardless of the `fail_on` setting in `.aint.yaml` — only `checks: <id>: off` silences a check in hook mode. This differs from `aint check`, which respects `fail_on` to gate exit code behavior. Hooks are stricter to catch issues before they land in your code.

Use `--global` to install into `~/.claude/settings.json` instead of the current project.

## Configuration

Optional `.aint.yaml` at the repo root — see the example file in this repo for the full set of options (`fail_on`, `ignore`, `checks`, `docs_base_url`).

## Checks

Run `aint list` for the full, current set. See `docs/checks/` for what each one flags and how to fix it.
