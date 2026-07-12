package version

import "testing"

func TestStringNonEmpty(t *testing.T) {
	if String() == "" {
		t.Fatal("version.String() returned an empty string")
	}
}

func TestLdflagsValueWins(t *testing.T) {
	old := value
	defer func() { value = old }()
	value = "v9.9.9-test"
	if got := String(); got != "v9.9.9-test" {
		t.Fatalf("String() = %q, want the ldflags value", got)
	}
}
