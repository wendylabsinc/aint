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
