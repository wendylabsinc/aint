// internal/hook/payload.go
package hook

import "encoding/json"

// PreToolUsePayload is the subset of Claude Code's PreToolUse hook JSON
// payload that aint needs: the tool name and the Bash command it's about
// to run.
type PreToolUsePayload struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
}

// PostToolUsePayload is the subset of Claude Code's PostToolUse hook JSON
// payload that aint needs: the tool name and the file path it just wrote.
type PostToolUsePayload struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
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
