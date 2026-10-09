package install

import (
	"encoding/json"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Real Project and Global entries must distinguish changed declarations from moved tags.
func TestReviewChangedDeclaredTagIsNotMovedTag(t *testing.T) {
	for _, global := range []bool{false, true} {
		label := "project"
		if global {
			label = "global"
		}
		t.Run(label, func(t *testing.T) {
			e := newEnv(t)
			setupLegacyScriptRepo(e, "consumer", "ctool", "#!/bin/sh\necho consumer-ok\n")
			root := e.project
			if global {
				e.globalDeclare("consumer")
				root = GlobalRoot(e.home)
			} else {
				e.declare("consumer")
			}
			run := func() Result {
				if global {
					return Global(e.cfg, t.TempDir(), Options{Platform: installPlatform(), StrictTags: true})
				}
				return e.install(Options{StrictTags: true})
			}
			if first := run(); first.Status != "ok" {
				t.Fatalf("first install: %+v", first)
			}
			repo := filepath.Join(e.skillsRoot, "consumer")
			e.write(repo, "references/new.md", "new release\n")
			e.git(repo, "add", ".")
			e.git(repo, "commit", "-qm", "release two")
			e.git(repo, "tag", "v2")
			path := filepath.Join(root, "Skillfile.json")
			payload, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var doc map[string]any
			if err := json.Unmarshal(payload, &doc); err != nil {
				t.Fatal(err)
			}
			doc["skills"].([]any)[0].(map[string]any)["tag"] = "v2"
			payload, err = json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, payload, 0600); err != nil {
				t.Fatal(err)
			}
			second := run()
			if second.Status != "ok" {
				t.Fatalf("unchanged v1 and v2 tags; explicit v1->v2 declaration must install under strict-tag policy: %+v", second)
			}
			if strings.Contains(strings.Join(second.Messages, "\n"), "moved tag") {
				t.Fatalf("unchanged distinct tag reported as moved: %+v", second)
			}
		})
	}
}

func TestReviewLegacyLockBindsWholeManifest(t *testing.T) {
	e := newEnv(t)
	setupLegacyDirectoryRepos(e, "skills/backend", "skills/helper")
	e.declare("consumer")
	if r := e.install(Options{}); r.Status != "ok" {
		t.Fatalf("first: %+v", r)
	}
	original := marker.Read(filepath.Join(e.project, ".agents", "skills", "backend"))
	if original == nil {
		t.Fatal("missing marker")
	}
	path := filepath.Join(e.project, manifest.Name)
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	doc["locale"] = "en"
	payload, err = json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, payload, 0600); err != nil {
		t.Fatal(err)
	}
	if r := e.install(Options{}); r.Status != "ok" {
		t.Fatalf("reinstall: %+v", r)
	}
	current := marker.Read(filepath.Join(e.project, ".agents", "skills", "backend"))
	if current == nil {
		t.Fatal("missing current marker")
	}
	if original.LockSHA256 == current.LockSHA256 {
		t.Fatal("manifest-only edit did not rebind lock digest")
	}
}

func TestReviewGlobalBuildPackageReceipt(t *testing.T) {
	e := newEnv(t)
	setupLegacyBuildRepo(e, "tooling")
	e.globalDeclare("tooling")
	deps, _, _ := draftBuildDeps(t)
	result := Global(e.cfg, t.TempDir(), Options{Platform: installPlatform(), Build: deps})
	if result.Status != "ok" {
		t.Fatalf("global build: %+v", result)
	}
	m := marker.Read(filepath.Join(GlobalRoot(e.home), "skills", "tooling"))
	if m == nil || m.Package == nil || m.SchemaVersion != 6 || m.Builds["ltool"].ReceiptSchemaVersion != 3 {
		t.Fatalf("global package receipt: %+v", m)
	}
}
