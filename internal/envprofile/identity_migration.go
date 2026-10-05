package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/transaction"
)

// identityMigration stages complete entry replacements. Immutable v2 store
// entries are populated before activation; existing v1 entries are never moved
// or renamed, because another profile or a retained lock can still name them.
type identityMigration struct {
	root    string
	serial  int
	targets map[string]string // empty staged path means removal
}

func (m *identityMigration) nextStage() string {
	m.serial++
	return filepath.Join(m.root, fmt.Sprintf("entry-%d", m.serial))
}

func (m *identityMigration) cleanup() { _ = os.RemoveAll(m.root) }

func (m *identityMigration) bytes(live string, payload []byte) error {
	staged := m.nextStage()
	if err := os.WriteFile(staged, payload, 0o600); err != nil {
		return err
	}
	m.targets[live] = staged
	return nil
}

func (m *identityMigration) link(live, target string) error {
	staged := m.nextStage()
	if err := os.Symlink(target, staged); err != nil {
		return err
	}
	m.targets[live] = staged
	return nil
}

func (m *identityMigration) publish(op *operation) error {
	plan := transaction.Plan{TransactionID: newTransactionID(), ProjectIdentity: "profile-hash-migration"}
	paths := make([]string, 0, len(m.targets))
	for path := range m.targets {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		digest, err := transaction.DigestTarget(transaction.KindEntry, path)
		if err != nil {
			return err
		}
		plan.Targets = append(plan.Targets, transaction.Target{Class: "profile-identity", Identifier: path, Kind: transaction.KindEntry, LivePath: path, StagedSource: m.targets[path], PreimageDigest: digest})
	}
	if _, err := op.engine.Prepare(op.lock, plan); err != nil {
		return fmt.Errorf("prepare profile identity migration: %w", err)
	}
	if err := op.engine.Commit(op.lock, plan.TransactionID); err != nil {
		return fmt.Errorf("publish profile identity migration: %w", err)
	}
	return nil
}

func markerHashVersion(marker *envmarker.Marker) hashing.Version {
	if marker.Version == envmarker.VersionV3 {
		return hashing.Version(marker.HashVersion) // #nosec G115 -- nonnegative int versions fit uint64; negative versions cannot alias v1 or v2
	}
	return hashing.VersionV1
}

// identityMigrationInputs binds every legacy profile dependency to a read-only
// plan, including siblings outside --env. It never hashes credential payloads.
func identityMigrationInputs(req *MigrateRequest, profiles []string) (map[string]string, error) {
	inputs := map[string]string{}
	if hashing.WriteVersion() != hashing.VersionV2 {
		return inputs, nil
	}
	adapters, err := migrationAdapters("")
	if err != nil {
		return nil, err
	}
	for _, profile := range profiles {
		if err := validateProfileStoreState(req.Home, profile, nil, nil); err != nil {
			return nil, err
		}
		_, lock, _, err := loadResolveInputs(req.Home, profile)
		if err != nil {
			return nil, err
		}
		if lock.ContentHashVersion() != hashing.VersionV1 {
			continue
		}
		paths := map[string]bool{lockPath(req.Home, profile): true, sourcePath(req.Home, profile): true}
		for _, member := range lock.Members {
			paths[contextstore.EntryDir(req.Home, member.Kind, member.Name, member.PinKey())] = true
		}
		for _, adapter := range adapters {
			if err := validateMarkerBoundary(req.Home, profile, adapter.ID, nil); err != nil {
				return nil, err
			}
			home := ManagedHomeDir(req.Home, profile, adapter.ID)
			marker, err := envmarker.Read(home)
			if err != nil {
				return nil, err
			}
			paths[filepath.Join(home, envmarker.Name)] = true
			if marker == nil {
				continue
			}
			if markerHashVersion(marker) != lock.ContentHashVersion() {
				return nil, fmt.Errorf("marker/lock hash version mismatch for %s/%s", profile, adapter.ID)
			}
			for _, surface := range marker.Surfaces {
				for _, path := range surface.Paths {
					live := filepath.Join(home, filepath.FromSlash(path))
					rel, err := managedRelative(EnvRoot(req.Home), live)
					if err != nil {
						return nil, err
					}
					if _, err := managedPath(EnvRoot(req.Home), rel, false); err != nil {
						return nil, err
					}
					paths[live] = true
					target, err := os.Readlink(live)
					if err == nil && sameStoreTree(target, ProfileDir(req.Home, profile)) {
						paths[target] = true
					}

				}
			}
		}
		for path := range paths {
			digest, err := transaction.DigestTarget(transaction.KindEntry, path)
			if err != nil {
				return nil, err
			}
			inputs[path] = digest
		}
	}
	return inputs, nil
}

// prepareIdentityMigration rebuilds the full profile projection from verified
// pinned snapshots, retaining unrelated metadata. Credential changes, when
// supplied, join this same transaction rather than a separate marker rewrite.
func prepareIdentityMigration(req *ResolveRequest, profiles []string, report *MigrateReport) (_ *identityMigration, resultErr error) {
	root, err := os.MkdirTemp("", "curator-identity-migration-*")
	if err != nil {
		return nil, err
	}
	migration := &identityMigration{root: root, targets: map[string]string{}}
	defer func() {
		if resultErr != nil {
			migration.cleanup()
		}
	}()
	adapters, err := migrationAdapters("")
	if err != nil {
		return nil, err
	}
	for _, profile := range profiles {
		if err := validateProfileStoreState(req.Home, profile, nil, req.boundaryOwnerLookup); err != nil {
			return nil, err
		}
		source, oldLock, oldHash, err := loadResolveInputs(req.Home, profile)
		if err != nil {
			return nil, err
		}
		if oldLock.ContentHashVersion() != hashing.VersionV1 {
			continue
		}
		lock := *oldLock
		lock.Members = append([]contextlock.Member{}, oldLock.Members...)
		lock.SchemaVersion = contextlock.SchemaVersion2
		lock.HashVersion = 2
		for i, member := range lock.Members {
			if member.StateHash == "" {
				continue
			}
			entry := contextstore.EntryDir(req.Home, member.Kind, member.Name, member.PinKey())
			dir, key, err := contextstore.EnsureState(req.Home, member.Kind, member.Name, entry)
			if err != nil {
				return nil, err
			}
			actual, err := hashing.ContentSHA256WithVersion(dir, map[string]bool{}, hashing.VersionV2)
			if err != nil {
				return nil, err
			}
			if hashing.Normalize(actual) != key {
				return nil, fmt.Errorf("v2 state entry does not match its rehashed pin")
			}
			lock.Members[i].StateHash = key
		}
		canonical, err := lock.Canonical()
		if err != nil {
			return nil, err
		}
		hash := contextlock.HashBytes(canonical)
		if err := migration.bytes(lockPath(req.Home, profile), canonical); err != nil {
			return nil, err
		}
		for _, adapter := range adapters {
			if err := validateMarkerBoundary(req.Home, profile, adapter.ID, req.boundaryOwnerLookup); err != nil {
				return nil, err
			}
			home := ManagedHomeDir(req.Home, profile, adapter.ID)
			prior, err := envmarker.Read(home)
			if err != nil {
				return nil, err
			}
			if prior == nil {
				continue
			}
			if prior.Mode != envmarker.ModeManagedHome || prior.Profile.Name != profile || prior.Profile.LockSHA256 != hashing.Normalize(oldHash) || markerHashVersion(prior) != hashing.VersionV1 {
				return nil, fmt.Errorf("profile identity migration refuses inconsistent marker for %s/%s", profile, adapter.ID)
			}
			orderBefore, err := contextmaterialize.EmittedOrder(oldLock, contextmaterialize.Precedence{Winner: prior.Precedence.Winner, Placement: prior.Precedence.Placement})
			if err != nil {
				return nil, err
			}
			if len(prior.Members) != len(orderBefore) {
				return nil, fmt.Errorf("marker member count does not match legacy lock")
			}
			for i, member := range orderBefore {
				recorded := prior.Members[i]
				if recorded.Name != member.Name || recorded.StateSHA256 != member.StateHash || recorded.Commit != member.Commit || recorded.Weight != member.Weight || recorded.Overlay != member.Overlay {
					return nil, fmt.Errorf("marker member %s does not match legacy lock", member.Name)
				}
			}
			rr := *req
			rr.Profile = profile
			rr.EnvID = adapter.ID
			rr.Policy.PrecedenceWinner = prior.Precedence.Winner
			rr.Policy.PrecedencePlacement = prior.Precedence.Placement
			var credentialHome *MigrateHome
			if report != nil {
				for i := range report.Homes {
					h := &report.Homes[i]
					if h.Profile == profile && h.EnvID == adapter.ID {
						credentialHome = h
						break
					}
				}
			}
			// Identity-only siblings retain their recorded credential mode. Only an
			// inventoried credential home may adopt the migration plan's new mode.
			isolation := ""
			if prior.Passthrough != nil && len(*prior.Passthrough) > 0 {
				isolation = (*prior.Passthrough)[0].Isolation
				if prior.Version == envmarker.VersionV1 {
					isolation = envregistry.IsolationShared
				}
			}
			if credentialHome != nil {
				isolation = credentialHome.Isolation
			}
			if isolation != "" {
				rr.Machine.Isolation = map[string]map[string]string{profile: {adapter.ID: isolation}}
			}
			rr.Machine.Forms = map[string]string{}
			if surface, ok := prior.Surfaces[envmarker.SurfaceRootContext]; ok {
				rr.Machine.Forms[adapter.ID] = surface.Form
			}
			rr.Machine.SystemPromptFiles = map[string]string{}
			for _, path := range prior.Surfaces[envmarker.SurfaceSystemPrompt].Paths {
				if path == "APPEND_SYSTEM.md" {
					rr.Machine.SystemPromptFiles[profile] = "append"
				}
				if path == "SYSTEM.md" {
					rr.Machine.SystemPromptFiles[profile] = "replace"
				}
			}
			precedence := rr.Policy.Precedence()
			order, err := contextmaterialize.EmittedOrder(&lock, precedence)
			if err != nil {
				return nil, err
			}
			plan, err := assembleHome(&rr, source, &lock, hash, precedence, order, adapter, prior)
			if err != nil {
				return nil, err
			}
			// Managed entry ownership is established by the old marker. Newly
			// generated paths must be absent; no implicit takeover is permitted.
			recorded := map[string]bool{}
			for _, surface := range prior.Surfaces {
				for _, path := range surface.Paths {
					recorded[path] = true
				}
			}
			for path := range plan.copies {
				if !recorded[path] {
					return nil, fmt.Errorf("identity migration cannot add unrecorded surface %s", path)
				}
			}
			for path := range plan.links {
				if !recorded[path] {
					return nil, fmt.Errorf("identity migration cannot add unrecorded surface %s", path)
				}
			}
			// Preserve symlink fallback copies, including whole skill trees.
			for name, surface := range prior.Surfaces {
				if surface.Copies == nil {
					continue
				}
				for _, copy := range *surface.Copies {
					if copy.Reason != envmarker.ReasonSymlinkFallback {
						continue
					}
					target, ok := plan.links[copy.Path]
					if !ok {
						continue
					}
					live := filepath.Join(home, filepath.FromSlash(copy.Path))
					if data, ok := plan.docs[target]; ok {
						if err := migration.bytes(live, data); err != nil {
							return nil, err
						}
					} else {
						staged := migration.nextStage()
						info, err := os.Stat(target)
						if err != nil {
							return nil, err
						}
						if info.IsDir() {
							if err := copyTree(target, root, filepath.Base(staged)); err != nil {
								return nil, err
							}
						} else {
							data, err := os.ReadFile(target) // #nosec G304 -- store link target recomputed from the verified materialization plan
							if err != nil {
								return nil, err
							}
							if err := os.WriteFile(staged, data, 0o600); err != nil {
								return nil, err
							}
						}
						migration.targets[live] = staged
					}
					delete(plan.links, copy.Path)
					updated := plan.marker.Surfaces[name]
					copies := append([]envmarker.Copy{}, (*updated.Copies)...)
					copies = append(copies, copy)
					updated.Copies = &copies
					plan.marker.Surfaces[name] = updated
				}
			}
			for path, data := range plan.docs {
				// Rendered documents stage with the other targets; the
				// unified route validation below covers them with the
				// profiles-tree link discipline.
				if err := migration.bytes(path, data); err != nil {
					return nil, err
				}
			}
			for path, data := range plan.copies {
				if err := migration.bytes(filepath.Join(home, filepath.FromSlash(path)), data); err != nil {
					return nil, err
				}
			}
			for path, target := range plan.links {
				if err := migration.link(filepath.Join(home, filepath.FromSlash(path)), target); err != nil {
					return nil, err
				}
			}
			for path := range recorded {
				if _, ok := plan.copies[path]; ok {
					continue
				}
				if _, ok := plan.links[path]; ok {
					continue
				}
				live := filepath.Join(home, filepath.FromSlash(path))
				if _, ok := migration.targets[live]; !ok {
					migration.targets[live] = ""
				}
			}
			marker := *prior
			marker.Version = envmarker.VersionV3
			marker.HashVersion = 2
			marker.Profile = plan.marker.Profile
			marker.Members = plan.marker.Members
			marker.Surfaces = plan.marker.Surfaces
			if prior.Version == envmarker.VersionV1 {
				records, err := rr.credentialRecords(adapter, plan.isolation, "migrated")
				if err != nil {
					return nil, err
				}
				if credentialHome == nil && !legacyCredentialProjectionMatches(*prior.Passthrough, records) {
					return nil, fmt.Errorf("%s: legacy credentials require explicit env migrate before repair", envregistry.DiagCredentialConflict)
				}
				marker.Passthrough = &records
			}
			if report != nil {
				for _, h := range report.Homes {
					if h.Profile == profile && h.EnvID == adapter.ID && (len(h.Ops) > 0 || prior.Version == envmarker.VersionV1) {
						records := append([]envmarker.Passthrough{}, h.credentialRecords...)
						marker.Passthrough = &records
					}
				}
			}
			payload, err := marker.Marshal()
			if err != nil {
				return nil, err
			}
			if err := migration.bytes(filepath.Join(home, envmarker.Name), payload); err != nil {
				return nil, err
			}
		}
	}
	// Validate every live route without following a leaf symlink. The
	// transaction engine records and restores the leaf as an entry.
	// Managed-home targets validate against the environments root: the
	// hardened boundary the ordinary repair path enforces (preflight and
	// validateMarkerBoundary). Lock and rendered-document targets are
	// contained to the profiles tree with the same component link-walk
	// the ordinary store-document publisher uses. The manager home
	// itself is not a hardened boundary root — Install publishes
	// profile records through plain parents and real operator homes
	// carry inherited access — so it cannot serve as a ValidateRoute
	// root: on Windows its inherited DACL fails the owner-only mutation
	// check for fixtures and operators alike.
	envsRoot := EnvRoot(req.Home)
	profilesRoot := ProfilesDir(req.Home)
	for path := range migration.targets {
		if _, err := managedRelative(envsRoot, path); err == nil {
			if err := pathboundary.ValidateRouteWithOwner(envsRoot, filepath.Dir(path), boundaryOwner(req.boundaryOwnerLookup)); err != nil {
				return nil, err
			}
			continue
		}
		rel, err := managedRelative(profilesRoot, path)
		if err != nil {
			return nil, fmt.Errorf("identity migration target %s is outside the managed trees", path)
		}
		if _, err := managedPath(profilesRoot, rel, false); err != nil {
			return nil, err
		}
	}
	if report != nil {
		for _, operation := range report.Ops() {
			live := filepath.Join(ManagedHomeDir(req.Home, operation.Profile, operation.EnvID), filepath.FromSlash(operation.Path))
			if err := validateMigrationOp(live, operation); err != nil {
				return nil, err
			}
			if operation.Kind == MigrateOpUnlink {
				migration.targets[live] = ""
			} else {
				if err := migration.link(live, operation.To); err != nil {
					return nil, err
				}
			}
		}
		// v2 homes participating in a mixed-profile credential plan still need
		// their credential record updates in the joint transaction.
		changed, _, err := changedMarkers(report)
		if err != nil {
			return nil, err
		}
		for path, payload := range changed {
			if _, ok := migration.targets[path]; !ok {
				if err := migration.bytes(path, payload); err != nil {
					return nil, err
				}
			}
		}
	}
	return migration, nil
}

func identityInputHash(inputs map[string]string) string {
	var text strings.Builder
	paths := make([]string, 0, len(inputs))
	for path := range inputs {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		fmt.Fprintf(&text, "%q %s\n", path, inputs[path])
	}
	return contextmaterialize.FileHash([]byte(text.String()))
}

// Schema-1 metadata can be expanded only if its recorded link strategies are
// exactly the projection of the current declaration; never invent evidence for
// a changed credential binding during an identity-only repair.
func legacyCredentialProjectionMatches(legacy, records []envmarker.Passthrough) bool {
	projected := map[string]string{}
	for _, entry := range records {
		if entry.Path == "" {
			continue
		}
		strategy := entry.Strategy
		if strategy == envregistry.StrategyKeyringPreferred && entry.Backend == "file" {
			strategy = envregistry.StrategyFileLink
		}
		projected[entry.Path] = strategy
	}
	if len(legacy) != len(projected) {
		return false
	}
	for _, entry := range legacy {
		if projected[entry.Path] != entry.Strategy {
			return false
		}
	}
	return true
}
