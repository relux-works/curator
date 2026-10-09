package install

import (
	"encoding/hex"
	"fmt"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/sourcelock"
)

// legacyMigrated reports whether one legacy-lane (Skillfile schema 1) node
// escalates to a package marker: no legacy schema can record manifest
// version 9 or a subdirectory selection, so those installations record a
// package marker on the schema-1 lane too — v6 exactly for schema-9
// installations, v5 for a subdirectory-selected installation of an earlier
// schema (skillfile-sources §4, draft-sources-v2).
func legacyMigrated(node *closure.Node) bool {
	return node.Spec.SchemaVersion == 9 || normalizedNodeDirectory(node) != "."
}

// legacyCommitFormat resolves the locked-commit object format from the
// resolved commit hex: 40 lowercase hex is sha1, 64 is sha256. Anything
// else is not a locked object the package arms can carry.
func legacyCommitFormat(commit string) (string, error) {
	raw, err := hex.DecodeString(commit)
	if err != nil || hex.EncodeToString(raw) != commit {
		if err == nil {
			err = fmt.Errorf("commit %q is not lowercase hex", commit)
		}
		return "", err
	}
	switch len(raw) {
	case 20:
		return "sha1", nil
	case 32:
		return "sha256", nil
	default:
		return "", fmt.Errorf("length %d is neither sha1 nor sha256", len(raw))
	}
}

// legacyLockPackage stages the frozen source-types schema-1 package
// identity of one legacy-lane node: the exact arms the draft resolve lane
// locks for Git members. A node with a canonical network identity records
// the network-git arm with the normalized selected directory; a node from
// a local configured source records the configured-git arm, whose
// directory is always "." — roots select "." by construction, and a
// subdirectory selection without a network identity is refused instead of
// recorded with a directory the arm forbids. The constructors validate
// the disjoint arm shapes, so only a statable identity is ever staged.
func legacyLockPackage(node *closure.Node) (sourcelock.Package, error) {
	directory := normalizedNodeDirectory(node)
	format, err := legacyCommitFormat(node.Resolved.Commit)
	if err != nil {
		return sourcelock.Package{}, fmt.Errorf("source_member_invalid: locked commit for %s is not a full object id: %w", node.Name, err)
	}
	commit := sourcelock.Commit{ObjectFormat: format, Hex: node.Resolved.Commit}
	if node.Identity != "" {
		return sourcelock.NetworkGitPackage(node.Identity, commit, directory)
	}
	if directory != "." {
		return sourcelock.Package{}, fmt.Errorf("source_selection_invalid: selecting a repository subdirectory for %s requires a canonical network repository identity", node.Name)
	}
	source := node.Decl.Source
	if source == "" {
		source = node.Name
	}
	return sourcelock.ConfiguredGitPackage(source, commit)
}

// legacyLanePlan is the package-marker migration of one legacy-lane
// install, derived once from the resolved closure so the marker, the
// runtime leaf, the build receipts, and the audit record all bind the
// same staged package identity. Packages, RuntimeKeys, and BuildPackages
// carry exactly the migrated nodes; LockSHA256 is the specified §3 digest
// of the effective lock those markers share, or "" when the plan carries
// no lock (a dry run builds no marker, so it stages no lock).
type legacyLanePlan struct {
	Packages      map[string]*marker.Package
	LockSHA256    string
	RuntimeKeys   map[string]string
	BuildPackages map[string]*buildmeta.Package
	Generation    []byte
}

// planLegacyLane derives the package-marker migration of one legacy-lane
// install from the parsed declaring Skillfile payload, the effective
// manifest (project declarations plus applicable hybrid entries), and the
// resolved closure. It returns nil when no node migrates, leaving every
// legacy marker, runtime key, and receipt byte-identical.
//
// Packages, runtime keys, and build identities stage on every path,
// including dry runs: a dry-run build plan must derive the same receipt-3
// keys the real install will record, and an unstatable package (a
// subdirectory without a network identity, a malformed commit) fails here,
// before any audit, registry, or compiler work. The effective lock — the
// content-hashed member list — is built only when needLock is set (a real
// install with at least one migrated node), because only a written marker
// carries its digest.
func planLegacyLane(payload []byte, effective *manifest.Manifest, nodes []*closure.Node, needLock bool) (*legacyLanePlan, error) {
	var migrated []*closure.Node
	for _, node := range nodes {
		if legacyMigrated(node) {
			migrated = append(migrated, node)
		}
	}
	if len(migrated) == 0 {
		return nil, nil
	}
	plan := &legacyLanePlan{
		Packages:      make(map[string]*marker.Package, len(migrated)),
		RuntimeKeys:   make(map[string]string, len(migrated)),
		BuildPackages: make(map[string]*buildmeta.Package, len(migrated)),
	}
	// The effective lock binds the resolved full closure, so a real
	// install stages every member; a dry run stages only the migrated
	// nodes whose build keys it plans, and never refuses on a legacy
	// node the real install would record byte-identically.
	staged := nodes
	if !needLock {
		staged = migrated
	}
	lockPackages := make(map[string]sourcelock.Package, len(staged))
	for _, node := range staged {
		pkg, err := legacyLockPackage(node)
		if err != nil {
			return nil, err
		}
		lockPackages[node.Name] = pkg
	}
	for _, node := range migrated {
		lockPackage := lockPackages[node.Name]
		markerPackage := draftMarkerPackage(lockPackage)
		digest, err := markerPackage.Digest()
		if err != nil {
			return nil, err
		}
		key, err := runtimestore.SourceV1Key(digest)
		if err != nil {
			return nil, err
		}
		plan.Packages[node.Name] = markerPackage
		plan.RuntimeKeys[node.Name] = key
		plan.BuildPackages[node.Name] = receiptPackage(lockPackage)
	}
	if !needLock {
		return plan, nil
	}
	lock, err := legacyEffectiveLock(payload, effective, nodes, lockPackages)
	if err != nil {
		return nil, err
	}
	plan.LockSHA256 = lock.LockSHA256
	plan.Generation, err = encodeLegacyGeneration(payload, lock, nodes)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

// legacyEffectiveLock builds the validated skillfile-lock schema-1
// document one legacy-lane install binds its package markers to
// (skillfile-sources §3): the CCJ-1 SHA-256 of the entire parsed declaring
// Skillfile over the resolved full closure — every member carrying its
// source-relative directory, package identity, and context content hash,
// root members carrying their zero-based selection index. The schema-1
// lane does not consume a frozen project lock. It retains the lock and matching
// manifest as installed-generation evidence; sourcelock.New computes
// lock_sha256 as the CCJ-1 SHA-256 of the
// lock with only lock_sha256 omitted, exactly the wire meaning §4 requires
// the marker to carry ("binds the installed selection and declared ref
// through the validated lock and matching manifest"). Selection indexes
// run over the effective manifest, so applicable hybrid roots keep their
// declaring position; a retargeted hybrid declaration still moves a member
// package and therefore the digest.
func legacyEffectiveLock(payload []byte, effective *manifest.Manifest, nodes []*closure.Node, lockPackages map[string]sourcelock.Package) (*sourcelock.Lock, error) {
	manifestSHA, err := sourcelock.ManifestDigest(payload)
	if err != nil {
		return nil, err
	}
	indexByName := make(map[string]int, len(effective.Skills))
	for index, decl := range effective.Skills {
		if decl.Name == "" {
			continue
		}
		if _, seen := indexByName[decl.Name]; !seen {
			indexByName[decl.Name] = index
		}
	}
	members := make([]sourcelock.Member, 0, len(nodes))
	for _, node := range nodes {
		directory := normalizedNodeDirectory(node)
		var selection *int
		if index, ok := indexByName[node.Name]; ok {
			value := index
			selection = &value
		}
		content, err := closure.ContentHashFor(node.Snapshot, node.Spec)
		if err != nil {
			return nil, err
		}
		member := sourcelock.Member{
			Name: node.Name, Selection: selection, Directory: directory,
			Package: lockPackages[node.Name], ContentSHA256: content,
		}
		if err := member.Validate("members"); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return sourcelock.New(manifestSHA, members)
}
