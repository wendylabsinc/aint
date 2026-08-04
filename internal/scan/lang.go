package scan

import "strings"

// LangForFile classifies a file path into a language tag by extension.
// Returns "" if the extension isn't recognized.
func LangForFile(path string) string {
	switch {
	case strings.HasSuffix(path, ".go"):
		return "go"
	case strings.HasSuffix(path, ".swift"):
		return "swift"
	case strings.HasSuffix(path, ".py"):
		return "python"
	case strings.HasSuffix(path, ".js"), strings.HasSuffix(path, ".ts"), strings.HasSuffix(path, ".mjs"):
		return "nodejs"
	case strings.HasSuffix(path, ".sh"):
		return "shell"
	case strings.HasSuffix(path, ".sql"):
		return "sql"
	case strings.HasSuffix(path, ".yml"), strings.HasSuffix(path, ".yaml"):
		return "yaml"
	default:
		return ""
	}
}

// MatchesLangs reports whether a check whose Langs field is langs applies
// to a file classified as fileLang. An empty/nil langs matches everything.
func MatchesLangs(langs []string, fileLang string) bool {
	if len(langs) == 0 {
		return true
	}
	for _, l := range langs {
		if l == fileLang {
			return true
		}
	}
	return false
}
