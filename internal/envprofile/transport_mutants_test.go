package envprofile

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Scoped overlays execute the public-entry behavioral suite. Each mutant
// weakens one member of the contract and must fail a named property test;
// merely failing compilation, timing out or crashing does not kill a mutant.
func TestDeclaredTransportNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ name, file, old, new, tests, killer, sibling string }{
		{"canonicalization-prefixed-ssh", "envprofile/gitsource.go", "Source: requirement.Git", "Source: identity.Canonical(requirement.Git)", "TestInstallDeclaredDependencyTransport", "TestInstallDeclaredDependencyTransport/ssh/stale-false", "TestInstallDeclaredDependencyTransport/https/stale-false"},
		{"missing-pin-ssh-to-https", "envprofile/gitsource.go", "gitops.FetchDeclaredCommit(dir, m.endpoint(source), commit)", "gitops.FetchDeclaredCommit(dir, strings.Replace(m.endpoint(source), \"ssh://git@\", \"https://\", 1), commit)", "TestInstallDeclaredDependencyTransport", "TestInstallDeclaredDependencyTransport/ssh/stale-true", "TestInstallDeclaredDependencyTransport/https/stale-true"},
		{"allow-one-denied-source", "envprofile/gitsource.go", "if identity.Allowed(trimmed, m.allowedSources) {", "if identity.Allowed(trimmed, m.allowedSources) || trimmed == \"fixture.invalid/Team/dep\" {", "(TestInstallDeclaredSourceAdmission|TestInstallDeclaredDependencyTransport)", "TestInstallDeclaredSourceAdmission/denied", "TestInstallDeclaredDependencyTransport/https/stale-false"},
		{"inherit-user-gitconfig", "gitops/gitops.go", `"GIT_CONFIG_GLOBAL="+emptyGitName`, `"GIT_CONFIG_GLOBAL="+os.Getenv("GIT_CONFIG_GLOBAL")`, "gitops:(TestDeclaredFetchIgnoreInheritedRemapping|TestDeclaredCloneFetchIgnoreInheritedRemapping|TestCloneAndResolve)", "TestDeclaredFetchIgnoreInheritedRemapping", "TestCloneAndResolve"},
		{"allow-terminal-prompt", "gitops/gitops.go", "GIT_TERMINAL_PROMPT=0", "GIT_TERMINAL_PROMPT=1", "(TestInstallDeclaredDependencyTransport|TestSharedGitDeclaredSSH)", "TestInstallDeclaredDependencyTransport/https/stale-false", "TestSharedGitDeclaredSSH"},
		{"reinstall-discards-current-endpoint", "envprofile/envprofile.go", "updateLocked(op, manager, home, name, options.Policy", "updateLocked(op, newGitManager(home).withPolicy(options.Policy), home, name, options.Policy", "(TestInstallReinstallDeclaredRootEndpoint|TestUpdateRetainsDeclaredCachedEndpoint)", "TestInstallReinstallDeclaredRootEndpoint", "TestUpdateRetainsDeclaredCachedEndpoint"},
		{"unreachable-ssh-falls-back", "gitops/acquisition.go", `return acquisitionError(endpoint, FetchCommitFromURLIsolated(repo, endpoint, commit))`, `err := acquisitionError(endpoint, FetchCommitFromURLIsolated(repo, endpoint, commit)); if failure, ok := err.(*AcquisitionError); ok && failure.Class == "transport_failed" && strings.HasPrefix(endpoint, "ssh://git@") { _ = FetchCommitFromURLIsolated(repo, strings.Replace(endpoint, "ssh://git@", "https://", 1), commit) }; return err`, "(TestInstallDeclaredEndpointRefusal|TestInstallDeclaredDependencyTransport)", "TestInstallDeclaredEndpointRefusal/unreachable/stale-true", "TestInstallDeclaredDependencyTransport/https/stale-false"},
		{"admit-ext-only", "gitops/gitops.go", `"GIT_ALLOW_PROTOCOL="+AllowedProtocols,`, `"GIT_ALLOW_PROTOCOL="+AllowedProtocols+":ext",`, "(TestReviewMissingPinPreservesProtocolAllowlist|TestSharedGitDeclaredSSH)", "TestReviewMissingPinPreservesProtocolAllowlist/serves", "TestSharedGitDeclaredSSH"},
		{"skip-one-update-refresh", "envprofile/gitsource.go", "if m.refreshResults != nil {\n\t\tif err := m.fetch(source);", "if m.refreshResults != nil && source != \"fixture.invalid/Team/dep\" {\n\t\tif err := m.fetch(source);", "(TestReinstallDeclaredPinnedDependencyEndpoint|TestUpdateRetainsDeclaredCachedEndpoint)", "TestReinstallDeclaredPinnedDependencyEndpoint", "TestUpdateRetainsDeclaredCachedEndpoint"},
		{"auth-stderr-retained", "gitops/acquisition.go", "class = \"authentication_failed\"", "class = \"authentication_failed\" + err.Error()", "TestInstallDeclaredEndpointRefusal", "TestInstallDeclaredEndpointRefusal/auth/stale-true", "TestInstallDeclaredEndpointRefusal/unreachable/stale-true"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := filepath.Join(root, "internal", filepath.FromSlash(tc.file))
			raw, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			raw = bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))
			if !strings.Contains(string(raw), tc.old) {
				t.Fatal("mutation anchor missing")
			}
			dir := t.TempDir()
			plant := filepath.Join(dir, "source.go")
			if err := os.WriteFile(plant, []byte(strings.Replace(string(raw), tc.old, tc.new, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{source: plant}})
			path := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(path, overlay, 0600); err != nil {
				t.Fatal(err)
			}
			pkg := "./internal/envprofile"
			if strings.HasPrefix(tc.tests, "gitops:") {
				pkg = "./internal/gitops"
				tc.tests = strings.TrimPrefix(tc.tests, "gitops:")
			}
			cmd := exec.Command("go", "test", "-overlay", path, "-count=1", "-timeout=3m", "-json", pkg, "-run", "^"+tc.tests+"$")
			cmd.Dir = root
			var output bytes.Buffer
			cmd.Stdout, cmd.Stderr = &output, &output
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			err = cmd.Wait()
			t.Logf("mutant=%s PID=%d exit=%d", tc.name, cmd.Process.Pid, cmd.ProcessState.ExitCode())
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 {
				t.Fatalf("survivor or invalid execution: %v: %s", err, output.String())
			}
			actions := map[string]string{}
			for _, line := range strings.Split(output.String(), "\n") {
				var event struct{ Action, Test string }
				if json.Unmarshal([]byte(line), &event) == nil && (event.Action == "pass" || event.Action == "fail") {
					actions[event.Test] = event.Action
				}
			}
			if actions[tc.killer] != "fail" || actions[tc.sibling] != "pass" {
				t.Fatalf("no named behavioral kill with green sibling: %v\n%s", actions, output.String())
			}
			t.Logf("killed by %s; valid sibling %s passed", tc.killer, tc.sibling)
		})
	}
}
