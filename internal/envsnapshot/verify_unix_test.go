//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package envsnapshot

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"golang.org/x/sys/unix"
)

func write(t *testing.T, root, p string, b []byte, mode os.FileMode) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(p))
	if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, b, mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(full, mode); err != nil {
		t.Fatal(err)
	}
}
func fixture(t *testing.T) string {
	t.Helper()
	root := temporary(t)
	if err := os.Chmod(root, 0700); err != nil {
		t.Fatal(err)
	}
	m := manifest(t)
	for _, f := range m.Files {
		mode := os.FileMode(0600)
		if f.Mode == "0700" {
			mode = 0700
		}
		write(t, root, f.Path, read(t, "golden/"+f.Path), mode)
	}
	write(t, root, "manifest.json", read(t, "golden/manifest.json"), 0600)
	write(t, root, "launch.json", read(t, "golden/launch.json"), 0600)
	return root
}
func saveManifest(t *testing.T, root string, m Manifest) {
	t.Helper()
	b, err := EncodeManifest(m)
	if err != nil {
		t.Fatal(err)
	}
	write(t, root, "manifest.json", b, 0600)
}
func saveDescriptor(t *testing.T, root string, m Manifest, d Descriptor) {
	t.Helper()
	b, err := EncodeDescriptor(d)
	if err != nil {
		t.Fatal(err)
	}
	write(t, root, "launch.json", b, 0600)
	m.LaunchSHA256 = Digest(b)
	saveManifest(t, root, m)
}
func verified(t *testing.T, root string) *Verified {
	t.Helper()
	v, err := Verify(root, "")
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// A private independent copy verifies; every pinned semantic corpus case
// reaches Verify, including unchanged forged metadata and execute-mode drift.
func TestPinnedTrees(t *testing.T) {
	verified(t, fixture(t))
	var vectors struct {
		Cases []struct {
			Name, Manifest, Launch, Expected string
			Content                          []struct{ Path, Source, Mode string }
		}
	}
	if err := json.Unmarshal(read(t, "vectors.json"), &vectors); err != nil {
		t.Fatal(err)
	}
	if len(vectors.Cases) != 19 {
		t.Fatalf("semantic inventory %d", len(vectors.Cases))
	}
	for _, v := range vectors.Cases {
		t.Run(v.Name, func(t *testing.T) {
			root := temporary(t)
			if err := os.Chmod(root, 0700); err != nil {
				t.Fatal(err)
			}
			// These cases cannot all be represented on case/normalization-folding
			// filesystems (nor can a regular file also be its own parent). Verify must
			// reject the metadata before content enumeration in all of these cases.
			if v.Expected != "path-collision" && v.Expected != "file-ancestor" {
				for _, c := range v.Content {
					mode := os.FileMode(0600)
					if c.Mode == "0700" {
						mode = 0700
					}
					write(t, root, c.Path, read(t, c.Source), mode)
				}
			}
			write(t, root, "manifest.json", read(t, v.Manifest), 0600)
			write(t, root, "launch.json", read(t, v.Launch), 0600)
			before, err := os.ReadFile(filepath.Join(root, "manifest.json"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = Verify(root, "")
			if v.Expected == "accept" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				// The corpus groups mode mismatches with file identity; the public
				// boundary provides a more precise mode refusal under the same code.
				reason := v.Expected
				if v.Name == "execute-mode" {
					reason = "file-mode"
				}
				requireReason(t, err, reason)
			}
			after, e := os.ReadFile(filepath.Join(root, "manifest.json"))
			if e != nil || !bytes.Equal(before, after) {
				t.Fatal("verification rewrote metadata", e)
			}
		})
	}
}

// Links, aliases, FIFOs, unreadable members, unknown directories and wrong
// privacy bits refuse through Verify without modifying the persisted baseline.
func TestActualTreeRefusals(t *testing.T) {
	cases := []struct {
		name, reason string
		mutate       func(*testing.T, string)
	}{
		{"extra-file", "entry-set", func(t *testing.T, r string) { write(t, r, "home/extra", []byte("extra"), 0600) }},
		{"extra-directory", "entry-set", func(t *testing.T, r string) {
			if err := os.Mkdir(filepath.Join(r, "home/empty"), 0700); err != nil {
				t.Fatal(err)
			}
		}},
		{"missing-file", "entry-set", func(t *testing.T, r string) {
			if err := os.Remove(filepath.Join(r, "home/AGENTS.md")); err != nil {
				t.Fatal(err)
			}
		}},
		{"execute-mode", "file-mode", func(t *testing.T, r string) {
			if err := os.Chmod(filepath.Join(r, "home/skills/sample/run.sh"), 0600); err != nil {
				t.Fatal(err)
			}
		}},
		{"private-mode", "file-mode", func(t *testing.T, r string) {
			if err := os.Chmod(filepath.Join(r, "home/AGENTS.md"), 0644); err != nil {
				t.Fatal(err)
			}
		}},
		{"symlink", "file-kind", func(t *testing.T, r string) {
			p := filepath.Join(r, "home/AGENTS.md")
			if err := os.Remove(p); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink("curator-mcp.config.toml", p); err != nil {
				t.Fatal(err)
			}
		}},
		{"hardlink-outside", "file-kind", func(t *testing.T, r string) {
			if err := os.Link(filepath.Join(r, "home/AGENTS.md"), filepath.Join(temporary(t), "alias")); err != nil {
				t.Fatal(err)
			}
		}},
		{"hardlink-inside", "file-kind", func(t *testing.T, r string) {
			p := filepath.Join(r, "home/curator-mcp.config.toml")
			if err := os.Remove(p); err != nil {
				t.Fatal(err)
			}
			if err := os.Link(filepath.Join(r, "home/AGENTS.md"), p); err != nil {
				t.Fatal(err)
			}
		}},
		{"fifo", "file-kind", func(t *testing.T, r string) {
			p := filepath.Join(r, "home/AGENTS.md")
			if err := os.Remove(p); err != nil {
				t.Fatal(err)
			}
			if err := unix.Mkfifo(p, 0600); err != nil {
				t.Fatal(err)
			}
		}},
		{"unreadable", "file-unreadable", func(t *testing.T, r string) {
			if err := os.Chmod(filepath.Join(r, "home/AGENTS.md"), 0000); err != nil {
				t.Fatal(err)
			}
		}},
		{"unreadable-directory", "tree-unreadable", func(t *testing.T, r string) {
			p := filepath.Join(r, "home/skills")
			if err := os.Chmod(p, 0000); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(p, 0700) })
		}},
		{"linked-directory", "directory-kind", func(t *testing.T, r string) {
			p := filepath.Join(r, "home/skills")
			saved := filepath.Join(temporary(t), "saved-skills")
			if err := os.Rename(p, saved); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(saved, p); err != nil {
				t.Fatal(err)
			}
		}},
		{"metadata-alias", "metadata-kind-mode", func(t *testing.T, r string) {
			if err := os.Link(filepath.Join(r, "manifest.json"), filepath.Join(temporary(t), "manifest-copy")); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := fixture(t)
			v := verified(t, r)
			baseline := v.ManifestBytes()
			c.mutate(t, r)
			_, err := Verify(r, v.ManifestSHA256())
			requireReason(t, err, c.reason)
			after, e := os.ReadFile(filepath.Join(r, "manifest.json"))
			if e != nil || !bytes.Equal(after, baseline) {
				t.Fatal("baseline changed", e)
			}
		})
	}
}

// Exact identity retains source provenance while portable equality excludes
// only that field; version, profile pin, descriptor and content still bind.
func TestPortableIdentity(t *testing.T) {
	a, b := fixture(t), fixture(t)
	av := verified(t, a)
	m := manifest(t)
	m.FragmentSHA256 = Digest([]byte("independent-source-home"))
	saveManifest(t, b, m)
	bv := verified(t, b)
	if av.ManifestSHA256() == bv.ManifestSHA256() {
		t.Fatal("source provenance lost")
	}
	equal, err := PortableEqual(a, b)
	if err != nil || !equal {
		t.Fatal(equal, err)
	}
	// Caller mutation of returned bytes cannot rewrite an admitted baseline.
	exposed := av.ManifestBytes()
	exposed[0] = 'x'
	if !bytes.Equal(av.ManifestBytes(), read(t, "golden/manifest.json")) {
		t.Fatal("mutable baseline")
	}
	for _, name := range []string{"version", "profile-name", "profile-lock", "descriptor", "file", "mode"} {
		t.Run(name, func(t *testing.T) {
			b := fixture(t)
			m := manifest(t)
			d := descriptor(t)
			switch name {
			case "version":
				m.CuratorVersion = "1.0.1"
				saveManifest(t, b, m)
			case "profile-name":
				m.Profile.Name = "other"
				d.Profile = m.Profile
				saveDescriptor(t, b, m, d)
			case "profile-lock":
				m.Profile.LockSHA256 = Digest([]byte("other-lock"))
				d.Profile = m.Profile
				saveDescriptor(t, b, m, d)
			case "descriptor":
				d.Precedence.Winner = "lower-weight"
				saveDescriptor(t, b, m, d)
			case "file":
				for i := range m.Files {
					if m.Files[i].Path == "home/AGENTS.md" {
						data := []byte("replacement")
						m.Files[i].Size = int64(len(data))
						m.Files[i].SHA256 = Digest(data)
						write(t, b, m.Files[i].Path, data, 0600)
					}
				}
				saveManifest(t, b, m)
			case "mode":
				for i := range m.Files {
					if m.Files[i].Path == "home/skills/sample/run.sh" {
						m.Files[i].Mode = "0600"
						if err := os.Chmod(filepath.Join(b, m.Files[i].Path), 0600); err != nil {
							t.Fatal(err)
						}
					}
				}
				saveManifest(t, b, m)
			}
			verified(t, b)
			eq, err := PortableEqual(a, b)
			if err != nil || eq {
				t.Fatal("binding omitted", eq, err)
			}
		})
	}
	// Forged unchanged metadata does not turn modified bytes into portable parity.
	write(t, b, "home/AGENTS.md", []byte("forged"), 0600)
	_, err = PortableEqual(a, b)
	requireReason(t, err, "file-identity")
}

// Recapturing a self-consistent new manifest cannot replace the original exact
// baseline. Repeated refusals preserve both the baseline and existing target.
func TestBaselineNeverRecaptured(t *testing.T) {
	r := fixture(t)
	v := verified(t, r)
	baseline := v.ManifestBytes()
	digest := v.ManifestSHA256()
	write(t, r, "home/AGENTS.md", []byte("changed"), 0600)
	_, err := Verify(r, digest)
	requireReason(t, err, "file-identity")
	m := manifest(t)
	for i := range m.Files {
		if m.Files[i].Path == "home/AGENTS.md" {
			m.Files[i].Size = 7
			m.Files[i].SHA256 = Digest([]byte("changed"))
		}
	}
	saveManifest(t, r, m)
	observed := verified(t, r)
	if observed.ManifestSHA256() == digest {
		t.Fatal("exact identity unchanged")
	}
	for i := 0; i < 2; i++ {
		_, err = Verify(r, digest)
		requireReason(t, err, "baseline-mismatch")
	}
	after, err := os.ReadFile(filepath.Join(r, "manifest.json"))
	if err != nil || !bytes.Equal(after, observed.ManifestBytes()) || !bytes.Equal(v.ManifestBytes(), baseline) {
		t.Fatal("baseline/target rewritten", err)
	}
}

// Muse's four layout parents are derived even without files. Root-context is
// an explicit fact, never inferred by a consuming launcher or omitted on drift.
func TestMuseDerivedDirectoriesAndRootFact(t *testing.T) {
	r := temporary(t)
	if err := os.Chmod(r, 0700); err != nil {
		t.Fatal(err)
	}
	d, err := DecodeDescriptor(canonical(t, read(t, "cases/snapshot-launch-valid-muse.json")))
	if err != nil {
		t.Fatal(err)
	}
	d.RootContext = nil
	m := manifest(t)
	m.Adapter = "muse"
	m.Files = []File{}
	for _, p := range []string{"home/config", "home/data", "home/state", "home/cache"} {
		if err := os.MkdirAll(filepath.Join(r, p), 0700); err != nil {
			t.Fatal(err)
		}
	}
	saveDescriptor(t, r, m, d)
	verified(t, r)
	if err := os.Remove(filepath.Join(r, "home/cache")); err != nil {
		t.Fatal(err)
	}
	_, err = Verify(r, "")
	requireReason(t, err, "entry-set")
	if err := os.Mkdir(filepath.Join(r, "home/cache"), 0700); err != nil {
		t.Fatal(err)
	}
	d.RootContext = &RootContext{Exclusive: true}
	saveDescriptor(t, r, m, d)
	_, err = Verify(r, "")
	requireReason(t, err, "root-context")
	data := []byte("root\n")
	write(t, r, "home/config/muse/AGENTS.md", data, 0600)
	m.Files = []File{{Path: "home/config/muse/AGENTS.md", Size: int64(len(data)), SHA256: Digest(data), Mode: "0600"}}
	saveDescriptor(t, r, m, d)
	verified(t, r)
	d.RootContext = nil
	saveDescriptor(t, r, m, d)
	_, err = Verify(r, "")
	requireReason(t, err, "root-context")
}

// Additional collision and reference neighbors exercise Unicode casefold,
// component-level aliases, sortedness and registry channel equality.
func TestMetadataNeighbors(t *testing.T) {
	for _, tc := range []struct {
		name, reason string
		change       func(*Manifest, *Descriptor)
	}{
		{"sharp-s-casefold", "path-collision", func(m *Manifest, _ *Descriptor) {
			f := m.Files[0]
			f.Path = "home/SS/a"
			g := f
			g.Path = "home/ß/b"
			m.Files = []File{f, g}
		}},
		{"NFD-components", "path-collision", func(m *Manifest, _ *Descriptor) {
			f := m.Files[0]
			f.Path = "home/café/a"
			g := f
			g.Path = "home/cafe\u0301/b"
			m.Files = []File{f, g}
			sort.Slice(m.Files, func(i, j int) bool { return m.Files[i].Path < m.Files[j].Path })
		}},
		{"wrong-fixed-layer", "channel-reference", func(_ *Manifest, d *Descriptor) { d.MCP.Path = "home/AGENTS.md" }},
		{"reserved-env", "launch-schema", func(_ *Manifest, d *Descriptor) { d.MCP.EnvNames = []string{"PATH"} }},
		{"unsorted-env", "channel-semantics", func(_ *Manifest, d *Descriptor) { d.MCP.EnvNames = []string{"Z", "A"} }},
		{"duplicate-env", "launch-schema", func(_ *Manifest, d *Descriptor) { d.MCP.EnvNames = []string{"A", "A"} }},
		{"system-empty", "channel-semantics", func(_ *Manifest, d *Descriptor) { d.SystemPrompt.Channels = []Channel{} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := fixture(t)
			verified(t, r)
			m, d := manifest(t), descriptor(t)
			tc.change(&m, &d)
			raw, e := json.Marshal(d)
			if e != nil {
				t.Fatal(e)
			}
			wire := canonical(t, raw)
			write(t, r, "launch.json", wire, 0600)
			m.LaunchSHA256 = Digest(wire)
			saveManifest(t, r, m)
			_, err := Verify(r, "")
			requireReason(t, err, tc.reason)
		})
	}
}

// Resolve only the test framework's own allocation (e.g. a platform /var alias).
// Production Verify never resolves caller-supplied linked ancestors.
func temporary(t *testing.T) string {
	t.Helper()
	p, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// A digest over launch JSON without its framing LF is not the launch identity.
func TestLaunchHashIncludesLF(t *testing.T) {
	r := fixture(t)
	verified(t, r)
	m := manifest(t)
	raw := read(t, "golden/launch.json")
	m.LaunchSHA256 = Digest(raw[:len(raw)-1])
	saveManifest(t, r, m)
	_, err := Verify(r, "")
	requireReason(t, err, "launch-digest")
}

// Linked target and linked ancestors are refused even when they name an intact
// snapshot. This test uses explicit links, not the platform's temp aliases.
func TestNoFollowSnapshotRoot(t *testing.T) {
	r := fixture(t)
	verified(t, r)
	parent := temporary(t)
	link := filepath.Join(parent, "linked")
	if err := os.Symlink(r, link); err != nil {
		t.Fatal(err)
	}
	_, err := Verify(link, "")
	requireReason(t, err, "tree-unreadable")
	container := temporary(t)
	target := filepath.Join(container, "snapshot")
	if err := os.Rename(r, target); err != nil {
		t.Fatal(err)
	}
	ancestor := filepath.Join(parent, "ancestor")
	if err := os.Symlink(container, ancestor); err != nil {
		t.Fatal(err)
	}
	_, err = Verify(filepath.Join(ancestor, "snapshot"), "")
	requireReason(t, err, "tree-unreadable")
}

// A link followed by '..' remains a linked ancestor; lexical cleaning must not
// erase it before the no-follow walk sees it.
func TestNoFollowBeforeParentSegment(t *testing.T) {
	parent := temporary(t)
	r := filepath.Join(parent, "snapshot")
	if err := os.Rename(fixture(t), r); err != nil {
		t.Fatal(err)
	}
	verified(t, r)
	if err := os.Symlink(parent, filepath.Join(parent, "linked")); err != nil {
		t.Fatal(err)
	}
	_, err := Verify(parent+"/linked/../snapshot", "")
	requireReason(t, err, "tree-unreadable")
}

// Core portable-path semantics forbid trailing dot/space in EVERY component;
// the snapshot schema's whole-path regex alone does not establish this.
func TestPortableComponentSuffixes(t *testing.T) {
	for _, suffix := range []string{".", " "} {
		t.Run(suffix, func(t *testing.T) {
			r := fixture(t)
			verified(t, r)
			m := manifest(t)
			m.Files[0].Path = "home/invalid" + suffix + "/asset"
			sort.Slice(m.Files, func(i, j int) bool { return m.Files[i].Path < m.Files[j].Path })
			saveManifest(t, r, m)
			_, err := Verify(r, "")
			requireReason(t, err, "path-invalid")
		})
	}
}

// Pi's fixed-file descriptors advertise alternative consumption channels, not
// one required filename for its inert system_prompt.path. All three registered
// descriptors must be retained without imposing contradictory file targets.
func TestPiAlternativeChannels(t *testing.T) {
	r := fixture(t)
	m := manifest(t)
	d := descriptor(t)
	m.Adapter = "pi"
	d.Environment = "pi"
	d.Env = map[string]string{"PI_CODING_AGENT_DIR": "home"}
	d.MCP = nil
	d.SystemPrompt.Channels = []Channel{{Kind: "flag", Semantics: "append", Flag: "--append-system-prompt", Argument: "path"}, {Kind: "file", Semantics: "append", Filename: "APPEND_SYSTEM.md"}, {Kind: "file", Semantics: "replace", Filename: "SYSTEM.md"}}
	saveDescriptor(t, r, m, d)
	verified(t, r)
	d.SystemPrompt.Channels[1].Filename = "UNREGISTERED.md"
	saveDescriptor(t, r, m, d)
	_, err := Verify(r, "")
	requireReason(t, err, "channel-semantics")
}
