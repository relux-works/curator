package rustsource

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/protocoljson"
)

// deriveRegistryTransform independently derives Cargo 0.92 registry vendoring.
func deriveRegistryTransform(origin registryOrigin) (VendorPackage, error) {
	if err := validatePackageKey(origin.Package); err != nil {
		return VendorPackage{}, fail(CodeRegistryIdentityInvalid, err.Error(), nil)
	}
	archiveDigest := digest(origin.Archive)
	if origin.Checksum != archiveDigest {
		return VendorPackage{}, fail(CodeRegistryIdentityInvalid, "archive checksum differs", map[string]string{"expected": origin.Checksum, "observed": archiveDigest})
	}
	var index struct {
		Name     string `json:"name"`
		Version  string `json:"vers"`
		Checksum string `json:"cksum"`
	}
	// Registry records are extensible, so decode the three authoritative fields separately.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(origin.IndexRecord, &raw); err != nil {
		return VendorPackage{}, fail(CodeRegistryIdentityInvalid, "registry index record is malformed", nil)
	}
	if err := json.Unmarshal(raw["name"], &index.Name); err != nil {
		return VendorPackage{}, fail(CodeRegistryIdentityInvalid, "index name missing", nil)
	}
	if err := json.Unmarshal(raw["vers"], &index.Version); err != nil {
		return VendorPackage{}, fail(CodeRegistryIdentityInvalid, "index version missing", nil)
	}
	if err := json.Unmarshal(raw["cksum"], &index.Checksum); err != nil {
		return VendorPackage{}, fail(CodeRegistryIdentityInvalid, "index checksum missing", nil)
	}
	if index.Name != origin.Package.Name || index.Version != origin.Package.Version || index.Checksum != origin.Checksum {
		return VendorPackage{}, fail(CodeRegistryIdentityInvalid, "lock, index, and archive identity differ", map[string]string{"package": origin.Package.String()})
	}
	leaves, err := readCrate(origin.Package, origin.Archive)
	if err != nil {
		return VendorPackage{}, err
	}
	result := VendorPackage{Package: origin.Package, Directory: packageDirectory(origin.Package), Kind: SourceRegistry}
	for _, leaf := range leaves {
		entry := TransformEntry{OriginPath: leaf.Path, VendorPath: leaf.Path, OriginSHA256: leaf.SHA256, ExpectedSHA256: leaf.SHA256, Size: leaf.Size, Disposition: CopyIdentical, Rule: "registry-copy-v1"}
		base := path.Base(leaf.Path)
		switch {
		case leaf.Path == ".cargo-checksum.json":
			return VendorPackage{}, fail(CodeRegistryIdentityInvalid, "origin contains manager-reserved checksum", map[string]string{"path": leaf.Path})
		case base == ".cargo-ok":
			entry.Disposition, entry.VendorPath, entry.Rule = OmitRegistryCargoOK, "", "registry-basename-cargo-ok-v1"
		case leaf.Path == ".git" || leaf.Path == ".gitignore" || leaf.Path == ".gitattributes":
			entry.Disposition, entry.VendorPath, entry.Rule = OmitReserved, "", "registry-exact-root-reserved-v1"
		default:
			result.Files = append(result.Files, cloneLeaf(leaf))
		}
		result.Entries = append(result.Entries, entry)
	}
	return addChecksum(result, origin.Checksum)
}

func readCrate(key PackageKey, payload []byte) ([]OriginLeaf, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(payload))
	if err != nil {
		return nil, fail(CodeRegistryIdentityInvalid, "crate is not a gzip stream", nil)
	}
	defer func() { _ = gzipReader.Close() }()
	reader := tar.NewReader(gzipReader)
	prefix := key.Name + "-" + key.Version + "/"
	leaves := []OriginLeaf{}
	seen := map[string]bool{}
	for {
		header, nextErr := reader.Next()
		if errors.Is(nextErr, io.EOF) {
			break
		}
		if nextErr != nil {
			return nil, fail(CodeRegistryIdentityInvalid, "crate tar is malformed", nil)
		}
		if !strings.HasPrefix(header.Name, prefix) {
			return nil, fail(CodeRegistryIdentityInvalid, "crate does not have exactly one package root", map[string]string{"path": header.Name})
		}
		rel := strings.TrimPrefix(header.Name, prefix)
		if rel == "" {
			continue
		}
		if !safeRelative(rel) {
			return nil, fail(CodeRegistryIdentityInvalid, "unsafe crate path", map[string]string{"path": header.Name})
		}
		if header.Typeflag == tar.TypeDir {
			continue
		}
		if header.Typeflag != tar.TypeReg {
			return nil, fail(CodeRegistryIdentityInvalid, "non-regular crate member", map[string]string{"path": header.Name})
		}
		if seen[rel] {
			return nil, fail(CodeRegistryIdentityInvalid, "duplicate crate member", map[string]string{"path": rel})
		}
		seen[rel] = true
		content, readErr := io.ReadAll(io.LimitReader(reader, header.Size+1))
		if readErr != nil || int64(len(content)) != header.Size {
			return nil, fail(CodeRegistryIdentityInvalid, "crate member size differs", map[string]string{"path": rel})
		}
		leaves = append(leaves, OriginLeaf{Path: rel, SHA256: digest(content), Size: int64(len(content)), Bytes: content})
	}
	sort.Slice(leaves, func(i, j int) bool { return leaves[i].Path < leaves[j].Path })
	return leaves, nil
}

// DeriveGitTransform derives one package projection from a complete admitted
// commit tree. NormalizedManifest must be independently produced by the pinned
// Cargo-normalizer implementation and is bound byte-for-byte here.
func deriveGitTransform(origin gitOrigin, derivation gitDerivation) (VendorPackage, error) {
	if derivation.receiptID == "" {
		return VendorPackage{}, fail(CodeGitIdentityInvalid, "manager-sealed Git derivation is missing", nil)
	}
	seal, err := gitDerivationSeal(derivation)
	if err != nil || seal != derivation.seal {
		return VendorPackage{}, fail(CodeGitIdentityInvalid, "sealed Git derivation was modified", nil)
	}
	include, _ := sortedUnique(origin.Include)
	originSubmodules, submoduleErr := canonicalSubmodules(origin.Submodules)
	if submoduleErr != nil {
		return VendorPackage{}, submoduleErr
	}
	if derivation.commit != origin.Commit || derivation.tree != origin.Tree || derivation.packagePath != origin.PackagePath || derivation.manifestTracked != origin.ManifestTracked || strings.Join(derivation.include, "\x00") != strings.Join(include, "\x00") || !reflect.DeepEqual(derivation.submodules, originSubmodules) {
		return VendorPackage{}, fail(CodeGitIdentityInvalid, "sealed Git derivation inputs differ from captured origin", nil)
	}
	if origin.Dirty || origin.UsesFilter || origin.IndexConflict || len(origin.Commit) != 40 || origin.Tree == "" {
		return VendorPackage{}, fail(CodeGitIdentityInvalid, "Git capture is not an exact clean commit/tree", map[string]string{"package": origin.Package.String()})
	}
	if derivation.mode != ProjectionGitIndexNoInclude && derivation.mode != ProjectionFilesystemInclude {
		return VendorPackage{}, fail(CodeGitIdentityInvalid, "unsupported Git projection branch", nil)
	}
	if derivation.mode == ProjectionGitIndexNoInclude && (len(origin.Include) != 0 || !origin.ManifestTracked) {
		return VendorPackage{}, fail(CodeGitIdentityInvalid, "git_index_no_include predicate is not proven", nil)
	}
	if derivation.mode == ProjectionFilesystemInclude && len(origin.Include) == 0 {
		return VendorPackage{}, fail(CodeGitIdentityInvalid, "filesystem_include predicate is not proven", nil)
	}
	if derivation.normalizerID != NormalizerID {
		return VendorPackage{}, fail(CodeVendorTransformUnsupported, "Git manifest normalizer is not pinned", map[string]string{"normalizer": derivation.normalizerID})
	}
	inputs, inputUnique := sortedUnique(derivation.normalizerInputs)
	if !inputUnique || len(inputs) == 0 {
		return VendorPackage{}, fail(CodeGitIdentityInvalid, "normalizer input set is empty or duplicated", nil)
	}
	selected, unique := sortedUnique(derivation.selected)
	if !unique {
		return VendorPackage{}, fail(CodeGitIdentityInvalid, "duplicate selected Git path", nil)
	}
	selectedSet := map[string]bool{}
	for _, item := range selected {
		if !safeRelative(item) {
			return VendorPackage{}, fail(CodeGitIdentityInvalid, "unsafe selected Git path", map[string]string{"path": item})
		}
		if derivation.mode == ProjectionFilesystemInclude && strings.HasPrefix(item, "target/") {
			return VendorPackage{}, fail(CodeGitIdentityInvalid, "filesystem_include selected package-root target", map[string]string{"path": item})
		}
		selectedSet[item] = true
	}
	for _, submodule := range origin.Submodules {
		if !safeRelative(submodule.Path) || len(submodule.Gitlink) != 40 || len(submodule.Commit) != 40 || submodule.TreeDigest == "" {
			return VendorPackage{}, fail(CodeGitIdentityInvalid, "submodule evidence is incomplete", map[string]string{"path": submodule.Path})
		}
	}
	result := VendorPackage{Package: origin.Package, Directory: packageDirectory(origin.Package), Kind: SourceGit}
	seen := map[string]bool{}
	hasManifest := false
	for _, leaf := range origin.Leaves {
		if !safeRelative(leaf.Path) || seen[leaf.Path] || leaf.SHA256 != digest(leaf.Bytes) || leaf.Size != int64(len(leaf.Bytes)) {
			return VendorPackage{}, fail(CodeGitIdentityInvalid, "invalid or duplicate Git leaf", map[string]string{"path": leaf.Path})
		}
		seen[leaf.Path] = true
		packageRelative, inside := trimPackagePath(origin.PackagePath, leaf.Path)
		entry := TransformEntry{OriginPath: leaf.Path, OriginSHA256: leaf.SHA256, Size: leaf.Size, Disposition: OmitUnselected, Rule: "git-outside-projection-v1"}
		if inside && selectedSet[packageRelative] {
			entry.VendorPath = packageRelative
			entry.ExpectedSHA256 = leaf.SHA256
			entry.Disposition = CopyIdentical
			entry.Rule = "git-copy-v1"
			switch packageRelative {
			case ".git", ".gitignore", ".gitattributes", ".cargo-ok":
				entry.Disposition, entry.VendorPath, entry.Rule = OmitReserved, "", "git-exact-root-reserved-v1"
			case "Cargo.toml":
				if len(derivation.normalizedManifest) == 0 {
					return VendorPackage{}, fail(CodeGitIdentityInvalid, "normalized Git manifest is missing", nil)
				}
				hasManifest = true
				entry.Disposition, entry.ExpectedSHA256, entry.Size, entry.Rule = ReplaceNormalizedManifest, digest(derivation.normalizedManifest), int64(len(derivation.normalizedManifest)), "cargo-0.92-normalized-manifest-v1"
				result.Files = append(result.Files, OriginLeaf{Path: "Cargo.toml", SHA256: entry.ExpectedSHA256, Size: entry.Size, Bytes: append([]byte(nil), derivation.normalizedManifest...)})
			default:
				result.Files = append(result.Files, OriginLeaf{Path: packageRelative, SHA256: leaf.SHA256, Size: leaf.Size, Bytes: append([]byte(nil), leaf.Bytes...)})
			}
		}
		result.Entries = append(result.Entries, entry)
	}
	for selectedPath := range selectedSet {
		full := selectedPath
		if origin.PackagePath != "" {
			full = path.Join(origin.PackagePath, selectedPath)
		}
		if !seen[full] {
			return VendorPackage{}, fail(CodeGitIdentityInvalid, "selected Git path is absent from commit", map[string]string{"path": selectedPath})
		}
	}
	if !hasManifest {
		return VendorPackage{}, fail(CodeGitIdentityInvalid, "selected Cargo.toml is missing", nil)
	}
	return addChecksum(result, "")
}

func trimPackagePath(root, value string) (string, bool) {
	if root == "" || root == "." {
		return value, true
	}
	root = strings.TrimSuffix(root, "/")
	if value == root {
		return "", true
	}
	if !strings.HasPrefix(value, root+"/") {
		return "", false
	}
	return strings.TrimPrefix(value, root+"/"), true
}
func cloneLeaf(leaf OriginLeaf) OriginLeaf {
	leaf.Bytes = append([]byte(nil), leaf.Bytes...)
	return leaf
}

func addChecksum(result VendorPackage, packageChecksum string) (VendorPackage, error) {
	sort.Slice(result.Files, func(i, j int) bool { return result.Files[i].Path < result.Files[j].Path })
	files := map[string]any{}
	seen := map[string]bool{}
	for _, leaf := range result.Files {
		if !safeRelative(leaf.Path) || seen[leaf.Path] {
			return VendorPackage{}, fail(CodeVendorIncomplete, "duplicate or unsafe vendor leaf", map[string]string{"path": leaf.Path})
		}
		seen[leaf.Path] = true
		files[leaf.Path] = leaf.SHA256
	}
	var packageValue any
	if packageChecksum != "" {
		packageValue = packageChecksum
	}
	checksum, err := protocoljson.MarshalCanonical(map[string]any{"files": files, "package": packageValue})
	if err != nil {
		return VendorPackage{}, err
	}
	checksumLeaf := OriginLeaf{Path: ".cargo-checksum.json", SHA256: digest(checksum), Size: int64(len(checksum)), Bytes: checksum}
	result.ChecksumBytes = append([]byte(nil), checksum...)
	result.Files = append(result.Files, checksumLeaf)
	result.Entries = append(result.Entries, TransformEntry{VendorPath: checksumLeaf.Path, Disposition: GenerateChecksum, ExpectedSHA256: checksumLeaf.SHA256, Size: checksumLeaf.Size, Rule: "cargo-checksum-compact-sorted-v1"})
	sort.Slice(result.Entries, func(i, j int) bool {
		a, b := result.Entries[i], result.Entries[j]
		return a.OriginPath+"\x00"+a.VendorPath < b.OriginPath+"\x00"+b.VendorPath
	})
	return result, nil
}

// VerifyVendor compares an absent-destination transform result against exact
// protected expected bytes. Cargo checksum metadata is corroboration only.
func VerifyVendor(root string, expected []VendorPackage) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return fail(CodeVendorIncomplete, "vendor destination is absent", nil)
	}
	expectedByDir := map[string]VendorPackage{}
	for _, item := range expected {
		if _, exists := expectedByDir[item.Directory]; exists {
			return fail(CodeVendorIncomplete, "multiple origins map to one vendor directory", map[string]string{"directory": item.Directory})
		}
		expectedByDir[item.Directory] = item
	}
	if len(entries) != len(expectedByDir) {
		return fail(CodeVendorIncomplete, "vendor package cardinality differs", map[string]string{"expected": fmt.Sprint(len(expectedByDir)), "observed": fmt.Sprint(len(entries))})
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			return fail(CodeVendorIncomplete, "unexpected vendor root leaf", map[string]string{"path": entry.Name()})
		}
		item, ok := expectedByDir[entry.Name()]
		if !ok {
			return fail(CodeVendorIncomplete, "vendor directory has no lock origin", map[string]string{"path": entry.Name()})
		}
		if err := verifyPackage(filepath.Join(root, entry.Name()), item); err != nil {
			return err
		}
	}
	return nil
}

func verifyPackage(root string, expected VendorPackage) error {
	observed := map[string]OriginLeaf{}
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == root {
			return nil
		}
		rel, relErr := filepath.Rel(root, current)
		if relErr != nil {
			return relErr
		}
		logical := filepath.ToSlash(rel)
		if entry.Type()&fs.ModeSymlink != 0 {
			return fail(CodeVendorIncomplete, "vendor contains link", map[string]string{"path": logical})
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fail(CodeVendorIncomplete, "vendor contains special node", map[string]string{"path": logical})
		}
		payload, readErr := os.ReadFile(current) // #nosec G304 -- WalkDir supplies a contained package member under the verified vendor root.
		if readErr != nil {
			return readErr
		}
		observed[logical] = OriginLeaf{Path: logical, SHA256: digest(payload), Size: int64(len(payload)), Bytes: payload}
		return nil
	})
	if err != nil {
		return err
	}
	if len(observed) != len(expected.Files) {
		return fail(CodeVendorIncomplete, "vendor leaf cardinality differs", map[string]string{"package": expected.Package.String()})
	}
	for _, leaf := range expected.Files {
		got, ok := observed[leaf.Path]
		if !ok {
			return fail(CodeVendorIncomplete, "expected vendor leaf missing", map[string]string{"path": leaf.Path})
		}
		if got.SHA256 != leaf.SHA256 || got.Size != leaf.Size || !bytes.Equal(got.Bytes, leaf.Bytes) {
			code := CodeRegistryIdentityInvalid
			if expected.Kind == SourceGit {
				code = CodeGitIdentityInvalid
			}
			return fail(code, "vendor leaf differs from protected transform", map[string]string{"path": leaf.Path, "expected": leaf.SHA256, "observed": got.SHA256})
		}
		delete(observed, leaf.Path)
	}
	if len(observed) != 0 {
		return fail(CodeVendorIncomplete, "unaccounted vendor leaf", nil)
	}
	return nil
}

// SourceReplacementConfig returns the one manager-owned Cargo replacement.
func SourceReplacementConfig(vendorDirectory string) ([]byte, error) {
	return DeriveSourceReplacementConfig(vendorDirectory, []LockPackage{{Kind: SourceRegistry, Key: PackageKey{Name: "placeholder", Version: "0", Source: "registry+https://github.com/rust-lang/crates.io-index"}}})
}

// DeriveSourceReplacementConfig emits one exact replacement for every remote
// lock source plus the single protected vendor directory.
func DeriveSourceReplacementConfig(vendorDirectory string, packages []LockPackage) ([]byte, error) {
	if !filepath.IsAbs(vendorDirectory) {
		return nil, fail(CodeConfigUntrusted, "vendor directory must be absolute", nil)
	}
	sources := map[string]SourceKind{}
	for _, pkg := range packages {
		if pkg.Kind == SourceRegistry || pkg.Kind == SourceGit {
			if prior, exists := sources[pkg.Key.Source]; exists && prior != pkg.Kind {
				return nil, fail(CodeVendorIncomplete, "one source has conflicting kinds", map[string]string{"source": pkg.Key.Source})
			}
			sources[pkg.Key.Source] = pkg.Kind
		}
	}
	keys := make([]string, 0, len(sources))
	for source := range sources {
		keys = append(keys, source)
	}
	sort.Strings(keys)
	var builder strings.Builder
	const cratesIO = "registry+https://github.com/rust-lang/crates.io-index"
	if _, present := sources[cratesIO]; present {
		builder.WriteString("[source.crates-io]\nreplace-with = \"vendored-sources\"\n\n")
	}
	for _, source := range keys {
		kind := sources[source]
		if source == cratesIO {
			continue
		}
		tableSource := source
		if kind == SourceGit {
			if index := strings.LastIndexByte(tableSource, '#'); index >= 0 {
				tableSource = tableSource[:index]
			}
		}
		builder.WriteString("[source.")
		builder.WriteString(strconv.Quote(tableSource))
		builder.WriteString("]\n")
		switch kind {
		case SourceRegistry:
			builder.WriteString("registry = ")
			builder.WriteString(strconv.Quote(strings.TrimPrefix(source, "registry+")))
			builder.WriteByte('\n')
		case SourceGit:
			gitURL := strings.TrimPrefix(source, "git+")
			selector := ""
			if index := strings.IndexByte(gitURL, '?'); index >= 0 {
				selector = gitURL[index+1:]
				if fragment := strings.IndexByte(selector, '#'); fragment >= 0 {
					selector = selector[:fragment]
				}
				gitURL = gitURL[:index]
			} else if index := strings.IndexByte(gitURL, '#'); index >= 0 {
				gitURL = gitURL[:index]
			}
			builder.WriteString("git = ")
			builder.WriteString(strconv.Quote(gitURL))
			builder.WriteByte('\n')
			if selector != "" {
				parts := strings.Split(selector, "&")
				if len(parts) != 1 {
					return nil, fail(CodeVendorIncomplete, "Git source has unsupported selector cardinality", map[string]string{"source": source})
				}
				key, value, found := strings.Cut(parts[0], "=")
				if !found || (key != "rev" && key != "branch" && key != "tag") || value == "" {
					return nil, fail(CodeVendorIncomplete, "Git source selector is unsupported", map[string]string{"source": source})
				}
				builder.WriteString(key)
				builder.WriteString(" = ")
				builder.WriteString(strconv.Quote(value))
				builder.WriteByte('\n')
			}
		default:
			return nil, fail(CodeVendorIncomplete, "unsupported replacement source", map[string]string{"source": source})
		}
		builder.WriteString("replace-with = \"vendored-sources\"\n\n")
	}
	builder.WriteString("[source.vendored-sources]\ndirectory = ")
	builder.WriteString(strconv.Quote(filepath.ToSlash(vendorDirectory)))
	builder.WriteByte('\n')
	return []byte(builder.String()), nil
}
