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
