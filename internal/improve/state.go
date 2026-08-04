// internal/improve/state.go
package improve

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// State is the offset cache: for each session file path, the last line
// number that has already been read and heuristic-checked. A file absent
// from Offsets has never been processed.
type State struct {
	Offsets map[string]int `json:"offsets"`
}

// LoadState reads path as JSON. A missing file returns an empty State,
// corrupt=false, err=nil. An unparsable file returns an empty State,
// corrupt=true, err=nil — callers should warn but keep going rather than
// fail the run over a cache-only file. Any other read error is returned as
// err.
func LoadState(path string) (State, bool, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return State{Offsets: map[string]int{}}, false, nil
	}
	if err != nil {
		return State{Offsets: map[string]int{}}, false, err
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return State{Offsets: map[string]int{}}, true, nil
	}
	if s.Offsets == nil {
		s.Offsets = map[string]int{}
	}
	return s, false, nil
}

// SaveState writes state to path as indented JSON, creating parent
// directories as needed.
func SaveState(path string, state State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
