package ciworkflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func isArtifactUploadAction(uses string) bool {
	const prefix = "actions/upload-artifact@"
	return strings.HasPrefix(uses, prefix) && len(strings.TrimPrefix(uses, prefix)) > 0
}

func workflowHasArtifactUploads(data []byte) (bool, error) {
	var wf workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return false, err
	}
	for _, job := range wf.Jobs {
		for _, current := range job.Steps {
			if isArtifactUploadAction(current.Uses) {
				return true, nil
			}
		}
	}
	return false, nil
}

func checkSanitizedArtifactUploads(data []byte) error {
	var wf workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return err
	}
	uploads := 0
	seenPaths := map[string]bool{}
	for jobName, job := range wf.Jobs {
		for i, current := range job.Steps {
			if !isArtifactUploadAction(current.Uses) {
				continue
			}
			uploads++
			path := current.With["path"]
			if !strings.HasPrefix(path, ".temp/public-artifacts-") || !strings.HasSuffix(path, "/") ||
				!strings.Contains(path, "${{ github.run_id }}-${{ github.run_attempt }}-") {
				return fmt.Errorf("%s: artifact upload uses an unsanitized path", jobName)
			}
			if seenPaths[path] {
				return fmt.Errorf("%s: sanitized upload destination is reused", jobName)
			}
			seenPaths[path] = true
			found := false
			for _, prior := range job.Steps[:i] {
				if prior.If == "always()" && prior.Shell == "bash" &&
					strings.Contains(prior.Run, "go run ./tools/sanitize-ci-evidence -source .temp/") &&
					strings.Contains(prior.Run, "-destination .temp/public-artifacts-$GITHUB_RUN_ID-$GITHUB_RUN_ATTEMPT-") {
					_, destination, ok := strings.Cut(prior.Run, "-destination ")
					want := strings.TrimSuffix(path, "/")
					want = strings.ReplaceAll(want, "${{ github.run_id }}", "$GITHUB_RUN_ID")
					want = strings.ReplaceAll(want, "${{ github.run_attempt }}", "$GITHUB_RUN_ATTEMPT")
					if ok && destination == want {
						found = true
					}
				}
			}
			if !found {
				return fmt.Errorf("%s: upload has no preceding sanitizer step", jobName)
			}
		}
	}
	if uploads == 0 {
		return fmt.Errorf("workflow contains no artifact uploads")
	}
	return nil
}

func TestCIArtifactUploadsUseSanitizedCopies(t *testing.T) {
	good := workflowBytes(t)
	if err := checkSanitizedArtifactUploads(good); err != nil {
		t.Fatal(err)
	}
	for name, pair := range map[string][2]string{
		"raw-upload-path":   {"path: .temp/public-artifacts-", "path: .temp/ci-evidence-"},
		"missing-sanitizer": {"go run ./tools/sanitize-ci-evidence", "echo skipped-sanitizer"},
		"wrong-destination": {"-destination .temp/public-artifacts-", "-destination .temp/ci-evidence-"},
	} {
		t.Run(name, func(t *testing.T) {
			planted := strings.Replace(string(good), pair[0], pair[1], 1)
			if planted == string(good) {
				t.Fatal("mutation did not apply")
			}
			if err := checkSanitizedArtifactUploads([]byte(planted)); err == nil {
				t.Fatal("unsanitized upload survived the control")
			}
		})
	}
	t.Run("other-upload-reference-raw-path", func(t *testing.T) {
		planted := strings.Replace(string(good), "actions/upload-artifact@v4", "actions/upload-artifact@v4.2.0", 1)
		planted = strings.Replace(planted, "path: .temp/public-artifacts-", "path: .temp/ci-evidence-", 1)
		if planted == string(good) {
			t.Fatal("mutation did not apply")
		}
		if err := checkSanitizedArtifactUploads([]byte(planted)); err == nil {
			t.Fatal("unsafe upload at another artifact-action ref survived the control")
		}
	})
}

func TestArtifactUploadGuardAcceptsVersionAndCommitRefs(t *testing.T) {
	good := workflowBytes(t)
	for _, ref := range []string{"v4.2.0", "0123456789abcdef0123456789abcdef01234567"} {
		t.Run(ref, func(t *testing.T) {
			variant := strings.Replace(string(good), "actions/upload-artifact@v4", "actions/upload-artifact@"+ref, 1)
			if variant == string(good) {
				t.Fatal("artifact action reference mutation did not apply")
			}
			if err := checkSanitizedArtifactUploads([]byte(variant)); err != nil {
				t.Fatalf("safe artifact upload at ref %q was rejected: %v", ref, err)
			}
		})
	}
}

func TestAllWorkflowArtifactUploadsUseSanitizedCopies(t *testing.T) {
	workflowDir := "../../.github/workflows"
	workflows, err := os.ReadDir(workflowDir)
	if err != nil {
		t.Fatal("read workflow directory:", err)
	}
	checked := 0
	for _, workflow := range workflows {
		if workflow.IsDir() {
			continue
		}
		extension := strings.ToLower(filepath.Ext(workflow.Name()))
		if extension != ".yml" && extension != ".yaml" {
			continue
		}
		path := filepath.Join(workflowDir, workflow.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read workflow %s: %v", path, err)
		}
		hasUploads, err := workflowHasArtifactUploads(data)
		if err != nil {
			t.Fatalf("parse workflow %s: %v", workflow.Name(), err)
		}
		if !hasUploads {
			continue
		}
		checked++
		if err := checkSanitizedArtifactUploads(data); err != nil {
			t.Errorf("%s: %v", path, err)
		}
	}
	if checked == 0 {
		t.Fatal("no workflow artifact uploads were checked")
	}
}

func TestPostLandingArtifactUploadUsesSanitizedCopy(t *testing.T) {
	data, err := os.ReadFile("../../.github/workflows/post-landing-mutation-tests.yml")
	if err != nil {
		t.Fatal("read workflow:", err)
	}
	if err := checkSanitizedArtifactUploads(data); err != nil {
		t.Fatal(err)
	}
}
