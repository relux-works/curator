package godriver

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// maxExecutableBytes bounds the manager executable hashed at the worker
// identity boundary.
const maxExecutableBytes = int64(512 * 1024 * 1024)

// ExecutableIdentity is the strong identity of the installed manager
// executable. It is recorded before the worker is launched and re-proved by the
// worker itself, at the launch boundary, and again after the last child exits.
type ExecutableIdentity struct {
	Path   string
	SHA256 string
	Size   int64

	info fs.FileInfo
}

// resolveExecutableIdentity canonicalizes path to a real installed regular
// file, rejects symlink, reparse-point, and hard-link substitution, records
// strong file identity, and hashes the bytes.
func resolveExecutableIdentity(path string) (ExecutableIdentity, error) {
	if path == "" {
		return ExecutableIdentity{}, diagnostic(CodeWorkerIdentityInvalid, "manager executable path is empty")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return ExecutableIdentity{}, diagnosticErr(CodeWorkerIdentityInvalid, err, "cannot resolve the manager executable")
	}
	canonical, err := physicalPath(filepath.Clean(absolute))
	if err != nil {
		return ExecutableIdentity{}, diagnosticErr(CodeWorkerIdentityInvalid, err, "cannot canonicalize the manager executable")
	}
	identity, err := readExecutableIdentity(canonical)
	if err != nil {
		return ExecutableIdentity{}, err
	}
	return identity, nil
}

func readExecutableIdentity(canonical string) (ExecutableIdentity, error) {
	info, err := os.Lstat(canonical)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
		return ExecutableIdentity{}, diagnosticErr(CodeWorkerIdentityInvalid, err, "the manager executable is not a canonical regular file")
	}
	if info.Size() <= 0 || info.Size() > maxExecutableBytes {
		return ExecutableIdentity{}, diagnostic(CodeWorkerIdentityInvalid, "the manager executable has an unusable size %d", info.Size())
	}
	multiple, err := artifactHasMultipleLinks(canonical, info)
	if err != nil {
		return ExecutableIdentity{}, diagnosticErr(CodeWorkerIdentityInvalid, err, "cannot inspect the manager executable link count")
	}
	if multiple {
		return ExecutableIdentity{}, diagnostic(CodeWorkerIdentityInvalid, "the manager executable has multiple filesystem links or is a reparse point")
	}
	file, err := os.Open(canonical) // #nosec G304 -- canonical path of this manager executable
	if err != nil {
		return ExecutableIdentity{}, diagnosticErr(CodeWorkerIdentityInvalid, err, "cannot open the manager executable")
	}
	defer func() { _ = file.Close() }()
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
		return ExecutableIdentity{}, diagnosticErr(CodeWorkerIdentityInvalid, err, "the manager executable changed while opening")
	}
	digest := sha256.New()
	written, err := io.CopyN(digest, file, opened.Size())
	var extra [1]byte
	extraCount, extraErr := file.Read(extra[:])
	if err != nil || written != opened.Size() || extraCount != 0 || (extraErr != nil && !errors.Is(extraErr, io.EOF)) {
		return ExecutableIdentity{}, diagnosticErr(CodeWorkerIdentityInvalid, err, "the manager executable changed while hashing")
	}
	return ExecutableIdentity{
		Path:   canonical,
		SHA256: "sha256:" + hex.EncodeToString(digest.Sum(nil)),
		Size:   opened.Size(),
		info:   info,
	}, nil
}

// Verify re-proves the recorded identity. It is called at the launch boundary
// so a replacement race cannot widen the process graph, and again after the
// last child exits and before publication.
func (identity ExecutableIdentity) Verify() error {
	current, err := readExecutableIdentity(identity.Path)
	if err != nil {
		return err
	}
	if identity.info != nil && !os.SameFile(identity.info, current.info) {
		return diagnostic(CodeWorkerIdentityInvalid, "the manager executable was replaced")
	}
	if current.SHA256 != identity.SHA256 || current.Size != identity.Size {
		return diagnostic(CodeWorkerIdentityInvalid, "the manager executable content changed")
	}
	return nil
}

// matches compares an expectation carried over the worker protocol.
func (identity ExecutableIdentity) matches(path, digest string, size int64) error {
	if identity.Path != path || identity.SHA256 != digest || identity.Size != size {
		return diagnostic(CodeWorkerIdentityInvalid,
			"worker identity proof %s/%s/%d does not match %s/%s/%d",
			path, digest, size, identity.Path, identity.SHA256, identity.Size)
	}
	return nil
}
