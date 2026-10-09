package main

// Wave-2 finding N9 regression at the CLI production entry: `audit
// --allow` records a defined creation time alongside the content
// identity, operator identity, and reason (pinned manager source-audit
// contract). A mutant that drops the timestamp leaves a pin with no
// created_at member, which this test refuses; every other pin field
// keeps its existing contract. This test runs on the hosted gate
// (R194), like every cmd/curator suite.

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAuditAllowRecordsPinCreationTime(t *testing.T) {
	digest := "sha256:" + strings.Repeat("c", 64)
	t.Setenv("USER", "synthetic-operator")

	for _, writer := range []struct {
		name string
		v2   bool
	}{
		{"v2-writer", true},
		{"v1-writer", false},
	} {
		t.Run(writer.name, func(t *testing.T) {
			pinAuditHashWriters(t, writer.v2)
			_, home := legacyProject(t)
			configPath := filepath.Join(home, "config.json")
			before := time.Now().UTC()
			if code, stdout, stderr := capture(t, configPath, "audit", "--allow", digest, "--reason", "synthetic approval"); code != exitOK {
				t.Fatalf("audit --allow = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
			}
			after := time.Now().UTC()

			raw := readCLIPinCarrier(t, home, digest)
			if raw["content_sha256"] != strings.ToLower(digest) {
				t.Fatalf("pin content_sha256 = %v, want %s", raw["content_sha256"], strings.ToLower(digest))
			}
			if raw["pinned"] != true {
				t.Fatalf("pin pinned = %v, want true", raw["pinned"])
			}
			if raw["pinned_by"] != "synthetic-operator" {
				t.Fatalf("pin pinned_by = %v, want synthetic-operator", raw["pinned_by"])
			}
			if raw["reason"] != "synthetic approval" {
				t.Fatalf("pin reason = %v, want synthetic approval", raw["reason"])
			}
			stamped, ok := raw["created_at"].(string)
			if !ok || stamped == "" {
				t.Fatalf("pin created_at = %v, want a defined creation timestamp", raw["created_at"])
			}
			if !strings.HasSuffix(stamped, "Z") {
				t.Fatalf("pin created_at = %q, want an RFC3339 UTC timestamp", stamped)
			}
			parsed, err := time.Parse(time.RFC3339, stamped)
			if err != nil {
				t.Fatalf("pin created_at = %q, want a valid timestamp: %v", stamped, err)
			}
			if parsed.Before(before.Add(-time.Minute)) || parsed.After(after.Add(time.Minute)) {
				t.Fatalf("pin created_at = %q outside the pin window %v..%v", stamped, before, after)
			}
		})
	}
}
