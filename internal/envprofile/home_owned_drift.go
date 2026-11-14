package envprofile

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/stateread"
)

// homeOwnedDrift compares derived output with its recorded receipt, independently
// of today's source bytes. Reassembling from a mismatching store alone cannot
// tell a changed copy from an unchanged copy of the earlier verified inputs.
func homeOwnedDrift(req *ResolveRequest, v *verification) bool {
	if v.marker == nil || v.plan == nil {
		return false
	}
	version := markerHashVersion(v.marker)
	for key, surface := range v.marker.Surfaces {
		expected, ok := v.plan.marker.Surfaces[key]
		if !ok || !equalStrings(surface.Paths, expected.Paths) {
			return true
		}
		if key == envmarker.SurfaceSkills {
			if recordedSkillSurfaceDrift(req, v, surface, version) {
				return true
			}
			continue
		}
		files := map[string][]byte{}
		complete := true
		copied := map[string]bool{}
		if surface.Copies != nil {
			for _, copy := range *surface.Copies {
				copied[copy.Path] = true
			}
		}
		for _, path := range surface.Paths {
			full := filepath.Join(v.plan.homeDir, filepath.FromSlash(path))
			state, err := stateread.Lstat(full)
			if err != nil || state.Kind != stateread.KindPresent || state.Info == nil {
				return true
			}
			target, linked := v.plan.links[path]
			if linked && !copied[path] {
				if state.Info.Mode()&os.ModeSymlink == 0 {
					return true
				}
				link, err := stateread.Readlink(full)
				if err != nil || link.Kind != stateread.KindPresent || link.Target != target {
					return true
				}
				// Changes through an intact raw-store link belong to source evidence.
				if sameStoreTree(target, contextstore.Root(req.Home)) {
					complete = false
					continue
				}
			} else if state.Info.Mode()&os.ModeSymlink != 0 {
				return true
			}
			observed, err := stateread.Stat(full)
			if err != nil || observed.Kind != stateread.KindPresent || observed.Info == nil {
				return true
			}
			if observed.Info.IsDir() {
				complete = false
				continue
			}
			if !observed.Info.Mode().IsRegular() {
				return true
			}
			data, err := stateread.ReadFile(full) // #nosec G304 -- marker path below the verified managed home
			if err != nil || data.Kind != stateread.KindPresent {
				return true
			}
			files[path] = data.Bytes
		}
		// Referenced root copies contain the include plan, independent of raw
		// module bytes. Mixed surfaces can compare that copy directly as well.
		if !complete && surface.Form == "referenced" {
			for path, data := range files {
				if expected, ok := v.plan.copies[path]; ok && contextmaterialize.FileHash(data) != contextmaterialize.FileHash(expected) {
					return true
				}
			}
		}
		if complete {
			actual, err := contextmaterialize.SurfaceHashWithVersion(files, version)
			if err != nil || actual != surface.ContentSHA256 {
				return true
			}
		}
	}
	return false
}

// A skill link's content-addressed projection target is itself a receipt.
// Compare that original identity with the marker, rather than a new projection
// path derived from mismatching skill source bytes.
func recordedSkillSurfaceDrift(req *ResolveRequest, v *verification, surface envmarker.Surface, version hashing.Version) bool {
	files := map[string][]byte{}
	for _, path := range surface.Paths {
		full := filepath.Join(v.plan.homeDir, filepath.FromSlash(path))
		info, err := os.Lstat(full)
		if err != nil {
			return true
		}
		var identity string
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(full)
			if err != nil {
				return true
			}
			digest := filepath.Base(target)
			raw, err := hex.DecodeString(digest)
			if err != nil || len(raw) != 32 {
				return true
			}
			expected := filepath.Join(contextstore.Root(req.Home), "skill-context", skillProjectionVersion, fmt.Sprintf("v%d", version), digest)
			if target != expected {
				return true
			}
			identity = "sha256:" + digest
			if err := verifySkillProjection(req.Home, skillOf{root: target, hash: identity, version: version}); err != nil {
				return true
			}
		} else if info.IsDir() {
			var err error
			identity, err = hashing.ContentSHA256WithVersion(full, nil, version)
			if err != nil {
				return true
			}
		} else {
			return true
		}
		files[path] = skillSurfaceIdentity(skillOf{hash: identity})
	}
	actual, err := contextmaterialize.SurfaceHashWithVersion(files, version)
	return err != nil || actual != surface.ContentSHA256
}

func homeInputsFailure(ownedHomeDrift bool, err error) error {
	if ownedHomeDrift {
		return fmt.Errorf("%s: home cannot be restored from unverified inputs; rerun with --repair after restoring source evidence: %w", DiagHomeStale, err)
	}
	return err
}
