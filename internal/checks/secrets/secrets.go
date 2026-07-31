// internal/checks/secrets/secrets.go
package secrets

import (
	"regexp"

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

func init() {
	check.Register(HardcodedKey)
	check.Register(PrivateKeyBlock)
}
