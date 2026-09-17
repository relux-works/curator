package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Production data-flow proof for the draft path-source boundary gate: every
// test drives the public Install entry with the draft switch explicitly on
// or off. The switch-off tests pin the legacy semantics byte-identically;
// the switch-on tests prove the admission gate runs before the store
// traverses anything.

func writeSeedPackage(t *testing.T, root, name string) {
	t.Helper()
	writeManifestPackage(t, root,
		`{"schema_version": 1, "name": "`+name+`", "version": "1.0.0",`+
			`"context": {"modules": [{"path": "a.md"}]}}`+"\n",
		map[string]string{"a.md": name + "\n"})
}

// TestLegacyPathInstallIgnoresStoreOverlap pins the legacy contract: with
// the draft switch off (the zero Policy, the only behavior a machine
// configuration produces today), a path operand inside the context store
// installs exactly as before. The draft gate would refuse this operand,
// so success here proves the gate is fully out of the legacy path.
func TestLegacyPathInstallIgnoresStoreOverlap(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	seed := filepath.Join(home, "contexts", "operand-seed")
	writeSeedPackage(t, seed, "seeded")
	info, _, _, err := Install(home, InstallOptions{Operand: seed})
	if err != nil {
		t.Fatalf("legacy install of a store-internal path: %v", err)
	}
	if info.Name != "seeded" {
		t.Fatalf("profile %q, want seeded", info.Name)
	}
}

// TestDraftPathInstallAdmitsOrdinaryDirectory proves the draft gate does
// not false-refuse: an ordinary directory installs with the switch on.
func TestDraftPathInstallAdmitsOrdinaryDirectory(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeSeedPackage(t, root, "acme")
	info, _, _, err := Install(home, InstallOptions{Operand: root, Policy: Policy{DraftSourcesV1: true}})
	if err != nil {
		t.Fatalf("draft install of an ordinary path: %v", err)
	}
	if info.Name != "acme" {
		t.Fatalf("profile %q, want acme", info.Name)
	}
	if _, err := os.Stat(lockPath(home, "acme")); err != nil {
		t.Fatalf("lock for the admitted install: %v", err)
	}
}

// TestDraftPathInstallRefusesStoreOverlap proves the gate runs before the
// store traversal on the production path-install entry: the same operand
// the legacy test installs is refused with the spec section 5 diagnostic
// leading, and nothing is published.
func TestDraftPathInstallRefusesStoreOverlap(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	seed := filepath.Join(home, "contexts", "operand-seed")
	writeSeedPackage(t, seed, "seeded")
	_, _, _, err := Install(home, InstallOptions{Operand: seed, Policy: Policy{DraftSourcesV1: true}})
	if err == nil || !strings.HasPrefix(err.Error(), "source_output_overlap") {
		t.Fatalf("draft install of a store-internal path err = %v, want leading source_output_overlap", err)
	}
	if _, statErr := os.Stat(lockPath(home, "seeded")); !os.IsNotExist(statErr) {
		t.Fatalf("refused install published a lock: %v", statErr)
	}
	if _, statErr := os.Stat(sourcePath(home, "seeded")); !os.IsNotExist(statErr) {
		t.Fatalf("refused install published a source record: %v", statErr)
	}
}

// TestDraftPathOverlayAdmitsOrdinaryDirectory proves path overlays keep
// joining the closure with the switch on.
func TestDraftPathOverlayAdmitsOrdinaryDirectory(t *testing.T) {
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
		DraftSourcesV1:       true,
	}
	info, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy})
	if err != nil {
		t.Fatalf("draft install with an ordinary overlay: %v", err)
	}
	member, ok := lockMember(info.Lock, "personal")
	if !ok || !member.Overlay {
		t.Fatalf("lock members %+v carry no overlay", info.Lock.Members)
	}
}

// TestDraftPathOverlayRefusesStoreOverlap proves the overlay funnel runs
// the same gate: an overlay declared inside the profile store refuses the
// whole install with the boundary diagnostic.
func TestDraftPathOverlayRefusesStoreOverlap(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeSeedPackage(t, root, "acme")
	overlay := filepath.Join(home, "profiles", "operand-seed")
	writeSeedPackage(t, overlay, "personal")
	policy := Policy{
		OverlaysAllowed:      true,
		OverlayDefaultWeight: 1000,
		Overlays:             map[string][]OverlaySpec{"acme": {{Source: overlay}}},
		DraftSourcesV1:       true,
	}
	_, _, _, err := Install(home, InstallOptions{Operand: root, Policy: policy})
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("draft install with a store-internal overlay err = %v, want source_output_overlap", err)
	}
	if _, statErr := os.Stat(lockPath(home, "acme")); !os.IsNotExist(statErr) {
		t.Fatalf("refused install published a lock: %v", statErr)
	}
}

// TestAdmitPathSourceRefusesManagerHome pins the owning-root mapping: the
// manager home itself selects the owning root, which would need operator
// root_inputs that profiles do not carry, so it is refused. The
// production funnel (Install through stateForPath) is proven by the
// overlap tests above; this pins the root-inputs branch of the mapping.
func TestAdmitPathSourceRefusesManagerHome(t *testing.T) {
	home := t.TempDir()
	_, err := admitPathSource(home, "acme", home)
	if err == nil || !strings.HasPrefix(err.Error(), "source_output_overlap") {
		t.Fatalf("admitPathSource(home) err = %v, want leading source_output_overlap", err)
	}
}
