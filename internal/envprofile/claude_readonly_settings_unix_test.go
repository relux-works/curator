//go:build unix

package envprofile

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

func enableClaudeReadOnly(req *ResolveRequest) {
	req.Machine.ClaudeSettingsMode = map[string]string{req.Profile: "read-only"}
}

// Fresh provisioning, repeated repair and re-materialisation keep mandatory
// settings regular/read-only; absent optional settings are not fabricated.
func TestClaudeReadOnlySettingsFresh(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	req := fx.request(envregistry.ClaudeCode)
	enableClaudeReadOnly(&req)
	req.Repair = true
	for i := 0; i < 2; i++ {
		if _, err := Resolve(req); err != nil {
			t.Fatal(err)
		}
		home := ManagedHomeDir(fx.home, fx.profile, req.EnvID)
		assertManagedMode(t, filepath.Join(home, "settings.json"), 0o400)
		for _, path := range []string{"cowork_settings.json", ".config.json"} {
			if _, err := os.Lstat(filepath.Join(home, path)); !os.IsNotExist(err) {
				t.Fatalf("optional %s created: %v", path, err)
			}
		}
		if i == 0 {
			if err := os.Remove(filepath.Join(home, "settings.json")); err != nil {
				t.Fatal(err)
			}
		}
	}
	req.Repair = false
	if _, err := Resolve(req); err != nil {
		t.Fatal("fresh repaired home is stale", err)
	}
}

// All three settings preserve bytes on mode repair. Open descriptors retain
// the old inode, bytes and mode across replacement, proving repair never
// chmods or writes the live entry. Protected tool state retains bytes/modes.
func TestClaudeReadOnlySettingsRepairAtomicAndProtected(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	req := fx.request(envregistry.ClaudeCode)
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	home := ManagedHomeDir(fx.home, fx.profile, req.EnvID)
	protected := map[string][]byte{}
	for _, path := range []string{".claude.json", ".credentials.json", "policy-limits.json", "remote-settings.json", "policy-limits.json.etag", "remote-settings.json.etag", "policy-limits.json.lock", "remote-settings.json.lock"} {
		full := filepath.Join(home, path)
		if path != ".claude.json" && path != ".credentials.json" {
			if err := os.WriteFile(full, []byte("opaque tool state\n"), 0o640); err != nil {
				t.Fatal(err)
			}
		} else if path == ".credentials.json" {
			// On macOS no credential link is created; on Linux preserve the link.
			if _, err := os.Lstat(full); os.IsNotExist(err) {
				if err := os.WriteFile(full, []byte("synthetic credential sentinel\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
		}
		data, err := os.ReadFile(full)
		if err != nil {
			t.Fatal(err)
		}
		protected[path] = data
	}
	before := map[string]os.FileInfo{}
	for path := range protected {
		info, err := os.Stat(filepath.Join(home, path))
		if err != nil {
			t.Fatal(err)
		}
		before[path] = info
	}
	enableClaudeReadOnly(&req)
	for _, path := range claudeSettingsPaths {
		full := filepath.Join(home, path)
		if err := os.WriteFile(full, []byte("{\"operator\":1}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	readOnly := req
	readOnly.Repair = false
	diagnostic := "Claude settings settings.json has write bits (0600); claude_settings_mode=read-only requires no write bits"
	if _, err := Resolve(readOnly); err == nil || !strings.Contains(err.Error(), diagnostic) {
		t.Fatalf("writable settings not reported: %v", err)
	}
	dry := req
	dry.DryRun = true
	if _, err := Resolve(dry); err == nil || !strings.Contains(err.Error(), diagnostic) {
		t.Fatalf("dry-run: %v", err)
	}
	for _, path := range claudeSettingsPaths {
		assertManagedMode(t, filepath.Join(home, path), 0o600)
	}
	for _, path := range claudeSettingsPaths {
		full := filepath.Join(home, path)
		if err := os.Chmod(full, 0o600); err != nil {
			t.Fatal(err)
		}
		old, err := os.Open(full)
		if err != nil {
			t.Fatal(err)
		}
		defer old.Close()
		oldInfo, err := old.Stat()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Resolve(req); err != nil {
			t.Fatal(err)
		}
		newInfo, err := os.Stat(full)
		if err != nil {
			t.Fatal(err)
		}
		if os.SameFile(oldInfo, newInfo) {
			t.Fatal("repair wrote in place")
		}
		held, err := old.Stat()
		if err != nil {
			t.Fatal(err)
		}
		// Only the first repair sees 0600; later old entries are already 0400.
		if held.Mode() != oldInfo.Mode() {
			t.Fatal("repair chmodded old live entry")
		}
		assertManagedMode(t, full, 0o400)
		data, err := io.ReadAll(old)
		if err != nil || string(data) != "{\"operator\":1}\n" {
			t.Fatalf("old bytes changed: %q %v", data, err)
		}
		// Exercise a content update through the same atomic writer used by repair.
		rel, err := managedRelative(EnvRoot(fx.home), full)
		if err != nil {
			t.Fatal(err)
		}
		if err := atomicManagedFile(EnvRoot(fx.home), rel, []byte("{\"operator\":2}\n"), claudeReadOnlySettingsMode); err != nil {
			t.Fatal(err)
		}
		if _, err := Resolve(req); err != nil {
			t.Fatal(err)
		}
		data, err = os.ReadFile(full)
		if err != nil || string(data) != "{\"operator\":2}\n" {
			t.Fatalf("updated bytes lost: %q %v", data, err)
		}
		assertManagedMode(t, full, 0o400)
	}
	for path, want := range protected {
		full := filepath.Join(home, path)
		data, err := os.ReadFile(full)
		if err != nil || string(data) != string(want) {
			t.Fatalf("protected %s bytes changed: %v", path, err)
		}
		info, err := os.Stat(full)
		if err != nil || info.Mode() != before[path].Mode() {
			t.Fatalf("protected %s mode changed: %v", path, err)
		}
	}
}

// Default/explicit native, a different profile and all other adapters ignore
// this Claude-only knob: existing settings keep 0600 and default authored
// context remains 0600.
func TestClaudeReadOnlySettingsScope(t *testing.T) {
	for _, id := range []string{envregistry.ClaudeCode, envregistry.CodexCLI, envregistry.Pi, envregistry.OpenCode, envregistry.Muse} {
		for _, mode := range []string{"", "native", "other-profile", "read-only"} {
			if id == envregistry.ClaudeCode && mode == "read-only" {
				continue
			}
			t.Run(id+"/"+mode, func(t *testing.T) {
				fx := writeManagedFixture(t, "acme")
				seedLiveNativeCredentials(t, fx)
				fx.native[envregistry.Muse] = t.TempDir()
				if err := os.WriteFile(filepath.Join(fx.native[envregistry.Muse], "auth.json"), []byte("synthetic\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				req := fx.request(id)
				req.Repair = true
				if _, err := Resolve(req); err != nil {
					t.Fatal(err)
				}
				home := ManagedHomeDir(fx.home, fx.profile, id)
				path := filepath.Join(home, "settings.json")
				if err := os.WriteFile(path, []byte("{\"operator\":true}\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				profile := fx.profile
				if mode == "other-profile" {
					profile = "other"
				}
				if mode != "" {
					req.Machine.ClaudeSettingsMode = map[string]string{profile: mode}
					if mode == "other-profile" {
						req.Machine.ClaudeSettingsMode[profile] = "read-only"
					}
				}
				if _, err := Resolve(req); err != nil {
					t.Fatal(err)
				}
				assertManagedMode(t, path, 0o600)
				data, err := os.ReadFile(path)
				if err != nil || string(data) != "{\"operator\":true}\n" {
					t.Fatal("unrelated settings changed", err)
				}
				if id == envregistry.ClaudeCode {
					assertManagedMode(t, filepath.Join(home, "CLAUDE.md"), 0o600)
				}
			})
		}
	}
}

// Present nonregular settings refuse before any home write; the symlink's
// foreign target and the existing marker are preserved. A regular 0444 file
// is admitted because the contract bans write bits, not group read access.
func TestClaudeReadOnlySettingsNonregular(t *testing.T) {
	for _, path := range claudeSettingsPaths {
		for _, kind := range []string{"link", "directory", "regular"} {
			t.Run(path+"/"+kind, func(t *testing.T) {
				fx := writeManagedFixture(t, "acme")
				seedLiveNativeCredentials(t, fx)
				req := fx.request(envregistry.ClaudeCode)
				req.Repair = true
				if _, err := Resolve(req); err != nil {
					t.Fatal(err)
				}
				home := ManagedHomeDir(fx.home, fx.profile, req.EnvID)
				markerPath := filepath.Join(home, ".agent-environment.json")
				marker, err := os.ReadFile(markerPath)
				if err != nil {
					t.Fatal(err)
				}
				target := filepath.Join(t.TempDir(), "foreign")
				if err := os.WriteFile(target, []byte("foreign\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				full := filepath.Join(home, path)
				switch kind {
				case "link":
					if err := os.Symlink(target, full); err != nil {
						t.Fatal(err)
					}
				case "directory":
					if err := os.Mkdir(full, 0o700); err != nil {
						t.Fatal(err)
					}
				case "regular":
					if err := os.WriteFile(full, []byte("{}\n"), 0o444); err != nil {
						t.Fatal(err)
					}
				}
				enableClaudeReadOnly(&req)
				_, err = Resolve(req)
				if kind == "regular" {
					if err != nil {
						t.Fatal(err)
					}
					return
				}
				if err == nil || !strings.Contains(err.Error(), "environment_claude_settings_invalid: "+path) {
					t.Fatalf("nonregular accepted: %v", err)
				}
				after, err := os.ReadFile(markerPath)
				if err != nil || string(after) != string(marker) {
					t.Fatal("refusal changed marker", err)
				}
				data, err := os.ReadFile(target)
				if err != nil || string(data) != "foreign\n" {
					t.Fatal("foreign target changed", err)
				}
				assertManagedMode(t, target, 0o600)
			})
		}
	}
}

// Owner, group and other write bits independently make every settings path
// stale. Read-only resolve preserves each violating mode; repair removes it.
// A regular 0444 positive control stays current without being rewritten.
func TestClaudeReadOnlySettingsWriteBits(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	req := fx.request(envregistry.ClaudeCode)
	enableClaudeReadOnly(&req)
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	home := ManagedHomeDir(fx.home, fx.profile, req.EnvID)
	for _, path := range claudeSettingsPaths {
		full := filepath.Join(home, path)
		if err := atomicManagedFile(EnvRoot(fx.home), mustSettingsRel(t, fx.home, full), []byte("{}\n"), 0o444); err != nil {
			t.Fatal(err)
		}
		readOnly := req
		readOnly.Repair = false
		if _, err := Resolve(readOnly); err != nil {
			t.Fatal("0444 rejected", err)
		}
		assertManagedMode(t, full, 0o444)
		for _, mode := range []os.FileMode{0o600, 0o420, 0o402} {
			if err := os.Chmod(full, mode); err != nil {
				t.Fatal(err)
			}
			if _, err := Resolve(readOnly); err == nil || !strings.Contains(err.Error(), "Claude settings "+path+" has write bits") {
				t.Fatalf("%s/%04o: %v", path, mode, err)
			}
			assertManagedMode(t, full, mode)
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			assertManagedMode(t, full, 0o400)
		}
	}
}

func mustSettingsRel(t *testing.T, home, full string) string {
	t.Helper()
	rel, err := managedRelative(EnvRoot(home), full)
	if err != nil {
		t.Fatal(err)
	}
	return rel
}
