// cmd/aint/hook.go
package main

import (
	"io"
	"os"

	"aint/internal/check"
	"aint/internal/config"
	"aint/internal/hook"
	"aint/internal/report"
	"aint/internal/scan"
)

func runHook(args []string) int {
	return runHookWithIO(args, os.Stdin, os.Stdout, os.Stderr)
}

func runHookWithIO(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		io.WriteString(stderr, "usage: aint hook <pre-bash|post-edit>\n")
		return 2
	}

	data, err := io.ReadAll(stdin)
	if err != nil {
		writeErr(stderr, "reading stdin", err)
		return 1
	}

	cfg, err := config.Load(".aint.yaml")
	if err != nil {
		writeErr(stderr, "loading config", err)
		return 1
	}

	switch args[0] {
	case "pre-bash":
		return runPreBash(data, cfg, stderr)
	case "post-edit":
		return runPostEdit(data, cfg, stderr)
	default:
		io.WriteString(stderr, "usage: aint hook <pre-bash|post-edit>\n")
		return 2
	}
}

func runPreBash(data []byte, cfg config.Config, stderr io.Writer) int {
	payload, err := hook.ParsePreToolUse(data)
	if err != nil {
		writeErr(stderr, "parsing hook payload", err)
		return 1
	}
	target := scan.CommandTarget(payload.ToolInput.Command)
	findings := scan.Run([]scan.Target{target}, check.All(), cfg)
	return reportHookFindings(findings, stderr)
}

func runPostEdit(data []byte, cfg config.Config, stderr io.Writer) int {
	payload, err := hook.ParsePostToolUse(data)
	if err != nil {
		writeErr(stderr, "parsing hook payload", err)
		return 1
	}
	content, err := os.ReadFile(payload.ToolInput.FilePath)
	if err != nil {
		writeErr(stderr, "reading file", err)
		return 1
	}
	target := scan.Target{
		Source:  payload.ToolInput.FilePath,
		Lang:    scan.LangForFile(payload.ToolInput.FilePath),
		Content: content,
	}
	findings := scan.Run([]scan.Target{target}, check.All(), cfg)
	return reportHookFindings(findings, stderr)
}

// reportHookFindings blocks on any finding at all, regardless of cfg.FailOn.
// Unlike the batch `aint check` command (which uses cfg.FailOn to decide
// pass/fail for CI), hooks run inline as Claude Code is about to act, so
// every reported finding — even Warning/Info severity ones — should be
// surfaced immediately. Users who want a check silenced in hooks can still
// disable it via .aint.yaml's `checks: <id>: off`.
func reportHookFindings(findings []check.Finding, stderr io.Writer) int {
	if len(findings) == 0 {
		return 0
	}
	report.WriteText(stderr, findings)
	return 2
}
