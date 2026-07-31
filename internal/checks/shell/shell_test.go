// internal/checks/shell/shell_test.go
package shell_test

import (
	"testing"

	"aint/internal/checks/shell"
)

func TestGCPRoleWildcardDetectsOwnerGrant(t *testing.T) {
	cmd := "gcloud projects add-iam-policy-binding my-project --member=user:x@example.com --role=roles/owner"
	findings := shell.GCPRoleWildcard.Run("<command>", []byte(cmd), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestGCPRoleWildcardDetectsEditorGrant(t *testing.T) {
	cmd := "gcloud projects add-iam-policy-binding my-project --role=roles/editor"
	findings := shell.GCPRoleWildcard.Run("<command>", []byte(cmd), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestGCPRoleWildcardIgnoresScopedRole(t *testing.T) {
	cmd := "gcloud projects add-iam-policy-binding my-project --role=roles/logging.viewer"
	findings := shell.GCPRoleWildcard.Run("<command>", []byte(cmd), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestChmodPermissiveDetects777(t *testing.T) {
	findings := shell.ChmodPermissive.Run("<command>", []byte("chmod 777 script.sh"), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestChmodPermissiveIgnoresNarrowMode(t *testing.T) {
	findings := shell.ChmodPermissive.Run("<command>", []byte("chmod 755 script.sh"), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

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
