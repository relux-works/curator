package main

import (
	"strings"
	"testing"
)

// Two declarations with equal source/revision/default weight survive add,
// config reload and list with their own directory, once each.
func TestProfileComposeListDirectoryRoundTrip(t *testing.T) {
	source := writeMachineConfig(t, `{"schema_version":2,"skills_root":"x","projects":{}}`)
	const repo = "https://example.com/profiles.git"
	const revision = "0123456789012345678901234567890123456789"
	for _, directory := range []string{"profiles/second", "profiles/first"} {
		code, _, stderr := runProfile(t, source, "profile", "compose", "sample", "add", repo, "--revision", revision, "--directory", directory)
		if code != exitOK {
			t.Fatalf("add %s: %d %s", directory, code, stderr)
		}
		source = reloadSource(t, source)
	}
	code, stdout, stderr := runProfile(t, source, "profile", "compose", "sample", "list")
	if code != exitOK {
		t.Fatalf("list: %d %s", code, stderr)
	}
	if len(strings.Split(strings.TrimSpace(stdout), "\n")) != 2 {
		t.Fatalf("expected two rows: %s", stdout)
	}
	for _, directory := range []string{"profiles/first", "profiles/second"} {
		row := repo + "\trevision=" + revision + "\t" + directory + "\t(default)\n"
		if strings.Count(stdout, row) != 1 {
			t.Fatalf("expected one %s row: %s", directory, stdout)
		}
	}
}
