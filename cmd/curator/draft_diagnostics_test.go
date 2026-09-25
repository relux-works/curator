package main

// Draft source workflow diagnostics at the CLI production entry.
//
// These tests drive the documented local/absolute/Git/collection
// examples, the machine policy setup, and every stable source/transport
// remediation through run(), never through the helpers under test. None
// of these tests run in parallel: they mutate process environment per
// run. Documented payloads below are quoted verbatim from docs/cli.md
// (placeholders substituted); TestDraftDocsPinExamples fails if the
// docs drift from the exercised shapes.

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/sourcelock"
)

// Documented acquisition shapes (docs/cli.md, README.md). Absolute and
// Git placeholders are substituted with fixture values before the run.
const (
	docLocalRelativePayload = `{"schema_version":2,"sources":{"local":{"path":"./pkgs"}},"skills":[{"name":"review","from":"local","directory":"review"}]}`
	docAbsolutePathShape    = `{"path": "/work/shared-agents"}`
	docGitTagShape          = `{"git": "https://example.org/kit.git", "tag": "v1.2.0"}`
	docLogicalBranchShape   = `{"repository": "example.org/kit", "branch": "main"}`
	docCollectionListShape  = `{"from": "team", "directory": "skills", "include": ["review", "docs"]}`
	docCollectionStarShape  = `{"from": "team", "directory": "skills", "include": ["*"]}`
	docCollectionExclude    = `{"from": "team", "directory": "skills", "include": ["*"], "exclude": ["release"]}`
)

func TestWithDraftRemediationTable(t *testing.T) {
	rows := []struct {
		name    string
		message string
		want    string
	}{
		{"alias", "source_alias_unknown: ghost", "declare the alias"},
		{"selection", "sources.s: source_selection_invalid: bad directory", "fix the named selector"},
		{"missing", "source_member_missing: ghost has no frozen tree", "add the named member directory"},
		{"invalid", "source_member_invalid: review: bad frontmatter", "fix the named package"},
		{"conflict", "source_name_conflict: review twice", "exactly one selection"},
		{"overlap", "source_output_overlap: .agents/skills", "move the authored package out of managed output"},
		{"changed", "source_snapshot_changed: stored snapshot diverged", "restore the declared source to the package identity and content_sha256 in Skillfile.lock.json"},
		{"unavailable", "source_snapshot_unavailable: declared path source cannot be reached", "restore access to the declared path or Git source, then retry install"},
		{"stale", "source_lock_stale: manifest differs", "run: curator project refresh"},
		{"endpoint", "repository_endpoint_unavailable: identity x: down", "verify the network path and operator authentication"},
		{"policy", "repository_policy_invalid: bad schema", "fix machine source-policy.json"},
		{"mirror", "repository_mirror_undeclared: host differs", `"mirror_of" equal to the entry key`},
		{"alias", "repository_alias_unknown: ghost", `"aliases" table of machine source-policy.json`},
		{"auditRejected", "review: source_audit_rejected: evidence mismatch", "re-resolve under trusted machine policy"},
		{"auditUnavailable", "review: source_audit_unavailable: no binding", "run the explicit attempt under trusted machine policy"},
		{"identityInvalid", "review.etool: build_repository_identity_invalid: transport plan endpoint 1 carries an explicit port or host alias outside the strict external-build lane grammar", "fix the endpoint entry in machine source-policy.json"},
	}
	for _, row := range rows {
		got := withDraftRemediation(row.message)
		if !strings.Contains(got, row.want) {
			t.Errorf("%s: withDraftRemediation(%q) = %q, want %q", row.name, row.message, got, row.want)
		}
		if !strings.HasPrefix(got, row.message) {
			t.Errorf("%s: remediation is not appended: %q", row.name, got)
		}
		// Idempotent: a second pass appends nothing.
		if again := withDraftRemediation(got); again != got {
			t.Errorf("%s: second pass changed %q to %q", row.name, got, again)
		}
	}
	already := []string{
		"source_snapshot_unavailable: no Skillfile lock; run explicit resolve first",
		"source_lock_stale: Skillfile changed since lock; explicit refresh required",
		"source_snapshot_changed: admitted membership changed during capture; retry the explicit attempt",
		"repository_endpoint_unavailable: identity x: down; verify the network path and operator authentication for the listed endpoints, then retry with machine source-policy.json",
	}
	for _, message := range already {
		if got := withDraftRemediation(message); got != message {
			t.Errorf("already-guided message changed:\n got %q\nwant %q", got, message)
		}
	}
	legacy := []string{
		"Skillfile.json not found",
		"up-to-date",
		"build_repository_source_unavailable: exact external source is unavailable",
		"source is not a portable identifier",
		// stbg4d pin: frozen v1 messages are byte-identical — the §7
		// remediation row must not fire on any other identity
		// diagnostic, synthetic or lane-produced.
		"build_repository_identity_invalid: wrong identity",
		"skill-a.ssh-cmd: build_repository_identity_invalid: SSH requires the exact manager wrapper",
	}
	for _, message := range legacy {
		if got := withDraftRemediation(message); got != message {
			t.Errorf("legacy message changed:\n got %q\nwant %q", got, message)
		}
	}
}

func TestDraftAttemptTransportTable(t *testing.T) {
	rows := []struct {
		url  string
		want string
	}{
		{"https://example.org/kit.git", "https"},
		{"https://example.org:8443/kit.git", "https"},
		{"http://example.org/kit.git", "http"},
		{"ssh://git@example.org/kit.git", "ssh"},
		{"ssh://git@example.org:2222/kit.git", "ssh"},
		{"git@example.org:kit.git", "ssh"},
		{"file:///tmp/kit.git", "file"},
		{"git://example.org/kit.git", "git"},
		{"example.org/kit", "unknown"},
		{"", "unknown"},
	}
	for _, row := range rows {
		if got := draftAttemptTransport(row.url); got != row.want {
			t.Errorf("draftAttemptTransport(%q) = %q, want %q", row.url, got, row.want)
		}
	}
}

func TestDraftAttemptClauseShape(t *testing.T) {
	clause := draftAttemptClause(1,
		config.Attempt{URL: "https://example.org/kit.git", Authentication: "team-https"},
		buildrepo.FailureTLS)
	if !strings.Contains(clause, `endpoint 1 (https, provider "team-https"): tls: TLS certificate validation failed`) {
		t.Fatalf("clause = %q", clause)
	}
	if strings.Contains(clause, "example.org") {
		t.Fatalf("clause leaks the endpoint URL: %q", clause)
	}
	bare := draftAttemptClause(2, config.Attempt{URL: "git@example.org:kit.git"}, buildrepo.FailureAuth)
	if !strings.Contains(bare, "endpoint 2 (ssh): auth: endpoint authentication unavailable or rejected") {
		t.Fatalf("bare clause = %q", bare)
	}
	if strings.Contains(bare, "example.org") || strings.Contains(bare, "provider") {
		t.Fatalf("bare clause leaks endpoint detail: %q", bare)
	}
}

func TestEndpointExhaustionCarriesRemediation(t *testing.T) {
	resolution := config.Resolution{
		Identity: "fixture.test/kit",
		Attempts: []config.Attempt{
			{URL: "https://fixture.test/kit.git", Authentication: "team-https"},
			{URL: "ssh://git@fixture.test/kit.git"},
		},
		Fallback: config.FallbackAvailabilityAuth,
	}
	exhausted := endpointExhaustion(resolution, []buildrepo.FailureClass{buildrepo.FailureAuth})
	for _, want := range []string{
		"identity fixture.test/kit",
		`endpoint 1 (https, provider "team-https"): auth: endpoint authentication unavailable or rejected`,
		"endpoint 2: not attempted",
		draftEndpointRemediation,
	} {
		if !strings.Contains(exhausted, want) {
			t.Errorf("endpointExhaustion misses %q:\n%s", want, exhausted)
		}
	}
	if strings.Contains(exhausted, "fixture.test/kit.git") {
		t.Errorf("endpointExhaustion leaks the endpoint URL:\n%s", exhausted)
	}
	fetched := fetchExhaustion(resolution, buildrepo.FailureAvailability)
	for _, want := range []string{
		"identity fixture.test/kit",
		"fetch of the existing checkout failed (availability: endpoint unavailable (DNS, connection, or HTTP 502/503/504 failure))",
		draftEndpointRemediation,
	} {
		if !strings.Contains(fetched, want) {
			t.Errorf("fetchExhaustion misses %q:\n%s", want, fetched)
		}
	}
}

// TestDraftProjectResolveHelp pins the project command help at the CLI entry.
// The flag spellings print the schema-2 workflow; the bare word "help"
// continues to resolve as a project alias.
func TestDraftProjectResolveHelp(t *testing.T) {
	root := t.TempDir()
	configPath, _ := setupCLIProject(t, root)
	helpDir := filepath.Join(root, "help")
	if err := os.MkdirAll(helpDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if code := runCode(t, configPath, []string{"project", "add", "help", helpDir, "--agents", "codex_cli"}); code != exitOK {
		t.Fatalf("project add help = %d", code)
	}
	if err := os.WriteFile(filepath.Join(helpDir, "Skillfile.json"), []byte(`{"schema_version":1,"agents":["codex_cli"],"skills":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	wantReport := "alias: help\npath: " + helpDir +
		"\nskillfile: " + filepath.Join(helpDir, "Skillfile.json") +
		"\nskills: " + filepath.Join(helpDir, ".agents", "skills") +
		"\nbin: " + filepath.Join(helpDir, ".agents", "bin") + "\n"

	for _, verb := range []string{"resolve", "refresh"} {
		for _, spelling := range []string{"-h", "--help"} {
			code, stdout, stderr := capture(t, configPath, "project", verb, spelling)
			if code != exitOK || stdout != projectResolveUsage || stderr != "" {
				t.Fatalf("project %s %s = code %d, stdout %d bytes, stderr %q; want the help golden on stdout", verb, spelling, code, len(stdout), stderr)
			}
		}
		code, stdout, stderr := capture(t, configPath, "project", verb, "help")
		if code != exitOK || stdout != wantReport || stderr != "" {
			t.Fatalf("project %s help = code %d, stdout %q, stderr %q; want the v1 report for alias help", verb, code, stdout, stderr)
		}
	}
	for _, marker := range []string{
		"Skillfile schema 2 project sources",
		"Skillfile schema 1 retains its exact meaning; no on-disk migration is implicit",
		"project resolve", "project refresh", "curator install", "curator status",
		`{"path": "./agents"}`, `{"path": "/work/shared-agents"}`,
		`{"git": "https://example.org/kit.git", "tag": "v1.2.0"}`,
		`{"repository": "example.org/kit", "branch": "main"}`,
		`{"from": "team", "directory": "skills", "include": ["*"]}`,
		"source-policy.json", `"root_inputs"`, "never rescan",
	} {
		if !strings.Contains(projectResolveUsage, marker) {
			t.Errorf("help golden misses %q", marker)
		}
	}
	for _, stale := range []string{"deprecated opt-out", "unreleased"} {
		if strings.Contains(projectResolveUsage, stale) {
			t.Errorf("help golden retains stale switch wording %q", stale)
		}
	}
}

func TestDraftInstallStatusHelpIncludesSchema2Workflow(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	for _, command := range []string{"install", "status"} {
		code, _, stderr := capture(t, configPath, command, "-h")
		if code != exitUsage {
			t.Fatalf("%s -h = %d, want %d", command, code, exitUsage)
		}
		if !strings.Contains(stderr, "Usage of "+command+":") || !strings.Contains(stderr, draftWorkflowSection) {
			t.Fatalf("%s -h misses usage or schema-2 workflow help:\n%s", command, stderr)
		}
	}
}

// writeDocPkgs lays out the documented local fixture: pkgs/review for
// the individual rows and pkgs/skills/{review,docs,release} for the
// collection rows (the exclude row drops release).
func writeDocPkgs(t *testing.T, pkgs string) {
	t.Helper()
	writeCLISkill(t, filepath.Join(pkgs, "review"), "review")
	writeCLISkill(t, filepath.Join(pkgs, "skills", "review"), "review")
	writeCLISkill(t, filepath.Join(pkgs, "skills", "docs"), "docs")
	writeCLISkill(t, filepath.Join(pkgs, "skills", "release"), "release")
}

// TestDraftDocumentedLocalShapesThroughCLI runs every documented local
// acquisition shape through the real verbs: resolve, install --dry-run,
// install, and status. The relative row is docs/cli.md verbatim.
func TestDraftDocumentedLocalShapesThroughCLI(t *testing.T) {
	rows := []struct {
		name    string
		payload func(project, pkgs string) string
		members []string
	}{
		{
			name:    "relative-individual",
			payload: func(string, string) string { return docLocalRelativePayload },
			members: []string{"review"},
		},
		{
			name: "absolute-individual",
			payload: func(_, pkgs string) string {
				return `{"schema_version":2,"sources":{"shared":{"path":` + quotePath(pkgs) + `}},"skills":[{"name":"review","from":"shared","directory":"review"}]}`
			},
			members: []string{"review"},
		},
		{
			name: "collection-list",
			payload: func(string, string) string {
				return `{"schema_version":2,"sources":{"team":{"path":"./pkgs"}},"skills":[` + docCollectionListShape + `]}`
			},
			members: []string{"docs", "review"},
		},
		{
			name: "collection-star",
			payload: func(string, string) string {
				return `{"schema_version":2,"sources":{"team":{"path":"./pkgs"}},"skills":[` + docCollectionStarShape + `]}`
			},
			members: []string{"docs", "release", "review"},
		},
		{
			name: "collection-exclude",
			payload: func(string, string) string {
				return `{"schema_version":2,"sources":{"team":{"path":"./pkgs"}},"skills":[` + docCollectionExclude + `]}`
			},
			members: []string{"docs", "review"},
		},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			root := t.TempDir()
			configPath, project := setupCLIProject(t, root)
			t.Setenv("CURATOR_CONFIG", configPath)
			pkgs := filepath.Join(project, "pkgs")
			writeDocPkgs(t, pkgs)
			payload := row.payload(project, pkgs)
			if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
				t.Fatal(err)
			}
			if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
				t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
			}
			lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
			if err != nil {
				t.Fatal(err)
			}
			if len(lock.Members) != len(row.members) {
				t.Fatalf("lock members = %d, want %d (%v)", len(lock.Members), len(row.members), row.members)
			}
			for _, want := range row.members {
				if _, ok := lock.Find(want); !ok {
					t.Fatalf("lock misses %s: %+v", want, lock.Members)
				}
			}
			if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitOK {
				t.Fatalf("install --dry-run = %d\nstderr:\n%s", code, stderr)
			}
			if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
				t.Fatalf("install = %d\nstderr:\n%s", code, stderr)
			}
			code, stdout, stderr := capture(t, configPath, "status", "app")
			if code != exitOK {
				t.Fatalf("status = %d\nstderr:\n%s", code, stderr)
			}
			for _, want := range row.members {
				if !strings.Contains(stdout, "app: "+want+" up-to-date") {
					t.Fatalf("status misses up-to-date %s:\n%s", want, stdout)
				}
			}
		})
	}
}

// quotePath renders one native path as a JSON string value.
func quotePath(path string) string {
	escaped := strings.ReplaceAll(path, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}

// setupRemediationProject bootstraps one draft project with pkgs/review
// present and returns its config path, project root, and home. It pins
// CURATOR_CONFIG at the test config so the explicit attempt loads
// machine source-policy.json from the test home, never the operator's.
func setupRemediationProject(t *testing.T) (configPath, project, home string) {
	t.Helper()
	root := t.TempDir()
	configPath, project = setupCLIProject(t, root)
	t.Setenv("CURATOR_CONFIG", configPath)
	writeCLISkill(t, filepath.Join(project, "pkgs", "review"), "review")
	return configPath, project, filepath.Dir(configPath)
}

func writeRemediationManifest(t *testing.T, project, payload string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestDraftRemediationThroughCLI fails one CLI verb per stable source or
// transport class and requires the class plus its sanitized remediation
// on stderr. Classes whose underlying message already guides the
// operator assert the single guidance instead of a duplicate.
func TestDraftRemediationThroughCLI(t *testing.T) {
	rows := []struct {
		name    string
		payload string
		verb    []string
		class   string
		remedy  string
		absent  string
		setup   func(t *testing.T, configPath, project, home string)
	}{
		{
			name:    "alias-unknown",
			payload: `{"schema_version":2,"sources":{"s":{"path":"./pkgs"}},"skills":[{"name":"review","from":"ghost","directory":"review"}]}`,
			verb:    []string{"project", "resolve", "app"},
			class:   "source_alias_unknown",
			remedy:  `declare the alias under "sources"`,
		},
		{
			name:    "selection-invalid",
			payload: `{"schema_version":2,"sources":{"s":{"path":"./pkgs"}},"skills":[{"name":"review","from":"s","directory":"/absolute"}]}`,
			verb:    []string{"project", "resolve", "app"},
			class:   "source_selection_invalid",
			remedy:  "fix the named selector",
		},
		{
			name:    "member-missing",
			payload: `{"schema_version":2,"sources":{"s":{"path":"./pkgs"}},"skills":[{"from":"s","directory":".","include":["ghost"]}]}`,
			verb:    []string{"project", "resolve", "app"},
			class:   "source_member_missing",
			remedy:  "add the named member directory",
		},
		{
			name:    "member-invalid",
			payload: `{"schema_version":2,"sources":{"s":{"path":"./pkgs"}},"skills":[{"from":"s","directory":".","include":["broken"]}]}`,
			verb:    []string{"project", "resolve", "app"},
			class:   "source_member_invalid",
			remedy:  "fix the named package",
			setup: func(t *testing.T, _, project, _ string) {
				broken := filepath.Join(project, "pkgs", "broken")
				if err := os.MkdirAll(broken, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(broken, "SKILL.md"), []byte("# no frontmatter\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:    "name-conflict",
			payload: `{"schema_version":2,"sources":{"s":{"path":"./pkgs"}},"skills":[{"name":"review","from":"s","directory":"review"},{"name":"review","from":"s","directory":"review"}]}`,
			verb:    []string{"project", "resolve", "app"},
			class:   "source_name_conflict",
			remedy:  "exactly one selection",
		},
		{
			name:    "output-overlap",
			payload: `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"name":"managed","from":"s","directory":".agents/skills"}]}`,
			verb:    []string{"project", "resolve", "app"},
			class:   "source_output_overlap",
			remedy:  "move the authored package out of managed output",
			setup: func(t *testing.T, _, project, _ string) {
				if err := os.MkdirAll(filepath.Join(project, ".agents", "skills"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:    "snapshot-changed",
			payload: `{"schema_version":2,"sources":{"s":{"path":"./pkgs"}},"skills":[{"name":"review","from":"s","directory":"review"}]}`,
			verb:    []string{"install", "app", "--dry-run"},
			class:   "source_snapshot_changed",
			remedy:  "restore the declared source to the package identity and content_sha256 in Skillfile.lock.json",
			setup: func(t *testing.T, configPath, project, home string) {
				if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
					t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
				}
				lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
				if err != nil {
					t.Fatal(err)
				}
				member, ok := lock.Find("review")
				if !ok {
					t.Fatalf("lock misses review: %+v", lock.Members)
				}
				digest := strings.TrimPrefix(member.Package.Snapshot, "sha256:")
				store := filepath.Join(home, "local-snapshots", digest, "snapshot", "references", "info.md")
				if err := os.WriteFile(store, []byte("tampered"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:    "snapshot-unavailable",
			payload: `{"schema_version":2,"sources":{"s":{"path":"./pkgs"}},"skills":[{"name":"review","from":"s","directory":"review"}]}`,
			verb:    []string{"install", "app", "--dry-run"},
			class:   "source_snapshot_unavailable",
			remedy:  "restore access to the declared path or Git source, then retry install",
			setup: func(t *testing.T, configPath, project, home string) {
				if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
					t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
				}
				if err := os.RemoveAll(filepath.Join(project, "pkgs")); err != nil {
					t.Fatal(err)
				}
				if err := os.RemoveAll(filepath.Join(home, "local-snapshots")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:    "lock-stale",
			payload: `{"schema_version":2,"sources":{"s":{"path":"./pkgs"}},"skills":[{"name":"review","from":"s","directory":"review"}]}`,
			verb:    []string{"install", "app", "--dry-run"},
			class:   "source_lock_stale",
			remedy:  "explicit refresh required",
			absent:  "changed since the lock; run",
			setup: func(t *testing.T, configPath, project, _ string) {
				if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
					t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
				}
				stale := `{"schema_version":2,"sources":{"s":{"path":"./pkgs"}},"skills":[{"name":"review","from":"s","directory":"review"},{"name":"docs","from":"s","directory":"review"}]}`
				writeRemediationManifest(t, project, stale)
			},
		},
		{
			name:    "endpoint-unavailable",
			payload: `{"schema_version":2,"sources":{"kit":{"repository":"fixture.test/kit","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`,
			verb:    []string{"project", "resolve", "app"},
			class:   "repository_endpoint_unavailable",
			remedy:  "retry with machine source-policy.json",
		},
		{
			name:    "policy-invalid",
			payload: `{"schema_version":2,"sources":{"kit":{"git":"https://fixture.test/kit.git","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`,
			verb:    []string{"project", "resolve", "app"},
			class:   "repository_policy_invalid",
			remedy:  "fix machine source-policy.json",
			setup: func(t *testing.T, _, _, home string) {
				if err := os.WriteFile(filepath.Join(home, "source-policy.json"), []byte(`{"schema_version":1`), 0o644); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:    "mirror-undeclared",
			payload: `{"schema_version":2,"sources":{"kit":{"git":"https://fixture.test/kit.git","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`,
			verb:    []string{"project", "resolve", "app"},
			class:   "repository_mirror_undeclared",
			remedy:  `"mirror_of" equal to the entry key`,
			setup: func(t *testing.T, _, _, home string) {
				policy := `{"schema_version":2,"repositories":{"fixture.test/kit":{"endpoints":[{"url":"https://mirror.test/kit.git","authentication":"team-https"}],"fallback":"none"}}}`
				if err := os.WriteFile(filepath.Join(home, "source-policy.json"), []byte(policy), 0o644); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:    "repository-alias-unknown",
			payload: `{"schema_version":2,"sources":{"kit":{"git":"https://fixture.test/kit.git","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`,
			verb:    []string{"project", "resolve", "app"},
			class:   "repository_alias_unknown",
			remedy:  `"aliases" table of machine source-policy.json`,
			setup: func(t *testing.T, _, _, home string) {
				policy := `{"schema_version":2,"repositories":{"fixture.test/kit":{"endpoints":[{"url":"https://fixture.test/kit.git","authentication":"team-https","alias":"ghost"}],"fallback":"none"}}}`
				if err := os.WriteFile(filepath.Join(home, "source-policy.json"), []byte(policy), 0o644); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:    "audit-unavailable",
			payload: `{"schema_version":2,"sources":{"s":{"path":"./pkgs"}},"skills":[{"name":"review","from":"s","directory":"review"}]}`,
			verb:    []string{"install", "app", "--dry-run", "--audit", "advisory"},
			class:   "source_audit_unavailable",
			remedy:  "run the explicit attempt under trusted machine policy",
			setup: func(t *testing.T, configPath, _, _ string) {
				if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
					t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
				}
			},
		},
		{
			name:    "audit-rejected",
			payload: `{"schema_version":2,"sources":{"s":{"path":"./pkgs"}},"skills":[{"name":"review","from":"s","directory":"review"}]}`,
			verb:    []string{"install", "app", "--dry-run", "--audit", "advisory"},
			class:   "source_audit_rejected",
			remedy:  "re-resolve under trusted machine policy",
			setup: func(t *testing.T, configPath, _, home string) {
				if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
					t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
				}
				if code, _, stderr := capture(t, configPath, "install", "app", "--audit", "advisory"); code != exitOK {
					t.Fatalf("install --audit advisory = %d\nstderr:\n%s", code, stderr)
				}
				reports, err := filepath.Glob(filepath.Join(home, "source-audit", "*.report.json"))
				if err != nil || len(reports) != 1 {
					t.Fatalf("audit reports = %v, %v", reports, err)
				}
				if err := os.WriteFile(reports[0], []byte(`{"tampered":true}`), 0o644); err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			configPath, project, home := setupRemediationProject(t)
			writeRemediationManifest(t, project, row.payload)
			if row.setup != nil {
				row.setup(t, configPath, project, home)
			}
			code, _, stderr := capture(t, configPath, row.verb...)
			if code != exitFail {
				t.Fatalf("%v = %d, want %d\nstderr:\n%s", row.verb, code, exitFail, stderr)
			}
			if !strings.Contains(stderr, row.class) {
				t.Fatalf("stderr misses class %q:\n%s", row.class, stderr)
			}
			if !strings.Contains(stderr, row.remedy) {
				t.Fatalf("stderr misses remediation %q:\n%s", row.remedy, stderr)
			}
			if row.absent != "" && strings.Contains(stderr, row.absent) {
				t.Fatalf("stderr repeats guidance %q:\n%s", row.absent, stderr)
			}
		})
	}
}

// setupKitBare builds the documented Git fixture: a bare repository with
// one package under skills/review at tag v1 on branch main. It returns
// the bare path and the tag commit.
func setupKitBare(t *testing.T, root string) (bare, commit string) {
	t.Helper()
	work := filepath.Join(root, "kit-work")
	bare = filepath.Join(root, "kit.git")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	writeCLISkill(t, filepath.Join(work, "skills", "review"), "review")
	runGit(t, work, "init", "-q", "-b", "main")
	runGit(t, work, "add", ".")
	runGit(t, work, "commit", "-qm", "fixture")
	runGit(t, work, "tag", "v1")
	runGit(t, "", "clone", "--quiet", "--bare", "--", work, bare)
	out, err := exec.Command("git", "--git-dir", bare, "rev-parse", "v1^{commit}").Output()
	if err != nil {
		t.Fatalf("rev-parse: %v", err)
	}
	return bare, strings.TrimSpace(string(out))
}

// withRewrittenGitURL puts a stand-in git on PATH that rewrites one
// declared fixture URL to a local bare repository.
func withRewrittenGitURL(t *testing.T, declaredURL, bare string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not available")
	}
	fakeDir := installFakeGitForDraftCLI(t, declaredURL, bare, realGit)
	savedPath, hadPath := os.LookupEnv("PATH")
	t.Cleanup(func() {
		if hadPath {
			_ = os.Setenv("PATH", savedPath)
		} else {
			_ = os.Unsetenv("PATH")
		}
	})
	if err := os.Setenv("PATH", fakeDir+string(os.PathListSeparator)+os.Getenv("PATH")); err != nil {
		t.Fatal(err)
	}
}

// installFailingGitForDraftCLI writes a stand-in git that fails clones of
// failingURL with fixed stderr (exit 128), rewrites rewriteURL to a local
// bare repository, and delegates everything else to the real git. It
// returns a PATH directory to prepend. An empty rewriteURL disables the
// rewrite arm. failingStderr is the fatal line; the script prefixes it
// with git's fixed "Cloning into '<dest>'..." progress line naming the
// clone destination, exactly as the real binary does, so classification
// sees the production shape.
func installFailingGitForDraftCLI(t *testing.T, failingURL, failingStderr, rewriteURL, bare, realGit string) string {
	t.Helper()
	fakeDir := t.TempDir()
	var script strings.Builder
	script.WriteString("#!/bin/sh\n")
	script.WriteString("fail=0; rewrite=0\n")
	script.WriteString("for arg in \"$@\"; do\n")
	script.WriteString("if [ \"$arg\" = '" + failingURL + "' ]; then fail=1; fi\n")
	if rewriteURL != "" {
		script.WriteString("if [ \"$arg\" = '" + rewriteURL + "' ]; then rewrite=1; fi\n")
	}
	script.WriteString("done\n")
	script.WriteString("if [ \"$1\" = \"clone\" ] && [ \"$fail\" = \"1\" ]; then\n")
	script.WriteString("dest=\"\"; for arg in \"$@\"; do dest=\"$arg\"; done\n")
	script.WriteString("printf \"Cloning into '%s'...\\n\" \"$dest\" >&2\n")
	script.WriteString("printf '%s\\n' \"" + failingStderr + "\" >&2\n")
	script.WriteString("exit 128\n")
	script.WriteString("fi\n")
	script.WriteString("args=\"\"; for arg in \"$@\"; do\n")
	if rewriteURL != "" {
		script.WriteString("if [ \"$rewrite\" = \"1\" ] && [ \"$arg\" = '" + rewriteURL + "' ]; then arg='file://" + bare + "'; fi\n")
	}
	script.WriteString("args=\"$args\n$arg\"; done\noldifs=$IFS; IFS='\n'; set -- $args; IFS=$oldifs\nexec '" + realGit + "' \"$@\"\n")
	if err := os.WriteFile(filepath.Join(fakeDir, "git"), []byte(script.String()), 0o700); err != nil {
		t.Fatal(err)
	}
	return fakeDir
}

// withDraftGitArms puts the failing/rewrite stand-in git on PATH with the
// same platform skips as withRewrittenGitURL.
func withDraftGitArms(t *testing.T, failingURL, failingStderr, rewriteURL, bare string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not available")
	}
	fakeDir := installFailingGitForDraftCLI(t, failingURL, failingStderr, rewriteURL, bare, realGit)
	savedPath, hadPath := os.LookupEnv("PATH")
	t.Cleanup(func() {
		if hadPath {
			_ = os.Setenv("PATH", savedPath)
		} else {
			_ = os.Unsetenv("PATH")
		}
	})
	if err := os.Setenv("PATH", fakeDir+string(os.PathListSeparator)+os.Getenv("PATH")); err != nil {
		t.Fatal(err)
	}
}

// TestDraftMachinePolicySetupThroughCLI proves the documented machine
// policy setup end to end: a logical repository declaration resolves
// through source-policy.json beside the manager configuration, and the
// installed marker binds the policy-selected commit. A malformed pin in
// the same file fails closed with remediation and publishes nothing.
func TestDraftMachinePolicySetupThroughCLI(t *testing.T) {
	root := t.TempDir()
	bare, commit := setupKitBare(t, root)
	const declaredURL = "https://fixture.test/kit.git"
	withRewrittenGitURL(t, declaredURL, bare)

	configPath, project := setupCLIProject(t, root)
	t.Setenv("CURATOR_CONFIG", configPath)
	home := filepath.Dir(configPath)
	policy := `{"schema_version":1,"repositories":{"fixture.test/kit":{"endpoints":[{"url":"` + declaredURL + `","authentication":"team-https"}],"fallback":"none"}}}`
	if err := os.WriteFile(filepath.Join(home, "source-policy.json"), []byte(policy), 0o644); err != nil {
		t.Fatal(err)
	}
	payload := `{"schema_version":2,"sources":{"kit":{"repository":"fixture.test/kit","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	writeRemediationManifest(t, project, payload)
	if code, stdout, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
		t.Fatalf("project resolve = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("review")
	if !ok || member.Package.Commit.Hex != commit {
		t.Fatalf("lock binds %+v, want commit %s", member.Package, commit)
	}
	if member.Package.Repository != "fixture.test/kit" {
		t.Fatalf("lock identity = %q, want the canonical key", member.Package.Repository)
	}
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("install = %d\nstderr:\n%s", code, stderr)
	}
	code, stdout, stderr := capture(t, configPath, "status", "app")
	if code != exitOK || !strings.Contains(stdout, "app: review up-to-date") {
		t.Fatalf("status = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}

	// A pin that names no listed endpoint fails closed before any
	// network I/O, with remediation, and the prior lock is untouched.
	lockBefore, err := os.ReadFile(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	badPin := `{"schema_version":1,"repositories":{"fixture.test/kit":{"endpoints":[{"url":"` + declaredURL + `","authentication":"team-https"}],"pin":"https://fixture.test/other.git","fallback":"none"}}}`
	if err := os.WriteFile(filepath.Join(home, "source-policy.json"), []byte(badPin), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, stderr = capture(t, configPath, "project", "refresh", "app")
	if code != exitFail || !strings.Contains(stderr, "repository_policy_invalid") || !strings.Contains(stderr, "fix machine source-policy.json") {
		t.Fatalf("bad-pin refresh = %d, stderr %q, want repository_policy_invalid with remediation", code, stderr)
	}
	lockAfter, err := os.ReadFile(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(lockAfter) != string(lockBefore) {
		t.Fatalf("failed refresh rewrote the lock")
	}
}

// TestDraftDocumentedGitRevisionThroughCLI runs the documented revision
// pin through the real verbs: resolve, install, and status.
func TestDraftDocumentedGitRevisionThroughCLI(t *testing.T) {
	root := t.TempDir()
	bare, commit := setupKitBare(t, root)
	const declaredURL = "https://fixture.test/kit.git"
	withRewrittenGitURL(t, declaredURL, bare)

	configPath, project := setupCLIProject(t, root)
	t.Setenv("CURATOR_CONFIG", configPath)
	payload := `{"schema_version":2,"sources":{"kit":{"git":"` + declaredURL + `","revision":"` + commit + `"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	writeRemediationManifest(t, project, payload)
	if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
		t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("install = %d\nstderr:\n%s", code, stderr)
	}
	code, stdout, stderr := capture(t, configPath, "status", "app")
	if code != exitOK || !strings.Contains(stdout, "app: review up-to-date") {
		t.Fatalf("status = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
}

// TestDraftFetchFailureSanitizedThroughCLI fails a Git acquisition with
// the production clone stderr (git's fixed progress line plus a DNS
// fatal) and requires the positively classified closed-vocabulary
// diagnostic: the availability class, one per-attempt clause, and
// remediation — never the URL, the raw tool output, a secret, or the old
// unknown misclassification. The fake-git arm replaces the former live
// DNS lookup, so the test needs no network.
func TestDraftFetchFailureSanitizedThroughCLI(t *testing.T) {
	const declared = "https://invalid.invalid/no-such-repo.git"
	const fatal = "fatal: unable to access 'https://invalid.invalid/no-such-repo.git/': Could not resolve host: invalid.invalid"
	withDraftGitArms(t, declared, fatal, "", "")
	configPath, project, _ := setupRemediationProject(t)
	payload := `{"schema_version":2,"sources":{"kit":{"git":"` + declared + `","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	writeRemediationManifest(t, project, payload)
	code, _, stderr := capture(t, configPath, "project", "resolve", "app")
	if code != exitFail {
		t.Fatalf("project resolve = %d, want %d\nstderr:\n%s", code, exitFail, stderr)
	}
	for _, want := range []string{
		"repository_endpoint_unavailable",
		"identity invalid.invalid/no-such-repo",
		"endpoint 1 (https): availability: endpoint unavailable",
		"verify the network path and operator authentication",
	} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("stderr misses %q:\n%s", want, stderr)
		}
	}
	// The portable identity is expected above; the endpoint URL, the
	// scheme, raw tool output, and the old unknown misclassification
	// must never appear.
	for _, leak := range []string{declared, "https://", "no-such-repo.git", "git clone failed", "failed: ", "unknown: unclassified failure", "Could not resolve host"} {
		if strings.Contains(stderr, leak) {
			t.Fatalf("stderr leaks endpoint detail %q:\n%s", leak, stderr)
		}
	}
	if _, err := os.Stat(filepath.Join(project, "Skillfile.lock.json")); !os.IsNotExist(err) {
		t.Fatalf("failed resolve published a lock: %v", err)
	}
}

// TestDraftAvailabilityFallbackThroughCLI proves a positively classified
// clone failure advances the availability-auth fallback at the CLI entry:
// the first endpoint fails with DNS stderr while the second serves the
// documented fixture, so resolve exits 0 with the lock bound to the
// alternate commit and only the first attempt rendered as a failure.
func TestDraftAvailabilityFallbackThroughCLI(t *testing.T) {
	const failingURL = "https://fixture.test/kit.git"
	const failingFatal = "fatal: unable to access 'https://fixture.test/kit.git/': Could not resolve host: fixture.test"
	const alternateURL = "git@fixture.test:kit.git"
	root := t.TempDir()
	bare, commit := setupKitBare(t, root)
	withDraftGitArms(t, failingURL, failingFatal, alternateURL, bare)

	configPath, project := setupCLIProject(t, root)
	t.Setenv("CURATOR_CONFIG", configPath)
	home := filepath.Dir(configPath)
	policy := `{"schema_version":1,"repositories":{"fixture.test/kit":{"endpoints":[{"url":"` + failingURL + `","authentication":"team-https"},{"url":"` + alternateURL + `","authentication":"team-alt"}],"fallback":"availability-auth"}}}`
	if err := os.WriteFile(filepath.Join(home, "source-policy.json"), []byte(policy), 0o644); err != nil {
		t.Fatal(err)
	}
	payload := `{"schema_version":2,"sources":{"kit":{"repository":"fixture.test/kit","tag":"v1"}},"skills":[{"name":"review","from":"kit","directory":"skills/review"}]}`
	writeRemediationManifest(t, project, payload)
	code, _, stderr := capture(t, configPath, "project", "resolve", "app")
	if code != exitOK {
		t.Fatalf("project resolve = %d, want %d\nstderr:\n%s", code, exitOK, stderr)
	}
	if want := `endpoint 1 (https, provider "team-https"): availability: endpoint unavailable`; !strings.Contains(stderr, want) {
		t.Fatalf("stderr misses %q:\n%s", want, stderr)
	}
	for _, leak := range []string{failingURL, alternateURL, "endpoint 2: not attempted", "unknown: unclassified failure", "Could not resolve host"} {
		if strings.Contains(stderr, leak) {
			t.Fatalf("stderr leaks or misreports %q:\n%s", leak, stderr)
		}
	}
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("review")
	if !ok || member.Package.Commit.Hex != commit {
		t.Fatalf("lock binds %+v, want alternate commit %s", member.Package, commit)
	}
	if member.Package.Repository != "fixture.test/kit" {
		t.Fatalf("lock identity = %q, want the canonical key", member.Package.Repository)
	}
}

// TestDraftStripCloneFramingTable pins the fail-closed strip: git's exact
// progress line is dropped before classification, while a near-match,
// diagnostic-only input, and empty input pass through unchanged.
func TestDraftStripCloneFramingTable(t *testing.T) {
	framing := "Cloning into '/tmp/kit-clone'..."
	fatal := "fatal: unable to access 'https://fixture.test/kit.git/': Could not resolve host: fixture.test"
	rows := []struct {
		name   string
		detail string
		want   buildrepo.FailureClass
	}{
		{"framing-plus-dns", framing + "\n" + fatal, buildrepo.FailureAvailability},
		{"dns-alone", fatal, buildrepo.FailureAvailability},
		{"framing-alone", framing, buildrepo.FailureUnknown},
		{"near-match-kept", "Cloning into '/tmp/kit-clone'... and more\n" + fatal, buildrepo.FailureUnknown},
		{"empty", "", buildrepo.FailureUnknown},
	}
	for _, row := range rows {
		if got := buildrepo.ClassifyFetchOutput(stripCloneFraming(row.detail)); got != row.want {
			t.Errorf("%s: class = %s, want %s", row.name, got, row.want)
		}
	}
	if got := stripCloneFraming(fatal); got != fatal {
		t.Errorf("strip changed diagnostic-only input: %q", got)
	}
}

// runDocShim executes one installed project shim and returns its output.
func runDocShim(t *testing.T, project, command string) string {
	t.Helper()
	out, err := exec.Command(filepath.Join(project, ".agents", "bin", command)).Output() // #nosec G204 -- the installed shim under test
	if err != nil {
		t.Fatalf("shim %s did not run: %v", command, err)
	}
	return string(out)
}

// TestDraftLaunchConsumesFrozenRuntimeThroughCLI proves launch never
// rescans live inputs at the CLI entry: the installed shim executes
// the frozen runtime before and after a live mutation, status stays
// current, a reinstall without refresh keeps the frozen bytes, and only
// an explicit refresh plus install publishes the new bytes.
func TestDraftLaunchConsumesFrozenRuntimeThroughCLI(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executes POSIX skill commands")
	}
	root := t.TempDir()
	configPath, project := setupCLIProject(t, root)
	t.Setenv("CURATOR_CONFIG", configPath)
	scriptDir := filepath.Join(project, "pkgs", "review")
	writeCLIScriptSkill(t, scriptDir, "review")
	payload := `{"schema_version":2,"sources":{"s":{"path":"./pkgs"}},"skills":[{"name":"review","from":"s","directory":"review"}]}`
	writeRemediationManifest(t, project, payload)
	if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
		t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("install = %d\nstderr:\n%s", code, stderr)
	}
	if got := runDocShim(t, project, "tool"); got != "original\n" {
		t.Fatalf("shim output = %q, want %q", got, "original\n")
	}
	// A live mutation changes neither the launched bytes, the status
	// verdict, nor a reinstall without refresh.
	script := filepath.Join(scriptDir, "scripts", "tool.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho mutated\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := runDocShim(t, project, "tool"); got != "original\n" {
		t.Fatalf("shim after live mutation = %q, want frozen %q", got, "original\n")
	}
	code, stdout, stderr := capture(t, configPath, "status", "app")
	if code != exitOK || !strings.Contains(stdout, "app: review up-to-date") {
		t.Fatalf("status after live mutation = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("reinstall without refresh = %d\nstderr:\n%s", code, stderr)
	}
	if got := runDocShim(t, project, "tool"); got != "original\n" {
		t.Fatalf("shim after reinstall without refresh = %q, want frozen %q", got, "original\n")
	}
	// The frozen runtime is verified on every consumption: a tampered
	// store fails instead of launching unverified bytes.
	lock, err := sourcelock.Read(filepath.Join(project, "Skillfile.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("review")
	if !ok {
		t.Fatalf("lock misses review: %+v", lock.Members)
	}
	digest := strings.TrimPrefix(member.Package.Snapshot, "sha256:")
	home := filepath.Dir(configPath)
	stored := filepath.Join(home, "local-snapshots", digest, "snapshot", "scripts", "tool.sh")
	storedBytes, err := os.ReadFile(stored)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stored, []byte("#!/bin/sh\necho tampered\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "install", "app", "--dry-run"); code != exitFail || !strings.Contains(stderr, "source_snapshot_changed") {
		t.Fatalf("tampered-store install = %d, stderr %q, want source_snapshot_changed", code, stderr)
	}
	if err := os.WriteFile(stored, storedBytes, 0o755); err != nil {
		t.Fatal(err)
	}
	// Only an explicit refresh plus install publishes the new bytes.
	if code, _, stderr := capture(t, configPath, "project", "refresh", "app"); code != exitOK {
		t.Fatalf("project refresh = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("install after refresh = %d\nstderr:\n%s", code, stderr)
	}
	if got := runDocShim(t, project, "tool"); got != "mutated\n" {
		t.Fatalf("shim after refresh plus install = %q, want %q", got, "mutated\n")
	}
}

// TestDraftRepairRestoresDriftedContentThroughCLI proves install repairs
// drifted installed bytes from the frozen lock at the CLI entry while
// the lock bytes stay identical.
func TestDraftRepairRestoresDriftedContentThroughCLI(t *testing.T) {
	configPath, project, _ := setupRemediationProject(t)
	payload := `{"schema_version":2,"sources":{"s":{"path":"./pkgs"}},"skills":[{"name":"review","from":"s","directory":"review"}]}`
	writeRemediationManifest(t, project, payload)
	if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
		t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
	}
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("install = %d\nstderr:\n%s", code, stderr)
	}
	lockPath := filepath.Join(project, "Skillfile.lock.json")
	lockBefore, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	installed := filepath.Join(project, ".agents", "skills", "review", "SKILL.md")
	staged, err := os.ReadFile(installed)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(installed, append(staged, []byte("drift")...), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := capture(t, configPath, "status", "app")
	if code != exitOK || !strings.Contains(stdout, "app: review content-drift") {
		t.Fatalf("status after drift = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("repair install = %d\nstderr:\n%s", code, stderr)
	}
	restored, err := os.ReadFile(installed)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != string(staged) {
		t.Fatalf("repair left drifted bytes behind")
	}
	lockAfter, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(lockAfter) != string(lockBefore) {
		t.Fatalf("repair rewrote the lock")
	}
	code, stdout, stderr = capture(t, configPath, "status", "app")
	if code != exitOK || !strings.Contains(stdout, "app: review up-to-date") {
		t.Fatalf("status after repair = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
}

// TestDraftDocsPinExamples pins the documented contract in the
// repository docs: every exercised acquisition shape, the default project lane,
// label, the machine policy setup, and every stable class remedy.
func TestDraftDocsPinExamples(t *testing.T) {
	cli := readRepoDoc(t, "docs", "cli.md")
	readme := readRepoDoc(t, "README.md")
	trouble := readRepoDoc(t, "docs", "troubleshooting.md")
	compact := compactJSON(cli)
	for _, shape := range []string{
		docLocalRelativePayload,
		docAbsolutePathShape,
		docGitTagShape,
		docLogicalBranchShape,
		docCollectionListShape,
		docCollectionStarShape,
		docCollectionExclude,
	} {
		if !strings.Contains(compact, compactJSON(shape)) {
			t.Errorf("docs/cli.md misses the exercised shape %s", shape)
		}
	}
	for _, marker := range []string{
		"Skillfile schema 2 is the default project reader and install path",
		"no on-disk migration is implicit",
		"source-policy.json",
		`"root_inputs"`,
		`"availability-auth"`,
		"curator project refresh",
		"never rescan",
		"credentials from the invoking environment",
		"is never consulted",
		"CURATOR_DRAFT_TRANSPORT_RESOLUTION",
		"allow-list",
		"does not honour proxy",
		"GIT_SSH_COMMAND",
		"ProxyCommand=none",
		"ssh-keyscan",
		"answers git's username and password prompts",
	} {
		if !strings.Contains(cli, marker) {
			t.Errorf("docs/cli.md misses %q", marker)
		}
	}
	for _, marker := range []string{
		"Skillfile schema 2 source declarations are supported by default",
		"no on-disk migration is implicit",
		`"path": "./pkgs"`,
		"source-policy.json",
		"never rescan live inputs",
	} {
		if !strings.Contains(readme, marker) {
			t.Errorf("README.md misses %q", marker)
		}
	}
	for _, class := range []string{
		"source_alias_unknown", "source_selection_invalid",
		"source_member_missing", "source_member_invalid",
		"source_name_conflict", "source_output_overlap",
		"source_snapshot_changed", "source_snapshot_unavailable",
		"source_lock_stale", "repository_endpoint_unavailable",
		"repository_policy_invalid", "repository_mirror_undeclared",
		"repository_alias_unknown", "source_audit_rejected",
		"source_audit_unavailable", "build_repository_identity_invalid",
	} {
		if !strings.Contains(trouble, "### "+class) {
			t.Errorf("docs/troubleshooting.md misses section %s", class)
		}
	}
	for _, marker := range []string{
		"These stable classes describe schema-2 project source resolution",
	} {
		if !strings.Contains(trouble, marker) {
			t.Errorf("docs/troubleshooting.md misses %q", marker)
		}
	}
	// Each remedy keyword below is asserted verbatim against the CLI
	// remediation output by TestWithDraftRemediationTable and
	// TestDraftRemediationThroughCLI; the docs must carry the same words.
	for class, keyword := range map[string]string{
		"source_alias_unknown":              "declare the alias",
		"source_selection_invalid":          "fix the named selector",
		"source_member_missing":             "add the named member directory",
		"source_member_invalid":             "fix the named package",
		"source_name_conflict":              "exactly one selection",
		"source_output_overlap":             "move the authored package out of managed output",
		"source_snapshot_changed":           "restore the declared source to the package identity and",
		"source_snapshot_unavailable":       "restore access to the declared path or Git source",
		"source_lock_stale":                 "curator project refresh",
		"repository_endpoint_unavailable":   "verify the network path",
		"repository_policy_invalid":         "fix machine source-policy.json",
		"repository_mirror_undeclared":      "equal to the entry key",
		"repository_alias_unknown":          "declare the alias in the",
		"source_audit_rejected":             "re-resolve under trusted machine policy",
		"source_audit_unavailable":          "run the explicit attempt under trusted machine policy",
		"build_repository_identity_invalid": "fix the endpoint entry",
	} {
		if !strings.Contains(trouble, keyword) {
			t.Errorf("docs/troubleshooting.md misses the %s remedy keyword %q", class, keyword)
		}
	}
	// Isolation wording (TASK-260920-3ccq6b): the askpass fix and the ssh
	// and proxy operator consequences pinned in the endpoint remedy.
	for _, marker := range []string{
		"answers git's username and password prompts",
		"embed the username in",
		"are not read",
		"ssh-keyscan",
		"is not honoured on this lane",
	} {
		if !strings.Contains(trouble, marker) {
			t.Errorf("docs/troubleshooting.md misses %q", marker)
		}
	}
}

// readRepoDoc reads one repository doc relative to the package directory.
func readRepoDoc(t *testing.T, elements ...string) string {
	t.Helper()
	parts := append([]string{"..", ".."}, elements...)
	payload, err := os.ReadFile(filepath.Join(parts...))
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}

// compactJSON strips JSON-insignificant whitespace so the docs pin
// tolerates pretty-printed manifests.
func compactJSON(text string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\n', '\r':
			return -1
		}
		return r
	}, text)
}
