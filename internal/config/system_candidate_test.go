package config

import (
	"os"
	"path/filepath"
	"testing"
)

// The freezeable policy candidate retains an explicit absent override, and
// becomes the same effective source after creation. It never needs user state.
func TestSystemCandidateIncludesAbsentPolicy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "system.json")
	t.Setenv("CURATOR_SYSTEM_CONFIG", path)
	if got := SystemCandidatePath(); got != path {
		t.Fatalf("candidate = %q", got)
	}
	if got := SystemPath(); got != path {
		t.Fatalf("override selection = %q", got)
	}
	if err := os.WriteFile(path, []byte(`{"schema_version":1}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if SystemPath() != path || SystemCandidatePath() != path {
		t.Fatal("creation changed the selected policy path")
	}
}
