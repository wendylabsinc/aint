// cmd/aint/improve.go
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"aint/internal/improve"
)

func runImprove(args []string) int {
	return runImproveWithIO(args, os.Stdout, os.Stderr, time.Now(), nil)
}

func runImproveWithIO(args []string, stdout, stderr io.Writer, now time.Time, runner improve.ClaudeRunner) int {
	home, err := os.UserHomeDir()
	if err != nil {
		writeErr(stderr, "resolving home directory", err)
		return 1
	}

	dir := filepath.Join(home, ".claude", "projects")
	statePath := filepath.Join(home, ".aint", "improve-state.json")
	outPath := ""
	claudeBin := "claude"
	limit := 50
	full := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--dir" && i+1 < len(args):
			i++
			dir = args[i]
		case strings.HasPrefix(arg, "--dir="):
			dir = strings.TrimPrefix(arg, "--dir=")
		case arg == "--out" && i+1 < len(args):
			i++
			outPath = args[i]
		case strings.HasPrefix(arg, "--out="):
			outPath = strings.TrimPrefix(arg, "--out=")
		case arg == "--state" && i+1 < len(args):
			i++
			statePath = args[i]
		case strings.HasPrefix(arg, "--state="):
			statePath = strings.TrimPrefix(arg, "--state=")
		case arg == "--claude-bin" && i+1 < len(args):
			i++
			claudeBin = args[i]
		case strings.HasPrefix(arg, "--claude-bin="):
			claudeBin = strings.TrimPrefix(arg, "--claude-bin=")
		case arg == "--limit" && i+1 < len(args):
			i++
			n, err := strconv.Atoi(args[i])
			if err != nil {
				writeErr(stderr, "parsing --limit", err)
				return 2
			}
			limit = n
		case strings.HasPrefix(arg, "--limit="):
			n, err := strconv.Atoi(strings.TrimPrefix(arg, "--limit="))
			if err != nil {
				writeErr(stderr, "parsing --limit", err)
				return 2
			}
			limit = n
		case arg == "--full":
			full = true
		default:
			writeErr(stderr, "unknown argument", fmt.Errorf("%s", arg))
			return 2
		}
	}

	if outPath == "" {
		outPath = filepath.Join(home, ".aint", "improve-reports", now.Format("2006-01-02")+".md")
	}

	state, corrupt, err := improve.LoadState(statePath)
	if err != nil {
		writeErr(stderr, "loading state", err)
		return 1
	}
	if corrupt {
		fmt.Fprintf(stderr, "aint: warning: state file %s was corrupt, starting fresh\n", statePath)
	}

	claudeUnavailable := ""
	if runner == nil {
		if _, err := exec.LookPath(claudeBin); err != nil {
			claudeUnavailable = fmt.Sprintf("claude CLI %q not found on PATH", claudeBin)
		} else {
			runner = improve.NewExecClaudeRunner(claudeBin)
		}
	}

	files, err := improve.FindSessionFiles(dir)
	if err != nil {
		writeErr(stderr, "scanning session directory", err)
		return 1
	}

	var incidents []improve.Incident
	callsUsed := 0
	candidatesFound := 0
	limitReached := false

filesLoop:
	for _, file := range files {
		startLine := 0
		if !full {
			startLine = state.Offsets[file]
		}

		// Always parse the whole file (startLine 0), never just from the
		// cached offset: this is what lets BuildContext look back at
		// assistant turns that came before a deferred/retried candidate.
		// The tradeoff is re-parsing already-seen bytes every run, which is
		// cheap CPU compared to the cost of a claude call.
		humans, assistants, totalLines, err := improve.ParseSessionFile(file, 0)
		if err != nil {
			writeErr(stderr, "parsing "+file, err)
			continue
		}

		newOffset := totalLines
		fileLimitReached := false

		for _, h := range humans {
			if h.Line <= startLine {
				continue
			}
			signals := improve.Detect(h.Text)
			if len(signals) == 0 {
				continue
			}
			candidatesFound++
			candidate := improve.Candidate{HumanMessage: h, Signals: signals}

			if claudeUnavailable == "" && callsUsed >= limit {
				newOffset = h.Line - 1
				fileLimitReached = true
				break
			}

			var incident improve.Incident
			include := true
			if claudeUnavailable != "" {
				incident = improve.Incident{Candidate: candidate, AnalysisFailed: claudeUnavailable}
			} else {
				candidate.Context = improve.BuildContext(humans, assistants, candidate)
				callsUsed++
				incident, include = improve.Analyze(runner, candidate)
			}

			if include {
				incidents = append(incidents, incident)
			}

			if incident.AnalysisFailed != "" {
				// A transient analysis failure on this candidate: pin the
				// offset just before it (so it's retried next run) and stop
				// processing this file, but don't abort the whole run —
				// other files may still have candidates worth analyzing.
				newOffset = h.Line - 1
				break
			}
		}

		if fileLimitReached {
			limitReached = true
		}

		state.Offsets[file] = newOffset
		if err := improve.SaveState(statePath, state); err != nil {
			writeErr(stderr, "saving state", err)
			return 1
		}

		if fileLimitReached {
			break filesLoop
		}
	}

	if len(incidents) == 0 {
		fmt.Fprintln(stdout, "No new incidents since last run.")
		if limitReached {
			fmt.Fprintln(stdout, "Stopped at --limit; run again to continue.")
		}
		return 0
	}

	failed := 0
	for _, inc := range incidents {
		if inc.AnalysisFailed != "" {
			failed++
		}
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0700); err != nil {
		writeErr(stderr, "creating report directory", err)
		return 1
	}
	// Append rather than truncate: outPath defaults to a date-based path, so
	// a second same-day run must not destroy the first run's report.
	f, err := os.OpenFile(outPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		writeErr(stderr, "creating report file", err)
		return 1
	}
	defer f.Close()

	if err := improve.WriteReport(f, incidents, now.Format("2006-01-02")); err != nil {
		writeErr(stderr, "writing report", err)
		return 1
	}

	fmt.Fprintf(stdout, "Found %d new incidents (%d analysis failures) from %d candidates. Report: %s\n", len(incidents), failed, candidatesFound, outPath)
	if limitReached {
		fmt.Fprintln(stdout, "Stopped at --limit; run again to continue.")
	}
	return 0
}
