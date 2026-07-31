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
