package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/manifest"
)

// The published manager-lifecycle document is the only source of expected
// values in this file. Only field names are mirrored.
type authoritativeBootstrapCase struct {
	Name      string `json:"name"`
	Config    string `json:"config"`
	Force     bool   `json:"force"`
	IfMissing bool   `json:"if_missing"`
	Outcome   string `json:"outcome"`
}

type authoritativeUpgradeCase struct {
	Name        string   `json:"name"`
	Scope       string   `json:"scope"`
	Selection   string   `json:"selection"`
	Fetch       []string `json:"fetch"`
	Exclude     []string `json:"exclude"`
	Deduplicate bool     `json:"deduplicate"`
}

type authoritativeLifecycle struct {
	BootstrapCases []authoritativeBootstrapCase `json:"bootstrap_cases"`
	UpgradeCases   []authoritativeUpgradeCase   `json:"upgrade_cases"`
}

func authoritativeLifecycleDocument(t *testing.T) authoritativeLifecycle {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "manager-lifecycle.json")) // #nosec G304 -- explicit authoritative conformance input
	if err != nil {
		t.Fatal(err)
	}
	var document authoritativeLifecycle
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatal(err)
	}
	return document
}

// invalidConfigBytes is a configuration a bootstrap that claims not to touch an
// existing file cannot even parse, so "unchanged" is provable byte for byte.
const invalidConfigBytes = "this is intentionally not valid JSON\n"

// TestAuthoritativeBootstrapCasesAreExecutable binds every published bootstrap
// case to a real CLI invocation. The published flags build the argument vector
// and the published outcome selects the assertion, so neither can drift from
// the suite without failing here.
func TestAuthoritativeBootstrapCasesAreExecutable(t *testing.T) {
	t.Parallel()
	document := authoritativeLifecycleDocument(t)
	if len(document.BootstrapCases) == 0 {
		t.Fatal("the authoritative suite publishes no bootstrap case to bind")
	}
	for _, published := range document.BootstrapCases {
		published := published
		t.Run(published.Name, func(t *testing.T) {
			// "either" means the outcome may not depend on the configuration, so
			// both configurations are exercised rather than one of them chosen.
			states := []string{published.Config}
			if published.Config == "either" {
				states = []string{"missing", "existing-invalid"}
			}
			for _, state := range states {
				runBootstrapCase(t, published, state)
			}
		})
	}
}

func runBootstrapCase(t *testing.T, published authoritativeBootstrapCase, state string) {
	t.Helper()
	root := t.TempDir()
	configPath := filepath.Join(root, "home", "config.json")
	skillsRoot := filepath.Join(root, "skills")

	switch state {
	case "missing":
	case "existing-invalid":
		if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(configPath, []byte(invalidConfigBytes), 0o600); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatalf("published configuration state %q has no executable binding", state)
	}

	arguments := []string{"bootstrap", "--non-interactive", "--skills-root", skillsRoot}
	if published.IfMissing {
		arguments = append(arguments, "--if-missing")
	}
	if published.Force {
		arguments = append(arguments, "--force")
	}
	code, stdout, stderr := capture(t, configPath, arguments...)

	switch published.Outcome {
	case "created":
		if code != exitOK {
			t.Fatalf("%v = %d\nstdout:\n%s\nstderr:\n%s", arguments, code, stdout, stderr)
		}
		loaded, err := config.Load(configPath, nil)
		if err != nil {
			t.Fatalf("bootstrap reported success but wrote no loadable configuration: %v", err)
		}
		if loaded.SkillsRoot != skillsRoot {
			t.Fatalf("skills_root = %q, want %q", loaded.SkillsRoot, skillsRoot)
		}
	case "unchanged-success":
		if code != exitOK {
			t.Fatalf("%v = %d\nstdout:\n%s\nstderr:\n%s", arguments, code, stdout, stderr)
		}
		payload, err := os.ReadFile(configPath) // #nosec G304 -- test fixture path
		if err != nil {
			t.Fatal(err)
		}
		if string(payload) != invalidConfigBytes {
			t.Fatalf("an existing configuration changed:\n%s", payload)
		}
	case "usage-error":
		if code != exitUsage {
			t.Fatalf("%v = %d, want the usage exit %d\nstdout:\n%s\nstderr:\n%s",
				arguments, code, exitUsage, stdout, stderr)
		}
		assertBootstrapLeftState(t, configPath, state)
	default:
		t.Fatalf("published bootstrap outcome %q has no executable binding", published.Outcome)
	}
}

// assertBootstrapLeftState proves a refused bootstrap changed nothing, in
// whichever configuration state the case started from.
func assertBootstrapLeftState(t *testing.T, configPath, state string) {
	t.Helper()
	payload, err := os.ReadFile(configPath) // #nosec G304 -- test fixture path
	switch state {
	case "missing":
		if !os.IsNotExist(err) {
			t.Fatalf("a refused bootstrap created a configuration: %v", err)
		}
	case "existing-invalid":
		if err != nil {
			t.Fatal(err)
		}
		if string(payload) != invalidConfigBytes {
			t.Fatalf("a refused bootstrap changed the existing configuration:\n%s", payload)
		}
	}
}

// upgradeFixture is one machine with three skill repositories: a directly
// declared skill, the skill it requires transitively, and an unrelated skill no
// published upgrade case may reach.
type upgradeFixture struct {
	home       string
	configPath string
	skillsRoot string
	projects   map[string]string
}

// newUpgradeFixture publishes each skill twice: an upstream repository and a
// working copy in the skills root whose origin points at it, so `upgrade`
// performs a real fetch instead of resolving purely locally.
func newUpgradeFixture(t *testing.T, projectAliases ...string) upgradeFixture {
	t.Helper()
	root := t.TempDir()
	upstream := filepath.Join(root, "upstream")
	fixture := upgradeFixture{
		home:       filepath.Join(root, "home"),
		configPath: filepath.Join(root, "home", "config.json"),
		skillsRoot: filepath.Join(root, "skills"),
		projects:   map[string]string{},
	}

	transitiveWorking := filepath.Join(fixture.skillsRoot, "transitive")
	publishSkillRepo(t, filepath.Join(upstream, "transitive"), "transitive", "")
	publishSkillRepo(t, filepath.Join(upstream, "unrelated"), "unrelated", "")
	publishSkillRepo(t, filepath.Join(upstream, "direct"), "direct", transitiveWorking)
	for _, name := range []string{"transitive", "unrelated", "direct"} {
		cloneSkillRepo(t, filepath.Join(upstream, name), filepath.Join(fixture.skillsRoot, name))
	}

	if err := config.Bootstrap(fixture.configPath, fixture.skillsRoot, "", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	for _, alias := range projectAliases {
		project := filepath.Join(root, "project-"+alias)
		writeFile(t, filepath.Join(project, ".gitignore"), ".agents/\n.codex/skills/\nSkillfile.dev.json\n")
		runGit(t, mkdirFor(t, project), "init", "-q", "-b", "main")
		if err := config.AddProject(fixture.configPath, alias, project, []string{"codex_cli"}); err != nil {
			t.Fatal(err)
		}
		fixture.projects[alias] = project
	}
	return fixture
}

func mkdirFor(t *testing.T, dir string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// publishSkillRepo writes one tagged skill repository. When requires is set the
// skill declares a full-mode requirement on that repository, which is the
// transitive edge a published closure case has to reach.
func publishSkillRepo(t *testing.T, dir, name, requires string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, "SKILL.md"), "---\nname: "+name+"\ndescription: d\n---\n# "+name+"\n")
	spec := map[string]any{
		"schema_version": 4,
		"capabilities":   map[string]any{},
		"commands":       map[string]any{},
	}
	if requires != "" {
		spec["dependencies"] = map[string]any{"skills": map[string]any{
			filepath.Base(requires): map[string]any{
				"git":  requires,
				"ref":  map[string]any{"kind": "tag", "value": "v1"},
				"mode": "full",
			},
		}}
	}
	payload, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "agent-skill.json"), string(payload))
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-qm", "publish "+name)
	runGit(t, dir, "tag", "v1")
}

func cloneSkillRepo(t *testing.T, upstream, working string) {
	t.Helper()
	runGit(t, mkdirFor(t, filepath.Dir(working)), "clone", "-q", upstream, working)
}

func (fixture upgradeFixture) declare(t *testing.T, alias string, names ...string) {
	t.Helper()
	skills := make([]map[string]any, 0, len(names))
	for _, name := range names {
		skills = append(skills, map[string]any{"name": name, "tag": "v1"})
	}
	payload, err := json.MarshalIndent(map[string]any{
		"schema_version": 1,
		"agents":         []string{"codex_cli"},
		"skills":         skills,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(fixture.projects[alias], manifest.Name), string(payload))
}

func (fixture upgradeFixture) declareGlobal(t *testing.T, names ...string) {
	t.Helper()
	if _, err := install.GlobalInit(fixture.home); err != nil {
		t.Fatal(err)
	}
	skills := make([]map[string]any, 0, len(names))
	for _, name := range names {
		skills = append(skills, map[string]any{"name": name, "tag": "v1"})
	}
	payload, err := json.MarshalIndent(map[string]any{
		"schema_version": 1,
		"agents":         []string{"codex_cli"},
		"skills":         skills,
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(install.GlobalRoot(fixture.home), manifest.Name), string(payload))
}

// TestAuthoritativeUpgradeCasesAreExecutable binds every published upgrade case
// to the CLI selection it names, then proves the published closure was fetched,
// the published exclusion was never reached, and — where the case requires it —
// a repository shared by two selected scopes is fetched exactly once.
func TestAuthoritativeUpgradeCasesAreExecutable(t *testing.T) {
	t.Parallel()
	document := authoritativeLifecycleDocument(t)
	if len(document.UpgradeCases) == 0 {
		t.Fatal("the authoritative suite publishes no upgrade case to bind")
	}
	for _, published := range document.UpgradeCases {
		published := published
		t.Run(published.Name, func(t *testing.T) {
			switch {
			case published.Scope == "project" && published.Selection == "one":
				runSelectedProjectUpgrade(t, published)
			case published.Scope == "project" && published.Selection == "all":
				runAllProjectsUpgrade(t, published)
			case published.Scope == "global" && published.Selection == "global":
				runGlobalUpgrade(t, published)
			default:
				t.Fatalf("published upgrade case %q (scope %q, selection %q) has no executable binding",
					published.Name, published.Scope, published.Selection)
			}
		})
	}
}

func runSelectedProjectUpgrade(t *testing.T, published authoritativeUpgradeCase) {
	t.Helper()
	fixture := newUpgradeFixture(t, "app", "other")
	fixture.declare(t, "app", "direct")
	fixture.declare(t, "other", "unrelated")

	code, stdout, stderr := capture(t, fixture.configPath, "upgrade", "app")
	if code != exitOK {
		t.Fatalf("upgrade app = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	installed := filepath.Join(fixture.projects["app"], ".agents", "skills")
	assertClosureFetched(t, published, installed, stdout)
	assertExclusionsUntouched(t, published, fixture, installed, stdout)
}

func runAllProjectsUpgrade(t *testing.T, published authoritativeUpgradeCase) {
	t.Helper()
	fixture := newUpgradeFixture(t, "app", "other")
	// Both selected projects declare the same skill, so the shared repository
	// is the one a deduplicating upgrade may fetch only once.
	fixture.declare(t, "app", "direct")
	fixture.declare(t, "other", "direct")

	code, stdout, stderr := capture(t, fixture.configPath, "upgrade", "--all")
	if code != exitOK {
		t.Fatalf("upgrade --all = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	for _, alias := range []string{"app", "other"} {
		if _, err := os.Stat(filepath.Join(fixture.projects[alias], ".agents", "skills", "direct")); err != nil {
			t.Fatalf("project %s was not installed by the all-projects selection: %v", alias, err)
		}
	}
	if published.Deduplicate {
		for _, repository := range []string{"direct", "transitive"} {
			if fetches := countFetches(stdout, repository); fetches != 1 {
				t.Fatalf("repository %q was fetched %d times across the selection, want exactly 1\n%s",
					repository, fetches, stdout)
			}
		}
	}
}

func runGlobalUpgrade(t *testing.T, published authoritativeUpgradeCase) {
	t.Helper()
	fixture := newUpgradeFixture(t)
	fixture.declareGlobal(t, "direct")

	code, stdout, stderr := capture(t, fixture.configPath, "global", "upgrade")
	if code != exitOK {
		t.Fatalf("global upgrade = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	installed := filepath.Join(fixture.home, "global", "skills")
	assertClosureFetched(t, published, installed, stdout)
	assertExclusionsUntouched(t, published, fixture, installed, stdout)
}

// assertClosureFetched maps each published closure member to the repository it
// names in this fixture and proves the selection actually reached it.
func assertClosureFetched(t *testing.T, published authoritativeUpgradeCase, installed, stdout string) {
	t.Helper()
	if len(published.Fetch) == 0 {
		t.Fatalf("published upgrade case %q names no closure to fetch", published.Name)
	}
	for _, member := range published.Fetch {
		repository := ""
		switch member {
		case "direct":
			repository = "direct"
		case "transitive":
			repository = "transitive"
		default:
			t.Fatalf("published closure member %q has no executable binding", member)
		}
		if _, err := os.Stat(filepath.Join(installed, repository)); err != nil {
			t.Fatalf("the %s closure member was not installed: %v\n%s", member, err, stdout)
		}
		if countFetches(stdout, repository) == 0 {
			t.Fatalf("the %s closure member was never fetched:\n%s", member, stdout)
		}
	}
}

// assertExclusionsUntouched proves the published exclusion was neither fetched
// nor installed by the selection under test.
func assertExclusionsUntouched(t *testing.T, published authoritativeUpgradeCase, fixture upgradeFixture, installed, stdout string) {
	t.Helper()
	for _, excluded := range published.Exclude {
		if excluded != "unrelated" {
			t.Fatalf("published exclusion %q has no executable binding", excluded)
		}
		if _, err := os.Stat(filepath.Join(installed, excluded)); !os.IsNotExist(err) {
			t.Fatalf("the excluded skill %q was installed by this selection: %v", excluded, err)
		}
		if countFetches(stdout, excluded) != 0 {
			t.Fatalf("the excluded skill %q was fetched by this selection:\n%s", excluded, stdout)
		}
		if _, err := os.Stat(filepath.Join(fixture.skillsRoot, excluded, ".git", "FETCH_HEAD")); !os.IsNotExist(err) {
			t.Fatalf("the excluded repository %q was contacted by this selection: %v", excluded, err)
		}
	}
}

// countFetches counts the manager's own "fetched <repository>" report lines,
// which is how a scope reports the repositories it actually contacted.
func countFetches(stdout, repository string) int {
	count := 0
	for _, line := range strings.Split(stdout, "\n") {
		if strings.HasSuffix(strings.TrimSpace(line), "fetched "+repository) {
			count++
		}
	}
	return count
}
