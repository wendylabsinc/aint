// internal/hook/install.go
package hook

import (
	"encoding/json"
	"os"
	"strings"
)

const preBashCommand = "aint hook pre-bash"
const postEditCommand = "aint hook post-edit"

// LoadSettings reads a Claude Code settings.json file into a generic map,
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
// hook entries into settings, leaving everything else untouched. It is
// idempotent: re-running it against its own output adds nothing. Returns
// the updated settings and a human-readable list of what was added.
func Install(settings map[string]interface{}) (map[string]interface{}, []string) {
	var added []string

	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = map[string]interface{}{}
	}

	if ensureHook(hooks, "PreToolUse", "Bash", preBashCommand) {
		added = append(added, "PreToolUse/Bash -> "+preBashCommand)
	}
	if ensureHook(hooks, "PostToolUse", "Write|Edit|MultiEdit", postEditCommand) {
		added = append(added, "PostToolUse/Write|Edit|MultiEdit -> "+postEditCommand)
	}

	settings["hooks"] = hooks
	return settings, added
}

func ensureHook(hooks map[string]interface{}, event, matcher, command string) bool {
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
			if cmd, _ := h["command"].(string); strings.Contains(cmd, command) {
				return false
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

// WriteSettings writes settings back out as indented JSON.
func WriteSettings(path string, settings map[string]interface{}) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
