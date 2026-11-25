// Package skillintegrity captures and verifies unsigned skill evidence bound to
// an already validated profile member. It never modifies a source or profile.
package skillintegrity

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/whitelist"
)

// Schema identifies the unsigned skill evidence format.
const Schema = "curator.skill-integrity/1"

// MaxFiles bounds the regular files across one selected skill closure.
const MaxFiles = 16384

// MaxBytes bounds the file bytes across one selected skill closure.
const MaxBytes int64 = 64 << 20

// VerificationTimeout bounds one shared skill verification pass.
const VerificationTimeout = 30 * time.Second
const syntheticDeclaration = `{"schema_version":3,"runtime_roots":[],"commands":{},"capabilities":{"network":"none","filesystem":"none","exec":"none","secrets":"none","env_read":"none"}}`

// Diagnostic classifies an integrity refusal or unavailable observation.
type Diagnostic struct {
	Code   string
	Detail string
}

func (d *Diagnostic) Error() string { return d.Code + ": " + d.Detail }

// Is matches diagnostics by their stable code rather than their detail.
func (d *Diagnostic) Is(target error) bool {
	other, ok := target.(*Diagnostic)
	return ok && d.Code == other.Code
}

// IsCode reports whether an error chain contains the given integrity code.
func IsCode(err error, code string) bool { return errors.Is(err, &Diagnostic{Code: code}) }
func failure(code, detail string) error  { return &Diagnostic{code, detail} }

// Identity must come from the existing strict lock member reader. Its source,
// Git/state pin and selected-directory validation remains that reader's contract.
type Identity struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Commit      string `json:"commit"`
	StateSHA256 string `json:"state_sha256"`
	Directory   string `json:"directory"`
}

// File records one portable regular file's content and executable bit.
type File struct {
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	Executable bool   `json:"executable"`
	SHA256     string `json:"sha256"`
}

// Manifest binds a complete file roster and declaration to a selected member.
type Manifest struct {
	Schema string `json:"schema"`
	Identity
	ManifestKind      string                   `json:"manifest_kind"`
	DeclarationPath   string                   `json:"declaration_path"`
	DeclarationSHA256 string                   `json:"declaration_sha256"`
	Declaration       json.RawMessage          `json:"declaration"`
	HashVersion       int                      `json:"hash_version"`
	ContentSHA256     string                   `json:"content_sha256"`
	Files             []File                   `json:"files"`
	FileSelection     *whitelist.FileSelection `json:"file_selection,omitempty"`
	ContextFiles      []string                 `json:"context_files,omitempty"`
}

// Entry pairs a manifest with its canonical CCJ-1 digest.
type Entry struct {
	Manifest       Manifest `json:"manifest"`
	ManifestSHA256 string   `json:"manifest_sha256"`
}

// Object projects only the admitted JSON domain for CCJ-1 encoding.
func (m Manifest) Object() map[string]any {
	files := make([]any, 0, len(m.Files))
	for _, file := range m.Files {
		files = append(files, map[string]any{"path": file.Path, "size": file.Size, "executable": file.Executable, "sha256": file.SHA256})
	}
	var declaration any
	if m.Declaration != nil {
		decoder := json.NewDecoder(bytes.NewReader(m.Declaration))
		decoder.UseNumber()
		// Validate checks the closed declaration before this projection is used.
		_ = decoder.Decode(&declaration)
	}
	object := map[string]any{"schema": m.Schema, "name": m.Name, "source": m.Source, "commit": m.Commit, "state_sha256": m.StateSHA256, "directory": m.Directory, "manifest_kind": m.ManifestKind, "declaration_path": m.DeclarationPath, "declaration_sha256": m.DeclarationSHA256, "declaration": declaration, "hash_version": m.HashVersion, "content_sha256": m.ContentSHA256, "files": files}
	if m.FileSelection != nil {
		object["file_selection"] = fileSelectionObject(m.FileSelection)
		object["context_files"] = m.ContextFiles
	}
	return object
}

func fileSelectionObject(selection *whitelist.FileSelection) map[string]any {
	include := append([]string{}, selection.Include...)
	exclude := append([]string{}, selection.Exclude...)
	return map[string]any{"include": include, "exclude": exclude}
}

// Object projects the entry's admitted JSON domain for canonical encoding.
func (e Entry) Object() map[string]any {
	return map[string]any{"manifest": e.Manifest.Object(), "manifest_sha256": e.ManifestSHA256}
}

// Budget is shared across every skill in one compose/verification pass.
type Budget struct {
	files   int
	bytes   int64
	started time.Time
	elapsed time.Duration
	paused  bool
}

// NewBudget starts the shared file, byte and deadline budget for one pass.
func NewBudget() *Budget { return &Budget{started: time.Now()} }

// Pause excludes source inspection and other work outside integrity capture
// from the shared verification deadline, while retaining accumulated file and
// byte limits.
func (b *Budget) Pause() { b.pauseAt(time.Now()) }

// Resume starts counting integrity time again after a paused interval.
func (b *Budget) Resume() { b.resumeAt(time.Now()) }

func (b *Budget) pauseAt(now time.Time) {
	if b.paused {
		return
	}
	b.elapsed += now.Sub(b.started)
	b.paused = true
}

func (b *Budget) resumeAt(now time.Time) {
	if !b.paused {
		return
	}
	b.started = now
	b.paused = false
}

func (b *Budget) elapsedAt(now time.Time) time.Duration {
	if b.paused {
		return b.elapsed
	}
	return b.elapsed + now.Sub(b.started)
}

func (b *Budget) check(ctx context.Context) error {
	if ctx.Err() != nil || b.elapsedAt(time.Now()) > VerificationTimeout {
		return failure("skill_integrity_unavailable", "verification interrupted or deadline reached")
	}
	return nil
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func validDigest(value string) bool {
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	for _, c := range value[7:] {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

func cloneFileSelection(selection *whitelist.FileSelection) *whitelist.FileSelection {
	if selection == nil {
		return nil
	}
	clone := &whitelist.FileSelection{
		Include: append([]string(nil), selection.Include...),
		Exclude: append([]string(nil), selection.Exclude...),
	}
	sort.Strings(clone.Include)
	sort.Strings(clone.Exclude)
	return clone
}

func equalFileSelection(left, right *whitelist.FileSelection) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	left, right = cloneFileSelection(left), cloneFileSelection(right)
	return slices.Equal(left.Include, right.Include) && slices.Equal(left.Exclude, right.Exclude)
}

type capture struct {
	data  map[string][]byte
	infos map[string]os.FileInfo
	files []File
	// Complete describes the initial traversal and reads, before revalidation.
	// A partial observation cannot establish that an expected file is absent.
	complete       bool
	filtered       bool
	selection      *whitelist.FileSelection
	contextFiles   []string
	includeScripts bool
	excludeRoots   []string
	preserveRoots  []string
	sourceFile     string
}

func (c *capture) reader(root string) skillspec.Reader {
	relative := func(path string) (string, error) {
		rel, err := filepath.Rel(root, path)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return "", fs.ErrPermission
		}
		return filepath.ToSlash(rel), nil
	}
	stat := func(path string) (os.FileInfo, error) {
		rel, err := relative(path)
		if err != nil {
			return nil, err
		}
		info, ok := c.infos[rel]
		if !ok {
			return nil, fs.ErrNotExist
		}
		return info, nil
	}
	return skillspec.Reader{Lstat: stat, Stat: stat, ReadFile: func(path string) ([]byte, error) {
		rel, err := relative(path)
		if err != nil {
			return nil, err
		}
		data, ok := c.data[rel]
		if !ok {
			return nil, fs.ErrNotExist
		}
		return bytes.Clone(data), nil
	}}
}

// Build uses one bounded capture for both hashing and native metadata parsing.
// The caller supplies a shared budget for a closure, or nil for one skill.
func Build(ctx context.Context, root string, identity Identity, budget *Budget) (Entry, error) {
	return build(ctx, root, identity, budget, nil, nil, nil)
}

// BuildWithProfileCapabilities synthesizes the same bounded evidence as Build
// while recording validated capabilities supplied by the containing profile.
// A nil declaration retains the existing all-none synthetic bytes.
func BuildWithProfileCapabilities(ctx context.Context, root string, identity Identity, declared *capabilities.Manifest, budget *Budget) (Entry, error) {
	return build(ctx, root, identity, budget, nil, nil, declared)
}

// BuildWithProfileSelection captures the exact profile-selected context while
// retaining declared runtime and build roots in the bounded evidence.
func BuildWithProfileSelection(ctx context.Context, root string, identity Identity, selection *whitelist.FileSelection, budget *Budget) (Entry, error) {
	return build(ctx, root, identity, budget, nil, selection, nil)
}

// BuildWithProfileCapabilitiesAndSelection is BuildWithProfileSelection for a
// manifestless skill whose capabilities are declared by the containing profile.
func BuildWithProfileCapabilitiesAndSelection(ctx context.Context, root string, identity Identity, declared *capabilities.Manifest, selection *whitelist.FileSelection, budget *Budget) (Entry, error) {
	return build(ctx, root, identity, budget, nil, selection, declared)
}

func build(ctx context.Context, root string, identity Identity, budget *Budget, expected *Manifest, selection *whitelist.FileSelection, profileCapabilities *capabilities.Manifest) (Entry, error) {
	if budget == nil {
		budget = NewBudget()
	}
	if err := budget.check(ctx); err != nil {
		return Entry{}, err
	}
	if expected != nil {
		if selection != nil && !equalFileSelection(selection, expected.FileSelection) {
			return Entry{}, failure("skill_integrity_mismatch", "profile file selection differs from pinned evidence")
		}
		selection = cloneFileSelection(expected.FileSelection)
	}
	selection = cloneFileSelection(selection)
	var c *capture
	var contextFiles []string
	var err error
	if selection == nil {
		c, err = captureTree(ctx, root, budget)
	} else {
		if err := whitelist.ValidateFileSelection(selection); err != nil {
			return Entry{}, failure("profile_skill_files_invalid", "file selection is invalid")
		}
		spec, specErr := skillspec.Load(root)
		if specErr != nil {
			return Entry{}, failure("skill_integrity_manifest_invalid", "skill declaration is invalid")
		}
		excludeRoots := skillContextExcludedRoots(spec)
		contextFiles, err = whitelist.ContextFiles(root, len(spec.Commands) == 0, excludeRoots, selection)
		if err != nil {
			return Entry{}, err
		}
		if expected != nil && !slices.Equal(contextFiles, expected.ContextFiles) {
			return Entry{}, failure("skill_integrity_mismatch", "selected context file set differs from pinned evidence")
		}
		c, err = captureSelectedTree(ctx, root, budget, contextFiles, spec, selection)
	}
	if err != nil {
		// A completed per-file mismatch is a failure even if a later read is
		// unavailable. A completed traversal also establishes missing members;
		// a partial capture never proves an expected file absent.
		if expected != nil && c != nil && observedMismatch(c.files, expected.Files, c.complete) {
			return Entry{}, errors.Join(failure("skill_integrity_mismatch", "observed file roster or metadata contradicted expected evidence"), err)
		}
		return Entry{}, err
	}
	// Supplied evidence was validated by Verify before capture. A completed
	// changed roster/content is installed drift, even when it no longer parses.
	if expected != nil && observedMismatch(c.files, expected.Files, true) {
		return Entry{}, failure("skill_integrity_mismatch", "installed skill differs from expected evidence")
	}
	reader := c.reader(root)
	// The existing resolver deliberately retains the legacy root-addressed
	// skill contract. Selected-directory skills use section-3 frontmatter;
	// legacy roots validate declarations without retroactive metadata rules.
	if identity.Directory == "" {
		if _, present := c.data["SKILL.md"]; !present {
			return Entry{}, failure("skill_integrity_manifest_invalid", "legacy skill has no SKILL.md")
		}
		if _, err := reader.Load(root); err != nil {
			return Entry{}, failure("skill_integrity_manifest_invalid", "skill declaration is invalid")
		}
	} else if _, _, err := manifest.ReadSkillPackageWith(root, reader); err != nil {
		return Entry{}, failure("skill_integrity_manifest_invalid", "skill declaration is invalid")
	}
	content, err := hashing.ContentSHA256Files(c.data, hashing.VersionV2)
	if err != nil {
		return Entry{}, failure("skill_integrity_manifest_invalid", "content hash could not be computed")
	}
	declaration := []byte(syntheticDeclaration)
	if expected != nil && expected.ManifestKind == "synthesized" {
		// Verify validates this frozen declaration before capture. Startup
		// compares it directly and never reloads the current profile.
		declaration = bytes.Clone(expected.Declaration)
	} else if profileCapabilities != nil {
		if source := reader.ManifestSourcePath(root); source != "" {
			return Entry{}, failure("profile_skill_capability_native_manifest", "profile capabilities cannot replace a native skill declaration")
		}
		var err error
		declaration, err = synthesizedDeclarationFor(*profileCapabilities)
		if err != nil {
			return Entry{}, failure("profile_skill_capability_invalid", "profile capabilities are invalid")
		}
	}
	m := Manifest{Schema: Schema, Identity: identity, HashVersion: 2, ContentSHA256: content, Files: c.files, ManifestKind: "synthesized", Declaration: json.RawMessage(declaration), FileSelection: cloneFileSelection(selection), ContextFiles: contextFiles}
	if source := reader.ManifestSourcePath(root); source != "" {
		m.ManifestKind = "native"
		m.DeclarationPath = source
		m.DeclarationSHA256 = digest(c.data[source])
		m.Declaration = nil
	}
	raw, err := protocoljson.MarshalCanonical(m.Object())
	if err != nil || int64(len(raw)) > MaxBytes {
		return Entry{}, failure("skill_integrity_manifest_invalid", "manifest exceeds evidence budget")
	}
	if err := budget.check(ctx); err != nil {
		return Entry{}, err
	}
	entry := Entry{Manifest: m, ManifestSHA256: digest(raw)}
	if err := entry.Validate(identity); err != nil {
		return Entry{}, err
	}
	return entry, nil
}

func synthesizedDeclarationFor(value capabilities.Manifest) ([]byte, error) {
	declared, err := capabilityObject(value)
	if err != nil {
		return nil, err
	}
	parsed, err := capabilities.Parse(declared)
	if err != nil {
		return nil, err
	}
	value = normalizeCapabilities(parsed)
	if allNoneCapabilities(value) {
		return []byte(syntheticDeclaration), nil
	}
	declared, err = capabilityObject(value)
	if err != nil {
		return nil, err
	}
	return protocoljson.MarshalCanonical(map[string]any{
		"schema_version": json.Number("3"),
		"runtime_roots":  []any{},
		"commands":       map[string]any{},
		"capabilities":   declared,
	})
}

func capabilityObject(value capabilities.Manifest) (map[string]any, error) {
	if value.Filesystem.Keyword != "" && len(value.Filesystem.Paths) != 0 {
		return nil, fmt.Errorf("filesystem cannot contain a keyword and path list")
	}
	declared := map[string]any{
		"network":    listOrNone(value.Network),
		"filesystem": filesystemValue(value.Filesystem),
		"exec":       listOrNone(value.Exec),
		"secrets":    listOrNone(value.Secrets),
		"env_read":   listOrNone(value.EnvRead),
	}
	if value.PromptScope != "" {
		declared["prompt_scope"] = value.PromptScope
	}
	return declared, nil
}

func normalizeCapabilities(value capabilities.Manifest) capabilities.Manifest {
	value.Network = sortedNetworkStrings(value.Network)
	value.Filesystem.Paths = sortedStrings(value.Filesystem.Paths)
	value.Exec = sortedStrings(value.Exec)
	value.Secrets = sortedStrings(value.Secrets)
	value.EnvRead = sortedStrings(value.EnvRead)
	return value
}

func sortedNetworkStrings(values []string) []string {
	values = sortedStrings(values)
	for index := range values {
		values[index] = strings.ToLower(values[index])
	}
	sort.Strings(values)
	return compactStrings(values)
}

func compactStrings(values []string) []string {
	if len(values) < 2 {
		return values
	}
	write := 1
	for read := 1; read < len(values); read++ {
		if values[read] == values[write-1] {
			continue
		}
		values[write] = values[read]
		write++
	}
	return values[:write]
}

func sortedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}

func allNoneCapabilities(value capabilities.Manifest) bool {
	return len(value.Network) == 0 && value.Filesystem.Keyword == "" && len(value.Filesystem.Paths) == 0 &&
		len(value.Exec) == 0 && len(value.Secrets) == 0 && len(value.EnvRead) == 0 && value.PromptScope == ""
}

func listOrNone(values []string) any {
	if len(values) == 0 {
		return "none"
	}
	result := make([]any, len(values))
	for index, value := range values {
		result[index] = value
	}
	return result
}

func filesystemValue(value capabilities.Filesystem) any {
	if value.Keyword != "" {
		return value.Keyword
	}
	return listOrNone(value.Paths)
}

func captureTree(ctx context.Context, root string, budget *Budget) (*capture, error) {
	c := &capture{data: map[string][]byte{}, infos: map[string]os.FileInfo{}, files: []File{}}
	folded := map[string]bool{}
	var observations []error
	err := filepath.WalkDir(root, func(path string, dir fs.DirEntry, walkErr error) error {
		if err := budget.check(ctx); err != nil {
			return err
		}
		if walkErr != nil {
			observations = append(observations, failure("skill_integrity_unavailable", "skill tree cannot be observed"))
			return nil
		}
		info, err := captureEntryInfo(path, dir)
		if err != nil {
			if IsCode(err, "skill_integrity_boundary_invalid") {
				return err
			}
			observations = append(observations, failure("skill_integrity_unavailable", "skill entry cannot be observed"))
			return nil
		}
		rel, err := admitEntry(root, path, info, folded)
		if err != nil {
			return err
		}
		c.infos[rel] = info
		if info.IsDir() {
			return nil
		}
		if budget.files >= MaxFiles || info.Size() < 0 || info.Size() > MaxBytes-budget.bytes {
			return failure("skill_integrity_manifest_invalid", "capture budget exceeded")
		}
		remaining := MaxBytes - budget.bytes
		// Reserve every observed regular file, including unavailable reads,
		// so closure limits cannot be bypassed with unreadable entries.
		budget.files++
		budget.bytes += info.Size()
		data, observed, err := readCapturedFile(path, info, remaining)
		if err != nil {
			if IsCode(err, "skill_integrity_unavailable") {
				observations = append(observations, err)
				return nil
			}
			return err
		}
		c.data[rel] = data
		c.infos[rel] = observed
		c.files = append(c.files, File{Path: rel, Size: int64(len(data)), Executable: observed.Mode().Perm()&0o111 != 0, SHA256: digest(data)})
		return budget.check(ctx)
	})
	if err != nil {
		observations = append(observations, err)
	}
	if err := errors.Join(observations...); err != nil {
		return c, err
	}
	c.complete = true
	if err := c.revalidate(ctx, root, budget); err != nil {
		return c, err
	}
	sort.Slice(c.files, func(i, j int) bool { return c.files[i].Path < c.files[j].Path })
	return c, nil
}

func captureSelectedTree(ctx context.Context, root string, budget *Budget, contextFiles []string, spec *skillspec.Spec, selection *whitelist.FileSelection) (*capture, error) {
	c := &capture{
		data:           map[string][]byte{},
		infos:          map[string]os.FileInfo{},
		files:          []File{},
		filtered:       true,
		selection:      cloneFileSelection(selection),
		contextFiles:   append([]string(nil), contextFiles...),
		includeScripts: len(spec.Commands) == 0,
		excludeRoots:   skillContextExcludedRoots(spec),
		preserveRoots:  skillContextExcludedRoots(spec),
		sourceFile:     spec.SourceFile,
	}
	selected := make(map[string]bool, len(contextFiles)+1)
	for _, file := range contextFiles {
		selected[file] = true
	}
	if spec.SourceFile != "" {
		selected[spec.SourceFile] = true
	}

	folded := map[string]bool{}
	var observations []error
	err := filepath.WalkDir(root, func(pathname string, entry fs.DirEntry, walkErr error) error {
		if err := budget.check(ctx); err != nil {
			return err
		}
		if walkErr != nil {
			observations = append(observations, failure("skill_integrity_unavailable", "selected skill paths cannot be observed"))
			return nil
		}
		rel, err := filepath.Rel(root, pathname)
		if err != nil {
			return failure("skill_integrity_boundary_invalid", "selected path is not relative to the skill root")
		}
		posix := whitelist.CanonicalContextPath(root, filepath.ToSlash(rel))
		if entry.IsDir() && posix != "." && !relevantSelectionPath(posix, selected, c.preserveRoots) {
			return filepath.SkipDir
		}
		if entry.IsDir() && whitelist.SelectionExcludesDirectory(selection, posix) {
			if intersectsRoots(posix, c.preserveRoots) {
				return failure("profile_skill_files_build_source_excluded", "file selection excludes a declared build or runtime root")
			}
			return filepath.SkipDir
		}
		needed := selected[posix] || isWithinRoots(posix, c.preserveRoots)
		if !needed && !entry.IsDir() {
			return nil
		}
		info, err := captureEntryInfo(pathname, entry)
		if err != nil {
			return err
		}
		if _, err := admitEntry(root, pathname, info, folded); err != nil {
			return err
		}
		c.infos[posix] = info
		if info.IsDir() {
			return nil
		}
		if !needed {
			return nil
		}
		if isWithinRoots(posix, c.preserveRoots) && whitelist.SelectionExcludesFile(selection, posix) {
			return failure("profile_skill_files_build_source_excluded", "file selection excludes a declared build or runtime input")
		}
		if budget.files >= MaxFiles || info.Size() < 0 || info.Size() > MaxBytes-budget.bytes {
			return failure("skill_integrity_manifest_invalid", "capture budget exceeded")
		}
		remaining := MaxBytes - budget.bytes
		budget.files++
		budget.bytes += info.Size()
		data, observed, err := readCapturedFile(pathname, info, remaining)
		if err != nil {
			if IsCode(err, "skill_integrity_unavailable") {
				observations = append(observations, err)
				return nil
			}
			return err
		}
		c.data[posix] = data
		c.infos[posix] = observed
		c.files = append(c.files, File{Path: posix, Size: int64(len(data)), Executable: observed.Mode().Perm()&0o111 != 0, SHA256: digest(data)})
		return budget.check(ctx)
	})
	if err != nil {
		observations = append(observations, err)
	}
	if err := errors.Join(observations...); err != nil {
		return c, err
	}
	c.complete = true
	if err := c.revalidate(ctx, root, budget); err != nil {
		return c, err
	}
	sort.Slice(c.files, func(i, j int) bool { return c.files[i].Path < c.files[j].Path })
	return c, nil
}

func relevantSelectionPath(path string, selected map[string]bool, roots []string) bool {
	if selected[path] || isWithinRoots(path, roots) {
		return true
	}
	for selectedPath := range selected {
		if path == "." || strings.HasPrefix(selectedPath, path+"/") {
			return true
		}
	}
	for _, root := range roots {
		if path == "." || strings.HasPrefix(root, path+"/") {
			return true
		}
	}
	return false
}

func isWithinRoots(path string, roots []string) bool {
	for _, root := range roots {
		if path == root || strings.HasPrefix(path, root+"/") {
			return true
		}
	}
	return false
}

func intersectsRoots(path string, roots []string) bool {
	if isWithinRoots(path, roots) {
		return true
	}
	for _, root := range roots {
		if path == "." || strings.HasPrefix(root, path+"/") {
			return true
		}
	}
	return false
}

func skillContextExcludedRoots(spec *skillspec.Spec) []string {
	var modules []string
	for _, command := range spec.Commands {
		if command.Driver == "go-v1" {
			modules = append(modules, command.Modules...)
		}
	}
	return whitelist.ContextExcludedRoots(spec.RuntimeRoots, spec.BuildRoots, modules)
}

// admitEntry is the positive path/kind admission used by both passes. A late
// forbidden entry must remain a hard boundary refusal at every startup level.
func admitEntry(root, path string, info os.FileInfo, folded map[string]bool) (string, error) {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return "", failure("skill_integrity_boundary_invalid", "invalid relative path")
	}
	rel = filepath.ToSlash(rel)
	if rel == "." && !info.IsDir() {
		return "", failure("skill_integrity_boundary_invalid", "skill root must be a directory")
	}
	if rel != "." {
		if !identifiers.PortablePath(rel) || folded[strings.ToLower(rel)] {
			return "", failure("skill_integrity_boundary_invalid", "nonportable or colliding path")
		}
		folded[strings.ToLower(rel)] = true
		for _, part := range strings.Split(rel, "/") {
			if strings.EqualFold(part, ".git") {
				return "", failure("skill_integrity_boundary_invalid", "Git metadata is not skill payload")
			}
		}
	}
	if info.Mode()&os.ModeSymlink != 0 || (!info.IsDir() && (!info.Mode().IsRegular() || !singleLink(path, info))) {
		return "", failure("skill_integrity_boundary_invalid", "payload must contain directories and single-link regular files")
	}
	if err := pathboundary.ValidateRouteWithOwner(root, path, pathboundary.DefaultOwnerLookup()); err != nil {
		return "", failure("skill_integrity_boundary_invalid", "skill path boundary is invalid")
	}
	return rel, nil
}

// ValidateLegacyBoundary checks the existing full snapshot without adopting
// lock6 capture limits or evidence. Unreadable entries never hide a later
// observed forbidden kind. Byte identity remains the legacy pin algorithm.
func ValidateLegacyBoundary(root string) error {
	folded := map[string]bool{}
	var unavailable error
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			unavailable = failure("skill_integrity_unavailable", "snapshot entry cannot be observed")
			return nil
		}
		info, err := captureEntryInfo(path, entry)
		if err != nil {
			if IsCode(err, "skill_integrity_boundary_invalid") {
				return err
			}
			unavailable = failure("skill_integrity_unavailable", "snapshot metadata cannot be observed")
			return nil
		}
		_, err = admitEntry(root, path, info, folded)
		if err != nil {
			var diagnostic *Diagnostic
			if errors.As(err, &diagnostic) {
				return failure(diagnostic.Code, fmt.Sprintf("%s: %s", path, diagnostic.Detail))
			}
			return fmt.Errorf("%s: %w", path, err)
		}
		return nil
	})
	return errors.Join(unavailable, err)
}

func (c *capture) revalidate(ctx context.Context, root string, budget *Budget) error {
	if c.filtered {
		return c.revalidateFiltered(ctx, root, budget)
	}
	seen := map[string]bool{}
	folded := map[string]bool{}
	failed, unavailable, complete := false, false, true
	priorFiles, priorBytes := budget.files-len(c.files), budget.bytes
	for _, file := range c.files {
		priorBytes -= file.Size
	}
	regular, total := 0, int64(0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if err := budget.check(ctx); err != nil {
			return err
		}
		if walkErr != nil {
			unavailable, complete = true, false
			return nil
		}
		info, err := captureEntryInfo(path, entry)
		if err != nil {
			if IsCode(err, "skill_integrity_boundary_invalid") {
				return err
			}
			unavailable, complete = true, false
			return nil
		}
		rel, err := admitEntry(root, path, info, folded)
		if err != nil {
			return err
		}
		seen[rel] = true
		if !info.IsDir() {
			regular++
			if priorFiles+regular > MaxFiles || info.Size() < 0 || info.Size() > MaxBytes-priorBytes-total {
				return failure("skill_integrity_manifest_invalid", "rechecked closure exceeds capture budget")
			}
			total += info.Size()
		}
		before, ok := c.infos[rel]
		if !ok {
			if info.IsDir() {
				unavailable = true
			} else {
				failed = true
			}
			return nil
		}
		if !os.SameFile(before, info) || before.Mode() != info.Mode() || before.Size() != info.Size() || !before.ModTime().Equal(info.ModTime()) {
			// Timestamp/inode changes alone do not contradict the frozen byte
			// identity. Keep scanning: a later forbidden entry is still fatal.
			if !info.IsDir() && (before.Size() != info.Size() || before.Mode().Perm()&0o111 != info.Mode().Perm()&0o111) {
				failed = true
			} else {
				unavailable = true
			}
		}
		return nil
	})
	if err == nil && complete {
		for path, before := range c.infos {
			if !seen[path] {
				if before.IsDir() {
					unavailable = true
				} else {
					failed = true
				}
			}
		}
	}
	var results []error
	if failed {
		results = append(results, failure("skill_integrity_mismatch", "skill roster or bound metadata changed during capture"))
	}
	if unavailable {
		results = append(results, failure("skill_integrity_unavailable", "skill capture could not be revalidated"))
	}
	if err != nil {
		results = append(results, err)
	}
	return errors.Join(results...)
}

func (c *capture) revalidateFiltered(ctx context.Context, root string, budget *Budget) error {
	spec, err := skillspec.Load(root)
	if err != nil {
		return failure("skill_integrity_unavailable", "selected skill declaration cannot be re-read")
	}
	if spec.SourceFile != c.sourceFile || !slices.Equal(skillContextExcludedRoots(spec), c.preserveRoots) {
		return failure("skill_integrity_mismatch", "declared build or runtime roots changed during capture")
	}
	excludeRoots := skillContextExcludedRoots(spec)
	contextFiles, err := whitelist.ContextFiles(root, c.includeScripts, excludeRoots, c.selection)
	if err != nil {
		return err
	}
	if !slices.Equal(contextFiles, c.contextFiles) {
		return failure("skill_integrity_mismatch", "selected context file set changed during capture")
	}
	selected := make(map[string]bool, len(contextFiles)+1)
	for _, file := range contextFiles {
		selected[file] = true
	}
	if spec.SourceFile != "" {
		selected[spec.SourceFile] = true
	}
	current := map[string]os.FileInfo{}
	folded := map[string]bool{}
	priorFiles := budget.files - len(c.files)
	priorBytes := budget.bytes
	for _, file := range c.files {
		priorBytes -= file.Size
	}
	regular := 0
	var total int64
	err = filepath.WalkDir(root, func(pathname string, entry fs.DirEntry, walkErr error) error {
		if err := budget.check(ctx); err != nil {
			return err
		}
		if walkErr != nil {
			return failure("skill_integrity_unavailable", "selected skill paths cannot be re-observed")
		}
		rel, err := filepath.Rel(root, pathname)
		if err != nil {
			return failure("skill_integrity_boundary_invalid", "selected path is not relative to the skill root")
		}
		posix := whitelist.CanonicalContextPath(root, filepath.ToSlash(rel))
		if entry.IsDir() && posix != "." && !relevantSelectionPath(posix, selected, c.preserveRoots) {
			return filepath.SkipDir
		}
		if entry.IsDir() && whitelist.SelectionExcludesDirectory(c.selection, posix) {
			if intersectsRoots(posix, c.preserveRoots) {
				return failure("profile_skill_files_build_source_excluded", "file selection excludes a declared build or runtime root")
			}
			return filepath.SkipDir
		}
		needed := selected[posix] || isWithinRoots(posix, c.preserveRoots)
		if !needed && !entry.IsDir() {
			return nil
		}
		info, err := captureEntryInfo(pathname, entry)
		if err != nil {
			return err
		}
		if _, err := admitEntry(root, pathname, info, folded); err != nil {
			return err
		}
		current[posix] = info
		if !info.IsDir() {
			regular++
			if priorFiles+regular > MaxFiles || info.Size() < 0 || info.Size() > MaxBytes-priorBytes-total {
				return failure("skill_integrity_manifest_invalid", "rechecked closure exceeds capture budget")
			}
			total += info.Size()
		}
		return nil
	})
	if err != nil {
		return err
	}
	failed, unavailable := false, false
	for path, before := range c.infos {
		now, ok := current[path]
		if !ok {
			if before.IsDir() {
				unavailable = true
			} else {
				failed = true
			}
			continue
		}
		if !os.SameFile(before, now) || before.Mode() != now.Mode() || before.Size() != now.Size() || !before.ModTime().Equal(now.ModTime()) {
			if !now.IsDir() && (before.Size() != now.Size() || before.Mode().Perm()&0o111 != now.Mode().Perm()&0o111) {
				failed = true
			} else {
				unavailable = true
			}
		}
	}
	for path, now := range current {
		if _, ok := c.infos[path]; !ok {
			if now.IsDir() {
				unavailable = true
			} else {
				failed = true
			}
		}
	}
	var results []error
	if failed {
		results = append(results, failure("skill_integrity_mismatch", "selected skill files changed during capture"))
	}
	if unavailable {
		results = append(results, failure("skill_integrity_unavailable", "selected skill paths could not be revalidated"))
	}
	return errors.Join(results...)
}

// Validate binds evidence to its already validated lock identity before I/O.
func (e Entry) Validate(expected Identity) error {
	m := e.Manifest
	invalid := func() error { return failure("skill_integrity_manifest_invalid", "invalid integrity evidence") }
	if m.Schema != Schema || m.HashVersion != 2 {
		return failure("skill_integrity_unsupported_version", "unsupported integrity version")
	}
	if m.Identity != expected {
		return failure("skill_integrity_manifest_invalid", "integrity identity differs from its lock member")
	}
	if !identifiers.Valid(m.Name) || !validDigest(m.ContentSHA256) || m.Files == nil || len(m.Files) > MaxFiles {
		return invalid()
	}
	var total int64
	folded := map[string]bool{}
	roster := map[string]File{}
	last := ""
	for _, file := range m.Files {
		if !identifiers.PortablePath(file.Path) || file.Path <= last || folded[strings.ToLower(file.Path)] || file.Size < 0 || file.Size > MaxBytes-total || !validDigest(file.SHA256) {
			return invalid()
		}
		last = file.Path
		folded[strings.ToLower(file.Path)] = true
		total += file.Size
		roster[file.Path] = file
	}
	if (m.FileSelection == nil) != (m.ContextFiles == nil) {
		return invalid()
	}
	if m.FileSelection != nil {
		if whitelist.ValidateFileSelection(m.FileSelection) != nil || whitelist.ValidateFileSet(m.ContextFiles) != nil {
			return invalid()
		}
		canonical := cloneFileSelection(m.FileSelection)
		if !slices.Equal(canonical.Include, m.FileSelection.Include) || !slices.Equal(canonical.Exclude, m.FileSelection.Exclude) {
			return invalid()
		}
		for _, path := range m.ContextFiles {
			if _, ok := roster[path]; !ok || whitelist.SelectionExcludesFile(m.FileSelection, path) {
				return invalid()
			}
		}
	}
	switch m.ManifestKind {
	case "native":
		if m.DeclarationPath != skillspec.CanonicalManifestName && m.DeclarationPath != skillspec.LegacyManifestName && m.DeclarationPath != skillspec.RuntimeFallbackName {
			return invalid()
		}
		file, ok := roster[m.DeclarationPath]
		if !ok || !validDigest(m.DeclarationSHA256) || file.SHA256 != m.DeclarationSHA256 || (len(m.Declaration) != 0 && !bytes.Equal(m.Declaration, []byte("null"))) {
			return invalid()
		}
	case "synthesized":
		if m.DeclarationPath != "" || m.DeclarationSHA256 != "" {
			return invalid()
		}
		if !validSyntheticDeclaration(m.Declaration) {
			return invalid()
		}
		for _, path := range []string{skillspec.CanonicalManifestName, skillspec.LegacyManifestName, skillspec.RuntimeFallbackName} {
			if _, present := roster[path]; present {
				return invalid()
			}
		}
	default:
		return invalid()
	}
	raw, err := protocoljson.MarshalCanonical(m.Object())
	if err != nil || int64(len(raw)) > MaxBytes || !validDigest(e.ManifestSHA256) || digest(raw) != e.ManifestSHA256 {
		return invalid()
	}
	return nil
}

func validSyntheticDeclaration(raw json.RawMessage) bool {
	if protocoljson.Validate(raw) != nil {
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var declaration any
	if decoder.Decode(&declaration) != nil {
		return false
	}
	fields, ok := declaration.(map[string]any)
	if !ok || len(fields) != 4 {
		return false
	}
	schema, ok := fields["schema_version"].(json.Number)
	if !ok || schema.String() != "3" {
		return false
	}
	roots, ok := fields["runtime_roots"].([]any)
	if !ok || len(roots) != 0 {
		return false
	}
	commands, ok := fields["commands"].(map[string]any)
	if !ok || len(commands) != 0 {
		return false
	}
	declared, ok := fields["capabilities"].(map[string]any)
	if !ok {
		return false
	}
	if _, err := capabilities.Parse(declared); err != nil {
		return false
	}
	return true
}

// Verify never creates evidence. Missing evidence is handled by the caller's
// typed unavailable observation; malformed supplied evidence is always fatal.
func Verify(ctx context.Context, root string, expected Identity, e Entry, budget *Budget) error {
	if err := e.Validate(expected); err != nil {
		return err
	}
	actual, err := build(ctx, root, expected, budget, &e.Manifest, nil, nil)
	if err != nil {
		return err
	}
	if actual.ManifestSHA256 != e.ManifestSHA256 {
		return failure("skill_integrity_mismatch", "skill roster, bytes, metadata or declaration changed")
	}
	return nil
}

// DecodeEntry enforces required fields as well as unknown-field rejection;
// a missing false/zero-valued member must not acquire a decoding default.
func DecodeEntry(raw []byte, expected Identity) (Entry, error) {
	invalid := func() (Entry, error) {
		return Entry{}, failure("skill_integrity_manifest_invalid", "invalid evidence JSON")
	}
	if int64(len(raw)) > MaxBytes || protocoljson.Validate(raw) != nil {
		return invalid()
	}
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil || !exactKeys(object, "manifest", "manifest_sha256") {
		return invalid()
	}
	var m map[string]json.RawMessage
	if json.Unmarshal(object["manifest"], &m) != nil {
		return invalid()
	}
	manifestKeys := []string{"schema", "name", "source", "commit", "state_sha256", "directory", "manifest_kind", "declaration_path", "declaration_sha256", "declaration", "hash_version", "content_sha256", "files"}
	hasSelection, hasContextFiles := m["file_selection"] != nil, m["context_files"] != nil
	if hasSelection != hasContextFiles {
		return invalid()
	}
	if hasSelection {
		manifestKeys = append(manifestKeys, "file_selection", "context_files")
	}
	if !exactKeys(m, manifestKeys...) {
		return invalid()
	}
	if hasSelection {
		var selectionObject map[string]json.RawMessage
		if json.Unmarshal(m["file_selection"], &selectionObject) != nil || !exactKeys(selectionObject, "include", "exclude") {
			return invalid()
		}
		for _, key := range []string{"include", "exclude"} {
			if raw, present := selectionObject[key]; present {
				var list []string
				if len(bytes.TrimSpace(raw)) == 0 || bytes.TrimSpace(raw)[0] != '[' || json.Unmarshal(raw, &list) != nil || list == nil {
					return invalid()
				}
			}
		}
		var contextFiles []string
		if len(bytes.TrimSpace(m["context_files"])) == 0 || bytes.TrimSpace(m["context_files"])[0] != '[' || json.Unmarshal(m["context_files"], &contextFiles) != nil || contextFiles == nil {
			return invalid()
		}
	}
	for _, key := range []string{"schema", "name", "source", "commit", "state_sha256", "directory", "manifest_kind", "declaration_path", "declaration_sha256", "content_sha256"} {
		if value := bytes.TrimSpace(m[key]); len(value) == 0 || value[0] != '"' {
			return invalid()
		}
	}
	filesDecoder := json.NewDecoder(bytes.NewReader(m["files"]))
	if token, err := filesDecoder.Token(); err != nil || token != json.Delim('[') {
		return invalid()
	}
	count := 0
	for filesDecoder.More() {
		if count >= MaxFiles {
			return invalid()
		}
		count++
		var file map[string]json.RawMessage
		if filesDecoder.Decode(&file) != nil || !exactKeys(file, "path", "size", "executable", "sha256") {
			return invalid()
		}
		boolean := bytes.TrimSpace(file["executable"])
		if !bytes.Equal(boolean, []byte("true")) && !bytes.Equal(boolean, []byte("false")) {
			return invalid()
		}
		sizeToken := string(bytes.TrimSpace(file["size"]))
		size, err := strconv.ParseInt(sizeToken, 10, 64)
		if err != nil || size < 0 || size > MaxBytes || strconv.FormatInt(size, 10) != sizeToken {
			return invalid()
		}
	}
	if _, err := filesDecoder.Token(); err != nil {
		return invalid()
	}
	var e Entry
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&e) != nil {
		return invalid()
	}
	if err := e.Validate(expected); err != nil {
		return Entry{}, err
	}
	return e, nil
}
func exactKeys(object map[string]json.RawMessage, keys ...string) bool {
	if len(object) != len(keys) {
		return false
	}
	for _, key := range keys {
		if _, ok := object[key]; !ok {
			return false
		}
	}
	return true
}

func observedMismatch(observed, expected []File, complete bool) bool {
	byPath := map[string]File{}
	for _, file := range expected {
		byPath[file.Path] = file
	}
	for _, file := range observed {
		if want, ok := byPath[file.Path]; !ok || file != want {
			return true
		}
	}
	return complete && len(observed) != len(expected)
}
