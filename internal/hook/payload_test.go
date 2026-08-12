// internal/hook/payload_test.go
package hook_test

import (
	"testing"

	"aint/internal/hook"
)

func TestParsePreToolUse(t *testing.T) {
	data := []byte(`{"tool_name":"Bash","tool_input":{"command":"chmod 777 x.sh"}}`)
	p, err := hook.ParsePreToolUse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ToolInput.Command != "chmod 777 x.sh" {
		t.Errorf("unexpected command: %q", p.ToolInput.Command)
	}
}

func TestParsePostToolUse(t *testing.T) {
	data := []byte(`{"tool_name":"Write","tool_input":{"file_path":"/tmp/main.go"}}`)
	p, err := hook.ParsePostToolUse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ToolInput.FilePath != "/tmp/main.go" {
		t.Errorf("unexpected file path: %q", p.ToolInput.FilePath)
	}
}

func TestPostToolUseEditedFilePathsFromCodexApplyPatch(t *testing.T) {
	data := []byte(`{
		"tool_name":"apply_patch",
		"tool_input":{"command":"*** Begin Patch\n*** Update File: old.go\n*** Move to: moved.go\n@@\n-old\n+new\n*** Add File: added.go\n+package added\n*** Delete File: removed.go\n*** End Patch"}
	}`)
	p, err := hook.ParsePostToolUse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := p.EditedFilePaths()
	want := []string{"old.go", "moved.go", "added.go", "removed.go"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}
