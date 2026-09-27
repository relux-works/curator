package envprofile

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
)

// TestUpdateSystemDeltaRefusesBeforePublication drives UpdateWithOptions,
// the profile update production entry, through revision-B refusal. The
// resolved delta is reported, but neither the old lock nor the candidate
// store entry is changed without the per-run confirmation flag.
func TestUpdateSystemDeltaRefusesBeforePublication(t *testing.T) {
	if _, err := gitLookPath(); err != nil {
		t.Skip("no git on PATH")
	}
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	manifest := `{"schema_version":1,"name":"delta","version":"1.0.0"}` + "\n"
	repo := gitRepo(t, map[string]string{"agent-context.json": manifest}, "v1.0.0")
	operand := ids.serve(repo, "https://example.com/delta")
	if _, _, _, err := Install(home, InstallOptions{Operand: operand}); err != nil {
		t.Fatalf("initial install: %v", err)
	}
	lockPath := lockPath(home, "delta")
	oldLock, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	writeGitFile(t, repo, "agent-context.json", `{"schema_version":1,"name":"delta","version":"1.0.1","context":{"modules":[{"path":"system.md","class":"system","environments":["claude_code"]}]}}`+"\n")
	writeGitFile(t, repo, "context/system.md", "system prompt\n")
	gitRun(t, repo, "add", ".")
	gitRun(t, repo, "commit", "-m", "system module")
	gitRun(t, repo, "tag", "v1.0.1")
	output, err := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	newCommit := strings.TrimSpace(string(output))
	info, moved, err := UpdateWithOptions(home, "delta", UpdateOptions{})
	if err == nil || !strings.Contains(err.Error(), DiagSystemDeltaConfirmationRequired) {
		t.Fatalf("unconfirmed UpdateWithOptions = moved:%t err:%v", moved, err)
	}
	if moved {
		t.Fatal("a refused resolved delta must not report a moved profile")
	}
	if len(info.Delta) != 1 || !strings.Contains(info.Delta[0], "lock-delta moved context delta") {
		t.Fatalf("refusal delta lines = %v", info.Delta)
	}
	currentLock, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(currentLock) != string(oldLock) {
		t.Fatal("confirmation refusal changed lock.json")
	}
	candidateEntry := contextstore.EntryDir(home, contextlock.KindContext, "delta", newCommit)
	if _, err := os.Lstat(candidateEntry); !os.IsNotExist(err) {
		t.Fatalf("confirmation refusal installed the candidate store entry: %v", err)
	}
}
