package scriptworker

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/scriptpolicy"
)

var (
	errFarmTargetNotAbsolute = errors.New("farm target is not absolute")
	errFarmTargetUnusable    = errors.New("farm target is not a usable regular file")
)

// ExecIdentity is the strong per-invocation identity of one manager-resolved
// exec name: the fixed path derivation found plus the hash the launch
// boundary re-checks.
type ExecIdentity struct {
	// Name is the declared bare executable name.
	Name   string
	Path   string
	SHA256 string
	Size   int64
	// These fields retain the platform and trusted manager System32 root
	// needed to repeat the source identity check immediately before launch.
	platform            string
	trustedSystem32Root string
}

// ResolveExec resolves one declared bare exec name to a fixed path through
// the manager mechanism: the manager-owned search directories in order, and
// nothing else. The caller's PATH, the repository, a runtime root, a
// command directory, a manifest value, a project file, and script bytes
// are never consulted. Candidates under a forbidden root are skipped, as
// are non-regular files, untrusted hard links, and (on unix) files without
// an execute bit. Protected files in the manager's default Windows System32
// directory may be multiply linked and are copied into the private farm.
//
// A name no directory resolves is reported, not an error at this resolver
// layer: found is false. Derivation keeps it out of the farm and records it
// in UnresolvedExec. Only a malformed name is an error here, refused as
// package influence.
func ResolveExec(name string, searchDirs, forbiddenRoots []string) (ExecIdentity, bool, error) {
	return resolveExecForPlatform(name, searchDirs, forbiddenRoots, runtime.GOOS, nil, searchDirs == nil)
}

// resolveExecForPlatform is the injectable manager-resolution path used by
// DeriveProfile. Only the default Windows search list grants the System32
// hard-link allowance. The bound is the canonical System32 directory
// derived from the manager's captured SYSTEMROOT, and a candidate must be
// physically below that root. A component-store hard link there is still the
// manager-selected System32 file: the alias count does not redirect lookup
// to an untrusted path, and the manager hashes and rechecks its bytes before
// copying them into the private farm. Explicit search directories keep the
// ordinary single-link gate.
func resolveExecForPlatform(name string, searchDirs, forbiddenRoots []string, platform string, managerEnvironment []string, useDefaultSearchDirs bool) (ExecIdentity, bool, error) {
	if !validExecName(name) {
		return ExecIdentity{}, false, diagnostic(CodePackageInfluenceForbidden,
			"exec entry %q is not a bare executable name", name)
	}
	directories := searchDirs
	if directories == nil {
		directories = defaultExecSearchDirs(platform, managerEnvironment)
	}
	trustedSystem32Root := ""
	// A nil environment lets the default search list consult the process
	// environment, but it is not an explicit manager-captured snapshot. Keep
	// that ambient value out of the System32 hard-link trust decision.
	if platform == "windows" && useDefaultSearchDirs && managerEnvironment != nil {
		trustedSystem32Root = windowsSystem32Directory(managerEnvironment)
	}
	for _, directory := range directories {
		if directory == "" || !filepath.IsAbs(directory) {
			continue
		}
		for _, candidate := range execCandidates(name, platform) {
			path := filepath.Join(directory, candidate)
			identity, ok, err := readExecIdentity(name, path, forbiddenRoots, platform, trustedSystem32Root)
			if err != nil {
				return ExecIdentity{}, false, err
			}
			if ok {
				return identity, true, nil
			}
		}
	}
	return ExecIdentity{}, false, nil
}

// VerifyExec re-proves a resolved exec identity at the launch boundary, so
// a replacement between derivation and worker start refuses instead of
// reaching the built PATH.
func VerifyExec(identity ExecIdentity) error {
	platform := identity.platform
	if platform == "" {
		platform = runtime.GOOS
	}
	current, found, err := readExecIdentityAt(identity.Name, identity.Path, platform, identity.trustedSystem32Root)
	if err != nil {
		return err
	}
	if !found {
		return diagnostic(CodeWorkerIdentityInvalid,
			"the resolved exec %s is no longer usable", identity.Name)
	}
	if current.SHA256 != identity.SHA256 || current.Size != identity.Size {
		return diagnostic(CodeWorkerIdentityInvalid,
			"the resolved exec %s content changed", identity.Name)
	}
	return nil
}

// DefaultExecSearchDirs is the manager-owned fixed directory list exec
// names resolve from. It is compiled in, so it is independent of package
// data and of the caller's PATH by construction. On Windows, System32 and
// the SystemRoot directory come from the manager environment. A declared
// exec outside this list stays absent from the PATH farm and is reported.
func DefaultExecSearchDirs() []string {
	return defaultExecSearchDirs(runtime.GOOS, nil)
}

// defaultExecSearchDirs resolves the Windows system root from the manager
// environment used for this invocation. The caller passes the same captured
// environment used to build the worker environment, so declared exec lookup
// and the manager-set SYSTEMROOT value cannot observe different sources.
func defaultExecSearchDirs(platform string, managerEnvironment []string) []string {
	if platform == "windows" {
		root := managerSystemRoot(managerEnvironment)
		if root == "" {
			return nil
		}
		return []string{
			filepath.Join(root, "System32"),
			root,
		}
	}
	directories := []string{"/usr/local/bin", "/usr/bin", "/bin"}
	if platform == "darwin" {
		directories = append([]string{"/opt/homebrew/bin"}, directories...)
	}
	return directories
}

func managerSystemRoot(managerEnvironment []string) string {
	managerEnv, err := parseHostEnvironment(managerEnvironment, "windows")
	if err != nil {
		return ""
	}
	root, present := managerEnv.lookup("SYSTEMROOT")
	if !present || root == "" || !filepath.IsAbs(root) {
		return ""
	}
	return root
}

func windowsSystem32Directory(managerEnvironment []string) string {
	root := managerSystemRoot(managerEnvironment)
	if root == "" {
		return ""
	}
	directory, err := godriver.CanonicalPhysicalPath(filepath.Join(root, "System32"))
	if err != nil {
		return ""
	}
	return directory
}

// execCandidates lists the file names one declared name may resolve to,
// in probe order. On unix a bare name resolves to exactly itself; on
// Windows the platform extension search applies, so the declared name
// plus the executable extensions are probed.
func execCandidates(name, platform string) []string {
	if platform != "windows" {
		return []string{name}
	}
	if filepath.Ext(name) != "" {
		return []string{name}
	}
	return []string{name, name + ".exe", name + ".com", name + ".bat", name + ".cmd"}
}

// validExecName mirrors the manifest parser's bare-name rule so hand-built
// declarations cannot smuggle a path or command line past derivation.
func validExecName(name string) bool {
	if name == "" || strings.HasPrefix(name, "-") || strings.HasPrefix(name, ".") {
		return false
	}
	if strings.ContainsAny(name, "/\\ \t\x00") {
		return false
	}
	if name == "." || name == ".." {
		return false
	}
	return true
}

// readExecIdentity verifies one candidate file and records its identity.
// A candidate that fails verification is skipped (found false) rather
// than an error: other directories may still hold the name.
func readExecIdentity(name, path string, forbiddenRoots []string, platform, trustedSystem32Root string) (ExecIdentity, bool, error) {
	canonical, err := godriver.CanonicalPhysicalPath(path)
	if err != nil {
		return ExecIdentity{}, false, nil
	}
	for _, root := range forbiddenRoots {
		resolved, resolveErr := godriver.CanonicalPhysicalPath(root)
		if resolveErr != nil {
			continue
		}
		if canonical == resolved || isBelow(canonical, resolved) {
			return ExecIdentity{}, false, nil
		}
	}
	return readExecIdentityAt(name, canonical, platform, trustedSystem32Root)
}

// readExecIdentityAt records the identity of one canonical candidate: a
// regular executable file hashed over the open handle. The sole multiple-link
// exception is a manager-resolved Windows System32 executable: the OS
// component store normally hard-links protected binaries into System32, and
// the Windows farm copies the checked bytes into its private directory.
func readExecIdentityAt(name, canonical, platform, trustedSystem32Root string) (ExecIdentity, bool, error) {
	info, err := os.Lstat(canonical)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
		return ExecIdentity{}, false, nil
	}
	if info.Size() <= 0 || info.Size() > maxInterpreterBytes {
		return ExecIdentity{}, false, nil
	}
	if !interpreterExecutable(info.Mode()) {
		return ExecIdentity{}, false, nil
	}
	multiple, err := godriver.HasMultipleLinks(canonical, info)
	if err != nil {
		return ExecIdentity{}, false, nil
	}
	trustedSystem32File := trustedSystem32Root != "" && pathWithinPlatform(canonical, trustedSystem32Root, platform)
	if multiple && !trustedSystem32File {
		return ExecIdentity{}, false, nil
	}
	file, err := os.Open(canonical) // #nosec G304 -- canonical manager-resolved exec path
	if err != nil {
		return ExecIdentity{}, false, nil
	}
	defer func() { _ = file.Close() }()
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
		return ExecIdentity{}, false, nil
	}
	digest := sha256.New()
	written, err := io.CopyN(digest, file, opened.Size())
	var extra [1]byte
	extraCount, extraErr := file.Read(extra[:])
	if err != nil || written != opened.Size() || extraCount != 0 || (extraErr != nil && !errors.Is(extraErr, io.EOF)) {
		return ExecIdentity{}, false, nil
	}
	return ExecIdentity{
		Name:                name,
		Path:                canonical,
		SHA256:              "sha256:" + hex.EncodeToString(digest.Sum(nil)),
		Size:                opened.Size(),
		platform:            platform,
		trustedSystem32Root: trustedSystem32Root,
	}, true, nil
}

func pathWithinPlatform(path, root, platform string) bool {
	path = filepath.Clean(path)
	root = filepath.Clean(root)
	if platform == "windows" {
		if strings.EqualFold(path, root) {
			return true
		}
		root = strings.TrimRight(root, "/\\") + string(filepath.Separator)
		return strings.HasPrefix(strings.ToLower(path), strings.ToLower(root))
	}
	return path == root || isBelow(path, root)
}

// farmRequest is one manager-built PATH directory: the resolved
// interpreter plus the resolved declared exec names, and nothing else.
type farmRequest struct {
	parent      string
	interpreter InterpreterIdentity
	exec        []ExecIdentity
}

// buildPathFarm creates the manager-owned PATH directory exposing exactly
// the resolved interpreter and the resolved declared exec names. Entries
// expose the verified bytes — symlinks on unix, copies on Windows — so
// bare-name resolution inside the invocation reaches the fixed paths and
// nothing else. Entries are never hard links: an extra link would trip
// the single-link identity check at the launch boundary. A farm that cannot
// be built or verified means the manager-built-path control cannot be
// applied, which refuses with control-unavailable rather than launching
// with a wider or narrower PATH.
func buildPathFarm(request farmRequest) (string, []string, error) {
	unavailable := func(format string, args ...any) (string, []string, error) {
		return "", nil, &Diagnostic{
			Code:   scriptpolicy.ControlUnavailable,
			Detail: "manager-built-path cannot be applied: " + fmt.Sprintf(format, args...),
		}
	}
	if request.parent == "" || !filepath.IsAbs(request.parent) {
		return unavailable("the farm parent is not an absolute manager-owned directory")
	}
	if info, err := os.Stat(request.parent); err != nil || !info.IsDir() {
		return unavailable("the farm parent is not a directory: %v", err)
	}
	farm, err := os.MkdirTemp(request.parent, ".curator-path-")
	if err != nil {
		return unavailable("cannot create the manager-owned PATH directory: %v", err)
	}
	entries := map[string]string{}
	link := func(entry, target string) error {
		if _, repeated := entries[entry]; repeated {
			return errors.New("duplicate PATH entry " + entry)
		}
		if err := linkFarmEntry(farm, entry, target); err != nil {
			return err
		}
		entries[entry] = target
		return nil
	}
	interpreterEntry := filepath.Base(request.interpreter.Path)
	if interpreterEntry == "" || interpreterEntry == "." || interpreterEntry == string(filepath.Separator) {
		_ = os.RemoveAll(farm)
		return unavailable("the resolved interpreter path has no file name")
	}
	if err := link(interpreterEntry, request.interpreter.Path); err != nil {
		_ = os.RemoveAll(farm)
		return unavailable("cannot expose the resolved interpreter: %v", err)
	}
	ordered := append([]ExecIdentity(nil), request.exec...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Name < ordered[j].Name })
	for _, identity := range ordered {
		entry := farmEntryName(identity.Name, identity.Path)
		if err := link(entry, identity.Path); err != nil {
			_ = os.RemoveAll(farm)
			return unavailable("cannot expose the resolved exec %s: %v", identity.Name, err)
		}
	}
	// The farm holds exactly the linked entries: anything else in the
	// directory refuses rather than widening bare-name resolution.
	observed, err := os.ReadDir(farm)
	if err != nil {
		_ = os.RemoveAll(farm)
		return unavailable("cannot verify the manager-owned PATH directory: %v", err)
	}
	if len(observed) != len(entries) {
		_ = os.RemoveAll(farm)
		return unavailable("the manager-owned PATH directory holds %d entries, want %d",
			len(observed), len(entries))
	}
	names := make([]string, 0, len(entries))
	for _, item := range observed {
		if item.IsDir() {
			_ = os.RemoveAll(farm)
			return unavailable("the manager-owned PATH directory holds a subdirectory")
		}
		target, ok := entries[item.Name()]
		if !ok {
			_ = os.RemoveAll(farm)
			return unavailable("the manager-owned PATH directory holds an unmanaged entry")
		}
		if !farmEntryResolves(filepath.Join(farm, item.Name()), target) {
			_ = os.RemoveAll(farm)
			return unavailable("the PATH entry %s does not resolve to its fixed path", item.Name())
		}
		names = append(names, item.Name())
	}
	sort.Strings(names)
	return farm, names, nil
}

// farmEntryName names one farm entry. The interpreter keeps its own file
// name; a declared exec name keeps its declared spelling, gaining the
// target extension on Windows when the declaration carries none, so
// bare-name resolution finds the entry through the platform search.
func farmEntryName(declared, target string) string {
	if runtime.GOOS != "windows" {
		return declared
	}
	if filepath.Ext(declared) != "" {
		return declared
	}
	return declared + filepath.Ext(target)
}

// farmEntryResolves reports whether one farm entry reaches its fixed path:
// the same file for a link, or identical bytes for a copy.
func farmEntryResolves(entry, target string) bool {
	resolved, err := godriver.CanonicalPhysicalPath(entry)
	if err != nil {
		return false
	}
	if resolved == target {
		return true
	}
	entryHash, err := hashFileBytes(entry)
	if err != nil {
		return false
	}
	targetHash, err := hashFileBytes(target)
	if err != nil {
		return false
	}
	return entryHash == targetHash
}

// hashFileBytes hashes one regular file for farm verification.
func hashFileBytes(path string) (string, error) {
	file, err := os.Open(path) // #nosec G304 -- manager-owned farm path under verification
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > maxInterpreterBytes {
		return "", errFarmTargetUnusable
	}
	digest := sha256.New()
	if _, err := io.CopyN(digest, file, info.Size()); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), nil
}
