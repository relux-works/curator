package envprofile

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Overlay plants must fail the named Resolve behavioral test with exit 1 and the intended diagnostic; setup/build failures do not count as kills.
func TestClaudeRefusalNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(root, "internal", "envprofile", "managed.go")
	payload, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(string(payload), "\r\n", "\n")
	for _, mutant := range []struct{ name, before, after, run, fails, message string }{
		{
			"trailing-empty-object",
			"if err := decoder.Decode(&trailing); err != io.EOF {",
			"if err := decoder.Decode(&trailing); err != io.EOF && string(payload) != \x60{} {}\x60 {",
			"^TestClaudePreserveLockRefusalIsZeroWrite$/^referenced$/^(trailing-object|trailing-garbage|valid-whitespace)$",
			"TestClaudePreserveLockRefusalIsZeroWrite/referenced/trailing-object", "want environment_seed_unreadable refusal",
		},
		{
			"premature-write-malformed-referenced",
			"if !ok {\n\t\t\treturn nil, fmt.Errorf(\"%s: managed .claude.json is not an object\", envregistry.DiagSeedUnreadable)",
			"if !ok {\nif string(payload) == \x60{\"broken\":\x60 && form == envregistry.FormReferenced { if err := atomicManagedFile(root, managedJoin(homeRel, \"CLAUDE.md\"), []byte(\"premature write\"), managedAuthoredFileMode); err != nil { return nil, err } }\n\t\t\treturn nil, fmt.Errorf(\"%s: managed .claude.json is not an object\", envregistry.DiagSeedUnreadable)",
			"^TestClaudePreserveLockRefusalIsZeroWrite$/^referenced$/^malformed$",
			"TestClaudePreserveLockRefusalIsZeroWrite/referenced/malformed", "refusal changed managed entry CLAUDE.md",
		},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			if strings.Count(source, mutant.before) != 1 {
				t.Fatal("mutation anchor is not unique")
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, "replacement.go")
			if err := os.WriteFile(replacement, []byte(strings.Replace(source, mutant.before, mutant.after, 1)), 0o600); err != nil {
				t.Fatal(err)
			}
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlayPath := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0o600); err != nil {
				t.Fatal(err)
			}
			command := exec.Command("go", "test", "-overlay", overlayPath, "-count=1", "-timeout=2m", "-v", "-run", mutant.run, "./internal/envprofile")
			command.Dir = root
			var outputBuffer bytes.Buffer
			command.Stdout = &outputBuffer
			command.Stderr = &outputBuffer
			if err := command.Start(); err != nil {
				t.Fatalf("cannot-start: %v", err)
			}
			t.Logf("owned child PID=%d", command.Process.Pid)
			err = command.Wait()
			output := outputBuffer.Bytes()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+mutant.fails+" ") || !strings.Contains(string(output), mutant.message) {
				t.Fatalf("mutant survived or lacked the intended behavioral failure: %v\n%s", err, output)
			}
			t.Logf("expected-red go test exit %d:\n%s", exit.ExitCode(), output)
		})
	}
}
