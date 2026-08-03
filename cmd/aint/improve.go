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

	state := improve.State{Offsets: map[string]int{}}
	if !full {
		loaded, corrupt, err := improve.LoadState(statePath)
		if err != nil {
			writeErr(stderr, "loading state", err)
			return 1
		}
		if corrupt {
			fmt.Fprintf(stderr, "aint: warning: state file %s was corrupt, starting fresh\n", statePath)
		}
		state = loaded
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

filesLoop:
	for _, file := range files {
		startLine := state.Offsets[file]
		humans, assistants, totalLines, err := improve.ParseSessionFile(file, startLine)
		if err != nil {
			writeErr(stderr, "parsing "+file, err)
			continue
		}

		newOffset := totalLines
		limitReached := false

		for _, h := range humans {
			signals := improve.Detect(h.Text)
			if len(signals) == 0 {
				continue
			}
			candidate := improve.Candidate{HumanMessage: h, Signals: signals}

			if claudeUnavailable != "" {
				incidents = append(incidents, improve.Incident{Candidate: candidate, AnalysisFailed: claudeUnavailable})
				continue
			}

			if callsUsed >= limit {
				newOffset = h.Line - 1
				limitReached = true
				break
			}

			candidate.Context = improve.BuildContext(humans, assistants, candidate)
			callsUsed++
			if incident, include := improve.Analyze(runner, candidate); include {
				incidents = append(incidents, incident)
			}
		}

		state.Offsets[file] = newOffset
		if err := improve.SaveState(statePath, state); err != nil {
			writeErr(stderr, "saving state", err)
			return 1
		}

		if limitReached {
			break filesLoop
		}
	}

	if len(incidents) == 0 {
		fmt.Fprintln(stdout, "No new incidents since last run.")
		return 0
	}

	failed := 0
	for _, inc := range incidents {
		if inc.AnalysisFailed != "" {
			failed++
		}
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		writeErr(stderr, "creating report directory", err)
		return 1
	}
	f, err := os.Create(outPath)
	if err != nil {
		writeErr(stderr, "creating report file", err)
		return 1
	}
	defer f.Close()

	if err := improve.WriteReport(f, incidents, now.Format("2006-01-02")); err != nil {
		writeErr(stderr, "writing report", err)
		return 1
	}

	fmt.Fprintf(stdout, "Found %d new incidents (%d analysis failures). Report: %s\n", len(incidents), failed, outPath)
	return 0
}
