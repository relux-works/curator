package transaction

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func cloneReadInputs(inputs map[string]string) map[string]string {
	if inputs == nil {
		return nil
	}
	cloned := make(map[string]string, len(inputs))
	for path, digest := range inputs {
		cloned[path] = digest
	}
	return cloned
}

// Input bytes are checked even when the in-memory guard was lost at restart.
// Rollback intentionally does not consult them: it restores the old generation.
func checkReadInputs(inputs map[string]string) error {
	paths := make([]string, 0, len(inputs))
	for path := range inputs {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		digest, err := DigestPath(path)
		if err != nil || digest != inputs[path] {
			return fmt.Errorf("transaction_input_changed: frozen input no longer matches: %s", path)
		}
	}
	return nil
}

func validateReadInputs(journal *Journal) error {
	for path, digest := range journal.ReadInputs {
		if !filepath.IsAbs(path) || filepath.Clean(path) != path || !validText(path) || !validDigest(digest) {
			return journalf("invalid frozen read input")
		}
		for _, target := range journal.Targets {
			// Read inputs must be independent of changing live and sidecar
			// namespaces. An unchanged lock is guarded as a target instead.
			if overlapsReadPath(path, target.LivePath) {
				return journalf("read input is a changing transaction target")
			}
			for _, sidecar := range []string{target.BackupPath, target.StagedPath, target.RollbackPath} {
				if overlapsReadPath(path, sidecar) {
					return journalf("read input aliases a transaction sidecar")
				}
			}
		}
	}
	return nil
}

func overlapsReadPath(a, b string) bool {
	return a == b || strings.HasPrefix(a, b+string(filepath.Separator)) || strings.HasPrefix(b, a+string(filepath.Separator))
}
