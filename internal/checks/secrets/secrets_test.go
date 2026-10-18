// internal/checks/secrets/secrets_test.go
package secrets_test

import (
	"strings"
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

func TestHighEntropyStringIgnoresGitObjectIDs(t *testing.T) {
	for _, value := range []string{
		"e28fdb1d3af7d4671d701f5cfef70493ffaf87fa",
		strings.Repeat("0123456789abcdef", 4),
	} {
		for _, value := range []string{value, strings.ToUpper(value), "E" + value[1:]} {
			for _, quote := range []string{"'", `"`} {
				t.Run(quote+value+quote, func(t *testing.T) {
					content := "commit = " + quote + value + quote
					findings := secrets.HighEntropyString.Run("sign_commits.py", []byte(content), "")
					if len(findings) != 0 {
						t.Fatalf("expected no findings for a Git object ID, got %+v", findings)
					}
				})
			}
		}
	}
}

func TestHighEntropyStringStillDetectsNonHashLiterals(t *testing.T) {
	for _, value := range []string{
		strings.Repeat("a", 41),
		strings.Repeat("a", 63),
		strings.Repeat("a", 65),
		strings.Repeat("a", 39) + "g",
		strings.Repeat("a", 63) + "/",
		strings.Repeat("a", 40) + "=",
		strings.Repeat("a", 64) + "==",
	} {
		t.Run(value, func(t *testing.T) {
			findings := secrets.HighEntropyString.Run("config.py", []byte(`value = "`+value+`"`), "")
			if len(findings) != 1 {
				t.Fatalf("expected 1 finding for a non-hash literal, got %+v", findings)
			}
		})
	}
}

func TestHighEntropyStringFindsLiteralAfterGitObjectIDs(t *testing.T) {
	const hash = `'e28fdb1d3af7d4671d701f5cfef70493ffaf87fa'`
	const blob = `"aGVsbG8gd29ybGQgdGhpcyBpcyBhIGxvbmcgYmFzZTY0IHN0cmluZw=="`
	const prefix = "values = [" + hash + ", " + hash + ", "
	content := "commit = " + hash + "\n" + prefix + blob + ", " + blob + "]\n" + "blob = " + blob
	findings := secrets.HighEntropyString.Run("config.py", []byte(content), "docs/checks/secret-high-entropy-string.md")
	if len(findings) != 2 {
		t.Fatalf("expected one finding per line with a non-hash literal, got %+v", findings)
	}
	if findings[0].Line != 2 || findings[0].Column != len(prefix)+1 {
		t.Errorf("expected the first blob at 2:%d, got %+v", len(prefix)+1, findings[0])
	}
	if findings[1].Line != 3 || findings[1].Column != len("blob = ")+1 {
		t.Errorf("expected the second blob at 3:8, got %+v", findings[1])
	}
}
