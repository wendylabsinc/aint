// cmd/aint/list.go
package main

import (
	"fmt"
	"os"

	"aint/internal/check"
)

func runList(args []string) int {
	for _, c := range check.All() {
		fmt.Fprintf(os.Stdout, "%-28s %-8s %-20v %s\n", c.ID, c.Severity, c.Langs, c.Title)
	}
	return 0
}
