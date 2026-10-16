package envsnapshot

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"syscall"
)

// Verified retains immutable metadata bytes, not a mutable caller-owned model.
// It is evidence at the Verify observation, not an authorization or a lease.
type Verified struct{ manifest, launch []byte }

// ManifestSHA256 returns the exact, source-bound LF-inclusive identity.
func (v *Verified) ManifestSHA256() string { return Digest(v.manifest) }

// LaunchSHA256 returns the exact LF-inclusive descriptor digest.
func (v *Verified) LaunchSHA256() string { return Digest(v.launch) }

// ManifestBytes returns an independent copy of the observed manifest.
func (v *Verified) ManifestBytes() []byte { return bytes.Clone(v.manifest) }

// DescriptorBytes returns an independent copy of the observed descriptor.
func (v *Verified) DescriptorBytes() []byte { return bytes.Clone(v.launch) }

// Verify reads all metadata and content through retained no-follow handles.
// expectedManifestSHA256 is the ORIGINAL externally admitted exact identity;
// empty means inspect an unadmitted candidate (e.g. destination idempotence).
// It never writes, repairs, recaptures or replaces a persisted baseline.
func Verify(root string, expectedManifestSHA256 string) (*Verified, error) {
	dir, err := openDirectory(root)
	if err != nil {
		return nil, refuse("tree-unreadable")
	}
	result, verifyErr := verifyDirectory(dir, expectedManifestSHA256)
	closeErr := dir.Close()
	if verifyErr != nil {
		return nil, verifyErr
	}
	if closeErr != nil {
		return nil, refuse("tree-unreadable")
	}
	return result, nil
}
func verifyDirectory(dir *os.File, expected string) (*Verified, error) {
	manifest, err := readMetadata(dir, "manifest.json")
	if err != nil {
		return nil, err
	}
	if expected != "" && Digest(manifest) != expected {
		return nil, refuse("baseline-mismatch")
	}
	launch, err := readMetadata(dir, "launch.json")
	if err != nil {
		return nil, err
	}
	m, err := DecodeManifest(manifest)
	if err != nil {
		return nil, err
	}
	d, err := DecodeDescriptor(launch)
	if err != nil {
		return nil, err
	}
	dirs, err := validateMetadata(m, d, launch)
	if err != nil {
		return nil, err
	}
	files := map[string]File{}
	for _, f := range m.Files {
		files[f.Path] = f
	}
	files["manifest.json"] = File{Path: "manifest.json", Size: int64(len(manifest)), SHA256: Digest(manifest), Mode: "0600"}
	files["launch.json"] = File{Path: "launch.json", Size: int64(len(launch)), SHA256: Digest(launch), Mode: "0600"}
	if err := verifyEntries(dir, "", dirs, files); err != nil {
		return nil, err
	}
	if len(files) != 0 || len(dirs) != 0 {
		return nil, refuse("entry-set")
	}
	return &Verified{bytes.Clone(manifest), bytes.Clone(launch)}, nil
}
func readMetadata(dir *os.File, name string) (data []byte, err error) {
	f, err := openMember(dir, name, false)
	if err != nil {
		return nil, refuse("metadata-unreadable")
	}
	defer func() {
		if closeErr := f.Close(); err == nil && closeErr != nil {
			data = nil
			err = refuse("metadata-unreadable")
		}
	}()
	info, err := f.Stat()
	if err != nil || !regularIndependent(info) || !modeMatches(info, "0600") {
		return nil, refuse("metadata-kind-mode")
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, refuse("metadata-unreadable")
	}
	return b, nil
}
func verifyEntries(dir *os.File, prefix string, dirs map[string]bool, files map[string]File) error {
	before, err := dir.Stat()
	if err == nil && before.Mode().Perm()&0500 != 0500 {
		return refuse("tree-unreadable")
	}
	if err != nil || !directoryMode(before) {
		return refuse("directory-mode")
	}
	entries, err := dir.ReadDir(-1)
	if err != nil {
		return refuse("tree-unreadable")
	}
	for _, entry := range entries {
		name := entry.Name()
		p := name
		if prefix != "" {
			p = prefix + "/" + name
		}
		if dirs[p] {
			child, err := openMember(dir, name, true)
			if err != nil {
				if errors.Is(err, syscall.EACCES) {
					return refuse("tree-unreadable")
				}
				return refuse("directory-kind")
			}
			err = verifyEntries(child, p, dirs, files)
			closeErr := child.Close()
			if err != nil {
				return err
			}
			if closeErr != nil {
				return refuse("tree-unreadable")
			}
			delete(dirs, p)
			continue
		}
		expected, ok := files[p]
		if !ok {
			return refuse("entry-set")
		}
		f, err := openMember(dir, name, false)
		if err != nil {
			if errors.Is(err, syscall.EACCES) {
				return refuse("file-unreadable")
			}
			return refuse("file-kind")
		}
		err = verifyFile(f, expected)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return refuse("file-unreadable")
		}
		delete(files, p)
	}
	after, err := dir.Stat()
	if err != nil || !os.SameFile(before, after) || !before.ModTime().Equal(after.ModTime()) {
		return refuse("tree-drift")
	}
	return nil
}
func verifyFile(f *os.File, want File) error {
	before, err := f.Stat()
	if err != nil || !regularIndependent(before) {
		return refuse("file-kind")
	}
	if before.Mode().Perm()&0400 == 0 {
		return refuse("file-unreadable")
	}
	if !modeMatches(before, want.Mode) {
		return refuse("file-mode")
	}
	if before.Size() != want.Size {
		return refuse("file-identity")
	}
	hash := sha256.New()
	n, err := io.Copy(hash, io.LimitReader(f, want.Size+1))
	if err != nil {
		return refuse("file-unreadable")
	}
	if n != want.Size || hex.EncodeToString(hash.Sum(nil)) != want.SHA256 {
		return refuse("file-identity")
	}
	after, err := f.Stat()
	if err != nil || !os.SameFile(before, after) || !regularIndependent(after) || !modeMatches(after, want.Mode) || after.Size() != before.Size() || !before.ModTime().Equal(after.ModTime()) {
		return refuse("tree-drift")
	}
	return nil
}

// PortableEqual verifies BOTH actual trees anew. Only fragment_sha256 is
// projected out; every other manifest field and exact descriptor byte binds.
func PortableEqual(left, right string) (bool, error) {
	a, err := Verify(left, "")
	if err != nil {
		return false, err
	}
	b, err := Verify(right, "")
	if err != nil {
		return false, err
	}
	am, err := DecodeManifest(a.manifest)
	if err != nil {
		return false, err
	}
	bm, err := DecodeManifest(b.manifest)
	if err != nil {
		return false, err
	}
	am.FragmentSHA256 = ""
	bm.FragmentSHA256 = ""
	// Encoding this projection directly is intentional: empty provenance is not
	// a valid full manifest, and must never be passed to the manifest codec.
	ac, err := project(am)
	if err != nil {
		return false, err
	}
	bc, err := project(bm)
	if err != nil {
		return false, err
	}
	return bytes.Equal(a.launch, b.launch) && bytes.Equal(ac, bc), nil
}
