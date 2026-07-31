// internal/checks/swift/swift_test.go
package swift_test

import (
	"testing"

	"aint/internal/checks/swift"
)

func TestForceUnwrapDetectsTryBang(t *testing.T) {
	findings := swift.ForceUnwrap.Run("main.swift", []byte("let x = try! risky()"), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestForceUnwrapDetectsAsBang(t *testing.T) {
	findings := swift.ForceUnwrap.Run("main.swift", []byte("let s = value as! String"), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestForceUnwrapIgnoresSafeTry(t *testing.T) {
	findings := swift.ForceUnwrap.Run("main.swift", []byte("let x = try? risky()"), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
