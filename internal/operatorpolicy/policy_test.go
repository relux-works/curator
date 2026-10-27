package operatorpolicy

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/projectcontext"
	"github.com/relux-works/curator/internal/testgitenv"
)

func TestMain(m *testing.M) {
	restore := testgitenv.Isolate()
	home := os.Getenv("HOME")
	root := filepath.Dir(os.Getenv("GIT_CONFIG_GLOBAL"))
	// Fail before any test if the neutral HOME/prefix boundary is unavailable.
	// PR #196 supplies it; validation overlays that exact helper until landing.
	if !filepath.IsAbs(home) || home != filepath.Join(root, "home") || os.Getenv("XDG_CONFIG_HOME") != filepath.Join(root, "xdg") {
		_, _ = os.Stderr.WriteString("operatorpolicy tests require the neutral testgitenv HOME boundary (PR #196)\n")
		restore()
		os.Exit(2)
	}
	if e := os.Setenv("USERPROFILE", home); e != nil {
		panic(e)
	}
	code := m.Run()
	restore()
	os.Exit(code)
}

type fixture struct {
	store           Store
	selection       Selection
	command         SignedCommand
	operator, actor ed25519.PrivateKey
}

func git(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "GIT_") {
			cmd.Env = append(cmd.Env, e)
		}
	}
	cmd.Env = append(cmd.Env, "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
	b, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("git: %v %s", e, b)
	}
	return strings.TrimSpace(string(b))
}
func write(t *testing.T, p string, b []byte) {
	t.Helper()
	if e := os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(p, b, 0600); e != nil {
		t.Fatal(e)
	}
}
func signed(t *testing.T, c Command, owner, actor ed25519.PrivateKey) SignedCommand {
	t.Helper()
	c.ActorPublicKey = base64.StdEncoding.EncodeToString(actor.Public().(ed25519.PublicKey))
	b, e := canonical(c)
	if e != nil {
		t.Fatal(e)
	}
	return SignedCommand{c, base64.StdEncoding.EncodeToString(ed25519.Sign(owner, b)), base64.StdEncoding.EncodeToString(ed25519.Sign(actor, b))}
}
func newFixture(t *testing.T) fixture {
	t.Helper()
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("protected registration filesystem identity is qualified on macOS/Linux")
	}
	base, e := filepath.EvalSymlinks(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	home := filepath.Join(base, "manager")
	root := filepath.Join(base, "checkout")
	write(t, filepath.Join(root, "README.md"), []byte("fixture"))
	git(t, root, "init", "-q")
	git(t, root, "add", "README.md")
	git(t, root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "Fixture")
	git(t, root, "remote", "add", "origin", "https://code.example/team/project.git")
	lock := &contextlock.Lock{SchemaVersion: 4, HashVersion: 2, Precedence: contextlock.Precedence{Winner: "higher-weight", Placement: "winner-last"}, Root: "base", Members: []contextlock.Member{{Kind: "context", Name: "base", Version: "1.0.0", StateHash: strings.Repeat("1", 64), RequiredBy: []string{}}}}
	b, e := lock.Canonical()
	if e != nil {
		t.Fatal(e)
	}
	write(t, filepath.Join(home, "profiles", "base", "lock.json"), b)
	reg, e := (projectcontext.Store{Home: home}).Register("fixture", root, "base", "codex_cli")
	if e != nil {
		t.Fatal(e)
	}
	pub, owner, e := ed25519.GenerateKey(rand.Reader)
	if e != nil {
		t.Fatal(e)
	}
	_, actor, e := ed25519.GenerateKey(rand.Reader)
	if e != nil {
		t.Fatal(e)
	}
	store := Store{Home: home}
	if e = store.PinOperator(context.Background(), "principal:operator", base64.StdEncoding.EncodeToString(pub)); e != nil {
		t.Fatal(e)
	}
	at := time.Now().UTC().Add(-time.Second).Format(time.RFC3339)
	issuer := Issuer{"operator", "principal:operator", strings.Repeat("a", 64), Actor{"keeper", "principal:keeper"}}
	sel := Selection{reg.CheckoutID, reg.ProjectRoot, "codex_cli", "lane:build", reg.Profile}
	policy := Policy{ID: "build", RootCommit: git(t, root, "rev-list", "--max-parents=0", "HEAD"), Destination: Destination{Transport: "https", Host: "code.example", Port: 443, Path: "/team/project.git"}, Spellings: []string{"https://code.example/team/project.git"}, Environments: []string{"codex_cli"}, Lanes: []string{"lane:build"}, ProjectTrust: true, Hooks: []Hook{{"pre-start", strings.Repeat("b", 64)}}, Issuer: issuer, Time: at, Reason: "Approve fixture", ReviewBy: time.Now().UTC().Add(time.Hour).Format(time.RFC3339)}
	command := Command{Kind: "operator-policy-command/1", Operation: "grant", Issuer: issuer, Time: at, Reason: policy.Reason, EntryID: policy.ID, Policy: &policy, Selection: &sel}
	return fixture{store, sel, signed(t, command, owner, actor), owner, actor}
}
func approve(t *testing.T, f fixture) Event {
	t.Helper()
	ev, e := f.store.Apply(context.Background(), f.command)
	if e != nil {
		t.Fatal(e)
	}
	return *ev
}
func entryRevision(t *testing.T, f fixture, id string) uint64 {
	t.Helper()
	list, e := f.store.List(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	return list.EntryRevisions[id]
}
func observeGranted(_ context.Context, _ Selection, ev Event) (hookObservation, error) {
	return hookObservation{append([]Hook{}, ev.Surfaces.Hooks...), []Hook{}}, nil
}
func require(t *testing.T, e error, code string) {
	t.Helper()
	if e == nil || !strings.HasPrefix(e.Error(), code+":") {
		t.Fatalf("error=%v want %s", e, code)
	}
}
func stateBytes(t *testing.T, f fixture) []byte {
	t.Helper()
	b, e := os.ReadFile(filepath.Join(f.store.Home, "operator-policy", "state.json"))
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func unchanged(t *testing.T, f fixture, b []byte) {
	t.Helper()
	if !bytes.Equal(stateBytes(t, f), b) {
		t.Fatal("refused operation changed private state")
	}
}

// Claim: authenticated operator/keeper commands record a schema-valid policy
// event; package trust/hook claims, forged signatures/actors and altered signed
// scope cannot create consent. The paired controls use real Ed25519 keys.
func TestPolicyAuthorityPairs(t *testing.T) {
	for _, name := range []string{"package-trust", "package-hooks", "unapproved-keeper", "forged-issuer", "actor-signature", "changed-scope", "approval-reuse"} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			ev := approve(t, f)
			if ev.Event != "policy-grant" || ev.Record != "workspace-trust-policy-grant-v1" || validEvent(ev) != nil {
				t.Fatal("invalid policy event")
			}
			before := stateBytes(t, f)
			bad := f.command
			switch name {
			case "package-trust":
				bad.Command.Issuer.Actor.Kind = "package"
			case "package-hooks":
				bad.Command.Issuer.Role = "package"
			case "unapproved-keeper":
				bad.ApprovalSignature = bad.ActorSignature
			case "forged-issuer":
				bad.Command.Issuer.Principal = "principal:other"
			case "actor-signature":
				bad.ActorSignature = bad.ApprovalSignature
			case "changed-scope":
				bad.Command.EntryID = "other"
			case "approval-reuse":
				bad.Command.Reason = "Changed approval"
			}
			out, e := f.store.Apply(context.Background(), bad)
			require(t, e, "workspace_trust_policy_authority")
			if out != nil {
				t.Fatal("refusal emitted event")
			}
			unchanged(t, f, before)
		})
	}
}

// Claim: an unchanged exact grant can be read through a test-only observer;
// revoked policy is denied on the next read and cannot be replayed to renew it.
// Production remains unqualified; this does not qualify native hook behavior.
func TestPolicyRevocationNextStart(t *testing.T) {
	f := newFixture(t)
	ev := approve(t, f)
	ctx := context.Background()
	out, e := f.store.read(ctx, f.selection, observeGranted)
	if e != nil || out == nil || out.GrantID != ev.ID {
		t.Fatalf("valid read %v %v", out, e)
	}
	c := Command{Kind: "operator-policy-command/1", Operation: "revoke", Issuer: f.command.Command.Issuer, Time: time.Now().UTC().Format(time.RFC3339), Reason: "Revoke fixture", EntryID: "build"}
	c.Issuer.Approval = strings.Repeat("c", 64)
	c.EntryRevision = entryRevision(t, f, c.EntryID)
	rev := signed(t, c, f.operator, f.actor)
	if _, e = f.store.Apply(ctx, rev); e != nil {
		t.Fatal(e)
	}
	before := stateBytes(t, f)
	out, e = f.store.read(ctx, f.selection, observeGranted)
	require(t, e, "workspace_trust_grant_revoked")
	if out != nil {
		t.Fatal("revoked result")
	}
	_, e = f.store.Apply(ctx, f.command)
	require(t, e, "workspace_trust_grant_revoked")
	unchanged(t, f, before)
	list, e := f.store.List(ctx)
	if e != nil || len(list.Events) != 1 || list.Events[0].State.Status != "revoked" || len(list.Events[0].History) != 1 {
		t.Fatalf("history %+v %v", list, e)
	}
	if _, e = f.store.Apply(ctx, rev); e != nil {
		t.Fatal("idempotent revoke", e)
	}
	unchanged(t, f, before)
}

// Claim: each changed effective/native hook digest (including native approvals
// with an empty grant), unreadable hook observations, or unqualified adapter
// refuses without a delivery or private/projection writes. Native behavior is
// deliberately not certified by this internal-observer test.
func TestPolicyHookPairs(t *testing.T) {
	for _, name := range []string{"hook-digest", "native-unlisted", "empty-hooks", "hook-unreadable", "unqualified"} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			if name == "empty-hooks" {
				f.command.Command.Policy.Hooks = []Hook{}
				f.command = signed(t, f.command.Command, f.operator, f.actor)
			}
			approve(t, f)
			before := stateBytes(t, f)
			if out, e := f.store.read(context.Background(), f.selection, observeGranted); e != nil || out == nil {
				t.Fatal("positive", e)
			}
			observer := func(ctx context.Context, s Selection, ev Event) (hookObservation, error) {
				good, _ := observeGranted(ctx, s, ev)
				switch name {
				case "hook-digest":
					good.Effective[0].SHA256 = strings.Repeat("d", 64)
				case "native-unlisted", "empty-hooks":
					good.NativeApprovals = []Hook{{"foreign", strings.Repeat("d", 64)}}
				case "hook-unreadable":
					good.NativeApprovals = nil
				case "unqualified":
					return good, refusal("workspace_trust_adapter_unqualified")
				}
				return good, nil
			}
			code := "workspace_trust_hook_digest"
			if name == "unqualified" {
				code = "workspace_trust_adapter_unqualified"
			}
			out, e := f.store.read(context.Background(), f.selection, observer)
			require(t, e, code)
			if out != nil {
				t.Fatal("refusal emitted delivery")
			}
			unchanged(t, f, before)
		})
	}
}

// Claim: exact project/environment/lane/profile/lock scope never broadens; the
// valid adjacent selection passes the core check, while each changed selection
// is refused without output or native state changes.
func TestPolicyScopePairs(t *testing.T) {
	for _, name := range []string{"unknown-project", "other-registered-project", "other-registered-environment", "project-descendant", "environment", "lane", "profile", "lock", "stale-installed-lock", "destination-spelling"} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			approve(t, f)
			before := stateBytes(t, f)
			if out, e := f.store.read(context.Background(), f.selection, observeGranted); e != nil || out == nil {
				t.Fatal("positive", e)
			}
			sel := f.selection
			code := "workspace_untrusted"
			switch name {
			case "unknown-project":
				sel.Checkout = "unknown"
			case "other-registered-project":
				other := filepath.Join(filepath.Dir(sel.Path), "other")
				git(t, filepath.Dir(sel.Path), "clone", "--no-local", sel.Path, other)
				git(t, other, "remote", "set-url", "origin", "https://code.example/team/project.git")
				r, err := (projectcontext.Store{Home: f.store.Home}).Register("other", other, "base", "codex_cli")
				if err != nil {
					t.Fatal(err)
				}
				sel.Checkout = r.CheckoutID
				sel.Path = r.ProjectRoot
			case "other-registered-environment":
				r, err := (projectcontext.Store{Home: f.store.Home}).Register("other-env", sel.Path, "base", "claude_code")
				if err != nil {
					t.Fatal(err)
				}
				sel.Checkout = r.CheckoutID
				sel.Environment = r.Environment

			case "project-descendant":
				sel.Path += "/child"
			case "environment":
				sel.Environment = "claude_code"
			case "lane":
				sel.Lane = "lane:other"
			case "profile":
				sel.Profile.Name = "other"
				code = "project_context_drift"
			case "lock":
				sel.Profile.LockSHA256 = strings.Repeat("f", 64)
				code = "project_context_drift"
			case "stale-installed-lock":
				p := filepath.Join(f.store.Home, "profiles", "base", "lock.json")
				b, e := os.ReadFile(p)
				if e != nil {
					t.Fatal(e)
				}
				write(t, p, bytes.ReplaceAll(b, []byte(strings.Repeat("1", 64)), []byte(strings.Repeat("2", 64))))
				code = "project_context_drift"
			case "destination-spelling":
				git(t, sel.Path, "remote", "set-url", "origin", "https://code.example/team/./project.git")
				code = "workspace_trust_repository_changed"
			}
			out, e := f.store.read(context.Background(), sel, observeGranted)
			require(t, e, code)
			if out != nil {
				t.Fatal("scope refusal emitted delivery")
			}
			if name != "destination-spelling" {
				unchanged(t, f, before)
			}
		})
	}
}

// Claim: complete-entry digest and signed audit are checked on every read;
// malformed/deleted state and policy edits fail closed rather than becoming an
// empty ledger or falling back to an older grant. Read-only list stays inert.
func TestPolicyPrivateStatePairs(t *testing.T) {
	for _, damage := range []string{"stale-digest", "missing-state", "malformed", "unknown-event", "missing-audit", "foreign-journal"} {
		t.Run(damage, func(t *testing.T) {
			f := newFixture(t)
			approve(t, f)
			before := stateBytes(t, f)
			if _, e := f.store.List(context.Background()); e != nil {
				t.Fatal(e)
			}
			unchanged(t, f, before)
			var state ledger
			if e := json.Unmarshal(before, &state); e != nil {
				t.Fatal(e)
			}
			code := "workspace_trust_ledger_unreadable"
			switch damage {
			case "stale-digest":
				p := state.Policies["build"]
				p.Reason = "Changed"
				state.Policies["build"] = p
				code = "workspace_trust_policy_digest"
			case "missing-state":
				if e := os.Remove(filepath.Join(f.store.Home, "operator-policy", "state.json")); e != nil {
					t.Fatal(e)
				}
			case "malformed":
				write(t, filepath.Join(f.store.Home, "operator-policy", "state.json"), []byte("{"))
			case "unknown-event":
				state.Events[0].Event = "explicit-grant"
			case "missing-audit":
				state.Audit = map[string]SignedCommand{}
				code = "workspace_trust_policy_authority"
			case "foreign-journal":
				state.Journal = []jsonJournal{{"foreign"}}
			}
			if damage != "missing-state" && damage != "malformed" {
				b, e := canonical(state)
				if e != nil {
					t.Fatal(e)
				}
				write(t, filepath.Join(f.store.Home, "operator-policy", "state.json"), b)
			}
			out, e := f.store.Read(context.Background(), f.selection)
			require(t, e, code)
			if out != nil {
				t.Fatal("damaged state emitted delivery")
			}
		})
	}
}

// Claim: editing even a reason revokes all events of the former entry; narrowing
// hooks needs a new signed approval and cannot reuse the old event/approval.
func TestPolicyEditRevokesOldEvent(t *testing.T) {
	f := newFixture(t)
	old := approve(t, f)
	c := f.command.Command
	c.Issuer.Approval = strings.Repeat("e", 64)
	c.Time = time.Now().UTC().Format(time.RFC3339)
	c.Reason = "Narrow hooks"
	c.EntryRevision = entryRevision(t, f, c.EntryID)
	p := *c.Policy
	p.Issuer = c.Issuer
	p.Time = c.Time
	p.Reason = c.Reason
	p.Hooks = []Hook{}
	c.Policy = &p
	ev, e := f.store.Apply(context.Background(), signed(t, c, f.operator, f.actor))
	if e != nil {
		t.Fatal(e)
	}
	if ev.ID == old.ID || len(ev.Surfaces.Hooks) != 0 {
		t.Fatal("edit reused old event")
	}
	out, e := f.store.read(context.Background(), f.selection, observeGranted)
	if e != nil || out.GrantID != ev.ID {
		t.Fatal(out, e)
	}
	list, e := f.store.List(context.Background())
	if e != nil || list.Events[0].State.Status != "revoked" || list.Events[0].Policy != old.Policy {
		t.Fatal("old provenance lost", e)
	}
}

// Claim: the real production read refuses unqualified adapters, including an
// empty hook list, and unknown checkouts refuse workspace_untrusted first.
func TestProductionReadQualification(t *testing.T) {
	f := newFixture(t)
	f.command.Command.Policy.Hooks = []Hook{}
	f.command = signed(t, f.command.Command, f.operator, f.actor)
	approve(t, f)
	before := stateBytes(t, f)
	out, e := f.store.Read(context.Background(), f.selection)
	require(t, e, "workspace_trust_adapter_unqualified")
	if out != nil {
		t.Fatal("unqualified delivery")
	}
	unchanged(t, f, before)
	unknown := f.selection
	unknown.Checkout = "unknown"
	out, e = f.store.Read(context.Background(), unknown)
	require(t, e, "workspace_untrusted")
	if out != nil {
		t.Fatal("unknown delivery")
	}
}

// Claim: the closed request parser accepts the actual signed envelope and
// rejects case aliases, missing required fields, duplicates and unknown fields
// before a command can be authenticated or any state written.
func TestSignedCommandClosedParser(t *testing.T) {
	f := newFixture(t)
	b, e := json.Marshal(f.command)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = ParseSignedCommand(b); e != nil {
		t.Fatal("valid parser control", e)
	}
	for _, bad := range [][]byte{
		bytes.Replace(b, []byte(`"command"`), []byte(`"Command"`), 1),
		bytes.Replace(b, []byte(`"kind":"operator-policy-command/1",`), nil, 1),
		append([]byte(`{"unknown":true,`), b[1:]...),
		append([]byte(`{"actor_signature":"duplicate",`), b[1:]...),
	} {
		if bytes.Equal(bad, b) {
			t.Fatal("control plant did not apply")
		}
		_, e = ParseSignedCommand(bad)
		require(t, e, "workspace_trust_policy_authority")
	}
}

// Claim: expired policy cannot be newly applied. A nearby valid policy passes
// before its real deadline; the next read durably lapses it at expiry and replay
// cannot renew it. No synthetic vendor qualification is inferred.
func TestPolicyExpiryLapse(t *testing.T) {
	f := newFixture(t)
	c := f.command.Command
	p := *c.Policy
	p.Time = time.Now().UTC().Add(-2 * time.Hour).Format(time.RFC3339)
	c.Time = p.Time
	p.ReviewBy = time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
	c.Policy = &p
	_, e := f.store.Apply(context.Background(), signed(t, c, f.operator, f.actor))
	require(t, e, "workspace_untrusted")
	c = f.command.Command
	p = *c.Policy
	p.ReviewBy = time.Now().UTC().Add(5 * time.Second).Format(time.RFC3339)
	c.Policy = &p
	f.command = signed(t, c, f.operator, f.actor)
	approve(t, f)
	if out, e := f.store.read(context.Background(), f.selection, observeGranted); e != nil || out == nil {
		t.Fatal("current deadline control", e)
	}
	deadline, _ := timestamp(p.ReviewBy)
	time.Sleep(time.Until(deadline) + 10*time.Millisecond)
	out, e := f.store.read(context.Background(), f.selection, observeGranted)
	require(t, e, "workspace_untrusted")
	if out != nil {
		t.Fatal("expired delivery")
	}
	list, e := f.store.List(context.Background())
	if e != nil || list.Events[0].State.Status != "lapsed" || list.Events[0].State.By != nil || len(list.Events[0].History) != 1 {
		t.Fatal("lapse not retained", e)
	}
	_, e = f.store.Apply(context.Background(), f.command)
	require(t, e, "workspace_trust_grant_revoked")
}

// Claim: loss of the retained checkout inode durably lapses consent; restoring
// the very same old directory after that observation does not revive a grant.
func TestPolicyFilesystemLapse(t *testing.T) {
	f := newFixture(t)
	approve(t, f)
	if out, e := f.store.read(context.Background(), f.selection, observeGranted); e != nil || out == nil {
		t.Fatal("valid inode control", e)
	}
	moved := f.selection.Path + "-moved"
	if e := os.Rename(f.selection.Path, moved); e != nil {
		t.Fatal(e)
	}
	git(t, filepath.Dir(moved), "clone", "--no-local", moved, f.selection.Path)
	git(t, f.selection.Path, "remote", "set-url", "origin", "https://code.example/team/project.git")
	out, e := f.store.read(context.Background(), f.selection, observeGranted)
	require(t, e, "workspace_trust_repository_changed")
	if out != nil {
		t.Fatal("changed inode delivery")
	}
	replacement := f.selection.Path + "-replacement"
	if e = os.Rename(f.selection.Path, replacement); e != nil {
		t.Fatal(e)
	}
	if e = os.Rename(moved, f.selection.Path); e != nil {
		t.Fatal(e)
	}
	out, e = f.store.read(context.Background(), f.selection, observeGranted)
	require(t, e, "workspace_untrusted")
	if out != nil {
		t.Fatal("restored inode reactivated consent")
	}
}

// Claim: enrollment is idempotent for the pinned operator only; another
// principal/key cannot replace an existing pin or write private state.
func TestPolicyOperatorPin(t *testing.T) {
	f := newFixture(t)
	before := stateBytes(t, f)
	ctx := context.Background()
	key := base64.StdEncoding.EncodeToString(f.operator.Public().(ed25519.PublicKey))
	if e := f.store.PinOperator(ctx, "principal:operator", key); e != nil {
		t.Fatal(e)
	}
	unchanged(t, f, before)
	for _, name := range []string{"principal:other", "principal:operator"} {
		e := f.store.PinOperator(ctx, name, base64.StdEncoding.EncodeToString(f.actor.Public().(ed25519.PublicKey)))
		require(t, e, "workspace_trust_policy_authority")
		unchanged(t, f, before)
	}
}

// Claim: real package manifest ingestion accepts an inert context package but
// refuses package-declared trust or hooks. Neither ingestion nor its rejection
// mints, edits or revokes the independently approved operator policy.
func TestPolicyPackageDeclarationPairs(t *testing.T) {
	f := newFixture(t)
	approve(t, f)
	before := stateBytes(t, f)
	valid := `{"schema_version":2,"name":"fixture","version":"1.0.0","context":{"modules":[{"path":"entry.md"}]}}`
	if _, e := contextpkg.ParseManifest([]byte(valid)); e != nil {
		t.Fatal("valid inert package", e)
	}
	for _, field := range []string{`"project_trust":true,`, `"hooks":[{"hook_id":"pre-start","sha256":"` + strings.Repeat("b", 64) + `"}],`} {
		bad := "{" + field + valid[1:]
		_, e := contextpkg.ParseManifest([]byte(bad))
		require(t, e, contextpkg.DiagManifestInvalid)
		unchanged(t, f, before)
	}
}

// Claim: an ordinary commit keeps consent bound to the same repository root
// ancestor, while incomplete/shallow history refuses identity verification.
func TestPolicyRepositoryHistoryPair(t *testing.T) {
	f := newFixture(t)
	ev := approve(t, f)
	before := stateBytes(t, f)
	write(t, filepath.Join(f.selection.Path, "next.md"), []byte("next"))
	git(t, f.selection.Path, "add", "next.md")
	git(t, f.selection.Path, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "Next")
	out, e := f.store.read(context.Background(), f.selection, observeGranted)
	if e != nil || out.GrantID != ev.ID {
		t.Fatal("ordinary commit lapsed consent", e)
	}
	unchanged(t, f, before)
	// Marking the original root as shallow retains the same visible unique root
	// but makes history explicitly incomplete, an independent refusal witness.
	write(t, filepath.Join(f.selection.Path, ".git", "shallow"), []byte(f.command.Command.Policy.RootCommit+"\n"))
	out, e = f.store.read(context.Background(), f.selection, observeGranted)
	require(t, e, "workspace_trust_repository_unverifiable")
	if out != nil {
		t.Fatal("incomplete history delivery")
	}
	unchanged(t, f, before)
}
