package main

import (
	"fmt"
	"os"
)

func main() {
	os.Exit(dispatch(os.Args))
}

func dispatch(args []string) int {
	if len(args) < 2 {
		printUsage()
		return 2
	}
	switch args[1] {
	default:
		printUsage()
		return 2
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `usage: aint <command> [args]

commands:
  check [paths...]            scan files/dirs for issues
  list                        list all registered checks
  install [--global]          wire aint into Claude Code hooks
  hook <pre-bash|post-edit>   internal: used by installed hooks`)
}
