package envprofile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
)

// Each pinned case executes the real manager: forbidden trust/MCP inheritance
// is paired with preserved settings, A/B controls, invalid-TOML zero home
// publication/lock mutation, a valid legacy-profile provisioning control, and
// byte/record preservation for existing homes under C.
func TestCodexSeedCSpecCases(t *testing.T) {
	var corpus struct {
		Cases []struct {
			Name, Operation, Revision, Config string
			Record                            *envmarker.CodexSeedRecord
			Expected                          struct {
				SeedConfig *string `json:"seed_config"`
				Record     *envmarker.CodexSeedRecord
				Diagnostic *string
				Names      []string
				Writes     []string
			}
		}
	}
	raw, err := os.ReadFile("testdata/codex-seed-c-v1/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &corpus); err != nil {
		t.Fatal(err)
	}
	if len(corpus.Cases) != 22 {
		t.Fatalf("case count = %d", len(corpus.Cases))
	}
	for _, tc := range corpus.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			var fx *managedFixture
			if tc.Expected.SeedConfig == nil || tc.Name == "c-preserves-unrelated" {
				fx = legacyIdentityFixture(t)
			} else {
				fx = writeManagedFixture(t, "example")
				seedLiveNativeCredentials(t, fx)
			}
			nativePath := filepath.Join(fx.native[envregistry.CodexCLI], "config.toml")
			put := func(path, body string) {
				t.Helper()
				if err := os.WriteFile(path, []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
			}
			put(nativePath, tc.Config)
			req := fx.request(envregistry.CodexCLI)
			req.Repair = true
			home := ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI)
			configPath := filepath.Join(home, "config.toml")
			markerPath := filepath.Join(home, envmarker.Name)
			var beforeMarker []byte
			if tc.Operation != "provision" {
				// Provision a historical carrier when the case describes an old home.
				revision := "B"
				if tc.Record != nil {
					revision = tc.Record.Revision
				}
				if revision != "C" {
					req.codexSeedRevisionForTest = revision
				}
				if _, err := Resolve(req); err != nil {
					t.Fatal(err)
				}
				marker := readManagedMarker(t, fx, envregistry.CodexCLI)
				marker.CodexSeedRecord = tc.Record
				encoded, err := marker.Marshal()
				if err != nil {
					t.Fatal(err)
				}
				put(markerPath, string(encoded))
				put(configPath, tc.Config)
				beforeMarker = encoded
				req.codexSeedRevisionForTest = ""
				put(nativePath, "model = \"changed-native\"\n[projects.\"/synthetic/new\"]\ntrust_level=\"trusted\"\n")
			} else if tc.Revision != "C" {
				req.codexSeedRevisionForTest = tc.Revision
			}
			lockBefore, err := os.ReadFile(lockPath(fx.home, fx.profile))
			if err != nil {
				t.Fatal(err)
			}
			nativeBefore, err := os.ReadFile(nativePath)
			if err != nil {
				t.Fatal(err)
			}
			homeBefore := ""
			if tc.Operation != "provision" {
				homeBefore = nativeSnapshot(t, home)
			}
			var warnings []string
			if tc.Operation == "status" {
				status, err := StatusOf(statusRequest(fx))
				if err != nil {
					t.Fatal(err)
				}
				if status.CodexSeedRule.Revision != "C" {
					t.Fatalf("shipped revision = %s", status.CodexSeedRule.Revision)
				}
				state := findHome(status, fx.profile, envregistry.CodexCLI)
				if state == nil || !state.Current {
					t.Fatalf("home not current: %+v", state)
				}
				warnings = state.Warnings
			} else {
				result, err := Resolve(req)
				if tc.Expected.SeedConfig == nil {
					if err == nil || !strings.Contains(err.Error(), envregistry.DiagSeedUnreadable) {
						t.Fatalf("invalid TOML refusal = %v", err)
					}
					lockAfter, err := os.ReadFile(lockPath(fx.home, fx.profile))
					if err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(lockBefore, lockAfter) {
						t.Fatal("invalid TOML changed profile lock before refusal")
					}
					if _, err := os.Lstat(home); !os.IsNotExist(err) {
						t.Fatalf("invalid TOML published home: %v", err)
					}
					after, _ := os.ReadFile(nativePath)
					if !bytes.Equal(after, nativeBefore) {
						t.Fatal("native changed on refusal")
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				warnings = result.Warnings
				// Repair does not itself print the status-only legacy warning.
				if tc.Operation == "repair" {
					status, err := StatusOf(statusRequest(fx))
					if err != nil {
						t.Fatal(err)
					}
					warnings = findHome(status, fx.profile, envregistry.CodexCLI).Warnings
				}
			}
			after, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatal(err)
			}
			if tc.Operation != "provision" {
				if nativeSnapshot(t, home) != homeBefore {
					t.Fatal("read-only/no-op operation changed home")
				}
				if string(after) != tc.Config {
					t.Fatal("existing seed bytes changed")
				}
				markerAfter, _ := os.ReadFile(markerPath)
				if !bytes.Equal(markerAfter, beforeMarker) {
					t.Fatal("existing marker bytes changed")
				}
			} else {
				var got, want map[string]any
				if err := toml.Unmarshal(after, &got); err != nil {
					t.Fatal(err)
				}
				if err := toml.Unmarshal([]byte(*tc.Expected.SeedConfig), &want); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(seedTOMLValue(got), seedTOMLValue(want)) {
					t.Fatalf("seed decoded values differ: got %#v want %#v", got, want)
				}
				var original map[string]any
				_ = toml.Unmarshal([]byte(tc.Config), &original)
				_, p := original["projects"]
				_, m := original["mcp_servers"]
				if (tc.Revision == "A" || (!p && !m)) && string(after) != tc.Config {
					t.Fatal("unfiltered seed bytes changed")
				}
				marker := readManagedMarker(t, fx, envregistry.CodexCLI)
				if tc.Revision == "C" && (marker.Version != 7 || marker.HashVersion != 2) {
					t.Fatalf("C carrier = %d/%d", marker.Version, marker.HashVersion)
				}
			}
			marker := readManagedMarker(t, fx, envregistry.CodexCLI)
			if !reflect.DeepEqual(marker.CodexSeedRecord, tc.Expected.Record) {
				t.Fatalf("record = %+v want %+v", marker.CodexSeedRecord, tc.Expected.Record)
			}
			if tc.Expected.Diagnostic != nil && countWarnings(warnings, *tc.Expected.Diagnostic) != 1 {
				t.Fatalf("missing diagnostic %s: %v", *tc.Expected.Diagnostic, warnings)
			}
			if tc.Expected.Diagnostic == nil && (countWarnings(warnings, envregistry.DiagMCPSeedUnstripped) > 0 || countWarnings(warnings, envregistry.DiagMCPNativeServersNotInherited) > 0 || countWarnings(warnings, envregistry.DiagMCPNativeServersUngoverned) > 0) {
				t.Fatalf("unexpected MCP warning %v", warnings)
			}
			for _, warning := range warnings {
				for _, label := range []string{"/foreign-host/", "/work/synthetic-", "trust_level"} {
					if strings.Contains(warning, label) {
						t.Fatalf("trust information leaked in warning: %s", warning)
					}
				}
			}
			nativeAfter, _ := os.ReadFile(nativePath)
			if !bytes.Equal(nativeAfter, nativeBefore) {
				t.Fatal("native source changed")
			}
		})
	}
}

// Actual missing, existing and symlink fixture paths all lose top-level trust;
// nested projects and arrays remain. Repeating real provisioning preserves
// output bytes, including an unfiltered source's comments and formatting.
func TestCodexSeedCPathsAndDeterminism(t *testing.T) {
	root := t.TempDir()
	existing := filepath.Join(root, "checkout")
	if err := os.Mkdir(existing, 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "alias")
	if err := os.Symlink(existing, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	inputs := []string{"# exact formatting\nmodel=\"example\"\n[profiles.example.projects]\nlabels=[\"keep\",\"both\"]\n"}
	for _, path := range []string{"/foreign-host/synthetic-checkout", filepath.Join(root, "missing"), existing, link} {
		inputs = append(inputs, fmt.Sprintf("model=\"example\"\n[projects.%q]\ntrust_level=\"trusted\"\n[profiles.example.projects]\nlabels=[\"keep\",\"both\"]\n", filepath.ToSlash(path)))
	}
	for i, input := range inputs {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			var prior []byte
			for repeat := 0; repeat < 2; repeat++ {
				fx := writeManagedFixture(t, "example")
				seedLiveNativeCredentials(t, fx)
				path := filepath.Join(fx.native[envregistry.CodexCLI], "config.toml")
				if err := os.WriteFile(path, []byte(input), 0600); err != nil {
					t.Fatal(err)
				}
				req := fx.request(envregistry.CodexCLI)
				req.Repair = true
				if _, err := Resolve(req); err != nil {
					t.Fatal(err)
				}
				got, err := os.ReadFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI), "config.toml"))
				if err != nil {
					t.Fatal(err)
				}
				var document map[string]any
				if err := toml.Unmarshal(got, &document); err != nil {
					t.Fatal(err)
				}
				if _, ok := document["projects"]; ok {
					t.Fatal("C inherited project trust")
				}
				labels := document["profiles"].(map[string]any)["example"].(map[string]any)["projects"].(map[string]any)["labels"]
				if !reflect.DeepEqual(labels, []any{"keep", "both"}) {
					t.Fatalf("nested array changed: %#v", labels)
				}
				if i == 0 && !bytes.Equal(got, []byte(input)) {
					t.Fatal("unfiltered bytes changed")
				}
				if repeat > 0 && !bytes.Equal(prior, got) {
					t.Fatal("non-deterministic provisioning")
				}
				prior = got
			}
		})
	}
}

// TOML inline arrays of tables and table arrays decode into different Go slice
// types; normalize containers without changing any scalar value or array order.
func seedTOMLValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		result := map[string]any{}
		for k, x := range v {
			result[k] = seedTOMLValue(x)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, x := range v {
			result[i] = seedTOMLValue(x)
		}
		return result
	case []map[string]any:
		result := make([]any, len(v))
		for i, x := range v {
			result[i] = seedTOMLValue(x)
		}
		return result
	default:
		return value
	}
}

// A real legacy B home retains seed bytes and its B carrier through surface
// repair, resolve and status. Independently repaired surfaces do not authorize
// stripping trust, refreshing native input, or upgrading the seed identity.
func TestCodexSeedCLegacyBRepair(t *testing.T) {
	fx, _ := nativeSkillsFixture(t)
	native := filepath.Join(fx.native[envregistry.CodexCLI], "config.toml")
	original := "model=\"legacy\"\n[projects.\"/synthetic/legacy\"]\ntrust_level=\"trusted\"\n[mcp_servers.original]\ncommand=\"synthetic-server\"\n"
	if err := os.WriteFile(native, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	req := fx.request(envregistry.CodexCLI)
	req.Repair = true
	req.codexSeedRevisionForTest = "B"
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	home := ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI)
	path := filepath.Join(home, "config.toml")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	record := readManagedMarker(t, fx, envregistry.CodexCLI).CodexSeedRecord
	if err := os.WriteFile(native, []byte("[projects.\"/synthetic/changed\"]\ntrust_level=\"untrusted\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(home, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	req.codexSeedRevisionForTest = ""
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	req.Repair = false
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	if _, err := StatusOf(statusRequest(fx)); err != nil {
		t.Fatal(err)
	}
	if _, err := UseWithPolicy(fx.home, fx.profile, envregistry.CodexCLI, "", false, Policy{}); err != nil {
		t.Fatal(err)
	}
	if _, err := SyncWithPolicy(fx.home, Policy{}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := UpdateWithPolicy(fx.home, fx.profile, Policy{}); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	marker := readManagedMarker(t, fx, envregistry.CodexCLI)
	if !bytes.Equal(before, after) || !reflect.DeepEqual(record, marker.CodexSeedRecord) {
		t.Fatal("legacy B seed or record changed")
	}
	if marker.Version != envmarker.VersionV3 {
		t.Fatalf("C adoption upgraded legacy B carrier: %d", marker.Version)
	}
}

// Full node-carrier parsing and per-home activation are positive controls.
// Node activation with either repair value refuses with the inherited source
// diagnostic, emits no fragment, and preserves home, lock, and native bytes.
// Migration planning also reports that authority refusal without writes.
func TestCodexSeedCNodeAuthorityIsNotPerHome(t *testing.T) {
	fx := writeManagedFixture(t, "example")
	seedLiveNativeCredentials(t, fx)
	req := fx.request(envregistry.CodexCLI)
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	for _, repair := range []bool{false, true} {
		req.Repair = repair
		result, err := Resolve(req)
		if err != nil || result == nil || len(result.Document) == 0 {
			t.Fatalf("per-home activation repair=%v: result=%+v err=%v", repair, result, err)
		}
	}
	marker := readManagedMarker(t, fx, envregistry.CodexCLI)
	emptyPass := []envmarker.Passthrough{}
	emptySeeds := []string{}
	marker.Passthrough = &emptyPass
	marker.Seeds = &emptySeeds
	marker.SeedLinks = nil
	marker.CredentialSource = &envmarker.CredentialSource{Source: &envmarker.NodeSource{Mode: "node", SourceRef: "local:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", HarnessChannel: "CLAUDE_CODE_OAUTH_TOKEN"}, SourceGeneration: 1, PolicyGeneration: 1}
	raw, err := marker.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := envmarker.Parse(raw); err != nil {
		t.Fatal(err)
	}
	home := ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI)
	path := filepath.Join(home, envmarker.Name)
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	before := nativeSnapshot(t, fx.home)
	nativeBefore := nativeSnapshot(t, fx.native[envregistry.CodexCLI])
	for _, repair := range []bool{false, true} {
		t.Run(fmt.Sprintf("repair=%v", repair), func(t *testing.T) {
			req.Repair = repair
			result, err := Resolve(req)
			// Pin the normative diagnostic independently of the implementation constant.
			if err == nil || !strings.HasPrefix(err.Error(), "credential_injection_unavailable:") {
				t.Fatalf("node activation did not refuse authority: %v", err)
			}
			if result != nil && len(result.Document) != 0 {
				t.Fatalf("node activation emitted a fragment: %+v", result)
			}
			if nativeSnapshot(t, fx.home) != before || nativeSnapshot(t, fx.native[envregistry.CodexCLI]) != nativeBefore {
				t.Fatal("node authority refusal changed home, lock, or native bytes")
			}
		})
	}
	migration, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	conflict := false
	for _, entry := range migration.Homes {
		for _, c := range entry.Conflicts {
			if strings.HasPrefix(c.Detail, "credential_injection_unavailable:") {
				conflict = true
			}
		}
	}
	if !conflict {
		t.Fatal("node migration did not refuse authority")
	}
	if nativeSnapshot(t, fx.home) != before || nativeSnapshot(t, fx.native[envregistry.CodexCLI]) != nativeBefore {
		t.Fatal("node migration planning changed home, lock, or native bytes")
	}
}

// C cannot label hash-v1 profile state as marker 7. The frozen B writer is a
// nearby positive control for that same legacy profile and absent native seed.
func TestCodexSeedCRequiresHash2BeforePublication(t *testing.T) {
	enableV1WritersForTest(t)
	fx := writeManagedFixture(t, "example")
	seedLiveNativeCredentials(t, fx)
	req := fx.request(envregistry.CodexCLI)
	req.Repair = true
	req.codexSeedRevisionForTest = ""
	if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), "Codex seed C requires hash_version 2") {
		t.Fatalf("C hash-v1 writer did not refuse before publication: %v", err)
	}
	if _, err := os.Lstat(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI)); !os.IsNotExist(err) {
		t.Fatalf("C hash-v1 writer published home: %v", err)
	}
	req.codexSeedRevisionForTest = "B"
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	marker := readManagedMarker(t, fx, envregistry.CodexCLI)
	if marker.Version != envmarker.VersionV2 || marker.CodexSeedRecord.Revision != "B" {
		t.Fatalf("legacy positive carrier = %+v", marker)
	}
}

// Repairing a C surface also preserves later user-authored project trust in
// its seed. A provisioning exclusion never becomes permission to refresh it.
func TestCodexSeedCRepairPreservesUserSeed(t *testing.T) {
	fx := writeManagedFixture(t, "example")
	seedLiveNativeCredentials(t, fx)
	req := fx.request(envregistry.CodexCLI)
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	home := ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI)
	path := filepath.Join(home, "config.toml")
	user := []byte("# user-authored after provisioning\n[projects.\"/synthetic/user-checkout\"]\ntrust_level=\"trusted\"\n")
	if err := os.WriteFile(path, user, 0600); err != nil {
		t.Fatal(err)
	}
	before := readManagedMarker(t, fx, envregistry.CodexCLI).CodexSeedRecord
	if err := os.Remove(filepath.Join(home, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(user, after) || !reflect.DeepEqual(before, readManagedMarker(t, fx, envregistry.CodexCLI).CodexSeedRecord) {
		t.Fatal("C repair refreshed user-authored seed or record")
	}
}
