package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Narrow one forbidden class at a time while retaining the surrounding gate.
// Only a named behavioral failure with go test exit 1 counts as a kill.
// The renderer plant drops only the mandatory backend; it is producer evidence,
// not the separately required real hosted-consumer override plant.
func TestMuseCredentialPolicyNarrowingMutants(t *testing.T) {
	mutants := []struct{ name, file, before, after, pkg, test string }{
		{"fragment-reobserves-muse-release", "internal/envprofile/managed.go", "policy := plan.credentialPolicy", "policy := plan.credentialPolicy\nif adapter.ID == envregistry.Muse { var err error; policy, err = req.resolveCredentialPolicy(adapter); if err != nil { return nil, err } }", "./internal/envprofile", "TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared"},
		{"muse-keychain-admitted", "internal/envprofile/credential_policy.go", "selection == envregistry.CredentialStoreNative || adapter.ID == envregistry.Muse && selection == envregistry.CredentialStoreFile", "selection == envregistry.CredentialStoreNative || adapter.ID == envregistry.Muse && (selection == envregistry.CredentialStoreFile || selection == envregistry.CredentialStoreKeychain)", "./internal/envprofile", "TestMuseCredentialPolicyMatrix/keychain/shared"},
		{"missing-shared-source-admitted", "internal/envprofile/muse.go", "return conflict(\"native credential target is absent; run TBH_CREDENTIAL_BACKEND=file muse login in the native XDG environment before resolving shared Muse\")", "return museAuthState{live: true}", "./internal/envprofile", "TestMuseMissingNativeFilePolicy/shared"},
		{"isolated-muse-native-link", "internal/envprofile/credential_policy.go", "if p.isolation == envregistry.IsolationIsolated {", "if p.isolation == envregistry.IsolationIsolated && adapter.ID != envregistry.Muse {", "./internal/envprofile", "TestMuseCredentialPolicyMatrix/file/isolated"},
		{"renderer-drops-only-backend", "internal/envfragment/envfragment.go", "for name := range f.Env {", "for name := range f.Env {\nif name == \"TBH_CREDENTIAL_BACKEND\" { continue }", "./cmd/curator", "TestMuseCLIStorePolicy"},
		{"header-only-root-fact-dropped", "internal/envprofile/managed.go", "fragment.RootContext = &envfragment.RootContext{Exclusive: true}", "fragment.RootContext = &envfragment.RootContext{Exclusive: true}\nif raw, e := os.ReadFile(filepath.Join(plan.homeDir,\"config/muse/AGENTS.md\")); e == nil && !strings.Contains(string(raw), \"## Context:\") { fragment.RootContext = nil }", "./cmd/curator", "TestMuseRootContextVectors/zero-applicable"},
		{"boundary-admits-HOME", "internal/envfragment/envfragment.go", "if !allowed[name] {", "if !allowed[name] && !(adapter.ID == envregistry.Muse && name == \"HOME\") {", "./internal/envfragment", "TestMuseFragmentBoundary/HOME"},
		{"boundary-admits-missing-literal", "internal/envfragment/envfragment.go", "fragment.Env[\"TBH_CREDENTIAL_BACKEND\"] != \"file\"", "(fragment.Env[\"TBH_CREDENTIAL_BACKEND\"] != \"file\" && fragment.Env[\"TBH_CREDENTIAL_BACKEND\"] != \"\")", "./internal/envfragment", "TestMuseFragmentBoundary/missing-literal"},
		{"boundary-admits-false-root", "internal/envfragment/envfragment.go", "adapter.ID != envregistry.Muse || !fragment.RootContext.Exclusive", "adapter.ID != envregistry.Muse", "./internal/envfragment", "TestMuseFragmentBoundary/false-root"},
		{"unrelated-root-mints-fact", "internal/envprofile/managed.go", "if _, ok := plan.marker.Surfaces[envmarker.SurfaceRootContext]; ok {", "_, ok := plan.marker.Surfaces[envmarker.SurfaceRootContext]; if !ok { _, e := os.Stat(filepath.Join(req.Home, \"AGENTS.md\")); ok = e == nil }; if ok {", "./cmd/curator", "TestMuseRootContextVectors/no-context"},
		{"linked-root-fact-dropped", "internal/envprofile/managed.go", "fragment.RootContext = &envfragment.RootContext{Exclusive: true}", "fragment.RootContext = &envfragment.RootContext{Exclusive: true}\nif info, e := os.Lstat(filepath.Join(plan.homeDir, \"config/muse/AGENTS.md\")); e == nil && info.Mode()&os.ModeSymlink != 0 { fragment.RootContext = nil }", "./cmd/curator", "TestMuseRootContextVectors/monolithic"},
		{"denied-root-admitted", "internal/envprofile/managed.go", "state, err := stateread.ReadFile(target) // #nosec G304 -- link target recomputed from the verified plan", "state, err := stateread.ReadFile(target) // #nosec G304 -- link target recomputed from the verified plan\nif plan.adapter.ID == envregistry.Muse && path == \"config/muse/AGENTS.md\" && errors.Is(err, os.ErrPermission) { continue }", "./cmd/curator", "TestMuseRootContextRefusals/denied"},
		{"shared-status-misreports-isolation", "internal/envprofile/status.go", "policy.effectiveStore, policy.isolation)", "policy.effectiveStore, map[string]string{\"shared\": \"isolated\", \"isolated\": \"isolated\"}[policy.isolation])", "./cmd/curator", "TestMuseCredentialPolicyStatus/shared"},
		{"readlink-denial-admitted", "internal/envprofile/muse.go", "link, err = stateread.ReadlinkWith(full, req.passthroughReadlink)\n\t\t\t}", "link, err = stateread.ReadlinkWith(full, req.passthroughReadlink)\n\t\t\t}\nif err != nil && link.Kind == stateread.KindUnreadable { return museAuthState{live: true} }", "./internal/envprofile", "TestMuseAuthRepairAdditionalStates/readlink-error"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			if runtime.GOOS == "windows" && mutant.name == "fragment-reobserves-muse-release" {
				t.Skip("shell sentinel is Unix-only")
			}
			if runtime.GOOS == "windows" && mutant.name == "denied-root-admitted" {
				t.Skip("POSIX mode-bit unreadability: chmod cannot establish Windows ACL denial")
			}
			output, err := runMuseProducerOverlay(t, mutant.file, mutant.before, mutant.after, mutant.pkg, mutant.test)
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+mutant.test) {
				t.Fatalf("mutant survived or lacked named behavioral failure %s: %v\n%s", mutant.test, err, output)
			}
			t.Logf("%s: expected-red go test exit %d; killed by %s", mutant.name, exit.ExitCode(), mutant.test)
		})
	}
}

// Run a production entry point with one bounded source overlay. No candidate
// files are rewritten; the CLI subprocess retains its normal git isolation.
func runMuseProducerOverlay(t *testing.T, file, before, after, pkg, test string) ([]byte, error) {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(root, filepath.FromSlash(file))
	payload, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	// Windows checkouts may use CRLF; overlays change Go behavior,
	// independently of the checkout's line-ending convention.
	source := strings.ReplaceAll(string(payload), "\r\n", "\n")
	if strings.Count(source, before) != 1 {
		t.Fatal("mutation anchor is not unique")
	}
	dir := t.TempDir()
	replacement := filepath.Join(dir, "replacement.go")
	if err := os.WriteFile(replacement, []byte(strings.Replace(source, before, after, 1)), 0600); err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(test, "/")
	for i := range parts {
		parts[i] = "^" + parts[i] + "$"
	}
	command := exec.Command("go", "test", "-overlay", overlayPath, "-count=1", "-timeout=5m", "-v", "-run", strings.Join(parts, "/"), pkg)
	command.Dir = root
	// Reuse the CLI helper protocol: the parent already holds the
	// package GOROOT lock, and these behavioral tests do not mutate it.
	// TestMain still isolates git in the child process.
	command.Env = append(os.Environ(), cliHelperProcess+"=1")
	return command.CombinedOutput()
}

// Force only Muse's root symlink to be unavailable. Resolve must take the real
// copyLinkFallback path, record a copied surface, then emit the verified fact
// on both repair and already-current resolve. This is a positive control.
func TestMuseRootContextCopyFallback(t *testing.T) {
	output, err := runMuseProducerOverlay(t, "internal/envprofile/managed.go",
		"if err := atomicManagedLink(root, rel, target); err != nil {\n\t\t\tvar unavailable *errSymlinkUnavailable",
		`var linkErr error
        if path == "config/muse/AGENTS.md" { linkErr = &errSymlinkUnavailable{err: os.ErrPermission} } else { linkErr = atomicManagedLink(root, rel, target) }
        if err := linkErr; err != nil {
            var unavailable *errSymlinkUnavailable`,
		"./cmd/curator", "TestMuseRootContextVectors/monolithic")
	if err != nil || !strings.Contains(string(output), "verified copied root") {
		t.Fatalf("copy fallback did not qualify: %v\n%s", err, output)
	}
	t.Log("copy-fallback production control: go test exit 0; verified copied root")
}
