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
	Message:  "JSONSerialization's untyped [String: Any] API is sloppy for typed payloads; decode with swift-json-schema instead (a @Schemable struct's Type.schema.decode, or JSONValue.parse for ad-hoc parsing)",
	DocsPath: "swift-jsonserialization.md",
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
}
