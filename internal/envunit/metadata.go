package envunit

import (
	"github.com/relux-works/curator/internal/pathboundary"
	"os"
)

// ObserveDirectory pins a single no-follow directory entry. Children are never
// enumerated, copied or hashed. The caller separately proves its parent route.
func ObserveDirectory(path string) (Image, error) {
	if !absolute(path) {
		return Image{}, invalid("directory path must be absolute")
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return Image{Kind: "absent"}, nil
	}
	if err != nil {
		return Image{}, invalid("directory metadata: %v", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return Image{}, invalid("expected no-follow directory")
	}
	if info.Mode()&(os.ModeSetuid|os.ModeSetgid|os.ModeSticky) != 0 {
		return Image{}, invalid("special directory protection is unsupported")
	}
	if err := pathboundary.CheckPrivateFile(path, info); err != nil {
		return Image{}, invalid("directory ownership/protection: %v", err)
	}
	owner, err := pathboundary.DefaultOwnerLookup()(path, info)
	if err != nil {
		return Image{}, invalid("directory owner: %v", err)
	}
	id, err := metadataIdentity(path, info)
	if err != nil {
		return Image{}, invalid("directory identity: %v", err)
	}
	mode, protection, err := metadataProtection(path, info)
	if err != nil {
		return Image{}, invalid("directory protection metadata: %v", err)
	}
	return Image{Kind: "directory", Identity: id, Owner: string(owner), Mode: mode, Protection: protection}, nil
}
