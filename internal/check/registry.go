// internal/check/registry.go
package check

var registry []Check

// Register adds a check to the global registry. Called from each
// checks/* package's init().
func Register(c Check) {
	registry = append(registry, c)
}

// All returns a copy of every registered check.
func All() []Check {
	out := make([]Check, len(registry))
	copy(out, registry)
	return out
}
