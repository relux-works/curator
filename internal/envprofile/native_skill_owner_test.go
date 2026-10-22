package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/envmarker"
)

func nativeOwnerFixture(t *testing.T) (*managedFixture, string, contextlock.Member) {
	t.Helper()
	fx, homes := nativeSkillsFixture(t)
	if _, err := UseWithPolicy(fx.home, fx.profile, "", "", false, Policy{}); err != nil {
		t.Fatal(err)
	}
	lock, _, err := readLock(fx.home, fx.profile)
	if err != nil {
		t.Fatal(err)
	}
	for _, member := range lock.Members {
		if member.Kind == contextlock.KindSkill && member.Name == "myskill" {
			return fx, homes["claude_code"], member
		}
	}
	t.Fatal("fixture skill missing")
	return nil, "", contextlock.Member{}
}

func writeNativeOwnerMarker(t *testing.T, native string, marker *envmarker.Marker) {
	t.Helper()
	payload, err := marker.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(native, envmarker.Name), payload, 0600); err != nil {
		t.Fatal(err)
	}
}

// The ownership bridge requires exact source, commit and directory agreement,
// a matching lock and surface record, and the recorded live projection. Each
// negative follows a valid positive control and leaves the native home intact.
func TestNativeSkillOwnerProof(t *testing.T) {
	for _, variant := range []string{"source", "commit", "directory", "lock", "surface", "live-link", "unrecorded", "absent-marker", "unreadable-marker", "parent-link"} {
		t.Run(variant, func(t *testing.T) {
			fx, native, member := nativeOwnerFixture(t)
			root := filepath.Join(native, "skills")
			owned, err := OwnsNativeSkill(fx.home, root, "myskill", &member)
			if err != nil || !owned {
				t.Fatalf("matching-owner positive control refused: owned=%v %v", owned, err)
			}
			marker := homeMarker(t, native)
			want := DiagSurfaceDrift
			switch variant {
			case "source":
				member.Source = "example.com/other-skill"
				want = "context_range_conflict"
			case "commit":
				member.Commit = strings.Repeat("1", 40)
				want = "context_range_conflict"
			case "directory":
				member.Directory = "nested"
				want = "context_range_conflict"
			case "lock":
				marker.Profile.LockSHA256 = strings.Repeat("0", 64)
				writeNativeOwnerMarker(t, native, marker)
			case "surface":
				surface := marker.Surfaces[envmarker.SurfaceSkills]
				surface.ContentSHA256 = "sha256:" + strings.Repeat("0", 64)
				marker.Surfaces[envmarker.SurfaceSkills] = surface
				writeNativeOwnerMarker(t, native, marker)
			case "live-link":
				target := filepath.Join(root, "myskill")
				if err := os.Remove(target); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink("../../foreign/missing", target); err != nil {
					t.Fatal(err)
				}
			case "unrecorded":
				lock, _, err := readLock(fx.home, fx.profile)
				if err != nil {
					t.Fatal(err)
				}
				surface := marker.Surfaces[envmarker.SurfaceSkills]
				surface.Paths = []string{}
				surface.ContentSHA256, err = contextmaterialize.SurfaceHashWithVersion(map[string][]byte{}, lock.ContentHashVersion())
				if err != nil {
					t.Fatal(err)
				}
				marker.Surfaces[envmarker.SurfaceSkills] = surface
				writeNativeOwnerMarker(t, native, marker)
				want = ""
			case "absent-marker", "unreadable-marker":
				if err := os.Remove(filepath.Join(native, envmarker.Name)); err != nil {
					t.Fatal(err)
				}
				want = ""
				if variant == "unreadable-marker" {
					if err := os.Mkdir(filepath.Join(native, envmarker.Name), 0700); err != nil {
						t.Fatal(err)
					}
					want = envmarker.DiagMarkerUnreadable
				}
			case "parent-link":
				moved := filepath.Join(native, "other-skills")
				if err := os.Rename(root, moved); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(moved, root); err != nil {
					t.Fatal(err)
				}
				want = DiagWriteWouldFollowLink
			}
			before := nativeSnapshot(t, native)
			owned, err = OwnsNativeSkill(fx.home, root, "myskill", &member)
			if owned || (want == "" && err != nil) || (want != "" && (err == nil || !strings.Contains(err.Error(), want))) {
				t.Fatalf("invalid native ownership admitted or wrong diagnostic: owned=%v err=%v want=%q", owned, err, want)
			}
			if after := nativeSnapshot(t, native); after != before {
				t.Fatal("ownership verification wrote native state")
			}
		})
	}
}

// A recorded copied projection is owned only while its bytes match the pin.
// Linked-mode fallback requires an explicit copy record; an unrecorded plain
// directory never gains ownership from the marker's path alone.
func TestNativeSkillOwnerCopyProof(t *testing.T) {
	for _, fallback := range []bool{false, true} {
		t.Run(map[bool]string{false: "copied", true: "fallback"}[fallback], func(t *testing.T) {
			fx, native, member := nativeOwnerFixture(t)
			root := filepath.Join(native, "skills")
			target := filepath.Join(root, "myskill")
			link, err := os.Readlink(target)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(target); err != nil {
				t.Fatal(err)
			}
			if err := copyTree(link, native, "skills/myskill"); err != nil {
				t.Fatal(err)
			}
			if owned, err := OwnsNativeSkill(fx.home, root, "myskill", &member); owned || err == nil || !strings.Contains(err.Error(), DiagSurfaceDrift) {
				t.Fatalf("unrecorded copied entry admitted: %v %v", owned, err)
			}
			marker := homeMarker(t, native)
			if fallback {
				surface := marker.Surfaces[envmarker.SurfaceSkills]
				copies := []envmarker.Copy{{Path: "skills/myskill", Reason: envmarker.ReasonSymlinkFallback}}
				surface.Copies = &copies
				marker.Surfaces[envmarker.SurfaceSkills] = surface
			} else {
				marker.Mode = envmarker.ModeCopied
				for name, surface := range marker.Surfaces {
					surface.Copies = nil
					marker.Surfaces[name] = surface
				}
			}
			writeNativeOwnerMarker(t, native, marker)
			if owned, err := OwnsNativeSkill(fx.home, root, "myskill", &member); !owned || err != nil {
				t.Fatalf("recorded-copy positive control refused: %v %v", owned, err)
			}
			if err := os.WriteFile(filepath.Join(target, "SKILL.md"), []byte("drifted copy\n"), 0600); err != nil {
				t.Fatal(err)
			}
			before := nativeSnapshot(t, native)
			if owned, err := OwnsNativeSkill(fx.home, root, "myskill", &member); owned || err == nil || !strings.Contains(err.Error(), DiagSurfaceDrift) {
				t.Fatalf("drifted copied entry admitted: %v %v", owned, err)
			}
			if nativeSnapshot(t, native) != before {
				t.Fatal("copied ownership check wrote native state")
			}
		})
	}
}
