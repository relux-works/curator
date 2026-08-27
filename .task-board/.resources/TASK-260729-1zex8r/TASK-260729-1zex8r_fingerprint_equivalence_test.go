package godriver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"go/build"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"
)

// legacyRecord and legacyFingerprintToolchain are a verbatim preservation of the
// traversal fingerprintToolchain used before the directory-scoped rework: one
// os.Root over all of GOROOT, fs.WalkDir over root.FS(), a full-path Lstat per
// entry and a full-path Open per file. They exist only so every test below can
// assert that the shipped implementation is byte-for-byte indistinguishable
// from it — same canonical records, same digest, same diagnostic code and same
// operator detail — instead of merely re-asserting the shipped behaviour
// against itself.
type legacyRecord struct {
	path       string
	filesystem string
	kind       byte
	link       string
	info       fs.FileInfo
}

func legacyFingerprintToolchain(ctx context.Context, rootPath, version string) (string, []toolchainState, error) {
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return "", nil, diagnosticErr("toolchain_unreadable", err, "cannot open GOROOT")
	}
	defer func() { _ = root.Close() }()

	records, err := legacyCollectRecords(ctx, root)
	if err != nil {
		return "", nil, err
	}
	return legacyDigestRecords(ctx, root, records, version)
}

// legacyCollectRecords and legacyDigestRecords split the preserved traversal at
// the same boundary the shipped one is split at, so a mutation applied between
// the two phases can be compared across both implementations.
func legacyCollectRecords(ctx context.Context, root *os.Root) ([]legacyRecord, error) {
	var records []legacyRecord
	encoded := make(map[string]struct{})
	err := fs.WalkDir(root.FS(), ".", func(name string, _ fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return diagnosticErr("toolchain_unreadable", walkErr, "cannot walk GOROOT")
		}
		if err := ctx.Err(); err != nil {
			return diagnosticErr("toolchain_timeout", err, "toolchain fingerprint deadline exceeded")
		}
		if name == "." {
			return nil
		}
		protocolPath := filepath.ToSlash(name)
		if !validToolchainPath(protocolPath) {
			return diagnostic("invalid_unicode", "toolchain path is not valid protocol UTF-8")
		}
		if err := claimEncodedPath(encoded, protocolPath); err != nil {
			return err
		}

		info, statErr := root.Lstat(name)
		if statErr != nil {
			return diagnosticErr("toolchain_unreadable", statErr, "cannot inspect toolchain path %q", protocolPath)
		}
		record := legacyRecord{path: protocolPath, filesystem: name, info: info}
		switch {
		case info.Mode().IsDir():
			record.kind = 'D'
		case info.Mode().IsRegular():
			record.kind = 'F'
		case info.Mode()&fs.ModeSymlink != 0:
			record.kind = 'L'
			target, readErr := root.Readlink(name)
			if readErr != nil {
				return diagnosticErr("toolchain_link_dangling", readErr, "cannot read toolchain link %q", protocolPath)
			}
			if !utf8.ValidString(target) || strings.ContainsRune(target, 0) {
				return diagnostic("invalid_unicode", "toolchain link %q has an invalid target", protocolPath)
			}
			if filepath.IsAbs(target) || filepath.VolumeName(target) != "" ||
				strings.HasPrefix(target, "/") || strings.HasPrefix(target, `\`) {
				return diagnostic("toolchain_link_absolute", "toolchain link %q is absolute", protocolPath)
			}
			resolvedName := filepath.Clean(filepath.Join(filepath.Dir(name), target))
			if resolvedName == ".." || strings.HasPrefix(resolvedName, ".."+string(filepath.Separator)) || filepath.IsAbs(resolvedName) {
				return diagnostic("toolchain_link_escape", "toolchain link %q escapes GOROOT", protocolPath)
			}
			if _, statErr := root.Stat(name); statErr != nil {
				return diagnosticErr("toolchain_link_dangling", statErr, "toolchain link %q does not resolve safely", protocolPath)
			}
			record.link = target
		default:
			return diagnostic("special_file_forbidden", "toolchain path %q is not a directory, regular file, or link", protocolPath)
		}
		records = append(records, record)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(records, func(left, right int) bool { return records[left].path < records[right].path })
	return records, nil
}

func legacyDigestRecords(ctx context.Context, root *os.Root, records []legacyRecord, version string) (string, []toolchainState, error) {
	digest := sha256.New()
	_, _ = io.WriteString(digest, toolchainDomain)
	state := make([]toolchainState, 0, len(records))
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			return "", nil, diagnosticErr("toolchain_timeout", err, "toolchain fingerprint deadline exceeded")
		}
		pathBytes := []byte(record.path)
		writeRecordHeader(digest, record.kind, pathBytes)
		item := toolchainState{path: record.path, kind: record.kind}
		switch record.kind {
		case 'D':
			writeLength(digest, 0)
		case 'L':
			payload := []byte(record.link)
			writeLength(digest, uint64(len(payload)))
			_, _ = digest.Write(payload)
			item.size = int64(len(payload))
			item.payload = sha256.Sum256(payload)
		case 'F':
			file, openErr := root.Open(record.filesystem)
			if openErr != nil {
				return "", nil, diagnosticErr("toolchain_unreadable", openErr, "cannot open toolchain file %q", record.path)
			}
			opened, statErr := file.Stat()
			if statErr != nil || !opened.Mode().IsRegular() || !os.SameFile(record.info, opened) {
				_ = file.Close()
				return "", nil, diagnosticErr("toolchain_mutated", statErr, "toolchain file %q changed while opening", record.path)
			}
			if opened.Size() < 0 {
				return "", nil, diagnostic("toolchain_mutated", "toolchain file %q reports a negative size", record.path)
			}
			writeLength(digest, uint64(opened.Size())) // #nosec G115 -- the size was just proved non-negative
			contentDigest := sha256.New()
			written, copyErr := copyWithContext(ctx, io.MultiWriter(digest, contentDigest), file)
			closeErr := file.Close()
			if copyErr != nil || closeErr != nil || written != opened.Size() {
				var lengthErr error
				if written != opened.Size() {
					lengthErr = errors.New("file length changed")
				}
				return "", nil, diagnosticErr("toolchain_mutated", errors.Join(copyErr, closeErr, lengthErr), "toolchain file %q changed while reading", record.path)
			}
			item.size = written
			copy(item.payload[:], contentDigest.Sum(nil))
		}
		state = append(state, item)
	}
	writeRecordHeader(digest, 'V', nil)
	writeLength(digest, uint64(len([]byte(version))))
	_, _ = io.WriteString(digest, version)
	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), state, nil
}

const equivalenceVersion = "go version go1.25.5 darwin/arm64"

// assertEquivalent runs both traversals over the same tree and requires that
// they are indistinguishable: identical digest, identical canonical record set
// in identical order, and on failure an identical diagnostic code and detail.
func assertEquivalent(t *testing.T, root string) (string, []toolchainState) {
	t.Helper()
	legacyDigest, legacyState, legacyErr := legacyFingerprintToolchain(context.Background(), root, equivalenceVersion)
	scopedDigest, scopedState, scopedErr := fingerprintToolchain(context.Background(), root, equivalenceVersion)

	switch {
	case legacyErr == nil && scopedErr != nil:
		t.Fatalf("legacy succeeded, scoped failed: %v", scopedErr)
	case legacyErr != nil && scopedErr == nil:
		t.Fatalf("legacy failed with %v, scoped succeeded", legacyErr)
	case legacyErr != nil:
		if DiagnosticCode(legacyErr) != DiagnosticCode(scopedErr) {
			t.Fatalf("diagnostic code: legacy %q, scoped %q (%v / %v)",
				DiagnosticCode(legacyErr), DiagnosticCode(scopedErr), legacyErr, scopedErr)
		}
		if legacyErr.Error() != scopedErr.Error() {
			t.Fatalf("diagnostic detail:\n legacy %q\n scoped %q", legacyErr.Error(), scopedErr.Error())
		}
		return "", nil
	}

	if legacyDigest != scopedDigest {
		t.Fatalf("digest: legacy %s, scoped %s", legacyDigest, scopedDigest)
	}
	if len(legacyState) != len(scopedState) {
		t.Fatalf("record count: legacy %d, scoped %d", len(legacyState), len(scopedState))
	}
	for index := range legacyState {
		if legacyState[index] != scopedState[index] {
			t.Fatalf("record %d:\n legacy %+v\n scoped %+v", index, legacyState[index], scopedState[index])
		}
	}
	return scopedDigest, scopedState
}

func writeTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, payload := range files {
		writeTestFile(t, filepath.Join(root, filepath.FromSlash(name)), []byte(payload), 0o644)
	}
}

func symlinkOrSkip(t *testing.T, target, name string) {
	t.Helper()
	if err := os.Symlink(target, name); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
}

// TestFingerprintTraversalMatchesLegacyOnRepresentativeTrees is the primary
// equivalence gate: every shape the canonical encoding can take must hash the
// same through both traversals.
func TestFingerprintTraversalMatchesLegacyOnRepresentativeTrees(t *testing.T) {
	cases := map[string]func(t *testing.T, root string){
		"empty root": func(*testing.T, string) {},
		"flat files": func(t *testing.T, root string) {
			writeTree(t, root, map[string]string{"a": "1", "b": "22", "c": ""})
		},
		"nested toolchain shape": func(t *testing.T, root string) {
			writeTree(t, root, map[string]string{
				"bin/go":                     "GO",
				"bin/gofmt":                  "FMT",
				"pkg/tool/darwin_arm64/link": "LINK",
				"src/runtime/proc.go":        strings.Repeat("x", 4096),
				"VERSION":                    "go1.25.5",
			})
		},
		// A directory and its own siblings interleave in sorted order whenever a
		// sibling name starts with a byte below '/'. This is the case a naive
		// one-slot directory cache would thrash on and a stack must not reorder.
		"siblings sorting around the separator": func(t *testing.T, root string) {
			writeTree(t, root, map[string]string{
				"b/c":   "in dir",
				"b-x":   "dash sorts before slash",
				"b.y":   "dot sorts before slash",
				"b!z":   "bang sorts before slash",
				"b0":    "digit sorts after slash",
				"bb":    "letter sorts after slash",
				"b/a/d": "deep",
			})
		},
		"deeply nested chain": func(t *testing.T, root string) {
			writeTree(t, root, map[string]string{
				"a/b/c/d/e/f/g/h/i/j/leaf": "deep",
				"a/b/c/d/e/f/g/sibling":    "mid",
				"a/top":                    "top",
			})
		},
		"empty directories": func(t *testing.T, root string) {
			for _, dir := range []string{"empty", "outer/inner-empty", "outer/kept"} {
				if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			writeTree(t, root, map[string]string{"outer/kept/file": "x"})
		},
		"unicode and space names": func(t *testing.T, root string) {
			writeTree(t, root, map[string]string{
				"пакет/файл.go":   "unicode",
				"with space/file": "space",
				"emoji-🙂/leaf":    "emoji",
			})
		},
		"large file crossing the copy buffer": func(t *testing.T, root string) {
			writeTree(t, root, map[string]string{
				"big": strings.Repeat("payload", 60000),
				"pkg/also-big": strings.Repeat("q", 128*1024) +
					strings.Repeat("r", 7),
			})
		},
		"relative links inside the same directory": func(t *testing.T, root string) {
			writeTree(t, root, map[string]string{"bin/go": "GO"})
			symlinkOrSkip(t, "go", filepath.Join(root, "bin", "go-alias"))
		},
		// The RC4 vector shape: a link whose target leaves its own directory but
		// stays inside GOROOT. Validation must stay anchored at the root.
		"link escaping its directory but not the root": func(t *testing.T, root string) {
			writeTree(t, root, map[string]string{"bin/go": "GO"})
			if err := os.MkdirAll(filepath.Join(root, "pkg"), 0o755); err != nil {
				t.Fatal(err)
			}
			symlinkOrSkip(t, "../bin/go", filepath.Join(root, "pkg", "tool-link"))
		},
		"link to a directory is recorded, not descended": func(t *testing.T, root string) {
			writeTree(t, root, map[string]string{"real/inner": "hidden"})
			symlinkOrSkip(t, "real", filepath.Join(root, "alias"))
		},
		"link to a link": func(t *testing.T, root string) {
			writeTree(t, root, map[string]string{"bin/go": "GO"})
			symlinkOrSkip(t, "go", filepath.Join(root, "bin", "first"))
			symlinkOrSkip(t, "first", filepath.Join(root, "bin", "second"))
		},
	}

	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			build(t, root)
			assertEquivalent(t, root)
		})
	}
}

// TestFingerprintTraversalMatchesLegacyOnFailClosedTrees pins the adversarial
// half: a rejected tree must be rejected identically, with the same stable code
// and the same operator detail.
func TestFingerprintTraversalMatchesLegacyOnFailClosedTrees(t *testing.T) {
	cases := map[string]struct {
		build func(t *testing.T, root string)
		code  string
	}{
		"absolute link": {
			build: func(t *testing.T, root string) {
				symlinkOrSkip(t, filepath.Join(string(filepath.Separator), "outside"), filepath.Join(root, "link"))
			},
			code: "toolchain_link_absolute",
		},
		"escaping link": {
			build: func(t *testing.T, root string) {
				if err := os.MkdirAll(filepath.Join(root, "pkg"), 0o755); err != nil {
					t.Fatal(err)
				}
				symlinkOrSkip(t, "../../outside", filepath.Join(root, "pkg", "link"))
			},
			code: "toolchain_link_escape",
		},
		"dangling link": {
			build: func(t *testing.T, root string) {
				symlinkOrSkip(t, "missing", filepath.Join(root, "link"))
			},
			code: "toolchain_link_dangling",
		},
		"dangling link deep in the tree": {
			build: func(t *testing.T, root string) {
				writeTree(t, root, map[string]string{"a/b/c/file": "x"})
				symlinkOrSkip(t, "missing", filepath.Join(root, "a", "b", "c", "link"))
			},
			code: "toolchain_link_dangling",
		},
		"link cycle": {
			build: func(t *testing.T, root string) {
				symlinkOrSkip(t, "second", filepath.Join(root, "first"))
				symlinkOrSkip(t, "first", filepath.Join(root, "second"))
			},
			code: "toolchain_link_dangling",
		},
	}

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			testCase.build(t, root)
			assertEquivalent(t, root)
			_, _, err := fingerprintToolchain(context.Background(), root, equivalenceVersion)
			if DiagnosticCode(err) != testCase.code {
				t.Fatalf("code = %q (%v), want %q", DiagnosticCode(err), err, testCase.code)
			}
		})
	}
}

// TestFingerprintTraversalMatchesLegacyOnErrorPrecedence guards the property
// that makes the scoped walk substitutable: when a tree carries more than one
// violation, both traversals must reject on the same one. Precedence follows
// name-sorted, pre-order descent, so it only holds while the scoped walk
// reproduces os.Root.FS()'s ordering.
func TestFingerprintTraversalMatchesLegacyOnErrorPrecedence(t *testing.T) {
	cases := map[string]func(t *testing.T, root string){
		"sibling violations in one directory": func(t *testing.T, root string) {
			symlinkOrSkip(t, "/absolute", filepath.Join(root, "b-absolute"))
			symlinkOrSkip(t, "missing", filepath.Join(root, "a-dangling"))
		},
		"shallow violation outranks a deeper one": func(t *testing.T, root string) {
			writeTree(t, root, map[string]string{"a/keep": "x"})
			symlinkOrSkip(t, "missing", filepath.Join(root, "a", "b-dangling"))
			symlinkOrSkip(t, "/absolute", filepath.Join(root, "z-absolute"))
		},
		"deeper violation reached before a later sibling": func(t *testing.T, root string) {
			writeTree(t, root, map[string]string{"a/keep": "x"})
			symlinkOrSkip(t, "/absolute", filepath.Join(root, "a", "z-absolute"))
			symlinkOrSkip(t, "missing", filepath.Join(root, "b-dangling"))
		},
		"violation inside a subtree that sorts before its siblings": func(t *testing.T, root string) {
			writeTree(t, root, map[string]string{"b/inner/keep": "x", "b-x": "sorts after b/"})
			symlinkOrSkip(t, "missing", filepath.Join(root, "b", "inner", "dangling"))
		},
	}

	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			build(t, root)
			assertEquivalent(t, root)
			if _, _, err := fingerprintToolchain(context.Background(), root, equivalenceVersion); err == nil {
				t.Fatal("tree with violations was accepted")
			}
		})
	}
}

// TestFingerprintTraversalMatchesLegacyOnRealToolchain is the representative
// case the optimisation exists for: the host GOROOT, with its real depth, file
// count, and symlinks.
func TestFingerprintTraversalMatchesLegacyOnRealToolchain(t *testing.T) {
	if testing.Short() {
		t.Skip("hashes the whole host GOROOT")
	}
	goroot := build.Default.GOROOT
	if goroot == "" {
		t.Skip("no GOROOT available")
	}
	if _, err := os.Stat(goroot); err != nil {
		t.Skipf("GOROOT %s unreadable: %v", goroot, err)
	}
	digest, state := assertEquivalent(t, goroot)
	if len(state) == 0 {
		t.Fatal("real GOROOT produced no records")
	}
	t.Logf("GOROOT %s: %d records, digest %s", goroot, len(state), digest)
}

// TestFingerprintIsStableAcrossRepeatedRuns rules out state leaking between
// calls through the reused directory handles.
func TestFingerprintIsStableAcrossRepeatedRuns(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"bin/go": "GO", "b-x": "sibling", "b/c/d": "deep", "pkg/tool/link": "L",
	})
	first, _, err := fingerprintToolchain(context.Background(), root, equivalenceVersion)
	if err != nil {
		t.Fatal(err)
	}
	for round := range 3 {
		again, _, repeatErr := fingerprintToolchain(context.Background(), root, equivalenceVersion)
		if repeatErr != nil {
			t.Fatalf("round %d: %v", round, repeatErr)
		}
		if again != first {
			t.Fatalf("round %d digest %s != %s", round, again, first)
		}
	}
}

// TestFingerprintDetectsMutationBetweenRuns keeps the fingerprint honest about
// content: the digest must move when any hashed byte moves, and must not move
// for metadata the canonical encoding deliberately excludes.
func TestFingerprintDetectsMutationBetweenRuns(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{"a/b/c": "before", "a/sibling": "x"})
	before, _, err := fingerprintToolchain(context.Background(), root, equivalenceVersion)
	if err != nil {
		t.Fatal(err)
	}
	writeTree(t, root, map[string]string{"a/b/c": "after"})
	after, _, err := fingerprintToolchain(context.Background(), root, equivalenceVersion)
	if err != nil {
		t.Fatal(err)
	}
	if before == after {
		t.Fatal("content change did not move the digest")
	}
	assertEquivalent(t, root)
}

// TestFingerprintCancellationStaysFailClosed covers the cancellation contract on
// both traversal phases; the scoped walk checks the context per directory and
// per entry exactly where the WalkDir callback did.
func TestFingerprintCancellationStaysFailClosed(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"a/b/c/d/leaf": strings.Repeat("x", 256*1024),
		"a/other":      "x",
		"z":            "x",
	})

	t.Run("cancelled before the walk", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, _, legacyErr := legacyFingerprintToolchain(ctx, root, equivalenceVersion)
		_, _, scopedErr := fingerprintToolchain(ctx, root, equivalenceVersion)
		if DiagnosticCode(legacyErr) != "toolchain_timeout" || DiagnosticCode(scopedErr) != "toolchain_timeout" {
			t.Fatalf("legacy %v, scoped %v", legacyErr, scopedErr)
		}
		if legacyErr.Error() != scopedErr.Error() {
			t.Fatalf("detail: legacy %q, scoped %q", legacyErr.Error(), scopedErr.Error())
		}
	})

	t.Run("cancelled on an empty root", func(t *testing.T) {
		empty := t.TempDir()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, _, err := fingerprintToolchain(ctx, empty, equivalenceVersion); DiagnosticCode(err) != "toolchain_timeout" {
			t.Fatalf("error = %v, want toolchain_timeout", err)
		}
	})

	t.Run("cancelled between the walk and the digest", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		go cancel()
		_, _, err := fingerprintToolchain(ctx, root, equivalenceVersion)
		if err != nil && DiagnosticCode(err) != "toolchain_timeout" {
			t.Fatalf("error = %v, want nil or toolchain_timeout", err)
		}
	})

	// The same window, but taken deterministically at the phase seam instead of
	// racing a goroutine, so both traversals are compared on the same event.
	t.Run("cancelled exactly at the phase boundary", func(t *testing.T) {
		handle, err := os.OpenRoot(root)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = handle.Close() }()

		ctx, cancel := context.WithCancel(context.Background())
		legacyRecords, legacyCollectErr := legacyCollectRecords(ctx, handle)
		scopedRecords, scopedCollectErr := collectToolchainRecords(ctx, handle)
		if legacyCollectErr != nil || scopedCollectErr != nil {
			t.Fatalf("collection failed before cancellation: %v / %v", legacyCollectErr, scopedCollectErr)
		}
		cancel()

		_, _, legacyErr := legacyDigestRecords(ctx, handle, legacyRecords, equivalenceVersion)
		_, _, scopedErr := digestToolchainRecords(ctx, handle, scopedRecords, equivalenceVersion)
		if DiagnosticCode(legacyErr) != "toolchain_timeout" || DiagnosticCode(scopedErr) != "toolchain_timeout" {
			t.Fatalf("legacy %v, scoped %v", legacyErr, scopedErr)
		}
		if legacyErr.Error() != scopedErr.Error() {
			t.Fatalf("detail: legacy %q, scoped %q", legacyErr.Error(), scopedErr.Error())
		}
	})
}

// TestFingerprintDoesNotDescendLinkedDirectories pins the containment property
// that lets the scoped walk open child handles at all: a symlinked directory is
// a leaf record, so no descended component is ever a link.
func TestFingerprintDoesNotDescendLinkedDirectories(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{"real/secret": "hidden"})
	symlinkOrSkip(t, "real", filepath.Join(root, "alias"))

	_, state, err := fingerprintToolchain(context.Background(), root, equivalenceVersion)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]byte{}
	for _, item := range state {
		seen[item.path] = item.kind
	}
	if seen["alias"] != 'L' {
		t.Fatalf("alias kind = %q, want 'L'", seen["alias"])
	}
	if _, descended := seen["alias/secret"]; descended {
		t.Fatalf("walk descended a linked directory: %v", seen)
	}
	if seen["real/secret"] != 'F' {
		t.Fatalf("real/secret kind = %q, want 'F'", seen["real/secret"])
	}
}

// assertPhaseBoundaryEquivalent builds the same tree for both traversals,
// collects the canonical record set, applies a mutation exactly between the
// collection and the digest phases, and requires both to reach the same outcome.
//
// This is the window a reused ancestor handle would open. The preserved
// traversal re-resolves every file path from GOROOT in the digest phase, so it
// observes whatever now occupies that path; the shipped traversal must observe
// the same thing rather than keep reading through a directory that has been
// renamed out of the tree.
func assertPhaseBoundaryEquivalent(t *testing.T, build, mutate func(t *testing.T, root string)) error {
	t.Helper()

	runLegacy := func() (string, error) {
		root := t.TempDir()
		build(t, root)
		handle, err := os.OpenRoot(root)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = handle.Close() }()
		records, collectErr := legacyCollectRecords(context.Background(), handle)
		if collectErr != nil {
			t.Fatalf("legacy collection failed before the mutation: %v", collectErr)
		}
		mutate(t, root)
		digest, _, digestErr := legacyDigestRecords(context.Background(), handle, records, equivalenceVersion)
		return digest, digestErr
	}
	runScoped := func() (string, error) {
		root := t.TempDir()
		build(t, root)
		handle, err := os.OpenRoot(root)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = handle.Close() }()
		records, collectErr := collectToolchainRecords(context.Background(), handle)
		if collectErr != nil {
			t.Fatalf("scoped collection failed before the mutation: %v", collectErr)
		}
		mutate(t, root)
		digest, _, digestErr := digestToolchainRecords(context.Background(), handle, records, equivalenceVersion)
		return digest, digestErr
	}

	legacyDigest, legacyErr := runLegacy()
	scopedDigest, scopedErr := runScoped()

	switch {
	case legacyErr == nil && scopedErr != nil:
		t.Fatalf("legacy accepted the mutated tree, scoped failed: %v", scopedErr)
	case legacyErr != nil && scopedErr == nil:
		t.Fatalf("legacy failed with %v, scoped accepted the mutated tree", legacyErr)
	case legacyErr != nil:
		if DiagnosticCode(legacyErr) != DiagnosticCode(scopedErr) {
			t.Fatalf("diagnostic code: legacy %q, scoped %q (%v / %v)",
				DiagnosticCode(legacyErr), DiagnosticCode(scopedErr), legacyErr, scopedErr)
		}
		if legacyErr.Error() != scopedErr.Error() {
			t.Fatalf("diagnostic detail:\n legacy %q\n scoped %q", legacyErr.Error(), scopedErr.Error())
		}
		return scopedErr
	}
	if legacyDigest != scopedDigest {
		t.Fatalf("digest: legacy %s, scoped %s", legacyDigest, scopedDigest)
	}
	return nil
}

// TestFingerprintDigestPhaseResolvesEveryFileFromTheRoot is the regression that
// the earlier directory-handle cache failed: hashed bytes must come from the
// file reachable at the canonical GOROOT path at read time, never from a handle
// to a directory that has since been renamed out of the tree.
func TestFingerprintDigestPhaseResolvesEveryFileFromTheRoot(t *testing.T) {
	build := func(t *testing.T, root string) {
		t.Helper()
		writeTree(t, root, map[string]string{
			"pkg/first":  "old-first",
			"pkg/second": "old-second",
			"z-last":     "tail",
		})
	}
	detach := func(t *testing.T, root string) {
		t.Helper()
		if err := os.Rename(filepath.Join(root, "pkg"), filepath.Join(root, "detached")); err != nil {
			t.Fatal(err)
		}
	}
	replaceFile := func(t *testing.T, name, payload string) {
		t.Helper()
		if err := os.Remove(name); err != nil {
			t.Fatal(err)
		}
		writeTestFile(t, name, []byte(payload), 0o644)
	}

	cases := map[string]struct {
		mutate func(t *testing.T, root string)
		code   string
	}{
		"unchanged between the phases": {
			mutate: func(*testing.T, string) {},
		},
		"directory renamed away and replaced": {
			mutate: func(t *testing.T, root string) {
				detach(t, root)
				writeTree(t, root, map[string]string{
					"pkg/first":  "new-first",
					"pkg/second": "new-second",
				})
			},
			code: "toolchain_mutated",
		},
		"directory renamed away with nothing in its place": {
			mutate: detach,
			code:   "toolchain_unreadable",
		},
		// os.Root refuses only the links that leave the root, so a link to the
		// detached copy still resolves and still reaches the same inodes. Both
		// traversals therefore accept it. This is pre-existing os.Root
		// behaviour, pinned here so the equivalence claim covers it rather than
		// quietly avoiding it.
		"directory replaced by a link to the detached copy": {
			mutate: func(t *testing.T, root string) {
				detach(t, root)
				symlinkOrSkip(t, "detached", filepath.Join(root, "pkg"))
			},
		},
		"single file swapped in place": {
			mutate: func(t *testing.T, root string) {
				replaceFile(t, filepath.Join(root, "pkg", "second"), "new-second")
			},
			code: "toolchain_mutated",
		},
		// Renaming the directory away and straight back leaves every inode
		// where the record set expects it. Neither traversal can distinguish
		// that from no mutation at all, and both accept it; the point of the
		// case is that they agree rather than that either detects it.
		"directory renamed away and back": {
			mutate: func(t *testing.T, root string) {
				detach(t, root)
				if err := os.Rename(filepath.Join(root, "detached"), filepath.Join(root, "pkg")); err != nil {
					t.Fatal(err)
				}
			},
		},
	}

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			err := assertPhaseBoundaryEquivalent(t, build, testCase.mutate)
			if DiagnosticCode(err) != testCase.code {
				t.Fatalf("code = %q (%v), want %q", DiagnosticCode(err), err, testCase.code)
			}
		})
	}
}

// TestFingerprintDigestPhaseResolvesReplacedAncestors is the same regression one
// level further up: it is not only the file's own directory that must be
// re-resolved, but every component above it.
func TestFingerprintDigestPhaseResolvesReplacedAncestors(t *testing.T) {
	build := func(t *testing.T, root string) {
		t.Helper()
		writeTree(t, root, map[string]string{
			"a/b/c/leaf":    "old-leaf",
			"a/b/c/sibling": "old-sibling",
			"a/top":         "old-top",
		})
	}
	mutate := func(t *testing.T, root string) {
		t.Helper()
		if err := os.Rename(filepath.Join(root, "a"), filepath.Join(root, "detached")); err != nil {
			t.Fatal(err)
		}
		writeTree(t, root, map[string]string{
			"a/b/c/leaf":    "new-leaf",
			"a/b/c/sibling": "new-sibling",
			"a/top":         "new-top",
		})
	}

	err := assertPhaseBoundaryEquivalent(t, build, mutate)
	if DiagnosticCode(err) != "toolchain_mutated" {
		t.Fatalf("code = %q (%v), want toolchain_mutated", DiagnosticCode(err), err)
	}
}

// TestToolchainWalkAnchorsDirectoryAndLinkMetadataAtTheRoot pins the asymmetry
// the collection phase relies on. A file's Lstat may be taken through the
// scoped directory handle because the digest phase re-opens that file from
// GOROOT and matches it with os.SameFile. A directory record and a link target
// are never revisited, so both are resolved from GOROOT instead.
//
// Pointing the walk's root at a tree that does not contain the entries proves
// which resolution each step actually uses: the directory and the link fail
// because they are looked up through the root, and the plain file does not.
func TestToolchainWalkAnchorsDirectoryAndLinkMetadataAtTheRoot(t *testing.T) {
	cases := map[string]struct {
		build func(t *testing.T, dir string)
		code  string
	}{
		"file metadata is taken through the scoped handle": {
			build: func(t *testing.T, dir string) {
				writeTree(t, dir, map[string]string{"entry": "x"})
			},
		},
		"directory metadata is taken through the root": {
			build: func(t *testing.T, dir string) {
				if err := os.MkdirAll(filepath.Join(dir, "entry"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			code: "toolchain_unreadable",
		},
		"link targets are read through the root": {
			build: func(t *testing.T, dir string) {
				writeTree(t, dir, map[string]string{"target": "x"})
				symlinkOrSkip(t, "target", filepath.Join(dir, "entry"))
			},
			code: "toolchain_link_dangling",
		},
	}

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			listed := t.TempDir()
			testCase.build(t, listed)
			listedHandle, err := os.OpenRoot(listed)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = listedHandle.Close() }()
			anchorHandle, err := os.OpenRoot(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = anchorHandle.Close() }()

			walk := toolchainWalk{root: anchorHandle, encoded: make(map[string]struct{})}
			walkErr := walk.descend(context.Background(), listedHandle, "")
			if DiagnosticCode(walkErr) != testCase.code {
				t.Fatalf("code = %q (%v), want %q", DiagnosticCode(walkErr), walkErr, testCase.code)
			}
		})
	}
}

// TestToolchainWalkRejectsDirectoryReplacedByFile pins the traversal decision
// made by fs.WalkDir: whether to descend comes from the directory entry that
// was listed, not from a later lstat. If that listed directory is replaced by a
// file before inspection, the legacy walk still attempts to descend and fails
// closed. The scoped walk must not reinterpret it as a regular-file leaf and
// silently omit the directory's former descendants.
func TestToolchainWalkRejectsDirectoryReplacedByFile(t *testing.T) {
	listed := t.TempDir()
	if err := os.MkdirAll(filepath.Join(listed, "entry"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeTree(t, listed, map[string]string{"entry/child": "old"})
	listedHandle, err := os.OpenRoot(listed)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listedHandle.Close() }()

	anchor := t.TempDir()
	writeTree(t, anchor, map[string]string{"entry": "replacement"})
	anchorHandle, err := os.OpenRoot(anchor)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = anchorHandle.Close() }()

	walk := toolchainWalk{root: anchorHandle, encoded: make(map[string]struct{})}
	if walkErr := walk.descend(context.Background(), listedHandle, ""); walkErr == nil {
		t.Fatalf("scoped walk accepted a file in place of a listed directory: %+v", walk.records)
	}
}

// TestFingerprintReportsUnreadableDirectoryIdentically covers the walk-error
// branch, which fs.WalkDir surfaced by re-invoking the callback with the
// ReadDir failure and the scoped walk surfaces at the same point.
func TestFingerprintReportsUnreadableDirectoryIdentically(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	root := t.TempDir()
	writeTree(t, root, map[string]string{"locked/file": "x", "readable": "y"})
	locked := filepath.Join(root, "locked")
	if err := os.Chmod(locked, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })

	assertEquivalent(t, root)
	_, _, err := fingerprintToolchain(context.Background(), root, equivalenceVersion)
	if DiagnosticCode(err) != "toolchain_unreadable" {
		t.Fatalf("code = %q (%v), want toolchain_unreadable", DiagnosticCode(err), err)
	}
}

func benchmarkRoot(b *testing.B) string {
	b.Helper()
	goroot := build.Default.GOROOT
	if goroot == "" {
		b.Skip("no GOROOT available")
	}
	if _, err := os.Stat(goroot); err != nil {
		b.Skipf("GOROOT %s unreadable: %v", goroot, err)
	}
	return goroot
}

func BenchmarkFingerprintToolchainLegacy(b *testing.B) {
	root := benchmarkRoot(b)
	for b.Loop() {
		if _, _, err := legacyFingerprintToolchain(context.Background(), root, equivalenceVersion); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFingerprintToolchainScoped(b *testing.B) {
	root := benchmarkRoot(b)
	for b.Loop() {
		if _, _, err := fingerprintToolchain(context.Background(), root, equivalenceVersion); err != nil {
			b.Fatal(err)
		}
	}
}
