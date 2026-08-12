package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunInstallCreatesSettingsFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude", "settings.json")

	var stdout, stderr bytes.Buffer
	code := runInstallWithIO([]string{}, path, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", code, stderr.String())
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected settings file to be created: %v", err)
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("written settings is not valid JSON: %v", err)
	}
	if _, ok := settings["hooks"]; !ok {
		t.Error("expected hooks key in written settings")
	}
}

func TestRunInstallTwiceIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude", "settings.json")

	var stdout1, stderr1 bytes.Buffer
	runInstallWithIO([]string{}, path, &stdout1, &stderr1)

	var stdout2, stderr2 bytes.Buffer
	code := runInstallWithIO([]string{}, path, &stdout2, &stderr2)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	first, _ := os.ReadFile(path)
	var stdout3, stderr3 bytes.Buffer
	runInstallWithIO([]string{}, path, &stdout3, &stderr3)
	second, _ := os.ReadFile(path)
	if string(first) != string(second) {
		t.Error("expected settings file to be byte-identical after a third install run")
	}
}

func TestRunGlobalInstallCreatesClaudeAndCodexHookFiles(t *testing.T) {
	dir := t.TempDir()
	paths := []string{
		filepath.Join(dir, ".claude", "settings.json"),
		filepath.Join(dir, ".codex", "hooks.json"),
	}

	var stdout, stderr bytes.Buffer
	code := runInstallWithPaths([]string{"--global"}, paths, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", code, stderr.String())
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("expected hook file %s to be created: %v", path, err)
		}
		var settings map[string]interface{}
		if err := json.Unmarshal(data, &settings); err != nil {
			t.Fatalf("hook file %s is not valid JSON: %v", path, err)
		}
		hooks, ok := settings["hooks"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected hooks object in %s", path)
		}
		if _, ok := hooks["PreToolUse"]; !ok {
			t.Errorf("expected PreToolUse hook in %s", path)
		}
		if _, ok := hooks["PostToolUse"]; !ok {
			t.Errorf("expected PostToolUse hook in %s", path)
		}
	}
}

func TestRunInstallUsesProvidedAbsoluteExecutable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude", "settings.json")

	var stdout, stderr bytes.Buffer
	code := runInstallWithExecutableIO(nil, path, "/tmp/custom-build/aint", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", code, stderr.String())
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatal(err)
	}
	hooks := settings["hooks"].(map[string]interface{})
	pre := hooks["PreToolUse"].([]interface{})[0].(map[string]interface{})
	command := pre["hooks"].([]interface{})[0].(map[string]interface{})["command"]
	if command != "/tmp/custom-build/aint hook pre-bash" {
		t.Errorf("expected selected executable in hook command, got %q", command)
	}
}
