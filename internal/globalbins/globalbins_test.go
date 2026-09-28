package globalbins

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/stateread"
)

func TestRefreshPublishesAndRemovesManagedUnixShims(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix symlinks are exercised on Linux and macOS")
	}
	root := t.TempDir()
	managerHome := filepath.Join(root, "manager")
	canonicalBin := filepath.Join(managerHome, "global", "bin")
	if err := os.MkdirAll(canonicalBin, 0o755); err != nil {
		t.Fatal(err)
	}
	canonical := filepath.Join(canonicalBin, "tool")
	if err := os.WriteFile(canonical, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	userHome := filepath.Join(root, "user")
	userBin := filepath.Join(userHome, ".local", "bin")
	if err := os.MkdirAll(userBin, 0o755); err != nil {
		t.Fatal(err)
	}
	messages := Refresh(managerHome, map[string]bool{"tool": true}, "unix", map[string]string{
		"PATH": userBin,
	}, userHome)
	if !containsMessage(messages, "command shims published") {
		t.Fatalf("publish messages: %v", messages)
	}
	published := filepath.Join(userBin, "tool")
	target, err := filepath.EvalSymlinks(published)
	targetInfo, targetErr := os.Stat(target)
	canonicalInfo, canonicalErr := os.Stat(canonical)
	if err != nil || targetErr != nil || canonicalErr != nil || !os.SameFile(targetInfo, canonicalInfo) {
		t.Fatalf("published shim = %q, %v; want %q", target, err, canonical)
	}
	payload, err := os.ReadFile(filepath.Join(userBin, managedFile))
	if err != nil {
		t.Fatal(err)
	}
	var recorded ledger
	if err := json.Unmarshal(payload, &recorded); err != nil || len(recorded.Entries) != 1 || recorded.Entries[0] != "tool" {
		t.Fatalf("ownership ledger = %+v, %v", recorded, err)
	}

	Refresh(managerHome, map[string]bool{}, "unix", map[string]string{"PATH": userBin}, userHome)
	if _, err := os.Lstat(published); !os.IsNotExist(err) {
		t.Fatalf("stale managed shim survived: %v", err)
	}
}

type adoptionFixture struct {
	managerHome string
	userHome    string
	userBin     string
	canonical   string
	published   string
	platform    string
	environment map[string]string
}

func newAdoptionFixture(t *testing.T, platform string, published []byte) adoptionFixture {
	t.Helper()
	root := t.TempDir()
	managerHome := filepath.Join(root, "manager")
	userHome := filepath.Join(root, "user")
	userBin := filepath.Join(userHome, ".local", "bin")
	canonicalBin := filepath.Join(managerHome, "global", "bin")
	if err := os.MkdirAll(canonicalBin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(userBin, 0o755); err != nil {
		t.Fatal(err)
	}
	canonical := shimPath(canonicalBin, "tool", platform)
	if err := os.WriteFile(canonical, []byte("canonical Curator target\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	publishedPath := shimPath(userBin, "tool", platform)
	if published != nil {
		if err := os.WriteFile(publishedPath, published, 0o751); err != nil {
			t.Fatal(err)
		}
		fixed := time.Date(2021, 4, 5, 6, 7, 8, 0, time.UTC)
		if err := os.Chtimes(publishedPath, fixed, fixed); err != nil {
			t.Fatal(err)
		}
	}
	return adoptionFixture{
		managerHome: managerHome,
		userHome:    userHome,
		userBin:     userBin,
		canonical:   canonical,
		published:   publishedPath,
		platform:    platform,
		environment: map[string]string{"PATH": userBin, UserBinEnv: userBin},
	}
}

func canonicalForwardingBytes(fixture adoptionFixture) []byte {
	if fixture.platform == "windows" {
		return []byte(runtimestore.WindowsShimContent(fixture.canonical, nil))
	}
	return []byte(runtimestore.UnixShimContent(fixture.canonical, nil))
}

func TestAdoptCanonicalShimBacksUpAndGlobalInstallStagesItAsManaged(t *testing.T) {
	for _, platform := range []string{"unix", "windows"} {
		t.Run(platform, func(t *testing.T) {
			fixture := newAdoptionFixture(t, platform, nil)
			if err := writeLedger(fixture.userBin, map[string]bool{"other": true}); err != nil {
				t.Fatal(err)
			}
			want := canonicalForwardingBytes(fixture)
			if err := os.WriteFile(fixture.published, want, 0o751); err != nil {
				t.Fatal(err)
			}
			fixed := time.Date(2021, 4, 5, 6, 7, 8, 0, time.UTC)
			if err := os.Chtimes(fixture.published, fixed, fixed); err != nil {
				t.Fatal(err)
			}
			before, err := os.Stat(fixture.published)
			if err != nil {
				t.Fatal(err)
			}

			adopted, err := Adopt(fixture.managerHome, "tool", platform, fixture.environment, fixture.userHome, false)
			if err != nil {
				t.Fatalf("Adopt() = %v", err)
			}
			if adopted.Path != fixture.published || adopted.Backup == "" || adopted.AlreadyManaged || adopted.DryRun {
				t.Fatalf("Adopt() = %+v, want a new backed-up adoption", adopted)
			}
			if filepath.Dir(adopted.Backup) != filepath.Join(fixture.managerHome, "backups", "global-bins") {
				t.Fatalf("backup path = %s, want Curator backup root", adopted.Backup)
			}
			backup, err := os.ReadFile(adopted.Backup)
			if err != nil || !bytes.Equal(backup, want) {
				t.Fatalf("backup bytes = %q, %v; want canonical shim %q", backup, err, want)
			}
			backupInfo, err := os.Stat(adopted.Backup)
			if err != nil || backupInfo.Mode().Perm() != before.Mode().Perm() || !backupInfo.ModTime().Equal(before.ModTime()) {
				t.Fatalf("backup metadata = (%v, %v); want mode %v and mtime %v", backupInfo, err, before.Mode().Perm(), before.ModTime())
			}
			managed, err := readLedger(fixture.userBin)
			if err != nil || !managed["tool"] || !managed["other"] {
				t.Fatalf("adoption ledger = %v, %v; want tool added without losing other", managed, err)
			}

			// install.stageGlobalScope calls StageForwarding. It must recognize
			// the adopted bytes and ledger entry as its own, without an unmanaged
			// conflict.
			forwarding, err := StageForwarding(t.TempDir(), fixture.managerHome, map[string]bool{"tool": true}, platform, fixture.environment, fixture.userHome)
			if err != nil || forwarding.Published != 1 || containsMessage(forwarding.Messages, "not managed by Curator") {
				t.Fatalf("StageForwarding() = (%+v, %v); want one managed command and no conflict", forwarding, err)
			}

			second, err := Adopt(fixture.managerHome, "tool", platform, fixture.environment, fixture.userHome, false)
			if err != nil || !second.AlreadyManaged || second.Backup != "" {
				t.Fatalf("second Adopt() = (%+v, %v); want idempotent no-op", second, err)
			}
			backups, err := os.ReadDir(filepath.Join(fixture.managerHome, "backups", "global-bins"))
			if err != nil || len(backups) != 1 {
				t.Fatalf("backup count after idempotent adoption = (%d, %v), want 1", len(backups), err)
			}
		})
	}
}

func TestAdoptDryRunWritesNothingForUnixAndWindowsShims(t *testing.T) {
	for _, platform := range []string{"unix", "windows"} {
		t.Run(platform, func(t *testing.T) {
			fixture := newAdoptionFixture(t, platform, nil)
			want := canonicalForwardingBytes(fixture)
			if err := os.WriteFile(fixture.published, want, 0o751); err != nil {
				t.Fatal(err)
			}

			result, err := Adopt(fixture.managerHome, "tool", platform, fixture.environment, fixture.userHome, true)
			if err != nil || !result.DryRun || result.AlreadyManaged || result.Backup != "" {
				t.Fatalf("dry-run Adopt() = (%+v, %v)", result, err)
			}
			if _, err := os.Lstat(filepath.Join(fixture.userBin, managedFile)); !os.IsNotExist(err) {
				t.Fatalf("dry-run wrote ownership marker: %v", err)
			}
			if _, err := os.Lstat(filepath.Join(fixture.managerHome, "backups")); !os.IsNotExist(err) {
				t.Fatalf("dry-run wrote backup root: %v", err)
			}
			got, err := os.ReadFile(fixture.published)
			if err != nil || !bytes.Equal(got, want) {
				t.Fatalf("dry-run changed source entry: %q, %v", got, err)
			}
		})
	}
}

func TestAdoptRefusalsDoNotWrite(t *testing.T) {
	cases := []struct {
		name       string
		prepare    func(*testing.T, adoptionFixture)
		wantReason string
	}{
		{
			name: "different bytes",
			prepare: func(t *testing.T, fixture adoptionFixture) {
				if err := os.WriteFile(fixture.published, []byte("manual replacement\n"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			wantReason: "bytes differ",
		},
		{
			name: "symbolic link",
			prepare: func(t *testing.T, fixture adoptionFixture) {
				if err := os.Remove(fixture.published); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(fixture.canonical, fixture.published); err != nil {
					t.Skipf("host cannot create symbolic links: %v", err)
				}
			},
			wantReason: "symbolic link",
		},
		{
			name: "special entry",
			prepare: func(t *testing.T, fixture adoptionFixture) {
				if err := os.Remove(fixture.published); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(fixture.published, 0o755); err != nil {
					t.Fatal(err)
				}
			},
			wantReason: "not a regular file",
		},
		{
			name: "missing entry",
			prepare: func(t *testing.T, fixture adoptionFixture) {
				if err := os.Remove(fixture.published); err != nil {
					t.Fatal(err)
				}
			},
			wantReason: "entry does not exist",
		},
		{
			name: "unknown command",
			prepare: func(t *testing.T, fixture adoptionFixture) {
				if err := os.Remove(fixture.canonical); err != nil {
					t.Fatal(err)
				}
			},
			wantReason: "no Curator global command has a canonical target",
		},
	}

	for _, platform := range []string{"unix", "windows"} {
		for _, tc := range cases {
			t.Run(platform+"/"+tc.name, func(t *testing.T) {
				fixture := newAdoptionFixture(t, platform, nil)
				if err := os.WriteFile(fixture.published, canonicalForwardingBytes(fixture), 0o755); err != nil {
					t.Fatal(err)
				}
				original, err := os.ReadFile(fixture.published)
				if err != nil {
					t.Fatal(err)
				}
				tc.prepare(t, fixture)

				_, err = Adopt(fixture.managerHome, "tool", platform, fixture.environment, fixture.userHome, false)
				if err == nil || !strings.Contains(err.Error(), fixture.published) || !strings.Contains(err.Error(), tc.wantReason) {
					t.Fatalf("Adopt() error = %v; want path %s and reason %q", err, fixture.published, tc.wantReason)
				}
				if _, err := os.Lstat(filepath.Join(fixture.userBin, managedFile)); !os.IsNotExist(err) {
					t.Fatalf("refusal wrote ownership marker: %v", err)
				}
				if _, err := os.Lstat(filepath.Join(fixture.managerHome, "backups")); !os.IsNotExist(err) {
					t.Fatalf("refusal wrote backup root: %v", err)
				}
				if tc.name == "different bytes" {
					got, err := os.ReadFile(fixture.published)
					if err != nil || string(got) != "manual replacement\n" {
						t.Fatalf("mismatch entry changed: %q, %v", got, err)
					}
				}
				if tc.name == "unknown command" {
					got, err := os.ReadFile(fixture.published)
					if err != nil || !bytes.Equal(got, original) {
						t.Fatalf("unknown-command entry changed: %q, %v", got, err)
					}
				}
			})
		}
	}
}

func TestRefreshAndStageRefuseAnUnreadableOwnershipLedger(t *testing.T) {
	root := t.TempDir()
	managerHome := filepath.Join(root, "manager")
	userHome := filepath.Join(root, "user")
	userBin := filepath.Join(userHome, "bin")
	if err := os.MkdirAll(userBin, 0o755); err != nil {
		t.Fatal(err)
	}
	ledgerPath := filepath.Join(userBin, managedFile)
	original := []byte(`{"schema_version":1,"entries":["old-tool"]}` + "\n")
	if err := os.WriteFile(ledgerPath, original, 0o600); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		t.Skip("platform-control: POSIX mode-bit unreadability")
	}
	if err := os.Chmod(ledgerPath, 0o000); err != nil {
		t.Skipf("this host cannot create mode-000 manager state: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(ledgerPath, 0o600) })
	if _, err := os.ReadFile(ledgerPath); err == nil {
		t.Skip("this environment can read a mode-000 file; unreadability is untestable here")
	}
	environment := map[string]string{"PATH": userBin}
	messages := Refresh(managerHome, map[string]bool{}, "unix", environment, userHome)
	if !containsMessage(messages, stateread.DiagUnreadable) {
		t.Fatalf("refresh messages for unreadable ledger = %v, want %s", messages, stateread.DiagUnreadable)
	}
	if _, err := StageForwarding(t.TempDir(), managerHome, map[string]bool{}, "unix", environment, userHome); err == nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) {
		t.Fatalf("stage with unreadable ledger error = %v, want %s", err, stateread.DiagUnreadable)
	}
	if err := os.Chmod(ledgerPath, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(ledgerPath)
	if err != nil || string(got) != string(original) {
		t.Fatalf("unreadable ledger was overwritten = (%q, %v), want original bytes", got, err)
	}
}

func TestRefreshNeverOverwritesUnmanagedCommand(t *testing.T) {
	root := t.TempDir()
	managerHome := filepath.Join(root, "manager")
	canonicalBin := filepath.Join(managerHome, "global", "bin")
	if err := os.MkdirAll(canonicalBin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(canonicalBin, "tool"), []byte("canonical"), 0o755); err != nil {
		t.Fatal(err)
	}
	userHome := filepath.Join(root, "user")
	userBin := filepath.Join(userHome, "bin")
	if err := os.MkdirAll(userBin, 0o755); err != nil {
		t.Fatal(err)
	}
	published := filepath.Join(userBin, "tool")
	if err := os.WriteFile(published, []byte("manual"), 0o755); err != nil {
		t.Fatal(err)
	}
	messages := Refresh(managerHome, map[string]bool{"tool": true}, "unix", map[string]string{
		"PATH": userBin,
	}, userHome)
	payload, err := os.ReadFile(published)
	if err != nil || string(payload) != "manual" {
		t.Fatalf("unmanaged command was overwritten: %q, %v", payload, err)
	}
	if !containsMessage(messages, "not managed by Curator") {
		t.Fatalf("missing unmanaged conflict warning: %v", messages)
	}
}

func TestRefreshDoesNotOverwriteFormerlyManagedCommandReplacedByUser(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix symlink ownership is exercised on Linux and macOS")
	}
	root := t.TempDir()
	managerHome := filepath.Join(root, "manager")
	canonicalBin := filepath.Join(managerHome, "global", "bin")
	if err := os.MkdirAll(canonicalBin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(canonicalBin, "tool"), []byte("canonical"), 0o755); err != nil {
		t.Fatal(err)
	}
	userHome := filepath.Join(root, "user")
	userBin := filepath.Join(userHome, ".local", "bin")
	if err := os.MkdirAll(userBin, 0o755); err != nil {
		t.Fatal(err)
	}
	environment := map[string]string{"PATH": userBin}
	Refresh(managerHome, map[string]bool{"tool": true}, "unix", environment, userHome)
	published := filepath.Join(userBin, "tool")
	if err := os.Remove(published); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(published, []byte("manual replacement"), 0o755); err != nil {
		t.Fatal(err)
	}

	messages := Refresh(managerHome, map[string]bool{"tool": true}, "unix", environment, userHome)
	payload, err := os.ReadFile(published)
	if err != nil || string(payload) != "manual replacement" {
		t.Fatalf("replacement was overwritten: %q, %v", payload, err)
	}
	if !containsMessage(messages, "not managed by Curator") {
		t.Fatalf("missing replacement conflict warning: %v", messages)
	}
}

func TestRefreshPublishesWindowsCommandWrapper(t *testing.T) {
	root := t.TempDir()
	managerHome := filepath.Join(root, "manager")
	canonicalBin := filepath.Join(managerHome, "global", "bin")
	if err := os.MkdirAll(canonicalBin, 0o755); err != nil {
		t.Fatal(err)
	}
	canonical := filepath.Join(canonicalBin, "tool.cmd")
	if err := os.WriteFile(canonical, []byte("@echo off\r\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	userHome := filepath.Join(root, "user")
	userBin := filepath.Join(userHome, ".local", "bin")
	if err := os.MkdirAll(userBin, 0o755); err != nil {
		t.Fatal(err)
	}
	messages := Refresh(managerHome, map[string]bool{"tool": true}, "windows", map[string]string{
		"PATH": userBin,
	}, userHome)
	payload, err := os.ReadFile(filepath.Join(userBin, "tool.cmd"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), `"`+canonical+`" %*`) {
		t.Fatalf("Windows forwarding wrapper:\n%s", payload)
	}
	if !containsMessage(messages, "command shims published") {
		t.Fatalf("publish messages: %v", messages)
	}
}

func TestSelectRejectsProtectedAndInvisibleBins(t *testing.T) {
	root := t.TempDir()
	userHome := filepath.Join(root, "user")
	managerHome := filepath.Join(userHome, ".curator")
	miseShims := filepath.Join(userHome, ".local", "share", "mise", "shims")
	if err := os.MkdirAll(miseShims, 0o755); err != nil {
		t.Fatal(err)
	}
	selection := Select(managerHome, "unix", map[string]string{"PATH": miseShims}, userHome)
	if selection.Path != "" || !strings.Contains(selection.Warning, "no safe PATH-visible") {
		t.Fatalf("protected selection = %+v", selection)
	}

	explicit := filepath.Join(userHome, "bin")
	selection = Select(managerHome, "unix", map[string]string{
		"PATH":     miseShims,
		UserBinEnv: explicit,
	}, userHome)
	if selection.Path != "" || !strings.Contains(selection.Warning, "is not on PATH") {
		t.Fatalf("invisible explicit selection = %+v", selection)
	}
}

func TestRefreshWarnsWhenNoSafeBinExists(t *testing.T) {
	root := t.TempDir()
	managerHome := filepath.Join(root, "manager")
	messages := Refresh(managerHome, map[string]bool{"tool": true}, "unix", map[string]string{
		"PATH": "/usr/bin",
	}, filepath.Join(root, "user"))
	if len(messages) != 1 || !strings.Contains(messages[0], filepath.Join(managerHome, "global", "bin")) ||
		!strings.Contains(messages[0], "curator shell-init --install") {
		t.Fatalf("fallback warning: %v", messages)
	}
}

// TestSafeSelectionFeedsStagedForwardingTargetWithoutLiveMutation runs on the
// host's own platform profile rather than a fixed "unix" one.
//
// The runtime store validates a script command against the semantics of the
// platform it was asked for, and the unix profile requires a POSIX execute bit.
// Windows has none to give -- os.WriteFile(0o755) there yields a file Go reports
// as 0666 -- so pinning "unix" made this case assert a permission model the host
// cannot express, and it failed on the executable check before reaching the
// staging behaviour it exists to cover. Asking for the host profile keeps the
// case running everywhere on the runtime shape that host actually uses.
func TestSafeSelectionFeedsStagedForwardingTargetWithoutLiveMutation(t *testing.T) {
	platform := runtimestore.Platform()
	root := t.TempDir()
	managerHome := filepath.Join(root, "manager")
	userHome := filepath.Join(root, "user")
	userBin := filepath.Join(userHome, ".local", "bin")
	if err := os.MkdirAll(userBin, 0o755); err != nil {
		t.Fatal(err)
	}
	selection := Select(managerHome, platform, map[string]string{"PATH": userBin}, userHome)
	if selection.Path != userBin {
		t.Fatalf("safe user bin selection = %+v", selection)
	}

	command := skillspec.Command{Name: "tool", Type: "script", UnixPath: "scripts/tool", WinPath: "scripts/tool.cmd"}
	scriptRel, scriptBody := "tool", "#!/bin/sh\n"
	liveName := "tool"
	if platform == "windows" {
		scriptRel, scriptBody = "tool.cmd", "@echo off\r\n"
		liveName = "tool.cmd"
	}
	snapshot := filepath.Join(root, "snapshot")
	if err := os.MkdirAll(filepath.Join(snapshot, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(snapshot, "scripts", scriptRel), []byte(scriptBody), 0o755); err != nil {
		t.Fatal(err)
	}
	runtimePlan, err := runtimestore.PrepareScriptRuntime(filepath.Join(root, "runtime-stage"), runtimestore.ScriptRuntimeSpec{
		Home: managerHome, SkillName: "skill-a", Commit: "commit-a", Snapshot: snapshot,
		RuntimeRoots: []string{"scripts"},
		Commands:     []skillspec.Command{command},
		Platform:     platform,
	})
	if err != nil {
		t.Fatal(err)
	}
	forward, err := runtimestore.NewManagedShim(runtimestore.SafeForwardingShim, selection.Path, "tool", platform)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := runtimestore.StageShimTransition(filepath.Join(root, "shim-stage"), []runtimestore.ShimSpec{{
		Destination: forward,
		Target:      runtimePlan.Commands["tool"],
	}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Desired) != 1 {
		t.Fatalf("forwarding transition = %+v", plan)
	}
	if _, err := os.Stat(filepath.Join(userBin, liveName)); !os.IsNotExist(err) {
		t.Fatalf("staging touched safe live user bin: %v", err)
	}
	if _, err := os.Stat(plan.Desired[0].StagedPath); err != nil {
		t.Fatalf("forwarding shim was not staged: %v", err)
	}
}

func containsMessage(messages []string, fragment string) bool {
	for _, message := range messages {
		if strings.Contains(message, fragment) {
			return true
		}
	}
	return false
}

// TestPublishedShimsTracksTheLedger pins the umbrella lookup's seam: a
// directory the manager never published into reports unpublished, and
// any ledger the manager wrote — including an emptied one — marks the
// directory as a publication target. Presence counts, not content.
func TestPublishedShimsTracksTheLedger(t *testing.T) {
	root := t.TempDir()
	fresh := filepath.Join(root, "fresh")
	if err := os.MkdirAll(fresh, 0o755); err != nil {
		t.Fatal(err)
	}
	if published, err := PublishedShims(fresh); err != nil || published {
		t.Fatal("a directory with no ledger reports published")
	}
	if published, err := PublishedShims(filepath.Join(root, "missing")); err != nil || published {
		t.Fatal("a missing directory reports published")
	}
	unreadable := filepath.Join(root, "unreadable")
	if err := os.MkdirAll(unreadable, 0o755); err != nil {
		t.Fatal(err)
	}
	loop := filepath.Join(unreadable, managedFile)
	if err := os.Symlink(loop, loop); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if published, err := PublishedShims(unreadable); published || err == nil {
		t.Fatalf("unreadable ledger = (%v, %v), want false and typed read failure", published, err)
	} else {
		var readErr *stateread.Error
		if !errors.As(err, &readErr) || readErr.Kind != stateread.KindUnreadable || readErr.Path != loop {
			t.Fatalf("unreadable ledger error = %v, want typed unreadable for %s", err, loop)
		}
	}
	published := filepath.Join(root, "published")
	if err := os.MkdirAll(published, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeLedger(published, map[string]bool{"tool": true}); err != nil {
		t.Fatal(err)
	}
	if got, err := PublishedShims(published); err != nil || !got {
		t.Fatal("a directory carrying the ownership ledger reports unpublished")
	}
	if err := writeLedger(published, map[string]bool{}); err != nil {
		t.Fatal(err)
	}
	if got, err := PublishedShims(published); err != nil || !got {
		t.Fatal("an emptied ledger stops marking its directory as published")
	}
}

// TestSelectReportsExplicit pins which selections carry the
// operator-declared publishing location: an explicitly configured bin
// reports Explicit, a scanned PATH entry does not.
func TestSelectReportsExplicit(t *testing.T) {
	root := t.TempDir()
	userHome := filepath.Join(root, "user")
	managerHome := filepath.Join(root, "manager")
	userBin := filepath.Join(userHome, ".local", "bin")
	if err := os.MkdirAll(userBin, 0o755); err != nil {
		t.Fatal(err)
	}
	explicit := Select(managerHome, "unix", map[string]string{
		"PATH":     userBin,
		UserBinEnv: userBin,
	}, userHome)
	if explicit.Path != userBin || !explicit.Explicit {
		t.Fatalf("explicit selection = %+v, want the declared bin marked explicit", explicit)
	}
	scanned := Select(managerHome, "unix", map[string]string{
		"PATH": userBin,
	}, userHome)
	if scanned.Path != userBin || scanned.Explicit {
		t.Fatalf("scanned selection = %+v, want the PATH entry without the explicit mark", scanned)
	}
}
