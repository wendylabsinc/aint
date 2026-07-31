// internal/scan/scan.go
package scan

import (
	"os"
	"path/filepath"

	"aint/internal/check"
	"aint/internal/config"
)

// Target is a single file or synthetic shell-command blob to run checks
// against.
type Target struct {
	Source  string
	Lang    string
	Content []byte
}

// CommandTarget wraps a raw shell command string (from a Claude Code
// PreToolUse Bash hook payload) as a Target classified as "shell".
func CommandTarget(command string) Target {
	return Target{Source: "<command>", Lang: "shell", Content: []byte(command)}
}

// Walk reads every non-ignored file under the given paths into Targets.
func Walk(paths []string, cfg config.Config) ([]Target, error) {
	var targets []Target
	for _, root := range paths {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if info.Mode()&os.ModeSymlink != 0 {
				return nil
			}
			ignoreCheck := path
			if rel, relErr := filepath.Rel(root, path); relErr == nil && rel != "." {
				ignoreCheck = rel
			}
			if cfg.IsIgnored(ignoreCheck) {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			targets = append(targets, Target{
				Source:  path,
				Lang:    LangForFile(path),
				Content: content,
			})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return targets, nil
}

// Run runs every check against every matching target, applying config
// severity overrides/disables, and returns all findings.
func Run(targets []Target, checks []check.Check, cfg config.Config) []check.Finding {
	var findings []check.Finding
	for _, c := range checks {
		sev, enabled := cfg.SeverityFor(c)
		if !enabled {
			continue
		}
		effective := c
		effective.Severity = sev
		docsURL := config.ResolveDocsURL(cfg, c.DocsPath)

		for _, t := range targets {
			if !MatchesLangs(c.Langs, t.Lang) {
				continue
			}
			findings = append(findings, effective.Run(t.Source, t.Content, docsURL)...)
		}
	}
	return findings
}
