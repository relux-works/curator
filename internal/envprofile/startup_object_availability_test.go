package envprofile

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

func removeStartupGitObject(t *testing.T, repo, oid string) {
	t.Helper()
	path := filepath.Join(repo, ".git", "objects", oid[:2], oid[2:])
	if err := os.Chmod(path, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}

func startupBlobID(t *testing.T, repo, commit string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", repo, "rev-parse", commit+":context/s.md").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

// Local Git-object loss is pin availability, never structural regular_types.
// Fresh fixtures cover every level/mode/drift combination; intact neighbors
// prove launch, and receipts plus bytes prove refusal does not mutate state.
func TestStartupObjectAvailabilityMatrix(t *testing.T) {
	for _, loss := range []string{"blob", "commit", "cache"} {
		t.Run(loss, func(t *testing.T) {
			for level := 1; level <= 4; level++ {
				t.Run(fmt.Sprintf("level%d", level), func(t *testing.T) {
					// Each level owns every fixture path and changes no process environment.
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
								member, entry := storeBoundaryRootMemberFor(fx)
								repo := newGitManager(fx.home).repoDir(member.Source)
								home := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.ClaudeCode), "CLAUDE.md")
								want, err := os.ReadFile(home)
								if err != nil {
									t.Fatal(err)
								}
								req.Policy.StartupVerificationLevel = level
								req.Repair = false
								if r, err := Resolve(req); err != nil || r == nil || len(r.Document) == 0 {
									t.Fatalf("intact object neighbor: %v", err)
								}
								if drift == "home" || drift == "combined" {
									if err := os.WriteFile(home, []byte("home drift\n"), 0600); err != nil {
										t.Fatal(err)
									}
								}
								if drift == "source" || drift == "combined" {
									if err := os.WriteFile(filepath.Join(entry, "context", "s.md"), []byte("source drift\n"), 0600); err != nil {
										t.Fatal(err)
									}
								}
								switch loss {
								case "blob":
									removeStartupGitObject(t, repo, startupBlobID(t, repo, member.Commit))
								case "commit":
									removeStartupGitObject(t, repo, member.Commit)
								case "cache":
									if err := os.RemoveAll(repo); err != nil {
										t.Fatal(err)
									}
								}
								req.Repair = mode != "observe"
								req.DryRun = mode == "dry-run"
								before := hashTreeForTest(t, fx.home)
								nodes := storeVerificationNodes(t, fx.home)
								r, err := Resolve(req)
								owned := drift == "home" || drift == "combined"
								// Tree metadata authenticates the intact store without reading blobs.
								launch := loss == "blob" && (drift == "none" || drift == "home" && (mode == "repair" || mode == "observe" && level == 1) || drift == "source" && level == 1 || drift == "combined" && mode == "observe" && level == 1)
								if launch {
									if err != nil || r == nil || len(r.Document) == 0 {
										t.Fatalf("available store refused after %s loss: %v", loss, err)
									}
									if owned && mode == "observe" && (len(r.Warnings) == 0 || !strings.HasPrefix(r.Warnings[0], DiagHomeStale+":") || !strings.Contains(r.Warnings[0], "--repair")) {
										t.Fatalf("home warning is not first: %v", r.Warnings)
									}
								} else {
									if err == nil || r != nil && len(r.Document) != 0 {
										t.Fatalf("unverified inputs launched: %v", err)
									}
									if strings.Contains(err.Error(), "regular_types") {
										t.Fatalf("object availability became structure: %v", err)
									}
									if owned {
										hintRequired := loss != "blob" || drift != "home" || mode != "dry-run"
										if !strings.HasPrefix(err.Error(), DiagHomeStale+":") || hintRequired && !strings.Contains(err.Error(), "--repair") {
											t.Fatalf("object loss hid home diagnostic: %v", err)
										}
									} else if loss != "blob" && (!strings.Contains(err.Error(), "pin_hash") || !strings.Contains(err.Error(), "pinned commit object unavailable")) {
										t.Fatalf("object loss lost pin class: %v", err)
									}
								}
								restored := launch && owned && mode == "repair"
								if restored {
									got, e := os.ReadFile(home)
									if e != nil || !bytes.Equal(got, want) {
										t.Fatalf("home-only blob-loss repair did not restore: %v", e)
									}
								} else if hashTreeForTest(t, fx.home) != before || !sameStoreVerificationNodes(nodes, storeVerificationNodes(t, fx.home)) {
									t.Fatal("object-loss observation/refusal wrote state")
								}
							})
						}
					}
				})
			}
		})
	}
}

// Object disappearance after structural inspection reaches pin revalidation,
// independently of the earlier missing-commit guard. The intact-home neighbor
// retains pin_hash; a drifted home leads with home_stale and never launches.
func TestStartupObjectLossBetweenInspectionAndPins(t *testing.T) {
	for _, drift := range []bool{false, true} {
		t.Run(fmt.Sprint(drift), func(t *testing.T) {
			fx := storeBoundaryFixture(t, "current", true)
			req := fx.request(envregistry.ClaudeCode)
			req.Repair = true
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			member, _ := storeBoundaryRootMemberFor(fx)
			repo := newGitManager(fx.home).repoDir(member.Source)
			if drift {
				if err := os.WriteFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.ClaudeCode), "CLAUDE.md"), []byte("home drift\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			called := false
			var before string
			var nodes map[string]storeVerificationNode
			req.afterStoreInspection = func() {
				called = true
				removeStartupGitObject(t, repo, member.Commit)
				before = hashTreeForTest(t, fx.home)
				nodes = storeVerificationNodes(t, fx.home)
			}
			r, err := Resolve(req)
			if !called {
				t.Fatal("object race seam not reached")
			}
			prefix := envregistry.DiagStoreUntrusted + ":"
			if drift {
				prefix = DiagHomeStale + ":"
			}
			if err == nil || !strings.HasPrefix(err.Error(), prefix) || r != nil && len(r.Document) != 0 {
				t.Fatalf("late object loss lost leading diagnostic: %v", err)
			}
			if hashTreeForTest(t, fx.home) != before || !sameStoreVerificationNodes(nodes, storeVerificationNodes(t, fx.home)) {
				t.Fatal("late object loss wrote state")
			}
		})
	}
}

// Narrowing each new gate must fail at its intended assertion, not setup.
func TestStartupObjectAvailabilityControls(t *testing.T) {
	for _, p := range []struct{ name, file, before, after, test, witness string }{
		{"blob-independent-roster", "../gitops/snapshot_paths.go", "sized.Write(meta)", "if _, err := runWithEnv(repo, env, \"cat-file\", \"-s\", strings.Fields(string(meta))[2]); err != nil { return nil, objectUnavailable(\"blob\", err) }; sized.Write(meta)", "TestStartupObjectAvailabilityMatrix/blob/level1/none/observe", "available store refused after blob loss"},
		{"late-home-wrap", "managed.go", "if pinnedCommitObjectUnavailable(failure) {\n\t\t\treturn nil, homeInputFailure(failure)\n\t\t}", "if pinnedCommitObjectUnavailable(failure) {\n\t\t\treturn nil, failure\n\t\t}", "TestStartupObjectLossBetweenInspectionAndPins/true", "late object loss lost leading diagnostic"},
		{"early-home-wrap", "managed.go", "return nil, homeInputsFailure(owned, err)", "_ = owned; return nil, err", "TestStartupObjectAvailabilityMatrix/cache/level1/home/observe", "object loss hid home diagnostic"},
		{"typed-pin-class", "store_availability.go", "if errors.As(err, &unavailable) {", "if false && errors.As(err, &unavailable) {", "TestStartupObjectAvailabilityMatrix/cache/level1/none/observe", "object availability became structure"},
	} {
		t.Run(p.name, func(t *testing.T) { runObservationControl(t, p.file, p.before, p.after, p.test, p.witness) })
	}
}
