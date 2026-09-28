package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/globalbins"
	"github.com/relux-works/curator/internal/runtimestore"
)

func TestGlobalInstallRecognizesAdoptedForwardingShim(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-g")
	if _, err := GlobalInit(e.home); err != nil {
		t.Fatal(err)
	}
	if err := manifestAddGlobal(e, "skill-g"); err != nil {
		t.Fatal(err)
	}

	userHome := t.TempDir()
	userBin := filepath.Join(userHome, ".local", "bin")
	inheritedPath := os.Getenv("PATH")
	t.Setenv(globalbins.UserBinEnv, "")
	t.Setenv("PATH", inheritedPath)
	initial := Global(e.cfg, userHome, Options{Platform: installPlatform()})
	if initial.Status != "ok" {
		t.Fatalf("initial global install without a selected user bin: %+v", initial)
	}

	if err := os.MkdirAll(userBin, 0o755); err != nil {
		t.Fatal(err)
	}
	command := "skill-g-tool"
	canonical := filepath.Join(GlobalRoot(e.home), "bin", shimName(command))
	published := filepath.Join(userBin, shimName(command))
	content := runtimestore.UnixShimContent(canonical, nil)
	if installPlatform() == "windows" {
		content = runtimestore.WindowsShimContent(canonical, nil)
	}
	if err := os.WriteFile(published, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(globalbins.UserBinEnv, userBin)
	t.Setenv("PATH", userBin+string(os.PathListSeparator)+inheritedPath)

	adopted, err := globalbins.Adopt(e.home, command, installPlatform(), nil, userHome, false)
	if err != nil || adopted.Path != published || adopted.Backup == "" {
		t.Fatalf("globalbins.Adopt() = (%+v, %v)", adopted, err)
	}

	result := Global(e.cfg, userHome, Options{Platform: installPlatform()})
	if result.Status != "ok" {
		t.Fatalf("global install after adoption: %+v", result)
	}
	for _, message := range result.Messages {
		if strings.Contains(message, "not managed by Curator") {
			t.Fatalf("global install rejected adopted shim: %s", message)
		}
	}
	if _, err := os.Stat(published); err != nil {
		t.Fatalf("global install did not publish the adopted command: %v", err)
	}
	ledgerPayload, err := os.ReadFile(filepath.Join(userBin, ".curator-managed.json"))
	if err != nil {
		t.Fatal(err)
	}
	var ownership struct {
		Entries []string `json:"entries"`
	}
	if err := json.Unmarshal(ledgerPayload, &ownership); err != nil || len(ownership.Entries) != 1 || ownership.Entries[0] != command {
		t.Fatalf("global install ownership after adoption = (%+v, %v), want %q managed", ownership, err, command)
	}
}
