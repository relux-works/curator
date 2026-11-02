package testhomeguard

import (
	"os"
	"path/filepath"
	"testing"
)

// The exception requires the native registration, directory identities and a
// regular coordination leaf. Name/ID collisions never exclude copied payloads,
// even when they mimic an otherwise registered runtime bucket.
func TestRuntimeMetadataLayout(t *testing.T) {
	for _, shape := range []string{"registered", "unregistered", "ordinary-name", "copied-synced", "manifest-directory", "manifest-link", "registration-link", "registration-directory"} {
		t.Run(shape, func(t *testing.T) {
			root := t.TempDir()
			bucket := fixtureBucketID
			if shape == "ordinary-name" {
				bucket = "references"
			}
			namespace := filepath.Join(root, "skills", "synced")
			leaf := filepath.Join(namespace, bucket, "manifest.json")
			must := func(err error) {
				t.Helper()
				if err != nil {
					t.Fatal(err)
				}
			}
			must(os.MkdirAll(filepath.Dir(leaf), 0700))
			registration := filepath.Join(namespace, ".bucket-"+bucket)
			switch shape {
			case "unregistered":
			case "registration-link":
				if err := os.Symlink("missing", registration); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
			case "registration-directory":
				must(os.Mkdir(registration, 0700))
			default:
				must(os.WriteFile(registration, []byte("synthetic registration"), 0600))
			}
			if shape == "copied-synced" {
				must(os.WriteFile(filepath.Join(namespace, "SKILL.md"), []byte("copied skill"), 0600))
			}
			if shape == "manifest-directory" {
				must(os.Mkdir(leaf, 0700))
				leaf = filepath.Join(leaf, "SKILL.md")
			}
			if shape == "manifest-link" {
				if err := os.Symlink("before", leaf); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
			} else {
				must(os.WriteFile(leaf, []byte("before"), 0600))
			}
			before, err := snapshot([]string{root})
			must(err)
			if shape == "manifest-link" {
				must(os.Remove(leaf))
				must(os.Symlink("after!", leaf))
			} else {
				info, err := os.Stat(leaf)
				must(err)
				must(os.WriteFile(leaf, []byte("after!"), 0600))
				must(os.Chtimes(leaf, info.ModTime(), info.ModTime()))
			}
			after, err := snapshot([]string{root})
			must(err)
			if same(before, after) != (shape == "registered") {
				t.Fatal("runtime layout exception did not distinguish registered metadata from payload")
			}
		})
	}
}

// Referenced context uses the same no-follow and credential metadata rules as
// skills. The container identity is observed before descending into modules.
func TestReferencedContextBoundary(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".agent-context", "modules", "fixture")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"auth.json", "instruction.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("synthetic"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	image, err := snapshot([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	if image[filepath.Join(dir, "auth.json")].hash != ([32]byte{}) || image[filepath.Join(dir, "instruction.md")].hash == ([32]byte{}) {
		t.Fatal("referenced context credential/instruction rules diverged")
	}
	other := t.TempDir()
	if err := os.Symlink(filepath.Join(root, ".agent-context"), filepath.Join(other, ".agent-context")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	image, err = snapshot([]string{other})
	if err != nil {
		t.Fatal(err)
	}
	if _, followed := image[filepath.Join(other, ".agent-context", "modules")]; followed {
		t.Fatal("linked context container traversed")
	}
	if image[filepath.Join(other, ".agent-context")].link == "" {
		t.Fatal("linked container identity missing")
	}
}
