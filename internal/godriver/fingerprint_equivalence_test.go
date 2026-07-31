package godriver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
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
			if err := digestCopyDiagnostic(record.path, written, opened.Size(), copyErr, closeErr); err != nil {
				return "", nil, err
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

	// A racing cancellation lands wherever the scheduler puts it, which on an
	// unloaded machine is almost always before the first file is read. That makes
	// this subtest a smoke check and nothing more: it is not what proves the
	// contract, because the window it needs to hit only opens wide enough to be
	// sampled when the machine is loaded — which is how the original failure was
	// found, and why repeating it here cannot be relied on to find the next one.
	// The deterministic subtests below take every cancellation point on purpose.
	t.Run("cancelled between the walk and the digest", func(t *testing.T) {
		for attempt := range 200 {
			ctx, cancel := context.WithCancel(context.Background())
			go cancel()
			_, _, err := fingerprintToolchain(ctx, root, equivalenceVersion)
			if err != nil && DiagnosticCode(err) != "toolchain_timeout" {
				t.Fatalf("attempt %d: error = %v, want nil or toolchain_timeout", attempt, err)
			}
		}
	})

	// The same window taken deterministically: the context ends on an exact
	// ctx.Err() call inside a record's copy loop, past the per-record check that
	// guards the top of the digest phase. A cancellation there stops the read
	// short, and a short read is the only thing a genuine concurrent write is
	// visible as, so this is precisely where the two must not be confused.
	t.Run("cancelled inside a record read", func(t *testing.T) {
		handle, err := os.OpenRoot(root)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = handle.Close() }()

		scopedRecords, scopedCollectErr := collectToolchainRecords(context.Background(), handle)
		legacyRecords, legacyCollectErr := legacyCollectRecords(context.Background(), handle)
		if scopedCollectErr != nil || legacyCollectErr != nil {
			t.Fatalf("collection failed before cancellation: %v / %v", legacyCollectErr, scopedCollectErr)
		}

		// The digest phase spends one check per record before it reaches the
		// first file, then one more per copied chunk. Counting the leading
		// records keeps the budgets anchored to the tree rather than to a
		// hand-written number that a later fixture edit would silently shift.
		checksToFirstFile := 0
		for _, record := range scopedRecords {
			checksToFirstFile++
			if record.kind == 'F' {
				break
			}
		}
		if checksToFirstFile == len(scopedRecords) && scopedRecords[len(scopedRecords)-1].kind != 'F' {
			t.Fatalf("fixture has no file record: %v", scopedRecords)
		}

		for _, budget := range []struct {
			name   string
			checks int
		}{
			{"before the first chunk", checksToFirstFile},
			{"after a partial chunk", checksToFirstFile + 1},
			{"after the last chunk", checksToFirstFile + 2},
		} {
			t.Run(budget.name, func(t *testing.T) {
				_, _, legacyErr := legacyDigestRecords(newCountdownContext(budget.checks), handle, legacyRecords, equivalenceVersion)
				_, _, scopedErr := digestToolchainRecords(newCountdownContext(budget.checks), handle, scopedRecords, equivalenceVersion)
				if DiagnosticCode(legacyErr) != "toolchain_timeout" || DiagnosticCode(scopedErr) != "toolchain_timeout" {
					t.Fatalf("legacy %v, scoped %v", legacyErr, scopedErr)
				}
				if legacyErr.Error() != scopedErr.Error() {
					t.Fatalf("detail: legacy %q, scoped %q", legacyErr.Error(), scopedErr.Error())
				}
				// The precise cause stays reachable for anything upstream that
				// inspects it; only the reported boundary code changes.
				if !errors.Is(scopedErr, context.Canceled) {
					t.Fatalf("scoped error %v does not unwrap to context.Canceled", scopedErr)
				}
			})
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

	// The contract in full, with nothing left to the scheduler: every cancellation
	// point the fingerprint can reach is taken in turn, and each one must produce
	// the deadline diagnostic or a clean fingerprint — never toolchain_mutated,
	// because a context ending is not the toolchain changing. Walking the budget
	// up rather than naming a few interesting values is what makes this a proof
	// instead of a sample: a check that a later edit adds inside the copy loop, or
	// anywhere else in either phase, is covered the moment it exists.
	t.Run("every cancellation point stays a deadline", func(t *testing.T) {
		// The sweep has to run past the last check to be exhaustive, and the
		// budget that first completes cleanly proves it got there.
		completed := false
		for budget := 0; budget < 200; budget++ {
			ctx := newCountdownContext(budget)
			_, _, err := fingerprintToolchain(ctx, root, equivalenceVersion)
			switch code := DiagnosticCode(err); {
			case err == nil:
				completed = true
			case code == "toolchain_timeout":
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("budget %d: %v does not unwrap to context.Canceled", budget, err)
				}
			default:
				t.Fatalf("budget %d: error = %v, want nil or toolchain_timeout", budget, err)
			}
			if completed {
				break
			}
		}
		if !completed {
			t.Fatal("no budget in the sweep completed the fingerprint; the sweep never reached the last check")
		}
	})
}

// countdownContext reports no error for a fixed number of Err calls and reports
// cancellation for every call after that. It exists so a test can place a
// cancellation on an exact check inside the digest phase — including checks
// taken between two chunks of one file — instead of racing a goroutine against
// a read and hoping it lands in the window under test. It is used from a single
// goroutine, which is how the digest phase consumes a context.
type countdownContext struct {
	context.Context
	remaining int
	done      chan struct{}
}

func newCountdownContext(checks int) *countdownContext {
	return &countdownContext{Context: context.Background(), remaining: checks, done: make(chan struct{})}
}

func (ctx *countdownContext) Err() error {
	if ctx.remaining > 0 {
		ctx.remaining--
		return nil
	}
	select {
	case <-ctx.done:
	default:
		close(ctx.done)
	}
	return context.Canceled
}

func (ctx *countdownContext) Done() <-chan struct{} { return ctx.done }

// TestDigestCopyDiagnosticPrecedence pins how narrow the cancellation branch is.
// Only the abort this package raises inside its own copy loop is read as a
// cancellation; every other way a copy can end badly still fails closed as
// toolchain_mutated, including a bare context error that did not come from that
// loop. Cancellation therefore cannot launder a concurrent write into a
// deadline, which is the half of the contract the fixture-driven tests above
// cannot force.
func TestDigestCopyDiagnosticPrecedence(t *testing.T) {
	abandoned := func(cause error) error { return fmt.Errorf("%w: %w", errCopyAbandoned, cause) }

	for _, testCase := range []struct {
		name              string
		written, size     int64
		copyErr, closeErr error
		code              string
	}{
		{name: "complete copy", written: 4, size: 4},
		{name: "abandoned before the first byte", written: 0, size: 4, copyErr: abandoned(context.Canceled), code: "toolchain_timeout"},
		{name: "abandoned mid-file", written: 2, size: 4, copyErr: abandoned(context.DeadlineExceeded), code: "toolchain_timeout"},
		{name: "abandoned after the last byte", written: 4, size: 4, copyErr: abandoned(context.Canceled), code: "toolchain_timeout"},
		{name: "bare context error is not this loop's abort", written: 2, size: 4, copyErr: context.Canceled, code: "toolchain_mutated"},
		{name: "read failure", written: 2, size: 4, copyErr: errors.New("input/output error"), code: "toolchain_mutated"},
		{name: "close failure", written: 4, size: 4, closeErr: errors.New("cannot close"), code: "toolchain_mutated"},
		{name: "file shrank under the read", written: 2, size: 4, code: "toolchain_mutated"},
		{name: "file grew under the read", written: 6, size: 4, code: "toolchain_mutated"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			err := digestCopyDiagnostic("bin/go", testCase.written, testCase.size, testCase.copyErr, testCase.closeErr)
			if DiagnosticCode(err) != testCase.code {
				t.Fatalf("code = %q, want %q (error %v)", DiagnosticCode(err), testCase.code, err)
			}
			if testCase.code == "" && err != nil {
				t.Fatalf("error = %v, want nil", err)
			}
		})
	}
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

// TestToolchainWalkRejectsDirectoryReplacedByFile is the tester's regression
// from RUN-260729 (tester-collection-race), kept under its original name and
// with its original assertion so the finding stays traceable. It is subsumed by
// TestToolchainWalkTakesDescentFromTheListedEntry below, which additionally
// pins the diagnostic code, the operator detail and the resulting record set.
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

// TestToolchainWalkTakesDescentFromTheListedEntry pins the traversal decision
// fs.WalkDir makes: whether to descend comes from the directory entry that was
// listed, never from a later lstat. Deciding it from the resolved metadata
// instead lets a listed directory that has become a regular file be recorded as
// a regular-file leaf, silently dropping every descendant it had — where
// fs.WalkDir still calls fs.ReadDir on it and fails closed.
//
// The two-handle construction places the mutation deterministically: the walk
// lists its entries from one tree and resolves them against another, which is
// exactly the state a rename between the readdir and the lstat produces, with
// no goroutine race to lose.
//
// The expectations are read off legacyCollectRecords rather than asserted from
// taste: an fs.ReadDir failure reaches its walk function as walkErr and becomes
// "cannot walk GOROOT"; a failing root.Lstat becomes "cannot inspect toolchain
// path %q". Both codes and both details must survive verbatim.
func TestToolchainWalkTakesDescentFromTheListedEntry(t *testing.T) {
	cases := map[string]struct {
		listed  func(t *testing.T, dir string)
		anchor  func(t *testing.T, dir string)
		code    string
		detail  string
		records []string
	}{
		// fs.ReadDir on a path that is no longer a directory fails, so
		// fs.WalkDir reports it through walkErr. The descendants the listing
		// promised must not survive as records.
		"listed directory replaced by a file": {
			listed: func(t *testing.T, dir string) {
				writeTree(t, dir, map[string]string{"entry/child": "old"})
			},
			anchor: func(t *testing.T, dir string) {
				writeTree(t, dir, map[string]string{"entry": "replacement"})
			},
			code:    "toolchain_unreadable",
			detail:  "cannot walk GOROOT",
			records: []string{"entry"},
		},
		// os.Root resolves links that stay inside the root, so fs.ReadDir on a
		// link to an in-root directory succeeds and fs.WalkDir descends it.
		// This is pre-existing os.Root behaviour, verified against
		// fs.ReadDir(root.FS(), ...) directly; the scoped walk must reproduce
		// it rather than tighten it, including the leaf 'L' record it leaves
		// at the link's own path.
		"listed directory replaced by an in-root link to a directory": {
			listed: func(t *testing.T, dir string) {
				writeTree(t, dir, map[string]string{"entry/child": "old"})
			},
			anchor: func(t *testing.T, dir string) {
				writeTree(t, dir, map[string]string{"target/child": "other"})
				symlinkOrSkip(t, "target", filepath.Join(dir, "entry"))
			},
			records: []string{"entry", "entry/child"},
		},
		// The full-path Lstat fires before the descent, exactly as it does in
		// legacyCollectRecords, so this reports the Lstat detail and not the
		// walk detail.
		"listed directory renamed away with nothing in its place": {
			listed: func(t *testing.T, dir string) {
				writeTree(t, dir, map[string]string{"entry/child": "old"})
			},
			anchor: func(*testing.T, string) {},
			code:   "toolchain_unreadable",
			detail: `cannot inspect toolchain path "entry"`,
		},
	}

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			listed := t.TempDir()
			testCase.listed(t, listed)
			listedHandle, err := os.OpenRoot(listed)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = listedHandle.Close() }()

			anchor := t.TempDir()
			testCase.anchor(t, anchor)
			anchorHandle, err := os.OpenRoot(anchor)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = anchorHandle.Close() }()

			// Ground truth for whether fs.WalkDir would have descended this
			// path, taken from the same fs.FS it used, so the expectation is
			// observed rather than assumed: fs.WalkDir descends exactly when
			// fs.ReadDir succeeds, and every shape that it cannot descend must
			// fail closed here.
			_, readDirErr := fs.ReadDir(anchorHandle.FS(), "entry")
			if (readDirErr != nil) != (testCase.code != "") {
				t.Fatalf("fs.ReadDir err=%v disagrees with the expected code %q", readDirErr, testCase.code)
			}

			walk := toolchainWalk{root: anchorHandle, encoded: make(map[string]struct{})}
			walkErr := walk.descend(context.Background(), listedHandle, "")
			if DiagnosticCode(walkErr) != testCase.code {
				t.Fatalf("code = %q (%v), want %q", DiagnosticCode(walkErr), walkErr, testCase.code)
			}
			if testCase.code != "" {
				var failure *Diagnostic
				if !errors.As(walkErr, &failure) || failure.Detail != testCase.detail {
					t.Fatalf("detail = %q, want %q", failure.Detail, testCase.detail)
				}
			}

			paths := make([]string, 0, len(walk.records))
			for _, record := range walk.records {
				paths = append(paths, record.path)
			}
			if strings.Join(paths, ",") != strings.Join(testCase.records, ",") {
				t.Fatalf("records = %v, want %v", paths, testCase.records)
			}
		})
	}
}

// TestFingerprintRejectsFileRecordResolvingToADirectory covers the opposite
// direction end to end. A listed regular file whose path is a directory by the
// time the digest phase opens it must not be hashed as a file: the scoped Lstat
// is trusted only because os.SameFile re-binds it against whatever GOROOT
// actually holds at read time.
func TestFingerprintRejectsFileRecordResolvingToADirectory(t *testing.T) {
	build := func(t *testing.T, root string) {
		writeTree(t, root, map[string]string{"entry": "payload"})
	}
	mutate := func(t *testing.T, root string) {
		entry := filepath.Join(root, "entry")
		if err := os.Remove(entry); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(entry, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	err := assertPhaseBoundaryEquivalent(t, build, mutate)
	if DiagnosticCode(err) != "toolchain_mutated" {
		t.Fatalf("code = %q (%v), want toolchain_mutated", DiagnosticCode(err), err)
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
	denyDirectoryListing(t, filepath.Join(root, "locked"))

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
