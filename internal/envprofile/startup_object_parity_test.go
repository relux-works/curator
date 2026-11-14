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
}
