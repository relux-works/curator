package scriptworker

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/scriptpolicy"
)

// Closed interpreter identifiers of protocol 1.0. Admitting another
// identifier is a specification revision, never a manager configuration
// option, so resolution rejects anything outside this set without consulting
// any other source.
var closedInterpreters = map[string]bool{"node-v1": true, "python3-v1": true}

// maxInterpreterBytes bounds the interpreter executable hashed at the launch
// boundary. An interpreter installation is host-owned, so the bound is the
// same ceiling the manager identity uses rather than a tighter claim about
// any one interpreter's size.
const maxInterpreterBytes = int64(512 * 1024 * 1024)

// InterpreterIdentity is the strong per-invocation identity of the resolved
// interpreter executable file.
type InterpreterIdentity struct {
	ID     string
	Path   string
	SHA256 string
	Size   int64
}

// ResolveInterpreter resolves the declared closed identifier to an
// operator-trusted executable, once per invocation, before the manager enters
// any package-controlled directory.
//
// The mapping is the whole input: an identifier absent from it is refused,
// and no other source is consulted — not the repository, a runtime root,
// .agents/bin, the user PATH, a manifest value, or a script byte. The
// resolved target must be a canonical regular native executable image;
// symlink, reparse-point, and hard-link substitution are rejected, wrapper
// scripts and batch files are refused (they would interpose another program
// between the worker and the interpreter), strong file identity is recorded,
// and the bytes are hashed. The caller re-checks the returned identity at
// the launch boundary, and the worker re-verifies it before the interpreter
// starts.
func ResolveInterpreter(identifier string, mapping map[string]string, forbiddenRoots []string) (InterpreterIdentity, error) {
	if !closedInterpreters[identifier] {
		return InterpreterIdentity{}, diagnostic(CodePackageInfluenceForbidden,
			"interpreter identifier %q is not an admitted closed identifier", identifier)
	}
	configured, ok := mapping[identifier]
	if !ok || configured == "" {
		return InterpreterIdentity{}, &scriptpolicy.Error{
			DiagnosticCode: scriptpolicy.ControlUnavailable,
			State:          scriptpolicy.StateUnsupported,
			Severity:       scriptpolicy.SeverityError,
			Detail: "interpreter " + identifier +
				" has no operator-trusted machine configuration entry, so interpreter resolution cannot be applied",
		}
	}
	if !filepath.IsAbs(configured) {
		return InterpreterIdentity{}, diagnostic(CodeWorkerIdentityInvalid,
			"interpreter %s binding %q is not absolute", identifier, configured)
	}
	canonical, err := godriver.CanonicalPhysicalPath(configured)
	if err != nil {
		return InterpreterIdentity{}, diagnosticErr(CodeWorkerIdentityInvalid, err,
			"cannot canonicalize the resolved interpreter %s", identifier)
	}
	for _, root := range forbiddenRoots {
		resolved, resolveErr := godriver.CanonicalPhysicalPath(root)
		if resolveErr != nil {
			continue
		}
		if canonical == resolved || isBelow(canonical, resolved) {
			return InterpreterIdentity{}, diagnostic(CodePackageInfluenceForbidden,
				"interpreter %s resolves under a repository or runtime root", identifier)
		}
	}
	identity, err := readInterpreterIdentity(identifier, canonical)
	if err != nil {
		return InterpreterIdentity{}, err
	}
	return identity, nil
}

// readInterpreterIdentity records the strong identity of one canonical
// interpreter path: regular file, single link, native image, open-then-hash.
// It reuses the go-v1 worker's canonicalization, link, and image-header
// primitives so both workers agree on what substitution is.
func readInterpreterIdentity(identifier, canonical string) (InterpreterIdentity, error) {
	info, err := os.Lstat(canonical)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
		return InterpreterIdentity{}, diagnosticErr(CodeWorkerIdentityInvalid, err,
			"the resolved interpreter %s is not a canonical regular file", identifier)
	}
	if info.Size() <= 0 || info.Size() > maxInterpreterBytes {
		return InterpreterIdentity{}, diagnostic(CodeWorkerIdentityInvalid,
			"the resolved interpreter %s has an unusable size %d", identifier, info.Size())
	}
	if !interpreterExecutable(info.Mode()) {
		return InterpreterIdentity{}, diagnostic(CodeWorkerIdentityInvalid,
			"the resolved interpreter %s is not executable", identifier)
	}
	if !interpreterImageNameOK(canonical) {
		return InterpreterIdentity{}, diagnostic(CodeWorkerIdentityInvalid,
			"the resolved interpreter %s binding must name the native .exe image itself", identifier)
	}
	multiple, err := godriver.HasMultipleLinks(canonical, info)
	if err != nil {
		return InterpreterIdentity{}, diagnosticErr(CodeWorkerIdentityInvalid, err,
			"cannot inspect the resolved interpreter %s link count", identifier)
	}
	if multiple {
		return InterpreterIdentity{}, diagnostic(CodeWorkerIdentityInvalid,
			"the resolved interpreter %s has multiple filesystem links or is a reparse point", identifier)
	}
	file, err := os.Open(canonical) // #nosec G304 -- canonical operator-configured interpreter path
	if err != nil {
		return InterpreterIdentity{}, diagnosticErr(CodeWorkerIdentityInvalid, err,
			"cannot open the resolved interpreter %s", identifier)
	}
	defer func() { _ = file.Close() }()
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
		return InterpreterIdentity{}, diagnosticErr(CodeWorkerIdentityInvalid, err,
			"the resolved interpreter %s changed while opening", identifier)
	}
	// The native-image gate: a wrapper script (a POSIX shebang shim, a
	// Windows batch file) would interpose another program between the
	// worker and the interpreter, so only a native executable image is
	// accepted. The header judgment is the go-v1 worker's own, read from
	// the same open handle that is hashed below. ReadAt leaves the read
	// offset untouched for the hash.
	var header [8]byte
	read, readErr := file.ReadAt(header[:], 0)
	if readErr != nil && !errors.Is(readErr, io.EOF) && !errors.Is(readErr, io.ErrUnexpectedEOF) {
		return InterpreterIdentity{}, diagnosticErr(CodeWorkerIdentityInvalid, readErr,
			"cannot inspect the resolved interpreter %s", identifier)
	}
	if !godriver.NativeExecutableHeader(header[:read]) {
		return InterpreterIdentity{}, diagnostic(CodeWorkerIdentityInvalid,
			"the resolved interpreter %s is not a native interpreter image", identifier)
	}
	digest := sha256.New()
	written, err := io.CopyN(digest, file, opened.Size())
	var extra [1]byte
	extraCount, extraErr := file.Read(extra[:])
	if err != nil || written != opened.Size() || extraCount != 0 || (extraErr != nil && !errors.Is(extraErr, io.EOF)) {
		return InterpreterIdentity{}, diagnosticErr(CodeWorkerIdentityInvalid, err,
			"the resolved interpreter %s changed while hashing", identifier)
	}
	return InterpreterIdentity{
		ID:     identifier,
		Path:   canonical,
		SHA256: "sha256:" + hex.EncodeToString(digest.Sum(nil)),
		Size:   opened.Size(),
	}, nil
}

// interpreterImageNameOK reports whether path names a file the platform
// executes itself. Windows rewrites an extensionless absolute path to a
// neighboring PATHEXT sibling (os/exec lookExtensions), so the binding must
// carry the .exe extension there or the verified file and the executed file
// differ. Elsewhere any name is accepted and the image header decides.
func interpreterImageNameOK(path string) bool {
	if !platformRequiresExecutableExtension {
		return true
	}
	return hasWindowsExecutableExtension(path)
}

// hasWindowsExecutableExtension reports whether path carries the .exe
// extension, so lookExtensions returns it unchanged and the verified file
// is the executed file.
func hasWindowsExecutableExtension(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".exe")
}

// VerifyInterpreter re-proves a resolved interpreter identity at the launch
// boundary, so a replacement race between resolution and exec cannot widen
// the process graph.
func VerifyInterpreter(identity InterpreterIdentity) error {
	current, err := readInterpreterIdentity(identity.ID, identity.Path)
	if err != nil {
		return err
	}
	if current.SHA256 != identity.SHA256 || current.Size != identity.Size {
		return diagnostic(CodeWorkerIdentityInvalid,
			"the resolved interpreter %s content changed", identity.ID)
	}
	return nil
}

// matchesExpectation compares a worker-observed interpreter identity to the
// manager's expectation carried over the session protocol.
func (identity InterpreterIdentity) matchesExpectation(path, digest string, size int64) error {
	if identity.Path != path || identity.SHA256 != digest || identity.Size != size {
		return diagnostic(CodeWorkerIdentityInvalid,
			"interpreter identity proof %s/%s/%d does not match %s/%s/%d",
			path, digest, size, identity.Path, identity.SHA256, identity.Size)
	}
	return nil
}

func isBelow(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
