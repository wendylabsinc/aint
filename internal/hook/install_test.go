package hook_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"aint/internal/hook"
)

func TestInstallOnEmptySettingsAddsBothHooks(t *testing.T) {
	settings := map[string]interface{}{}
	updated, added := hook.Install(settings, "/opt/aint/bin/aint")

	if len(added) != 2 {
		t.Fatalf("expected 2 additions, got %d: %v", len(added), added)
	}

	hooks := updated["hooks"].(map[string]interface{})
	if _, ok := hooks["PreToolUse"]; !ok {
		t.Error("expected PreToolUse to be set")
	}
	if _, ok := hooks["PostToolUse"]; !ok {
		t.Error("expected PostToolUse to be set")
	}
	pre := hooks["PreToolUse"].([]interface{})[0].(map[string]interface{})
	command := pre["hooks"].([]interface{})[0].(map[string]interface{})["command"]
	if command != "/opt/aint/bin/aint hook pre-bash" {
		t.Errorf("expected absolute executable path in hook command, got %q", command)
	}
}

func TestInstallIsIdempotent(t *testing.T) {
	settings := map[string]interface{}{}
	updated, _ := hook.Install(settings, "/opt/aint/bin/aint")
	_, addedSecondRun := hook.Install(updated, "/opt/aint/bin/aint")

	if len(addedSecondRun) != 0 {
		t.Fatalf("expected no additions on second run, got %v", addedSecondRun)
	}
}

func TestInstallPreservesUnrelatedSettings(t *testing.T) {
	settings := map[string]interface{}{
		"model": "claude-sonnet-5",
		"hooks": map[string]interface{}{
			"PreToolUse": []interface{}{
				map[string]interface{}{
					"matcher": "Read",
					"hooks": []interface{}{
						map[string]interface{}{"type": "command", "command": "some-other-tool check"},
					},
				},
			},
		},
	}
	updated, added := hook.Install(settings, "/opt/aint/bin/aint")

	if updated["model"] != "claude-sonnet-5" {
		t.Error("expected unrelated top-level key to survive")
	}
	pre := updated["hooks"].(map[string]interface{})["PreToolUse"].([]interface{})
	if len(pre) != 2 {
		t.Fatalf("expected existing PreToolUse entry preserved plus aint's own, got %d entries", len(pre))
	}
	if len(added) != 2 {
		t.Fatalf("expected 2 additions (Bash pre and edit post), got %d: %v", len(added), added)
	}
}

func TestInstallQuotesExecutablePathAndReplacesLegacyCommands(t *testing.T) {
	settings := map[string]interface{}{
		"hooks": map[string]interface{}{
			"PreToolUse": []interface{}{
				map[string]interface{}{
					"matcher": "Bash",
					"hooks": []interface{}{
						map[string]interface{}{"type": "command", "command": "aint hook pre-bash"},
					},
				},
			},
		},
	}

	updated, changed := hook.Install(settings, "/Applications/Aint Tools/aint")
	if len(changed) != 2 {
		t.Fatalf("expected legacy pre hook to be replaced and post hook added, got %v", changed)
	}

	hooks := updated["hooks"].(map[string]interface{})
	pre := hooks["PreToolUse"].([]interface{})
	if len(pre) != 1 {
		t.Fatalf("expected legacy hook to be replaced in place, got %d entries", len(pre))
	}
	entry := pre[0].(map[string]interface{})
	command := entry["hooks"].([]interface{})[0].(map[string]interface{})["command"]
	if command != "'/Applications/Aint Tools/aint' hook pre-bash" {
		t.Errorf("unexpected quoted command: %q", command)
	}
}

func TestLoadSettingsReturnsEmptyMapWhenFileMissing(t *testing.T) {
	settings, err := hook.LoadSettings(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(settings) != 0 {
		t.Errorf("expected empty settings, got %v", settings)
	}
}

func TestWriteSettingsThenLoadRoundTrips(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	settings := map[string]interface{}{"model": "claude-sonnet-5"}

	if err := hook.WriteSettings(path, settings); err != nil {
		t.Fatalf("unexpected error writing: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("unexpected error reading back: %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("written file is not valid JSON: %v", err)
	}
	if decoded["model"] != "claude-sonnet-5" {
		t.Errorf("unexpected round-tripped content: %v", decoded)
	}
}
