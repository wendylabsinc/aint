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

func TestErrorDirectComparisonDetectsSentinelComparison(t *testing.T) {
	content := []byte("package p\nfunc f(err error) { if err == io.EOF {} }")
	findings := golang.ErrorDirectComparison.Run("main.go", content, "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestErrorDirectComparisonAllowsNilCheck(t *testing.T) {
	content := []byte("package p\nfunc f(err error) { if err != nil {} }")
	findings := golang.ErrorDirectComparison.Run("main.go", content, "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestErrorStringComparisonDetectsEqualityAndContains(t *testing.T) {
	content := []byte(`package p
func f(err error) {
	_ = err.Error() == "not found"
	_ = strings.Contains(err.Error(), "timeout")
}`)
	findings := golang.ErrorStringComparison.Run("main.go", content, "")
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d: %+v", len(findings), findings)
	}
}

func TestErrorStringComparisonAllowsErrorsIs(t *testing.T) {
	content := []byte("package p\nfunc f(err error) { _ = errors.Is(err, io.EOF) }")
	findings := golang.ErrorStringComparison.Run("main.go", content, "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestPanicWithErrorDetectsErrorValues(t *testing.T) {
	content := []byte(`package p
func f(err error) {
	panic(err)
	panic(fmt.Errorf("broken: %w", err))
}`)
	findings := golang.PanicWithError.Run("main.go", content, "")
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d: %+v", len(findings), findings)
	}
}

func TestPanicWithErrorAllowsInvariantMessage(t *testing.T) {
	content := []byte(`package p
func f() { panic("unreachable") }`)
	findings := golang.PanicWithError.Run("main.go", content, "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestDeferInLoopDetectsRangeAndForLoops(t *testing.T) {
	content := []byte(`package p
func f(items []string) {
	for range items { defer closeOne() }
	for i := 0; i < 1; i++ { defer closeTwo() }
}`)
	findings := golang.DeferInLoop.Run("main.go", content, "")
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d: %+v", len(findings), findings)
	}
}

func TestDeferInLoopAllowsPerIterationFunction(t *testing.T) {
	content := []byte(`package p
func f(items []string) {
	for range items { func() { defer closeOne() }() }
}`)
	findings := golang.DeferInLoop.Run("main.go", content, "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestGoroutineInLoopDetectsUnboundedPattern(t *testing.T) {
	content := []byte(`package p
func f(items []string) {
	for _, item := range items { go process(item) }
}`)
	findings := golang.GoroutineInLoop.Run("main.go", content, "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestGoroutineInLoopIgnoresNestedFunctionBody(t *testing.T) {
	content := []byte(`package p
func f(items []string) {
	for range items { funcs = append(funcs, func() { go process() }) }
}`)
	findings := golang.GoroutineInLoop.Run("main.go", content, "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestGetterPrefixDetectsConventionalGetter(t *testing.T) {
	content := []byte(`package p
type User struct{}
func (User) GetName() string { return "" }
func (User) Get(name string) string { return name }
func (User) Name() string { return "" }`)
	findings := golang.GetterPrefix.Run("main.go", content, "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestLargeInterfaceDetectsMoreThanFiveMembers(t *testing.T) {
	content := []byte(`package p
type Small interface { A(); B(); C(); D(); E() }
type Large interface { A(); B(); C(); D(); E(); F() }`)
	findings := golang.LargeInterface.Run("main.go", content, "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestSyntaxChecksIgnoreIncompleteGo(t *testing.T) {
	findings := golang.DeferInLoop.Run("main.go", []byte("not valid Go"), "")
	if len(findings) != 0 {
		t.Fatalf("expected parse failures to produce no findings, got %+v", findings)
	}
}
