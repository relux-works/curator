package transaction

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/text/unicode/norm"
)

type targetNamespacePath struct {
	owner string
	kind  string
	path  string
	key   string
	// entry marks a path the manager owns as the directory entry it is. Its
	// final component is not resolved and its identity is read with Lstat, so an
	// owned symbolic link is never confused with whatever it points at.
	entry           bool
	caseInsensitive bool
	normInsensitive bool
}

// resolvedNamespacePath is one declared path as a single validation pass sees
// it. The pass names O(P) distinct paths but compares O(P^2) pairs of them, and
// every question a pair asks — how the key splits into components, what object
// the path currently names — has the same answer for every pair the path takes
// part in. Reading those answers once per pass and holding them here is what
// keeps the filesystem work at O(P) while the comparison itself stays exhaustive.
//
// The snapshot is strictly per pass. Nothing here outlives the slice
// validateIndependentTargetNamespaces builds, so every saveJournal call resolves
// and re-reads the live filesystem from scratch, and a symlink or alias that
// changes between two saves is seen by the second one.
type resolvedNamespacePath struct {
	targetNamespacePath
	// volume and parts are the key split for comparison, and volumeNFD and
	// partsNFD are the same split with each component NFD-normalized.
	// Normalization is applied per component and depends only on the component,
	// so pre-normalizing asks exactly the comparison the pair would have asked.
	volume    string
	parts     []string
	volumeNFD string
	partsNFD  []string
	// identityRead records that this pass has already asked the filesystem what
	// the path names; identityInfo and identityErr are that one answer, returned
	// unchanged to every later pair.
	identityRead bool
	identityInfo os.FileInfo
	identityErr  error
}

// resolveNamespacePath performs the per-path half of the pairwise sweep once.
// The filesystem identity is deliberately not read here: the original sweep only
// asks for it when a pair survives the containment test, and reading it eagerly
// would report an inspection failure for a path that a containment overlap would
// have rejected first.
func resolveNamespacePath(candidate targetNamespacePath) resolvedNamespacePath {
	volume, parts := namespaceComponents(candidate.key)
	resolved := resolvedNamespacePath{
		targetNamespacePath: candidate,
		volume:              volume,
		parts:               parts,
		volumeNFD:           norm.NFD.String(volume),
		partsNFD:            make([]string, len(parts)),
	}
	for index, part := range parts {
		resolved.partsNFD[index] = norm.NFD.String(part)
	}
	return resolved
}

// identity answers what object the path names, asking the filesystem at most
// once per pass. A failure is remembered exactly as it was returned, so a pass
// that cannot interrogate a path still fails closed at the same pair it would
// have failed at before.
func (resolved *resolvedNamespacePath) identity() (os.FileInfo, error) {
	if !resolved.identityRead {
		resolved.identityInfo, resolved.identityErr = namespaceIdentity(resolved.targetNamespacePath)
		resolved.identityRead = true
	}
	return resolved.identityInfo, resolved.identityErr
}

func validateIndependentTargetNamespaces(targets []TargetRecord, reserved ...targetNamespacePath) error {
	paths := make([]resolvedNamespacePath, 0, len(targets)*7+len(reserved))
	for index := range targets {
		target := &targets[index]
		entry := target.Kind == KindEntry
		for _, candidate := range []struct {
			kind string
			path string
		}{
			{kind: "live", path: target.LivePath},
			{kind: "staged", path: target.StagedPath},
			{kind: "backup", path: target.BackupPath},
			{kind: "rollback", path: target.RollbackPath},
		} {
			if candidate.path == "" {
				continue
			}
			key, err := canonicalNamespaceTargetPath(candidate.path, entry)
			if err != nil {
				return fmt.Errorf("target %d %s path: %w", index, candidate.kind, err)
			}
			caseInsensitive, err := namespaceCaseInsensitive(key)
			if err != nil {
				return fmt.Errorf("target %d %s path: determine filesystem case behavior: %w", index, candidate.kind, err)
			}
			normInsensitive, err := namespaceNormalizationInsensitive(key)
			if err != nil {
				return fmt.Errorf("target %d %s path: determine filesystem normalization behavior: %w", index, candidate.kind, err)
			}
			paths = append(paths, resolveNamespacePath(targetNamespacePath{
				owner:           fmt.Sprintf("target %d", index),
				kind:            candidate.kind,
				path:            candidate.path,
				key:             key,
				entry:           entry,
				caseInsensitive: caseInsensitive,
				normInsensitive: normInsensitive,
			}))
			if candidate.kind != "live" {
				tombPath := candidate.path + ".delete"
				tombKey, err := canonicalNamespaceTargetPath(tombPath, entry)
				if err != nil {
					return fmt.Errorf("target %d %s tomb path: %w", index, candidate.kind, err)
				}
				paths = append(paths, resolveNamespacePath(targetNamespacePath{
					owner:           fmt.Sprintf("target %d", index),
					kind:            candidate.kind + " cleanup tomb",
					path:            tombPath,
					key:             tombKey,
					entry:           entry,
					caseInsensitive: caseInsensitive,
					normInsensitive: normInsensitive,
				}))
			}
		}
	}
	for _, candidate := range reserved {
		key, err := canonicalNamespacePath(candidate.path)
		if err != nil {
			return fmt.Errorf("%s %s path: %w", candidate.owner, candidate.kind, err)
		}
		caseInsensitive, err := namespaceCaseInsensitive(key)
		if err != nil {
			return fmt.Errorf("%s %s path: determine filesystem case behavior: %w", candidate.owner, candidate.kind, err)
		}
		candidate.key = key
		candidate.caseInsensitive = caseInsensitive
		normInsensitive, err := namespaceNormalizationInsensitive(key)
		if err != nil {
			return fmt.Errorf("%s %s path: determine filesystem normalization behavior: %w", candidate.owner, candidate.kind, err)
		}
		candidate.normInsensitive = normInsensitive
		paths = append(paths, resolveNamespacePath(candidate))
	}
	for leftIndex := range paths {
		for rightIndex := leftIndex + 1; rightIndex < len(paths); rightIndex++ {
			left := &paths[leftIndex]
			right := &paths[rightIndex]
			overlap, err := namespacePathsOverlap(left, right)
			if err != nil {
				return err
			}
			if overlap {
				return fmt.Errorf("%s %s path %q overlaps %s %s path %q", left.owner, left.kind, left.path, right.owner, right.kind, right.path)
			}
		}
	}
	return nil
}

// canonicalNamespaceTargetPath canonicalizes one target path for independence
// checking. A byte target is resolved completely, which is what makes two
// targets aliasing one object through a link detectable. An owned directory
// entry keeps its final component: the entry is the target, so resolving it
// would compare the manager's own link against whatever it points at.
func canonicalNamespaceTargetPath(path string, entry bool) (string, error) {
	if !entry {
		return canonicalNamespacePath(path)
	}
	if path == "" || !validText(path) || !filepath.IsAbs(path) {
		return "", fmt.Errorf("path is not valid absolute filesystem text")
	}
	absolute := filepath.Clean(path)
	parent := filepath.Dir(absolute)
	if parent == absolute {
		return absolute, nil
	}
	resolvedParent, err := canonicalNamespacePath(parent)
	if err != nil {
		return "", err
	}
	return filepath.Join(resolvedParent, filepath.Base(absolute)), nil
}

func canonicalNamespacePath(path string) (string, error) {
	if path == "" || !validText(path) || !filepath.IsAbs(path) {
		return "", fmt.Errorf("path is not valid absolute filesystem text")
	}
	absolute := filepath.Clean(path)
	resolved, err := filepath.EvalSymlinks(absolute)
	if err == nil {
		return filepath.Clean(resolved), nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("resolve symlinks for %q: %w", path, err)
	}

	prefix := absolute
	missing := make([]string, 0, 4)
	for {
		parent := filepath.Dir(prefix)
		if parent == prefix {
			return "", fmt.Errorf("resolve existing prefix for %q: %w", path, err)
		}
		missing = append(missing, filepath.Base(prefix))
		prefix = parent
		resolved, err = filepath.EvalSymlinks(prefix)
		if err == nil {
			for index := len(missing) - 1; index >= 0; index-- {
				resolved = filepath.Join(resolved, missing[index])
			}
			return filepath.Clean(resolved), nil
		}
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("resolve symlinks for %q: %w", path, err)
		}
	}
}

func namespacePathsOverlap(left, right *resolvedNamespacePath) (bool, error) {
	caseInsensitive := left.caseInsensitive || right.caseInsensitive
	normInsensitive := left.normInsensitive || right.normInsensitive
	if namespaceContains(left, right, caseInsensitive, normInsensitive) || namespaceContains(right, left, caseInsensitive, normInsensitive) {
		return true, nil
	}
	leftInfo, leftErr := left.identity()
	rightInfo, rightErr := right.identity()
	if leftErr != nil && !os.IsNotExist(leftErr) {
		return false, fmt.Errorf("inspect %s %s path: %w", left.owner, left.kind, leftErr)
	}
	if rightErr != nil && !os.IsNotExist(rightErr) {
		return false, fmt.Errorf("inspect %s %s path: %w", right.owner, right.kind, rightErr)
	}
	return leftErr == nil && rightErr == nil && os.SameFile(leftInfo, rightInfo), nil
}

// namespaceIdentity reads the object a path currently names, complete. A byte
// target is resolved through links so an alias of another target is detected;
// an owned directory entry is read as itself, because replacing the manager's
// own link cannot disturb the object at its destination.
//
// The stat alone is not the whole answer on every host — see
// completeNamespaceIdentity — so the identity is completed here, while the pass
// still holds the object it named, rather than at the comparison that consumes
// it. Everything the sweep records as one pass's answer is therefore fixed by
// the time identity() returns it.
func namespaceIdentity(candidate targetNamespacePath) (os.FileInfo, error) {
	info, err := namespaceStat(candidate)
	if err != nil {
		return nil, err
	}
	return completeNamespaceIdentity(candidate, info)
}

func namespaceStat(candidate targetNamespacePath) (os.FileInfo, error) {
	if candidate.entry {
		return os.Lstat(candidate.path)
	}
	return os.Stat(candidate.path)
}

func namespaceContains(parent, child *resolvedNamespacePath, caseInsensitive, normInsensitive bool) bool {
	parentVolume, parentParts := parent.volume, parent.parts
	childVolume, childParts := child.volume, child.parts
	if normInsensitive {
		parentVolume, parentParts = parent.volumeNFD, parent.partsNFD
		childVolume, childParts = child.volumeNFD, child.partsNFD
	}
	if !namespaceComponentEqual(parentVolume, childVolume, caseInsensitive) || len(parentParts) > len(childParts) {
		return false
	}
	for index := range parentParts {
		if !namespaceComponentEqual(parentParts[index], childParts[index], caseInsensitive) {
			return false
		}
	}
	return true
}

func namespaceComponents(path string) (string, []string) {
	clean := filepath.Clean(path)
	volume := filepath.VolumeName(clean)
	remainder := clean[len(volume):]
	parts := strings.FieldsFunc(remainder, func(value rune) bool {
		if value == rune(filepath.Separator) {
			return true
		}
		return runtime.GOOS == "windows" && (value == '/' || value == '\\')
	})
	return volume, parts
}

// namespaceComponentEqual compares two path components that have already been
// normalized as the pair requires. Normalization moved to resolveNamespacePath
// because it depends only on the component, while case folding stays here
// because it depends on the pair.
func namespaceComponentEqual(left, right string, caseInsensitive bool) bool {
	if caseInsensitive {
		return strings.EqualFold(left, right)
	}
	return left == right
}
