package pnpmsource

import (
	"bytes"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/closuregraph"
)

var unifiedHunkHeader = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)

type patchFileChange struct {
	oldPath, newPath string
	hunks            []patchHunk
}

type patchHunk struct {
	oldStart, oldCount, newStart, newCount int
	lines                                  []string
}

func derivePatchedInventories(graph Graph, tarballs map[string]capturedBlob, patches map[string]capturedBlob) ([]PatchTransformEvidence, error) {
	receipts := make([]PatchTransformEvidence, 0, len(graph.Patches))
	for _, patch := range graph.Patches {
		item, ok := tarballs[patch.Selector]
		if !ok {
			return nil, fail(CodeGraphIncomplete, "pnpm patch has no admitted package input", map[string]string{"selector": patch.Selector})
		}
		patchInput, ok := patches[patch.Path]
		if !ok {
			return nil, fail(CodeOfflineInputMissing, "pnpm patch transform has no admitted patch input", map[string]string{"path": patch.Path})
		}
		payload := graph.patchBytes[patch.Path]
		patched, err := applyUnifiedPatch(item.contents, payload)
		if err != nil {
			return nil, fail(CodeIntegrityMismatch, "declared pnpm patch cannot be applied to admitted package: "+err.Error(), map[string]string{"selector": patch.Selector, "path": patch.Path})
		}
		files := inventoryContentMap(patched, item.files)
		fileValues := make([]any, len(files))
		for i, file := range files {
			fileValues[i] = map[string]any{"path": file.Path, "sha256": string(file.SHA256), "size": file.Size, "executable": file.Executable}
		}
		id, err := closuregraph.DomainID("pnpm-patch-transform-receipt-v1", map[string]any{
			"schema_id":            "pnpm-patch-transform-receipt-v1",
			"selector":             patch.Selector,
			"snapshot_keys":        append([]string(nil), patch.SnapshotKeys...),
			"manager_hash":         patch.ManagerHash,
			"tarball_receipt_id":   string(item.receiptID),
			"patch_receipt_id":     string(patchInput.receiptID),
			"expected_file_values": fileValues,
		})
		if err != nil {
			return nil, err
		}
		item.files = files
		item.contents = patched
		tarballs[patch.Selector] = item
		receipts = append(receipts, PatchTransformEvidence{Selector: patch.Selector, SnapshotKeys: append([]string(nil), patch.SnapshotKeys...), ManagerHash: patch.ManagerHash, TarballReceiptID: item.receiptID, PatchReceiptID: patchInput.receiptID, ExpectedFiles: files, ReceiptID: id})
	}
	return receipts, nil
}

func applyUnifiedPatch(original map[string][]byte, payload []byte) (map[string][]byte, error) {
	changes, err := parseUnifiedPatch(payload)
	if err != nil {
		return nil, err
	}
	result := cloneContentMap(original)
	for _, change := range changes {
		if change.oldPath == "package.json" || change.newPath == "package.json" {
			return nil, fmt.Errorf("patching package.json is unsupported because it can mutate the closed graph")
		}
		var current []byte
		if change.oldPath != "" {
			value, ok := result[change.oldPath]
			if !ok {
				return nil, fmt.Errorf("patch source %q is absent", change.oldPath)
			}
			current = value
		} else if _, exists := result[change.newPath]; exists {
			return nil, fmt.Errorf("patch creates existing path %q", change.newPath)
		}
		updated, err := applyPatchHunks(current, change.hunks)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", firstNonempty(change.newPath, change.oldPath), err)
		}
		if change.newPath == "" {
			delete(result, change.oldPath)
			continue
		}
		if change.oldPath != "" && change.oldPath != change.newPath {
			delete(result, change.oldPath)
		}
		result[change.newPath] = updated
	}
	return result, nil
}

func parseUnifiedPatch(payload []byte) ([]patchFileChange, error) {
	if len(payload) == 0 || bytes.IndexByte(payload, 0) >= 0 {
		return nil, fmt.Errorf("patch is empty or contains NUL")
	}
	normalized := strings.ReplaceAll(string(payload), "\r\n", "\n")
	lines := strings.SplitAfter(normalized, "\n")
	changes := []patchFileChange{}
	for i := 0; i < len(lines); {
		line := strings.TrimSuffix(lines[i], "\n")
		if !strings.HasPrefix(line, "diff --git ") {
			if strings.TrimSpace(line) == "" {
				i++
				continue
			}
			return nil, fmt.Errorf("unexpected patch preamble line %q", line)
		}
		fields := strings.Fields(line)
		if len(fields) != 4 || !strings.HasPrefix(fields[2], "a/") || !strings.HasPrefix(fields[3], "b/") {
			return nil, fmt.Errorf("unsupported diff header %q", line)
		}
		change := patchFileChange{}
		i++
		for i < len(lines) && !strings.HasPrefix(lines[i], "--- ") {
			metadata := strings.TrimSuffix(lines[i], "\n")
			if strings.HasPrefix(metadata, "diff --git ") || strings.HasPrefix(metadata, "rename ") || strings.HasPrefix(metadata, "copy ") || strings.HasPrefix(metadata, "old mode ") || strings.HasPrefix(metadata, "new mode ") {
				return nil, fmt.Errorf("unsupported patch metadata %q", metadata)
			}
			i++
		}
		if i+1 >= len(lines) || !strings.HasPrefix(lines[i], "--- ") || !strings.HasPrefix(lines[i+1], "+++ ") {
			return nil, fmt.Errorf("diff has no closed old/new path headers")
		}
		oldPath, err := parsePatchPath(strings.TrimSuffix(strings.TrimPrefix(lines[i], "--- "), "\n"), "a/")
		if err != nil {
			return nil, err
		}
		newPath, err := parsePatchPath(strings.TrimSuffix(strings.TrimPrefix(lines[i+1], "+++ "), "\n"), "b/")
		if err != nil {
			return nil, err
		}
		if oldPath == "" && newPath == "" {
			return nil, fmt.Errorf("patch cannot delete and create /dev/null")
		}
		if oldPath != "" && newPath != "" && oldPath != newPath {
			return nil, fmt.Errorf("patch renames are unsupported")
		}
		if oldPath != "" && strings.TrimPrefix(fields[2], "a/") != oldPath {
			return nil, fmt.Errorf("diff and old path headers disagree")
		}
		if newPath != "" && strings.TrimPrefix(fields[3], "b/") != newPath {
			return nil, fmt.Errorf("diff and new path headers disagree")
		}
		change.oldPath, change.newPath = oldPath, newPath
		i += 2
		for i < len(lines) && strings.HasPrefix(lines[i], "@@ ") {
			hunk, next, err := parsePatchHunk(lines, i)
			if err != nil {
				return nil, err
			}
			change.hunks = append(change.hunks, hunk)
			i = next
		}
		if len(change.hunks) == 0 {
			return nil, fmt.Errorf("patch file change has no hunks")
		}
		changes = append(changes, change)
	}
	if len(changes) == 0 {
		return nil, fmt.Errorf("patch contains no file changes")
	}
	return changes, nil
}

func parsePatchPath(value, prefix string) (string, error) {
	value = strings.SplitN(value, "\t", 2)[0]
	if value == "/dev/null" {
		return "", nil
	}
	if !strings.HasPrefix(value, prefix) {
		return "", fmt.Errorf("patch path %q lacks %q prefix", value, prefix)
	}
	value = strings.TrimPrefix(value, prefix)
	if value == "" || path.IsAbs(value) || path.Clean(value) != value || value == ".." || strings.HasPrefix(value, "../") || strings.Contains(value, "\\") {
		return "", fmt.Errorf("patch path %q escapes package", value)
	}
	return value, nil
}

func parsePatchHunk(lines []string, start int) (patchHunk, int, error) {
	header := strings.TrimSuffix(lines[start], "\n")
	match := unifiedHunkHeader.FindStringSubmatch(header)
	if match == nil {
		return patchHunk{}, start, fmt.Errorf("invalid hunk header %q", header)
	}
	values := make([]int, 4)
	for i := range values {
		if match[i+1] == "" {
			values[i] = 1
			continue
		}
		value, err := strconv.Atoi(match[i+1])
		if err != nil {
			return patchHunk{}, start, err
		}
		values[i] = value
	}
	hunk := patchHunk{oldStart: values[0], oldCount: values[1], newStart: values[2], newCount: values[3]}
	oldSeen, newSeen := 0, 0
	i := start + 1
	for (i < len(lines) && oldSeen < hunk.oldCount) || newSeen < hunk.newCount {
		if i >= len(lines) {
			return patchHunk{}, start, fmt.Errorf("truncated hunk")
		}
		line := lines[i]
		if strings.HasPrefix(line, "\\ No newline at end of file") {
			return patchHunk{}, start, fmt.Errorf("no-newline patch markers are unsupported")
		}
		if line == "" {
			return patchHunk{}, start, fmt.Errorf("empty unprefixed hunk line")
		}
		switch line[0] {
		case ' ':
			oldSeen++
			newSeen++
		case '-':
			oldSeen++
		case '+':
			newSeen++
		default:
			return patchHunk{}, start, fmt.Errorf("invalid hunk line prefix %q", line[0])
		}
		hunk.lines = append(hunk.lines, line)
		i++
	}
	if oldSeen != hunk.oldCount || newSeen != hunk.newCount {
		return patchHunk{}, start, fmt.Errorf("hunk line counts do not match header")
	}
	return hunk, i, nil
}

func applyPatchHunks(payload []byte, hunks []patchHunk) ([]byte, error) {
	original := splitLines(payload)
	result := []string{}
	cursor := 0
	for _, hunk := range hunks {
		start := hunk.oldStart - 1
		if hunk.oldStart == 0 {
			start = 0
		}
		if start < cursor || start > len(original) {
			return nil, fmt.Errorf("hunk starts outside source")
		}
		result = append(result, original[cursor:start]...)
		cursor = start
		for _, line := range hunk.lines {
			content := line[1:]
			switch line[0] {
			case ' ':
				if cursor >= len(original) || original[cursor] != content {
					return nil, fmt.Errorf("hunk context does not match source")
				}
				result = append(result, content)
				cursor++
			case '-':
				if cursor >= len(original) || original[cursor] != content {
					return nil, fmt.Errorf("hunk deletion does not match source")
				}
				cursor++
			case '+':
				result = append(result, content)
			}
		}
	}
	result = append(result, original[cursor:]...)
	return []byte(strings.Join(result, "")), nil
}

func splitLines(payload []byte) []string {
	if len(payload) == 0 {
		return nil
	}
	return strings.SplitAfter(string(payload), "\n")
}

func inventoryContentMap(contents map[string][]byte, originals []packageFile) []packageFile {
	executable := map[string]bool{}
	for _, file := range originals {
		executable[file.Path] = file.Executable
	}
	files := make([]packageFile, 0, len(contents))
	for name, payload := range contents {
		files = append(files, packageFile{Path: name, SHA256: digestID(payload), Size: int64(len(payload)), Executable: executable[name]})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files
}

func cloneContentMap(values map[string][]byte) map[string][]byte {
	result := make(map[string][]byte, len(values))
	for key, value := range values {
		result[key] = append([]byte(nil), value...)
	}
	return result
}

func firstNonempty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
