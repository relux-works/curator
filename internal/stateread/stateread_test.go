package stateread

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestRegularFileAndLinkReadersKeepFailureClasses(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "state.json")
	if err := os.WriteFile(file, []byte("state\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := ReadRegularFile(file); err != nil || got.Kind != KindPresent || string(got.Bytes) != "state\n" {
		t.Fatalf("ReadRegularFile(file) = (%+v, %v), want present bytes", got, err)
	}
	if got, err := ReadRegularFile(filepath.Join(root, "missing")); err != nil || got.Kind != KindAbsent {
		t.Fatalf("ReadRegularFile(missing) = (%+v, %v), want absent", got, err)
	}
	if got, err := ReadRegularFile(root); err == nil || got.Kind != KindUnreadable {
		t.Fatalf("ReadRegularFile(directory) = (%+v, %v), want unreadable", got, err)
	}

	linkPath := filepath.Join(root, "link")
	if err := os.Symlink(file, linkPath); err != nil {
		t.Fatal(err)
	}
	if got, err := ReadRegularFile(linkPath); err == nil || got.Kind != KindUnreadable {
		t.Fatalf("ReadRegularFile(symlink) = (%+v, %v), want unreadable", got, err)
	}
	if got, err := Readlink(linkPath); err != nil || got.Kind != KindPresent || got.Target != file {
		t.Fatalf("Readlink(link) = (%+v, %v), want target %q", got, err, file)
	}
	if got, err := Readlink(filepath.Join(root, "missing-link")); err != nil || got.Kind != KindAbsent {
		t.Fatalf("Readlink(missing) = (%+v, %v), want absent", got, err)
	}
	failedLstat := func(string) (os.FileInfo, error) { return nil, errors.New("injected inspection failure") }
	if got, err := LstatWith(file, failedLstat); err == nil || got.Kind != KindUnreadable {
		t.Fatalf("LstatWith(failed) = (%+v, %v), want unreadable", got, err)
	}
	failedReadlink := func(string) (string, error) { return "", errors.New("injected link read failure") }
	if got, err := ReadlinkWith(linkPath, failedReadlink); err == nil || got.Kind != KindUnreadable {
		t.Fatalf("ReadlinkWith(failed) = (%+v, %v), want unreadable", got, err)
	}
}

// TestReadsDistinguishAbsentFromBlockedParent proves that a missing leaf in
// a traversable directory is absence, while read/stat failures beneath a
// regular-file ancestor remain unreadable on every platform.
func TestReadsDistinguishAbsentFromBlockedParent(t *testing.T) {
	root := t.TempDir()
	absent := filepath.Join(root, "absent")
	if got, err := ReadFile(absent); err != nil || got.Kind != KindAbsent {
		t.Fatalf("ReadFile(absent) = (%+v, %v), want absent", got, err)
	}
	if got, err := ReadDir(absent); err != nil || got.Kind != KindAbsent {
		t.Fatalf("ReadDir(absent) = (%+v, %v), want absent", got, err)
	}
	if got, err := Lstat(absent); err != nil || got.Kind != KindAbsent {
		t.Fatalf("Lstat(absent) = (%+v, %v), want absent", got, err)
	}

	blocker := filepath.Join(root, "blocker")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	targetDirectory := filepath.Join(root, "target-directory")
	if err := os.Mkdir(targetDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	directoryLink := filepath.Join(root, "directory-link")
	if err := os.Symlink(targetDirectory, directoryLink); err != nil {
		t.Fatal(err)
	}
	if got, err := ReadFile(filepath.Join(directoryLink, "absent")); err != nil || got.Kind != KindAbsent {
		t.Fatalf("ReadFile(absent below directory symlink) = (%+v, %v), want absent", got, err)
	}
	brokenLink := filepath.Join(root, "broken-link")
	if err := os.Symlink(filepath.Join(root, "missing-target"), brokenLink); err != nil {
		t.Fatal(err)
	}
	if got, err := ReadFile(filepath.Join(brokenLink, "child")); err == nil || got.Kind != KindUnreadable {
		t.Fatalf("ReadFile(child below broken symlink) = (%+v, %v), want unreadable", got, err)
	}
	blocked := filepath.Join(blocker, "child")
	// A Windows ERROR_PATH_NOT_FOUND for a child beneath a regular file is
	// equivalent to os.ErrNotExist at this classifier boundary. Keep that
	// shape explicit on every runner so narrowing mutants are portable.
	if missingPath(blocked, os.ErrNotExist) {
		t.Fatalf("missingPath(%q, os.ErrNotExist) = true below an existing file", blocked)
	}
	if missingPath(blocked, syscall.ENOTDIR) {
		t.Fatalf("missingPath(%q, ENOTDIR) = true; ENOTDIR is always unreadable", blocked)
	}
	windowsAbsence := func(err error) bool {
		return os.IsNotExist(err) || errors.Is(err, syscall.ENOTDIR)
	}
	if missingPathWith(blocked, syscall.ENOTDIR, windowsAbsence) {
		t.Fatalf("missingPathWith(%q, ENOTDIR, Windows IsNotExist) = true below a regular-file parent", blocked)
	}
	if !missingPath(absent, os.ErrNotExist) {
		t.Fatal("missingPath for a missing leaf under a traversable directory = false, want true")
	}
	if missingPath(absent, syscall.EACCES) {
		t.Fatal("missingPath treated permission denied as absence")
	}
	if got, err := ReadlinkWith(filepath.Join(blocker, "child-link"), func(string) (string, error) {
		return "", os.ErrNotExist
	}); err == nil || got.Kind != KindUnreadable {
		t.Fatalf("ReadlinkWith(ENOENT below non-directory parent) = (%+v, %v), want unreadable", got, err)
	}
	checks := []struct {
		name string
		read func() (Kind, error)
	}{
		{"read-file", func() (Kind, error) { result, err := ReadFile(blocked); return result.Kind, err }},
		{"read-dir", func() (Kind, error) { result, err := ReadDir(blocked); return result.Kind, err }},
		{"stat", func() (Kind, error) { result, err := Stat(blocked); return result.Kind, err }},
		{"lstat", func() (Kind, error) { result, err := Lstat(blocked); return result.Kind, err }},
		{"read-regular-file", func() (Kind, error) { result, err := ReadRegularFile(blocked); return result.Kind, err }},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			kind, err := check.read()
			var stateErr *Error
			if kind != KindUnreadable || !errors.As(err, &stateErr) || stateErr.Kind != KindUnreadable || stateErr.Path != blocked {
				t.Fatalf("read below regular-file parent = (%q, %v), want typed unreadable for %s", kind, err, blocked)
			}
		})
	}
}

func TestReadableRegularFileMetadataOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(path, []byte("synthetic"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := CheckReadableRegularFile(path); err != nil {
		t.Fatal(err)
	}
	if err := CheckReadableRegularFile(filepath.Dir(path)); err == nil {
		t.Fatal("directory accepted as readable regular credential")
	}
	absent, err := StatWith(path+".missing", nil)
	if err != nil || absent.Kind != KindAbsent {
		t.Fatalf("absence: %v %v", absent, err)
	}
	unreadable, err := StatWith(path, func(string) (os.FileInfo, error) { return nil, os.ErrPermission })
	if err == nil || unreadable.Kind != KindUnreadable {
		t.Fatalf("failed observation is absence: %v %v", unreadable, err)
	}
}
