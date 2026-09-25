package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
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

type namedSchemaCase struct {
	Name  string
	Valid bool
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
func schemaCaseValidity(t *testing.T, root, family string) []namedSchemaCase {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(root, "schema-cases", "index.json")) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var entries []schemaIndexEntry
	if err := json.Unmarshal(payload, &entries); err != nil {
		t.Fatal(err)
	}
	var cases []namedSchemaCase
	indexed := map[string]bool{}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Instance, family+"/") {
			continue
		}
		name := strings.TrimPrefix(entry.Instance, family+"/")
		cases = append(cases, namedSchemaCase{Name: name, Valid: entry.Valid})
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
	return cases
}

// TestManagerConfigV2SchemaCases runs the published manager-config-v2
// family through Load: every indexed valid case parses and every indexed
// invalid case is rejected.
func TestManagerConfigV2SchemaCases(t *testing.T) {
	root := conformanceRoot(t)
	cases := schemaCaseValidity(t, root, "manager-config-v2")
	conformancecoverage.RunOutcomes(t, "manager-config-v2/schema-cases", cases,
		func(tc namedSchemaCase) string { return tc.Name }, func(t *testing.T, tc namedSchemaCase) conformancecoverage.Observation {
			payload, err := os.ReadFile(filepath.Join(root, "schema-cases", "manager-config-v2", tc.Name)) // #nosec G304 -- explicit conformance input
			if err != nil {
				t.Fatal(err)
			}
			path := writeConfig(t, t.TempDir(), "config.json", string(payload))
			_, err = Load(path, nil)
			if tc.Valid && err != nil {
				return conformancecoverage.Observation{FailureReason: fmt.Sprintf("Load rejected published-valid case: %v", err)}
			}
			if !tc.Valid && err == nil {
				return conformancecoverage.Observation{FailureReason: "Load accepted published-invalid case"}
			}
			if !tc.Valid {
				if reason := systemModuleSchemaFailure("manager-config-v2", tc.Name, err); reason != "" {
					return conformancecoverage.Observation{FailureReason: reason}
				}
			}
			return conformancecoverage.Observation{}
		})
}

// TestSystemConfigV2SchemaCases runs the published system-config-v2 family
// through Load as the system file over a fixed minimal schema-2 machine
// file: every indexed valid case merges and every indexed invalid case
// fails the load.
func TestSystemConfigV2SchemaCases(t *testing.T) {
	root := conformanceRoot(t)
	const user = `{"schema_version": 2, "skills_root": "/tmp/skills", "projects": {}}`
	cases := schemaCaseValidity(t, root, "system-config-v2")
	conformancecoverage.RunOutcomes(t, "system-config-v2/schema-cases", cases,
		func(tc namedSchemaCase) string { return tc.Name }, func(t *testing.T, tc namedSchemaCase) conformancecoverage.Observation {
			payload, err := os.ReadFile(filepath.Join(root, "schema-cases", "system-config-v2", tc.Name)) // #nosec G304 -- explicit conformance input
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			userPath := writeConfig(t, dir, "config.json", user)
			systemPath := writeConfig(t, dir, "system.json", string(payload))
			t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
			_, err = Load(userPath, nil)
			if tc.Valid && err != nil {
				return conformancecoverage.Observation{FailureReason: fmt.Sprintf("Load rejected published-valid case: %v", err)}
			}
			if !tc.Valid && err == nil {
				return conformancecoverage.Observation{FailureReason: "Load accepted published-invalid case"}
			}
			if !tc.Valid {
				if reason := systemModuleSchemaFailure("system-config-v2", tc.Name, err); reason != "" {
					return conformancecoverage.Observation{FailureReason: reason}
				}
			}
			return conformancecoverage.Observation{}
		})
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
	conformancecoverage.RunOutcomes(t, "manager-config-v2/vectors", cases,
		func(tc vectorCase) string { return tc.Name }, func(t *testing.T, tc vectorCase) conformancecoverage.Observation {
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
					return conformancecoverage.Observation{FailureReason: "Parse accepted published-invalid vector"}
				}
				return conformancecoverage.Observation{}
			}
			if err != nil {
				return conformancecoverage.Observation{FailureReason: fmt.Sprintf("Parse rejected published-valid vector: %v", err)}
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
				gotValue := got[key]
				if !reflect.DeepEqual(gotValue, want) {
					wantJSON, _ := json.Marshal(want)
					gotJSON, _ := json.Marshal(gotValue)
					return conformancecoverage.Observation{FailureReason: fmt.Sprintf("effective member %q got %s, want %s", key, gotJSON, wantJSON)}
				}
			}
			return conformancecoverage.Observation{}
		})
}

// systemModuleSchemaFailure keeps the E2 schema cases' stronger diagnostic
// assertion in the counted family driver. This avoids running the same
// published case a second time in TestSystemModuleSchemaSubset and ensures a
// failed diagnostic is attributed by the coverage ledger like any other case.
func systemModuleSchemaFailure(family, file string, err error) string {
	for _, tc := range systemModuleSchemaCases {
		if tc.family != family || tc.file != file || tc.knob == "" || strings.Contains(err.Error(), tc.knob) {
			continue
		}
		return fmt.Sprintf("Load rejected the E2 case for %q instead of naming %q: %v", tc.knob, tc.knob, err)
	}
	return ""
}

// TestOverlayGapOwnersMatchFirstProductionBlocker keeps the 28 legacy
// path-kind overlay rows attributed to the new fields that actually block
// them at the pinned root. Schema cases are driven through Load; vectors are
// driven through Parse and compared with their published EffectiveJSON.
func TestOverlayGapOwnersMatchFirstProductionBlocker(t *testing.T) {
	root := conformanceRoot(t)
	_, gaps, err := conformancecoverage.Load()
	if err != nil {
		t.Fatal(err)
	}
	gapByCase := make(map[string]conformancecoverage.Gap, len(gaps))
	for _, gap := range gaps {
		gapByCase[gap.Family+"\x00"+gap.CaseID] = gap
	}
	lookup := func(family, caseID string) conformancecoverage.Gap {
		t.Helper()
		gap, ok := gapByCase[family+"\x00"+caseID]
		if !ok {
			t.Fatalf("missing owner row for %s/%s", family, caseID)
		}
		return gap
	}

	const managerSchema = "manager-config-v2/schema-cases"
	schemaCases := 0
	for _, tc := range schemaCaseValidity(t, root, "manager-config-v2") {
		if !strings.HasPrefix(tc.Name, "valid-overlay-") {
			continue
		}
		schemaCases++
		if !tc.Valid {
			t.Errorf("published overlay case %s is not marked valid", tc.Name)
			continue
		}
		payload, err := os.ReadFile(filepath.Join(root, "schema-cases", "manager-config-v2", tc.Name)) // #nosec G304 -- explicit conformance input
		if err != nil {
			t.Fatal(err)
		}
		var input map[string]any
		if err := json.Unmarshal(payload, &input); err != nil {
			t.Fatal(err)
		}
		environments, ok := input["environments"].(map[string]any)
		if !ok {
			t.Errorf("%s has no environments object", tc.Name)
			continue
		}
		for _, field := range []string{"permissions", "source_signers", "require_source_signers"} {
			if _, exists := environments[field]; !exists {
				t.Errorf("%s does not carry the new %q blocker whose owner is being verified", tc.Name, field)
			}
		}
		_, loadErr := Load(writeConfig(t, t.TempDir(), "config.json", string(payload)), nil)
		if loadErr == nil {
			t.Errorf("Load now accepts published overlay case %s; remove its gap row", tc.Name)
			continue
		}
		owner, field := overlaySchemaFailureOwner(loadErr)
		if owner == "" {
			t.Errorf("Load's first blocker for %s is not a known new field: %v", tc.Name, loadErr)
			continue
		}
		gap := lookup(managerSchema, tc.Name)
		const wantOwners = "STORY-260916-ioemse+STORY-260922-1cenbr"
		if gap.Owner != wantOwners {
			t.Errorf("%s owner = %q, the case contains both E1 and 0017/0018 fields; want %q", tc.Name, gap.Owner, wantOwners)
		}
		if !strings.Contains(loadErr.Error(), field) {
			t.Errorf("%s first Load blocker %q is not present in observed error %v", tc.Name, field, loadErr)
		}
		for _, requiredField := range []string{"permissions", "source_signers", "require_source_signers"} {
			if !strings.Contains(gap.Reason, requiredField) {
				t.Errorf("%s reason %q omits unsupported field %q", tc.Name, gap.Reason, requiredField)
			}
		}
		if strings.Contains(gap.Reason, "path-kind") {
			t.Errorf("%s reason %q misattributes the Load failure to path-kind", tc.Name, gap.Reason)
		}
	}
	if schemaCases != 14 {
		t.Fatalf("published valid overlay schema cases = %d, want 14", schemaCases)
	}

	const managerVectors = "manager-config-v2/vectors"
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "manager-config-v2.json")) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var vectors []vectorCase
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	selectedVectors := map[string]bool{}
	for _, gap := range gaps {
		if gap.Family == managerVectors && strings.HasPrefix(gap.CaseID, "schema2-overlay-") {
			selectedVectors[gap.CaseID] = true
		}
	}
	vectorCases := 0
	for _, tc := range vectors {
		if !selectedVectors[tc.Name] {
			continue
		}
		vectorCases++
		encoded, err := json.Marshal(tc.Input)
		if err != nil {
			t.Fatal(err)
		}
		var object map[string]any
		if err := json.Unmarshal(encoded, &object); err != nil {
			t.Fatal(err)
		}
		cfg, parseErr := Parse(object, "vector.json")
		if parseErr != nil {
			t.Errorf("Parse's first blocker for %s is %v; expected the published output comparison to identify its owners", tc.Name, parseErr)
			continue
		}
		rendered, err := json.Marshal(cfg.EffectiveJSON())
		if err != nil {
			t.Fatal(err)
		}
		var got map[string]any
		if err := json.Unmarshal(rendered, &got); err != nil {
			t.Fatal(err)
		}
		gotEnv, gotOK := got["environments"].(map[string]any)
		wantEnv, wantOK := tc.Expected["environments"].(map[string]any)
		if !gotOK || !wantOK {
			t.Errorf("%s lacks the environments EffectiveJSON object", tc.Name)
			continue
		}
		for _, field := range []string{"permissions", "source_signers", "require_source_signers"} {
			if _, exists := wantEnv[field]; !exists {
				t.Errorf("%s expected EffectiveJSON omits the published new %q field whose owner is being verified", tc.Name, field)
			}
		}
		differences := jsonDifferencePaths("environments", gotEnv, wantEnv)
		changed := map[string]bool{}
		for _, path := range differences {
			switch {
			case strings.Contains(path, "permissions"):
				changed["permissions"] = true
			case strings.Contains(path, "source_signers"), strings.Contains(path, "require_source_signers"):
				changed["source_signers"] = true
			default:
				t.Errorf("%s differs at unrelated EffectiveJSON field %s", tc.Name, path)
			}
		}
		if !changed["permissions"] || !changed["source_signers"] {
			t.Errorf("%s differences = %v, want both permissions and source_signers fields", tc.Name, differences)
			continue
		}
		gap := lookup(managerVectors, tc.Name)
		const wantOwners = "STORY-260916-ioemse+STORY-260922-1cenbr"
		if gap.Owner != wantOwners {
			t.Errorf("%s owner = %q, changed fields are owned by %q", tc.Name, gap.Owner, wantOwners)
		}
		if !strings.Contains(gap.Reason, "permissions") || !strings.Contains(gap.Reason, "source_signers") || strings.Contains(gap.Reason, "path-kind") {
			t.Errorf("%s reason %q does not describe both changed fields", tc.Name, gap.Reason)
		}
	}
	if vectorCases != 14 {
		t.Fatalf("published valid overlay vectors = %d, want 14", vectorCases)
	}
}

func overlaySchemaFailureOwner(err error) (owner, field string) {
	message := err.Error()
	switch {
	case strings.Contains(message, `unsupported field "permissions"`):
		return "STORY-260922-1cenbr", "permissions"
	case strings.Contains(message, `unsupported field "source_signers"`):
		return "STORY-260916-ioemse", "source_signers"
	case strings.Contains(message, `unsupported field "require_source_signers"`):
		return "STORY-260916-ioemse", "require_source_signers"
	default:
		return "", ""
	}
}

func jsonDifferencePaths(prefix string, got, want any) []string {
	gotObject, gotMap := got.(map[string]any)
	wantObject, wantMap := want.(map[string]any)
	if !gotMap || !wantMap {
		if reflect.DeepEqual(got, want) {
			return nil
		}
		return []string{prefix}
	}
	keys := make(map[string]bool, len(gotObject)+len(wantObject))
	for key := range gotObject {
		keys[key] = true
	}
	for key := range wantObject {
		keys[key] = true
	}
	ordered := make([]string, 0, len(keys))
	for key := range keys {
		ordered = append(ordered, key)
	}
	sort.Strings(ordered)
	var paths []string
	for _, key := range ordered {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		gotValue, gotOK := gotObject[key]
		wantValue, wantOK := wantObject[key]
		if !gotOK || !wantOK {
			paths = append(paths, path)
			continue
		}
		paths = append(paths, jsonDifferencePaths(path, gotValue, wantValue)...)
	}
	return paths
}
