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

	paths := []string{".claude/settings.json"}
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			writeErr(os.Stderr, "resolving home dir", err)
			return 1
		}
		paths = []string{
			filepath.Join(home, ".claude", "settings.json"),
			filepath.Join(home, ".codex", "hooks.json"),
		}
	}

	return runInstallWithPaths(args, paths, os.Stdout, os.Stderr)
}

func runInstallWithIO(args []string, path string, stdout, stderr io.Writer) int {
	return runInstallWithPaths(args, []string{path}, stdout, stderr)
}

func runInstallWithPaths(args []string, paths []string, stdout, stderr io.Writer) int {
	executable, err := os.Executable()
	if err != nil {
		writeErr(stderr, "resolving aint executable", err)
		return 1
	}
	executable, err = filepath.Abs(executable)
	if err != nil {
		writeErr(stderr, "resolving absolute aint executable path", err)
		return 1
	}
	return runInstallWithExecutablePathsIO(args, paths, executable, stdout, stderr)
}

func runInstallWithExecutableIO(args []string, path, executable string, stdout, stderr io.Writer) int {
	return runInstallWithExecutablePathsIO(args, []string{path}, executable, stdout, stderr)
}

func runInstallWithExecutablePathsIO(args []string, paths []string, executable string, stdout, stderr io.Writer) int {
	installedCodexHooks := false
	for _, path := range paths {
		if code := installHooks(path, executable, stdout, stderr); code != 0 {
			return code
		}
		if filepath.Base(path) == "hooks.json" && filepath.Base(filepath.Dir(path)) == ".codex" {
			installedCodexHooks = true
		}
	}
	if installedCodexHooks {
		fmt.Fprintln(stdout, "aint: open /hooks in Codex to review and trust the installed hooks")
	}
	return 0
}

func installHooks(path, executable string, stdout, stderr io.Writer) int {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		writeErr(stderr, "creating settings directory", err)
		return 1
	}

	settings, err := hook.LoadSettings(path)
	if err != nil {
		writeErr(stderr, "loading settings", err)
		return 1
	}

	settings, added := hook.Install(settings, executable)

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
