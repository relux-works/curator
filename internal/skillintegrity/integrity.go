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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/skillspec"
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
	ManifestKind      string          `json:"manifest_kind"`
	DeclarationPath   string          `json:"declaration_path"`
	DeclarationSHA256 string          `json:"declaration_sha256"`
	Declaration       json.RawMessage `json:"declaration"`
	HashVersion       int             `json:"hash_version"`
	ContentSHA256     string          `json:"content_sha256"`
	Files             []File          `json:"files"`
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
	return map[string]any{"schema": m.Schema, "name": m.Name, "source": m.Source, "commit": m.Commit, "state_sha256": m.StateSHA256, "directory": m.Directory, "manifest_kind": m.ManifestKind, "declaration_path": m.DeclarationPath, "declaration_sha256": m.DeclarationSHA256, "declaration": declaration, "hash_version": m.HashVersion, "content_sha256": m.ContentSHA256, "files": files}
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
}

// NewBudget starts the shared file, byte and deadline budget for one pass.
func NewBudget() *Budget { return &Budget{started: time.Now()} }
func (b *Budget) check(ctx context.Context) error {
	if ctx.Err() != nil || time.Since(b.started) > VerificationTimeout {
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

type capture struct {
	data  map[string][]byte
	infos map[string]os.FileInfo
	files []File
	// Complete describes the initial traversal and reads, before revalidation.
	// A partial observation cannot establish that an expected file is absent.
	complete bool
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
	return build(ctx, root, identity, budget, nil)
}

func build(ctx context.Context, root string, identity Identity, budget *Budget, expected *Manifest) (Entry, error) {
	if budget == nil {
		budget = NewBudget()
	}
	if err := budget.check(ctx); err != nil {
		return Entry{}, err
	}
	c, err := captureTree(ctx, root, budget)
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
	m := Manifest{Schema: Schema, Identity: identity, HashVersion: 2, ContentSHA256: content, Files: c.files, ManifestKind: "synthesized", Declaration: json.RawMessage(syntheticDeclaration)}
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
		var declaration any
		decoder := json.NewDecoder(bytes.NewReader(m.Declaration))
		decoder.UseNumber()
		if protocoljson.Validate(m.Declaration) != nil || decoder.Decode(&declaration) != nil {
			return invalid()
		}
		var declarationFields map[string]json.RawMessage
		if json.Unmarshal(m.Declaration, &declarationFields) != nil || !bytes.Equal(bytes.TrimSpace(declarationFields["schema_version"]), []byte("3")) {
			return invalid()
		}
		got, err := protocoljson.MarshalCanonical(declaration)
		if err != nil {
			return invalid()
		}
		var synthetic any
		decoder = json.NewDecoder(strings.NewReader(syntheticDeclaration))
		decoder.UseNumber()
		_ = decoder.Decode(&synthetic)
		want, _ := protocoljson.MarshalCanonical(synthetic)
		if !bytes.Equal(got, want) {
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

// Verify never creates evidence. Missing evidence is handled by the caller's
// typed unavailable observation; malformed supplied evidence is always fatal.
func Verify(ctx context.Context, root string, expected Identity, e Entry, budget *Budget) error {
	if err := e.Validate(expected); err != nil {
		return err
	}
	actual, err := build(ctx, root, expected, budget, &e.Manifest)
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
	if json.Unmarshal(object["manifest"], &m) != nil || !exactKeys(m, "schema", "name", "source", "commit", "state_sha256", "directory", "manifest_kind", "declaration_path", "declaration_sha256", "declaration", "hash_version", "content_sha256", "files") {
		return invalid()
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
