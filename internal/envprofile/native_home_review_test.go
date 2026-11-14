package envprofile

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

// Every row uses production Resolve --repair, with a fresh outside-origin
// positive control. Refused selectors emit no fragment or fallback-success
// warning, and never publish a home. No live home or credential is inspected.
func TestNativeHomeReviewRepair(t *testing.T) {
	for _, mode := range []string{"xdg-managed", "xdg-child-alias", "all-managed", "all-managed-unset", "parent-traversal", "native-ancestor", "absent-root-case", "muse-managed", "muse-explicit"} {
		t.Run(mode, func(t *testing.T) {
			requireLinkCapability(t)
			fx := writeManagedFixture(t, "selection-review")
			operator := t.TempDir()
			t.Setenv("HOME", operator)
			t.Setenv("USERPROFILE", operator)
			t.Setenv("CURATOR_CONFIG", filepath.Join(fx.home, "config.json"))
			id := "codex_cli"
			if strings.HasPrefix(mode, "xdg-") {
				id = "opencode"
			}
			if strings.HasPrefix(mode, "muse-") {
				id = envregistry.Muse
			}
			native := filepath.Join(operator, ".codex")
			xdg := filepath.Join(operator, ".config")
			if id != "codex_cli" {
				native = filepath.Join(xdg, id)
			}
			if err := os.MkdirAll(native, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(native, "auth.json"), []byte("synthetic\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if id == "codex_cli" {
				if err := os.WriteFile(filepath.Join(native, "config.toml"), []byte("cli_auth_credentials_store = \"file\"\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			for _, name := range []string{"git", "gh", "ssh"} {
				if err := os.MkdirAll(filepath.Join(xdg, name), 0700); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("CODEX_HOME", native)
			t.Setenv("XDG_CONFIG_HOME", xdg)
			req := fx.request(id)
			req.NativeHomeOf = nil
			req.OperatorXDG = ""
			req.Repair = true
			req.Machine.XDGSeedAllowlist = []string{"git", "gh", "ssh"}
			// Provision a separate manager to prove the same native inputs are valid,
			// while leaving the negative manager's publication path absent.
			good := req
			if !strings.HasPrefix(mode, "muse-") {
				control := writeManagedFixture(t, fx.profile)
				good.Home = control.home
			}
			if result, err := Resolve(good); err != nil || result == nil {
				t.Fatalf("outside repair positive control: %v", err)
			}
			if id == "opencode" {
				for _, name := range []string{"git", "gh", "ssh"} {
					got, err := os.Readlink(filepath.Join(ManagedParent(good.Home, fx.profile, id), name))
					if err != nil || got != filepath.Join(xdg, name) {
						t.Fatalf("outside seed control %s: %q %v", name, got, err)
					}
				}
			}
			// Rejection preserves the profile and publication state, whether absent
			// or already provisioned by the Muse positive control.
			before := managedCallerSnapshot(t, fx, ManagedParent(fx.home, fx.profile, id))
			switch mode {
			case "xdg-managed", "muse-managed", "muse-explicit":
				managed := filepath.Join(EnvRoot(fx.home), "other", "opencode")
				for _, name := range []string{"git", "gh", "ssh", "muse"} {
					if err := os.MkdirAll(filepath.Join(managed, name), 0700); err != nil {
						t.Fatal(err)
					}
				}
				t.Setenv("XDG_CONFIG_HOME", managed)
				if mode == "muse-explicit" {
					req.OperatorXDG = managed
					t.Setenv("XDG_CONFIG_HOME", xdg)
				}
				if strings.HasPrefix(mode, "muse-") {
					result, err := Resolve(req)
					if err != nil || result == nil || !strings.Contains(strings.Join(result.Warnings, "\n"), nativeHomeFallbackNotice) {
						t.Fatalf("Muse shared fallback: %v", err)
					}
					target, err := os.Readlink(filepath.Join(ManagedHomeDir(fx.home, fx.profile, id), "config", "muse", "auth.json"))
					if err != nil || target != filepath.Join(native, "auth.json") {
						t.Fatalf("Muse fallback credential origin %q %v", target, err)
					}
					return
				}
			case "xdg-child-alias":
				managed := filepath.Join(EnvRoot(fx.home), "other", "seed")
				if err := os.MkdirAll(managed, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(filepath.Join(xdg, "git")); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(managed, filepath.Join(xdg, "git")); err != nil {
					t.Fatal(err)
				}
			case "all-managed", "all-managed-unset":
				home := filepath.Join(EnvRoot(fx.home), "caller", "codex_cli")
				if err := os.MkdirAll(home, 0700); err != nil {
					t.Fatal(err)
				}
				t.Setenv("HOME", home)
				t.Setenv("USERPROFILE", home)
				t.Setenv("CODEX_HOME", home)
				if mode == "all-managed-unset" {
					t.Setenv("CODEX_HOME", "")
				}
			case "parent-traversal":
				managed := filepath.Join(EnvRoot(fx.home), "other", "sub")
				if err := os.MkdirAll(managed, 0700); err != nil {
					t.Fatal(err)
				}
				alias := filepath.Join(t.TempDir(), "alias")
				if err := os.Symlink(managed, alias); err != nil {
					t.Fatal(err)
				}
				t.Setenv("CODEX_HOME", alias+string(filepath.Separator)+".."+string(filepath.Separator)+"codex_cli")
			case "native-ancestor":
				t.Setenv("CODEX_HOME", fx.home)
			case "absent-root-case":
				// This manager has no environments root. The real repair must refuse
				// the unresolved case alias before publishing that root.
				absent := fx.home
				selected := filepath.Join(absent, "ENVIRONMENTS", "caller", "codex_cli")
				t.Setenv("CODEX_HOME", selected)
			}
			result, err := Resolve(req)
			if err == nil || !strings.HasPrefix(err.Error(), envregistry.DiagCredentialConflict+":") || !strings.Contains(err.Error(), "run repair with the intended external native home") {
				t.Fatalf("named native refusal: %v", err)
			}
			if strings.HasPrefix(mode, "all-managed") && err.Error() != "environment_credential_conflict: native home "+filepath.Join(os.Getenv("HOME"), ".codex")+" is inside managed environments; run repair with the intended external native home" {
				t.Fatalf("refusal lacks default path: %v", err)
			}
			if mode == "absent-root-case" && !strings.Contains(err.Error(), "ambiguous case identity") {
				t.Fatalf("absent-root case refusal: %v", err)
			}
			if result != nil && (len(result.Document) != 0 || strings.Contains(strings.Join(result.Warnings, "\n"), nativeHomeFallbackNotice)) {
				t.Fatalf("refusal emitted fragment or fallback-success notice: %+v", result)
			}
			if !sameReviewSnapshot(before, managedCallerSnapshot(t, fx, ManagedParent(fx.home, fx.profile, id))) {
				t.Fatal("refused repair changed managed/profile state")
			}
		})
	}
}

func sameReviewSnapshot(a, b map[string]string) bool { return reflect.DeepEqual(a, b) }

// The public selector, shared by switch/unmanage/import/purge, refuses
// a managed caller where its signature cannot carry a fallback notice. A real
// outside home still succeeds; Resolve's reporting fallback has its own control.
func TestNativeHomeReviewUnreportedFallbackRefuses(t *testing.T) {
	manager := t.TempDir()
	operator := t.TempDir()
	t.Setenv("CURATOR_CONFIG", filepath.Join(manager, "config.json"))
	t.Setenv("HOME", operator)
	t.Setenv("USERPROFILE", operator)
	adapter, _ := adapterByID("codex_cli")
	t.Setenv("CODEX_HOME", filepath.Join(operator, ".codex"))
	if got, err := NativeHome(adapter); err != nil || got != os.Getenv("CODEX_HOME") {
		t.Fatalf("outside public selector: %q %v", got, err)
	}
	// These are the actual selection seams used by switch, import and unmanage.
	// Their outside controls succeed before the same operations refuse a caller.
	selections := []struct {
		name string
		run  func() error
	}{
		{"switch", func() error { _, err := resolveNativeHome(adapter, nil); return err }},
		{"import", func() error { _, err := importNativeHome(ImportOptions{}, adapter.ID); return err }},
		{"unmanage", func() error { _, err := planUnmanageHomes(UnmanageRequest{EnvID: adapter.ID}); return err }},
		{"purge", func() error { return purgeHomes(manager, "absent") }},
	}
	for _, operation := range selections {
		if err := operation.run(); err != nil {
			t.Fatalf("outside %s control: %v", operation.name, err)
		}
	}
	caller := filepath.Join(EnvRoot(manager), "caller", "codex_cli")
	if err := os.MkdirAll(caller, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODEX_HOME", caller)
	if got, err := NativeHome(adapter); err == nil || got != "" || !strings.HasPrefix(err.Error(), envregistry.DiagCredentialConflict+":") {
		t.Fatalf("silent fallback admitted: %q %v", got, err)
	}
	for _, operation := range selections {
		if err := operation.run(); err == nil || !strings.HasPrefix(err.Error(), envregistry.DiagCredentialConflict+":") {
			t.Fatalf("%s hid selector refusal: %v", operation.name, err)
		}
	}
}

// Global reconciliation calls the real repair path and retains its fallback
// notice in the public result. An unchanged outside caller has no such notice.
func TestNativeHomeReviewGlobalNotice(t *testing.T) {
	fx, req := managedCallerFixture(t)
	for _, managed := range []bool{false, true} {
		if managed {
			t.Setenv("CODEX_HOME", ManagedHomeDir(fx.home, fx.profile, "codex_cli"))
		}
		results, _ := syncManagedProfile(fx.home, fx.profile, req.Policy, req.Machine, nil)
		found := false
		for _, entry := range results {
			if entry.Adapter != "codex_cli" {
				continue
			}
			found = true
			notice := strings.Contains(strings.Join(entry.Warnings, "\n"), nativeHomeFallbackNotice)
			if !entry.OK || notice != managed {
				t.Fatalf("global fallback notice lost: managed=%v entry=%+v", managed, entry)
			}
		}
		if !found {
			t.Fatal("global Codex repair witness missing")
		}
	}
}
