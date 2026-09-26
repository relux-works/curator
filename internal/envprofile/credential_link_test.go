// Production entry points under test: Resolve and StatusOf for the
// Decision 0017 fix-first credential-link repairs (environments §7.4,
// §10.1; manager §12.4). Every row drives a production entry on a
// temporary store; helper-direct assertions are bounds, not rows.
package envprofile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
)

// TestSharedToIsolatedRemovesStaleLink is row (a): a shared→isolated turn
// leaves no stale recorded link behind. Repair never migrates: it
// refuses the stale link with environment_credential_conflict pointing
// at the explicit migration, which unlinks it preserving the native
// bytes. A mutant that restores the unlink inside repair makes the
// repair succeed and fails the refusal; one that skips the migration
// unlink leaves the link on disk and fails the absence assertion.
func TestSharedToIsolatedRemovesStaleLink(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	nativeAuth := []byte("{\"t\":\"operator-credential\"}\n")
	if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "auth.json"), nativeAuth, 0o600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
	link := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
	if _, err := os.Readlink(link); err != nil {
		t.Fatalf("shared provision links auth.json: %v", err)
	}
	isolated := envregistry.DefaultMachineConfig()
	isolated.Isolation = map[string]map[string]string{"acme": {"codex_cli": "isolated"}}
	stale := fx.request("codex_cli")
	stale.Machine = isolated
	if _, err := Resolve(stale); err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
		t.Fatalf("the mode turn must be stale, got %v", err)
	}
	repair := fx.request("codex_cli")
	repair.Machine = isolated
	repair.Repair = true
	_, err := Resolve(repair)
	if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
		t.Fatalf("repair must refuse the stale link with %s, got %v", envregistry.DiagCredentialConflict, err)
	}
	for _, want := range []string{"migration needed", "env migrate", link} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("the refusal points at the explicit migration, want %q in %v", want, err)
		}
	}
	if target, err := os.Readlink(link); err != nil {
		t.Fatalf("the refused link is untouched: %v", err)
	} else if target != filepath.Join(fx.native["codex_cli"], "auth.json") {
		t.Fatalf("the refused link still targets the declared store: %q", target)
	}
	mreq := fx.migrateRequest()
	mreq.Machine = isolated
	report, err := PlanMigration(mreq)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Ops()) != 1 || report.Ops()[0].Kind != MigrateOpUnlink {
		t.Fatalf("the migration plans the unlink: %+v", report.Ops())
	}
	mreq.Expect = report.Hash
	if _, err := ApplyMigration(mreq); err != nil {
		t.Fatalf("the migration applies the isolated turn: %v", err)
	}
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Fatalf("no stale link survives shared→isolated, lstat: %v", err)
	}
	if payload, err := os.ReadFile(filepath.Join(fx.native["codex_cli"], "auth.json")); err != nil || string(payload) != string(nativeAuth) {
		t.Fatalf("unlinking preserves the native bytes: %q (%v)", payload, err)
	}
	marker := readManagedMarker(t, fx, "codex_cli")
	if marker.Passthrough == nil || len(*marker.Passthrough) != 0 {
		t.Fatalf("the record drops the stale entry: %+v", marker.Passthrough)
	}
	if _, err := Resolve(stale); err != nil {
		t.Fatalf("the isolated home is current: %v", err)
	}
}

// TestCredentialLinkRegularFileRefuses is row (b): a regular file at a
// wanted link path refuses with environment_credential_conflict naming
// the path, and the bytes — managed and native — are untouched. The
// empty-file subtest is the narrowing killer for a mutant that refuses
// only non-empty files; restoring the unconditional Remove passes
// neither subtest.
func TestCredentialLinkRegularFileRefuses(t *testing.T) {
	requireLinkCapability(t)
	for _, tc := range []struct {
		name  string
		bytes []byte
	}{
		{"non-empty", []byte("detached-credential\n")},
		{"empty", []byte{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := writeManagedFixture(t, "acme")
			nativeAuth := []byte("{\"t\":\"operator-credential\"}\n")
			if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "auth.json"), nativeAuth, 0o600); err != nil {
				t.Fatal(err)
			}
			provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
			link := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
			_ = os.Remove(link)
			if err := os.WriteFile(link, tc.bytes, 0o600); err != nil {
				t.Fatal(err)
			}
			stale, err := Resolve(fx.request("codex_cli"))
			if err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
				t.Fatalf("a file at the link path must be stale, got %v", err)
			}
			reasons := strings.Join(stale.StaleReasons, "; ")
			if !strings.Contains(reasons, envregistry.DiagCredentialConflict) || !strings.Contains(reasons, "auth.json") {
				t.Fatalf("stale reasons carry conflict-class wording: %q", reasons)
			}
			req := fx.request("codex_cli")
			req.Repair = true
			_, err = Resolve(req)
			if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
				t.Fatalf("repair must refuse with %s, got %v", envregistry.DiagCredentialConflict, err)
			}
			if !strings.Contains(err.Error(), link) {
				t.Fatalf("the refusal names the path %s: %v", link, err)
			}
			if strings.Contains(err.Error(), DiagRepairFailed) {
				t.Fatalf("a credential refusal is not wrapped as %s: %v", DiagRepairFailed, err)
			}
			info, err := os.Lstat(link)
			if err != nil || info.Mode()&os.ModeSymlink != 0 {
				t.Fatalf("the managed file is untouched, still a regular file: %v", err)
			}
			if payload, err := os.ReadFile(link); err != nil || string(payload) != string(tc.bytes) {
				t.Fatalf("managed bytes untouched: %q (%v)", payload, err)
			}
			if payload, err := os.ReadFile(filepath.Join(fx.native["codex_cli"], "auth.json")); err != nil || string(payload) != string(nativeAuth) {
				t.Fatalf("native bytes untouched: %q (%v)", payload, err)
			}
		})
	}
}

// TestStaleCredentialLinkRefusals pins the fix-first bound on the removal
// side: a stale recorded link path holding a regular file, a foreign
// symlink, or a directory refuses with
// environment_credential_conflict naming the path, removing nothing. The
// foreign symlink shares the expected basename on purpose: a mutant that
// compares basenames instead of full targets unlinks it and fails.
func TestStaleCredentialLinkRefusals(t *testing.T) {
	requireLinkCapability(t)
	setup := func(t *testing.T) (*managedFixture, string) {
		t.Helper()
		fx := writeManagedFixture(t, "acme")
		if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "auth.json"), []byte("{\"t\":\"operator\"}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
		return fx, filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
	}
	isolatedRepair := func(fx *managedFixture) ResolveRequest {
		machine := envregistry.DefaultMachineConfig()
		machine.Isolation = map[string]map[string]string{"acme": {"codex_cli": "isolated"}}
		req := fx.request("codex_cli")
		req.Machine = machine
		req.Repair = true
		return req
	}
	t.Run("regular file", func(t *testing.T) {
		fx, link := setup(t)
		_ = os.Remove(link)
		managed := []byte("managed-credential\n")
		if err := os.WriteFile(link, managed, 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := Resolve(isolatedRepair(fx)); err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
			t.Fatalf("a file at the stale path must refuse, got %v", err)
		} else if !strings.Contains(err.Error(), link) {
			t.Fatalf("the refusal names the path: %v", err)
		}
		if payload, err := os.ReadFile(link); err != nil || string(payload) != string(managed) {
			t.Fatalf("managed bytes untouched: %q (%v)", payload, err)
		}
	})
	t.Run("foreign symlink", func(t *testing.T) {
		fx, link := setup(t)
		foreign := filepath.Join(t.TempDir(), "auth.json")
		if err := os.WriteFile(foreign, []byte("foreign\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		_ = os.Remove(link)
		if err := os.Symlink(foreign, link); err != nil {
			t.Fatal(err)
		}
		if _, err := Resolve(isolatedRepair(fx)); err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
			t.Fatalf("a foreign link at the stale path must refuse, got %v", err)
		} else if !strings.Contains(err.Error(), link) {
			t.Fatalf("the refusal names the path: %v", err)
		}
		if target, err := os.Readlink(link); err != nil || target != foreign {
			t.Fatalf("the foreign link is untouched: %q (%v)", target, err)
		}
	})
	t.Run("directory", func(t *testing.T) {
		fx, link := setup(t)
		_ = os.Remove(link)
		if err := os.Mkdir(link, 0o755); err != nil {
			t.Fatal(err)
		}
		kept := filepath.Join(link, "kept")
		if err := os.WriteFile(kept, []byte("kept\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Resolve(isolatedRepair(fx)); err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
			t.Fatalf("a directory at the stale path must refuse, got %v", err)
		} else if !strings.Contains(err.Error(), link) {
			t.Fatalf("the refusal names the path: %v", err)
		}
		if payload, err := os.ReadFile(kept); err != nil || string(payload) != "kept\n" {
			t.Fatalf("directory contents untouched: %q (%v)", payload, err)
		}
	})
}

// TestDanglingPiLinkReportedDetached is row (c), the operator-machine
// reproduction: the managed Pi link still targets the pre-0017 native
// path, which does not exist, while the real credential lives at the
// agent root. Status and resolve report the recorded passthrough as
// detached with environment_credential_conflict-class wording — never
// silence — and repair refuses without touching bytes, pointing at the
// explicit migration that heals the home
// (TestMigratePiWrongTargetToAgentRoot). A mutant that restores the old
// Pi target or compares basenames only reports current and fails; one
// that re-points inside repair makes the repair succeed and fails the
// refusal below.
func TestDanglingPiLinkReportedDetached(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	native := fx.native["pi"]
	agentAuth := filepath.Join(native, "agent", "auth.json")
	if err := os.MkdirAll(filepath.Dir(agentAuth), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agentAuth, []byte("{\"t\":\"operator-pi-credential\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	link := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
	// Reproduce the operator state: the managed link targets the old
	// native path, which exists nowhere.
	_ = os.Remove(link)
	oldTarget := filepath.Join(native, "auth.json")
	if err := os.Symlink(oldTarget, link); err != nil {
		t.Fatal(err)
	}
	status, err := StatusOf(statusRequest(fx))
	if err != nil {
		t.Fatal(err)
	}
	row := findHome(status, "acme", "pi")
	if row == nil || row.Current {
		t.Fatalf("the mis-targeted Pi home is non-current: %+v", row)
	}
	findings := strings.Join(row.Findings, "; ")
	for _, want := range []string{envregistry.DiagPassthroughDetached, envregistry.DiagCredentialConflict, "auth.json", agentAuth} {
		if !strings.Contains(findings, want) {
			t.Fatalf("status findings carry detached conflict-class wording, want %q in %q", want, findings)
		}
	}
	stale, err := Resolve(fx.request("pi"))
	if err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
		t.Fatalf("the mis-targeted Pi home must be stale, got %v", err)
	}
	reasons := strings.Join(stale.StaleReasons, "; ")
	for _, want := range []string{"is detached", envregistry.DiagCredentialConflict, "auth.json", agentAuth} {
		if !strings.Contains(reasons, want) {
			t.Fatalf("stale reasons carry detached conflict-class wording, want %q in %q", want, reasons)
		}
	}
	req := fx.request("pi")
	req.Repair = true
	_, err = Resolve(req)
	if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
		t.Fatalf("repair must refuse the mis-targeted link with %s, got %v", envregistry.DiagCredentialConflict, err)
	}
	for _, want := range []string{"migration needed", "env migrate", link, oldTarget, agentAuth} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("the refusal points at the explicit migration, want %q in %v", want, err)
		}
	}
	if strings.Contains(err.Error(), DiagRepairFailed) {
		t.Fatalf("a credential refusal is not wrapped as %s: %v", DiagRepairFailed, err)
	}
	if target, err := os.Readlink(link); err != nil || target != oldTarget {
		t.Fatalf("the refused link is untouched: %q (%v), want %q", target, err, oldTarget)
	}
	if payload, err := os.ReadFile(agentAuth); err != nil || string(payload) != "{\"t\":\"operator-pi-credential\"}\n" {
		t.Fatalf("native bytes untouched: %q (%v)", payload, err)
	}
	if _, err := os.Lstat(oldTarget); !os.IsNotExist(err) {
		t.Fatalf("repair creates nothing at the old target: %v", err)
	}
	if _, err := Resolve(fx.request("pi")); err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
		t.Fatalf("the refused home stays stale, got %v", err)
	}
}

// TestCodexIsolatedAdmission is row (d): codex_cli isolated is admitted
// under effective file storage only. An absent config file and an absent
// key both resolve to file; keyring and auto refuse with
// environment_isolated_unsupported; any other selector fails closed with
// environment_credential_unsupported — under shared too. A mutant that
// drops the auto arm admits auto; one that broadens the gate to
// store-is-not-file reports the unknown selector as isolated-unsupported
// instead of credential-unsupported; both fail below.
func TestCodexIsolatedAdmission(t *testing.T) {
	for _, tc := range []struct {
		name    string
		config  *string
		wantErr string
	}{
		{"absent file admits", nil, ""},
		{"absent key admits", strptr("model = \"o3\"\n"), ""},
		{"file admits", strptr("cli_auth_credentials_store = \"file\"\n"), ""},
		{"keyring refuses", strptr("cli_auth_credentials_store = \"keyring\"\n"), envregistry.DiagIsolatedUnsupported},
		{"auto refuses", strptr("cli_auth_credentials_store = \"auto\"\n"), envregistry.DiagIsolatedUnsupported},
		{"unknown fails closed", strptr("cli_auth_credentials_store = \"ephemeral\"\n"), envregistry.DiagCredentialUnsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := writeManagedFixture(t, "acme")
			if tc.config != nil {
				if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "config.toml"), []byte(*tc.config), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			machine := envregistry.DefaultMachineConfig()
			machine.Isolation = map[string]map[string]string{"acme": {"codex_cli": "isolated"}}
			req := fx.request("codex_cli")
			req.Machine = machine
			req.Repair = true
			_, err := Resolve(req)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("isolated must be admitted: %v", err)
				}
				marker := readManagedMarker(t, fx, "codex_cli")
				if marker.Passthrough == nil || len(*marker.Passthrough) != 0 {
					t.Fatalf("an isolated home records no passthrough: %+v", marker.Passthrough)
				}
				if _, err := os.Lstat(filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")); !os.IsNotExist(err) {
					t.Fatalf("an isolated home links nothing: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("want %s, got %v", tc.wantErr, err)
			}
			if tc.wantErr == envregistry.DiagIsolatedUnsupported && !strings.Contains(err.Error(), "codex_cli") {
				t.Fatalf("the refusal names the environment: %v", err)
			}
		})
	}
}

// TestCodexUnknownStoreSharedRefuses pins the shared side of the unknown
// selector: provisioning with a storage selector outside the verified
// file/keyring/auto set fails closed with
// environment_credential_unsupported before the first write, so fixing
// the selector and re-running provisions cleanly.
func TestCodexUnknownStoreSharedRefuses(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	config := filepath.Join(fx.native["codex_cli"], "config.toml")
	if err := os.WriteFile(config, []byte("cli_auth_credentials_store = \"ephemeral\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	req := fx.request("codex_cli")
	req.Repair = true
	if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialUnsupported) {
		t.Fatalf("an unknown selector must fail closed, got %v", err)
	}
	if _, err := os.Lstat(ManagedHomeDir(fx.home, "acme", "codex_cli")); !os.IsNotExist(err) {
		t.Fatalf("a refused provisioning writes nothing: %v", err)
	}
	if err := os.WriteFile(config, []byte("cli_auth_credentials_store = \"file\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(req); err != nil {
		t.Fatalf("fixing the selector provisions: %v", err)
	}
}

// TestPiProvisionTargetsAgentRoot is row (e): new Pi provisioning links
// the managed auth.json at the 0017 native root ~/.pi/agent, and the
// marker records path and strategy only — the frozen v1 schema gains no
// fields. Provisioning with no native credential yet succeeds — the link
// is established, the native side is the operator's — warning loudly
// with the detached-pending state; bare resolve succeeds the same way
// until the native credential exists. The pending warning carries no
// conflict diagnostic: it is the expected pre-login shape, not a
// refusal. A mutant that reports the pending link stale fails the bare
// resolve; one that silences it fails the warning assertions.
func TestPiProvisionTargetsAgentRoot(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	req := fx.request("pi")
	req.Repair = true
	provisioned, err := Resolve(req)
	if err != nil {
		t.Fatalf("provisioning with no native credential yet succeeds: %v", err)
	}
	warnings := strings.Join(provisioned.Warnings, "; ")
	for _, want := range []string{"detached-pending", "does not exist yet", "pi"} {
		if !strings.Contains(warnings, want) {
			t.Fatalf("provision warns loudly about the pending target, want %q in %q", want, warnings)
		}
	}
	if strings.Contains(warnings, envregistry.DiagCredentialConflict) {
		t.Fatalf("the pending warning is not a conflict: %q", warnings)
	}
	link := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
	want := filepath.Join(fx.native["pi"], "agent", "auth.json")
	if target, err := os.Readlink(link); err != nil || target != want {
		t.Fatalf("new Pi homes link the agent root: %q (%v), want %q", target, err, want)
	}
	current, err := Resolve(fx.request("pi"))
	if err != nil {
		t.Fatalf("a dangling-but-correct link succeeds with its finding, got %v", err)
	}
	resolved := strings.Join(current.Warnings, "; ")
	for _, want := range []string{"detached-pending", "does not exist yet", "pi"} {
		if !strings.Contains(resolved, want) {
			t.Fatalf("bare resolve carries the pending finding, want %q in %q", want, resolved)
		}
	}
	if strings.Contains(resolved, envregistry.DiagCredentialConflict) {
		t.Fatalf("the pending finding is not a conflict: %q", resolved)
	}
	marker := readManagedMarker(t, fx, "pi")
	if marker.Version != envmarker.VersionV2 || marker.Passthrough == nil || len(*marker.Passthrough) != 1 {
		t.Fatalf("one passthrough entry recorded: %+v", marker.Passthrough)
	}
	entry := (*marker.Passthrough)[0]
	if entry.Path != "auth.json" || entry.Isolation != envregistry.IsolationShared || entry.Strategy != envregistry.StrategyFileLink ||
		entry.SourceRole != "native" || entry.Backend != "file" || entry.BackendVersion != "0.84.2" || entry.Provenance != "provisioned" {
		t.Fatalf("recorded entry %+v", entry)
	}
	payload, err := json.Marshal(marker.Passthrough)
	if err != nil {
		t.Fatal(err)
	}
	var decoded []map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	for key := range decoded[0] {
		if key != "path" && key != "isolation" && key != "strategy" && key != "source_role" && key != "backend" && key != "backend_version" && key != "provenance" {
			t.Fatalf("schema-v2 credential record has an unknown field %q in %s", key, payload)
		}
	}
	// With the native credential present the link reads through.
	if err := os.MkdirAll(filepath.Join(fx.native["pi"], "agent"), 0o755); err != nil {
		t.Fatal(err)
	}
	nativeAuth := []byte("{\"t\":\"pi-credential\"}\n")
	if err := os.WriteFile(want, nativeAuth, 0o600); err != nil {
		t.Fatal(err)
	}
	if payload, err := os.ReadFile(link); err != nil || string(payload) != string(nativeAuth) {
		t.Fatalf("the link serves the agent-root bytes: %q (%v)", payload, err)
	}
	healed, err := Resolve(fx.request("pi"))
	if err != nil {
		t.Fatalf("the linked home is current: %v", err)
	}
	if warnings := strings.Join(healed.Warnings, "; "); strings.Contains(warnings, "detached-pending") {
		t.Fatalf("the pending finding clears once the native target exists: %q", warnings)
	}
}

// TestCredentialLinkDirectory pins the §7.4 directory rule through
// Resolve: an empty directory at a wanted link path is replaced by the
// link, while a non-empty one refuses with
// environment_credential_conflict and its contents are untouched.
func TestCredentialLinkDirectory(t *testing.T) {
	requireLinkCapability(t)
	t.Run("empty is re-linked", func(t *testing.T) {
		fx := writeManagedFixture(t, "acme")
		seedLiveNativeCredentials(t, fx)
		provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
		link := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
		_ = os.Remove(link)
		if err := os.Mkdir(link, 0o755); err != nil {
			t.Fatal(err)
		}
		if _, err := Resolve(fx.request("codex_cli")); err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
			t.Fatalf("a directory at the link path must be stale, got %v", err)
		}
		req := fx.request("codex_cli")
		req.Repair = true
		if _, err := Resolve(req); err != nil {
			t.Fatalf("repair re-links over an empty directory: %v", err)
		}
		if target, err := os.Readlink(link); err != nil || target != filepath.Join(fx.native["codex_cli"], "auth.json") {
			t.Fatalf("repair left %q (%v)", target, err)
		}
	})
	t.Run("non-empty refuses", func(t *testing.T) {
		fx := writeManagedFixture(t, "acme")
		seedLiveNativeCredentials(t, fx)
		provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
		link := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
		_ = os.Remove(link)
		if err := os.Mkdir(link, 0o755); err != nil {
			t.Fatal(err)
		}
		kept := filepath.Join(link, "kept")
		if err := os.WriteFile(kept, []byte("kept\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		req := fx.request("codex_cli")
		req.Repair = true
		_, err := Resolve(req)
		if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
			t.Fatalf("a non-empty directory must refuse, got %v", err)
		}
		if !strings.Contains(err.Error(), link) {
			t.Fatalf("the refusal names the path: %v", err)
		}
		if payload, err := os.ReadFile(kept); err != nil || string(payload) != "kept\n" {
			t.Fatalf("directory contents untouched: %q (%v)", payload, err)
		}
	})
}

// TestCredentialLinkUnrecordedSymlinkRefuses pins the recorded-only bound
// on the migration re-point: a symlink to an unexpected target at a
// wanted link path the marker does NOT record is foreign, so repair (and
// the migration) refuse with environment_credential_conflict naming the
// path, removing nothing. The setup turns a keyring-ambient codex home
// (nothing linked, nothing recorded) into a file-link home while the
// operator's own symlink sits at the link path. A mutant that drops the
// recorded check and re-points every symlink unlinks the foreign link
// and fails; the recorded re-point itself lives in the explicit
// migration (TestMigratePiWrongTargetToAgentRoot).
func TestCredentialLinkUnrecordedSymlinkRefuses(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "config.toml"), []byte("cli_auth_credentials_store = \"keyring\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
	marker := readManagedMarker(t, fx, "codex_cli")
	if marker.Passthrough == nil || len(*marker.Passthrough) != 1 {
		t.Fatalf("a keyring home records its linkless credential strategy: %+v", marker.Passthrough)
	}
	entry := (*marker.Passthrough)[0]
	if entry.Path != "" || entry.Isolation != envregistry.IsolationShared || entry.Strategy != envregistry.StrategyKeyringPreferred || entry.Backend != "ambient" || entry.SourceRole != "native" || entry.BackendVersion != "0.153.2" {
		t.Fatalf("keyring home records a pathless ambient credential: %+v", entry)
	}
	// The native store turns to file, and the operator's own symlink sits
	// at the now-wanted link path.
	if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "config.toml"), []byte("cli_auth_credentials_store = \"file\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "auth.json"), []byte("{\"t\":\"operator\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
	foreign := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(foreign, []byte("foreign\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(foreign, link); err != nil {
		t.Fatal(err)
	}
	req := fx.request("codex_cli")
	req.Repair = true
	_, err := Resolve(req)
	if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
		t.Fatalf("an unrecorded symlink at the wanted path must refuse, got %v", err)
	}
	if !strings.Contains(err.Error(), link) {
		t.Fatalf("the refusal names the path: %v", err)
	}
	// A recorded mis-targeted link refuses with "migration needed" (the
	// explicit step re-points it); an unrecorded link is foreign, so
	// the refusal must not point at migration — a mutant collapsing
	// the two refusals fails here while the Pi row still passes.
	if strings.Contains(err.Error(), "migration needed") {
		t.Fatalf("an unrecorded link is foreign, not a migration case: %v", err)
	}
	if target, err := os.Readlink(link); err != nil || target != foreign {
		t.Fatalf("the foreign link is untouched: %q (%v)", target, err)
	}
}

func strptr(value string) *string { return &value }

// TestReviewerDanglingExpectedTarget is the revision-1 reviewer's probe,
// revised under rework-2 as the detached-pending row: a recorded Pi link
// pointing at the declared agent-root target with no native file is
// REPORTED — never silently current — but no longer fails. StatusOf and
// Resolve both surface the detached-pending finding; Resolve succeeds
// with it as a warning. The probe's original stale-plus-conflict
// expectation now belongs to the mis-targeted state only
// (TestDanglingPiLinkReportedDetached).
func TestReviewerDanglingExpectedTarget(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	status, err := StatusOf(statusRequest(fx))
	if err != nil {
		t.Fatal(err)
	}
	row := findHome(status, "acme", "pi")
	if row == nil {
		t.Fatal("the pi home row is missing")
	}
	reported := strings.Join(row.Findings, ";") + ";" + strings.Join(row.Warnings, ";")
	for _, want := range []string{"detached-pending", "does not exist yet"} {
		if !strings.Contains(reported, want) {
			t.Errorf("dangling expected target must be reported, want %q in %+v", want, row)
		}
	}
	if strings.Contains(reported, envregistry.DiagCredentialConflict) {
		t.Errorf("the pending finding is not a conflict: %+v", row)
	}
	current, err := Resolve(fx.request("pi"))
	if err != nil {
		t.Errorf("Resolve surfaces the pending finding without failing, got %v", err)
	} else if warnings := strings.Join(current.Warnings, "; "); !strings.Contains(warnings, "detached-pending") {
		t.Errorf("Resolve warnings carry the pending finding: %q", warnings)
	}
}

// TestReviewerLiteralCodexSelector is the revision-1 reviewer's probe,
// committed verbatim: valid TOML literal spellings of the native
// cli_auth_credentials_store selector refuse exactly like the
// double-quoted spelling — keyring and auto with
// environment_isolated_unsupported, any other value with
// environment_credential_unsupported — never read as absent.
func TestReviewerLiteralCodexSelector(t *testing.T) {
	for _, value := range []string{"keyring", "auto", "ephemeral"} {
		t.Run(value, func(t *testing.T) {
			fx := writeManagedFixture(t, "acme")
			config := "cli_auth_credentials_store = '" + value + "'\n"
			if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "config.toml"), []byte(config), 0600); err != nil {
				t.Fatal(err)
			}
			req := fx.request("codex_cli")
			req.Repair = true
			req.Machine = envregistry.DefaultMachineConfig()
			req.Machine.Isolation = map[string]map[string]string{"acme": {"codex_cli": "isolated"}}
			_, err := Resolve(req)
			want := envregistry.DiagIsolatedUnsupported
			if value == "ephemeral" {
				want = envregistry.DiagCredentialUnsupported
			}
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("valid TOML literal selector must refuse with %s, got %v", want, err)
			}
		})
	}
}

// TestDanglingExpectedTargetRepairSucceeds pins the repair half of the
// detached-pending contract (rework-2): a correctly targeted link whose
// native target went missing is already exactly what repair would
// establish, so repair leaves the link untouched and SUCCEEDS, carrying
// the pending finding as a warning — restoring the native target clears
// it. This is the narrowing partner of TestDanglingPiLinkReportedDetached:
// a mutant that collapses mis-targeted and dangling-to-declared into one
// state fails one of the two — stale-plus-conflict here breaks the
// success below, pending-warning there breaks the Pi staleness.
func TestDanglingExpectedTargetRepairSucceeds(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	nativeAuth := filepath.Join(fx.native["codex_cli"], "auth.json")
	if err := os.WriteFile(nativeAuth, []byte("{\"t\":\"operator\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
	link := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
	if _, err := Resolve(fx.request("codex_cli")); err != nil {
		t.Fatalf("the live home is current: %v", err)
	}
	// The native credential goes missing: the link dangles with the
	// expected target.
	if err := os.Remove(nativeAuth); err != nil {
		t.Fatal(err)
	}
	current, err := Resolve(fx.request("codex_cli"))
	if err != nil {
		t.Fatalf("a dangling-to-declared link succeeds with its finding, got %v", err)
	}
	warnings := strings.Join(current.Warnings, "; ")
	for _, want := range []string{"detached-pending", "does not exist yet", "auth.json"} {
		if !strings.Contains(warnings, want) {
			t.Fatalf("resolve warnings carry the pending finding, want %q in %q", want, warnings)
		}
	}
	if strings.Contains(warnings, envregistry.DiagCredentialConflict) {
		t.Fatalf("the pending finding is not a conflict: %q", warnings)
	}
	req := fx.request("codex_cli")
	req.Repair = true
	repaired, err := Resolve(req)
	if err != nil {
		t.Fatalf("repair of a dangling-to-declared link succeeds: %v", err)
	}
	if warnings := strings.Join(repaired.Warnings, "; "); !strings.Contains(warnings, "detached-pending") {
		t.Fatalf("repair carries the pending finding: %q", warnings)
	}
	if target, err := os.Readlink(link); err != nil || target != nativeAuth {
		t.Fatalf("repair touches nothing: %q (%v)", target, err)
	}
	// Restoring the native target clears the finding.
	if err := os.WriteFile(nativeAuth, []byte("{\"t\":\"operator\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	healed, err := Resolve(req)
	if err != nil {
		t.Fatalf("repair converges once the native target is back: %v", err)
	}
	if warnings := strings.Join(healed.Warnings, "; "); strings.Contains(warnings, "detached-pending") {
		t.Fatalf("the pending finding clears once the native target is back: %q", warnings)
	}
	if _, err := Resolve(fx.request("codex_cli")); err != nil {
		t.Fatalf("the healed home is current: %v", err)
	}
}

// TestCredentialLinkTargetInspectionFailure pins the absent-versus-
// uninspectable distinction: a correctly targeted link whose native
// target cannot be inspected is a conflict/inspection diagnostic — never
// silence, and never the "does not exist" absence wording. A regular file
// where the agent directory should be fails the target stat
// deterministically on Unix, even as root; Windows maps a stat through a
// file to path-not-found, so the row is Unix-only and Windows keeps the
// absence wording there as a stated bound.
func TestCredentialLinkTargetInspectionFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("this host cannot create an uninspectable link target: Windows maps a stat through a file to path-not-found, so the inspection diagnostic runs on the unix runners")
	}
	fx := writeManagedFixture(t, "acme")
	native := fx.native["pi"]
	agentAuth := filepath.Join(native, "agent", "auth.json")
	if err := os.MkdirAll(filepath.Dir(agentAuth), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agentAuth, []byte("{\"t\":\"operator-pi\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	if _, err := Resolve(fx.request("pi")); err != nil {
		t.Fatalf("the live home is current: %v", err)
	}
	// Break the target's parent: a file where the agent directory was.
	if err := os.RemoveAll(filepath.Join(native, "agent")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(native, "agent"), []byte("not a directory\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stale, err := Resolve(fx.request("pi"))
	if err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
		t.Fatalf("an uninspectable target must be stale, got %v", err)
	}
	reasons := strings.Join(stale.StaleReasons, "; ")
	for _, want := range []string{"cannot be inspected", envregistry.DiagCredentialConflict, "auth.json"} {
		if !strings.Contains(reasons, want) {
			t.Fatalf("stale reasons carry the inspection diagnostic, want %q in %q", want, reasons)
		}
	}
	if strings.Contains(reasons, "does not exist") {
		t.Fatalf("an inspection failure is not absence: %q", reasons)
	}
	status, err := StatusOf(statusRequest(fx))
	if err != nil {
		t.Fatal(err)
	}
	row := findHome(status, "acme", "pi")
	if row == nil || row.Current {
		t.Fatalf("the uninspectable home is non-current: %+v", row)
	}
	findings := strings.Join(row.Findings, "; ")
	for _, want := range []string{"cannot be inspected", envregistry.DiagCredentialConflict} {
		if !strings.Contains(findings, want) {
			t.Fatalf("status findings carry the inspection diagnostic, want %q in %q", want, findings)
		}
	}
}

// TestCodexCredentialStoreTOMLSpellings drives the native selector reader
// through isolated Resolve at the production entry: every valid TOML
// spelling of the top-level key resolves identically, a nested same-named
// key is not the selector and stays absent, and malformed or mistyped
// config fails closed instead of reading as absent. Narrowing killers: a
// mutant that drops single-quote handling admits the literal rows; one
// that matches nested keys refuses the table rows; one that defaults
// invalid TOML to file admits the malformed row.
func TestCodexCredentialStoreTOMLSpellings(t *testing.T) {
	for _, tc := range []struct {
		name    string
		config  string
		wantErr string
	}{
		{"single-quoted file admits", "cli_auth_credentials_store = 'file'\n", ""},
		{"multi-line basic keyring refuses", "cli_auth_credentials_store = \"\"\"keyring\"\"\"\n", envregistry.DiagIsolatedUnsupported},
		{"multi-line literal auto refuses", "cli_auth_credentials_store = '''auto'''\n", envregistry.DiagIsolatedUnsupported},
		{"indented keyring refuses", "  cli_auth_credentials_store = \"keyring\"\n", envregistry.DiagIsolatedUnsupported},
		{"trailing comment auto refuses", "cli_auth_credentials_store = \"auto\" # operator note\n", envregistry.DiagIsolatedUnsupported},
		{"nested table key stays absent", "[profile]\ncli_auth_credentials_store = \"keyring\"\n", ""},
		{"dotted key stays absent", "profile.cli_auth_credentials_store = \"keyring\"\n", ""},
		{"invalid TOML fails closed", "cli_auth_credentials_store = \n", "not valid TOML"},
		{"integer selector fails closed", "cli_auth_credentials_store = 42\n", envregistry.DiagCredentialUnsupported},
		{"boolean selector fails closed", "cli_auth_credentials_store = true\n", envregistry.DiagCredentialUnsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := writeManagedFixture(t, "acme")
			if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "config.toml"), []byte(tc.config), 0o644); err != nil {
				t.Fatal(err)
			}
			machine := envregistry.DefaultMachineConfig()
			machine.Isolation = map[string]map[string]string{"acme": {"codex_cli": "isolated"}}
			req := fx.request("codex_cli")
			req.Machine = machine
			req.Repair = true
			_, err := Resolve(req)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("isolated must be admitted: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("want %s, got %v", tc.wantErr, err)
			}
		})
	}
}

// TestCodexMalformedStoreSharedRefuses pins the shared side of malformed
// native config: invalid TOML and non-string selectors fail closed with a
// diagnostic naming the file before the first write — never the file
// default — so fixing the config and re-running provisions cleanly.
func TestCodexMalformedStoreSharedRefuses(t *testing.T) {
	requireLinkCapability(t)
	for _, tc := range []struct {
		name      string
		config    string
		wantErr   string
		namesFile bool
	}{
		{"invalid TOML", "cli_auth_credentials_store = \n", "not valid TOML", true},
		{"integer selector", "cli_auth_credentials_store = 42\n", envregistry.DiagCredentialUnsupported, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := writeManagedFixture(t, "acme")
			config := filepath.Join(fx.native["codex_cli"], "config.toml")
			if err := os.WriteFile(config, []byte(tc.config), 0o644); err != nil {
				t.Fatal(err)
			}
			req := fx.request("codex_cli")
			req.Repair = true
			_, err := Resolve(req)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("malformed config must fail closed, got %v", err)
			}
			// A file-level failure names the file; a value-level
			// failure names the key, like the unknown-selector
			// diagnostic it shares the class with.
			if tc.namesFile && !strings.Contains(err.Error(), "config.toml") {
				t.Fatalf("the refusal names the file: %v", err)
			}
			if !tc.namesFile && !strings.Contains(err.Error(), "cli_auth_credentials_store") {
				t.Fatalf("the refusal names the key: %v", err)
			}
			if _, err := os.Lstat(ManagedHomeDir(fx.home, "acme", "codex_cli")); !os.IsNotExist(err) {
				t.Fatalf("a refused provisioning writes nothing: %v", err)
			}
			if err := os.WriteFile(config, []byte("cli_auth_credentials_store = \"file\"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := Resolve(req); err != nil {
				t.Fatalf("fixing the config provisions: %v", err)
			}
		})
	}
}
