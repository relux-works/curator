package globalbins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/skillspec"
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
