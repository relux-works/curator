package install

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func r5Run(t *testing.T, e *env, global bool, userHome string, opts Options) Result {
	if global {
		opts.Platform = installPlatform()
		return Global(e.cfg, userHome, opts)
	}
	return e.install(opts)
}

func r5Setup(t *testing.T, global bool) (*env, string, string) {
	e := newEnv(t)
	first := setupLegacyScriptRepo(e, "consumer", "ctool", "#!/bin/sh\necho consumer-ok\n")
	root := e.project
	if global {
		e.globalDeclare("consumer")
		root = GlobalRoot(e.home)
	} else {
		e.declare("consumer")
	}
	return e, first, root
}

func r5Move(e *env, alias string, from string) string {
	repo := filepath.Join(e.skillsRoot, "consumer")
	e.git(repo, "tag", alias, from)
	e.write(repo, "references/m-"+alias+".md", alias+"\n")
	e.git(repo, "add", ".")
	e.git(repo, "commit", "-qm", "move "+alias)
	e.git(repo, "tag", "-f", "v1")
	return e.gitOutput(repo, "rev-parse", "HEAD")
}

// Evidence must be refreshed by a non-strict rebind: a second move with a new alias still refuses,
// in dry-run planning as well as in a real strict install.
func TestReviewR5RefreshedEvidenceThenMoveAgain(t *testing.T) {
	for _, global := range []bool{false, true} {
		name := "project"
		if global {
			name = "global"
		}
		t.Run(name, func(t *testing.T) {
			e, first, _ := r5Setup(t, global)
			home := t.TempDir()
			if r := r5Run(t, e, global, home, Options{StrictTags: true}); r.Status != "ok" {
				t.Fatalf("initial: %+v", r)
			}
			second := r5Move(e, "archive1", first)
			if r := r5Run(t, e, global, home, Options{StrictTags: true, DryRun: true}); r.Status != "failed" || !strings.Contains(strings.Join(r.Errors, "\n"), "moved tag for consumer: v1 "+first+" -> "+second) {
				t.Fatalf("dry strict after first move must refuse: %+v", r)
			}
			if r := r5Run(t, e, global, home, Options{}); r.Status != "ok" {
				t.Fatalf("non-strict rebind: %+v", r)
			}
			third := r5Move(e, "archive2", second)
			if r := r5Run(t, e, global, home, Options{StrictTags: true, DryRun: true}); r.Status != "failed" || !strings.Contains(strings.Join(r.Errors, "\n"), "moved tag for consumer: v1 "+second+" -> "+third) {
				t.Fatalf("dry strict after second move must refuse: %+v", r)
			}
			if r := r5Run(t, e, global, home, Options{StrictTags: true}); r.Status != "failed" || !strings.Contains(strings.Join(r.Errors, "\n"), "moved tag for consumer: v1 "+second+" -> "+third) {
				t.Fatalf("strict after second move must refuse: %+v", r)
			}
		})
	}
}

// A same-commit declaration change v1 -> v2 must refresh the recorded declaration: moving v2 later
// refuses as v2, moving the abandoned v1 later is not a moved declared tag.
func TestReviewR5SameCommitRedeclareThenMove(t *testing.T) {
	for _, global := range []bool{false, true} {
		name := "project"
		if global {
			name = "global"
		}
		t.Run(name, func(t *testing.T) {
			e, first, root := r5Setup(t, global)
			home := t.TempDir()
			if r := r5Run(t, e, global, home, Options{StrictTags: true}); r.Status != "ok" {
				t.Fatalf("initial: %+v", r)
			}
			repo := filepath.Join(e.skillsRoot, "consumer")
			e.git(repo, "tag", "v2", first)
			path := filepath.Join(root, "Skillfile.json")
			payload, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			e.write(root, "Skillfile.json", string(bytes.ReplaceAll(payload, []byte(`"v1"`), []byte(`"v2"`))))
			if r := r5Run(t, e, global, home, Options{StrictTags: true}); r.Status != "ok" {
				t.Fatalf("redeclare: %+v", r)
			}
			// move abandoned v1: not a declared tag any more
			e.write(repo, "references/x.md", "x\n")
			e.git(repo, "add", ".")
			e.git(repo, "commit", "-qm", "x")
			e.git(repo, "tag", "-f", "v1")
			if r := r5Run(t, e, global, home, Options{StrictTags: true}); r.Status != "ok" || strings.Contains(strings.Join(r.Messages, "\n"), "moved tag") {
				t.Fatalf("moving abandoned v1 must not refuse: %+v", r)
			}
			// move declared v2
			e.write(repo, "references/y.md", "y\n")
			e.git(repo, "add", ".")
			e.git(repo, "commit", "-qm", "y")
			e.git(repo, "tag", "-f", "v2")
			moved := e.gitOutput(repo, "rev-parse", "HEAD")
			if r := r5Run(t, e, global, home, Options{StrictTags: true}); r.Status != "failed" || !strings.Contains(strings.Join(r.Errors, "\n"), "moved tag for consumer: v2 ") || !strings.Contains(strings.Join(r.Errors, "\n"), moved) {
				t.Fatalf("moving declared v2 must refuse: %+v", r)
			}
		})
	}
}
