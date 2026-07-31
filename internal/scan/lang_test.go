package scan_test

import (
	"testing"

	"aint/internal/scan"
)

func TestLangForFile(t *testing.T) {
	cases := map[string]string{
		"main.go":       "go",
		"App.swift":     "swift",
		"script.py":     "python",
		"index.js":      "nodejs",
		"index.ts":      "nodejs",
		"deploy.sh":     "shell",
		"README.md":     "",
		"path/to/a.go":  "go",
	}
	for path, want := range cases {
		if got := scan.LangForFile(path); got != want {
			t.Errorf("LangForFile(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestMatchesLangs(t *testing.T) {
	if !scan.MatchesLangs(nil, "go") {
		t.Error("nil Langs should match any file language")
	}
	if !scan.MatchesLangs([]string{}, "shell") {
		t.Error("empty Langs should match any file language, including shell")
	}
	if !scan.MatchesLangs([]string{"go", "swift"}, "swift") {
		t.Error("swift should match when listed")
	}
	if scan.MatchesLangs([]string{"go"}, "python") {
		t.Error("python should not match when only go is listed")
	}
}
