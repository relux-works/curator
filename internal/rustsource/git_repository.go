package rustsource

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1" // #nosec G505 -- Git's pinned object identity is SHA-1 by protocol.
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type gitIndexEntry struct {
	path string
	mode uint32
	oid  [20]byte
}

type gitRepositoryState struct {
	commit, tree string
	tracked      map[string]bool
	entries      []gitIndexEntry
	submodules   []SubmoduleEvidence
}

func inspectGitRepository(root, lockedCommit string) (gitRepositoryState, error) {
	gitDir, err := resolveGitDir(root)
	if err != nil {
		return gitRepositoryState{}, fail(CodeGitIdentityInvalid, "Git administrative directory is absent", nil)
	}
	return inspectGitRepositoryContained(root, lockedCommit, gitDir)
}

func inspectGitRepositoryContained(root, lockedCommit, allowedGitRoot string) (gitRepositoryState, error) {
	gitDir, err := resolveGitDir(root)
	if err != nil || !contained(allowedGitRoot, gitDir) {
		return gitRepositoryState{}, fail(CodeGitIdentityInvalid, "Git administrative directory escapes the captured object store", nil)
	}
	head, err := resolveGitHEAD(gitDir)
	if err != nil || head != lockedCommit {
		return gitRepositoryState{}, fail(CodeGitIdentityInvalid, "Git HEAD differs from lock", nil)
	}
	objectType, commitPayload, err := readGitObject(gitDir, head)
	if err != nil || objectType != "commit" {
		return gitRepositoryState{}, fail(CodeGitIdentityInvalid, "locked Git commit object is unavailable", nil)
	}
	tree := ""
	for _, line := range strings.Split(string(commitPayload), "\n") {
		if strings.HasPrefix(line, "tree ") {
			tree = strings.TrimPrefix(line, "tree ")
			break
		}
	}
	if !validLowerHex(tree, 40) {
		return gitRepositoryState{}, fail(CodeGitIdentityInvalid, "locked Git tree identity is invalid", nil)
	}
	entries, err := readGitIndex(filepath.Join(gitDir, "index"))
	if err != nil {
		return gitRepositoryState{}, err
	}
	computedTree, err := gitIndexTree(entries)
	if err != nil || computedTree != tree {
		return gitRepositoryState{}, fail(CodeGitIdentityInvalid, "Git index differs from locked tree", nil)
	}
	tracked := make(map[string]bool, len(entries))
	submodules := []SubmoduleEvidence{}
	gitlinks := []string{}
	for _, entry := range entries {
		tracked[entry.path] = true
		if entry.mode == 0o160000 {
			commit := hex.EncodeToString(entry.oid[:])
			nested, nestedErr := inspectGitRepositoryContained(filepath.Join(root, filepath.FromSlash(entry.path)), commit, allowedGitRoot)
			if nestedErr != nil {
				return gitRepositoryState{}, nestedErr
			}
			submodules = append(submodules, SubmoduleEvidence{Path: entry.path, Gitlink: commit, Commit: nested.commit, TreeDigest: nested.tree})
			for _, child := range nested.submodules {
				child.Path = filepath.ToSlash(filepath.Join(entry.path, child.Path))
				submodules = append(submodules, child)
			}
			gitlinks = append(gitlinks, entry.path)
			continue
		}
		payload, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(entry.path))) // #nosec G304 -- index-bound contained path.
		if readErr != nil {
			return gitRepositoryState{}, fail(CodeGitIdentityInvalid, "tracked Git leaf is absent", map[string]string{"path": entry.path})
		}
		if gitBlobOID(payload) != hex.EncodeToString(entry.oid[:]) {
			return gitRepositoryState{}, fail(CodeGitIdentityInvalid, "Git worktree differs from index", map[string]string{"path": entry.path})
		}
		if filepath.Base(entry.path) == ".gitattributes" {
			text := strings.ToLower(string(payload))
			if strings.Contains(text, "filter=") || strings.Contains(text, "filter ") || strings.Contains(text, "filter\t") {
				return gitRepositoryState{}, fail(CodeGitIdentityInvalid, "Git filter or LFS declaration is unsupported", map[string]string{"path": entry.path})
			}
		}
	}
	leaves, err := managerInventory(root)
	if err != nil {
		return gitRepositoryState{}, err
	}
	for _, leaf := range leaves {
		insideSubmodule := false
		for _, prefix := range gitlinks {
			if strings.HasPrefix(leaf.Path, prefix+"/") {
				insideSubmodule = true
				break
			}
		}
		if !tracked[leaf.Path] && !insideSubmodule {
			return gitRepositoryState{}, fail(CodeGitIdentityInvalid, "Git worktree contains an untracked leaf", map[string]string{"path": leaf.Path})
		}
	}
	sort.Slice(submodules, func(i, j int) bool { return submodules[i].Path < submodules[j].Path })
	return gitRepositoryState{commit: head, tree: tree, tracked: tracked, entries: entries, submodules: submodules}, nil
}

func resolveGitDir(root string) (string, error) {
	marker := filepath.Join(root, ".git")
	info, err := os.Lstat(marker)
	if err != nil {
		return "", err
	}
	if info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
		return marker, nil
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("git directory marker is invalid")
	}
	payload, err := os.ReadFile(marker) // #nosec G304 -- fixed nested-worktree marker.
	if err != nil || !strings.HasPrefix(string(payload), "gitdir: ") {
		return "", fmt.Errorf("git directory marker is malformed")
	}
	value := strings.TrimSpace(strings.TrimPrefix(string(payload), "gitdir: "))
	if filepath.IsAbs(value) {
		return filepath.Clean(value), nil
	}
	return filepath.Clean(filepath.Join(root, value)), nil
}

func resolveGitHEAD(gitDir string) (string, error) {
	payload, err := os.ReadFile(filepath.Join(gitDir, "HEAD")) // #nosec G304 -- fixed administrative path.
	if err != nil {
		return "", err
	}
	value := strings.TrimSpace(string(payload))
	if validLowerHex(value, 40) {
		return value, nil
	}
	if !strings.HasPrefix(value, "ref: refs/") || strings.Contains(value, "..") {
		return "", fmt.Errorf("invalid Git HEAD")
	}
	ref := strings.TrimPrefix(value, "ref: ")
	refBytes, readErr := os.ReadFile(filepath.Join(gitDir, filepath.FromSlash(ref))) // #nosec G304 -- validated relative ref.
	if readErr == nil {
		result := strings.TrimSpace(string(refBytes))
		if validLowerHex(result, 40) {
			return result, nil
		}
	}
	packed, packedErr := os.ReadFile(filepath.Join(gitDir, "packed-refs")) // #nosec G304 -- fixed administrative path.
	if packedErr != nil {
		return "", readErr
	}
	for _, line := range strings.Split(string(packed), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == ref && validLowerHex(fields[0], 40) {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("git HEAD ref is unresolved")
}

func readLooseGitObject(gitDir, oid string) (string, []byte, error) {
	if !validLowerHex(oid, 40) {
		return "", nil, fmt.Errorf("invalid Git object id")
	}
	path := filepath.Join(gitDir, "objects", oid[:2], oid[2:])
	file, err := os.Open(path) // #nosec G304 -- object path derives from validated hex.
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = file.Close() }()
	reader, err := zlib.NewReader(file)
	if err != nil {
		return "", nil, err
	}
	raw, err := io.ReadAll(io.LimitReader(reader, 256<<20))
	closeErr := reader.Close()
	if err != nil || closeErr != nil {
		return "", nil, fmt.Errorf("read Git object: %v %v", err, closeErr)
	}
	separator := bytes.IndexByte(raw, 0)
	if separator <= 0 {
		return "", nil, fmt.Errorf("malformed Git object")
	}
	header := strings.Fields(string(raw[:separator]))
	if len(header) != 2 {
		return "", nil, fmt.Errorf("malformed Git object header")
	}
	size, err := strconv.ParseInt(header[1], 10, 64)
	if err != nil || size != int64(len(raw)-separator-1) {
		return "", nil, fmt.Errorf("git object size differs")
	}
	sum := sha1.Sum(raw) // #nosec G401 -- exact Git protocol identity verification.
	if hex.EncodeToString(sum[:]) != oid {
		return "", nil, fmt.Errorf("git object digest differs")
	}
	return header[0], raw[separator+1:], nil
}

func readGitObject(gitDir, oid string) (string, []byte, error) {
	kind, payload, err := readLooseGitObject(gitDir, oid)
	if err == nil {
		return kind, payload, nil
	}
	return readPackedGitObject(gitDir, oid)
}

type gitPackIndex struct {
	packPath string
	offsets  map[string]uint64
}

func readPackedGitObject(gitDir, oid string) (string, []byte, error) {
	indexes, err := filepath.Glob(filepath.Join(gitDir, "objects", "pack", "*.idx"))
	if err != nil {
		return "", nil, err
	}
	for _, indexPath := range indexes {
		index, indexErr := loadGitPackIndex(indexPath)
		if indexErr != nil {
			return "", nil, indexErr
		}
		if offset, ok := index.offsets[oid]; ok {
			kind, payload, readErr := index.readObjectAt(offset, map[uint64]bool{})
			if readErr != nil {
				return "", nil, readErr
			}
			if gitObjectOID(kind, payload) != oid {
				return "", nil, fmt.Errorf("packed Git object digest differs")
			}
			return kind, payload, nil
		}
	}
	return "", nil, fmt.Errorf("packed Git object is unavailable")
}

func gitObjectOID(kind string, payload []byte) string {
	hash := sha1.New() // #nosec G401 -- Git object identity protocol.
	_, _ = fmt.Fprintf(hash, "%s %d\x00", kind, len(payload))
	_, _ = hash.Write(payload)
	return hex.EncodeToString(hash.Sum(nil))
}

func loadGitPackIndex(path string) (gitPackIndex, error) {
	payload, err := os.ReadFile(path) // #nosec G304 -- contained Git pack index.
	if err != nil || len(payload) < 8+256*4+40 || binary.BigEndian.Uint32(payload[:4]) != 0xff744f63 || binary.BigEndian.Uint32(payload[4:8]) != 2 {
		return gitPackIndex{}, fmt.Errorf("unsupported Git pack index")
	}
	indexSum := sha1.Sum(payload[:len(payload)-20]) // #nosec G401 -- Git pack-index checksum protocol.
	if !bytes.Equal(indexSum[:], payload[len(payload)-20:]) {
		return gitPackIndex{}, fmt.Errorf("git pack index checksum differs")
	}
	fanout := payload[8 : 8+256*4]
	count := int(binary.BigEndian.Uint32(fanout[255*4 : 256*4]))
	namesStart := 8 + 256*4
	crcStart := namesStart + count*20
	offsetStart := crcStart + count*4
	if count < 0 || offsetStart+count*4 > len(payload)-40 {
		return gitPackIndex{}, fmt.Errorf("truncated Git pack index")
	}
	largeStart := offsetStart + count*4
	offsets := make(map[string]uint64, count)
	for index := 0; index < count; index++ {
		name := hex.EncodeToString(payload[namesStart+index*20 : namesStart+(index+1)*20])
		raw := binary.BigEndian.Uint32(payload[offsetStart+index*4 : offsetStart+(index+1)*4])
		var offset uint64
		if raw&0x80000000 == 0 {
			offset = uint64(raw)
		} else {
			largeIndex := int(raw & 0x7fffffff)
			position := largeStart + largeIndex*8
			if position+8 > len(payload)-40 {
				return gitPackIndex{}, fmt.Errorf("invalid large Git pack offset")
			}
			offset = binary.BigEndian.Uint64(payload[position : position+8])
		}
		offsets[name] = offset
	}
	return gitPackIndex{packPath: strings.TrimSuffix(path, ".idx") + ".pack", offsets: offsets}, nil
}

func (index gitPackIndex) readObjectAt(offset uint64, visiting map[uint64]bool) (string, []byte, error) {
	if visiting[offset] {
		return "", nil, fmt.Errorf("cyclic Git pack delta")
	}
	visiting[offset] = true
	defer delete(visiting, offset)
	file, err := os.Open(index.packPath) // #nosec G304 -- paired contained pack path.
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = file.Close() }()
	if offset > math.MaxInt64 {
		return "", nil, fmt.Errorf("git pack offset overflows")
	}
	if _, err = file.Seek(int64(offset), io.SeekStart); err != nil {
		return "", nil, err
	}
	first := []byte{0}
	if _, err = io.ReadFull(file, first); err != nil {
		return "", nil, err
	}
	objectType := (first[0] >> 4) & 7
	size := uint64(first[0] & 0x0f)
	shift := uint(4)
	current := first[0]
	for current&0x80 != 0 {
		if _, err = io.ReadFull(file, first); err != nil {
			return "", nil, err
		}
		current = first[0]
		size |= uint64(current&0x7f) << shift
		shift += 7
		if shift > 63 {
			return "", nil, fmt.Errorf("git pack size overflows")
		}
	}
	baseOffset := uint64(0)
	baseOID := ""
	switch objectType {
	case 6:
		if _, err = io.ReadFull(file, first); err != nil {
			return "", nil, err
		}
		distance := uint64(first[0] & 0x7f)
		for first[0]&0x80 != 0 {
			if _, err = io.ReadFull(file, first); err != nil {
				return "", nil, err
			}
			distance = ((distance + 1) << 7) | uint64(first[0]&0x7f)
		}
		if distance > offset {
			return "", nil, fmt.Errorf("invalid Git OFS delta")
		}
		baseOffset = offset - distance
	case 7:
		base := make([]byte, 20)
		if _, err = io.ReadFull(file, base); err != nil {
			return "", nil, err
		}
		baseOID = hex.EncodeToString(base)
	}
	reader, err := zlib.NewReader(file)
	if err != nil {
		return "", nil, err
	}
	payload, err := io.ReadAll(io.LimitReader(reader, int64(size)+(64<<20)))
	closeErr := reader.Close()
	if err != nil || closeErr != nil {
		return "", nil, fmt.Errorf("read packed Git object: %v %v", err, closeErr)
	}
	if objectType >= 1 && objectType <= 4 {
		if uint64(len(payload)) != size {
			return "", nil, fmt.Errorf("packed Git object size differs")
		}
		kinds := map[byte]string{1: "commit", 2: "tree", 3: "blob", 4: "tag"}
		return kinds[objectType], payload, nil
	}
	var baseKind string
	var basePayload []byte
	switch objectType {
	case 6:
		baseKind, basePayload, err = index.readObjectAt(baseOffset, visiting)
	case 7:
		baseKind, basePayload, err = index.readObjectByOID(baseOID, visiting)
	default:
		return "", nil, fmt.Errorf("unsupported Git pack object type")
	}
	if err != nil {
		return "", nil, err
	}
	result, err := applyGitDelta(basePayload, payload)
	if err != nil || uint64(len(result)) != size {
		return "", nil, fmt.Errorf("apply Git pack delta: %w", err)
	}
	return baseKind, result, nil
}

func (index gitPackIndex) readObjectByOID(oid string, visiting map[uint64]bool) (string, []byte, error) {
	offset, ok := index.offsets[oid]
	if !ok {
		return "", nil, fmt.Errorf("git delta base object is absent")
	}
	return index.readObjectAt(offset, visiting)
}

func applyGitDelta(base, delta []byte) ([]byte, error) {
	readSize := func(offset *int) (uint64, error) {
		var value uint64
		var shift uint
		for {
			if *offset >= len(delta) || shift > 63 {
				return 0, fmt.Errorf("truncated Git delta size")
			}
			current := delta[*offset]
			(*offset)++
			value |= uint64(current&0x7f) << shift
			if current&0x80 == 0 {
				return value, nil
			}
			shift += 7
		}
	}
	offset := 0
	baseSize, err := readSize(&offset)
	if err != nil || baseSize != uint64(len(base)) {
		return nil, fmt.Errorf("git delta base size differs")
	}
	resultSize, err := readSize(&offset)
	if err != nil || resultSize > 256<<20 {
		return nil, fmt.Errorf("git delta result size is invalid")
	}
	result := make([]byte, 0, resultSize)
	for offset < len(delta) {
		opcode := delta[offset]
		offset++
		if opcode&0x80 == 0 {
			count := int(opcode)
			if count == 0 || offset+count > len(delta) {
				return nil, fmt.Errorf("invalid Git delta insert")
			}
			result = append(result, delta[offset:offset+count]...)
			offset += count
			continue
		}
		copyOffset, copySize := uint32(0), uint32(0)
		for bit, shift := byte(1), uint(0); bit <= 8; bit, shift = bit<<1, shift+8 {
			if opcode&bit != 0 {
				if offset >= len(delta) {
					return nil, fmt.Errorf("truncated Git delta copy offset")
				}
				copyOffset |= uint32(delta[offset]) << shift
				offset++
			}
		}
		for bit, shift := byte(0x10), uint(0); bit <= 0x40; bit, shift = bit<<1, shift+8 {
			if opcode&bit != 0 {
				if offset >= len(delta) {
					return nil, fmt.Errorf("truncated Git delta copy size")
				}
				copySize |= uint32(delta[offset]) << shift
				offset++
			}
		}
		if copySize == 0 {
			copySize = 0x10000
		}
		end := uint64(copyOffset) + uint64(copySize)
		if end > uint64(len(base)) {
			return nil, fmt.Errorf("git delta copy escapes base")
		}
		result = append(result, base[copyOffset:end]...)
	}
	if uint64(len(result)) != resultSize {
		return nil, fmt.Errorf("git delta result size differs")
	}
	return result, nil
}

func readGitIndex(path string) ([]gitIndexEntry, error) {
	payload, err := os.ReadFile(path) // #nosec G304 -- fixed admitted repository metadata.
	if err != nil || len(payload) < 12+20 || string(payload[:4]) != "DIRC" {
		return nil, fail(CodeGitIdentityInvalid, "Git index is absent or malformed", nil)
	}
	version := binary.BigEndian.Uint32(payload[4:8])
	if version != 2 {
		return nil, fail(CodeGitIdentityInvalid, "Git index version is unsupported", map[string]string{"version": fmt.Sprint(version)})
	}
	count := int(binary.BigEndian.Uint32(payload[8:12]))
	offset := 12
	entries := make([]gitIndexEntry, 0, count)
	for index := 0; index < count; index++ {
		start := offset
		if offset+62 > len(payload)-20 {
			return nil, fail(CodeGitIdentityInvalid, "Git index entry is truncated", nil)
		}
		mode := binary.BigEndian.Uint32(payload[offset+24 : offset+28])
		flags := binary.BigEndian.Uint16(payload[offset+60 : offset+62])
		if flags&0x3000 != 0 || flags&0x4000 != 0 {
			return nil, fail(CodeGitIdentityInvalid, "Git index contains conflict or extended flags", nil)
		}
		offset += 62
		end := bytes.IndexByte(payload[offset:len(payload)-20], 0)
		if end < 0 {
			return nil, fail(CodeGitIdentityInvalid, "Git index path is unterminated", nil)
		}
		name := string(payload[offset : offset+end])
		if !safeRelative(filepath.ToSlash(name)) {
			return nil, fail(CodeGitIdentityInvalid, "Git index path is unsafe", map[string]string{"path": name})
		}
		var oid [20]byte
		copy(oid[:], payload[start+40:start+60])
		entries = append(entries, gitIndexEntry{path: filepath.ToSlash(name), mode: mode, oid: oid})
		offset = start + ((62+end+1+7)/8)*8
	}
	indexSum := sha1.Sum(payload[:len(payload)-20]) // #nosec G401 -- Git index checksum protocol.
	if !bytes.Equal(indexSum[:], payload[len(payload)-20:]) {
		return nil, fail(CodeGitIdentityInvalid, "Git index checksum differs", nil)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].path < entries[j].path })
	return entries, nil
}

func gitBlobOID(payload []byte) string {
	hash := sha1.New() // #nosec G401 -- Git object identity protocol.
	_, _ = fmt.Fprintf(hash, "blob %d\x00", len(payload))
	_, _ = hash.Write(payload)
	return hex.EncodeToString(hash.Sum(nil))
}

type gitTreeNode struct {
	files    []gitIndexEntry
	children map[string]*gitTreeNode
}

func gitIndexTree(entries []gitIndexEntry) (string, error) {
	root := &gitTreeNode{children: map[string]*gitTreeNode{}}
	for _, entry := range entries {
		parts := strings.Split(entry.path, "/")
		node := root
		for _, part := range parts[:len(parts)-1] {
			if node.children[part] == nil {
				node.children[part] = &gitTreeNode{children: map[string]*gitTreeNode{}}
			}
			node = node.children[part]
		}
		entry.path = parts[len(parts)-1]
		node.files = append(node.files, entry)
	}
	return hashGitTree(root)
}

func hashGitTree(node *gitTreeNode) (string, error) {
	type item struct {
		name, mode string
		oid        []byte
		directory  bool
	}
	items := []item{}
	for _, file := range node.files {
		mode := fmt.Sprintf("%o", file.mode)
		items = append(items, item{name: file.path, mode: mode, oid: append([]byte(nil), file.oid[:]...)})
	}
	for name, child := range node.children {
		oid, err := hashGitTree(child)
		if err != nil {
			return "", err
		}
		decoded, _ := hex.DecodeString(oid)
		items = append(items, item{name: name, mode: "40000", oid: decoded, directory: true})
	}
	sort.Slice(items, func(i, j int) bool {
		left, right := items[i].name, items[j].name
		if items[i].directory {
			left += "/"
		}
		if items[j].directory {
			right += "/"
		}
		return left < right
	})
	var payload bytes.Buffer
	for _, entry := range items {
		_, _ = payload.WriteString(entry.mode + " " + entry.name)
		_ = payload.WriteByte(0)
		_, _ = payload.Write(entry.oid)
	}
	hash := sha1.New() // #nosec G401 -- Git tree identity protocol.
	_, _ = fmt.Fprintf(hash, "tree %d\x00", payload.Len())
	_, _ = hash.Write(payload.Bytes())
	return hex.EncodeToString(hash.Sum(nil)), nil
}
