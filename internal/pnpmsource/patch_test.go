package pnpmsource

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClosedUnifiedPatchTransformModifiesCreatesAndDeletes(t *testing.T) {
	original := map[string][]byte{
		"a.txt":   []byte("one\ntwo\nthree\n"),
		"old.txt": []byte("old\n"),
	}
	patch := []byte(`diff --git a/a.txt b/a.txt
--- a/a.txt
+++ b/a.txt
@@ -1,3 +1,3 @@
 one
-two
+TWO
 three
diff --git a/new.txt b/new.txt
new file mode 100644
--- /dev/null
+++ b/new.txt
@@ -0,0 +1,2 @@
+new
+file
diff --git a/old.txt b/old.txt
deleted file mode 100644
--- a/old.txt
+++ /dev/null
@@ -1 +0,0 @@
-old
`)
	result, err := applyUnifiedPatch(original, patch)
	if err != nil {
		t.Fatal(err)
	}
	if string(result["a.txt"]) != "one\nTWO\nthree\n" || string(result["new.txt"]) != "new\nfile\n" {
		t.Fatalf("unexpected patched files: %#v", result)
	}
	if _, exists := result["old.txt"]; exists {
		t.Fatal("deleted patch target remains in expected inventory")
	}
	if string(original["a.txt"]) != "one\ntwo\nthree\n" {
		t.Fatal("patch transform mutated admitted source inventory")
	}
}

func TestClosedUnifiedPatchTransformRejectsUnsupportedOrDriftedInputs(t *testing.T) {
	original := map[string][]byte{"a.txt": []byte("one\n")}
	for _, testCase := range []struct {
		name  string
		patch string
	}{
		{"empty", ""},
		{"escaping path", "diff --git a/../a b/../a\n--- a/../a\n+++ b/../a\n@@ -1 +1 @@\n-one\n+two\n"},
		{"metadata mutation", "diff --git a/package.json b/package.json\n--- a/package.json\n+++ b/package.json\n@@ -1 +1 @@\n-old\n+new\n"},
		{"source drift", "diff --git a/a.txt b/a.txt\n--- a/a.txt\n+++ b/a.txt\n@@ -1 +1 @@\n-other\n+two\n"},
		{"rename", "diff --git a/a.txt b/b.txt\nrename from a.txt\nrename to b.txt\n--- a/a.txt\n+++ b/b.txt\n@@ -1 +1 @@\n-one\n+two\n"},
		{"no newline marker", "diff --git a/a.txt b/a.txt\n--- a/a.txt\n+++ b/a.txt\n@@ -1 +1 @@\n-one\n\\ No newline at end of file\n+two\n"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := applyUnifiedPatch(original, []byte(testCase.patch)); err == nil {
				t.Fatal("invalid patch was accepted")
			}
		})
	}
}

func TestVirtualStoreMemberAndDirectoryEncoding(t *testing.T) {
	root := t.TempDir()
	scope := filepath.Join(root, "@scope")
	if err := os.MkdirAll(filepath.Join(scope, "package"), 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "dependency")
	if err := os.MkdirAll(target, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(scope, "dependency")); err != nil {
		t.Fatal(err)
	}
	members, err := readVirtualPackageMembers(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 2 || members["@scope/package"] == "" || members["@scope/dependency"] == "" {
		t.Fatalf("scoped members not enumerated exactly: %+v", members)
	}
	if got := pnpmSnapshotDirectory("@scope/package@1.0.0(peer@2.0.0)"); got != "@scope+package@1.0.0_peer@2.0.0" {
		t.Fatalf("snapshot directory = %q", got)
	}
	long := pnpmSnapshotDirectory("MixedCase@1.0.0(" + strings.Repeat("peer@1.0.0", 20) + ")")
	if len(long) != pnpmVirtualStoreDirMaxLength || !strings.Contains(long, "_") {
		t.Fatalf("long/mixed-case snapshot directory is not pinned: %q", long)
	}
}
