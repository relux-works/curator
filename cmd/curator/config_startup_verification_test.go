//go:build unix

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The real file-backed CLI reads default3, changes only the selected machine
// level, preserves unrelated source values, and rejects malformed options
// without reflecting rejected values or changing source bytes.
func TestConfigStartupVerificationCLI(t *testing.T) {
	system := filepath.Join(t.TempDir(), "system.json")
	if err := os.WriteFile(system, []byte(`{"schema_version":2,"locked":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", system)
	path := filepath.Join(t.TempDir(), "selected.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":7,"skills_root":"fixture","projects":{},"preferred_locale":"en"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	source := fileConfigSource(path)
	invoke := func(args ...string) (int, string, string) {
		var out, errs bytes.Buffer
		code := run(append([]string{"config", "startup-verification"}, args...), source, &out, &errs)
		return code, out.String(), errs.String()
	}
	code, out, errs := invoke("--json")
	var projection map[string]any
	if json.Unmarshal([]byte(out), &projection) != nil || code != exitOK || projection["level"] != float64(3) || projection["origin"] != "default" || len(projection) != 3 {
		t.Fatalf("read: %d %s %s", code, out, errs)
	}
	code, out, errs = invoke("--level", "2")
	if code != exitOK || !strings.Contains(out, "2 (machine)") {
		t.Fatalf("edit: %d %s %s", code, out, errs)
	}
	before, _ := os.ReadFile(path)
	for _, args := range [][]string{{"--level", "0"}, {"--level", "3.0"}, {"--level", "3", "--level", "2"}, {"--unknown", "rejected-private-value"}, {"--json", "extra"}} {
		code, _, errs = invoke(args...)
		if code != exitUsage || strings.Contains(errs, "rejected-private-value") {
			t.Fatalf("invalid options: %d %s", code, errs)
		}
		after, _ := os.ReadFile(path)
		if !bytes.Equal(before, after) {
			t.Fatal("invalid options changed configuration")
		}
	}
}
