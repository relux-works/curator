package install

import (
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/sourcelock"
	"os"
	"path/filepath"
	"testing"
)

func TestReviewLegacyRuntimeMatchesPackageMarker(t *testing.T) {
	e := newEnv(t)
	consumer := filepath.Join(e.skillsRoot, "consumer")
	writeDraftScriptSkill(t, consumer, "consumer", "ctool", "#!/bin/sh\necho consumer-ok\n", func(spec map[string]any) { spec["schema_version"] = 9 })
	e.git(consumer, "init", "-q", "-b", "main")
	e.git(consumer, "add", ".")
	e.git(consumer, "commit", "-qm", "schema9 script")
	e.git(consumer, "tag", "v1")
	e.declare("consumer")
	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("install: %+v", result)
	}
	recorded, _ := readLegacyMarker(t, filepath.Join(e.project, ".agents", "skills", "consumer"))
	pkg := sourcelock.Package{Kind: recorded.Package.Kind, Source: recorded.Package.Source, Directory: recorded.Package.Directory, Commit: sourcelock.Commit{ObjectFormat: recorded.Package.Commit.ObjectFormat, Hex: recorded.Package.Commit.Hex}}
	digest, err := pkg.Digest()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := runtimestore.SourceV1Dir(e.home, "consumer", digest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "scripts", "ctool.sh")); err != nil {
		t.Fatalf("package marker runtime missing under source-v1 namespace: %v", err)
	}
}
