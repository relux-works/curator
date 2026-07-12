package verr

import (
	"errors"
	"testing"
)

func TestErrorCarriesPath(t *testing.T) {
	err := New("dependencies.skills.x.ref.kind", "must be %q or %q", "tag", "revision")
	var v *Error
	if !errors.As(err, &v) {
		t.Fatal("New did not produce a *verr.Error")
	}
	if v.Path != "dependencies.skills.x.ref.kind" {
		t.Fatalf("Path = %q", v.Path)
	}
	want := `dependencies.skills.x.ref.kind: must be "tag" or "revision"`
	if err.Error() != want {
		t.Fatalf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestErrorWithoutPath(t *testing.T) {
	err := New("", "top-level message")
	if err.Error() != "top-level message" {
		t.Fatalf("Error() = %q", err.Error())
	}
}
