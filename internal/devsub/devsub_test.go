package devsub

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/verr"
)

func write(t *testing.T, text string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, Name), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func mustFail(t *testing.T, dir, wantPathPrefix string) {
	t.Helper()
	_, err := Load(dir)
	if err == nil {
		t.Fatalf("expected error with path prefix %q", wantPathPrefix)
	}
	var v *verr.Error
	if !errors.As(err, &v) {
		t.Fatalf("not a validation error: %v", err)
	}
	if !strings.HasPrefix(v.Path, wantPathPrefix) {
		t.Fatalf("error path %q does not start with %q (%s)", v.Path, wantPathPrefix, v.Message)
	}
}

func TestMissingFileYieldsEmpty(t *testing.T) {
	subs, err := Load(t.TempDir())
	if err != nil || len(subs) != 0 {
		t.Fatalf("Load = %v, %v", subs, err)
	}
}

func TestPathSubstitutionResolvesAgainstProject(t *testing.T) {
	dir := write(t, `{"substitutions": {"skill-a": {"path": "../skill-a"}}}`)
	subs, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	sub := subs["skill-a"]
	if !filepath.IsAbs(sub.Path) {
		t.Fatalf("path not resolved: %q", sub.Path)
	}
	if !strings.HasPrefix(sub.Describe(), "path ") {
		t.Fatalf("describe: %q", sub.Describe())
	}
}

func TestGitSubstitutionAllowsBranch(t *testing.T) {
	dir := write(t, `{"substitutions": {"skill-a": {
		"git": "git@example.com:forks/skill-a.git",
		"ref": {"kind": "branch", "value": "fix-pagination"}}}}`)
	subs, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	sub := subs["skill-a"]
	if sub.RefKind != "branch" || sub.RefValue != "fix-pagination" {
		t.Fatalf("sub: %+v", sub)
	}
	if want := "git git@example.com:forks/skill-a.git branch fix-pagination"; sub.Describe() != want {
		t.Fatalf("describe: %q", sub.Describe())
	}
}

func TestRejections(t *testing.T) {
	cases := []struct {
		name string
		text string
		path string
	}{
		{"unknown top", `{"subs": {}}`, Name},
		{"both path and git", `{"substitutions": {"a": {"path": "p", "git": "u", "ref": {"kind": "tag", "value": "v"}}}}`, "substitutions.a"},
		{"neither", `{"substitutions": {"a": {}}}`, "substitutions.a"},
		{"path with ref", `{"substitutions": {"a": {"path": "p", "ref": {"kind": "tag", "value": "v"}}}}`, "substitutions.a"},
		{"git without ref", `{"substitutions": {"a": {"git": "u"}}}`, "substitutions.a"},
		{"bad kind", `{"substitutions": {"a": {"git": "u", "ref": {"kind": "semver", "value": "v"}}}}`, "substitutions.a.ref.kind"},
		{"empty value", `{"substitutions": {"a": {"git": "u", "ref": {"kind": "tag", "value": ""}}}}`, "substitutions.a.ref.value"},
		{"unknown in ref", `{"substitutions": {"a": {"git": "u", "ref": {"kind": "tag", "value": "v", "x": 1}}}}`, "substitutions.a.ref"},
		{"unknown in entry", `{"substitutions": {"a": {"path": "p", "extra": 1}}}`, "substitutions.a"},
		{"bad name", `{"substitutions": {"-a": {"path": "p"}}}`, "substitutions.-a"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { mustFail(t, write(t, tc.text), tc.path) })
	}
}
