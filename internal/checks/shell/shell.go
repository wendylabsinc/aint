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

var GrepSearchCodebase = check.Check{
	ID:       "shell-grep-search-codebase",
	Title:    "grep/rg used to search a codebase from the shell",
	Severity: check.SeverityWarning,
	Langs:    []string{"shell"},
	Pattern:  regexp.MustCompile(`(^|[|;&]\s*)(rg\s|grep\s+(-\w*[rRl]\w*\b|--include\b))`),
	Message:  "use the Grep/Read tool instead of shelling out to grep/rg for codebase search - when you already know the file, Read it directly",
	DocsPath: "shell-grep-search-codebase.md",
}

var GitAddBroad = check.Check{
	ID:       "shell-git-add-broad",
	Title:    "Broad git add (-A/--all/-u/.)",
	Severity: check.SeverityWarning,
	Langs:    []string{"shell"},
	Pattern:  regexp.MustCompile(`\bgit\s+add\s+(-A\b|--all\b|-u\b|\.(\s|$))`),
	Message:  "a broad `git add` can sweep pre-existing uncommitted WIP into this commit - check `git status` first and stage only the files this task actually changed",
	DocsPath: "shell-git-add-broad.md",
}

// wendyos boundary excludes sibling repos/worktrees like wendyos-builder or
// wendyos-update, which are separate trees, not the shared main checkout.
var GitCheckoutSharedWendyosTree = check.Check{
	ID:       "shell-git-checkout-shared-wendyos-tree",
	Title:    "Branch switch in the shared wendyos main tree",
	Severity: check.SeverityError,
	Langs:    []string{"shell"},
	Pattern: regexp.MustCompile(
		`(?:(?:^|[\s/"'])wendyos(?:[\s/"'&|;]|$).{0,200}\b(?:checkout|switch)\b)` +
			`|(?:\b(?:checkout|switch)\b.{0,200}(?:^|[\s/"'])wendyos(?:[\s/"'&|;]|$))`,
	),
	Message:  "the wendyos main tree is shared by concurrent Claude sessions - switching branches here can redirect another session's commits; do multi-commit work in an isolated git worktree instead",
	DocsPath: "shell-git-checkout-shared-wendyos-tree.md",
}

var PsqlInlineSQLVariable = check.Check{
	ID:       "shell-psql-inline-sql-variable",
	Title:    "psql -c with shell-interpolated SQL",
	Severity: check.SeverityError,
	Langs:    []string{"shell"},
	Pattern:  regexp.MustCompile(`psql\b[^\n]*-c\s+["'][^"']*\$`),
	Message:  "building `psql -c` SQL by shell-interpolating a variable is defeated by comment/quote smuggling; enforce read-only (or any other guarantee) at the protocol level with a real client library, not string concatenation",
	DocsPath: "shell-psql-inline-sql-variable.md",
}

var ContainerCachePurgeForce = check.Check{
	ID:       "shell-container-cache-purge-force",
	Title:    "Forced container build-cache purge",
	Severity: check.SeverityWarning,
	Langs:    []string{"shell"},
	Pattern:  regexp.MustCompile(`\b(container\s+builder\s+delete|docker\s+(builder|image|container)\s+prune)\b[^\n]*(--force\b|-\w*f\w*\b)`),
	Message:  "this force-purges a container build cache with no confirmation; the next build runs cold. Confirm with the user before running it",
	DocsPath: "shell-container-cache-purge-force.md",
}

var YoctoTmpdirSymlink = check.Check{
	ID:       "shell-yocto-tmpdir-symlink",
	Title:    "Symlinking Yocto's build/tmp (TMPDIR)",
	Severity: check.SeverityError,
	Langs:    []string{"shell"},
	Pattern:  regexp.MustCompile(`\bln\s+-s\S*\b[^\n]*\b(build/tmp|TMPDIR)\b`),
	Message:  "Yocto's pseudo (fakeroot) does not tolerate TMPDIR/build/tmp behind a symlink - do_rootfs postinst scriptlets fail (surfaces as an unrelated-looking rpcbind failure). Bind-mount the scratch onto build/tmp instead",
	DocsPath: "shell-yocto-tmpdir-symlink.md",
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
	check.Register(GrepSearchCodebase)
	check.Register(GitAddBroad)
	check.Register(GitCheckoutSharedWendyosTree)
	check.Register(PsqlInlineSQLVariable)
	check.Register(ContainerCachePurgeForce)
	check.Register(YoctoTmpdirSymlink)
}
