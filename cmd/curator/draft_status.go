package main

// Draft source currentness (skillfile-sources §4): `curator status` for a
// schema-2 project on the draft lane.
//
// The legacy drift surface resolves live checkouts and compares legacy
// identity fields, neither of which exists here: status consumes the
// frozen lock only — never a collection rescan, branch advance, or
// snapshot replacement — and compares package, lock, attestation, and
// substitution against the effective plan the read-only dry run derived,
// in addition to every retained comparison (installed presence, marker
// validity, content hash). Changed registry/status/key, substitution
// identifier, declared ref (through lock staleness), package, or lock
// makes the installation non-current. Missing or unreadable evidence is
// never current; checking stays read-only and returns nonzero through
// the existing --check verdict.

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"

	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/sourcelock"
)

// errDraftManifestMoved fails a draft status check closed when the
// manifest no longer parses as the schema-2 document the caller peeked.
var errDraftManifestMoved = errors.New("the Skillfile changed while status was classifying it; re-run status")

// draftStatusApplies reports whether one project takes the schema-2 status
// lane. Schema 1 and missing manifests keep the legacy surface.
func draftStatusApplies(manifestRoot string) bool {
	payload, err := os.ReadFile(manifest.PathIn(manifestRoot)) // #nosec G304 -- operator-selected project Skillfile
	if err != nil {
		return false
	}
	var peek struct {
		SchemaVersion int `json:"schema_version"`
	}
	if json.Unmarshal(payload, &peek) != nil {
		return false
	}
	return peek.SchemaVersion == 2
}

// draftStatusDrift compares every locked member against its installed
// marker. effective carries the registry evidence the read-only dry run
// selected, by skill name; it is compared, never trusted. An error means
// the frozen inputs themselves are unreadable or no longer bind the
// declaration — a race with a concurrent resolve, or a manifest edit
// between the dry run and classification — and the caller fails closed,
// read-only, without a verdict.
func draftStatusDrift(manifestRoot, skillsDir string, effective map[string]*marker.Attestation) (map[string]string, error) {
	payload, err := os.ReadFile(manifest.PathIn(manifestRoot)) // #nosec G304 -- operator-selected project Skillfile
	if err != nil {
		return nil, err
	}
	projectManifest, err := manifest.ParseBytes(payload, manifest.PathIn(manifestRoot))
	if err != nil {
		return nil, err
	}
	if projectManifest.SchemaVersion != 2 {
		// The manifest moved under the status check (the caller only
		// diverts schema-2 projects here): no verdict, fail closed.
		return nil, errDraftManifestMoved
	}
	lock, err := sourcelock.Read(sourcelock.PathIn(manifestRoot))
	if err != nil {
		return nil, err
	}
	drift := make(map[string]string, len(lock.Members))
	if err := lock.CheckStale(payload); err != nil {
		// The lock no longer binds the declaration: the row set cannot
		// be derived from the new manifest without a collection rescan,
		// which status must not do, so a stale lock is not a
		// per-member verdict at all. Fail closed with the refusal.
		return nil, err
	}
	for _, member := range lock.Members {
		drift[member.Name] = classifyDraftMember(skillsDir, member, lock.LockSHA256, effective[member.Name])
	}
	return drift, nil
}

// classifyDraftMember is the per-member currentness verdict: installed
// presence, marker validity, content hash, frozen package, binding lock,
// substitution absence, and recorded attestation against the effective
// plan's selected evidence. Every unknown is non-current; only an exact
// match is up-to-date.
func classifyDraftMember(skillsDir string, member sourcelock.Member, lockSHA256 string, effective *marker.Attestation) string {
	installed := filepath.Join(skillsDir, member.Name)
	if _, err := os.Lstat(installed); err != nil {
		if os.IsNotExist(err) {
			return stateNotInstalled
		}
		return stateUnresolvable
	}
	recorded := marker.Read(installed)
	if recorded == nil {
		state, _ := markerRefusal(installed)
		return state
	}
	actualHash, err := hashing.ContentSHA256(installed, nil)
	if err != nil {
		return stateUnresolvable
	}
	if actualHash != recorded.ContentSHA256 {
		return stateContentDrift
	}
	if recorded.Package == nil {
		// A legacy marker can never describe a locked package; the
		// installation predates the lock and must be installed again.
		return stateNeedsInstall
	}
	recordedDigest, err := recorded.Package.Digest()
	if err != nil {
		return stateUnresolvable
	}
	memberDigest, err := member.Package.Digest()
	if err != nil {
		return stateUnresolvable
	}
	if recordedDigest != memberDigest || recorded.LockSHA256 != lockSHA256 {
		return stateNeedsInstall
	}
	// The effective plan on the draft lane never substitutes: frozen
	// consumption carries no substitution identifier, and install
	// refuses development substitutions with draft selectors before any
	// planning. A recorded identifier therefore disagrees with the plan.
	if recorded.Substituted != "" {
		return stateNeedsInstall
	}
	if !reflect.DeepEqual(recorded.Attestation, effective) {
		return stateNeedsInstall
	}
	return stateUpToDate
}
