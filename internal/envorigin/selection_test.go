package envorigin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Proves production selection with only raw-account evidence injected. It does
// not qualify OS account transport or a launcher/native provider process.
func TestProvisioningOriginAccountPolicy(t *testing.T) {
	root := t.TempDir()
	manager := filepath.Join(root, "manager")
	anchor := filepath.Join(root, "operator")
	if err := os.Mkdir(anchor, 0700); err != nil {
		t.Fatal(err)
	}
	real, effective, err := processAccountIdentity()
	if err != nil {
		t.Fatal(err)
	}
	account := func() (Account, error) {
		return Account{UID: effective, RealUID: real, EffectiveUID: effective, Home: anchor}, nil
	}
	cases := []struct{ id, key, selector, want, defaultWant string }{
		{"codex_cli", "CODEX_HOME", filepath.Join(root, "a"), filepath.Join(root, "a"), filepath.Join(anchor, ".codex")},
		{"claude_code", "CLAUDE_CONFIG_DIR", filepath.Join(root, "claude"), filepath.Join(root, "claude"), filepath.Join(anchor, ".claude")},
		{"pi", "PI_CODING_AGENT_DIR", filepath.Join(root, "pi", "agent"), filepath.Join(root, "pi", "agent"), filepath.Join(anchor, ".pi", "agent")},
		{"opencode", "XDG_CONFIG_HOME", filepath.Join(root, "xdg"), filepath.Join(root, "xdg", "opencode"), filepath.Join(anchor, ".config", "opencode")},
		{"muse", "XDG_CONFIG_HOME", filepath.Join(root, "xdg"), filepath.Join(root, "xdg", "muse"), filepath.Join(anchor, ".config", "muse")},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			t.Setenv(tc.key, tc.selector)
			b := Binding{Profile: "p", LockSHA256: "sha256:" + strings.Repeat("a", 64), ManagedTarget: ManagedTarget(manager, "p", tc.id)}
			r, err := Select(manager, b, tc.id, account)
			if err != nil || r.NativeRoot != tc.want || r.Selector.Value == nil || *r.Selector.Value != tc.selector {
				t.Fatalf("selection %+v: %v", r, err)
			}
			for _, state := range []string{"absent", "present-empty", "managed"} {
				t.Run(state, func(t *testing.T) {
					if state == "absent" {
						t.Setenv(tc.key, "unused")
						_ = os.Unsetenv(tc.key)
					} else if state == "present-empty" {
						t.Setenv(tc.key, "")
					} else {
						t.Setenv(tc.key, ManagedTarget(manager, "other", tc.id))
					}
					r, err := Select(manager, b, tc.id, account)
					if err != nil {
						t.Fatal(err)
					}
					if r.NativeRoot != tc.defaultWant {
						t.Fatalf("default %q want%q", r.NativeRoot, tc.defaultWant)
					}
					if state != "managed" && r.Selector.State != state {
						t.Fatal("original selector presence lost")
					}
					if state == "managed" && (r.Selector.Value == nil || *r.Selector.Value != ManagedTarget(manager, "other", tc.id)) {
						t.Fatal("managed original selector lost")
					}
				})
			}
		})
	}
	t.Setenv("HOME", filepath.Join(root, "poison"))
	b := Binding{Profile: "p", LockSHA256: "sha256:" + strings.Repeat("a", 64), ManagedTarget: ManagedTarget(manager, "p", "codex_cli")}
	bad := func() (Account, error) {
		return Account{UID: effective, RealUID: real + 1, EffectiveUID: effective, Home: anchor}, nil
	}
	if _, err := Select(manager, b, "codex_cli", bad); err == nil {
		t.Fatal("ambiguous account identity accepted")
	}
}

// Proves lexical AND physical exclusion without creating missing tails, including
// PR179's absent-EnvRoot case. Capability branches execute assertions, not skips.
func TestProvisioningOriginPhysicalBoundary(t *testing.T) {
	root := t.TempDir()
	manager := filepath.Join(root, "manager")
	if err := os.Mkdir(manager, 0700); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(root, "outside")
	if err := os.Mkdir(outside, 0700); err != nil {
		t.Fatal(err)
	}
	missingManaged := filepath.Join(manager, "environments", "p", "codex_cli")
	if err := ValidateExternal(manager, missingManaged); err == nil {
		t.Fatal("lexical managed tail accepted")
	}
	alias := filepath.Join(root, "alias")
	if err := os.Symlink(manager, alias); err != nil {
		t.Fatal(err)
	}
	if err := ValidateExternal(manager, filepath.Join(alias, "environments", "p", "codex_cli")); err == nil {
		t.Fatal("absent EnvRoot alias accepted")
	}
	caseManager := filepath.Join(root, "MANAGER")
	mi, _ := os.Stat(manager)
	ci, ce := os.Stat(caseManager)
	if ce == nil && os.SameFile(mi, ci) {
		t.Log("case capability: differently cased manager names identify the same inode; exercising absent-root refusal")
		if err := ValidateExternal(manager, filepath.Join(caseManager, "environments", "p", "codex_cli")); err == nil {
			t.Fatal("same-inode absent EnvRoot case alias accepted")
		}
	} else {
		t.Log("case capability: distinct case-sensitive manager directories; exercising outside positive")
		if err := os.Mkdir(caseManager, 0700); err != nil {
			t.Fatal(err)
		}
		if err := ValidateExternal(manager, filepath.Join(caseManager, "environments", "p", "codex_cli")); err != nil {
			t.Fatalf("distinct case-sensitive outside rejected: %v", err)
		}
	}
	if err := ValidateExternal(manager, filepath.Join(outside, "clean", "absent")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(outside, "clean")); !os.IsNotExist(err) {
		t.Fatal("validator created absent tail")
	}
	if err := os.Mkdir(filepath.Join(manager, "environments"), 0700); err != nil {
		t.Fatal(err)
	}
	insideOut := filepath.Join(manager, "environments", "escape")
	if err := os.Symlink(outside, insideOut); err != nil {
		t.Fatal(err)
	}
	if err := ValidateExternal(manager, insideOut); err == nil {
		t.Fatal("managed spelling laundered through outside symlink")
	}
	goodLink := filepath.Join(root, "outside-link")
	if err := os.Symlink(outside, goodLink); err != nil {
		t.Fatal(err)
	}
	if err := ValidateExternal(manager, goodLink); err != nil {
		t.Fatal(err)
	}
	broken := filepath.Join(root, "broken")
	if err := os.Symlink(filepath.Join(root, "none"), broken); err != nil {
		t.Fatal(err)
	}
	if err := ValidateExternal(manager, filepath.Join(broken, "child")); err == nil {
		t.Fatal("broken route treated as clean absence")
	}
	file := filepath.Join(root, "file")
	if err := os.WriteFile(file, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := ValidateExternal(manager, filepath.Join(file, "child")); err == nil {
		t.Fatal("file parent treated as absence")
	}
}
