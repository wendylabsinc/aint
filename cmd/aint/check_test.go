// cmd/aint/check_test.go
package main

import (
	"bytes"
	"testing"
)

func TestRunCheckFindsSeedViolationAndReturnsNonZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCheckWithIO([]string{"testdata/checkfixture/main.go"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d (stderr: %s)", code, stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte("go-ignored-error")) {
		t.Errorf("expected output to mention go-ignored-error, got: %s", stdout.String())
	}
}

func TestRunCheckOnCleanFileReturnsZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCheckWithIO([]string{"testdata/checkfixture/clean.go"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
	}
}
