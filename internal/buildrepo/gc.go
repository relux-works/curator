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
func Collect(root string, referenced []string, now time.Time, grace time.Duration) (removed []string, err error) {
	store := &DiskProtectedStore{Root: root}
	if _, statErr := os.Lstat(root); os.IsNotExist(statErr) {
		return nil, nil
	}
	if err := store.prepareFor(false); err != nil {
		return nil, err
	}
	keep := map[string]bool{}
	for _, key := range referenced {
		keep[strings.TrimPrefix(key, "sha256:")] = true
	}
	reachableSnapshots := map[string]bool{}
	artifactsRoot := filepath.Join(root, "artifacts")
	entries, readErr := os.ReadDir(artifactsRoot)
	if readErr != nil && !os.IsNotExist(readErr) {
		return nil, readErr
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(artifactsRoot, entry.Name())
		info, infoErr := entry.Info()
		if infoErr != nil {
			return removed, infoErr
		}
		if !keep[entry.Name()] && (grace <= 0 || !info.ModTime().Add(grace).After(now)) {
			if err := os.RemoveAll(path); err != nil {
				return removed, err
			}
			removed = append(removed, "external:sha256:"+entry.Name())
			continue
		}
		receipt, err := store.readProtectedFile(filepath.Join(path, "receipt.json"), 4<<20)
		if err != nil || protocoljson.Validate(receipt) != nil {
			return removed, fmt.Errorf("retained external receipt %s is unreadable", entry.Name())
		}
		var object map[string]any
		decoder := json.NewDecoder(strings.NewReader(string(receipt)))
		decoder.UseNumber()
		if decoder.Decode(&object) != nil {
			return removed, fmt.Errorf("retained external receipt %s is invalid", entry.Name())
		}
		if key, ok := snapshotKeyFromReceipt(object); ok {
			reachableSnapshots[strings.TrimPrefix(key, "sha256:")] = true
		}
	}
	snapshotsRoot := filepath.Join(root, "snapshots")
	snapshots, readErr := os.ReadDir(snapshotsRoot)
	if readErr != nil && !os.IsNotExist(readErr) {
		return removed, readErr
	}
	for _, entry := range snapshots {
		if entry.IsDir() && !reachableSnapshots[entry.Name()] {
			if err := os.RemoveAll(filepath.Join(snapshotsRoot, entry.Name())); err != nil {
				return removed, err
			}
			removed = append(removed, "external-snapshot:sha256:"+entry.Name())
		}
	}
	sort.Strings(removed)
	return removed, nil
}

func snapshotKeyFromReceipt(receipt map[string]any) (string, bool) {
	input, ok := receipt["input"].(map[string]any)
	if !ok {
		return "", false
	}
	source, ok := input["source"].(map[string]any)
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
