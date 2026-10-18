package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunHookPreBashBlocksOverscopedGrant(t *testing.T) {
	stdin := bytes.NewBufferString(`{"tool_name":"Bash","tool_input":{"command":"gcloud projects add-iam-policy-binding p --role=roles/owner"}}`)
	var stdout, stderr bytes.Buffer
	code := runHookWithIO([]string{"pre-bash"}, stdin, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d (stderr: %s)", code, stderr.String())
	}
	if !bytes.Contains(stderr.Bytes(), []byte("shell-gcp-role-wildcard")) {
		t.Errorf("expected stderr to mention shell-gcp-role-wildcard, got: %s", stderr.String())
	}
}

func TestRunHookPreBashAllowsScopedGrant(t *testing.T) {
	stdin := bytes.NewBufferString(`{"tool_name":"Bash","tool_input":{"command":"gcloud projects add-iam-policy-binding p --role=roles/logging.viewer"}}`)
	var stdout, stderr bytes.Buffer
	code := runHookWithIO([]string{"pre-bash"}, stdin, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", code, stderr.String())
	}
}

func TestRunHookPostEditReportsFindingsInEditedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte("_ = err"), 0644); err != nil {
		t.Fatal(err)
	}

	stdin := bytes.NewBufferString(`{"tool_name":"Write","tool_input":{"file_path":"` + path + `"}}`)
	var stdout, stderr bytes.Buffer
	code := runHookWithIO([]string{"post-edit"}, stdin, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d (stderr: %s)", code, stderr.String())
	}
	if !bytes.Contains(stderr.Bytes(), []byte("go-ignored-error")) {
		t.Errorf("expected stderr to mention go-ignored-error, got: %s", stderr.String())
	}
}

func TestRunHookPostEditReportsFindingsFromCodexApplyPatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte("_ = err"), 0644); err != nil {
		t.Fatal(err)
	}

	stdin := bytes.NewBufferString(`{"cwd":"` + dir + `","tool_name":"apply_patch","tool_input":{"command":"*** Begin Patch\n*** Update File: main.go\n@@\n-old\n+new\n*** End Patch"}}`)
	var stdout, stderr bytes.Buffer
	code := runHookWithIO([]string{"post-edit"}, stdin, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d (stderr: %s)", code, stderr.String())
	}
	if !bytes.Contains(stderr.Bytes(), []byte("go-ignored-error")) {
		t.Errorf("expected stderr to mention go-ignored-error, got: %s", stderr.String())
	}
}

func TestRunHookPostEditDistinguishesGitHashesFromSecrets(t *testing.T) {
	const sha1 = "0123456789abcdef0123456789abcdef01234567"
	const sha256 = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	const blob = "aGVsbG8gd29ybGQgdGhpcyBpcyBhIGxvbmcgYmFzZTY0IHN0cmluZw=="
	tests := []struct {
		name      string
		content   string
		wantCode  int
		wantCheck string
	}{
		{
			name:    "commit hashes pass",
			content: "revisions = [\"" + sha1 + "\", \"" + sha256 + "\"]\n",
		},
		{
			name:      "base64 after a commit hash blocks",
			content:   "values = [\"" + sha1 + "\", \"" + blob + "\"]\n",
			wantCode:  2,
			wantCheck: "secret-high-entropy-string",
		},
		{
			name:      "hex token assignment blocks",
			content:   "token = \"" + sha1 + "\"\n",
			wantCode:  2,
			wantCheck: "secret-hardcoded-key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "sign_commits.py"), []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}
			payload, err := json.Marshal(map[string]any{
				"cwd":       dir,
				"tool_name": "apply_patch",
				"tool_input": map[string]string{
					"command": "*** Begin Patch\n*** Add File: sign_commits.py\n+" + tt.content + "*** End Patch",
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			code := runHookWithIO([]string{"post-edit"}, bytes.NewReader(payload), &stdout, &stderr)
			if code != tt.wantCode {
				t.Fatalf("expected exit code %d, got %d (stderr: %s)", tt.wantCode, code, stderr.String())
			}
			if tt.wantCheck == "" {
				if stderr.Len() != 0 {
					t.Fatalf("expected no findings, got: %s", stderr.String())
				}
			} else if !bytes.Contains(stderr.Bytes(), []byte("["+tt.wantCheck+"]")) {
				t.Errorf("expected stderr to mention %s, got: %s", tt.wantCheck, stderr.String())
			}
		})
	}
}
