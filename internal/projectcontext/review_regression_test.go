package projectcontext

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Claim R1: two unapproved previews cannot remove an active required approval;
// explicit approval of optionality or revocation permits subsequent removal.
func TestRequiredApprovalSurvivesPreview(t *testing.T) {
	for _, action := range []string{"revoke", "approve-optional"} {
		t.Run(action, func(t *testing.T) {
			f := newFixture(t)
			id := f.registration.CheckoutID
			packageApproved(t, f, false)
			if _, err := f.store.Admitted(id, []Selection{{"guide", false}}); err != nil {
				t.Fatal(err)
			}
			if _, err := f.store.Preview(id, Declaration{1, []Request{}}); err == nil {
				t.Fatal("required removal accepted")
			}
			p, err := f.store.PackageInstruction(id, "guide", "AGENTS.md", false)
			if err != nil {
				t.Fatal(err)
			}
			_, err = f.store.Admitted(id, []Selection{{"guide", false}})
			required(t, err, "project_item_unapproved")
			if _, err = f.store.Preview(id, Declaration{1, []Request{}}); err == nil {
				t.Fatal("two-preview required removal accepted")
			}
			retained, err := f.store.Review(id)
			if err != nil || retained.PackageSHA256 != p.PackageSHA256 {
				t.Fatal("rejected removal changed preview", err)
			}
			if action == "revoke" {
				err = f.store.Revoke(id, "guide")
			} else {
				_, err = f.store.Approve(id, "guide", p.Review.Items[0].ItemSHA256, p.PackageSHA256)
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err = f.store.Preview(id, Declaration{1, []Request{}}); err != nil {
				t.Fatal(err)
			}
			if got, err := f.store.Admitted(id, nil); err != nil || len(got.Items) != 0 {
				t.Fatal(got, err)
			}
		})
	}
}

// Claim R2: parser and protected-record readers require exact case at every
// closed nesting level. Dynamic dependency names retain their owning grammar.
func TestExactCaseProjectRecords(t *testing.T) {
	good := `{"version":1,"items":[{"id":"guide","type":"context","environment":"codex_cli","required":true,"source":"package","dependencies":{}}]}`
	if _, err := ParseDeclaration([]byte(good)); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{
		strings.Replace(good, `"required":true`, `"required":true,"Required":false`, 1),
		strings.Replace(good, `"version"`, `"Version"`, 1), strings.Replace(good, `"items"`, `"Items"`, 1),
		strings.Replace(good, `"dependencies":{}`, `"dependencies":{},"Packaging":null`, 1),
		strings.Replace(good, `"dependencies":{}`, `"dependencies":{},"Packaging":""`, 1),
	} {
		if _, err := ParseDeclaration([]byte(bad)); err == nil {
			t.Fatal("case alias accepted", bad)
		}
	}
	r := Registration{Kind: "project-registration/1", CheckoutID: "p1", ProjectRoot: "/fixture", RootIdentity: strings.Repeat("1", 64), Profile: Profile{"base", strings.Repeat("2", 64)}, Environment: "codex_cli", HomeRef: "home:p1", WorktreeRef: "worktree:p1"}
	b, _ := canonical(r)
	var out Registration
	if err := decode(b, &out); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{strings.Replace(string(b), `"lock_sha256"`, `"Lock_sha256"`, 1), strings.Replace(string(b), `"kind"`, `"Kind"`, 1)} {
		if err := decode([]byte(bad), &out); err == nil {
			t.Fatal("nested record alias accepted")
		}
	}
}

// Claim R2/R8: owning context and skill manifests reject case variants without
// changing the current review; their valid siblings produce a retained preview.
func TestExactCaseManifests(t *testing.T) {
	for _, kind := range []string{"context", "skill"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			source := filepath.Join(f.root, "package")
			contextDir(t, source, "guide")
			file := filepath.Join(source, "agent-context.json")
			if kind == "skill" {
				file = filepath.Join(source, "agent-skill.json")
				write(t, file, skillManifest(`{}`, `{}`))
				write(t, filepath.Join(source, "SKILL.md"), validSkillDocument)
			}
			b, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			d := declaration("guide", "package", false)
			d.Items[0].Type = kind
			p, err := f.store.Preview(f.registration.CheckoutID, d)
			if err != nil {
				t.Fatal(err)
			}
			write(t, file, strings.Replace(string(b), `"schema_version"`, `"Schema_version"`, 1))
			_, err = f.store.Preview(f.registration.CheckoutID, d)
			required(t, err, "project_context_invalid")
			after, err := f.store.Review(f.registration.CheckoutID)
			if err != nil || after.PackageSHA256 != p.PackageSHA256 {
				t.Fatal("bad manifest changed preview", err)
			}
		})
	}
}

func skillManifest(commands, dependencies string) string {
	return `{"schema_version":3,"commands":` + commands + `,"capabilities":{"network":"none","filesystem":"none","exec":"none","secrets":"none","env_read":"none","prompt_scope":"none"},"dependencies":` + dependencies + `}`
}

// Claim R3: inert skill review displays retained instructions, executable
// declarations, command owners, capability requests and policy effects. A
// command change is visible against approval and changes the package digest.
func TestSkillSemanticReview(t *testing.T) {
	f := newFixture(t)
	id := f.registration.CheckoutID
	dir := filepath.Join(f.root, "skill")
	write(t, filepath.Join(dir, "agent-skill.json"), skillManifest(`{"inspect-data":{"type":"system","command":"true"}}`, `{}`))
	write(t, filepath.Join(dir, "SKILL.md"), validSkillDocument+"Inspect retained instructions\n")
	d := declaration("review", "skill", false)
	d.Items[0].Type = "skill"
	p, err := f.store.Preview(id, d)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Review.Changes[0].Semantics.After.CommandOwners) != 1 {
		t.Fatal("skill semantic review omitted command owners")
	}
	view, err := f.store.ReviewContent(id)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(view)
	for _, want := range []string{"inspect-data", "Inspect retained instructions", "requested_capabilities", "mcp_endpoints", "manager policy only"} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("skill semantic review omitted %s", want)
		}
	}
	if _, err = f.store.Approve(id, "review", p.Review.Items[0].ItemSHA256, p.PackageSHA256); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "agent-skill.json"), skillManifest(`{"other-tool":{"type":"system","command":"false"}}`, `{}`))
	changed, err := f.store.Preview(id, d)
	if err != nil {
		t.Fatal(err)
	}
	semantics := changed.Review.Changes[0].Semantics
	if semantics == nil || semantics.Before == nil || semantics.Before.CommandOwners[0].Name != "inspect-data" || semantics.After.CommandOwners[0].Name != "other-tool" || semantics.After.CommandOwners[0].ItemID != "review" || changed.PackageSHA256 == p.PackageSHA256 {
		t.Fatal("command semantic delta missing")
	}
	capabilityChange := strings.Replace(skillManifest(`{"other-tool":{"type":"system","command":"false"}}`, `{}`), `"env_read":"none"`, `"env_read":["LANG"]`, 1)
	write(t, filepath.Join(dir, "agent-skill.json"), capabilityChange)
	capabilityPreview, err := f.store.Preview(id, d)
	if err != nil {
		t.Fatal(err)
	}
	if capabilityPreview.PackageSHA256 == changed.PackageSHA256 || !reflect.DeepEqual(capabilityPreview.Review.Changes[0].Semantics.After.RequestedCapabilities["env_read"], []string{"LANG"}) {
		t.Fatal("capability-only change missing from semantic review identity")
	}
	_, err = f.store.Admitted(id, []Selection{{"review", false}})
	required(t, err, "project_item_unapproved")
}

// Claim R4: execute-only roots cannot register or retain lookup authority;
// readable/searchable roots register. Observed loss retains a protected lapse;
// failed new registration cannot grant authority.
func TestExecuteOnlyRootRefuses(t *testing.T) {
	f := newFixture(t)
	id := f.registration.CheckoutID
	if _, err := f.store.LocateFor(f.root, "base", "codex_cli"); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(filepath.Join(f.store.Home, "project-context/state.json"))
	if err := os.Chmod(f.root, 0100); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(f.root, 0700)
	_, err := f.store.LocateFor(f.root, "base", "codex_cli")
	required(t, err, "project_registration_mismatch")
	_, err = f.store.Register("another", f.root, "base", "claude_code")
	required(t, err, "project_registration_mismatch")
	after, _ := os.ReadFile(filepath.Join(f.store.Home, "project-context/state.json"))
	if string(after) == string(before) {
		t.Fatal("observed unreadable root did not retain authority lapse")
	}
	var retained state
	if err := decode(after, &retained); err != nil || !retained.Registrations[id].Lapsed {
		t.Fatal("missing protected root-authority lapse", err)
	}
	if err = os.Chmod(f.root, 0700); err != nil {
		t.Fatal(err)
	}
	if _, err = f.store.Rebind(id, f.root); err != nil {
		t.Fatal(err)
	}
}

// Claim R5: each well-formed unsupported skill closure is a named, content-bound
// omission only when optional. Malformed and unreadable optional inputs refuse.
func TestOptionalUnsupportedSkillClosures(t *testing.T) {
	for _, dep := range []string{
		`{"skills":{"helper":{"git":"https://example.invalid/helper.git","ref":{"kind":"tag","value":"v1.0.0"}}}}`,
		`{"mcp_servers":{"service":{"hint":"Separate server admission","transport":"http"}}}`,
		`external`,
	} {
		t.Run(dep, func(t *testing.T) {
			f := newFixture(t)
			id := f.registration.CheckoutID
			dir := filepath.Join(f.root, "skill")
			file := filepath.Join(dir, "agent-skill.json")
			body := skillManifest(`{}`, dep)
			if strings.Contains(dep, `"skills"`) {
				body = strings.Replace(body, `"schema_version":3`, `"schema_version":4`, 1)
			}
			if strings.Contains(dep, `"mcp_servers"`) {
				body = strings.Replace(body, `"schema_version":3`, `"schema_version":5`, 1)
			}
			if dep == "external" {
				body = `{"schema_version":7,"commands":{"tool":{"type":"build","driver":"go-repository-v1","repository":"helper","target":"sample-tool"}},"capabilities":{"filesystem":"none"},"build_repositories":{"helper":{"git":"https://example.invalid/helper.git","locked_commit":{"object_format":"sha1","hex":"` + strings.Repeat("1", 40) + `"}}}}`
			}
			write(t, file, body)
			write(t, filepath.Join(dir, "SKILL.md"), validSkillDocument)
			d := declaration("review", "skill", false)
			d.Items[0].Type = "skill"
			_, err := f.store.Preview(id, d)
			required(t, err, "project_context_unsupported")
			d.Items[0].Required = false
			p, err := f.store.Preview(id, d)
			if err != nil {
				t.Fatal("optional unsupported closure aborted", err)
			}
			if len(p.Review.Items) != 0 || len(p.Review.Omitted) != 1 || p.Review.Omitted[0].ItemID != "review" || p.Review.Omitted[0].Reason != "unsupported" || !hexRE.MatchString(p.Review.Omitted[0].SourceSHA256) {
				t.Fatal("missing named unsupported omission", p)
			}
			write(t, filepath.Join(dir, "SKILL.md"), "---\nname: guide\ndescription: Changed fixture\n---\n# Changed\n")
			changed, err := f.store.Preview(id, d)
			if err != nil || changed.PackageSHA256 == p.PackageSHA256 {
				t.Fatal("omission identity ignored content", err)
			}
			write(t, file, "{")
			_, err = f.store.Preview(id, d)
			required(t, err, "project_context_invalid")
			write(t, file, body)
			if err = os.Chmod(file, 0000); err != nil {
				t.Fatal(err)
			}
			_, err = f.store.Preview(id, d)
			required(t, err, "project_context_unreadable")
			_ = os.Chmod(file, 0600)
			write(t, file, skillManifest(`{}`, `{}`))
			if p, err = f.store.Preview(id, d); err != nil || len(p.Review.Items) != 1 || len(p.Review.Omitted) != 0 {
				t.Fatal("valid optional skill refused", err)
			}
		})
	}
}

// Claim R6: root spelling obeys the leaf-1/workspace grammar. A tab-bearing
// actual rebind refuses without modifying registration; Unicode roots emit.
func TestRootPathGrammar(t *testing.T) {
	for _, name := range []string{"checkout\tbad", "checkout\x7fbad", "checkout-é"} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			dest := filepath.Join(filepath.Dir(f.root), name)
			before, _ := os.ReadFile(filepath.Join(f.store.Home, "project-context/state.json"))
			if err := os.Rename(f.root, dest); err != nil {
				t.Fatal(err)
			}
			got, err := f.store.Rebind(f.registration.CheckoutID, dest)
			if strings.ContainsAny(name, "\t\x7f") {
				required(t, err, "project_registration_mismatch")
				after, _ := os.ReadFile(filepath.Join(f.store.Home, "project-context/state.json"))
				if string(after) != string(before) {
					t.Fatal("invalid root changed authority")
				}
			} else if canonical, _ := filepath.EvalSymlinks(dest); err != nil || got.ProjectRoot != canonical || !validRoot(got.ProjectRoot) {
				t.Fatal(got, err)
			}
		})
	}
}

// Claim R7: repeated previews compare file bytes against the approved candidate,
// preserving their file delta and package digest until approval changes.
func TestPreviewUsesApprovedBaseline(t *testing.T) {
	f := newFixture(t)
	id := f.registration.CheckoutID
	p := packageApproved(t, f, false)
	write(t, filepath.Join(f.root, "AGENTS.md"), "Changed project context.\n")
	first, err := f.store.PackageInstruction(id, "guide", "AGENTS.md", true)
	if err != nil {
		t.Fatal(err)
	}
	second, err := f.store.PackageInstruction(id, "guide", "AGENTS.md", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Review.Changes[0].Files) != 1 || !reflect.DeepEqual(first, second) || first.PackageSHA256 == p.PackageSHA256 {
		t.Fatal("preview lost approved file delta")
	}
	if _, err = f.store.Approve(id, "guide", second.Review.Items[0].ItemSHA256, second.PackageSHA256); err != nil {
		t.Fatal(err)
	}
	same, err := f.store.PackageInstruction(id, "guide", "AGENTS.md", true)
	if err != nil || len(same.Review.Changes[0].Files) != 0 || same.Review.Changes[0].Effect != "unchanged" {
		t.Fatal("approved baseline did not advance", err)
	}
}

// Claim R8: actual malformed manifests and missing declaration reads use the
// canonical invalid/unreadable distinction, preserve the review and reveal no
// filesystem path. An actual offline two-package cycle maps to invalid; its
// acyclic sibling succeeds, and protected snapshot drift remains distinct.
func TestCanonicalProjectDiagnostics(t *testing.T) {
	f := newFixture(t)
	p := packageApproved(t, f, false)
	_, err := ReadDeclaration(filepath.Join(f.root, "absent.json"))
	required(t, err, "project_context_unreadable")
	if strings.Contains(err.Error(), f.root) {
		t.Fatal("diagnostic leaked path")
	}
	contextDir(t, filepath.Join(f.root, "package"), "guide")
	write(t, filepath.Join(f.root, "package/agent-context.json"), "{")
	_, err = f.store.Preview(f.registration.CheckoutID, declaration("guide", "package", false))
	required(t, err, "project_context_invalid")

	after, err := f.store.Review(f.registration.CheckoutID)
	if err != nil || after.PackageSHA256 != p.PackageSHA256 {
		t.Fatal("invalid input changed review", err)
	}
	// Tag references make a real cyclic closure possible without cyclic Git
	// commit ancestry. All objects remain offline fixture repositories.
	deps := map[string]string{}
	for _, name := range []string{"alpha", "beta"} {
		dir := filepath.Join(f.root, "dependencies", name)
		initRepo(t, dir)
		contextDir(t, dir, name)
		gitTest(t, dir, "remote", "add", "origin", "https://example.invalid/"+name+".git")
		other := "alpha"
		if name == "alpha" {
			other = "beta"
		}
		write(t, filepath.Join(dir, "agent-context.json"), `{"schema_version":2,"name":"`+name+`","version":"1.0.0","context":{"modules":[{"path":"entry.md"}]},"requires":{"contexts":{"`+other+`":{"git":"https://example.invalid/`+other+`.git","tag":"v1.0.0"}}}}`)
		gitTest(t, dir, "add", ".")
		gitTest(t, dir, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "Closure")
		gitTest(t, dir, "-c", "tag.gpgsign=false", "tag", "v1.0.0")
		deps[name] = "dependencies/" + name
	}
	write(t, filepath.Join(f.root, "package/agent-context.json"), `{"schema_version":2,"name":"guide","version":"1.0.0","context":{"modules":[{"path":"entry.md"}]},"requires":{"contexts":{"alpha":{"git":"https://example.invalid/alpha.git","tag":"v1.0.0"}}}}`)
	d := declaration("guide", "package", false)
	d.Items[0].Dependencies = deps
	_, err = f.store.Preview(f.registration.CheckoutID, d)
	required(t, err, "project_context_invalid")
	contextDir(t, filepath.Join(f.root, "package"), "guide")
	valid, err := f.store.Preview(f.registration.CheckoutID, d)
	if err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(f.store.Home, "project-context", "snapshots", valid.Review.Items[0].Item.Payload.SnapshotSHA256, "files", "context/entry.md")
	write(t, entry, "Tampered.\n")
	_, err = f.store.Review(f.registration.CheckoutID)
	required(t, err, "project_context_drift")
}

// Claim R1: admission independently rejects an empty retained candidate review
// while a required approval is active, including records from the earlier
// preview implementation. Revocation makes the same empty review admissible.
func TestAdmissionRetainsRequiredAuthority(t *testing.T) {
	f := newFixture(t)
	id := f.registration.CheckoutID
	packageApproved(t, f, false)
	op, err := f.store.begin(false)
	if err != nil {
		t.Fatal(err)
	}
	p := op.state.Reviews[id]
	p.Review.Items = []Candidate{}
	p.Review.Changes = []Change{}
	p.PackageSHA256 = identity(p.Review)
	op.state.Reviews[id] = p
	err = op.save()
	op.close()
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.store.Admitted(id, nil)
	required(t, err, "project_item_unapproved")
	if err = f.store.Revoke(id, "guide"); err != nil {
		t.Fatal(err)
	}
	if got, err := f.store.Admitted(id, nil); err != nil || len(got.Items) != 0 {
		t.Fatal(got, err)
	}
}

// Claim: an old approval with an exact retained preview migrates to an
// immutable baseline before another preview; missing legacy evidence refuses
// rather than treating a new preview as approval. Revocation remains available.
func TestLegacyApprovalBaseline(t *testing.T) {
	for _, missing := range []bool{false, true} {
		t.Run(map[bool]string{false: "exact", true: "missing"}[missing], func(t *testing.T) {
			f := newFixture(t)
			id := f.registration.CheckoutID
			packageApproved(t, f, false)
			op, err := f.store.begin(false)
			if err != nil {
				t.Fatal(err)
			}
			a := op.state.Approvals[id+":guide"]
			a.Candidate = nil
			op.state.Approvals[id+":guide"] = a
			if missing {
				p := op.state.Reviews[id]
				p.Review.Items = []Candidate{}
				p.PackageSHA256 = identity(p.Review)
				op.state.Reviews[id] = p
			}
			err = op.save()
			op.close()
			if err != nil {
				t.Fatal(err)
			}
			_, err = f.store.PackageInstruction(id, "guide", "AGENTS.md", false)
			if missing {
				required(t, err, "project_context_drift")
				if err = f.store.Revoke(id, "guide"); err != nil {
					t.Fatal(err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				_, err = f.store.Preview(id, Declaration{1, []Request{}})
				required(t, err, "project_context_unsupported")
			}
		})
	}
}

// Claim R2: the project manifest gate accepts exact legacy declarations and
// rejects shadow/unrelated fields the legacy owning reader permits. This does
// not qualify legacy skill admission; the subsequent audit can still refuse it.
func TestLegacySkillClosedMembers(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "agent-skill.json")
	good := `{"schema_version":1,"commands":{"inspect-data":{"type":"system","command":"true"}}}`
	write(t, file, good)
	if err := closedSkill(dir); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{strings.Replace(good, `"schema_version":1`, `"schema_version":1,"Commands":{}`, 1), strings.Replace(good, `"command":"true"`, `"command":"true","Command":"false"`, 1), strings.Replace(good, `"command":"true"`, `"command":"true","unknown":false`, 1)} {
		write(t, file, bad)
		required(t, closedSkill(dir), "project_context_invalid")
	}
}
