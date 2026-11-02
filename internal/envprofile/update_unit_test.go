package envprofile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/transaction"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Actual UpdateWithOptions enlists current and scoped native publication with
// inactive managed homes and lock/previous-lock. A late fault and unmanaged
// native conflict preserve exact L0. Repeated valid updates preserve capture.
// Raw account evidence is fixture-only; no account/provider process runs.
func TestUpdateWholeInvocationUnit(t *testing.T) {
	for _, state := range []string{"current", "scoped", "noncurrent", "native-conflict", "late-fault"} {
		t.Run(state, func(t *testing.T) {
			enableV1WritersForTest(t)
			pinOperatorHome(t, t.TempDir())
			homes := pinHomes(t)
			home := gitFixtureHome(t)
			ids := newGitIdentities(t)
			skill := gitRepo(t, map[string]string{"SKILL.md": "# myskill one\n"}, "v1.0.0")
			operand := ids.serve(skill, "https://example.com/update-unit-skill")
			root := gitRepo(t, map[string]string{"agent-context.json": `{"schema_version":1,"name":"update-unit","version":"1.0.0","context":{"modules":[{"path":"a.md"}]},"requires":{"skills":{"myskill":{"git":"` + operand + `","range":"*"}}}}`, "context/a.md": "hello\n"}, "v1.0.0")
			if _, _, _, err := Install(home, InstallOptions{Operand: ids.serve(root, "https://example.com/update-unit"), Use: true}); err != nil {
				t.Fatal(err)
			}
			fx := &managedFixture{home: home, profile: "update-unit", native: homes, xdg: filepath.Dir(homes["opencode"]), launch: t.TempDir(), accountAnchor: t.TempDir()}
			_, fx.lockHash, _ = readLock(home, fx.profile)
			for _, id := range []string{"codex_cli", "claude_code"} {
				req := fx.request(id)
				req.Repair = true
				if _, err := Resolve(req); err != nil {
					t.Fatal(err)
				}
			}
			if state == "scoped" || state == "late-fault" {
				if err := SetScoped(home, "env:codex_cli", fx.profile, false); err != nil {
					t.Fatal(err)
				}
				scoped, err := ScopedCurrents(home)
				if err != nil || scoped["env:codex_cli"] != fx.profile {
					t.Fatalf("scoped continuation fixture missing: %v %v", scoped, err)
				}
			}
			if state == "noncurrent" {
				if err := SetCurrent(home, DefaultProfile); err != nil {
					t.Fatal(err)
				}
			}
			if state == "native-conflict" {
				if err := os.Remove(filepath.Join(homes["codex_cli"], envmarker.Name)); err != nil {
					t.Fatal(err)
				}
				target := filepath.Join(homes["codex_cli"], "skills", "myskill")
				if err := os.Remove(target); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(target, []byte("operator entry"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			originPath := filepath.Join(ManagedHomeDir(home, fx.profile, "codex_cli"), envorigin.Name)
			initial := nativeRead(t, originPath)
			var capture envorigin.Record
			if err := json.Unmarshal(initial, &capture); err != nil {
				t.Fatal(err)
			}
			// Exclude immutable newly installed store objects and normal lock-file mtime.
			snapshot := func() string {
				var b strings.Builder
				for _, path := range []string{ProfileDir(home, fx.profile), ManagedHomeDir(home, fx.profile, "codex_cli"), ManagedHomeDir(home, fx.profile, "claude_code"), filepath.Dir(CurrentFile(home)), homes["claude_code"], homes["codex_cli"], homes["pi"], homes["opencode"]} {
					b.WriteString(nativeSnapshot(t, path))
				}
				return b.String()
			}
			for revision := 1; revision <= 2; revision++ {
				writeGitFile(t, skill, "SKILL.md", fmt.Sprintf("# myskill revision %d\n", revision))
				gitRun(t, skill, "add", ".")
				gitRun(t, skill, "commit", "-m", "skill revision")
				gitRun(t, skill, "tag", fmt.Sprintf("v1.0.%d", revision))
				before := snapshot()
				oldLock := nativeRead(t, lockPath(home, fx.profile))
				hit := false
				earlierNative := false
				saved := newOperationEngine
				if state == "late-fault" {
					newOperationEngine = func(h string, opts ...transaction.Option) (*transaction.Engine, error) {
						return saved(h, append(opts, transaction.WithHooks(transaction.Hooks{Fault: func(e transaction.Event) error {
							if e.Point == transaction.PointAfterInstall && e.LivePath == filepath.Join(homes["claude_code"], "skills", "myskill") {
								earlierNative = true
							}
							if e.Point == transaction.PointAfterInstall && e.LivePath == filepath.Join(homes["codex_cli"], "skills", "myskill") {
								hit = true
								return errors.New("late update owner fault")
							}
							return nil
						}}))...)
					}
				}
				_, moved, err := UpdateWithOptions(home, fx.profile, UpdateOptions{})
				newOperationEngine = saved
				if state == "native-conflict" || state == "late-fault" {
					want := DiagUsePartial
					if state == "late-fault" {
						want = "late update owner fault"
						if !hit || !earlierNative {
							t.Fatalf("late native stage not reached: %v", err)
						}
					}
					if err == nil || !strings.Contains(err.Error(), want) || moved {
						t.Fatalf("owner refusal: moved=%t err=%v want=%s", moved, err, want)
					}
					if got := snapshot(); got != before {
						t.Fatal("update refusal changed L0 lock/home/native/selection preimages")
					}
					return
				}
				if err != nil || !moved {
					t.Fatalf("actual update: moved=%t %v", moved, err)
				}
				newLock := nativeRead(t, lockPath(home, fx.profile))
				if bytes.Equal(newLock, oldLock) {
					t.Fatal("update did not reach changed lock publication")
				}
				var after envorigin.Record
				if err := json.Unmarshal(nativeRead(t, originPath), &after); err != nil {
					t.Fatal(err)
				}
				if after.Capture != capture.Capture || after.NativeRoot != capture.NativeRoot || after.Binding.LockSHA256 == capture.Binding.LockSHA256 {
					t.Fatal("lock transition recaptured or did not rebind origin")
				}
				if state != "noncurrent" {
					for _, id := range []string{"codex_cli", "claude_code"} {
						target, err := filepath.EvalSymlinks(filepath.Join(homes[id], "skills", "myskill", "SKILL.md"))
						if err != nil {
							t.Fatal(err)
						}
						if got := string(nativeRead(t, target)); !strings.Contains(got, fmt.Sprintf("revision %d", revision)) {
							t.Fatal("native continuation retained stale skill")
						}
					}
				}
			}
		})
	}
}
