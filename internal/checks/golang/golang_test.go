// internal/checks/golang/golang_test.go
package golang_test

import (
	"testing"

	"aint/internal/checks/golang"
)

func TestIgnoredErrorDetectsDiscard(t *testing.T) {
	findings := golang.IgnoredError.Run("main.go", []byte("_ = err"), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestIgnoredErrorIgnoresHandledError(t *testing.T) {
	findings := golang.IgnoredError.Run("main.go", []byte("result, err := doSomething()"), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

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
