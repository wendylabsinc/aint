// internal/checks/swift/swift.go
package swift

import (
	"regexp"

	"aint/internal/check"
)

var ForceUnwrap = check.Check{
	ID:       "swift-force-unwrap",
	Title:    "Force unwrap or force cast",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\b(try|as)!`),
	Message:  "force unwrap/cast can crash at runtime; prefer try?/as? with explicit handling",
	DocsPath: "swift-force-unwrap.md",
}

var PrintStatement = check.Check{
	ID:       "swift-print-statement",
	Title:    "print/debugPrint/NSLog instead of a structured logger",
	Severity: check.SeverityInfo,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\b(print|debugPrint|dump|NSLog)\s*\(`),
	Message:  "use swift-log's Logger instead of print/debugPrint/dump/NSLog for production output",
	DocsPath: "swift-print-statement.md",
}

var UnownedReference = check.Check{
	ID:       "swift-unowned-reference",
	Title:    "unowned reference",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\bunowned(\([^)]*\))?\s+(self|let|var)\b`),
	Message:  "unowned crashes if the referenced object has been deallocated; prefer weak with explicit nil-handling",
	DocsPath: "swift-unowned-reference.md",
}

var HTTPNotHTTPS = check.Check{
	ID:       "swift-http-not-https",
	Title:    "Hardcoded http:// URL",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`"http://[^"]*"`),
	Message:  "hardcoded http:// URL; use https:// to protect data in transit",
	DocsPath: "swift-http-not-https.md",
}

var WebViewArbitraryLoad = check.Check{
	ID:       "swift-webview-arbitrary-load",
	Title:    "WKWebView loading a URLRequest",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\.load\(\s*URLRequest\(url:`),
	Message:  "loading a URLRequest into a WKWebView; verify the URL comes from a trusted, validated source",
	DocsPath: "swift-webview-arbitrary-load.md",
}

var UserDefaultsSensitiveKey = check.Check{
	ID:       "swift-userdefaults-sensitive-key",
	Title:    "Sensitive-looking value stored in UserDefaults",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`(?i)UserDefaults\.\w+\.set\(.*forKey:\s*"[^"]*(password|token|secret|api[_-]?key)[^"]*"`),
	Message:  "storing sensitive-looking data in UserDefaults; use the Keychain instead",
	DocsPath: "swift-userdefaults-sensitive-key.md",
}

var FatalError = check.Check{
	ID:       "swift-fatal-error",
	Title:    "fatalError usage",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\bfatalError\s*\(`),
	Message:  "fatalError terminates the process; prefer throwing an error or precondition for programming invariants",
	DocsPath: "swift-fatal-error.md",
}

var TodoComment = check.Check{
	ID:       "swift-todo-comment",
	Title:    "TODO/FIXME/HACK comment",
	Severity: check.SeverityInfo,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`(?i)//.*\b(TODO|FIXME|HACK)\b`),
	Message:  "TODO/FIXME/HACK comment; track in an issue tracker and resolve before release",
	DocsPath: "swift-todo-comment.md",
}

var ImplicitlyUnwrappedOptional = check.Check{
	ID:       "swift-implicitly-unwrapped-optional",
	Title:    "Implicitly unwrapped optional",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`:\s*[A-Za-z_][A-Za-z0-9_.\[\]<>]*!`),
	Message:  "implicitly unwrapped optional (Type!) crashes if force-accessed while nil; use a regular optional",
	DocsPath: "swift-implicitly-unwrapped-optional.md",
}

var JSONSerializationUsage = check.Check{
	ID:       "swift-jsonserialization",
	Title:    "Foundation JSONSerialization usage",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\bJSONSerialization\b`),
	Message:  "JSONSerialization's untyped [String: Any] API is sloppy; use https://github.com/wendylabsinc/swift-json-schema for generated typed models and/or https://github.com/orlandos-nl/swift-json (IkigaJSON) for raw or ad-hoc JSON",
	DocsPath: "swift-jsonserialization.md",
}

var UnstructuredTask = check.Check{
	ID:       "swift-unstructured-task",
	Title:    "Unstructured Task creation",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\bTask(\.detached)?\s*(<[^>]*>)?\s*[(){]`),
	Message:  "unstructured Task{} escapes its parent's cancellation and error propagation; use a task group, async let, or a structured child task instead",
	DocsPath: "swift-unstructured-task.md",
}

var NonisolatedUnsafe = check.Check{
	ID:       "swift-nonisolated-unsafe",
	Title:    "nonisolated(unsafe) usage",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\bnonisolated\(unsafe\)`),
	Message:  "nonisolated(unsafe) disables the compiler's data-race checking for this access; only justified with a documented safety invariant (e.g. an actor-free C global) and a follow-up to remove it",
	DocsPath: "swift-nonisolated-unsafe.md",
}

var UnsafeMutablePointerUsage = check.Check{
	ID:       "swift-unsafe-mutable-pointer",
	Title:    "UnsafeMutablePointer family usage",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\bUnsafeMutable(Raw)?(Buffer)?Pointer\b`),
	Message:  "UnsafeMutablePointer/UnsafeMutableRawPointer/UnsafeMutableBufferPointer bypass Swift's memory safety and lifetime checks; prefer Span, InlineArray, ContiguousArray, or withUnsafeMutableBufferPointer's scoped closure form",
	DocsPath: "swift-unsafe-mutable-pointer.md",
}

var TaskGroupCancelAllNotDeferred = check.Check{
	ID:       "swift-taskgroup-cancelall-not-deferred",
	Title:    "TaskGroup cancelAll() not wrapped in defer",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`^\s*[A-Za-z_][A-Za-z0-9_]*\.cancelAll\(\)\s*$`),
	Message:  "cancelAll() as a bare statement is skipped if the preceding awaited call throws, leaking the group's remaining tasks; wrap it in defer { ... } so cancellation always runs",
	DocsPath: "swift-taskgroup-cancelall-not-deferred.md",
}

var CommandLineArguments = check.Check{
	ID:       "swift-commandline-arguments",
	Title:    "Manual CommandLine argument handling",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\bCommandLine\.(arguments|argc|unsafeArgv)\b`),
	Message:  "hand-rolling argv parsing off CommandLine drifts into missing --help, no validation, and ad-hoc flag handling; use swift-argument-parser (a @main struct conforming to ParsableCommand with @Argument/@Option/@Flag) instead",
	DocsPath: "swift-commandline-arguments.md",
}

var ProcessInfoEnvironment = check.Check{
	ID:       "swift-processinfo-environment",
	Title:    "Direct ProcessInfo environment access",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\bProcessInfo\s*\.\s*processInfo\s*\.\s*environment\b`),
	Message:  "direct ProcessInfo.processInfo.environment access creates stringly typed, scattered configuration; use swift-configuration's ConfigReader with EnvironmentVariablesProvider instead",
	DocsPath: "swift-processinfo-environment.md",
}

var DispatchSourceSignal = check.Check{
	ID:       "swift-dispatchsource-signal",
	Title:    "Hand-rolled DispatchSource signal handling",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\bDispatchSource\.makeSignalSource\b`),
	Message:  "a DispatchSource signal handler runs outside structured concurrency and gives no ordered shutdown of dependent services; use swift-service-lifecycle's ServiceGroup with gracefulShutdownSignals: [.sigterm, .sigint] instead",
	DocsPath: "swift-dispatchsource-signal.md",
}

var SIMDFixedSizeStorage = check.Check{
	ID:       "swift-simd-fixed-size-storage",
	Title:    "SIMD vector type used as fixed-size storage",
	Severity: check.SeverityInfo,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`\bSIMD\d+\b`),
	Message:  "SIMDn is a lane-parallel math type, not a general fixed-size buffer; unless you need vector arithmetic, use InlineArray (e.g. InlineArray<4, Float>) for inline fixed-capacity storage",
	DocsPath: "swift-simd-fixed-size-storage.md",
}

// ToolsVersionOutdated matches a `// swift-tools-version:` header at 6.2 or
// earlier. RE2 has no lookahead, so the version alternatives are anchored to
// end-of-line (or the legacy `;`-suffixed form) to keep 6.3, 6.10 and 7.x from
// matching the shorter `6` / `6.1` branches.
var ToolsVersionOutdated = check.Check{
	ID:       "swift-tools-version-outdated",
	Title:    "Package.swift below the latest swift-tools-version",
	Severity: check.SeverityWarning,
	Langs:    []string{"swift"},
	Pattern:  regexp.MustCompile(`^//\s*swift-tools-version:\s*(?:[0-5](?:\.\d+)*|6|6\.[0-2])\s*(?:;.*)?$`),
	Message:  "Package.swift declares an outdated swift-tools-version; new packages should always declare the latest (// swift-tools-version: 6.3), never 6.2 or earlier",
	DocsPath: "swift-tools-version-outdated.md",
}

func init() {
	check.Register(ForceUnwrap)
	check.Register(PrintStatement)
	check.Register(UnownedReference)
	check.Register(HTTPNotHTTPS)
	check.Register(WebViewArbitraryLoad)
	check.Register(UserDefaultsSensitiveKey)
	check.Register(FatalError)
	check.Register(TodoComment)
	check.Register(ImplicitlyUnwrappedOptional)
	check.Register(JSONSerializationUsage)
	check.Register(UnstructuredTask)
	check.Register(NonisolatedUnsafe)
	check.Register(UnsafeMutablePointerUsage)
	check.Register(TaskGroupCancelAllNotDeferred)
	check.Register(CommandLineArguments)
	check.Register(ProcessInfoEnvironment)
	check.Register(DispatchSourceSignal)
	check.Register(SIMDFixedSizeStorage)
	check.Register(ToolsVersionOutdated)
}
