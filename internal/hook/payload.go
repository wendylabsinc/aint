// internal/hook/payload.go
package hook

import (
	"encoding/json"
	"strings"
)

// PreToolUsePayload is the subset of Claude Code and Codex PreToolUse hook JSON
// payload that aint needs: the tool name and the Bash command it's about
// to run.
type PreToolUsePayload struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
}

// PostToolUsePayload is the subset of Claude Code and Codex PostToolUse hook
// JSON payload that aint needs to locate the files just written.
type PostToolUsePayload struct {
	Cwd       string `json:"cwd"`
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath string `json:"file_path"`
		Command  string `json:"command"`
	} `json:"tool_input"`
}

// EditedFilePaths returns the files affected by either Claude Code's file
// editing tools or Codex's apply_patch tool. Codex sends the patch text in
// tool_input.command, so paths are read from its operation headers.
func (p PostToolUsePayload) EditedFilePaths() []string {
	if p.ToolInput.FilePath != "" {
		return []string{p.ToolInput.FilePath}
	}

	var paths []string
	seen := make(map[string]bool)
	for _, line := range strings.Split(p.ToolInput.Command, "\n") {
		for _, prefix := range []string{
			"*** Add File: ",
			"*** Update File: ",
			"*** Delete File: ",
			"*** Move to: ",
		} {
			if !strings.HasPrefix(line, prefix) {
				continue
			}
			path := strings.TrimSpace(strings.TrimPrefix(line, prefix))
			if path != "" && !seen[path] {
				paths = append(paths, path)
				seen[path] = true
			}
			break
		}
	}
	return paths
}

// ParsePreToolUse parses a PreToolUse hook payload from raw JSON.
func ParsePreToolUse(data []byte) (PreToolUsePayload, error) {
	var p PreToolUsePayload
	err := json.Unmarshal(data, &p)
	return p, err
}

// ParsePostToolUse parses a PostToolUse hook payload from raw JSON.
func ParsePostToolUse(data []byte) (PostToolUsePayload, error) {
	var p PostToolUsePayload
	err := json.Unmarshal(data, &p)
	return p, err
}
