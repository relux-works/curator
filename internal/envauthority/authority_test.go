package envauthority

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/privatedir"
)

// All registered managed layouts (including nested OpenCode) and orphaned
// profiles are checked. Full valid per-home carriers remain admitted; full node
// carriers refuse with the normative diagnostic without changing marker bytes.
func TestManagedAuthorityInventory(t *testing.T) {
	raw, err := os.ReadFile("../envmarker/testdata/seed-c/node.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, adapter := range envregistry.Registry {
		for _, node := range []bool{true, false} {
			t.Run(adapter.ID+map[bool]string{true: "/node", false: "/per-home"}[node], func(t *testing.T) {
				home := t.TempDir()
				dir := filepath.Join(home, "environments", "orphan", adapter.ID)
				if adapter.ID == envregistry.OpenCode {
					dir = filepath.Join(dir, "opencode")
				}
				if err := privatedir.MakeAll(dir); err != nil {
					t.Fatal(err)
				}
				payload := raw
				if !node {
					var doc map[string]any
					if err := json.Unmarshal(raw, &doc); err != nil {
						t.Fatal(err)
					}
					doc["credential_source"] = map[string]any{"mode": "per-home"}
					doc["seeds"] = []string{"config.toml"}
					payload, err = json.Marshal(doc)
					if err != nil {
						t.Fatal(err)
					}
				}
				if _, err := envmarker.Parse(payload); err != nil {
					t.Fatalf("full carrier setup: %v", err)
				}
				path := filepath.Join(dir, envmarker.Name)
				file, err := privatedir.Create(path, 0600)
				if err != nil {
					t.Fatal(err)
				}
				_, writeErr := file.Write(payload)
				closeErr := file.Close()
				if writeErr != nil || closeErr != nil {
					t.Fatalf("private marker setup: %v %v", writeErr, closeErr)
				}
				// Shared creators must establish private objects before the
				// authority verdict, independently of the parent temp ACL.
				for current := dir; current != home; current = filepath.Dir(current) {
					if err := privatedir.Validate(current); err != nil {
						t.Fatalf("private managed directory setup: %v", err)
					}
				}
				if err := privatedir.ValidateFile(path); err != nil {
					t.Fatalf("private managed marker setup: %v", err)
				}
				err = Validate(home)
				if node {
					if err == nil || !strings.HasPrefix(err.Error(), "credential_injection_unavailable:") {
						t.Fatalf("node diagnostic: %v", err)
					}
				} else if err != nil {
					t.Fatalf("per-home positive: %v", err)
				}
				after, e := os.ReadFile(path)
				if e != nil || string(after) != string(payload) {
					t.Fatal("authority check changed marker")
				}
			})
		}
	}
	// A home with no environments is a valid first-provisioning input.
	if err := Validate(filepath.Join(t.TempDir(), "absent")); err != nil {
		t.Fatal(err)
	}
}

// Unsafe and malformed authority cannot authorize mutation. Neighboring valid
// per-home carriers above prevent an always-refuse gate satisfying this check.
func TestManagedAuthorityBoundary(t *testing.T) {
	raw, err := os.ReadFile("../envmarker/testdata/seed-c/node.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, variant := range []string{"root-link", "profile-link", "marker-link", "writable", "malformed"} {
		t.Run(variant, func(t *testing.T) {
			home := t.TempDir()
			root := filepath.Join(home, "environments")
			dir := filepath.Join(root, "example", envregistry.CodexCLI)
			if err := privatedir.MakeAll(dir); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, envmarker.Name)
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			want := "environment_store_untrusted:"
			switch variant {
			case "root-link", "profile-link", "marker-link":
				selected := path
				if variant == "root-link" {
					selected = root
				}
				if variant == "profile-link" {
					selected = filepath.Dir(dir)
				}
				saved := filepath.Join(t.TempDir(), "saved")
				if err := os.Rename(selected, saved); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(saved, selected); err != nil {
					t.Skipf("symlink fixture unavailable: %v", err)
				}
			case "writable":
				if runtime.GOOS == "windows" {
					t.Skip("POSIX mode-bit contract")
				}
				if err := os.Chmod(path, 0666); err != nil {
					t.Fatal(err)
				}
			case "malformed":
				if err := os.WriteFile(path, []byte("{"), 0600); err != nil {
					t.Fatal(err)
				}
				want = "environment_marker_unreadable:"
			}
			err := Validate(home)
			if err == nil || !strings.HasPrefix(err.Error(), want) {
				t.Fatalf("authority boundary diagnostic: %v; want %s", err, want)
			}
		})
	}
}
