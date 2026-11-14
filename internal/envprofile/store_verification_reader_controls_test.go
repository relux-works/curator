package envprofile

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The actual array schema accepts a valid row, while malformed carriers and
// broken references refuse with the qualification code and no returned rows.
// These are reader tests, not filesystem or launch qualification.
func TestStoreVerificationReaderClosedCarrier(t *testing.T) {
	valid, err := os.ReadFile("testdata/store-verification-v2/valid-case.json")
	if err != nil {
		t.Fatal(err)
	}
	if rows, err := decodeStoreVerificationCases(valid); err != nil || len(rows) != 1 {
		t.Fatalf("valid neighbor: %d %v", len(rows), err)
	}
	replace := func(from, to string) []byte {
		if !bytes.Contains(valid, []byte(from)) {
			t.Fatalf("missing control plant %q", from)
		}
		return bytes.Replace(valid, []byte(from), []byte(to), 1)
	}
	tests := map[string][]byte{
		"inadmissible-mcp-local":  bytes.ReplaceAll(replace(`"member_kind": "context"`, `"member_kind": "mcp"`), []byte(`"source_kind": "git"`), []byte(`"source_kind": "local"`)),
		"inadmissible-skill-path": bytes.ReplaceAll(replace(`"member_kind": "context"`, `"member_kind": "skill"`), []byte(`"source_kind": "git"`), []byte(`"source_kind": "path"`)),
		"old-object":              []byte(`{"resolve_cases":[]}`),
		"empty":                   []byte(`[]`), "null": []byte(`null`),
		"unknown-version":    replace(`"curator.environment-store-case/2"`, `"curator.environment-store-case/3"`),
		"unknown-capability": replace(`"agent-environments"`, `"unknown"`),
		"revision-null":      replace(`"capability_revision": 2`, `"capability_revision": null`),
		"revision-missing":   replace(`"capability_revision": 2,`, ``),
		"revision-fraction":  replace(`"capability_revision": 2`, `"capability_revision": 2.0`),
		"revision-exponent":  replace(`"capability_revision": 2`, `"capability_revision": 2e0`),
		"level-fraction":     replace("      1,", "      1.0,"),
		"level-exponent":     replace("      1,", "      1e0,"),
		"level-duplicate":    replace("      2,", "      1,"),
		"unknown-field":      replace(`"schema":`, `"unknown": true, "schema":`),
		"duplicate-key":      replace(`"schema":`, `"schema":"old", "schema":`),
		"escaped-duplicate":  replace(`"schema":`, `"sch\u0065ma":"old", "schema":`),
		"nested-unknown":     replace(`"fragment_emitted": true`, `"unknown":true, "fragment_emitted":true`),
		"nested-null":        replace(`"fragment_emitted": true`, `"fragment_emitted":null`),
		"nested-missing":     replace(`"fragment_emitted": true,`, ``),
		"writes-null":        replace(`"writes": []`, `"writes":null`),
		"missing-reference":  replace(`"positive_control": "control-resolve-context-intact"`, `"positive_control":"absent"`),
		"trailing":           append(append([]byte{}, valid...), []byte(`{}`)...),
		"invalid-utf8":       replace(`"agent-environments"`, string([]byte{'"', 0xff, '"'})),
		"excess-depth":       []byte(strings.Repeat("[", 34) + "0" + strings.Repeat("]", 34)),
		"excess-bytes":       bytes.Repeat([]byte(" "), (4<<20)+1),
	}
	// The tightened member/source constraint retains path-backed contexts and
	// Git-backed skills/MCPs beside the two inadmissible cross-products above.
	for _, raw := range [][]byte{replace(`"source_kind": "git"`, `"source_kind": "path"`), replace(`"member_kind": "context"`, `"member_kind": "mcp"`), replace(`"member_kind": "context"`, `"member_kind": "skill"`)} {
		if rows, err := decodeStoreVerificationCases(raw); err != nil || len(rows) != 1 {
			t.Fatalf("admitted source neighbor: rows=%d err=%v", len(rows), err)
		}
	}
	var duplicate []json.RawMessage
	if err := json.Unmarshal(valid, &duplicate); err != nil {
		t.Fatal(err)
	}
	duplicate = append(duplicate, duplicate[0])
	tests["duplicate-row"], err = json.Marshal(duplicate)
	if err != nil {
		t.Fatal(err)
	}
	mixed := bytes.Replace(duplicate[1], []byte(`"curator.environment-store-case/2"`), []byte(`"curator.environment-store-case/3"`), 1)
	tests["mixed-versions"] = append(append(append([]byte("["), duplicate[0]...), ','), append(mixed, ']')...)
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			rows, err := decodeStoreVerificationCases(raw)
			if err == nil || !strings.HasPrefix(err.Error(), storeVerificationUnsupported+":") || rows != nil {
				t.Fatalf("reader accepted plant: rows=%d err=%v", len(rows), err)
			}
		})
	}
}

func storeVerificationLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) { value, ok := values[key]; return value, ok }
}

// Absent selection is inert; partial, empty and unmatched selectors refuse
// before any corpus or fixture can be returned. Full pinned reading is below.
func TestStoreVerificationSelectionRefusesIncompletePin(t *testing.T) {
	rows, selected, err := readStoreVerificationCorpus(storeVerificationLookup(nil))
	if rows != nil || selected || err != nil {
		t.Fatalf("unset positive: %v %v %v", rows, selected, err)
	}
	for name, values := range map[string]map[string]string{
		"empty-root":  {storeVerificationRootKey: ""},
		"root-only":   {storeVerificationRootKey: t.TempDir()},
		"pin-only":    {storeVerificationPinKey: storeVerificationManifestPin},
		"relative":    {storeVerificationRootKey: "relative", storeVerificationPinKey: storeVerificationManifestPin},
		"unknown-pin": {storeVerificationRootKey: t.TempDir(), storeVerificationPinKey: strings.Repeat("0", 64)},
	} {
		t.Run(name, func(t *testing.T) {
			rows, selected, err := readStoreVerificationCorpus(storeVerificationLookup(values))
			if rows != nil || !selected || err == nil || !strings.HasPrefix(err.Error(), storeVerificationUnsupported+":") {
				t.Fatalf("partial input accepted: %v %v %v", rows, selected, err)
			}
		})
	}
}

// Both the real canonical gate and its Go boundary refuse a present-empty
// selector. With it unset the real script reaches its ordinary usage check,
// before Go execution or fixture setup. This does not claim a full suite run.
func TestStoreVerificationCanonicalSelectorGuard(t *testing.T) {
	for name, values := range map[string]map[string]string{
		"unset": nil, "empty": {storeVerificationRootKey: ""}, "set": {storeVerificationRootKey: "/fixture"}, "pin-only": {storeVerificationPinKey: storeVerificationManifestPin},
	} {
		t.Run(name, func(t *testing.T) {
			err := storeVerificationCanonicalGuard(storeVerificationLookup(values))
			if (err != nil) != (name != "unset") {
				t.Fatalf("Go canonical guard: %v", err)
			}
			cmd := exec.Command("bash", filepath.Join("..", "..", ".github", "ci", "test-gate.sh"))
			for _, entry := range os.Environ() {
				if !strings.HasPrefix(entry, storeVerificationRootKey+"=") && !strings.HasPrefix(entry, storeVerificationPinKey+"=") {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			for key, value := range values {
				cmd.Env = append(cmd.Env, key+"="+value)
			}
			out, runErr := cmd.CombinedOutput()
			want := "canonical conformance refuses store qualification"
			if name == "unset" {
				want = "usage: test-gate.sh"
			}
			if runErr == nil || !strings.Contains(string(out), want) {
				t.Fatalf("canonical gate did not reach intended boundary: err=%v out=%s", runErr, out)
			}
		})
	}
}

// Scan real production sources and prove the scanner detects a new getenv
// read; an unrelated source string remains a valid neighbor. The planted file
// exists only in a temporary fixture and is never built into the product.
func TestStoreVerificationProductionNonreference(t *testing.T) {
	if paths, err := storeVerificationProductionReferences(filepath.Join("..", "..")); err != nil || len(paths) != 0 {
		t.Fatalf("production reference: %v %v", paths, err)
	}
	root := t.TempDir()
	file := filepath.Join(root, "main.go")
	write := func(text string) {
		if err := os.WriteFile(file, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("package main\nvar value=\"ordinary input\"\n")
	if paths, err := storeVerificationProductionReferences(root); err != nil || len(paths) != 0 {
		t.Fatalf("positive scanner neighbor: %v %v", paths, err)
	}
	write("package main\nimport \"os\"\nvar value=os.Getenv(\"" + storeVerificationRootKey + "\")\n")
	if paths, err := storeVerificationProductionReferences(root); err != nil || len(paths) != 1 || paths[0] != file {
		t.Fatalf("known-bad production read escaped: %v %v", paths, err)
	}
}

// This opt-in read proves the entire exact draft roster qualifies before any
// operation driver runs. It is not runtime or canonical conformance evidence.
func TestStoreVerificationPinnedCorpusReader(t *testing.T) {
	rows, selected, err := readStoreVerificationCorpus(os.LookupEnv)
	if err != nil {
		t.Fatal(err)
	}
	if !selected {
		t.Skip("separate store revision-2 draft qualification is not selected")
	}
	implementation, controls, executions, controlExecutions := 0, 0, 0, 0
	for _, row := range rows {
		if row.Kind == "implementation" {
			implementation++
			executions += len(row.Levels)
		} else {
			controls++
			controlExecutions += len(row.Levels)
		}
	}
	if implementation != storeVerificationImplementationCount || controls != storeVerificationControlCount || executions != storeVerificationImplementationLevels || controlExecutions != storeVerificationControlLevels {
		t.Fatalf("incomplete roster: %d/%d rows, %d/%d levels", implementation, controls, executions, controlExecutions)
	}
	t.Log("non-canonical draft reader: 179 implementation/6 checker-control rows, 519/22 declared levels; no runtime executions claimed")
}

// Actual frozen v1 bytes remain readable by their existing object type. That
// reader refuses the revision-2 array, and the new reader refuses the v1 object.
// This is format refusal in both directions, not runtime equivalence.
func TestStoreVerificationActualReaderDirections(t *testing.T) {
	if _, selected := os.LookupEnv(storeVerificationRootKey); !selected {
		t.Skip("separate store revision-2 draft qualification is not selected")
	}
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Fatal("selected draft qualification requires the frozen v1 root for both reader directions")
	}
	old, err := storeVerificationReadPinned(filepath.Join(root, "vectors", storeBoundaryVectorFile), storeVerificationPredecessorPin)
	if err != nil {
		t.Fatal(err)
	}
	var legacy storeBoundaryVectors
	if err := json.Unmarshal(old, &legacy); err != nil || len(legacy.Resolve)+len(legacy.DryRun)+len(legacy.Repair)+len(legacy.Status) != 38 {
		t.Fatalf("actual v1 positive refused: %v", err)
	}
	if rows, err := decodeStoreVerificationCases(old); err == nil || rows != nil || !strings.HasPrefix(err.Error(), storeVerificationUnsupported+":") {
		t.Fatalf("v2 reader admitted actual v1 object: %v", err)
	}
	rows, selected, err := readStoreVerificationCorpus(os.LookupEnv)
	if err != nil {
		t.Fatal(err)
	}
	if !selected {
		t.Fatal("draft selection disappeared")
	}
	raw, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	legacy = storeBoundaryVectors{}
	if err := json.Unmarshal(raw, &legacy); err == nil {
		t.Fatal("actual old object reader admitted revision-2 array")
	}
}

// The canonical stage selects only the sibling corpus from SPEC_PIN's checkout,
// never an environment-selected alternative. It checks carriers and declared
// coverage; this reader does not execute the revision-2 filesystem operations.
func TestStoreVerificationCanonicalPinnedCorpusReader(t *testing.T) {
	if os.Getenv("CI_STORE_V2_CANONICAL") != "1" {
		t.Skip("canonical store reader stage is not selected")
	}
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Fatal("canonical store reader requires the SPEC_PIN root")
	}
	root, err := filepath.Abs(filepath.Join(root, "..", "environment-store-verification-v2"))
	if err != nil {
		t.Fatal(err)
	}
	rows, selected, err := readStoreVerificationCorpus(storeVerificationLookup(map[string]string{storeVerificationRootKey: root, storeVerificationPinKey: storeVerificationManifestPin}))
	if err != nil || !selected || len(rows) != storeVerificationCaseCount {
		t.Fatalf("canonical pin requires the landed revision-2 corpus: rows=%d err=%v", len(rows), err)
	}
	t.Logf("corpus pin: spec=%s manifest=%s rows=%d", storeVerificationSpecCommit, storeVerificationManifestPin, len(rows))
}
