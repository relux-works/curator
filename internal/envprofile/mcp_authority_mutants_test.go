package envprofile

import (
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
		{"profile-publication", "native_mcp.go", "validateInputs := func() error {", "^TestProfileMCPCredentialPublicationBoundary$/^node=true$", "MCP publication authority refusal wrote after the marker changed"},
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
			if tc.anchor == "" {
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
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, "test", "-overlay", overlay, "-count=1", "-timeout=4m", "-v", "-run", tc.run, "./internal/envprofile") // #nosec G204 -- fixed bounded test arguments and selected Go toolchain
			cmd.Dir = root
			for _, entry := range os.Environ() {
				name, _, _ := strings.Cut(entry, "=")
				if !strings.EqualFold(name, "GIT_DIR") && !strings.EqualFold(name, "GIT_WORK_TREE") {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			output, err := cmd.CombinedOutput()
			var exit *exec.ExitError
			if ctx.Err() != nil || !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: ") || !strings.Contains(string(output), tc.witness) {
				t.Fatalf("MCP authority plant did not reach intended assertion: %v\n%s", err, output)
			}
			t.Logf("expected failure: %s", tc.witness)
		})
	}
}
