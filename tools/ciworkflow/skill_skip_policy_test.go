package ciworkflow

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The shipped gate admits exactly the three guarded skill-fixture capability
// reasons. Unknown and decorated reasons fail at classification, rather than
// silently shrinking the suite. This isolates class policy from the case ledger.
func TestSkillCapabilitySkipPolicy(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("no bash")
	}
	ci, err := filepath.Abs("../../.github/ci")
	if err != nil {
		t.Fatal(err)
	}
	reasons := []string{
		"filesystem cannot represent the colliding pair",
		"filesystem cannot materialize the executable-bit mutation",
		"fixture requires Unix permission denial without root bypass",
	}
	for _, goos := range []string{"darwin", "linux", "windows"} {
		for _, reason := range append(reasons, "unclassified fixture omission") {
			for _, suffix := range []string{"", " extra"} {
				t.Run(goos+"/"+reason+suffix, func(t *testing.T) {
					dir := t.TempDir()
					stream := filepath.Join(dir, "test.json")
					ledger := filepath.Join(dir, "cases.tsv")
					var data []byte
					for _, event := range []map[string]string{
						{"Action": "output", "Package": "example.test/fixture/skill", "Test": "TestFixture", "Output": "    fixture_test.go:12: " + reason + suffix + "\n"},
						{"Action": "skip", "Package": "example.test/fixture/skill", "Test": "TestFixture"},
					} {
						line, err := json.Marshal(event)
						if err != nil {
							t.Fatal(err)
						}
						data = append(append(data, line...), '\n')
					}
					if err := os.WriteFile(stream, data, 0o600); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(ledger, []byte("# Isolated class-policy witness.\n"), 0o600); err != nil {
						t.Fatal(err)
					}
					cmd := exec.Command(bash, filepath.ToSlash(filepath.Join(ci, "platform-case-gate.sh")), filepath.ToSlash(stream), filepath.ToSlash(dir))
					for _, entry := range os.Environ() {
						key, _, _ := strings.Cut(entry, "=")
						if strings.HasPrefix(key, "CI_GATE_") || key == "CI_SKIP_CLASSES" || key == "CI_PLATFORM_CASES" || key == "CI_DEFERRED_PKGS" || key == "CI_EXCLUDED_PKGS" {
							continue
						}
						cmd.Env = append(cmd.Env, entry)
					}
					cmd.Env = append(cmd.Env, "CI_GATE_GOOS="+goos, "CI_GATE_MODULE=example.test/fixture", "CI_SKIP_CLASSES="+filepath.ToSlash(filepath.Join(ci, "skip-classes.tsv")), "CI_PLATFORM_CASES="+filepath.ToSlash(ledger))
					out, err := cmd.CombinedOutput()
					allowed := suffix == "" && reason != "unclassified fixture omission"
					if allowed {
						if err != nil {
							t.Fatalf("legitimate capability refusal: %v\n%s", err, out)
						}
						return
					}
					if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 || !strings.Contains(string(out), "skip with an unrecognised reason") {
						t.Fatalf("missing exact policy refusal: %v\n%s", err, out)
					}
				})
			}
		}
	}
}
