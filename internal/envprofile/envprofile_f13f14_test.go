package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/identity"
)

// Production entry points under test: Install, Current, Use.
//
// F13 covers install activation as a §9.2 switch: first install and
// InstallOptions.Use must attempt every entry, record the new current only
// when the whole scope materialized, and report profile_use_partial when it
// did not. A mutant that only moves the current pointer (the pre-fix shape)
// or that drops materialization for one adapter must fail these tests.
//
// F14 covers the syntactic path-vs-git distinction (§9.1). The kind of an
// install operand is decided by spelling alone through the one
// discriminator (identity.ClassifySource) and is never probed from the
// filesystem, so a directory planted in the operator's working directory
// can never shadow a git identity and carry it past the core §6.1 network
// allowlist. A mutant that stats the operand to decide must fail these
// tests.

func TestInstallUseSwitchesAndAgrees(t *testing.T) {
	home := t.TempDir()
	homes := pinHomes(t)
	first := t.TempDir()
	writePackage(t, first, "alpha", "1.0.0", "alpha\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: first}); err != nil {
		t.Fatalf("install alpha: %v", err)
	}
	second := t.TempDir()
	writePackage(t, second, "beta", "1.0.0", "beta\n")
	info, activated, _, err := Install(home, InstallOptions{Operand: second, Use: true})
	if err != nil {
		t.Fatalf("install beta --use: %v", err)
	}
	if !activated {
		t.Fatal("install --use must report activated")
	}
	if len(info.Activation) != len(Adapters) {
		t.Fatalf("activation results = %d, want %d", len(info.Activation), len(Adapters))
	}
	for _, result := range info.Activation {
		if !result.OK {
			t.Fatalf("activation %s: %s", result.Adapter, result.Detail)
		}
	}
	current, err := Current(home)
	if err != nil {
		t.Fatal(err)
	}
	if current != "beta" {
		t.Fatalf("current = %q, want beta", current)
	}
	payload, err := os.ReadFile(filepath.Join(homes["claude_code"], "CLAUDE.md")) // #nosec G304 -- test home
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), "## Context: beta 1.0.0") {
		t.Fatalf("CLAUDE.md does not carry beta:\n%s", payload)
	}
	marker, err := envmarker.Read(homes["claude_code"])
	if err != nil || marker == nil || marker.Profile.Name != "beta" {
		t.Fatalf("marker %+v %v", marker, err)
	}
}

// TestInstallUsePartialLeavesCurrent forces one adapter to fail and checks
// the §9.2 partial shape through Install: the scope is attempted, the
// recorded current is unchanged, and the error carries profile_use_partial.
// A mutant that only moves the pointer (no materialization, no failure)
// succeeds the install and moves current, and must fail this test.
func TestInstallUsePartialLeavesCurrent(t *testing.T) {
	home := t.TempDir()
	homes := pinHomes(t)
	first := t.TempDir()
	writePackage(t, first, "alpha", "1.0.0", "alpha\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: first}); err != nil {
		t.Fatalf("install alpha: %v", err)
	}
	// Break one adapter home: replace its directory with a regular file so
	// MkdirAll inside materializeOne fails for exactly that entry.
	claude := homes["claude_code"]
	if err := os.RemoveAll(claude); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(claude, []byte("blocker"), 0o644); err != nil {
		t.Fatal(err)
	}
	second := t.TempDir()
	writePackage(t, second, "beta", "1.0.0", "beta\n")
	info, activated, _, err := Install(home, InstallOptions{Operand: second, Use: true})
	if err == nil || !strings.Contains(err.Error(), DiagUsePartial) {
		t.Fatalf("install beta --use err = %v, want %s", err, DiagUsePartial)
	}
	if activated {
		t.Fatal("partial activation must not report activated")
	}
	failed := false
	for _, result := range info.Activation {
		if !result.OK {
			failed = true
		}
	}
	if !failed {
		t.Fatalf("activation %+v has no failing entry", info.Activation)
	}
	current, err := Current(home)
	if err != nil {
		t.Fatal(err)
	}
	if current != "alpha" {
		t.Fatalf("current = %q, want alpha (unchanged)", current)
	}
}

// TestOperandShadowedByDirectoryResolvesAsGit checks F14 through Install: an
// operand whose spelling is a git identity resolves as git even when a
// directory of that name exists in the working directory, so the locked
// machine allowlist still refuses it. A mutant that stats the operand
// classifies it as path, bypasses the allowlist, and installs the planted
// bytes — and must fail this test.
func TestOperandShadowedByDirectoryResolvesAsGit(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	work := t.TempDir()
	planted := filepath.Join(work, "github.com", "evil-org", "pkg")
	if err := os.MkdirAll(planted, 0o755); err != nil {
		t.Fatal(err)
	}
	writePackage(t, planted, "planted", "9.9.9", "PLANTED LOCAL CONTENT\n")
	t.Chdir(work)
	policy := Policy{AllowedSources: []string{"github.com/relux-works"}}
	_, _, _, err := Install(home, InstallOptions{Operand: "github.com/evil-org/pkg", Policy: policy})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
		t.Fatalf("shadowed operand err = %v, want %s", err, DiagSourceInvalid)
	}
	if !strings.Contains(err.Error(), "allowed sources") {
		t.Fatalf("shadowed operand err = %v, want the allowlist reason", err)
	}
	if _, err := readSource(home, "planted"); err == nil {
		t.Fatal("refused install must leave no profile record")
	}
	if entries, readErr := os.ReadDir(filepath.Join(home, "profile-repos")); readErr == nil && len(entries) != 0 {
		t.Fatalf("refused install cloned: %v", entries)
	}
}

// TestAbsentPathOperandWithRequirementIsRefConflict checks F14 through
// Install: a syntactic path operand that does not exist is still a path, so
// a requirement flag on it is profile_install_ref_conflict — it must never
// reach a git clone. A stat-based classifier routes both to git and reports
// a clone failure instead, and must fail this test.
func TestAbsentPathOperandWithRequirementIsRefConflict(t *testing.T) {
	pinHomes(t)
	work := t.TempDir()
	t.Chdir(work)
	absent := filepath.Join(work, "definitely-not-here")
	cases := []InstallOptions{
		{Operand: absent, Range: "^1.0.0"},
		{Operand: "./missing-relative-pkg", Tag: "v1.0.0"},
	}
	for _, options := range cases {
		home := t.TempDir()
		_, _, _, err := Install(home, options)
		if err == nil || !strings.Contains(err.Error(), DiagRefConflict) {
			t.Fatalf("operand %q err = %v, want %s", options.Operand, err, DiagRefConflict)
		}
		if entries, readErr := os.ReadDir(filepath.Join(home, "profile-repos")); readErr == nil && len(entries) != 0 {
			t.Fatalf("operand %q cloned: %v", options.Operand, entries)
		}
	}
}

// installOperandCases is the `profile install <git-url|path>` operand
// matrix. The kind comes from the one discriminator
// (identity.ClassifySource) plus this row's stated widening: a spelling the
// discriminator calls `path` that is also a valid core §6.1 canonical
// network identity stays `git`, because `curator profile install
// github.com/example/x` clones it over https today and following the
// classification would silently turn a network install into a local one.
var installOperandCases = []struct {
	operand string
	want    identity.SourceKind
	why     string
}{
	{"/tmp/pkg", identity.SourcePath, "absolute"},
	{"/", identity.SourcePath, "the root"},
	{"./pkg", identity.SourcePath, "explicitly relative"},
	{"../pkg", identity.SourcePath, "parent-relative"},
	{".", identity.SourcePath, "the working directory"},
	{"..", identity.SourcePath, "the parent directory"},
	{`.\pkg`, identity.SourcePath, "windows relative"},
	{`..\pkg`, identity.SourcePath, "windows parent-relative"},
	{`C:/pkg`, identity.SourcePath, "drive, slash"},
	{`C:\pkg`, identity.SourcePath, "drive, backslash"},
	{`\\host\share\pkg`, identity.SourcePath, "UNC"},
	{"~/pkg", identity.SourcePath, "home-relative carries no network identity"},
	{"pkg", identity.SourcePath, "a bare name carries no network identity"},
	{"MyOrg/pkg", identity.SourcePath, "an uppercase host is not a canonical identity"},
	{"github.com/evil-org/pkg", identity.SourceGit, "a canonical identity stays git and stays allowlisted"},
	{"github.com/relux-works/pkg", identity.SourceGit, "same"},
	{"example.com/org/pkg", identity.SourceGit, "same"},
	{"https://example.com/org/pkg", identity.SourceGit, "https"},
	{"ssh://git@example.com/org/pkg", identity.SourceGit, "ssh"},
	{"git@example.com:org/pkg.git", identity.SourceGit, "scp"},
	{"c:example/pkg", identity.SourceGit, "an SCP remote on a one-character §6.1 host"},
	{"D:", identity.SourceInvalid, "a bare drive letter names nothing"},
	{"", identity.SourceInvalid, "an empty operand is no operand"},
	{"my_host:example/pkg", identity.SourceInvalid, "outside the §6.1 host grammar"},
	{"github.com:/org/pkg", identity.SourceInvalid, "a slash first path character is not an SCP remote"},
	{"svn://example.com/org/pkg", identity.SourceInvalid, "an unsupported scheme"},
	{"file:///tmp/pkg", identity.SourceInvalid, "file carries no network identity"},
}

// TestInstallOperandKindIsSyntactic pins the install classifier table: the
// kind is decided by spelling alone. A mutant that re-admits the stat
// fallback fails the git rows whenever the test working directory contains
// such a name (TestOperandShadowedByDirectoryResolvesAsGit drives that
// through Install).
func TestInstallOperandKindIsSyntactic(t *testing.T) {
	for _, tc := range installOperandCases {
		if got := installOperandKind(tc.operand); got != tc.want {
			t.Errorf("installOperandKind(%q) = %s, want %s (%s)", tc.operand, got, tc.want, tc.why)
		}
	}
}

// TestInstallNeverDemotesANetworkIdentityToAPath is the invariant behind the
// widening above, stated as a property over the same matrix: an operand that
// carries a core §6.1 canonical network identity is never classified `path`,
// because a `path` source bypasses the network allowlist by design. A mutant
// that drops the widening and follows the overlay classification verbatim
// turns `github.com/evil-org/pkg` into a path install and fails here.
func TestInstallNeverDemotesANetworkIdentityToAPath(t *testing.T) {
	checked := 0
	for _, tc := range installOperandCases {
		canonical, err := identity.Parse(strings.TrimSpace(tc.operand))
		if err != nil || canonical == "" {
			continue
		}
		checked++
		if got := installOperandKind(tc.operand); got == identity.SourcePath {
			t.Errorf("operand %q carries network identity %q but classified %s", tc.operand, canonical, got)
		}
	}
	if checked == 0 {
		t.Fatal("the matrix carries no network identity: the invariant checked nothing")
	}
	t.Logf("checked %d network-identity operands", checked)
}

// TestInstallRefusesANonSourceOperand drives Install with each refused
// spelling: the operand never reaches a clone and never becomes a local
// install of whatever the spelling happens to name.
func TestInstallRefusesANonSourceOperand(t *testing.T) {
	pinHomes(t)
	for _, tc := range installOperandCases {
		if tc.want != identity.SourceInvalid {
			continue
		}
		t.Run(tc.operand, func(t *testing.T) {
			home := t.TempDir()
			_, _, _, err := Install(home, InstallOptions{Operand: tc.operand})
			if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) {
				t.Fatalf("operand %q err = %v, want %s", tc.operand, err, DiagSourceInvalid)
			}
			// The refusal must come from the kind gate, not from a
			// downstream clone or manifest failure carrying the same
			// diagnostic.
			if !strings.Contains(err.Error(), "neither a git source nor a path") &&
				!strings.Contains(err.Error(), "no network identity") {
				t.Fatalf("operand %q err = %v, want the source-kind refusal", tc.operand, err)
			}
			if entries, readErr := os.ReadDir(filepath.Join(home, "profile-repos")); readErr == nil && len(entries) != 0 {
				t.Fatalf("refused operand %q cloned: %v", tc.operand, entries)
			}
		})
	}
}
