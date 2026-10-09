package marker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// The closed v6 shape is pinned against the draft schema vendored under
// testdata/draft-sources-v2 (byte-identical to curator-spec
// schemas/draft-sources-v2/install-marker-v6.schema.json at 7eaeb73f).
// Draft marker v6 is marker v5 with schema_version 6 and
// skill_schema_version through 9; every other field, requiredness rule,
// and currentness comparison is unchanged (skillfile-sources §4).

func loadV6Schema(t *testing.T) map[string]any {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join("testdata", "draft-sources-v2", "install-marker-v6.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(payload, &schema); err != nil {
		t.Fatal(err)
	}
	return schema
}

// TestMarkerV6TopLevelMatchesDraftSchema pins the closed v6 top level
// against the vendored draft schema: the schema is closed, its property
// and required sets are exactly the v5 migration-table sets, the five
// replaced legacy fields are not properties at all, the version const is
// 6 with skill schemas through 9, and the local arm forbids attestation
// and substituted.
func TestMarkerV6TopLevelMatchesDraftSchema(t *testing.T) {
	schema := loadV6Schema(t)
	if closed, ok := schema["additionalProperties"].(bool); !ok || closed {
		t.Fatalf("install-marker-v6 must be closed (additionalProperties:false)")
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("install-marker-v6 properties missing")
	}
	if len(properties) != len(v5SchemaTopLevel) {
		t.Fatalf("schema properties = %d members, table = %d", len(properties), len(v5SchemaTopLevel))
	}
	for _, member := range v5SchemaTopLevel {
		if _, ok := properties[member]; !ok {
			t.Fatalf("schema property %q missing from the table", member)
		}
	}
	for member := range properties {
		found := false
		for _, want := range v5SchemaTopLevel {
			if member == want {
				found = true
			}
		}
		if !found {
			t.Fatalf("schema property %q missing from the table", member)
		}
	}
	for _, field := range []string{"source", "git", "ref_kind", "ref", "commit"} {
		if _, ok := properties[field]; ok {
			t.Fatalf("schema admits replaced legacy field %q", field)
		}
	}
	if version, ok := properties["schema_version"].(map[string]any); !ok || version["const"] != float64(SchemaV6) {
		t.Fatalf("schema_version const = %v, want %d", properties["schema_version"], SchemaV6)
	}
	skill, ok := properties["skill_schema_version"].(map[string]any)
	if !ok || skill["minimum"] != float64(1) || skill["maximum"] != float64(9) {
		t.Fatalf("skill_schema_version range = %v, want 1..9", properties["skill_schema_version"])
	}
	requiredAny, ok := schema["required"].([]any)
	if !ok {
		t.Fatalf("install-marker-v6 required missing")
	}
	if len(requiredAny) != len(v5SchemaRequired) {
		t.Fatalf("schema required = %v, table = %v", requiredAny, v5SchemaRequired)
	}
	for _, reqAny := range requiredAny {
		req, ok := reqAny.(string)
		if !ok {
			t.Fatalf("schema required entry not a string: %v", reqAny)
		}
		found := false
		for _, want := range v5SchemaRequired {
			if req == want {
				found = true
			}
		}
		if !found {
			t.Fatalf("schema required %q missing from the table", req)
		}
	}
}

// validateV6AgainstSchema checks one raw v6 document against the closed
// sets of the vendored draft schema, mirroring validateV5AgainstSchema:
// every top-level member is a schema property, every required member is
// present, the schema version is 6, the skill schema is in 1..9, and the
// package matches its arm's closed shape exactly.
func validateV6AgainstSchema(t *testing.T, raw map[string]json.RawMessage) error {
	t.Helper()
	schema := loadV6Schema(t)
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("install-marker-v6 properties missing")
	}
	for member := range raw {
		if _, ok := properties[member]; !ok {
			return fmt.Errorf("document carries member %q outside the schema properties", member)
		}
	}
	for _, req := range v5SchemaRequired {
		if _, ok := raw[req]; !ok {
			return fmt.Errorf("document misses required member %q", req)
		}
	}
	var version int
	if err := json.Unmarshal(raw["schema_version"], &version); err != nil || version != SchemaV6 {
		return fmt.Errorf("schema_version = %s, want %d", raw["schema_version"], SchemaV6)
	}
	var skillSchema int
	if err := json.Unmarshal(raw["skill_schema_version"], &skillSchema); err != nil ||
		skillSchema < 1 || skillSchema > 9 {
		return fmt.Errorf("skill_schema_version = %s, want 1..9", raw["skill_schema_version"])
	}
	var pkg struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(raw["package"], &pkg); err != nil {
		return fmt.Errorf("package is not an object: %v", err)
	}
	arm, ok := v5PackageArms[pkg.Kind]
	if !ok {
		return fmt.Errorf("package kind %q is not a schema arm", pkg.Kind)
	}
	var pkgRaw map[string]json.RawMessage
	if err := json.Unmarshal(raw["package"], &pkgRaw); err != nil {
		t.Fatal(err)
	}
	if !closedRawShape(pkgRaw, arm) {
		return fmt.Errorf("package does not match the closed %s arm: %s", pkg.Kind, raw["package"])
	}
	if commitRaw, present := pkgRaw["commit"]; present {
		var commitRawMap map[string]json.RawMessage
		if err := json.Unmarshal(commitRaw, &commitRawMap); err != nil ||
			!closedRawShape(commitRawMap, v5CommitShape) {
			return fmt.Errorf("commit is not the closed lockedCommit shape: %s", commitRaw)
		}
	}
	if pkg.Kind == "local-snapshot" {
		if _, present := raw["attestation"]; present {
			return fmt.Errorf("local-snapshot document carries attestation")
		}
		if _, present := raw["substituted"]; present {
			return fmt.Errorf("local-snapshot document carries substituted")
		}
	}
	return nil
}

// TestMarkerV6WrittenMarkersValidateAgainstDraftSchema validates a
// written v6 marker of every arm — network-git, configured-git and
// local-snapshot — against the vendored draft schema, so a writer that
// emits a non-conformant document fails here instead of shipping. The
// trailing controls prove the check is not vacuous.
func TestMarkerV6WrittenMarkersValidateAgainstDraftSchema(t *testing.T) {
	configured := v5GitAttestedConfiguredMarker()
	configured.SkillSchemaVersion = 9
	local := v5LocalMarker()
	local.SkillSchemaVersion = 9
	for name, base := range map[string]*Marker{
		"network-git":    v6NetworkMarker(),
		"configured-git": configured,
		"local-snapshot": local,
	} {
		t.Run(name, func(t *testing.T) {
			dir, _ := writeV5(t, base)
			payload, err := os.ReadFile(filepath.Join(dir, Name))
			if err != nil {
				t.Fatal(err)
			}
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(payload, &raw); err != nil {
				t.Fatal(err)
			}
			if err := validateV6AgainstSchema(t, raw); err != nil {
				t.Fatal(err)
			}
		})
	}
	t.Run("control/replaced-field-refused", func(t *testing.T) {
		dir, _ := writeV5(t, v6NetworkMarker())
		spliceV5Member(t, dir, `"source":"review"`)
		payload, err := os.ReadFile(filepath.Join(dir, Name))
		if err != nil {
			t.Fatal(err)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(payload, &raw); err != nil {
			t.Fatal(err)
		}
		if err := validateV6AgainstSchema(t, raw); err == nil {
			t.Fatalf("schema validation admitted a v6 Git marker carrying source")
		}
	})
	t.Run("control/hash-version-refused", func(t *testing.T) {
		dir, _ := writeV5(t, v6NetworkMarker())
		spliceV5Member(t, dir, `"hash_version":2`)
		if recorded := Read(dir); recorded != nil {
			t.Fatalf("Read admitted a v6 marker carrying hash_version: %+v", recorded)
		}
	})
}
