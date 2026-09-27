package envprofile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
)

type readFailureVectorCase struct {
	Name         string `json:"name"`
	Operation    string `json:"operation"`
	FileClass    string `json:"file_class"`
	EntryKind    string `json:"entry_kind"`
	FailureClass string `json:"failure_class"`
	Conforming   *bool  `json:"conforming"`
	Expected     struct {
		Currency              string  `json:"currency"`
		Diagnostic            *string `json:"diagnostic"`
		FragmentEmitted       *bool   `json:"fragment_emitted"`
		RowCurrent            *bool   `json:"row_current"`
		Rebuilt               *bool   `json:"rebuilt"`
		Written               *bool   `json:"written"`
		ProvisioningContinues *bool   `json:"provisioning_continues"`
		RepairRelinks         *bool   `json:"repair_relinks"`
	} `json:"expected"`
}

func TestEnvironmentsReadFailureVectors(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	path := filepath.Join(root, "vectors", "environments-read-failure.json")
	payload, err := os.ReadFile(path) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var vectors struct {
		ProtocolVersion string                  `json:"protocol_version"`
		Cases           []readFailureVectorCase `json:"cases"`
	}
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	if vectors.ProtocolVersion != "1.0.0-rc.13" || len(vectors.Cases) != 39 {
		t.Fatalf("pinned read-failure vectors = protocol %q, %d cases; want rc.13 and 39 cases", vectors.ProtocolVersion, len(vectors.Cases))
	}
	var profileCases []readFailureVectorCase
	var restoreCases []string
	for _, tc := range vectors.Cases {
		if tc.Operation == "restore" {
			restoreCases = append(restoreCases, tc.Name)
			continue
		}
		profileCases = append(profileCases, tc)
	}
	wantRestoreCases := map[string]bool{
		"backup-record-absent-restore-nothing":   false,
		"backup-record-unreadable-restore-stops": false,
	}
	for _, name := range restoreCases {
		if _, ok := wantRestoreCases[name]; !ok {
			t.Fatalf("unexpected restore vector %q", name)
		}
		wantRestoreCases[name] = true
	}
	if len(profileCases) != 37 || len(restoreCases) != len(wantRestoreCases) {
		t.Fatalf("read-failure vector split = %d profile + %d restore; want 37 + 2", len(profileCases), len(restoreCases))
	}
	for name, seen := range wantRestoreCases {
		if !seen {
			t.Fatalf("restore vector %q is missing", name)
		}
	}
	conformancecoverage.RunOutcomes(t, "environments-read-failure/cases", profileCases,
		func(tc readFailureVectorCase) string { return tc.Name }, func(t *testing.T, tc readFailureVectorCase) conformancecoverage.Observation {
			switch tc.FileClass {
			case "marker":
				runMarkerReadFailureVector(t, tc)
			case "lock":
				runLockReadFailureVector(t, tc)
			case "seed":
				runSeedReadFailureVector(t, tc)
			case "passthrough":
				runPassthroughReadFailureVector(t, tc)
			case "backup-record":
				runBackupReadFailureVector(t, tc)
			default:
				t.Fatalf("unknown read-failure file class %q", tc.FileClass)
			}
			return conformancecoverage.Observation{}
		})
}

func vectorFailure(class string) error {
	switch class {
	case "permission-denied":
		return syscall.EACCES
	case "io-error":
		return syscall.EIO
	case "parent-not-directory":
		return syscall.ENOTDIR
	default:
		return errors.New("injected unreadable failure")
	}
}

func unreadableFile(path string, cause error) (stateread.File, error) {
	return stateread.File{Kind: stateread.KindUnreadable}, stateread.UnusableError(path, cause)
}

func vectorRegularFileReader(target string, failure error) func(string) (stateread.File, error) {
	return func(path string) (stateread.File, error) {
		if path == target {
			return unreadableFile(path, failure)
		}
		return stateread.ReadRegularFile(path)
	}
}

func vectorStateFileReader(target string, failure error) func(string) (stateread.File, error) {
	return func(path string) (stateread.File, error) {
		if path == target {
			return unreadableFile(path, failure)
		}
		return stateread.ReadFile(path)
	}
}

func runMarkerReadFailureVector(t *testing.T, tc readFailureVectorCase) {
	t.Helper()
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
	marker := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), envmarker.Name)
	var readStateFile func(string) (stateread.File, error)
	switch tc.FailureClass {
	case "permission-denied", "io-error":
		readStateFile = vectorStateFileReader(marker, vectorFailure(tc.FailureClass))
	case "unparseable-content":
		if err := os.WriteFile(marker, []byte("{\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	case "schema-invalid-content":
		if err := os.WriteFile(marker, []byte("{}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	case "": // marker absence is known stale state
		if err := os.Remove(marker); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatalf("unsupported marker failure class %q", tc.FailureClass)
	}
	req := fx.request("codex_cli")
	req.readStateFile = readStateFile
	result, err := Resolve(req)
	want := envmarker.DiagMarkerUnreadable
	if tc.FailureClass == "" {
		want = DiagHomeStale
	}
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("Resolve marker case error = %v; want %s", err, want)
	}
	if tc.FailureClass != "" && !strings.HasPrefix(err.Error(), envmarker.DiagMarkerUnreadable) {
		t.Fatalf("Resolve collapsed unreadable marker into stale: %v", err)
	}
	if tc.Conforming != nil && !*tc.Conforming && strings.HasPrefix(err.Error(), DiagHomeStale) {
		t.Fatalf("unreadable marker was reported as absence: %v", err)
	}
	if result != nil && len(result.Document) != 0 {
		t.Fatal("Resolve emitted a fragment after an unreadable marker")
	}
	statusReq := statusRequest(fx)
	statusReq.readStateFile = readStateFile
	status, err := StatusOf(statusReq)
	if err != nil {
		t.Fatal(err)
	}
	row := findHome(status, "acme", "codex_cli")
	if row == nil || row.Current {
		t.Fatalf("unreadable marker status row = %+v; want non-current currency unknown", row)
	}
}

func runLockReadFailureVector(t *testing.T, tc readFailureVectorCase) {
	t.Helper()
	fx := writeManagedFixture(t, "acme")
	lock := lockPath(fx.home, "acme")
	var readRegularFile func(string) (stateread.File, error)
	switch tc.FailureClass {
	case "permission-denied", "io-error", "parent-not-directory":
		readRegularFile = vectorRegularFileReader(lock, vectorFailure(tc.FailureClass))
	case "unparseable-content":
		if err := os.WriteFile(lock, []byte("{\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	case "schema-invalid-content":
		if err := os.WriteFile(lock, []byte("{\"schema_version\":1}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	case "symlink-where-regular-required":
		outside := filepath.Join(t.TempDir(), "lock.json")
		if err := os.WriteFile(outside, []byte("{}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(lock); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, lock); err != nil {
			t.Fatal(err)
		}
	case "directory-where-file-expected":
		if err := os.Remove(lock); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(lock, 0o700); err != nil {
			t.Fatal(err)
		}
	case "": // the known-absent lock vector
		if err := os.Remove(lock); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatalf("unsupported lock failure class %q", tc.FailureClass)
	}
	if tc.Operation == "update" {
		broken, err := os.ReadFile(lock)
		if err != nil {
			t.Fatal(err)
		}
		_, _, err = UpdateWithOptions(fx.home, "acme", UpdateOptions{Policy: Policy{}, readRegularFile: readRegularFile})
		if err == nil || !strings.Contains(err.Error(), envregistry.DiagStoreUntrusted) {
			t.Fatalf("Update error = %v; want %s", err, envregistry.DiagStoreUntrusted)
		}
		after, err := os.ReadFile(lock)
		if err != nil || string(after) != string(broken) {
			t.Fatalf("update changed unreadable lock: %v", err)
		}
		if _, err := os.Lstat(ProfileDir(fx.home, DefaultProfile)); !os.IsNotExist(err) {
			t.Fatalf("update wrote default profile before refusing unreadable lock: %v", err)
		}
		return
	}
	if tc.Operation == "env-status" {
		statusReq := statusRequest(fx)
		statusReq.readRegularFile = readRegularFile
		status, err := StatusOf(statusReq)
		if err != nil {
			t.Fatal(err)
		}
		row := findHome(status, "acme", "codex_cli")
		if row == nil || row.Current || !containsText(row.Findings, envregistry.DiagStoreUntrusted) {
			t.Fatalf("unreadable lock status row = %+v; want non-current currency unknown %s", row, envregistry.DiagStoreUntrusted)
		}
		return
	}
	req := fx.request("codex_cli")
	req.Repair = tc.Operation == "repair"
	req.readRegularFile = readRegularFile
	result, err := Resolve(req)
	if tc.FailureClass == "" {
		if err == nil || !strings.Contains(err.Error(), DiagProfileUnknown) {
			t.Fatalf("absent lock error = %v; want %s", err, DiagProfileUnknown)
		}
		return
	}
	if err == nil || !strings.HasPrefix(err.Error(), envregistry.DiagStoreUntrusted) {
		t.Fatalf("Resolve(%s) error = %v; want %s", tc.Operation, err, envregistry.DiagStoreUntrusted)
	}
	if tc.Conforming != nil && !*tc.Conforming && strings.Contains(err.Error(), DiagProfileUnknown) {
		t.Fatalf("unreadable lock was reported as profile unknown: %v", err)
	}
	if result != nil && len(result.Document) != 0 {
		t.Fatal("Resolve emitted a fragment from an unreadable lock")
	}
	if tc.Operation == "repair" {
		if _, err := os.Lstat(filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), envmarker.Name)); !os.IsNotExist(err) {
			t.Fatalf("repair rebuilt from unreadable lock: %v", err)
		}
	}
}

func runSeedReadFailureVector(t *testing.T, tc readFailureVectorCase) {
	t.Helper()
	fx := writeManagedFixture(t, "acme")
	seed := filepath.Join(fx.native["codex_cli"], "config.toml")
	var readRegularFile func(string) (stateread.File, error)
	switch tc.FailureClass {
	case "permission-denied", "io-error", "parent-not-directory":
		readRegularFile = vectorRegularFileReader(seed, vectorFailure(tc.FailureClass))
	case "unparseable-content":
		if err := os.WriteFile(seed, []byte("not = [toml\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	case "symlink-where-regular-required":
		target := filepath.Join(t.TempDir(), "config.toml")
		if err := os.WriteFile(target, []byte("cli_auth_credentials_store = \"file\"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, seed); err != nil {
			t.Fatal(err)
		}
	case "directory-where-file-expected":
		if err := os.Mkdir(seed, 0o700); err != nil {
			t.Fatal(err)
		}
	case "": // absent native seed is allowed
	default:
		t.Fatalf("unsupported seed failure class %q", tc.FailureClass)
	}
	req := fx.request("codex_cli")
	req.Repair = true
	req.readRegularFile = readRegularFile
	_, err := Resolve(req)
	if tc.FailureClass == "" {
		if err != nil {
			t.Fatalf("absent seed should not block provisioning: %v", err)
		}
		return
	}
	if err == nil || !strings.Contains(err.Error(), envregistry.DiagSeedUnreadable) {
		t.Fatalf("Resolve with unreadable seed error = %v; want %s", err, envregistry.DiagSeedUnreadable)
	}
	if _, err := os.Lstat(filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), envmarker.Name)); !os.IsNotExist(err) {
		t.Fatalf("provisioning continued after unreadable seed: %v", err)
	}
}

func runPassthroughReadFailureVector(t *testing.T, tc readFailureVectorCase) {
	t.Helper()
	path := "auth.json"
	if tc.FailureClass == "parent-not-directory" {
		// The stock Codex passthrough is a root-level file, whose parent is
		// also required to contain the marker. Give this production-entry
		// case a valid nested registry path so the failure is a real
		// non-directory ancestor on every platform, including Windows where
		// ERROR_PATH_NOT_FOUND also satisfies os.IsNotExist.
		path = "nested/auth.json"
		originalRegistry := envregistry.Registry
		registry := append([]envregistry.Adapter(nil), originalRegistry...)
		found := false
		for index := range registry {
			if registry[index].ID != envregistry.CodexCLI {
				continue
			}
			passthrough := make(map[string][]envregistry.Passthrough, len(registry[index].Passthrough))
			for goos, entries := range registry[index].Passthrough {
				passthrough[goos] = append([]envregistry.Passthrough(nil), entries...)
			}
			entries := passthrough["default"]
			if len(entries) != 1 {
				t.Fatalf("Codex default passthrough entries = %d; want one", len(entries))
			}
			entries[0].Path = path
			passthrough["default"] = entries
			registry[index].Passthrough = passthrough
			found = true
			break
		}
		if !found {
			t.Fatal("Codex adapter is missing from the environment registry")
		}
		envregistry.Registry = registry
		t.Cleanup(func() { envregistry.Registry = originalRegistry })
	}
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
	link := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), filepath.FromSlash(path))
	var lstat func(string) (os.FileInfo, error)
	var readlink func(string) (string, error)
	if tc.FailureClass != "" && tc.FailureClass != "parent-not-directory" {
		failure := vectorFailure(tc.FailureClass)
		lstat = func(path string) (os.FileInfo, error) {
			if path == link {
				return nil, &os.PathError{Op: "lstat", Path: path, Err: failure}
			}
			return os.Lstat(path)
		}
		if tc.Name == "passthrough-readlink-io-error-unreadable" {
			lstat = func(path string) (os.FileInfo, error) { return os.Lstat(path) }
			readlink = func(path string) (string, error) {
				if path == link {
					return "", failure
				}
				return os.Readlink(path)
			}
		}
	} else if tc.FailureClass == "" && tc.EntryKind == "missing" {
		if err := os.Remove(link); err != nil {
			t.Fatal(err)
		}
	} else if tc.FailureClass == "" && tc.EntryKind == "directory" {
		if err := os.Remove(link); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(link, 0o700); err != nil {
			t.Fatal(err)
		}
	} else if tc.FailureClass == "parent-not-directory" {
		parent := filepath.Dir(link)
		if err := os.Remove(link); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(parent); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(parent, []byte("blocking parent"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	if tc.Operation == "env-resolve" {
		req := fx.request("codex_cli")
		req.passthroughLstat = lstat
		req.passthroughReadlink = readlink
		result, err := Resolve(req)
		want := DiagHomeStale
		if tc.FailureClass != "" {
			want = envregistry.DiagPassthroughUnreadable
		}
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("Resolve error = %v; want %s", err, want)
		}
		if tc.FailureClass != "" && !strings.HasPrefix(err.Error(), envregistry.DiagPassthroughUnreadable) {
			t.Fatalf("Resolve collapsed unreadable passthrough into stale: %v", err)
		}
		if result != nil && len(result.Document) != 0 {
			t.Fatal("Resolve emitted a fragment for stale or unreadable passthrough")
		}
		return
	}
	statusReq := statusRequest(fx)
	statusReq.passthroughLstat = lstat
	statusReq.passthroughReadlink = readlink
	status, err := StatusOf(statusReq)
	if err != nil {
		t.Fatal(err)
	}
	row := findHome(status, "acme", "codex_cli")
	if row == nil || row.Current {
		t.Fatalf("passthrough status row = %+v; want non-current", row)
	}
	if tc.FailureClass != "" {
		if !containsText(row.Findings, envregistry.DiagPassthroughUnreadable) || containsText(row.Findings, envregistry.DiagPassthroughDetached) {
			t.Fatalf("unreadable status findings = %v", row.Findings)
		}
	} else if !containsText(row.Findings, envregistry.DiagPassthroughDetached) {
		t.Fatalf("detached status findings = %v", row.Findings)
	}
	if tc.Expected.RepairRelinks != nil && *tc.Expected.RepairRelinks {
		req := fx.request("codex_cli")
		req.Repair = true
		if _, err := Resolve(req); err != nil {
			t.Fatalf("repair detached entry: %v", err)
		}
		if _, err := os.Lstat(link); err != nil {
			t.Fatalf("repair did not restore detached link: %v", err)
		}
	}
}

func runBackupReadFailureVector(t *testing.T, tc readFailureVectorCase) {
	t.Helper()
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
	managed := ManagedHomeDir(fx.home, "acme", "codex_cli")
	backupRoot := filepath.Join(managed, ".agent-environment-backup")
	if tc.FailureClass != "" {
		if err := os.Mkdir(backupRoot, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	statusReq := statusRequest(fx)
	if tc.FailureClass != "" {
		statusReq.readStateDirectory = func(path string) (stateread.Directory, error) {
			if path == backupRoot {
				return stateread.Directory{Kind: stateread.KindUnreadable}, stateread.UnusableError(path, vectorFailure(tc.FailureClass))
			}
			return stateread.ReadDir(path)
		}
	}
	status, err := StatusOf(statusReq)
	if err != nil {
		t.Fatal(err)
	}
	row := findHome(status, "acme", "codex_cli")
	if row == nil {
		t.Fatal("backup status omitted the profile row")
	}
	if tc.FailureClass != "" {
		if row.BackupsKnown || row.Current || !containsText(row.Findings, envregistry.DiagBackupRecordUnreadable) {
			t.Fatalf("unreadable backup inventory row = %+v", row)
		}
		if tc.Conforming != nil && !*tc.Conforming && !strings.Contains(strings.Join(row.Findings, " "), envregistry.DiagBackupRecordUnreadable) {
			t.Fatal("unreadable backup inventory was reported as empty/current")
		}
	} else if !row.BackupsKnown || row.Backups != 0 {
		t.Fatalf("absent backup inventory row = %+v; want known zero", row)
	}
}
