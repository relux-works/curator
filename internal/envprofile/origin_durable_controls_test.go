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

// These overlays change actual production owners. Each child must fail the
// named contract assertion, with no setup/build/panic/timeout substituted for
// detection. Nearby healthy leaves are part of the unmodified owner tests;
// overlays never change source on disk. Historical callback-model plants are
// separately labelled and provide no runtime/restart qualification.
func TestProvisioningOriginDurableNarrowingControls(t *testing.T) {
	isolateDetectorParent(t)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, plant := range []struct {
		name, file, before, after, run, assertion string
		count                                     int
	}{
		{"reject-retained-neighbor", "internal/envauthority/lifecycle.go", "if name != profile {", "if name != profile && entry.Removed == nil {", "^TestProvisioningOriginRemoveRetainOwner$", "disjoint actual Q continuation:", 1},
		{"exempt-origin-orphan", "internal/envauthority/authority.go", "if !lifecycle.retained[target] {", "if !lifecycle.retained[target] && marker != nil {", "^TestProvisioningOriginCommonAdmission$/^orphan-sidecar$", "common owner admitted invalid authority", 1},
		{"drop-retained-L1", "internal/transaction/global.go", "j.Global.State = \"retaining\"", "return engine.rollbackGlobal(j, &RetainedGlobalConflict{})\n j.Global.State = \"retaining\"", "^TestProvisioningOriginGlobalRetainedConflict$", "handled conflict dropped retained L1:", 1},
		{"bless-stable-drift", "internal/envauthority/origin.go", "func validateRecorded(manager, profile, adapter, target string, marker *envmarker.Marker, intent *envunit.Unit) error {", "func validateRecorded(manager, profile, adapter, target string, marker *envmarker.Marker, intent *envunit.Unit) error {\n if marker != nil { return nil }", "^TestProvisioningOriginCommonAdmission$/^lock-drift$", "common owner admitted invalid authority", 1},
		{"skip-migration-scan", "internal/envprofile/migrate.go", "if err := envauthority.Validate(req.Home); err != nil {\n\t\treturn nil, err\n\t}", "// planted: selected inventory without common admission", "^TestProvisioningOriginMigrationCommonBarrier$/^legacy-neighbor$", "migration common barrier before backend:", 1},
		{"stale-backend-cache", "internal/envprofile/managed.go", "req.credentialPolicy = nil", "// planted: reuse pre-lock backend cache", "^TestProvisioningOriginBackendCacheBoundary$/^changed$", "intended backend transition refusal:", 2},
	} {
		t.Run(plant.name, func(t *testing.T) {
			path := filepath.Join(root, filepath.FromSlash(plant.file))
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(raw), "\r\n", "\n")
			if count := strings.Count(source, plant.before); count != plant.count {
				t.Fatalf("plant anchor count %d, want %d", count, plant.count)
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, "replacement.go")
			if err := os.WriteFile(replacement, []byte(strings.ReplaceAll(source, plant.before, plant.after)), 0600); err != nil {
				t.Fatal(err)
			}
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{path: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlayPath := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
				t.Fatal(err)
			}
			command := exec.Command("go", "test", "-p=1", "-overlay", overlayPath, "-count=1", "-timeout=90s", "-v", "./internal/envprofile", "-run", plant.run)
			command.Dir = root
			// The focused parent harness supplies only owned disposable values. Keep
			// the child allowlist closed; never inherit selectors, Git or SSH settings.
			for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "GOTMPDIR", "GOCACHE", "GOPATH", "GOMODCACHE", "GOTOOLCHAIN", "GOPROXY", "GOSUMDB", "GOFLAGS", "GOENV", "GOTELEMETRY", "CGO_ENABLED", "LANG", "LC_ALL", "CURATOR_FIXTURE_PID_LOG"} {
				if value, ok := os.LookupEnv(key); ok {
					command.Env = append(command.Env, key+"="+value)
				}
			}
			var output bytes.Buffer
			command.Stdout = &output
			command.Stderr = &output
			if err := command.Start(); err != nil {
				t.Fatal(err)
			}
			pid := command.Process.Pid
			err = command.Wait()
			t.Logf("detector child PID %d reaped; exit %d", pid, command.ProcessState.ExitCode())
			exit, ok := err.(*exec.ExitError)
			failure := strings.NewReplacer("^", "", "$", "").Replace(plant.run)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(output.String(), "--- FAIL: "+failure+" ") || !strings.Contains(output.String(), plant.assertion) || strings.Contains(output.String(), "panic:") || strings.Contains(output.String(), "[build failed]") || strings.Contains(output.String(), "no tests to run") {
				t.Fatalf("plant survived or missed intended owner assertion: %v\n%s", err, output.String())
			}
			t.Logf("expected-red: %s: %s", failure, plant.assertion)
		})
	}
}
