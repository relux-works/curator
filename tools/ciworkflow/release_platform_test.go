package ciworkflow

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func rejectReleasePlatformTag(data []byte) error {
	if strings.Contains(string(data), "curator_test_platform") {
		return fmt.Errorf("release path contains the platform test tag")
	}
	return nil
}

// Both real release entrypoints exclude the override tag. Plants in each
// document must fail this gate; default-build tests also prove the selector inert.
func TestReleaseBuildExcludesPlatformTestTag(t *testing.T) {
	for _, path := range []string{"../../.goreleaser.yml", "../../.github/workflows/release.yml"} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := rejectReleasePlatformTag(raw); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		plant := append(append([]byte(nil), raw...), []byte("\nenv: {GOFLAGS: '-tags=curator_test_platform'}\n")...)
		if err := rejectReleasePlatformTag(plant); err == nil {
			t.Fatalf("release-tag plant survived in %s", path)
		}
	}
}

// An absent, renamed or skipped test cannot satisfy a filtered CI witness.
func TestRequiredTestsRejectMissingAndSkipped(t *testing.T) {
	for _, event := range []string{"pass", "skip", "absent", "renamed"} {
		t.Run(event, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "tests.json")
			row := `{"Action":"` + event + `","Test":"Witness"}` + "\n"
			if event == "absent" {
				row = ""
			}
			if event == "renamed" {
				row = `{"Action":"pass","Test":"Different"}` + "\n"
			}
			if err := os.WriteFile(path, []byte(row), 0600); err != nil {
				t.Fatal(err)
			}
			output, err := exec.Command("python3", "../../.github/ci/required-tests.py", path, "Witness").CombinedOutput()
			if event == "pass" {
				if err != nil {
					t.Fatalf("positive required-set: %v %s", err, output)
				}
			} else if err == nil || !strings.Contains(string(output), "required tests did not pass: Witness") {
				t.Fatalf("negative required-set: %v %s", err, output)
			}
		})
	}
}
