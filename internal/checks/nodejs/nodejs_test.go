// internal/checks/nodejs/nodejs_test.go
package nodejs_test

import (
	"testing"

	"aint/internal/checks/nodejs"
)

func TestEvalDetectsCall(t *testing.T) {
	findings := nodejs.Eval.Run("index.js", []byte(`eval(userInput)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestEvalIgnoresSimilarIdentifier(t *testing.T) {
	findings := nodejs.Eval.Run("index.js", []byte(`evaluate(userInput)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

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
