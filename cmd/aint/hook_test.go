package main

import (
	"bytes"
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
