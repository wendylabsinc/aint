// internal/hook/install.go
package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// LoadSettings reads a Claude Code settings.json or Codex hooks.json file into a generic map,
// returning an empty map (not an error) if the file doesn't exist yet.
func LoadSettings(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]interface{}{}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return map[string]interface{}{}, nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// Install merges aint's PreToolUse/Bash and PostToolUse/Write|Edit|MultiEdit
// hook entries into Claude Code or Codex settings, leaving everything else untouched. executable
// is the absolute path of the aint binary that the hooks should run. Install
// also replaces hooks written by older versions, which used a bare "aint".
// It is idempotent: re-running it against its own output changes nothing.
func Install(settings map[string]interface{}, executable string) (map[string]interface{}, []string) {
	var added []string
	preBashCommand := command(executable, "pre-bash")
	postEditCommand := command(executable, "post-edit")

	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = map[string]interface{}{}
	}

	if ensureHook(hooks, "PreToolUse", "Bash", "pre-bash", preBashCommand) {
		added = append(added, "PreToolUse/Bash -> "+preBashCommand)
	}
	if ensureHook(hooks, "PostToolUse", "Write|Edit|MultiEdit", "post-edit", postEditCommand) {
		added = append(added, "PostToolUse/Write|Edit|MultiEdit -> "+postEditCommand)
	}

	settings["hooks"] = hooks
	return settings, added
}

func command(executable, action string) string {
	return shellQuote(executable) + " hook " + action
}

func shellQuote(value string) string {
	if value != "" && !strings.ContainsAny(value, " \t\n\r'\"\\$`!&|;()<>*?[]{}") {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func ensureHook(hooks map[string]interface{}, event, matcher, action, command string) bool {
	list, _ := hooks[event].([]interface{})

	for _, entryRaw := range list {
		entry, _ := entryRaw.(map[string]interface{})
		if entry == nil {
			continue
		}
		hookList, _ := entry["hooks"].([]interface{})
		for _, hRaw := range hookList {
			h, _ := hRaw.(map[string]interface{})
			if h == nil {
				continue
			}
			if cmd, _ := h["command"].(string); cmd == command {
				return false
			} else if isAintHookCommand(cmd, action) {
				h["command"] = command
				return true
			}
		}
	}

	list = append(list, map[string]interface{}{
		"matcher": matcher,
		"hooks": []interface{}{
			map[string]interface{}{"type": "command", "command": command},
		},
	})
	hooks[event] = list
	return true
}

func isAintHookCommand(command, action string) bool {
	suffix := " hook " + action
	if !strings.HasSuffix(strings.TrimSpace(command), suffix) {
		return false
	}
	executable := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(command), suffix))
	executable = strings.Trim(executable, "'\"")
	return filepath.Base(executable) == "aint"
}

// WriteSettings writes settings back out as indented JSON.
func WriteSettings(path string, settings map[string]interface{}) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
