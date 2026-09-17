// Package sourcelock implements the portable Skillfile package lock
// (Skillfile.lock.json, skillfile-lock schema 1) of the unreleased
// skillfile-sources-v1 extension.
//
// A package lock freezes the resolved closure: exact per-member package
// identities, the declaring Skillfile digest, and the selection indexes.
// It carries portable content identity only; machine-resolved absolute
// paths, endpoint choices, and credentials MUST NOT enter it. Those live
// in the machine-private Bindings record instead.
//
// This is NOT the environment context lock (internal/contextlock,
// context-lock-v1): different file, different schema, different identity
// arms. The two must never be parsed or consumed as each other.
//
// Diagnostic classes follow skillfile-sources §5: malformed lock shape is
// source_selection_invalid, malformed member records are
// source_member_invalid, duplicate names are source_name_conflict, a lock
// that no longer matches its manifest, plan, or machine bindings is
// source_lock_stale, and a plan member absent from the lock is
// source_member_missing.
package sourcelock

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/verr"
)

// SchemaVersion is the only Skillfile lock schema this reader accepts.
const SchemaVersion = 1

// FileName is the portable lock file next to the declaring Skillfile.
const FileName = "Skillfile.lock.json"

// Package identity arms (source-types schema 1). The arms are disjoint:
// a local snapshot never carries Git fields and a Git package never
// carries a snapshot digest.
const (
	KindLocalSnapshot = "local-snapshot"
	KindNetworkGit    = "network-git"
	KindConfiguredGit = "configured-git"
)

var (
	digestRE    = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	sha1RE      = regexp.MustCompile(`^[0-9a-f]{40}$`)
	sha256HexRE = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// Commit is a locked Git object: format-bound lowercase hex, never a
// snapshot digest and never prefixed.
type Commit struct {
	ObjectFormat string // "sha1" or "sha256"
	Hex          string
}

// Validate enforces the lockedCommit shape: sha1 binds 40 hex, sha256 64.
func (c Commit) Validate(path string) error {
	switch c.ObjectFormat {
	case "sha1":
		if !sha1RE.MatchString(c.Hex) {
			return verr.New(path, "source_member_invalid: sha1 commit must be 40 lowercase hex")
		}
	case "sha256":
		if !sha256HexRE.MatchString(c.Hex) {
			return verr.New(path, "source_member_invalid: sha256 commit must be 64 lowercase hex")
		}
	default:
		return verr.New(path, "source_member_invalid: commit object_format must be sha1 or sha256")
	}
	return nil
}

// Package is one frozen package identity. Exactly one arm is populated;
// Validate rejects any cross-arm field.
type Package struct {
	Kind       string
	Snapshot   string // local-snapshot only: sha256:<hex> inventory digest
	Repository string // network-git only: canonical host/path
	Source     string // configured-git only: legacy configured-root path
	Commit     Commit // Git arms only
	Directory  string // Git arms only: effective per-member directory
}

// LocalPackage builds a local-snapshot identity over an inventory digest.
func LocalPackage(snapshot string) (Package, error) {
	p := Package{Kind: KindLocalSnapshot, Snapshot: snapshot}
	if err := p.Validate("package"); err != nil {
		return Package{}, err
	}
	return p, nil
}

// NetworkGitPackage builds a network-git identity. Directory is the
// effective per-member source-relative directory and must equal the
// lock member directory.
func NetworkGitPackage(repository string, commit Commit, directory string) (Package, error) {
	p := Package{Kind: KindNetworkGit, Repository: repository, Commit: commit, Directory: directory}
	if err := p.Validate("package"); err != nil {
		return Package{}, err
	}
	return p, nil
}

// ConfiguredGitPackage builds a configured-git identity for a legacy entry
// whose repository has no canonical network identity. The directory is
// always "." relative to that entry's selected repository.
func ConfiguredGitPackage(source string, commit Commit) (Package, error) {
	p := Package{Kind: KindConfiguredGit, Source: source, Commit: commit, Directory: "."}
	if err := p.Validate("package"); err != nil {
		return Package{}, err
	}
	return p, nil
}

// IsGit reports whether the package uses a Git identity arm.
func (p Package) IsGit() bool {
	return p.Kind == KindNetworkGit || p.Kind == KindConfiguredGit
}

// Validate enforces the disjoint arm shapes of source-types schema 1.
func (p Package) Validate(path string) error {
	switch p.Kind {
	case KindLocalSnapshot:
		if !digestRE.MatchString(p.Snapshot) {
			return verr.New(path+".snapshot", "source_member_invalid: snapshot must be sha256:<64 lowercase hex>")
		}
		if p.Repository != "" || p.Source != "" || p.Commit != (Commit{}) || p.Directory != "" {
			return verr.New(path, "source_member_invalid: local-snapshot must not carry Git identity")
		}
	case KindNetworkGit:
		if p.Snapshot != "" || p.Source != "" {
			return verr.New(path, "source_member_invalid: network-git must not carry snapshot or configured source")
		}
		if !validRepository(p.Repository) {
			return verr.New(path+".repository", "source_member_invalid: repository must be canonical host/path")
		}
		if err := p.Commit.Validate(path + ".commit"); err != nil {
			return err
		}
		if !validLockDirectory(p.Directory) {
			return verr.New(path+".directory", "source_member_invalid: directory must be '.' or a portable contained path")
		}
	case KindConfiguredGit:
		if p.Snapshot != "" || p.Repository != "" {
			return verr.New(path, "source_member_invalid: configured-git must not carry snapshot or network repository")
		}
		if !identifiers.PortablePath(p.Source) {
			return verr.New(path+".source", "source_member_invalid: source must be a portable configured-root path")
		}
		if err := p.Commit.Validate(path + ".commit"); err != nil {
			return err
		}
		if p.Directory != "." {
			return verr.New(path+".directory", "source_member_invalid: configured-git directory must be '.'")
		}
	default:
		return verr.New(path+".kind", "source_selection_invalid: unknown package kind %q", p.Kind)
	}
	return nil
}

// validRepository enforces repository-transport revision 1 §1: a lock
// member's repository MUST already be canonical (lowercase host,
// case-sensitive path, exactly one terminal ".git" removed, no transport
// or username). identity.ValidCanonical covers grammar, host case, and
// transport/username rejection; the terminal ".git" suffix is enforced
// here so legacy identity semantics elsewhere stay untouched.
func validRepository(value string) bool {
	if value == "" || utf8.RuneCountInString(value) > 4096 {
		return false
	}
	if strings.HasSuffix(value, ".git") {
		return false
	}
	return identity.ValidCanonical(value)
}

// validLockDirectory admits "." and portable contained paths without glob
// metacharacters, mirroring the manifest selector rule.
func validLockDirectory(value string) bool {
	return value == "." || (identifiers.PortablePath(value) && !strings.ContainsAny(value, "*?[]"))
}

func (p Package) object() map[string]any {
	switch p.Kind {
	case KindNetworkGit:
		return map[string]any{
			"kind":       p.Kind,
			"repository": p.Repository,
			"commit":     map[string]any{"object_format": p.Commit.ObjectFormat, "hex": p.Commit.Hex},
			"directory":  p.Directory,
		}
	case KindConfiguredGit:
		return map[string]any{
			"kind":      p.Kind,
			"source":    p.Source,
			"commit":    map[string]any{"object_format": p.Commit.ObjectFormat, "hex": p.Commit.Hex},
			"directory": p.Directory,
		}
	default:
		return map[string]any{"kind": p.Kind, "snapshot": p.Snapshot}
	}
}

// Canonical returns the CCJ-1 bytes of the package identity after validation.
// Runtime store keys hash these bytes; callers must never substitute a
// snapshot digest into a Git commit field or vice versa.
func (p Package) Canonical() ([]byte, error) {
	if err := p.Validate("package"); err != nil {
		return nil, err
	}
	return protocoljson.MarshalCanonical(p.object())
}

// Digest returns the sha256:<hex> content identity of the package over its
// CCJ-1 bytes. Rebinding the same bytes on another machine preserves it.
func (p Package) Digest() (string, error) {
	canonical, err := p.Canonical()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// Member is one frozen closure member. Selection is the zero-based root
// skills index, shared by collection-expanded siblings, or nil for a
// transitive member. Directory is source-relative; for both Git arms it
// must equal the package directory.
type Member struct {
	Name          string
	Selection     *int
	Directory     string
	Package       Package
	ContentSHA256 string // context content hash, not package identity
}

// Validate enforces the member shape and the Git directory-agreement rule.
func (m Member) Validate(path string) error {
	if !identifiers.Valid(m.Name) {
		return verr.New(path+".name", "source_member_invalid: name %s", identifiers.Rule)
	}
	if m.Selection != nil && (*m.Selection < 0 || int64(*m.Selection) > protocoljson.MaxSafeInteger) {
		return verr.New(path+".selection", "source_selection_invalid: selection must be null or a non-negative safe integer")
	}
	if !validLockDirectory(m.Directory) {
		return verr.New(path+".directory", "source_member_invalid: directory must be '.' or a portable contained path")
	}
	if err := m.Package.Validate(path + ".package"); err != nil {
		return err
	}
	if m.Package.IsGit() && m.Directory != m.Package.Directory {
		return verr.New(path+".directory", "source_member_invalid: member directory %q differs from package directory %q", m.Directory, m.Package.Directory)
	}
	if !digestRE.MatchString(m.ContentSHA256) {
		return verr.New(path+".content_sha256", "source_member_invalid: content_sha256 must be sha256:<64 lowercase hex>")
	}
	return nil
}

func (m Member) object() map[string]any {
	var selection any
	if m.Selection != nil {
		selection = *m.Selection
	}
	return map[string]any{
		"name":           m.Name,
		"selection":      selection,
		"directory":      m.Directory,
		"package":        m.Package.object(),
		"content_sha256": m.ContentSHA256,
	}
}

// equalRecord reports whether two members agree on every frozen field.
func (m Member) equalRecord(other Member) bool {
	if m.Name != other.Name || m.Directory != other.Directory || m.ContentSHA256 != other.ContentSHA256 {
		return false
	}
	if (m.Selection == nil) != (other.Selection == nil) {
		return false
	}
	if m.Selection != nil && *m.Selection != *other.Selection {
		return false
	}
	a, b := m.Package, other.Package
	return a.Kind == b.Kind && a.Snapshot == b.Snapshot && a.Repository == b.Repository &&
		a.Source == b.Source && a.Commit == b.Commit && a.Directory == b.Directory
}

// Lock is one parsed Skillfile.lock.json object: the declaring manifest
// digest plus the frozen closure sorted by UTF-8 skill name.
type Lock struct {
	ManifestSHA256 string
	Members        []Member
	LockSHA256     string
}

// New builds a lock over a manifest digest, sorting members into canonical
// UTF-8 name order and computing lock_sha256.
func New(manifestSHA256 string, members []Member) (*Lock, error) {
	lock := &Lock{ManifestSHA256: manifestSHA256, Members: append([]Member(nil), members...)}
	sort.SliceStable(lock.Members, func(i, j int) bool {
		return bytes.Compare([]byte(lock.Members[i].Name), []byte(lock.Members[j].Name)) < 0
	})
	if err := lock.validateStructure(); err != nil {
		return nil, err
	}
	digest, err := lock.digest()
	if err != nil {
		return nil, err
	}
	lock.LockSHA256 = digest
	return lock, nil
}

// Find returns the locked member with the given installed name.
func (lock *Lock) Find(name string) (Member, bool) {
	if lock == nil {
		return Member{}, false
	}
	for _, member := range lock.Members {
		if member.Name == name {
			return member, true
		}
	}
	return Member{}, false
}

// digestObject renders the lock minus lock_sha256, the digest preimage.
func (lock *Lock) digestObject() map[string]any {
	members := make([]any, 0, len(lock.Members))
	for _, member := range lock.Members {
		members = append(members, member.object())
	}
	return map[string]any{
		"schema_version":  SchemaVersion,
		"manifest_sha256": lock.ManifestSHA256,
		"members":         members,
	}
}

// Object renders the complete lock including lock_sha256.
func (lock *Lock) Object() map[string]any {
	object := lock.digestObject()
	object["lock_sha256"] = lock.LockSHA256
	return object
}

// digest hashes the CCJ-1 bytes of the lock with lock_sha256 omitted.
func (lock *Lock) digest() (string, error) {
	canonical, err := protocoljson.MarshalCanonical(lock.digestObject())
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// Digest validates the lock structure and returns its lock_sha256.
func (lock *Lock) Digest() (string, error) {
	if lock == nil {
		return "", verr.New("lock", "source_selection_invalid: lock is nil")
	}
	if err := lock.validateStructure(); err != nil {
		return "", err
	}
	return lock.digest()
}

// Validate enforces the full lock shape and the lock_sha256 self-integrity.
func (lock *Lock) Validate() error {
	if lock == nil {
		return verr.New("lock", "source_selection_invalid: lock is nil")
	}
	if err := lock.validateStructure(); err != nil {
		return err
	}
	if !digestRE.MatchString(lock.LockSHA256) {
		return verr.New("lock_sha256", "source_selection_invalid: lock_sha256 must be sha256:<64 lowercase hex>")
	}
	expected, err := lock.digest()
	if err != nil {
		return err
	}
	if lock.LockSHA256 != expected {
		return verr.New("lock_sha256", "source_selection_invalid: lock_sha256 mismatch: lock was modified without refresh")
	}
	return nil
}

func (lock *Lock) validateStructure() error {
	if !digestRE.MatchString(lock.ManifestSHA256) {
		return verr.New("manifest_sha256", "source_selection_invalid: manifest_sha256 must be sha256:<64 lowercase hex>")
	}
	seen := map[string]string{}
	for index, member := range lock.Members {
		path := "members[" + strconv.Itoa(index) + "]"
		if err := member.Validate(path); err != nil {
			return err
		}
		folded := strings.ToLower(member.Name)
		if previous, ok := seen[folded]; ok {
			return verr.New(path+".name", "source_name_conflict: %s repeats destination %s", member.Name, previous)
		}
		seen[folded] = member.Name
		if index > 0 && bytes.Compare([]byte(lock.Members[index-1].Name), []byte(member.Name)) >= 0 {
			return verr.New(path+".name", "source_selection_invalid: members are not sorted by UTF-8 skill name")
		}
	}
	return nil
}

// ManifestDigest binds a lock to its declaring Skillfile: the CCJ-1 SHA-256
// of the entire parsed manifest, including declared paths, with no field
// excluded. Callers pass the bytes of a successfully parsed Skillfile.
func ManifestDigest(payload []byte) (string, error) {
	if err := protocoljson.Validate(payload); err != nil {
		return "", fmt.Errorf("source_selection_invalid: malformed Skillfile JSON: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return "", fmt.Errorf("source_selection_invalid: malformed Skillfile JSON: %w", err)
	}
	if _, ok := value.(map[string]any); !ok {
		return "", verr.New("Skillfile", "source_selection_invalid: manifest must contain a JSON object")
	}
	canonical, err := protocoljson.MarshalCanonical(value)
	if err != nil {
		return "", fmt.Errorf("source_selection_invalid: manifest is not CCJ-1 encodable: %w", err)
	}
	sum := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// CheckStale requires the lock to bind the given declaring Skillfile bytes.
// A changed manifest fails source_lock_stale; only explicit refresh may
// replace the lock.
func (lock *Lock) CheckStale(manifestPayload []byte) error {
	if lock == nil {
		return verr.New("lock", "source_selection_invalid: lock is nil")
	}
	digest, err := ManifestDigest(manifestPayload)
	if err != nil {
		return err
	}
	return lock.CheckStaleDigest(digest)
}

// CheckStaleDigest requires the lock to bind a manifest digest.
func (lock *Lock) CheckStaleDigest(digest string) error {
	if lock == nil {
		return verr.New("lock", "source_selection_invalid: lock is nil")
	}
	if lock.ManifestSHA256 != digest {
		return verr.New("manifest_sha256", "source_lock_stale: Skillfile changed since lock; explicit refresh required")
	}
	return nil
}

// CheckNames requires the locked member set to equal exactly the expected
// installed names from manifest expansion. Launch and install consume the
// locked set; a plan member absent from the lock fails
// source_member_missing, and a locked member outside the plan fails
// source_lock_stale. Callers pass expansion names, never live filesystem
// members, so live collection extras never widen the frozen set.
func (lock *Lock) CheckNames(expected []string) error {
	if lock == nil {
		return verr.New("lock", "source_selection_invalid: lock is nil")
	}
	locked := map[string]bool{}
	for _, member := range lock.Members {
		locked[member.Name] = true
	}
	want := append([]string(nil), expected...)
	sort.Strings(want)
	for _, name := range want {
		if !locked[name] {
			return verr.New("members", "source_member_missing: %s is not in the lock", name)
		}
	}
	planned := map[string]bool{}
	for _, name := range expected {
		planned[name] = true
	}
	for _, member := range lock.Members {
		if !planned[member.Name] {
			return verr.New("members", "source_lock_stale: lock contains %s outside the plan; explicit refresh required", member.Name)
		}
	}
	return nil
}

// CheckMembership requires the locked member records to equal exactly the
// expected frozen set: names, selection indexes, directories, package
// identities, and content hashes.
func (lock *Lock) CheckMembership(expected []Member) error {
	if lock == nil {
		return verr.New("lock", "source_selection_invalid: lock is nil")
	}
	want := append([]Member(nil), expected...)
	sort.SliceStable(want, func(i, j int) bool { return want[i].Name < want[j].Name })
	for _, member := range want {
		locked, ok := lock.Find(member.Name)
		if !ok {
			return verr.New("members", "source_member_missing: %s is not in the lock", member.Name)
		}
		if !locked.equalRecord(member) {
			return verr.New("members", "source_member_invalid: locked %s differs from the plan", member.Name)
		}
	}
	planned := map[string]bool{}
	for _, member := range expected {
		planned[member.Name] = true
	}
	for _, member := range lock.Members {
		if !planned[member.Name] {
			return verr.New("members", "source_lock_stale: lock contains %s outside the plan; explicit refresh required", member.Name)
		}
	}
	return nil
}

// PathIn returns the lock path for a project root.
func PathIn(projectRoot string) string {
	return filepath.Join(projectRoot, FileName)
}

// Parse decodes one lock payload under the strict reader discipline:
// protocol JSON, no unknown fields, full structural validation, and
// lock_sha256 self-integrity. Any serialization parses; only the value
// must verify.
func Parse(payload []byte) (*Lock, error) {
	if err := protocoljson.Validate(payload); err != nil {
		return nil, fmt.Errorf("source_selection_invalid: malformed lock JSON: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var raw any
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("source_selection_invalid: malformed lock JSON: %w", err)
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New("lock", "source_selection_invalid: lock must contain a JSON object")
	}
	if unknown := unknownFields(obj, "schema_version", "manifest_sha256", "members", "lock_sha256"); len(unknown) > 0 {
		return nil, verr.New("lock", "source_selection_invalid: unsupported field(s): %s", strings.Join(unknown, ", "))
	}
	version, ok := obj["schema_version"].(json.Number)
	if !ok || version.String() != strconv.Itoa(SchemaVersion) {
		return nil, verr.New("schema_version", "source_selection_invalid: unsupported lock schema_version; this lock requires a newer tool")
	}
	manifest, _ := obj["manifest_sha256"].(string)
	rawMembers, ok := obj["members"].([]any)
	if !ok {
		return nil, verr.New("members", "source_selection_invalid: lock requires field 'members' as a list")
	}
	members := make([]Member, 0, len(rawMembers))
	for index, entry := range rawMembers {
		member, err := parseMember(entry, "members["+strconv.Itoa(index)+"]")
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	digest, _ := obj["lock_sha256"].(string)
	lock := &Lock{ManifestSHA256: manifest, Members: members, LockSHA256: digest}
	if err := lock.Validate(); err != nil {
		return nil, err
	}
	return lock, nil
}

func parseMember(entry any, path string) (Member, error) {
	obj, ok := entry.(map[string]any)
	if !ok {
		return Member{}, verr.New(path, "source_member_invalid: must be an object")
	}
	if unknown := unknownFields(obj, "name", "selection", "directory", "package", "content_sha256"); len(unknown) > 0 {
		return Member{}, verr.New(path, "source_member_invalid: unsupported field(s): %s", strings.Join(unknown, ", "))
	}
	name, _ := obj["name"].(string)
	rawSelection, present := obj["selection"]
	if !present {
		return Member{}, verr.New(path+".selection", "source_selection_invalid: selection is required and must be null or an index")
	}
	selection, err := parseSelection(rawSelection, path+".selection")
	if err != nil {
		return Member{}, err
	}
	directory, _ := obj["directory"].(string)
	rawPackage, present := obj["package"]
	if !present {
		return Member{}, verr.New(path+".package", "source_member_invalid: package is required")
	}
	parsed, err := parsePackage(rawPackage, path+".package")
	if err != nil {
		return Member{}, err
	}
	content, _ := obj["content_sha256"].(string)
	member := Member{Name: name, Selection: selection, Directory: directory, Package: parsed, ContentSHA256: content}
	if err := member.Validate(path); err != nil {
		return Member{}, err
	}
	return member, nil
}

func parseSelection(raw any, path string) (*int, error) {
	if raw == nil {
		return nil, nil
	}
	number, ok := raw.(json.Number)
	if !ok {
		return nil, verr.New(path, "source_selection_invalid: selection must be null or a non-negative safe integer")
	}
	value, err := strconv.ParseInt(number.String(), 10, 64)
	if err != nil || strconv.FormatInt(value, 10) != number.String() || value < 0 || value > protocoljson.MaxSafeInteger {
		return nil, verr.New(path, "source_selection_invalid: selection must be null or a non-negative safe integer")
	}
	index := int(value)
	return &index, nil
}

func parsePackage(raw any, path string) (Package, error) {
	obj, ok := raw.(map[string]any)
	if !ok {
		return Package{}, verr.New(path, "source_member_invalid: package must be an object")
	}
	kind, _ := obj["kind"].(string)
	var parsed Package
	switch kind {
	case KindLocalSnapshot:
		if unknown := unknownFields(obj, "kind", "snapshot"); len(unknown) > 0 {
			return Package{}, verr.New(path, "source_member_invalid: local-snapshot admits only kind and snapshot")
		}
		snapshot, _ := obj["snapshot"].(string)
		parsed = Package{Kind: kind, Snapshot: snapshot}
	case KindNetworkGit:
		if unknown := unknownFields(obj, "kind", "repository", "commit", "directory"); len(unknown) > 0 {
			return Package{}, verr.New(path, "source_member_invalid: network-git admits only kind, repository, commit and directory")
		}
		repository, _ := obj["repository"].(string)
		commit, err := parseCommit(obj["commit"], path+".commit")
		if err != nil {
			return Package{}, err
		}
		directory, _ := obj["directory"].(string)
		parsed = Package{Kind: kind, Repository: repository, Commit: commit, Directory: directory}
	case KindConfiguredGit:
		if unknown := unknownFields(obj, "kind", "source", "commit", "directory"); len(unknown) > 0 {
			return Package{}, verr.New(path, "source_member_invalid: configured-git admits only kind, source, commit and directory")
		}
		source, _ := obj["source"].(string)
		commit, err := parseCommit(obj["commit"], path+".commit")
		if err != nil {
			return Package{}, err
		}
		directory, _ := obj["directory"].(string)
		parsed = Package{Kind: kind, Source: source, Commit: commit, Directory: directory}
	default:
		return Package{}, verr.New(path+".kind", "source_selection_invalid: unknown package kind %q", kind)
	}
	if err := parsed.Validate(path); err != nil {
		return Package{}, err
	}
	return parsed, nil
}

func parseCommit(raw any, path string) (Commit, error) {
	obj, ok := raw.(map[string]any)
	if !ok {
		return Commit{}, verr.New(path, "source_member_invalid: commit must be an object")
	}
	if unknown := unknownFields(obj, "object_format", "hex"); len(unknown) > 0 {
		return Commit{}, verr.New(path, "source_member_invalid: unsupported field(s): %s", strings.Join(unknown, ", "))
	}
	format, _ := obj["object_format"].(string)
	hexValue, _ := obj["hex"].(string)
	commit := Commit{ObjectFormat: format, Hex: hexValue}
	if err := commit.Validate(path); err != nil {
		return Commit{}, err
	}
	return commit, nil
}

func unknownFields(object map[string]any, allowed ...string) []string {
	set := map[string]bool{}
	for _, field := range allowed {
		set[field] = true
	}
	var unknown []string
	for field := range object {
		if !set[field] {
			unknown = append(unknown, field)
		}
	}
	sort.Strings(unknown)
	return unknown
}

// Read loads and validates the lock at path. A missing file returns the
// filesystem error, never an empty lock: absence is not validity.
func Read(path string) (*Lock, error) {
	payload, err := os.ReadFile(path) // #nosec G304 -- caller-supplied lock path
	if err != nil {
		return nil, err
	}
	lock, err := Parse(payload)
	if err != nil {
		return nil, fmt.Errorf("lock %s: %w", path, err)
	}
	return lock, nil
}

// Write validates the lock and stores its canonical bytes at path
// atomically. The lock must already carry its lock_sha256; New computes it.
func Write(path string, lock *Lock) error {
	if lock == nil {
		return verr.New("lock", "source_selection_invalid: lock is nil")
	}
	if err := lock.Validate(); err != nil {
		return err
	}
	canonical, err := protocoljson.MarshalCanonical(lock.Object())
	if err != nil {
		return err
	}
	return writeFileAtomic(path, canonical, 0o644, 0o755)
}

func writeFileAtomic(path string, payload []byte, fileMode, dirMode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), dirMode); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".lock-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer func() { _ = os.Remove(name) }()
	if err := tmp.Chmod(fileMode); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(payload); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(name, path); err != nil {
		return err
	}
	return nil
}
