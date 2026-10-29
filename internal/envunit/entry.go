package envunit

import (
	"github.com/relux-works/curator/internal/entrydigest"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/pathboundary"
	"io"
	"os"
)

// ValidateEntryRead checks type without opening a payload. It precedes every
// executor digest and planner preimage: KindEntry alone permits recursive
// hashing, whereas nonrecursive home publication classes do not. Credential
// entries use their stricter link/empty-directory metadata contract.
func ValidateEntryRead(path, class string) error {
	if class == "credential-entry" {
		_, err := ObserveEntry(path, true)
		return err
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.IsDir() && !recursiveClass(class) {
		return invalid("tool state is outside nonrecursive publication scope")
	}
	return nil
}

// ObserveEntry reads precisely one declared owned entry. Credential entries
// admit only absence or a link: regular payloads are never opened or hashed.
// Directory images describe metadata only, never their descendants.
func ObserveEntry(path string, credential bool) (Image, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return Image{Kind: "absent"}, nil
	}
	if err != nil {
		return Image{}, err
	}
	if credential && info.Mode()&os.ModeSymlink == 0 {
		if !info.IsDir() {
			return Image{}, invalid("credential entry is not a link or empty directory")
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return Image{}, err
		}
		if len(entries) != 0 {
			return Image{}, invalid("credential directory is not empty")
		}
	}
	owner, err := pathboundary.DefaultOwnerLookup()(path, info)
	if err != nil {
		return Image{}, err
	}
	id, err := metadataIdentity(path, info)
	if err != nil {
		return Image{}, err
	}
	mode, protection, err := metadataProtection(path, info)
	if err != nil {
		return Image{}, err
	}
	im := Image{Identity: id, Owner: string(owner), Mode: mode, Protection: protection}
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		im.Kind = "link"
		im.LinkTarget, err = os.Readlink(path)
	case info.Mode().IsRegular():
		if err := pathboundary.CheckPrivateFile(path, info); err != nil {
			return Image{}, err
		}
		im.Kind = "file"
		f, openErr := pathboundary.OpenReadNoFollow(path)
		if openErr != nil {
			return Image{}, openErr
		}
		opened, statErr := f.Stat()
		if statErr != nil || !os.SameFile(info, opened) {
			_ = f.Close()
			return Image{}, invalid("owned file changed while opening")
		}
		raw, readErr := io.ReadAll(f)
		closeErr := f.Close()
		if readErr != nil {
			return Image{}, readErr
		}
		if closeErr != nil {
			return Image{}, closeErr
		}
		after, statErr := os.Lstat(path)
		if statErr != nil || !os.SameFile(info, after) || info.Size() != after.Size() || info.ModTime() != after.ModTime() || int64(len(raw)) != info.Size() {
			return Image{}, invalid("owned file changed while reading")
		}
		if err := pathboundary.CheckPrivateFile(path, after); err != nil {
			return Image{}, err
		}
		im.Digest = envorigin.Digest(raw)
	case info.IsDir():
		im.Kind = "directory"
		if credential {
			im.Kind = "empty-directory"
		}
	default:
		return Image{}, invalid("unsupported owned entry type")
	}
	after, statErr := os.Lstat(path)
	if statErr != nil || !os.SameFile(info, after) || info.Mode() != after.Mode() {
		return Image{}, invalid("owned entry changed while observing")
	}
	if err == nil {
		modeAfter, protectionAfter, protectionErr := metadataProtection(path, after)
		if protectionErr != nil || modeAfter != im.Mode || protectionAfter != im.Protection {
			return Image{}, invalid("owned entry protection changed while observing: %v", protectionErr)
		}
	}
	return im, err
}

// ObserveTarget permits recursive byte identity only for an explicitly owned
// projection/backup tree. Home directories and credentials never use this
// class: their descendants may include unrelated sessions or secrets.
func ObserveTarget(path, class string) (Image, error) {
	im, err := ObserveEntry(path, class == "credential-entry")
	if err != nil || im.Kind != "directory" || !recursiveClass(class) {
		return im, err
	}
	im.Digest, err = entrydigest.DigestTarget("entry", path)
	if err != nil {
		return Image{}, err
	}
	im.Kind = "tree"
	return im, nil
}
