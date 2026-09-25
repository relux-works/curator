package crossconformance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/sourcelock"
)

// Selection semantic rows (Skillfile §2): unknown aliases, collection
// membership errors, name conflicts, and frozen membership.

func init() {
	registerDraftSemantic("unknown-alias", driveUnknownAlias)
	registerDraftSemantic("missing-excluded-literal", driveMissingExcludedLiteral)
	registerDraftSemantic("bad-wildcard-member", driveBadWildcardMember)
	registerDraftSemantic("duplicate-name", driveDuplicateName)
	registerDraftSemantic("frozen-membership", driveFrozenMembership)
}

func driveUnknownAlias(t *testing.T, _ draftSemanticCase) {
	project := t.TempDir()
	payload := `{"schema_version":2,"skills":[{"name":"review","from":"absent","directory":"."}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := manifest.Load(project)
	if err == nil || !strings.Contains(err.Error(), "source_alias_unknown") {
		t.Fatalf("err = %v, want source_alias_unknown", err)
	}
}

func draftCollectionProject(t *testing.T, payload string, setup func(collections string)) string {
	t.Helper()
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	collections := filepath.Join(project, "collections")
	if err := os.MkdirAll(collections, 0o755); err != nil {
		t.Fatal(err)
	}
	setup(collections)
	return project
}

func expandDraft(t *testing.T, project string) ([]manifest.Selection, error) {
	t.Helper()
	m, err := manifest.Load(project)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return manifest.Expand(m, manifest.ExpansionOptions{})
}

func driveMissingExcludedLiteral(t *testing.T, _ draftSemanticCase) {
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"collections","include":["missing"],"exclude":["missing"]}]}`
	project := draftCollectionProject(t, payload, func(string) {})
	_, err := expandDraft(t, project)
	if err == nil || !strings.Contains(err.Error(), "source_member_missing") {
		t.Fatalf("err = %v, want source_member_missing", err)
	}
}

func driveBadWildcardMember(t *testing.T, _ draftSemanticCase) {
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"collections","include":["*"]}]}`
	project := draftCollectionProject(t, payload, func(collections string) {
		writeDraftSkill(t, filepath.Join(collections, "review"), "review")
		if err := os.MkdirAll(filepath.Join(collections, "docs"), 0o755); err != nil {
			t.Fatal(err)
		}
	})
	_, err := expandDraft(t, project)
	if err == nil || !strings.Contains(err.Error(), "source_member_invalid") {
		t.Fatalf("err = %v, want source_member_invalid", err)
	}
}

func driveDuplicateName(t *testing.T, _ draftSemanticCase) {
	local, err := sourcelock.LocalPackage("sha256:" + strings.Repeat("11", 32))
	if err != nil {
		t.Fatal(err)
	}
	members := []sourcelock.Member{
		{Name: "review", Directory: "a", Package: local, ContentSHA256: "sha256:" + strings.Repeat("22", 32)},
		{Name: "review", Directory: "b", Package: local, ContentSHA256: "sha256:" + strings.Repeat("33", 32)},
	}
	if _, err := sourcelock.New("sha256:"+strings.Repeat("44", 32), members); err == nil || !strings.Contains(err.Error(), "source_name_conflict") {
		t.Fatalf("New err = %v, want source_name_conflict", err)
	}
	member := func(name, dir, content string, selection int) map[string]any {
		return map[string]any{
			"name": name, "selection": selection, "directory": dir,
			"package":        map[string]any{"kind": "local-snapshot", "snapshot": "sha256:" + strings.Repeat("11", 32)},
			"content_sha256": content,
		}
	}
	doc := map[string]any{
		"schema_version":  1,
		"manifest_sha256": "sha256:" + strings.Repeat("44", 32),
		"members": []any{
			member("review", "a", "sha256:"+strings.Repeat("22", 32), 0),
			member("review", "b", "sha256:"+strings.Repeat("33", 32), 1),
		},
		"lock_sha256": "sha256:" + strings.Repeat("55", 32),
	}
	payload, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sourcelock.Parse(payload); err == nil || !strings.Contains(err.Error(), "source_name_conflict") {
		t.Fatalf("Parse err = %v, want source_name_conflict", err)
	}
}

// driveFrozenMembership proves launch/install consumes the locked
// membership only: the lock binds review, the live tree grows docs,
// and the install still publishes exactly review.
func driveFrozenMembership(t *testing.T, _ draftSemanticCase) {
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project, home := draftProject(t, payload, map[string]string{"skills/review": "review"})
	resolveDraftPlan(t, project, home, payload)
	writeDraftSkill(t, filepath.Join(project, "skills", "docs"), "docs")
	result := draftInstall(t, project, home, install.Options{})
	if result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	skillsDir := filepath.Join(project, ".agents", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if len(names) != 1 || names[0] != "review" {
		t.Fatalf("installed = %v, want [review]", names)
	}
}
