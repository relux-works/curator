package envprofile

import (
	"os"
	"path/filepath"
	"testing"
)

// Profile path-source behavior follows environments §1 and remains outside
// the project Skillfile source protocol.

func writeSeedPackage(t *testing.T, root, name string) {
	t.Helper()
	writeManifestPackage(t, root,
		`{"schema_version": 1, "name": "`+name+`", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": name + "\n"})
}

// TestPathInstallAdmitsOrdinaryDirectory proves ordinary profile paths remain
// available through the environment context store.
func TestPathInstallAdmitsOrdinaryDirectory(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeSeedPackage(t, root, "acme")
	info, _, _, err := Install(home, InstallOptions{Operand: root})
	if err != nil {
		t.Fatalf("profile install of an ordinary path: %v", err)
	}
	if info.Name != "acme" {
		t.Fatalf("profile %q, want acme", info.Name)
	}
	if _, err := os.Stat(lockPath(home, "acme")); err != nil {
		t.Fatalf("lock for the admitted install: %v", err)
	}
}

// TestLegacyPathInstallIgnoresStoreOverlap pins the environment profile
// contract: profile path sources use the environment context store and do
// not inherit project Skillfile source admission rules.
func TestLegacyPathInstallIgnoresStoreOverlap(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	seed := filepath.Join(home, "contexts", "operand-seed")
	writeSeedPackage(t, seed, "seeded")
	info, _, _, err := Install(home, InstallOptions{Operand: seed})
	if err != nil {
		t.Fatalf("profile install from a context-store path: %v", err)
	}
	if info.Name != "seeded" {
		t.Fatalf("profile %q, want seeded", info.Name)
	}
}

// TestPathOverlayAdmitsOrdinaryDirectory proves path overlays keep joining
// the profile closure through the environment context store.
func TestPathOverlayAdmitsOrdinaryDirectory(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeSeedPackage(t, root, "acme")
	overlay := filepath.Join(t.TempDir(), "overlay")
	writeSeedPackage(t, overlay, "personal")
	policy := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays:             map[string][]OverlaySpec{"acme": {{Source: overlay}}},
	}
	info, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy})
	if err != nil {
		t.Fatalf("path install with an ordinary overlay: %v", err)
	}
	member, ok := lockMember(info.Lock, "personal")
	if !ok || !member.Overlay {
		t.Fatalf("lock members %+v carry no overlay", info.Lock.Members)
	}
}
