package permissionsmemory

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Claim: missing opt-in and a foreign digest refuse before creating a fixture;
// a foreign file is never executed. Real exact-pin positives run separately.
func TestIntake(t *testing.T) {
	if os.Getenv("QUAL_INTAKE_CHILD") == "1" {
		setupBinary(t, "QUAL_INTAKE_BINARY", map[string]bool{museSHA: true})
		return
	}
	dir := t.TempDir()
	wrong := filepath.Join(dir, "foreign")
	write(t, wrong, "not a vendor executable")
	for _, kind := range []string{"missing", "digest"} {
		t.Run(kind, func(t *testing.T) {
			root := filepath.Join(dir, kind, ".temp", "run")
			binary := ""
			if kind == "digest" {
				binary = wrong
			}
			cmd := exec.Command(os.Args[0], "-test.run=^TestIntake$", "-test.v")
			for _, e := range os.Environ() {
				if !strings.HasPrefix(e, "QUAL_INTAKE_") && !strings.HasPrefix(e, "QUAL_RUN_ROOT=") {
					cmd.Env = append(cmd.Env, e)
				}
			}
			cmd.Env = append(cmd.Env, "QUAL_INTAKE_CHILD=1", "QUAL_INTAKE_BINARY="+binary, "QUAL_RUN_ROOT="+root)
			out, err := cmd.CombinedOutput()
			if kind == "missing" {
				if err != nil || !strings.Contains(string(out), "unset; no vendor process or fixture") {
					t.Fatal("missing opt-in was not specific")
				}
			} else if err == nil || !strings.Contains(string(out), "unqualified binary digest") {
				t.Fatal("foreign digest did not refuse")
			}
			if _, err := os.Stat(root); !os.IsNotExist(err) {
				t.Fatal("intake refusal created fixture")
			}
		})
	}
}

// Claim: each opt-in narrowing plant fails its intended native or synthetic
// assertion. No vendor runs under a weakened network/Keychains fence. Skips,
// setup errors and unrelated failures never count as a caught plant.
func TestNarrowing(t *testing.T) {
	if os.Getenv("QUAL_RUN_MUTANTS") != "1" {
		t.Skip("explicit bounded mutation lane")
	}
	if os.Getenv("QUAL_CODEX_BINARY") == "" || os.Getenv("QUAL_MUSE_BINARY") == "" {
		t.Fatal("both exact binary locators required")
	}
	root := os.Getenv("QUAL_RUN_ROOT")
	if root == "" {
		t.Fatal("fresh task-local root required")
	}
	for _, row := range []struct{ name, test, key, value, witness string }{
		{"memory-on", "TestCodexMemoryRead", "QUAL_MUTANT", "memory-on", "memory canary escaped off gate"},
		{"muse-web-enabled", "TestMuseToolDenials/web_fetch", "QUAL_MUTANT", "muse-tool-enabled", "Muse disabled web_fetch executed"},
		{"muse-tool-enabled", "TestMuseToolDenials/write_file", "QUAL_MUTANT", "muse-tool-enabled", "Muse disabled write_file executed"},
		{"mcp-enabled", "TestCodexMCPDenial", "QUAL_MUTANT", "mcp-enabled", "disabled MCP tool executed"},
		{"mcp-prompt-auto", "TestCodexMCPDefaultPrompt", "QUAL_MUTANT", "mcp-prompt-auto", "disabled MCP tool executed"},
		{"dedicated-memory-on", "TestCodexDedicatedMemory/write", "QUAL_MUTANT", "memory-on", "memory tool escaped off gate"},
		{"muse-edit-enabled", "TestMuseToolDenials/edit_file", "QUAL_MUTANT", "muse-tool-enabled", "Muse disabled edit_file executed"},
		{"muse-memory-write-enabled", "TestMuseToolDenials/add_memory", "QUAL_MUTANT", "muse-tool-enabled", "Muse disabled add_memory executed"},
		{"muse-shell-enabled", "TestMuseToolDenials/legacy-bash", "QUAL_MUTANT", "muse-tool-enabled", "Muse disabled legacy-bash executed"},
		{"nested-as-policy", "TestFilesystemDenialAttribution", "QUAL_MUTANT", "nested-as-policy", "nested startup misattributed to manager denial"},
		{"request-attribution", "TestRequestAttribution", "QUAL_MUTANT", "request-attribution", "background request misattributed to main turn"},
		{"result-attribution", "TestFunctionResults", "QUAL_MUTANT", "result-attribution", "tool-result attribution lost"},
		{"unknown-qualified", "TestVerdictGate", "QUAL_MUTANT", "unknown-qualified", "false qualification"},
		{"collector-prompt", "TestCollector/bounded", "QUAL_COLLECT_MUTANT", "prompt", "native prompt leaked"},
		{"digest", "TestIntake", "QUAL_MUTANT", "digest", "foreign digest did not refuse"},
		{"network", "TestOfflineBoundary", "QUAL_BOUNDARY_MUTANT", "network", "unrelated loopback not denied"},
		{"keychains", "TestOfflineBoundary", "QUAL_BOUNDARY_MUTANT", "keychains", "synthetic Keychains read not denied"},
	} {
		t.Run(row.name, func(t *testing.T) {
			pattern := "^" + strings.Split(row.test, "/")[0] + "$"
			if strings.Contains(row.test, "/") {
				pattern += "/" + strings.Split(row.test, "/")[1] + "$"
			}
			cmd := exec.Command(os.Args[0], "-test.run="+pattern, "-test.v")
			for _, e := range os.Environ() {
				if !strings.HasPrefix(e, "QUAL_RUN_ROOT=") {
					cmd.Env = append(cmd.Env, e)
				}
			}
			cmd.Env = append(cmd.Env, "QUAL_RUN_ROOT="+filepath.Join(root, row.name), row.key+"="+row.value, "QUAL_EXPECT_DENY=1")
			out, err := cmd.CombinedOutput()
			text := string(out)
			if err == nil || !strings.Contains(text, row.witness) || !strings.Contains(text, "--- FAIL: "+strings.Split(row.test, "/")[0]) || strings.Contains(text, "panic:") || strings.Contains(text, "unqualified binary") || strings.Contains(text, "--- SKIP:") {
				t.Fatalf("plant missed intended witness: %v\n%s", err, out)
			}
			t.Log("caught intended assertion: " + row.witness)
		})
	}
}
