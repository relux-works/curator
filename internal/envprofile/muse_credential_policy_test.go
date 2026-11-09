package envprofile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

// Shared native/unset must leave Muse's default store available (the global
// Keychain on the measured macOS release), retaining the auth.json fallback
// link. Fixtures prove projection only; they never access a real Keychain.
func TestMuseSharedNativeKeepsDefaultBackend(t *testing.T) {
	for _, selection := range []envregistry.CredentialStoreSelection{"native", ""} {
		t.Run(string(selection), func(t *testing.T) {
			fx, auth, native := museFixture(t)
			req := fx.request(envregistry.Muse)
			if selection != "" {
				req.Machine.CredentialStore = map[string]map[string]envregistry.CredentialStoreSelection{fx.profile: {envregistry.Muse: selection}}
			}
			result, err := Resolve(req)
			if err != nil {
				t.Fatal(err)
			}
			var fragment struct {
				Env             map[string]string
				CredentialStore struct{ Selection string } `json:"credential_store"`
			}
			if err := json.Unmarshal(result.Document, &fragment); err != nil {
				t.Fatal(err)
			}
			if backend, present := fragment.Env["TBH_CREDENTIAL_BACKEND"]; present {
				t.Fatalf("shared native must not force a backend, got %q", backend)
			}
			if fragment.CredentialStore.Selection != "native" {
				t.Fatalf("effective selection=%q, want native", fragment.CredentialStore.Selection)
			}
			if strings.Contains(strings.Join(result.Warnings, ";"), "maps configured native") {
				t.Fatal("shared native still reports a file mapping")
			}
			if got, err := os.Readlink(auth); err != nil || got != native {
				t.Fatalf("fallback link=%q, %v; want %q", got, err, native)
			}
		})
	}
}

// Shared native/unset preserves the vendor default; explicit file still forces
// file. Isolated native/unset still maps to file and warns, preventing access
// to a global Keychain login. Selectors and fallback link rules stay unchanged.
// Isolated never imports a native credential; shared requires a live native file.
// Keychain refuses before any home, marker, link or credential mutation.
func TestMuseCredentialPolicyMatrix(t *testing.T) {
	requireLinkCapability(t)
	for _, selection := range []envregistry.CredentialStoreSelection{"", "native", "file", "keychain"} {
		for _, isolation := range []string{"shared", "isolated"} {
			t.Run(string(selection)+"/"+isolation, func(t *testing.T) {
				fx := writeManagedFixture(t, "muse-policy")
				fx.native[envregistry.Muse] = t.TempDir()
				native := filepath.Join(fx.native[envregistry.Muse], "auth.json")
				if err := os.WriteFile(native, []byte("synthetic-native\n"), 0600); err != nil {
					t.Fatal(err)
				}
				req := fx.request(envregistry.Muse)
				req.Repair = true
				req.Machine.Isolation[fx.profile] = map[string]string{envregistry.Muse: isolation}
				if selection != "" {
					req.Machine.CredentialStore = map[string]map[string]envregistry.CredentialStoreSelection{fx.profile: {envregistry.Muse: selection}}
				}
				before := snapshotCredentialScope(t, fx)
				configured, _ := json.Marshal(req.Machine.CredentialStore)
				result, err := Resolve(req)
				if selection == "keychain" {
					if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialStoreUnsupported) || result != nil {
						t.Fatalf("keychain: %v %v", result, err)
					}
					if !reflect.DeepEqual(before, snapshotCredentialScope(t, fx)) {
						t.Fatal("refusal changed forbidden state")
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				after, _ := json.Marshal(req.Machine.CredentialStore)
				if string(configured) != string(after) {
					t.Fatal("selector rewritten")
				}
				var fragment struct {
					Fragment        string
					Env             map[string]string
					CredentialStore struct {
						Selection string `json:"selection"`
					} `json:"credential_store"`
					RootContext any `json:"root_context"`
				}
				if err := json.Unmarshal(result.Document, &fragment); err != nil {
					t.Fatal(err)
				}
				wantStore, wantBackend, wantEnvCount := "file", "file", 5
				if isolation == "shared" && selection != "file" {
					wantStore, wantBackend, wantEnvCount = "native", "", 4
				}
				backend, forced := fragment.Env["TBH_CREDENTIAL_BACKEND"]
				if fragment.Fragment != "launch-env-fragment-v4" || len(fragment.Env) != wantEnvCount || backend != wantBackend || forced != (wantStore == "file") || fragment.CredentialStore.Selection != wantStore || fragment.RootContext == nil {
					t.Fatalf("projection: %s", result.Document)
				}
				wantWarning := isolation == "isolated" && selection != "file"
				if strings.Contains(strings.Join(result.Warnings, ";"), "maps configured native") != wantWarning {
					t.Fatalf("mapping warning: %v, want warning=%v", result.Warnings, wantWarning)
				}
				auth := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.Muse), "config", "muse", "auth.json")
				if isolation == "shared" {
					if got, err := os.Readlink(auth); err != nil || got != native {
						t.Fatalf("shared source: %q %v", got, err)
					}
				} else if _, err := os.Lstat(auth); !os.IsNotExist(err) {
					t.Fatalf("isolated imported or synthesized auth: %v", err)
				}
				for _, format := range []string{"json", "env", "shell"} {
					req.Format = format
					req.Repair = false
					r, err := Resolve(req)
					if err != nil {
						t.Fatal(err)
					}
					if strings.Contains(string(r.Document), "TBH_CREDENTIAL_BACKEND") != (wantStore == "file") {
						t.Fatalf("%s: %v %v", format, r, err)
					}
				}
				if got, err := os.ReadFile(native); err != nil || string(got) != "synthetic-native\n" {
					t.Fatal("native bytes changed")
				}
			})
		}
	}
}

// A missing shared source refuses with actionable native-file guidance and no
// writes; the same missing native file is irrelevant to a new isolated home.
func TestMuseMissingNativeFilePolicy(t *testing.T) {
	for _, mode := range []string{"shared", "isolated"} {
		t.Run(mode, func(t *testing.T) {
			fx := writeManagedFixture(t, "muse-missing")
			fx.native[envregistry.Muse] = t.TempDir()
			req := fx.request(envregistry.Muse)
			req.Repair = true
			req.Machine.Isolation[fx.profile] = map[string]string{envregistry.Muse: mode}
			before := snapshotCredentialScope(t, fx)
			result, err := Resolve(req)
			if mode == "shared" {
				if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) || !strings.Contains(err.Error(), "TBH_CREDENTIAL_BACKEND=file muse login") || result != nil {
					t.Fatalf("missing shared source: %v %v", result, err)
				}
				if !reflect.DeepEqual(before, snapshotCredentialScope(t, fx)) {
					t.Fatal("missing native source caused writes")
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Lstat(filepath.Join(fx.native[envregistry.Muse], "auth.json")); !os.IsNotExist(err) {
				t.Fatal("native placeholder created")
			}
		})
	}
}

// Repair cannot apply a shared-to-isolated policy change or erase a detached
// vendor-owned file. The unchanged shared control resolves before each switch.
func TestMusePolicySwitchPreservesCredentials(t *testing.T) {
	for _, detached := range []bool{false, true} {
		t.Run(map[bool]string{false: "link", true: "detached-file"}[detached], func(t *testing.T) {
			fx, auth, _ := museFixture(t)
			if _, err := Resolve(fx.request(envregistry.Muse)); err != nil {
				t.Fatal(err)
			}
			if detached {
				if err := os.Remove(auth); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(auth, []byte("synthetic-detached\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			req := fx.request(envregistry.Muse)
			req.Repair = true
			req.Machine.Isolation[fx.profile] = map[string]string{envregistry.Muse: "isolated"}
			before := snapshotCredentialScope(t, fx)
			result, err := Resolve(req)
			if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) || result != nil && len(result.Document) != 0 {
				t.Fatalf("switch bypassed migration: %v %v", result, err)
			}
			if !reflect.DeepEqual(before, snapshotCredentialScope(t, fx)) {
				t.Fatal("repair migrated or deleted credentials")
			}
		})
	}
}

// Explicit inspect/plan/apply may unlink the verified shared route and publish
// isolated records, preserving native credential bytes. Repair alone may not.
func TestMusePolicyExplicitMigration(t *testing.T) {
	fx, auth, native := museFixture(t)
	before, err := os.ReadFile(native)
	if err != nil {
		t.Fatal(err)
	}
	req := fx.migrateRequest()
	req.Profile = fx.profile
	req.EnvID = envregistry.Muse
	req.Machine.Isolation[fx.profile] = map[string]string{envregistry.Muse: envregistry.IsolationIsolated}
	report, err := PlanMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Conflicts()) != 0 || len(report.Ops()) != 1 || report.Ops()[0].Kind != MigrateOpUnlink {
		t.Fatalf("plan: %+v", report)
	}
	req.Expect = report.Hash
	if _, err := ApplyMigration(req); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(auth); !os.IsNotExist(err) {
		t.Fatalf("shared link survived migration: %v", err)
	}
	if got, err := os.ReadFile(native); err != nil || string(got) != string(before) {
		t.Fatal("migration changed native bytes")
	}
	resolve := fx.request(envregistry.Muse)
	resolve.Machine = req.Machine
	if _, err := Resolve(resolve); err != nil {
		t.Fatalf("migrated isolated home not current: %v", err)
	}
}

// The Muse-only change leaves every other adapter's native selection and
// launch environment untouched; explicit unsupported mappings remain covered
// by TestCredentialStorePolicyActivation and its no-provisioning assertions.
func TestMusePolicyLeavesOtherAdaptersNative(t *testing.T) {
	for _, id := range []string{envregistry.ClaudeCode, envregistry.CodexCLI, envregistry.Pi, envregistry.OpenCode} {
		t.Run(id, func(t *testing.T) {
			fx := writeManagedFixture(t, "other-native")
			req := fx.request(id)
			adapter, err := envregistry.ByID(id)
			if err != nil {
				t.Fatal(err)
			}
			policy, err := req.computeCredentialPolicy(adapter, "unknown")
			if err != nil {
				t.Fatal(err)
			}
			if policy.effectiveStore != envregistry.CredentialStoreNative || len(policy.launchEnv) != 0 {
				t.Fatalf("native policy changed: %+v", policy)
			}
		})
	}
}
