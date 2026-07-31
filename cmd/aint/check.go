// cmd/aint/check.go
package main

import (
	"io"
	"os"
	"strings"

	"aint/internal/check"
	"aint/internal/config"
	"aint/internal/report"
	"aint/internal/scan"
)

func runCheck(args []string) int {
	return runCheckWithIO(args, os.Stdout, os.Stderr)
}

func runCheckWithIO(args []string, stdout, stderr io.Writer) int {
	format := "text"
	var paths []string
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--format" && i+1 < len(args):
			i++
			format = args[i]
		case strings.HasPrefix(args[i], "--format="):
			format = strings.TrimPrefix(args[i], "--format=")
		default:
			paths = append(paths, args[i])
		}
	}
	if len(paths) == 0 {
		paths = []string{"."}
	}

	cfg, err := config.Load(".aint.yaml")
	if err != nil {
		writeErr(stderr, "loading config", err)
		return 1
	}

	targets, err := scan.Walk(paths, cfg)
	if err != nil {
		writeErr(stderr, "scanning", err)
		return 1
	}

	findings := scan.Run(targets, check.All(), cfg)

	if format == "json" {
		if err := report.WriteJSON(stdout, findings); err != nil {
			writeErr(stderr, "writing JSON output", err)
			return 1
		}
	} else {
		report.WriteText(stdout, findings)
	}

	for _, f := range findings {
		if f.Severity.AtLeast(cfg.FailOn) {
			return 1
		}
	}
	return 0
}

func writeErr(w io.Writer, action string, err error) {
	io.WriteString(w, "aint: "+action+": "+err.Error()+"\n")
}
