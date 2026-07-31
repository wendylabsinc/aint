// internal/checks/secrets/secrets_test.go
package secrets_test

import (
	"testing"

	"aint/internal/checks/secrets"
)

func TestHardcodedKeyDetectsAWSKey(t *testing.T) {
	findings := secrets.HardcodedKey.Run("test.go", []byte(`key := "AKIAABCDEFGHIJKLMNOP"`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestHardcodedKeyDetectsGenericSecretAssignment(t *testing.T) {
	findings := secrets.HardcodedKey.Run("config.py", []byte(`api_key = "sk-abcdefghijklmnopqrstuvwx"`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestHardcodedKeyIgnoresShortStrings(t *testing.T) {
	findings := secrets.HardcodedKey.Run("test.go", []byte(`name := "hello"`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestPrivateKeyBlockDetectsCommittedKey(t *testing.T) {
	content := "-----BEGIN RSA PRIVATE KEY-----\nMIIB...\n-----END RSA PRIVATE KEY-----"
	findings := secrets.PrivateKeyBlock.Run("id_rsa", []byte(content), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestPrivateKeyBlockIgnoresUnrelatedFiles(t *testing.T) {
	findings := secrets.PrivateKeyBlock.Run("readme.md", []byte("# hello world"), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
