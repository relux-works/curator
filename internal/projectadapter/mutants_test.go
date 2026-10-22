package projectadapter

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Claim: each opt-in source plant reaches its named forbidden behavior and the
// matching regression assertion fails. Skips, crashes, compile/setup failures
// and unrelated test failures never count as a caught narrowing mutant.
func TestCodexNarrowingMutants(t *testing.T) {
	if os.Getenv("CURATOR_CODEX_ADAPTER_MUTANTS") != "1" {
		t.Skip("explicit bounded mutation lane")
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ name, file, pkg, test, old, new, witness string }{
		{"project-first", "../projectcontext/composition_types.go", "../projectcontext", "TestCodexAdapterCompositionEndToEnd", `return bytes.Join(parts, []byte{'\n'})`, `if len(parts)>1 { parts[0],parts[len(parts)-1]=parts[len(parts)-1],parts[0] }; return bytes.Join(parts, []byte{'\n'})`, "profile-first root bytes differ"},
		{"repository-approval", "codex.go", "../projectcontext", "TestCodexAdapterCompositionEndToEnd", `replace("approval_policy", req.Approval)`, `repoBytes,readErr:=os.ReadFile(filepath.Join(req.Worktree,".codex/config.toml")); if readErr!=nil { return nil,readErr }; repo:=map[string]any{}; if _,decodeErr:=toml.Decode(string(repoBytes),&repo);decodeErr!=nil{return nil,decodeErr}; replace("approval_policy",repo["approval_policy"])`, "repository approval imported"},
		{"seed-mcp-survives", "codex.go", ".", "TestCodexCompositionAndSeed", `replace("mcp_servers", servers)`, `if seed,ok:=config["mcp_servers"].(map[string]any);ok {for name,v:=range seed{servers[name]=v}};replace("mcp_servers", servers)`, "seed MCP survived composition"},
		{"receipt-drift", "codex.go", ".", "TestCodexRefusals/drift", `if err := validateInput(req.Files, req.Generated); err != nil {`, `if err := validateInput(req.Files, req.Generated); err != nil && false {`, "want receipt refusal"},
		{"strict-certification", "codex.go", ".", "TestCodexLaunchAndCertification", `return &Unsupported{NativeSourceExclusion, codexExclusionReason}`, `return nil`, "strict launch did not refuse"},
		{"strict-mode", "codex.go", ".", "TestCodexRefusals/strict", `if req.Mode != NativeProject {`, `if req.Mode != NativeProject && false {`, "want only native-project refusal"},
		{"native-project-opt-in", "codex.go", ".", "TestCodexRefusals/no-opt-in", `if !req.NativeProjectEnabled {`, `if !req.NativeProjectEnabled && false {`, "want not enabled by the operator refusal"},
		{"project-root-installed", "codex.go", "../projectcontext", "TestCodexAdapterCompositionEndToEnd", `// Native-project never installs concatenated project instructions.`, `p.files[name]=bytes.Clone(b)`, "native-project managed root is not profile-only"},
		{"release-pin", "codex.go", ".", "TestCodexRefusals/version", `if req.ProviderVersion != "0.162.1" || req.Platform != "macos" {`, `if (req.ProviderVersion != "0.162.1" || req.Platform != "macos") && false {`, "want release/platform refusal"},
		{"seed-revision", "codex.go", ".", "TestCodexRefusals/seed-revision", `if req.SeedRevision != envregistry.CodexSeedRevisionA && req.SeedRevision != envregistry.CodexSeedRevisionB {`, `if (req.SeedRevision != envregistry.CodexSeedRevisionA && req.SeedRevision != envregistry.CodexSeedRevisionB) && false {`, "want seed revision refusal"},
		{"manager-policy", "codex.go", ".", "TestCodexRefusals/policy", `if (req.Approval != "never" && req.Approval != "on-request") || (req.Sandbox != "read-only" && req.Sandbox != "workspace-write") {`, `if ((req.Approval != "never" && req.Approval != "on-request") || (req.Sandbox != "read-only" && req.Sandbox != "workspace-write")) && false {`, "want manager approval refusal"},
		{"safe-path", "codex.go", ".", "TestCodexRefusals/parent", `if !safePath(name) {`, `if !safePath(name) && false {`, "want unsafe output refusal"},
		{"path-alias", "codex.go", ".", "TestCodexInputBoundaries/alias", `if aliases[alias] {`, `if aliases[alias] && false {`, "want output path alias refusal"},
		{"seed-alternate-source", "codex.go", ".", "TestCodexRefusals/seed-hooks", `return nil, fmt.Errorf("project_context_unsupported: seed control key %s requires separate admission", key)`, `continue`, "want seed control key hooks refusal"},
		{"http-unqualified", "codex.go", ".", "TestCodexInputBoundaries/mcp-http", `if _, http := table["url"]; http {`, `if _, http := table["url"]; http && false {`, "want HTTP native entry loading is unqualified refusal"},
		{"memory-on", "codex.go", ".", "TestCodexLaunchAndCertification", `setTable(row.table, row.key, false)`, `setTable(row.table, row.key, true)`, "memory overrides missing"},
		{"project-trusted", "codex.go", ".", "TestCodexLaunchAndCertification", `"trust_level": "untrusted"`, `"trust_level": "trusted"`, "project controls not applied"},
		{"allocation", "materialize.go", ".", "TestCodexPublicationBoundaries/allocation", `parts[1] != "project-views"`, `false`, "publication boundary accepted"},
		{"generated-hash", "materialize.go", ".", "TestCodexMaterializeAndDrift/bytes", `surface(b, expected.Mode) != expected`, `(surface(b, expected.Mode) != expected && false)`, "drift accepted"},
		{"prepared-lease-id", "publisher.go", ".", "TestCodexPreparedPublisher", `!identifiers.Valid(leaseID)`, `(!identifiers.Valid(leaseID) && false)`, "invalid publication accepted"},
		{"home-lock", "materialize.go", ".", "TestCodexPublicationBoundaries/zero-lock", `if err := held.Validate(managerHome); err != nil {`, `if err := held.Validate(managerHome); err != nil && false {`, "publication boundary accepted"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			file := filepath.Clean(filepath.Join(cwd, tc.file))
			source, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(source), tc.old) != 1 {
				t.Fatal("mutation anchor differs")
			}
			mutant := strings.Replace(string(source), tc.old, tc.new, 1)
			if tc.name == "http-unqualified" {
				mutant = strings.Replace(mutant, `if !ok || command == "" {`, `if (!ok || command == "") && false {`, 1)
			}

			if tc.name == "repository-approval" {
				mutant = strings.Replace(mutant, `"bytes"`, `"bytes"`+"\n\"os\"", 1)
			}
			// The second witness check must not mask the deliberately weakened gate.
			if tc.name == "home-lock" {
				mutant = strings.Replace(mutant, `if err = held.Validate(managerHome); err != nil {`, `if err = held.Validate(managerHome); err != nil && false {`, 1)
			}
			dir := t.TempDir()
			replacement := filepath.Join(dir, "mutant.go")
			if err = os.WriteFile(replacement, []byte(mutant), 0600); err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(dir, "overlay.json")
			raw, _ := json.Marshal(map[string]any{"Replace": map[string]string{file: replacement}})
			if err = os.WriteFile(overlay, raw, 0600); err != nil {
				t.Fatal(err)
			}
			goBinary := os.Getenv("CURATOR_GO")
			if goBinary == "" {
				t.Fatal("CURATOR_GO required for bounded mutation lane")
			}
			cmd := exec.Command(goBinary, "test", "-overlay", overlay, "-count=1", "-run", "^"+strings.Split(tc.test, "/")[0]+"$"+strings.TrimPrefix(tc.test, strings.Split(tc.test, "/")[0]), tc.pkg)
			cmd.Dir = cwd
			for _, v := range os.Environ() {
				if !strings.HasPrefix(v, "GIT_DIR=") && !strings.HasPrefix(v, "GIT_WORK_TREE=") {
					cmd.Env = append(cmd.Env, v)
				}
			}
			out, err := cmd.CombinedOutput()
			text := string(out)
			if err == nil || !strings.Contains(text, "--- FAIL: "+strings.Split(tc.test, "/")[0]) || !strings.Contains(text, tc.witness) || strings.Contains(text, "[build failed]") || strings.Contains(text, "panic:") {
				t.Fatalf("mutant did not hit intended witness: %v\n%s", err, out)
			}
			t.Logf("%s: caught intended assertion %q", tc.name, tc.witness)
		})
	}
}
