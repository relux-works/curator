package envprofile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/envregistry"
)

// Production entry points under test: Install, UpdateWithPolicy, Resolve,
// StatusOf. Every fixture builds a root -> mid -> leaf git closure where
// the leaf carries a class: system module, so the leaf is transitive and
// its system module is admitted only by waiver.

// admissionRepos serves a sysroot -> sysmid -> sysleaf closure. The leaf
// always carries a system module; rootSystem and midSystem select whether
// the root and mid carry one too.
func admissionRepos(t *testing.T, ids *gitIdentities, rootSystem, midSystem bool) string {
	t.Helper()
	leaf := gitRepo(t, map[string]string{
		"agent-context.json": `{"schema_version": 1, "name": "sysleaf", "version": "1.0.0",` +
			`"context": {"modules": [{"path": "90-system.md", "class": "system"}]}}` + "\n",
		"context/90-system.md": "Leaf system prompt.\n",
	}, "v1.0.0")
	leafURL := ids.serve(leaf, "https://example.com/sysleaf")
	midModules := `[{"path": "00-mid.md"}]`
	midFiles := map[string]string{
		"context/00-mid.md": "# Mid\n\nMid context.\n",
	}
	if midSystem {
		midModules = `[{"path": "00-mid.md"}, {"path": "90-system.md", "class": "system"}]`
		midFiles["context/90-system.md"] = "Mid system prompt.\n"
	}
	midFiles["agent-context.json"] = `{"schema_version": 1, "name": "sysmid", "version": "1.0.0",` +
		`"context": {"modules": ` + midModules + `},` +
		`"requires": {"contexts": {"sysleaf": {"git": "` + leafURL + `", "range": "*"}}}}` + "\n"
	mid := gitRepo(t, midFiles, "v1.0.0")
	midURL := ids.serve(mid, "https://example.com/sysmid")
	rootModules := `[{"path": "00-root.md"}]`
	rootFiles := map[string]string{
		"context/00-root.md": "# Sysroot\n\nRoot context.\n",
	}
	if rootSystem {
		rootModules = `[{"path": "00-root.md"}, {"path": "90-system.md", "class": "system"}]`
		rootFiles["context/90-system.md"] = "Root system prompt.\n"
	}
	rootFiles["agent-context.json"] = `{"schema_version": 1, "name": "sysroot", "version": "1.0.0",` +
		`"context": {"modules": ` + rootModules + `},` +
		`"requires": {"contexts": {"sysmid": {"git": "` + midURL + `", "range": "*"}}}}` + "\n"
	root := gitRepo(t, rootFiles, "v1.0.0")
	return ids.serve(root, "https://example.com/sysroot")
}

func admissionPolicy() Policy {
	return Policy{TransitiveSystemModules: contextmaterialize.TransitiveDrop}
}

// TestInstallErrorRefusesTransitiveSystemModule drives the production
// Install under the error policy: resolution fails with
// context_system_module_transitive naming the transitive package and
// module, and no lock is written.
func TestInstallErrorRefusesTransitiveSystemModule(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	operand := admissionRepos(t, ids, true, true)
	policy := admissionPolicy()
	policy.TransitiveSystemModules = contextmaterialize.TransitiveError
	_, _, _, err := Install(home, InstallOptions{Operand: operand, Policy: policy})
	if err == nil {
		t.Fatalf("error policy installed a transitive system module")
	}
	var refusal *contextmaterialize.TransitiveSystemModuleError
	if !errors.As(err, &refusal) {
		t.Fatalf("err %v is not the %s refusal", err, contextmaterialize.DiagSystemModuleTransitive)
	}
	if refusal.Package != "sysleaf" || refusal.Module != "90-system.md" {
		t.Fatalf("refusal %+v names neither the package nor the module", refusal)
	}
	if _, statErr := os.Stat(lockPath(home, "sysroot")); !os.IsNotExist(statErr) {
		t.Fatalf("refused install wrote a lock: %v", statErr)
	}
}

// TestInstallErrorWaiverAdmits drives the production Install under the
// error policy with a waiver for the transitive package: the waiver
// admits the module and the install succeeds.
func TestInstallErrorWaiverAdmits(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	operand := admissionRepos(t, ids, true, true)
	policy := admissionPolicy()
	policy.TransitiveSystemModules = contextmaterialize.TransitiveError
	policy.SystemModuleWaivers = []SystemModuleWaiver{{Package: "sysleaf", Reason: "reviewed"}}
	info, _, _, err := Install(home, InstallOptions{Operand: operand, Policy: policy})
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Lock.Members) != 3 {
		t.Fatalf("lock members %+v", info.Lock.Members)
	}
}

// TestUpdateErrorLeavesLockUnchanged installs under drop, then updates
// under error: the update fails with the refusal and the old lock bytes
// are untouched.
func TestUpdateErrorLeavesLockUnchanged(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	operand := admissionRepos(t, ids, true, true)
	if _, _, _, err := Install(home, InstallOptions{Operand: operand, Policy: admissionPolicy()}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(lockPath(home, "sysroot"))
	if err != nil {
		t.Fatal(err)
	}
	policy := admissionPolicy()
	policy.TransitiveSystemModules = contextmaterialize.TransitiveError
	_, _, err = UpdateWithPolicy(home, "sysroot", policy)
	if err == nil {
		t.Fatalf("error policy updated past a transitive system module")
	}
	if !strings.Contains(err.Error(), contextmaterialize.DiagSystemModuleTransitive) ||
		!strings.Contains(err.Error(), "sysleaf") || !strings.Contains(err.Error(), "90-system.md") {
		t.Fatalf("err %v names neither the diagnostic, the package, nor the module", err)
	}
	after, err := os.ReadFile(lockPath(home, "sysroot"))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("refused update changed the lock:\n%s\n%s", before, after)
	}
}

// admissionNative makes the one native-home base shared by repair and
// status. The passthrough liveness row compares the recorded link target
// against the effective one, so a split base reports every file-link
// entry detached — visible on Linux, where claude_code carries the
// .credentials.json entry (darwin and windows link nothing).
func admissionNative(t *testing.T) func(string) (string, error) {
	t.Helper()
	native := t.TempDir()
	return func(id string) (string, error) {
		dir := filepath.Join(native, id)
		_ = os.MkdirAll(dir, 0o755)
		return dir, nil
	}
}

// admissionResolve provisions the managed home for envID and returns the
// resolve result: the production Resolve with repair under policy. The
// caller supplies the native-home base so status can re-verify against
// the same one.
func admissionResolve(t *testing.T, home, profile, envID string, policy Policy, nativeHomeOf func(string) (string, error)) *ResolveResult {
	t.Helper()
	result, err := Resolve(ResolveRequest{
		Home:         home,
		Profile:      profile,
		EnvID:        envID,
		LaunchDir:    t.TempDir(),
		Machine:      envregistry.DefaultMachineConfig(),
		Repair:       true,
		Format:       "json",
		Policy:       policy,
		Detect:       func(envregistry.Adapter) string { return "unknown" },
		NativeHomeOf: nativeHomeOf,
		OperatorXDG:  t.TempDir(),
	})
	if err != nil {
		t.Fatalf("resolve --repair: %v", err)
	}
	return result
}

// TestDropRepairWarnsAndFragmentFollowsAdmitted installs under drop and
// repairs the managed home: the repair warns
// context_system_module_dropped naming package and module, the inert
// system-prompt file holds exactly the admitted modules' bytes, and the
// fragment carries the system_prompt section for the admitted set.
func TestDropRepairWarnsAndFragmentFollowsAdmitted(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	operand := admissionRepos(t, ids, true, true)
	info, _, _, err := Install(home, InstallOptions{Operand: operand, Policy: admissionPolicy()})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, warning := range info.Warnings {
		if strings.Contains(warning, "context-system-module-present") && strings.Contains(warning, "sysleaf") {
			found = true
		}
	}
	if !found {
		t.Fatalf("install warnings %+v lack the always-warn finding for the transitive member", info.Warnings)
	}
	result := admissionResolve(t, home, "sysroot", "claude_code", admissionPolicy(), admissionNative(t))
	warned := false
	for _, warning := range result.Warnings {
		if strings.Contains(warning, contextmaterialize.DiagSystemModuleDropped) &&
			strings.Contains(warning, "sysleaf") && strings.Contains(warning, "90-system.md") {
			warned = true
		}
	}
	if !warned {
		t.Fatalf("resolve warnings %+v lack the drop warning naming package and module", result.Warnings)
	}
	payload, err := os.ReadFile(filepath.Join(ManagedHomeDir(home, "sysroot", "claude_code"), ".agent-context", "system-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	if want := "Mid system prompt.\n\nRoot system prompt.\n"; string(payload) != want {
		t.Fatalf("system-prompt.md %q, want exactly the admitted modules' bytes %q", payload, want)
	}
	var fragment struct {
		SystemPrompt *struct {
			Path string `json:"path"`
		} `json:"system_prompt"`
	}
	if err := json.Unmarshal(result.Document, &fragment); err != nil {
		t.Fatal(err)
	}
	if fragment.SystemPrompt == nil || fragment.SystemPrompt.Path == "" {
		t.Fatalf("fragment lacks system_prompt for a non-empty admitted set: %s", result.Document)
	}
}

// TestDropAllDroppedOmitsFragmentSection repairs a profile whose only
// system module is transitive: nothing is written and the fragment
// carries no system_prompt section.
func TestDropAllDroppedOmitsFragmentSection(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	operand := admissionRepos(t, ids, false, false)
	if _, _, _, err := Install(home, InstallOptions{Operand: operand, Policy: admissionPolicy()}); err != nil {
		t.Fatal(err)
	}
	result := admissionResolve(t, home, "sysroot", "claude_code", admissionPolicy(), admissionNative(t))
	if _, err := os.Stat(filepath.Join(ManagedHomeDir(home, "sysroot", "claude_code"), ".agent-context", "system-prompt.md")); !os.IsNotExist(err) {
		t.Fatalf("all-dropped materialization wrote a system-prompt file: %v", err)
	}
	var fragment struct {
		SystemPrompt *struct{} `json:"system_prompt"`
	}
	if err := json.Unmarshal(result.Document, &fragment); err != nil {
		t.Fatal(err)
	}
	if fragment.SystemPrompt != nil {
		t.Fatalf("fragment carries system_prompt with an empty admitted set: %s", result.Document)
	}
}

// TestStatusReportsPolicyAndDropped proves the §12 posture: the profile
// row reports the effective policy value with every dropped module by
// package and path, and the drop warning never makes a home row
// non-current.
func TestStatusReportsPolicyAndDropped(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	operand := admissionRepos(t, ids, true, true)
	if _, _, _, err := Install(home, InstallOptions{Operand: operand, Policy: admissionPolicy()}); err != nil {
		t.Fatal(err)
	}
	nativeHomeOf := admissionNative(t)
	admissionResolve(t, home, "sysroot", "claude_code", admissionPolicy(), nativeHomeOf)
	status, err := StatusOf(StatusRequest{
		Home:         home,
		Machine:      envregistry.DefaultMachineConfig(),
		Detect:       func(envregistry.Adapter) string { return "unknown" },
		NativeHomeOf: nativeHomeOf,
		OperatorXDG:  t.TempDir(),
		LaunchDir:    t.TempDir(),
		Policy:       admissionPolicy(),
	})
	if err != nil {
		t.Fatal(err)
	}
	var profile *ProfileState
	for i := range status.Profiles {
		if status.Profiles[i].Profile == "sysroot" {
			profile = &status.Profiles[i]
		}
	}
	if profile == nil {
		t.Fatalf("no profile row for sysroot")
	}
	if profile.TransitiveSystemModules != contextmaterialize.TransitiveDrop {
		t.Fatalf("policy row %q, want drop", profile.TransitiveSystemModules)
	}
	if len(profile.DroppedSystemModules) != 1 ||
		profile.DroppedSystemModules[0] != (DroppedSystemModule{Package: "sysleaf", Path: "90-system.md"}) {
		t.Fatalf("dropped %+v, want the transitive module by package and path", profile.DroppedSystemModules)
	}
	for _, state := range status.Homes {
		if state.Profile != "sysroot" || state.Environment != "claude_code" {
			continue
		}
		if !state.Current {
			t.Fatalf("drop warnings made the row non-current: %+v", state.Findings)
		}
		warned := false
		for _, warning := range state.Warnings {
			if strings.Contains(warning, contextmaterialize.DiagSystemModuleDropped) {
				warned = true
			}
		}
		if !warned {
			t.Fatalf("home warnings %+v lack the drop warning", state.Warnings)
		}
	}
}

// TestStatusErrorReportsPolicyWithoutDropped proves the error posture: the
// profile row reports error with an empty dropped list, since nothing is
// skipped under error.
func TestStatusErrorReportsPolicyWithoutDropped(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	ids := newGitIdentities(t)
	operand := admissionRepos(t, ids, true, true)
	if _, _, _, err := Install(home, InstallOptions{Operand: operand, Policy: admissionPolicy()}); err != nil {
		t.Fatal(err)
	}
	policy := admissionPolicy()
	policy.TransitiveSystemModules = contextmaterialize.TransitiveError
	native := t.TempDir()
	status, err := StatusOf(StatusRequest{
		Home:         home,
		Machine:      envregistry.DefaultMachineConfig(),
		Detect:       func(envregistry.Adapter) string { return "unknown" },
		NativeHomeOf: func(id string) (string, error) { return filepath.Join(native, id), nil },
		OperatorXDG:  t.TempDir(),
		LaunchDir:    t.TempDir(),
		Policy:       policy,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, profile := range status.Profiles {
		if profile.Profile != "sysroot" {
			continue
		}
		if profile.TransitiveSystemModules != contextmaterialize.TransitiveError {
			t.Fatalf("policy row %q, want error", profile.TransitiveSystemModules)
		}
		if len(profile.DroppedSystemModules) != 0 {
			t.Fatalf("dropped %+v under error, want empty", profile.DroppedSystemModules)
		}
	}
}

// TestPolicyFromConfigCarriesAdmission proves the machine knobs reach the
// profile policy: the error value and the waiver list.
func TestPolicyFromConfigCarriesAdmission(t *testing.T) {
	cfg := &config.Config{Env: config.Environments{
		TransitiveSystemModules: contextmaterialize.TransitiveError,
		SystemModuleWaivers:     []config.SystemModuleWaiver{{Package: "sysleaf", Reason: "reviewed"}},
	}}
	policy := PolicyFromConfig(cfg)
	if policy.TransitiveSystemModules != contextmaterialize.TransitiveError {
		t.Fatalf("policy %q", policy.TransitiveSystemModules)
	}
	if len(policy.SystemModuleWaivers) != 1 || policy.SystemModuleWaivers[0].Package != "sysleaf" {
		t.Fatalf("waivers %+v", policy.SystemModuleWaivers)
	}
	admission := policy.Admission()
	if effective, err := admission.Policy(); err != nil || effective != contextmaterialize.TransitiveError {
		t.Fatalf("admission %q %v", effective, err)
	}
	if !admission.Waived("sysleaf") || admission.Waived("sysmid") {
		t.Fatalf("waived set %+v", admission.Waivers)
	}
	if effective, err := (Policy{}).Admission().Policy(); err != nil || effective != contextmaterialize.TransitiveDrop {
		t.Fatalf("zero policy admission %q %v", effective, err)
	}
}
