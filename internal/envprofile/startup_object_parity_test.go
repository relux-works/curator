package envprofile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

// This unchanged witness runs on main and the rework head: losing only a blob
// does not invalidate an intact store, including home-only repair. Losing the
// repository refuses as pin_hash, never as structural regular_types.
func TestStartupObjectMainParity(t *testing.T) {
	for _, loss := range []string{"blob", "cache"} {
		for level := 1; level <= 4; level++ {
			for _, mode := range []string{"observe", "repair", "home-repair"} {
				t.Run(fmt.Sprintf("%s/level%d/%s", loss, level, mode), func(t *testing.T) {
					fx := storeBoundaryFixture(t, "current", true)
					req := fx.request(envregistry.ClaudeCode)
					req.Repair = true
					if _, err := Resolve(req); err != nil {
						t.Fatal(err)
					}
					member, _ := storeBoundaryRootMemberFor(fx)
					repo := newGitManager(fx.home).repoDir(member.Source)
					home := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.ClaudeCode), "CLAUDE.md")
					want, err := os.ReadFile(home)
					if err != nil {
						t.Fatal(err)
					}
					if mode == "home-repair" {
						if err := os.WriteFile(home, []byte("home drift\n"), 0600); err != nil {
							t.Fatal(err)
						}
					}
					if loss == "cache" {
						if err := os.RemoveAll(repo); err != nil {
							t.Fatal(err)
						}
					} else {
						out, err := exec.Command("git", "-C", repo, "rev-parse", member.Commit+":context/s.md").Output()
						if err != nil {
							t.Fatal(err)
						}
						oid := strings.TrimSpace(string(out))
						path := filepath.Join(repo, ".git", "objects", oid[:2], oid[2:])
						if err := os.Chmod(path, 0600); err != nil {
							t.Fatal(err)
						}
						if err := os.Remove(path); err != nil {
							t.Fatal(err)
						}
					}
					// Main predates the level field: it exercises the same default path.
					if field := reflect.ValueOf(&req.Policy).Elem().FieldByName("StartupVerificationLevel"); field.IsValid() {
						field.SetInt(int64(level))
					}
					req.Repair = mode != "observe"
					r, err := Resolve(req)
					if loss == "blob" {
						if err != nil || r == nil || len(r.Document) == 0 {
							t.Fatalf("main-parity intact blob-loss launch: %v", err)
						}
						got, e := os.ReadFile(home)
						if e != nil || string(got) != string(want) {
							t.Fatalf("main-parity home restoration: %v", e)
						}
					} else {
						// Main's pin diagnostic is the contract under comparison; home-first
						// precedence is asserted separately by the full rework matrix.
						if err == nil || !strings.Contains(err.Error(), "pin_hash") || strings.Contains(err.Error(), "regular_types") || r != nil && len(r.Document) != 0 {
							t.Fatalf("main-parity missing cache pin refusal: %v", err)
						}
					}
				})
			}
		}
	}
	t.Run("broken-repository", startupBrokenRepositoryParity)
}

// Main has no level policy and repeats its default route under the five labels.
// The rework runs actual L0-L4. Both retain pin_hash and byte/identity receipts;
// the rework additionally gives home_stale precedence for home/combined drift.
func startupBrokenRepositoryParity(t *testing.T) {
	for _, name := range []string{"HEAD", "objects"} {
		for level := 0; level <= 4; level++ {
			t.Run(fmt.Sprintf("%s/level%d", name, level), func(t *testing.T) {
				t.Parallel()
				for _, drift := range []string{"none", "home", "source", "combined"} {
					for _, mode := range []string{"observe", "dry-run", "repair"} {
						t.Run(drift+"/"+mode, func(t *testing.T) {
							fx := storeBoundaryFixture(t, "current", true)
							req := fx.request(envregistry.ClaudeCode)
							req.Repair = true
							if _, err := Resolve(req); err != nil {
								t.Fatal(err)
							}
							field := reflect.ValueOf(&req.Policy).Elem().FieldByName("StartupVerificationLevel")
							current := field.IsValid()
							if current {
								field.SetInt(int64(level))
							}
							req.Repair = false
							if result, err := Resolve(req); err != nil || result == nil || len(result.Document) == 0 {
								t.Fatalf("main-parity intact repository neighbor: %v", err)
							}
							member, entry := storeBoundaryRootMemberFor(fx)
							repo := newGitManager(fx.home).repoDir(member.Source)
							owned := drift == "home" || drift == "combined"
							if owned {
								path := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.ClaudeCode), "CLAUDE.md")
								if err := os.WriteFile(path, []byte("home drift\n"), 0600); err != nil {
									t.Fatal(err)
								}
							}
							if drift == "source" || drift == "combined" {
								if err := os.WriteFile(filepath.Join(entry, "context", "s.md"), []byte("source drift\n"), 0600); err != nil {
									t.Fatal(err)
								}
							}
							if err := os.RemoveAll(filepath.Join(repo, ".git", name)); err != nil {
								t.Fatal(err)
							}
							before := hashTreeForTest(t, fx.home)
							nodes := startupParityIdentities(t, fx.home)
							req.Repair, req.DryRun = mode != "observe", mode == "dry-run"
							result, err := Resolve(req)
							if err == nil || !strings.Contains(err.Error(), "pin_hash") || strings.Contains(err.Error(), "regular_types") || result != nil && len(result.Document) != 0 {
								t.Fatalf("main-parity invalid repository pin refusal: %v", err)
							}
							if current && !owned && strings.Contains(err.Error(), DiagHomeStale) {
								t.Fatalf("main-parity repository loss borrowed home precedence: %v", err)
							}
							if current && owned && (!strings.HasPrefix(err.Error(), "environment_home_stale:") || !strings.Contains(err.Error(), "--repair")) {
								t.Fatalf("main-parity invalid repository hid home diagnostic: %v", err)
							}
							after := startupParityIdentities(t, fx.home)
							if hashTreeForTest(t, fx.home) != before || len(nodes) != len(after) {
								t.Fatal("main-parity invalid repository wrote state")
							}
							for path, original := range nodes {
								if now, ok := after[path]; !ok || !os.SameFile(original, now) || original.Mode() != now.Mode() || original.ModTime() != now.ModTime() {
									t.Fatalf("main-parity invalid repository replaced state %q", path)
								}
							}
						})
					}
				}
			})
		}
	}
}

func startupParityIdentities(t *testing.T, root string) map[string]os.FileInfo {
	t.Helper()
	result := map[string]os.FileInfo{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() || info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			info, err = file.Stat()
			closeErr := file.Close()
			if err != nil {
				return err
			}
			if closeErr != nil {
				return closeErr
			}
		}
		result[path] = info
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
