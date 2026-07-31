// internal/checks/python/python.go
package python

import (
	"regexp"

	"aint/internal/check"
)

var ShellTrue = check.Check{
	ID:       "python-shell-true",
	Title:    "subprocess call with shell=True",
	Severity: check.SeverityWarning,
	Langs:    []string{"python"},
	Pattern:  regexp.MustCompile(`subprocess\.\w+\([^)]*shell\s*=\s*True`),
	Message:  "shell=True risks shell injection if any part of the command is untrusted input",
	DocsPath: "python-shell-true.md",
}

var EvalExec = check.Check{
	ID:       "python-eval-exec",
	Title:    "Use of eval()/exec()",
	Severity: check.SeverityError,
	Langs:    []string{"python"},
	Pattern:  regexp.MustCompile(`\b(eval|exec)\s*\(`),
	Message:  "eval()/exec() executes arbitrary strings as code; avoid it or use a safer parser",
	DocsPath: "python-eval-exec.md",
}

var OSSystem = check.Check{
	ID:       "python-os-system",
	Title:    "Use of os.system()",
	Severity: check.SeverityError,
	Langs:    []string{"python"},
	Pattern:  regexp.MustCompile(`\bos\.system\s*\(`),
	Message:  "os.system() runs its argument through the shell; risks command injection with untrusted input",
	DocsPath: "python-os-system.md",
}

var PickleLoad = check.Check{
	ID:       "python-pickle-load",
	Title:    "Unpickling data",
	Severity: check.SeverityError,
	Langs:    []string{"python"},
	Pattern:  regexp.MustCompile(`\bpickle\.(load|loads)\s*\(`),
	Message:  "unpickling untrusted data can execute arbitrary code; use json or another safe serialization format",
	DocsPath: "python-pickle-load.md",
}

var YAMLUnsafeLoad = check.Check{
	ID:       "python-yaml-unsafe-load",
	Title:    "yaml.load() without a safe loader",
	Severity: check.SeverityError,
	Langs:    []string{"python"},
	Pattern:  regexp.MustCompile(`\byaml\.load\s*\(`),
	Message:  "yaml.load() can execute arbitrary code via !!python/object tags unless Loader=yaml.SafeLoader; prefer yaml.safe_load()",
	DocsPath: "python-yaml-unsafe-load.md",
}

var RequestsNoVerify = check.Check{
	ID:       "python-requests-no-verify",
	Title:    "requests call with verify=False",
	Severity: check.SeverityError,
	Langs:    []string{"python"},
	Pattern:  regexp.MustCompile(`requests\.\w+\([^)]*verify\s*=\s*False`),
	Message:  "verify=False disables TLS certificate verification, allowing man-in-the-middle attacks",
	DocsPath: "python-requests-no-verify.md",
}

var WeakHash = check.Check{
	ID:       "python-weak-hash",
	Title:    "Weak cryptographic hash (MD5/SHA1)",
	Severity: check.SeverityWarning,
	Langs:    []string{"python"},
	Pattern:  regexp.MustCompile(`hashlib\.(md5|sha1)\s*\(`),
	Message:  "MD5/SHA1 are not safe for password hashing; use a dedicated password-hashing library (bcrypt, argon2)",
	DocsPath: "python-weak-hash.md",
}

var AssertSecurityCheck = check.Check{
	ID:       "python-assert-security-check",
	Title:    "assert used for an authorization check",
	Severity: check.SeverityWarning,
	Langs:    []string{"python"},
	Pattern:  regexp.MustCompile(`(?i)assert\s+.*(is_admin|authenticated|authorized|has_permission)`),
	Message:  "assert is stripped when Python runs with -O; don't use it to gate authorization/permission checks",
	DocsPath: "python-assert-security-check.md",
}

func init() {
	check.Register(ShellTrue)
	check.Register(EvalExec)
	check.Register(OSSystem)
	check.Register(PickleLoad)
	check.Register(YAMLUnsafeLoad)
	check.Register(RequestsNoVerify)
	check.Register(WeakHash)
	check.Register(AssertSecurityCheck)
}
