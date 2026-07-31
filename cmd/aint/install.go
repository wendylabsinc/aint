// cmd/aint/install.go
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"aint/internal/hook"
)

func runInstall(args []string) int {
	global := false
	for _, a := range args {
		if a == "--global" {
			global = true
		}
	}

	path := ".claude/settings.json"
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			writeErr(os.Stderr, "resolving home dir", err)
			return 1
		}
		path = filepath.Join(home, ".claude", "settings.json")
	}

	return runInstallWithIO(args, path, os.Stdout, os.Stderr)
}

func runInstallWithIO(args []string, path string, stdout, stderr io.Writer) int {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		writeErr(stderr, "creating settings directory", err)
		return 1
	}

	settings, err := hook.LoadSettings(path)
	if err != nil {
		writeErr(stderr, "loading settings", err)
		return 1
	}

	settings, added := hook.Install(settings)

	if err := hook.WriteSettings(path, settings); err != nil {
		writeErr(stderr, "writing settings", err)
		return 1
	}

	if len(added) == 0 {
		fmt.Fprintln(stdout, "aint: hooks already installed in", path)
	} else {
		fmt.Fprintln(stdout, "aint: installed into", path)
		for _, a := range added {
			fmt.Fprintln(stdout, "  +", a)
		}
	}
	return 0
}
