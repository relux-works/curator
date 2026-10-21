package envprofile

import (
	"bytes"
	"encoding/json"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hashing"
)

type directorySnapshot struct {
	Files    map[string]string
	Symlinks map[string]string
}
type directoryFixture struct {
	CandidateLock   json.RawMessage `json:"candidate_lock"`
	Repository      string
	ProfileManifest map[string]any `json:"profile_manifest"`
	Snapshot        directorySnapshot
	RootSkill       directorySnapshot `json:"root_skill_snapshot"`
	ContextSkill    directorySnapshot `json:"context_skill_snapshot"`
}
type directoryVector struct {
	Name, Entry       string
	Changes, Expected map[string]any
}

func readDirectoryJSON(t *testing.T, name string, out any) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("../contextpkg/testdata/profile-skill-directories-v1", name))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, out); err != nil {
		t.Fatal(err)
	}
}
func directoryJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
func directoryFixtureLoad(t *testing.T) directoryFixture {
	t.Helper()
	var f directoryFixture
	readDirectoryJSON(t, "fixtures.json", &f)
	return f
}
func directoryRequirement(f directoryFixture) map[string]any {
	return f.ProfileManifest["requires"].(map[string]any)["skills"].(map[string]any)["formatter"].(map[string]any)
}

// Every published lifecycle vector drives Install/Update/Sync/EnsureDefault,
// followed by Use and Resolve for successful materialization. Git OIDs replace
// symbolic fixture pins; the historical root notes.md contact witness moves to
// references/notes.md under core §3. No network or real operator home is used.
func TestSkillDirectoryLifecycleVectors(t *testing.T) {
	var corpus struct{ Cases []directoryVector }
	readDirectoryJSON(t, "vectors.json", &corpus)
	if len(corpus.Cases) != 24 {
		t.Fatalf("vector count=%d", len(corpus.Cases))
	}
	for _, tc := range corpus.Cases {
		t.Run(tc.Name, func(t *testing.T) { driveDirectoryVector(t, tc) })
	}
}

// The selected package's visible URL is admitted with its own host declared;
// this is the nearby positive control for the published undeclared-contact case.
func TestSkillDirectoryMarkdownDeclaredControl(t *testing.T) {
	driveDirectoryVector(t, directoryVector{
		Name: "selected-visible-declared-host", Entry: "profile install",
		Changes:  map[string]any{"snapshot.files.skills/formatter/references/notes.md": "Contact https://api.example.net for formatting guidance.\n"},
		Expected: map[string]any{"outcome": "materialized"},
	})
}

// Selected skill metadata must satisfy core section 3 before any profile or
// store publication; the valid control retains the same production install.
func TestSkillDirectoryMetadata(t *testing.T) {
	for _, tc := range []struct {
		name, skill string
		valid       bool
	}{
		{"valid", "---\nname: formatter\ndescription: Valid skill\n---\n# Formatter\n", true},
		{"missing-frontmatter", "# Formatter\n", false},
		{"invalid-yaml", "---\nname: [\ndescription: Test\n---\n", false},
		{"wrong-name", "---\nname: sibling\ndescription: Test\n---\n", false},
		{"empty-description", "---\nname: formatter\ndescription: ''\n---\n", false},
		{"invalid-trigger", "---\nname: formatter\ndescription: Test\ntriggers: ['']\n---\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expected := map[string]any{"outcome": "materialized"}
			if !tc.valid {
				expected = map[string]any{"outcome": "refused", "diagnostic": "manifest_invalid"}
			}
			driveDirectoryVector(t, directoryVector{Name: tc.name, Entry: "profile install", Changes: map[string]any{"snapshot.files.skills/formatter/SKILL.md": tc.skill}, Expected: expected})
		})
	}
}

func driveDirectoryVector(t *testing.T, tc directoryVector) {
	t.Helper()
	f := directoryFixtureLoad(t)
	home := t.TempDir()
	operator := t.TempDir()
	pinOperatorHome(t, operator)
	native := pinHomes(t)
	ids := newGitIdentities(t)
	if tc.Entry != "first profile-surface use" {
		if err := EnsureDefaultWithPolicy(home, Policy{}); err != nil {
			t.Fatal(err)
		}
	}
	req := directoryRequirement(f)
	snapshot := f.Snapshot
	if alternate, ok := tc.Changes["snapshot"].(string); ok {
		switch alternate {
		case "root_skill_snapshot":
			snapshot = f.RootSkill
		case "context_skill_snapshot":
			snapshot = f.ContextSkill
		default:
			t.Fatal(alternate)
		}
	}
	// The unchanged vendored CIP-0021 corpus predates the core §3 Markdown
	// scope fix. Keep its selected-package/capability/refusal contract, but place
	// its contact witness in agent-facing context instead of package-root notes.
	if finding, ok := tc.Expected["finding_file"].(string); ok && finding == "notes.md" {
		directory, _ := req["directory"].(string)
		source := filepath.ToSlash(filepath.Join(directory, finding))
		text, exists := snapshot.Files[source]
		if !exists {
			t.Fatal("published Markdown contact witness is missing")
		}
		visible := "references/" + finding
		snapshot.Files[filepath.ToSlash(filepath.Join(directory, visible))] = text
		delete(snapshot.Files, source)
		tc.Expected = maps.Clone(tc.Expected)
		tc.Expected["finding_file"] = visible
	}
	baseFiles := map[string]string{}
	for p, b := range snapshot.Files {
		baseFiles[p] = b
	}
	for key, value := range tc.Changes {
		switch {
		case strings.HasPrefix(key, "requirement."):
			field := strings.TrimPrefix(key, "requirement.")
			if field == "remove" {
				for _, k := range value.([]any) {
					delete(req, k.(string))
				}
			} else {
				req[field] = value
			}
		case key == "requirement":
			req = value.(map[string]any)
			f.ProfileManifest["requires"].(map[string]any)["skills"].(map[string]any)["formatter"] = req
		case strings.HasPrefix(key, "snapshot.files."):
			snapshot.Files[strings.TrimPrefix(key, "snapshot.files.")] = value.(string)
		case key == "snapshot.remove":
			for _, p := range value.([]any) {
				delete(snapshot.Files, p.(string))
			}
		case key == "snapshot.symlinks":
			for p, target := range value.(map[string]any) {
				snapshot.Symlinks[p] = target.(string)
			}
		}
	}
	// The parent-component witness also has a valid normalized target,
	// so grammar admission cannot be masked by a missing-directory error.
	if req["directory"] == "skills/../formatter" {
		for p, b := range snapshot.Files {
			if strings.HasPrefix(p, "skills/formatter/") {
				snapshot.Files[strings.TrimPrefix(p, "skills/")] = b
			}
		}
	}
	// Inside-link controls contain a valid selected package at their target.
	for link, target := range snapshot.Symlinks {
		destination := filepath.ToSlash(filepath.Clean(filepath.Join(filepath.Dir(link), target)))
		for p, b := range baseFiles {
			if strings.HasPrefix(p, link+"/") {
				delete(snapshot.Files, p)
				snapshot.Files[destination+strings.TrimPrefix(p, link)] = b
			}
		}
	}
	// update vectors preinstall the exact base, then publish a new immutable
	// skill and root commit; sync reuses its existing lock without re-resolving.
	initialFiles := snapshot.Files
	if tc.Entry == "profile update" {
		initialFiles = baseFiles
	}
	skillRepo := gitRepo(t, initialFiles, "v1.0.0")
	gitRun(t, skillRepo, "tag", "-d", "v1.0.0")
	ids.serve(skillRepo, f.Repository)
	ids.serve(skillRepo, strings.TrimSuffix(f.Repository, ".git"))
	commit := fixtureGitOutput(t, skillRepo, "rev-parse", "HEAD")
	if len(snapshot.Symlinks) > 0 {
		for link, target := range snapshot.Symlinks {
			c := exec.Command("git", "hash-object", "-w", "--stdin")
			c.Dir = skillRepo
			c.Stdin = strings.NewReader(target)
			out, err := c.Output()
			if err != nil {
				t.Fatal(err)
			}
			gitRun(t, skillRepo, "update-index", "--add", "--cacheinfo", "120000,"+strings.TrimSpace(string(out))+","+link)
		}
		gitRun(t, skillRepo, "commit", "-m", "link fixture")
		commit = fixtureGitOutput(t, skillRepo, "rev-parse", "HEAD")
	}
	req["revision"] = commit
	if _, ok := req["path"]; ok {
		delete(req, "revision")
	}
	if consumers, ok := tc.Changes["requirers"].(map[string]any); ok {
		delete(f.ProfileManifest["requires"].(map[string]any), "skills")
		contexts := map[string]any{}
		for name, v := range consumers {
			r := v.(map[string]any)
			r["revision"] = commit
			if r["git"] != f.Repository {
				ids.serve(skillRepo, strings.TrimSuffix(r["git"].(string), ".git"))
			}
			manifest := map[string]any{"schema_version": 2, "name": name, "version": "1.0.0", "requires": map[string]any{"skills": map[string]any{"formatter": r}}}
			repo := gitRepo(t, map[string]string{"agent-context.json": directoryJSON(t, manifest)}, "v1.0.0")
			operand := ids.serve(repo, "https://github.com/example/"+name)
			contexts[name] = map[string]any{"git": operand, "revision": fixtureGitOutput(t, repo, "rev-parse", "HEAD")}
		}
		f.ProfileManifest["requires"].(map[string]any)["contexts"] = contexts
	}
	rootRepo := gitRepo(t, map[string]string{"agent-context.json": directoryJSON(t, f.ProfileManifest)}, "v1.0.0")
	operand := ids.serve(rootRepo, "https://github.com/example/sample-profile.git")
	var info Info
	var err error
	if tc.Entry == "first profile-surface use" {
		global := filepath.Join(home, "global")
		writeGitFile(t, global, "Skillfile.json", directoryJSON(t, map[string]any{"schema_version": 1, "skills": []any{map[string]any{"name": "formatter", "git": f.Repository, "revision": commit}}}))
		err = EnsureDefaultWithPolicy(home, Policy{})
		if err == nil {
			info.Lock, info.LockHash, err = readLock(home, DefaultProfile)
			info.Name = DefaultProfile
		}
	} else {
		info, _, _, err = Install(home, InstallOptions{Operand: operand, Range: "*"})
		if err == nil && tc.Entry == "profile update" {
			oldRoot, readErr := os.ReadFile(lockPath(home, info.Name))
			if readErr != nil {
				t.Fatal(readErr)
			}
			before := directoryContentHash(t, home, info.Lock)
			for p, b := range snapshot.Files {
				writeGitFile(t, skillRepo, p, b)
			}
			gitRun(t, skillRepo, "add", "--all")
			gitRun(t, skillRepo, "commit", "--allow-empty", "-m", "updated skill")
			req["revision"] = fixtureGitOutput(t, skillRepo, "rev-parse", "HEAD")
			f.ProfileManifest["version"] = "1.0.1"
			writeGitFile(t, rootRepo, "agent-context.json", directoryJSON(t, f.ProfileManifest))
			gitRun(t, rootRepo, "add", "--all")
			gitRun(t, rootRepo, "commit", "-m", "updated profile")
			gitRun(t, rootRepo, "tag", "v1.0.1")
			info, _, err = UpdateWithPolicy(home, info.Name, Policy{})
			if err != nil {
				after, _ := os.ReadFile(lockPath(home, "sample-profile"))
				if !bytes.Equal(oldRoot, after) {
					t.Fatal("refused update changed lock")
				}
			}
			if err == nil {
				if equal, ok := tc.Expected["skill_content_sha256_equal_to_base"].(bool); ok && (before == directoryContentHash(t, home, info.Lock)) != equal {
					t.Fatal("selected content identity comparison mismatch")
				}
			}
		}
		if err == nil && tc.Entry == "profile sync" {
			before, errRead := os.ReadFile(lockPath(home, info.Name))
			if errRead != nil {
				t.Fatal(errRead)
			}
			_, err = SyncWithPolicy(home, Policy{})
			after, _ := os.ReadFile(lockPath(home, info.Name))
			if !bytes.Equal(before, after) {
				t.Fatal("sync changed lock")
			}
		}
	}
	if tc.Expected["outcome"] == "refused" {
		diagnostic := tc.Expected["diagnostic"].(string)
		if err == nil || !strings.Contains(err.Error(), diagnostic) {
			t.Fatalf("want refusal %s; got %v", diagnostic, err)
		}
		for _, field := range []string{"finding_file", "finding_host"} {
			if value, ok := tc.Expected[field].(string); ok && !strings.Contains(err.Error(), value) {
				t.Fatalf("refusal missing %s %s: %v", field, value, err)
			}
		}
		if required, ok := tc.Expected["diagnostic_requirers"].([]any); ok {
			for _, name := range required {
				if !strings.Contains(err.Error(), name.(string)) {
					t.Fatalf("missing requirer: %v", err)
				}
			}
			for _, field := range []string{"source", "directory"} {
				if !strings.Contains(err.Error(), field) {
					t.Fatalf("missing conflict field: %v", err)
				}
			}
		}
		if _, err := os.Stat(lockPath(home, "sample-profile")); !os.IsNotExist(err) {
			t.Fatalf("refusal published profile lock: %v", err)
		}
		for _, kind := range []string{"context", "skill"} {
			name := "sample-profile"
			if kind == "skill" {
				name = "formatter"
			}
			path := filepath.Join(contextstore.Root(home), kind, name)
			if entries, e := os.ReadDir(path); e == nil && len(entries) != 0 {
				t.Fatalf("refusal published store: %s", path)
			}
		}
		if _, err := os.Stat(sourcePath(home, "sample-profile")); !os.IsNotExist(err) {
			t.Fatalf("refusal published source: %v", err)
		}
		for _, path := range native {
			entries, e := os.ReadDir(path)
			if e == nil && len(entries) > 0 {
				t.Fatalf("refusal published native surface: %s", path)
			}
		}
		return
	}
	if err != nil {
		t.Fatalf("valid vector refused: %v", err)
	}
	member, ok := info.Lock.Find(contextlock.KindSkill, "formatter")
	if !ok {
		t.Fatal("formatter missing")
	}
	directory, _ := req["directory"].(string)
	if tc.Entry == "first profile-surface use" {
		directory = ""
	}
	if consumers, ok := tc.Changes["requirers"].(map[string]any); ok {
		directory, _ = consumers["consumer-a"].(map[string]any)["directory"].(string)
	}
	if member.Directory != directory {
		t.Fatalf("selected directory=%q want %q", member.Directory, directory)
	}
	if _, ok := tc.Expected["skill_lock_sha256"]; ok {
		expected, e := contextlock.Parse(f.CandidateLock)
		if e != nil {
			t.Fatal(e)
		}
		for i := range expected.Members {
			m := &expected.Members[i]
			if m.Kind == contextlock.KindContext {
				m.Commit = fixtureGitOutput(t, rootRepo, "rev-parse", "HEAD")
				m.Version = f.ProfileManifest["version"].(string)
			} else {
				m.Commit = req["revision"].(string)
			}
		}
		expectedPin, e := expected.Hash()
		if e != nil {
			t.Fatal(e)
		}
		if expectedPin != info.LockHash {
			t.Fatalf("substituted fixture lock identity mismatch: want %s got %s; lock=%s", expectedPin, info.LockHash, directoryJSON(t, info.Lock.Object()))
		}
	}
	if directory != "" && info.Lock.SchemaVersion != 4 {
		t.Fatalf("schema=%d want 4", info.Lock.SchemaVersion)
	}
	if _, ok := tc.Changes["requirers"]; ok {
		if !reflect.DeepEqual(member.RequiredBy, []string{"consumer-a", "consumer-b"}) {
			t.Fatalf("required_by=%v", member.RequiredBy)
		}
		n := 0
		for _, m := range info.Lock.Members {
			if m.Kind == contextlock.KindSkill {
				n++
			}
		}
		if n != 1 {
			t.Fatalf("skill count=%d", n)
		}
	}
	// Pin is computed from the actual emitted complete carrier, then round-trips.
	canonical, e := info.Lock.Canonical()
	if e != nil {
		t.Fatal(e)
	}
	if contextlock.HashBytes(canonical) != info.LockHash {
		t.Fatal("lock pin mismatch")
	}
	if _, e := contextlock.Parse(canonical); e != nil {
		t.Fatal(e)
	}
	root := packageRoot(newGitManager(home).entryPath(home, resolvedOf(member)), directory)
	if paths, ok := tc.Expected["skill_content_paths"].([]any); ok {
		var got []string
		err := filepath.WalkDir(root, func(p string, d os.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if !d.IsDir() {
				r, _ := filepath.Rel(root, p)
				got = append(got, filepath.ToSlash(r))
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		sort.Strings(got)
		want := []string{}
		for _, p := range paths {
			want = append(want, p.(string))
		}
		sort.Strings(want)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("content paths %v want %v", got, want)
		}
	}
	directoryAssertAudit(t, home, info.Lock, root, member)
	directoryAssertLinks(t, home, info.Name, native, root)
}
func directoryContentHash(t *testing.T, home string, lock *contextlock.Lock) string {
	t.Helper()
	m, ok := lock.Find(contextlock.KindSkill, "formatter")
	if !ok {
		t.Fatal("missing formatter")
	}
	root := packageRoot(newGitManager(home).entryPath(home, resolvedOf(m)), m.Directory)
	h, err := hashing.ContentSHA256WithVersion(root, nil, lock.ContentHashVersion())
	if err != nil {
		t.Fatal(err)
	}
	return h
}
func directoryAssertLinks(t *testing.T, home, profile string, native map[string]string, root string) {
	t.Helper()
	results, err := UseWithPolicy(home, profile, "codex_cli", "", false, Policy{})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if !r.OK {
			t.Fatalf("in-place materialization: %+v", r)
		}
	}
	// Directory selection feeds the canonical core projection, shared by native
	// and managed homes; sibling package bytes cannot become installed context.
	lock, _, err := readLock(home, profile)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := projectSkill(home, root, lock.ContentHashVersion())
	if err != nil {
		t.Fatal(err)
	}
	root = projection.root
	// NATSKILL places Codex skills directly in its native skills directory.
	nativeLink := filepath.Join(native["codex_cli"], "skills", "formatter")
	// Symlinks are preferred; Windows may use the protected fallback copy.
	check := func(path string) {
		t.Helper()
		target, e := os.Readlink(path)
		if e == nil {
			if !filepath.IsAbs(target) {
				target = filepath.Join(filepath.Dir(path), target)
			}
			if filepath.Clean(target) != filepath.Clean(root) {
				t.Fatalf("skill target=%s want %s", target, root)
			}
			return
		}
		got, e := os.ReadFile(filepath.Join(path, "SKILL.md"))
		want, we := os.ReadFile(filepath.Join(root, "SKILL.md"))
		if e != nil || we != nil || !bytes.Equal(got, want) {
			t.Fatalf("skill target=copy at %s has wrong root content: %v %v", path, e, we)
		}
	}
	check(nativeLink)
	req := ResolveRequest{Home: home, Profile: profile, EnvID: "codex_cli", LaunchDir: t.TempDir(), Repair: true, Machine: envregistry.DefaultMachineConfig(), Detect: func(envregistry.Adapter) string { return "unknown" }, NativeHomeOf: func(id string) (string, error) { return native[id], nil }, OperatorXDG: t.TempDir()}
	if _, err := Resolve(req); err != nil {
		t.Fatalf("managed materialization: %v", err)
	}
	check(filepath.Join(ManagedHomeDir(home, profile, "codex_cli"), "skills", "formatter"))
}

// Persisted audit evidence must bind the selected content, directory and
// capability declaration. Root/sibling files must not supply findings.
func directoryAssertAudit(t *testing.T, home string, lock *contextlock.Lock, root string, member contextlock.Member) {
	t.Helper()
	hash := directoryContentHash(t, home, lock)
	found := false
	err := filepath.WalkDir(home, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasPrefix(d.Name(), "verdict-") || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		var r struct {
			Skill, Commit, Directory string
			Hash                     string `json:"content_sha256"`
			Capabilities             struct{ Network []string }
			Findings                 []struct{ ID, File, Evidence string }
		}
		if e = json.Unmarshal(b, &r); e != nil {
			return e
		}
		if r.Skill != "formatter" || r.Commit != member.Commit {
			return nil
		}
		found = true
		directory := member.Directory
		if directory == "" {
			directory = "."
		}
		if r.Directory != directory || r.Hash != hash {
			t.Fatalf("audit identity does not bind selected tree: %s", b)
		}
		manifest, e := os.ReadFile(filepath.Join(root, "agent-skill.json"))
		if e == nil && bytes.Contains(manifest, []byte("api.example.net")) && !reflect.DeepEqual(r.Capabilities.Network, []string{"api.example.net"}) {
			t.Fatalf("capabilities not read from selected manifest: %s", b)
		}
		for _, f := range r.Findings {
			if f.ID == "audit.capability.network-undeclared" || strings.HasPrefix(f.File, "skills/") || strings.Contains(f.Evidence, "other.example.net") {
				t.Fatalf("root or sibling contributes a finding: %+v", f)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("no selected-skill audit verdict persisted")
	}
}
