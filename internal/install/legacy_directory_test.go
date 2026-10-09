package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/gitops"
	manifestpkg "github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/sourcelock"
)

// legacyDirectoryGit is the canonical network identity shared by the
// subdirectory-selected fixtures below.
const legacyDirectoryGit = "https://github.com/example/role-skills.git"

func (e *env) gitOutput(dir string, args ...string) string {
	e.t.Helper()
	gitArgs := append([]string{"-c", "commit.gpgsign=false", "-c", "tag.gpgSign=false"}, args...)
	out := gitFixture(e.t, dir, args, gitArgs, []string{
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
	})
	return strings.TrimSpace(string(out))
}

// writeLegacyPackageSkill writes one skill package at dir with the given
// manifest schema and skill requirements.
func writeLegacyPackageSkill(e *env, dir, name string, schema int, requirements map[string]any) {
	e.t.Helper()
	e.write(dir, "SKILL.md", "---\nname: "+name+"\ndescription: Test "+name+"\n---\n# "+name+"\n")
	e.write(dir, "references/info.md", "context\n")
	spec := map[string]any{
		"schema_version": schema,
		"capabilities":   map[string]any{},
		"commands":       map[string]any{},
		"dependencies":   map[string]any{"skills": requirements},
	}
	payload, err := json.Marshal(spec)
	if err != nil {
		e.t.Fatal(err)
	}
	e.write(dir, "agent-skill.json", string(payload))
}

// setupLegacyDirectoryRepos builds the schema-1 lane fixture: a provider
// repository holding schema-9 backend and schema-8 helper packages in
// subfolders, and a schema-9 consumer repository requiring both at their
// subfolders. It returns the provider commit the requirements pin.
//
// The provider checkout lives at the first-sorted requirement name: the
// legacy lane resolves a requirement source below the skills root, and
// the second requirement reuses the same repository by canonical network
// identity, so both subfolders share one snapshot at one commit.
func setupLegacyDirectoryRepos(e *env, backendDir, helperDir string) string {
	e.t.Helper()
	provider := filepath.Join(e.skillsRoot, "backend")
	if err := os.MkdirAll(provider, 0o755); err != nil {
		e.t.Fatal(err)
	}
	e.git(provider, "init", "-q", "-b", "main")
	writeLegacyPackageSkill(e, filepath.Join(provider, "skills", "backend"), "backend", 9, map[string]any{})
	writeLegacyPackageSkill(e, filepath.Join(provider, "skills", "helper"), "helper", 8, map[string]any{})
	e.git(provider, "add", ".")
	e.git(provider, "commit", "-qm", "role packages")
	commit := e.gitOutput(provider, "rev-parse", "HEAD")

	consumer := filepath.Join(e.skillsRoot, "consumer")
	if err := os.MkdirAll(consumer, 0o755); err != nil {
		e.t.Fatal(err)
	}
	e.git(consumer, "init", "-q", "-b", "main")
	writeLegacyPackageSkill(e, consumer, "consumer", 9, map[string]any{
		"backend": map[string]any{
			"git":       legacyDirectoryGit,
			"ref":       map[string]any{"kind": "revision", "value": commit},
			"directory": backendDir,
		},
		"helper": map[string]any{
			"git":       legacyDirectoryGit,
			"ref":       map[string]any{"kind": "revision", "value": commit},
			"directory": helperDir,
		},
	})
	e.git(consumer, "add", ".")
	e.git(consumer, "commit", "-qm", "init")
	e.git(consumer, "tag", "v1")
	return commit
}

func readLegacyMarker(t *testing.T, dir string) (*marker.Marker, map[string]json.RawMessage) {
	t.Helper()
	recorded := marker.Read(dir)
	if recorded == nil {
		t.Fatalf("marker missing in %s", dir)
	}
	payload, err := os.ReadFile(filepath.Join(dir, marker.Name))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	return recorded, raw
}

// TestLegacyLaneRecordsDependencyDirectory installs a schema-1 project
// whose schema-9 root requires subdirectory-selected packages through
// the real install entry. The schema-9 installation records draft marker
// v6 with the normalized directory; the schema-8 subdirectory selection
// records v5 with its directory; the schema-9 root records v6 with the
// configured-git root selection; and the local audit records bind the
// same directories (core §4.4, skillfile-sources §4, draft-sources-v2).
// A second install through the same entry is up-to-date.
func TestLegacyLaneRecordsDependencyDirectory(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	commit := setupLegacyDirectoryRepos(e, "skills/backend", "skills/helper")
	e.declare("consumer")
	e.cfg.Audit.Enabled = true

	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("install: %+v", result)
	}

	backend, backendRaw := readLegacyMarker(t, filepath.Join(e.project, ".agents", "skills", "backend"))
	if backend.SchemaVersion != marker.SchemaV6 || backend.SkillSchemaVersion != 9 {
		t.Fatalf("backend marker = schema %d skill %d, want marker 6 skill 9",
			backend.SchemaVersion, backend.SkillSchemaVersion)
	}
	if backend.Package == nil || backend.Package.Kind != "network-git" ||
		backend.Package.Repository != "github.com/example/role-skills" ||
		backend.Package.Commit == nil || backend.Package.Commit.ObjectFormat != "sha1" ||
		backend.Package.Commit.Hex != commit || backend.Package.Directory != "skills/backend" {
		t.Fatalf("backend marker lost the selected package identity: %+v", backend.Package)
	}
	if backend.LockSHA256 == "" {
		t.Fatalf("backend marker carries no lock binding")
	}
	for _, field := range []string{"source", "git", "ref_kind", "ref", "commit"} {
		if _, present := backendRaw[field]; present {
			t.Fatalf("backend v6 marker carries replaced legacy field %q", field)
		}
	}

	helper, _ := readLegacyMarker(t, filepath.Join(e.project, ".agents", "skills", "helper"))
	if helper.SchemaVersion != marker.SchemaV5 || helper.SkillSchemaVersion != 8 {
		t.Fatalf("helper marker = schema %d skill %d, want marker 5 skill 8",
			helper.SchemaVersion, helper.SkillSchemaVersion)
	}
	if helper.Package == nil || helper.Package.Kind != "network-git" ||
		helper.Package.Directory != "skills/helper" {
		t.Fatalf("helper marker lost the selected package identity: %+v", helper.Package)
	}

	consumer, _ := readLegacyMarker(t, filepath.Join(e.project, ".agents", "skills", "consumer"))
	if consumer.SchemaVersion != marker.SchemaV6 || consumer.SkillSchemaVersion != 9 {
		t.Fatalf("consumer marker = schema %d skill %d, want marker 6 skill 9",
			consumer.SchemaVersion, consumer.SkillSchemaVersion)
	}
	if consumer.Package == nil || consumer.Package.Kind != "configured-git" ||
		consumer.Package.Source != "consumer" || consumer.Package.Directory != "." {
		t.Fatalf("consumer marker lost the root package identity: %+v", consumer.Package)
	}

	// F2: every migrated marker of one install shares the specified
	// skillfile-sources §3 digest of the effective lock — the CCJ-1
	// SHA-256 of the lock with only lock_sha256 omitted — binding the
	// installed selection and declared ref through the validated lock
	// and matching manifest (§4). An invented per-node preimage differs
	// per node and never equals the reconstructed lock digest.
	if backend.LockSHA256 == "" || helper.LockSHA256 != backend.LockSHA256 ||
		consumer.LockSHA256 != backend.LockSHA256 {
		t.Fatalf("migrated markers do not share one lock digest: backend=%q helper=%q consumer=%q",
			backend.LockSHA256, helper.LockSHA256, consumer.LockSHA256)
	}
	if want := legacyEffectiveLockDigest(t, e); backend.LockSHA256 != want {
		t.Fatalf("marker lock_sha256 = %q, want the specified effective-lock digest %q", backend.LockSHA256, want)
	}

	// F5: the local audit records bind the complete package identity —
	// repository, commit, and directory — equal to the installed
	// marker packages (core §4.4, skillfile-sources §4 cache equality).
	type auditPackage struct {
		Kind       string `json:"kind"`
		Repository string `json:"repository"`
		Source     string `json:"source"`
		Commit     struct {
			ObjectFormat string `json:"object_format"`
			Hex          string `json:"hex"`
		} `json:"commit"`
		Directory string `json:"directory"`
	}
	type auditRecord struct {
		Skill     string       `json:"skill"`
		Commit    string       `json:"commit"`
		Directory string       `json:"directory"`
		Package   auditPackage `json:"package"`
	}
	audited := map[string]auditRecord{}
	matches, err := filepath.Glob(filepath.Join(e.home, "audit", "*", "verdict-*.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range matches {
		payload, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var record auditRecord
		if err := json.Unmarshal(payload, &record); err != nil {
			t.Fatal(err)
		}
		audited[record.Skill] = record
	}
	for skill, marked := range map[string]*marker.Marker{"backend": backend, "helper": helper, "consumer": consumer} {
		record, ok := audited[skill]
		if !ok {
			t.Fatalf("no audit record for %s (records: %v)", skill, audited)
		}
		if record.Directory != marked.Package.Directory {
			t.Fatalf("audit record for %s binds directory %q, want %q", skill, record.Directory, marked.Package.Directory)
		}
		if record.Commit != marked.Package.Commit.Hex {
			t.Fatalf("audit record for %s binds commit %q, want %q", skill, record.Commit, marked.Package.Commit.Hex)
		}
		if record.Package.Kind != marked.Package.Kind ||
			record.Package.Repository != marked.Package.Repository ||
			record.Package.Source != marked.Package.Source ||
			record.Package.Commit.ObjectFormat != marked.Package.Commit.ObjectFormat ||
			record.Package.Commit.Hex != marked.Package.Commit.Hex ||
			record.Package.Directory != marked.Package.Directory {
			t.Fatalf("audit record for %s binds package %+v, want the marker identity %+v",
				skill, record.Package, marked.Package)
		}
	}

	second := e.install(Options{})
	if second.Status != "ok" {
		t.Fatalf("second: %+v", second)
	}
	if joined := strings.Join(second.Messages, "\n"); !strings.Contains(joined, "up-to-date") {
		t.Fatalf("second install must be up-to-date:\n%s", joined)
	}
}

// legacyEffectiveLockDigest independently reconstructs the specified
// skillfile-sources §3 lock digest of one installed fixture: the
// manifest digest over the declaring Skillfile bytes plus one member per
// closure node, built only from exported lock primitives. It proves the
// marker's lock_sha256 is the CCJ-1 digest of the validated effective
// lock, not an invented per-node preimage.
func legacyEffectiveLockDigest(t *testing.T, e *env) string {
	t.Helper()
	skillfilePath := filepath.Join(e.project, "Skillfile.json")
	payload, err := os.ReadFile(skillfilePath)
	if err != nil {
		t.Fatal(err)
	}
	projectManifest, err := manifestpkg.ParseBytes(payload, skillfilePath)
	if err != nil {
		t.Fatal(err)
	}
	manifestSHA, err := sourcelock.ManifestDigest(payload)
	if err != nil {
		t.Fatal(err)
	}
	nodes, err := closure.Build(closure.Options{
		SkillsRoot: e.skillsRoot, Home: e.home, AllowedSources: e.cfg.AllowedSources,
		FetchedRepos: map[string]bool{},
	}, projectManifest, map[string]devsub.Substitution{})
	if err != nil {
		t.Fatal(err)
	}
	indexByName := map[string]int{}
	for index, decl := range projectManifest.Skills {
		indexByName[decl.Name] = index
	}
	members := make([]sourcelock.Member, 0, len(nodes))
	for _, node := range nodes {
		directory := node.Directory
		if directory == "" {
			directory = "."
		}
		var selection *int
		if index, ok := indexByName[node.Name]; ok {
			value := index
			selection = &value
		}
		commit := sourcelock.Commit{Hex: node.Resolved.Commit}
		switch len(commit.Hex) {
		case 40:
			commit.ObjectFormat = "sha1"
		case 64:
			commit.ObjectFormat = "sha256"
		default:
			t.Fatalf("resolved commit %q for %s is not a full object id", commit.Hex, node.Name)
		}
		var pkg sourcelock.Package
		if node.Identity != "" {
			pkg, err = sourcelock.NetworkGitPackage(node.Identity, commit, directory)
		} else {
			source := node.Decl.Source
			if source == "" {
				source = node.Name
			}
			pkg, err = sourcelock.ConfiguredGitPackage(source, commit)
		}
		if err != nil {
			t.Fatal(err)
		}
		content, err := closure.ContentHashFor(node.Snapshot, node.Spec)
		if err != nil {
			t.Fatal(err)
		}
		members = append(members, sourcelock.Member{
			Name: node.Name, Selection: selection, Directory: directory,
			Package: pkg, ContentSHA256: content,
		})
	}
	lock, err := sourcelock.New(manifestSHA, members)
	if err != nil {
		t.Fatal(err)
	}
	return lock.LockSHA256
}

// TestLegacyLockPackageArms pins the legacy-lane package staging: a
// network identity records the network-git arm with the normalized
// directory, a local source records the configured-git root arm, and a
// subdirectory selection without a network identity fails closed instead
// of recording a directory the configured-git arm forbids. A malformed
// commit stages nothing.
func TestLegacyLockPackageArms(t *testing.T) {
	t.Parallel()
	commit := strings.Repeat("ab", 20)
	network := &closure.Node{
		Name:      "backend",
		Decl:      manifestpkg.Decl{Name: "backend", Source: "backend", Git: legacyDirectoryGit, Ref: manifestpkg.Ref{Kind: "revision", Value: commit}},
		Directory: "skills/backend",
		Resolved:  gitops.ResolvedRef{Kind: "revision", Ref: commit, Commit: commit},
		Identity:  "github.com/example/role-skills",
	}
	pkg, err := legacyLockPackage(network)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Kind != "network-git" || pkg.Repository != "github.com/example/role-skills" ||
		pkg.Commit.ObjectFormat != "sha1" || pkg.Commit.Hex != commit ||
		pkg.Directory != "skills/backend" {
		t.Fatalf("network-git arm = %+v", pkg)
	}

	local := &closure.Node{
		Name:      "consumer",
		Decl:      manifestpkg.Decl{Name: "consumer", Source: "consumer", Ref: manifestpkg.Ref{Kind: "tag", Value: "v1"}},
		Directory: ".",
		Resolved:  gitops.ResolvedRef{Kind: "tag", Ref: "v1", Commit: commit},
	}
	localPkg, err := legacyLockPackage(local)
	if err != nil {
		t.Fatal(err)
	}
	if localPkg.Kind != "configured-git" || localPkg.Source != "consumer" || localPkg.Directory != "." {
		t.Fatalf("configured-git arm = %+v", localPkg)
	}

	subdirLocal := *local
	subdirLocal.Directory = "skills/backend"
	if _, err := legacyLockPackage(&subdirLocal); err == nil ||
		!strings.Contains(err.Error(), "source_selection_invalid") {
		t.Fatalf("subdir without network identity = %v, want source_selection_invalid", err)
	}

	badCommit := *network
	badCommit.Resolved.Commit = "not-a-commit"
	if _, err := legacyLockPackage(&badCommit); err == nil {
		t.Fatalf("malformed commit was staged")
	}
}

// TestBuildMarkerRecordsReceipt3ForMigratedLegacyNode pins the migrated
// marker shape: a schema-9 node with compiled commands records the staged
// package and the shared effective-lock digest (core §4.4 permits builds
// from schema-9 providers), and its build entries bind receipt version 3
// with the execution policy on every driver (skillfile-sources §4). A
// migrated node without a staged migration fails closed instead of
// recording a legacy marker no reader may accept for it.
func TestBuildMarkerRecordsReceipt3ForMigratedLegacyNode(t *testing.T) {
	t.Parallel()
	commit := strings.Repeat("ab", 20)
	node := &closure.Node{
		Name:      "backend",
		Decl:      manifestpkg.Decl{Name: "backend", Source: "backend", Git: legacyDirectoryGit, Ref: manifestpkg.Ref{Kind: "revision", Value: commit}},
		Directory: "skills/backend",
		Resolved:  gitops.ResolvedRef{Kind: "revision", Ref: commit, Commit: commit},
		Identity:  "github.com/example/role-skills",
		Spec: &skillspec.Spec{
			SchemaVersion: 9,
			Commands:      map[string]skillspec.Command{"tool": {Name: "tool", Type: "build", Driver: "go-v1"}},
			Dependencies:  map[string]skillspec.CommandDependency{},
			Requirements:  map[string]skillspec.Requirement{},
		},
		Edges: []closure.Edge{{Consumer: closure.ProjectEdge, Mode: "full"}},
	}
	lockPackage, err := legacyLockPackage(node)
	if err != nil {
		t.Fatal(err)
	}
	binding := "sha256:" + strings.Repeat("cd", 32)
	legacy := &legacyLanePlan{
		Packages:   map[string]*marker.Package{"backend": draftMarkerPackage(lockPackage)},
		LockSHA256: binding,
	}
	builds := map[string]marker.Build{"tool": {Driver: "go-v1"}}
	staged, err := buildMarker(node, "en", []string{"claude_code"}, []string{"tool"}, nil, nil, builds, nil, legacy)
	if err != nil {
		t.Fatal(err)
	}
	if staged.Package == nil || staged.Package.Directory != "skills/backend" {
		t.Fatalf("migrated marker lost the staged package: %+v", staged.Package)
	}
	if staged.LockSHA256 != binding {
		t.Fatalf("migrated marker lock = %q, want the shared digest", staged.LockSHA256)
	}
	recorded, ok := staged.Builds["tool"]
	if !ok || recorded.ReceiptSchemaVersion != 3 || recorded.ExecutionPolicy != "manager-worker-v1" {
		t.Fatalf("migrated build record = %+v, want receipt 3 with the execution policy", staged.Builds["tool"])
	}
	for _, field := range []string{staged.Source, staged.RefKind, staged.Ref, staged.Commit, staged.Git} {
		if field != "" {
			t.Fatalf("migrated marker carries a replaced legacy identity field: %+v", staged)
		}
	}

	if _, err := buildMarker(node, "en", []string{"claude_code"}, []string{"tool"}, nil, nil, builds, nil, nil); err == nil ||
		!strings.Contains(err.Error(), "source_member_missing") {
		t.Fatalf("migrated node without a staged migration = %v, want source_member_missing", err)
	}
}

// TestLegacyLaneRefusesGlobDependencyDirectory pins the dependency-path
// glob class through the real install entry: `?` and `[` are refused on
// the dependency path exactly like the spec's `*` vector, the diagnostic
// names the directory field, and nothing is materialized.
func TestLegacyLaneRefusesGlobDependencyDirectory(t *testing.T) {
	t.Parallel()
	for _, directory := range []string{"skills/back?end", "skills/back[eo]nd", "skills/*"} {
		t.Run(strings.ReplaceAll(directory, "/", "_"), func(t *testing.T) {
			t.Parallel()
			e := newEnv(t)
			setupLegacyDirectoryRepos(e, directory, "skills/helper")
			e.declare("consumer")

			result := e.install(Options{})
			if result.Status != "failed" {
				t.Fatalf("install with directory %q = %+v, want failed", directory, result)
			}
			joined := strings.Join(append(append([]string{}, result.Errors...), result.Messages...), "\n")
			if !strings.Contains(joined, "dependencies.skills.backend.directory") {
				t.Fatalf("error does not name the directory field:\n%s", joined)
			}
			if !strings.Contains(joined, "portable contained path") {
				t.Fatalf("error does not carry the directory grammar diagnostic:\n%s", joined)
			}
			if _, err := os.Stat(filepath.Join(e.project, ".agents")); !os.IsNotExist(err) {
				t.Fatalf("refused install materialized state: %v", err)
			}
		})
	}
}
