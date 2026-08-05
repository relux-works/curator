package conformanceconsumer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// CorpusFile maps one authenticated corpus file into an isolated fixture tree.
type CorpusFile struct {
	Source string
	Target string
	Mode   fs.FileMode
}

// MaterializedFile records the source and digest of a generated fixture file.
// It is provenance metadata, not normative release evidence.
type MaterializedFile struct {
	Source string `json:"source"`
	Target string `json:"target"`
	SHA256 string `json:"sha256"`
}

// MaterializeCorpusFiles deterministically creates a fixture tree from
// manifest-authenticated corpus inputs. Existing targets and symlinks are
// rejected so a run cannot overwrite or escape its isolated workspace.
func MaterializeCorpusFiles(corpus *Corpus, root string, files []CorpusFile) ([]MaterializedFile, error) {
	if corpus == nil {
		return nil, fmt.Errorf("corpus is required")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve fixture root: %w", err)
	}
	ordered := append([]CorpusFile(nil), files...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Target < ordered[j].Target })
	seen := make(map[string]struct{}, len(ordered))
	for _, file := range ordered {
		if err := validateCorpusPath(file.Target); err != nil {
			return nil, fmt.Errorf("fixture target %q: %w", file.Target, err)
		}
		if _, duplicate := seen[file.Target]; duplicate {
			return nil, fmt.Errorf("duplicate fixture target %q", file.Target)
		}
		seen[file.Target] = struct{}{}
		if file.Mode&^fs.FileMode(0o777) != 0 || file.Mode.Perm() == 0 {
			return nil, fmt.Errorf("fixture target %q has invalid mode %04o", file.Target, file.Mode)
		}
	}
	if err := os.MkdirAll(absolute, 0o755); err != nil {
		return nil, fmt.Errorf("create fixture root: %w", err)
	}
	rootInfo, err := os.Lstat(absolute)
	if err != nil {
		return nil, fmt.Errorf("inspect fixture root: %w", err)
	}
	if rootInfo.Mode()&os.ModeSymlink != 0 || !rootInfo.IsDir() {
		return nil, fmt.Errorf("fixture root must be a real directory")
	}
	result := make([]MaterializedFile, 0, len(ordered))
	for _, file := range ordered {
		payload, digest, err := corpus.Read(file.Source)
		if err != nil {
			return nil, err
		}
		target := filepath.Join(absolute, filepath.FromSlash(file.Target))
		if err := ensureSafeParents(absolute, filepath.Dir(target)); err != nil {
			return nil, fmt.Errorf("fixture target %q: %w", file.Target, err)
		}
		handle, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, file.Mode.Perm())
		if err != nil {
			return nil, fmt.Errorf("create fixture target %q: %w", file.Target, err)
		}
		_, writeErr := handle.Write(payload)
		closeErr := handle.Close()
		if writeErr != nil {
			return nil, fmt.Errorf("write fixture target %q: %w", file.Target, writeErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close fixture target %q: %w", file.Target, closeErr)
		}
		if err := os.Chmod(target, file.Mode.Perm()); err != nil {
			return nil, fmt.Errorf("set fixture target %q mode: %w", file.Target, err)
		}
		result = append(result, MaterializedFile{Source: file.Source, Target: file.Target, SHA256: digest})
	}
	return result, nil
}

func ensureSafeParents(root, parent string) error {
	relative, err := filepath.Rel(root, parent)
	if err != nil || relative == ".." || filepath.IsAbs(relative) {
		return fmt.Errorf("target escapes fixture root")
	}
	cursor := root
	if relative == "." {
		return nil
	}
	for _, part := range splitPath(relative) {
		cursor = filepath.Join(cursor, part)
		info, statErr := os.Lstat(cursor)
		if statErr == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
				return fmt.Errorf("parent %q is not a real directory", cursor)
			}
			continue
		}
		if !os.IsNotExist(statErr) {
			return statErr
		}
		if err := os.Mkdir(cursor, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func splitPath(value string) []string {
	parts := make([]string, 0)
	for value != "." && value != "" {
		dir, base := filepath.Split(value)
		if base != "" {
			parts = append([]string{base}, parts...)
		}
		value = filepath.Clean(dir)
	}
	return parts
}
