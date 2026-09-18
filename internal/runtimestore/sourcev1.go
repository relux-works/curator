package runtimestore

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/relux-works/curator/internal/identifiers"
)

// SourceV1Namespace names the distinct runtime-store namespace for
// skillfile-sources draft installations (skillfile-sources §4). Draft
// runtime trees never share a leaf with a legacy commit-keyed entry:
// legacy leaves are bare 40/64-hex commits while draft leaves always
// carry this prefix, so neither lane can adopt the other's bytes.
const SourceV1Namespace = "source-v1"

var sourceV1DigestRE = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// SourceV1Key derives the portable runtime-store leaf for one frozen draft
// package digest. The digest is SHA-256 over the CCJ-1 bytes of the locked
// package identity — never the bare snapshot digest and never a Git commit
// substituted into the other arm. A runtime edit followed by refresh
// changes the locked package, so it changes this key and requires new
// runtime publication instead of reusing the previous tree.
func SourceV1Key(packageDigest string) (string, error) {
	if !sourceV1DigestRE.MatchString(packageDigest) {
		return "", fmt.Errorf("draft package digest must be sha256:<64 lowercase hex>")
	}
	key := SourceV1Namespace + "-" + packageDigest[len("sha256:"):]
	if !identifiers.Valid(key) {
		return "", fmt.Errorf("draft runtime key is not a portable identifier")
	}
	return key, nil
}

// SourceV1Dir returns the runtime store location for a draft skill package:
// home/runtime/<skill>/<source-v1 key>. It keeps the existing per-skill
// layout so protected-store sweeping keeps working; only the leaf carries
// the namespace.
func SourceV1Dir(home, skillName, packageDigest string) (string, error) {
	if !identifiers.Valid(skillName) {
		return "", fmt.Errorf("draft runtime skill must be a portable identifier")
	}
	key, err := SourceV1Key(packageDigest)
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "runtime", skillName, key), nil
}
