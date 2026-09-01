package buildrepo

import (
	"bytes"
	"context"
	"crypto/sha1" // #nosec G505 -- validates Git SHA-1 container checksums.
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// LocalRequest selects a local Git worktree for inert admission.
type LocalRequest struct {
	Path            string
	Tool            GitTool
	Limits          Limits
	afterObjectCopy func()
}

type localFileProof struct {
	path    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	digest  [32]byte
}

// AdmitLocal parses a narrow ordinary worktree as inert bytes, copies only its
// admitted object inventory, re-proves every source file, and applies the same
// raw-object boundary as network acquisition. It never invokes Git in the
// selected source repository.
func AdmitLocal(ctx context.Context, request LocalRequest) (snapshot *Snapshot, err error) {
	if err := ValidateGitTool(ctx, request.Tool); err != nil {
		return nil, err
	}
	rootInfo, err := os.Lstat(request.Path)
	if err != nil || !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return nil, admissionError(CodeLocalLayoutUnsafe, "local selection is not a link-free directory")
	}
	gitDir := filepath.Join(request.Path, ".git")
	gitInfo, err := os.Lstat(gitDir)
	if err != nil {
		if looksBare(request.Path) {
			return nil, admissionError(CodeLocalBareUnsupported, "bare local repository is unsupported")
		}
		return nil, admissionError(CodeLocalLayoutUnsafe, "local .git directory is missing")
	}
	if gitInfo.Mode().IsRegular() {
		return nil, admissionError(CodeLocalGitfileUnsupported, "local .git gitfile is unsupported")
	}
	if !gitInfo.IsDir() || gitInfo.Mode()&os.ModeSymlink != 0 {
		return nil, admissionError(CodeLocalLayoutUnsafe, "local .git is not an ordinary directory")
	}
	for _, name := range []string{"commondir", "worktrees", "config.worktree"} {
		if _, err := os.Lstat(filepath.Join(gitDir, name)); err == nil {
			return nil, admissionError(CodeLocalLinkedUnsupported, "linked worktree state is unsupported")
		}
	}
	configBytes, proof, err := readProvedFile(filepath.Join(gitDir, "config"), 1<<20)
	if err != nil {
		return nil, admissionError(CodeLocalLayoutUnsafe, "local config is unsafe")
	}
	format, err := parseLocalConfig(configBytes)
	if err != nil {
		return nil, err
	}
	selected, refProofs, err := readLocalHEAD(gitDir, format)
	if err != nil {
		return nil, err
	}
	proofs := append([]localFileProof{proof}, refProofs...)

	limits := normalizedLimits(request.Limits)
	ctx, cancel := context.WithTimeout(ctx, limits.Timeout)
	defer cancel()
	privateRoot, err := os.MkdirTemp("", "curator-buildrepo-local-")
	if err != nil {
		return nil, admissionError(CodeLocalLayoutUnsafe, "cannot create private state")
	}
	defer func() {
		// The object store is sealed read-only below, so the release has to
		// restore write permission before it can remove the tree. A private
		// root that survives the admission is a real defect, not a detail to
		// discard: a successful admission is refused rather than leaking it.
		if releaseErr := releasePrivateRoot(privateRoot); releaseErr != nil && err == nil {
			snapshot, err = nil, admissionError(CodeLocalLayoutUnsafe, "private state could not be released")
		}
	}()
	paths, err := makePrivatePaths(privateRoot)
	if err != nil {
		return nil, admissionError(CodeLocalLayoutUnsafe, "cannot initialize private state")
	}
	env := cleanGitEnvironment(paths, request.Tool, "")
	initArgs := []string{"--git-dir=" + paths.repo, "-c", "init.defaultBranch=curator-invalid", "init", "--bare", "--quiet", "--template=" + paths.template, "--object-format=" + format, "--ref-format=files"}
	if err := runGit(ctx, request.Tool.Executable, paths.work, env, initArgs...); err != nil {
		return nil, admissionError(CodeLocalObjectFormatUnsupported, "private Git initialization failed")
	}
	objectProofs, err := copyLocalObjects(filepath.Join(gitDir, "objects"), filepath.Join(paths.repo, "objects"), format, limits)
	if err != nil {
		return nil, err
	}
	proofs = append(proofs, objectProofs...)
	if request.afterObjectCopy != nil {
		request.afterObjectCopy()
	}
	if err := validateLocalAdministration(gitDir, format); err != nil {
		return nil, err
	}
	for _, item := range proofs {
		if err := recheckProof(item); err != nil {
			return nil, admissionError(CodeLocalLayoutUnsafe, "local repository changed during admission")
		}
	}
	if err := sealObjectStore(filepath.Join(paths.repo, "objects")); err != nil {
		return nil, admissionError(CodeLocalLayoutUnsafe, "private object store could not be sealed")
	}
	snapshot, err = proveRepository(ctx, request.Tool, env, paths, format, selected, "", limits)
	if err != nil {
		return nil, err
	}
	for _, item := range proofs {
		if err := recheckProof(item); err != nil {
			return nil, admissionError(CodeLocalLayoutUnsafe, "local repository changed during raw-object proof")
		}
	}
	return snapshot, nil
}

func looksBare(root string) bool {
	for _, name := range []string{"HEAD", "config", "objects", "refs"} {
		if _, err := os.Lstat(filepath.Join(root, name)); err != nil {
			return false
		}
	}
	return true
}

var configToken = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]*$`)

type configEntry struct{ section, subsection, key, value string }

func parseLocalConfig(data []byte) (string, error) {
	if len(data) > 1<<20 || !utf8.Valid(data) || bytes.HasPrefix(data, []byte{0xef, 0xbb, 0xbf}) || bytes.IndexByte(data, 0) >= 0 || bytes.Contains(data, []byte("\r\n")) && bytes.Contains(bytes.ReplaceAll(data, []byte("\r\n"), nil), []byte{'\r'}) || (!bytes.Contains(data, []byte("\r\n")) && bytes.IndexByte(data, '\r') >= 0) {
		return "", admissionError(CodeLocalFormatUnsupported, "local Git config encoding is unsupported")
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(text, "\n")
	section, subsection := "", ""
	entries := make([]configEntry, 0)
	for lineNo, raw := range lines {
		if lineNo == len(lines)-1 && raw == "" {
			continue
		}
		trimmed := strings.Trim(raw, " \t")
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			continue
		}
		if strings.HasSuffix(raw, "\\") {
			return "", admissionError(CodeLocalFormatUnsupported, "config continuation is unsupported")
		}
		if strings.HasPrefix(trimmed, "[") {
			if !strings.HasSuffix(trimmed, "]") {
				return "", admissionError(CodeLocalFormatUnsupported, "malformed config section")
			}
			inside := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
			name, sub, ok := parseConfigSection(inside)
			if !ok {
				return "", admissionError(CodeLocalFormatUnsupported, "malformed config section")
			}
			section, subsection = strings.ToLower(name), strings.ToLower(sub)
			if section == "include" || section == "includeif" {
				return "", admissionError(CodeLocalFormatUnsupported, "config includes are unsupported")
			}
			continue
		}
		if section == "" {
			return "", admissionError(CodeLocalFormatUnsupported, "config assignment precedes section")
		}
		key, value, ok := parseConfigAssignment(raw)
		if !ok {
			return "", admissionError(CodeLocalFormatUnsupported, "malformed config assignment")
		}
		entries = append(entries, configEntry{section: section, subsection: subsection, key: strings.ToLower(key), value: value})
	}
	security := map[string]string{}
	for _, entry := range entries {
		identity := entry.section + "." + entry.subsection + "." + entry.key
		isSecurity := entry.section == "core" && (entry.key == "repositoryformatversion" || entry.key == "bare") || entry.section == "extensions" || entry.section == "remote" && (entry.key == "promisor" || entry.key == "partialclonefilter")
		if !isSecurity {
			continue
		}
		if _, exists := security[identity]; exists {
			return "", admissionError(CodeLocalFormatUnsupported, "duplicate security-relevant config key")
		}
		security[identity] = entry.value
	}
	version, versionOK := security["core..repositoryformatversion"]
	bare, bareOK := security["core..bare"]
	if !versionOK || !bareOK {
		return "", admissionError(CodeLocalFormatUnsupported, "required core config keys are missing")
	}
	bareValue, ok := parseGitBool(bare)
	if !ok || bareValue {
		if ok && bareValue {
			return "", admissionError(CodeLocalBareUnsupported, "local repository is bare")
		}
		return "", admissionError(CodeLocalFormatUnsupported, "core.bare is invalid")
	}
	for key, value := range security {
		if strings.HasPrefix(key, "remote.") {
			if strings.HasSuffix(key, ".promisor") {
				promisor, valid := parseGitBool(value)
				if !valid || promisor {
					return "", admissionError(CodeLocalFormatUnsupported, "promisor state is unsupported")
				}
			} else if value != "" {
				return "", admissionError(CodeLocalFormatUnsupported, "partial clone state is unsupported")
			}
		}
	}
	extensions := map[string]string{}
	for key, value := range security {
		if strings.HasPrefix(key, "extensions..") {
			extensions[strings.TrimPrefix(key, "extensions..")] = strings.ToLower(value)
		}
	}
	switch version {
	case "0":
		if len(extensions) != 0 {
			return "", admissionError(CodeLocalFormatUnsupported, "SHA-1 repository has extensions")
		}
		return "sha1", nil
	case "1":
		if extensions["objectformat"] != "sha256" || len(extensions) > 2 || (extensions["refstorage"] != "" && extensions["refstorage"] != "files") {
			return "", admissionError(CodeLocalFormatUnsupported, "repository extensions are unsupported")
		}
		for key := range extensions {
			if key != "objectformat" && key != "refstorage" {
				return "", admissionError(CodeLocalFormatUnsupported, "unknown repository extension")
			}
		}
		return "sha256", nil
	default:
		return "", admissionError(CodeLocalFormatUnsupported, "repository format version is unsupported")
	}
}

func parseConfigSection(value string) (name, subsection string, ok bool) {
	space := strings.IndexAny(value, " \t")
	if space < 0 {
		return value, "", configToken.MatchString(value)
	}
	name = value[:space]
	rest := strings.TrimSpace(value[space:])
	if !configToken.MatchString(name) || len(rest) < 2 || rest[0] != '"' || rest[len(rest)-1] != '"' {
		return "", "", false
	}
	decoded, valid := decodeConfigQuoted(rest[1:len(rest)-1], false)
	return name, decoded, valid
}

func parseConfigAssignment(line string) (key, value string, ok bool) {
	trimmed := strings.Trim(line, " \t")
	equal := strings.IndexByte(trimmed, '=')
	if equal < 0 {
		key = strings.TrimSpace(stripConfigComment(trimmed))
		return key, "true", configToken.MatchString(key)
	}
	key = strings.TrimSpace(trimmed[:equal])
	if !configToken.MatchString(key) {
		return "", "", false
	}
	raw := strings.TrimSpace(trimmed[equal+1:])
	if strings.HasPrefix(raw, "\"") {
		end := quotedEnd(raw)
		if end < 0 || strings.TrimSpace(stripConfigComment(raw[end+1:])) != "" {
			return "", "", false
		}
		decoded, valid := decodeConfigQuoted(raw[1:end], true)
		return key, decoded, valid
	}
	raw = strings.TrimSpace(stripConfigComment(raw))
	if strings.ContainsAny(raw, "\"\\") {
		return "", "", false
	}
	for _, r := range raw {
		if unicodeControlExceptTab(r) {
			return "", "", false
		}
	}
	return key, raw, true
}

func stripConfigComment(value string) string {
	for index, r := range value {
		if r == '#' || r == ';' {
			return value[:index]
		}
	}
	return value
}

func quotedEnd(value string) int {
	escaped := false
	for i := 1; i < len(value); i++ {
		if escaped {
			escaped = false
			continue
		}
		//nolint:staticcheck // The two-branch escape grammar is clearer than a tagged switch here.
		if value[i] == '\\' {
			escaped = true
		} else if value[i] == '"' {
			return i
		}
	}
	return -1
}

func decodeConfigQuoted(value string, assignment bool) (string, bool) {
	var out strings.Builder
	for i := 0; i < len(value); i++ {
		if value[i] != '\\' {
			out.WriteByte(value[i])
			continue
		}
		i++
		if i >= len(value) {
			return "", false
		}
		switch value[i] {
		case '\\', '"':
			out.WriteByte(value[i])
		case 'n':
			if !assignment {
				return "", false
			}
			out.WriteByte('\n')
		case 't':
			if !assignment {
				return "", false
			}
			out.WriteByte('\t')
		case 'b':
			if !assignment {
				return "", false
			}
			out.WriteByte('\b')
		default:
			return "", false
		}
	}
	return out.String(), utf8.ValidString(out.String())
}

func unicodeControlExceptTab(r rune) bool { return r < 0x20 && r != '\t' || r == 0x7f }

func parseGitBool(value string) (bool, bool) {
	switch strings.ToLower(value) {
	case "true", "yes", "on", "1":
		return true, true
	case "false", "no", "off", "0":
		return false, true
	default:
		return false, false
	}
}

func readLocalHEAD(gitDir, format string) (string, []localFileProof, error) {
	payload, headProof, err := readProvedFile(filepath.Join(gitDir, "HEAD"), 512)
	if err != nil {
		return "", nil, admissionError(CodeLocalLayoutUnsafe, "HEAD is unsafe")
	}
	value, err := exactOneLine(payload)
	if err != nil {
		return "", nil, admissionError(CodeLocalFormatUnsupported, "HEAD is malformed")
	}
	proofs := []localFileProof{headProof}
	if strings.HasPrefix(value, "ref: ") {
		ref := strings.TrimPrefix(value, "ref: ")
		if !strings.HasPrefix(ref, "refs/heads/") || !ValidRefName(strings.TrimPrefix(ref, "refs/heads/")) {
			return "", nil, admissionError(CodeLocalFormatUnsupported, "HEAD ref is invalid")
		}
		refBytes, refProof, readErr := readProvedFile(filepath.Join(gitDir, filepath.FromSlash(ref)), 512)
		if readErr == nil {
			value, err = exactOneLine(refBytes)
			if err != nil {
				return "", nil, admissionError(CodeLocalFormatUnsupported, "selected head ref is malformed")
			}
			proofs = append(proofs, refProof)
		} else if os.IsNotExist(unwrapPathError(readErr)) {
			packed, packedProof, packedErr := readPackedRefs(gitDir, format)
			if packedErr != nil {
				return "", nil, packedErr
			}
			var found bool
			value, found = packed[ref]
			if !found {
				return "", nil, admissionError(CodeLocalFormatUnsupported, "selected head ref is missing")
			}
			proofs = append(proofs, packedProof)
		} else {
			return "", nil, admissionError(CodeLocalLayoutUnsafe, "selected head ref is unsafe")
		}
	}
	if !validOID(value, format) {
		return "", nil, admissionError(CodeLocalFormatUnsupported, "HEAD object ID is invalid")
	}
	return value, proofs, nil
}

func unwrapPathError(err error) error {
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return pathErr.Err
	}
	return err
}

func readPackedRefs(gitDir, format string) (map[string]string, localFileProof, error) {
	path := filepath.Join(gitDir, "packed-refs")
	data, proof, err := readProvedFile(path, 8<<20)
	if err != nil {
		return nil, localFileProof{}, admissionError(CodeLocalLayoutUnsafe, "packed-refs is unsafe")
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !utf8.ValidString(text) || strings.ContainsRune(text, '\r') {
		return nil, localFileProof{}, admissionError(CodeLocalFormatUnsupported, "packed-refs encoding is invalid")
	}
	refs := map[string]string{}
	previousTag := false
	headerSeen := false
	for index, line := range strings.Split(strings.TrimSuffix(text, "\n"), "\n") {
		if line == "" {
			return nil, localFileProof{}, admissionError(CodeLocalFormatUnsupported, "empty packed-refs record")
		}
		if strings.HasPrefix(line, "#") {
			if index != 0 || headerSeen || !strings.HasPrefix(line, "# pack-refs with:") {
				return nil, localFileProof{}, admissionError(CodeLocalFormatUnsupported, "unsupported packed-refs header")
			}
			headerSeen = true
			for _, trait := range strings.Fields(strings.TrimSpace(strings.TrimPrefix(line, "# pack-refs with:"))) {
				if trait != "peeled" && trait != "fully-peeled" && trait != "sorted" {
					return nil, localFileProof{}, admissionError(CodeLocalFormatUnsupported, "unsupported packed-refs trait")
				}
			}
			continue
		}
		if strings.HasPrefix(line, "^") {
			if !previousTag || !validOID(strings.TrimPrefix(line, "^"), format) {
				return nil, localFileProof{}, admissionError(CodeLocalFormatUnsupported, "misplaced packed-ref peel")
			}
			previousTag = false
			continue
		}
		space := strings.IndexByte(line, ' ')
		if space <= 0 {
			return nil, localFileProof{}, admissionError(CodeLocalFormatUnsupported, "malformed packed ref")
		}
		oid, ref := line[:space], line[space+1:]
		if !validOID(oid, format) || !strings.HasPrefix(ref, "refs/") || !ValidRefName(strings.TrimPrefix(ref, "refs/")) {
			return nil, localFileProof{}, admissionError(CodeLocalFormatUnsupported, "invalid packed ref")
		}
		if strings.HasPrefix(ref, "refs/replace/") {
			return nil, localFileProof{}, admissionError(CodeLocalFormatUnsupported, "replace refs are unsupported")
		}
		if _, exists := refs[ref]; exists {
			return nil, localFileProof{}, admissionError(CodeLocalFormatUnsupported, "duplicate packed ref")
		}
		refs[ref] = oid
		previousTag = strings.HasPrefix(ref, "refs/tags/")
	}
	return refs, proof, nil
}

func exactOneLine(data []byte) (string, error) {
	if bytes.HasSuffix(data, []byte("\r\n")) {
		data = data[:len(data)-2]
	} else if bytes.HasSuffix(data, []byte{'\n'}) {
		data = data[:len(data)-1]
	}
	if len(data) == 0 || bytes.ContainsAny(data, "\r\n") {
		return "", errors.New("not one line")
	}
	return string(data), nil
}

func readProvedFile(path string, maxBytes int64) ([]byte, localFileProof, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, localFileProof{}, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() > maxBytes {
		return nil, localFileProof{}, errors.New("unsafe file")
	}
	file, err := os.Open(path) // #nosec G304 -- caller-selected local repository, lstat/recheck bounded.
	if err != nil {
		return nil, localFileProof{}, err
	}
	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	closeErr := file.Close()
	if err != nil || closeErr != nil || int64(len(data)) > maxBytes {
		return nil, localFileProof{}, errors.New("file read failed")
	}
	digest := sha256.Sum256(data)
	return data, localFileProof{path: path, size: info.Size(), mode: info.Mode(), modTime: info.ModTime(), digest: digest}, nil
}

func recheckProof(proof localFileProof) error {
	data, current, err := readProvedFile(proof.path, proof.size)
	if err != nil || current.size != proof.size || current.mode != proof.mode || !current.modTime.Equal(proof.modTime) || sha256.Sum256(data) != proof.digest {
		return errors.New("changed")
	}
	return nil
}

func validateLocalAdministration(gitDir, format string) error {
	allowed := map[string]bool{"HEAD": true, "config": true, "index": true, "objects": true, "refs": true, "packed-refs": true}
	entries, err := os.ReadDir(gitDir)
	if err != nil {
		return admissionError(CodeLocalLayoutUnsafe, "cannot inventory .git")
	}
	for _, entry := range entries {
		if !allowed[entry.Name()] {
			return admissionError(CodeLocalFormatUnsupported, "unexpected local Git administration child")
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return admissionError(CodeLocalLayoutUnsafe, "link in local Git administration")
		}
	}
	refsRoot := filepath.Join(gitDir, "refs")
	if err := filepath.WalkDir(refsRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return admissionError(CodeLocalLayoutUnsafe, "unsafe ref entry")
		}
		if entry.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(refsRoot, path)
		ref := "refs/" + filepath.ToSlash(rel)
		if strings.HasPrefix(ref, "refs/replace/") {
			return admissionError(CodeLocalFormatUnsupported, "replace refs are unsupported")
		}
		if !ValidRefName(strings.TrimPrefix(ref, "refs/")) {
			return admissionError(CodeLocalFormatUnsupported, "invalid loose ref name")
		}
		data, _, err := readProvedFile(path, 512)
		if err != nil {
			return admissionError(CodeLocalLayoutUnsafe, "unsafe loose ref")
		}
		value, err := exactOneLine(data)
		if err != nil || !validOID(value, format) {
			return admissionError(CodeLocalFormatUnsupported, "invalid loose ref")
		}
		return nil
	}); err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, forbidden := range []string{"info/grafts", "shallow", "objects/info/alternates", "objects/info/http-alternates", "objects/info/commit-graph", "objects/pack/multi-pack-index"} {
		if _, err := os.Lstat(filepath.Join(gitDir, filepath.FromSlash(forbidden))); err == nil {
			return admissionError(CodeLocalFormatUnsupported, "forbidden local Git state")
		}
	}
	if _, err := os.Lstat(filepath.Join(gitDir, "packed-refs")); err == nil {
		if _, _, err := readPackedRefs(gitDir, format); err != nil {
			return err
		}
	}
	return nil
}

func copyLocalObjects(source, destination, format string, limits Limits) ([]localFileProof, error) {
	hashHex := 40
	if format == "sha256" {
		hashHex = 64
	}
	entries, err := os.ReadDir(source)
	if err != nil {
		return nil, admissionError(CodeLocalLayoutUnsafe, "objects directory is unreadable")
	}
	proofs := []localFileProof{}
	packFiles := map[string]map[string]string{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.Type()&os.ModeSymlink != 0 || !entry.IsDir() {
			return nil, admissionError(CodeLocalLayoutUnsafe, "unsafe object inventory entry")
		}
		switch {
		case name == "info":
			children, err := os.ReadDir(filepath.Join(source, name))
			if err != nil {
				return nil, admissionError(CodeLocalLayoutUnsafe, "object info is unreadable")
			}
			if len(children) != 0 {
				return nil, admissionError(CodeLocalFormatUnsupported, "object info sidecars are unsupported")
			}
		case name == "pack":
			children, err := os.ReadDir(filepath.Join(source, name))
			if err != nil {
				return nil, admissionError(CodeLocalLayoutUnsafe, "pack directory is unreadable")
			}
			for _, child := range children {
				if child.Type()&os.ModeSymlink != 0 || !child.Type().IsRegular() {
					return nil, admissionError(CodeLocalLayoutUnsafe, "unsafe pack entry")
				}
				base, ext, ok := packName(child.Name(), hashHex)
				if !ok {
					return nil, admissionError(CodeLocalFormatUnsupported, "unsupported pack sidecar")
				}
				if packFiles[base] == nil {
					packFiles[base] = map[string]string{}
				}
				if packFiles[base][ext] != "" {
					return nil, admissionError(CodeLocalObjectFormatUnsupported, "duplicate pack member")
				}
				packFiles[base][ext] = filepath.Join(source, name, child.Name())
			}
		case len(name) == 2 && lowercaseHex(name):
			children, err := os.ReadDir(filepath.Join(source, name))
			if err != nil {
				return nil, admissionError(CodeLocalLayoutUnsafe, "loose object directory is unreadable")
			}
			for _, child := range children {
				if child.Type()&os.ModeSymlink != 0 || !child.Type().IsRegular() || len(child.Name()) != hashHex-2 || !lowercaseHex(child.Name()) {
					return nil, admissionError(CodeLocalObjectFormatUnsupported, "invalid loose object name")
				}
				src := filepath.Join(source, name, child.Name())
				dst := filepath.Join(destination, name, child.Name())
				if info, err := child.Info(); err != nil || info.Size() > limits.MaxObjectBytes {
					return nil, admissionError(CodeIncompleteSource, "loose object size limit exceeded")
				}
				proof, err := copyProvedFile(src, dst, limits.MaxObjectBytes)
				if err != nil {
					return nil, admissionError(CodeLocalLayoutUnsafe, "loose object copy failed")
				}
				proofs = append(proofs, proof)
			}
		default:
			return nil, admissionError(CodeLocalFormatUnsupported, "unexpected object inventory child")
		}
	}
	for base, members := range packFiles {
		if members[".pack"] == "" || members[".idx"] == "" {
			return nil, admissionError(CodeLocalObjectFormatUnsupported, "pack/index pair is incomplete")
		}
		pack, packProof, err := readProvedFile(members[".pack"], limits.MaxExpandedBytes)
		if err != nil {
			return nil, admissionError(CodeLocalLayoutUnsafe, "pack is unsafe")
		}
		index, indexProof, err := readProvedFile(members[".idx"], limits.MaxExpandedBytes)
		if err != nil {
			return nil, admissionError(CodeLocalLayoutUnsafe, "index is unsafe")
		}
		if err := validatePackIndex(base, pack, index, format); err != nil {
			return nil, err
		}
		for ext, data := range map[string][]byte{".pack": pack, ".idx": index} {
			if err := os.MkdirAll(filepath.Join(destination, "pack"), 0o700); err != nil {
				return nil, err
			}
			if err := os.WriteFile(filepath.Join(destination, "pack", base+ext), data, 0o600); err != nil {
				return nil, err
			}
		}
		proofs = append(proofs, packProof, indexProof)
	}
	return proofs, nil
}

func packName(name string, hashHex int) (string, string, bool) {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	return base, ext, (ext == ".pack" || ext == ".idx") && strings.HasPrefix(base, "pack-") && len(strings.TrimPrefix(base, "pack-")) == hashHex && lowercaseHex(strings.TrimPrefix(base, "pack-"))
}

func lowercaseHex(value string) bool {
	_, err := hex.DecodeString(value)
	return err == nil && value == strings.ToLower(value)
}

func copyProvedFile(source, destination string, maxBytes int64) (localFileProof, error) {
	data, proof, err := readProvedFile(source, maxBytes)
	if err != nil {
		return localFileProof{}, err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return localFileProof{}, err
	}
	if err := os.WriteFile(destination, data, 0o600); err != nil {
		return localFileProof{}, err
	}
	return proof, nil
}

func validatePackIndex(base string, pack, index []byte, format string) error {
	hashBytes := 20
	if format == "sha256" {
		hashBytes = 32
	}
	if len(pack) < 12+hashBytes || string(pack[:4]) != "PACK" {
		return admissionError(CodeLocalObjectFormatUnsupported, "malformed pack header")
	}
	version := binary.BigEndian.Uint32(pack[4:8])
	if version != 2 && version != 3 {
		return admissionError(CodeLocalObjectFormatUnsupported, "unsupported pack version")
	}
	packHash := objectHash(format, pack[:len(pack)-hashBytes])
	trailer := pack[len(pack)-hashBytes:]
	if !bytes.Equal(packHash, trailer) || "pack-"+hex.EncodeToString(trailer) != base {
		return admissionError(CodeLocalObjectFormatUnsupported, "pack checksum or name mismatch")
	}
	if len(index) < 8+256*4+2*hashBytes || !bytes.Equal(index[:4], []byte{0xff, 't', 'O', 'c'}) || binary.BigEndian.Uint32(index[4:8]) != 2 {
		return admissionError(CodeLocalObjectFormatUnsupported, "unsupported index")
	}
	previous := uint32(0)
	for i := 0; i < 256; i++ {
		current := binary.BigEndian.Uint32(index[8+i*4:])
		if current < previous {
			return admissionError(CodeLocalObjectFormatUnsupported, "non-monotonic fanout")
		}
		previous = current
	}
	count := int(previous)
	if uint32(count) != binary.BigEndian.Uint32(pack[8:12]) {
		return admissionError(CodeLocalObjectFormatUnsupported, "pack/index object counts differ")
	}
	fixed := 8 + 256*4 + count*hashBytes + count*4 + count*4
	minimum := fixed + 2*hashBytes
	if len(index) < minimum || !bytes.Equal(index[len(index)-2*hashBytes:len(index)-hashBytes], trailer) || !bytes.Equal(objectHash(format, index[:len(index)-hashBytes]), index[len(index)-hashBytes:]) {
		return admissionError(CodeLocalObjectFormatUnsupported, "index checksum or length mismatch")
	}
	oids := index[8+256*4 : 8+256*4+count*hashBytes]
	for i := 1; i < count; i++ {
		if bytes.Compare(oids[(i-1)*hashBytes:i*hashBytes], oids[i*hashBytes:(i+1)*hashBytes]) >= 0 {
			return admissionError(CodeLocalObjectFormatUnsupported, "index object IDs are not unique and sorted")
		}
	}
	crcStart := 8 + 256*4 + count*hashBytes
	offsetStart := crcStart + count*4
	largeStart := offsetStart + count*4
	largeBytes := len(index) - largeStart - 2*hashBytes
	if largeBytes < 0 || largeBytes%8 != 0 {
		return admissionError(CodeLocalObjectFormatUnsupported, "invalid large-offset table")
	}
	largeCount := largeBytes / 8
	usedLarge := make([]bool, largeCount)
	offsets := make([]uint64, count)
	seenOffsets := map[uint64]bool{}
	for i := 0; i < count; i++ {
		raw := binary.BigEndian.Uint32(index[offsetStart+i*4:])
		var offset uint64
		if raw&0x80000000 != 0 {
			position := int(raw & 0x7fffffff)
			if position >= largeCount || usedLarge[position] {
				return admissionError(CodeLocalObjectFormatUnsupported, "invalid large-offset reference")
			}
			usedLarge[position] = true
			offset = binary.BigEndian.Uint64(index[largeStart+position*8:])
		} else {
			offset = uint64(raw)
		}
		packEnd := len(pack) - hashBytes
		if packEnd < 0 || offset < 12 || offset >= uint64(packEnd) || seenOffsets[offset] { // #nosec G115 -- non-negative packEnd is proved before conversion.
			return admissionError(CodeLocalObjectFormatUnsupported, "invalid or duplicate pack offset")
		}
		seenOffsets[offset] = true
		offsets[i] = offset
	}
	for _, used := range usedLarge {
		if !used {
			return admissionError(CodeLocalObjectFormatUnsupported, "unreferenced large offset")
		}
	}
	type indexedOffset struct {
		offset uint64
		crc    uint32
	}
	ordered := make([]indexedOffset, count)
	for i := range offsets {
		ordered[i] = indexedOffset{offset: offsets[i], crc: binary.BigEndian.Uint32(index[crcStart+i*4:])}
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].offset < ordered[j].offset })
	for i, item := range ordered {
		packEnd := len(pack) - hashBytes
		if packEnd < 0 {
			return admissionError(CodeLocalObjectFormatUnsupported, "invalid pack length")
		}
		end := uint64(packEnd) // #nosec G115 -- packEnd is non-negative above.
		if i+1 < len(ordered) {
			end = ordered[i+1].offset
		}
		if end <= item.offset || crc32.ChecksumIEEE(pack[item.offset:end]) != item.crc {
			return admissionError(CodeLocalObjectFormatUnsupported, "pack entry CRC mismatch")
		}
	}
	return nil
}

func objectHash(format string, data []byte) []byte {
	if format == "sha1" {
		sum := sha1.Sum(data) // #nosec G401 -- validates Git SHA-1 container identity, not a security digest.
		return sum[:]
	}
	sum := sha256.Sum256(data)
	return sum[:]
}

func sealObjectStore(root string) error {
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		paths = append(paths, path)
		if !entry.IsDir() {
			return os.Chmod(path, 0o400)
		}
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(paths, func(i, j int) bool { return len(paths[i]) > len(paths[j]) })
	for _, path := range paths {
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := os.Chmod(path, 0o500); err != nil { // #nosec G302 -- sealed directories require owner execute for traversal.
				return err
			}
		}
	}
	return nil
}

// releasePrivateRoot removes an operation-private root after sealObjectStore
// took owner write permission away from every directory beneath it. RemoveAll
// cannot unlink an entry from a read-only directory, so the release first
// hands owner write back to each directory and only then removes the tree.
// A root that never existed is already released.
func releasePrivateRoot(root string) error {
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}
		return os.Chmod(path, 0o700) // #nosec G302 -- owner-only permission on a private directory about to be removed.
	})
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.RemoveAll(root)
}

var _ = fmt.Sprintf
var _ = strconv.IntSize
