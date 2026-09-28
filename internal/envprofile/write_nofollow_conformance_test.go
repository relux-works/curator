package envprofile

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
)

type writeNofollowVector struct {
	Cases    []writeNofollowCase             `json:"cases"`
	Fixtures map[string]writeNofollowFixture `json:"fixtures"`
}

type writeNofollowFixture struct {
	BytesBase64 string `json:"bytes_base64"`
	SHA256      string `json:"sha256"`
}

type writeNofollowCase struct {
	Name                string `json:"name"`
	Operation           string `json:"operation"`
	Mode                string `json:"mode"`
	TakeoverAuthorized  bool   `json:"takeover_authorized"`
	MarkerRecordsTarget bool   `json:"marker_records_target"`
	ParentLink          *struct {
		LinkPoints string `json:"link_points"`
	} `json:"parent_link"`
	Target struct {
		Kind       string  `json:"kind"`
		LinkOwner  *string `json:"link_owner"`
		LinkPoints *string `json:"link_points"`
	} `json:"target"`
	Expected struct {
		Outcome                  string  `json:"outcome"`
		Diagnostic               *string `json:"diagnostic"`
		EntryAfter               string  `json:"entry_after"`
		BackupHoldsLink          bool    `json:"backup_holds_link"`
		ForeignTargetSHA256After *string `json:"foreign_target_sha256_after"`
	} `json:"expected"`
}

// TestEnvironmentWriteNofollowVectors drives rc.13 materialize, takeover,
// and repair cases through UseWithPolicy and Resolve. The backup-source
// vector for an unauthorized, unowned symlink is classified separately:
// no production operation requests that backup. Ordinary takeover refuses
// the path during foreign-manager inventory; authorized takeover preserves
// the link itself in the backup. Those production paths remain driven by
// the adjacent takeover vectors.
func TestEnvironmentWriteNofollowVectors(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "environments-write-nofollow.json")) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var vectors writeNofollowVector
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	if len(vectors.Cases) != 11 {
		t.Fatalf("rc.13 write-nofollow vector has %d cases, want 11", len(vectors.Cases))
	}
	if len(vectors.Fixtures) == 0 {
		t.Fatal("rc.13 write-nofollow vector has no fixtures")
	}
	conformancecoverage.RunOutcomes(t, "environments-write-nofollow/cases", vectors.Cases,
		func(tc writeNofollowCase) string { return tc.Name }, func(caseT *testing.T, tc writeNofollowCase) conformancecoverage.Observation {
			if tc.Operation == "backup" && tc.Name == "backup-symlinked-target-refused" {
				return conformancecoverage.Observation{BoundReason: "no production entry requests backup of an unauthorized, unowned link: ordinary takeover refuses it in the foreign-manager inventory, and authorized takeover preserves the link itself; both production paths are driven by the adjacent takeover vectors"}
			}
			runWriteNofollowVectorCase(caseT, tc, vectors.Fixtures)
			return conformancecoverage.Observation{}
		})
}

func TestResolveRenderedDocumentReplacesStoreSymlinkWithoutFollowing(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	request := fx.request("codex_cli")
	request.Repair = true
	storeDocument := storeDocPath(fx.home, fx.profile, "codex_cli", ".agent-context/system-prompt.md")
	foreign := filepath.Join(t.TempDir(), "foreign.md")
	if err := os.WriteFile(foreign, []byte("foreign store target\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(storeDocument), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(foreign, storeDocument); err != nil {
		t.Skipf("symlink fixture unavailable: %v", err)
	}

	if _, err := Resolve(request); err != nil {
		t.Fatalf("resolve with planted rendered-document link: %v", err)
	}
	assertManagedEntry(t, storeDocument, "managed-file", ProfilesDir(fx.home))
	if got, err := os.ReadFile(foreign); err != nil || string(got) != "foreign store target\n" {
		t.Fatalf("foreign rendered-document target changed to %q (%v)", got, err)
	}
}

func runWriteNofollowVectorCase(t *testing.T, tc writeNofollowCase, fixtures map[string]writeNofollowFixture) {
	t.Helper()
	switch tc.Name {
	case "takeover-symlinked-target-authorized-replaced", "takeover-symlinked-target-unauthorized-stopped":
		runSwitchSymlinkTakeoverCase(t, tc, fixtures)
	case "materialize-symlinked-parent-refused", "takeover-symlinked-parent-authorized-still-refused":
		runResolveSymlinkParentCase(t, tc, fixtures)
	case "repair-planted-link-replaced":
		runResolvePlantedLinkRepairCase(t, tc, fixtures)
	case "repair-manager-owned-link-replaced":
		runResolveOwnedLinkRepairCase(t, tc)
	case "backup-symlinked-destination-refused":
		runSwitchBackupParentCase(t, tc, fixtures)
	case "materialize-clean-path-written":
		runResolveCleanMaterializeCase(t, tc)
	case "takeover-inside-link-unauthorized-unmanaged-conflict":
		runSwitchInsideStoreLinkCase(t, tc, fixtures)
	case "materialize-recorded-file-replaced":
		runResolveRecordedFileRepairCase(t, tc)
	default:
		t.Fatalf("unhandled rc.13 write-nofollow case %q", tc.Name)
	}
}

func fixtureBytes(t *testing.T, fixtures map[string]writeNofollowFixture, name string) []byte {
	t.Helper()
	fixture, ok := fixtures[name]
	if !ok {
		t.Fatalf("vector fixture %q is absent", name)
	}
	payload, err := base64.StdEncoding.DecodeString(fixture.BytesBase64)
	if err != nil {
		t.Fatalf("decode fixture %q: %v", name, err)
	}
	hash := sha256.Sum256(payload)
	if got := hex.EncodeToString(hash[:]); got != fixture.SHA256 {
		t.Fatalf("fixture %q hash = %s, want %s", name, got, fixture.SHA256)
	}
	return payload
}

func writeExternalTarget(t *testing.T, fixtures map[string]writeNofollowFixture) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "foreign-notes.md")
	if err := os.WriteFile(path, fixtureBytes(t, fixtures, "foreign-notes"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertForeignHash(t *testing.T, path string, want *string) {
	t.Helper()
	if want == nil {
		return
	}
	payload, err := os.ReadFile(path) // #nosec G304 -- vector-owned fixture path
	if err != nil {
		t.Fatalf("read foreign target: %v", err)
	}
	hash := sha256.Sum256(payload)
	got := hex.EncodeToString(hash[:])
	if got != *want {
		t.Fatalf("foreign target sha256 = %s, want %s", got, *want)
	}
}

func assertExpectedDiagnostic(t *testing.T, err error, want *string) {
	t.Helper()
	if want == nil {
		if err != nil {
			t.Fatalf("unexpected operation error: %v", err)
		}
		return
	}
	if err == nil || !strings.Contains(err.Error(), *want) {
		t.Fatalf("operation error = %v, want diagnostic %s", err, *want)
	}
}

func assertSwitchDiagnostic(t *testing.T, err error, results []EntryResult, want *string) {
	t.Helper()
	if want == nil {
		if err != nil {
			t.Fatalf("unexpected switch error: %v", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("switch error = nil, want diagnostic %s", *want)
	}
	for _, result := range results {
		if strings.Contains(result.Detail, *want) {
			return
		}
	}
	t.Fatalf("switch results %+v do not report %s (aggregate error %v)", results, *want, err)
}

func assertManagedEntry(t *testing.T, path, want string, storeRoot string) {
	t.Helper()
	info, err := os.Lstat(path)
	if want == "unchanged" && os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatalf("lstat managed entry %s: %v", path, err)
	}
	switch want {
	case "managed-file":
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			t.Fatalf("entry %s mode = %s, want a regular managed file", path, info.Mode())
		}
	case "managed-link":
		if info.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("entry %s mode = %s, want a managed link", path, info.Mode())
		}
		linkText, err := os.Readlink(path)
		if err != nil || !sameStoreTree(linkText, storeRoot) {
			t.Fatalf("entry %s link = %q (%v), not into store %s", path, linkText, err, storeRoot)
		}
	case "unchanged":
		// The operation-specific assertions below also verify the target
		// link text or bytes for each unchanged row.
	default:
		t.Fatalf("unknown vector entry_after value %q", want)
	}
}

func setupVectorSwitch(t *testing.T, fixtures map[string]writeNofollowFixture) (home, native string) {
	t.Helper()
	home = t.TempDir()
	pinHomes(t)
	source := t.TempDir()
	writePackage(t, source, "acme", "1.0.0", string(fixtureBytes(t, fixtures, "managed-root-context")))
	if _, _, _, err := Install(home, InstallOptions{Operand: source}); err != nil {
		t.Fatalf("install vector profile: %v", err)
	}
	native = claudeHome(t)
	return home, native
}

func runSwitchSymlinkTakeoverCase(t *testing.T, tc writeNofollowCase, fixtures map[string]writeNofollowFixture) {
	home, native := setupVectorSwitch(t, fixtures)
	foreign := writeExternalTarget(t, fixtures)
	target := filepath.Join(native, "CLAUDE.md")
	if err := os.MkdirAll(native, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(foreign, target); err != nil {
		t.Skipf("symlink fixture unavailable: %v", err)
	}
	results, err := UseWithPolicy(home, "acme", "claude_code", "", false, Policy{Takeover: tc.TakeoverAuthorized})
	assertSwitchDiagnostic(t, err, results, tc.Expected.Diagnostic)
	if tc.Expected.Outcome == "replaced" {
		assertManagedEntry(t, target, tc.Expected.EntryAfter, contextstore.Root(home))
		backup := filepath.Join(native, ".agent-environment-backup", "1", "CLAUDE.md")
		info, err := os.Lstat(backup)
		if err != nil || info.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("takeover backup is not the original symlink: mode=%v err=%v", info, err)
		}
		if got, err := os.Readlink(backup); err != nil || got != foreign {
			t.Fatalf("takeover backup link = %q, %v; want %q", got, err, foreign)
		}
	} else {
		if info, err := os.Lstat(target); err != nil || info.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("refused takeover changed target entry: mode=%v err=%v", info, err)
		}
		if _, err := os.Lstat(filepath.Join(native, ".agent-environment-backup")); !os.IsNotExist(err) {
			t.Fatalf("refused takeover created a backup: %v", err)
		}
	}
	assertForeignHash(t, foreign, tc.Expected.ForeignTargetSHA256After)
}

func runResolveSymlinkParentCase(t *testing.T, tc writeNofollowCase, fixtures map[string]writeNofollowFixture) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	request := fx.request("codex_cli")
	request.Repair = true
	request.Policy.Takeover = tc.TakeoverAuthorized
	homeDir := ManagedHomeDir(fx.home, "acme", "codex_cli")
	parent := filepath.Join(homeDir, ".agent-context")
	outside := t.TempDir()
	foreign := writeExternalTarget(t, fixtures)
	if tc.Target.Kind == "symlink" {
		if err := os.Symlink(foreign, filepath.Join(outside, "system-prompt.md")); err != nil {
			t.Skipf("symlink fixture unavailable: %v", err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(parent), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, parent); err != nil {
		t.Skipf("symlink fixture unavailable: %v", err)
	}
	_, err := Resolve(request)
	assertExpectedDiagnostic(t, err, tc.Expected.Diagnostic)
	if got, err := os.Readlink(parent); err != nil || got != outside {
		t.Fatalf("managed parent entry = %q (%v), want original symlink to %q", got, err, outside)
	}
	assertForeignHash(t, foreign, tc.Expected.ForeignTargetSHA256After)
	if tc.Expected.BackupHoldsLink {
		t.Fatal("parent-link refusal unexpectedly backed up the target")
	}
	if _, err := os.Lstat(filepath.Join(homeDir, ".agent-environment-backup")); !os.IsNotExist(err) {
		t.Fatalf("parent-link refusal created a backup: %v", err)
	}
}

func runResolvePlantedLinkRepairCase(t *testing.T, tc writeNofollowCase, fixtures map[string]writeNofollowFixture) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	request := fx.request("claude_code")
	request.Repair = true
	if _, err := Resolve(request); err != nil {
		t.Fatalf("initial materialization: %v", err)
	}
	marker := readManagedMarker(t, fx, "claude_code")
	path := marker.Surfaces[envmarker.SurfaceRootContext].Paths[0]
	target := filepath.Join(ManagedHomeDir(fx.home, "acme", "claude_code"), filepath.FromSlash(path))
	foreign := writeExternalTarget(t, fixtures)
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(foreign, target); err != nil {
		t.Skipf("symlink fixture unavailable: %v", err)
	}
	_, err := Resolve(request)
	assertExpectedDiagnostic(t, err, tc.Expected.Diagnostic)
	assertManagedEntry(t, target, tc.Expected.EntryAfter, contextstore.Root(fx.home))
	assertForeignHash(t, foreign, tc.Expected.ForeignTargetSHA256After)
	if tc.Expected.BackupHoldsLink {
		t.Fatal("repair unexpectedly backed up the planted link")
	}
}

func runResolveOwnedLinkRepairCase(t *testing.T, tc writeNofollowCase) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	request := fx.request("codex_cli")
	request.Repair = true
	if _, err := Resolve(request); err != nil {
		t.Fatalf("initial materialization: %v", err)
	}
	homeDir := ManagedHomeDir(fx.home, "acme", "codex_cli")
	link := filepath.Join(homeDir, ".agent-context", "system-prompt.md")
	linkText, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("system prompt surface is not linked: %v", err)
	}
	rootContext := filepath.Join(homeDir, "AGENTS.md")
	if err := os.Remove(rootContext); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(rootContext, []byte("drift\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = Resolve(request)
	assertExpectedDiagnostic(t, err, tc.Expected.Diagnostic)
	assertManagedEntry(t, link, tc.Expected.EntryAfter, ProfilesDir(fx.home))
	if got, err := os.Readlink(link); err != nil || got != linkText {
		t.Fatalf("repaired link = %q, %v; want %q", got, err, linkText)
	}
	if tc.Expected.BackupHoldsLink {
		t.Fatal("manager-owned link repair unexpectedly backed up the link")
	}
}

func runSwitchBackupParentCase(t *testing.T, tc writeNofollowCase, fixtures map[string]writeNofollowFixture) {
	home, native := setupVectorSwitch(t, fixtures)
	target := filepath.Join(native, "CLAUDE.md")
	original := fixtureBytes(t, fixtures, "foreign-notes")
	if err := os.MkdirAll(native, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, original, 0o644); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	backupRoot := filepath.Join(native, ".agent-environment-backup")
	if err := os.Symlink(outside, backupRoot); err != nil {
		t.Skipf("symlink fixture unavailable: %v", err)
	}
	results, err := UseWithPolicy(home, "acme", "claude_code", "", false, Policy{Takeover: true})
	assertSwitchDiagnostic(t, err, results, tc.Expected.Diagnostic)
	if got, err := os.ReadFile(target); err != nil || string(got) != string(original) {
		t.Fatalf("backup refusal changed the takeover target: %q (%v)", got, err)
	}
	entries, err := os.ReadDir(outside)
	if err != nil || len(entries) != 0 {
		t.Fatalf("backup path symlink received writes: entries=%v err=%v", entries, err)
	}
	assertForeignHash(t, target, tc.Expected.ForeignTargetSHA256After)
}

func runResolveCleanMaterializeCase(t *testing.T, tc writeNofollowCase) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	request := fx.request("claude_code")
	request.Repair = true
	_, err := Resolve(request)
	assertExpectedDiagnostic(t, err, tc.Expected.Diagnostic)
	marker := readManagedMarker(t, fx, "claude_code")
	path := marker.Surfaces[envmarker.SurfaceRootContext].Paths[0]
	target := filepath.Join(ManagedHomeDir(fx.home, "acme", "claude_code"), filepath.FromSlash(path))
	assertManagedEntry(t, target, tc.Expected.EntryAfter, contextstore.Root(fx.home))
}

func runSwitchInsideStoreLinkCase(t *testing.T, tc writeNofollowCase, fixtures map[string]writeNofollowFixture) {
	home, native := setupVectorSwitch(t, fixtures)
	storeTarget := filepath.Join(contextstore.Root(home), "vector", "target.md")
	if err := os.MkdirAll(filepath.Dir(storeTarget), 0o755); err != nil {
		t.Fatal(err)
	}
	foreign := fixtureBytes(t, fixtures, "foreign-notes")
	if err := os.WriteFile(storeTarget, foreign, 0o644); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(native, "CLAUDE.md")
	if err := os.MkdirAll(native, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(storeTarget, target); err != nil {
		t.Skipf("symlink fixture unavailable: %v", err)
	}
	results, err := UseWithPolicy(home, "acme", "claude_code", "", false, Policy{})
	assertSwitchDiagnostic(t, err, results, tc.Expected.Diagnostic)
	if got, err := os.Readlink(target); err != nil || got != storeTarget {
		t.Fatalf("refused inside-store link changed to %q (%v)", got, err)
	}
	assertForeignHash(t, storeTarget, tc.Expected.ForeignTargetSHA256After)
}

func runResolveRecordedFileRepairCase(t *testing.T, tc writeNofollowCase) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	request := fx.request("claude_code")
	request.Repair = true
	if _, err := Resolve(request); err != nil {
		t.Fatalf("initial materialization: %v", err)
	}
	marker := readManagedMarker(t, fx, "claude_code")
	path := marker.Surfaces[envmarker.SurfaceRootContext].Paths[0]
	target := filepath.Join(ManagedHomeDir(fx.home, "acme", "claude_code"), filepath.FromSlash(path))
	if err := os.WriteFile(target, []byte("operator drift\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Resolve(request)
	assertExpectedDiagnostic(t, err, tc.Expected.Diagnostic)
	assertManagedEntry(t, target, tc.Expected.EntryAfter, contextstore.Root(fx.home))
	if tc.Expected.BackupHoldsLink {
		t.Fatal("regular-file repair unexpectedly backed up a symlink")
	}
}
