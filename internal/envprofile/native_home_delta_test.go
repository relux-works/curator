package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envregistry"
)

// An XDG seed parent may contain the machine home: only its selected children
// are linked. Both that documented layout and a separate XDG parent provision
// and verify through Resolve, retaining the external seed link in each case.
func TestRev2OpenCodeXDGContainsMachineHome(t *testing.T) {
	for _, layout := range []string{"contains-machine-home", "separate-xdg"} {
		t.Run(layout, func(t *testing.T) {
			requireLinkCapability(t)
			fx := writeManagedFixture(t, "xdg-parent")
			xdg := t.TempDir()
			if layout == "contains-machine-home" {
				xdg = filepath.Dir(fx.home)
			}
			req := deltaOpenCodeXDG(t, fx, xdg)
			if result, err := Resolve(req); err != nil || result == nil || len(result.Document) == 0 {
				t.Fatalf("outside XDG parent repair: %v", err)
			}
			link := filepath.Join(ManagedParent(fx.home, fx.profile, envregistry.OpenCode), "git")
			if target, err := os.Readlink(link); err != nil || target != filepath.Join(xdg, "git") {
				t.Fatalf("external XDG seed link: %q %v", target, err)
			}
			req.Repair = false
			if result, err := Resolve(req); err != nil || result == nil || len(result.Document) == 0 {
				t.Fatalf("outside XDG parent verification: %v", err)
			}
		})
	}
}

func deltaOpenCodeXDG(t *testing.T, fx *managedFixture, xdg string) ResolveRequest {
	t.Helper()
	pinOperatorHome(t, t.TempDir())
	t.Setenv("CURATOR_CONFIG", filepath.Join(fx.home, "config.json"))
	t.Setenv("XDG_CONFIG_HOME", xdg)
	for _, name := range []string{"opencode", "git"} {
		if err := os.MkdirAll(filepath.Join(xdg, name), 0700); err != nil {
			t.Fatal(err)
		}
	}
	req := fx.request(envregistry.OpenCode)
	req.NativeHomeOf = nil
	req.OperatorXDG = ""
	req.Repair = true
	req.Machine.XDGSeedAllowlist = []string{"git"}
	return req
}

// Read-only verification refuses when an already linked external XDG seed is
// redirected into managed ownership. Intact controls before and after pass;
// refusal emits no fragment and preserves the managed home, profile and plant.
func TestRev2OpenCodeXDGVerifyAfterSeedMoves(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "xdg-verification")
	xdg := t.TempDir()
	req := deltaOpenCodeXDG(t, fx, xdg)
	if _, err := Resolve(req); err != nil {
		t.Fatalf("outside seed provisioning control: %v", err)
	}
	req.Repair = false
	if result, err := Resolve(req); err != nil || result == nil || len(result.Document) == 0 {
		t.Fatalf("outside seed verification control: %v", err)
	}
	seed := filepath.Join(xdg, "git")
	managed := filepath.Join(EnvRoot(fx.home), "other", "seed")
	if err := os.MkdirAll(managed, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(seed); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(managed, seed); err != nil {
		t.Fatal(err)
	}
	parent := ManagedParent(fx.home, fx.profile, envregistry.OpenCode)
	before := managedCallerSnapshot(t, fx, parent, xdg, managed)
	result, err := Resolve(req)
	if err == nil || !strings.HasPrefix(err.Error(), DiagHomeStale+":") || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) || !strings.Contains(err.Error(), "XDG seed target is inside managed environments") {
		t.Fatalf("verification admitted managed XDG target: %v", err)
	}
	if result != nil && len(result.Document) != 0 {
		t.Fatal("verification refusal emitted a fragment")
	}
	if !sameReviewSnapshot(before, managedCallerSnapshot(t, fx, parent, xdg, managed)) {
		t.Fatal("verification refusal changed managed state or the XDG plant")
	}
	if err := os.Remove(seed); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(seed, 0700); err != nil {
		t.Fatal(err)
	}
	if result, err := Resolve(req); err != nil || result == nil || len(result.Document) == 0 {
		t.Fatalf("restored outside seed verification control: %v", err)
	}
}

// Public Import must surface a managed-caller selection refusal even with
// --allow-lossy. These fresh fixtures, with/without a current record, create
// no default/imported profile, source or store when selection refuses.
// Restoring the outside selector imports its actual content on the same manager.
func TestRev2ImportManagedCallerVisible(t *testing.T) {
	for _, name := range []string{"strict", "allow-lossy", "current-profile"} {
		t.Run(name, func(t *testing.T) {
			home := gitFixtureHome(t)
			operator := t.TempDir()
			pinOperatorHome(t, operator)
			homes := pinHomes(t)
			t.Setenv("CURATOR_CONFIG", filepath.Join(home, "config.json"))
			outside := homes[envregistry.CodexCLI]
			caller := filepath.Join(EnvRoot(home), "caller", envregistry.CodexCLI)
			writeNativeFile(t, outside, "AGENTS.md", "selected outside\n")
			writeNativeFile(t, filepath.Join(operator, ".codex"), "AGENTS.md", "default outside\n")
			writeNativeFile(t, caller, "AGENTS.md", "managed caller\n")
			t.Setenv("CODEX_HOME", caller)
			if name == "current-profile" {
				seedCurrentDefault(t, home)
			}
			options := ImportOptions{AllowLossy: name == "allow-lossy"}
			info, activated, updated, err := Import(home, options)
			if err == nil || !strings.HasPrefix(err.Error(), envregistry.DiagCredentialConflict+":") || !strings.Contains(err.Error(), caller) {
				t.Fatalf("Import hid native selection refusal: %v", err)
			}
			if info.Name != "" || activated || updated || len(info.Warnings) != 0 {
				t.Fatalf("refused Import reported success: %+v %v %v", info, activated, updated)
			}
			absent := []string{ProfileDir(home, DefaultProfile), ProfileDir(home, ImportProfileDefault), filepath.Join(home, "imports"), contextstore.Root(home)}
			if name != "current-profile" {
				absent = append(absent, ProfilesDir(home))
			} else if data, err := os.ReadFile(CurrentFile(home)); err != nil || string(data) != DefaultProfile+"\n" {
				t.Fatalf("refused Import changed the current profile: %q %v", data, err)
			}
			for _, path := range absent {
				if _, err := os.Lstat(path); !os.IsNotExist(err) {
					t.Fatalf("refused Import created profile/source/store state: %s: %v", filepath.Base(path), err)
				}
			}
			for path, want := range map[string]string{caller: "managed caller\n", outside: "selected outside\n", filepath.Join(operator, ".codex"): "default outside\n"} {
				if data, err := os.ReadFile(filepath.Join(path, "AGENTS.md")); err != nil || string(data) != want {
					t.Fatalf("Import changed a native input: %q %v", data, err)
				}
			}
			t.Setenv("CODEX_HOME", outside)
			// Keep the successful import inactive; activation has separate coverage.
			seedCurrentDefault(t, home)
			info, _, _, err = Import(home, options)
			if err != nil || info.Name != ImportProfileDefault {
				t.Fatalf("outside Import positive control: %v", err)
			}
			source, err := readSource(home, info.Name)
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(source.Path, "context", "codex_cli.md"))
			if err != nil || string(data) != "selected outside\n" {
				t.Fatalf("outside Import lost selected content: %q %v", data, err)
			}
		})
	}
}
