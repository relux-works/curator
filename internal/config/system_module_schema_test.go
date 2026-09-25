// E2 system-module schema cases (environments §12.1, §12.2): the seven
// schema cases the finding publishes — five manager-config-v2 and two
// system-config-v2 — run here, selected by name, with exact published bytes
// through the production Load entry point.
//
// The general consumers (TestManagerConfigV2SchemaCases,
// TestSystemConfigV2SchemaCases) iterate whatever the root's index lists,
// so a root predating the E2 cases runs them green with no E2 case at
// all. This driver is the explicit subset accounting: it runs every E2
// case, and fails loud when the root publishes only some of them — or
// none, since the committed pin serves the family and a root-content
// skip would hide a regression.
//
// An invalid E2 case must be rejected for its E2 reason, not merely
// rejected: a sibling knob the published bytes also carry (E4's
// provider_directories at the current root) must not turn the case into
// a false green. The rejection has to name the E2 knob.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// systemModuleSchemaCases names the E2 schema cases by family and file,
// with the knob substring an invalid case's rejection must name. The
// index — not this table — decides each case's validity.
var systemModuleSchemaCases = []struct {
	family string
	file   string
	knob   string
	valid  bool
}{
	{"manager-config-v2", "valid-system-module-waiver.json", "", true},
	{"manager-config-v2", "invalid-transitive-system-modules-value.json", "transitive_system_modules", false},
	{"manager-config-v2", "invalid-system-module-waiver-missing-reason.json", "system_module_waivers", false},
	{"manager-config-v2", "invalid-system-module-waiver-package-grammar.json", "system_module_waivers", false},
	{"manager-config-v2", "invalid-system-module-waiver-unknown-field.json", "system_module_waivers", false},
	{"system-config-v2", "invalid-transitive-system-modules-drop-direction.json", "transitive_system_modules", false},
	{"system-config-v2", "invalid-transitive-system-modules-value.json", "transitive_system_modules", false},
}

// TestSystemModuleSchemaSubset requires all seven E2 cases to remain
// published. The counted manager/system family tests drive their exact bytes
// through Load and assert both rejection and the E2 diagnostic.
func TestSystemModuleSchemaSubset(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	indexPath := filepath.Join(root, "schema-cases", "index.json")
	payload, err := os.ReadFile(indexPath) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var entries []schemaIndexEntry
	if err := json.Unmarshal(payload, &entries); err != nil {
		t.Fatal(err)
	}
	validity := map[string]bool{}
	for _, entry := range entries {
		validity[entry.Instance] = entry.Valid
	}
	var missing []string
	present := 0
	for _, tc := range systemModuleSchemaCases {
		if valid, ok := validity[tc.family+"/"+tc.file]; ok {
			if valid != tc.valid {
				t.Errorf("%s/%s validity = %v, want %v", tc.family, tc.file, valid, tc.valid)
			}
			present++
		} else {
			missing = append(missing, tc.family+"/"+tc.file)
		}
	}
	if len(missing) != 0 {
		t.Fatalf("conformance root %s publishes only %d of %d system-module schema cases, missing %s",
			root, present, len(systemModuleSchemaCases), strings.Join(missing, ", "))
	}
}
