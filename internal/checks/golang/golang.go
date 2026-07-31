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
	Message:  "error value discarded via `_ = err`",
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
	Message:  "http.Client{} has no Timeout set; a hung server can block this goroutine forever",
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

func init() {
	check.Register(IgnoredError)
	check.Register(ExecCommandInjection)
	check.Register(TLSInsecureSkipVerify)
	check.Register(WeakCryptoHash)
	check.Register(HTTPClientNoTimeout)
	check.Register(UncheckedTypeAssertion)
	check.Register(SQLStringConcat)
}
