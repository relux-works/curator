package envprofile

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/skillintegrity"
)

// inspectStoreAvailability distinguishes an absent leaf from a present unsafe
// entry. It checks every selected member before returning availability failures;
// an earlier absent entry must never conceal a later structural violation.
// The low-level named-target validators deliberately keep rejecting absence.
func inspectStoreAvailability(home string, lock *contextlock.Lock, owner pathboundary.OwnerLookup) ([]*storeEntryFailure, error) {
	root := contextstore.Root(home)
	var absent []*storeEntryFailure
	for _, member := range lock.Members {
		entry := memberEntryPath(home, member)
		failure := func(err error) *storeEntryFailure {
			return &storeEntryFailure{Member: member, Path: entry, Check: boundaryCheck(err), Err: err}
		}
		parent := filepath.Dir(entry)
		if err := pathboundary.ValidateRouteWithOwner(root, parent, boundaryOwner(owner)); err != nil {
			return nil, failure(err)
		}
		parentInfo, err := os.Lstat(parent)
		if err != nil || !parentInfo.IsDir() {
			return nil, failure(&pathboundary.Failure{Root: root, Path: parent, Check: pathboundary.CheckRegular, Reason: "store entry parent is not an existing directory", Cause: err})
		}
		_, err = os.Lstat(entry)
		if errors.Is(err, fs.ErrNotExist) {
			// Recheck the no-follow route so disappearance of an intermediate
			// ancestor cannot turn a failed lookup into leaf availability.
			if routeErr := pathboundary.ValidateRouteWithOwner(root, parent, boundaryOwner(owner)); routeErr != nil {
				return nil, failure(routeErr)
			}
			absent = append(absent, failure(&pathboundary.Failure{Root: root, Path: entry, Check: pathboundary.CheckRegular, Cause: err}))
			continue
		}
		if err != nil {
			return nil, failure(&pathboundary.Failure{Root: root, Path: entry, Check: pathboundary.CheckRegular, Cause: err})
		}
		one := &contextlock.Lock{Members: []contextlock.Member{member}}
		if unsafe := validateNamedStoreBoundaries(home, one, owner); unsafe != nil {
			return nil, unsafe
		}
		if member.Kind == contextlock.KindSkill {
			if err := skillintegrity.ValidateLegacyBoundary(entry); err != nil {
				return nil, err
			}
		}
		if failure := validateRecordedStorePaths(home, lock, member, entry, owner); failure != nil {
			return nil, failure
		}
	}
	return absent, nil
}

func (req *ResolveRequest) validateRepairStoreStructure() error {
	_, lock, _, err := loadResolveInputsWithOwner(req.Home, req.Profile, req.readRegularFile, req.boundaryOwnerLookup)
	if err != nil {
		return err
	}
	_, err = inspectStoreAvailability(req.Home, lock, req.boundaryOwnerLookup)
	return err
}

func storeBoundaryRefusal(err error) bool {
	var failure *storeEntryFailure
	return errors.As(err, &failure) || skillintegrity.IsCode(err, "skill_integrity_boundary_invalid")
}

func absentStoreRepairRefusal(failure *storeEntryFailure) error {
	return fmt.Errorf("%s: store entry remained absent after repair: %w", DiagRepairFailed, failure)
}

// Presence comes from the verified recorded roster or the existing declaration
// contract, never from a new universal filename rule for native skills.
func validateRecordedStorePaths(home string, lock *contextlock.Lock, member contextlock.Member, entry string, owner pathboundary.OwnerLookup) *storeEntryFailure {
	var paths []string
	if member.Commit != "" {
		directory := skillSnapshotDirectory(member.Kind, member.Directory)
		if entry == contextstore.EntryDir(home, member.Kind, member.Name, member.Commit) {
			directory = ""
		}
		roster, err := gitops.PinnedDirectorySnapshotPaths(newGitManager(home).repoDir(member.Source), member.Commit, directory)
		if err != nil {
			var unavailable *gitops.ObjectUnavailableError
			if errors.As(err, &unavailable) {
				return &storeEntryFailure{Member: member, Path: entry, Check: diagPinHash, Err: &pinUnavailable{cause: fmt.Errorf("pinned commit object unavailable for member %s/%s: %w", member.Kind, member.Name, err)}}
			}
			return &storeEntryFailure{Member: member, Path: entry, Check: pathboundary.CheckRegular, Err: fmt.Errorf("pinned snapshot path metadata: %w", err)}
		}
		for _, path := range roster {
			paths = append(paths, filepath.Join(entry, filepath.FromSlash(path)))
		}
	} else {
		root := packageRoot(entry, member.Directory)
		switch member.Kind {
		case contextlock.KindContext:
			paths = append(paths, filepath.Join(root, contextpkg.ManifestName))
			manifest, err := contextpkg.LoadManifest(root)
			if err == nil {
				for _, module := range manifest.Modules {
					paths = append(paths, filepath.Join(root, contextpkg.ContextDir, filepath.FromSlash(module.Path)))
				}
			} else {
				var failed *os.PathError
				if errors.As(err, &failed) {
					paths = append(paths, failed.Path)
				}
			}
		case contextlock.KindMCP:
			paths = append(paths, filepath.Join(root, contextpkg.MCPManifestName))
		case contextlock.KindSkill:
			for _, evidence := range lock.SkillIntegrity {
				if evidence.Manifest.Name == member.Name && evidence.Manifest.StateSHA256 == member.StateHash && evidence.Manifest.Directory == member.Directory {
					for _, file := range evidence.Manifest.Files {
						paths = append(paths, filepath.Join(root, filepath.FromSlash(file.Path)))
					}
				}
			}
		}
	}
	for _, path := range paths {
		if err := pathboundary.ValidateSnapshotWithinWithOwner(entry, path, boundaryOwner(owner)); err != nil {
			return &storeEntryFailure{Member: member, Path: path, Check: boundaryCheck(err), Err: err}
		}
	}
	return nil
}
