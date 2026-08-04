# AINT - Make AI more deterministic

`aint` is a fast, single-binary static analyzer that catches the mistakes that actually hurt: hardcoded secrets, overscoped cloud permissions, and language-specific footguns - in your code *and* in the shell commands run against it. It's built to sit directly in Claude Code's hook pipeline, so an agent gets caught **before** it ships a `roles/owner` grant or commits an API key, not after.

![](logo.jpeg)

```
$ aint check .
internal/api/client.go:14:9: error [secret-hardcoded-key] possible hardcoded API key or token - docs/checks/secret-hardcoded-key.md
internal/api/client.go:42:2: warning [go-ignored-error] error value discarded via `_ = err` - docs/checks/go-ignored-error.md
scripts/deploy.sh:8:1: error [shell-gcp-role-wildcard] granting roles/owner or roles/editor is overscoped; use a narrower predefined or custom role - docs/checks/shell-gcp-role-wildcard.md

exit code: 1
```

## Why

Agentic coding tools are extremely good at writing code that *runs*. They're not always good at writing code that's *safe* - a plausible-looking `gcloud` command with `--role=roles/owner`, a `subprocess.run(cmd, shell=True)` that looked fine in the sandbox, a `console.log` that ships to prod with an API key in it. Nobody's watching every line an agent generates.

`aint` is that watcher. It's:

- **Fast** - a single static Go binary, no runtime, no network calls, scans a repo in milliseconds. (The one exception is `aint improve`, an opt-in subcommand that shells out to the local `claude` CLI and does make network calls on your behalf.)
- **Cross-language** - Go, Swift, Python, Node.js today; the check catalog is designed to grow.
- **Shell-aware** - it treats a raw shell command the same as a file, so it can statically verify IaC and cloud CLI invocations before they run, not just source code.
- **Configurable** - every check can be silenced or have its severity tuned per-project via `.aint.yaml`.
- **Actionable** - every finding is `file:line:col`, a one-line explanation, and a link to a doc page with the real fix.

## Quick start

```bash
go build -o aint ./cmd/aint     # or: make build
./aint check .                  # scan the current directory
./aint list                     # see every registered check
```

Wire it into Claude Code so it runs automatically:

```bash
./aint install                  # writes hooks into .claude/settings.json
```

From then on:
- Every `Bash` command Claude Code is about to run is checked first - an overscoped IAM grant or a `chmod 777` gets **blocked outright**, with the reason shown back to Claude.
- Every file Claude writes or edits is checked right after - findings come back as feedback Claude can act on in its next turn.

## What it catches today

| Check | Severity | Applies to | Catches |
|---|---|---|---|
| `secret-hardcoded-key` | error | any file | AWS-style keys, `sk-…` tokens, `api_key = "…"`-shaped assignments |
| `secret-private-key-block` | error | any file | A `-----BEGIN PRIVATE KEY-----` block committed to a file |
| `shell-gcp-role-wildcard` | error | shell / `.sh` | `gcloud … add-iam-policy-binding … --role=roles/owner\|editor` |
| `shell-chmod-permissive` | warning | shell / `.sh` | `chmod 777` / `a+rwx` / `ugo+rwx` |
| `go-ignored-error` | warning | `.go` | `_ = err` silently discarding an error |
| `swift-force-unwrap` | warning | `.swift` | `try!` / `as!` |
| `python-shell-true` | warning | `.py` | `subprocess.*(shell=True)` |
| `node-eval` | warning | `.js` / `.ts` | `eval(…)` |

This table shows a representative sample - run `aint list` for the full, current set (46 checks as of this batch, across secrets, shell/cloud IAM, Go, Swift, Python, and Node.js). See `docs/checks/<id>.md` for the full explanation and fix for each one.

## Configuration

Drop a `.aint.yaml` at your repo root to tune behavior - see the example at the repo root for the full option set:

```yaml
fail_on: error          # info | warning | error - findings at/above this fail `aint check`
ignore:
  - vendor/**
  - node_modules/**
checks:
  go-ignored-error: error   # bump a check's severity
  node-eval: off             # or turn it off entirely
```

**One important distinction:** `aint check`'s exit code is gated by `fail_on` (for CI). Claude Code hook mode is stricter by design - it surfaces *every* finding regardless of `fail_on`, since it's catching things inline, before they land. Only `checks: <id>: off` silences a check in hook mode.

## Roadmap

- Deeper coverage per language (this is a seed set, not the ceiling)
- Real IaC parsing (Terraform HCL, Kubernetes manifests) beyond today's regex-based shell checks
- Per-repo baseline/allowlisting for gradual adoption on existing codebases

## Commands

```
aint check [paths...]           # scan files/dirs (default: .)
aint check --format=json ...    # machine-readable output
aint list                       # list all registered checks
aint install [--global]         # wire aint into Claude Code hooks
aint improve                    # mine ~/.claude/projects for incidents, report suggested aint/lint rules + doc fixes
                                 # flags: --dir --out --state --claude-bin --limit --full
                                 # shells out to your installed `claude` CLI to analyze flagged excerpts -
                                 # this sends conversation text to Anthropic and costs tokens/time
```
