package identity

import "testing"

func TestCanonical(t *testing.T) {
	cases := []struct {
		url  string
		want string
	}{
		// ssh and https of one repository share one identity
		{"git@git.example.com:skills/skill-a.git", "git.example.com/skills/skill-a"},
		{"https://git.example.com/skills/skill-a.git", "git.example.com/skills/skill-a"},
		{"ssh://git@git.example.com/skills/skill-a.git", "git.example.com/skills/skill-a"},
		{"https://GIT.Example.COM/Skills/Skill-A.git", "git.example.com/Skills/Skill-A"},
		{"git://host/path", "host/path"},
		{"http://host/a/b/", "host/a/b"},
		{"user@host.io:group/repo", "host.io/group/repo"},
		// local sources: no identity
		{"/abs/path/repo", ""},
		{"./relative", ""},
		{"../up", ""},
		{"~/home/repo", ""},
		{"file:///abs/repo", ""},
		{`C:\repos\skill`, ""},
		{"C:/repos/skill", ""},
		{"", ""},
		{"   ", ""},
		{"just-a-name", ""},
	}
	for _, tc := range cases {
		if got := Canonical(tc.url); got != tc.want {
			t.Errorf("Canonical(%q) = %q, want %q", tc.url, got, tc.want)
		}
	}
}

func TestMatchesPrefix(t *testing.T) {
	cases := []struct {
		identity, prefix string
		want             bool
	}{
		{"h/skills/x", "h/skills", true},
		{"h/skills", "h/skills", true},
		{"h/skills-evil", "h/skills", false}, // segment aware
		{"h/skills/x", "h/skills/", true},    // trailing slash trimmed
		{"h/skills/x", "", false},
		{"h/skills/x", "  ", false},
		{"other/skills/x", "h/skills", false},
	}
	for _, tc := range cases {
		if got := MatchesPrefix(tc.identity, tc.prefix); got != tc.want {
			t.Errorf("MatchesPrefix(%q, %q) = %v, want %v", tc.identity, tc.prefix, got, tc.want)
		}
	}
}

func TestParseRejectsAmbiguousNetworkForms(t *testing.T) {
	for _, source := range []string{
		"https://git.example.com:8443/skills/a",
		"https://git.example.com/skills%2Fa",
		"https://git.example.com/skills/a?q=1",
		"ftp://git.example.com/skills/a",
		"git@example.com:CON/repo",
		"git@git.example.com:skills/a b",
		"git@git.example.com:skills/a#fragment",
	} {
		if _, err := Parse(source); err == nil {
			t.Errorf("Parse(%q) must fail", source)
		}
	}
	if identity, err := Parse("file:///tmp/repo"); err != nil || identity != "" {
		t.Fatalf("local source = %q, %v", identity, err)
	}
}

// TestParseAcceptsLocalFilesystemSourcesWithoutError pins the *error* half of a
// local source, which TestCanonical cannot see because Canonical discards it.
// A Windows checkout is declared by its native path, and every backslash form
// used to leave Parse through the scp fallback, which saw the drive colon and
// refused the source as a malformed network remote — so `curator` could not
// install from a local path on Windows at all.
func TestParseAcceptsLocalFilesystemSourcesWithoutError(t *testing.T) {
	for _, source := range []string{
		`C:\repos\skill`,
		"C:/repos/skill",
		`c:\Users\RUNNER~1\AppData\Local\Temp\build\origin\build-skill`,
		`Z:\share\skills\skill-a.git`,
		`\\server\share\skills\skill-a`,
		"/abs/path/repo",
		"./relative",
		"~/home/repo",
	} {
		identity, err := Parse(source)
		if err != nil || identity != "" {
			t.Errorf("Parse(%q) = %q, %v; want a local source with no identity", source, identity, err)
		}
	}
	// The drive carve-out is one letter wide: a real hostname before the colon
	// stays a network remote and stays subject to the allowlist.
	if identity, err := Parse(`ci.example.com:skills/a`); err != nil || identity != "ci.example.com/skills/a" {
		t.Fatalf("Parse of an scp remote = %q, %v", identity, err)
	}
}

func TestValidCanonical(t *testing.T) {
	for _, identity := range []string{"git.example.com/skills/a", "git.example.com/Skills/文書"} {
		if !ValidCanonical(identity) {
			t.Errorf("ValidCanonical(%q) = false", identity)
		}
	}
	for _, identity := range []string{
		"GIT.example.com/skills/a",
		"git.example.com/skills/a b",
		"git.example.com/skills/a#fragment",
		"git.example.com/CON/a",
		"git.example.com/skills/",
	} {
		if ValidCanonical(identity) {
			t.Errorf("ValidCanonical(%q) = true", identity)
		}
	}
}

func TestAllowed(t *testing.T) {
	allowlist := []string{"git.example.com/skills"}
	if !Allowed("git.example.com/skills/skill-a", allowlist) {
		t.Fatal("in-prefix identity must pass")
	}
	if Allowed("git.example.com/other/skill-a", allowlist) {
		t.Fatal("out-of-prefix identity must fail")
	}
	if !Allowed("", allowlist) {
		t.Fatal("local source (no identity) must pass")
	}
	if !Allowed("anything/at/all", nil) {
		t.Fatal("empty allowlist allows everything")
	}
}
