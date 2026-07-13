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
	} {
		if _, err := Parse(source); err == nil {
			t.Errorf("Parse(%q) must fail", source)
		}
	}
	if identity, err := Parse("file:///tmp/repo"); err != nil || identity != "" {
		t.Fatalf("local source = %q, %v", identity, err)
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
