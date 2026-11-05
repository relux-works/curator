package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/staging"
)

// A colliding foreign file refuses the complete install and preserves its
// bytes and the previous marker/lock. Absent entries are a positive control.
func TestProjectBinForeignCollision(t *testing.T) {
	for _, installed := range []bool{false, true} {
		t.Run(fmt.Sprint(installed), func(t *testing.T) {
			e := newEnv(t)
			e.skill("tool")
			e.declare("tool")
			shim := filepath.Join(e.project, ".agents", "bin", shimName("tool-tool"))
			markerPath := filepath.Join(e.project, ".agents", "skills", "tool", ".csk-install.json")
			if installed {
				if r := e.install(Options{}); r.Status != "ok" {
					t.Fatalf("positive install: %+v", r)
				}
			}
			previousMarker, _ := os.ReadFile(markerPath)
			lockPath := filepath.Join(e.project, "Skillfile.lock")
			previousLock, _ := os.ReadFile(lockPath)
			if err := os.MkdirAll(filepath.Dir(shim), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(shim); err != nil && !os.IsNotExist(err) {
				t.Fatal(err)
			}
			foreign := []byte("foreign operator command\n")
			if err := os.WriteFile(shim, foreign, 0o755); err != nil {
				t.Fatal(err)
			}
			r := e.install(Options{})
			text := fmt.Sprint(r)
			if r.Status == "ok" || !strings.Contains(text, projectBinConflict) || !strings.Contains(text, shim) {
				t.Fatalf("expected named collision: %+v", r)
			}
			for path, want := range map[string][]byte{shim: foreign, markerPath: previousMarker, lockPath: previousLock} {
				got, _ := os.ReadFile(path)
				if string(got) != string(want) {
					t.Fatalf("refusal changed %s", path)
				}
			}
		})
	}
}

// An unrelated foreign entry survives install, reinstall and skill removal,
// with a named report. A live shim swapped by the user also survives removal.
func TestProjectBinPreservesForeignEntriesAndRemoval(t *testing.T) {
	e := newEnv(t)
	e.skill("tool")
	e.declare("tool")
	bin := filepath.Join(e.project, ".agents", "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	unrelated := filepath.Join(bin, "foreign-file")
	const bytes = "foreign content\n"
	if err := os.WriteFile(unrelated, []byte(bytes), 0o644); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		r := e.install(Options{})
		if r.Status != "ok" || !strings.Contains(fmt.Sprint(r.Messages), projectBinPreserved) || !strings.Contains(fmt.Sprint(r.Messages), unrelated) {
			t.Fatalf("install: %+v", r)
		}
		got, _ := os.ReadFile(unrelated)
		if string(got) != bytes {
			t.Fatal("foreign entry changed")
		}
	}
	shim := filepath.Join(bin, shimName("tool-tool"))
	if err := os.WriteFile(shim, []byte(bytes), 0o755); err != nil {
		t.Fatal(err)
	}
	e.declare()
	r := e.install(Options{})
	if r.Status != "ok" {
		t.Fatalf("remove: %+v", r)
	}
	for _, path := range []string{shim, unrelated} {
		got, _ := os.ReadFile(path)
		if string(got) != bytes {
			t.Fatalf("removal changed %s", path)
		}
	}
	if !strings.Contains(fmt.Sprint(r.Messages), shim) {
		t.Fatalf("swapped shim was not reported: %+v", r)
	}
}

// A prior installed marker and an exact legacy script/.cmd or relative link
// permit upgrade into a fingerprint ledger and later safe removal. Windows
// includes the original, lifecycle and single-pass-call historical bytes. Merely
// similar script text fails the exact renderer check in the adjacent tests.
func TestProjectBinLegacyUpgradeAndRemoval(t *testing.T) {
	forms := []string{"script"}
	if installPlatform() == "unix" {
		forms = append(forms, "symlink")
	} else {
		forms = append(forms, "cmd-original", "cmd-lifecycle", "cmd-single-pass")
	}
	for _, form := range forms {
		for _, upgrade := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/upgrade=%v", form, upgrade), func(t *testing.T) {
				e := newEnv(t)
				if installPlatform() == "windows" {
					e.home = filepath.Join(e.home, "runtime % space")
					e.cfg.Path = filepath.Join(e.home, "config.json")
				}
				e.skill("tool")
				e.declare("tool")
				if r := e.install(Options{}); r.Status != "ok" {
					t.Fatalf("install: %+v", r)
				}
				bin := filepath.Join(e.project, ".agents", "bin")
				shim := filepath.Join(bin, shimName("tool-tool"))
				ledger := filepath.Join(filepath.Dir(bin), projectBinLedgerName)
				if err := os.Remove(ledger); err != nil {
					t.Fatal(err)
				}
				if form == "symlink" {
					target := ""
					if !runtimestore.MatchesLegacyShim(shim, installPlatform(), func(path string) bool { target = path; return true }) {
						t.Fatal("installed shim not recognized")
					}
					if err := os.Remove(shim); err != nil {
						t.Fatal(err)
					}
					rel, err := filepath.Rel(bin, target)
					if err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(rel, shim); err != nil {
						t.Fatal(err)
					}
				}
				if strings.HasPrefix(form, "cmd-") {
					target := ""
					if !runtimestore.MatchesLegacyShim(shim, "windows", func(path string) bool { target = path; return true }) {
						t.Fatal("installed cmd not recognized")
					}
					escaped := strings.ReplaceAll(target, "%", "%%")
					content := "@echo off\r\n\"" + target + "\" %*\r\n"
					if form == "cmd-lifecycle" {
						prefix := strings.ReplaceAll(filepath.Join(e.home, "runtime deps"), "%", "%%")
						content = "@echo off\r\nsetlocal DisableDelayedExpansion\r\nset \"PATH=" + prefix + ";%PATH%\"\r\ncall \"" + escaped + "\" %*\r\nexit /b %ERRORLEVEL%\r\n"
					}
					if form == "cmd-single-pass" {
						content = strings.Replace(runtimestore.WindowsShimContent(target, nil), "call \""+strings.ReplaceAll(escaped, "%", "%%")+"\"", "call \""+escaped+"\"", 1)
					}
					if err := os.WriteFile(shim, []byte(content), 0o755); err != nil {
						t.Fatal(err)
					}
				}
				if upgrade {
					if r := e.install(Options{}); r.Status != "ok" {
						t.Fatalf("legacy upgrade: %+v", r)
					}
					if _, err := os.Stat(ledger); err != nil {
						t.Fatal(err)
					}
				}
				e.declare()
				if r := e.install(Options{}); r.Status != "ok" {
					t.Fatalf("removal: %+v", r)
				}
				if _, err := os.Lstat(shim); !os.IsNotExist(err) {
					t.Fatalf("managed shim survived removal: %v", err)
				}
			})
		}
	}
}

// A corrupt ledger refuses without changing it or the foreign command. The
// preceding install tests provide valid-ledger positive controls.
func TestProjectBinInvalidOwnershipLedger(t *testing.T) {
	e := newEnv(t)
	e.skill("tool")
	e.declare("tool")
	bin := filepath.Join(e.project, ".agents", "bin")
	if err := os.MkdirAll(bin, 0755); err != nil {
		t.Fatal(err)
	}
	ledger := filepath.Join(filepath.Dir(bin), projectBinLedgerName)
	const bad = "foreign ledger data"
	if err := os.WriteFile(ledger, []byte(bad), 0600); err != nil {
		t.Fatal(err)
	}
	r := e.install(Options{})
	if r.Status == "ok" || !strings.Contains(fmt.Sprint(r), "project_bin_ownership_invalid") || !strings.Contains(fmt.Sprint(r), ledger) {
		t.Fatalf("refusal: %+v", r)
	}
	if got, _ := os.ReadFile(ledger); string(got) != bad {
		t.Fatal("ledger changed")
	}
	if _, err := os.Lstat(filepath.Join(bin, shimName("tool-tool"))); !os.IsNotExist(err) {
		t.Fatalf("refusal published shim: %v", err)
	}
}

// Ownership checked during staging remains pinned when the journal captures
// preimages: a later foreign replacement refuses before any journal/write.
func TestProjectBinOwnershipSwapBeforeJournal(t *testing.T) {
	root := t.TempDir()
	live := filepath.Join(root, "tool")
	staged := filepath.Join(root, "staged")
	for _, p := range []string{live, staged} {
		if err := os.WriteFile(p, []byte("managed"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	expected, err := shimFingerprint(live)
	if err != nil {
		t.Fatal(err)
	}
	targets := scopeTargets{projectBinPreimages: map[string]string{live: expected}}
	targets.plan.Replace(staging.ClassCanonicalShim, "tool", live, staged)
	req := commitRequest{commit: CommitDeps{NewTransactionID: func() (string, error) { return "fixture", nil }}}
	if _, _, err := journalPlan(req, targets); err != nil {
		t.Fatalf("unchanged positive control: %v", err)
	}
	if err := os.WriteFile(live, []byte("foreign"), 0755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := journalPlan(req, targets); err == nil || !strings.Contains(err.Error(), projectBinConflict) || !strings.Contains(err.Error(), live) {
		t.Fatalf("swap refusal: %v", err)
	}
	if got, _ := os.ReadFile(live); string(got) != "foreign" {
		t.Fatal("foreign replacement changed")
	}
}

// Missing inventories and entries are valid empty controls. A non-directory
// ancestor or unreadable marker cannot become absence or authorize adoption;
// the probes leave the obstructing bytes and marker directory untouched.
func TestProjectBinOwnershipReadFailuresRemainUnreadable(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	bin := filepath.Join(root, "project", ".agents", "bin")
	if _, err := inspectProjectBin(home, bin, installPlatform(), "", nil); err != nil {
		t.Fatalf("absent inventory: %v", err)
	}
	if fingerprint, err := shimFingerprint(filepath.Join(bin, "missing")); err != nil || fingerprint != "" {
		t.Fatalf("absent entry = %q, %v", fingerprint, err)
	}
	obstructed := filepath.Join(root, "obstructed")
	const payload = "foreign ancestor\n"
	if err := os.WriteFile(obstructed, []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
	requireUnreadable := func(err error) {
		t.Helper()
		if err == nil || !strings.Contains(err.Error(), "manager_state_unreadable") {
			t.Fatalf("want unreadable state, got %v", err)
		}
	}
	_, err := inspectProjectBin(home, filepath.Join(obstructed, ".agents", "bin"), installPlatform(), "", nil)
	requireUnreadable(err)
	_, err = shimFingerprint(filepath.Join(obstructed, "command"))
	requireUnreadable(err)
	_, err = previousProjectCommands(home, filepath.Join(obstructed, "bin"))
	requireUnreadable(err)
	if got, err := os.ReadFile(obstructed); err != nil || string(got) != payload {
		t.Fatalf("ancestor changed: %q, %v", got, err)
	}
	markerDir := filepath.Join(filepath.Dir(bin), "skills", "tool", ".csk-install.json")
	if err := os.MkdirAll(markerDir, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err = previousProjectCommands(home, bin)
	requireUnreadable(err)
	if info, err := os.Stat(markerDir); err != nil || !info.IsDir() {
		t.Fatalf("marker directory changed: %v, %v", info, err)
	}
}
