package envprofile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
)

const codexSeedVectorFile = "environments-codex-seed.json"

type codexSeedProvisioningVector struct {
	Name         string `json:"name"`
	Revision     string `json:"revision"`
	NativeConfig string `json:"native_config_toml"`
	Expected     struct {
		SeededHasMCP    bool                       `json:"seeded_has_mcp_servers"`
		SeededTopLevel  []string                   `json:"seeded_top_level_members"`
		SeededMembers   map[string]any             `json:"seeded_members"`
		Names           []string                   `json:"names"`
		Diagnostic      *string                    `json:"diagnostic"`
		MigrationHint   bool                       `json:"migration_hint"`
		CodexSeedRecord *envmarker.CodexSeedRecord `json:"codex_seed_record"`
	} `json:"expected"`
}

type codexSeedPostureVector struct {
	Name            string                     `json:"name"`
	RevisionShipped string                     `json:"revision_shipped"`
	Environment     string                     `json:"environment"`
	Record          *envmarker.CodexSeedRecord `json:"codex_seed_record"`
	Expected        struct {
		CodexSeedRow string   `json:"codex_seed_row"`
		Diagnostics  []string `json:"status_diagnostics"`
		NamesListed  []string `json:"names_listed"`
		ListedAs     string   `json:"listed_as"`
		RepairHint   bool     `json:"repair_hint"`
		RowCurrent   bool     `json:"row_current"`
	} `json:"expected"`
}

type codexSeedVectorSet struct {
	Provisioning []codexSeedProvisioningVector `json:"provisioning_cases"`
	Posture      []codexSeedPostureVector      `json:"posture_cases"`
}

// TestCodexSeedProvisioningAndStatus exercises Resolve and StatusOf, the
// production entry points for revision-B Codex seed handling. It also pins
// malformed native config to a fail-closed provisioning result.
func TestCodexSeedProvisioningAndStatus(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	native := "model = \"gpt-5-codex\"\n\n[projects.\"/Users/operator/work\"]\ntrust_level = \"trusted\"\n\n[tui]\ntheme = \"dark\"\n\n[mcp_servers.gh]\ncommand = \"gh\"\nargs = [\"mcp\"]\n\n[mcp_servers.figma]\ncommand = \"npx\"\nargs = [\"-y\", \"figma-developer-mcp\", \"--stdio\"]\n"
	if err := os.WriteFile(filepath.Join(fx.native[envregistry.CodexCLI], "config.toml"), []byte(native), 0o600); err != nil {
		t.Fatal(err)
	}

	req := fx.request(envregistry.CodexCLI)
	req.Repair = true
	result, err := Resolve(req)
	if err != nil {
		t.Fatalf("Resolve provisioning: %v", err)
	}
	if !result.Provisioned {
		t.Fatal("Resolve did not report first provisioning")
	}
	if got := countWarnings(result.Warnings, envregistry.DiagMCPNativeServersNotInherited); got != 1 {
		t.Fatalf("provision warnings = %v; mcp_native_servers_not_inherited count = %d, want 1", result.Warnings, got)
	}
	if warning := warningWithPrefix(result.Warnings, envregistry.DiagMCPNativeServersNotInherited); !strings.Contains(warning, "figma, gh") {
		t.Fatalf("provision warning %q does not name the sorted native servers", warning)
	}

	managedHome := ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI)
	seededPayload, err := os.ReadFile(filepath.Join(managedHome, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	var seeded map[string]any
	if err := toml.Unmarshal(seededPayload, &seeded); err != nil {
		t.Fatalf("managed config.toml is not valid TOML: %v", err)
	}
	if _, exists := seeded["mcp_servers"]; exists {
		t.Fatalf("managed config.toml retained mcp_servers: %s", seededPayload)
	}
	var original map[string]any
	if err := toml.Unmarshal([]byte(native), &original); err != nil {
		t.Fatal(err)
	}
	delete(original, "mcp_servers")
	if !reflect.DeepEqual(seeded, original) {
		t.Fatalf("managed config.toml members = %#v, want all native members except mcp_servers %#v", seeded, original)
	}

	marker := readManagedMarker(t, fx, envregistry.CodexCLI)
	wantRecord := &envmarker.CodexSeedRecord{Revision: envregistry.CodexSeedRevisionB, NativeMCPServers: []string{"figma", "gh"}}
	if !reflect.DeepEqual(marker.CodexSeedRecord, wantRecord) {
		t.Fatalf("codex_seed_record = %+v, want %+v", marker.CodexSeedRecord, wantRecord)
	}

	status, err := StatusOf(statusRequest(fx))
	if err != nil {
		t.Fatal(err)
	}
	if status.CodexSeedRule.Revision != envregistry.CodexSeedRevisionB || status.CodexSeedRule.Provenance != "shipped" {
		t.Fatalf("codex-seed status row = %+v", status.CodexSeedRule)
	}
	home := findHome(status, fx.profile, envregistry.CodexCLI)
	if home == nil {
		t.Fatal("env status omitted the managed codex_cli home")
	}
	if !home.Current {
		t.Fatalf("Codex seed warnings made the home non-current: %+v", home)
	}
	if !reflect.DeepEqual(home.NativeMCPServers, wantRecord.NativeMCPServers) || home.NativeMCPServersDisposition != "not-inherited" {
		t.Fatalf("env status Codex seed row = %+v", home)
	}
	if got := countWarnings(home.Warnings, envregistry.DiagMCPNativeServersNotInherited); got != 1 {
		t.Fatalf("env status warnings = %v; not-inherited count = %d, want 1", home.Warnings, got)
	}
}

func TestCodexSeedStripsInlineMCPTable(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	native := "model = \"gpt-5-codex\"\nmcp_servers = { local = { command = \"server\", args = [] } }\n"
	if err := os.WriteFile(filepath.Join(fx.native[envregistry.CodexCLI], "config.toml"), []byte(native), 0o600); err != nil {
		t.Fatal(err)
	}
	req := fx.request(envregistry.CodexCLI)
	req.Repair = true
	result, err := Resolve(req)
	if err != nil {
		t.Fatalf("Resolve provisioning: %v", err)
	}
	if got := countWarnings(result.Warnings, envregistry.DiagMCPNativeServersNotInherited); got != 1 {
		t.Fatalf("provision warnings = %v; not-inherited count = %d, want 1", result.Warnings, got)
	}
	managedHome := ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI)
	seededPayload, err := os.ReadFile(filepath.Join(managedHome, "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	var seeded map[string]any
	if err := toml.Unmarshal(seededPayload, &seeded); err != nil {
		t.Fatalf("managed config.toml is not valid TOML: %v", err)
	}
	if _, exists := seeded["mcp_servers"]; exists {
		t.Fatalf("managed config.toml retained inline mcp_servers: %s", seededPayload)
	}
	want := &envmarker.CodexSeedRecord{Revision: envregistry.CodexSeedRevisionB, NativeMCPServers: []string{"local"}}
	if got := readManagedMarker(t, fx, envregistry.CodexCLI).CodexSeedRecord; !reflect.DeepEqual(got, want) {
		t.Fatalf("codex_seed_record = %+v, want %+v", got, want)
	}
}

func TestCodexSeedRejectsInvalidTOMLBeforePublishingHome(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	if err := os.WriteFile(filepath.Join(fx.native[envregistry.CodexCLI], "config.toml"), []byte("[mcp_servers.figma\ncommand = \"npx\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	req := fx.request(envregistry.CodexCLI)
	req.Repair = true
	if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), envregistry.DiagSeedUnreadable) {
		t.Fatalf("Resolve error = %v, want %s refusal", err, envregistry.DiagSeedUnreadable)
	}
	if _, err := os.Lstat(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI)); !os.IsNotExist(err) {
		t.Fatalf("invalid native config published managed home: lstat error = %v", err)
	}
}

func TestCodexSeedSnapshotAndBytesDoNotRefreshOnRepair(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	native := "model = \"gpt-5-codex\"\n\n[mcp_servers.original]\ncommand = \"server\"\nargs = []\n"
	if err := os.WriteFile(filepath.Join(fx.native[envregistry.CodexCLI], "config.toml"), []byte(native), 0o600); err != nil {
		t.Fatal(err)
	}
	req := fx.request(envregistry.CodexCLI)
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	managedHome := ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI)
	seedPath := filepath.Join(managedHome, "config.toml")
	before, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatal(err)
	}
	beforeRecord := readManagedMarker(t, fx, envregistry.CodexCLI).CodexSeedRecord
	if err := os.WriteFile(filepath.Join(fx.native[envregistry.CodexCLI], "config.toml"), []byte("model = \"changed\"\n\n[mcp_servers.later]\ncommand = \"server\"\nargs = []\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(managedHome, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	result, err := Resolve(req)
	if err != nil {
		t.Fatal(err)
	}
	if got := countWarnings(result.Warnings, envregistry.DiagMCPNativeServersNotInherited); got != 0 {
		t.Fatalf("repair repeated the provisioning warning: %v", result.Warnings)
	}
	after, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("repair refreshed the Codex seed: before %q, after %q", before, after)
	}
	if got := readManagedMarker(t, fx, envregistry.CodexCLI).CodexSeedRecord; !reflect.DeepEqual(got, beforeRecord) {
		t.Fatalf("repair refreshed codex_seed_record: before %+v, after %+v", beforeRecord, got)
	}
}

// TestEnvironmentsCodexSeedVectors drives the published rc.13 revision-B
// provisioning vectors and every revision-B status vector through Resolve and
// StatusOf. Revision-A producer/status postures are prior-manager behavior;
// they are explicitly reported as bounded while A-record homes under the
// shipped B manager remain production-driven.
func TestEnvironmentsCodexSeedVectors(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	path := filepath.Join(root, "vectors", codexSeedVectorFile)
	payload, err := os.ReadFile(path) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var vectors codexSeedVectorSet
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	if len(vectors.Provisioning) == 0 || len(vectors.Posture) == 0 {
		t.Fatalf("%s is missing provisioning or posture cases", codexSeedVectorFile)
	}
	conformancecoverage.RunOutcomes(t, "environments-codex-seed/provisioning-cases", vectors.Provisioning,
		func(tc codexSeedProvisioningVector) string { return tc.Name }, func(t *testing.T, tc codexSeedProvisioningVector) conformancecoverage.Observation {
			if tc.Revision != envregistry.CodexSeedRevisionB {
				return conformancecoverage.Observation{BoundReason: "revision A is prior-manager provisioning behavior; this manager ships revision B"}
			}
			runCodexSeedProvisioningVector(t, tc)
			return conformancecoverage.Observation{}
		})
	conformancecoverage.RunOutcomes(t, "environments-codex-seed/posture-cases", vectors.Posture,
		func(tc codexSeedPostureVector) string { return tc.Name }, func(t *testing.T, tc codexSeedPostureVector) conformancecoverage.Observation {
			if tc.RevisionShipped != envregistry.CodexSeedRevisionB {
				return conformancecoverage.Observation{BoundReason: "revision A is prior-manager status behavior; this manager ships revision B"}
			}
			runCodexSeedPostureVector(t, tc)
			return conformancecoverage.Observation{}
		})
}

func runCodexSeedProvisioningVector(t *testing.T, tc codexSeedProvisioningVector) {
	t.Helper()
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	if err := os.WriteFile(filepath.Join(fx.native[envregistry.CodexCLI], "config.toml"), []byte(tc.NativeConfig), 0o600); err != nil {
		t.Fatal(err)
	}
	req := fx.request(envregistry.CodexCLI)
	req.Repair = true
	result, err := Resolve(req)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !result.Provisioned {
		t.Fatal("Resolve did not provision the Codex home")
	}
	seedPath := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI), "config.toml")
	seedBytes, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatal(err)
	}
	var seeded map[string]any
	if err := toml.Unmarshal(seedBytes, &seeded); err != nil {
		t.Fatalf("seeded config.toml is not valid TOML: %v", err)
	}
	if _, exists := seeded["mcp_servers"]; exists != tc.Expected.SeededHasMCP {
		t.Fatalf("seeded_has_mcp_servers = %v, want %v", exists, tc.Expected.SeededHasMCP)
	}
	topLevel := make([]string, 0, len(seeded))
	for key := range seeded {
		topLevel = append(topLevel, key)
	}
	sort.Strings(topLevel)
	if !reflect.DeepEqual(topLevel, tc.Expected.SeededTopLevel) {
		t.Fatalf("seeded top-level members = %v, want %v", topLevel, tc.Expected.SeededTopLevel)
	}
	if !reflect.DeepEqual(seeded, tc.Expected.SeededMembers) {
		t.Fatalf("seeded members = %#v, want %#v", seeded, tc.Expected.SeededMembers)
	}
	marker := readManagedMarker(t, fx, envregistry.CodexCLI)
	if !reflect.DeepEqual(marker.CodexSeedRecord, tc.Expected.CodexSeedRecord) {
		t.Fatalf("codex_seed_record = %+v, want %+v", marker.CodexSeedRecord, tc.Expected.CodexSeedRecord)
	}
	if !reflect.DeepEqual(marker.CodexSeedRecord.NativeMCPServers, tc.Expected.Names) {
		t.Fatalf("recorded native server names = %v, want %v", marker.CodexSeedRecord.NativeMCPServers, tc.Expected.Names)
	}
	if tc.Expected.Diagnostic == nil {
		if got := countWarnings(result.Warnings, envregistry.DiagMCPNativeServersNotInherited); got != 0 {
			t.Fatalf("provision warnings = %v; no stripped-server warning expected", result.Warnings)
		}
	} else {
		if *tc.Expected.Diagnostic != envregistry.DiagMCPNativeServersNotInherited {
			t.Fatalf("unexpected revision-B diagnostic %q", *tc.Expected.Diagnostic)
		}
		if got := countWarnings(result.Warnings, *tc.Expected.Diagnostic); got != 1 {
			t.Fatalf("provision warnings = %v; %s count = %d, want 1", result.Warnings, *tc.Expected.Diagnostic, got)
		}
		warning := warningWithPrefix(result.Warnings, *tc.Expected.Diagnostic)
		for _, name := range tc.Expected.Names {
			if !strings.Contains(warning, name) {
				t.Fatalf("provision warning %q omits native server %q", warning, name)
			}
		}
	}
	if tc.Expected.MigrationHint {
		t.Fatalf("revision-B warning unexpectedly requires a migration hint: %v", result.Warnings)
	}
}

func runCodexSeedPostureVector(t *testing.T, tc codexSeedPostureVector) {
	t.Helper()
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	provision(t, fx, tc.Environment, envregistry.DefaultMachineConfig())
	marker := readManagedMarker(t, fx, tc.Environment)
	marker.CodexSeedRecord = cloneCodexSeedRecord(tc.Record)
	if tc.Record != nil && tc.Record.Revision == "A" && len(tc.Record.NativeMCPServers) > 0 {
		var config bytes.Buffer
		for _, name := range tc.Record.NativeMCPServers {
			fmt.Fprintf(&config, "\n[mcp_servers.%q]\ncommand = \"server\"\nargs = []\n", name)
		}
		if err := os.WriteFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, tc.Environment), "config.toml"), config.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	markerBytes, err := marker.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	markerPath := filepath.Join(ManagedHomeDir(fx.home, fx.profile, tc.Environment), envmarker.Name)
	if err := os.WriteFile(markerPath, markerBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	status, err := StatusOf(statusRequest(fx))
	if err != nil {
		t.Fatal(err)
	}
	if status.CodexSeedRule.Revision != tc.Expected.CodexSeedRow {
		t.Fatalf("codex-seed row revision = %q, want %q", status.CodexSeedRule.Revision, tc.Expected.CodexSeedRow)
	}
	home := findHome(status, fx.profile, tc.Environment)
	if home == nil {
		t.Fatal("env status omitted the expected home")
	}
	if home.Current != tc.Expected.RowCurrent {
		t.Fatalf("home current = %v, want %v; findings=%v", home.Current, tc.Expected.RowCurrent, home.Findings)
	}
	if !equalOrdered(home.NativeMCPServers, tc.Expected.NamesListed) {
		t.Fatalf("listed native servers = %v, want %v", home.NativeMCPServers, tc.Expected.NamesListed)
	}
	if home.NativeMCPServersDisposition != tc.Expected.ListedAs {
		t.Fatalf("listed-as disposition = %q, want %q", home.NativeMCPServersDisposition, tc.Expected.ListedAs)
	}
	var diagnostics []string
	repairHint := false
	for _, warning := range home.Warnings {
		for _, code := range []string{envregistry.DiagMCPNativeServersUngoverned, envregistry.DiagMCPNativeServersNotInherited, envregistry.DiagMCPSeedUnstripped} {
			if strings.HasPrefix(warning, code+":") {
				diagnostics = append(diagnostics, code)
				if code == envregistry.DiagMCPSeedUnstripped && strings.Contains(warning, "re-provision") {
					repairHint = true
				}
			}
		}
	}
	if !equalOrdered(diagnostics, tc.Expected.Diagnostics) {
		t.Fatalf("status diagnostics = %v, want %v; warnings=%v", diagnostics, tc.Expected.Diagnostics, home.Warnings)
	}
	if repairHint != tc.Expected.RepairHint {
		t.Fatalf("repair hint = %v, want %v; warnings=%v", repairHint, tc.Expected.RepairHint, home.Warnings)
	}
}

func countWarnings(warnings []string, code string) int {
	count := 0
	for _, warning := range warnings {
		if strings.HasPrefix(warning, code+":") {
			count++
		}
	}
	return count
}

func warningWithPrefix(warnings []string, code string) string {
	for _, warning := range warnings {
		if strings.HasPrefix(warning, code+":") {
			return warning
		}
	}
	return ""
}
