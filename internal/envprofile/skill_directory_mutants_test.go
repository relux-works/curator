package envprofile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/audit"
)

type directoryMutation struct{ file, before, after string }

// Overlays narrow real production gates; only go test exit 1 plus the named
// behavioral assertion counts as a kill. No deletion-only plants or static
// source checks stand in for the lifecycle suite.
func TestSkillDirectoryNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, pkg, run, fails, message string
		changes                        []directoryMutation
	}{
		{"schema-four-string-zero-weight", "contextpkg", "^TestReviewSkillDirectoryStringWeight$", "TestReviewSkillDirectoryStringWeight/weight-context/string", "lock Read admitted string weight", []directoryMutation{
			{"contextlock/contextlock.go", `weight, err := strconv.ParseInt(string(member.Weight), 10, 64)`, `if schemaVersion == SchemaVersion4 && string(member.Weight) == ` + "`\"0\"`" + ` { member.Weight = json.RawMessage("0") }
 weight, err := strconv.ParseInt(string(member.Weight), 10, 64)`},
		}},
		{"reuse-pre-markdown-ruleset", "audit", "^TestReviewPriorMarkdownAllowance$", "TestReviewPriorMarkdownAllowance/v1", "prior ruleset-4 allowance bypassed current Markdown network gate", []directoryMutation{
			{"audit/audit.go", `RulesetVersion = "` + audit.RulesetVersion + `"`, `RulesetVersion = "4"`},
		}},
		{"schema-one-directory", "contextpkg", "^TestSkillDirectorySchemaVectors$/^manifest-v1-still-refuses-skill-directory$", "TestSkillDirectorySchemaVectors/manifest-v1-still-refuses-skill-directory", "valid=false", []directoryMutation{
			{"contextpkg/contextpkg.go", `if schemaVersion == 2 {`, `if schemaVersion == 2 || schemaVersion == 1 {`},
		}},
		{"selected-frontmatter-description", "envprofile", "^TestSkillDirectoryMetadata$/^empty-description$", "TestSkillDirectoryMetadata/empty-description", "want refusal manifest_invalid", []directoryMutation{
			{"manifest/expand.go", `if !ok || strings.TrimSpace(description) == "" {`, `if !ok || (strings.TrimSpace(description) == "" && name != "formatter") {`},
		}},
		{"selected-frontmatter-name", "envprofile", "^TestSkillDirectoryMetadata$/^wrong-name$", "TestSkillDirectoryMetadata/wrong-name", "want refusal manifest_invalid", []directoryMutation{
			{"envprofile/gitsource.go", `if skillName != name {`, `if skillName != name && !(name == "formatter" && skillName == "sibling") {`},
		}},
		{"root-versus-selected-agreement", "contextresolve", "^TestSkillDirectoryAgreement$/^omitted-selector$", "TestSkillDirectoryAgreement/omitted-selector", "want context_range_conflict, got <nil>", []directoryMutation{
			{"contextresolve/contextresolve.go", `if directorySeen && directory != c.requirement.Directory {`, `if directorySeen && directory != c.requirement.Directory && directory != "" && c.requirement.Directory != "" {`},
		}},
		{"parent-component", "envprofile", "^TestSkillDirectoryLifecycleVectors$/^parent-component-refused$", "TestSkillDirectoryLifecycleVectors/parent-component-refused", "got <nil>", []directoryMutation{
			{"contextpkg/contextpkg.go", `if !ok || !identifiers.PortablePath(text) {`, `if !ok || (!identifiers.PortablePath(text) && text != "skills/../formatter") {`},
			{"envprofile/gitsource.go", `if !identifiers.PortablePath(directory) {`, `if !identifiers.PortablePath(directory) && directory != "skills/../formatter" {`},
			{"contextlock/contextlock.go", `if member.Directory != "" && !identifiers.PortablePath(member.Directory) {`, `if member.Directory != "" && !identifiers.PortablePath(member.Directory) && member.Directory != "skills/../formatter" {`},
			{"contextlock/contextlock.go", `if err := json.Unmarshal(member.Directory, &directory); err != nil || !identifiers.PortablePath(directory) {`, `if err := json.Unmarshal(member.Directory, &directory); err != nil || (!identifiers.PortablePath(directory) && directory != "skills/../formatter") {`},
		}},
		{"inside-directory-symlink", "envprofile", "^TestSkillDirectoryLifecycleVectors$/^symlinked-leaf-inside-refused$", "TestSkillDirectoryLifecycleVectors/symlinked-leaf-inside-refused", "got <nil>", []directoryMutation{
			// Admit only this one tree link, preserving all other link refusals.
			// Expand its valid in-snapshot target as an unsafe following extractor
			// would; otherwise the directory-exists gate would mask the witness.
			{"gitops/gitops.go", `case entry.mode == "120000":`, `case entry.mode == "120000" && entry.path == "skills/formatter":
 entry.mode = "100644"
 case entry.mode == "120000":`},
			{"gitops/gitops.go", `plan, err := planWrites(destRoot, entries)`, `var admitted []treeEntry
 for _, entry := range entries {
 if entry.path != "skills/formatter" { admitted = append(admitted, entry); continue }
 for _, target := range entries { if strings.HasPrefix(target.path, "sibling/") { target.path = "skills/formatter/" + strings.TrimPrefix(target.path, "sibling/"); admitted = append(admitted, target) } }
 }
 entries = admitted
 plan, err := planWrites(destRoot, entries)`},
		}},
		{"borrow-root-capabilities", "envprofile", "^TestSkillDirectoryLifecycleVectors$/^selected-root-undeclared-contact$", "TestSkillDirectoryLifecycleVectors/selected-root-undeclared-contact", "want refusal audit.capability.network-undeclared", []directoryMutation{
			{"envprofile/envprofile.go", `declared = spec.Capabilities`, `declared = spec.Capabilities
 if resolved.Directory == "skills/formatter" && len(declared.Network) == 0 {
 rootSpec, rootErr := skillspec.Load(entry); if rootErr == nil { declared = rootSpec.Capabilities }
 }`},
		}},
		{"markdown-contact-ignored", "envprofile", "^TestSkillDirectoryLifecycleVectors$/^selected-root-undeclared-contact$", "TestSkillDirectoryLifecycleVectors/selected-root-undeclared-contact", "want refusal audit.capability.network-undeclared", []directoryMutation{
			{"audit/audit.go", `if !strings.HasPrefix(posix, "scripts/") && posix != manifestName && !markdown {`, `if !strings.HasPrefix(posix, "scripts/") && posix != manifestName && (!markdown || posix == "references/notes.md") {`},
		}},
		{"alternate-directory-identity", "contextpkg", "^TestSkillDirectoryLockIdentityVectors$/^alternate$", "TestSkillDirectoryLockIdentityVectors/alternate", "canonical or identity mismatch", []directoryMutation{
			{"contextlock/contextlock.go", `object["directory"] = member.Directory`, `object["directory"] = member.Directory
 if member.Kind == KindSkill && member.Directory == "skills/alternate" { object["directory"] = "skills/formatter" }`},
		}},
		{"materialize-repository-root", "envprofile", "^TestSkillDirectoryLifecycleVectors$/^profile-install-selected-root$", "TestSkillDirectoryLifecycleVectors/profile-install-selected-root", "skill target=", []directoryMutation{
			{"envprofile/managed.go", `root := packageRoot(entry, member.Directory)`, `root := packageRoot(entry, member.Directory)
 if member.Name == "formatter" && member.Directory == "skills/formatter" { root = entry }`},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			sources := map[string]string{}
			for _, change := range tc.changes {
				path := filepath.Join(root, "internal", filepath.FromSlash(change.file))
				source, ok := sources[path]
				if !ok {
					b, e := os.ReadFile(path)
					if e != nil {
						t.Fatal(e)
					}
					source = strings.ReplaceAll(string(b), "\r\n", "\n")
				}
				if strings.Count(source, change.before) != 1 {
					t.Fatalf("mutation anchor is not unique: %s", change.file)
				}
				sources[path] = strings.Replace(source, change.before, change.after, 1)
			}
			replacements := map[string]string{}
			for path, source := range sources {
				replacement := filepath.Join(dir, filepath.Base(path))
				if e := os.WriteFile(replacement, []byte(source), 0600); e != nil {
					t.Fatal(e)
				}
				replacements[path] = replacement
			}
			overlay, e := json.Marshal(map[string]any{"Replace": replacements})
			if e != nil {
				t.Fatal(e)
			}
			overlayPath := filepath.Join(dir, "overlay.json")
			if e := os.WriteFile(overlayPath, overlay, 0600); e != nil {
				t.Fatal(e)
			}
			command := exec.Command("go", "test", "-overlay", overlayPath, "-count=1", "-timeout=2m", "-v", "-run", tc.run, "./internal/"+tc.pkg)
			command.Dir = root
			output, e := command.CombinedOutput()
			exit, ok := e.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: "+tc.fails+" ") || !strings.Contains(string(output), tc.message) {
				t.Fatalf("mutant survived or lacked intended behavioral failure: %v\n%s", e, output)
			}
			t.Logf("expected-red go test exit %d:\n%s", exit.ExitCode(), output)
		})
	}
}
