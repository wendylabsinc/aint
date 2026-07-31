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

func TestPrintStatementDetectsPrint(t *testing.T) {
	findings := swift.PrintStatement.Run("main.swift", []byte(`print("debug: \(value)")`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestPrintStatementDetectsNSLog(t *testing.T) {
	findings := swift.PrintStatement.Run("main.swift", []byte(`NSLog("debug: %@", value)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestPrintStatementIgnoresLogger(t *testing.T) {
	findings := swift.PrintStatement.Run("main.swift", []byte(`logger.info("debug: \(value)")`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestUnownedReferenceDetectsCaptureList(t *testing.T) {
	findings := swift.UnownedReference.Run("main.swift", []byte(`task = Task { [unowned self] in await self.run() }`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestUnownedReferenceDetectsPropertyDeclaration(t *testing.T) {
	findings := swift.UnownedReference.Run("main.swift", []byte(`unowned let delegate: SomeDelegate`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestUnownedReferenceIgnoresWeak(t *testing.T) {
	findings := swift.UnownedReference.Run("main.swift", []byte(`weak var delegate: SomeDelegate?`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestHTTPNotHTTPSDetectsPlainHTTP(t *testing.T) {
	findings := swift.HTTPNotHTTPS.Run("main.swift", []byte(`let url = "http://api.example.com/data"`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestHTTPNotHTTPSIgnoresHTTPS(t *testing.T) {
	findings := swift.HTTPNotHTTPS.Run("main.swift", []byte(`let url = "https://api.example.com/data"`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestWebViewArbitraryLoadDetectsURLRequestLoad(t *testing.T) {
	findings := swift.WebViewArbitraryLoad.Run("main.swift", []byte(`webView.load(URLRequest(url: someURL))`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestWebViewArbitraryLoadIgnoresLoadHTMLString(t *testing.T) {
	findings := swift.WebViewArbitraryLoad.Run("main.swift", []byte(`webView.loadHTMLString(staticHTML, baseURL: nil)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestUserDefaultsSensitiveKeyDetectsTokenKey(t *testing.T) {
	findings := swift.UserDefaultsSensitiveKey.Run("main.swift", []byte(`UserDefaults.standard.set(authToken, forKey: "userAuthToken")`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestUserDefaultsSensitiveKeyIgnoresUsernameKey(t *testing.T) {
	findings := swift.UserDefaultsSensitiveKey.Run("main.swift", []byte(`UserDefaults.standard.set(username, forKey: "username")`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestFatalErrorDetectsCall(t *testing.T) {
	findings := swift.FatalError.Run("main.swift", []byte(`fatalError("unreachable")`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestFatalErrorIgnoresPreconditionFailure(t *testing.T) {
	findings := swift.FatalError.Run("main.swift", []byte(`preconditionFailure("unreachable")`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestTodoCommentDetectsTODO(t *testing.T) {
	findings := swift.TodoComment.Run("main.swift", []byte(`// TODO: handle this edge case`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestTodoCommentIgnoresRegularComment(t *testing.T) {
	findings := swift.TodoComment.Run("main.swift", []byte(`// this function handles the edge case`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestImplicitlyUnwrappedOptionalDetectsDeclaration(t *testing.T) {
	findings := swift.ImplicitlyUnwrappedOptional.Run("main.swift", []byte(`var window: UIWindow!`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestImplicitlyUnwrappedOptionalIgnoresRegularOptional(t *testing.T) {
	findings := swift.ImplicitlyUnwrappedOptional.Run("main.swift", []byte(`var window: UIWindow?`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}
