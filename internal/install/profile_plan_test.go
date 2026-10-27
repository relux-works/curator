package install

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/godriver"
)

// The ordinary Go session re-execs its manager image. Give the fixture image
// the same fixed worker dispatch as cmd/curator without replacing the compiler.
func init() {
	if len(os.Args) == 2 && os.Args[1] == godriver.WorkerMode {
		os.Exit(godriver.RunWorker(os.Stdin, os.Stdout))
	}
}

// PSR05 preparation subset: a real signed package passes strict audit and the
// ordinary private Go compiler/receipt path before record publication. Invalid
// Go and post-receipt artifact replacement each refuse with the old lock and
// routing intact. This proves preparation only, not command/PATH adoption.
func TestProfileSkillPlanRealCompileBeforeLock(t *testing.T) {
	for _, mode := range []string{"success", "compiler-failure", "artifact-tamper"} {
		t.Run(mode, func(t *testing.T) {
			e := newEnv(t)
			e.buildSkill("compiler", "profile-tool")
			repo := filepath.Join(e.skillsRoot, "compiler")
			key := filepath.Join(t.TempDir(), "signer")
			if out, err := exec.Command("ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", key).CombinedOutput(); err != nil {
				t.Fatalf("fixture signer: %v %s", err, out)
			}
			if mode == "compiler-failure" {
				e.write(repo, "src/cmd/profile-tool/main.go", "package main\nfunc main() { broken syntax }\n")
			}
			e.git(repo, "add", ".")
			e.git(repo, "-c", "gpg.format=ssh", "-c", "gpg.ssh.program=ssh-keygen", "-c", "user.signingkey="+key, "commit", "--amend", "-S", "--no-edit")
			pin, err := gitops.Resolve(repo, "revision", "HEAD")
			if err != nil {
				t.Fatal(err)
			}
			pub, err := os.ReadFile(key + ".pub")
			if err != nil {
				t.Fatal(err)
			}
			gitConfig := filepath.Join(t.TempDir(), "gitconfig")
			e.write(filepath.Dir(gitConfig), filepath.Base(gitConfig), fmt.Sprintf("[url %q]\n insteadOf = https://example.com/compiler\n", draftReplayFileURL(repo)))
			t.Setenv("GIT_CONFIG_GLOBAL", gitConfig)
			t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
			root := t.TempDir()
			e.write(root, "agent-context.json", fmt.Sprintf(`{"schema_version":1,"name":"profile","version":"1.0.0","requires":{"skills":{"compiler":{"git":"https://example.com/compiler","revision":%q}}}}`, pin.Commit))
			e.write(root, "context/main.md", "# Profile\n")
			_, state, err := contextstore.EnsureState(e.home, "context", "profile", root)
			if err != nil {
				t.Fatal(err)
			}
			lockPath := filepath.Join(e.home, "profiles/profile/lock.json")
			hash, err := contextlock.Write(lockPath, &contextlock.Lock{SchemaVersion: 2, HashVersion: 2, Root: "profile", Members: []contextlock.Member{{Kind: "context", Name: "profile", Version: "1.0.0", StateHash: state, RequiredBy: []string{}}}})
			if err != nil {
				t.Fatal(err)
			}
			e.write(e.home, "profiles/current", "other\n")
			global := "{\"schema_version\":1,\"skills\":[]}\n"
			e.write(e.home, "global/Skillfile.json", global)
			old, _ := os.ReadFile(lockPath)
			called := false
			result, err := PrepareProfileSkillPlan(e.cfg, "profile", hash, envprofile.Policy{RequireSourceSigners: true, SourceSigners: map[string][]contextresolve.Signer{"example.com/compiler": {{Type: "ssh", Key: strings.TrimSpace(string(pub))}}}}, Options{OnStaged: func(staged Staged) error {
				called = true
				during, _ := os.ReadFile(lockPath)
				if !bytes.Equal(old, during) {
					t.Fatal("lock published before receipt observation")
				}
				builds := staged.Builds()
				if len(builds) != 1 || builds[0].Artifact().Size == 0 || builds[0].Receipt().Validate() != nil || builds[0].Receipt().Artifact != builds[0].Artifact() {
					t.Fatal("missing real compiled receipt")
				}
				if mode == "artifact-tamper" {
					return os.WriteFile(builds[0].Path(), []byte("tampered"), 0755)
				}
				return nil
			}})
			after, _ := os.ReadFile(lockPath)
			if mode == "success" {
				if err != nil {
					t.Fatal(err)
				}
				if !called || bytes.Equal(old, after) || result.Lock.SchemaVersion != 5 {
					t.Fatal("compiled plan not published")
				}
			} else {
				want := "profile_skill_compile_failed"
				if mode == "artifact-tamper" {
					want = "profile_skill_plan_binding_invalid"
				}
				if err == nil || !strings.Contains(err.Error(), want) || (mode == "compiler-failure" && !strings.Contains(err.Error(), "go_build_failed")) {
					t.Fatalf("intended compilation refusal: %v", err)
				}
				if !bytes.Equal(old, after) {
					t.Fatal("failed compile published lock")
				}
				if (mode == "artifact-tamper") != called {
					t.Fatal("failure did not reach intended boundary")
				}
			}
			current, _ := os.ReadFile(filepath.Join(e.home, "profiles/current"))
			declarations, _ := os.ReadFile(filepath.Join(e.home, "global/Skillfile.json"))
			if string(current) != "other\n" || string(declarations) != global {
				t.Fatal("preparation changed routing/declarations")
			}
			if _, err := os.Stat(filepath.Join(e.home, "global/bin")); !os.IsNotExist(err) {
				t.Fatal("published commands")
			}
		})
	}
}
