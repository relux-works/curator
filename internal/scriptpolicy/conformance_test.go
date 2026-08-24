package scriptpolicy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/skillspec"
)

// scriptHostExecutionPolicyVector is the published script-worker-v1 behavioural
// family. Only the sections curator's surface answers are decoded; the rest of
// the file is checked by name in TestScriptHostExecutionPolicySectionsAreAllClassified,
// so a section this build does not read can never arrive unnoticed.
type scriptHostExecutionPolicyVector struct {
	SchemaVersion   int      `json:"schema_version"`
	ProtocolVersion string   `json:"protocol_version"`
	ExecutionPolicy string   `json:"execution_policy"`
	Interpreters    []string `json:"interpreters"`
	OptInCases      []struct {
		Name            string  `json:"name"`
		ManifestSchema  int     `json:"manifest_schema"`
		ExecutionPolicy *string `json:"execution_policy"`
		Interpreter     *string `json:"interpreter"`
		Mode            *string `json:"mode"`
		Accepted        bool    `json:"accepted"`
	} `json:"opt_in_cases"`
}

// Section classifications. Every top-level section of the published family
// carries exactly one of these, and the classification is asserted against the
// file's real key set rather than trusted.
const (
	// consumedHere: this file drives the section against curator's behaviour.
	consumedHere = "consumed by this package"
	// refusedBeforeReached: the section describes worker behaviour that only a
	// manager implementing script-worker-v1 can exhibit. curator refuses every
	// enforced command at admission (manager profile §3.6), so no run of this
	// build can reach it. TestARefusalPrecedesEveryWorkerSurface asserts the
	// refusal that makes that true, instead of leaving the section unread.
	refusedBeforeReached = "unreachable: curator refuses enforced commands before any worker surface"
	// notImplementedYet: real curator surface that this build does not yet
	// carry. Named with its owner so the gap is declared, not silent.
	notImplementedYet = "not implemented: audit warning classes, owned by STORY-260822-2h0v9j"
)

var scriptHostExecutionPolicySections = map[string]string{
	"schema_version":              consumedHere,
	"protocol_version":            consumedHere,
	"execution_policy":            consumedHere,
	"interpreters":                consumedHere,
	"opt_in_cases":                consumedHere,
	"audit_label_cases":           notImplementedYet,
	"capability_derivation_cases": refusedBeforeReached,
	"capability_evidence_cases":   refusedBeforeReached,
	"capability_evidence_record":  refusedBeforeReached,
	"mandatory_controls":          refusedBeforeReached,
	"native_control_inventory":    refusedBeforeReached,
	"preflight_cases":             refusedBeforeReached,
}

// loadScriptHostExecutionPolicyVector reads the published family. The read is
// unguarded on purpose: `.github/ci/root-artifacts.tsv` declares this file for
// this package, so a root that stops publishing it defers the package and is
// fatal in any lane that requires a fully serving root. Skipping here instead
// would let a candidate be qualified against a suite this build never read.
func loadScriptHostExecutionPolicyVector(t *testing.T) (string, []byte) {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	path := filepath.Join(root, "vectors", "script-host-execution-policy.json")
	payload, err := os.ReadFile(path) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	return root, payload
}

// TestScriptHostExecutionPolicySectionsAreAllClassified proves the file this
// build reads is the file the suite publishes. A section the protocol adds --
// a new case family, a new record shape -- fails here on its first run rather
// than being silently ignored by a decoder that only names what it wants.
func TestScriptHostExecutionPolicySectionsAreAllClassified(t *testing.T) {
	_, payload := loadScriptHostExecutionPolicyVector(t)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	if len(raw) == 0 {
		t.Fatal("the script-host-execution-policy family published no sections")
	}
	var unclassified, stale []string
	for section := range raw {
		if _, ok := scriptHostExecutionPolicySections[section]; !ok {
			unclassified = append(unclassified, section)
		}
	}
	for section := range scriptHostExecutionPolicySections {
		if _, ok := raw[section]; !ok {
			stale = append(stale, section)
		}
	}
	sort.Strings(unclassified)
	sort.Strings(stale)
	if len(unclassified) != 0 {
		t.Errorf("the root publishes sections this build does not classify: %s\n"+
			"classify each in scriptHostExecutionPolicySections and consume it, or record why it is unreachable",
			strings.Join(unclassified, ", "))
	}
	if len(stale) != 0 {
		t.Errorf("this build classifies sections the root no longer publishes: %s\n"+
			"a classification that names nothing proves nothing; delete it or fix the name",
			strings.Join(stale, ", "))
	}
}

// TestScriptExecutionPolicyIdentityMatchesTheSuite binds the two closed
// identities this build hard-codes to the suite's own bytes. A protocol that
// renamed the policy or widened the interpreter set would otherwise leave
// curator silently accepting the old spelling and rejecting the new one.
func TestScriptExecutionPolicyIdentityMatchesTheSuite(t *testing.T) {
	_, payload := loadScriptHostExecutionPolicyVector(t)
	var vector scriptHostExecutionPolicyVector
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatal(err)
	}
	if vector.ExecutionPolicy != skillspec.ScriptExecutionPolicy {
		t.Fatalf("the suite names execution policy %q; this build hard-codes %q",
			vector.ExecutionPolicy, skillspec.ScriptExecutionPolicy)
	}
	if len(vector.Interpreters) == 0 {
		t.Fatal("the suite published no interpreter identities")
	}
	published := append([]string(nil), vector.Interpreters...)
	sort.Strings(published)
	accepted := make([]string, 0, len(skillspec.ScriptInterpreters))
	for name := range skillspec.ScriptInterpreters {
		accepted = append(accepted, name)
	}
	sort.Strings(accepted)
	if strings.Join(published, ",") != strings.Join(accepted, ",") {
		t.Fatalf("the suite publishes interpreters [%s]; this build accepts [%s]",
			strings.Join(published, ", "), strings.Join(accepted, ", "))
	}
}

// TestScriptExecutionOptInCases runs every published opt-in case through the
// two decisions curator owns: whether the manifest parses, and -- when it does
// -- whether the command is enforced or declared-only. An enforced command
// must additionally be refused with the closed §4.1.1 diagnostic, because this
// manager has no worker and the profile forbids installing it anyway.
func TestScriptExecutionOptInCases(t *testing.T) {
	_, payload := loadScriptHostExecutionPolicyVector(t)
	var vector scriptHostExecutionPolicyVector
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatal(err)
	}
	if len(vector.OptInCases) == 0 {
		t.Fatal("the script-host-execution-policy family published no opt-in cases")
	}
	for _, testCase := range vector.OptInCases {
		t.Run(testCase.Name, func(t *testing.T) {
			command := map[string]any{"type": "script", "unix_path": "scripts/tool"}
			if testCase.ExecutionPolicy != nil {
				command["execution_policy"] = *testCase.ExecutionPolicy
			}
			if testCase.Interpreter != nil {
				command["interpreter"] = *testCase.Interpreter
			}
			spec, err := skillspec.Load(materializeScriptSkill(t, testCase.ManifestSchema, command))
			if (err == nil) != testCase.Accepted {
				t.Fatalf("accepted = %v, want %v (error %v)", err == nil, testCase.Accepted, err)
			}
			if !testCase.Accepted {
				// A rejected manifest has no command to classify, and the case
				// records no mode for it.
				if testCase.Mode != nil {
					t.Fatalf("the suite records mode %q for a rejected case", *testCase.Mode)
				}
				return
			}
			if testCase.Mode == nil {
				t.Fatal("the suite accepts this case but records no mode for it")
			}
			parsed, ok := spec.Commands["tool"]
			if !ok {
				t.Fatalf("the parsed manifest carries no command: %+v", spec.Commands)
			}
			wantEnforced := *testCase.Mode == "enforced"
			if Enforced(parsed) != wantEnforced {
				t.Fatalf("Enforced = %v, want %v for mode %q (policy %q)",
					Enforced(parsed), wantEnforced, *testCase.Mode, parsed.ExecutionPolicy)
			}
			admission := Admit(spec.Commands)
			if !wantEnforced {
				if admission != nil {
					t.Fatalf("a declared-only command was refused: %v", admission)
				}
				return
			}
			if Code(admission) != PolicyUnsupported {
				t.Fatalf("Admit code = %q, want %q (error %v)", Code(admission), PolicyUnsupported, admission)
			}
			if parsed.Interpreter == "" {
				t.Fatal("an enforced command parsed without its bound interpreter")
			}
		})
	}
}

// TestARefusalPrecedesEveryWorkerSurface is the assertion the sections
// classified `refusedBeforeReached` rest on. Every enforced shape the suite
// itself declares -- the closed policy identity against each published
// interpreter -- is refused at admission, so no control probe, no capability
// derivation, no evidence record and no preflight of this family is reachable
// from this build. When the worker lands, this test is what has to change
// first.
func TestARefusalPrecedesEveryWorkerSurface(t *testing.T) {
	_, payload := loadScriptHostExecutionPolicyVector(t)
	var vector scriptHostExecutionPolicyVector
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatal(err)
	}
	if len(vector.Interpreters) == 0 {
		t.Fatal("the suite published no interpreter identities")
	}
	for _, interpreter := range vector.Interpreters {
		t.Run(interpreter, func(t *testing.T) {
			spec, err := skillspec.Load(materializeScriptSkill(t, 8, map[string]any{
				"type":             "script",
				"unix_path":        "scripts/tool",
				"execution_policy": vector.ExecutionPolicy,
				"interpreter":      interpreter,
			}))
			if err != nil {
				t.Fatalf("the suite's own enforced shape did not parse: %v", err)
			}
			if Code(Admit(spec.Commands)) != PolicyUnsupported {
				t.Fatalf("an enforced %s command was not refused with %s", interpreter, PolicyUnsupported)
			}
		})
	}
}

// capabilitiesRequiredFrom is the first manifest schema that requires a
// `capabilities` object. The fixtures below carry an empty one from there on so
// a case fails for its execution-policy rule and not for a missing field.
const capabilitiesRequiredFrom = 3

// materializeScriptSkill writes a one-command skill snapshot at the requested
// manifest schema and lays out the script the command names, so a case fails
// for its execution-policy rule rather than for a missing file.
func materializeScriptSkill(t *testing.T, schema int, command map[string]any) string {
	t.Helper()
	dir := t.TempDir()
	if path, ok := command["unix_path"].(string); ok {
		full := filepath.Join(dir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("fixture"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	manifest := map[string]any{
		"schema_version": schema,
		"commands":       map[string]any{"tool": command},
	}
	if schema >= capabilitiesRequiredFrom {
		manifest["capabilities"] = map[string]any{}
	}
	payload, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, skillspec.CanonicalManifestName), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}
