package identity

import (
	"os"
	"path/filepath"
	"testing"
)

// TestClassifySourceMatrix is the classification matrix of the landed
// manager-config-v2 `$defs/overlay` discriminator. Every spelling the
// committed conformance corpus publishes appears here with the kind the
// corpus decides, alongside the decided edges the corpus states in prose.
// The corpus itself is driven through config.Load in internal/config; this
// table pins the helper directly so a misclassification names the arm.
func TestClassifySourceMatrix(t *testing.T) {
	cases := []struct {
		spelling string
		want     SourceKind
		why      string
	}{
		// --- git: an admitted `://` scheme, any letter case ---------------
		{"https://github.com/example/x", SourceGit, "https"},
		{"http://github.com/example/extra-context", SourceGit, "http"},
		{"ssh://github.com/example/personal-context", SourceGit, "ssh"},
		{"git://github.com/example/team-context", SourceGit, "git"},
		{"HTTPS://github.com/example/x", SourceGit, "uppercase https"},
		{"HTTP://github.com/example/extra-context", SourceGit, "uppercase http"},
		{"SSH://github.com/example/personal-context", SourceGit, "uppercase ssh"},
		{"GIT://github.com/example/team-context", SourceGit, "uppercase git"},
		{"HtTpS://github.com/example/x", SourceGit, "mixed-case https"},

		// --- git: SCP [user@]host:path on the §6.1 host grammar -----------
		{"git@github.com:example/team-context", SourceGit, "scp with user"},
		{"github.com:example/team-context", SourceGit, "scp without user"},
		{"c:example/team-context", SourceGit, "one-character §6.1 host is a host, not a drive"},
		{"github.com:a", SourceGit, "shortest scp path"},

		// --- path: no colon before the first slash or whitespace ----------
		{"/Users/operator/context", SourcePath, "absolute"},
		{"./packages/personal", SourcePath, "explicitly relative"},
		{"../packages/personal", SourcePath, "parent-relative"},
		{".", SourcePath, "the working directory"},
		{"..", SourcePath, "the parent directory"},
		{`.\packages\personal`, SourcePath, "windows relative"},
		{`\\host\share\pkg`, SourcePath, "UNC"},
		{"packages/team-context", SourcePath, "project-relative"},
		{"packages/team:context", SourcePath, "a colon in a later segment"},
		{"pkg", SourcePath, "a bare name"},
		{"github.com/example/x", SourcePath, "a bare canonical identity carries no colon"},

		// --- path: the Windows drive letter, carved out ahead of SCP ------
		{`C:\Users\operator\context`, SourcePath, "drive, backslash"},
		{"C:/Users/operator/context", SourcePath, "drive, slash"},
		{"C://Users/operator/context", SourcePath, "a one-character scheme is not a URL"},
		{`c:\users\operator\context`, SourcePath, "lowercase drive"},

		// --- refused: a `://` URL on an unsupported scheme ----------------
		{"file:///Users/operator/context", SourceInvalid, "file is not an admitted transport"},
		{"svn://github.com/example/x", SourceInvalid, "svn is not an admitted transport"},
		{"ftp://github.com/example/x", SourceInvalid, "ftp is not an admitted transport"},
		{"HTTPX://github.com/example/x", SourceInvalid, "https? does not admit httpx"},

		// --- refused: colon-bearing, and no valid §6.1 network form -------
		{"C:", SourceInvalid, "a bare drive letter names nothing"},
		{"file:", SourceInvalid, "neither a URL, a drive, nor an SCP remote"},
		{"my_host:example/x", SourceInvalid, "underscore is outside the §6.1 host grammar"},
		{"git@my_host:example/x", SourceInvalid, "same, with a user"},
		{`github.com:\example\x`, SourceInvalid, "a backslash first path character"},
		{"github.com:/example/x", SourceInvalid, "a slash first path character"},
		{"github.com: example/x", SourceInvalid, "a whitespace first path character"},
		{"-host:example/x", SourceInvalid, "a host may not start with a dash"},
		{"git@:example/x", SourceInvalid, "an empty host"},
		{"github.com:", SourceInvalid, "an empty path"},

		// --- refused: the empty spelling ----------------------------------
		{"", SourceInvalid, "an empty source is no source"},
	}
	for _, tc := range cases {
		if got := ClassifySource(tc.spelling); got != tc.want {
			t.Errorf("ClassifySource(%q) = %s, want %s (%s)", tc.spelling, got, tc.want, tc.why)
		}
	}
}

// TestClassifySourceNeverProbesTheFilesystem drives the discriminator from
// a working directory in which each git spelling that the platform can host
// as a file name also names a real directory. A classifier that stats the
// operand would call those `path`, hand them the network-allowlist bypass a
// `path` source has by design, and fail here. Every git spelling carries a
// colon, which Windows forbids in a path component, so on Windows no such
// directory can exist at all and the verdicts are asserted unplanted --
// that is the bound, and it is a bound in the safe direction.
func TestClassifySourceNeverProbesTheFilesystem(t *testing.T) {
	work := t.TempDir()
	t.Chdir(work)
	spellings := []string{
		"github.com:example/team-context",
		"c:example/team-context",
		"git@github.com:example/x",
	}
	planted := 0
	for _, spelling := range spellings {
		if err := os.MkdirAll(filepath.Join(work, filepath.FromSlash(spelling)), 0o755); err != nil {
			t.Logf("%q is not a hostable directory name on this platform: %v", spelling, err)
			continue
		}
		planted++
	}
	t.Logf("planted %d of %d git spellings as directories", planted, len(spellings))
	for _, spelling := range spellings {
		if got := ClassifySource(spelling); got != SourceGit {
			t.Errorf("ClassifySource(%q) = %s with a directory of that name present, want %s", spelling, got, SourceGit)
		}
	}
	// The other direction: a path spelling stays a path when nothing of
	// that name exists, so the verdict never depends on the filesystem in
	// either direction.
	if got := ClassifySource("packages/definitely-not-here"); got != SourcePath {
		t.Errorf("ClassifySource of an absent path = %s, want %s", got, SourcePath)
	}
}

// TestClassifySourceEcmaWhitespace pins the one place the Go transcription
// could drift from the ECMA-262 engine the committed schema is evaluated
// by: Go's own \s is [\t\n\f\r ], while ECMA-262 \s also covers the
// vertical tab, the Unicode space separators, and U+FEFF. A source whose
// colon sits behind one of those must classify exactly as the schema
// classifies it.
func TestClassifySourceEcmaWhitespace(t *testing.T) {
	// U+00A0 is ECMA-262 whitespace, so the colon after it is not "reached
	// without crossing whitespace": the colon rule does not fire and the
	// spelling falls through to `path`, exactly as `a b:c` does.
	for _, space := range []string{"\u00a0", "\u2028", "\u3000", "\ufeff", "\v"} {
		spelling := "a" + space + "b:c"
		if got := ClassifySource(spelling); got != SourcePath {
			t.Errorf("ClassifySource(%q) = %s, want %s", spelling, got, SourcePath)
		}
	}
	// The same spellings are refused when the whitespace follows the colon:
	// the SCP form's first path character may not be whitespace.
	for _, space := range []string{"\u00a0", "\u2028", "\u3000", "\ufeff", "\v"} {
		spelling := "github.com:" + space + "x"
		if got := ClassifySource(spelling); got != SourceInvalid {
			t.Errorf("ClassifySource(%q) = %s, want %s", spelling, got, SourceInvalid)
		}
	}
}

// TestParseOneCharacterHostIsANetworkIdentity pins the canonicalization side
// of the landed valid-overlay-git-single-letter-host case: core §6.1's host
// grammar is [A-Za-z0-9][A-Za-z0-9.-]*, which admits one character, so
// `c:example/x` is a network source and canonicalizes like any other SCP
// remote. The Windows drive spellings never reach that rule — they are local
// before it, in both the forward-slash and backslash forms — so widening the
// host grammar here does not turn a drive path into a remote.
func TestParseOneCharacterHostIsANetworkIdentity(t *testing.T) {
	network := map[string]string{
		"c:example/x":     "c/example/x",
		"C:example/x":     "c/example/x",
		"git@c:example/x": "c/example/x",
	}
	for spelling, want := range network {
		got, err := Parse(spelling)
		if err != nil || got != want {
			t.Errorf("Parse(%q) = %q, %v, want %q", spelling, got, err, want)
		}
	}
	for _, spelling := range []string{"C:/Users/operator/context", `C:\Users\operator\context`, `c:\users\operator\context`} {
		got, err := Parse(spelling)
		if err != nil || got != "" {
			t.Errorf("Parse(%q) = %q, %v, want the local verdict", spelling, got, err)
		}
	}
}
