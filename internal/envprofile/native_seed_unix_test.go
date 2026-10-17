//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package envprofile

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/envregistry"
	"golang.org/x/sys/unix"
)

// Invalid linked seeds refuse with the seed diagnostic before provisioning and
// leave existing target bytes/mtime intact. The reader's unit cases pin each
// underlying type/mode/chain refusal independently of this integration wrapper.
func TestNativeSeedRefusals(t *testing.T) {
	for _, kind := range []string{"dangling", "directory", "fifo", "group-write", "other-write", "loop"} {
		t.Run(kind, func(t *testing.T) {
			fx := writeManagedFixture(t, "seed-refusal")
			path := filepath.Join(fx.native[envregistry.CodexCLI], "config.toml")
			target := filepath.Join(filepath.Dir(path), "target")
			payload := []byte("model = 'fixture'\n")
			var err error
			switch kind {
			case "directory":
				err = os.Mkdir(target, 0o700)
			case "fifo":
				err = unix.Mkfifo(target, 0o600)
			case "loop":
				err = os.Symlink("config.toml", target)
			case "group-write", "other-write":
				err = os.WriteFile(target, payload, 0o600)
				if err == nil {
					mode := os.FileMode(0o620)
					if kind == "other-write" {
						mode = 0o602
					}
					err = os.Chmod(target, mode)
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink("target", path); err != nil {
				t.Fatal(err)
			}
			var before os.FileInfo
			if kind == "group-write" || kind == "other-write" {
				before, err = os.Stat(target)
				if err != nil {
					t.Fatal(err)
				}
			}
			req := fx.request(envregistry.CodexCLI)
			req.Repair = true
			result, err := Resolve(req)
			if err == nil || result != nil || !strings.Contains(err.Error(), envregistry.DiagSeedUnreadable) {
				t.Fatalf("want seed refusal: %+v %v", result, err)
			}
			t.Logf("intended seed refusal: %v", err)
			if _, err := os.Lstat(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI)); !os.IsNotExist(err) {
				t.Fatalf("refusal published a home: %v", err)
			}
			if got, err := os.Readlink(path); err != nil || got != "target" {
				t.Fatal("refusal changed native link")
			}
			if before != nil {
				got, err := os.ReadFile(target)
				after, statErr := os.Stat(target)
				if err != nil || statErr != nil || !bytes.Equal(got, payload) || !before.ModTime().Equal(after.ModTime()) {
					t.Fatal("refusal changed target bytes/mtime")
				}
			}
		})
	}
}

// Initial provisioning reads a linked seed through both credential inspection
// and copy gathering; later resolve/repair preserve native bytes and mtime.
func TestNativeSeedSymlinkProvision(t *testing.T) {
	fx := writeManagedFixture(t, "seed-provision")
	seedLiveNativeCredentials(t, fx)
	path := filepath.Join(fx.native[envregistry.CodexCLI], "config.toml")
	target := filepath.Join(filepath.Dir(path), "target")
	payload := []byte("model = 'fixture'\ncli_auth_credentials_store = 'file'\n")
	if err := os.WriteFile(target, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	stamp := time.Unix(1600000000, 0)
	if err := os.Chtimes(target, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("target", path); err != nil {
		t.Fatal(err)
	}
	req := fx.request(envregistry.CodexCLI)
	req.Repair = true
	initial, err := Resolve(req)
	if err != nil || initial == nil || !initial.Provisioned {
		t.Fatalf("linked initial provisioning: %+v %v", initial, err)
	}
	managed := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.CodexCLI), "config.toml")
	got, err := os.ReadFile(managed)
	if err != nil || !bytes.Equal(got, payload) {
		t.Fatalf("linked seed copy: %q %v", got, err)
	}
	for _, repair := range []bool{false, true} {
		req.Repair = repair
		result, err := Resolve(req)
		if err != nil || result == nil || !bytes.Equal(result.Document, initial.Document) {
			t.Fatalf("post-provision resolve/repair: %+v %v", result, err)
		}
	}
	got, err = os.ReadFile(target)
	info, statErr := os.Stat(target)
	if err != nil || statErr != nil || !bytes.Equal(got, payload) || !info.ModTime().Equal(stamp) {
		t.Fatal("provision/resolve/repair wrote native target")
	}
}
