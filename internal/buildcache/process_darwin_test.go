package buildcache

import (
	"os"
	"testing"
)

func TestDarwinSavedExecutablePath(t *testing.T) {
	want, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	got, err := darwinExecutablePath(os.Getpid())
	if err != nil || got != want {
		t.Fatalf("native executable = %q, %v; want %q", got, err, want)
	}
	path, err := executableFromDarwinArgs([]byte("\x01\x00\x00\x00/real/image\x00\x00/forged/argv0\x00"))
	if err != nil || path != "/real/image" {
		t.Fatalf("saved path = %q, %v", path, err)
	}
	for _, input := range []string{"", "1234", "1234/unterminated", "1234relative\x00"} {
		if _, err := executableFromDarwinArgs([]byte(input)); err == nil {
			t.Fatalf("malformed process arguments accepted: %q", input)
		}
	}
}
