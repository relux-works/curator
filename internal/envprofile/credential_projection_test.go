package envprofile

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
)

// Resolve must use one native selector observation for admission, links and
// marker records, including repair's verification passes. A later operation
// must observe a changed selector rather than reuse the previous projection.
func TestCredentialProjectionSingleObservation(t *testing.T) {
	requireLinkCapability(t)
	for _, mode := range []string{envregistry.IsolationShared, envregistry.IsolationIsolated} {
		t.Run(mode, func(t *testing.T) {
			fx := writeManagedFixture(t, "projection")
			seedLiveNativeCredentials(t, fx)
			req := fx.request(envregistry.CodexCLI)
			req.Repair = true
			req.Machine.Isolation[fx.profile] = map[string]string{envregistry.CodexCLI: mode}
			path := filepath.Join(fx.native[envregistry.CodexCLI], "config.toml")
			if err := os.WriteFile(path, []byte("cli_auth_credentials_store = 'file'\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			// Force the repair path without asking the one-time provisioning
			// seed reader to share the policy observation.
			if err := os.Remove(filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI), "AGENTS.md")); err != nil {
				t.Fatal(err)
			}
			reads := 0
			req.readRegularFile = func(name string) (stateread.File, error) {
				if name == path {
					reads++
					if reads > 1 {
						return stateread.File{}, os.ErrPermission
					}
				}
				return stateread.ReadRegularFile(name)
			}
			result, err := Resolve(req)
			if err != nil || result == nil || len(result.Document) == 0 {
				t.Fatalf("one file observation should resolve: result=%+v err=%v reads=%d", result, err, reads)
			}
			if reads != 1 {
				t.Fatalf("native selector observed %d times, want exactly once", reads)
			}
			marker := readManagedMarker(t, fx, envregistry.CodexCLI)
			if marker.Passthrough == nil {
				t.Fatal("missing credential records")
			}
			link := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI), "auth.json")
			if mode == envregistry.IsolationShared {
				if len(*marker.Passthrough) != 1 || (*marker.Passthrough)[0].Backend != "file" {
					t.Fatalf("records disagree with observed file policy: %+v", marker.Passthrough)
				}
				if target, err := os.Readlink(link); err != nil || target != filepath.Join(fx.native[envregistry.CodexCLI], "auth.json") {
					t.Fatalf("file policy link = %s, %v", target, err)
				}
			} else if _, err := os.Lstat(link); !os.IsNotExist(err) || len(*marker.Passthrough) != 0 {
				t.Fatalf("isolated policy must not share credentials: %v %+v", err, marker.Passthrough)
			}

			// The same public request value is reusable; the projection belongs
			// to one operation, not to the caller or a process-wide cache.
			reads = 0
			if _, err := Resolve(req); err != nil || reads != 1 {
				t.Fatalf("fresh operation must reobserve: err=%v reads=%d", err, reads)
			}
			if err := os.WriteFile(path, []byte("cli_auth_credentials_store = 'invalid'\n"), 0600); err != nil {
				t.Fatal(err)
			}
			before := snapshotCredentialScope(t, fx)
			reads = 0
			result, err = Resolve(req)
			if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialUnsupported) || result != nil && len(result.Document) != 0 || reads != 1 {
				t.Fatalf("changed selector must refuse: result=%+v err=%v reads=%d", result, err, reads)
			}
			if !reflect.DeepEqual(before, snapshotCredentialScope(t, fx)) {
				t.Fatal("selector refusal changed configuration, markers, links or credentials")
			}
		})
	}
}

// Native keyring and auto remain forbidden for isolated homes. The adjacent
// file control proves admission is reachable; refusals preserve all credential
// scope bytes and links even when repair is requested.
func TestCredentialProjectionIsolationRefusals(t *testing.T) {
	for _, store := range []string{"file", "keyring", "auto"} {
		t.Run(store, func(t *testing.T) {
			fx := writeManagedFixture(t, "isolation")
			seedLiveNativeCredentials(t, fx)
			path := filepath.Join(fx.native[envregistry.CodexCLI], "config.toml")
			if err := os.WriteFile(path, []byte("cli_auth_credentials_store = '"+store+"'\n"), 0600); err != nil {
				t.Fatal(err)
			}
			req := fx.request(envregistry.CodexCLI)
			req.Repair = true
			req.Machine.Isolation[fx.profile] = map[string]string{envregistry.CodexCLI: envregistry.IsolationIsolated}
			before := snapshotCredentialScope(t, fx)
			result, err := Resolve(req)
			if store == "file" {
				if err != nil || result == nil || len(result.Document) == 0 {
					t.Fatalf("file positive control: result=%+v err=%v", result, err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), envregistry.DiagIsolatedUnsupported) || result != nil && len(result.Document) != 0 {
				t.Fatalf("native %s isolation must refuse: result=%+v err=%v", store, result, err)
			}
			if !reflect.DeepEqual(before, snapshotCredentialScope(t, fx)) {
				t.Fatal("isolation refusal changed forbidden state")
			}
		})
	}
}

// StatusOf uses the same policy as Resolve, without a second native selector
// read. Failed observation reports non-current without modifying the home;
// a later readable observation restores current status rather than caching it.
func TestCredentialProjectionStatusObservation(t *testing.T) {
	fx := writeManagedFixture(t, "status-projection")
	seedLiveNativeCredentials(t, fx)
	provision(t, fx, envregistry.CodexCLI, envregistry.DefaultMachineConfig())
	path := filepath.Join(fx.native[envregistry.CodexCLI], "config.toml")
	for _, refused := range []bool{false, true, false} {
		req := statusRequest(fx)
		reads := 0
		req.readRegularFile = func(name string) (stateread.File, error) {
			if name == path {
				reads++
				if refused || reads > 1 {
					return stateread.File{}, os.ErrPermission
				}
			}
			return stateread.ReadRegularFile(name)
		}
		before := snapshotCredentialScope(t, fx)
		status, err := StatusOf(req)
		if err != nil {
			t.Fatal(err)
		}
		row := findHome(status, fx.profile, envregistry.CodexCLI)
		if row == nil || row.Current == refused || reads != 1 {
			t.Fatalf("status projection: row=%+v refused=%v reads=%d", row, refused, reads)
		}
		if refused && !strings.Contains(strings.Join(row.Findings, ";"), envregistry.DiagSeedUnreadable) {
			t.Fatalf("failed observation must retain unreadable diagnostic: %+v", row)
		}
		if !reflect.DeepEqual(before, snapshotCredentialScope(t, fx)) {
			t.Fatal("status observation changed forbidden state")
		}
	}
}

// Identity-only staging must preserve a sibling's recorded sharing policy,
// even when the caller already resolved a different isolation. The internal
// staging entry is the bound here: Resolve/ApplyMigration integration is
// exercised separately by the existing RC14 identity-migration tests.
func TestCredentialProjectionIdentitySubrequest(t *testing.T) {
	for _, mode := range []string{envregistry.IsolationShared, envregistry.IsolationIsolated} {
		t.Run(mode, func(t *testing.T) {
			fx := legacyIdentityFixture(t, envregistry.Pi)
			marker := readManagedMarker(t, fx, envregistry.Pi)
			marker.Version = envmarker.VersionV1
			records := []envmarker.Passthrough{}
			for _, record := range *marker.Passthrough {
				records = append(records, envmarker.Passthrough{Path: record.Path, Strategy: record.Strategy})
			}
			marker.Passthrough = &records
			payload, err := marker.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			markerPath := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.Pi), envmarker.Name)
			if err := os.WriteFile(markerPath, payload, 0600); err != nil {
				t.Fatal(err)
			}
			req := fx.request(envregistry.Pi)
			req.Machine.Isolation[fx.profile] = map[string]string{envregistry.Pi: mode}
			adapter, err := envregistry.ByID(envregistry.Pi)
			if err != nil {
				t.Fatal(err)
			}
			policy, err := req.resolveCredentialPolicy(adapter)
			if err != nil || policy.isolation != mode {
				t.Fatalf("caller policy: %+v, %v", policy, err)
			}
			before := snapshotCredentialScope(t, fx)
			op, err := beginOperation(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := op.close(); err != nil {
					t.Error(err)
				}
			}()
			migration, err := prepareIdentityMigration(op, &req, []string{fx.profile}, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer migration.cleanup()
			if !reflect.DeepEqual(before, snapshotCredentialScope(t, fx)) {
				t.Fatal("identity staging changed live credential scope")
			}
			staged, err := os.ReadFile(migration.targets[markerPath])
			if err != nil {
				t.Fatal(err)
			}
			projected, err := envmarker.Parse(staged)
			if err != nil {
				t.Fatal(err)
			}
			if projected.Passthrough == nil || len(*projected.Passthrough) != 1 || (*projected.Passthrough)[0].Isolation != envregistry.IsolationShared || (*projected.Passthrough)[0].Path != "auth.json" {
				t.Fatalf("identity-only staging must keep recorded shared policy: %+v", projected.Passthrough)
			}
		})
	}
}
