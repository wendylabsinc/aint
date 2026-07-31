// internal/checks/python/python_test.go
package python_test

import (
	"testing"

	"aint/internal/checks/python"
)

func TestShellTrueDetectsShellInjectionRisk(t *testing.T) {
	findings := python.ShellTrue.Run("script.py", []byte(`subprocess.run(cmd, shell=True)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestShellTrueIgnoresSafeCall(t *testing.T) {
	findings := python.ShellTrue.Run("script.py", []byte(`subprocess.run(cmd)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
