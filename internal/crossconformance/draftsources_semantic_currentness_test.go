package crossconformance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/registry"
)

// Currentness, substitution, and marker-plan semantic rows (§4): the
// attested/substituted positives report current, the strict and
// forbidden substitutions refuse before publication, and every
// marker/effective-plan mismatch reports noncurrent without mutation.

func init() {
	registerDraftSemantic("attested-network-current", driveAttestedNetworkCurrent)
	registerDraftSemantic("legacy-substitution-current", driveLegacySubstitutionCurrent)
	registerDraftSemantic("legacy-substitution-strict", driveLegacySubstitutionStrict)
	registerDraftSemantic("selector-substitution-forbidden", driveSelectorSubstitutionForbidden)
	registerDraftSemantic("local-required-registry", driveLocalRequiredRegistry)
	registerDraftSemantic("external-substitution-strict", driveExternalSubstitutionStrict)
	for _, field := range []string{"registry", "status", "key_id", "substituted", "package", "lock_sha256"} {
		field := field
		registerDraftSemantic("marker-plan-mismatch-"+field, func(t *testing.T, c draftSemanticCase) {
			driveMarkerPlanMismatch(t, c, field)
		})
	}
}

// driveAttestedCLIInstall resolves, installs, and verifies one
// attested network-git skill through the compiled CLI, returning the
// configured project for mutation rows.
func driveAttestedCLIInstall(t *testing.T, stub *attestStub) (configPath, project, home string, pathEnv []string) {
	t.Helper()
	realGit := requireGit(t)
	root := t.TempDir()
	bare, _ := draftCLIKitBare(t, root)
	const declared = "https://fixture.test/kit.git"
	fakeDir, _ := installDraftTransportShim(t, "https://never.invalid/x.git", "fatal: unexpected clone in transport fixture", declared, bare, realGit)
	pathEnv = draftTransportPATH(t, fakeDir)
	configPath, project, home = setupCLIProject(t, root)
	mergeRegistryConfig(t, configPath, stub)
	payload := `{"schema_version":2,"sources":{"kit":{"git":"` + declared + `","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	stub.set(attestGood)
	if code, stdout, stderr := runCurator(t, home, configPath, pathEnv, "project", "resolve", "app"); code != 0 {
		t.Fatalf("resolve = %d\n%s\n%s", code, stdout, stderr)
	}
	if code, stdout, stderr := runCurator(t, home, configPath, pathEnv, "install", "app"); code != 0 {
		t.Fatalf("install = %d\n%s\n%s", code, stdout, stderr)
	}
	if code, stdout, stderr := runCurator(t, home, configPath, pathEnv, "status", "app"); code != 0 {
		t.Fatalf("attested status = %d, want up-to-date\n%s\n%s", code, stdout, stderr)
	}
	return configPath, project, home, pathEnv
}

func driveAttestedNetworkCurrent(t *testing.T, _ draftSemanticCase) {
	stub := newAttestStub(t)
	_, project, _, _ := driveAttestedCLIInstall(t, stub)
	recorded := marker.Read(filepath.Join(project, ".agents", "skills", "review"))
	if recorded == nil || recorded.Attestation == nil {
		t.Fatalf("marker attestation = %+v, want a recorded attestation", recorded)
	}
	if recorded.Attestation.Registry != "one" || recorded.Attestation.Status != registry.StatusAudited || recorded.Attestation.KeyID == "" {
		t.Fatalf("attestation = %+v, want the registry/status/key binding", recorded.Attestation)
	}
}

// draftLegacyProject builds a schema-1 project resolving one tagged
// skill from a local skills root.
func draftLegacyProject(t *testing.T) (project, home, skillsRoot string) {
	t.Helper()
	project = t.TempDir()
	home = t.TempDir()
	skillsRoot = t.TempDir()
	skillDir := filepath.Join(skillsRoot, "review")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	runDraftGit(t, skillDir, "init", "-q", "-b", "main")
	writeLegacySkillFiles(t, skillDir, "review")
	runDraftGit(t, skillDir, "add", ".")
	runDraftGit(t, skillDir, "commit", "-qm", "init")
	runDraftGit(t, skillDir, "tag", "v1")
	payload := `{"schema_version":1,"agents":["claude_code"],"skills":[{"name":"review","tag":"v1"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	runDraftGit(t, project, "init", "-q")
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\nSkillfile.dev.json\n.claude/skills/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return project, home, skillsRoot
}

func writeLegacySkillFiles(t *testing.T, dir, name string) {
	t.Helper()
	files := map[string]string{
		"SKILL.md":                  "---\nname: " + name + "\ndescription: d\n---\n# " + name + "\n",
		"references/info.md":        "ref",
		"scripts/" + name + "-tool": "#!/bin/sh\necho " + name + "\n",
		"README.md":                 "dev docs",
	}
	for rel, content := range files {
		full := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	spec := map[string]any{
		"schema_version": 4, "capabilities": map[string]any{}, "runtime_roots": []string{"scripts"},
		"commands": map[string]any{
			name + "-tool": map[string]any{"type": "script", "unix_path": "scripts/" + name + "-tool", "win_path": "scripts/" + name + "-tool"},
		},
	}
	payload, _ := json.MarshalIndent(spec, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "csk-skill.json"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
}

func draftLegacyConfig(home, skillsRoot string) *config.Config {
	return &config.Config{Path: filepath.Join(home, "config.json"), SkillsRoot: skillsRoot, DefaultAgents: []string{"claude_code"}, AdapterMode: "auto"}
}

func writeDevSubstitution(t *testing.T, project, payload string) {
	t.Helper()
	if err := os.WriteFile(devsub.PathIn(project), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeDevPathSubstitution records one operator path substitution for
// the named skill. The document is JSON-encoded (never string
// interpolation) so a native Windows path stays valid JSON.
func writeDevPathSubstitution(t *testing.T, project, name, path string) {
	t.Helper()
	doc := map[string]any{"substitutions": map[string]any{name: map[string]any{"path": path}}}
	payload, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	writeDevSubstitution(t, project, string(payload))
}

// driveLegacySubstitutionCurrent proves an advisory install records an
// operator substitution and reports current through the compiled CLI.
func driveLegacySubstitutionCurrent(t *testing.T, _ draftSemanticCase) {
	root := t.TempDir()
	configPath, project, home := setupCLIProject(t, root)
	_, _, skillsRoot := draftLegacyProject(t)
	// Reconfigure the CLI project as the legacy project: same skills
	// root, legacy manifest, operator substitution.
	legacyPayload := `{"schema_version":1,"agents":["codex_cli"],"skills":[{"name":"review","tag":"v1"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(legacyPayload), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\nSkillfile.dev.json\n.codex/skills/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mergeSkillsRoot(t, configPath, skillsRoot)
	alt := t.TempDir()
	writeLegacySkillFiles(t, alt, "review")
	runDraftGit(t, alt, "init", "-q", "-b", "main")
	runDraftGit(t, alt, "add", ".")
	runDraftGit(t, alt, "commit", "-qm", "alt")
	writeDevPathSubstitution(t, project, "review", alt)
	if code, stdout, stderr := runCurator(t, home, configPath, nil, "install", "app"); code != 0 {
		t.Fatalf("install = %d\n%s\n%s", code, stdout, stderr)
	}
	recorded := marker.Read(filepath.Join(project, ".agents", "skills", "review"))
	if recorded == nil || recorded.Substituted == "" {
		t.Fatalf("marker substitution = %+v, want a recorded operator substitution", recorded)
	}
	if code, stdout, stderr := runCurator(t, home, configPath, nil, "status", "app"); code != 0 {
		t.Fatalf("status = %d, want current\n%s\n%s", code, stdout, stderr)
	}
}

// mergeSkillsRoot repoints a bootstrapped CLI config at a fixture
// skills root.
func mergeSkillsRoot(t *testing.T, configPath, skillsRoot string) {
	t.Helper()
	payload, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	doc["skills_root"] = skillsRoot
	merged, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, merged, 0o644); err != nil {
		t.Fatal(err)
	}
}

func driveLegacySubstitutionStrict(t *testing.T, _ draftSemanticCase) {
	project, home, skillsRoot := draftLegacyProject(t)
	alt := t.TempDir()
	writeLegacySkillFiles(t, alt, "review")
	writeDevPathSubstitution(t, project, "review", alt)
	cfg := draftLegacyConfig(home, skillsRoot)
	cfg.Audit.Enabled = true
	cfg.Audit.Mode = "strict"
	before := treeDigest(t, project) + treeDigest(t, home)
	result := install.Project(cfg, project, "test", install.Options{Platform: draftPlatform()})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "strict audit refuses substituted installs") {
		t.Fatalf("install = %+v, want the strict substitution refusal", result)
	}
	if after := treeDigest(t, project) + treeDigest(t, home); after != before {
		t.Fatal("refused install published state")
	}
}

func driveSelectorSubstitutionForbidden(t *testing.T, _ draftSemanticCase) {
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project, home := draftProject(t, payload, map[string]string{"skills/review": "review"})
	resolveDraftPlan(t, project, home, payload)
	alt := t.TempDir()
	writeDraftSkill(t, alt, "review")
	writeDevPathSubstitution(t, project, "review", alt)
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\n.claude/skills/\nSkillfile.dev.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := treeDigest(t, project) + treeDigest(t, home)
	cfg := draftInstallConfig(home)
	result := install.Project(cfg, project, "test", install.Options{DraftSourcesV1: true, Platform: draftPlatform()})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_selection_invalid") {
		t.Fatalf("install = %+v, want the from-selector substitution refusal", result)
	}
	if !strings.Contains(strings.Join(result.Errors, ";"), "development substitutions are not admitted with draft selectors") {
		t.Fatalf("install = %+v, want the draft-selector substitution diagnostic", result)
	}
	if after := treeDigest(t, project) + treeDigest(t, home); after != before {
		t.Fatal("refused install published state")
	}
}

func driveLocalRequiredRegistry(t *testing.T, _ draftSemanticCase) {
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project, home := draftProject(t, payload, map[string]string{"skills/review": "review"})
	resolveDraftPlan(t, project, home, payload)
	cfg := draftInstallConfig(home)
	cfg.Audit.RegistryPolicy = "strict"
	before := treeDigest(t, project) + treeDigest(t, home)
	result := install.Project(cfg, project, "test", install.Options{DraftSourcesV1: true, Platform: draftPlatform()})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "no network attestation identity") {
		t.Fatalf("install = %+v, want the local attestation refusal", result)
	}
	if after := treeDigest(t, project) + treeDigest(t, home); after != before {
		t.Fatal("refused install published state")
	}
}

func driveExternalSubstitutionStrict(t *testing.T, _ draftSemanticCase) {
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project, home := draftProject(t, payload, map[string]string{"skills/review": "review"})
	resolveDraftPlan(t, project, home, payload)
	alt := filepath.Join(project, "alt")
	writeDraftSkill(t, alt, "review")
	writeDevSubstitution(t, project, `{"schema_version":2,"substitutions":{},"build_repository_substitutions":{"review":{"tools":{"path":"alt"}}}}`)
	cfg := draftInstallConfig(home)
	cfg.Audit.Enabled = true
	cfg.Audit.Mode = "strict"
	before := treeDigest(t, project) + treeDigest(t, home)
	result := install.Project(cfg, project, "test", install.Options{DraftSourcesV1: true, Platform: draftPlatform()})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "strict audit refuses substituted installs") {
		t.Fatalf("install = %+v, want the strict substitution refusal", result)
	}
	if after := treeDigest(t, project) + treeDigest(t, home); after != before {
		t.Fatal("refused install published state")
	}
}

// driveMarkerPlanMismatch mutates one recorded marker field of an
// attested install and proves CLI status reports noncurrent without
// mutating anything.
func driveMarkerPlanMismatch(t *testing.T, _ draftSemanticCase, field string) {
	stub := newAttestStub(t)
	configPath, project, home, pathEnv := driveAttestedCLIInstall(t, stub)
	markerPath := filepath.Join(project, ".agents", "skills", "review", marker.Name)
	payload, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	switch field {
	case "registry", "status", "key_id":
		attestation, ok := doc["attestation"].(map[string]any)
		if !ok {
			t.Fatalf("baseline marker has no attestation: %v", doc["attestation"])
		}
		switch field {
		case "registry":
			attestation["registry"] = "other"
		case "status":
			attestation["status"] = "deprecated"
		case "key_id":
			attestation["key_id"] = "0000000000000000"
		}
	case "substituted":
		doc["substituted"] = "dev-checkout"
	case "package":
		pkg, ok := doc["package"].(map[string]any)
		if !ok {
			t.Fatal("baseline marker has no package")
		}
		commit, ok := pkg["commit"].(map[string]any)
		if !ok {
			t.Fatal("baseline marker package has no commit")
		}
		commit["hex"] = strings.Repeat("b", 40)
	case "lock_sha256":
		doc["lock_sha256"] = "sha256:" + strings.Repeat("9", 64)
	}
	mutated, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(markerPath, mutated, 0o644); err != nil {
		t.Fatal(err)
	}
	stub.set(attestGood)
	before := draftInstallStateDigest(t, project, home)
	code, stdout, stderr := runCurator(t, home, configPath, pathEnv, "status", "--check", "app")
	if code == 0 {
		t.Fatalf("status --check with a %s mismatch succeeded:\n%s\n%s", field, stdout, stderr)
	}
	if !strings.Contains(stdout+stderr, "needs-install") {
		t.Fatalf("status verdict misses needs-install:\n%s\n%s", stdout, stderr)
	}
	if after := draftInstallStateDigest(t, project, home); after != before {
		t.Fatal("status mutated state")
	}
}
