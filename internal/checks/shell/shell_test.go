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

func TestGrepSearchCodebaseDetectsRecursiveGrep(t *testing.T) {
	findings := shell.GrepSearchCodebase.Run("<command>", []byte(`grep -rn "TODO" .`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestGrepSearchCodebaseDetectsRipgrep(t *testing.T) {
	findings := shell.GrepSearchCodebase.Run("<command>", []byte(`rg "TODO" src/`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestGrepSearchCodebaseIgnoresSingleFileGrep(t *testing.T) {
	findings := shell.GrepSearchCodebase.Run("<command>", []byte(`grep -n "TODO" file.go`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestGrepSearchCodebaseIgnoresFilteringProcessOutput(t *testing.T) {
	findings := shell.GrepSearchCodebase.Run("<command>", []byte("ps aux | grep node"), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestGitAddBroadDetectsDashA(t *testing.T) {
	findings := shell.GitAddBroad.Run("<command>", []byte("git add -A"), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestGitAddBroadDetectsBareDot(t *testing.T) {
	findings := shell.GitAddBroad.Run("<command>", []byte("git add ."), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestGitAddBroadIgnoresSpecificFile(t *testing.T) {
	findings := shell.GitAddBroad.Run("<command>", []byte("git add internal/apps_install.go"), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestGitCheckoutSharedWendyosTreeDetectsCdAndCheckout(t *testing.T) {
	cmd := "cd ~/git/wendy/wendyos && git checkout jo/app-install"
	findings := shell.GitCheckoutSharedWendyosTree.Run("<command>", []byte(cmd), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestGitCheckoutSharedWendyosTreeDetectsDashCSwitch(t *testing.T) {
	cmd := "git -C ~/git/wendy/wendyos switch main"
	findings := shell.GitCheckoutSharedWendyosTree.Run("<command>", []byte(cmd), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestGitCheckoutSharedWendyosTreeIgnoresSiblingRepo(t *testing.T) {
	cmd := "cd ~/git/wendy/wendyos-builder && git checkout jo/foo"
	findings := shell.GitCheckoutSharedWendyosTree.Run("<command>", []byte(cmd), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestGitCheckoutSharedWendyosTreeIgnoresUnrelatedCheckout(t *testing.T) {
	cmd := "cd ~/git/wendy/pascal && git checkout jo/foo"
	findings := shell.GitCheckoutSharedWendyosTree.Run("<command>", []byte(cmd), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestPsqlInlineSQLVariableDetectsInterpolatedSQL(t *testing.T) {
	cmd := `psql -c "BEGIN READ ONLY; $SQL; COMMIT;"`
	findings := shell.PsqlInlineSQLVariable.Run("<command>", []byte(cmd), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestPsqlInlineSQLVariableIgnoresStaticSQL(t *testing.T) {
	cmd := `psql -c "SELECT 1;"`
	findings := shell.PsqlInlineSQLVariable.Run("<command>", []byte(cmd), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestContainerCachePurgeForceDetectsDockerBuilderPruneForce(t *testing.T) {
	findings := shell.ContainerCachePurgeForce.Run("<command>", []byte("docker builder prune -af"), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestContainerCachePurgeForceDetectsContainerBuilderDeleteForce(t *testing.T) {
	findings := shell.ContainerCachePurgeForce.Run("<command>", []byte("container builder delete --force"), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestContainerCachePurgeForceIgnoresPlainPrune(t *testing.T) {
	findings := shell.ContainerCachePurgeForce.Run("<command>", []byte("docker container prune"), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestYoctoTmpdirSymlinkDetectsSymlinkedBuildTmp(t *testing.T) {
	findings := shell.YoctoTmpdirSymlink.Run("<command>", []byte("ln -s /wendy/build-tmp build/tmp"), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestYoctoTmpdirSymlinkIgnoresUnrelatedSymlink(t *testing.T) {
	findings := shell.YoctoTmpdirSymlink.Run("<command>", []byte("ln -s foo.txt bar.txt"), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
