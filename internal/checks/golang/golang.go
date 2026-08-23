// internal/checks/golang/golang.go
package golang

import (
	"regexp"

	"aint/internal/check"
)

var IgnoredError = check.Check{
	ID:       "go-ignored-error",
	Title:    "Ignored error return value",
	Severity: check.SeverityWarning,
	Langs:    []string{"go"},
	Pattern:  regexp.MustCompile(`\b_\s*=\s*err\b`),
	Message:  "error value discarded by assigning it to the blank identifier",
	DocsPath: "go-ignored-error.md",
}

var ExecCommandInjection = check.Check{
	ID:       "go-exec-command-injection",
	Title:    "Shell spawned via exec.Command -c",
	Severity: check.SeverityError,
	Langs:    []string{"go"},
	Pattern:  regexp.MustCompile(`exec\.Command\(\s*"(sh|bash|/bin/sh|/bin/bash)"\s*,\s*"-c"`),
	Message:  "spawning a shell via exec.Command with -c risks command injection if any argument isn't a fixed literal",
	DocsPath: "go-exec-command-injection.md",
}

var TLSInsecureSkipVerify = check.Check{
	ID:       "go-tls-insecure-skip-verify",
	Title:    "TLS certificate verification disabled",
	Severity: check.SeverityError,
	Langs:    []string{"go"},
	Pattern:  regexp.MustCompile(`InsecureSkipVerify:\s*true`),
	Message:  "disabling TLS certificate verification allows man-in-the-middle attacks",
	DocsPath: "go-tls-insecure-skip-verify.md",
}

var WeakCryptoHash = check.Check{
	ID:       "go-weak-crypto-hash",
	Title:    "Weak cryptographic hash (MD5/SHA1)",
	Severity: check.SeverityWarning,
	Langs:    []string{"go"},
	Pattern:  regexp.MustCompile(`\b(md5|sha1)\.(New|Sum)\(`),
	Message:  "MD5/SHA1 are not safe for password hashing or security-sensitive checksums; use bcrypt/argon2 or SHA-256+",
	DocsPath: "go-weak-crypto-hash.md",
}

var HTTPClientNoTimeout = check.Check{
	ID:       "go-http-client-no-timeout",
	Title:    "http.Client created with no Timeout",
	Severity: check.SeverityWarning,
	Langs:    []string{"go"},
	Pattern:  regexp.MustCompile(`http\.Client\{\s*\}`),
	Message:  "an empty http.Client literal has no Timeout set; a hung server can block this goroutine forever",
	DocsPath: "go-http-client-no-timeout.md",
}

var UncheckedTypeAssertion = check.Check{
	ID:       "go-unchecked-type-assertion",
	Title:    "Type assertion without comma-ok",
	Severity: check.SeverityWarning,
	Langs:    []string{"go"},
	Pattern:  regexp.MustCompile(`^\s*[A-Za-z_][A-Za-z0-9_]*\s*:=\s*[A-Za-z_][A-Za-z0-9_.]*\.\([A-Za-z_][A-Za-z0-9_.]*\)\s*$`),
	Message:  "type assertion without the comma-ok form panics if the assertion fails",
	DocsPath: "go-unchecked-type-assertion.md",
}

var SQLStringConcat = check.Check{
	ID:       "go-sql-string-concat",
	Title:    "SQL query built via string concatenation",
	Severity: check.SeverityError,
	Langs:    []string{"go"},
	Pattern:  regexp.MustCompile(`\.(Query|Exec|QueryRow)\w*\([^)]*(fmt\.Sprintf|"\s*\+|\+\s*")`),
	Message:  "building a SQL query via string concatenation/formatting risks SQL injection; use parameterized queries",
	DocsPath: "go-sql-string-concat.md",
}

var ErrorDirectComparison = check.Check{
	ID:       "go-error-direct-comparison",
	Title:    "Direct comparison of an error",
	Severity: check.SeverityWarning,
	Langs:    []string{"go"},
	Analyze:  errorDirectComparison,
	Message:  "direct error comparison misses wrapped errors; use errors.Is unless identity comparison is intentional",
	DocsPath: "go-error-direct-comparison.md",
}

var ErrorStringComparison = check.Check{
	ID:       "go-error-string-comparison",
	Title:    "Error inspected by its message string",
	Severity: check.SeverityWarning,
	Langs:    []string{"go"},
	Analyze:  errorStringComparison,
	Message:  "matching err.Error() text is brittle; use errors.Is/errors.As or a typed error",
	DocsPath: "go-error-string-comparison.md",
}

var PanicWithError = check.Check{
	ID:       "go-panic-with-error",
	Title:    "Error passed to panic",
	Severity: check.SeverityWarning,
	Langs:    []string{"go"},
	Analyze:  panicWithError,
	Message:  "panic(err) turns a normal failure into a process crash; return or handle the error unless an invariant is broken",
	DocsPath: "go-panic-with-error.md",
}

var DeferInLoop = check.Check{
	ID:       "go-defer-in-loop",
	Title:    "defer used inside a loop",
	Severity: check.SeverityWarning,
	Langs:    []string{"go"},
	Analyze:  deferInLoop,
	Message:  "deferred cleanup runs when the surrounding function returns, so resources accumulate across loop iterations",
	DocsPath: "go-defer-in-loop.md",
}

var GoroutineInLoop = check.Check{
	ID:       "go-goroutine-in-loop",
	Title:    "Goroutine started inside a loop",
	Severity: check.SeverityWarning,
	Langs:    []string{"go"},
	Analyze:  goroutineInLoop,
	Message:  "starting goroutines in a loop can create unbounded concurrency; use an explicit limit and cancellation path",
	DocsPath: "go-goroutine-in-loop.md",
}

var GetterPrefix = check.Check{
	ID:       "go-getter-prefix",
	Title:    "Getter method uses a Get prefix",
	Severity: check.SeverityInfo,
	Langs:    []string{"go"},
	Analyze:  getterPrefix,
	Message:  "Go getter methods conventionally omit the Get prefix (Name rather than GetName)",
	DocsPath: "go-getter-prefix.md",
}

var LargeInterface = check.Check{
	ID:       "go-large-interface",
	Title:    "Interface has more than five members",
	Severity: check.SeverityInfo,
	Langs:    []string{"go"},
	Analyze:  largeInterface,
	Message:  "large interfaces are difficult to implement and mock; define the smallest interface needed by the consumer",
	DocsPath: "go-large-interface.md",
}

func init() {
	check.Register(IgnoredError)
	check.Register(ExecCommandInjection)
	check.Register(TLSInsecureSkipVerify)
	check.Register(WeakCryptoHash)
	check.Register(HTTPClientNoTimeout)
	check.Register(UncheckedTypeAssertion)
	check.Register(SQLStringConcat)
	check.Register(ErrorDirectComparison)
	check.Register(ErrorStringComparison)
	check.Register(PanicWithError)
	check.Register(DeferInLoop)
	check.Register(GoroutineInLoop)
	check.Register(GetterPrefix)
	check.Register(LargeInterface)
}
