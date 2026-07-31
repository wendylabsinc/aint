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
