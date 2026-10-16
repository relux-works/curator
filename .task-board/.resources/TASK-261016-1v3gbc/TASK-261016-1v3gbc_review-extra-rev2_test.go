//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package envsnapshot

import ("os"; "path/filepath"; "sort"; "testing")

// Five neighboring full-fold classes refuse file and component collisions; singleton controls verify and manifest bytes remain unchanged.
func TestReviewAdditionalCaselessCollisions(t *testing.T) {
	for _, tc := range []struct{ name, a, b string }{{"sharp-s", "straße", "STRASSE"}, {"sigma", "ς", "σ"}, {"ligature", "ﬀ", "FF"}, {"long-s", "ſ", "S"}, {"kelvin", "K", "k"}} {
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
