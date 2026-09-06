package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Schema-2 conformance drivers. Every manager-config-v2 and
// system-config-v2 case the root publishes runs through the production Load
// entry point: a user-config case loads as the machine file, a system-config
// case loads as the system file over a fixed minimal schema-2 machine file.
// The families are named explicitly, so a root that stops publishing one
// fails here instead of quietly narrowing the check. The packages are
// registered in .github/ci/root-artifacts.tsv, so the default lane defers
// them against a root predating the families and the candidate lane
// (CI_REQUIRE_FULL_ROOT=1) fails closed on a root that drops them.

type vectorCase struct {
	Name     string         `json:"name"`
	Valid    bool           `json:"valid"`
	Input    map[string]any `json:"input"`
	Expected map[string]any `json:"expected"`
}

type schemaIndexEntry struct {
	Instance string `json:"instance"`
	Schema   string `json:"schema"`
	Valid    bool   `json:"valid"`
}

func conformanceRoot(t *testing.T) string {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	return root
}

// schemaCaseValidity loads the published index and reports the recorded
// validity of each instance in the family. The index — not the file-name
// prefix — decides: an unindexed file is logged and skipped, and a family
// the root publishes with no indexed cases fails instead of passing empty.
func schemaCaseValidity(t *testing.T, root, family string) map[string]bool {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(root, "schema-cases", "index.json")) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var entries []schemaIndexEntry
	if err := json.Unmarshal(payload, &entries); err != nil {
		t.Fatal(err)
	}
	validity := map[string]bool{}
	indexed := map[string]bool{}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Instance, family+"/") {
			continue
		}
		name := strings.TrimPrefix(entry.Instance, family+"/")
		validity[name] = entry.Valid
		indexed[name] = true
	}
	disk, err := os.ReadDir(filepath.Join(root, "schema-cases", family))
	if err != nil {
		t.Fatal(err)
	}
	seen := 0
	for _, entry := range disk {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		if !indexed[entry.Name()] {
			t.Logf("schema-cases/%s/%s is published but unindexed: skipped", family, entry.Name())
			continue
		}
		seen++
	}
	if seen == 0 {
		t.Fatalf("the root publishes schema-cases/%s but it contains no cases", family)
	}
	return validity
}

// TestManagerConfigV2SchemaCases runs the published manager-config-v2
// family through Load: every indexed valid case parses and every indexed
// invalid case is rejected.
func TestManagerConfigV2SchemaCases(t *testing.T) {
	root := conformanceRoot(t)
	for name, wantValid := range schemaCaseValidity(t, root, "manager-config-v2") {
		name, wantValid := name, wantValid
		t.Run(name, func(t *testing.T) {
			payload, err := os.ReadFile(filepath.Join(root, "schema-cases", "manager-config-v2", name)) // #nosec G304 -- explicit conformance input
			if err != nil {
				t.Fatal(err)
			}
			path := writeConfig(t, t.TempDir(), "config.json", string(payload))
			_, err = Load(path, nil)
			if wantValid && err != nil {
				t.Fatalf("valid case rejected: %v", err)
			}
			if !wantValid && err == nil {
				t.Fatalf("invalid case accepted")
			}
		})
	}
}

// TestSystemConfigV2SchemaCases runs the published system-config-v2 family
// through Load as the system file over a fixed minimal schema-2 machine
// file: every indexed valid case merges and every indexed invalid case
// fails the load.
func TestSystemConfigV2SchemaCases(t *testing.T) {
	root := conformanceRoot(t)
	const user = `{"schema_version": 2, "skills_root": "/tmp/skills", "projects": {}}`
	for name, wantValid := range schemaCaseValidity(t, root, "system-config-v2") {
		name, wantValid := name, wantValid
		t.Run(name, func(t *testing.T) {
			payload, err := os.ReadFile(filepath.Join(root, "schema-cases", "system-config-v2", name)) // #nosec G304 -- explicit conformance input
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			userPath := writeConfig(t, dir, "config.json", user)
			systemPath := writeConfig(t, dir, "system.json", string(payload))
			t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
			_, err = Load(userPath, nil)
			if wantValid && err != nil {
				t.Fatalf("valid case rejected: %v", err)
			}
			if !wantValid && err == nil {
				t.Fatalf("invalid case accepted")
			}
		})
	}
}

// TestManagerConfigV2Vectors runs the published manager-config-v2 vector
// family through Parse: every valid case renders the expected effective
// configuration and every invalid case is rejected.
func TestManagerConfigV2Vectors(t *testing.T) {
	root := conformanceRoot(t)
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "manager-config-v2.json")) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var cases []vectorCase
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatalf("vectors/manager-config-v2.json publishes no cases")
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			roundTrip, err := json.Marshal(tc.Input)
			if err != nil {
				t.Fatal(err)
			}
			var object map[string]any
			if err := json.Unmarshal(roundTrip, &object); err != nil {
				t.Fatal(err)
			}
			cfg, err := Parse(object, "vector.json")
			if !tc.Valid {
				if err == nil {
					t.Fatalf("invalid vector accepted")
				}
				return
			}
			if err != nil {
				t.Fatalf("valid vector rejected: %v", err)
			}
			rendered, err := json.Marshal(cfg.EffectiveJSON())
			if err != nil {
				t.Fatal(err)
			}
			var got map[string]any
			if err := json.Unmarshal(rendered, &got); err != nil {
				t.Fatal(err)
			}
			for key, want := range tc.Expected {
				if !reflect.DeepEqual(got[key], want) {
					wantJSON, _ := json.Marshal(want)
					gotJSON, _ := json.Marshal(got[key])
					t.Fatalf("member %q:\n got %s\nwant %s", key, gotJSON, wantJSON)
				}
			}
		})
	}
}
