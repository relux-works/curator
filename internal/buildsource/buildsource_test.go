package buildsource

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/hashing"
)

func writeTestFile(t *testing.T, root, path string, content []byte) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, content, 0o644); err != nil {
		t.Fatal(err)
	}
}

func validateTestTree(t *testing.T, root string) *Token {
	t.Helper()
	token, err := Validate(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = token.Close() })
	return token
}

func TestAuthoritativeFramingVectors(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		token := validateTestTree(t, t.TempDir())
		sum := sha256.Sum256([]byte(domain))
		want := "sha256:" + hex.EncodeToString(sum[:])
		if got := token.Identity().ContentSHA256; got != want {
			t.Fatalf("empty digest = %s, want %s", got, want)
		}
	})

	t.Run("binary ordering marker and framing", func(t *testing.T) {
		root := t.TempDir()
		writeTestFile(t, root, "z-binary.bin", []byte{0, 0xff, 1})
		writeTestFile(t, root, "empty", nil)
		writeTestFile(t, root, ".csk-install.json", []byte("{\"variant\":\"fixture\"}\n"))
		writeTestFile(t, root, "é.txt", []byte("utf8\n"))

		got := validateTestTree(t, root).Identity().ContentSHA256
		const want = "sha256:68008c9a1131c1295d78f4f7d184c3df5f7382a88d8d40333be7cf02b2ee4de9"
		if got != want {
			t.Fatalf("digest = %s, want %s", got, want)
		}
	})

	t.Run("legacy structural collision is separated", func(t *testing.T) {
		one := t.TempDir()
		writeTestFile(t, one, "a", []byte{'x', 0, 'b', 0, 'y'})
		two := t.TempDir()
		writeTestFile(t, two, "a", []byte("x"))
		writeTestFile(t, two, "b", []byte("y"))

		oneDigest := validateTestTree(t, one).Identity().ContentSHA256
		twoDigest := validateTestTree(t, two).Identity().ContentSHA256
		if oneDigest != "sha256:96e3ed15c69a125ac033997b1f53baababece8a0be0831590f9282431ab6bc85" {
			t.Fatalf("one-file digest = %s", oneDigest)
		}
		if twoDigest != "sha256:15068f03268e971a11b928800ca920c6841dee76747bb3411ae93ff4ab77a334" {
			t.Fatalf("two-file digest = %s", twoDigest)
		}
		if oneDigest == twoDigest {
			t.Fatal("length framing did not separate legacy-colliding streams")
		}
	})
}

func TestMarkerChangesBuildSourceButNotLegacyHash(t *testing.T) {
	left := t.TempDir()
	right := t.TempDir()
	for _, root := range []string{left, right} {
		writeTestFile(t, root, "go.mod", []byte("module example\n"))
	}
	writeTestFile(t, left, hashing.MarkerName, []byte("{\"variant\":\"A\"}\n"))
	writeTestFile(t, right, hashing.MarkerName, []byte("{\"variant\":\"B\"}\n"))

	leftLegacy, err := hashing.ContentSHA256(left, nil)
	if err != nil {
		t.Fatal(err)
	}
	rightLegacy, err := hashing.ContentSHA256(right, nil)
	if err != nil {
		t.Fatal(err)
	}
	if leftLegacy != rightLegacy {
		t.Fatalf("legacy hashes differ: %s != %s", leftLegacy, rightLegacy)
	}

	leftBuild := validateTestTree(t, left).Identity().ContentSHA256
	rightBuild := validateTestTree(t, right).Identity().ContentSHA256
	if leftBuild == rightBuild {
		t.Fatal("root marker bytes must affect build-source identity")
	}
}

func TestModesAndTimestampsAreNotFrozenInputs(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "file", []byte("content"))
	token := validateTestTree(t, root)
	before := token.Identity()
	file := filepath.Join(root, "file")
	if err := os.Chmod(file, 0o755); err != nil {
		t.Fatal(err)
	}
	changedTime := time.Unix(1_893_456_000, 0)
	if err := os.Chtimes(file, changedTime, changedTime); err != nil {
		t.Fatal(err)
	}
	if err := token.Recheck(); err != nil {
		t.Fatalf("mode/timestamp-only change invalidated token: %v", err)
	}
	if token.Identity() != before {
		t.Fatal("mode/timestamp changed frozen identity")
	}
}

func TestRejectsInvalidPathsLinksAndCollisions(t *testing.T) {
	t.Run("invalid protocol path", func(t *testing.T) {
		root := t.TempDir()
		name := writeInvalidProtocolPathEntry(t, root)
		token, err := Validate(root)
		if token != nil {
			// A token holds an open root handle, and on Windows that handle
			// blocks the temporary directory's own removal, so a case that
			// wrongly accepted the tree would fail the *next* case's cleanup
			// instead of its own assertion.
			t.Cleanup(func() { _ = token.Close() })
		}
		if !errors.Is(err, ErrInvalidSnapshot) {
			t.Fatalf("Validate error for %q = %v", name, err)
		}
	})

	t.Run("symbolic link", func(t *testing.T) {
		root := t.TempDir()
		writeTestFile(t, root, "target", []byte("x"))
		if err := os.Symlink("target", filepath.Join(root, "link")); err != nil {
			t.Skipf("symlink unavailable: %v", err)
		}
		if _, err := Validate(root); !errors.Is(err, ErrInvalidSnapshot) {
			t.Fatalf("Validate error = %v", err)
		}
	})

	t.Run("symbolic link root", func(t *testing.T) {
		parent := t.TempDir()
		target := filepath.Join(parent, "target")
		if err := os.Mkdir(target, 0o755); err != nil {
			t.Fatal(err)
		}
		root := filepath.Join(parent, "root")
		if err := os.Symlink("target", root); err != nil {
			t.Skipf("symlink unavailable: %v", err)
		}
		if _, err := Validate(root); !errors.Is(err, ErrInvalidSnapshot) {
			t.Fatalf("Validate error = %v", err)
		}
	})

	t.Run("duplicate encoded path", func(t *testing.T) {
		paths := newPathSet()
		if err := paths.add("same"); err != nil {
			t.Fatal(err)
		}
		if err := paths.add("same"); !errors.Is(err, ErrInvalidSnapshot) {
			t.Fatalf("duplicate error = %v", err)
		}
	})

	t.Run("invalid UTF-8", func(t *testing.T) {
		if err := newPathSet().add(string([]byte{0xff})); !errors.Is(err, ErrInvalidSnapshot) {
			t.Fatalf("invalid UTF-8 error = %v", err)
		}
	})

	t.Run("case-distinct encoded paths", func(t *testing.T) {
		paths := newPathSet()
		if err := paths.add("Dir/File"); err != nil {
			t.Fatal(err)
		}
		if err := paths.add("dir/file"); err != nil {
			t.Fatalf("distinct encoded path error = %v", err)
		}
	})

	t.Run("normalization-distinct encoded paths", func(t *testing.T) {
		paths := newPathSet()
		if err := paths.add("é.txt"); err != nil {
			t.Fatal(err)
		}
		if err := paths.add("e\u0301.txt"); err != nil {
			t.Fatalf("distinct encoded path error = %v", err)
		}
	})
}

func TestFrozenTokenRejectsMutation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, string)
	}{
		{
			name: "file bytes",
			mutate: func(t *testing.T, root string) {
				writeTestFile(t, root, "file", []byte("changed"))
			},
		},
		{
			name: "tree",
			mutate: func(t *testing.T, root string) {
				if err := os.Mkdir(filepath.Join(root, "empty-directory"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "file type",
			mutate: func(t *testing.T, root string) {
				if err := os.Remove(filepath.Join(root, "file")); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink("other", filepath.Join(root, "file")); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			writeTestFile(t, root, "file", []byte("original"))
			writeTestFile(t, root, "other", []byte("other"))
			token := validateTestTree(t, root)
			test.mutate(t, root)
			if err := token.Recheck(); !errors.Is(err, ErrSnapshotMutated) {
				t.Fatalf("Recheck error = %v", err)
			}
		})
	}
}

// TestFrozenTokenRejectsRootReplacement proves the frozen token refuses a root
// whose name now leads to a different directory instance, even when the
// replacement carries byte-identical content.
//
// How a root is replaced is not portable, so the fixture is. On POSIX the
// validated directory is renamed aside while its handle is open and a
// same-content directory takes the name. Windows refuses that move outright —
// os.OpenRoot opens the root with FILE_SHARE_READ|FILE_SHARE_WRITE and no
// FILE_SHARE_DELETE, so the kernel pins the directory for as long as the token
// is frozen — and newFrozenRootCase there asserts that refusal and then drives
// the replacement the platform does allow, through a directory reparse point in
// the path prefix. Both runners reach this same Recheck assertion; neither is
// exempted from it.
func TestFrozenTokenRejectsRootReplacement(t *testing.T) {
	fixture := newFrozenRootCase(t, t.TempDir())
	writeTestFile(t, fixture.root, "file", []byte("same"))
	token := validateTestTree(t, fixture.root)
	fixture.replace(t)
	if err := token.Recheck(); !errors.Is(err, ErrSnapshotMutated) {
		t.Fatalf("Recheck error = %v", err)
	}
}

func TestWithValidatedOrdersCallbackAndRejectsMutation(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "file", []byte("before"))
	called := false
	err := WithValidated(root, func(token *Token) error {
		called = true
		if token.Identity().Algorithm != Algorithm || token.Identity().ContentSHA256 == "" {
			t.Fatalf("callback received incomplete identity: %#v", token.Identity())
		}
		writeTestFile(t, root, "file", []byte("after"))
		return nil
	})
	if !called {
		t.Fatal("callback was not called after validation")
	}
	if !errors.Is(err, ErrSnapshotMutated) {
		t.Fatalf("WithValidated error = %v", err)
	}

	invalid := t.TempDir()
	writeTestFile(t, invalid, "target", []byte("x"))
	if err := os.Symlink("target", filepath.Join(invalid, "link")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	called = false
	err = WithValidated(invalid, func(*Token) error {
		called = true
		return nil
	})
	if !errors.Is(err, ErrInvalidSnapshot) || called {
		t.Fatalf("invalid snapshot error = %v, callback called = %v", err, called)
	}
}
