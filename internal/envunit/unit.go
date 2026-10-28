// Package envunit defines the closed durable evidence shared by origin
// admission and the transaction engine. Parsing a descriptor is not admission:
// the engine must also bind its actual targets, journal phase and live pins.
package envunit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/staging"
)

const Version = "environment-unit/1"
const MaxBytes = 4 << 20

type Action string

const (
	Provision    Action = "provision"
	Adopt        Action = "adopt"
	Maintain     Action = "maintain"
	Rebind       Action = "profile-update"
	GlobalExtend Action = "global-extend"
	Takeover     Action = "retained-takeover"
	Remove       Action = "remove"
	Purge        Action = "remove-purge"
	Reuse        Action = "purged-name-reuse"
)

var identifier = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)
var hash = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var identity = regexp.MustCompile(`^[0-9]+:[0-9]+$`)

type Manager struct {
	Path     string `json:"path"`
	Identity string `json:"identity"`
	Owner    string `json:"owner"`
}

// Image contains metadata and, only for manager-owned noncredential bytes, a
// digest. Credential entries are absent/link/empty-directory images exclusively.
// Directory images identify the entry, never a recursive content snapshot.
type Image struct {
	Protection string `json:"protection,omitempty"`
	Kind       string `json:"kind"`
	Identity   string `json:"identity,omitempty"`
	Owner      string `json:"owner,omitempty"`
	Mode       uint32 `json:"mode,omitempty"`
	Digest     string `json:"digest,omitempty"`
	LinkTarget string `json:"link_target,omitempty"`
}
type RecordImage struct {
	Kind   string `json:"kind"`
	Bytes  []byte `json:"bytes,omitempty"`
	Digest string `json:"digest,omitempty"`
}
type ProfileTransition struct {
	Name      string      `json:"name"`
	Source    string      `json:"source"`
	OldSource string      `json:"old_source,omitempty"`
	Closure   []string    `json:"closure"`
	OldLock   RecordImage `json:"old_lock"`
	NewLock   RecordImage `json:"new_lock"`
}
type HomeImage struct {
	Kind   string `json:"kind"`
	Marker []byte `json:"marker,omitempty"`
	Origin []byte `json:"origin,omitempty"`
}
type Home struct {
	Profile         string    `json:"profile"`
	Adapter         string    `json:"adapter"`
	Mode            string    `json:"mode"`
	Target          string    `json:"target"`
	Root            string    `json:"root"`
	DirectoryBefore Image     `json:"directory_before"`
	DirectoryAfter  Image     `json:"directory_after"`
	Before          HomeImage `json:"before"`
	After           HomeImage `json:"after"`
}
type Target struct {
	Class  string `json:"class"`
	Path   string `json:"path"`
	Before Image  `json:"before"`
	After  Image  `json:"after"`
}
type Unit struct {
	LeasePreimage     *ReferencePreimage      `json:"lease_preimage,omitempty"`
	Legacy            *LegacyPublication      `json:"legacy_publication,omitempty"`
	Record            string                  `json:"record"`
	ID                string                  `json:"id"`
	Manager           Manager                 `json:"manager"`
	Action            Action                  `json:"action"`
	Profiles          []ProfileTransition     `json:"profiles"`
	Native            []Native                `json:"native,omitempty"`
	Homes             []Home                  `json:"homes"`
	Targets           []Target                `json:"targets"`
	BoundaryProof     staging.DurableSnapshot `json:"boundary_proof"`
	LedgerPublication *LedgerPublication      `json:"ledger_publication,omitempty"`
	RetainedEntry     *RetainedEntry          `json:"retained_entry,omitempty"`
}

// RetainedEntry binds consumption to one exact private receipt. HomeLocks is
// required only for retained-L1 takeover: installed L1 and old-home L0 are
// different authorities, both captured durably, never a generic prev-lock waiver.
type RetainedEntry struct {
	Serialized []byte                 `json:"serialized,omitempty"`
	ID         string                 `json:"id"`
	Digest     string                 `json:"digest"`
	HomeLocks  map[string]RecordImage `json:"home_locks,omitempty"`
}

// LedgerPublication binds the structured shared-container CAS separately from
// its byte replacement. Its postimage is validated structurally and as an
// ordinary journal target. Keeping that digest in the enclosing journal avoids
// a recursive hash when a terminal entry retains its creating unit evidence.
type LedgerPublication struct {
	Path           string   `json:"path"`
	PreimageDigest string   `json:"preimage_digest"`
	Revision       uint64   `json:"revision"`
	NextRevision   uint64   `json:"next_revision"`
	EntryIDs       []string `json:"entry_ids"`
}

func invalid(format string, args ...any) error {
	return fmt.Errorf(envorigin.DiagInvalid+": durable unit: "+format, args...)
}
func absolute(s string) bool {
	return s != "" && filepath.IsAbs(s) && filepath.Clean(s) == s && !strings.ContainsAny(s, "\x00\r\n")
}
func within(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && (rel == "." || rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel))
}
func (a Action) valid() bool {
	switch a {
	case Provision, Adopt, Maintain, Rebind, GlobalExtend, Takeover, Remove, Purge, Reuse:
		return true
	}
	return false
}

// ObserveManager captures no-follow directory identity/owner from one metadata
// observation. Unsupported transports refuse; HOME and selectors are irrelevant.
func ObserveManager(path string) (Manager, error) {
	if !absolute(path) {
		return Manager{}, invalid("manager must be clean absolute")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return Manager{}, invalid("manager metadata: %v", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return Manager{}, invalid("manager is not an owned directory")
	}
	if err := pathboundary.CheckPrivateFile(path, info); err != nil {
		return Manager{}, invalid("manager boundary: %v", err)
	}
	owner, err := pathboundary.DefaultOwnerLookup()(path, info)
	if err != nil {
		return Manager{}, invalid("manager owner: %v", err)
	}
	id, err := metadataIdentity(path, info)
	if err != nil {
		return Manager{}, invalid("manager identity: %v", err)
	}
	return Manager{Path: path, Identity: id, Owner: string(owner)}, nil
}

func (m Manager) Validate() error {
	if !absolute(m.Path) || !identity.MatchString(m.Identity) || m.Owner == "" || len(m.Owner) > 256 || strings.ContainsAny(m.Owner, "\x00\r\n") {
		return invalid("manager identity grammar")
	}
	return nil
}
func (m Manager) Verify() error {
	if err := m.Validate(); err != nil {
		return err
	}
	actual, err := ObserveManager(m.Path)
	if err != nil {
		return err
	}
	if actual != m {
		return invalid("manager identity changed")
	}
	return nil
}

func Parse(raw []byte, manager Manager) (*Unit, error) {
	if len(raw) == 0 || len(raw) > MaxBytes {
		return nil, invalid("unit size")
	}
	var u Unit
	if err := protocoljson.DecodeClosed(raw, &u); err != nil {
		return nil, invalid("closed unit: %v", err)
	}
	if u.Manager != manager {
		return nil, invalid("unit manager binding")
	}
	if err := u.Validate(); err != nil {
		return nil, err
	}
	return &u, nil
}
func (u *Unit) Marshal() ([]byte, error) {
	if err := u.Validate(); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(u)
	if len(raw) > MaxBytes {
		return nil, invalid("unit size")
	}
	return raw, err
}

func (i Image) validate(credential bool) error { return i.validateImage(credential, false) }
func (i Image) validateImage(credential, desired bool) error {
	if i.Kind == "absent" {
		if i != (Image{Kind: "absent"}) {
			return invalid("absent entry carries metadata")
		}
		return nil
	}
	if (!identity.MatchString(i.Identity) && !(desired && i.Identity == "" && (i.Kind == "file" || i.Kind == "link" || i.Kind == "tree"))) || i.Owner == "" || i.Mode > 07777 || !(desired && i.Identity == "" && i.Protection == "") && !ValidProtection(i.Protection) {
		return invalid("entry identity/owner/mode")
	}
	switch i.Kind {
	case "tree":
		if credential || !hash.MatchString(i.Digest) || i.LinkTarget != "" {
			return invalid("owned tree image must bind only noncredential projection bytes")
		}
	case "file":
		if credential || !hash.MatchString(i.Digest) || i.LinkTarget != "" {
			return invalid("file image cannot snapshot a credential")
		}
	case "link":
		if i.Digest != "" || i.LinkTarget == "" || strings.ContainsRune(i.LinkTarget, 0) {
			return invalid("link image must contain only its target text")
		}
	case "directory", "empty-directory":
		if i.Digest != "" || i.LinkTarget != "" || credential && i.Kind != "empty-directory" {
			return invalid("directory image is metadata only")
		}
	default:
		return invalid("unknown entry image kind")
	}
	return nil
}
func (i RecordImage) validate() error {
	if i.Kind == "absent" {
		if len(i.Bytes) != 0 || i.Digest != "" {
			return invalid("absent record carries bytes")
		}
		return nil
	}
	if i.Kind != "file" || len(i.Bytes) == 0 || len(i.Bytes) > 1<<20 || envorigin.Digest(i.Bytes) != i.Digest {
		return invalid("record byte identity")
	}
	return nil
}

// Validate checks self-contained declaration invariants. Actual filesystem
// effects/aliases, target ordering and reachable live states remain mandatory
// engine checks; this method cannot authorize a forward write by itself.
func (u *Unit) Validate() error {
	if u == nil || u.Record != Version || !identifier.MatchString(u.ID) || !u.Action.valid() || !absolute(u.Manager.Path) || !identity.MatchString(u.Manager.Identity) || u.Manager.Owner == "" {
		return invalid("unit header")
	}
	if err := u.validateReferenceCapture(); err != nil {
		return err
	}
	if err := u.Manager.Validate(); err != nil {
		return err
	}
	needsRetained := u.Action == Takeover || u.Action == Purge || u.Action == Reuse
	if needsRetained != (u.RetainedEntry != nil) {
		return invalid("action/retained receipt binding")
	}
	if r := u.RetainedEntry; r != nil {
		if !identifier.MatchString(r.ID) || !hash.MatchString(r.Digest) || u.LedgerPublication == nil {
			return invalid("retained receipt identity/CAS")
		}
		if u.Action != Takeover && r.HomeLocks != nil {
			return invalid("historical home lock exception is takeover-only")
		}
	}
	if ledger := u.LedgerPublication; ledger != nil {
		if ledger.Path != filepath.Join(u.Manager.Path, ".state", "environment-lifecycle.json") || (ledger.PreimageDigest != "absent" && !hash.MatchString(ledger.PreimageDigest)) || ledger.Revision == ^uint64(0) || ledger.NextRevision != ledger.Revision+1 || len(ledger.EntryIDs) > 1024 {
			return invalid("structured ledger publication binding")
		}
		seen := map[string]bool{}
		for _, id := range ledger.EntryIDs {
			if !identifier.MatchString(id) || seen[id] {
				return invalid("structured ledger entry identity")
			}
			seen[id] = true
		}
		if (ledger.Revision == 0) != (ledger.PreimageDigest == "absent") {
			return invalid("structured ledger absent/revision mismatch")
		}
	}
	if len(u.Profiles) == 0 || len(u.Profiles) > 256 || (len(u.Homes) == 0 && len(u.Native) == 0 && u.Action != Remove && u.Action != Purge && u.Action != Reuse) || len(u.Homes)+len(u.Native) > 2048 || (len(u.Targets) == 0 && !(u.Action == Purge && len(u.Homes) == 0 && len(u.Native) == 0 && u.LedgerPublication != nil)) || len(u.Targets) > 8192 {
		return invalid("unit inventory bounds")
	}
	profiles := map[string]ProfileTransition{}
	for _, p := range u.Profiles {
		if p.OldSource != "" && !hash.MatchString(p.OldSource) {
			return invalid("old source identity")
		}
		if !identifier.MatchString(p.Name) || !hash.MatchString(p.Source) || len(p.Closure) == 0 || len(p.Closure) > 4096 {
			return invalid("profile/source/closure identity")
		}
		if _, ok := profiles[p.Name]; ok {
			return invalid("duplicate profile")
		}
		profiles[p.Name] = p
		refs := map[string]bool{}
		for _, ref := range p.Closure {
			if !hash.MatchString(ref) || refs[ref] {
				return invalid("closure identity/duplicate")
			}
			refs[ref] = true
		}
		if err := p.OldLock.validate(); err != nil {
			return err
		}
		if err := p.NewLock.validate(); err != nil {
			return err
		}
		for _, image := range []RecordImage{p.OldLock, p.NewLock} {
			if image.Kind == "absent" {
				continue
			}
			lock, err := contextlock.Parse(image.Bytes)
			if err != nil {
				return invalid("installed lock grammar: %v", err)
			}
			canonical, err := lock.Canonical()
			if err != nil || !bytes.Equal(canonical, image.Bytes) {
				return invalid("installed lock is not exact canonical authority")
			}
		}
		if u.Action != Remove && u.Action != Purge && p.NewLock.Kind != "file" {
			return invalid("forward unit requires exact new lock")
		}
		if (u.Action == Adopt || u.Action == Maintain || u.Action == Takeover) && (!bytes.Equal(p.OldLock.Bytes, p.NewLock.Bytes) || p.OldLock.Digest != p.NewLock.Digest) {
			return invalid("adopt/maintain/takeover cannot change installed lock")
		}
	}
	if u.Action == Takeover {
		if len(u.RetainedEntry.HomeLocks) != len(profiles) {
			return invalid("takeover lacks exact old-home locks")
		}
		for name, lock := range u.RetainedEntry.HomeLocks {
			_, known := profiles[name]
			firstUse := lock.Kind == "absent"
			for _, home := range u.Homes {
				if home.Profile == name {
					firstUse = false
				}
			}
			// A genuine first-use retained publication has no L0 or managed
			// homes. Exact predecessor validation still binds this absent image
			// to the serialized creating global unit; it grants no home waiver.
			if !known || lock.Kind != "file" && !firstUse {
				return invalid("takeover old-home lock outside group")
			}
			if err := lock.validate(); err != nil {
				return err
			}
		}
	}
	if err := u.validateLegacy(); err != nil {
		return err
	}
	if err := u.validateNative(profiles); err != nil {
		return err
	}
	targets := map[string]Target{}
	for _, t := range u.Targets {
		if err := u.ValidateReadScope(t.Class, t.Path); err != nil {
			return err
		}
		if !absolute(t.Path) || (!within(u.Manager.Path, t.Path) && !u.nativeTarget(t) && !u.LegacyOwns(t.Class, t.Path) && !((t.Class == "parent" || t.Class == "home-directory") && t.Before.Kind == "absent" && t.After.Kind == "directory" && u.LegacyParent(t.Path))) {
			return invalid("target outside manager namespace")
		}
		if _, ok := targets[t.Path]; ok {
			return invalid("duplicate target")
		}
		targets[t.Path] = t
		switch t.Class {
		case "profile-mcp", "parent", "home-directory", "owned-bytes", "owned-tree", "marker", "origin", "credential-entry", "activation", "profile-record", "installed-profile", "quarantine", BackupQuarantine, "native-marker", "reference-record":
		default:
			if !u.LegacyOwns(t.Class, t.Path) {
				return invalid("unknown target class")
			}
		}
		if t.Class == "installed-profile" && (t.Before.Kind != "tree" || t.After.Kind != "absent") {
			return invalid("installed profile removal image differs")
		}
		if t.Class == "reference-record" && (u.Action != Remove || t.Path != filepath.Join(u.Manager.Path, "state", "project-leases", "v1", "ledger.json")) {
			return invalid("reference publication outside exact removal ledger")
		}
		if t.Class == "activation" && t.After.Kind != "absent" {
			bound := false
			for name := range profiles {
				bound = bound || t.After.Kind == "file" && t.After.Digest == envorigin.Digest([]byte(name+"\n"))
			}
			if !bound {
				return invalid("activation does not select a whole-unit profile")
			}
		}
		if t.Path == filepath.Join(u.Manager.Path, ".state", "environment-lifecycle.json") {
			return invalid("shared ledger requires typed structured publication")
		}
		if err := t.Before.validate(t.Class == "credential-entry"); err != nil {
			return err
		}
		if err := t.After.validateImage(t.Class == "credential-entry", true); err != nil {
			return err
		}
		if (t.Before.Kind == "tree" || t.After.Kind == "tree") && !recursiveClass(t.Class) {
			return invalid("recursive image outside an owned projection target")
		}
		if (t.Class == "origin" || t.Class == "lifecycle" || t.Class == "56-hook-approvals") && t.After.Kind != "absent" && (t.After.Kind != "file" || t.After.Mode&0077 != 0) {
			return invalid("private authority postimage")
		}
		if t.Class == BackupQuarantine {
			if _, ok := u.BackupAdapter(t.Path); !ok || t.Before.Kind != "directory" || t.After.Kind != "absent" {
				return invalid("backup quarantine lacks exact generation custody")
			}
		}
		if t.Class == "quarantine" && (u.Action != Purge || t.After.Kind != "absent" || t.Before.Kind != "directory") {
			return invalid("opaque home removal requires purge action")
		}
		proof, ok := u.BoundaryProof.Targets[t.Path]
		if !ok || !absolute(proof.Canonical) || len(proof.Ancestors) == 0 {
			return invalid("missing durable target/parent proof")
		}
	}
	if u.Legacy != nil {
		for _, entry := range u.Legacy.Entries {
			t, ok := targets[entry.Path]
			if !ok || t.Class != entry.Class {
				return invalid("legacy declared target missing from journal")
			}
		}
	}
	for _, n := range u.Native {
		for _, rel := range n.Paths {
			if _, ok := targets[filepath.Join(n.Root, filepath.FromSlash(rel))]; !ok {
				return invalid("native declared target missing from actual journal")
			}
		}
		marker, ok := targets[filepath.Join(n.Root, envmarker.Name)]
		if !ok || marker.Class != "native-marker" || marker.After.Kind != "file" || marker.After.Mode&0077 != 0 || marker.After.Digest != envorigin.Digest(n.MarkerAfter) {
			return invalid("native marker target differs from custody")
		}
		if len(n.MarkerBefore) == 0 {
			if marker.Before.Kind != "absent" {
				return invalid("native absent marker preimage differs")
			}
		} else if marker.Before.Kind != "file" || marker.Before.Digest != envorigin.Digest(n.MarkerBefore) {
			return invalid("native marker preimage differs")
		}
		if n.DirectoryBefore != n.DirectoryAfter {
			dir, ok := targets[n.Root]
			if !ok || dir.Before != n.DirectoryBefore || dir.After != n.DirectoryAfter || dir.Class != "home-directory" {
				return invalid("native directory change omitted")
			}
		}
	}
	for _, p := range u.Profiles {
		if p.OldLock.Kind == "absent" {
			if u.Action != GlobalExtend && u.Action != Reuse {
				return invalid("absent installed lock requires explicit first-use global or name-reuse unit")
			}
			for _, name := range []string{"lock.json", "source.json"} {
				target, ok := targets[filepath.Join(u.Manager.Path, "profiles", p.Name, name)]
				want := p.Source
				if name == "lock.json" {
					want = p.NewLock.Digest
				}
				if !ok || target.Class != "profile-record" || target.Before.Kind != "absent" || target.After.Kind != "file" || target.After.Digest != want {
					return invalid("first-use installed target binding differs")
				}
			}
		}
		if p.OldSource != "" {
			target, ok := targets[filepath.Join(u.Manager.Path, "profiles", p.Name, "source.json")]
			if !ok || target.Class != "profile-record" || target.Before.Digest != p.OldSource || target.After.Digest != p.Source || (u.Action != Rebind && u.Action != GlobalExtend) {
				return invalid("source transition is not an exact record target")
			}
		}
	}
	if len(u.BoundaryProof.Targets) != len(targets) {
		return invalid("boundary proof differs from target inventory")
	}
	homes := map[string]bool{}
	for _, h := range u.Homes {
		p, ok := profiles[h.Profile]
		if !ok || h.Target != envorigin.ManagedTarget(u.Manager.Path, h.Profile, h.Adapter) || h.Root != envorigin.ManagedRoot(u.Manager.Path, h.Profile, h.Adapter) || h.Mode != "managed-home" {
			return invalid("home/profile/adapter/mode binding")
		}
		if homes[h.Target] {
			return invalid("duplicate home")
		}
		homes[h.Target] = true
		if err := h.DirectoryBefore.validate(false); err != nil {
			return err
		}
		if err := h.DirectoryAfter.validate(false); err != nil {
			return err
		}
		for _, state := range []struct {
			home      HomeImage
			directory Image
		}{{h.Before, h.DirectoryBefore}, {h.After, h.DirectoryAfter}} {
			if state.home.Kind == "absent" && state.directory.Kind != "absent" && !(u.Action == Provision && state.directory.Kind == "directory") || state.home.Kind != "absent" && state.directory.Kind != "directory" {
				return invalid("home directory pre/postimage differs from authority state")
			}
		}
		if h.DirectoryBefore != h.DirectoryAfter {
			target, ok := targets[h.Root]
			if !ok || target.Before != h.DirectoryBefore || target.After != h.DirectoryAfter || (target.Class != "home-directory" && target.Class != "quarantine") {
				return invalid("home directory change is missing from complete target set")
			}
		}
		beforeLock := p.OldLock
		if u.Action == Takeover {
			beforeLock = u.RetainedEntry.HomeLocks[h.Profile]
		}
		if err := u.validateHomeImage(h, h.Before, beforeLock, false); err != nil {
			return err
		}
		afterLock := p.NewLock
		if u.Action == Remove {
			afterLock = p.OldLock
		}
		if err := u.validateHomeImage(h, h.After, afterLock, true); err != nil {
			return err
		}
		if h.Before.Kind == "recorded" && h.After.Kind == "recorded" {
			before, _ := envorigin.CaptureFacts(h.Before.Origin)
			after, _ := envorigin.CaptureFacts(h.After.Origin)
			if !bytes.Equal(before, after) {
				return invalid("maintenance/rebind changed original capture bytes")
			}
		}
		if u.Action == Provision && h.Before.Kind != "absent" || u.Action == Adopt && (h.Before.Kind != "legacy" || h.After.Kind != "recorded" || !bytes.Equal(h.Before.Marker, h.After.Marker)) {
			return invalid("new/adopt preconditions")
		}
		if u.Action == Remove && (!bytes.Equal(h.Before.Marker, h.After.Marker) || !bytes.Equal(h.Before.Origin, h.After.Origin) || h.DirectoryBefore != h.DirectoryAfter) {
			return invalid("remove-retain changes retained home authority")
		}
		authName := "auth.json"
		if h.Adapter == "claude_code" {
			authName = ".credentials.json"
		}
		if h.Adapter == "muse" {
			authName = filepath.Join("config", "muse", "auth.json")
		}
		if credential, ok := targets[filepath.Join(h.Target, authName)]; ok && h.Adapter != "opencode" && credential.Class != "credential-entry" {
			return invalid("credential bytes cannot enter an ordinary target snapshot")
		}
		for _, item := range []struct {
			name, class   string
			before, after []byte
		}{{envmarker.Name, "marker", h.Before.Marker, h.After.Marker}, {envorigin.Name, "origin", h.Before.Origin, h.After.Origin}} {
			if u.Action == Purge {
				target, ok := targets[h.Root]
				if !ok || target.Class != "quarantine" || h.After.Kind != "absent" {
					return invalid("purge lacks exact opaque owned-root target")
				}
				continue
			}
			path := filepath.Join(h.Target, item.name)
			target, changed := targets[path]
			if bytes.Equal(item.before, item.after) {
				if changed && (target.Before.Digest != target.After.Digest || target.Before.Kind != target.After.Kind) {
					return invalid("unchanged authority target changes identity")
				}
				continue
			}
			if !changed || target.Class != item.class {
				return invalid("authority change missing from complete target set")
			}
			for _, x := range []struct {
				data  []byte
				image Image
			}{{item.before, target.Before}, {item.after, target.After}} {
				if len(x.data) == 0 {
					if x.image.Kind != "absent" {
						return invalid("authority absence pre/postimage mismatch")
					}
				} else if x.image.Kind != "file" || x.image.Digest != envorigin.Digest(x.data) {
					return invalid("authority bytes differ from target pre/postimage")
				}
			}
		}
	}
	return nil
}
func (u *Unit) validateHomeImage(h Home, image HomeImage, lock RecordImage, after bool) error {
	switch image.Kind {
	case "absent":
		if len(image.Marker) != 0 || len(image.Origin) != 0 {
			return invalid("absent home carries authority")
		}
		return nil
	case "legacy":
		if after || u.Action != Adopt || len(image.Origin) != 0 {
			return invalid("legacy home is only an adoption preimage")
		}
	case "recorded":
		if len(image.Origin) == 0 {
			return invalid("recorded home lacks origin")
		}
	default:
		return invalid("unknown home state")
	}
	if lock.Kind != "file" || len(image.Marker) == 0 {
		return invalid("home lacks exact lock/marker")
	}
	marker, err := envmarker.Parse(image.Marker)
	if err != nil {
		return invalid("marker schema: %v", err)
	}
	if marker.Mode != envmarker.ModeManagedHome || marker.Profile.Name != h.Profile || strings.TrimPrefix(lock.Digest, "sha256:") != marker.Profile.LockSHA256 {
		return invalid("marker binding differs from unit")
	}
	if marker.CredentialSource != nil && marker.CredentialSource.Source != nil {
		return invalid("node credential source cannot enter per-home unit")
	}
	if image.Kind == "recorded" {
		r, err := envorigin.ParseCaptured(image.Origin, envorigin.Binding{Profile: h.Profile, LockSHA256: lock.Digest, ManagedTarget: h.Target}, h.Adapter, u.Manager.Path)
		if err != nil {
			return err
		}
		if after && (u.Action == Provision || u.Action == Reuse || h.Before.Kind == "absent") && r.Capture.Method != "provisioning" || after && u.Action == Adopt && r.Capture.Method != "adoption" {
			return invalid("capture method differs from action")
		}
	}
	return nil
}

// Progress is supplied from the strictly decoded real journal and fresh target
// metadata, never from an arbitrary admission callback. Rename-window handling
// needs exact backup/stage evidence in addition to this stable-state validator.
type Progress struct {
	State string
	Live  Image
}

func (u *Unit) ValidateProgress(phase string, states []Progress) error {
	if err := u.Validate(); err != nil {
		return err
	}
	if len(states) != len(u.Targets) {
		return invalid("journal target count differs from admitted unit")
	}
	seenPending := false
	for i, p := range states {
		target := u.Targets[i]
		switch phase {
		case "preparing", "prepared":
			if p.State != "pending" || p.Live != target.Before {
				return invalid("live mutation before committing phase")
			}
		case "committing":
			if p.State == "pending" {
				seenPending = true
				if p.Live != target.Before {
					return invalid("remaining target preimage changed")
				}
			} else if p.State != "committed" || seenPending || p.Live != target.After {
				return invalid("unreachable forward target prefix")
			}
		case "cleanup":
			if p.State != "committed" || p.Live != target.After {
				return invalid("cleanup lacks exact committed postconditions")
			}
		case "rollback_cleanup":
			if p.State != "rolled_back" || p.Live != target.Before {
				return invalid("rollback cleanup lacks exact preconditions")
			}
		default:
			return invalid("unsupported or incomplete phase proof")
		}
	}
	return nil
}
