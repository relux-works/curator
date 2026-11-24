package buildrepo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/protocoljson"
)

// Collect removes unreferenced external receipt-v2 artifacts after grace and
// retains exactly the snapshot roots reachable from retained receipts. It is
// called only while the manager-home mutation lock is held by the lifecycle
// layer; this package independently revalidates the protected boundary.
func Collect(root string, referenced []string, now time.Time, grace time.Duration) ([]string, error) {
	return CollectChecked(root, referenced, now, grace, func() error { return nil })
}

// CollectChecked revalidates the acquisition that produced referenced after
// potentially blocking reads and immediately before each artifact/snapshot removal.
func CollectChecked(root string, referenced []string, now time.Time, grace time.Duration, check func() error) (removed []string, err error) {
	removed, _, err = collectChecked(root, referenced, now, grace, check, false)
	return removed, err
}

// PreviewChecked classifies external artifacts and snapshots that collection
// would remove without changing the cache root.
func PreviewChecked(root string, referenced []string, now time.Time, grace time.Duration, check func() error) (wouldRemove []string, err error) {
	_, wouldRemove, err = collectChecked(root, referenced, now, grace, check, true)
	return wouldRemove, err
}

func collectChecked(root string, referenced []string, now time.Time, grace time.Duration, check func() error, dryRun bool) (removed, wouldRemove []string, err error) {
	if check == nil {
		return nil, nil, fmt.Errorf("external sweep ownership check is required")
	}
	if err := check(); err != nil {
		return nil, nil, err
	}
	store := &DiskProtectedStore{Root: root}
	if _, statErr := os.Lstat(root); os.IsNotExist(statErr) {
		return nil, nil, nil
	}
	if err := store.prepareFor(false); err != nil {
		return nil, nil, err
	}
	keep := map[string]bool{}
	for _, key := range referenced {
		keep[strings.TrimPrefix(key, "sha256:")] = true
	}
	reachableSnapshots := map[string]bool{}
	for _, version := range []int{LegacyReceiptSchemaVersion, SourceAwareReceiptSchemaVersion} {
		removed, wouldRemove, err = collectArtifacts(store, filepath.Join(root, ArtifactsDir(version)), keep, reachableSnapshots, now, grace, dryRun, removed, wouldRemove, check)
		if err != nil {
			return removed, wouldRemove, err
		}
	}
	snapshotsRoot := filepath.Join(root, "snapshots")
	snapshots, readErr := os.ReadDir(snapshotsRoot)
	if readErr != nil && !os.IsNotExist(readErr) {
		return removed, wouldRemove, readErr
	}
	for _, entry := range snapshots {
		if entry.IsDir() && !reachableSnapshots[entry.Name()] {
			if err := check(); err != nil {
				return removed, wouldRemove, err
			}
			if dryRun {
				wouldRemove = append(wouldRemove, "external-snapshot:sha256:"+entry.Name())
				continue
			}
			if err := os.RemoveAll(filepath.Join(snapshotsRoot, entry.Name())); err != nil {
				return removed, wouldRemove, err
			}
			removed = append(removed, "external-snapshot:sha256:"+entry.Name())
		}
	}
	sort.Strings(removed)
	sort.Strings(wouldRemove)
	return removed, wouldRemove, nil
}

// collectArtifacts sweeps one artifact namespace and records the snapshot
// roots its retained receipts still reach.
func collectArtifacts(store *DiskProtectedStore, artifactsRoot string, keep, reachableSnapshots map[string]bool, now time.Time, grace time.Duration, dryRun bool, removed, wouldRemove []string, check func() error) ([]string, []string, error) {
	entries, readErr := os.ReadDir(artifactsRoot)
	if readErr != nil && !os.IsNotExist(readErr) {
		return removed, wouldRemove, readErr
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(artifactsRoot, entry.Name())
		info, infoErr := entry.Info()
		if infoErr != nil {
			return removed, wouldRemove, infoErr
		}
		if !keep[entry.Name()] && (grace <= 0 || !info.ModTime().Add(grace).After(now)) {
			if err := check(); err != nil {
				return removed, wouldRemove, err
			}
			if dryRun {
				wouldRemove = append(wouldRemove, "external:sha256:"+entry.Name())
				continue
			}
			if err := os.RemoveAll(path); err != nil {
				return removed, wouldRemove, err
			}
			removed = append(removed, "external:sha256:"+entry.Name())
			continue
		}
		receipt, err := store.readProtectedFile(filepath.Join(path, "receipt.json"), 4<<20)
		if err != nil || protocoljson.Validate(receipt) != nil {
			return removed, wouldRemove, fmt.Errorf("retained external receipt %s is unreadable", entry.Name())
		}
		var object map[string]any
		decoder := json.NewDecoder(strings.NewReader(string(receipt)))
		decoder.UseNumber()
		if decoder.Decode(&object) != nil {
			return removed, wouldRemove, fmt.Errorf("retained external receipt %s is invalid", entry.Name())
		}
		if key, ok := snapshotKeyFromReceipt(object); ok {
			reachableSnapshots[strings.TrimPrefix(key, "sha256:")] = true
		}
	}
	return removed, wouldRemove, nil
}

func snapshotKeyFromReceipt(receipt map[string]any) (string, bool) {
	input, ok := receipt["input"].(map[string]any)
	if !ok {
		return "", false
	}
	// A receipt-3 wrapper keeps the effective source inside its build object.
	source, ok := driverInputOf(input)["source"].(map[string]any)
	if !ok {
		return "", false
	}
	effectiveObject, ok := source["effective"].(map[string]any)
	if !ok {
		return "", false
	}
	identityObject, ok := effectiveObject["identity"].(map[string]any)
	if !ok {
		return "", false
	}
	kind, _ := identityObject["kind"].(string)
	value, _ := identityObject["value"].(string)
	format, _ := effectiveObject["object_format"].(string)
	commit, _ := effectiveObject["commit"].(string)
	buildSource, ok := effectiveObject["build_source"].(map[string]any)
	if !ok {
		return "", false
	}
	digest, _ := buildSource["content_sha256"].(string)
	key, err := SnapshotKey(EffectiveState{IdentityKind: kind, Identity: value, ObjectFormat: format, Commit: commit}, digest)
	return key, err == nil
}
