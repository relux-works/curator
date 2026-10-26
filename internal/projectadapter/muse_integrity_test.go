package projectadapter

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Claim: an external hardlink to either approved context or a nested skill file
// is drift, while the same allocation before linking and after unlinking verifies.
// Rejection neither rewrites the approved file nor changes its retained receipt.
func TestMuseOfflineHardlinks(t *testing.T) {
	for _, name := range []string{"config/muse/AGENTS.md", "data/muse/skills/helper/scripts/tool"} {
		t.Run(name, func(t *testing.T) {
			in := offlineFixture()
			home := offlineHome(t)
			got, err := MaterializeMuseOffline(context.Background(), home, in)
			if err != nil {
				t.Fatal(err)
			}
			if err = VerifyMuseOffline(home, got.Generated); err != nil {
				t.Fatal("valid neighbour:", err)
			}
			target := filepath.Join(home, filepath.FromSlash(name))
			external := filepath.Join(filepath.Dir(home), "external-link")
			if err = os.Link(target, external); err != nil {
				t.Fatal(err)
			}
			if err = VerifyMuseOffline(home, got.Generated); err == nil || !strings.Contains(err.Error(), "project_context_drift: offline hardlink detected") {
				t.Fatalf("hardlink was not rejected: %v", err)
			}
			data, err := os.ReadFile(target)
			if err != nil || !bytes.Equal(data, in.Files[name]) || got.Generated[name] != in.Generated[name] {
				t.Fatal("rejection changed protected output")
			}
			if err = os.Remove(external); err != nil {
				t.Fatal(err)
			}
			if err = VerifyMuseOffline(home, got.Generated); err != nil {
				t.Fatal("unlinked neighbour:", err)
			}
		})
	}
}

// Claim: every XDG parent must remain present/private and every nested directory
// must retain mode 0700; verification rejects drift without repairing it.
// Empty extra private directories remain accepted, as in the original verifier.
func TestMuseOfflineDirectoryIntegrity(t *testing.T) {
	for _, tc := range []struct{ name, dir, action string }{
		{"missing-state", "state", "remove"}, {"missing-cache", "cache", "remove"},
		{"state-mode", "state", "mode"}, {"cache-mode", "cache", "mode"},
		{"nested-mode", "data/muse/skills/helper/scripts", "mode"}, {"extra-mode", "extra", "mode"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := offlineHome(t)
			got, err := MaterializeMuseOffline(context.Background(), home, offlineFixture())
			if err != nil {
				t.Fatal(err)
			}
			if err = os.Mkdir(filepath.Join(home, "extra"), 0700); err != nil {
				t.Fatal(err)
			}
			if err = VerifyMuseOffline(home, got.Generated); err != nil {
				t.Fatal("valid private neighbour:", err)
			}
			dir := filepath.Join(home, filepath.FromSlash(tc.dir))
			reason := "offline directory mode differs"
			if tc.action == "remove" {
				err = os.Remove(dir)
				reason = "offline required parent unavailable"
			} else {
				err = os.Chmod(dir, 0755)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err = VerifyMuseOffline(home, got.Generated); err == nil || !strings.Contains(err.Error(), "project_context_drift: "+reason) {
				t.Fatalf("directory drift was not rejected: %v", err)
			}
			info, statErr := os.Lstat(dir)
			if tc.action == "remove" {
				if !os.IsNotExist(statErr) {
					t.Fatal("verifier recreated parent")
				}
				err = os.Mkdir(dir, 0700)
			} else {
				if statErr != nil || info.Mode().Perm() != 0755 {
					t.Fatal("verifier repaired mode")
				}
				err = os.Chmod(dir, 0700)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err = VerifyMuseOffline(home, got.Generated); err != nil {
				t.Fatal("restored neighbour:", err)
			}
		})
	}
}
