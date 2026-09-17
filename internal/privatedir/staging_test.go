package privatedir

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTempStagingIsPrivateAndValidated(t *testing.T) {
	parent := t.TempDir()
	directory, err := TempStaging(parent, "snapshot-")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(directory) }()
	if filepath.Dir(directory) != parent {
		t.Fatalf("staging directory %q is not under %q", directory, parent)
	}
	if err := Validate(directory); err != nil {
		t.Fatalf("staged snapshot directory rejected: %v", err)
	}
	child := filepath.Join(directory, "input")
	if err := os.WriteFile(child, []byte("payload"), 0o600); err != nil {
		t.Fatalf("owner cannot write inside staged directory: %v", err)
	}
}

func TestTempStagingDefaultsToSystemTemp(t *testing.T) {
	directory, err := TempStaging("", "snapshot-")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(directory) }()
	if err := Validate(directory); err != nil {
		t.Fatalf("default staging directory rejected: %v", err)
	}
}
