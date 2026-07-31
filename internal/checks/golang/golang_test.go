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
