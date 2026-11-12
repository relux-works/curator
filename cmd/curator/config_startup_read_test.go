package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Read/default and usage behavior is portable even where safe file editing is
// unqualified. Invalid options preserve the valid source and omit rejected text.
func TestConfigStartupReadCLI(t *testing.T) {
	system := filepath.Join(t.TempDir(), "system.json")
	if err := os.WriteFile(system, []byte(`{"schema_version":2,"locked":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_SYSTEM_CONFIG", system)
	path := filepath.Join(t.TempDir(), "machine.json")
	raw := []byte(`{"schema_version":7,"skills_root":"fixture","projects":{}}`)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	source := fileConfigSource(path)
	var out, errs bytes.Buffer
	code := run([]string{"config", "startup-verification", "--json"}, source, &out, &errs)
	var p map[string]any
	if code != exitOK || json.Unmarshal(out.Bytes(), &p) != nil || len(p) != 3 || p["level"] != float64(3) || p["origin"] != "default" {
		t.Fatalf("default read: %d %s %s", code, out.String(), errs.String())
	}
	for _, args := range [][]string{{"--level", "3.0"}, {"--unknown", "rejected-private-value"}, {"--json", "extra"}} {
		out.Reset()
		errs.Reset()
		code = run(append([]string{"config", "startup-verification"}, args...), source, &out, &errs)
		if code != exitUsage || strings.Contains(errs.String(), "rejected-private-value") {
			t.Fatalf("usage: %d %s", code, errs.String())
		}
		after, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(raw, after) {
			t.Fatal("invalid options changed source")
		}
	}
}
