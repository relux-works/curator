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
		{"project-untrusted", "codex.go", ".", "TestCodexLaunchAndCertification", `"trust_level": "trusted"`, `"trust_level": "untrusted"`, "project controls not applied"},
		{"native-project-launch-refusal", "codex.go", ".", "TestCodexNativeProjectRefusal", `return &Unsupported{Permissions, codexNativeProjectReason}`, `return nil`, "incompatible native-project launch accepted"},
		{"native-project-permission-certification", "codex.go", ".", "TestCodexNativeProjectRefusal", `Certified: []SurfaceName{Context, Skills, MCP, MemoryOff}`, `Certified: []SurfaceName{Context, Skills, MCP, Permissions, MemoryOff}`, "unsafe native-project permissions certified"},
		{"native-bundle-delivery", "codex.go", ".", "TestCodexSystemBundle", `p := &Plan{files: bundle,`, `bundle = map[string][]byte{}; p := &Plan{files: bundle,`, "pinned native bundle missing"},
		{"native-bundle-provenance", "codex_bundle.go", ".", "TestCodexBundleProvenance", `if fmt.Sprintf("%x", sha256.Sum256(archive)) != codexSystemBundleSHA {`, `if fmt.Sprintf("%x", sha256.Sum256(archive)) != codexSystemBundleSHA && false {`, "unpinned native bundle accepted"},
		{"native-bundle-reservation", "codex.go", ".", "TestCodexSystemBundle", `if folded == "skills/.system" || strings.HasPrefix(folded, "skills/.system/") {`, `if (folded == "skills/.system" || strings.HasPrefix(folded, "skills/.system/")) && false {`, "reserved native bundle replacement accepted"},
		{"absent-root-inventory", "materialize.go", ".", "TestCodexInstructionInventory", `[]string{"AGENTS.md", "AGENTS.override.md"}`, `[]string{"AGENTS.override.md"}`, "instruction inventory drift accepted"},
		{"native-bundle-extras", "materialize.go", ".", "TestCodexSystemBundleDrift/extra", `if _, ok := p.generated[filepath.ToSlash(rel)]; !ok {`, `if strings.HasPrefix(filepath.ToSlash(rel), "skills/.system/") { return nil }; if _, ok := p.generated[filepath.ToSlash(rel)]; !ok {`, "native bundle drift accepted"},
		{"guarded-absence", "materialize.go", ".", "TestCodexManagerMetadata", `return stateread.LstatWith(filepath.Join(base, filepath.FromSlash(name)), func(string) (os.FileInfo, error) { return root.Lstat(name) })`, `info, err := root.Lstat(name); if os.IsNotExist(err) { return stateread.Metadata{Kind:stateread.KindAbsent},nil }; if err!=nil {return stateread.Metadata{Kind:stateread.KindUnreadable},err};return stateread.Metadata{Kind:stateread.KindPresent,Info:info},nil`, "unreadable ancestor collapsed to absence: broken/child"},
		{"mcp-enumeration-unknown", "../../.research/project-exclusion-codex/adapter_rework_test.go", "../../.research/project-exclusion-codex", "TestMCPCeilingParser", `if !allowed {`, `if !allowed && false {`, "unsafe MCP candidate input accepted"},
		{"config-fence-refusal-path", "codex.go", ".", "TestCodexNativeProjectRefusal", `project_config_changed digest preflight`, `digest preflight`, "config-fence refusal omits measured path"},
		{"config-fence-skipped", "../../.research/project-exclusion-codex/config_fence_test.go", "../../.research/project-exclusion-codex", "TestConfigFenceAdmission/thread/late-server", `if err := f.check(); err != nil {`, `if err := f.check(); err != nil && false {`, "changed config reached thread dispatch"},
		{"config-fence-file-list-only", "../../.research/project-exclusion-codex/config_fence_test.go", "../../.research/project-exclusion-codex", "TestConfigFenceAdmission/thread/disabled-tools", `if !reflect.DeepEqual(current, f.baseline) {`, `if !reflect.DeepEqual(configPaths(current), configPaths(f.baseline)) {`, "changed config reached thread dispatch"},
		{"config-fence-parent-link", "../../.research/project-exclusion-codex/config_fence_test.go", "../../.research/project-exclusion-codex", "TestConfigFenceAdmission/thread/linked-directory", `if metadata.Kind == stateread.KindPresent && metadata.Info.Mode()&os.ModeSymlink != 0 {`, `if metadata.Kind == stateread.KindPresent && metadata.Info.Mode()&os.ModeSymlink != 0 && false {`, "changed config reached thread dispatch"},
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
