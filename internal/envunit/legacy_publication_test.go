package envunit

import (
	"path/filepath"
	"strings"
	"testing"
)

// This is a grammar/read-scope control, not an install or account witness.
// Exact legacy publication leaves are admitted; labels cannot authorize a
// credential, session, adapter root, or arbitrary user-home tree read.
func TestLegacyPublicationReadScope(t *testing.T) {
	manager := "/fixture/manager"
	user := "/fixture/user"
	for _, tc := range []struct {
		class, path string
		want        bool
	}{
		{"50-env-file", manager + "/global/env.sh", true},
		{"10-context", manager + "/global/skills/tool", true},
		{"20-runtime", manager + "/runtime/tool/abcdef", true},
		{"10-context", user + "/.codex/skills/tool", false},
		{"60-adapter-ledger", user + "/.codex/skills/tool", true},
		{"60-adapter-ledger", user + "/.codex/skills/.csk-managed.json", true},
		{"40-shim-forwarding", user + "/bin/tool", true},
		{"10-context", user + "/.codex", false},
		{"10-context", user + "/.codex/auth.json", false},
		{"10-context", user + "/.codex/sessions", false},
		{"10-context", manager + "/environments/p/codex_cli", false},
		{"50-env-file", manager + "/global/auth.json", false},
	} {
		t.Run(tc.class+"/"+filepath.Base(tc.path), func(t *testing.T) {
			u := &Unit{Manager: Manager{Path: manager}, Action: GlobalExtend,
				Legacy: &LegacyPublication{UserHome: user, UserBin: user + "/bin", Entries: []LegacyTarget{{Class: tc.class, Path: tc.path}}}}
			err := u.ValidateReadScope(tc.class, tc.path)
			if (err == nil) != tc.want {
				t.Fatalf("read scope admitted=%t, want=%t: %v", err == nil, tc.want, err)
			}
		})
	}
}

// Artifact keys and receipt namespaces are closed independently of two
// publishers agreeing. An explicit transport root cannot overlap native,
// account-root, manager authority, or PATH consumer custody.
func TestLegacyExternalPublicationScope(t *testing.T) {
	manager, user := "/fixture/manager", "/fixture/user"
	for _, tc := range []struct {
		name, root, namespace, key, class string
		want                              bool
	}{
		{"legacy", manager + "/external-build-cache", "artifacts", strings.Repeat("a", 64), "05-external-cache", true},
		{"source-aware", "/fixture/custom-cache", "artifacts-receipt-3", strings.Repeat("b", 64), "05-external-cache", true},
		{"snapshot", "/fixture/custom-cache", "snapshots", strings.Repeat("c", 64), "05-external-snapshot", true},
		{"invented-namespace", "/fixture/custom-cache", "artifacts-v2", strings.Repeat("a", 64), "05-external-cache", false},
		{"short-key", "/fixture/custom-cache", "artifacts", "abc", "05-external-cache", false},
		{"uppercase-key", "/fixture/custom-cache", "artifacts", strings.Repeat("A", 64), "05-external-cache", false},
		{"native", user + "/.codex/cache", "artifacts", strings.Repeat("a", 64), "05-external-cache", false},
		{"account", user, "artifacts", strings.Repeat("a", 64), "05-external-cache", false},
		{"manager-authority", manager + "/profiles/cache", "artifacts", strings.Repeat("a", 64), "05-external-cache", false},
		{"bin", user + "/bin/cache", "artifacts", strings.Repeat("a", 64), "05-external-cache", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(tc.root, tc.namespace, tc.key)
			u := &Unit{Manager: Manager{Path: manager}, Action: GlobalExtend, Legacy: &LegacyPublication{UserHome: user, UserBin: user + "/bin", ExternalStore: tc.root, Entries: []LegacyTarget{{Class: tc.class, Path: path}}}}
			err := u.ValidateReadScope(tc.class, path)
			if (err == nil) != tc.want {
				t.Fatalf("external scope admitted=%t want=%t: %v", err == nil, tc.want, err)
			}
		})
	}
}
