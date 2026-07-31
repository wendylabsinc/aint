package config

import (
	"os"
	"path/filepath"
	"strings"

	"aint/internal/check"
	"gopkg.in/yaml.v3"
)

type Config struct {
	FailOn      check.Severity    `yaml:"fail_on"`
	Ignore      []string          `yaml:"ignore"`
	Checks      map[string]string `yaml:"checks"`
	DocsBaseURL string            `yaml:"docs_base_url"`
}

func Default() Config {
	return Config{FailOn: check.SeverityError}
}

// Load reads and parses path as YAML, returning Default() unmodified if
// the file does not exist.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Default(), nil
	}
	if err != nil {
		return Config{}, err
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.FailOn == "" {
		cfg.FailOn = check.SeverityError
	}
	return cfg, nil
}

// IsIgnored reports whether path matches any of the configured ignore
// globs. Supports a "dir/**" prefix form and single-segment "*.ext" globs
// via filepath.Match against the basename or full path.
func (c Config) IsIgnored(path string) bool {
	path = filepath.ToSlash(path)
	for _, pattern := range c.Ignore {
		pattern = filepath.ToSlash(pattern)
		if strings.HasSuffix(pattern, "/**") {
			prefix := strings.TrimSuffix(pattern, "/**")
			if path == prefix || strings.HasPrefix(path, prefix+"/") {
				return true
			}
			continue
		}
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
	}
	return false
}

// SeverityFor returns the effective severity for chk given config
// overrides, and whether the check is enabled at all ("off" disables it).
func (c Config) SeverityFor(chk check.Check) (check.Severity, bool) {
	override, ok := c.Checks[chk.ID]
	if !ok {
		return chk.Severity, true
	}
	if override == "off" {
		return "", false
	}
	return check.Severity(override), true
}

// ResolveDocsURL builds the docs link for a finding: docsBaseURL + file if
// configured, else a relative docs/checks/<file> path.
func ResolveDocsURL(c Config, docsFile string) string {
	if c.DocsBaseURL != "" {
		return strings.TrimSuffix(c.DocsBaseURL, "/") + "/" + docsFile
	}
	return "docs/checks/" + docsFile
}
