# aint Check Catalog Expansion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expand aint's check catalog from 8 to 46 checks by adding 38 new regex-based checks across all six existing categories (secrets, shell, Go, Swift, Python, Node.js), plus docs pages for every new check.

**Architecture:** No engine changes. Every new check is a `check.Check{}` value appended to its existing category package file, registered via that package's existing `init()`. Same TDD pattern as the seed checks: a positive and negative table test per check.

**Tech Stack:** Go, `regexp` (RE2 syntax — no lookahead/lookbehind/backreferences).

## Global Constraints

- No AST/tree-sitter parsing — every check remains a line-based regex match, per the original design's non-goal.
- No changes to `internal/check`, `internal/config`, `internal/scan`, `internal/report`, or `cmd/aint` — this batch only adds `check.Check` values to the six existing `internal/checks/*` packages and their docs pages.
- Out of scope for this batch (see `docs/roadmap-checks.md`'s "Deferred" section): `secret-dotenv-committed` (needs filename-pattern matching, not content regex — a `Check` engine capability that doesn't exist yet), the three `iac-*` checks (need real HCL/YAML parsing), and the swift-server-lint Concurrency category (needs real actor/task-boundary AST awareness).
- Every new `Check.DocsPath` follows the existing convention: `<id>.md`, resolved via `docs/checks/<id>.md`.
- Spec/roadmap reference: `docs/roadmap-checks.md` (see its "Status (2026-07-31)" note and "Claude-skills-sourced checks" section for the full rationale behind this batch's scope).

---

### Task 1: Secrets checks batch

**Files:**
- Modify: `internal/checks/secrets/secrets.go`
- Modify: `internal/checks/secrets/secrets_test.go`

**Interfaces:**
- Consumes: `check.Check`, `check.Register`, `check.SeverityError`, `check.SeverityWarning`, `check.SeverityInfo` (existing, from Task 2 of the original framework plan).
- Produces: package-level vars `secrets.GenericConnectionString`, `secrets.JWTToken`, `secrets.SlackWebhook`, `secrets.HighEntropyString`, each registered via `init()`. Existing `secrets.HardcodedKey`/`secrets.PrivateKeyBlock` are untouched.

- [ ] **Step 1: Write the failing tests**

Add these test functions to the end of `internal/checks/secrets/secrets_test.go` (keep the existing 5 tests and imports unchanged):

```go
func TestGenericConnectionStringDetectsEmbeddedCredentials(t *testing.T) {
	findings := secrets.GenericConnectionString.Run("config.go", []byte(`dsn := "postgres://admin:s3cr3t@db.example.com:5432/mydb"`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestGenericConnectionStringIgnoresCredentiallessURL(t *testing.T) {
	findings := secrets.GenericConnectionString.Run("config.go", []byte(`dsn := "postgres://db.example.com:5432/mydb"`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestJWTTokenDetectsHardcodedToken(t *testing.T) {
	content := `token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"`
	findings := secrets.JWTToken.Run("config.go", []byte(content), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestJWTTokenIgnoresShortPrefix(t *testing.T) {
	findings := secrets.JWTToken.Run("config.go", []byte(`s := "eyJ short"`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestSlackWebhookDetectsHardcodedURL(t *testing.T) {
	content := `url := "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"`
	findings := secrets.SlackWebhook.Run("config.go", []byte(content), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestSlackWebhookIgnoresUnrelatedSlackURL(t *testing.T) {
	findings := secrets.SlackWebhook.Run("config.go", []byte(`url := "https://slack.com/api/chat.postMessage"`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestHighEntropyStringDetectsLongBase64Literal(t *testing.T) {
	content := `blob := "aGVsbG8gd29ybGQgdGhpcyBpcyBhIGxvbmcgYmFzZTY0IHN0cmluZw=="`
	findings := secrets.HighEntropyString.Run("config.go", []byte(content), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestHighEntropyStringIgnoresShortString(t *testing.T) {
	findings := secrets.HighEntropyString.Run("config.go", []byte(`s := "hello world"`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/checks/secrets/... -v`
Expected: FAIL — `secrets.GenericConnectionString`, `secrets.JWTToken`, `secrets.SlackWebhook`, `secrets.HighEntropyString` undefined.

- [ ] **Step 3: Implement**

Replace `internal/checks/secrets/secrets.go` in full with:

```go
// internal/checks/secrets/secrets.go
package secrets

import (
	"regexp"

	"aint/internal/check"
)

var HardcodedKey = check.Check{
	ID:       "secret-hardcoded-key",
	Title:    "Hardcoded API key or token",
	Severity: check.SeverityError,
	Pattern: regexp.MustCompile(
		`AKIA[0-9A-Z]{16}|sk-[a-zA-Z0-9]{20,}|(?i)(api_key|apikey|secret|token)\s*[:=]\s*["'][^"']{16,}["']`,
	),
	Message:  "possible hardcoded API key or token",
	DocsPath: "secret-hardcoded-key.md",
}

var PrivateKeyBlock = check.Check{
	ID:       "secret-private-key-block",
	Title:    "Committed private key material",
	Severity: check.SeverityError,
	Pattern:  regexp.MustCompile(`-----BEGIN (RSA |EC )?PRIVATE KEY-----`),
	Message:  "private key material committed to file",
	DocsPath: "secret-private-key-block.md",
}

var GenericConnectionString = check.Check{
	ID:       "secret-generic-connection-string",
	Title:    "Connection string with embedded credentials",
	Severity: check.SeverityError,
	Pattern:  regexp.MustCompile(`(postgres|postgresql|mysql|mongodb(\+srv)?):\/\/[^:\/\s]+:[^@\/\s]+@`),
	Message:  "connection string contains an embedded username/password",
	DocsPath: "secret-generic-connection-string.md",
}

var JWTToken = check.Check{
	ID:       "secret-jwt-token",
	Title:    "Hardcoded JWT token",
	Severity: check.SeverityWarning,
	Pattern:  regexp.MustCompile(`eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}`),
	Message:  "hardcoded JWT-shaped token found in source",
	DocsPath: "secret-jwt-token.md",
}

var SlackWebhook = check.Check{
	ID:       "secret-slack-webhook",
	Title:    "Hardcoded Slack webhook URL",
	Severity: check.SeverityError,
	Pattern:  regexp.MustCompile(`hooks\.slack\.com/services/[A-Za-z0-9/]+`),
	Message:  "hardcoded Slack webhook URL found in source",
	DocsPath: "secret-slack-webhook.md",
}

var HighEntropyString = check.Check{
	ID:       "secret-high-entropy-string",
	Title:    "High-entropy string literal",
	Severity: check.SeverityInfo,
	Pattern:  regexp.MustCompile(`["'][A-Za-z0-9+/]{40,}={0,2}["']`),
	Message:  "long, high-entropy-looking string literal — verify this isn't a leaked credential",
	DocsPath: "secret-high-entropy-string.md",
}

func init() {
	check.Register(HardcodedKey)
	check.Register(PrivateKeyBlock)
	check.Register(GenericConnectionString)
	check.Register(JWTToken)
	check.Register(SlackWebhook)
	check.Register(HighEntropyString)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/checks/secrets/... -v`
Expected: PASS (13 tests: 5 existing + 8 new).

- [ ] **Step 5: Commit**

```bash
git add internal/checks/secrets
git commit -m "add secret-generic-connection-string, secret-jwt-token, secret-slack-webhook, secret-high-entropy-string checks"
```

---

### Task 2: Shell / cloud IAM checks batch

**Files:**
- Modify: `internal/checks/shell/shell.go`
- Modify: `internal/checks/shell/shell_test.go`

**Interfaces:**
- Consumes: `check.Check`, `check.Register`, `check.SeverityError`, `check.SeverityWarning` (existing).
- Produces: package-level vars `shell.AWSIAMWildcard`, `shell.AWSIAMAttachAdmin`, `shell.AzureRoleOwner`, `shell.CurlPipeShell`, `shell.DockerPrivileged`, `shell.GCPServiceAccountKeyDownload`, `shell.DisableHostFirewall`, each registered via `init()`. Existing `shell.GCPRoleWildcard`/`shell.ChmodPermissive` are untouched.

- [ ] **Step 1: Write the failing tests**

Add these test functions to the end of `internal/checks/shell/shell_test.go`:

```go
func TestAWSIAMWildcardDetectsWildcardAction(t *testing.T) {
	cmd := `aws iam put-role-policy --policy-document '{"Statement":[{"Effect":"Allow","Action":"*","Resource":"*"}]}'`
	findings := shell.AWSIAMWildcard.Run("<command>", []byte(cmd), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestAWSIAMWildcardIgnoresScopedAction(t *testing.T) {
	cmd := `aws iam put-role-policy --policy-document '{"Action":"s3:GetObject"}'`
	findings := shell.AWSIAMWildcard.Run("<command>", []byte(cmd), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestAWSIAMAttachAdminDetectsAdministratorAccess(t *testing.T) {
	cmd := "aws iam attach-role-policy --role-name deploy --policy-arn arn:aws:iam::aws:policy/AdministratorAccess"
	findings := shell.AWSIAMAttachAdmin.Run("<command>", []byte(cmd), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestAWSIAMAttachAdminIgnoresScopedPolicy(t *testing.T) {
	cmd := "aws iam attach-role-policy --role-name deploy --policy-arn arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
	findings := shell.AWSIAMAttachAdmin.Run("<command>", []byte(cmd), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestAzureRoleOwnerDetectsOwnerGrant(t *testing.T) {
	cmd := "az role assignment create --assignee x@example.com --role Owner --scope /subscriptions/xxx"
	findings := shell.AzureRoleOwner.Run("<command>", []byte(cmd), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestAzureRoleOwnerIgnoresReaderRole(t *testing.T) {
	cmd := "az role assignment create --assignee x@example.com --role Reader --scope /subscriptions/xxx"
	findings := shell.AzureRoleOwner.Run("<command>", []byte(cmd), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestCurlPipeShellDetectsPipeToBash(t *testing.T) {
	cmd := "curl -sSL https://example.com/install.sh | bash"
	findings := shell.CurlPipeShell.Run("<command>", []byte(cmd), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestCurlPipeShellIgnoresDownloadToFile(t *testing.T) {
	cmd := "curl -sSL https://example.com/install.sh -o install.sh"
	findings := shell.CurlPipeShell.Run("<command>", []byte(cmd), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestDockerPrivilegedDetectsPrivilegedFlag(t *testing.T) {
	findings := shell.DockerPrivileged.Run("<command>", []byte("docker run --privileged -it myimage"), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestDockerPrivilegedDetectsSocketMount(t *testing.T) {
	findings := shell.DockerPrivileged.Run("<command>", []byte("docker run -v /var/run/docker.sock:/var/run/docker.sock myimage"), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestDockerPrivilegedIgnoresPlainRun(t *testing.T) {
	findings := shell.DockerPrivileged.Run("<command>", []byte("docker run -it myimage"), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestGCPServiceAccountKeyDownloadDetectsKeyCreate(t *testing.T) {
	cmd := "gcloud iam service-accounts keys create key.json --iam-account=sa@project.iam.gserviceaccount.com"
	findings := shell.GCPServiceAccountKeyDownload.Run("<command>", []byte(cmd), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestGCPServiceAccountKeyDownloadIgnoresList(t *testing.T) {
	findings := shell.GCPServiceAccountKeyDownload.Run("<command>", []byte("gcloud iam service-accounts list"), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestDisableHostFirewallDetectsSetenforce0(t *testing.T) {
	findings := shell.DisableHostFirewall.Run("<command>", []byte("setenforce 0"), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestDisableHostFirewallIgnoresSetenforce1(t *testing.T) {
	findings := shell.DisableHostFirewall.Run("<command>", []byte("setenforce 1"), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/checks/shell/... -v`
Expected: FAIL — new vars undefined.

- [ ] **Step 3: Implement**

Replace `internal/checks/shell/shell.go` in full with:

```go
// internal/checks/shell/shell.go
package shell

import (
	"regexp"

	"aint/internal/check"
)

var GCPRoleWildcard = check.Check{
	ID:       "shell-gcp-role-wildcard",
	Title:    "Overscoped GCP IAM role grant",
	Severity: check.SeverityError,
	Langs:    []string{"shell"},
	Pattern:  regexp.MustCompile(`add-iam-policy-binding.*--role[= ]roles/(owner|editor)\b`),
	Message:  "granting roles/owner or roles/editor is overscoped; use a narrower predefined or custom role",
	DocsPath: "shell-gcp-role-wildcard.md",
}

var ChmodPermissive = check.Check{
	ID:       "shell-chmod-permissive",
	Title:    "World-writable chmod",
	Severity: check.SeverityWarning,
	Langs:    []string{"shell"},
	Pattern:  regexp.MustCompile(`chmod\s+(-R\s+)?(777|a\+rwx|ugo\+rwx)\b`),
	Message:  "world-writable permissions granted via chmod",
	DocsPath: "shell-chmod-permissive.md",
}

var AWSIAMWildcard = check.Check{
	ID:       "shell-aws-iam-wildcard",
	Title:    "Wildcard AWS IAM policy",
	Severity: check.SeverityError,
	Langs:    []string{"shell"},
	Pattern:  regexp.MustCompile(`"(Action|Resource)"\s*:\s*"\*"`),
	Message:  "IAM policy grants wildcard Action or Resource; scope this down",
	DocsPath: "shell-aws-iam-wildcard.md",
}

var AWSIAMAttachAdmin = check.Check{
	ID:       "shell-aws-iam-attach-admin",
	Title:    "AdministratorAccess policy attached",
	Severity: check.SeverityError,
	Langs:    []string{"shell"},
	Pattern:  regexp.MustCompile(`attach-(user|role|group)-policy.*--policy-arn\s+\S*AdministratorAccess`),
	Message:  "attaching AdministratorAccess grants full account access; use a narrower managed or custom policy",
	DocsPath: "shell-aws-iam-attach-admin.md",
}

var AzureRoleOwner = check.Check{
	ID:       "shell-azure-role-owner",
	Title:    "Overscoped Azure Owner role grant",
	Severity: check.SeverityError,
	Langs:    []string{"shell"},
	Pattern:  regexp.MustCompile(`az\s+role\s+assignment\s+create.*--role[= ]["']?Owner["']?`),
	Message:  "granting the Owner role is overscoped; use a narrower built-in or custom role",
	DocsPath: "shell-azure-role-owner.md",
}

var CurlPipeShell = check.Check{
	ID:       "shell-curl-pipe-shell",
	Title:    "curl piped directly into a shell",
	Severity: check.SeverityError,
	Langs:    []string{"shell"},
	Pattern:  regexp.MustCompile(`curl\s+[^|]*\|\s*(sudo\s+)?(bash|sh)\b`),
	Message:  "piping curl output directly into a shell executes remote code with no verification",
	DocsPath: "shell-curl-pipe-shell.md",
}

var DockerPrivileged = check.Check{
	ID:       "shell-docker-privileged",
	Title:    "Privileged or docker-socket-mounted container",
	Severity: check.SeverityError,
	Langs:    []string{"shell"},
	Pattern:  regexp.MustCompile(`docker\s+run\b.*(--privileged\b|/var/run/docker\.sock)`),
	Message:  "running a container privileged or with the Docker socket mounted grants effective host root",
	DocsPath: "shell-docker-privileged.md",
}

var GCPServiceAccountKeyDownload = check.Check{
	ID:       "shell-gcp-service-account-key-download",
	Title:    "Long-lived GCP service account key creation",
	Severity: check.SeverityWarning,
	Langs:    []string{"shell"},
	Pattern:  regexp.MustCompile(`gcloud\s+iam\s+service-accounts\s+keys\s+create\b`),
	Message:  "creating a long-lived service account key; prefer workload identity federation where possible",
	DocsPath: "shell-gcp-service-account-key-download.md",
}

var DisableHostFirewall = check.Check{
	ID:       "shell-disable-host-firewall",
	Title:    "Host firewall/SELinux disabled",
	Severity: check.SeverityWarning,
	Langs:    []string{"shell"},
	Pattern:  regexp.MustCompile(`\bsetenforce\s+0\b|ufw\s+disable\b|systemctl\s+stop\s+firewalld\b`),
	Message:  "disabling SELinux/firewall reduces host defense in depth",
	DocsPath: "shell-disable-host-firewall.md",
}

func init() {
	check.Register(GCPRoleWildcard)
	check.Register(ChmodPermissive)
	check.Register(AWSIAMWildcard)
	check.Register(AWSIAMAttachAdmin)
	check.Register(AzureRoleOwner)
	check.Register(CurlPipeShell)
	check.Register(DockerPrivileged)
	check.Register(GCPServiceAccountKeyDownload)
	check.Register(DisableHostFirewall)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/checks/shell/... -v`
Expected: PASS (19 tests: 5 existing + 14 new).

- [ ] **Step 5: Commit**

```bash
git add internal/checks/shell
git commit -m "add AWS/Azure IAM, curl-pipe-shell, docker-privileged, GCP key download, and firewall-disable shell checks"
```

---

### Task 3: Go checks batch

**Files:**
- Modify: `internal/checks/golang/golang.go`
- Modify: `internal/checks/golang/golang_test.go`

**Interfaces:**
- Consumes: `check.Check`, `check.Register`, `check.SeverityError`, `check.SeverityWarning` (existing).
- Produces: package-level vars `golang.ExecCommandInjection`, `golang.TLSInsecureSkipVerify`, `golang.WeakCryptoHash`, `golang.HTTPClientNoTimeout`, `golang.UncheckedTypeAssertion`, `golang.SQLStringConcat`, each registered via `init()`. Existing `golang.IgnoredError` is untouched.

- [ ] **Step 1: Write the failing tests**

Add these test functions to the end of `internal/checks/golang/golang_test.go`:

```go
func TestExecCommandInjectionDetectsShellDashC(t *testing.T) {
	findings := golang.ExecCommandInjection.Run("main.go", []byte(`exec.Command("sh", "-c", cmd)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestExecCommandInjectionIgnoresDirectArgs(t *testing.T) {
	findings := golang.ExecCommandInjection.Run("main.go", []byte(`exec.Command("ls", "-la")`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestTLSInsecureSkipVerifyDetectsTrue(t *testing.T) {
	findings := golang.TLSInsecureSkipVerify.Run("main.go", []byte(`tls.Config{InsecureSkipVerify: true}`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestTLSInsecureSkipVerifyIgnoresFalse(t *testing.T) {
	findings := golang.TLSInsecureSkipVerify.Run("main.go", []byte(`tls.Config{InsecureSkipVerify: false}`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestWeakCryptoHashDetectsMD5(t *testing.T) {
	findings := golang.WeakCryptoHash.Run("main.go", []byte(`h := md5.New()`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestWeakCryptoHashIgnoresSHA256(t *testing.T) {
	findings := golang.WeakCryptoHash.Run("main.go", []byte(`h := sha256.New()`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestHTTPClientNoTimeoutDetectsEmptyStruct(t *testing.T) {
	findings := golang.HTTPClientNoTimeout.Run("main.go", []byte(`client := &http.Client{}`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestHTTPClientNoTimeoutIgnoresClientWithTimeout(t *testing.T) {
	findings := golang.HTTPClientNoTimeout.Run("main.go", []byte(`client := &http.Client{Timeout: 10 * time.Second}`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestUncheckedTypeAssertionDetectsSingleValueForm(t *testing.T) {
	findings := golang.UncheckedTypeAssertion.Run("main.go", []byte(`val := raw.(string)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestUncheckedTypeAssertionIgnoresCommaOkForm(t *testing.T) {
	findings := golang.UncheckedTypeAssertion.Run("main.go", []byte(`val, ok := raw.(string)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestSQLStringConcatDetectsSprintf(t *testing.T) {
	findings := golang.SQLStringConcat.Run("main.go", []byte(`db.Query(fmt.Sprintf("SELECT * FROM users WHERE id = %s", id))`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestSQLStringConcatIgnoresParameterizedQuery(t *testing.T) {
	findings := golang.SQLStringConcat.Run("main.go", []byte(`db.Query("SELECT * FROM users WHERE id = $1", id)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/checks/golang/... -v`
Expected: FAIL — new vars undefined.

- [ ] **Step 3: Implement**

Replace `internal/checks/golang/golang.go` in full with:

```go
// internal/checks/golang/golang.go
package golang

import (
	"regexp"

	"aint/internal/check"
)

var IgnoredError = check.Check{
	ID:       "go-ignored-error",
	Title:    "Ignored error return value",
	Severity: check.SeverityWarning,
	Langs:    []string{"go"},
	Pattern:  regexp.MustCompile(`\b_\s*=\s*err\b`),
	Message:  "error value discarded via `_ = err`",
	DocsPath: "go-ignored-error.md",
}

var ExecCommandInjection = check.Check{
	ID:       "go-exec-command-injection",
	Title:    "Shell spawned via exec.Command -c",
	Severity: check.SeverityError,
	Langs:    []string{"go"},
	Pattern:  regexp.MustCompile(`exec\.Command\(\s*"(sh|bash|/bin/sh|/bin/bash)"\s*,\s*"-c"`),
	Message:  "spawning a shell via exec.Command with -c risks command injection if any argument isn't a fixed literal",
	DocsPath: "go-exec-command-injection.md",
}

var TLSInsecureSkipVerify = check.Check{
	ID:       "go-tls-insecure-skip-verify",
	Title:    "TLS certificate verification disabled",
	Severity: check.SeverityError,
	Langs:    []string{"go"},
	Pattern:  regexp.MustCompile(`InsecureSkipVerify:\s*true`),
	Message:  "disabling TLS certificate verification allows man-in-the-middle attacks",
	DocsPath: "go-tls-insecure-skip-verify.md",
}

var WeakCryptoHash = check.Check{
	ID:       "go-weak-crypto-hash",
	Title:    "Weak cryptographic hash (MD5/SHA1)",
	Severity: check.SeverityWarning,
	Langs:    []string{"go"},
	Pattern:  regexp.MustCompile(`\b(md5|sha1)\.(New|Sum)\(`),
	Message:  "MD5/SHA1 are not safe for password hashing or security-sensitive checksums; use bcrypt/argon2 or SHA-256+",
	DocsPath: "go-weak-crypto-hash.md",
}

var HTTPClientNoTimeout = check.Check{
	ID:       "go-http-client-no-timeout",
	Title:    "http.Client created with no Timeout",
	Severity: check.SeverityWarning,
	Langs:    []string{"go"},
	Pattern:  regexp.MustCompile(`http\.Client\{\s*\}`),
	Message:  "http.Client{} has no Timeout set; a hung server can block this goroutine forever",
	DocsPath: "go-http-client-no-timeout.md",
}

var UncheckedTypeAssertion = check.Check{
	ID:       "go-unchecked-type-assertion",
	Title:    "Type assertion without comma-ok",
	Severity: check.SeverityWarning,
	Langs:    []string{"go"},
	Pattern:  regexp.MustCompile(`^\s*[A-Za-z_][A-Za-z0-9_]*\s*:=\s*[A-Za-z_][A-Za-z0-9_.]*\.\([A-Za-z_][A-Za-z0-9_.]*\)\s*$`),
	Message:  "type assertion without the comma-ok form panics if the assertion fails",
	DocsPath: "go-unchecked-type-assertion.md",
}

var SQLStringConcat = check.Check{
	ID:       "go-sql-string-concat",
	Title:    "SQL query built via string concatenation",
	Severity: check.SeverityError,
	Langs:    []string{"go"},
	Pattern:  regexp.MustCompile(`\.(Query|Exec|QueryRow)\w*\([^)]*(fmt\.Sprintf|"\s*\+|\+\s*")`),
	Message:  "building a SQL query via string concatenation/formatting risks SQL injection; use parameterized queries",
	DocsPath: "go-sql-string-concat.md",
}

func init() {
	check.Register(IgnoredError)
	check.Register(ExecCommandInjection)
	check.Register(TLSInsecureSkipVerify)
	check.Register(WeakCryptoHash)
	check.Register(HTTPClientNoTimeout)
	check.Register(UncheckedTypeAssertion)
	check.Register(SQLStringConcat)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/checks/golang/... -v`
Expected: PASS (14 tests: 2 existing + 12 new).

- [ ] **Step 5: Commit**

```bash
git add internal/checks/golang
git commit -m "add go-exec-command-injection, go-tls-insecure-skip-verify, go-weak-crypto-hash, go-http-client-no-timeout, go-unchecked-type-assertion, go-sql-string-concat checks"
```

---

### Task 4: Swift checks batch

**Files:**
- Modify: `internal/checks/swift/swift.go`
- Modify: `internal/checks/swift/swift_test.go`

**Interfaces:**
- Consumes: `check.Check`, `check.Register`, `check.SeverityWarning`, `check.SeverityInfo` (existing).
- Produces: package-level vars `swift.PrintStatement`, `swift.UnownedReference`, `swift.HTTPNotHTTPS`, `swift.WebViewArbitraryLoad`, `swift.UserDefaultsSensitiveKey`, `swift.FatalError`, `swift.TodoComment`, `swift.ImplicitlyUnwrappedOptional`, each registered via `init()`. Existing `swift.ForceUnwrap` is untouched. Note: `swift.PrintStatement` covers both the roadmap's "print-statement" check and claude-skills' `general.logger-usage` rule (both flag the same `print`/`debugPrint`/`dump`/`NSLog` calls) — do not create a separate logger-usage check.

- [ ] **Step 1: Write the failing tests**

Add these test functions to the end of `internal/checks/swift/swift_test.go`:

```go
func TestPrintStatementDetectsPrint(t *testing.T) {
	findings := swift.PrintStatement.Run("main.swift", []byte(`print("debug: \(value)")`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestPrintStatementDetectsNSLog(t *testing.T) {
	findings := swift.PrintStatement.Run("main.swift", []byte(`NSLog("debug: %@", value)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestPrintStatementIgnoresLogger(t *testing.T) {
	findings := swift.PrintStatement.Run("main.swift", []byte(`logger.info("debug: \(value)")`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestUnownedReferenceDetectsCaptureList(t *testing.T) {
	findings := swift.UnownedReference.Run("main.swift", []byte(`task = Task { [unowned self] in await self.run() }`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestUnownedReferenceDetectsPropertyDeclaration(t *testing.T) {
	findings := swift.UnownedReference.Run("main.swift", []byte(`unowned let delegate: SomeDelegate`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestUnownedReferenceIgnoresWeak(t *testing.T) {
	findings := swift.UnownedReference.Run("main.swift", []byte(`weak var delegate: SomeDelegate?`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestHTTPNotHTTPSDetectsPlainHTTP(t *testing.T) {
	findings := swift.HTTPNotHTTPS.Run("main.swift", []byte(`let url = "http://api.example.com/data"`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestHTTPNotHTTPSIgnoresHTTPS(t *testing.T) {
	findings := swift.HTTPNotHTTPS.Run("main.swift", []byte(`let url = "https://api.example.com/data"`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestWebViewArbitraryLoadDetectsURLRequestLoad(t *testing.T) {
	findings := swift.WebViewArbitraryLoad.Run("main.swift", []byte(`webView.load(URLRequest(url: someURL))`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestWebViewArbitraryLoadIgnoresLoadHTMLString(t *testing.T) {
	findings := swift.WebViewArbitraryLoad.Run("main.swift", []byte(`webView.loadHTMLString(staticHTML, baseURL: nil)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestUserDefaultsSensitiveKeyDetectsTokenKey(t *testing.T) {
	findings := swift.UserDefaultsSensitiveKey.Run("main.swift", []byte(`UserDefaults.standard.set(authToken, forKey: "userAuthToken")`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestUserDefaultsSensitiveKeyIgnoresUsernameKey(t *testing.T) {
	findings := swift.UserDefaultsSensitiveKey.Run("main.swift", []byte(`UserDefaults.standard.set(username, forKey: "username")`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestFatalErrorDetectsCall(t *testing.T) {
	findings := swift.FatalError.Run("main.swift", []byte(`fatalError("unreachable")`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestFatalErrorIgnoresPreconditionFailure(t *testing.T) {
	findings := swift.FatalError.Run("main.swift", []byte(`preconditionFailure("unreachable")`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestTodoCommentDetectsTODO(t *testing.T) {
	findings := swift.TodoComment.Run("main.swift", []byte(`// TODO: handle this edge case`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestTodoCommentIgnoresRegularComment(t *testing.T) {
	findings := swift.TodoComment.Run("main.swift", []byte(`// this function handles the edge case`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestImplicitlyUnwrappedOptionalDetectsDeclaration(t *testing.T) {
	findings := swift.ImplicitlyUnwrappedOptional.Run("main.swift", []byte(`var window: UIWindow!`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestImplicitlyUnwrappedOptionalIgnoresRegularOptional(t *testing.T) {
	findings := swift.ImplicitlyUnwrappedOptional.Run("main.swift", []byte(`var window: UIWindow?`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/checks/swift/... -v`
Expected: FAIL — new vars undefined.

- [ ] **Step 3: Implement**

Replace `internal/checks/swift/swift.go` in full with:

```go
// internal/checks/swift/swift.go
package swift

import (
	"regexp"

	"aint/internal/check"
)

var ForceUnwrap = check.Check{
	ID:       "swift-force-unwrap",
	Title:    "Force unwrap or force cast",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\b(try|as)!`),
	Message:  "force unwrap/cast can crash at runtime; prefer try?/as? with explicit handling",
	DocsPath: "swift-force-unwrap.md",
}

var PrintStatement = check.Check{
	ID:       "swift-print-statement",
	Title:    "print/debugPrint/NSLog instead of a structured logger",
	Severity: check.SeverityInfo,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\b(print|debugPrint|dump|NSLog)\s*\(`),
	Message:  "use swift-log's Logger instead of print/debugPrint/dump/NSLog for production output",
	DocsPath: "swift-print-statement.md",
}

var UnownedReference = check.Check{
	ID:       "swift-unowned-reference",
	Title:    "unowned reference",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\bunowned(\([^)]*\))?\s+(self|let|var)\b`),
	Message:  "unowned crashes if the referenced object has been deallocated; prefer weak with explicit nil-handling",
	DocsPath: "swift-unowned-reference.md",
}

var HTTPNotHTTPS = check.Check{
	ID:       "swift-http-not-https",
	Title:    "Hardcoded http:// URL",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`"http://[^"]*"`),
	Message:  "hardcoded http:// URL; use https:// to protect data in transit",
	DocsPath: "swift-http-not-https.md",
}

var WebViewArbitraryLoad = check.Check{
	ID:       "swift-webview-arbitrary-load",
	Title:    "WKWebView loading a URLRequest",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\.load\(\s*URLRequest\(url:`),
	Message:  "loading a URLRequest into a WKWebView; verify the URL comes from a trusted, validated source",
	DocsPath: "swift-webview-arbitrary-load.md",
}

var UserDefaultsSensitiveKey = check.Check{
	ID:       "swift-userdefaults-sensitive-key",
	Title:    "Sensitive-looking value stored in UserDefaults",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`(?i)UserDefaults\.\w+\.set\(.*forKey:\s*"[^"]*(password|token|secret|api[_-]?key)[^"]*"`),
	Message:  "storing sensitive-looking data in UserDefaults; use the Keychain instead",
	DocsPath: "swift-userdefaults-sensitive-key.md",
}

var FatalError = check.Check{
	ID:       "swift-fatal-error",
	Title:    "fatalError usage",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\bfatalError\s*\(`),
	Message:  "fatalError terminates the process; prefer throwing an error or precondition for programming invariants",
	DocsPath: "swift-fatal-error.md",
}

var TodoComment = check.Check{
	ID:       "swift-todo-comment",
	Title:    "TODO/FIXME/HACK comment",
	Severity: check.SeverityInfo,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`(?i)//.*\b(TODO|FIXME|HACK)\b`),
	Message:  "TODO/FIXME/HACK comment; track in an issue tracker and resolve before release",
	DocsPath: "swift-todo-comment.md",
}

var ImplicitlyUnwrappedOptional = check.Check{
	ID:       "swift-implicitly-unwrapped-optional",
	Title:    "Implicitly unwrapped optional",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`:\s*[A-Za-z_][A-Za-z0-9_.\[\]<>]*!`),
	Message:  "implicitly unwrapped optional (Type!) crashes if force-accessed while nil; use a regular optional",
	DocsPath: "swift-implicitly-unwrapped-optional.md",
}

func init() {
	check.Register(ForceUnwrap)
	check.Register(PrintStatement)
	check.Register(UnownedReference)
	check.Register(HTTPNotHTTPS)
	check.Register(WebViewArbitraryLoad)
	check.Register(UserDefaultsSensitiveKey)
	check.Register(FatalError)
	check.Register(TodoComment)
	check.Register(ImplicitlyUnwrappedOptional)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/checks/swift/... -v`
Expected: PASS (19 tests: 3 existing + 16 new).

- [ ] **Step 5: Commit**

```bash
git add internal/checks/swift
git commit -m "add print-statement, unowned-reference, http-not-https, webview-arbitrary-load, userdefaults-sensitive-key, fatal-error, todo-comment, implicitly-unwrapped-optional swift checks"
```

---

### Task 5: Python checks batch

**Files:**
- Modify: `internal/checks/python/python.go`
- Modify: `internal/checks/python/python_test.go`

**Interfaces:**
- Consumes: `check.Check`, `check.Register`, `check.SeverityError`, `check.SeverityWarning` (existing).
- Produces: package-level vars `python.EvalExec`, `python.OSSystem`, `python.PickleLoad`, `python.YAMLUnsafeLoad`, `python.RequestsNoVerify`, `python.WeakHash`, `python.AssertSecurityCheck`, each registered via `init()`. Existing `python.ShellTrue` is untouched.

- [ ] **Step 1: Write the failing tests**

Add these test functions to the end of `internal/checks/python/python_test.go`:

```go
func TestEvalExecDetectsEval(t *testing.T) {
	findings := python.EvalExec.Run("script.py", []byte(`eval(user_input)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestEvalExecDetectsExec(t *testing.T) {
	findings := python.EvalExec.Run("script.py", []byte(`exec(user_input)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestEvalExecIgnoresSimilarIdentifier(t *testing.T) {
	findings := python.EvalExec.Run("script.py", []byte(`evaluate(user_input)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestOSSystemDetectsCall(t *testing.T) {
	findings := python.OSSystem.Run("script.py", []byte(`os.system(cmd)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestOSSystemIgnoresUnrelatedOSCall(t *testing.T) {
	findings := python.OSSystem.Run("script.py", []byte(`os.path.join(a, b)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestPickleLoadDetectsLoads(t *testing.T) {
	findings := python.PickleLoad.Run("script.py", []byte(`data = pickle.loads(payload)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestPickleLoadIgnoresJSONLoads(t *testing.T) {
	findings := python.PickleLoad.Run("script.py", []byte(`data = json.loads(payload)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestYAMLUnsafeLoadDetectsLoad(t *testing.T) {
	findings := python.YAMLUnsafeLoad.Run("script.py", []byte(`config = yaml.load(f)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestYAMLUnsafeLoadIgnoresSafeLoad(t *testing.T) {
	findings := python.YAMLUnsafeLoad.Run("script.py", []byte(`config = yaml.safe_load(f)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestRequestsNoVerifyDetectsVerifyFalse(t *testing.T) {
	findings := python.RequestsNoVerify.Run("script.py", []byte(`requests.get(url, verify=False)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestRequestsNoVerifyIgnoresDefaultVerify(t *testing.T) {
	findings := python.RequestsNoVerify.Run("script.py", []byte(`requests.get(url)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestWeakHashDetectsMD5(t *testing.T) {
	findings := python.WeakHash.Run("script.py", []byte(`hashlib.md5(password.encode())`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestWeakHashIgnoresSHA256(t *testing.T) {
	findings := python.WeakHash.Run("script.py", []byte(`hashlib.sha256(password.encode())`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestAssertSecurityCheckDetectsIsAdmin(t *testing.T) {
	findings := python.AssertSecurityCheck.Run("script.py", []byte(`assert user.is_admin`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestAssertSecurityCheckIgnoresGenericAssert(t *testing.T) {
	findings := python.AssertSecurityCheck.Run("script.py", []byte(`assert len(items) > 0`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/checks/python/... -v`
Expected: FAIL — new vars undefined.

- [ ] **Step 3: Implement**

Replace `internal/checks/python/python.go` in full with:

```go
// internal/checks/python/python.go
package python

import (
	"regexp"

	"aint/internal/check"
)

var ShellTrue = check.Check{
	ID:       "python-shell-true",
	Title:    "subprocess call with shell=True",
	Severity: check.SeverityWarning,
	Langs:    []string{"python"},
	Pattern:  regexp.MustCompile(`subprocess\.\w+\([^)]*shell\s*=\s*True`),
	Message:  "shell=True risks shell injection if any part of the command is untrusted input",
	DocsPath: "python-shell-true.md",
}

var EvalExec = check.Check{
	ID:       "python-eval-exec",
	Title:    "Use of eval()/exec()",
	Severity: check.SeverityError,
	Langs:    []string{"python"},
	Pattern:  regexp.MustCompile(`\b(eval|exec)\s*\(`),
	Message:  "eval()/exec() executes arbitrary strings as code; avoid it or use a safer parser",
	DocsPath: "python-eval-exec.md",
}

var OSSystem = check.Check{
	ID:       "python-os-system",
	Title:    "Use of os.system()",
	Severity: check.SeverityError,
	Langs:    []string{"python"},
	Pattern:  regexp.MustCompile(`\bos\.system\s*\(`),
	Message:  "os.system() runs its argument through the shell; risks command injection with untrusted input",
	DocsPath: "python-os-system.md",
}

var PickleLoad = check.Check{
	ID:       "python-pickle-load",
	Title:    "Unpickling data",
	Severity: check.SeverityError,
	Langs:    []string{"python"},
	Pattern:  regexp.MustCompile(`\bpickle\.(load|loads)\s*\(`),
	Message:  "unpickling untrusted data can execute arbitrary code; use json or another safe serialization format",
	DocsPath: "python-pickle-load.md",
}

var YAMLUnsafeLoad = check.Check{
	ID:       "python-yaml-unsafe-load",
	Title:    "yaml.load() without a safe loader",
	Severity: check.SeverityError,
	Langs:    []string{"python"},
	Pattern:  regexp.MustCompile(`\byaml\.load\s*\(`),
	Message:  "yaml.load() can execute arbitrary code via !!python/object tags unless Loader=yaml.SafeLoader; prefer yaml.safe_load()",
	DocsPath: "python-yaml-unsafe-load.md",
}

var RequestsNoVerify = check.Check{
	ID:       "python-requests-no-verify",
	Title:    "requests call with verify=False",
	Severity: check.SeverityError,
	Langs:    []string{"python"},
	Pattern:  regexp.MustCompile(`requests\.\w+\([^)]*verify\s*=\s*False`),
	Message:  "verify=False disables TLS certificate verification, allowing man-in-the-middle attacks",
	DocsPath: "python-requests-no-verify.md",
}

var WeakHash = check.Check{
	ID:       "python-weak-hash",
	Title:    "Weak cryptographic hash (MD5/SHA1)",
	Severity: check.SeverityWarning,
	Langs:    []string{"python"},
	Pattern:  regexp.MustCompile(`hashlib\.(md5|sha1)\s*\(`),
	Message:  "MD5/SHA1 are not safe for password hashing; use a dedicated password-hashing library (bcrypt, argon2)",
	DocsPath: "python-weak-hash.md",
}

var AssertSecurityCheck = check.Check{
	ID:       "python-assert-security-check",
	Title:    "assert used for an authorization check",
	Severity: check.SeverityWarning,
	Langs:    []string{"python"},
	Pattern:  regexp.MustCompile(`(?i)assert\s+.*(is_admin|authenticated|authorized|has_permission)`),
	Message:  "assert is stripped when Python runs with -O; don't use it to gate authorization/permission checks",
	DocsPath: "python-assert-security-check.md",
}

func init() {
	check.Register(ShellTrue)
	check.Register(EvalExec)
	check.Register(OSSystem)
	check.Register(PickleLoad)
	check.Register(YAMLUnsafeLoad)
	check.Register(RequestsNoVerify)
	check.Register(WeakHash)
	check.Register(AssertSecurityCheck)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/checks/python/... -v`
Expected: PASS (16 tests: 2 existing + 14 new).

- [ ] **Step 5: Commit**

```bash
git add internal/checks/python
git commit -m "add python-eval-exec, python-os-system, python-pickle-load, python-yaml-unsafe-load, python-requests-no-verify, python-weak-hash, python-assert-security-check checks"
```

---

### Task 6: Node.js checks batch

**Files:**
- Modify: `internal/checks/nodejs/nodejs.go`
- Modify: `internal/checks/nodejs/nodejs_test.go`

**Interfaces:**
- Consumes: `check.Check`, `check.Register`, `check.SeverityError`, `check.SeverityWarning`, `check.SeverityInfo` (existing).
- Produces: package-level vars `nodejs.ChildProcessExec`, `nodejs.TLSRejectUnauthorizedFalse`, `nodejs.InsecureRandom`, `nodejs.JWTAlgNone`, `nodejs.ConsoleLog`, `nodejs.CORSWildcardCredentials`, each registered via `init()`. Existing `nodejs.Eval` is untouched.

- [ ] **Step 1: Write the failing tests**

Add these test functions to the end of `internal/checks/nodejs/nodejs_test.go`:

```go
func TestChildProcessExecDetectsCall(t *testing.T) {
	findings := nodejs.ChildProcessExec.Run("index.js", []byte(`child_process.exec(cmd)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestChildProcessExecIgnoresExecFile(t *testing.T) {
	findings := nodejs.ChildProcessExec.Run("index.js", []byte(`child_process.execFile(cmd, args)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestTLSRejectUnauthorizedFalseDetectsOption(t *testing.T) {
	findings := nodejs.TLSRejectUnauthorizedFalse.Run("index.js", []byte(`https.request({ rejectUnauthorized: false })`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestTLSRejectUnauthorizedFalseDetectsEnvVar(t *testing.T) {
	findings := nodejs.TLSRejectUnauthorizedFalse.Run("index.js", []byte(`process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0'`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestTLSRejectUnauthorizedFalseIgnoresTrue(t *testing.T) {
	findings := nodejs.TLSRejectUnauthorizedFalse.Run("index.js", []byte(`https.request({ rejectUnauthorized: true })`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestInsecureRandomDetectsSessionToken(t *testing.T) {
	findings := nodejs.InsecureRandom.Run("index.js", []byte(`const sessionToken = Math.random().toString(36)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestInsecureRandomIgnoresUnrelatedRandom(t *testing.T) {
	findings := nodejs.InsecureRandom.Run("index.js", []byte(`const jitter = Math.random() * 100`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestJWTAlgNoneDetectsNoneAlgorithm(t *testing.T) {
	findings := nodejs.JWTAlgNone.Run("index.js", []byte(`jwt.verify(token, secret, { algorithms: ['none'] })`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestJWTAlgNoneIgnoresHS256(t *testing.T) {
	findings := nodejs.JWTAlgNone.Run("index.js", []byte(`jwt.verify(token, secret, { algorithms: ['HS256'] })`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestConsoleLogDetectsCall(t *testing.T) {
	findings := nodejs.ConsoleLog.Run("index.js", []byte(`console.log(debugValue)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestConsoleLogIgnoresLogger(t *testing.T) {
	findings := nodejs.ConsoleLog.Run("index.js", []byte(`logger.info(debugValue)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestCORSWildcardCredentialsDetectsWildcardWithCredentials(t *testing.T) {
	findings := nodejs.CORSWildcardCredentials.Run("index.js", []byte(`cors({ origin: '*', credentials: true })`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestCORSWildcardCredentialsIgnoresScopedOrigin(t *testing.T) {
	findings := nodejs.CORSWildcardCredentials.Run("index.js", []byte(`cors({ origin: 'https://example.com', credentials: true })`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/checks/nodejs/... -v`
Expected: FAIL — new vars undefined.

- [ ] **Step 3: Implement**

Replace `internal/checks/nodejs/nodejs.go` in full with:

```go
// internal/checks/nodejs/nodejs.go
package nodejs

import (
	"regexp"

	"aint/internal/check"
)

var Eval = check.Check{
	ID:       "node-eval",
	Title:    "Use of eval()",
	Severity: check.SeverityWarning,
	Langs:    []string{"nodejs"},
	Pattern:  regexp.MustCompile(`\beval\s*\(`),
	Message:  "eval() executes arbitrary strings as code; avoid it or use a safer parser",
	DocsPath: "node-eval.md",
}

var ChildProcessExec = check.Check{
	ID:       "node-child-process-exec",
	Title:    "child_process.exec() usage",
	Severity: check.SeverityError,
	Langs:    []string{"nodejs"},
	Pattern:  regexp.MustCompile(`child_process\.exec\s*\(`),
	Message:  "child_process.exec runs its argument through a shell; use execFile/spawn with an argv array instead",
	DocsPath: "node-child-process-exec.md",
}

var TLSRejectUnauthorizedFalse = check.Check{
	ID:       "node-tls-reject-unauthorized-false",
	Title:    "TLS certificate verification disabled",
	Severity: check.SeverityError,
	Langs:    []string{"nodejs"},
	Pattern:  regexp.MustCompile(`rejectUnauthorized\s*:\s*false|NODE_TLS_REJECT_UNAUTHORIZED\s*=\s*['"]?0`),
	Message:  "disabling TLS certificate verification allows man-in-the-middle attacks",
	DocsPath: "node-tls-reject-unauthorized-false.md",
}

var InsecureRandom = check.Check{
	ID:       "node-insecure-random",
	Title:    "Math.random() used for a security-sensitive value",
	Severity: check.SeverityWarning,
	Langs:    []string{"nodejs"},
	Pattern:  regexp.MustCompile(`(?i)(token|session|secret|password|otp)\w*\s*=\s*.*Math\.random\(\)`),
	Message:  "Math.random() is not cryptographically secure; use crypto.randomBytes() for tokens/sessions/secrets",
	DocsPath: "node-insecure-random.md",
}

var JWTAlgNone = check.Check{
	ID:       "node-jwt-alg-none",
	Title:    "JWT 'none' algorithm allowed",
	Severity: check.SeverityError,
	Langs:    []string{"nodejs"},
	Pattern:  regexp.MustCompile(`algorithms?\s*:\s*\[?\s*['"]none['"]`),
	Message:  "allowing the 'none' JWT algorithm lets an attacker forge unsigned tokens",
	DocsPath: "node-jwt-alg-none.md",
}

var ConsoleLog = check.Check{
	ID:       "node-console-log",
	Title:    "Leftover console.log",
	Severity: check.SeverityInfo,
	Langs:    []string{"nodejs"},
	Pattern:  regexp.MustCompile(`\bconsole\.log\s*\(`),
	Message:  "leftover console.log; use a structured logger for production output",
	DocsPath: "node-console-log.md",
}

var CORSWildcardCredentials = check.Check{
	ID:       "node-cors-wildcard-credentials",
	Title:    "Wildcard CORS origin with credentials",
	Severity: check.SeverityError,
	Langs:    []string{"nodejs"},
	Pattern:  regexp.MustCompile(`origin:\s*['"]\*['"].*credentials:\s*true|credentials:\s*true.*origin:\s*['"]\*['"]`),
	Message:  "wildcard CORS origin combined with credentials:true lets any site read authenticated responses",
	DocsPath: "node-cors-wildcard-credentials.md",
}

func init() {
	check.Register(Eval)
	check.Register(ChildProcessExec)
	check.Register(TLSRejectUnauthorizedFalse)
	check.Register(InsecureRandom)
	check.Register(JWTAlgNone)
	check.Register(ConsoleLog)
	check.Register(CORSWildcardCredentials)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/checks/nodejs/... -v`
Expected: PASS (14 tests: 2 existing + 12 new).

- [ ] **Step 5: Commit**

```bash
git add internal/checks/nodejs
git commit -m "add node-child-process-exec, node-tls-reject-unauthorized-false, node-insecure-random, node-jwt-alg-none, node-console-log, node-cors-wildcard-credentials checks"
```

---

### Task 7: Docs pages for all 38 new checks, README/roadmap updates

**Files:**
- Create: 38 files under `docs/checks/` — one per new check ID from Tasks 1-6 (full list below).
- Modify: `README.md`
- Modify: `docs/roadmap-checks.md`

**Interfaces:**
- None — pure documentation, no code. Depends on Tasks 1-6 being committed (their exact `Message`/`Pattern`/`Severity`/`Title` values are the source of truth for each doc page's content).

**The 38 doc pages to create** (grouped by task for reference — see that task's implementation code above for each check's exact `Message`, `Pattern`, and `Severity`):

- From Task 1 (secrets): `secret-generic-connection-string.md`, `secret-jwt-token.md`, `secret-slack-webhook.md`, `secret-high-entropy-string.md`
- From Task 2 (shell): `shell-aws-iam-wildcard.md`, `shell-aws-iam-attach-admin.md`, `shell-azure-role-owner.md`, `shell-curl-pipe-shell.md`, `shell-docker-privileged.md`, `shell-gcp-service-account-key-download.md`, `shell-disable-host-firewall.md`
- From Task 3 (go): `go-exec-command-injection.md`, `go-tls-insecure-skip-verify.md`, `go-weak-crypto-hash.md`, `go-http-client-no-timeout.md`, `go-unchecked-type-assertion.md`, `go-sql-string-concat.md`
- From Task 4 (swift): `swift-print-statement.md`, `swift-unowned-reference.md`, `swift-http-not-https.md`, `swift-webview-arbitrary-load.md`, `swift-userdefaults-sensitive-key.md`, `swift-fatal-error.md`, `swift-todo-comment.md`, `swift-implicitly-unwrapped-optional.md`
- From Task 5 (python): `python-eval-exec.md`, `python-os-system.md`, `python-pickle-load.md`, `python-yaml-unsafe-load.md`, `python-requests-no-verify.md`, `python-weak-hash.md`, `python-assert-security-check.md`
- From Task 6 (nodejs): `node-child-process-exec.md`, `node-tls-reject-unauthorized-false.md`, `node-insecure-random.md`, `node-jwt-alg-none.md`, `node-console-log.md`, `node-cors-wildcard-credentials.md`

- [ ] **Step 1: Write each doc page**

Every page follows the same four-section shape used by the 8 existing pages in `docs/checks/`: what it flags, why it matters, how to fix it, one flagged-vs-fixed code example. Read the corresponding check's `Message`, `Pattern`, and `Title` from Tasks 1-6 above to ground the content — don't write generic filler. Two full examples (repeat this shape for the other 36, substituting the real detection/fix/example each check's `Message`/`Pattern` already gives you):

```markdown
<!-- docs/checks/python-yaml-unsafe-load.md -->
# python-yaml-unsafe-load

**Flags:** a call to `yaml.load(...)`.

**Why it matters:** `yaml.load()` can deserialize arbitrary Python objects via
`!!python/object` tags, which can execute arbitrary code when the loaded
document isn't trusted. Note: this check flags every `yaml.load(...)` call,
including ones that already pass `Loader=yaml.SafeLoader` — aint's regex
engine can't see keyword arguments reliably, so a call that's actually safe
may still be flagged. Prefer `yaml.safe_load()` outright, which sidesteps
the ambiguity entirely.

**Fix:** use `yaml.safe_load()`, which only ever constructs plain Python
types (dicts, lists, strings, numbers).

```python
# Flags this:
config = yaml.load(f)

# Prefer this:
config = yaml.safe_load(f)
```
```

```markdown
<!-- docs/checks/shell-curl-pipe-shell.md -->
# shell-curl-pipe-shell

**Flags:** `curl ... | bash` or `curl ... | sh` (with an optional `sudo`
before the shell).

**Why it matters:** piping a remote script straight into a shell executes
it with zero verification — no checksum, no signature, no review of what
actually ran. If the remote host is compromised, or the connection is
intercepted, this runs arbitrary code with whatever privileges the shell
has.

**Fix:** download the script first, review it (or verify a checksum/
signature), then execute it explicitly.

```bash
# Flags this:
curl -sSL https://example.com/install.sh | bash

# Prefer this:
curl -sSL https://example.com/install.sh -o install.sh
sha256sum install.sh   # verify against a published checksum
bash install.sh
```
```

Write the remaining 36 pages the same way, one per ID listed above, using each check's real `Message` and a realistic example derived from its test cases in Tasks 1-6.

- [ ] **Step 2: Update the README's check section**

In `README.md`, update the check table to note the catalog has grown and point to `aint list` and `docs/checks/` for the full current set (the table doesn't need to enumerate all 46 checks — keep the existing 8-row table as illustrative examples and add a line noting the total count and where to find the rest):

Find this line in `README.md`:
```markdown
Run `aint list` for the live, current set — it's designed to keep growing. See `docs/checks/<id>.md` for the full explanation and fix for each one.
```

Replace it with:
```markdown
This table shows a representative sample — run `aint list` for the full, current set (46 checks as of this batch, across secrets, shell/cloud IAM, Go, Swift, Python, and Node.js). See `docs/checks/<id>.md` for the full explanation and fix for each one.
```

- [ ] **Step 3: Update the roadmap doc's status line**

In `docs/roadmap-checks.md`, the "Status (2026-07-31)" note already describes this batch prospectively. Update it to reflect completion — change:

```markdown
**Status (2026-07-31):** every "Fits today" check below across all six
categories was implemented in one batch (see
```

to:

```markdown
**Status (2026-07-31, complete):** every "Fits today" check below across all six
categories was implemented in one batch (see
```

- [ ] **Step 4: Verify the whole build and check count**

Run: `go build ./... && go test ./... -count=1`
Expected: PASS across every package.

Run: `go run ./cmd/aint list | wc -l`
Expected: `46`

Run: `ls docs/checks/*.md | wc -l`
Expected: `46`

- [ ] **Step 5: Commit**

```bash
git add docs/checks README.md docs/roadmap-checks.md
git commit -m "add docs pages for all 38 new checks, update README and roadmap status"
```

---

## Plan Self-Review Notes

- **Spec coverage:** every "Fits today" row in `docs/roadmap-checks.md`'s six category tables (minus `secret-dotenv-committed`, explicitly deferred) maps to a check in Tasks 1-6; the claude-skills General-category rules map to `swift-fatal-error`/`swift-todo-comment`/`swift-implicitly-unwrapped-optional` in Task 4, with `general.logger-usage` merged into `swift-print-statement` rather than duplicated. Docs coverage (Task 7) accounts for all 38 new checks by exact ID.
- **Type consistency verified:** every new check uses the existing `check.Check{ID, Title, Severity, Langs, Pattern, Message, DocsPath}` struct and `check.Register`/`check.SeverityError`/`check.SeverityWarning`/`check.SeverityInfo` exactly as defined in the original framework plan — no new fields, no engine changes.
- **No placeholders:** every check has a complete, tested regex and message; Task 7's two full doc-page examples plus the exact per-check `Message`/`Pattern` values from Tasks 1-6 give the doc-writer everything needed for the remaining 36 pages with zero ambiguity.
- **RE2 constraint check:** every pattern uses only RE2-supported syntax (no lookahead/lookbehind/backreferences) — verified by hand-tracing each positive/negative test case against its regex during design.
