package envprofile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/pathboundary"
)

const storeBoundaryVectorFile = "environments-store-boundary.json"

type storeBoundaryVectors struct {
	Resolve []storeBoundaryVector `json:"resolve_cases"`
	DryRun  []storeBoundaryVector `json:"dry_run_cases"`
	Repair  []storeBoundaryVector `json:"repair_cases"`
	Status  []storeBoundaryVector `json:"status_cases"`
}

type storeBoundaryVector struct {
	Name                 string `json:"name"`
	Object               string `json:"object"`
	Home                 string `json:"home"`
	Surface              string `json:"surface"`
	EntryKind            string `json:"entry_kind"`
	Diagnostic           string `json:"diagnostic"`
	FailingCheck         string `json:"failing_check"`
	NamesFailingCheck    string `json:"names_failing_check"`
	Outcome              string `json:"outcome"`
	Conforming           *bool  `json:"conforming"`
	Ownership            *bool  `json:"ownership"`
	FragmentEmitted      bool   `json:"fragment_emitted"`
	RowCurrent           bool   `json:"row_current"`
	RebuiltFromSnapshot  bool   `json:"rebuilt_from_snapshot"`
	ReappliedBeforeTrust bool   `json:"reapplied_before_trust"`
	Mutated              bool   `json:"mutated"`
}

func readStoreBoundaryVectors(t *testing.T) storeBoundaryVectors {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", storeBoundaryVectorFile)) // #nosec G304 -- explicit pinned conformance input
	if err != nil {
		t.Fatal(err)
	}
	var vectors storeBoundaryVectors
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	return vectors
}

func TestResolveRepairKeepsEnclosingStoreRootFailureClass(t *testing.T) {
	fx := storeBoundaryFixture(t, "current", true)
	storeRoot := contextstore.Root(fx.home)
	before := hashTreeForTest(t, fx.home)
	req := fx.request(envregistry.CodexCLI)
	req.Repair = true
	req.boundaryOwnerLookup = storeBoundaryOwnerLookup(storeRoot, false)

	result, err := Resolve(req)
	if err == nil || !strings.Contains(err.Error(), "profile store root failed ownership check:") {
		t.Fatalf("Resolve repair = (%v, %v), want the enclosing profile-store failure class", result, err)
	}
	if strings.Contains(err.Error(), DiagRepairFailed) || strings.Contains(err.Error(), envregistry.DiagWouldRebuildUntrustedStore) {
		t.Fatalf("enclosing profile-store failure entered the entry-rebuild path: %v", err)
	}
	if result != nil && len(result.Document) != 0 {
		t.Fatalf("untrusted Resolve repair emitted fragment %s", result.Document)
	}
	if after := hashTreeForTest(t, fx.home); after != before {
		t.Fatal("enclosing profile-store refusal mutated manager state")
	}
}

func TestStoreBoundaryResolveVectorsDriveResolve(t *testing.T) {
	vectors := readStoreBoundaryVectors(t)
	conformancecoverage.RunOutcomes(t, "environments-store-boundary/resolve_cases", vectors.Resolve,
		func(tc storeBoundaryVector) string { return tc.Name }, func(caseT *testing.T, tc storeBoundaryVector) conformancecoverage.Observation {
			return driveStoreBoundaryResolve(caseT, tc)
		})
}

func TestStoreBoundaryDryRunVectorsDriveResolveRepairPlan(t *testing.T) {
	vectors := readStoreBoundaryVectors(t)
	conformancecoverage.RunOutcomes(t, "environments-store-boundary/dry_run_cases", vectors.DryRun,
		func(tc storeBoundaryVector) string { return tc.Name }, func(caseT *testing.T, tc storeBoundaryVector) conformancecoverage.Observation {
			return driveStoreBoundaryDryRun(caseT, tc)
		})
}

func TestStoreBoundaryRepairVectorsDriveResolveRepair(t *testing.T) {
	vectors := readStoreBoundaryVectors(t)
	conformancecoverage.RunOutcomes(t, "environments-store-boundary/repair_cases", vectors.Repair,
		func(tc storeBoundaryVector) string { return tc.Name }, func(caseT *testing.T, tc storeBoundaryVector) conformancecoverage.Observation {
			return driveStoreBoundaryRepair(caseT, tc)
		})
}

func TestStoreBoundaryStatusVectorsDriveStatus(t *testing.T) {
	vectors := readStoreBoundaryVectors(t)
	conformancecoverage.RunOutcomes(t, "environments-store-boundary/status_cases", vectors.Status,
		func(tc storeBoundaryVector) string { return tc.Name }, func(caseT *testing.T, tc storeBoundaryVector) conformancecoverage.Observation {
			return driveStoreBoundaryStatus(caseT, tc)
		})
}

func driveStoreBoundaryResolve(t *testing.T, tc storeBoundaryVector) conformancecoverage.Observation {
	t.Helper()
	fx := storeBoundaryFixture(t, tc.Home, false)
	if tc.Home == "stale-old-marker" {
		changeStoreBoundaryLock(t, fx)
	}
	ownerPath := storeBoundaryOwnerPath(fx, tc)
	owner := pathboundary.OwnerLookup(nil)
	if ownerPath != "" && storeBoundaryNeedsOwnerFailure(tc) {
		owner = storeBoundaryOwnerLookup(ownerPath, false)
	}
	if reason := mutateStoreBoundaryVector(t, fx, tc); reason != "" {
		return conformancecoverage.Observation{BoundReason: reason}
	}
	req := fx.request(envregistry.CodexCLI)
	req.boundaryOwnerLookup = owner
	if tc.Name == "unreadable-marker-non-current" {
		marker := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI), envmarker.Name)
		req.readStateFile = vectorStateFileReader(marker, os.ErrPermission)
	}
	result, err := Resolve(req)
	if tc.Name == "intact-resolve-emits-fragment" {
		if err != nil || result == nil || len(result.Document) == 0 {
			t.Fatalf("Resolve = (%v, %v), want a current fragment", result, err)
		}
		return conformancecoverage.Observation{}
	}
	wantDiagnostic := tc.Diagnostic
	if wantDiagnostic == "" {
		wantDiagnostic = envregistry.DiagStoreUntrusted
	}
	if err == nil || !strings.Contains(err.Error(), wantDiagnostic) {
		t.Fatalf("Resolve error = %v, want %s", err, wantDiagnostic)
	}
	if result != nil && len(result.Document) != 0 {
		t.Fatalf("untrusted resolve emitted fragment %s", result.Document)
	}
	if tc.FailingCheck != "" && tc.FailingCheck != "home_currency" && !strings.Contains(err.Error(), "failed "+tc.FailingCheck+" check:") {
		t.Fatalf("Resolve error %q does not name failing check %q", err, tc.FailingCheck)
	}
	return conformancecoverage.Observation{}
}

func driveStoreBoundaryDryRun(t *testing.T, tc storeBoundaryVector) conformancecoverage.Observation {
	t.Helper()
	fx := storeBoundaryFixture(t, tc.Home, true)
	ownerPath := storeBoundaryOwnerPath(fx, tc)
	owner := pathboundary.OwnerLookup(nil)
	if ownerPath != "" && storeBoundaryNeedsOwnerFailure(tc) {
		owner = storeBoundaryOwnerLookup(ownerPath, false)
	}
	if reason := mutateStoreBoundaryVector(t, fx, tc); reason != "" {
		return conformancecoverage.Observation{BoundReason: reason}
	}
	_, entry := storeBoundaryRootMemberFor(fx)
	entryBefore, err := os.Lstat(entry)
	if err != nil {
		t.Fatalf("lstat store entry before dry-run: %v", err)
	}
	before := hashTreeForTest(t, fx.home)
	req := fx.request(envregistry.CodexCLI)
	req.Repair = true
	req.DryRun = true
	req.boundaryOwnerLookup = owner
	result, resolveErr := Resolve(req)
	after := hashTreeForTest(t, fx.home)
	if before != after {
		t.Fatalf("dry-run repair mutated manager state for %s", tc.Name)
	}
	entryAfter, statErr := os.Lstat(entry)
	if statErr != nil || !os.SameFile(entryBefore, entryAfter) {
		t.Fatalf("dry-run repair replaced store entry for %s: before=%v after=%v err=%v", tc.Name, entryBefore, entryAfter, statErr)
	}
	switch tc.Name {
	case "dry-run-intact-plans-nothing":
		if resolveErr != nil || result == nil || len(result.Document) == 0 {
			t.Fatalf("intact Resolve = (%v, %v), want a current fragment", result, resolveErr)
		}
	case "dry-run-enclosing-no-rebuild":
		if resolveErr == nil || (result != nil && len(result.Document) != 0) || !strings.Contains(resolveErr.Error(), envregistry.DiagStoreUntrusted) {
			t.Fatalf("dry-run repair = (%v, %v), want fail-closed %s", result, resolveErr, envregistry.DiagStoreUntrusted)
		}
	case "dry-run-untrusted-reports-would-rebuild", "dry-run-mutates":
		if resolveErr == nil || (result != nil && len(result.Document) != 0) || !strings.Contains(resolveErr.Error(), envregistry.DiagWouldRebuildUntrustedStore) {
			t.Fatalf("dry-run repair = (%v, %v), want %s", result, resolveErr, envregistry.DiagWouldRebuildUntrustedStore)
		}
	default:
		t.Fatalf("unknown dry-run vector %q", tc.Name)
	}
	return conformancecoverage.Observation{}
}

func driveStoreBoundaryRepair(t *testing.T, tc storeBoundaryVector) conformancecoverage.Observation {
	t.Helper()
	gitRoot := tc.EntryKind == "git"
	fx := storeBoundaryFixture(t, tc.Home, gitRoot)
	if tc.Home == "unprovisioned" {
		seedLiveNativeCredentials(t, fx)
	}
	if tc.Home == "stale-old-marker" {
		changeStoreBoundaryLock(t, fx)
	}
	ownerPath := storeBoundaryOwnerPath(fx, tc)
	owner := pathboundary.OwnerLookup(nil)
	if tc.Name == "repair-path-entry-cannot-rebuild" {
		ownerPath = storeBoundaryEntry(t, fx, contextlock.KindContext)
	}
	if ownerPath != "" && storeBoundaryNeedsOwnerFailure(tc) {
		oneShot := strings.Contains(tc.Name, "rebuilds-entry-boundary") || tc.Name == "repair-reapplies-untrusted"
		owner = storeBoundaryOwnerLookup(ownerPath, oneShot)
	}
	if reason := mutateStoreBoundaryVector(t, fx, tc); reason != "" {
		return conformancecoverage.Observation{BoundReason: reason}
	}
	beforeState := hashTreeForTest(t, fx.home)
	req := fx.request(envregistry.CodexCLI)
	req.Repair = true
	req.boundaryOwnerLookup = owner
	result, err := Resolve(req)
	switch tc.Name {
	case "repair-path-entry-cannot-rebuild", "repair-local-entry-cannot-rebuild":
		if err == nil || (result != nil && len(result.Document) != 0) || !strings.Contains(err.Error(), DiagRepairFailed) {
			t.Fatalf("repair Resolve = (%v, %v), want no fragment and %s", result, err, DiagRepairFailed)
		}
	case "repair-enclosing-refuses-no-rebuild":
		if err == nil || (result != nil && len(result.Document) != 0) || !strings.Contains(err.Error(), envregistry.DiagStoreUntrusted) {
			t.Fatalf("enclosing-boundary repair = (%v, %v), want refusal %s", result, err, envregistry.DiagStoreUntrusted)
		}
		if after := hashTreeForTest(t, fx.home); after != beforeState {
			t.Fatal("enclosing-boundary refusal mutated manager state")
		}
	case "repair-rebuilds-git-entry-from-snapshot", "repair-rebuilds-entry-boundary-failure", "repair-stale-old-marker-succeeds", "repair-swapped-old-marker-never-adopted", "repair-unprovisioned-intact-provisions", "repair-unprovisioned-swapped-rebuilds", "repair-reapplies-untrusted":
		if err != nil || result == nil || len(result.Document) == 0 {
			t.Fatalf("repair Resolve = (%v, %v), want a trusted fragment", result, err)
		}
		if tc.RebuiltFromSnapshot {
			root, _ := storeBoundaryRootMember(t, fx)
			if root == nil {
				t.Fatal("repaired profile lost its context member")
			}
			lock, _, err := readLock(fx.home, fx.profile)
			if err != nil {
				t.Fatal(err)
			}
			if failure := validateNamedStorePins(fx.home, lock); failure != nil {
				t.Fatalf("rebuilt store entry did not match its pin: %v", failure)
			}
		}
	default:
		t.Fatalf("unknown repair vector %q", tc.Name)
	}
	return conformancecoverage.Observation{}
}

func driveStoreBoundaryStatus(t *testing.T, tc storeBoundaryVector) conformancecoverage.Observation {
	t.Helper()
	fx := storeBoundaryFixture(t, tc.Home, false)
	provisionStoreBoundaryHome(t, fx)
	ownerPath := storeBoundaryOwnerPath(fx, tc)
	owner := pathboundary.OwnerLookup(nil)
	if ownerPath != "" && storeBoundaryNeedsOwnerFailure(tc) {
		owner = storeBoundaryOwnerLookup(ownerPath, false)
	}
	if reason := mutateStoreBoundaryVector(t, fx, tc); reason != "" {
		return conformancecoverage.Observation{BoundReason: reason}
	}
	req := statusRequest(fx)
	req.boundaryOwnerLookup = owner
	status, err := StatusOf(req)
	if err != nil {
		t.Fatalf("StatusOf: %v", err)
	}
	var row *HomeState
	for i := range status.Homes {
		if status.Homes[i].Profile == fx.profile && status.Homes[i].Environment == envregistry.CodexCLI {
			row = &status.Homes[i]
			break
		}
	}
	if row == nil {
		t.Fatalf("StatusOf has no %s/%s row", fx.profile, envregistry.CodexCLI)
	}
	if tc.Name == "status-intact-current" {
		if !row.Current || !row.Provisioned {
			t.Fatalf("intact status row = %+v, want current and provisioned", *row)
		}
		return conformancecoverage.Observation{}
	}
	if row.Current || !strings.Contains(strings.Join(row.Findings, " "), envregistry.DiagStoreUntrusted) {
		t.Fatalf("untrusted status row = %+v, want non-current with %s", *row, envregistry.DiagStoreUntrusted)
	}
	wantCheck := tc.NamesFailingCheck
	if wantCheck == "" {
		wantCheck = tc.FailingCheck
	}
	if wantCheck != "" && !strings.Contains(strings.Join(row.Findings, " "), wantCheck) {
		t.Fatalf("status findings %q do not name check %q", strings.Join(row.Findings, " "), wantCheck)
	}
	return conformancecoverage.Observation{}
}

func storeBoundaryFixture(t *testing.T, homeState string, gitRoot bool) *managedFixture {
	t.Helper()
	fx := writeManagedFixture(t, "acme")
	if gitRoot {
		convertStoreBoundaryRootToGit(t, fx)
	}
	if homeState != "unprovisioned" {
		provisionStoreBoundaryHome(t, fx)
	}
	return fx
}

func provisionStoreBoundaryHome(t *testing.T, fx *managedFixture) {
	t.Helper()
	seedLiveNativeCredentials(t, fx)
	req := fx.request(envregistry.CodexCLI)
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatalf("provision store-boundary fixture: %v", err)
	}
}

func convertStoreBoundaryRootToGit(t *testing.T, fx *managedFixture) {
	t.Helper()
	lock, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	_, ok := lock.Find(contextlock.KindContext, fx.profile)
	if !ok {
		t.Fatal("fixture has no root context member")
	}
	source := "github.com/example/context-" + fx.profile
	commit := writeGitStoreFixture(t, fx.home, contextlock.KindContext, fx.profile, source, map[string]string{
		"agent-context.json": `{"schema_version": 1, "name": "` + fx.profile + `", "version": "1.0.0", "context": {"modules": [{"path": "a.md"}, {"path": "s.md", "class": "system"}]}}`,
		"context/a.md":       "hello\n",
		"context/s.md":       "Be terse.\n",
	})
	for i := range lock.Members {
		if lock.Members[i].Kind == contextlock.KindContext && lock.Members[i].Name == fx.profile {
			lock.Members[i].Source = source
			lock.Members[i].Commit = commit
			lock.Members[i].StateHash = ""
		}
	}
	lock.Sort()
	hash, err := contextlock.Write(lockPath(fx.home, fx.profile), lock)
	if err != nil {
		t.Fatal(err)
	}
	fx.lockHash = hash
	payload, err := json.Marshal(Source{Kind: KindGit, Git: source, Req: Requirement{Revision: commit}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath(fx.home, fx.profile), append(payload, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

func changeStoreBoundaryLock(t *testing.T, fx *managedFixture) {
	t.Helper()
	lock, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	for i := range lock.Members {
		if lock.Members[i].Kind == contextlock.KindContext && lock.Members[i].Name == fx.profile {
			lock.Members[i].Weight++
		}
	}
	hash, err := contextlock.Write(lockPath(fx.home, fx.profile), lock)
	if err != nil {
		t.Fatal(err)
	}
	fx.lockHash = hash
}

func storeBoundaryRootMember(t *testing.T, fx *managedFixture) (*contextlock.Member, string) {
	t.Helper()
	lock, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find(contextlock.KindContext, fx.profile)
	if !ok {
		return nil, ""
	}
	return &member, contextstore.EntryDir(fx.home, member.Kind, member.Name, member.PinKey())
}

func storeBoundaryEntry(t *testing.T, fx *managedFixture, kind string) string {
	t.Helper()
	lock, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	for _, member := range lock.Members {
		if member.Kind == kind {
			return contextstore.EntryDir(fx.home, member.Kind, member.Name, member.PinKey())
		}
	}
	t.Fatalf("fixture has no %s member", kind)
	return ""
}

func storeBoundaryOwnerPath(fx *managedFixture, tc storeBoundaryVector) string {
	switch tc.Object {
	case "store-entry":
		_, entry := storeBoundaryRootMemberFor(fx)
		return entry
	case "lock-file":
		return lockPath(fx.home, fx.profile)
	case "marker-file":
		return filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI), envmarker.Name)
	case "environments-root":
		return EnvRoot(fx.home)
	case "store-root":
		return contextstore.Root(fx.home)
	default:
		return ""
	}
}

func storeBoundaryNeedsOwnerFailure(tc storeBoundaryVector) bool {
	return tc.Ownership != nil && !*tc.Ownership
}

func storeBoundaryRootMemberFor(fx *managedFixture) (*contextlock.Member, string) {
	lock, _, err := contextlock.Read(lockPath(fx.home, fx.profile))
	if err != nil {
		return nil, ""
	}
	member, ok := lock.Find(contextlock.KindContext, fx.profile)
	if !ok {
		return nil, ""
	}
	return &member, contextstore.EntryDir(fx.home, member.Kind, member.Name, member.PinKey())
}

func storeBoundaryOwnerLookup(target string, once bool) pathboundary.OwnerLookup {
	defaultLookup := pathboundary.DefaultOwnerLookup()
	used := false
	return func(path string, info os.FileInfo) (pathboundary.OwnerIdentity, error) {
		if filepath.Clean(path) == filepath.Clean(target) && (!once || !used) {
			used = true
			return pathboundary.OwnerIdentity("foreign-owner"), nil
		}
		return defaultLookup(path, info)
	}
}

func mutateStoreBoundaryVector(t *testing.T, fx *managedFixture, tc storeBoundaryVector) string {
	t.Helper()
	member, entry := storeBoundaryRootMemberFor(fx)
	if member == nil {
		t.Fatal("could not read root context member")
	}
	switch tc.Name {
	case "swapped-system-prompt-bytes-untrusted", "swapped-updated-store-old-marker-untrusted", "swapped-bytes-emits-fragment":
		return writeSwappedStoreSurface(t, entry, "context/s.md")
	case "swapped-root-context-bytes-untrusted", "unprovisioned-swapped-untrusted":
		return writeSwappedStoreSurface(t, entry, "context/a.md")
	case "repair-rebuilds-git-entry-from-snapshot", "repair-swapped-old-marker-never-adopted":
		return writeSwappedStoreSurface(t, entry, "context/s.md")
	case "repair-unprovisioned-swapped-rebuilds":
		return writeSwappedStoreSurface(t, entry, "context/a.md")
	case "symlinked-entry-root-untrusted", "untrusted-without-diagnostic":
		other := storeBoundaryEntry(t, fx, contextlock.KindMCP)
		return replaceWithSymlink(t, entry, other)
	case "wrong-permissions-untrusted":
		return makeStoreBoundaryPermissionsFailure(t, entry)
	case "repair-local-entry-cannot-rebuild":
		return makeStoreBoundaryPermissionsFailure(t, entry)
	case "containment-escape-untrusted":
		outside := filepath.Join(t.TempDir(), "outside.md")
		if err := os.WriteFile(outside, []byte("outside\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		return replaceWithSymlink(t, filepath.Join(entry, "context", "a.md"), outside)
	case "non-regular-component-untrusted":
		if err := makeFIFOVectorForTest(filepath.Join(entry, "context", "pipe")); err != nil {
			return "host cannot create the non-regular-file vector: " + err.Error()
		}
	case "marker-symlink-untrusted":
		marker := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI), envmarker.Name)
		return replaceWithSymlink(t, marker, marker+".saved")
	case "store-root-symlinked-untrusted":
		root := contextstore.Root(fx.home)
		saved := root + ".saved"
		if err := os.Rename(root, saved); err != nil {
			t.Fatalf("move store root for symlink vector: %v", err)
		}
		if err := os.Symlink(saved, root); err != nil {
			return "host cannot create the store-root symlink: " + err.Error()
		}
	case "intact-resolve-emits-fragment", "intact-updated-store-old-marker-stale", "unprovisioned-intact-stale", "unreadable-marker-non-current",
		"wrong-ownership-untrusted", "lock-file-wrong-owner-untrusted", "environments-root-wrong-owner-untrusted", "untrusted-reported-current",
		"dry-run-untrusted-reports-would-rebuild", "dry-run-intact-plans-nothing", "dry-run-enclosing-no-rebuild", "dry-run-mutates",
		"repair-path-entry-cannot-rebuild", "repair-rebuilds-entry-boundary-failure", "repair-enclosing-refuses-no-rebuild",
		"repair-stale-old-marker-succeeds", "repair-unprovisioned-intact-provisions", "repair-reapplies-untrusted",
		"status-intact-current", "status-names-failing-check", "status-enclosing-names-boundary", "status-hides-failing-check":
		// These rows are driven by the injected owner lookup or by their
		// intact/stale fixture state; no filesystem mutation is required.
	default:
		t.Fatalf("no store-boundary mutator for vector %q", tc.Name)
	}
	return ""
}

func makeStoreBoundaryPermissionsFailure(t *testing.T, path string) string {
	t.Helper()
	restore, err := makeWorldWritableDirectoryForTest(path)
	if err != nil {
		return "host cannot create the private-permission failure: " + err.Error()
	}
	t.Cleanup(func() {
		if err := restore(); err != nil {
			t.Errorf("restore store entry permissions: %v", err)
		}
	})
	return ""
}

func writeSwappedStoreSurface(t *testing.T, entry, relative string) string {
	t.Helper()
	path := filepath.Join(entry, filepath.FromSlash(relative))
	if err := os.WriteFile(path, []byte("swapped bytes\n"), 0o600); err != nil {
		t.Fatalf("swap store bytes at %s: %v", path, err)
	}
	return ""
}

func replaceWithSymlink(t *testing.T, path, target string) string {
	t.Helper()
	if err := os.Rename(path, path+".saved"); err != nil {
		t.Fatalf("move %s for symlink vector: %v", path, err)
	}
	if err := os.Symlink(target, path); err != nil {
		return "host cannot create the symlink needed by the vector: " + err.Error()
	}
	return ""
}

// TestStoreBoundaryCheckedRootsExcludeCuratorHome pins the checked-root
// scope of environments §4: the Curator home and the profiles directory are
// operator directories (inherited ACEs on Windows), so a foreign owner or
// shared mutation permissions there must not refuse env resolve, while the
// same state on the environments root or profile store root must.
func TestStoreBoundaryCheckedRootsExcludeCuratorHome(t *testing.T) {
	fx := storeBoundaryFixture(t, "current", false)
	for _, operatorDir := range []string{fx.home, ProfilesDir(fx.home), ProfileDir(fx.home, fx.profile)} {
		req := fx.request(envregistry.CodexCLI)
		req.boundaryOwnerLookup = storeBoundaryOwnerLookup(operatorDir, false)
		if _, err := Resolve(req); err != nil {
			t.Fatalf("Resolve with foreign-owned operator dir %s = %v, want success", operatorDir, err)
		}
	}
	if restore, err := makeWorldWritableSingleDirectoryForTest(fx.home); err != nil {
		t.Log("host cannot create the private-permission failure: " + err.Error())
	} else {
		t.Cleanup(func() {
			if err := restore(); err != nil {
				t.Errorf("restore Curator home permissions: %v", err)
			}
		})
		if _, err := Resolve(fx.request(envregistry.CodexCLI)); err != nil {
			t.Fatalf("Resolve with shared-permission Curator home = %v, want success", err)
		}
	}
	for _, root := range []string{EnvRoot(fx.home), contextstore.Root(fx.home)} {
		req := fx.request(envregistry.CodexCLI)
		req.boundaryOwnerLookup = storeBoundaryOwnerLookup(root, false)
		_, err := Resolve(req)
		if err == nil || !strings.Contains(err.Error(), envregistry.DiagStoreUntrusted) || !strings.Contains(err.Error(), "failed ownership check:") {
			t.Fatalf("Resolve with foreign-owned %s = %v, want %s ownership", root, err, envregistry.DiagStoreUntrusted)
		}
	}
}
