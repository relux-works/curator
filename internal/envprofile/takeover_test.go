package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
)

// Production entry points under test: Install, UseWithPolicy,
// SyncWithPolicy, Resolve, StatusOf, List. Every fixture builds paths with
// filepath.Join over platform-absolute temporary bases.

// installIdleProfile installs a path profile through the production
// Install. The first install activates into the pinned test homes; the
// caller then redirects a native home at a fresh directory so the next
// switch meets pristine unmanaged state there.
func installIdleProfile(t *testing.T, home, name string) {
	t.Helper()
	source := filepath.Join(t.TempDir(), "source")
	writeManifestPackage(t, source,
		`{"schema_version": 1, "name": "`+name+`", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "managed\n"})
	if _, _, _, err := Install(home, InstallOptions{Operand: source}); err != nil {
		t.Fatal(err)
	}
}

// claudeHome redirects the claude_code native home at a fresh directory
// carrying no marker, so the next switch inventories unmanaged state.
func claudeHome(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "claude")
	t.Setenv("CLAUDE_CONFIG_DIR", dir)
	return dir
}

// TestUseWithoutTakeoverRefusesUnmanaged drives the production switch
// against a native home carrying an unmanaged root-context file: the entry
// fails with environment_surface_unmanaged_conflict and writes nothing.
func TestUseWithoutTakeoverRefusesUnmanaged(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	installIdleProfile(t, home, "acme")
	native := claudeHome(t)
	if err := os.MkdirAll(native, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(native, "CLAUDE.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	results, err := UseWithPolicy(home, "acme", "", "", false, Policy{})
	if err == nil {
		t.Fatal("switch over unmanaged files must fail without --takeover")
	}
	found := false
	for _, result := range results {
		if result.Adapter == "claude_code" && !result.OK &&
			strings.Contains(result.Detail, DiagUnmanagedConflict) {
			found = true
		}
	}
	if !found {
		t.Fatalf("results %+v carry no %s", results, DiagUnmanagedConflict)
	}
	if payload, _ := os.ReadFile(filepath.Join(native, "CLAUDE.md")); string(payload) != "mine\n" {
		t.Fatal("failed switch must not touch the unmanaged file")
	}
	if _, err := os.Stat(filepath.Join(native, ".agent-environment-backup")); !os.IsNotExist(err) {
		t.Fatal("failed switch must not write a backup")
	}
}

// TestUseTakeoverBacksUpAndNotifies drives the production switch with the
// section 9.5 takeover flag: every replaced file is copied into the next
// backup generation before the first write, the managed bytes land, and
// the replace notice names the backup.
func TestUseTakeoverBacksUpAndNotifies(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	installIdleProfile(t, home, "acme")
	native := claudeHome(t)
	if err := os.MkdirAll(native, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(native, "CLAUDE.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	results, err := UseWithPolicy(home, "acme", "", "", false, Policy{Takeover: true})
	if err != nil {
		t.Fatal(err)
	}
	var claude *EntryResult
	for index := range results {
		if results[index].Adapter == "claude_code" {
			claude = &results[index]
		}
		if !results[index].OK {
			t.Fatalf("%s: %s", results[index].Adapter, results[index].Detail)
		}
	}
	if claude == nil {
		t.Fatal("no claude_code entry")
	}
	if claude.Notice == "" || !strings.Contains(claude.Notice, ".agent-environment-backup") {
		t.Fatalf("notice %q names no backup", claude.Notice)
	}
	backup := filepath.Join(native, ".agent-environment-backup", "1", "CLAUDE.md")
	payload, err := os.ReadFile(backup) // #nosec G304 -- test backup file
	if err != nil || string(payload) != "mine\n" {
		t.Fatalf("backup = %q, err = %v; want the replaced bytes", payload, err)
	}
	managed, err := os.ReadFile(filepath.Join(native, "CLAUDE.md"))
	if err != nil || string(managed) == "mine\n" {
		t.Fatal("takeover must install the managed bytes")
	}
	current, err := Current(home)
	if err != nil || current != "acme" {
		t.Fatalf("takeover switch must record the current, got %q", current)
	}
}

// TestForeignSymlinkStopsSwitch drives the production switch against a
// managed-surface path that is already a symlink outside the manager's
// store: the entry stops with environment_foreign_manager_detected and the
// explicit abort-or-take-over choice; the flag takes over with backup.
func TestForeignSymlinkStopsSwitch(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	installIdleProfile(t, home, "acme")
	native := claudeHome(t)
	if err := os.MkdirAll(native, 0o755); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(t.TempDir(), "foreign.md")
	if err := os.WriteFile(foreign, []byte("foreign\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(native, "CLAUDE.md")
	if err := os.Symlink(foreign, link); err != nil {
		t.Skipf("this host cannot create symlinks: %v", err)
	}
	results, err := UseWithPolicy(home, "acme", "", "", false, Policy{})
	if err == nil {
		t.Fatal("foreign-manager symlink must stop the switch without --takeover")
	}
	found := false
	for _, result := range results {
		if result.Adapter == "claude_code" && strings.Contains(result.Detail, DiagForeignManager) &&
			strings.Contains(result.Detail, "take over") {
			found = true
		}
	}
	if !found {
		t.Fatalf("results %+v carry no %s with the take-over choice", results, DiagForeignManager)
	}
	taken, err := UseWithPolicy(home, "acme", "", "", false, Policy{Takeover: true})
	if err != nil {
		t.Fatalf("takeover over a foreign symlink: %v", err)
	}
	for _, result := range taken {
		if !result.OK {
			t.Fatalf("%s: %s", result.Adapter, result.Detail)
		}
	}
	// The takeover must replace the link with a regular file, never write
	// through it: the surface stops being a symlink and the foreign file's
	// bytes are unchanged (environments §9.5 takes over with backup, never
	// absorbs in the wrong direction).
	if info, err := os.Lstat(link); err != nil {
		t.Fatalf("takeover removed the surface: %v", err)
	} else if info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("takeover left %q a symlink; a copied surface must never follow a link", link)
	}
	if payload, err := os.ReadFile(foreign); err != nil {
		t.Fatalf("foreign file unreadable after takeover: %v", err)
	} else if string(payload) != "foreign\n" {
		t.Fatalf("takeover overwrote the foreign file: %q", payload)
	}
	// The backup preserves the foreign bytes before the first write.
	backups, err := os.ReadDir(filepath.Join(native, ".agent-environment-backup"))
	if err != nil || len(backups) == 0 {
		t.Fatalf("takeover wrote no backup: %v %v", backups, err)
	}
}

// TestTakeoverWarnsDotfileHeuristic drives the production switch with a
// takeover while a dotfile manager's closed-list state location exists:
// the non-blocking environment_foreign_manager_suspected warning names
// the manager, and the switch still converges.
func TestTakeoverWarnsDotfileHeuristic(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	installIdleProfile(t, home, "acme")
	native := claudeHome(t)
	if err := os.MkdirAll(native, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(native, "CLAUDE.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	operator := t.TempDir()
	t.Setenv("HOME", operator)
	if err := os.MkdirAll(filepath.Join(operator, ".local", "share", "chezmoi"), 0o755); err != nil {
		t.Fatal(err)
	}
	results, err := UseWithPolicy(home, "acme", "", "", false, Policy{Takeover: true})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, result := range results {
		if !result.OK {
			t.Fatalf("%s: %s", result.Adapter, result.Detail)
		}
		for _, warning := range result.Warnings {
			if strings.Contains(warning, DiagForeignSuspect) && strings.Contains(warning, "chezmoi") {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("results %+v carry no %s naming chezmoi", results, DiagForeignSuspect)
	}
}

// TestReadOnlyCommandsNeverOnboard drives the read-only surface over
// unmanaged state: list, status, and bare resolve report without writing
// a backup, a marker, or a prompt.
func TestReadOnlyCommandsNeverOnboard(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	installIdleProfile(t, home, "acme")
	native := claudeHome(t)
	if err := os.MkdirAll(native, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(native, "CLAUDE.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := List(home); err != nil {
		t.Fatal(err)
	}
	status, err := StatusOf(StatusRequest{Home: home, Machine: envregistry.DefaultMachineConfig()})
	if err != nil {
		t.Fatal(err)
	}
	_ = status
	if _, err := Resolve(ResolveRequest{
		Home: home, Profile: "acme", EnvID: "claude_code",
		LaunchDir: t.TempDir(), Machine: envregistry.DefaultMachineConfig(),
		Detect: func(envregistry.Adapter) string { return "unknown" },
	}); err == nil {
		t.Fatal("bare resolve of an unprovisioned home must stay stale")
	}
	if _, err := os.Stat(filepath.Join(native, ".agent-environment-backup")); !os.IsNotExist(err) {
		t.Fatal("read-only commands must never write a backup")
	}
	if _, err := os.Stat(filepath.Join(native, envmarker.Name)); !os.IsNotExist(err) {
		t.Fatal("read-only commands must never write a marker")
	}
}

// TestRepairTakeoverProvisionsManagedHome drives the production Resolve
// with --repair over a managed home carrying an unmanaged file: without
// the flag the repair fails, with it the file is backed up and the home
// provisions.
func TestRepairTakeoverProvisionsManagedHome(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	installIdleProfile(t, home, "acme")
	managed := ManagedHomeDir(home, "acme", "claude_code")
	if err := os.MkdirAll(managed, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(managed, "CLAUDE.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	base := ResolveRequest{
		Home: home, Profile: "acme", EnvID: "claude_code",
		LaunchDir: t.TempDir(), Machine: envregistry.DefaultMachineConfig(),
		Detect: func(envregistry.Adapter) string { return "unknown" },
		Repair: true, Format: "json",
	}
	if _, err := Resolve(base); err == nil {
		t.Fatal("repair over unmanaged files must fail without --takeover")
	}
	req := base
	req.Policy = Policy{Takeover: true}
	result, err := Resolve(req)
	if err != nil {
		t.Fatalf("takeover repair: %v", err)
	}
	if !result.Provisioned {
		t.Fatal("takeover repair must provision")
	}
	backup := filepath.Join(managed, ".agent-environment-backup", "1", "CLAUDE.md")
	payload, err := os.ReadFile(backup) // #nosec G304 -- test backup file
	if err != nil || string(payload) != "mine\n" {
		t.Fatalf("backup = %q, err = %v; want the replaced bytes", payload, err)
	}
}
