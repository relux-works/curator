package projectcontext

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// Claim: a failed second observation and each changed identity field produce
// distinct, path-free refusals; unchanged observations pass beside every plant.
// This exercises the diagnostic boundary on every platform, without Git or a
// filesystem-identity fixture, and does not reproduce the hosted transient.
func TestBindingRecheckDiagnostics(t *testing.T) {
	remote := "https://example.invalid/repository.git"
	fresh := func() registered {
		b := binding{repository{"/private-checkout", &remote, strings.Repeat("1", 40)}, map[string]string{"device": "3", "inode": "4"}}
		return registered{Record: Registration{RootIdentity: identity(b)}, Binding: b}
	}
	for _, tc := range []struct {
		name   string
		change func(*registered, *binding) error
		want   string
	}{
		{"failed-bind", func(_ *registered, _ *binding) error {
			return bindingFailure("primary remote configuration is unreadable")
		}, "checkout binding recheck failed: primary remote configuration is unreadable"},
		{"failed-bind-with-remedy", func(_ *registered, _ *binding) error { return bindingFailure("root cannot be read; explicitly rebind") }, "checkout binding recheck failed: root cannot be read"},
		{"git-exit", func(_ *registered, _ *binding) error { return &localGitError{code: 128, class: "exit"} }, "checkout binding recheck failed: local Git identity read failed: exit 128"},
		{"git-start", func(_ *registered, _ *binding) error {
			return &localGitError{code: -1, class: "start refused", errno: "EACCES"}
		}, "checkout binding recheck failed: local Git identity read failed: start refused (EACCES)"},
		{"git-timeout", func(_ *registered, _ *binding) error { return &localGitError{code: -1, class: "timed out"} }, "checkout binding recheck failed: local Git identity read failed: timed out"},
		{"git-signal", func(_ *registered, _ *binding) error { return &localGitError{code: -1, class: "signal"} }, "checkout binding recheck failed: local Git identity read failed: signal"},
		{"git-output-bound", func(_ *registered, _ *binding) error { return &localGitError{code: -1, class: "output bound exceeded"} }, "checkout binding recheck failed: local Git identity read failed: output bound exceeded"},
		{"untrusted-error", func(_ *registered, _ *binding) error {
			return &os.PathError{Op: "open", Path: "/do-not-report", Err: os.ErrPermission}
		}, "checkout binding recheck failed: checkout identity is unavailable"},
		{"unknown-error", func(_ *registered, _ *binding) error { return errors.New("/do-not-report") }, "checkout binding recheck failed: checkout identity is unavailable"},
		{"root", func(_ *registered, b *binding) error { b.Repository.Root = "/do-not-report"; return nil }, "checkout binding changed: repository.root"},
		{"remote-value", func(_ *registered, b *binding) error {
			v := "https://example.invalid/do-not-report.git"
			b.Repository.PrimaryRemote = &v
			return nil
		}, "checkout binding changed: repository.primary_remote"},
		{"remote-removed", func(_ *registered, b *binding) error { b.Repository.PrimaryRemote = nil; return nil }, "checkout binding changed: repository.primary_remote"},
		{"remote-added", func(r *registered, _ *binding) error {
			r.Binding.Repository.PrimaryRemote = nil
			r.Record.RootIdentity = identity(r.Binding)
			return nil
		}, "checkout binding changed: repository.primary_remote"},
		{"root-commit", func(_ *registered, b *binding) error { b.Repository.RootCommit = strings.Repeat("2", 40); return nil }, "checkout binding changed: repository.root_commit"},
		{"device", func(_ *registered, b *binding) error { b.Filesystem["device"] = "5"; return nil }, "checkout binding changed: filesystem.device"},
		{"inode", func(_ *registered, b *binding) error { b.Filesystem["inode"] = "5"; return nil }, "checkout binding changed: filesystem.inode"},
		{"missing-inode", func(_ *registered, b *binding) error { delete(b.Filesystem, "inode"); return nil }, "checkout binding changed: filesystem.inode"},
		{"unknown-field", func(_ *registered, b *binding) error { b.Filesystem["/do-not-report"] = "5"; return nil }, "checkout binding changed: filesystem"},
		{"digest", func(r *registered, _ *binding) error { r.Record.RootIdentity = strings.Repeat("0", 64); return nil }, "checkout binding changed: root_identity"},
		{"multiple-fields", func(_ *registered, b *binding) error {
			b.Repository.RootCommit = strings.Repeat("2", 40)
			b.Filesystem["inode"] = "5"
			return nil
		}, "checkout binding changed: repository.root_commit, filesystem.inode"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := fresh()
			if err := verifyBindingResult(r, r.Binding, nil); err != nil {
				t.Fatal("unchanged positive control", err)
			}
			b := fresh().Binding
			remoteCopy := *b.Repository.PrimaryRemote
			b.Repository.PrimaryRemote = &remoteCopy
			if err := verifyBindingResult(r, b, nil); err != nil {
				t.Fatal("equal remote value at a different address control", err)
			}
			err := tc.change(&r, &b)
			before := identity(struct {
				Saved    registered
				Observed binding
			}{r, b})
			got := verifyBindingResult(r, b, err)
			want := "project_registration_mismatch: " + tc.want + "; explicitly rebind"
			if got == nil || got.Error() != want {
				t.Fatalf("got %v, want %s", got, want)
			}
			if after := identity(struct {
				Saved    registered
				Observed binding
			}{r, b}); after != before {
				t.Fatal("refusal mutated saved or observed authority")
			}
			if strings.Contains(got.Error(), "do-not-report") || strings.Contains(got.Error(), remote) {
				t.Fatal("diagnostic exposed identity values or an underlying path")
			}
		})
	}
}

// The existing native replacement witness calls this beside its changed-inode
// row. Real Git controls prove that both diagnostic branches are wired through
// verifyBinding; these direct verifyBinding calls do not write a lapse or rewrite the checkout.
// An ambiguous remote is a real configuration change, not proof of a transient.
func bindingRecheckLiveDiagnostics(t *testing.T, root string, reg Registration) {
	t.Helper()
	b, err := bind(root)
	if err != nil {
		t.Fatal(err)
	}
	r := registered{Record: reg, Binding: b}
	config := filepath.Join(root, ".git", "config")
	original, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.WriteFile(config, original, 0600); err != nil {
			t.Error(err)
		}
	})
	for _, kind := range []string{"ambiguous-remote-config-change", "changed-remote"} {
		t.Run(kind, func(t *testing.T) {
			if err := verifyBinding(r); err != nil {
				t.Fatal("unchanged native control", err)
			}
			gitTest(t, root, "config", "--add", "remote.origin.url", "https://example.invalid/first.git")
			want := "project_registration_mismatch: checkout binding changed: repository.primary_remote; explicitly rebind"
			if kind == "ambiguous-remote-config-change" {
				gitTest(t, root, "config", "--add", "remote.origin.url", "https://example.invalid/second.git")
				want = "project_registration_mismatch: checkout binding recheck failed: primary remote is ambiguous; explicitly rebind"
			}
			planted, err := os.ReadFile(config)
			if err != nil {
				t.Fatal(err)
			}
			before := identity(r)
			err = verifyBinding(r)
			if err == nil || err.Error() != want {
				t.Fatalf("got %v, want %s", err, want)
			}
			after, err := os.ReadFile(config)
			if err != nil || string(after) != string(planted) || identity(r) != before {
				t.Fatal("refusal mutated checkout or saved authority", err)
			}
			if err := os.WriteFile(config, original, 0600); err != nil {
				t.Fatal(err)
			}
			if err := verifyBinding(r); err != nil {
				t.Fatal("restored native control", err)
			}
		})
	}
}

// Claim: known execution failures retain only static labels and start errno
// names. Timeout and output limits cannot be mistaken for a refused spawn;
// every identity-read site carries each class without exposing its path.
func TestBindingGitFailureClasses(t *testing.T) {
	for _, site := range []string{"rev-parse", "rev-list", "config"} {
		t.Run(site, func(t *testing.T) {
			for _, tc := range []struct {
				name     string
				err, ctx error
				bounded  bool
				want     string
			}{
				{"start", &os.PathError{Op: "fork/exec", Path: "/do-not-report", Err: syscall.EACCES}, nil, false, "start refused (EACCES)"},
				{"timeout", errors.New("/do-not-report"), context.DeadlineExceeded, false, "timed out"},
				{"output", errors.New("/do-not-report"), nil, true, "output bound exceeded"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					r := registered{Binding: binding{}}
					r.Record.RootIdentity = identity(r.Binding)
					if err := verifyBindingResult(r, r.Binding, nil); err != nil {
						t.Fatal("equal control", err)
					}
					e := localGitFailure(tc.err, tc.ctx, tc.bounded, site)
					err := verifyBindingResult(r, r.Binding, e)
					if err == nil || !strings.Contains(err.Error(), "at "+site+": "+tc.want) || strings.Contains(err.Error(), "do-not-report") {
						t.Fatalf("got %v", err)
					}
				})
			}
		})
	}
}
