package config

import (
	"bytes"
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

type sourceSignerMergeVectorFile struct {
	MergeCases []struct {
		Name     string                         `json:"name"`
		Locked   bool                           `json:"locked"`
		System   map[string][]sourceSignerWire  `json:"system"`
		Machine  *map[string][]sourceSignerWire `json:"machine"`
		Expected struct {
			Effective map[string][]sourceSignerWire `json:"effective"`
			Warnings  []string                      `json:"warnings"`
		} `json:"expected"`
	} `json:"merge_cases"`
}

type sourceSignerWire struct {
	Type        string `json:"type"`
	Key         string `json:"key,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

// TestSourceSignerMergeVectorsAtLoad drives every rc.13 source_signers
// precedence vector through Load, including per-source system locks and
// unlocked whole-knob defaults (manager §1, environments §12.1).
func TestSourceSignerMergeVectorsAtLoad(t *testing.T) {
	root := conformanceRoot(t)
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "environments-source-signers.json")) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var vectors sourceSignerMergeVectorFile
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	if len(vectors.MergeCases) != 6 {
		t.Fatalf("pinned rc.13 merge case count = %d, want 6", len(vectors.MergeCases))
	}
	for _, tc := range vectors.MergeCases {
		t.Run(tc.Name, func(t *testing.T) {
			normalizeMergeVectorKeys(tc.System)
			if tc.Machine != nil {
				normalizeMergeVectorKeys(*tc.Machine)
			}
			normalizeMergeVectorKeys(tc.Expected.Effective)
			userEnv := map[string]any{}
			if tc.Machine != nil {
				userEnv["source_signers"] = *tc.Machine
			}
			userPayload, err := json.Marshal(map[string]any{
				"schema_version": 2, "skills_root": "/tmp/skills", "projects": map[string]any{}, "environments": userEnv,
			})
			if err != nil {
				t.Fatal(err)
			}
			systemEnv := map[string]any{"source_signers": tc.System}
			systemData := map[string]any{"schema_version": 2, "environments": systemEnv}
			if tc.Locked {
				systemData["locked"] = []string{"environments.source_signers"}
			}
			systemPayload, err := json.Marshal(systemData)
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			userPath := writeConfig(t, dir, "config.json", string(userPayload))
			systemPath := writeConfig(t, dir, "system.json", string(systemPayload))
			t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
			var warnings []string
			cfg, err := Load(userPath, func(message string) { warnings = append(warnings, message) })
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			got := cfg.Env.SourceSigners
			if got == nil {
				got = map[string][]SourceSigner{}
			}
			want := decodeSourceSignerWire(tc.Expected.Effective)
			if want == nil {
				want = map[string][]SourceSigner{}
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("effective source_signers = %#v, want %#v", got, want)
			}
			if len(warnings) != len(tc.Expected.Warnings) {
				t.Fatalf("Load warnings = %v, want source suffixes %v", warnings, tc.Expected.Warnings)
			}
			for i, suffix := range tc.Expected.Warnings {
				if !strings.Contains(warnings[i], suffix) {
					t.Errorf("warning %q does not name expected source %q", warnings[i], suffix)
				}
			}
		})
	}
}

func decodeSourceSignerWire(wire map[string][]sourceSignerWire) map[string][]SourceSigner {
	if wire == nil {
		return nil
	}
	decoded := make(map[string][]SourceSigner, len(wire))
	for source, signers := range wire {
		for _, signer := range signers {
			decoded[source] = append(decoded[source], SourceSigner(signer))
		}
	}
	return decoded
}

func normalizeMergeVectorKeys(signers map[string][]sourceSignerWire) {
	const (
		operatorKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIM2viPc9tcg5ax615zmMJHL/WEOIxD87zH2o+/Znd0Ga operator@example"
		intruderKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAINZFvyk44Hs0dZ8KKN5icz5Hv56DJuhxW6/eRroAMHVG intruder@example"
	)
	for source, list := range signers {
		for i := range list {
			if list[i].Type != "ssh" {
				continue
			}
			if strings.HasSuffix(list[i].Key, "intruder@example") {
				list[i].Key = intruderKey
			} else {
				list[i].Key = operatorKey
			}
		}
		signers[source] = list
	}
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

// TestSystemConfigV2IsolationDirectionsFromPinnedCases drives the isolation
// members of both published system-config-v2 direction cases through Load.
// The full cases are also driven above; at the current pin those documents
// carry source_signers and permissions locks outside this implementation's
// task scope, so this companion projects only the published isolation knob
// and its required locked entry rather than treating those unrelated gaps as
// evidence about isolation.
func TestSystemConfigV2IsolationDirectionsFromPinnedCases(t *testing.T) {
	root := conformanceRoot(t)
	for _, tc := range []struct {
		name string
		mode string
	}{
		{name: "valid-isolation-shared-direction.json", mode: "shared"},
		{name: "valid-isolation-isolated-direction.json", mode: "isolated"},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			payload, err := os.ReadFile(filepath.Join(root, "schema-cases", "system-config-v2", tc.name)) // #nosec G304 -- pinned conformance input
			if err != nil {
				t.Fatal(err)
			}
			var published map[string]any
			if err := json.Unmarshal(payload, &published); err != nil {
				t.Fatalf("decode published system case: %v", err)
			}
			publishedEnv, ok := published["environments"].(map[string]any)
			if !ok {
				t.Fatal("published system case has no environments object")
			}
			isolation, ok := publishedEnv["isolation"]
			if !ok {
				t.Fatal("published system case has no isolation knob")
			}
			system, err := json.Marshal(map[string]any{
				"schema_version": float64(2),
				"locked":         []any{"environments.isolation"},
				"environments":   map[string]any{"isolation": isolation},
			})
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			userPath := writeConfig(t, dir, "config.json", `{"schema_version": 2, "skills_root": "/tmp/skills", "projects": {}}`)
			systemPath := writeConfig(t, dir, "system.json", string(system))
			t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
			cfg, err := Load(userPath, nil)
			if err != nil {
				t.Fatalf("Load rejected projected published isolation case: %v", err)
			}
			wantJSON, err := json.Marshal(isolation)
			if err != nil {
				t.Fatal(err)
			}
			var want map[string]map[string]string
			if err := json.Unmarshal(wantJSON, &want); err != nil {
				t.Fatal(err)
			}
			if !cfg.Locked["environments.isolation"] || !reflect.DeepEqual(cfg.Env.Isolation, want) {
				t.Fatalf("Load isolation = %+v, locked=%v; want %s projection %+v", cfg.Env.Isolation, cfg.Locked, tc.mode, want)
			}
			for profile, environments := range want {
				for env, mode := range environments {
					if mode != tc.mode {
						t.Fatalf("published %s case has %s for %s/%s", tc.mode, mode, profile, env)
					}
				}
			}
		})
	}
}

// TestManagerConfigV2Vectors runs the published manager-config-v2 vector
// family through Parse: every valid case renders the exact expected
// effective configuration and every invalid case is rejected.
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
			wantJSON, err := json.Marshal(tc.Expected)
			if err != nil {
				t.Fatal(err)
			}
			if !managerEffectiveJSONMatchesVector(got, tc.Expected) {
				return conformancecoverage.Observation{FailureReason: fmt.Sprintf("effective config got %s, want %s", rendered, wantJSON)}
			}
			return conformancecoverage.Observation{}
		})
}

// exactManagerEffectiveJSON compares canonical JSON bytes for the full
// normalized object. In particular, extra actual keys are a mismatch too.
func exactManagerEffectiveJSON(actual, expected map[string]any) bool {
	actualJSON, actualErr := json.Marshal(actual)
	expectedJSON, expectedErr := json.Marshal(expected)
	return actualErr == nil && expectedErr == nil && bytes.Equal(actualJSON, expectedJSON)
}

func TestManagerEffectiveJSONComparisonRejectsExtraKnobs(t *testing.T) {
	actual := map[string]any{
		"existing":                  true,
		"transitive_system_modules": "drop",
	}
	expected := map[string]any{"existing": true}
	if exactManagerEffectiveJSON(actual, expected) {
		t.Fatal("normalized output with an unexpected manager knob matched the expected vector")
	}
}

// managerEffectiveJSONMatchesVector handles the published vectors that
// project the normalized object to their declared top-level keys. Each
// projected value still compares as canonical JSON, so nested knobs remain
// strict; full-object vectors retain exactManagerEffectiveJSON's strict rule.
func managerEffectiveJSONMatchesVector(actual, expected map[string]any) bool {
	if len(expected) == len(actual) {
		return exactManagerEffectiveJSON(actual, expected)
	}
	if len(expected) > len(actual) {
		return false
	}
	for key, expectedValue := range expected {
		actualValue, ok := actual[key]
		if !ok {
			return false
		}
		actualJSON, actualErr := json.Marshal(actualValue)
		expectedJSON, expectedErr := json.Marshal(expectedValue)
		if actualErr != nil || expectedErr != nil || !bytes.Equal(actualJSON, expectedJSON) {
			return false
		}
	}
	return true
}

func TestProjectedManagerEffectiveJSONRejectsExtraEnvironmentKnob(t *testing.T) {
	actual := map[string]any{
		"environments": map[string]any{
			"existing":                  true,
			"transitive_system_modules": "drop",
		},
		"adapter_mode": "auto",
	}
	expected := map[string]any{"environments": map[string]any{"existing": true}}
	if managerEffectiveJSONMatchesVector(actual, expected) {
		t.Fatal("projected vector matched despite an unexpected environments knob")
	}
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

// TestOverlayConformancePassesWithPermissionsAndSourceSigners drives the
// published overlay schema cases through Load and their vectors through Parse.
// With both the permissions and source signer knobs implemented, these cases
// must match their published behavior and have no remaining gap rows.
func TestOverlayConformancePassesWithPermissionsAndSourceSigners(t *testing.T) {
	root := conformanceRoot(t)
	_, gaps, err := conformancecoverage.Load()
	if err != nil {
		t.Fatal(err)
	}
	gapByCase := make(map[string]conformancecoverage.Gap, len(gaps))
	for _, gap := range gaps {
		gapByCase[gap.Family+"\x00"+gap.CaseID] = gap
	}
	assertNoGap := func(family, caseID string) {
		t.Helper()
		if gap, ok := gapByCase[family+"\x00"+caseID]; ok {
			t.Errorf("passing conformance case %s/%s still has gap row owned by %s: %s", family, caseID, gap.Owner, gap.Reason)
		}
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
				t.Errorf("%s does not carry the published %q field", tc.Name, field)
			}
		}
		if _, err := Load(writeConfig(t, t.TempDir(), "config.json", string(payload)), nil); err != nil {
			t.Errorf("Load rejects published valid overlay case %s: %v", tc.Name, err)
			continue
		}
		assertNoGap(managerSchema, tc.Name)
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
	selectedVectors := map[string]bool{
		"schema2-overlay-git-git-uppercase":            true,
		"schema2-overlay-git-http-uppercase":           true,
		"schema2-overlay-git-https-uppercase":          true,
		"schema2-overlay-git-scp":                      true,
		"schema2-overlay-git-scp-no-user":              true,
		"schema2-overlay-git-single-letter-host":       true,
		"schema2-overlay-git-ssh-uppercase":            true,
		"schema2-overlay-path-colon-later-segment":     true,
		"schema2-overlay-path-relative":                true,
		"schema2-overlay-path-source":                  true,
		"schema2-overlay-path-windows-backslash":       true,
		"schema2-overlay-path-windows-double-slash":    true,
		"schema2-overlay-path-windows-lowercase-drive": true,
		"schema2-overlay-path-windows-slash":           true,
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
			t.Errorf("Parse rejects published valid overlay vector %s: %v", tc.Name, parseErr)
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
				t.Errorf("%s expected EffectiveJSON omits the published %q field", tc.Name, field)
			}
		}
		if differences := jsonDifferencePaths("environments", gotEnv, wantEnv); len(differences) != 0 {
			t.Errorf("%s EffectiveJSON differences = %v, want none", tc.Name, differences)
		}
		assertNoGap(managerVectors, tc.Name)
	}
	if vectorCases != 14 {
		t.Fatalf("published valid overlay vectors = %d, want 14", vectorCases)
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
