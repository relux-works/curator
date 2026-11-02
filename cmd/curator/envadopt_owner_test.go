package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/privatedir"
)

// Actual CLI dispatch adopts only explicit historical authority for a synthetic
// legacy home, without initial account discovery. Refusals preserve the legacy
// entries; successful retries preserve the original capture and every other
// entry, lock and source. This is not an account/container/process witness.
func TestEnvAdoptOwner(t *testing.T) {
	for _, state := range []string{"healthy", "wrong-lock", "wrong-target", "missing-selector-value", "different-attestation"} {
		t.Run(state, func(t *testing.T) {
			source, manager := profileHome(t)
			profile := "fixture"
			native, anchor, today := t.TempDir(), t.TempDir(), t.TempDir()
			t.Setenv("CODEX_HOME", today)
			if err := os.WriteFile(filepath.Join(native, "auth.json"), nil, 0600); err != nil {
				t.Fatal(err)
			}
			lock := &contextlock.Lock{SchemaVersion: 1, Root: profile, Members: []contextlock.Member{{Kind: contextlock.KindContext, Name: profile, Version: "1.0.0", StateHash: strings.Repeat("a", 64), Weight: 100}}}
			lockfile := filepath.Join(envprofile.ProfileDir(manager, profile), "lock.json")
			hash, err := contextlock.Write(lockfile, lock)
			if err != nil {
				t.Fatal(err)
			}
			sourcefile := filepath.Join(envprofile.ProfileDir(manager, profile), "source.json")
			raw, err := json.Marshal(envprofile.Source{Kind: envprofile.KindLocal})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(sourcefile, raw, 0600); err != nil {
				t.Fatal(err)
			}
			target := envprofile.ManagedHomeDir(manager, profile, "codex_cli")
			if err := privatedir.MakeAll(target); err != nil {
				t.Fatal(err)
			}
			passthrough := []envmarker.Passthrough{{Path: "auth.json", Strategy: "file-link"}}
			seeds := []string{}
			marker := &envmarker.Marker{Version: 1, Mode: envmarker.ModeManagedHome, Profile: envmarker.Profile{Name: profile, Root: profile, Kind: "local", LockSHA256: strings.TrimPrefix(hash, "sha256:")}, Members: []envmarker.Member{{Name: profile, Version: "1.0.0", StateSHA256: strings.Repeat("a", 64), Weight: 100}}, Precedence: envmarker.Precedence{Winner: "higher-weight", Placement: "winner-last"}, Surfaces: map[string]envmarker.Surface{}, Passthrough: &passthrough, Seeds: &seeds}
			markerRaw, err := marker.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(target, envmarker.Name), markerRaw, 0600); err != nil {
				t.Fatal(err)
			}
			linkTarget := filepath.Join(native, "auth.json")
			if state == "wrong-target" {
				linkTarget = filepath.Join(today, "auth.json")
			}
			if err := os.Symlink(linkTarget, filepath.Join(target, "auth.json")); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(target, "session-fixture"), []byte("opaque retained session fixture"), 0600); err != nil {
				t.Fatal(err)
			}
			before := nativeOwnerSnapshot(t, target)
			lockRaw, err := os.ReadFile(lockfile)
			if err != nil {
				t.Fatal(err)
			}
			args := []string{"env", "adopt", "codex_cli", "--profile", profile, "--managed-target", target, "--lock-sha256", hash, "--operator-home", anchor, "--selector-state", "present-value", "--native-root", native}
			if state != "missing-selector-value" {
				args = append(args, "--selector-value", native)
			}
			if state == "wrong-lock" {
				for i, v := range args {
					if v == "--lock-sha256" {
						args[i+1] = "sha256:" + strings.Repeat("f", 64)
					}
				}
			}
			code, out, diag := runProfile(t, source, args...)
			originPath := filepath.Join(target, envorigin.Name)
			if state != "healthy" && state != "different-attestation" {
				want := envorigin.DiagInvalid
				if state == "wrong-target" {
					want = "credential_conflict"
				}
				if state == "missing-selector-value" {
					want = "presence differs"
				}
				if code == exitOK || out != "" || !strings.Contains(diag, want) {
					t.Fatalf("intended CLI refusal: %d %q %q", code, out, diag)
				}
				if !reflect.DeepEqual(before, nativeOwnerSnapshot(t, target)) {
					t.Fatal("adoption refusal changed legacy entries")
				}
				if _, err := os.Lstat(originPath); !os.IsNotExist(err) {
					t.Fatalf("refusal published origin: %v", err)
				}
			} else {
				if code != exitOK || out != "" {
					t.Fatalf("CLI adoption failed: %d %q %q", code, out, diag)
				}
				adopted, err := os.ReadFile(originPath)
				if err != nil {
					t.Fatal(err)
				}
				record, err := envorigin.Parse(adopted, envorigin.Binding{Profile: profile, ManagedTarget: target, LockSHA256: hash}, "codex_cli", manager)
				if err != nil || record.Capture.Method != "adoption" || record.NativeRoot != native || record.OperatorHome != anchor {
					t.Fatalf("explicit authority lost: %+v %v", record, err)
				}
				after := nativeOwnerSnapshot(t, target)
				delete(after, envorigin.Name)
				if !reflect.DeepEqual(before, after) {
					t.Fatal("CLI adoption changed another legacy entry")
				}
				if state == "different-attestation" {
					for i, v := range args {
						if v == "--operator-home" {
							args[i+1] = today
						}
					}
				}
				code, out, diag = runProfile(t, source, args...)
				if state == "different-attestation" {
					if code == exitOK || !strings.Contains(diag, "attestation differs") {
						t.Fatalf("authority replacement admitted: %d %q", code, diag)
					}
				} else if code != exitOK || out != "" {
					t.Fatalf("identical retry failed: %d %q %q", code, out, diag)
				}
				again, err := os.ReadFile(originPath)
				if err != nil || !bytes.Equal(again, adopted) {
					t.Fatal("CLI retry replaced capture")
				}
			}
			if got, err := os.ReadFile(lockfile); err != nil || !bytes.Equal(got, lockRaw) {
				t.Fatal("adoption changed installed lock")
			}
			if got, err := os.ReadFile(sourcefile); err != nil || !bytes.Equal(got, raw) {
				t.Fatal("adoption changed installed source")
			}
		})
	}
}
