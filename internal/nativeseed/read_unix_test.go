//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package nativeseed

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/stateread"
	"golang.org/x/sys/unix"
)

// Real filesystem cases distinguish absence from an unreadable present link;
// both regular and chained valid inputs retain their bytes and target mtime.
func TestNativeSeedReadContract(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		mode       os.FileMode
	}{
		{"regular", "", 0o600}, {"relative-link", "", 0o600}, {"chain", "", 0o600},
		{"absent", "absent", 0}, {"dangling", "resolve native seed", 0},
		{"directory", "must be a regular file", 0}, {"fifo", "must be a regular file", 0},
		{"group-write", "must not be group- or other-writable", 0o620},
		{"other-write", "must not be group- or other-writable", 0o602},
		{"loop", "resolve native seed", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "input")
			target := filepath.Join(root, "target")
			payload := []byte("fixture seed\n")
			if tc.mode != 0 {
				if err := os.WriteFile(target, payload, 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(target, tc.mode); err != nil {
					t.Fatal(err)
				}
				stamp := time.Unix(1600000000, 0)
				if err := os.Chtimes(target, stamp, stamp); err != nil {
					t.Fatal(err)
				}
			}
			var err error
			switch tc.name {
			case "absent":
			case "regular":
				path = target
			case "directory":
				err = os.Mkdir(target, 0o700)
			case "fifo":
				err = unix.Mkfifo(target, 0o600)
			case "loop":
				err = os.Symlink("input", target)
			case "chain":
				err = os.Symlink("target", filepath.Join(root, "middle"))
			}
			if err != nil {
				t.Fatal(err)
			}
			if tc.name != "absent" && tc.name != "regular" {
				dest := "target"
				if tc.name == "chain" {
					dest = "middle"
				}
				if err := os.Symlink(dest, path); err != nil {
					t.Fatal(err)
				}
			}
			state, err := ReadFile(path)
			if tc.want == "absent" {
				if err != nil || state.Kind != stateread.KindAbsent || len(state.Bytes) != 0 {
					t.Fatalf("absent: %+v %v", state, err)
				}
				return
			}
			if tc.want != "" {
				if err == nil || state.Kind != stateread.KindUnreadable || len(state.Bytes) != 0 || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("want %q, no bytes: %+v %v", tc.want, state, err)
				}
			} else if err != nil || state.Kind != stateread.KindPresent || !bytes.Equal(state.Bytes, payload) {
				t.Fatalf("valid read: %+v %v", state, err)
			}
			if tc.mode != 0 {
				got, err := os.ReadFile(target)
				info, statErr := os.Stat(target)
				if err != nil || statErr != nil || !bytes.Equal(got, payload) || !info.ModTime().Equal(time.Unix(1600000000, 0)) {
					t.Fatal("read altered native target")
				}
			}
		})
	}
}

type seedStatOverride struct {
	os.FileInfo
	stat syscall.Stat_t
}

func (s seedStatOverride) Sys() any { return &s.stat }

// Synthetic fstat ownership metadata exercises the production owner predicate
// without chown/root; the matching-uid control proves it does not reject all.
func TestNativeSeedForeignOwner(t *testing.T) {
	path := filepath.Join(t.TempDir(), "seed")
	if err := os.WriteFile(path, []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	uid := uint32(os.Geteuid())
	stat := *info.Sys().(*syscall.Stat_t)
	if err := checkSeedInfo(info, uid); err != nil {
		t.Fatal(err)
	}
	stat.Uid = uid + 1
	if err := checkSeedInfo(seedStatOverride{info, stat}, uid); err == nil || !strings.Contains(err.Error(), "not owned by the current user") {
		t.Fatalf("foreign uid: %v", err)
	}
}

// A pathname swap after open cannot supply the bytes or permissions used for
// an already-open descriptor. A final symlink swapped in before open refuses.
func TestNativeSeedOpenDescriptorBinding(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "seed")
	other := filepath.Join(root, "other")
	if err := os.WriteFile(path, []byte("checked bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(other, []byte("replacement bytes"), 0o622); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(other, 0o622); err != nil {
		t.Fatal(err)
	}
	state, err := readWithOpen(path, func(resolved string) (*os.File, error) {
		f, err := openSeed(resolved)
		if err != nil {
			return nil, err
		}
		if err := os.Rename(other, path); err != nil {
			_ = f.Close()
			return nil, err
		}
		return f, nil
	})
	if err != nil || string(state.Bytes) != "checked bytes" {
		t.Fatalf("read was not bound to open fd: %+v %v", state, err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	state, err = readWithOpen(path, func(resolved string) (*os.File, error) {
		if err := os.Rename(resolved, other); err != nil {
			return nil, err
		}
		if err := os.Symlink(other, resolved); err != nil {
			return nil, err
		}
		return openSeed(resolved)
	})
	if !errors.Is(err, syscall.ELOOP) || state.Kind != stateread.KindUnreadable || len(state.Bytes) != 0 {
		t.Fatalf("replacement link admitted: %+v %v", state, err)
	}
}

// A safe pre-open pathname cannot conceal a writable or non-regular open fd.
// The FIFO case also proves a raced FIFO can be opened without blocking.
func TestNativeSeedRejectsUnsafeOpenDescriptor(t *testing.T) {
	for _, fifo := range []bool{false, true} {
		t.Run(map[bool]string{false: "writable", true: "fifo"}[fifo], func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "seed")
			bad := filepath.Join(root, "bad")
			if err := os.WriteFile(path, []byte("valid"), 0o600); err != nil {
				t.Fatal(err)
			}
			want := "must not be group- or other-writable"
			if fifo {
				if err := unix.Mkfifo(bad, 0o600); err != nil {
					t.Fatal(err)
				}
				want = "must be a regular file"
			} else {
				if err := os.WriteFile(bad, []byte("forbidden"), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(bad, 0o622); err != nil {
					t.Fatal(err)
				}
			}
			state, err := readWithOpen(path, func(string) (*os.File, error) { return openSeed(bad) })
			if err == nil || !strings.Contains(err.Error(), want) || state.Kind != stateread.KindUnreadable || len(state.Bytes) != 0 {
				t.Fatalf("unsafe fd admitted: %+v %v", state, err)
			}
		})
	}
}
