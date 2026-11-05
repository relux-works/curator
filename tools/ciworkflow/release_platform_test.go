package ciworkflow

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
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

func checkColdRequiredWitnessSets(data []byte) error {
	var wf workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return err
	}
	job, ok := wf.Jobs["cold-repository"]
	if !ok {
		return fmt.Errorf("cold job missing")
	}
	if len(job.Steps) < 2 || job.Steps[1].Run != isolationCommand {
		return fmt.Errorf("cold private Go initialization missing")
	}
	count := 0
	for _, step := range job.Steps {
		if !strings.Contains(step.Run, "go test ") {
			continue
		}
		count++
		if !strings.Contains(step.Run, "-json") || !strings.Contains(step.Run, "python3 .github/ci/required-tests.py ") {
			return fmt.Errorf("cold filtered tests lost required PASS set")
		}
	}
	if count != 3 {
		return fmt.Errorf("cold witness step count=%d", count)
	}
	return nil
}

// The real filtered cold workflow wires the required-set checker to all three
// test steps. Removing it from any one step is detected independently of Go.
func TestColdWorkflowRequiresNamedWitnesses(t *testing.T) {
	data, err := os.ReadFile("../../.github/workflows/cold-repository.yml")
	if err != nil {
		t.Fatal(err)
	}
	data = []byte(strings.ReplaceAll(string(data), "\r\n", "\n"))
	if err := checkColdRequiredWitnessSets(data); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"measurement-tests.json", "install-tests.json", "deadline-tests.json"} {
		lines := strings.Split(string(data), "\n")
		planted := false
		for i, line := range lines {
			if strings.Contains(line, "python3 .github/ci/required-tests.py "+name+" ") {
				lines[i] = "          true"
				planted = true
			}
		}
		if !planted {
			t.Fatalf("required-set plant not applied: %s", name)
		}
		err := checkColdRequiredWitnessSets([]byte(strings.Join(lines, "\n")))
		if err == nil || !strings.Contains(err.Error(), "required PASS set") {
			t.Fatalf("required-set wiring plant survived: %v", err)
		}
	}
}
