package improve_test

import (
	"os"
	"path/filepath"
	"testing"

	"aint/internal/improve"
)

func TestLoadStateMissingFileReturnsEmpty(t *testing.T) {
	state, corrupt, err := improve.LoadState(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if corrupt {
		t.Error("expected corrupt=false for a missing file")
	}
	if len(state.Offsets) != 0 {
		t.Errorf("expected empty offsets, got %v", state.Offsets)
	}
}

func TestLoadStateCorruptFileReturnsEmptyAndFlag(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := os.WriteFile(path, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	state, corrupt, err := improve.LoadState(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !corrupt {
		t.Error("expected corrupt=true for an unparsable file")
	}
	if len(state.Offsets) != 0 {
		t.Errorf("expected empty offsets, got %v", state.Offsets)
	}
}

func TestSaveAndLoadStateRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "state.json")
	want := improve.State{Offsets: map[string]int{"a.jsonl": 5, "b.jsonl": 12}}

	if err := improve.SaveState(path, want); err != nil {
		t.Fatalf("unexpected error saving: %v", err)
	}

	got, corrupt, err := improve.LoadState(path)
	if err != nil {
		t.Fatalf("unexpected error loading: %v", err)
	}
	if corrupt {
		t.Error("expected corrupt=false after a valid save")
	}
	if got.Offsets["a.jsonl"] != 5 || got.Offsets["b.jsonl"] != 12 {
		t.Errorf("unexpected offsets after round trip: %v", got.Offsets)
	}
}
