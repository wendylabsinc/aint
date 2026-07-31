# Roadmap: next checks

The 8 seed checks prove the framework works end-to-end. This is the backlog for
expanding real coverage, organized by category and priority. `Fits today`
means it's a plain regex/line-based check like the seed set; `Needs parser`
means it really wants a real IaC/AST parser (Terraform HCL, Kubernetes YAML,
a real AST) to avoid excessive false positives/negatives, and should wait
for that capability rather than being forced into a regex.

Priority: **P0** = do next, **P1** = solid value, **P2** = nice to have / revisit later.

**Status (2026-07-31):** every "Fits today" check below across all six
categories was implemented in one batch (see
`docs/superpowers/plans/2026-07-31-aint-check-expansion.md`), plus the
claude-skills-sourced Swift checks further down this doc. `swift-print-statement`
absorbed claude-skills' `general.logger-usage` as one check rather than two.
Only `secret-dotenv-committed` and the `iac-*`/Concurrency items remain
deferred — see the "Deferred" section at the bottom for why.

## Secrets & credentials

| ID | Severity | Fits today? | Detects |
|---|---|---|---|
| `secret-generic-connection-string` | error | ✅ | `postgres://user:pass@host`, `mongodb+srv://user:pass@...` with a literal password |
| `secret-jwt-token` | warning | ✅ | A hardcoded JWT-shaped string (`eyJ...`) committed to source |
| `secret-slack-webhook` | error | ✅ | `hooks.slack.com/services/...` hardcoded |
| `secret-dotenv-committed` | error | ✅ | A `.env`/`.env.local`-style filename being scanned at all (path-based, not content) |
| `secret-high-entropy-string` | info | ⚠️ regex-approximate | Generic high-entropy literal in a suspicious assignment; higher false-positive risk than the others, ship as `info` by default |

**Priority:** P0 — `secret-generic-connection-string`, `secret-dotenv-committed` (highest signal, lowest effort). P1 — the rest.

## Shell / cloud IAM / IaC

| ID | Severity | Fits today? | Detects |
|---|---|---|---|
| `shell-aws-iam-wildcard` | error | ✅ | `aws iam ...` policy documents/CLI args with `"Action": "*"` or `"Resource": "*"` |
| `shell-aws-iam-attach-admin` | error | ✅ | `aws iam attach-*-policy --policy-arn ...AdministratorAccess` |
| `shell-azure-role-owner` | error | ✅ | `az role assignment create --role Owner` (or `Contributor` at broad scope) |
| `shell-curl-pipe-shell` | error | ✅ | `curl ... \| bash` / `curl ... \| sh` — supply-chain risk, no verification before execution |
| `shell-docker-privileged` | error | ✅ | `docker run --privileged` or mounting `/var/run/docker.sock` |
| `shell-gcp-service-account-key-download` | warning | ✅ | `gcloud iam service-accounts keys create` — long-lived key creation instead of workload identity |
| `shell-disable-host-firewall` | warning | ✅ | `setenforce 0`, `ufw disable`, `systemctl stop firewalld` |
| `iac-terraform-iam-wildcard` | error | 🔜 needs parser | `Action = "*"` / `Resource = "*"` inside an actual IAM policy block in `.tf` — regex on raw HCL risks false positives outside policy blocks |
| `iac-terraform-public-storage` | error | 🔜 needs parser | `acl = "public-read"` on an `aws_s3_bucket`/`google_storage_bucket` resource |
| `iac-k8s-privileged-pod` | error | 🔜 needs parser | `privileged: true`, `hostNetwork: true`, or host path mounts in a pod/deployment spec |

**Priority:** P0 — `shell-aws-iam-wildcard`, `shell-curl-pipe-shell`, `shell-docker-privileged` (mirror the existing GCP check's value directly). P1 — the rest of the shell-only rows. The `iac-*` rows are P2 and gated on adding a real Terraform/YAML parser — don't force them into the regex engine, they're exactly the false-positive trap the design doc's "Non-goals" section warns about.

## Go

| ID | Severity | Fits today? | Detects |
|---|---|---|---|
| `go-exec-command-injection` | error | ✅ | `exec.Command("sh", "-c", ...)` or similar built from string concatenation |
| `go-tls-insecure-skip-verify` | error | ✅ | `InsecureSkipVerify: true` |
| `go-weak-crypto-hash` | warning | ✅ | `md5.New()`/`sha1.New()` used in a hashing context that looks password/token-related |
| `go-http-client-no-timeout` | warning | ✅ | `http.Client{}` constructed with no `Timeout` field set on the same line/block |
| `go-unchecked-type-assertion` | warning | ⚠️ approximate | `x := y.(SomeType)` single-value form (no `, ok`) — regex can catch the common shape but will miss some |
| `go-sql-string-concat` | error | ⚠️ approximate | `fmt.Sprintf`/`+` building a SQL string passed to `Query`/`Exec` |

**Priority:** P0 — `go-exec-command-injection`, `go-tls-insecure-skip-verify` (direct security impact, clean regex fit). P1 — `go-weak-crypto-hash`, `go-http-client-no-timeout`. P2 — the two "approximate" ones; revisit once/if the engine grows a lightweight AST option, since regex false-positive rates here are meaningfully higher.

## Swift

| ID | Severity | Fits today? | Detects |
|---|---|---|---|
| `swift-print-statement` | info | ✅ | Leftover `print(...)` debug statements |
| `swift-unowned-reference` | warning | ✅ | `unowned` instead of `weak` — crashes instead of silently becoming nil |
| `swift-http-not-https` | warning | ✅ | A hardcoded `http://` URL literal (should be `https://`) |
| `swift-webview-arbitrary-load` | warning | ✅ | `.load(URLRequest(url: ...))` on a `WKWebView` fed from a non-literal/untrusted source |
| `swift-userdefaults-sensitive-key` | warning | ⚠️ approximate | `UserDefaults` key/value pair whose name looks like a credential (`password`, `token`, `secret`) — should live in Keychain instead |

**Priority:** P0 — `swift-unowned-reference`, `swift-http-not-https`. P1 — `swift-print-statement`, `swift-webview-arbitrary-load`. P2 — `swift-userdefaults-sensitive-key` (name-based heuristic, moderate false-positive risk).

## Python

| ID | Severity | Fits today? | Detects |
|---|---|---|---|
| `python-eval-exec` | error | ✅ | `eval(...)` / `exec(...)` call (mirrors `node-eval`) |
| `python-os-system` | error | ✅ | `os.system(...)` — same shell-injection class as `python-shell-true` |
| `python-pickle-load` | error | ✅ | `pickle.load(...)`/`pickle.loads(...)` on data that isn't obviously a trusted local file |
| `python-yaml-unsafe-load` | error | ✅ | `yaml.load(...)` without `Loader=yaml.SafeLoader` |
| `python-requests-no-verify` | error | ✅ | `requests.*(..., verify=False)` — disables TLS certificate verification |
| `python-weak-hash` | warning | ✅ | `hashlib.md5(...)`/`hashlib.sha1(...)` in a password/token-hashing context |
| `python-assert-security-check` | warning | ⚠️ approximate | `assert` used to gate an auth/permission check (stripped entirely under `-O`) |

**Priority:** P0 — `python-eval-exec`, `python-os-system`, `python-yaml-unsafe-load`, `python-requests-no-verify` (all direct, well-known vulnerability classes with clean regex fit — this is the single highest-value batch on the list). P1 — `python-pickle-load`, `python-weak-hash`. P2 — `python-assert-security-check`.

## Node.js / TypeScript

| ID | Severity | Fits today? | Detects |
|---|---|---|---|
| `node-child-process-exec` | error | ✅ | `child_process.exec(...)` (as opposed to `execFile`/`spawn` with an argv array) built from a template string |
| `node-tls-reject-unauthorized-false` | error | ✅ | `rejectUnauthorized: false` or `process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0'` |
| `node-insecure-random` | warning | ✅ | `Math.random()` used in a token/session/id-generation context |
| `node-jwt-alg-none` | error | ✅ | JWT verify/sign call allowing `algorithm: 'none'` |
| `node-console-log` | info | ✅ | Leftover `console.log(...)` (already referenced as the example in the original design doc's config sample) |
| `node-cors-wildcard-credentials` | error | ⚠️ approximate | `Access-Control-Allow-Origin: '*'` combined with `credentials: true` nearby |

**Priority:** P0 — `node-child-process-exec`, `node-tls-reject-unauthorized-false` (same class as the Go/Python entries above — this triangulates a "no shell-injection, no disabled TLS verification" baseline across all four languages). P1 — `node-jwt-alg-none`, `node-insecure-random`. P2 — `node-console-log` (low severity, mostly a hygiene check), `node-cors-wildcard-credentials` (proximity-based, needs care to avoid false positives).

## Suggested next batch (highest value per unit of effort)

If picking a single next PR's worth of checks, this set gives the broadest
new coverage with the lowest false-positive risk, one per language plus one
cross-cutting:

1. `python-eval-exec` + `python-os-system` + `python-yaml-unsafe-load` + `python-requests-no-verify`
2. `node-child-process-exec` + `node-tls-reject-unauthorized-false`
3. `go-exec-command-injection` + `go-tls-insecure-skip-verify`
4. `swift-unowned-reference` + `swift-http-not-https`
5. `shell-aws-iam-wildcard` + `shell-curl-pipe-shell` + `shell-docker-privileged`
6. `secret-generic-connection-string` + `secret-dotenv-committed`

That's 15 checks, doubling the current catalog, all regex-only, all in the
existing engine with no new dependencies — the same shape as the current
seed checks, just more of them.

## Claude-skills-sourced checks (Swift, added 2026-07-31)

`/Users/joannisorlandos/git/wendy/claude-skills` hosts `swift-server-lint`, an
AST-based Swift linter with ~50 rules across General, Concurrency,
Hummingbird, NIO, Postgres, and LibraryDesign categories. Only the
**General** category was mined for aint — it's broadly applicable to any
Swift codebase and regex-portable. The other five categories are specific to
a particular Swift server framework stack (NIO event loops, Hummingbird
routing, Postgres pooling, actor/Sendable concurrency) and need real AST
awareness aint's regex engine doesn't have.

| ID | Severity | Source rule | Detects |
|---|---|---|---|
| `swift-fatal-error` | warning | `general.fatal-error` | `fatalError(...)` calls |
| `swift-todo-comment` | info | `general.todo-comment` | `// TODO`/`FIXME`/`HACK` comments |
| `swift-implicitly-unwrapped-optional` | warning | `general.implicitly-unwrapped-optional` | `: Type!` declarations |

`general.logger-usage` (flags `print`/`debugPrint`/`dump`/`NSLog`) was merged
into the roadmap's own `swift-print-statement` rather than shipped as a
separate check — they detect the same thing.

**Concurrency category — deferred, not this batch.** Requested for a later
pass: actor reentrancy, Sendable closures, detached-task-in-actor, and
similar rules. These need real understanding of task/actor boundaries: a
regex approximation would be noisy and more likely to mislead than help.
Revisit once/if aint grows an AST option for Swift.

## Deferred until a real parser or engine capability exists

Don't force these into the regex engine — they're the exact false-positive
trap the design doc's "Non-goals" section warns about:

- `iac-terraform-iam-wildcard`, `iac-terraform-public-storage` — need real
  HCL parsing to know whether a `resource` block is actually an IAM
  policy/storage bucket, not just a string match anywhere in a `.tf` file.
- `iac-k8s-privileged-pod` — needs real YAML structure awareness (is
  `privileged: true` inside a `securityContext` under a container spec, or
  an unrelated field that happens to share a name?).
- `secret-dotenv-committed` — needs to match on **filename** (e.g. a file
  literally named `.env`), not file content. aint's `Check` type only
  supports content-regex matching today; this needs a `PathPattern` glob
  option added to the engine first, which is out of scope for a
  checks-only batch.
