package envprofile

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
)

// Production entry points under test: Import, List. Native homes and the
// global skills surface are pinned through the import seams, so no test
// touches the operator's real homes.

// importSeams pins every detected surface below temporary directories:
// one native home per adapter id and one global skills surface.
type importSeams struct {
	native map[string]string
	global string
}

func pinImportSeams(t *testing.T) importSeams {
	t.Helper()
	seams := importSeams{native: map[string]string{}, global: filepath.Join(t.TempDir(), "global-skills")}
	for _, id := range []string{"claude_code", "codex_cli", "opencode", "pi"} {
		seams.native[id] = filepath.Join(t.TempDir(), id)
	}
	return seams
}

func (s importSeams) options(policy Policy) ImportOptions {
	return ImportOptions{
		Policy: policy,
		NativeHomeOf: func(id string) (string, error) {
			return s.native[id], nil
		},
		GlobalSkillsOf: func() (string, error) {
			return s.global, nil
		},
	}
}

func writeNativeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// seedCurrentDefault records the builtin default as machine current so an
// import installs without attempting activation: the reassembly and lock
// surface without the switch.
func seedCurrentDefault(t *testing.T, home string) {
	t.Helper()
	if err := SetCurrent(home, DefaultProfile); err != nil {
		t.Fatal(err)
	}
}

// TestImportLosslessInstallsThroughPathPipeline drives the production
// Import over one detected root-context file: the profile installs named
// imported at version 1.0.0 through the ordinary path pipeline —
// state-hash pin, strict audit, lock — and the native file is untouched.
func TestImportLosslessInstallsThroughPathPipeline(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	seams := pinImportSeams(t)
	writeNativeFile(t, seams.native["claude_code"], "CLAUDE.md", "hello\n")
	seedCurrentDefault(t, home)
	info, activated, _, err := Import(home, seams.options(Policy{}))
	if err != nil {
		t.Fatal(err)
	}
	if activated {
		t.Fatal("import without --use on a machine with a current profile must not activate")
	}
	if info.Name != "imported" {
		t.Fatalf("profile %q, want imported", info.Name)
	}
	root, ok := info.Lock.RootMember()
	if !ok || root.Version != "1.0.0" {
		t.Fatalf("root %+v, want version 1.0.0", root)
	}
	if root.Name != "imported" {
		t.Fatalf("root %q, want imported", root.Name)
	}
	if len(info.Lock.Members) != 1 {
		t.Fatalf("lock members %+v, want the root alone", info.Lock.Members)
	}
	source, err := readSource(home, "imported")
	if err != nil {
		t.Fatal(err)
	}
	if source.Kind != KindPath || !source.ImportedFromNative {
		t.Fatalf("source %+v, want an imported path record", source)
	}
	if payload, _ := os.ReadFile(filepath.Join(seams.native["claude_code"], "CLAUDE.md")); string(payload) != "hello\n" {
		t.Fatal("the import writes nothing into any native home by itself")
	}
	if _, err := os.Stat(filepath.Join(seams.native["claude_code"], envmarker.Name)); !os.IsNotExist(err) {
		t.Fatal("import without activation writes no marker")
	}
}

// TestImportNormalizesOrdersAndPins drives Import over CRLF/CR root files
// on two adapters: reassembly normalizes to LF with exactly one trailing
// LF, lists modules in ascending environment-identifier order, and stores
// the exact bytes.
func TestImportNormalizesOrdersAndPins(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	seams := pinImportSeams(t)
	writeNativeFile(t, seams.native["codex_cli"], "AGENTS.md", "codex\r\nline\r\n")
	writeNativeFile(t, seams.native["claude_code"], "CLAUDE.md", "claude\rline\n\n")
	seedCurrentDefault(t, home)
	info, _, _, err := Import(home, seams.options(Policy{}))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lockMember(info.Lock, "imported")
	if !ok || member.StateHash == "" {
		t.Fatalf("lock members %+v carry no state pin", info.Lock.Members)
	}
	entry := contextstore.EntryDir(home, "context", "imported", member.StateHash)
	for _, tc := range []struct{ file, want string }{
		{"context/claude_code.md", "claude\nline\n"},
		{"context/codex_cli.md", "codex\nline\n"},
	} {
		payload, err := os.ReadFile(filepath.Join(entry, filepath.FromSlash(tc.file))) // #nosec G304 -- store entry file
		if err != nil || string(payload) != tc.want {
			t.Fatalf("%s = %q, err = %v; want %q", tc.file, payload, err, tc.want)
		}
	}
	manifest, err := os.ReadFile(filepath.Join(entry, "agent-context.json")) // #nosec G304 -- store entry file
	if err != nil {
		t.Fatal(err)
	}
	claude := strings.Index(string(manifest), "claude_code.md")
	codex := strings.Index(string(manifest), "codex_cli.md")
	if claude < 0 || codex < 0 || claude > codex {
		t.Fatalf("modules are not in ascending environment order:\n%s", manifest)
	}
}

// TestNormalizeImportText checks the section 9.6 normalization exactly:
// every CRLF and bare CR becomes LF and the content ends with exactly one
// trailing LF, applied only here — snapshot modules are never normalized.
func TestNormalizeImportText(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"a\r\nb\r\n", "a\nb\n"},
		{"a\rb\r", "a\nb\n"},
		{"a\nb\n\n\n", "a\nb\n"},
		{"a", "a\n"},
		{"", "\n"},
		{"a\rm\nb\r\nc", "a\nm\nb\nc\n"},
	} {
		if got := string(normalizeImportText([]byte(tc.in))); got != tc.want {
			t.Fatalf("normalize(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestImportLossyStopsWithoutConsent drives Import over an unrecoverable
// skills entry: the lossy classification stops with
// environment_import_lossy and the loss list, installing nothing.
func TestImportLossyStopsWithoutConsent(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	seams := pinImportSeams(t)
	writeNativeFile(t, seams.native["claude_code"], "CLAUDE.md", "hello\n")
	skills := filepath.Join(seams.native["claude_code"], "skills", "mystery")
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, "SKILL.md"), []byte("# mystery\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	seedCurrentDefault(t, home)
	_, _, _, err := Import(home, seams.options(Policy{}))
	if err == nil || !strings.Contains(err.Error(), DiagImportLossy) {
		t.Fatalf("err = %v, want %s", err, DiagImportLossy)
	}
	if !strings.Contains(err.Error(), "mystery") {
		t.Fatalf("err = %v, want the loss list naming the entry", err)
	}
	if _, statErr := readSource(home, "imported"); statErr == nil {
		t.Fatal("a stopped import installs nothing")
	}
}

// TestImportLossyProceedsWithConsent drives the same classification with
// the per-operation consent flag: the import proceeds and re-reports the
// loss list as warnings under the same diagnostic.
func TestImportLossyProceedsWithConsent(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	seams := pinImportSeams(t)
	writeNativeFile(t, seams.native["claude_code"], "CLAUDE.md", "hello\n")
	skills := filepath.Join(seams.native["claude_code"], "skills", "mystery")
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, "SKILL.md"), []byte("# mystery\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	seedCurrentDefault(t, home)
	options := seams.options(Policy{})
	options.AllowLossy = true
	info, _, _, err := Import(home, options)
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "imported" {
		t.Fatalf("profile %q", info.Name)
	}
	found := false
	for _, warning := range info.Warnings {
		if strings.Contains(warning, DiagImportLossy) && strings.Contains(warning, "mystery") {
			found = true
		}
	}
	if !found {
		t.Fatalf("warnings %+v re-report no loss list", info.Warnings)
	}
}

// TestImportUnreadableRootIsLoss drives Import over a native root-context
// file that cannot be read: the failed read is always a loss, never an
// absence.
func TestImportUnreadableRootIsLoss(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	seams := pinImportSeams(t)
	writeNativeFile(t, seams.native["claude_code"], "CLAUDE.md", "hello\n")
	root := filepath.Join(seams.native["claude_code"], "CLAUDE.md")
	// The probe below decides: a host that reads through mode 000
	// (superuser, or Windows ACL semantics) skips under the classified
	// host-capability reason instead of asserting untestable behaviour.
	if err := os.Chmod(root, 0o000); err != nil {
		t.Skipf("this environment can read a mode-000 directory: chmod refused: %v", err)
	}
	defer func() { _ = os.Chmod(root, 0o644) }()
	if _, err := os.ReadFile(root); err == nil {
		t.Skip("this environment can read a mode-000 directory; unreadability is untestable here")
	}
	seedCurrentDefault(t, home)
	_, _, _, err := Import(home, seams.options(Policy{}))
	if err == nil || !strings.Contains(err.Error(), DiagImportLossy) {
		t.Fatalf("err = %v, want %s", err, DiagImportLossy)
	}
	if strings.Contains(err.Error(), "never") {
		t.Fatalf("err = %v, a failed read must never report absence", err)
	}
}

// TestImportInvalidUTF8IsLoss drives Import over a native root-context
// file that is not valid UTF-8: normalization is content-preserving, so
// the file is a loss.
func TestImportInvalidUTF8IsLoss(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	seams := pinImportSeams(t)
	writeNativeFile(t, seams.native["claude_code"], "CLAUDE.md", "hello\xff\n")
	seedCurrentDefault(t, home)
	_, _, _, err := Import(home, seams.options(Policy{}))
	if err == nil || !strings.Contains(err.Error(), DiagImportLossy) {
		t.Fatalf("err = %v, want %s", err, DiagImportLossy)
	}
	if !strings.Contains(err.Error(), "UTF-8") {
		t.Fatalf("err = %v, want the UTF-8 reason", err)
	}
}

// TestImportPartialActivationKeepsLock drives Import on a fresh machine
// whose native home already carries the detected file: the ordinary path
// pipeline installs the lock, and the activation it attempts under the
// section 9.1 rules stops on the unmanaged file without --takeover —
// converging it remains the takeover switch's job.
func TestImportPartialActivationKeepsLock(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	seams := pinImportSeams(t)
	writeNativeFile(t, seams.native["claude_code"], "CLAUDE.md", "hello\n")
	t.Setenv("CLAUDE_CONFIG_DIR", seams.native["claude_code"])
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex"))
	t.Setenv("PI_CODING_AGENT_DIR", filepath.Join(t.TempDir(), "pi"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "xdg"))
	_, activated, _, err := Import(home, seams.options(Policy{}))
	// The activation stops per entry with the unmanaged conflict and
	// the switch reports the ordinary partial diagnostic.
	if err == nil || !strings.Contains(err.Error(), DiagUsePartial) {
		t.Fatalf("err = %v, want %s", err, DiagUsePartial)
	}
	if activated {
		t.Fatal("a conflicted activation must not report activated")
	}
	if _, statErr := readSource(home, "imported"); statErr != nil {
		t.Fatal("the ordinary pipeline keeps the installed lock past a conflicted activation")
	}
}

// TestImportNameTakenBeforeAnyWrite drives Import with a chosen name that
// is already installed: the import stops with profile_import_name_taken
// before any write, leaving the installed profile alone.
func TestImportNameTakenBeforeAnyWrite(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	seams := pinImportSeams(t)
	writeNativeFile(t, seams.native["claude_code"], "CLAUDE.md", "hello\n")
	source := filepath.Join(t.TempDir(), "source")
	writeManifestPackage(t, source,
		`{"schema_version": 1, "name": "legacy", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": "legacy\n"})
	before, _, _, err := Install(home, InstallOptions{Operand: source, As: "imported"})
	if err != nil {
		t.Fatal(err)
	}
	options := seams.options(Policy{})
	options.As = "imported"
	_, _, _, err = Import(home, options)
	if err == nil || !strings.Contains(err.Error(), DiagImportNameTaken) {
		t.Fatalf("err = %v, want %s", err, DiagImportNameTaken)
	}
	lock, hash, err := readLock(home, "imported")
	if err != nil || hash != before.LockHash || lock.Root != "legacy" {
		t.Fatal("a refused import must leave the installed profile alone")
	}
}

// gitSkillCheckout clones a served skill repository below dir as name with
// the fake network identity as its origin and a clean HEAD, returning the
// resolved commit.
func gitSkillCheckout(t *testing.T, ids *gitIdentities, dir, name, identity string) string {
	t.Helper()
	repo := gitRepo(t, map[string]string{"SKILL.md": "# " + name + "\n"}, "v1.0.0")
	operand := ids.serve(repo, identity)
	entry := filepath.Join(dir, name)
	cmd := exec.Command("git", "clone", repo, entry)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("clone: %v\n%s", err, out)
	}
	for _, args := range [][]string{
		{"remote", "set-url", "origin", operand},
		{"config", "user.email", "t@t"},
		{"config", "user.name", "t"},
	} {
		cmd := exec.Command("git", append([]string{"-C", entry}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	out, err := exec.Command("git", "-C", entry, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

// TestImportSkillForeignPinnedByRevision drives Import over a skills entry
// kept as a clean git checkout: the entry maps onto one requires.skills
// entry pinned by revision, installs through the path pipeline, and warns
// environment_import_skill_foreign.
func TestImportSkillForeignPinnedByRevision(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	seams := pinImportSeams(t)
	writeNativeFile(t, seams.native["claude_code"], "CLAUDE.md", "hello\n")
	ids := newGitIdentities(t)
	skills := filepath.Join(seams.native["claude_code"], "skills")
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	commit := gitSkillCheckout(t, ids, skills, "foreign", "https://example.com/skills/foreign")
	seedCurrentDefault(t, home)
	info, _, _, err := Import(home, seams.options(Policy{}))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lockMember(info.Lock, "foreign")
	if !ok {
		t.Fatalf("lock members %+v carry no imported skill", info.Lock.Members)
	}
	if member.Kind != "skill" || member.Commit != commit || member.Source != "example.com/skills/foreign" {
		t.Fatalf("skill member %+v, want the checkout HEAD pinned", member)
	}
	found := false
	for _, warning := range info.Warnings {
		if strings.Contains(warning, DiagImportSkillForeign) && strings.Contains(warning, "foreign") {
			found = true
		}
	}
	if !found {
		t.Fatalf("warnings %+v carry no %s", info.Warnings, DiagImportSkillForeign)
	}
}

// TestImportSkillMarkerRecovery drives Import over a skills entry carrying
// a valid install marker: the marker's source identity and resolved commit
// recover the exact declaration pinned by revision.
func TestImportSkillMarkerRecovery(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	seams := pinImportSeams(t)
	writeNativeFile(t, seams.native["claude_code"], "CLAUDE.md", "hello\n")
	ids := newGitIdentities(t)
	repo := gitRepo(t, map[string]string{"SKILL.md": "# marked\n"}, "v1.0.0")
	operand := ids.serve(repo, "https://example.com/skills/marked")
	out, err := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	head := strings.TrimSpace(string(out))
	entry := filepath.Join(seams.native["claude_code"], "skills", "marked")
	if err := os.MkdirAll(entry, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(entry, "SKILL.md"), []byte("# marked\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sha := strings.Repeat("0", 64)
	markerJSON := `{"schema_version": 1, "name": "marked", "source": "skills/marked",` +
		` "ref_kind": "tag", "ref": "v1.0.0", "commit": "` + head + `",` +
		` "content_sha256": "sha256:` + sha + `", "locale": null,` +
		` "agents": [], "commands": [], "dependencies": [], "skill_schema_version": 1,` +
		` "runtime_roots": [], "installed_at": "2026-01-02T15:04:05Z", "files": ["SKILL.md"],` +
		` "git": "` + operand + `"}` + "\n"
	if err := os.WriteFile(filepath.Join(entry, ".csk-install.json"), []byte(markerJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	seedCurrentDefault(t, home)
	info, _, _, err := Import(home, seams.options(Policy{}))
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lockMember(info.Lock, "marked")
	if !ok {
		t.Fatalf("lock members %+v carry no imported skill", info.Lock.Members)
	}
	if member.Kind != "skill" || member.Commit != head || member.Source != "example.com/skills/marked" {
		t.Fatalf("skill member %+v, want the marker's resolved commit pinned", member)
	}
}

// TestImportRecordsImportedFromNative drives Import to an installed lock
// and then takes over the switch: the environment markers record
// imported_from_native and the profile list reports the imported marker.
func TestImportRecordsImportedFromNative(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	seams := pinImportSeams(t)
	writeNativeFile(t, seams.native["claude_code"], "CLAUDE.md", "hello\n")
	seedCurrentDefault(t, home)
	if _, _, _, err := Import(home, seams.options(Policy{})); err != nil {
		t.Fatal(err)
	}
	// The detected file is unmanaged in the fresh native home, so only a
	// takeover switch converges it; the marker then records the import.
	t.Setenv("CLAUDE_CONFIG_DIR", seams.native["claude_code"])
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex"))
	t.Setenv("PI_CODING_AGENT_DIR", filepath.Join(t.TempDir(), "pi"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "xdg"))
	results, err := UseWithPolicy(home, "imported", "", "", false, Policy{Takeover: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range results {
		if !result.OK {
			t.Fatalf("%s: %s", result.Adapter, result.Detail)
		}
	}
	marker, err := envmarker.Read(seams.native["claude_code"])
	if err != nil || marker == nil {
		t.Fatalf("marker %v", err)
	}
	if !marker.Profile.ImportedFromNative {
		t.Fatalf("marker %+v records no imported_from_native", marker.Profile)
	}
	if err := marker.Validate(); err != nil {
		t.Fatalf("imported marker invalid: %v", err)
	}
	profiles, err := List(home)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, info := range profiles {
		if info.Name == "imported" {
			found = true
			if !info.Source.ImportedFromNative {
				t.Fatalf("listed source %+v carries no import record", info.Source)
			}
		}
	}
	if !found {
		t.Fatal("imported profile is not listed")
	}
}
