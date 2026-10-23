package envprofile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Bounded plants must exit 1 at the intended behavioral assertion. Compilation,
// setup failures and skipped tests do not prove that the gate detects a defect.
func TestCodexSeedCNarrowingControls(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range []struct{ name, file, before, after, pkg, run, fails, message string }{
		{"one-foreign-trust", "internal/envprofile/managed.go", `delete(document, "projects")`, `if projects, ok := document["projects"].(map[string]any); !ok || projects["/foreign-host/synthetic-project"] == nil { delete(document, "projects") }`, "./internal/envprofile", "^TestCodexSeedCSpecCases$/^c-strips-foreign$", "TestCodexSeedCSpecCases/c-strips-foreign", "seed decoded values differ"},
		{"one-nested-array", "internal/envprofile/managed.go", `delete(document, "projects")`, `delete(document, "projects"); if sizes,ok:=document["window_sizes"].([]any);ok&&len(sizes)>0{sizes[0]=int64(81)}`, "./internal/envprofile", "^TestCodexSeedCSpecCases$/^c-preserves-nested-tables-and-arrays$", "TestCodexSeedCSpecCases/c-preserves-nested-tables-and-arrays", "seed decoded values differ"},
		{"unfiltered-comments", "internal/envprofile/managed.go", `if !hasMCP && !hasProjects {`, `if !hasMCP && !hasProjects && !bytes.HasPrefix(payload, []byte("# exact formatting")) {`, "./internal/envprofile", "^TestCodexSeedCPathsAndDeterminism$/^0$", "TestCodexSeedCPathsAndDeterminism/0", "unfiltered bytes changed"},
		{"legacy-B-record-refresh", "internal/envprofile/managed.go", `record.NativeMCPServers = append([]string{}, prior.CodexSeedRecord.NativeMCPServers...)`, `record.NativeMCPServers = append([]string{}, prior.CodexSeedRecord.NativeMCPServers...); if record.Revision == "B" { record.NativeMCPServers = []string{"changed"} }`, "./internal/envprofile", "^TestCodexSeedCLegacyBRepair$", "TestCodexSeedCLegacyBRepair", "legacy B seed or record changed"},
		{"full-carrier-bypass", "internal/envmarker/schema_v7.go", `func validateMarker7(raw []byte) error {`, `func validateMarker7(raw []byte) error { return nil;`, "./internal/envmarker", "^TestMarker7FullCarrier$/^invalid-ref$", "TestMarker7FullCarrier/invalid-ref", "invalid carrier accepted"},
		{"legacy-null-source", "internal/envmarker/envmarker.go", `if _, present := members[name]; present {`, `if _, present := members[name]; present && name != "credential_source" {`, "./internal/envmarker", "^TestMarker7FullCarrier$", "TestMarker7FullCarrier", "invalid carrier accepted"},
		{"C-under-hash1", "internal/envprofile/managed.go", `if prior == nil && adapter.CodexSeedRevision == envregistry.CodexSeedRevisionC && lock.ContentHashVersion() != hashing.VersionV2 {`, `if false && prior == nil && adapter.CodexSeedRevision == envregistry.CodexSeedRevisionC && lock.ContentHashVersion() != hashing.VersionV2 {`, "./internal/envprofile", "^TestCodexSeedCRequiresHash2BeforePublication$", "TestCodexSeedCRequiresHash2BeforePublication", "C hash-v1 writer did not refuse before publication"},
		{"late-seed-refusal", "internal/envprofile/managed.go", `if adapter.CodexSeedRevision == envregistry.CodexSeedRevisionC {`, `if false && adapter.CodexSeedRevision == envregistry.CodexSeedRevisionC {`, "./internal/envprofile", "^TestCodexSeedCSpecCases$/^c-invalid-toml$", "TestCodexSeedCSpecCases/c-invalid-toml", "invalid TOML changed profile lock before refusal"},
		{"node-authority-downgrade", "internal/envprofile/managed.go", `if marker != nil && marker.CredentialSource != nil && marker.CredentialSource.Source != nil {`, `if false && marker != nil && marker.CredentialSource != nil && marker.CredentialSource.Source != nil {`, "./internal/envprofile", "^TestCodexSeedCNodeAuthorityIsNotPerHome$", "TestCodexSeedCNodeAuthorityIsNotPerHome", "node activation did not refuse authority"},
	} {
		t.Run(m.name, func(t *testing.T) {
			original := filepath.Join(root, filepath.FromSlash(m.file))
			payload, err := os.ReadFile(original)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(payload), "\r\n", "\n")
			if strings.Count(source, m.before) != 1 {
				t.Fatal("plant anchor is not unique")
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, "replacement.go")
			if err := os.WriteFile(replacement, []byte(strings.Replace(source, m.before, m.after, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
			overlayPath := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
				t.Fatal(err)
			}
			binary := os.Getenv("CURATOR_GO")
			if binary == "" {
				binary = "go"
			}
			cmd := exec.Command(binary, "test", "-overlay", overlayPath, "-count=1", "-timeout=2m", "-v", "-run", m.run, m.pkg)
			cmd.Dir = root
			output, err := cmd.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+m.fails+" ") || !strings.Contains(string(output), m.message) {
				t.Fatalf("plant survived or failed for an unintended reason: %v\n%s", err, output)
			}
			t.Logf("expected-red exit 1: %s\n%s", m.fails, output)
		})
	}
}
