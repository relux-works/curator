package install

import (
	"fmt"
	"github.com/relux-works/curator/internal/globalbins"
	"github.com/relux-works/curator/internal/transaction"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWave2ConcurrentProjectInstalls(t *testing.T) {
	e := newEnv(t)
	e.skill("wave-skill")
	e.declare("wave-skill")
	other := &env{t: t, skillsRoot: e.skillsRoot, home: e.home, project: t.TempDir(), cfg: e.cfg}
	other.git(other.project, "init", "-q")
	other.declare("wave-skill")
	ready := make(chan struct{}, 2)
	release := make(chan struct{})
	results := make(chan Result, 2)
	for _, project := range []string{e.project, other.project} {
		go func(project string) {
			results <- Project(e.cfg, project, "wave", Options{Platform: installPlatform(), OnStaged: func(_ Staged) error {
				ready <- struct{}{}
				select {
				case <-release:
					return nil
				case <-time.After(30 * time.Second):
					return fmt.Errorf("barrier timed out")
				}
			}})
		}(project)
	}
	for i := 0; i < 2; i++ {
		select {
		case <-ready:
		case <-time.After(30 * time.Second):
			close(release)
			t.Fatal("both installs did not reach staging")
		}
	}
	close(release)
	for i := 0; i < 2; i++ {
		r := <-results
		t.Logf("Project status=%s errors=%v", r.Status, r.Errors)
		if r.Status != "ok" {
			t.Fatalf("concurrent install: %+v", r)
		}
	}
	for _, project := range []string{e.project, other.project} {
		got, err := exec.Command(filepath.Join(project, ".agents", "bin", shimName("wave-skill-tool"))).CombinedOutput()
		t.Logf("published shim child exit=%v", err)
		if err != nil || strings.TrimSpace(string(got)) != "wave-skill" {
			t.Fatalf("inconsistent published command: %q %v", got, err)
		}
	}
	assertNoJournalRemains(t, e.home)
}
func TestWave2GlobalOwnership(t *testing.T) {
	for _, shape := range []string{"owned-regular", "foreign-regular", "owned-legacy-link", "foreign-after-prepare"} {
		t.Run(shape, func(t *testing.T) {
			e := newEnv(t)
			e.skill("wave-skill")
			e.globalDeclareAll("wave-skill")
			user := t.TempDir()
			bin := filepath.Join(user, "bin")
			if err := os.Mkdir(bin, 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv(globalbins.UserBinEnv, bin)
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			opts := Options{Platform: installPlatform()}
			first := Global(e.cfg, user, opts)
			if first.Status != "ok" {
				t.Fatalf("initial global install: %+v", first)
			}
			name := "wave-skill-tool"
			published := filepath.Join(bin, shimName(name))
			switch shape {
			case "foreign-regular":
				if err := os.WriteFile(published, []byte("foreign bytes\n"), 0700); err != nil {
					t.Fatal(err)
				}
			case "owned-legacy-link":
				messages := globalbins.Refresh(e.home, map[string]bool{name: true}, installPlatform(), nil, user)
				info, err := os.Lstat(published)
				if err != nil || info.Mode()&os.ModeSymlink == 0 {
					t.Fatalf("Refresh did not make legacy link: %v %v", err, messages)
				}
			case "foreign-after-prepare":
				opts.Commit.Hooks = transaction.Hooks{Fault: func(event transaction.Event) error {
					if event.Point == transaction.PointPrepared {
						return os.WriteFile(published, []byte("foreign bytes\n"), 0700)
					}
					return nil
				}}
			}
			result := Global(e.cfg, user, opts)
			t.Logf("Global status=%s errors=%v", result.Status, result.Errors)
			if strings.HasPrefix(shape, "foreign") {
				got, err := os.ReadFile(published)
				if err != nil || string(got) != "foreign bytes\n" {
					t.Fatalf("foreign shim changed: %q %v", got, err)
				}
				if shape == "foreign-after-prepare" && result.Status == "ok" {
					t.Fatal("preimage drift must fail")
				}
			} else if result.Status != "ok" {
				t.Fatalf("manager-owned forwarding shim cannot be reconciled: %+v", result)
			}
		})
	}
}
