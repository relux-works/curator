package main

// Rework-2 regression for revision-3 finding F2 at the CLI production
// entry: `audit --allow` records the writer framing explicitly, so a new
// pin carries hash_version 2 under the v2 writer and the frozen v1 shape
// under the v1 writer. Legacy v1 pins keep authorizing v1 reads only.
// This test runs on the hosted gate (R194), like every cmd/curator suite.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
)

// The CLI pin writer records the framing in force: a v2-writer pin is a
// schema-2 carrier with hash_version 2, while a v1-writer pin keeps the
// frozen schema-1 shape with no hash_version member.
func TestAuditAllowWritesWriterVersionPinCarrier(t *testing.T) {
	digest := "sha256:" + strings.Repeat("b", 64)

	t.Run("v2-writer", func(t *testing.T) {
		pinAuditHashWriters(t, true)
		_, home := legacyProject(t)
		configPath := filepath.Join(home, "config.json")
		if code, stdout, stderr := capture(t, configPath, "audit", "--allow", digest, "--reason", "regression"); code != exitOK {
			t.Fatalf("audit --allow = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		raw := readCLIPinCarrier(t, home, digest)
		if raw["schema_version"] != float64(2) || raw["hash_version"] != float64(2) {
			t.Fatalf("CLI v2 pin carrier = schema_version %v hash_version %v, want 2/2",
				raw["schema_version"], raw["hash_version"])
		}
	})

	t.Run("v1-writer", func(t *testing.T) {
		pinAuditHashWriters(t, false)
		_, home := legacyProject(t)
		configPath := filepath.Join(home, "config.json")
		if code, stdout, stderr := capture(t, configPath, "audit", "--allow", digest, "--reason", "regression"); code != exitOK {
			t.Fatalf("audit --allow = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		raw := readCLIPinCarrier(t, home, digest)
		if raw["schema_version"] != float64(1) {
			t.Fatalf("CLI v1 pin carrier schema_version = %v, want 1", raw["schema_version"])
		}
		if _, present := raw["hash_version"]; present {
			t.Fatalf("CLI v1 pin carrier gained hash_version %v, want the frozen shape", raw["hash_version"])
		}
	})
}

func readCLIPinCarrier(t *testing.T, home, digest string) map[string]any {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(home, "audit", hashing.Normalize(digest), "trust.json"))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	return raw
}
