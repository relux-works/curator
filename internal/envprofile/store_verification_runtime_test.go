package envprofile

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/privatedir"
)

// These are fixture shapes only, preserving the predecessor's enclosing object,
// source, surface and home. Canonical verdicts come exclusively from revision 2.
//
//go:embed testdata/store-verification-v2/predecessor-fixtures.json
var storeVerificationFixtureShapes []byte

func canonicalStoreVerificationRows(t *testing.T) []storeVerificationCase {
	t.Helper()
	if err := storeVerificationCanonicalGuard(os.LookupEnv); err != nil {
		t.Fatal(err)
	}
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Fatal("canonical store runtime requires the SPEC_PIN root")
	}
	root, err := filepath.Abs(filepath.Join(root, "..", "environment-store-verification-v2"))
	if err != nil {
		t.Fatal(err)
	}
	rows, selected, err := readStoreVerificationCorpus(storeVerificationLookup(map[string]string{storeVerificationRootKey: root, storeVerificationPinKey: storeVerificationManifestPin}))
	if err != nil || !selected {
		t.Fatalf("canonical store corpus: %v", err)
	}
	return rows
}

// The four existing runtime families execute every predecessor successor and
// its declared level splits. Other rows remain metadata-only. Host inability
// to plant a required violation is a failure here, never a skip or bound pass.
func runStoreVerificationSuccessors(t *testing.T, group string) {
	t.Helper()
	rows := canonicalStoreVerificationRows(t)
	var shapes map[string]storeBoundaryVector
	if err := json.Unmarshal(storeVerificationFixtureShapes, &shapes); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, row := range rows {
		var name string
		if row.Predecessor != nil && row.Predecessor.Group == group {
			name = row.Predecessor.Name
		}
		if row.Predecessor == nil && strings.HasPrefix(row.Name, "level") {
			for _, prefix := range []string{"level1-legacy-", "level2-legacy-"} {
				if strings.HasPrefix(row.Name, prefix) {
					candidate := strings.TrimPrefix(row.Name, prefix)
					if _, ok := shapes[group+"/"+candidate]; ok {
						name = candidate
					}
				}
			}
		}
		if name == "" {
			continue
		}
		shape, ok := shapes[group+"/"+name]
		if !ok {
			t.Fatalf("missing preserved fixture for %s", row.Name)
		}
		count++
		t.Run(row.Name, func(t *testing.T) {
			for _, level := range row.Levels {
				t.Run(fmt.Sprintf("level%d", level), func(t *testing.T) { driveStoreVerificationSuccessor(t, row, shape, level) })
			}
		})
	}
	if count == 0 {
		t.Fatal("canonical family executed no successor rows")
	}
	t.Logf("revision-2 runtime rows=%d family=%s", count, group)
}

func storeVerificationTree(t *testing.T, path string) string {
	t.Helper()
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return "absent"
	} else if err != nil {
		t.Fatal(err)
	}
	return hashTreeForTest(t, path)
}

func driveStoreVerificationSuccessor(t *testing.T, row storeVerificationCase, shape storeBoundaryVector, level int) {
	t.Helper()
	fx := writeManagedFixture(t, "acme")
	if row.Source == "git" {
		convertStoreBoundaryRootToGit(t, fx)
	}
	if row.Source == "local" {
		raw, err := json.Marshal(Source{Kind: KindLocal})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(sourcePath(fx.home, fx.profile), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	source, err := readSource(fx.home, fx.profile)
	if err != nil || source.Kind != row.Source {
		t.Fatalf("v2 source fixture kind: %v %v", source, err)
	}
	if shape.Home != "unprovisioned" {
		provisionStoreBoundaryHome(t, fx)
	}
	if shape.Home == "unprovisioned" {
		seedLiveNativeCredentials(t, fx)
	}
	if shape.Home == "stale-old-marker" {
		changeStoreBoundaryLock(t, fx)
	}
	ownerPath := storeBoundaryOwnerPath(fx, shape)
	if shape.Name == "repair-path-entry-cannot-rebuild" {
		_, ownerPath = storeBoundaryRootMemberFor(fx)
	}
	owner := storeBoundaryOwnerLookup(ownerPath, false)
	if !storeBoundaryNeedsOwnerFailure(shape) {
		owner = nil
	}
	_, trustedEntry := storeBoundaryRootMemberFor(fx)
	trustedStore := storeVerificationTree(t, trustedEntry)
	if shape.Name == "non-regular-component-untrusted" {
		_, entry := storeBoundaryRootMemberFor(fx)
		plantStoreVerificationSocket(t, filepath.Join(entry, "context", "pipe"))
	} else if reason := mutateStoreBoundaryVector(t, fx, shape); reason != "" {
		t.Fatalf("required plant unavailable: %s", reason)
	}
	_, entry := storeBoundaryRootMemberFor(fx)
	entryBefore, err := storeVerificationFileInfo(entry)
	if err != nil {
		t.Fatal(err)
	}
	before := hashTreeForTest(t, fx.home)
	beforeNodes := storeVerificationNodes(t, fx.home)
	storeBefore := storeVerificationTree(t, entry)
	storeNodesBefore := storeVerificationNodes(t, entry)
	lockInfoBefore, err := storeVerificationFileInfo(lockPath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	lockBefore, err := os.ReadFile(lockPath(fx.home, fx.profile))
	if err != nil {
		t.Fatal(err)
	}
	homePath := ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI)
	homeBefore := storeVerificationTree(t, homePath)
	req := fx.request(envregistry.CodexCLI)
	req.Policy.StartupVerificationLevel = level
	req.Repair = row.Operation == "repair" || row.Operation == "dry-run"
	req.DryRun = row.Operation == "dry-run"
	req.boundaryOwnerLookup = owner
	if shape.Name == "unreadable-marker-non-current" {
		req.readStateFile = vectorStateFileReader(filepath.Join(homePath, envmarker.Name), os.ErrPermission)
	}
	var result *ResolveResult
	var operationErr error
	var current bool
	if row.Operation == "status" {
		request := statusRequest(fx)
		request.boundaryOwnerLookup = owner
		status, err := StatusOf(request)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, home := range status.Homes {
			if home.Profile == fx.profile && home.Environment == envregistry.CodexCLI {
				found = true
				current = home.Current
				if row.Expected.Diagnostic != "" && !strings.Contains(strings.Join(home.Findings, " "), row.Expected.Diagnostic) {
					t.Fatalf("v2 status diagnostic missing: %+v", home)
				}
				if row.Expected.PathNamed && !strings.Contains(strings.Join(home.Findings, " "), ownerPath) {
					t.Fatalf("v2 status offending path missing: %+v", home)
				}
				if row.Expected.Check != "" && !strings.Contains(strings.Join(home.Findings, " "), row.Expected.Check) {
					t.Fatalf("v2 status check missing: %+v", home)
				}
			}
		}
		if !found {
			t.Fatal("v2 status row missing")
		}
	} else {
		result, operationErr = Resolve(req)
		if row.Expected.Diagnostic == "" || row.Expected.Fragment {
			if operationErr != nil {
				t.Fatalf("v2 expected successful operation: %v", operationErr)
			}
		} else if operationErr == nil || !strings.HasPrefix(operationErr.Error(), row.Expected.Diagnostic+":") {
			t.Fatalf("v2 leading diagnostic: got %v want %s", operationErr, row.Expected.Diagnostic)
		}
		if row.Expected.Diagnostic != "" && row.Expected.Fragment {
			warnings := ""
			if result != nil {
				warnings = strings.Join(result.Warnings, ";")
			}
			if !strings.Contains(warnings, row.Expected.Diagnostic+":") {
				t.Fatalf("v2 warning diagnostic missing: %s", warnings)
			}
		}
		fragment := result != nil && len(result.Document) != 0
		if fragment != row.Expected.Fragment {
			t.Fatalf("v2 fragment=%v want %v", fragment, row.Expected.Fragment)
		}
		if row.Expected.Check != "" && row.Expected.Check != "home_currency" && (operationErr == nil || !strings.Contains(operationErr.Error(), "failed "+row.Expected.Check+" check:")) {
			t.Fatalf("v2 failing check %s missing: %v", row.Expected.Check, operationErr)
		}
		if row.Expected.WarningEmitted != nil {
			warning := result != nil && len(result.Warnings) != 0
			if warning != *row.Expected.WarningEmitted {
				t.Fatalf("v2 warning=%v want %v", warning, *row.Expected.WarningEmitted)
			}
		}
		if row.Expected.RepairHint != nil {
			detail := ""
			if operationErr != nil {
				detail = operationErr.Error()
			}
			if result != nil {
				detail += strings.Join(result.Warnings, ";")
			}
			if strings.Contains(detail, "--repair") != *row.Expected.RepairHint {
				t.Fatalf("v2 repair hint: %s", detail)
			}
		}
		statusReq := statusRequest(fx)
		statusReq.boundaryOwnerLookup = owner
		statusReq.readStateFile = req.readStateFile
		status, err := StatusOf(statusReq)
		if err == nil {
			for _, home := range status.Homes {
				if home.Profile == fx.profile && home.Environment == envregistry.CodexCLI {
					current = home.Current
				}
			}
		}
	}
	if current != row.Expected.Current {
		t.Fatalf("v2 current=%v want %v", current, row.Expected.Current)
	}
	lockAfter, err := os.ReadFile(lockPath(fx.home, fx.profile))
	if err != nil || string(lockAfter) != string(lockBefore) {
		t.Fatalf("v2 operation changed lock: %v", err)
	}
	lockInfoAfter, err := storeVerificationFileInfo(lockPath(fx.home, fx.profile))
	if err != nil || !os.SameFile(lockInfoBefore, lockInfoAfter) || !lockInfoBefore.ModTime().Equal(lockInfoAfter.ModTime()) {
		t.Fatalf("v2 operation replaced or touched lock: %v", err)
	}
	entryAfter, err := storeVerificationFileInfo(entry)
	if err != nil {
		t.Fatal(err)
	}
	rebuilt := !os.SameFile(entryBefore, entryAfter)
	t.Logf("store rebuild observation: rebuilt=%v expected=%v source=%s level=%d", rebuilt, row.Expected.Rebuilt, source.Kind, level)
	if rebuilt != row.Expected.Rebuilt {
		t.Fatalf("v2 rebuilt=%v want %v", rebuilt, row.Expected.Rebuilt)
	}
	if rebuilt && storeVerificationTree(t, entry) != trustedStore {
		t.Fatal("v2 rebuild did not restore verified snapshot bytes")
	}
	if len(row.Expected.Writes) == 0 && (hashTreeForTest(t, fx.home) != before || !sameStoreVerificationNodes(beforeNodes, storeVerificationNodes(t, fx.home))) {
		t.Fatal("v2 read-only/refusal mutated manager state")
	}
	if row.Expected.PlantPreserved && (!sameStoreVerificationNodes(storeNodesBefore, storeVerificationNodes(t, entry)) || storeVerificationTree(t, entry) != storeBefore) {
		t.Fatal("v2 operation erased structural plant")
	}
	if row.Expected.PathNamed && row.Operation != "status" {
		// Every predecessor chooses an enclosing object or a concrete payload. The
		// boundary diagnostic must identify that object and its declared rule.
		if operationErr == nil || !strings.Contains(operationErr.Error(), ownerPath) {
			t.Fatalf("v2 offending path missing: %v", operationErr)
		}
	}
	wantsHomeWrite := false
	for _, write := range row.Expected.Writes {
		if write == "home" {
			wantsHomeWrite = true
		}
	}
	if wantsHomeWrite && storeVerificationTree(t, homePath) == homeBefore {
		t.Fatal("v2 repair did not update derived home")
	}
}

// A real socket exercises regular_types on Unix and Windows. Bind at a short
// private path, then move the node to the predecessor's exact payload path;
// Windows and Darwin sockaddr paths cannot hold a deep managed-store path.
func plantStoreVerificationSocket(t *testing.T, path string) {
	t.Helper()
	dir, err := os.MkdirTemp(os.TempDir(), "sv2-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Error(err)
		}
	})
	if err := privatedir.Protect(dir); err != nil {
		t.Fatal(err)
	}
	socket := filepath.Join(dir, "s")
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: socket, Net: "unix"})
	if err != nil {
		t.Fatalf("required real socket plant: %v", err)
	}
	listener.SetUnlinkOnClose(false)
	t.Cleanup(func() {
		if err := listener.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := os.Rename(socket, path); err != nil {
		t.Fatalf("move required socket plant: %v", err)
	}
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSocket == 0 {
		t.Fatalf("special-file plant is not a real socket: %v %v", info, err)
	}
	t.Log("non-regular fixture kind=socket")
}

type storeVerificationNode struct {
	info   os.FileInfo
	target string
}

func storeVerificationNodes(t *testing.T, root string) map[string]storeVerificationNode {
	t.Helper()
	nodes := map[string]storeVerificationNode{}
	err := filepath.WalkDir(root, func(path string, _ os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := storeVerificationFileInfo(path)
		if err != nil {
			return err
		}
		target := ""
		if info.Mode()&os.ModeSymlink != 0 {
			target, err = os.Readlink(path)
			if err != nil {
				return err
			}
		}
		nodes[path] = storeVerificationNode{info, target}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return nodes
}
func sameStoreVerificationNodes(before, after map[string]storeVerificationNode) bool {
	if len(before) != len(after) {
		return false
	}
	for path, a := range before {
		b, ok := after[path]
		if !ok || a.target != b.target || !os.SameFile(a.info, b.info) || a.info.Mode() != b.info.Mode() || !a.info.ModTime().Equal(b.info.ModTime()) {
			return false
		}
	}
	return true
}
