package envprofile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
)

func TestPassthroughReadFailuresAreUnknownAndNeverRepaired(t *testing.T) {
	for _, failure := range []string{"lstat", "readlink"} {
		t.Run(failure, func(t *testing.T) {
			fx := writeManagedFixture(t, "acme")
			auth := filepath.Join(fx.native["codex_cli"], "auth.json")
			if err := os.WriteFile(auth, []byte("{\"openai_api_key\":null}\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
			managed := ManagedHomeDir(fx.home, "acme", "codex_cli")
			link := filepath.Join(managed, "auth.json")
			wantTarget, err := os.Readlink(link)
			if err != nil {
				t.Fatal(err)
			}
			markerPath := filepath.Join(managed, envmarker.Name)
			markerBefore, err := os.ReadFile(markerPath)
			if err != nil {
				t.Fatal(err)
			}

			lstat := func(path string) (os.FileInfo, error) { return os.Lstat(path) }
			readlink := func(path string) (string, error) { return os.Readlink(path) }
			if failure == "lstat" {
				lstat = func(path string) (os.FileInfo, error) {
					if path == link {
						return nil, &os.PathError{Op: "lstat", Path: path, Err: os.ErrPermission}
					}
					return os.Lstat(path)
				}
			} else {
				readlink = func(path string) (string, error) {
					if path == link {
						return "", errors.New("injected readlink I/O error")
					}
					return os.Readlink(path)
				}
			}

			statusReq := statusRequest(fx)
			statusReq.passthroughLstat = lstat
			statusReq.passthroughReadlink = readlink
			status, err := StatusOf(statusReq)
			if err != nil {
				t.Fatal(err)
			}
			row := findHome(status, "acme", "codex_cli")
			if row == nil || row.Current || !containsText(row.Findings, envregistry.DiagPassthroughUnreadable) {
				t.Fatalf("status row = %+v; want unknown, non-current %s", row, envregistry.DiagPassthroughUnreadable)
			}

			for _, repair := range []bool{false, true} {
				req := fx.request("codex_cli")
				req.Repair = repair
				req.passthroughLstat = lstat
				req.passthroughReadlink = readlink
				result, resolveErr := Resolve(req)
				if resolveErr == nil || !strings.Contains(resolveErr.Error(), envregistry.DiagPassthroughUnreadable) {
					t.Fatalf("Resolve(repair=%v) error = %v; want %s", repair, resolveErr, envregistry.DiagPassthroughUnreadable)
				}
				if !strings.HasPrefix(resolveErr.Error(), envregistry.DiagPassthroughUnreadable) {
					t.Fatalf("Resolve(repair=%v) collapsed unreadable passthrough into stale/detached: %v", repair, resolveErr)
				}
				if result != nil && len(result.Document) != 0 {
					t.Fatalf("Resolve(repair=%v) emitted a fragment after an unreadable passthrough", repair)
				}
			}
			if got, err := os.Readlink(link); err != nil || got != wantTarget {
				t.Fatalf("repair changed passthrough link to %q (%v), want %q", got, err, wantTarget)
			}
			markerAfter, err := os.ReadFile(markerPath)
			if err != nil || string(markerAfter) != string(markerBefore) {
				t.Fatalf("repair changed marker after read failure: %v", err)
			}
		})
	}
}

func TestUnreadableLockIsUntrustedAcrossStatusResolveRepairAndUpdate(t *testing.T) {
	for _, kind := range []string{"malformed", "directory", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			fx := writeManagedFixture(t, "acme")
			lock := lockPath(fx.home, "acme")
			switch kind {
			case "malformed":
				if err := os.WriteFile(lock, []byte("{\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			case "directory":
				if err := os.Remove(lock); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(lock, 0o700); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				outside := filepath.Join(t.TempDir(), "lock.json")
				if err := os.WriteFile(outside, []byte("{}\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(lock); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, lock); err != nil {
					t.Fatal(err)
				}
			}

			before, err := os.Lstat(lock)
			if err != nil {
				t.Fatal(err)
			}
			request := fx.request("codex_cli")
			if _, err := Resolve(request); err == nil || !strings.Contains(err.Error(), envregistry.DiagStoreUntrusted) {
				t.Fatalf("Resolve error = %v; want %s", err, envregistry.DiagStoreUntrusted)
			}
			request.Repair = true
			if _, err := Resolve(request); err == nil || !strings.Contains(err.Error(), envregistry.DiagStoreUntrusted) {
				t.Fatalf("Resolve --repair error = %v; want %s", err, envregistry.DiagStoreUntrusted)
			}
			status, err := StatusOf(statusRequest(fx))
			if err != nil {
				t.Fatal(err)
			}
			row := findHome(status, "acme", "codex_cli")
			if row == nil || row.Current || !containsText(row.Findings, envregistry.DiagStoreUntrusted) {
				t.Fatalf("status row = %+v; want non-current %s", row, envregistry.DiagStoreUntrusted)
			}
			if _, err := os.Lstat(filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), envmarker.Name)); !os.IsNotExist(err) {
				t.Fatalf("repair rebuilt managed home from unreadable lock: marker lstat error = %v", err)
			}
			after, err := os.Lstat(lock)
			if err != nil || before.Mode().Type() != after.Mode().Type() {
				t.Fatalf("lock entry changed after refused repair: before=%v after=%v err=%v", before.Mode(), after, err)
			}
		})
	}
}

func TestUnreadableLockRefusesUpdateWithoutRewriting(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	policy := Policy{}
	lock := lockPath(fx.home, "acme")
	broken := []byte("{\n")
	if err := os.WriteFile(lock, broken, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := UpdateWithOptions(fx.home, "acme", UpdateOptions{Policy: policy}); err == nil || !strings.Contains(err.Error(), envregistry.DiagStoreUntrusted) {
		t.Fatalf("Update error = %v; want %s", err, envregistry.DiagStoreUntrusted)
	}
	after, err := os.ReadFile(lock)
	if err != nil || string(after) != string(broken) {
		t.Fatalf("update rewrote unreadable lock: %q (%v)", after, err)
	}
	if _, err := os.Lstat(ProfileDir(fx.home, DefaultProfile)); !os.IsNotExist(err) {
		t.Fatalf("update wrote default profile before refusing unreadable lock: %v", err)
	}
}

func TestAbsentLockRemainsProfileUnknown(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	if err := os.Remove(lockPath(fx.home, "acme")); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(fx.request("codex_cli")); err == nil || !strings.Contains(err.Error(), DiagProfileUnknown) {
		t.Fatalf("Resolve with absent lock error = %v; want %s", err, DiagProfileUnknown)
	}
}

func TestUnreadableCodexSeedDoesNotSelectAbsentDefault(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	seed := filepath.Join(fx.native["codex_cli"], "config.toml")
	reads := 0
	req := fx.request("codex_cli")
	req.Repair = true
	req.readRegularFile = func(path string) (stateread.File, error) {
		if path == seed {
			reads++
			if reads == 1 {
				return stateread.File{Kind: stateread.KindUnreadable}, stateread.UnusableError(path, os.ErrPermission)
			}
			return stateread.File{Kind: stateread.KindAbsent}, nil
		}
		return stateread.ReadRegularFile(path)
	}
	if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), envregistry.DiagSeedUnreadable) {
		t.Fatalf("Resolve with unreadable Codex seed = %v; want %s", err, envregistry.DiagSeedUnreadable)
	}
	if reads != 1 {
		t.Fatalf("unreadable seed was retried and could take absence fallback: %d reads", reads)
	}
	if _, err := os.Lstat(filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), envmarker.Name)); !os.IsNotExist(err) {
		t.Fatalf("provisioning continued after unreadable seed: %v", err)
	}
}
