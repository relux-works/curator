package ciworkflow

import (
	"fmt"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func checkStoreCompatibilityWiring(data []byte) error {
	var wf struct {
		Jobs map[string]struct {
			Needs any
			Env   map[string]string
			Steps []step
		}
	}
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return err
	}
	producer, ok := wf.Jobs["store-v1-compatibility"]
	if !ok {
		return fmt.Errorf("missing compatibility producer")
	}
	replay, upload := 0, 0
	for _, s := range producer.Steps {
		if strings.Contains(s.Run, "-run '^TestStoreBoundary(Resolve|DryRun|Repair|Status)VectorsDrive'") && strings.Contains(s.Run, "go test -json -count=1") && strings.Contains(s.Run, `test "$replay_rc" = 1`) && strings.Contains(s.Run, `go run ./tools/check-store-compatibility -dir "$evidence" -commit "$GITHUB_SHA"`) && strings.Contains(s.Run, "go test -count=1 ./internal/storecompatibility") && s.Env["CURATOR_CONFORMANCE_ROOT"] == "${{ github.workspace }}/protocol-spec/conformance/v1" {
			replay++
		}
		if s.Uses == "actions/upload-artifact@v4" && s.With["name"] == "store-v1-compatibility" && s.With["if-no-files-found"] == "error" {
			upload++
		}
	}
	if replay != 1 || upload != 1 {
		return fmt.Errorf("compatibility producer must replay and check exact IDs before upload")
	}
	for _, name := range []string{"test", "race", "windows-test-shards", "unix-test-shards", "envprofile-race"} {
		job := wf.Jobs[name]
		need := false
		if list, ok := job.Needs.([]any); ok {
			for _, item := range list {
				if item == "store-v1-compatibility" {
					need = true
				}
			}
		}
		if !need || job.Env["CI_STORE_COMPATIBILITY"] != "${{ github.workspace }}/.temp/store-v1-compatibility" || job.Env["CI_STORE_V2_CANONICAL"] != "1" {
			return fmt.Errorf("%s: compatibility dependency and canonical pin required", name)
		}
		lastCheckout, download := -1, -1
		for i, s := range job.Steps {
			if s.Uses == "actions/checkout@v4" {
				lastCheckout = i
			}
			if s.Uses == "actions/download-artifact@v4" && s.With["name"] == "store-v1-compatibility" && s.With["path"] == ".temp/store-v1-compatibility" && s.If == "" {
				download = i
			}
		}
		if download <= lastCheckout {
			return fmt.Errorf("%s: same-run evidence must be downloaded after checkout", name)
		}
	}
	return nil
}

// Real workflow wiring delegates only after the full replay checker, and every
// canonical inventory/aggregate consumes that same-run artifact. These plants
// prove the wiring detector, not execution on hosted platforms.
func TestStoreCompatibilityWiring(t *testing.T) {
	good := workflowBytes(t)
	if err := checkStoreCompatibilityWiring(good); err != nil {
		t.Fatal(err)
	}
	for name, pair := range map[string][2]string{
		"missing-checker":    {`go run ./tools/check-store-compatibility -dir "$evidence" -commit "$GITHUB_SHA"`, `echo unverified`},
		"wrong-family":       {`-run '^TestStoreBoundary(Resolve|DryRun|Repair|Status)VectorsDrive'`, `-run '^TestUnrelated'`},
		"missing-dependency": {"needs: [store-v1-compatibility]", "needs: [unrelated]"},
		"missing-proof":      {"CI_STORE_COMPATIBILITY: ${{ github.workspace }}/.temp/store-v1-compatibility", "CI_OTHER_COMPATIBILITY: unused"},
		"wrong-artifact":     {"name: store-v1-compatibility\n          path: .temp/store-v1-compatibility", "name: unrelated\n          path: .temp/store-v1-compatibility"},
	} {
		t.Run(name, func(t *testing.T) {
			planted := strings.Replace(string(good), pair[0], pair[1], 1)
			if planted == string(good) {
				t.Fatal("plant did not apply")
			}
			if err := checkStoreCompatibilityWiring([]byte(planted)); err == nil {
				t.Fatal("compatibility wiring violation survived")
			}
		})
	}
}
