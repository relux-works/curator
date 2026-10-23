package projectmcpstate

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/projectmcp"
)

// Rehashing every carrier digest cannot admit a resolved version outside its
// range, including an unrequested prerelease; satisfying stable/prerelease
// neighbors remain valid. Revision and tag tests elsewhere retain their gates.
func TestLockRangeConsistency(t *testing.T) {
	for _, tc := range []struct {
		selector, version string
		valid             bool
	}{
		{"^1.2.3", "1.2.3", true}, {"^1.2.3", "1.9.0", true},
		{"^1.2.3", "9.0.0", false}, {"^1.2.3", "1.2.2", false},
		{"^1.2.3", "1.3.0-beta.1", false},
		{"^1.3.0-beta.1", "1.3.0-beta.2", true},
	} {
		t.Run(tc.selector+"/"+tc.version, func(t *testing.T) {
			l := testLock(t)
			p := l.Entries[0].Declaration["package"].(map[string]any)
			delete(p, "tag")
			p["range"] = tc.selector
			l.Entries[0].DeclarationSHA256, _ = projectmcp.Digest(l.Entries[0].Declaration)
			l.Entries[0].Resolved.Version = tc.version
			assertRehashedLock(t, l, tc.valid, "MCP resolved range")
		})
	}
}

// Package replay retains UTF-8 byte ceilings as well as inline scalar limits.
// Each exact-byte boundary passes, each next multibyte value fails, while those
// same oversized package servers remain valid inline declarations.
func TestLockPackageByteBounds(t *testing.T) {
	for _, field := range []string{"command", "args", "url"} {
		for _, over := range []bool{false, true} {
			t.Run(field+map[bool]string{false: "/boundary", true: "/over"}[over], func(t *testing.T) {
				l := testLock(t)
				n := 0
				if over {
					n = 1
				}
				server := map[string]any{"transport": "stdio", "command": "runtime", "args": []string{}, "env_names": []string{}}
				switch field {
				case "command":
					server["command"] = strings.Repeat("é", 127+n) + strings.Repeat("x", 1-n)
				case "args":
					server["args"] = []string{strings.Repeat("é", 4096+n)}
				case "url":
					server = map[string]any{"transport": "http", "url": "https://a.example/" + strings.Repeat("é", 2039+n), "env_names": []string{}}
				}
				normalized, err := projectmcp.NormalizeServer(server)
				if err != nil {
					t.Fatalf("inline scalar control: %v", err)
				}
				l.Entries[0].Resolved.Server = normalized
				l.Entries[0].Resolved.ServerSHA256, _ = projectmcp.Digest(normalized)
				assertRehashedLock(t, l, !over, "resolved package server")
			})
		}
	}
}

func assertRehashedLock(t *testing.T, l *Lock, valid bool, reason string) {
	t.Helper()
	var err error
	l.LockSHA256, err = carrierHash(l, "lock_sha256")
	if err != nil {
		t.Fatal(err)
	}
	payload, err := Canonical(l)
	if err != nil {
		t.Fatal(err)
	}
	_, err = DecodeLock(payload)
	if valid {
		if err != nil {
			t.Fatal(err)
		}
	} else if err == nil || !strings.Contains(err.Error(), projectmcp.Invalid) || !strings.Contains(err.Error(), reason) {
		t.Fatalf("rehashed violation admitted or wrong refusal: %v", err)
	}
}
