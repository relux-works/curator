package envprofile

// Onboarding import (environments §9.6): the detected native context
// becomes one installed profile through the ordinary path pipeline of
// section 9.1.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/marker"
)

// envAdapterByID resolves the registry adapter for import skills-surface
// resolution; a registered switch adapter is always registered there.
func envAdapterByID(id string) (envregistry.Adapter, error) {
	return envregistry.ByID(id)
}

// Diagnostics for the onboarding import (environments §9.6, §9.7).
const (
	DiagImportLossy        = "environment_import_lossy"
	DiagImportSkillForeign = "environment_import_skill_foreign"
)

// ImportProfileDefault is the profile name an import takes unless the
// operator supplies one under the core §2 grammar (environments §9.6).
const ImportProfileDefault = "imported"

// ImportOptions selects one onboarding import.
type ImportOptions struct {
	// As names the profile; empty takes ImportProfileDefault.
	As string
	// AllowLossy is the explicit per-operation consent flag: a lossy
	// import proceeds only under it, re-reporting the loss list as
	// warnings. Machine configuration never pre-records consent, so no
	// policy knob feeds this field.
	AllowLossy bool
	// Use activates the installed profile under the section 9.1 rules.
	Use    bool
	Policy Policy
	// NativeHomeOf resolves native homes; nil means the process homes. A
	// test seam: production never sets it.
	NativeHomeOf func(id string) (string, error)
	// GlobalSkillsOf resolves the machine-global native skills surface
	// (the opencode skills target, §7.1); nil resolves it below the
	// operator home. A test seam: production never sets it.
	GlobalSkillsOf func() (string, error)
}

// ImportLoss is one detected surface that maps onto no supported surface
// (environments §9.6): the adapter, the platform path, and the reason. An
// absent surface is never a loss; a failed read always is (§8.4).
type ImportLoss struct {
	Adapter string
	Path    string
	Reason  string
}

func (l ImportLoss) String() string {
	return l.Adapter + " " + l.Path + ": " + l.Reason
}

// detectedRoot is one detected native root-context file.
type detectedRoot struct {
	envID string
	data  []byte
}

// detectedSkill is one skills entry with a recovered exact declaration,
// pinned by revision to its resolved commit.
type detectedSkill struct {
	envID    string
	name     string
	git      string
	revision string
	path     string
}

// Import turns the section 9.5 inventory into an installed profile through
// the ordinary path pipeline: detect, classify, gate on consent,
// reassemble, and install with the always-strict audit. The import writes
// nothing into any native home by itself; replacing native files remains
// the section 9.5 takeover path with its notice and backup. A chosen name
// that is already installed stops the import with
// profile_import_name_taken before any write.
func Import(home string, options ImportOptions) (Info, bool, bool, error) {
	op, err := beginOperation(home)
	if err != nil {
		return Info{}, false, false, err
	}
	defer func() { _ = op.close() }()
	return importLocked(op, home, options)
}

func importLocked(op *operation, home string, options ImportOptions) (Info, bool, bool, error) {
	if err := ensureDefault(op, home, options.Policy); err != nil {
		return Info{}, false, false, err
	}
	name := options.As
	if name == "" {
		name = ImportProfileDefault
	}
	if !validProfileName(name) {
		return Info{}, false, false, fmt.Errorf("%s: profile name %q is not a portable identifier", DiagSourceInvalid, name)
	}
	if _, err := readSource(home, name); err == nil {
		return Info{}, false, false, fmt.Errorf("%s: profile %q is already installed", DiagImportNameTaken, name)
	}
	roots, skills, losses := detectNative(home, options)
	// A mapping skills entry that cannot be represented in the single
	// requires.skills map is a loss, not a silent collapse (§9.6: one
	// entry per mapping entry, and the loss list names every loss). Two
	// adapters carrying the same skill name at the same commit and source
	// collapse to one entry with no loss — one declaration, not two. Two
	// adapters carrying the same name at different commits or sources are
	// divergent: the first by ascending environment identifier wins the
	// entry and each dropped declaration joins the loss list, making the
	// import lossy. The deciding sentence is the reassembly one — "one
	// requires.skills entry per mapping skills entry" — which a JSON
	// object cannot satisfy for duplicate keys, so the excess mapping
	// entry is an unmappable detected surface, hence a loss.
	skills, dupLosses := deduplicateSkills(skills)
	losses = append(losses, dupLosses...)
	if len(losses) > 0 && !options.AllowLossy {
		lines := make([]string, 0, len(losses))
		for _, loss := range losses {
			lines = append(lines, "- "+loss.String())
		}
		return Info{}, false, false, fmt.Errorf("%s: %d detected surfaces cannot be carried over:\n%s",
			DiagImportLossy, len(losses), strings.Join(lines, "\n"))
	}
	var warnings []string
	for _, loss := range losses {
		warnings = append(warnings, DiagImportLossy+": "+loss.String())
	}
	for _, skill := range skills {
		warnings = append(warnings, DiagImportSkillForeign+": "+skill.name+
			" was managed by other means; re-declare it from its upstream source")
	}
	stage, err := reassembleImport(home, name, roots, skills)
	if err != nil {
		return Info{}, false, false, err
	}
	defer func() { _ = os.RemoveAll(stage) }()
	installOptions := InstallOptions{
		Operand: stage, As: name, Use: options.Use,
		Policy: options.Policy, Imported: true,
	}
	info, activated, updated, err := installLocked(op, home, installOptions)
	if err != nil {
		return Info{}, false, false, err
	}
	info.Warnings = append(warnings, info.Warnings...)
	return info, activated, updated, nil
}

// importNativeHome resolves the native default home of one environment.
func importNativeHome(options ImportOptions, id string) (string, error) {
	if options.NativeHomeOf != nil {
		return options.NativeHomeOf(id)
	}
	adapter, ok := adapterByID(id)
	if !ok {
		return "", fmt.Errorf("%s: unregistered environment %q", DiagUnknownEnvironment, id)
	}
	return NativeHome(adapter)
}

// importSkillsDir resolves the skills surface one adapter's detected
// entries are read from: the home-local skills directory, or the
// machine-global native surface where the adapter declares none
// (opencode, split-brain by construction, §7.1).
func importSkillsDir(options ImportOptions, id, native string) (string, error) {
	adapter, err := envAdapterByID(id)
	if err != nil {
		return "", err
	}
	if adapter.SkillsDir != "" {
		return filepath.Join(native, adapter.SkillsDir), nil
	}
	if options.GlobalSkillsOf != nil {
		return options.GlobalSkillsOf()
	}
	operator, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("no home directory for the global skills surface: %v", err)
	}
	return filepath.Join(operator, ".agents", "skills"), nil
}

// detectNative inventories the closed revision-1 detected-surface list
// (environments §9.6): per registered adapter over its native default
// home, the root-context file and every unledgered skills entry. A
// surface that is absent is simply not detected.
func detectNative(home string, options ImportOptions) ([]detectedRoot, []detectedSkill, []ImportLoss) {
	var roots []detectedRoot
	var skills []detectedSkill
	var losses []ImportLoss
	for _, adapter := range Adapters {
		native, err := importNativeHome(options, adapter.ID)
		if err != nil || native == "" {
			continue
		}
		rootPath := filepath.Join(native, adapter.Target)
		if data, loss := readRootSurface(adapter.ID, native, rootPath); loss != nil {
			losses = append(losses, *loss)
		} else if data != nil {
			roots = append(roots, detectedRoot{envID: adapter.ID, data: data})
		}
		skillsDir, err := importSkillsDir(options, adapter.ID, native)
		if err != nil || skillsDir == "" {
			continue
		}
		mapped, dirLosses := readSkillsSurface(home, adapter.ID, skillsDir)
		skills = append(skills, mapped...)
		losses = append(losses, dirLosses...)
	}
	sort.Slice(roots, func(i, j int) bool { return roots[i].envID < roots[j].envID })
	sort.Slice(skills, func(i, j int) bool {
		if skills[i].envID != skills[j].envID {
			return skills[i].envID < skills[j].envID
		}
		return skills[i].name < skills[j].name
	})
	sort.Slice(losses, func(i, j int) bool {
		if losses[i].Adapter != losses[j].Adapter {
			return losses[i].Adapter < losses[j].Adapter
		}
		return losses[i].Path < losses[j].Path
	})
	return roots, skills, losses
}

// deduplicateSkills collapses same-named skills across adapters without
// silent loss. Skills arrive sorted by ascending environment identifier.
// The first declaration of a name wins the requires.skills entry; an
// identical repeat (same git and revision) is the same declaration and
// carries no loss; a divergent repeat (different commit or source) cannot
// be represented beside the winner in one JSON object, so each dropped
// declaration joins the loss list.
func deduplicateSkills(skills []detectedSkill) ([]detectedSkill, []ImportLoss) {
	seen := map[string]detectedSkill{}
	var kept []detectedSkill
	var losses []ImportLoss
	for _, skill := range skills {
		winner, ok := seen[skill.name]
		if !ok {
			seen[skill.name] = skill
			kept = append(kept, skill)
			continue
		}
		if winner.git == skill.git && winner.revision == skill.revision {
			continue
		}
		losses = append(losses, ImportLoss{
			Adapter: skill.envID, Path: skill.path,
			Reason: fmt.Sprintf("duplicate skill %q diverges from %s at %s; only the first declaration carries over", skill.name, winner.envID, winner.revision),
		})
	}
	sort.Slice(losses, func(i, j int) bool {
		if losses[i].Adapter != losses[j].Adapter {
			return losses[i].Adapter < losses[j].Adapter
		}
		return losses[i].Path < losses[j].Path
	})
	return kept, losses
}

// readRootSurface detects one native root-context file: absent is not
// detected; an unreadable file and a file that is not valid UTF-8 are
// losses. Reassembly normalization is content-preserving and never makes
// an import lossy. A curator-managed home is never native input: when the
// native directory carries a valid environment marker, its root-context
// file reaches managed state through §9.2, never through import (§9.5
// step 1 inventories unmanaged files, and §9.6's detected-surface list is
// that inventory), so it is skipped — neither detected nor a loss. Absence
// and a failed read stay different facts (§8.4): no marker file is absence
// and detects the surface; a marker that exists but cannot be read or
// decoded is a failed read of the evidence that classifies the surface, so
// the surface is neither detected nor silently dropped — the marker file
// itself joins the loss list and the consent gate decides.
func readRootSurface(envID, native, path string) ([]byte, *ImportLoss) {
	marker, err := envmarker.Read(native)
	if err != nil {
		return nil, &ImportLoss{Adapter: envID, Path: filepath.Join(native, envmarker.Name), Reason: "cannot be read: " + err.Error()}
	}
	if marker != nil {
		return nil, nil
	}
	payload, err := os.ReadFile(path) // #nosec G304 -- native home file named by the adapter registry
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, &ImportLoss{Adapter: envID, Path: path, Reason: "cannot be read: " + err.Error()}
	}
	if !utf8.Valid(payload) {
		return nil, &ImportLoss{Adapter: envID, Path: path, Reason: "is not valid UTF-8"}
	}
	return payload, nil
}

// readSkillsSurface detects the unledgered skills entries of one surface:
// a ledgered entry belongs to the machine-global scope and reaches
// managed state through the section 9.4 migration, never through import.
// An entry with no recoverable exact declaration is a loss. A ledger
// that exists but cannot be read or decoded is never treated as absence
// (§8.4): the manager cannot prove any entry unledgered, so the surface
// contributes no skills and the ledger file itself is the loss.
func readSkillsSurface(home, envID, dir string) ([]detectedSkill, []ImportLoss) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, []ImportLoss{{Adapter: envID, Path: dir, Reason: "cannot be read: " + err.Error()}}
	}
	ledgered, ledgerErr := readSkillsLedger(dir)
	if ledgerErr != nil {
		ledgerPath := filepath.Join(dir, ".csk-managed.json")
		return nil, []ImportLoss{{Adapter: envID, Path: ledgerPath, Reason: "cannot be read: " + ledgerErr.Error()}}
	}
	var skills []detectedSkill
	var losses []ImportLoss
	for _, entry := range entries {
		name := entry.Name()
		if name == ".csk-managed.json" {
			continue
		}
		if ledgered[name] {
			continue
		}
		full := filepath.Join(dir, name)
		if isStoreEntry(home, full) {
			continue
		}
		if !identifiers.Valid(name) {
			losses = append(losses, ImportLoss{Adapter: envID, Path: full, Reason: "is not a portable identifier"})
			continue
		}
		git, revision, ok := recoverSkill(full)
		if !ok {
			losses = append(losses, ImportLoss{Adapter: envID, Path: full, Reason: "names no recoverable exact declaration"})
			continue
		}
		skills = append(skills, detectedSkill{envID: envID, name: name, git: git, revision: revision, path: full})
	}
	return skills, losses
}

// readSkillsLedger returns the skill names the manager's adapter ledger
// records at the surface root. Absence (no ledger file) is legitimate and
// returns an empty set with no error; any other read or decode failure is
// a failed read, never absence (§8.4), and returns the error so the caller
// records the ledger itself as the loss instead of importing entries it
// cannot prove unledgered.
func readSkillsLedger(dir string) (map[string]bool, error) {
	recorded := map[string]bool{}
	payload, err := os.ReadFile(filepath.Join(dir, ".csk-managed.json")) // #nosec G304 -- ledger beside the scanned surface
	if err != nil {
		if os.IsNotExist(err) {
			return recorded, nil
		}
		return nil, err
	}
	var data struct {
		SchemaVersion int      `json:"schema_version"`
		Entries       []string `json:"entries"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}
	if data.SchemaVersion != 1 || data.Entries == nil {
		return nil, fmt.Errorf("adapter ledger is not a schema_version 1 entry list")
	}
	for _, entry := range data.Entries {
		if identifiers.Valid(entry) {
			recorded[entry] = true
		}
	}
	return recorded, nil
}

// isStoreEntry reports whether path is a symlink into the manager's
// immutable profile store: manager-staged content the section 9.4
// migration owns, never import input.
func isStoreEntry(home, path string) bool {
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		return false
	}
	target, err := os.Readlink(path)
	if err != nil {
		return false
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}
	return sameStoreTree(target, contextstore.Root(home))
}

// recoverSkill recovers a complete exact declaration from the entry's own
// records: a valid install marker recording the source identity, declared
// ref, and resolved commit, or a git checkout whose origin canonicalizes
// and whose HEAD carries no staged, dirty, or untracked bytes.
func recoverSkill(entry string) (string, string, bool) {
	if m := marker.Read(entry); m != nil {
		source := m.Git
		if source == "" {
			source = m.Source
		}
		if source != "" && m.Ref != "" && isHexCommit(m.Commit) {
			if canonical, err := canonicalGit(source); err == nil && canonical != "" {
				return canonical, m.Commit, true
			}
		}
	}
	if _, err := os.Stat(filepath.Join(entry, ".git")); err != nil {
		return "", "", false
	}
	// The configured origin, not `remote get-url`: the latter applies
	// insteadOf rewrites and would report the transport rather than the
	// declared identity the recovery must canonicalize.
	origin, err := gitOutput(entry, "config", "--get", "remote.origin.url")
	if err != nil {
		return "", "", false
	}
	canonical, err := canonicalGit(strings.TrimSpace(origin))
	if err != nil || canonical == "" {
		return "", "", false
	}
	head, err := gitOutput(entry, "rev-parse", "HEAD")
	if err != nil || !isHexCommit(strings.TrimSpace(head)) {
		return "", "", false
	}
	status, err := gitOutput(entry, "status", "--porcelain")
	if err != nil || strings.TrimSpace(status) != "" {
		return "", "", false
	}
	return canonical, strings.TrimSpace(head), true
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...) // #nosec G204 -- fixed git argv over the scanned entry
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func isHexCommit(commit string) bool {
	if len(commit) != 40 && len(commit) != 64 {
		return false
	}
	for _, c := range commit {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// normalizeImportText applies the section 9.6 reassembly normalization
// exactly: every CRLF and bare-CR line ending becomes LF, and the content
// ends with exactly one trailing LF.
func normalizeImportText(data []byte) []byte {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return []byte(strings.TrimRight(text, "\n") + "\n")
}

// reassembleImport assembles the context-package-shaped directory inside
// the machine home: agent-context.json (schema_version 1, version 1.0.0,
// weight 0, no weights), one normalized module per detected root-context
// file in ascending environment-identifier order, and one requires.skills
// entry per mapping entry pinned by revision. An import with no detected
// root-context file emits no context member.
func reassembleImport(home, name string, roots []detectedRoot, skills []detectedSkill) (string, error) {
	stage, err := os.MkdirTemp(home, ".import-*")
	if err != nil {
		return "", err
	}
	manifest := map[string]any{
		"schema_version": 1,
		"name":           name,
		"version":        "1.0.0",
		"weight":         0,
	}
	if len(roots) > 0 {
		modules := make([]any, 0, len(roots))
		for _, root := range roots {
			modules = append(modules, map[string]any{
				"path":         root.envID + ".md",
				"class":        "root",
				"environments": []any{root.envID},
			})
		}
		manifest["context"] = map[string]any{"modules": modules}
	}
	if len(skills) > 0 {
		entries := map[string]any{}
		for _, skill := range skills {
			entries[skill.name] = map[string]any{"git": skill.git, "revision": skill.revision}
		}
		manifest["requires"] = map[string]any{"skills": entries}
	}
	payload, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		_ = os.RemoveAll(stage)
		return "", err
	}
	if err := os.WriteFile(filepath.Join(stage, "agent-context.json"), append(payload, '\n'), 0o644); err != nil {
		_ = os.RemoveAll(stage)
		return "", err
	}
	if len(roots) > 0 {
		contextDir := filepath.Join(stage, "context")
		if err := os.MkdirAll(contextDir, 0o755); err != nil {
			_ = os.RemoveAll(stage)
			return "", err
		}
		for _, root := range roots {
			if err := os.WriteFile(filepath.Join(contextDir, root.envID+".md"), normalizeImportText(root.data), 0o644); err != nil {
				_ = os.RemoveAll(stage)
				return "", err
			}
		}
	}
	return stage, nil
}
