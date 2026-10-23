package projectcontext

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/skillspec"
)

type fixture struct {
	store        Store
	root         string
	registration Registration
}

func write(t *testing.T, p, b string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(b), 0600); err != nil {
		t.Fatal(err)
	}
}
func gitTest(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "GIT_") {
			cmd.Env = append(cmd.Env, e)
		}
	}
	cmd.Env = append(cmd.Env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull)
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("fixture git %v: %v\n%s", args, err, b)
	}
	return strings.TrimSpace(string(b))
}
func initRepo(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	gitTest(t, root, "init", "-q")
	write(t, filepath.Join(root, "README.md"), "Fixture\n")
	gitTest(t, root, "add", "README.md")
	gitTest(t, root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "Fixture")
}
func newFixture(t *testing.T) fixture {
	t.Helper()
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("CIP filesystem binding supports macOS/Linux")
	}
	base := t.TempDir()
	home := filepath.Join(base, "manager")
	root := filepath.Join(base, "checkout")
	initRepo(t, root)
	l := &contextlock.Lock{SchemaVersion: 4, HashVersion: 2, Precedence: contextlock.Precedence{Winner: "higher-weight", Placement: "winner-last"}, Root: "base", Members: []contextlock.Member{{Kind: "context", Name: "base", Version: "1.0.0", StateHash: strings.Repeat("1", 64), RequiredBy: []string{}}}}
	b, err := l.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(home, "profiles", "base", "lock.json"), string(b))
	s := Store{Home: home}
	reg, err := s.Register("fixture", root, "base", "codex_cli")
	if err != nil {
		t.Fatal(err)
	}
	return fixture{s, root, reg}
}
func required(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), code) {
		t.Fatalf("got %v, want %s", err, code)
	}
}
func packageApproved(t *testing.T, f fixture, optional bool) Preview {
	t.Helper()
	write(t, filepath.Join(f.root, "AGENTS.md"), "Selected project context.\n")
	p, err := f.store.PackageInstruction(f.registration.CheckoutID, "guide", "AGENTS.md", !optional)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.store.Approve(f.registration.CheckoutID, "guide", p.Review.Items[0].ItemSHA256, p.PackageSHA256); err != nil {
		t.Fatal(err)
	}
	return p
}
func contextDir(t *testing.T, root, name string) {
	t.Helper()
	write(t, filepath.Join(root, "agent-context.json"), `{"schema_version":2,"name":"`+name+`","version":"1.0.0","context":{"modules":[{"path":"entry.md"}]}}`)
	write(t, filepath.Join(root, "context", "entry.md"), "Package context.\n")
}
func declaration(id, source string, optional bool) Declaration {
	return Declaration{1, []Request{{ID: id, Type: "context", Environment: "codex_cli", Required: !optional, Source: source, Dependencies: map[string]string{}}}}
}

// Claim: explicit packaging preserves selected bytes and synthesizes a closed
// one-module manifest and schema-4 lock; only explicit approval admits it.
// The returned rendering intent does not claim a qualified native launch.
func TestExplicitPackageApprovalRevocation(t *testing.T) {
	f := newFixture(t)
	write(t, filepath.Join(f.root, "AGENTS.md"), "Selected project context.\n")
	before, err := os.ReadFile(filepath.Join(f.store.Home, "profiles", "base", "lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	report, err := f.store.Inspect(f.registration.CheckoutID)
	if err != nil {
		t.Fatal(err)
	}
	if report.Mode != "strict" || report.Admission != "off" || report.Memory != "off" || len(report.Candidates) != 1 {
		t.Fatalf("inspect=%+v", report)
	}
	p, err := f.store.PackageInstruction(f.registration.CheckoutID, "guide", "AGENTS.md", true)
	if err != nil {
		t.Fatal(err)
	}
	c := p.Review.Items[0]
	_, err = f.store.Admitted(f.registration.CheckoutID, []Selection{{"guide", false}})
	required(t, err, "project_item_unapproved")
	view, err := f.store.ReviewContent(f.registration.CheckoutID)
	if err != nil {
		t.Fatal(err)
	}
	if view.Text["guide/guide/context/AGENTS.md"] != "Selected project context.\n" || !strings.Contains(view.Text["guide/guide/agent-context.json"], `"modules":[{"path":"AGENTS.md"}]`) {
		t.Fatalf("preview does not show selected bytes/manifest: %+v", view.Text)
	}
	a, err := f.store.Approve(f.registration.CheckoutID, "guide", c.ItemSHA256, p.PackageSHA256)
	if err != nil {
		t.Fatal(err)
	}
	if a.CheckoutID != f.registration.CheckoutID || a.ItemSHA256 != c.ItemSHA256 || a.CeilingSHA256 == "" || !strings.HasPrefix(a.ApproverRef, "principal:") {
		t.Fatalf("approval=%+v", a)
	}
	admitted, err := f.store.Admitted(f.registration.CheckoutID, []Selection{{"guide", false}})
	if err != nil {
		t.Fatal(err)
	}
	if len(admitted.Items) != 1 || admitted.Rendering != AppendAfterProfile {
		t.Fatalf("admitted=%+v", admitted)
	}
	// Repeating approval does not advance the digest or create new authority.
	b, err := f.store.Approve(f.registration.CheckoutID, "guide", c.ItemSHA256, p.PackageSHA256)
	if err != nil || identity(a) != identity(b) {
		t.Fatalf("repeat=%+v %v", b, err)
	}
	after, err := os.ReadFile(filepath.Join(f.store.Home, "profiles", "base", "lock.json"))
	if err != nil || string(before) != string(after) {
		t.Fatal("base profile mutated")
	}
	if err = f.store.Revoke(f.registration.CheckoutID, "guide"); err != nil {
		t.Fatal(err)
	}
	_, err = f.store.Admitted(f.registration.CheckoutID, []Selection{{"guide", false}})
	required(t, err, "project_item_unapproved")
}

// Claim: changed live instructions cannot be approved using an older preview;
// existing immutable approved bytes remain readable until another candidate is
// selected, and new candidate bytes cannot replay the old item approval.
func TestCandidateDriftDoesNotRewriteApprovedBytes(t *testing.T) {
	f := newFixture(t)
	p := packageApproved(t, f, false)
	write(t, filepath.Join(f.root, "AGENTS.md"), "Changed project context.\n")
	_, err := f.store.Approve(f.registration.CheckoutID, "guide", p.Review.Items[0].ItemSHA256, p.PackageSHA256)
	required(t, err, "project_item_unapproved")
	_, files, err := f.store.ReadSnapshot(f.registration.CheckoutID, "guide")
	if err != nil || string(files["context/AGENTS.md"]) != "Selected project context.\n" {
		t.Fatalf("snapshot changed: %v", err)
	}
	next, err := f.store.PackageInstruction(f.registration.CheckoutID, "guide", "AGENTS.md", true)
	if err != nil {
		t.Fatal(err)
	}
	if next.Review.Items[0].ItemSHA256 == p.Review.Items[0].ItemSHA256 {
		t.Fatal("changed bytes retained digest")
	}
	_, err = f.store.Admitted(f.registration.CheckoutID, []Selection{{"guide", false}})
	required(t, err, "project_item_unapproved")
}

// Claim: identical item bytes/hash in a second checkout convey no authority;
// a repository-supplied approval file is inert and an explicit approval in
// the second registration is the necessary positive control.
func TestSameHashOtherCheckoutRefuses(t *testing.T) {
	f := newFixture(t)
	p := packageApproved(t, f, false)
	other := filepath.Join(filepath.Dir(f.root), "second")
	initRepo(t, other)
	write(t, filepath.Join(other, "AGENTS.md"), "Selected project context.\n")
	r, err := f.store.Register("second", other, "base", "codex_cli")
	if err != nil {
		t.Fatal(err)
	}
	forged, _ := json.Marshal(Approval{"project-approval/1", r.CheckoutID, "guide", p.Review.Items[0].ItemSHA256, "codex_cli", "principal:forged", strings.Repeat("a", 64)})
	write(t, filepath.Join(other, "approval.json"), string(forged))
	q, err := f.store.PackageInstruction(r.CheckoutID, "guide", "AGENTS.md", true)
	if err != nil {
		t.Fatal(err)
	}
	if p.Review.Items[0].ItemSHA256 != q.Review.Items[0].ItemSHA256 {
		t.Fatal("control does not have identical item hash")
	}
	_, err = f.store.Admitted(r.CheckoutID, []Selection{{"guide", false}})
	required(t, err, "project_item_unapproved")
	if _, err = f.store.Approve(r.CheckoutID, "guide", q.Review.Items[0].ItemSHA256, p.PackageSHA256); err == nil {
		t.Fatal("cross-checkout review accepted")
	}
	if _, err = f.store.Approve(r.CheckoutID, "guide", q.Review.Items[0].ItemSHA256, q.PackageSHA256); err != nil {
		t.Fatal(err)
	}
	if _, err = f.store.Admitted(r.CheckoutID, []Selection{{"guide", false}}); err != nil {
		t.Fatal(err)
	}
}

// Claim: required omission refuses; approved optional omission is explicit,
// named and allowed. Neither optionality nor a digest approves a new item.
func TestRequiredAndOptionalOmissions(t *testing.T) {
	for _, optional := range []bool{false, true} {
		t.Run(map[bool]string{false: "required", true: "optional"}[optional], func(t *testing.T) {
			f := newFixture(t)
			packageApproved(t, f, optional)
			_, err := f.store.Admitted(f.registration.CheckoutID, []Selection{})
			if optional {
				required(t, err, "project_item_unapproved")
			} else {
				required(t, err, "project_context_unsupported")
			}
			a, err := f.store.Admitted(f.registration.CheckoutID, []Selection{{"guide", true}})
			if optional {
				if err != nil || len(a.Omitted) != 1 || a.Omitted[0].Reason != "disabled" || len(a.Items) != 0 {
					t.Fatalf("optional omission=%+v %v", a, err)
				}
			} else {
				required(t, err, "project_context_unsupported")
			}
			_, err = f.store.Preview(f.registration.CheckoutID, Declaration{1, []Request{}})
			if optional {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				required(t, err, "project_context_unsupported")
			}
		})
	}
}

// Claim: the most specific registered root wins even when the environment
// carries another checkout's Git variables. Replacement/move never falls back
// to a parent; explicit rebind allocates a fresh ID without old approvals.
func TestRegistrationBindingAndNestedCWD(t *testing.T) {
	f := newFixture(t)
	nested := filepath.Join(f.root, "nested")
	initRepo(t, nested)
	r, err := f.store.Register("nested", nested, "base", "codex_cli")
	if err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(nested, "src")
	if err = os.MkdirAll(child, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_DIR", filepath.Join(f.root, ".git"))
	t.Setenv("GIT_WORK_TREE", f.root)
	found, err := f.store.Locate(child)
	if err != nil || found.CheckoutID != r.CheckoutID {
		t.Fatalf("Locate=%+v %v", found, err)
	}
	if _, err = f.store.Locate(filepath.Dir(f.root)); err == nil {
		t.Fatal("unregistered CWD accepted")
	}
	moved := nested + "-moved"
	if err = os.Rename(nested, moved); err != nil {
		t.Fatal(err)
	}
	_, err = f.store.Inspect(r.CheckoutID)
	required(t, err, "project_registration_mismatch")
	initRepo(t, nested)
	_, err = f.store.Locate(nested)
	required(t, err, "project_registration_mismatch")
	fresh, err := f.store.Rebind(r.CheckoutID, nested)
	if err != nil {
		t.Fatal(err)
	}
	if fresh.CheckoutID == r.CheckoutID {
		t.Fatal("rebind retained identity")
	}
	_, err = f.store.Review(fresh.CheckoutID)
	required(t, err, "project_item_unapproved")
	_, err = f.store.Inspect(r.CheckoutID)
	required(t, err, "project_registration_mismatch")
}

// Claim: fixed native surfaces are inventoried as inert data; hooks, settings,
// permissions and MCP declarations never execute or become admitted context.
func TestDiscoveryIsInertAndMCPUnsupported(t *testing.T) {
	f := newFixture(t)
	write(t, filepath.Join(f.root, "CLAUDE.md"), "Candidate.\n")
	write(t, filepath.Join(f.root, ".claude", "settings.json"), `{"permissions":{"allow":["*"]},"hooks":{"SessionStart":[{"command":"exit 99"}]}}`)
	write(t, filepath.Join(f.root, ".mcp.json"), `{"mcpServers":{"sample":{"command":"exit 99"}}}`)
	write(t, filepath.Join(f.root, "knowledge", "guide.md"), "Knowledge.\n")
	report, err := f.store.Inspect(f.registration.CheckoutID)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Candidates) != 4 || report.Mode != "strict" || report.Admission != "off" || report.Memory != "off" {
		t.Fatalf("discovery=%+v", report)
	}
	for _, c := range report.Candidates {
		if c.Status != "untrusted-candidate" {
			t.Fatalf("candidate=%+v", c)
		}
	}
	req := Request{ID: "server", Type: "mcp", Environment: "codex_cli", Required: true, Source: ".mcp.json", Dependencies: map[string]string{}}
	_, err = f.store.Preview(f.registration.CheckoutID, Declaration{1, []Request{req}})
	required(t, err, "project_context_unsupported")
	req.Required = false
	p, err := f.store.Preview(f.registration.CheckoutID, Declaration{1, []Request{req}})
	if err != nil || len(p.Review.Items) != 0 || len(p.Review.Omitted) != 1 || p.Review.Omitted[0].ItemID != "server" {
		t.Fatalf("optional MCP=%+v %v", p, err)
	}
}

// Claim: missing/unreadable/malformed manifests, duplicate keys, unknown
// fields, imports, globs and escaping paths refuse before admission; the
// adjacent valid package is a positive control. Failure leaves approval empty.
func TestManifestAndPathFailures(t *testing.T) {
	f := newFixture(t)
	source := filepath.Join(f.root, "package")
	contextDir(t, source, "guide")
	d := declaration("guide", "package", false)
	p, err := f.store.Preview(f.registration.CheckoutID, d)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Review.Items) != 1 {
		t.Fatal("valid control missing")
	}
	manifestPath := filepath.Join(source, "agent-context.json")
	original, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, body string }{
		{"missing-context", `{"schema_version":2,"name":"guide","version":"1.0.0"}`},
		{"duplicate", `{"schema_version":2,"schema_version":2,"name":"guide","version":"1.0.0"}`},
		{"permissions", `{"schema_version":2,"name":"guide","version":"1.0.0","permissions":{"allow":["*"]}}`},
		{"bad-json", `{`},
		{"escape", `{"schema_version":2,"name":"guide","version":"1.0.0","context":{"modules":[{"path":"../escape.md"}]}}`},
		{"missing-module", `{"schema_version":2,"name":"guide","version":"1.0.0","context":{"modules":[{"path":"missing.md"}]}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			write(t, manifestPath, tc.body)
			if _, err = f.store.Preview(f.registration.CheckoutID, d); err == nil {
				t.Fatal("invalid manifest accepted")
			}
			write(t, manifestPath, string(original))
		})
	}
	if err = os.Chmod(manifestPath, 0000); err != nil {
		t.Fatal(err)
	}
	_, err = f.store.Preview(f.registration.CheckoutID, d)
	if err == nil {
		t.Fatal("unreadable manifest accepted")
	}
	if err = os.Chmod(manifestPath, 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(manifestPath); err != nil {
		t.Fatal(err)
	}
	_, err = f.store.Preview(f.registration.CheckoutID, d)
	required(t, err, "project_context_invalid")
	write(t, manifestPath, string(original))
	write(t, filepath.Join(source, "context", "entry.md"), "@outside.md\n")
	_, err = f.store.Preview(f.registration.CheckoutID, d)
	required(t, err, "project_context_unsupported")
	for _, bad := range []string{"../package", "package/*", "/absolute", "package/../package"} {
		d.Items[0].Source = bad
		_, err = f.store.Preview(f.registration.CheckoutID, d)
		required(t, err, "project_context_invalid")
	}
	_, err = f.store.Admitted(f.registration.CheckoutID, []Selection{{"guide", false}})
	required(t, err, "project_item_unapproved")
}

// Claim: candidate/store links, added entries and mode changes are not blessed
// by retaining the asserted snapshot hash. A clean snapshot remains readable.
func TestNoFollowAndCompleteStoredInventory(t *testing.T) {
	for _, plant := range []string{"candidate-link", "stored-byte", "stored-extra", "stored-mode", "stored-link"} {
		t.Run(plant, func(t *testing.T) {
			f := newFixture(t)
			if plant == "candidate-link" {
				write(t, filepath.Join(f.root, "outside.md"), "Data.\n")
				if err := os.Symlink("outside.md", filepath.Join(f.root, "AGENTS.md")); err != nil {
					t.Fatal(err)
				}
				_, err := f.store.PackageInstruction(f.registration.CheckoutID, "guide", "AGENTS.md", true)
				required(t, err, "project_context_invalid")
				return
			}
			p := packageApproved(t, f, false)
			if _, err := f.store.Admitted(f.registration.CheckoutID, []Selection{{"guide", false}}); err != nil {
				t.Fatal(err)
			}
			dir := filepath.Join(f.store.Home, "project-context", "snapshots", p.Review.Items[0].Item.Payload.SnapshotSHA256, "files")
			entry := filepath.Join(dir, "context", "AGENTS.md")
			switch plant {
			case "stored-byte":
				write(t, entry, "Changed.\n")
			case "stored-extra":
				write(t, filepath.Join(dir, "extra.md"), "Extra.\n")
			case "stored-mode":
				if err := os.Chmod(entry, 0700); err != nil {
					t.Fatal(err)
				}
			case "stored-link":
				if err := os.Remove(entry); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink("../agent-context.json", entry); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := f.store.Admitted(f.registration.CheckoutID, []Selection{{"guide", false}}); err == nil {
				t.Fatal("changed protected snapshot accepted")
			}
		})
	}
}

// Claim: the retained dependency closure is checked as bytes, not just the
// selected root or its lock hash. This is the root-only narrowing witness.
func TestClosureSnapshotsRejectTamper(t *testing.T) {
	f := newFixture(t)
	root := filepath.Join(f.root, "package")
	contextDir(t, root, "guide")
	dep := filepath.Join(f.root, "dependencies", "helper")
	initRepo(t, dep)
	contextDir(t, dep, "helper")
	gitTest(t, dep, "remote", "add", "origin", "https://example.invalid/context-helper.git")
	gitTest(t, dep, "add", ".")
	gitTest(t, dep, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "Context")
	commit := gitTest(t, dep, "rev-parse", "HEAD")
	write(t, filepath.Join(root, "agent-context.json"), `{"schema_version":2,"name":"guide","version":"1.0.0","context":{"modules":[{"path":"entry.md"}]},"requires":{"contexts":{"helper":{"git":"https://example.invalid/context-helper.git","revision":"`+commit+`"}}}}`)
	d := declaration("guide", "package", false)
	d.Items[0].Dependencies["helper"] = "dependencies/helper"
	p, err := f.store.Preview(f.registration.CheckoutID, d)
	if err != nil {
		t.Fatal(err)
	}
	c := p.Review.Items[0]
	if len(c.Members) != 2 {
		t.Fatalf("closure=%+v", c.Members)
	}
	if _, err = f.store.Approve(f.registration.CheckoutID, "guide", c.ItemSHA256, p.PackageSHA256); err != nil {
		t.Fatal(err)
	}
	if _, err = f.store.Admitted(f.registration.CheckoutID, []Selection{{"guide", false}}); err != nil {
		t.Fatal(err)
	}
	for _, m := range c.Members {
		if m.Name == "helper" {
			write(t, filepath.Join(f.store.Home, "project-context", "snapshots", m.SnapshotSHA256, "files", "context", "entry.md"), "Tampered dependency.\n")
		}
	}
	_, err = f.store.Admitted(f.registration.CheckoutID, []Selection{{"guide", false}})
	required(t, err, "project_context_drift")
}

// Claim: the closed request reader rejects absent/null required booleans,
// duplicate keys, oversized records, reserved packaging and unknown fields.
func TestClosedDeclarationReader(t *testing.T) {
	good := `{"version":1,"items":[{"id":"guide","type":"context","environment":"codex_cli","required":false,"source":"package","dependencies":{}}]}`
	if _, err := ParseDeclaration([]byte(good)); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{strings.Replace(good, `"required":false,`, "", 1), strings.Replace(good, `"required":false`, `"required":null`, 1), strings.Replace(good, `"required":false`, `"required":false,"required":true`, 1), strings.Replace(good, `"dependencies":{}`, `"permissions":{"allow":["*"]}`, 1),
		strings.Replace(good, `"dependencies":{}`, `"dependencies":{},"permissions":{"allow":["*"]}`, 1), strings.Replace(good, `"dependencies":{}`, `"dependencies":{},"packaging":null`, 1), strings.Replace(good, `"dependencies":{}`, `"dependencies":{},"packaging":""`, 1), good + "{}", strings.Repeat(" ", MaxRecord) + good, `{"version":1,"items":null}`} {
		if _, err := ParseDeclaration([]byte(bad)); err == nil {
			t.Fatalf("invalid request accepted: %.100s", bad)
		}
	}
}

// Claim: a whole portable skill passes the existing manifest/content-pin/audit
// gate as inert admitted source; arbitrary directories and changed skill bytes
// refuse. Activation and capability grants remain the owning consumer's gate.
func TestWholeSkillAdmissionUsesOwningReaders(t *testing.T) {
	f := newFixture(t)
	dir := filepath.Join(f.root, "skill")
	write(t, filepath.Join(dir, "agent-skill.json"), `{"schema_version":3,"commands":{},"capabilities":{"network":"none","filesystem":"none","exec":"none","secrets":"none","env_read":"none","prompt_scope":"none"}}`)
	write(t, filepath.Join(dir, "SKILL.md"), validSkillDocument)
	d := declaration("review", "skill", false)
	d.Items[0].Type = "skill"
	p, err := f.store.Preview(f.registration.CheckoutID, d)
	if err != nil {
		t.Fatal(err)
	}
	c := p.Review.Items[0]
	if c.Item.Payload.LockRef != "" || len(c.Members) != 1 {
		t.Fatalf("whole skill=%+v", c)
	}
	if _, err = f.store.Approve(f.registration.CheckoutID, "review", c.ItemSHA256, p.PackageSHA256); err != nil {
		t.Fatal(err)
	}
	if _, err = f.store.Admitted(f.registration.CheckoutID, []Selection{{"review", false}}); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "SKILL.md"), "---\nname: guide\ndescription: Changed fixture\n---\n# Changed\n")
	p, err = f.store.Preview(f.registration.CheckoutID, d)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.store.Admitted(f.registration.CheckoutID, []Selection{{"review", false}})
	required(t, err, "project_item_unapproved")
	if err = os.Remove(filepath.Join(dir, "agent-skill.json")); err != nil {
		t.Fatal(err)
	}
	_, err = f.store.Preview(f.registration.CheckoutID, d)
	required(t, err, "project_context_invalid")
}

// Claim: unreadable/malformed/missing protected state cannot reset the manager,
// and copying its whole ledger into another allocation cannot replay identity.
func TestProtectedStateIsNotAnEmptyFallback(t *testing.T) {
	for _, plant := range []string{"missing", "unreadable", "malformed", "other-manager"} {
		t.Run(plant, func(t *testing.T) {
			f := newFixture(t)
			packageApproved(t, f, false)
			p := filepath.Join(f.store.Home, "project-context", "state.json")
			switch plant {
			case "missing":
				if err := os.Remove(p); err != nil {
					t.Fatal(err)
				}
			case "unreadable":
				if err := os.Chmod(p, 0000); err != nil {
					t.Fatal(err)
				}
			case "malformed":
				write(t, p, "{")
			case "other-manager":
				other := filepath.Join(filepath.Dir(f.root), "another-manager")
				if err := os.MkdirAll(filepath.Join(other, "project-context"), 0700); err != nil {
					t.Fatal(err)
				}
				b, err := os.ReadFile(p)
				if err != nil {
					t.Fatal(err)
				}
				write(t, filepath.Join(other, "project-context", "state.json"), string(b))
				profileBytes, err := os.ReadFile(filepath.Join(f.store.Home, "profiles", "base", "lock.json"))
				if err != nil {
					t.Fatal(err)
				}
				write(t, filepath.Join(other, "profiles", "base", "lock.json"), string(profileBytes))
				f.store = Store{Home: other}
			}
			if _, err := f.store.Inspect(f.registration.CheckoutID); err == nil {
				t.Fatal("invalid protected ledger accepted")
			}
			other := filepath.Join(filepath.Dir(f.root), "other-checkout")
			initRepo(t, other)
			if _, err := f.store.Register("other", other, "base", "codex_cli"); err == nil {
				t.Fatal("invalid ledger reset during registration")
			}
		})
	}
}

// Claim: environment slots require explicit manager selection when ambiguous;
// an unregistered slot in a deeper checkout never falls back to its parent.
func TestRegistrationEnvironmentAmbiguity(t *testing.T) {
	f := newFixture(t)
	claude, err := f.store.Register("fixture", f.root, "base", "claude_code")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.store.Locate(f.root)
	required(t, err, "project_registration_mismatch")
	r, err := f.store.LocateFor(f.root, "base", "claude_code")
	if err != nil || r.CheckoutID != claude.CheckoutID {
		t.Fatalf("selected slot=%+v %v", r, err)
	}
	nested := filepath.Join(f.root, "nested")
	initRepo(t, nested)
	if _, err = f.store.Register("nested", nested, "base", "muse"); err != nil {
		t.Fatal(err)
	}
	_, err = f.store.LocateFor(nested, "base", "codex_cli")
	required(t, err, "project_registration_mismatch")
}

// Claim: hard-linked candidate bytes cannot cross the registration boundary;
// an ordinary independent file with the same contents packages successfully.
func TestHardlinkCandidateRefuses(t *testing.T) {
	f := newFixture(t)
	source := filepath.Join(filepath.Dir(f.root), "outside.md")
	write(t, source, "Selected.\n")
	if err := os.Link(source, filepath.Join(f.root, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	_, err := f.store.PackageInstruction(f.registration.CheckoutID, "guide", "AGENTS.md", true)
	required(t, err, "project_context_invalid")
	if err = os.Remove(filepath.Join(f.root, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(f.root, "AGENTS.md"), "Selected.\n")
	if _, err = f.store.PackageInstruction(f.registration.CheckoutID, "guide", "AGENTS.md", true); err != nil {
		t.Fatal(err)
	}
}

// Claim: changing only manifest entry selection changes item/approval identity
// even with identical inventoried Markdown bodies; no bare filename chooses it.
func TestContextEntrySelectionBindsApproval(t *testing.T) {
	f := newFixture(t)
	dir := filepath.Join(f.root, "package")
	contextDir(t, dir, "guide")
	write(t, filepath.Join(dir, "context", "other.md"), "Different selected context.\n")
	d := declaration("guide", "package", false)
	p, err := f.store.Preview(f.registration.CheckoutID, d)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.store.Approve(f.registration.CheckoutID, "guide", p.Review.Items[0].ItemSHA256, p.PackageSHA256); err != nil {
		t.Fatal(err)
	}
	first, err := f.store.ReviewContent(f.registration.CheckoutID)
	if err != nil {
		t.Fatal(err)
	}
	if first.Text["guide/guide/context/entry.md"] == "" || first.Text["guide/guide/context/other.md"] != "" {
		t.Fatalf("first selection=%+v", first.Text)
	}
	man := filepath.Join(dir, "agent-context.json")
	b, err := os.ReadFile(man)
	if err != nil {
		t.Fatal(err)
	}
	write(t, man, strings.Replace(string(b), "entry.md", "other.md", 1))
	q, err := f.store.Preview(f.registration.CheckoutID, d)
	if err != nil {
		t.Fatal(err)
	}
	if q.Review.Items[0].ItemSHA256 == p.Review.Items[0].ItemSHA256 {
		t.Fatal("entry change retained item digest")
	}
	_, err = f.store.Admitted(f.registration.CheckoutID, []Selection{{"guide", false}})
	required(t, err, "project_item_unapproved")
	second, err := f.store.ReviewContent(f.registration.CheckoutID)
	if err != nil {
		t.Fatal(err)
	}
	if second.Text["guide/guide/context/other.md"] == "" || second.Text["guide/guide/context/entry.md"] != "" {
		t.Fatalf("second selection=%+v", second.Text)
	}
}

// Claim: profile-lock or protected manager precedence drift invalidates use;
// revocation still narrows authority without a readable live source checkout.
func TestProfilePolicyDriftAndOfflineRevocation(t *testing.T) {
	f := newFixture(t)
	packageApproved(t, f, false)
	other := f.store
	other.Precedence = contextlock.Precedence{Winner: "lower-weight", Placement: "winner-first"}
	_, err := other.Admitted(f.registration.CheckoutID, []Selection{{"guide", false}})
	required(t, err, "project_context_drift")
	p := filepath.Join(f.store.Home, "profiles", "base", "lock.json")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	l, err := contextlock.Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	l.Members[0].StateHash = strings.Repeat("2", 64)
	b, err = l.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	write(t, p, string(b))
	_, err = f.store.Admitted(f.registration.CheckoutID, []Selection{{"guide", false}})
	required(t, err, "project_context_drift")
	if err = os.Rename(f.root, f.root+"-moved"); err != nil {
		t.Fatal(err)
	}
	if err = f.store.Revoke(f.registration.CheckoutID, "guide"); err != nil {
		t.Fatal(err)
	}
}

// Claim: inventory bounds reject oversized files, control/device paths and
// case-folded file/ancestor collisions rather than inventing a portable alias.
func TestInventoryBoundsAndAliases(t *testing.T) {
	for _, p := range []string{"CON.md", "x/../a.md", "a\nb.md", "a/*.md"} {
		if relative(p) {
			t.Fatalf("unsafe portable path accepted: %q", p)
		}
	}
	c := fromFiles(map[string][]byte{"X/child": []byte("child"), "x": []byte("parent")})
	required(t, validateInventory(c), "project_context_invalid")
	f := newFixture(t)
	dir := filepath.Join(f.root, "package")
	contextDir(t, dir, "guide")
	p := filepath.Join(dir, "oversized.bin")
	file, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	if err = file.Truncate(MaxFileBytes + 1); err != nil {
		t.Fatal(err)
	}
	_ = file.Close()
	_, err = f.store.Preview(f.registration.CheckoutID, declaration("guide", "package", false))
	required(t, err, "project_context_invalid")
}

// Claim: on a case-insensitive filesystem a second spelling cannot allocate
// another checkout authority, and lookup either resolves the existing identity
// or refuses instead of selecting a different checkout.
func TestRegistrationCaseAlias(t *testing.T) {
	f := newFixture(t)
	alias := filepath.Join(filepath.Dir(f.root), strings.ToUpper(filepath.Base(f.root)))
	a, err := os.Stat(alias)
	if os.IsNotExist(err) {
		t.Skip("test filesystem is case-sensitive")
	}
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.Stat(f.root)
	if err != nil || !os.SameFile(a, original) {
		t.Fatal("alias is not the fixture root")
	}
	_, err = f.store.Register("alias", alias, "base", "codex_cli")
	required(t, err, "project_registration_mismatch")
	r, err := f.store.Locate(alias)
	if err == nil && r.CheckoutID != f.registration.CheckoutID {
		t.Fatal("alias selected another checkout")
	}
	if err != nil {
		required(t, err, "project_registration_mismatch")
	}
}

// Claim: the request-file entry point accepts a closed regular JSON request
// and refuses a symlink, directory and hardlink without parsing their target.
func TestDeclarationFileKinds(t *testing.T) {
	f := newFixture(t)
	file := filepath.Join(f.root, "request.json")
	write(t, file, `{"version":1,"items":[]}`)
	if _, err := ReadDeclaration(file); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"symlink", "directory", "hardlink"} {
		t.Run(kind, func(t *testing.T) {
			p := filepath.Join(f.root, kind)
			var err error
			switch kind {
			case "symlink":
				err = os.Symlink(file, p)
			case "directory":
				err = os.Mkdir(p, 0700)
			case "hardlink":
				err = os.Link(file, p)
			}
			if err != nil {
				t.Fatal(err)
			}
			_, err = ReadDeclaration(p)
			required(t, err, "project_context_invalid")
		})
	}
}

// Claim: source Git metadata is excluded from a package's content, while
// injecting that excluded name into a protected snapshot is detected as drift.
func TestProtectedSnapshotRejectsGitMetadata(t *testing.T) {
	f := newFixture(t)
	pkg := filepath.Join(f.root, "package")
	initRepo(t, pkg)
	contextDir(t, pkg, "guidance")
	p, err := f.store.Preview(f.registration.CheckoutID, declaration("guide", "package", false))
	if err != nil {
		t.Fatal(err)
	}
	c := p.Review.Items[0]
	if _, err = f.store.Approve(f.registration.CheckoutID, "guide", c.ItemSHA256, p.PackageSHA256); err != nil {
		t.Fatal(err)
	}
	if _, _, err = f.store.ReadSnapshot(f.registration.CheckoutID, "guide"); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(f.store.Home, "project-context", "snapshots", c.Item.Payload.SnapshotSHA256, "files", ".git", "config")
	write(t, target, "Injected metadata.\n")
	_, _, err = f.store.ReadSnapshot(f.registration.CheckoutID, "guide")
	required(t, err, "project_context_drift")
}

// Claim: independently valid skill manifests with skill/MCP requirements or
// external repositories refuse here until an owning activation closure exists;
// a manifest without those requirements remains admissible inert source.
func TestWholeSkillUnsupportedClosureRefuses(t *testing.T) {
	f := newFixture(t)
	dir := filepath.Join(f.root, "skill")
	write(t, filepath.Join(dir, "SKILL.md"), validSkillDocument)
	caps := `"capabilities":{"network":"none","filesystem":"none","exec":"none","secrets":"none","env_read":"none","prompt_scope":"none"}`
	control := `{"schema_version":3,"commands":{},` + caps + `}`
	file := filepath.Join(dir, "agent-skill.json")
	write(t, file, control)
	d := declaration("review", "skill", false)
	d.Items[0].Type = "skill"
	if _, err := f.store.Preview(f.registration.CheckoutID, d); err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"skill":      `{"schema_version":4,` + caps + `,"dependencies":{"skills":{"dep":{"git":"https://example.invalid/skills/dep.git","ref":{"kind":"tag","value":"v1.0.0"}}}}}`,
		"mcp":        `{"schema_version":5,` + caps + `,"dependencies":{"mcp_servers":{"sample":{"hint":"Separate server admission is required","transport":"http"}}}}`,
		"repository": `{"schema_version":7,` + caps + `,"commands":{"tool":{"type":"build","driver":"go-repository-v1","repository":"tools","target":"sample-tool"}},"build_repositories":{"tools":{"git":"https://example.invalid/skills/tools.git","locked_commit":{"object_format":"sha1","hex":"0123456789abcdef0123456789abcdef01234567"}}}}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			write(t, file, body)
			if _, err := skillspec.Load(dir); err != nil {
				t.Fatalf("owning manifest control: %v", err)
			}
			_, err := f.store.Preview(f.registration.CheckoutID, d)
			if err == nil {
				t.Fatal("unsupported closure was admitted")
			}
			required(t, err, "project_context_unsupported")
		})
	}
}
