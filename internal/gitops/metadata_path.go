package gitops

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf16"
)

// MetadataPathError refuses a new Windows clone before creating cache parents.
// Actual and Limit count UTF-16 units, excluding the terminating NUL, in the
// longest reserved metadata path; RootLimit is the corresponding clone budget.
type MetadataPathError struct {
	Actual, Limit, RootLength, RootLimit int
}

func (e *MetadataPathError) Error() string {
	return fmt.Sprintf("git_metadata_path_too_long: projected Windows Git metadata length %d UTF-16 units exceeds bound %d; expanded clone root length %d exceeds bound %d", e.Actual, e.Limit, e.RootLength, e.RootLimit)
}

// CheckCloneDestination is a read-only Windows metadata budget check. Callers
// that prepare cache parents must call it before creating those parents.
// Other platforms retain their existing path behavior.
func CheckCloneDestination(destination string) error {
	return checkCloneDestination(destination)
}

// Reserve both supported object formats before contacting the remote. The
// longest fixed path is an incremental SHA-256 multi-pack-index bitmap. Loose
// objects, ordinary pack sidecars and commit graphs participate independently
// so changes to one layout cannot silently remove another from the budget.
func metadataPaths() []string {
	hash := strings.Repeat("0", 64)
	return []string{
		".git/objects/00/" + hash[2:],
		".git/objects/pack/pack-" + hash + ".keep",
		".git/objects/pack/pack-" + hash + ".promisor",
		".git/objects/pack/pack-" + hash + ".bitmap",
		".git/objects/info/commit-graphs/graph-" + hash + ".graph",
		".git/objects/pack/multi-pack-index.d/multi-pack-index-" + hash + ".bitmap",
		".git/objects/pack/multi-pack-index.d/multi-pack-index-chain.lock",
	}
}

func metadataPathLength(root string) (actual, overhead int) {
	rootLength := utf16PathLength(root)
	for _, path := range metadataPaths() {
		length := utf16PathLength(filepath.Join(root, filepath.FromSlash(path)))
		if length > actual {
			actual = length
		}
	}
	return actual, actual - rootLength
}

func utf16PathLength(path string) int {
	return len(utf16.Encode([]rune(path)))
}

func checkMetadataPathLength(root string, limit int) error {
	actual, overhead := metadataPathLength(root)
	if actual > limit {
		return &MetadataPathError{Actual: actual, Limit: limit, RootLength: utf16PathLength(root), RootLimit: limit - overhead}
	}
	return nil
}
