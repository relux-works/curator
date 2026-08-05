package buildrepo_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/skillspec"
)

const guideRelativePath = "../../docs/external-build-repositories.md"

func readGuide(t *testing.T) string {
	t.Helper()
	payload, err := os.ReadFile(guideRelativePath)
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}

func guideExamples(t *testing.T, guide string) map[string][]string {
	t.Helper()
	pattern := regexp.MustCompile(`(?s)<!-- example:([a-z-]+) -->\s*` + "```json\\n" + `(.*?)\n` + "```")
	examples := map[string][]string{}
	for _, match := range pattern.FindAllStringSubmatch(guide, -1) {
		examples[match[1]] = append(examples[match[1]], match[2])
	}
	return examples
}

func TestExternalRepositoryGuideExamplesValidate(t *testing.T) {
	examples := guideExamples(t, readGuide(t))
	if len(examples["manifest"]) != 1 || len(examples["descriptor"]) != 1 || len(examples["substitution"]) != 2 {
		t.Fatalf("guide example inventory = %#v", examples)
	}

	for _, manifest := range examples["manifest"] {
		directory := t.TempDir()
		if err := os.WriteFile(filepath.Join(directory, skillspec.CanonicalManifestName), []byte(manifest), 0o600); err != nil {
			t.Fatal(err)
		}
		spec, err := skillspec.Load(directory)
		if err != nil {
			t.Fatalf("schema-7 manifest example is invalid: %v", err)
		}
		if spec.SchemaVersion != 7 {
			t.Fatalf("manifest schema = %d, want 7", spec.SchemaVersion)
		}
	}

	for _, descriptor := range examples["descriptor"] {
		if _, err := buildrepo.ParseDescriptor([]byte(descriptor)); err != nil {
			t.Fatalf("skill-build.json example is invalid: %v", err)
		}
	}

	for _, substitution := range examples["substitution"] {
		if _, err := devsub.ParseManifestBytes([]byte(substitution), t.TempDir()); err != nil {
			t.Fatalf("Skillfile.dev.json example is invalid: %v", err)
		}
	}
}

func TestExternalRepositoryGuideFutureDriverAdmissionTable(t *testing.T) {
	guide := readGuide(t)
	start := strings.Index(guide, "<!-- future-driver-table:start -->")
	end := strings.Index(guide, "<!-- future-driver-table:end -->")
	if start < 0 || end <= start {
		t.Fatal("future-driver admission table markers are missing")
	}
	lines := strings.Split(guide[start:end], "\n")
	rows := map[string][]string{}
	for _, line := range lines {
		if !strings.HasPrefix(line, "| ") || strings.HasPrefix(line, "| ---") || strings.HasPrefix(line, "| Language ") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "| "), " | ")
		if len(cells) != 7 {
			t.Fatalf("future-driver row has %d cells, want 7: %s", len(cells), line)
		}
		rows[cells[0]] = cells
	}

	for _, language := range []string{"Rust", "Swift", "Kotlin/JVM", "C/C++", ".NET"} {
		cells, ok := rows[language]
		if !ok {
			t.Errorf("future-driver table is missing %s", language)
			continue
		}
		if !strings.Contains(cells[1], "Unsupported") || !strings.Contains(cells[1], "separately versioned closed driver") {
			t.Errorf("%s admission does not remain closed and unsupported: %q", language, cells[1])
		}
		for column, cell := range cells[2:] {
			if strings.TrimSpace(cell) == "" {
				t.Errorf("%s threat-review column %d is empty", language, column+2)
			}
		}
	}
	if len(rows) != 5 {
		t.Fatalf("future-driver table has %d language rows, want 5", len(rows))
	}
}

func TestExternalRepositoryGuideLocalLinksResolve(t *testing.T) {
	guide := readGuide(t)
	linkPattern := regexp.MustCompile(`\[[^]]+\]\(([^)]+)\)`)
	guideDirectory := filepath.Dir(guideRelativePath)
	for _, match := range linkPattern.FindAllStringSubmatch(guide, -1) {
		target := strings.SplitN(match[1], "#", 2)[0]
		if target == "" || strings.Contains(target, "://") {
			continue
		}
		if _, err := os.Stat(filepath.Clean(filepath.Join(guideDirectory, filepath.FromSlash(target)))); err != nil {
			t.Errorf("guide link %q does not resolve: %v", match[1], err)
		}
	}
}
