// internal/checks/swift/swift_test.go
package swift_test

import (
	"strings"
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

func TestJSONSerializationUsageDetectsJSONObjectWith(t *testing.T) {
	findings := swift.JSONSerializationUsage.Run("main.swift", []byte(`let obj = try JSONSerialization.jsonObject(with: data)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestJSONSerializationUsageIgnoresSwiftJSONSchema(t *testing.T) {
	findings := swift.JSONSerializationUsage.Run("main.swift", []byte(`let user = try User(json: jsonObject)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestJSONSerializationUsageFeedbackNamesBothAlternatives(t *testing.T) {
	message := swift.JSONSerializationUsage.Message
	for _, alternative := range []string{"https://github.com/wendylabsinc/swift-json-schema", "https://github.com/orlandos-nl/swift-json"} {
		if !strings.Contains(message, alternative) {
			t.Errorf("expected feedback to mention %q, got %q", alternative, message)
		}
	}
}

func TestUnstructuredTaskDetectsPlainTask(t *testing.T) {
	findings := swift.UnstructuredTask.Run("main.swift", []byte(`Task { await run() }`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestUnstructuredTaskDetectsDetached(t *testing.T) {
	findings := swift.UnstructuredTask.Run("main.swift", []byte(`Task.detached(priority: .background) { await run() }`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestUnstructuredTaskIgnoresAddTask(t *testing.T) {
	findings := swift.UnstructuredTask.Run("main.swift", []byte(`group.addTask { await run() }`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestUnstructuredTaskIgnoresTypeDeclaration(t *testing.T) {
	findings := swift.UnstructuredTask.Run("main.swift", []byte(`var task: Task<Void, Never>?`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestNonisolatedUnsafeDetectsUsage(t *testing.T) {
	findings := swift.NonisolatedUnsafe.Run("main.swift", []byte(`nonisolated(unsafe) let out = stdout`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestNonisolatedUnsafeIgnoresPlainNonisolated(t *testing.T) {
	findings := swift.NonisolatedUnsafe.Run("main.swift", []byte(`nonisolated func compute() -> Int { 42 }`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestUnsafeMutablePointerUsageDetectsBasePointer(t *testing.T) {
	findings := swift.UnsafeMutablePointerUsage.Run("main.swift", []byte(`let buf = UnsafeMutablePointer<UInt8>.allocate(capacity: 16)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestUnsafeMutablePointerUsageDetectsBufferVariant(t *testing.T) {
	findings := swift.UnsafeMutablePointerUsage.Run("main.swift", []byte(`func fill(_ buf: UnsafeMutableBufferPointer<UInt8>) {}`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestUnsafeMutablePointerUsageIgnoresImmutablePointer(t *testing.T) {
	findings := swift.UnsafeMutablePointerUsage.Run("main.swift", []byte(`func read(_ buf: UnsafePointer<UInt8>) {}`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestTaskGroupCancelAllNotDeferredDetectsBareStatement(t *testing.T) {
	findings := swift.TaskGroupCancelAllNotDeferred.Run("main.swift", []byte("try await group.next()\ngroup.cancelAll()"), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestTaskGroupCancelAllNotDeferredIgnoresDeferOneLiner(t *testing.T) {
	findings := swift.TaskGroupCancelAllNotDeferred.Run("main.swift", []byte("defer { group.cancelAll() }\ntry await group.next()"), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestCommandLineArgumentsDetectsArguments(t *testing.T) {
	findings := swift.CommandLineArguments.Run("main.swift", []byte(`let path = CommandLine.arguments[1]`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestCommandLineArgumentsDetectsUnsafeArgv(t *testing.T) {
	findings := swift.CommandLineArguments.Run("main.swift", []byte(`let argv = CommandLine.unsafeArgv`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestCommandLineArgumentsIgnoresArgumentParser(t *testing.T) {
	findings := swift.CommandLineArguments.Run("main.swift", []byte(`@Argument(help: "Path.") var path: String`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestProcessInfoEnvironmentDetectsDirectAccess(t *testing.T) {
	findings := swift.ProcessInfoEnvironment.Run("main.swift", []byte(`let port = ProcessInfo.processInfo.environment["PORT"]`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestProcessInfoEnvironmentDetectsAccessWithWhitespace(t *testing.T) {
	findings := swift.ProcessInfoEnvironment.Run("main.swift", []byte(`let environment = ProcessInfo . processInfo . environment`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestProcessInfoEnvironmentIgnoresSwiftConfiguration(t *testing.T) {
	findings := swift.ProcessInfoEnvironment.Run("main.swift", []byte(`let config = ConfigReader(provider: EnvironmentVariablesProvider())`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestDispatchSourceSignalDetectsMakeSignalSource(t *testing.T) {
	findings := swift.DispatchSourceSignal.Run("main.swift", []byte(`let src = DispatchSource.makeSignalSource(signal: SIGTERM, queue: .main)`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestDispatchSourceSignalIgnoresTimerSource(t *testing.T) {
	findings := swift.DispatchSourceSignal.Run("main.swift", []byte(`let src = DispatchSource.makeTimerSource(queue: .main)`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestSIMDFixedSizeStorageDetectsSIMD4(t *testing.T) {
	findings := swift.SIMDFixedSizeStorage.Run("main.swift", []byte(`var channels = SIMD4<Float>()`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestSIMDFixedSizeStorageIgnoresInlineArray(t *testing.T) {
	findings := swift.SIMDFixedSizeStorage.Run("main.swift", []byte(`var channels: InlineArray<4, Float> = [0, 0, 0, 0]`), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestToolsVersionOutdatedDetects62(t *testing.T) {
	findings := swift.ToolsVersionOutdated.Run("Package.swift", []byte("// swift-tools-version: 6.2\nimport PackageDescription"), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestToolsVersionOutdatedDetectsLegacySuffixedForm(t *testing.T) {
	findings := swift.ToolsVersionOutdated.Run("Package.swift", []byte(`// swift-tools-version:5.3;-swift-version:4`), "")
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
}

func TestToolsVersionOutdatedIgnoresLatest(t *testing.T) {
	findings := swift.ToolsVersionOutdated.Run("Package.swift", []byte("// swift-tools-version: 6.3\nimport PackageDescription"), "")
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d: %+v", len(findings), findings)
	}
}

func TestToolsVersionOutdatedIgnoresFutureVersions(t *testing.T) {
	for _, line := range []string{"// swift-tools-version: 6.10", "// swift-tools-version: 7.0", "// swift-tools-version:6.4"} {
		findings := swift.ToolsVersionOutdated.Run("Package.swift", []byte(line), "")
		if len(findings) != 0 {
			t.Fatalf("expected 0 findings for %q, got %d: %+v", line, len(findings), findings)
		}
	}
}
