package scriptworker

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"testing"
)

//go:embed testdata/executable_identity_cases.json
var pinnedExecutableIdentityFixture embed.FS

const pinnedExecutableIdentityFixtureSHA256 = "sha256:124e00757b3add8c2ba639a8403cba97eec7f5d1bdd923d14b51db9a104292a2"

type pinnedExecutableIdentityCase = executableIdentityCase

// TestUncapturedSystemRootHardlinkRegression proves an ambient SYSTEMROOT
// cannot authorize the System32 hard-link exception. The exhaustive family
// driver accounts for the other published cases, without known gaps.
func loadPinnedExecutableIdentityCases(t *testing.T) []executableIdentityCase {
	t.Helper()
	payload, err := pinnedExecutableIdentityFixture.ReadFile("testdata/executable_identity_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(payload)
	if got := "sha256:" + hex.EncodeToString(digest[:]); got != pinnedExecutableIdentityFixtureSHA256 {
		t.Fatalf("executable identity fixture digest = %s, want %s", got, pinnedExecutableIdentityFixtureSHA256)
	}
	var cases []pinnedExecutableIdentityCase
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 8 {
		t.Fatalf("executable identity case count = %d, want 8", len(cases))
	}
	return cases
}

func TestUncapturedSystemRootHardlinkRegression(t *testing.T) {
	cases := loadPinnedExecutableIdentityCases(t)
	var target *pinnedExecutableIdentityCase
	for i := range cases {
		if cases[i].Name != "windows-exec-uncaptured-systemroot-hardlinks" {
			continue
		}
		if target != nil {
			t.Fatal("pinned identity fixture repeats the uncaptured-SystemRoot case")
		}
		target = &cases[i]
	}
	if target == nil {
		t.Fatal("pinned identity fixture omits the uncaptured-SystemRoot case")
	}
	if target.Platform != "windows" || target.Use != "declared-exec-name" ||
		target.Accepted || target.SystemRoot == nil || *target.SystemRoot != "caller-or-package-value" {
		t.Fatalf("uncaptured-SystemRoot case has unexpected contract: %+v", *target)
	}
	if failure := driveDeclaredExecIdentityCase(t, *target); failure != "" {
		t.Fatal(failure)
	}
}
