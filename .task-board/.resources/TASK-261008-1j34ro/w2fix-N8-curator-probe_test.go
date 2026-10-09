package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestWave2AuditPinNamespace(t *testing.T) {
	source, home := profileHome(t)
	t.Setenv("USER", "fixture")
	code, _, _ := runProfile(t, source, "audit", "--allow", "../wave2-outside-audit", "--reason", "synthetic approval")
	_, err := os.Stat(filepath.Join(home, "wave2-outside-audit", "trust.json"))
	t.Logf("audit --allow exit=%d outside_audit_record_exists=%v", code, err == nil)
	if code == 0 || err == nil {
		t.Fatal("non-digest hash escaped the audit namespace")
	}
}
func TestWave2AuditPinCreationTime(t *testing.T) {
	source, home := profileHome(t)
	t.Setenv("USER", "fixture")
	hash := strings.Repeat("a", 64)
	code, _, stderr := runProfile(t, source, "audit", "--allow", hash, "--reason", "synthetic approval")
	if code != 0 {
		t.Fatalf("pin exit=%d stderr=%s", code, stderr)
	}
	raw, err := os.ReadFile(filepath.Join(home, "audit", hash, "trust.json"))
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err = json.Unmarshal(raw, &record); err != nil {
		t.Fatal(err)
	}
	t.Logf("audit --allow exit=%d creation_time_present=%v", code, record["created_at"] != nil)
	if record["created_at"] == nil {
		t.Fatal("pin omits spec-required creation time")
	}
}

// Production umbrella dispatch, with a local recording provider and no shell
// interpretation of the supplied argv values.
func TestWave2UmbrellaArgvEnv(t *testing.T) {
	source, _ := profileHome(t)
	bin := t.TempDir()
	body := "#!/bin/sh\nprintf '%s\\000' \"$@\"\nprintf '%s\\000' \"$WAVE2_VALUE\"\n"
	if err := os.WriteFile(filepath.Join(bin, "curator-run"), []byte(body), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	t.Setenv("WAVE2_VALUE", "line\nwith=equals")
	want := []string{"claude_code", "--permissions=native", "--", "space and quote '\"", "", "line\nbreak", "$(literal)", "line\nwith=equals"}
	args := append([]string{"run", "claude"}, want[1:len(want)-1]...)
	code, stdout, stderr := runProfile(t, source, args...)
	t.Logf("umbrella exit=%d outside-roots-warning=%v", code, strings.Contains(stderr, "subcommand_provider_outside_trust_roots"))
	got := strings.Split(strings.TrimSuffix(stdout, "\x00"), "\x00")
	if code != 0 || !reflect.DeepEqual(got, want) {
		t.Fatalf("argv/env transport drift: exit=%d got=%q want=%q", code, got, want)
	}
}
