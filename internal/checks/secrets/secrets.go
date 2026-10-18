// internal/checks/secrets/secrets.go
package secrets

import (
	"regexp"
	"strings"

	"aint/internal/check"
)

var HardcodedKey = check.Check{
	ID:       "secret-hardcoded-key",
	Title:    "Hardcoded API key or token",
	Severity: check.SeverityError,
	Pattern: regexp.MustCompile(
		`AKIA[0-9A-Z]{16}|sk-[a-zA-Z0-9]{20,}|(?i)(api_key|apikey|secret|token)\s*[:=]\s*["'][^"']{16,}["']`,
	),
	Message:  "possible hardcoded API key or token",
	DocsPath: "secret-hardcoded-key.md",
}

var PrivateKeyBlock = check.Check{
	ID:       "secret-private-key-block",
	Title:    "Committed private key material",
	Severity: check.SeverityError,
	Pattern:  regexp.MustCompile(`-----BEGIN (RSA |EC )?PRIVATE KEY-----`),
	Message:  "private key material committed to file",
	DocsPath: "secret-private-key-block.md",
}

var GenericConnectionString = check.Check{
	ID:       "secret-generic-connection-string",
	Title:    "Connection string with embedded credentials",
	Severity: check.SeverityError,
	Pattern:  regexp.MustCompile(`(postgres|postgresql|mysql|mongodb(\+srv)?):\/\/[^:\/\s]+:[^@\/\s]+@`),
	Message:  "connection string contains an embedded username/password",
	DocsPath: "secret-generic-connection-string.md",
}

var JWTToken = check.Check{
	ID:       "secret-jwt-token",
	Title:    "Hardcoded JWT token",
	Severity: check.SeverityWarning,
	Pattern:  regexp.MustCompile(`eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}`),
	Message:  "hardcoded JWT-shaped token found in source",
	DocsPath: "secret-jwt-token.md",
}

var SlackWebhook = check.Check{
	ID:       "secret-slack-webhook",
	Title:    "Hardcoded Slack webhook URL",
	Severity: check.SeverityError,
	Pattern:  regexp.MustCompile(`hooks\.slack\.com/services/[A-Za-z0-9/]+`),
	Message:  "hardcoded Slack webhook URL found in source",
	DocsPath: "secret-slack-webhook.md",
}

var HighEntropyString = check.Check{
	ID:       "secret-high-entropy-string",
	Title:    "High-entropy string literal",
	Severity: check.SeverityInfo,
	Analyze:  highEntropyStrings,
	Message:  "long, high-entropy-looking string literal — verify this isn't a leaked credential",
	DocsPath: "secret-high-entropy-string.md",
}

var highEntropyLiteralPattern = regexp.MustCompile(`["'][A-Za-z0-9+/]{40,}={0,2}["']`)
var gitObjectIDPattern = regexp.MustCompile(`^(?:[A-Fa-f0-9]{40}|[A-Fa-f0-9]{64})$`)

func highEntropyStrings(_ string, content []byte) []check.Match {
	var matches []check.Match
	for i, line := range strings.Split(string(content), "\n") {
		for _, loc := range highEntropyLiteralPattern.FindAllStringIndex(line, -1) {
			value := line[loc[0]+1 : loc[1]-1]
			// Full SHA-1 and SHA-256 Git object IDs are common non-secret
			// literals. Explicit credential assignments are checked separately.
			if gitObjectIDPattern.MatchString(value) {
				continue
			}
			matches = append(matches, check.Match{Line: i + 1, Column: loc[0] + 1})
			break // Preserve one finding per line, after any ignored hashes.
		}
	}
	return matches
}

func init() {
	check.Register(HardcodedKey)
	check.Register(PrivateKeyBlock)
	check.Register(GenericConnectionString)
	check.Register(JWTToken)
	check.Register(SlackWebhook)
	check.Register(HighEntropyString)
}
