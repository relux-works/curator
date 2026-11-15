package ciworkflow

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Synthetic streams prove each store/skill reason is bound to its exact case,
// platform and measured capability. Shared reasons must admit both named cases;
// unrelated cases, decorated reasons and sensitive-filesystem collisions refuse.
// The probe fixtures test gate policy, not native filesystem qualification.
func TestSkillCapabilitySkipPolicy(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("no bash")
	}
	ci, err := filepath.Abs("../../.github/ci")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ pkg, name, reason, goos, capability string }{
		{"internal/skillintegrity", "TestEvidenceStructuralRefusals/case-collision", "filesystem cannot represent the colliding pair", "linux,darwin,windows", "case-folding"},
		{"internal/skillintegrity", "TestEvidenceRejectsEveryDriftDirection/execute", "filesystem cannot materialize the executable-bit mutation", "windows", "-"},
		{"internal/skillintegrity", "TestUnreadableNativeIsNotAbsence", "fixture requires Unix permission denial without root bypass", "linux,darwin,windows", "-"},
		{"internal/skillintegrity", "TestCaptureContinuesAfterUnavailableRead", "fixture requires Unix permission denial without root bypass", "linux,darwin,windows", "-"},
		{"internal/envprofile", "TestStoreVerificationPinnedCorpusReader", "separate store revision-2 draft qualification is not selected", "linux,darwin,windows", "-"},
		{"internal/envprofile", "TestStoreVerificationActualReaderDirections", "separate store revision-2 draft qualification is not selected", "linux,darwin,windows", "-"},
	}
	for _, row := range cases {
		for _, goos := range []string{"darwin", "linux", "windows"} {
			for _, variant := range []string{"exact", "package", "test", "child", "prefix", "suffix", "sensitive"} {
				t.Run(row.name+"/"+goos+"/"+variant, func(t *testing.T) {
					dir := t.TempDir()
					stream := filepath.Join(dir, "test.json")
					ledger := filepath.Join(dir, "cases.tsv")
					pkg, name, reason := row.pkg, row.name, row.reason
					want := ""
					switch variant {
					case "package":
						pkg = "internal/unrelated"
						want = "FATAL-scope"
					case "test":
						name = "TestUnrelated"
						want = "FATAL-scope"
					case "child":
						name += "/extra"
						want = "FATAL-scope"
					case "prefix":
						reason = "unexpected: " + reason
						want = "FATAL-unclassified"
					case "suffix":
						reason += " extra"
						want = "FATAL-unclassified"
					}
					if variant == "sensitive" && row.capability == "case-folding" {
						want = "FATAL-scope"
					}
					if !strings.Contains(row.goos, goos) && want == "" {
						want = "FATAL-scope"
					}
					var data []byte
					for _, event := range []map[string]string{
						{"Action": "output", "Package": "example.test/fixture/" + pkg, "Test": name, "Output": "    fixture_test.go:12: " + reason + "\n"},
						{"Action": "skip", "Package": "example.test/fixture/" + pkg, "Test": name},
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
					if err := os.WriteFile(ledger, []byte("# Scope policy in isolation.\n"), 0o600); err != nil {
						t.Fatal(err)
					}
					gate, err := os.ReadFile(filepath.Join(ci, "platform-case-gate.sh"))
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(dir, "gate.sh"), gate, 0o600); err != nil {
						t.Fatal(err)
					}
					capability := "folding"
					if variant == "sensitive" {
						capability = "sensitive"
					}
					if err := os.WriteFile(filepath.Join(dir, "case-folding-probe.sh"), []byte("echo "+capability+"\n"), 0o600); err != nil {
						t.Fatal(err)
					}
					cmd := exec.Command(bash, filepath.ToSlash(filepath.Join(dir, "gate.sh")), filepath.ToSlash(stream), filepath.ToSlash(dir))
					for _, entry := range os.Environ() {
						key, _, _ := strings.Cut(entry, "=")
						if strings.HasPrefix(key, "CI_GATE_") || key == "CI_SKIP_CLASSES" || key == "CI_PLATFORM_CASES" || key == "CI_DEFERRED_PKGS" || key == "CI_EXCLUDED_PKGS" {
							continue
						}
						cmd.Env = append(cmd.Env, entry)
					}
					cmd.Env = append(cmd.Env, "CI_GATE_GOOS="+goos, "CI_GATE_MODULE=example.test/fixture", "CI_SKIP_CLASSES="+filepath.ToSlash(filepath.Join(ci, "skip-classes.tsv")), "CI_PLATFORM_CASES="+filepath.ToSlash(ledger))
					out, err := cmd.CombinedOutput()
					if want == "" {
						if err != nil {
							t.Fatalf("legitimate scoped skip: %v\n%s", err, out)
						}
						return
					}
					if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 {
						t.Fatalf("missing scope rejection: %v\n%s", err, out)
					}
					verdict, err := os.ReadFile(filepath.Join(dir, "skips-observed.tsv"))
					if err != nil || !strings.Contains(string(verdict), want) {
						t.Fatalf("wanted %s: %v\n%s", want, err, verdict)
					}
				})
			}
		}
	}
}
