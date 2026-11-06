//go:build linux || darwin

package projectcontext

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type bindingRecords struct{ state, lapses []byte }

func readBindingRecords(t *testing.T, f fixture) bindingRecords {
	t.Helper()
	a, e := os.ReadFile(filepath.Join(f.store.Home, "project-context", "state.json"))
	if e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(filepath.Join(f.store.Home, "project-context", "lapses.json"))
	if e != nil {
		t.Fatal(e)
	}
	return bindingRecords{a, b}
}
func sameBindingRecords(t *testing.T, f fixture, before bindingRecords) {
	t.Helper()
	after := readBindingRecords(t, f)
	if string(before.state) != string(after.state) || string(before.lapses) != string(after.lapses) {
		t.Fatal("refusal rewrote state.json or lapses.json")
	}
}
func retainedBindingLapse(t *testing.T, f fixture, before bindingRecords) {
	t.Helper()
	after := readBindingRecords(t, f)
	if string(before.state) != string(after.state) {
		t.Fatal("lapse changed state.json")
	}
	var record rootLapses
	if err := json.Unmarshal(after.lapses, &record); err != nil {
		t.Fatal(err)
	}
	for _, id := range record.CheckoutIDs {
		if id == f.registration.CheckoutID {
			return
		}
	}
	t.Fatal("observed identity change did not retain its lapse")
}
func unknownResult(t *testing.T, result any, err error) error {
	t.Helper()
	if err != nil && !reflect.ValueOf(result).IsZero() {
		t.Fatal("refusal returned usable result", result)
	}
	return err
}

// Claim: real refused Git starts deny every tested public entry point, preserve
// both registration records, and allow the same tuple once Git is readable.
// Final-recheck cases also reach Inspect/Preview/Approve's second observation.
// This proves refusal and recovery, not the cause of the hosted intermittent.
func TestRegistrationUnknownRead(t *testing.T) {
	for _, entry := range []string{"Inspect", "Preview", "Review", "Approve", "Admitted", "Locate", "LocateFor", "PackageInstruction", "ReadSnapshot", "ReviewContent"} {
		t.Run(entry, func(t *testing.T) {
			sites := []string{"rev-parse", "rev-list", "config"}
			if entry == "Inspect" || entry == "Preview" || entry == "Approve" {
				sites = append(sites, "final-recheck")
			}
			for index, site := range sites {
				t.Run(site, func(t *testing.T) {
					f := newFixture(t)
					id := f.registration.CheckoutID
					var preview Preview
					if entry != "Preview" {
						preview = packageApproved(t, f, false)
					}
					invoke := func() error {
						switch entry {
						case "Inspect":
							v, e := f.store.Inspect(id)
							return unknownResult(t, v, e)
						case "Preview":
							v, e := f.store.Preview(id, Declaration{1, []Request{}})
							return unknownResult(t, v, e)
						case "Review":
							v, e := f.store.Review(id)
							return unknownResult(t, v, e)
						case "Approve":
							v, e := f.store.Approve(id, "guide", preview.Review.Items[0].ItemSHA256, preview.PackageSHA256)
							return unknownResult(t, v, e)
						case "Admitted":
							v, e := f.store.Admitted(id, []Selection{{"guide", false}})
							return unknownResult(t, v, e)
						case "Locate":
							v, e := f.store.Locate(f.root)
							return unknownResult(t, v, e)
						case "LocateFor":
							v, e := f.store.LocateFor(f.root, "base", "codex_cli")
							return unknownResult(t, v, e)
						case "PackageInstruction":
							v, e := f.store.PackageInstruction(id, "guide", "AGENTS.md", true)
							return unknownResult(t, v, e)
						case "ReadSnapshot":
							v, files, e := f.store.ReadSnapshot(id, "guide")
							if e != nil && len(files) != 0 {
								t.Fatal("unknown returned snapshot files")
							}
							return unknownResult(t, v, e)
						case "ReviewContent":
							v, e := f.store.ReviewContent(id)
							return unknownResult(t, v, e)
						}
						panic("unhandled entry")
					}
					if err := invoke(); err != nil {
						t.Fatal("healthy control", err)
					}
					before := readBindingRecords(t, f)
					orig := os.Getenv("PATH")
					dir, _ := bindingGitPlant(t, index+1, "spawn")
					t.Setenv("PATH", dir+string(os.PathListSeparator)+orig)
					err := invoke()
					if !isCode(err, "project_registration_mismatch") || !strings.Contains(err.Error(), "start refused (EACCES)") {
						t.Fatalf("unknown did not refuse with its class: %v", err)
					}
					sameBindingRecords(t, f, before)
					t.Setenv("PATH", orig)
					if err := invoke(); err != nil {
						t.Fatal("same tuple did not recover", err)
					}
					// Reopen and re-read to ensure recovery did not silently rebind the ID.
					if got, err := f.store.Inspect(id); err != nil || got.Registration.CheckoutID != id {
						t.Fatal("recovery replaced registration", err)
					}
				})
			}
		})
	}
}

// Claim: a real root, remote or unique root-ancestor change remains a durable
// lapse even when preceded or temporarily hidden by an unknown read. Restoring
// the tuple cannot revive it. A no-such-key config answer is an observed change.
func TestRegistrationChangeAfterUnknown(t *testing.T) {
	for _, kind := range []string{"changed-root", "changed-remote", "removed-remote", "changed-root-ancestor"} {
		for _, when := range []string{"during", "after"} {
			t.Run(kind+"/"+when, func(t *testing.T) {
				f := newFixture(t)
				gitTest(t, f.root, "config", "remote.origin.url", "https://example.invalid/original.git")
				reg, err := f.store.Rebind(f.registration.CheckoutID, f.root)
				if err != nil {
					t.Fatal(err)
				}
				f.registration = reg
				id := reg.CheckoutID
				if _, err := f.store.Inspect(id); err != nil {
					t.Fatal("healthy control", err)
				}
				before := readBindingRecords(t, f)
				config := filepath.Join(f.root, ".git", "config")
				original, err := os.ReadFile(config)
				if err != nil {
					t.Fatal(err)
				}
				restore := func() {}
				change := func() {
					switch kind {
					case "changed-remote":
						gitTest(t, f.root, "config", "remote.origin.url", "https://example.invalid/changed.git")
						restore = func() { write(t, config, string(original)) }
					case "removed-remote":
						gitTest(t, f.root, "config", "--unset-all", "remote.origin.url")
						restore = func() { write(t, config, string(original)) }
					case "changed-root-ancestor":
						head := gitTest(t, f.root, "rev-parse", "HEAD")
						replacement := filepath.Join(t.TempDir(), "replacement")
						initRepo(t, replacement)
						write(t, filepath.Join(replacement, "different"), "different root\n")
						gitTest(t, replacement, "add", "different")
						gitTest(t, replacement, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "--amend", "-qm", "different root")
						other := gitTest(t, replacement, "rev-parse", "HEAD")
						gitTest(t, f.root, "fetch", replacement, "HEAD")
						gitTest(t, f.root, "update-ref", "HEAD", other)
						restore = func() { gitTest(t, f.root, "update-ref", "HEAD", head) }
					case "changed-root":
						moved := f.root + "-saved"
						if err := os.Rename(f.root, moved); err != nil {
							t.Fatal(err)
						}
						if err := os.Mkdir(f.root, 0700); err != nil {
							t.Fatal(err)
						}
						if err := os.Rename(filepath.Join(moved, ".git"), filepath.Join(f.root, ".git")); err != nil {
							t.Fatal(err)
						}
						write(t, filepath.Join(f.root, "README.md"), "Fixture\n")
						restore = func() {
							if err := os.Rename(filepath.Join(f.root, ".git"), filepath.Join(moved, ".git")); err != nil {
								t.Fatal(err)
							}
							if err := os.Rename(f.root, f.root+"-replaced"); err != nil {
								t.Fatal(err)
							}
							if err := os.Rename(moved, f.root); err != nil {
								t.Fatal(err)
							}
						}
					}
				}
				if when == "during" {
					change()
				}
				orig := os.Getenv("PATH")
				dir, _ := bindingGitPlant(t, 1, "spawn")
				t.Setenv("PATH", dir+string(os.PathListSeparator)+orig)
				_, err = f.store.Inspect(id)
				if !isCode(err, "project_registration_mismatch") || !strings.Contains(err.Error(), "start refused (EACCES)") {
					t.Fatal("unknown authorized or lost class", err)
				}
				sameBindingRecords(t, f, before)
				t.Setenv("PATH", orig)
				if when == "after" {
					change()
				}
				_, err = f.store.Inspect(id)
				want := map[string]string{"changed-root": "filesystem.inode", "changed-remote": "repository.primary_remote", "removed-remote": "repository.primary_remote", "changed-root-ancestor": "repository.root_commit"}[kind]
				if !isCode(err, "project_registration_mismatch") || !strings.Contains(err.Error(), "checkout binding changed:") || !strings.Contains(err.Error(), want) {
					t.Fatalf("change did not refuse by field %s: %v", want, err)
				}
				retainedBindingLapse(t, f, before)
				restore()
				_, err = f.store.Inspect(id)
				if !isCode(err, "project_registration_mismatch") || !strings.Contains(err.Error(), "authority lapsed") {
					t.Fatal("restored tuple revived a lapse", err)
				}
				retainedBindingLapse(t, f, before)
			})
		}
	}
}

// Claim: the three explicitly undecided cases retain their existing durable
// refuse-and-lapse behavior, including after restoration. This pins the current
// conservative default; it does not resolve the open normative questions.
func TestRegistrationOpenLapseCases(t *testing.T) {
	for _, kind := range []string{"missing-root", "not-a-repository", "multiple-roots"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			id := f.registration.CheckoutID
			if _, err := f.store.Inspect(id); err != nil {
				t.Fatal("healthy control", err)
			}
			before := readBindingRecords(t, f)
			restore := func() {}
			switch kind {
			case "missing-root":
				moved := f.root + "-moved"
				if err := os.Rename(f.root, moved); err != nil {
					t.Fatal(err)
				}
				restore = func() {
					if err := os.Rename(moved, f.root); err != nil {
						t.Fatal(err)
					}
				}
			case "not-a-repository":
				original := filepath.Join(f.root, ".git")
				moved := f.root + "-git"
				if err := os.Rename(original, moved); err != nil {
					t.Fatal(err)
				}
				restore = func() {
					if err := os.Rename(moved, original); err != nil {
						t.Fatal(err)
					}
				}
			case "multiple-roots":
				original := gitTest(t, f.root, "rev-parse", "HEAD")
				other := filepath.Join(t.TempDir(), "other")
				initRepo(t, other)
				write(t, filepath.Join(other, "different"), "another root\n")
				gitTest(t, other, "add", "different")
				gitTest(t, other, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "--amend", "-qm", "another root")
				second := gitTest(t, other, "rev-parse", "HEAD")
				gitTest(t, f.root, "fetch", other, "HEAD")
				tree := gitTest(t, f.root, "rev-parse", "HEAD^{tree}")
				merge := gitTest(t, f.root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit-tree", tree, "-p", original, "-p", second, "-m", "two roots")
				gitTest(t, f.root, "update-ref", "HEAD", merge)
				restore = func() { gitTest(t, f.root, "update-ref", "HEAD", original) }
			}
			_, err := f.store.Inspect(id)
			required(t, err, "project_registration_mismatch")
			retainedBindingLapse(t, f, before)
			restore()
			_, err = f.store.Inspect(id)
			if !isCode(err, "project_registration_mismatch") || !strings.Contains(err.Error(), "authority lapsed") {
				t.Fatal("open-case lapse was renewed", err)
			}
			retainedBindingLapse(t, f, before)
		})
	}
}

// Claim: unavailable child-CWD observation still refuses; a failed metadata
// read cannot lapse its healthy registered ancestor or authorize the child.
// After Git recovers, the root remains usable and the missing child refuses.
func TestRegistrationUnavailableCWDUnknownRead(t *testing.T) {
	for index, site := range []string{"rev-parse", "rev-list", "config"} {
		t.Run(site, func(t *testing.T) {
			f := newFixture(t)
			id := f.registration.CheckoutID
			if _, err := f.store.Locate(f.root); err != nil {
				t.Fatal("root control", err)
			}
			before := readBindingRecords(t, f)
			orig := os.Getenv("PATH")
			dir, _ := bindingGitPlant(t, index+1, "spawn")
			t.Setenv("PATH", dir+string(os.PathListSeparator)+orig)
			v, err := f.store.Locate(filepath.Join(f.root, "missing-child"))
			if v.CheckoutID != "" || !isCode(err, "project_registration_mismatch") || !strings.Contains(err.Error(), "start refused (EACCES)") {
				t.Fatal("unknown CWD authorized or lost class", err)
			}
			sameBindingRecords(t, f, before)
			t.Setenv("PATH", orig)
			v, err = f.store.Locate(filepath.Join(f.root, "missing-child"))
			if v.CheckoutID != "" || !isCode(err, "project_registration_mismatch") || !strings.Contains(err.Error(), "CWD is unreadable") {
				t.Fatal("missing child authorized after Git recovered", err)
			}
			sameBindingRecords(t, f, before)
			if v, err := f.store.Inspect(id); err != nil || v.Registration.CheckoutID != id {
				t.Fatal("root failed to recover with same ID", err)
			}
		})
	}
}

// Claim: local-manager composition preview, publication and held-lock admission
// also refuse unknown metadata without lapsing or returning usable authority.
// Existing records stay intact and the same tuple recovers without rebind.
func TestCompositionUnknownRead(t *testing.T) {
	for _, entry := range []string{"preview", "publish", "resolve", "authorize"} {
		t.Run(entry, func(t *testing.T) {
			f := compositionFixture(t, "codex_cli", nil)
			packageApproved(t, f, false)
			req := compositionRequest(f)
			a := Authority{Store: f.store, CWD: f.root}
			owner := principal(t, f)
			p, err := f.store.PreviewComposition(context.Background(), req)
			if err != nil {
				t.Fatal("preview control", err)
			}
			c, err := a.PublishComposition(context.Background(), req, p.SHA256, owner)
			if err != nil {
				t.Fatal("publication control", err)
			}
			invoke := func() error {
				switch entry {
				case "preview":
					v, e := f.store.PreviewComposition(context.Background(), req)
					return unknownResult(t, v, e)
				case "publish":
					v, e := a.PublishComposition(context.Background(), req, p.SHA256, owner)
					return unknownResult(t, v, e)
				case "resolve":
					v, e := resolveComposition(t, a, c.Ref, c.SHA256)
					return unknownResult(t, v, e)
				case "authorize":
					return authorizeComposition(t, a, c, owner)
				}
				panic("unhandled composition entry")
			}
			if err := invoke(); err != nil {
				t.Fatal("healthy control", err)
			}
			before := readBindingRecords(t, f)
			record := filepath.Join(f.store.Home, "project-context", "compositions", c.SHA256+".json")
			bytes, err := os.ReadFile(record)
			if err != nil {
				t.Fatal(err)
			}
			orig := os.Getenv("PATH")
			dir, _ := bindingGitPlant(t, 1, "spawn")
			t.Setenv("PATH", dir+string(os.PathListSeparator)+orig)
			err = invoke()
			if !isCode(err, "project_registration_mismatch") || !strings.Contains(err.Error(), "start refused (EACCES)") {
				t.Fatal("unknown composition authorized or lost class", err)
			}
			sameBindingRecords(t, f, before)
			after, err := os.ReadFile(record)
			if err != nil || string(after) != string(bytes) {
				t.Fatal("unknown rewrote composition", err)
			}
			t.Setenv("PATH", orig)
			if err := invoke(); err != nil {
				t.Fatal("same tuple composition did not recover", err)
			}
		})
	}
}
