package main

import (
	"encoding/json"
	"github.com/relux-works/curator/internal/marker"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A second real CLI audit with identical bytes under another repository cannot adopt the first verdict.
func TestReviewCLIAuditRepositoryChangeMissesCache(t *testing.T) {
	cfg, home, _, _, _ := auditDirectoryProject(t)
	if code, out, err := capture(t, cfg, "audit", "app"); code != exitOK {
		t.Fatalf("initial audit: %d %s %s", code, out, err)
	}
	files, err := filepath.Glob(filepath.Join(home, "audit", "*", "verdict-*.json"))
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var doc map[string]any
		if err := json.Unmarshal(b, &doc); err != nil {
			t.Fatal(err)
		}
		if doc["skill"] != "backend" {
			continue
		}
		pkg := doc["package"].(map[string]any)
		pkg["repository"] = "github.com/example/other-skills"
		b, err = json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, b, 0600); err != nil {
			t.Fatal(err)
		}
		changed = true
	}
	if !changed {
		t.Fatal("backend record absent")
	}
	if code, out, err := capture(t, cfg, "audit", "app"); code != exitOK {
		t.Fatalf("second audit: %d %s %s", code, out, err)
	}
	files, err = filepath.Glob(filepath.Join(home, "audit", "*", "verdict-*.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var doc map[string]any
		if err := json.Unmarshal(b, &doc); err != nil {
			t.Fatal(err)
		}
		if doc["skill"] == "backend" && doc["package"].(map[string]any)["repository"] != "github.com/example/role-skills" {
			t.Fatalf("CLI reused mismatching repository cache: %s", b)
		}
	}
}

func TestReviewCLIStatusDetectsChangedTagAtSameCommit(t *testing.T) {
	cfg, home, project, _, _ := auditDirectoryProject(t)
	if code, out, err := capture(t, cfg, "install", "app"); code != exitOK {
		t.Fatalf("install: %d %s %s", code, out, err)
	}
	runGit(t, filepath.Join(filepath.Dir(home), "skills", "consumer"), "tag", "v2")
	path := filepath.Join(project, "Skillfile.json")
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
	code, out, stderr := capture(t, cfg, "status", "app", "--json")
	if code != exitOK {
		t.Fatalf("status: %d %s %s", code, out, stderr)
	}
	report := decodeStatus(t, out)
	if report.Skills["consumer"] == stateUpToDate {
		t.Errorf("changed declaration is incorrectly current: %s", out)
	}
	if code, out, stderr := capture(t, cfg, "status", "app", "--check"); code == exitOK {
		t.Errorf("status --check admitted stale ref binding: %s %s", out, stderr)
	}
}

func TestReviewCLIRealSchema9BuildAndLaunch(t *testing.T) {
	requireNativeControlInventoryPlatform(t)
	project, home := compiledProject(t)
	repo := filepath.Join(filepath.Dir(home), "skills", "build-skill")
	path := filepath.Join(repo, "agent-skill.json")
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	payload = []byte(strings.Replace(string(payload), "\"schema_version\": 6", "\"schema_version\": 9", 1))
	if err := os.WriteFile(path, payload, 0600); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-qm", "schema9 build")
	runGit(t, repo, "tag", "-f", "v1")
	cfg := filepath.Join(home, "config.json")
	if code, out, err := capture(t, cfg, "install", "app"); code != exitOK {
		t.Fatalf("real schema9 build: %d %s %s", code, out, err)
	}
	m := marker.Read(filepath.Join(project, ".agents", "skills", "build-skill"))
	if m == nil || m.SchemaVersion != 6 || m.Package == nil || m.Builds["build-tool"].ReceiptSchemaVersion != 3 {
		t.Fatalf("real build marker: %+v", m)
	}
	out, err := exec.Command(filepath.Join(project, ".agents", "bin", "build-tool")).CombinedOutput()
	if err != nil || !strings.Contains(string(out), "curator") {
		t.Fatalf("compiled launch: %v %s", err, out)
	}
	if code, out, err := capture(t, cfg, "install", "app"); code != exitOK {
		t.Fatalf("real build reinstall: %d %s %s", code, out, err)
	}
	if code, out, err := capture(t, cfg, "status", "app", "--check"); code != exitOK {
		t.Fatalf("compiled currentness: %d %s %s", code, out, err)
	}
}
