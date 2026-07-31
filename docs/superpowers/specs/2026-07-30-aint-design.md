# aint — static analysis CLI + Claude Code hook

Date: 2026-07-30

## Purpose

`aint` is a static-analysis CLI that sits between a developer (or Claude Code)
and the code/commands they produce. It scans files and shell commands for
common issues — hardcoded secrets, overscoped IAM/shell permissions, and
per-language footguns — and reports them as human-readable
`file:line:col` findings with a link to docs explaining the issue and the fix.

The first milestone (this spec) is the **framework**: the check engine, CLI,
config format, a small seed set of checks across five languages/domains, and
a `aint install` command that wires `aint` into Claude Code's hooks so it
automatically checks code Claude writes and shell commands Claude runs.

Future milestones (explicitly out of scope here) will expand the check
catalog per language and add IaC-specific parsing (Terraform HCL, k8s
manifests) beyond the regex-based seed checks.

## Non-goals (v1)

- No AST/tree-sitter parsing — all checks are regex/line-based.
- No auto-fix — `aint` reports, it does not modify code.
- No uninstall command — removing hooks is a manual settings.json edit.
- No remote/cloud config or telemetry.
- Not an exhaustive check catalog — just enough checks per category to prove
  the framework works end-to-end.

## Architecture

Single static Go binary, no runtime dependencies.

```
aint/
  cmd/aint/main.go        # CLI entrypoint (cobra or stdlib flag-based subcommands)
  internal/
    check/                 # Check, Finding types + global registry
    checks/
      secrets/             # secret-hardcoded-key, secret-private-key-block
      shell/                # shell-gcp-role-wildcard, shell-chmod-permissive
      golang/               # go-ignored-error
      swift/                # swift-force-unwrap
      python/               # python-shell-true
      nodejs/               # node-eval
    scan/                   # file walking, language classification, ignore globs
    config/                 # .aint.yaml loading + defaults
    report/                 # human-readable + json formatters
    hook/                   # Claude Code hook JSON parsing + install merge logic
  docs/checks/<id>.md       # one doc page per check, linked from findings
  .aint.yaml                # example config (not required for aint to run)
  go.mod
```

**`internal/check`** defines:

```go
type Severity string // "info" | "warning" | "error"

type Kind string // "file" | "shell"

type Check struct {
    ID       string
    Title    string
    Severity Severity
    Kind     Kind
    Langs    []string // file extensions/language tags this applies to; ignored for Kind=="shell"
    Pattern  *regexp.Regexp
    Message  string
    DocsURL  string // resolved at report time: docs_base_url + ID if docs_base_url is set, else a relative docs/checks/<ID>.md path
}

type Finding struct {
    CheckID  string
    Severity Severity
    Source   string // file path, or "<command>" for shell-mode
    Line     int
    Column   int
    Message  string
    DocsURL  string
}
```

Checks register themselves into a package-level registry via `init()` in
each `checks/*` package; `cmd/aint/main.go` blank-imports all check packages
so registration happens on startup.

**`internal/scan`** walks the given paths (or wraps a single in-memory shell
command string for hook mode), classifies each file by extension/shebang,
applies `.aint.yaml` ignore globs, and runs every registered `Check` whose
`Kind`/`Langs` match against each target, collecting `Finding`s.

**`internal/report`** renders findings either as human-readable lines:

```
path/to/file.go:42:5: error [go-ignored-error] error return value is discarded — https://.../docs/checks/go-ignored-error.md
```

or as `--format=json` for scripting.

## CLI

```
aint check [paths...]           # scan files/dirs (default: .) — exit 0 clean, 1 if findings >= fail_on
aint check --format=json ...    # machine-readable output
aint check --stdin-command      # read a shell command string from stdin, run Kind=="shell" checks only
aint list                       # list all registered checks: ID, severity, langs, description, docs link
aint install [--global]         # merge aint hooks into Claude Code settings.json
```

## Config — `.aint.yaml`

Optional; built-in defaults apply if absent. Located at the repo root (or
nearest ancestor directory containing one, similar to how `.git` is found).

```yaml
fail_on: error          # info|warning|error — findings at/above this exit 1 for `aint check`; hook mode always blocks/feeds back on any finding regardless of fail_on (see Claude Code hook integration below)
ignore:
  - vendor/**
  - "*.pb.go"
checks:
  go-ignored-error: error
  node-console-log: off   # off|info|warning|error — override severity or disable a check entirely
docs_base_url: ""         # optional; if unset, DocsURL is a relative docs/checks/<id>.md path
```

## Seed checks (v1)

| ID | Kind/Langs | Detects |
|---|---|---|
| `secret-hardcoded-key` | file, any | Common API-key/token literal shapes (AWS `AKIA...`, `sk-...`, generic `(api_key\|secret\|token)\s*=\s*["'][16+ chars]["']`) |
| `secret-private-key-block` | file, any | `-----BEGIN (RSA \|EC )?PRIVATE KEY-----` committed to a file |
| `shell-gcp-role-wildcard` | shell + `.sh` | `gcloud ... add-iam-policy-binding ... --role=roles/(owner\|editor)` |
| `shell-chmod-permissive` | shell + `.sh` | `chmod` with world-writable modes (`777`, `a+rwx`, etc.) |
| `go-ignored-error` | file, `.go` | `_ = err` discarding an error value |
| `swift-force-unwrap` | file, `.swift` | `try!` or `as!` |
| `python-shell-true` | file, `.py` | `subprocess.*shell=True` |
| `node-eval` | file, `.js`/`.ts` | `eval(` call |

Each check ships a `docs/checks/<id>.md`: what it flags, why it matters, how
to fix it, and one safe-vs-unsafe code example.

## Claude Code hook integration

Two hook subcommands, both reading Claude Code's hook JSON payload from stdin:

- **`aint hook pre-bash`** — wired to `PreToolUse` matching `Bash`. Extracts
  `tool_input.command`, runs shell-applicable checks against it. If any
  finding exists at all (regardless of `fail_on` — hooks run inline as
  Claude is about to act, so every finding should surface immediately;
  only `checks: <id>: off` silences a check here): prints findings to
  stderr and **exits 2**, which Claude Code interprets as blocking the
  command (stderr is shown to Claude as the reason).
- **`aint hook post-edit`** — wired to `PostToolUse` matching
  `Write|Edit|MultiEdit`. Extracts the written file path from the payload,
  runs the full file-check set against it. If findings exist: prints them to
  stderr and **exits 2** — the write already happened, but Claude Code
  surfaces the stderr to Claude as feedback for its next action.

`aint install [--global]`:

1. Targets `.claude/settings.json` by default, `~/.claude/settings.json`
   with `--global`; creates the file if it doesn't exist.
2. Reads existing JSON (if any) and merges in the two hook entries under
   `hooks.PreToolUse` / `hooks.PostToolUse`, checking for an existing entry
   whose `command` contains `aint hook` before appending — re-running
   `install` is idempotent and never duplicates entries.
3. Writes the file back with standard JSON formatting, leaving unrelated
   existing hooks/settings untouched.
4. Prints a summary of what was added vs. already present.

No uninstall command in v1.

## Testing

- **Per-check table tests**: true-positive and true-negative fixture
  snippets for every check (e.g. `roles/owner` flags, `roles/logging.viewer`
  doesn't; `try!` flags, `try?` doesn't).
- **Config tests**: load `.aint.yaml`, defaults when absent, severity
  override, `off` disables a check, ignore globs exclude paths.
- **Install tests**: merge into an empty settings.json, into one with
  unrelated existing hooks, and idempotency (running install twice produces
  identical output).
- **End-to-end smoke test**: fixture directory with one violation per
  language; run `aint check` against it and assert exit code + output shape.

No AST/tree-sitter dependency and no network calls anywhere in tests.
