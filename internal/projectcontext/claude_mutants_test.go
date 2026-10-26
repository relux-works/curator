package projectcontext

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Claim: each bounded source plant breaks its intended adapter assertion, not
// setup or compilation. Production witnesses run separately without mutations.
func TestClaudeNarrowingMutants(t *testing.T) {
	if os.Getenv("CURATOR_PROJECT_CLAUDE_MUTANTS") != "1" {
		t.Skip("set CURATOR_PROJECT_CLAUDE_MUTANTS=1 for the explicit bounded mutation lane")
	}
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		t.Skip("native contribution witnesses require measured host tuple")
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ name, file, test, old, new, witness string }{
		{"uncanonical-empty-grants", "claude_adapter.go", "TestClaudeEmptyControlsPublication", `out.AllowTools = append([]string{}, input.AllowTools...)`, `out.AllowTools = input.AllowTools`, "published empty controls cannot resolve"},
		{"project-before-profile", "composition_types.go", "TestClaudeCompositionMaterialization", `return bytes.Join(parts, []byte{'\n'})`, `if len(parts)>1 {parts[0],parts[len(parts)-1]=parts[len(parts)-1],parts[0]}; return bytes.Join(parts, []byte{'\n'})`, "Claude profile-first bytes differ"},
		{"repository-permissions", "composition.go", "TestClaudeCompositionMaterialization", "\tpolicyObject := struct {", "if req.Claude!=nil { b,_:=os.ReadFile(filepath.Join(r.Record.ProjectRoot,\".claude/settings.json\"));out.Files[\"settings.json\"]=b };\n\tpolicyObject := struct {", "repository permission imported"},
		{"omit-setting-sources", "claude_adapter.go", "TestClaudeCompositionMaterialization", `"--setting-sources", "user", `, ``, "Claude source controls missing"},
		{"allow-receipt-byte-drift", "claude_adapter.go", "TestClaudeReceiptAndPublicationRefusals", `digest(b) != f.SHA256`, `false`, "want project_context_drift"},
		{"allow-receipt-mode-drift", "claude_adapter.go", "TestClaudeReceiptAndPublicationRefusals", `uint32(st.Mode().Perm()) != uint32(mode)`, `uint32(st.Mode().Perm()) != uint32(mode) && false`, "want project_context_drift"},
		{"allow-resume", "claude_adapter.go", "TestClaudeCertificationAndResume", `if entry != ClaudeFresh {`, `if entry != ClaudeFresh { return out,nil }; if entry != ClaudeFresh {`, "resume refusal lost"},
		{"qualify-late-reloads", "claude_adapter.go", "TestClaudeCertificationAndResume", `// No complete tuple has native qualification in this release.`, `return nil // planted certification bypass`, "want project_context_unsupported"},
		{"allow-unqualified-tuple", "composition.go", "TestClaudeNativeOptIn", `if req.Claude != nil && (req.Adapter.Environment`, `if req.Claude != nil && false && (req.Adapter.Environment`, "want project_context_unsupported"},
		{"drop-policy-version", "claude_adapter.go", "TestClaudeManagerControls", `if c.Version != 1 {`, `if c.Version != 1 && false {`, "want project_context_invalid"},
		{"drop-policy-order", "claude_adapter.go", "TestClaudeManagerControls", `!sort.StringsAreSorted(names)`, `!sort.StringsAreSorted(names) && false`, "want project_context_invalid"},
		{"drop-policy-name-check", "claude_adapter.go", "TestClaudeManagerControls", `!claudeToolName.MatchString(n)`, `!claudeToolName.MatchString(n) && false`, "want project_context_invalid"},
		{"enable-mcp-tools-by-default", "claude_adapter.go", "TestClaudeManagerControls", `"mcp__*"`, `"Read"`, "manager policy not programmed"},
		{"omit-memory-off", "claude_adapter.go", "TestClaudeCompositionMaterialization", `out.MemoryDisabled = true`, `out.MemoryDisabled = false`, "managed home or memory control missing"},
		{"drop-tool-preset-check", "claude_adapter.go", "TestClaudeManagerControls", `n == "default"`, `n == "default" && false`, "want project_context_invalid"},
		{"drop-policy-unique", "claude_adapter.go", "TestClaudeManagerControls", `(i > 0 && names[i-1] == n)`, `(i > 0 && names[i-1] == n && false)`, "want project_context_invalid"},
		{"drop-policy-bound", "claude_adapter.go", "TestClaudeManagerControls", `len(names) > 256`, `len(names) > 256 && false`, "want project_context_invalid"},
		{"drop-mcp-policy-name-check", "claude_adapter.go", "TestClaudeManagerControls", `strings.HasPrefix(n, "mcp__")`, `strings.HasPrefix(n, "mcp__") && false`, "want project_context_invalid"},
		{"allow-legacy-opt-in", "claude_adapter.go", "TestClaudeNativeOptIn", `if req.Claude == nil || req.Adapter.Environment != "claude_code" {`, `if req.Claude == nil && false || req.Adapter.Environment != "claude_code" {`, "want project_context_unsupported"},
		{"allow-root-mode-drift", "claude_adapter.go", "TestClaudeReceiptAndPublicationRefusals", `if err := privatedir.Validate(home); err != nil {`, `if err := privatedir.Validate(home); err != nil && false {`, "want project_context_drift"},
		{"ignore-mcp-transport", "claude_adapter.go", "TestClaudeManagerControls", `bytes.TrimSuffix(files[contextmaterialize.MCPClaudePath], []byte{'\n'})`, `bytes.TrimSuffix([]byte("{\"mcpServers\":{}}"),nil)`, "want project_context_unsupported"},
		{"native-project-prose", "claude_native_project.go", "TestClaudeNativeProjectPreparation", `files := cloneClaudeFiles(documents)`, `files := cloneClaudeFiles(preview.Files)`, "native profile-only root differs"},
		{"native-drop-profile-system", "claude_native_project.go", "TestClaudeNativeProjectPreparation", `files := cloneClaudeFiles(documents)`, `files := cloneClaudeFiles(documents);delete(files,".agent-context/system-prompt.md")`, "native profile-only system differs"},
		{"native-launch-without-qualification", "claude_native_project.go", "TestClaudeNativeProjectQualification", `return out, &ClaudeDiagnostic{Code: "project_context_unsupported", Detail: "Claude native-project requires qualified manager permissions and memory-off", Mode: mode, Notice: r.Notice, Refused: r.Refused}`, `_ = r; return out, nil`, "native launch qualification bypass"},
		{"native-omit-notice", "claude_native_project.go", "TestClaudeNativeProjectQualification", `r.Notice = projectadapter.NativeProjectNotice`, `r.Notice = ""`, "native status notice missing"},
		{"native-policy-rebinding", "claude_native_project.go", "TestClaudeNativeProjectInputs", `if !bytes.Equal(files["settings.json"], preview.Files["settings.json"]) || !bytes.Equal(files[claudeOwnershipPath], preview.Files[claudeOwnershipPath]) {`, `if false && (!bytes.Equal(files["settings.json"], preview.Files["settings.json"]) || !bytes.Equal(files[claudeOwnershipPath], preview.Files[claudeOwnershipPath])) {`, "want project_context_invalid"},
		{"native-profile-surfaces", "claude_native_project.go", "TestClaudeNativeProjectInputs", `(f.Path != "CLAUDE.md" && f.Path != ".agent-context/system-prompt.md")`, `false`, "want project_context_invalid"},
		{"native-profile-mode", "claude_native_project.go", "TestClaudeNativeProjectInputs", `f.Mode != "0600"`, `false`, "want project_context_invalid"},
		{"native-receipt-grammar", "claude_native_project.go", "TestClaudeNativeFileSet", `if err := validateInventory(captured{Manifest: Snapshot{1, receipt}}); err != nil {`, `if err := validateInventory(captured{Manifest: Snapshot{1, receipt}}); err != nil && false {`, "want project_context_invalid"},
		{"native-receipt-digest", "claude_native_project.go", "TestClaudeNativeFileSet", `digest(b) != f.SHA256`, `false`, "want project_context_invalid"},
		{"native-receipt-size", "claude_native_project.go", "TestClaudeNativeFileSet", `int64(len(b)) != f.Size`, `false`, "want project_context_invalid"},
		{"native-receipt-extra", "claude_native_project.go", "TestClaudeNativeFileSet", `if len(files) != len(receipt) {`, `if len(files) != len(receipt) && false {`, "want project_context_invalid"},
		{"native-implicit-mode", "claude_native_project.go", "TestClaudeNativeProjectInputs", "func requireClaudeNativeMode(mode string) error {\n\tif mode != projectadapter.NativeProject {", "func requireClaudeNativeMode(mode string) error {\n\tif false {", "want project_context_invalid"},
		{"native-base-ref-binding", "claude_native_project.go", "TestClaudeNativeProjectInputs", `preview.Ref != "composition:"+hashID(preview.SHA256)`, `false`, "want project_context_invalid"},
		{"native-base-sha-binding", "claude_native_project.go", "TestClaudeNativeProjectInputs", `identity(preview.Record) != preview.SHA256`, `false`, "want project_context_invalid"},
		{"native-base-receipt-binding", "claude_native_project.go", "TestClaudeNativeProjectInputs", `identity(preview.Receipt) != preview.Record.MaterializationSHA256`, `false`, "want project_context_invalid"},
		{"native-base-environment-binding", "claude_native_project.go", "TestClaudeNativeProjectInputs", `preview.Record.Environment != "claude_code"`, `false`, "want project_context_invalid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			file := filepath.Join(cwd, tc.file)
			source, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(source), tc.old) != 1 {
				t.Fatalf("plant anchor count differs: %s", tc.name)
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, "plant.go")
			if err = os.WriteFile(replacement, []byte(strings.Replace(string(source), tc.old, tc.new, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(dir, "overlay.json")
			raw, _ := json.Marshal(map[string]any{"Replace": map[string]string{file: replacement}})
			if err = os.WriteFile(overlay, raw, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(compositionGo(), "test", "-overlay", overlay, "-count=1", "-run", "^"+tc.test+"$", ".")
			cmd.Dir = cwd
			for _, e := range os.Environ() {
				if !strings.HasPrefix(e, "GIT_DIR=") && !strings.HasPrefix(e, "GIT_WORK_TREE=") {
					cmd.Env = append(cmd.Env, e)
				}
			}
			out, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(out), "--- FAIL: "+tc.test) || !strings.Contains(string(out), tc.witness) || strings.Contains(string(out), "[build failed]") {
				t.Fatalf("mutant missed intended witness: %v\n%s", err, out)
			}
			t.Logf("%s: caught intended witness", tc.name)
		})
	}
}
