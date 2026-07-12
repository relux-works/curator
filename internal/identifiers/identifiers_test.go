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
	}
	for _, tc := range cases {
		if got := ValidSourcePath(tc.value); got != tc.want {
			t.Errorf("ValidSourcePath(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}
