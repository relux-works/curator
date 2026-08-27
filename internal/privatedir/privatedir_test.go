package privatedir

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMakeCreatesADirectoryValidateAccepts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private")
	if err := Make(path); err != nil {
		t.Fatal(err)
	}
	if err := Validate(path); err != nil {
		t.Fatalf("fresh private directory rejected: %v", err)
	}
	if err := Make(path); !errors.Is(err, os.ErrExist) {
		t.Fatalf("second Make = %v, want os.ErrExist", err)
	}
}

func TestMakeAllCreatesEveryMissingLevel(t *testing.T) {
	base := t.TempDir()
	leaf := filepath.Join(base, "a", "b", "c")
	if err := MakeAll(leaf); err != nil {
		t.Fatal(err)
	}
	if err := Validate(leaf); err != nil {
		t.Fatalf("MakeAll leaf rejected: %v", err)
	}
	// Idempotent over an existing tree, like os.MkdirAll.
	if err := MakeAll(leaf); err != nil {
		t.Fatalf("repeated MakeAll = %v", err)
	}
}

func TestValidateRejectsWhatIsNotAPrivateDirectory(t *testing.T) {
	base := t.TempDir()

	missing := filepath.Join(base, "missing")
	if err := Validate(missing); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing path error = %v, want os.ErrNotExist", err)
	}

	file := filepath.Join(base, "file")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if Validate(file) == nil {
		t.Fatal("a regular file validated as a private directory")
	}

	// A directory created by plain os.Mkdir carries the platform's default
	// shape (group/other bits from the mode on Unix, an inherited DACL on
	// Windows) and must be rejected: the boundary only trusts what it made.
	ordinary := filepath.Join(base, "ordinary")
	if err := os.Mkdir(ordinary, 0o755); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		// Pin the loose shape regardless of the process umask.
		if err := os.Chmod(ordinary, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if Validate(ordinary) == nil {
		t.Fatal("an ordinary directory validated as private")
	}

	if runtime.GOOS != "windows" {
		linked := filepath.Join(base, "linked")
		target := filepath.Join(base, "target")
		if err := Make(target); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, linked); err != nil {
			t.Fatal(err)
		}
		if Validate(linked) == nil {
			t.Fatal("a symlink to a private directory validated as private")
		}
	}
}

func TestProtectUpgradesAnExistingDirectory(t *testing.T) {
	// os.MkdirTemp is the one creation path the boundary uses that cannot
	// attach the private shape at creation; Protect must close that gap.
	path, err := os.MkdirTemp(t.TempDir(), "blob-")
	if err != nil {
		t.Fatal(err)
	}
	if err := Protect(path); err != nil {
		t.Fatal(err)
	}
	if err := Validate(path); err != nil {
		t.Fatalf("protected directory rejected: %v", err)
	}
}

func TestPrivateDirectoryIsUsableByItsOwner(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private")
	if err := Make(path); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(path, "child.txt")
	if err := os.WriteFile(child, []byte("payload"), 0o600); err != nil {
		t.Fatalf("owner cannot write inside its private directory: %v", err)
	}
	payload, err := os.ReadFile(child)
	if err != nil || string(payload) != "payload" {
		t.Fatalf("owner cannot read back its file: %v %q", err, payload)
	}
	nested := filepath.Join(path, "nested")
	if err := os.Mkdir(nested, 0o700); err != nil {
		t.Fatalf("owner cannot create a nested directory: %v", err)
	}
	if err := os.RemoveAll(path); err != nil {
		t.Fatalf("owner cannot remove its private tree: %v", err)
	}
}
