// internal/checks/nodejs/nodejs.go
package nodejs

import (
	"regexp"

	"aint/internal/check"
)

var Eval = check.Check{
	ID:       "node-eval",
	Title:    "Use of eval()",
	Severity: check.SeverityWarning,
	Langs:    []string{"nodejs"},
	Pattern:  regexp.MustCompile(`\beval\s*\(`),
	Message:  "eval() executes arbitrary strings as code; avoid it or use a safer parser",
	DocsPath: "node-eval.md",
}

var ChildProcessExec = check.Check{
	ID:       "node-child-process-exec",
	Title:    "child_process.exec() usage",
	Severity: check.SeverityError,
	Langs:    []string{"nodejs"},
	Pattern:  regexp.MustCompile(`child_process\.exec\s*\(`),
	Message:  "child_process.exec runs its argument through a shell; use execFile/spawn with an argv array instead",
	DocsPath: "node-child-process-exec.md",
}

var TLSRejectUnauthorizedFalse = check.Check{
	ID:       "node-tls-reject-unauthorized-false",
	Title:    "TLS certificate verification disabled",
	Severity: check.SeverityError,
	Langs:    []string{"nodejs"},
	Pattern:  regexp.MustCompile(`rejectUnauthorized\s*:\s*false|NODE_TLS_REJECT_UNAUTHORIZED\s*=\s*['"]?0`),
	Message:  "disabling TLS certificate verification allows man-in-the-middle attacks",
	DocsPath: "node-tls-reject-unauthorized-false.md",
}

var InsecureRandom = check.Check{
	ID:       "node-insecure-random",
	Title:    "Math.random() used for a security-sensitive value",
	Severity: check.SeverityWarning,
	Langs:    []string{"nodejs"},
	Pattern:  regexp.MustCompile(`(?i)(token|session|secret|password|otp)\w*\s*=\s*.*Math\.random\(\)`),
	Message:  "Math.random() is not cryptographically secure; use crypto.randomBytes() for tokens/sessions/secrets",
	DocsPath: "node-insecure-random.md",
}

var JWTAlgNone = check.Check{
	ID:       "node-jwt-alg-none",
	Title:    "JWT 'none' algorithm allowed",
	Severity: check.SeverityError,
	Langs:    []string{"nodejs"},
	Pattern:  regexp.MustCompile(`algorithms?\s*:\s*\[?\s*['"]none['"]`),
	Message:  "allowing the 'none' JWT algorithm lets an attacker forge unsigned tokens",
	DocsPath: "node-jwt-alg-none.md",
}

var ConsoleLog = check.Check{
	ID:       "node-console-log",
	Title:    "Leftover console.log",
	Severity: check.SeverityInfo,
	Langs:    []string{"nodejs"},
	Pattern:  regexp.MustCompile(`\bconsole\.log\s*\(`),
	Message:  "leftover console.log; use a structured logger for production output",
	DocsPath: "node-console-log.md",
}

var CORSWildcardCredentials = check.Check{
	ID:       "node-cors-wildcard-credentials",
	Title:    "Wildcard CORS origin with credentials",
	Severity: check.SeverityError,
	Langs:    []string{"nodejs"},
	Pattern:  regexp.MustCompile(`origin:\s*['"]\*['"].*credentials:\s*true|credentials:\s*true.*origin:\s*['"]\*['"]`),
	Message:  "wildcard CORS origin combined with credentials:true lets any site read authenticated responses",
	DocsPath: "node-cors-wildcard-credentials.md",
}

var PgClientRawConnect = check.Check{
	ID:       "node-pg-client-raw-connect",
	Title:    "Raw pg.Client instantiation",
	Severity: check.SeverityWarning,
	Langs:    []string{"nodejs"},
	Pattern:  regexp.MustCompile(`new\s+(pg\.)?Client\s*\(`),
	Message:  "node-postgres emits 'error' on the client instance on any abrupt disconnect - attach client.on('error', ...) before .connect() or an uncaught exception kills the whole process, not just this query",
	DocsPath: "node-pg-client-raw-connect.md",
}

func init() {
	check.Register(Eval)
	check.Register(ChildProcessExec)
	check.Register(TLSRejectUnauthorizedFalse)
	check.Register(InsecureRandom)
	check.Register(JWTAlgNone)
	check.Register(ConsoleLog)
	check.Register(CORSWildcardCredentials)
	check.Register(PgClientRawConnect)
}
