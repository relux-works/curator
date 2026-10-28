package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
)

// Common mutation admission scans the actual complete fixed namespace before
// lock/recovery or generic writes. Valid recorded authority reaches a write;
// legacy, unsafe, orphaned and unknown entries cannot be hidden by selecting Q.
func TestProvisioningOriginCommonAdmission(t *testing.T) {
	for _, state := range []string{"healthy", "legacy", "unsafe-origin", "orphan-sidecar", "unknown-adapter", "broken-root", "absent-tail", "lock-drift"} {
		t.Run(state, func(t *testing.T) {
			fx, _, _ := recordedOriginFixture(t, "a")
			target := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
			switch state {
			case "lock-drift":
				lock, _, e := readLock(fx.home, fx.profile)
				if e != nil {
					t.Fatal(e)
				}
				lock.Members[0].Weight++
				if _, e := contextlock.Write(lockPath(fx.home, fx.profile), lock); e != nil {
					t.Fatal(e)
				}
			case "legacy":
				if e := os.Remove(filepath.Join(target, envorigin.Name)); e != nil {
					t.Fatal(e)
				}
			case "unsafe-origin":
				if e := os.Chmod(filepath.Join(target, envorigin.Name), 0644); e != nil {
					t.Fatal(e)
				}
			case "orphan-sidecar":
				if e := os.Remove(filepath.Join(target, envmarker.Name)); e != nil {
					t.Fatal(e)
				}
			case "unknown-adapter":
				if e := os.Mkdir(filepath.Join(EnvRoot(fx.home), fx.profile, "unknown"), 0700); e != nil {
					t.Fatal(e)
				}
			case "broken-root", "absent-tail":
				root := fx.native["codex_cli"]
				if e := os.Rename(root, root+"-moved"); e != nil {
					t.Fatal(e)
				}
				if state == "broken-root" {
					if e := os.Symlink(root+"-missing", root); e != nil {
						t.Fatal(e)
					}
				}
			}
			before := nativeSnapshot(t, target)
			op, e := beginOperation(fx.home)
			if state != "healthy" && state != "absent-tail" {
				code := envorigin.DiagInvalid
				if state == "legacy" {
					code = envorigin.DiagRequired
				}
				if e == nil {
					_ = op.close()
					t.Fatal("common owner admitted invalid authority")
				}
				if !strings.Contains(e.Error(), code) {
					t.Fatalf("intended common refusal: %v", e)
				}
				if nativeSnapshot(t, target) != before {
					t.Fatal("common refusal changed home")
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			defer op.close()
			live := filepath.Join(ProfileDir(fx.home, fx.profile), "healthy-control")
			if e := op.publish(map[string][]byte{live: []byte("reached")}); e != nil {
				t.Fatal(e)
			}
			if raw, e := os.ReadFile(live); e != nil || string(raw) != "reached" {
				t.Fatal("healthy actual owner did not write")
			}
		})
	}
}

// The real common mutation owner must inspect the installed source as private
// authority, not count any existing directory entry as an installed profile.
// Healthy SetCurrent reaches its actual pointer write; wrong-kind/link/fork
// entries refuse before writes and preserve both the home and foreign bytes.
func TestProvisioningOriginInstalledSourceEntry(t *testing.T) {
	for _, state := range []string{"healthy", "symlink", "dangling", "directory", "hardlink"} {
		t.Run(state, func(t *testing.T) {
			fx, _, _ := recordedOriginFixture(t, "a")
			source := sourcePath(fx.home, fx.profile)
			original, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			foreign := filepath.Join(t.TempDir(), "source.json")
			if err := os.WriteFile(foreign, original, 0600); err != nil {
				t.Fatal(err)
			}
			if state != "healthy" {
				if err := os.Remove(source); err != nil {
					t.Fatal(err)
				}
				switch state {
				case "symlink":
					err = os.Symlink(foreign, source)
				case "dangling":
					err = os.Symlink(foreign+"-absent", source)
				case "directory":
					err = os.Mkdir(source, 0700)
				case "hardlink":
					err = os.Link(foreign, source)
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			target := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
			before := nativeSnapshot(t, target)
			pointer, pointerErr := os.ReadFile(CurrentFile(fx.home))
			err = SetCurrent(fx.home, fx.profile)
			if state == "healthy" {
				if err != nil {
					t.Fatalf("healthy actual pointer owner: %v", err)
				}
				selected, e := Current(fx.home)
				if e != nil || selected != fx.profile {
					t.Fatalf("pointer continuation not reached: %q %v", selected, e)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) {
					t.Fatalf("installed source entry must refuse specifically: %v", err)
				}
				after, afterErr := os.ReadFile(CurrentFile(fx.home))
				if string(after) != string(pointer) || (pointerErr == nil) != (afterErr == nil) {
					t.Fatal("refusal changed current pointer")
				}
			}
			if nativeSnapshot(t, target) != before {
				t.Fatal("source admission changed managed home")
			}
			if got, e := os.ReadFile(foreign); e != nil || string(got) != string(original) {
				t.Fatal("source admission changed foreign preimage")
			}
		})
	}
}
