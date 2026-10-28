//go:build unix

package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
)

func assertManagedMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("%s mode = %04o, want %04o (permissions must not widen)", filepath.Base(path), got, want)
	}
}

// The real fallback writer called by applyPlan keeps copies private and
// preserves owner execution bits in trees. This does not simulate symlink
// rejection by a filesystem; fallback selection has separate coverage.
func TestManagedFallbackPermissions(t *testing.T) {
	for _, tree := range []bool{false, true} {
		t.Run(fmt.Sprintf("tree-%t", tree), func(t *testing.T) {
			root := t.TempDir()
			source := filepath.Join(t.TempDir(), "source")
			if tree {
				if err := os.Mkdir(source, 0o755); err != nil {
					t.Fatal(err)
				}
				for _, mode := range []os.FileMode{0o644, 0o755, 0o400} {
					path := filepath.Join(source, fmt.Sprintf("%04o", mode))
					if err := os.WriteFile(path, []byte("fixture\n"), mode); err != nil {
						t.Fatal(err)
					}
					if err := os.Chmod(path, mode); err != nil {
						t.Fatal(err)
					}
				}
			} else if err := os.WriteFile(source, []byte("fixture\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			plan := &homePlan{marker: &envmarker.Marker{}, links: map[string]string{"copy": source}}
			if err := copyLinkFallback(root, "copy", source, "copy", plan); err != nil {
				t.Fatal(err)
			}
			if len(plan.links) != 0 {
				t.Fatal("fallback retained the pending link")
			}
			modes := map[string]os.FileMode{"": 0o600}
			if tree {
				assertManagedMode(t, filepath.Join(root, "copy"), 0o700)
				modes = map[string]os.FileMode{"0644": 0o600, "0755": 0o700, "0400": 0o400}
			}
			for name, want := range modes {
				path := filepath.Join(root, "copy", name)
				assertManagedMode(t, path, want)
				payload, err := os.ReadFile(path)
				if err != nil || string(payload) != "fixture\n" {
					t.Fatalf("fallback bytes changed: %q, %v", payload, err)
				}
			}
		})
	}
}

// Resolve provisioning preserves native rwx bits, including read-only and
// group-readable controls; revision-B TOML transformation must not widen them.
func TestManagedSeedPreservesNativePermissions(t *testing.T) {
	for _, mode := range []os.FileMode{0o600, 0o400, 0o640, 0o644} {
		t.Run(fmt.Sprintf("%04o", mode), func(t *testing.T) {
			fx := writeManagedFixture(t, "acme")
			seedLiveNativeCredentials(t, fx)
			seed := filepath.Join(fx.native[envregistry.CodexCLI], "config.toml")
			if err := os.WriteFile(seed, []byte("model = 'test'\n[mcp_servers.native]\ncommand = 'server'\n"), mode); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(seed, mode); err != nil {
				t.Fatal(err)
			}
			req := fx.request(envregistry.CodexCLI)
			req.Repair = true
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			managedSeed := filepath.Join(ManagedHomeDir(fx.home, fx.profile, req.EnvID), "config.toml")
			assertManagedMode(t, managedSeed, mode)
			payload, err := os.ReadFile(managedSeed)
			if err != nil || strings.Contains(string(payload), "mcp_servers") || !strings.Contains(string(payload), "model") {
				t.Fatalf("seed transformation failed: %q, %v", payload, err)
			}
		})
	}
}

// Resolve creates private roots/homes for every adapter, and repairs 0755
// directories even when content is current, preserving tool-owned bytes.
func TestManagedHomeRepairTightensDirectories(t *testing.T) {
	for _, id := range []string{envregistry.ClaudeCode, envregistry.CodexCLI, envregistry.OpenCode, envregistry.Pi, envregistry.Muse} {
		t.Run(id, func(t *testing.T) {
			fx := writeManagedFixture(t, "acme")
			seedLiveNativeCredentials(t, fx)
			fx.native[envregistry.Muse] = filepath.Join(fx.xdg, "muse")
			if err := os.MkdirAll(fx.native[envregistry.Muse], 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(fx.native[envregistry.Muse], "auth.json"), []byte("synthetic-native\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			req := fx.request(id)
			req.Repair = true
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			home := ManagedHomeDir(fx.home, fx.profile, id)
			paths := []string{EnvRoot(fx.home), filepath.Dir(ManagedParent(fx.home, fx.profile, id)), ManagedParent(fx.home, fx.profile, id), home}
			for _, path := range paths {
				assertManagedMode(t, path, 0o700)
			}
			for _, path := range paths {
				if err := os.Chmod(path, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			keep := filepath.Join(home, "tool-owned.txt")
			if err := os.WriteFile(keep, []byte("keep\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			// Verification and dry-run must never silently repair permissions.
			readOnly := req
			readOnly.Repair = false
			if _, err := Resolve(readOnly); err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
				t.Fatalf("wide home must be stale: %v", err)
			}
			dry := req
			dry.DryRun = true
			if _, err := Resolve(dry); err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
				t.Fatalf("dry repair must report drift: %v", err)
			}
			assertManagedMode(t, home, 0o755)
			if _, err := Resolve(req); err != nil {
				t.Fatalf("repair failed to restore private directories: %v", err)
			}
			for _, path := range paths {
				assertManagedMode(t, path, 0o700)
			}
			if payload, err := os.ReadFile(keep); err != nil || string(payload) != "keep\n" {
				t.Fatalf("repair lost contents: %q, %v", payload, err)
			}
		})
	}
}

// Resolve writes Claude's trust seed, copied context and marker owner-only;
// repair of an existing wider trust seed retains unrelated tool state.
func TestManagedAuthoredFilesArePrivate(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	req := fx.request(envregistry.ClaudeCode)
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	home := ManagedHomeDir(fx.home, fx.profile, req.EnvID)
	for _, name := range []string{".claude.json", "CLAUDE.md", envmarker.Name} {
		assertManagedMode(t, filepath.Join(home, name), 0o600)
	}
	trust := filepath.Join(home, ".claude.json")
	if err := os.WriteFile(trust, []byte(`{"projects":{},"toolOwned":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(trust, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(home, "CLAUDE.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	assertManagedMode(t, trust, 0o600)
	if payload, err := os.ReadFile(trust); err != nil || !strings.Contains(string(payload), `"toolOwned":true`) {
		t.Fatalf("trust repair lost tool state: %q, %v", payload, err)
	}
}

// A native symlink, unreadable regular file or FIFO refuses through Resolve
// before publishing any managed home. Regular readable seeds above are controls.
func TestManagedSeedRefusalsPreserved(t *testing.T) {
	for _, kind := range []string{"symlink", "unreadable", "fifo"} {
		t.Run(kind, func(t *testing.T) {
			fx := writeManagedFixture(t, "acme")
			seed := filepath.Join(fx.native[envregistry.CodexCLI], "config.toml")
			switch kind {
			case "symlink":
				target := filepath.Join(t.TempDir(), "config.toml")
				if err := os.WriteFile(target, []byte("model = 'test'\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, seed); err != nil {
					t.Fatal(err)
				}
			case "unreadable":
				if err := os.WriteFile(seed, []byte("model = 'test'\n"), 0o000); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(seed, 0o600) })
			case "fifo":
				if err := makeFIFOVectorForTest(seed); err != nil {
					t.Fatal(err)
				}
			}
			req := fx.request(envregistry.CodexCLI)
			req.Repair = true
			if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), envregistry.DiagSeedUnreadable) {
				t.Fatalf("%s seed must refuse as unreadable: %v", kind, err)
			}
			if _, err := os.Lstat(ManagedHomeDir(fx.home, fx.profile, req.EnvID)); !os.IsNotExist(err) {
				t.Fatalf("seed refusal published a home: %v", err)
			}
		})
	}
}
