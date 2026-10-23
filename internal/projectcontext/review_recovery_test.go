package projectcontext

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const validSkillDocument = "---\nname: guide\ndescription: Review inert project data\n---\n# Guide\n"

func skillFixture(t *testing.T, f fixture) Declaration {
	t.Helper()
	dir := filepath.Join(f.root, "skill")
	write(t, filepath.Join(dir, "agent-skill.json"), skillManifest(`{"inspect-data":{"type":"system","command":"true"}}`, `{}`))
	write(t, filepath.Join(dir, "SKILL.md"), validSkillDocument)
	d := declaration("guide", "skill", true)
	d.Items[0].Type = "skill"
	return d
}

// Claim R4-r2: an observed unreadable, missing or moved root stays lapsed after
// restoration on the same inode, across store operations. Explicit rebind gets
// a fresh ID and requires new approval; unrelated CWD loss does not lapse a root.
func TestRootAuthorityRequiresExplicitRecovery(t *testing.T) {
	for _, cause := range []string{"permissions", "missing", "moved", "missing-cwd"} {
		t.Run(cause, func(t *testing.T) {
			f := newFixture(t)
			id := f.registration.CheckoutID
			packageApproved(t, f, false)
			if got, err := f.store.Admitted(id, []Selection{{"guide", false}}); err != nil || len(got.Items) != 1 {
				t.Fatal("valid admission control", err)
			}
			var restore func()
			if cause == "permissions" {
				if err := os.Chmod(f.root, 0100); err != nil {
					t.Fatal(err)
				}
				restore = func() {
					if err := os.Chmod(f.root, 0700); err != nil {
						t.Fatal(err)
					}
				}
			} else if cause != "missing-cwd" {
				moved := f.root + "-moved"
				if err := os.Rename(f.root, moved); err != nil {
					t.Fatal(err)
				}
				restore = func() {
					if err := os.Rename(moved, f.root); err != nil {
						t.Fatal(err)
					}
				}
			} else {
				_, err := f.store.LocateFor(filepath.Join(f.root, "absent-child"), "base", "codex_cli")
				required(t, err, "project_registration_mismatch")
				if _, err = f.store.Admitted(id, []Selection{{"guide", false}}); err != nil {
					t.Fatal("unrelated CWD error lapsed valid root", err)
				}
				return
			}
			defer func() { _ = os.Chmod(f.root, 0700) }()
			var err error
			if cause == "moved" {
				_, err = f.store.Inspect(id)
			} else {
				_, err = f.store.LocateFor(f.root, "base", "codex_cli")
			}
			required(t, err, "project_registration_mismatch")
			restore()
			// A separate Store value must observe the retained lapse, not process state.
			restored := Store{Home: f.store.Home}
			if _, err = restored.Admitted(id, []Selection{{"guide", false}}); err == nil {
				t.Fatal("restored root revived old authority")
			} else {
				required(t, err, "project_registration_mismatch")
			}
			_, err = restored.Inspect(id)
			required(t, err, "project_registration_mismatch")
			_, err = restored.LocateFor(f.root, "base", "codex_cli")
			required(t, err, "project_registration_mismatch")
			_, err = restored.Register("replacement", f.root, "base", "codex_cli")
			required(t, err, "project_registration_mismatch")
			rebound, err := restored.Rebind(id, f.root)
			if err != nil || rebound.CheckoutID == id {
				t.Fatal("explicit recovery failed", err)
			}
			p, err := restored.PackageInstruction(rebound.CheckoutID, "guide", "AGENTS.md", true)
			if err != nil {
				t.Fatal(err)
			}
			_, err = restored.Admitted(rebound.CheckoutID, []Selection{{"guide", false}})
			required(t, err, "project_item_unapproved")
			if _, err = restored.Approve(rebound.CheckoutID, "guide", p.Review.Items[0].ItemSHA256, p.PackageSHA256); err != nil {
				t.Fatal(err)
			}
			if got, err := restored.Admitted(rebound.CheckoutID, []Selection{{"guide", false}}); err != nil || len(got.Items) != 1 {
				t.Fatal("rebound approved control", err)
			}
			_, err = restored.Admitted(id, []Selection{{"guide", false}})
			required(t, err, "project_registration_mismatch")
		})
	}
}

// Claim R11: known native authentication entries (including case aliases and
// empty files) refuse before snapshot publication for context and skill, even
// an optional unsupported skill. Prior state, retained files and review text do
// not acquire the synthetic marker. Ordinary non-auth metadata remains visible.
// This path exclusion is not a general secret detector.
func TestNativeAuthenticationNeverPublished(t *testing.T) {
	for _, kind := range []string{"context", "required-context", "skill", "required-skill", "optional-unsupported-skill"} {
		for _, name := range []string{"config/muse/auth.json", "config/muse/Auth.json", "CONFIG/MUSE/auth.json", "config/muse/auth.json/child", "auth.json", ".credentials.json"} {
			for _, empty := range []bool{false, true} {
				t.Run(kind+"/"+name+map[bool]string{false: "", true: "-empty"}[empty], func(t *testing.T) {
					f := newFixture(t)
					id := f.registration.CheckoutID
					d := skillFixture(t, f)
					source := filepath.Join(f.root, "skill")
					if strings.HasSuffix(kind, "context") {
						source = filepath.Join(f.root, "context")
						contextDir(t, source, "guide")
						d = declaration("guide", "context", true)
					}
					d.Items[0].Required = strings.HasPrefix(kind, "required-")
					write(t, filepath.Join(source, "config/muse/metadata.json"), "ordinary fixture metadata\n")
					good, err := f.store.Preview(id, d)
					if err != nil {
						t.Fatal("ordinary metadata control", err)
					}
					view, err := f.store.ReviewContent(id)
					if err != nil {
						t.Fatal(err)
					}
					if !strings.HasSuffix(kind, "context") && view.Text["guide/guide/config/muse/metadata.json"] != "ordinary fixture metadata\n" {
						t.Fatal("ordinary metadata filtered")
					}
					before := protectedFiles(t, f.store.Home)
					marker := "inert-native-state-marker"
					if empty {
						marker = ""
					}
					write(t, filepath.Join(source, filepath.FromSlash(name)), marker)
					if kind == "optional-unsupported-skill" {
						body := strings.Replace(skillManifest(`{}`, `{"skills":{"helper":{"git":"https://example.invalid/helper.git","ref":{"kind":"tag","value":"v1.0.0"}}}}`), `"schema_version":3`, `"schema_version":4`, 1)
						write(t, filepath.Join(source, "agent-skill.json"), body)
					}
					_, err = f.store.Preview(id, d)
					if err == nil {
						t.Fatal("native authentication candidate was published")
					}
					required(t, err, "project_context_invalid")
					if strings.Contains(err.Error(), "inert-native-state-marker") {
						t.Fatal("native marker in diagnostic")
					}
					if !reflect.DeepEqual(before, protectedFiles(t, f.store.Home)) {
						t.Fatal("native authentication changed retained storage")
					}
					after, err := f.store.ReviewContent(id)
					if err != nil || after.Preview.PackageSHA256 != good.PackageSHA256 {
						t.Fatal("native authentication replaced prior review", err)
					}
					b, _ := json.Marshal(after)
					if strings.Contains(string(b), "inert-native-state-marker") {
						t.Fatal("native authentication in review text")
					}
				})
			}
		}
	}
}

func protectedFiles(t *testing.T, home string) map[string]string {
	t.Helper()
	out := map[string]string{}
	root := filepath.Join(home, "project-context")
	err := filepath.WalkDir(root, func(p string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		out[rel] = string(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Claim R3-r2: the approved skill's command/capability semantics remain visible
// through removal or replacement by context; reverse and same-kind transitions
// show only their applicable sides. Repeated previews retain the approved
// baseline and digest. These previews do not change the approved candidate.
func TestSkillTransitionSemantics(t *testing.T) {
	for _, transition := range []string{"removed", "skill-to-context", "context-to-skill", "skill-to-skill"} {
		t.Run(transition, func(t *testing.T) {
			f := newFixture(t)
			id := f.registration.CheckoutID
			skill := skillFixture(t, f)
			contextDir(t, filepath.Join(f.root, "context"), "guide")
			ctx := declaration("guide", "context", true)
			before, next := skill, skill
			switch transition {
			case "removed":
				next = Declaration{1, []Request{}}
			case "skill-to-context":
				next = ctx
			case "context-to-skill":
				before = ctx
			}
			approved, err := f.store.Preview(id, before)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = f.store.Approve(id, "guide", approved.Review.Items[0].ItemSHA256, approved.PackageSHA256); err != nil {
				t.Fatal(err)
			}
			if transition == "skill-to-skill" {
				write(t, filepath.Join(f.root, "skill/agent-skill.json"), skillManifest(`{"other-tool":{"type":"system","command":"false"}}`, `{}`))
			}
			p, err := f.store.Preview(id, next)
			if err != nil {
				t.Fatal(err)
			}
			again, err := f.store.Preview(id, next)
			if err != nil || !reflect.DeepEqual(p, again) {
				t.Fatal("unstable transition preview", err)
			}
			sem := p.Review.Changes[0].Semantics
			if sem == nil {
				t.Fatal("skill transition omitted semantic delta")
			}
			if transition != "context-to-skill" {
				if sem.Before == nil || len(sem.Before.CommandOwners) != 1 || sem.Before.CommandOwners[0].Name != "inspect-data" || !reflect.DeepEqual(sem.Before, approved.Review.Changes[0].Semantics.After) {
					t.Fatal("skill transition lost approved semantics")
				}
			} else if sem.Before != nil {
				t.Fatal("invented before skill semantics")
			}
			if transition == "removed" || transition == "skill-to-context" {
				if sem.After != nil {
					t.Fatal("invented after skill semantics")
				}
			} else {
				want := "inspect-data"
				if transition == "skill-to-skill" {
					want = "other-tool"
				}
				if sem.After == nil || len(sem.After.CommandOwners) != 1 || sem.After.CommandOwners[0].Name != want {
					t.Fatal("missing after skill semantics")
				}
			}
			view, err := f.store.ReviewContent(id)
			if err != nil {
				t.Fatal(err)
			}
			b, _ := json.Marshal(view)
			if !strings.Contains(string(b), "inspect-data") {
				t.Fatal("approved owner absent from full review")
			}
		})
	}
}

// Claim R10: whole-skill admission requires the owning core package reader's
// document/frontmatter contract, even for optional unsupported closures.
// Valid LF/CRLF and multiline YAML descriptions preview, approve and admit;
// invalid input cannot replace the existing review or approvals.
func TestWholeSkillDocumentShape(t *testing.T) {
	cases := map[string]string{
		"missing": "", "no-frontmatter": "# Guide\n", "unterminated": "---\nname: guide\n", "missing-name": "---\ndescription: d\n---\n",
		"empty-name": "---\nname: ''\ndescription: d\n---\n", "nonstring-name": "---\nname: 7\ndescription: d\n---\n", "missing-description": "---\nname: guide\n---\n", "empty-description": "---\nname: guide\ndescription: ' '\n---\n", "nonstring-description": "---\nname: guide\ndescription: [d]\n---\n", "invalid-yaml": "---\nname: [\n---\n", "bad-triggers": "---\nname: guide\ndescription: d\ntriggers: ['']\n---\n",
	}
	for _, unsupported := range []bool{false, true} {
		for name, body := range cases {
			t.Run(name+map[bool]string{false: "", true: "-unsupported"}[unsupported], func(t *testing.T) {
				f := newFixture(t)
				id := f.registration.CheckoutID
				d := skillFixture(t, f)
				good, err := f.store.Preview(id, d)
				if err != nil {
					t.Fatal(err)
				}
				if _, err = f.store.Approve(id, "guide", good.Review.Items[0].ItemSHA256, good.PackageSHA256); err != nil {
					t.Fatal(err)
				}
				if got, err := f.store.Admitted(id, []Selection{{"guide", false}}); err != nil || len(got.Items) != 1 {
					t.Fatal("valid skill control", err)
				}
				doc := filepath.Join(f.root, "skill/SKILL.md")
				if name == "missing" {
					if err = os.Remove(doc); err != nil {
						t.Fatal(err)
					}
				} else {
					write(t, doc, body)
				}
				if unsupported {
					write(t, filepath.Join(f.root, "skill/agent-skill.json"), strings.Replace(skillManifest(`{}`, `{"skills":{"helper":{"git":"https://example.invalid/helper.git","ref":{"kind":"tag","value":"v1.0.0"}}}}`), `"schema_version":3`, `"schema_version":4`, 1))
				}
				_, err = f.store.Preview(id, d)
				if err == nil {
					t.Fatal("invalid whole-skill document was admitted")
				}
				required(t, err, "project_context_invalid")
				review, err := f.store.Review(id)
				if err != nil || review.PackageSHA256 != good.PackageSHA256 {
					t.Fatal("invalid document replaced approved review", err)
				}
			})
		}
	}
	for _, doc := range []string{validSkillDocument, strings.ReplaceAll(validSkillDocument, "\n", "\r\n"), "---\nname: guide\ndescription: |\n  Review data\ntriggers: [review]\n---\n# Guide\n"} {
		t.Run("valid", func(t *testing.T) {
			f := newFixture(t)
			d := skillFixture(t, f)
			write(t, filepath.Join(f.root, "skill/SKILL.md"), doc)
			p, err := f.store.Preview(f.registration.CheckoutID, d)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = f.store.Approve(f.registration.CheckoutID, "guide", p.Review.Items[0].ItemSHA256, p.PackageSHA256); err != nil {
				t.Fatal(err)
			}
			if got, err := f.store.Admitted(f.registration.CheckoutID, []Selection{{"guide", false}}); err != nil || len(got.Items) != 1 {
				t.Fatal("valid YAML control", err)
			}
		})
	}
}

// Claim R4-r2: observing one root's loss lapses every registration slot bound
// to that root and leaves its nested shadow in place. Parent authority and a
// still-verifiable root survive unrelated child/profile-read errors.
func TestRootLapseSlotsAndShadow(t *testing.T) {
	f := newFixture(t)
	second, err := f.store.Register("second", f.root, "base", "claude_code")
	if err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(f.root, "nested")
	initRepo(t, child)
	nested, err := f.store.Register("nested", child, "base", "codex_cli")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(child, 0100); err != nil {
		t.Fatal(err)
	}
	_, err = f.store.Inspect(nested.CheckoutID)
	required(t, err, "project_registration_mismatch")
	if err = os.Chmod(child, 0700); err != nil {
		t.Fatal(err)
	}
	_, err = f.store.LocateFor(child, "base", "codex_cli")
	required(t, err, "project_registration_mismatch")
	if _, err = f.store.Inspect(f.registration.CheckoutID); err != nil {
		t.Fatal("parent control", err)
	}
	lock := filepath.Join(f.store.Home, "profiles/base/lock.json")
	b, err := os.ReadFile(lock)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(lock); err != nil {
		t.Fatal(err)
	}
	_, err = f.store.Inspect(f.registration.CheckoutID)
	required(t, err, "project_context_unreadable")
	write(t, lock, string(b))
	if _, err = f.store.Inspect(f.registration.CheckoutID); err != nil {
		t.Fatal("profile error lapsed valid root", err)
	}
	if err = os.Chmod(f.root, 0100); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(f.root, 0700)
	_, err = f.store.Inspect(f.registration.CheckoutID)
	required(t, err, "project_registration_mismatch")
	if err = os.Chmod(f.root, 0700); err != nil {
		t.Fatal(err)
	}
	if _, err = f.store.Inspect(second.CheckoutID); err == nil {
		t.Fatal("sibling slot revived root authority")
	} else {
		required(t, err, "project_registration_mismatch")
	}
}

// Claim R11: even a matching-hash snapshot planted by an older permissive
// writer refuses on protected read before review can emit its synthetic data.
// The neighboring safe snapshot remains readable; contents are never filtered.
func TestRetainedNativeAuthenticationRefuses(t *testing.T) {
	f := newFixture(t)
	d := skillFixture(t, f)
	p, err := f.store.Preview(f.registration.CheckoutID, d)
	if err != nil {
		t.Fatal(err)
	}
	op, err := f.store.begin(false)
	if err != nil {
		t.Fatal(err)
	}
	defer op.close()
	c := p.Review.Items[0]
	safe, err := op.get(c.Item.Payload.SnapshotRef, c.Item.Payload.SnapshotSHA256)
	if err != nil {
		t.Fatal("safe retained control", err)
	}
	files := map[string][]byte{}
	for name, b := range safe.Bytes {
		files[name] = b
	}
	files["CONFIG/MUSE/Auth.json"] = []byte("inert-retained-native-state-marker")
	unsafe := fromFiles(files)
	h := identity(unsafe.Manifest)
	dir := filepath.Join(op.root, "snapshots", h)
	for name, b := range unsafe.Bytes {
		write(t, filepath.Join(dir, "files", filepath.FromSlash(name)), string(b))
	}
	b, _ := canonical(unsafe.Manifest)
	write(t, filepath.Join(dir, "manifest.json"), string(b))
	if _, err = op.get(snapshotRef(h), h); err == nil {
		t.Fatal("retained native authentication snapshot was read")
	} else {
		required(t, err, "project_context_drift")
	}
	if _, err = op.get(c.Item.Payload.SnapshotRef, c.Item.Payload.SnapshotSHA256); err != nil {
		t.Fatal("unsafe read changed safe control", err)
	}
}

// Claim: replacing a registered directory before any missing-root observation
// cannot reuse authority even with identical Git metadata. A still-bound root
// and an explicit fresh-ID rebind succeed. This isolates filesystem identity
// from the persistent lapse gate, so a root-path-only checker is detected.
func TestRootReplacementBindingStillRequired(t *testing.T) {
	f := newFixture(t)
	id := f.registration.CheckoutID
	if _, err := f.store.Inspect(id); err != nil {
		t.Fatal("original root control", err)
	}
	moved := f.root + "-original"
	if err := os.Rename(f.root, moved); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(f.root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(moved, ".git"), filepath.Join(f.root, ".git")); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(f.root, "README.md"), "Fixture\n")
	if _, err := f.store.Inspect(id); err == nil {
		t.Fatal("replacement reused native root identity")
	} else {
		required(t, err, "project_registration_mismatch")
	}
	r, err := f.store.Rebind(id, f.root)
	if err != nil || r.CheckoutID == id {
		t.Fatal("replacement rebind control", err)
	}
	if _, err = f.store.Inspect(r.CheckoutID); err != nil {
		t.Fatal(err)
	}
}
