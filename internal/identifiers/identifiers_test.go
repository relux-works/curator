package identifiers

import "testing"

func TestValid(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"skill-youtrack", true},
		{"ytx", true},
		{"a", true},
		{"9lives", true},
		{"a.b_c-d", true},
		{"", false},
		{"-leading-dash", false},
		{".hidden", false},
		{"_underscore", false},
		{"has space", false},
		{"path/inside", false},
		{`back\slash`, false},
		{"..", false},
		{"semi;colon", false},
		{"CON", false},
		{"nul.txt", false},
		{"COM1", false},
		{"Lpt9.log", false},
		{"trailing.", false},
	}
	for _, tc := range cases {
		if got := Valid(tc.value); got != tc.want {
			t.Errorf("Valid(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}

func TestValidSourcePath(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"skill-youtrack", true},
		{"internal/skill-metrics", true},
		{"a/b/c", true},
		{"", false},
		{"/absolute", false},
		{"trailing/", false},
		{"has/../dotdot", false},
		{"-option/like", false},
		{`win\path`, false},
		{"dir/NUL.txt", false},
	}
	for _, tc := range cases {
		if got := ValidSourcePath(tc.value); got != tc.want {
			t.Errorf("ValidSourcePath(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}

func TestValidLocale(t *testing.T) {
	for value, want := range map[string]bool{
		"en": true, "pt-BR": true, "zh-Hans-CN": true,
		"": false, "-en": false, "en-": false, "pt_BR": false,
		"../en": false, "русский": false,
	} {
		if got := ValidLocale(value); got != want {
			t.Errorf("ValidLocale(%q) = %v, want %v", value, got, want)
		}
	}
}

func TestPortablePath(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"directory with space/文書.md", true},
		{"a/b", true},
		{"", false},
		{"a/", false},
		{"a//b", false},
		{"a/..", false},
		{"a:b", false},
		{"a\\b", false},
		{"control\u0085name", false},
		{string([]byte{'a', 0xff}), false},
	}
	for _, tc := range cases {
		if got := PortablePath(tc.value); got != tc.want {
			t.Errorf("PortablePath(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}
