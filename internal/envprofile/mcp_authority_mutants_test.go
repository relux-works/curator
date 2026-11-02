package envprofile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Each plant removes the named admission or publication requirement. The
// bounded child must fail at the specific behavioral assertion; setup errors,
// compilation failures, skips and timeouts cannot count as killed mutants.
func TestMCPCredentialAuthorityNarrowingControls(t *testing.T) {
	isolateDetectorParent(t)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, file, anchor, run, witness string }{
		{"profile-reporting", "native_mcp.go", "", "^TestProfileMCPCredentialAuthority$/^status$/^node=true$", "profile MCP authority admission missing"},
		{"project-acquisition", "project_mcp.go", "func ResolveProjectMCP(", "^TestProjectMCPAuthorityBeforeState$/^resolve$/^node=true$", "MCP early authority admission missing"},
		{"project-publication", "project_mcp.go", "func PublishProjectMCPLock(", "^TestProjectMCPAuthorityBeforeState$/^publish$/^node=true$", "MCP early authority admission missing"},
		{"project-replay", "project_mcp.go", "func replayProjectMCP(", "^TestProjectMCPAuthorityBeforeState$/^replay$/^node=true$", "MCP early authority admission missing"},
		{"project-restore", "project_mcp.go", "func replayProjectMCP(", "^TestProjectMCPAuthorityBeforeState$/^restore$/^node=true$", "MCP early authority admission missing"},
		// Removing one redundant prepublication admission is no longer a policy
		// violation: the unit and engine independently re-admit authority. Plant
		// the actual IR5 omission instead; healthy publication must reach and
		// reject a missing complete unit at the engine, never a setup failure.
		{"profile-unit-omitted", "native_mcp.go", "Unit: unit, Directories: directories", "^TestProfileMCPFirstInstall$", "managed home effect requires a complete environment unit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(root, "internal", "envprofile", tc.file)
			body, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(body), "\r\n", "\n")
			call := "envauthority.Validate(home)"
			if tc.file == "native_mcp.go" {
				call = "envauthority.Validate(req.Home)"
			}
			replacement := "func() error { _ = envauthority.Validate; return nil }()"
			if tc.name == "profile-unit-omitted" {
				if strings.Count(source, tc.anchor) != 1 {
					t.Fatal("unit plant anchor must match once")
				}
				source = strings.Replace(source, tc.anchor, "Unit: nil, Directories: directories", 1)
			} else if tc.anchor == "" {
				if strings.Count(source, call) != 2 {
					t.Fatal("profile authority plants must match twice")
				}
				source = strings.ReplaceAll(source, call, replacement)
			} else {
				if strings.Count(source, tc.anchor) != 1 {
					t.Fatal("authority plant anchor must match once")
				}
				start := strings.Index(source, tc.anchor)
				offset := strings.Index(source[start:], call)
				if offset < 0 {
					t.Fatal("authority plant call is absent")
				}
				at := start + offset
				source = source[:at] + replacement + source[at+len(call):]
			}
			dir := t.TempDir()
			plant := filepath.Join(dir, "plant.go")
			if err := os.WriteFile(plant, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			body, err = json.Marshal(map[string]any{"Replace": map[string]string{path: plant}})
			if err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlay, body, 0600); err != nil {
				t.Fatal(err)
			}
			binary := os.Getenv("CURATOR_GO")
			if binary == "" {
				binary = "go"
			}
			// The outer five-minute and child four-minute limits only bound hangs.
			// Detection requires exit 1 and the named authority assertion below.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, "test", "-overlay", overlay, "-count=1", "-timeout=4m", "-v", "-run", tc.run, "./internal/envprofile") // #nosec G204 -- fixed bounded test arguments and selected Go toolchain
			cmd.Dir = root
			cmd.Env = detectorChildEnv()
			var buffer bytes.Buffer
			cmd.Stdout, cmd.Stderr = &buffer, &buffer
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			pid := cmd.Process.Pid
			err = cmd.Wait()
			t.Logf("detector child PID %d reaped; exit %d", pid, cmd.ProcessState.ExitCode())
			output := buffer.Bytes()
			var exit *exec.ExitError
			if ctx.Err() != nil || !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: ") || !strings.Contains(string(output), tc.witness) {
				t.Fatalf("MCP authority plant did not reach intended assertion: %v\n%s", err, output)
			}
			t.Logf("expected failure: %s", tc.witness)
		})
	}
}
