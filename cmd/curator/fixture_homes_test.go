package main

import (
	"os"
	"testing"
)

// Explicit child roots override isolated parent roots; an empty override stays
// present, while an unselected optional root stays unset after restoration.
func TestCLIUnitFixtureHomesPreserveExplicitSelection(t *testing.T) {
	parentRoot := t.TempDir()
	childRoot := t.TempDir()
	t.Setenv("CODEX_HOME", parentRoot)
	payload := forwardCLIUnitHomes(t, map[string]string{
		"CODEX_HOME": childRoot, "CLAUDE_CONFIG_DIR": "",
	})
	t.Setenv(cliHelperHomes, payload)
	restoreCLIUnitHomes(t)
	if got := os.Getenv("CODEX_HOME"); got != childRoot {
		t.Fatalf("selected CODEX_HOME = %q, want %q", got, childRoot)
	}
	if value, exists := os.LookupEnv("CLAUDE_CONFIG_DIR"); !exists || value != "" {
		t.Fatalf("explicit empty root = %q, present=%v", value, exists)
	}
	if _, exists := os.LookupEnv("PI_CODING_AGENT_DIR"); exists {
		t.Fatal("unselected optional root became present")
	}
}

// A malformed, incomplete or unrelated-key payload must be rejected without
// changing HOME. A complete fixture payload is the neighboring positive.
func TestCLIUnitFixtureHomesRejectInvalidPayloads(t *testing.T) {
	home := os.Getenv("HOME")
	for _, payload := range []string{
		`{`, `{"HOME":"fixture"}`,
		`{"HOME":"fixture","USERPROFILE":"fixture","XDG_CONFIG_HOME":"fixture","PATH":"unrelated"}`,
	} {
		t.Setenv(cliHelperHomes, payload)
		if _, err := cliUnitHomes(); err == nil {
			t.Errorf("accepted invalid fixture payload %s", payload)
		}
		if os.Getenv("HOME") != home {
			t.Fatal("decoder changed HOME after rejection")
		}
	}
	t.Setenv(cliHelperHomes, forwardCLIUnitHomes(t, nil))
	if _, err := cliUnitHomes(); err != nil {
		t.Fatalf("valid fixture payload rejected: %v", err)
	}
}
