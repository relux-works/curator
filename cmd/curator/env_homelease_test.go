package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/homelease"
)

type homeLeaseCLIImage struct {
	entries map[string]helpTreeEntry
	infos   map[string]os.FileInfo
}

// Only synthetic manager fixtures are passed here; operator homes are guarded
// separately by TestMain without recursively reading their contents.
func homeLeaseCLIImageOf(t *testing.T, root string) homeLeaseCLIImage {
	t.Helper()
	image := homeLeaseCLIImage{entries: nativeOwnerSnapshot(t, root), infos: map[string]os.FileInfo{}}
	for rel := range image.entries {
		info, err := os.Lstat(filepath.Join(root, rel))
		if err != nil {
			t.Fatal(err)
		}
		image.infos[rel] = info
	}
	return image
}

func homeLeaseCLIUnchanged(t *testing.T, before homeLeaseCLIImage, root string) {
	t.Helper()
	after := homeLeaseCLIImageOf(t, root)
	if !reflect.DeepEqual(before.entries, after.entries) {
		t.Fatal("lease command changed fixture bytes, links, modes or directory membership")
	}
	for rel, info := range before.infos {
		if got := after.infos[rel]; got == nil || !os.SameFile(info, got) || !info.ModTime().Equal(got.ModTime()) {
			t.Fatalf("lease command changed fixture identity at %q", rel)
		}
	}
}

func runHomeLeaseCLI(t *testing.T, source stubConfigSource, args ...string) homelease.Result {
	t.Helper()
	code, stdout, _ := runProfile(t, source, append([]string{"env", "lease"}, args...)...)
	if strings.Count(stdout, "\n") != 1 || !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("lease command did not emit one JSON line: %q", stdout)
	}
	result, err := homelease.DecodeResult([]byte(strings.TrimSuffix(stdout, "\n")))
	if err != nil || result.Exit != code {
		t.Fatalf("closed lease result/exit mismatch: %d %q %v", code, stdout, err)
	}
	return result
}

// Recognized usage failures pass through the real top-level router rather than
// generic help, and preserve the manager tree. Read-only list is the control.
func TestEnvLeaseCLIEnvelope(t *testing.T) {
	source, home := profileHome(t)
	for _, args := range [][]string{{"acquire", "--help"}, {"list", "--format", "text"}, {"break", "--lease", "absent", "--if-revision", "1", "--reason", "acknowledged"}} {
		before := homeLeaseCLIImageOf(t, home)
		result := runHomeLeaseCLI(t, source, args...)
		if result.Exit != exitUsage || result.OK || result.Code == nil || *result.Code != "usage" {
			t.Fatalf("recognized usage: %+v", result)
		}
		homeLeaseCLIUnchanged(t, before, home)
	}
	before := homeLeaseCLIImageOf(t, home)
	if result := runHomeLeaseCLI(t, source, "list"); !result.OK || result.Reason != "listed" {
		t.Fatalf("list control: %+v", result)
	}
	homeLeaseCLIUnchanged(t, before, home)
}

// Real CLI acquisition blocks sync and unconfirmed release without writes.
// An explicit audited break permits sync; exact terminal replay and unknown
// release stay side-effect free. No production observer is simulated here.
func TestEnvLeaseCLIWriterAndTerminalReplay(t *testing.T) {
	source, home := profileHome(t)
	writeNativeCredentials(t) // Synthetic provider roots from profileHome only.
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "lease CLI control\n")
	for _, args := range [][]string{{"profile", "install", pkg}, {"env", "resolve", "codex_cli", "--repair"}} {
		if code, _, stderr := runProfile(t, source, args...); code != exitOK {
			t.Fatalf("fixture command %q: %d %s", args[0], code, stderr)
		}
	}
	lock, err := os.ReadFile(filepath.Join(envprofile.ProfileDir(home, "acme"), "lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(lock)
	holder, err := json.Marshal(homelease.Holder{Kind: "prepared", Launcher: "curator-run", ManagerID: "cli-fixture", ManagerRegisteredAt: "2026-10-27T00:00:00Z", RequestID: "cli-request"})
	if err != nil {
		t.Fatal(err)
	}
	acquire := []string{"acquire", "--env", "codex_cli", "--profile", "acme", "--lock-sha256", hex.EncodeToString(digest[:]), "--request-id", "cli-request", "--holder", string(holder)}
	before := homeLeaseCLIImageOf(t, home)
	native := os.Getenv("CODEX_HOME")
	wrongNative := filepath.Join(t.TempDir(), "different-native")
	t.Setenv("CODEX_HOME", wrongNative)
	result := runHomeLeaseCLI(t, source, acquire...)
	if result.Exit != exitFail || result.Reason != "home_not_current" {
		t.Fatalf("detached native override admitted: %+v", result)
	}
	homeLeaseCLIUnchanged(t, before, home)
	if _, err := os.Lstat(wrongNative); !os.IsNotExist(err) {
		t.Fatalf("refusal provisioned the new native target: %v", err)
	}
	t.Setenv("CODEX_HOME", native)
	result = runHomeLeaseCLI(t, source, acquire...)
	if !result.OK || result.Reason != "acquired" {
		t.Fatalf("acquire control: %+v", result)
	}
	var outcome homelease.Outcome
	raw, _ := json.Marshal(result.Result)
	if err := json.Unmarshal(raw, &outcome); err != nil || outcome.Lease == nil || !outcome.Changed {
		t.Fatalf("acquire outcome: %s %v", raw, err)
	}
	id := outcome.Lease.LeaseID
	before = homeLeaseCLIImageOf(t, home)
	code, stdout, stderr := runProfile(t, source, "profile", "sync")
	if code != exitFail || stdout != "" || !strings.Contains(stderr, "environment_home_leased") || !strings.Contains(stderr, id) {
		t.Fatalf("leased sync: %d %q %q", code, stdout, stderr)
	}
	homeLeaseCLIUnchanged(t, before, home)
	result = runHomeLeaseCLI(t, source, "release", "--lease", id)
	if result.Exit != exitFail || result.Reason != "session_state_unknown" {
		t.Fatalf("unconfirmed release: %+v", result)
	}
	homeLeaseCLIUnchanged(t, before, home)
	result = runHomeLeaseCLI(t, source, "break", "--lease", id, "--if-revision", "1", "--reason", "explicit abandonment")
	if result.Exit != exitUsage || result.Reason != "operator_confirmation_required" {
		t.Fatalf("unauthorized break: %+v", result)
	}
	homeLeaseCLIUnchanged(t, before, home)
	result = runHomeLeaseCLI(t, source, "break", "--lease", id, "--if-revision", "1", "--operator", "--reason", "explicit abandonment")
	if !result.OK || result.Reason != "broken" {
		t.Fatalf("break control: %+v", result)
	}
	before = homeLeaseCLIImageOf(t, home)
	result = runHomeLeaseCLI(t, source, acquire...)
	raw, _ = json.Marshal(result.Result)
	if err := json.Unmarshal(raw, &outcome); err != nil || !result.OK || result.Reason != "already_broken" || outcome.Changed || outcome.Lease == nil || outcome.Lease.LeaseID != id {
		t.Fatalf("terminal replay: %+v %s %v", result, raw, err)
	}
	homeLeaseCLIUnchanged(t, before, home)
	result = runHomeLeaseCLI(t, source, "release", "--lease", "absent")
	if !result.OK || result.Reason != "unknown_lease" {
		t.Fatalf("unknown release: %+v", result)
	}
	homeLeaseCLIUnchanged(t, before, home)
	if code, _, stderr := runProfile(t, source, "profile", "sync"); code != exitOK {
		t.Fatalf("unleased sync control: %d %s", code, stderr)
	}
}
