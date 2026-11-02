package main

import (
	"flag"
	"io"
	"strings"
	"testing"
)

// Proves the real CLI parser preserves original selector absence versus an
// explicitly empty value, and rejects missing historical inputs before an owner
// is invoked. This parser-only test is not a CLI/account process witness.
func TestEnvAdoptPresence(t *testing.T) {
	base := []string{"codex_cli", "--profile", "p", "--managed-target", "/fixture/manager/environments/p/codex_cli", "--lock-sha256", "sha256:" + strings.Repeat("a", 64), "--operator-home", "/fixture/operator", "--native-root", "/fixture/a"}
	for _, test := range []struct {
		name, state string
		value       []string
		valid       bool
	}{
		{"absent", "absent", nil, true}, {"empty", "present-empty", []string{"--selector-value", ""}, true}, {"value", "present-value", []string{"--selector-value", "/fixture/a"}, true},
		{"absent-with-value", "absent", []string{"--selector-value", ""}, false}, {"empty-without-value", "present-empty", nil, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			flags := flag.NewFlagSet("fixture", flag.ContinueOnError)
			flags.SetOutput(io.Discard)
			args := append(append(append([]string{}, base...), "--selector-state", test.state), test.value...)
			req, err := parseOriginAttestation(flags, args)
			if test.valid {
				if err != nil || req.SelectorState != test.state || (req.SelectorValue == nil) != (test.state == "absent") {
					t.Fatalf("selector presence lost: %+v %v", req, err)
				}
			} else if err == nil || !strings.Contains(err.Error(), "presence differs") {
				t.Fatalf("intended presence refusal: %v", err)
			}
		})
	}
	flags := flag.NewFlagSet("fixture", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	if _, err := parseOriginAttestation(flags, []string{"codex_cli"}); err == nil || !strings.Contains(err.Error(), "requires explicit --profile") {
		t.Fatalf("missing authority not rejected: %v", err)
	}
}
