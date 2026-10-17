package envprofile

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/testseed"
)

// The same native bytes produce the same fragment through a regular file, a
// relative symlink, and a chain; verification and repair never alter the target.
func TestNativeSeedSymlinkResolve(t *testing.T) {
	fx := writeManagedFixture(t, "native-seed")
	seedLiveNativeCredentials(t, fx)
	config := filepath.Join(fx.native[envregistry.CodexCLI], "config.toml")
	payload := []byte("model = 'fixture'\ncli_auth_credentials_store = 'file'\n")
	if err := testseed.WriteFile(config, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	req := fx.request(envregistry.CodexCLI)
	req.Repair = true
	initial, err := Resolve(req)
	if err != nil {
		t.Fatalf("regular seed: %v", err)
	}
	managed := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI), "config.toml")
	wantSeed, err := os.ReadFile(managed)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(config); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(filepath.Dir(config), "seed.toml")
	if err := testseed.WriteFile(target, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	stamp := time.Unix(1600000000, 123000000)
	if err := os.Chtimes(target, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	for _, chain := range []bool{false, true} {
		name := "symlink"
		linkTarget := "seed.toml"
		if chain {
			name = "chain"
			if err := os.Remove(config); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink("seed.toml", filepath.Join(filepath.Dir(config), "middle.toml")); err != nil {
				t.Fatal(err)
			}
			linkTarget = "middle.toml"
		}
		if err := os.Symlink(linkTarget, config); err != nil {
			t.Fatal(err)
		}
		for _, repair := range []bool{false, true} {
			t.Run(name+map[bool]string{false: "/resolve", true: "/repair"}[repair], func(t *testing.T) {
				req.Repair = repair
				if repair {
					if err := os.Remove(filepath.Join(filepath.Dir(managed), "AGENTS.md")); err != nil && !os.IsNotExist(err) {
						t.Fatal(err)
					}
				}
				result, err := Resolve(req)
				if err != nil || result == nil {
					t.Fatalf("native linked seed refused: %v", err)
				}
				if !bytes.Equal(result.Document, initial.Document) {
					t.Fatal("same native bytes changed resolved fragment")
				}
				got, err := os.ReadFile(managed)
				if err != nil || !bytes.Equal(got, wantSeed) {
					t.Fatalf("managed seed changed: %v", err)
				}
				got, err = os.ReadFile(target)
				after, statErr := os.Stat(target)
				if err != nil || statErr != nil || !bytes.Equal(got, payload) || !after.ModTime().Equal(before.ModTime()) {
					t.Fatal("resolve/repair modified native target bytes or mtime")
				}
				if got, err := os.Readlink(config); err != nil || got != linkTarget {
					t.Fatalf("native link changed: %v", err)
				}
			})
		}
	}
}

// Manager-owned lock symlinks remain a hard refusal, without provisioning or
// writes to the target, even when repair was requested.
func TestNativeSeedDoesNotRelaxManagerState(t *testing.T) {
	fx := writeManagedFixture(t, "native-manager")
	path := lockPath(fx.home, fx.profile)
	target := path + ".regular"
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(path, target); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	state, err := stateread.ReadRegularFile(path)
	if err == nil || state.Kind != stateread.KindUnreadable || !strings.Contains(err.Error(), "path must be a regular file") {
		t.Fatalf("strict manager reader: %+v %v", state, err)
	}
	for _, repair := range []bool{false, true} {
		req := fx.request(envregistry.CodexCLI)
		req.Repair = repair
		result, err := Resolve(req)
		if err == nil || result != nil || !strings.Contains(err.Error(), envregistry.DiagStoreUntrusted) {
			t.Fatalf("manager symlink resolve: %+v %v", result, err)
		}
		if _, err := os.Lstat(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI)); !os.IsNotExist(err) {
			t.Fatalf("refusal provisioned home: %v", err)
		}
		got, err := os.ReadFile(target)
		if err != nil || !bytes.Equal(got, original) {
			t.Fatal("refusal changed manager target")
		}
	}
}
