//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package envsnapshot

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// Optional nulls must not erase a present invalid channel through typed decoding.
func TestReviewerNullChannel(t *testing.T) {
	r := fixture(t)
	verified(t, r)
	var d map[string]any
	if err := json.Unmarshal(read(t, "golden/launch.json"), &d); err != nil {
		t.Fatal(err)
	}
	d["mcp"] = nil
	raw, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	wire := canonical(t, raw)
	write(t, r, "launch.json", wire, 0600)
	m := manifest(t)
	m.LaunchSHA256 = Digest(wire)
	saveManifest(t, r, m)
	_, err = Verify(r, "")
	requireReason(t, err, "launch-schema")
}

// A metadata-only directory cannot pass either public tree-verification path.
func TestReviewerManifestOnly(t *testing.T) {
	r := fixture(t)
	verified(t, r)
	if err := os.RemoveAll(filepath.Join(r, "home")); err != nil {
		t.Fatal(err)
	}
	_, err := Verify(r, "")
	requireReason(t, err, "entry-set")
	_, err = PortableEqual(r, r)
	requireReason(t, err, "entry-set")
}

// Both sides must be re-observed; identical saved metadata cannot hide corrupt left bytes.
func TestReviewerPortableLeftDrift(t *testing.T) {
	a, b := fixture(t), fixture(t)
	eq, err := PortableEqual(a, b)
	if err != nil || !eq {
		t.Fatal(eq, err)
	}
	write(t, a, "home/AGENTS.md", []byte("forged"), 0600)
	_, err = PortableEqual(a, b)
	requireReason(t, err, "file-identity")
}

// Recapture does not change an original externally retained baseline on either retry.
func TestReviewerRecapture(t *testing.T) {
	r := fixture(t)
	v := verified(t, r)
	m := manifest(t)
	m.CuratorVersion = "9.9.9"
	saveManifest(t, r, m)
	for i := 0; i < 2; i++ {
		_, err := Verify(r, v.ManifestSHA256())
		requireReason(t, err, "baseline-mismatch")
	}
}

// Metadata aliases and special mode bits cannot be laundered by valid hashes.
func TestReviewerFreeHunt(t *testing.T) {
	t.Run("launch-hardlink", func(t *testing.T) {
		r := fixture(t)
		v := verified(t, r)
		if err := os.Link(filepath.Join(r, "launch.json"), filepath.Join(temporary(t), "alias")); err != nil {
			t.Fatal(err)
		}
		_, err := Verify(r, v.ManifestSHA256())
		requireReason(t, err, "metadata-kind-mode")
	})
	t.Run("sticky-content", func(t *testing.T) {
		r := fixture(t)
		v := verified(t, r)
		if err := os.Chmod(filepath.Join(r, "home/AGENTS.md"), 0600|os.ModeSticky); err != nil {
			t.Fatal(err)
		}
		_, err := Verify(r, v.ManifestSHA256())
		requireReason(t, err, "file-mode")
	})
	t.Run("missing-metadata", func(t *testing.T) {
		r := fixture(t)
		v := verified(t, r)
		if err := os.Remove(filepath.Join(r, "launch.json")); err != nil {
			t.Fatal(err)
		}
		_, err := Verify(r, v.ManifestSHA256())
		requireReason(t, err, "metadata-unreadable")
	})
}

// Portable non-ASCII spellings and distinct non-colliding components remain valid.
func TestReviewerUnicodePositive(t *testing.T) {
	for _, p := range []string{"home/café/asset", "home/cafe\u0301/asset", "home/日本語/asset", "home/😀/asset"} {
		t.Run(p, func(t *testing.T) {
			r := fixture(t)
			m := manifest(t)
			data := []byte("unicode")
			write(t, r, p, data, 0600)
			m.Files = append(m.Files, File{Path: p, Size: int64(len(data)), SHA256: Digest(data), Mode: "0600"})
			sort.Slice(m.Files, func(i, j int) bool { return m.Files[i].Path < m.Files[j].Path })
			saveManifest(t, r, m)
			verified(t, r)
		})
	}
}

// Canonically caseless-equivalent Greek components must refuse through Verify.
// Each singleton is a positive control; two distinct entries are representable
// on a case-sensitive filesystem. The original manifest must remain unchanged.
func TestReviewerGreekCollision(t *testing.T) {
	for _, tc := range []struct{ name, a, b string }{{"iota", "ι\u0308\u0301", "Ι\u0308\u0301"}, {"upsilon", "υ\u0308\u0301", "Υ\u0308\u0301"}} {
		for _, kind := range []string{"directory", "file"} {
			t.Run(tc.name+"/"+kind, func(t *testing.T) {
				paths := []string{"home/" + tc.a, "home/" + tc.b}
				if kind == "directory" {
					paths[0] += "/a"
					paths[1] += "/b"
				}
				makeTree := func(ps []string) string {
					r := fixture(t)
					m := manifest(t)
					data := []byte("collision")
					for _, p := range ps {
						write(t, r, p, data, 0600)
						m.Files = append(m.Files, File{Path: p, Size: int64(len(data)), SHA256: Digest(data), Mode: "0600"})
					}
					sort.Slice(m.Files, func(i, j int) bool { return m.Files[i].Path < m.Files[j].Path })
					saveManifest(t, r, m)
					return r
				}
				for _, p := range paths {
					verified(t, makeTree([]string{p}))
				}
				r := makeTree(paths)
				before, err := os.ReadFile(filepath.Join(r, "manifest.json"))
				if err != nil {
					t.Fatal(err)
				}
				_, err = Verify(r, "")
				after, e := os.ReadFile(filepath.Join(r, "manifest.json"))
				if e != nil || string(before) != string(after) {
					t.Fatal("baseline rewritten", e)
				}
				requireReason(t, err, "path-collision")
			})
		}
	}
}
