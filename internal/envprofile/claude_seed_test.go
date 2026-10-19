package envprofile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/envregistry"
)

func claudeRepairFixture(t *testing.T, form string) (*managedFixture, ResolveRequest, string) {
	t.Helper()
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	req := fx.request(envregistry.ClaudeCode)
	req.Machine.Forms = map[string]string{envregistry.ClaudeCode: form}
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	return fx, req, filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.ClaudeCode), ".claude.json")
}

func claudeNumberObject(t *testing.T, payload []byte) map[string]any {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var object map[string]any
	if err := decoder.Decode(&object); err != nil {
		t.Fatal(err)
	}
	return object
}

func readClaudeState(t *testing.T, path string) []byte {
	t.Helper()
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

// Resolve repairs real-world numeric tool state, preserving every unowned value and number lexeme, then verifies it without repair.
func TestClaudeRepairPreservesNumberText(t *testing.T) {
	for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
		for _, variant := range []string{"new-launch", "existing-launch", "missing-trust", "missing-approval", "missing-projects", "non-object-projects", "overflow"} {
			t.Run(form+"/"+variant, func(t *testing.T) {
				_, req, path := claudeRepairFixture(t, form)
				// Force repair even for monolithic, whose project entry is not a freshness condition.
				if err := os.Remove(filepath.Join(filepath.Dir(path), "CLAUDE.md")); err != nil {
					t.Fatal(err)
				}
				object := claudeNumberObject(t, []byte(`{
					"cachedGrowthBookFeatures":{"probability":0.125,"rollout":1e-7},
					"projects":{"/other":{"lastCost":0.000123,"lastFpsAverage":59.9400,"costUSD":-1.25}},
					"large":9007199254740993123456789,"negative":-42,"minusZero":-0,
					"exponent":1E+09,"nested":[0.25,[-2.5,1e-7],{"unknown":null}],
					"unknown":{"text":"tool state","enabled":false}
				}`))
				switch variant {
				case "overflow":
					object["overflow"] = json.Number("1e999")
				case "existing-launch", "missing-trust", "missing-approval":
					entry := map[string]any{
						"lastCost": json.Number("0.000000125"), "hasTrustDialogAccepted": false,
						"hasClaudeMdExternalIncludesApproved": "pending", "unknown": []any{json.Number("1e-7"), nil},
					}
					if variant == "missing-trust" {
						entry["hasClaudeMdExternalIncludesApproved"] = true
					}
					if variant == "missing-approval" {
						entry["hasTrustDialogAccepted"] = true
					}
					object["projects"].(map[string]any)[req.LaunchDir] = entry
				case "missing-projects":
					delete(object, "projects")
				case "non-object-projects":
					object["projects"] = []any{json.Number("0.5")}
				}
				before, err := json.Marshal(object)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, before, 0o600); err != nil {
					t.Fatal(err)
				}
				// Independent expected merge: legacy non-object projects are replaced; all other values survive.
				want := claudeNumberObject(t, before)
				projects, ok := want["projects"].(map[string]any)
				if !ok {
					projects = map[string]any{}
					want["projects"] = projects
				}
				entry, ok := projects[req.LaunchDir].(map[string]any)
				if !ok {
					entry = map[string]any{}
					projects[req.LaunchDir] = entry
				}
				entry["hasTrustDialogAccepted"] = true
				if form == envregistry.FormReferenced {
					entry["hasClaudeMdExternalIncludesApproved"] = true
				}
				if _, err := Resolve(req); err != nil {
					t.Fatalf("repair numeric tool state: %v", err)
				}
				got := claudeNumberObject(t, readClaudeState(t, path))
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("repair changed unowned values or number texts:\ngot  %#v\nwant %#v", got, want)
				}
				req.Repair = false
				if _, err := Resolve(req); err != nil {
					t.Fatalf("verify numeric tool state: %v", err)
				}
			})
		}
	}
}

// Resolve's repair path leaves satisfied Claude state byte-identical and retains its inode and mtime even when another surface needs repair.
func TestClaudeRepairSatisfiedIsNoOp(t *testing.T) {
	for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
		t.Run(form, func(t *testing.T) {
			_, req, path := claudeRepairFixture(t, form)
			key, err := json.Marshal(req.LaunchDir)
			if err != nil {
				t.Fatal(err)
			}
			// A monolithic home does not require or overwrite external-includes approval.
			approval := "false"
			if form == envregistry.FormReferenced {
				approval = "true"
			}
			before := []byte("{\n  \"unknown\": 1e-7, \"projects\": {" + string(key) + ": {\"hasTrustDialogAccepted\":true,\"hasClaudeMdExternalIncludesApproved\":" + approval + ",\"lastCost\":0.125}}\n}\n")
			if err := os.WriteFile(path, before, 0o600); err != nil {
				t.Fatal(err)
			}
			stamp := time.Unix(1234567890, 0)
			if err := os.Chtimes(path, stamp, stamp); err != nil {
				t.Fatal(err)
			}
			old, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(filepath.Join(filepath.Dir(path), "CLAUDE.md")); err != nil {
				t.Fatal(err)
			}
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			current, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(readClaudeState(t, path), before) || !current.ModTime().Equal(old.ModTime()) || !os.SameFile(old, current) {
				t.Fatal("satisfied .claude.json was rewritten")
			}
		})
	}
}

// Resolve provisioning retains the pre-fix CCJ-1 seed bytes for both forms (the launch path is the only dynamic value).
func TestClaudeAbsentSeedGolden(t *testing.T) {
	for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
		t.Run(form, func(t *testing.T) {
			_, req, path := claudeRepairFixture(t, form)
			key, err := json.Marshal(req.LaunchDir)
			if err != nil {
				t.Fatal(err)
			}
			entry := `{"hasTrustDialogAccepted":true}`
			if form == envregistry.FormReferenced {
				entry = `{"hasClaudeMdExternalIncludesApproved":true,"hasTrustDialogAccepted":true}`
			}
			want := "{\"hasCompletedOnboarding\":true,\"projects\":{" + string(key) + ":" + entry + "}}\n"
			if got := string(readClaudeState(t, path)); got != want {
				t.Fatalf("seed bytes changed: got %q, want %q", got, want)
			}
		})
	}
}

// Resolve refuses invalid/unreadable seed state without replacing the leaf or changing its bytes/target; chmod and symlink cases require host support.
func TestClaudeRepairRefusesUnsafeState(t *testing.T) {
	for _, variant := range []string{"array", "null", "scalar", "malformed", "symlink", "directory", "unreadable"} {
		t.Run(variant, func(t *testing.T) {
			if variant == "unreadable" && runtime.GOOS == "windows" {
				t.Skip("POSIX mode-bit unreadability")
			}
			if variant == "unreadable" && os.Geteuid() == 0 {
				t.Skip("this environment can read a mode-000 file")
			}
			_, req, path := claudeRepairFixture(t, envregistry.FormReferenced)
			before := []byte(`{"toolState":0.125}`)
			diagnostic := envregistry.DiagSeedUnreadable
			switch variant {
			case "array":
				before = []byte(`[]`)
			case "null":
				before = []byte(`null`)
			case "scalar":
				before = []byte(`42`)
			case "malformed":
				before = []byte(`{"broken":`)
			case "symlink", "directory":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if variant == "directory" {
					if err := os.Mkdir(path, 0o700); err != nil {
						t.Fatal(err)
					}
				} else {
					target := filepath.Join(t.TempDir(), "tool-state.json")
					if err := os.WriteFile(target, before, 0o600); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(target, path); err != nil {
						t.Fatal(err)
					}
					diagnostic = DiagWriteWouldFollowLink
				}
			}
			if variant != "symlink" && variant != "directory" {
				if err := os.WriteFile(path, before, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if variant == "unreadable" {
				// Seed validation refuses the read before any backup or managed write.
				if err := os.Chmod(path, 0); err != nil {
					t.Fatal(err)
				}
			}
			old, err := os.Lstat(path)
			if err != nil {
				t.Fatal(err)
			}
			_, repairErr := Resolve(req)
			if repairErr == nil || !strings.Contains(repairErr.Error(), diagnostic) {
				t.Fatalf("Resolve = %v, want %s refusal", repairErr, diagnostic)
			}
			current, err := os.Lstat(path)
			if err != nil {
				t.Fatal(err)
			}
			if !os.SameFile(old, current) || old.Mode() != current.Mode() || !old.ModTime().Equal(current.ModTime()) {
				t.Fatal("refusal changed the tool-owned leaf")
			}
			if variant == "unreadable" {
				if err := os.Chmod(path, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if variant == "directory" {
				entries, err := os.ReadDir(path)
				if err != nil || len(entries) != 0 {
					t.Fatalf("refusal changed directory: %v, %v", entries, err)
				}
			} else if !bytes.Equal(readClaudeState(t, path), before) {
				t.Fatal("refusal changed tool bytes (or the symlink target)")
			}
		})
	}
}
