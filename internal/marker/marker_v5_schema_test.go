package marker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// The closed v5 shape is pinned against the accepted schemas vendored
// under testdata/draft-sources-v1 (byte-identical to curator-spec
// schemas/draft-sources-v1 at 802caee, landed in a4fcaf0). The drift
// rows below fail when the spec moves; the written-bytes rows fail when
// this implementation emits a document the schema refuses — the exact
// class revision 2 shipped on the Git arm (the five replaced legacy
// fields admitted alongside the package).

// v5SchemaTopLevel is the closed top-level member set of
// install-marker-v5.schema.json: every property the schema admits.
var v5SchemaTopLevel = []string{
	"name", "content_sha256", "build_source", "locale", "agents", "commands",
	"dependencies", "runtime_roots", "build_roots", "installed_at", "files",
	"requirements", "mcp_servers", "activation", "requirers", "schema_version",
	"skill_schema_version", "package", "lock_sha256", "builds",
	"attestation", "substituted",
}

// v5SchemaRequired is the required member set of install-marker-v5.schema.json.
var v5SchemaRequired = []string{
	"name", "content_sha256", "locale", "agents", "commands", "dependencies",
	"runtime_roots", "build_roots", "installed_at", "files", "schema_version",
	"skill_schema_version", "package", "lock_sha256", "builds",
}

func loadV5Schema(t *testing.T) map[string]any {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join("testdata", "draft-sources-v1", "install-marker-v5.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(payload, &schema); err != nil {
		t.Fatal(err)
	}
	return schema
}

func loadSourceTypesSchema(t *testing.T) map[string]any {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join("testdata", "draft-sources-v1", "source-types-v1.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(payload, &schema); err != nil {
		t.Fatal(err)
	}
	return schema
}

// TestMarkerV5TopLevelMatchesAcceptedSchema pins the closed v5 top level
// against the vendored accepted schema: the schema is closed, its
// property and required sets are exactly the migration table's, the five
// replaced legacy fields are not properties at all, and the local arm
// forbids attestation and substituted.
func TestMarkerV5TopLevelMatchesAcceptedSchema(t *testing.T) {
	schema := loadV5Schema(t)
	if closed, ok := schema["additionalProperties"].(bool); !ok || closed {
		t.Fatalf("install-marker-v5 must be closed (additionalProperties:false)")
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("install-marker-v5 properties missing")
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
	requiredAny, ok := schema["required"].([]any)
	if !ok {
		t.Fatalf("install-marker-v5 required missing")
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
	// The local-snapshot branch forbids attestation and substituted.
	allOf, ok := schema["allOf"].([]any)
	if !ok || len(allOf) == 0 {
		t.Fatalf("install-marker-v5 allOf missing")
	}
	foundLocalForbid := false
	for _, branchAny := range allOf {
		branch, ok := branchAny.(map[string]any)
		if !ok {
			continue
		}
		encoded, _ := json.Marshal(branch)
		var probe struct {
			If struct {
				Properties struct {
					Package struct {
						Properties struct {
							Kind struct {
								Const string `json:"const"`
							} `json:"kind"`
						} `json:"properties"`
					} `json:"package"`
				} `json:"properties"`
			} `json:"if"`
			Then struct {
				Not struct {
					AnyOf []struct {
						Required []string `json:"required"`
					} `json:"anyOf"`
				} `json:"not"`
			} `json:"then"`
		}
		if err := json.Unmarshal(encoded, &probe); err != nil {
			continue
		}
		if probe.If.Properties.Package.Properties.Kind.Const != "local-snapshot" {
			continue
		}
		forbids := map[string]bool{}
		for _, option := range probe.Then.Not.AnyOf {
			for _, req := range option.Required {
				forbids[req] = true
			}
		}
		if forbids["attestation"] && forbids["substituted"] {
			foundLocalForbid = true
		}
	}
	if !foundLocalForbid {
		t.Fatalf("schema must forbid attestation and substituted for local-snapshot")
	}
}

// TestMarkerV5PackageArmsMatchAcceptedSchema pins the in-code closed
// package-arm table (v5PackageArms) against the vendored source-types-v1
// schema: three arms, each closed, with required and property sets equal
// to the table.
func TestMarkerV5PackageArmsMatchAcceptedSchema(t *testing.T) {
	schema := loadSourceTypesSchema(t)
	defs, ok := schema["$defs"].(map[string]any)
	if !ok {
		t.Fatalf("$defs missing")
	}
	pkgDef, ok := defs["package"].(map[string]any)
	if !ok {
		t.Fatalf("$defs.package missing")
	}
	arms, ok := pkgDef["oneOf"].([]any)
	if !ok || len(arms) != 3 {
		t.Fatalf("package oneOf must have exactly 3 arms, got %v", arms)
	}
	seenKinds := map[string]bool{}
	for _, armAny := range arms {
		arm, ok := armAny.(map[string]any)
		if !ok {
			t.Fatalf("arm is not an object: %v", armAny)
		}
		if closed, ok := arm["additionalProperties"].(bool); !ok || closed {
			t.Fatalf("arm must declare additionalProperties:false: %v", arm)
		}
		props, ok := arm["properties"].(map[string]any)
		if !ok {
			t.Fatalf("arm properties missing: %v", arm)
		}
		kindProp, ok := props["kind"].(map[string]any)
		if !ok {
			t.Fatalf("arm kind property missing: %v", arm)
		}
		kind, ok := kindProp["const"].(string)
		if !ok {
			t.Fatalf("arm kind const missing: %v", arm)
		}
		seenKinds[kind] = true
		table, ok := v5PackageArms[kind]
		if !ok {
			t.Fatalf("schema arm kind %q missing from the table", kind)
		}
		requiredAny, ok := arm["required"].([]any)
		if !ok {
			t.Fatalf("arm required missing: %v", arm)
		}
		if len(requiredAny) != len(table) {
			t.Fatalf("arm %q required = %v, table = %v", kind, requiredAny, table)
		}
		for _, reqAny := range requiredAny {
			req, ok := reqAny.(string)
			if !ok {
				t.Fatalf("arm %q required entry not a string: %v", kind, reqAny)
			}
			if _, ok := table[req]; !ok {
				t.Fatalf("arm %q required %q missing from the table", kind, req)
			}
		}
		for key := range props {
			if _, ok := table[key]; !ok {
				t.Fatalf("arm %q schema property %q missing from the table", kind, key)
			}
		}
		for key := range table {
			if _, ok := props[key]; !ok {
				t.Fatalf("arm %q table member %q missing from schema properties", kind, key)
			}
		}
	}
	for _, kind := range []string{"local-snapshot", "network-git", "configured-git"} {
		if !seenKinds[kind] {
			t.Fatalf("schema arms missing kind %q", kind)
		}
	}
}

// validateV5AgainstSchema checks one raw v5 document against the closed
// sets of the vendored accepted schemas: every top-level member is a
// schema property, every required member is present, the schema version
// is 5, the skill schema is in range, the package matches its arm's
// closed shape exactly, and a local snapshot carries no attestation or
// substitution. Value grammars behind $refs (v4/common) stay the job of
// validMarker's unit rows; this check pins the closed shape, the class
// that regressed.
func validateV5AgainstSchema(t *testing.T, raw map[string]json.RawMessage) error {
	t.Helper()
	schema := loadV5Schema(t)
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("install-marker-v5 properties missing")
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
	if err := json.Unmarshal(raw["schema_version"], &version); err != nil || version != SchemaV5 {
		return fmt.Errorf("schema_version = %s, want %d", raw["schema_version"], SchemaV5)
	}
	var skillSchema int
	if err := json.Unmarshal(raw["skill_schema_version"], &skillSchema); err != nil ||
		skillSchema < 1 || skillSchema > 8 {
		return fmt.Errorf("skill_schema_version = %s, want 1..8", raw["skill_schema_version"])
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

// TestMarkerV5WrittenMarkersValidateAgainstAcceptedSchema validates a
// written v5 marker of every arm — network-git, configured-git and
// local-snapshot — against the vendored accepted schema, so a writer
// that emits a non-conformant document fails here instead of shipping.
// The trailing control proves the check is not vacuous: the same Git
// document with one replaced legacy field spliced back in is refused.
func TestMarkerV5WrittenMarkersValidateAgainstAcceptedSchema(t *testing.T) {
	for name, base := range map[string]*Marker{
		"network-git":    v5GitAttestedMarker(),
		"configured-git": v5GitAttestedConfiguredMarker(),
		"local-snapshot": v5LocalMarker(),
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
			if err := validateV5AgainstSchema(t, raw); err != nil {
				t.Fatal(err)
			}
		})
	}
	t.Run("control/replaced-field-refused", func(t *testing.T) {
		dir, _ := writeV5(t, v5GitAttestedMarker())
		spliceV5Member(t, dir, `"source":"review"`)
		payload, err := os.ReadFile(filepath.Join(dir, Name))
		if err != nil {
			t.Fatal(err)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(payload, &raw); err != nil {
			t.Fatal(err)
		}
		if err := validateV5AgainstSchema(t, raw); err == nil {
			t.Fatalf("schema validation admitted a v5 Git marker carrying source")
		}
	})
}
