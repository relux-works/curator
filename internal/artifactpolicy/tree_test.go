package artifactpolicy

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDependencyDirectoryAdmissionAndPortableIdentity(t *testing.T) {
	first := makeSourceTree(t)
	second := makeSourceTree(t)
	descriptor := fixtureDescriptor(nil, ProfileGoV1)
	firstDigest := treeDigestFromRejected(t, first, "source", descriptor)
	secondDigest := treeDigestFromRejected(t, second, "source", descriptor)
	if firstDigest != secondDigest {
		t.Fatalf("location-independent tree digests differ: %s != %s", firstDigest, secondDigest)
	}
	descriptor.Origin = OriginEvidence{
		Locator: "fixture://tree", ImmutableID: "tree-revision-1", LockRecord: "tree-lock-1",
		ChecksumSHA256: firstDigest, Verified: true,
	}
	result, err := NewService().AdmitDependencyDirectory(t.Context(), DirectoryRequest{
		Descriptor: descriptor, Root: first, VirtualRoot: "source",
	})
	if err != nil {
		t.Fatal(err)
	}
	requireDecision(t, result, DecisionAdmitInput)
	if result.Manifest.RawPayload.Kind != "canonical_tree" || result.Manifest.RawPayload.SHA256 != firstDigest {
		t.Fatalf("tree payload identity = %+v", result.Manifest.RawPayload)
	}
	if got := requireNode(t, result, "source/cmd/tool/main.go").Class; got != ClassSourceAuthoredText {
		t.Fatalf("tree source class = %q", got)
	}
	if result.Admission == nil {
		t.Fatal("admitted tree returned no token")
	}
}

func TestDependencyDirectoryRejectsCompiledLinkAndBundleNodes(t *testing.T) {
	t.Run("extensionless compiled bytes", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "tool"), makeMachO(2), 0o644); err != nil {
			t.Fatal(err)
		}
		descriptor := fixtureDescriptor(nil, ProfileCommonV1)
		digest := treeDigestFromRejected(t, root, "tree", descriptor)
		descriptor.Origin = OriginEvidence{Locator: "fixture://tree", ImmutableID: "r1", LockRecord: "l1", ChecksumSHA256: digest, Verified: true}
		result, err := NewService().AdmitDependencyDirectory(t.Context(), DirectoryRequest{Descriptor: descriptor, Root: root, VirtualRoot: "tree"})
		requireCode(t, err, CodeCompiledDependency)
		if got := requireNode(t, result, "tree/tool").Class; got != ClassNativeExecutable {
			t.Fatalf("compiled tree class = %q", got)
		}
	})

	t.Run("framework bundle root", func(t *testing.T) {
		root := t.TempDir()
		bundle := filepath.Join(root, "Thing.framework")
		if err := os.Mkdir(bundle, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(bundle, "README.txt"), []byte("resource\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		descriptor := fixtureDescriptor(nil, ProfileSwiftPMV1)
		digest := treeDigestFromRejected(t, root, "tree", descriptor)
		descriptor.Origin = OriginEvidence{Locator: "fixture://tree", ImmutableID: "r1", LockRecord: "l1", ChecksumSHA256: digest, Verified: true}
		result, err := NewService().AdmitDependencyDirectory(t.Context(), DirectoryRequest{Descriptor: descriptor, Root: root, VirtualRoot: "tree"})
		requireCode(t, err, CodeCompiledDependency)
		if got := requireNode(t, result, "tree/Thing.framework").Class; got != ClassAppleFramework {
			t.Fatalf("framework class = %q", got)
		}
	})

	if runtime.GOOS != "windows" {
		t.Run("filesystem symlink", func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "target.go"), []byte("package target\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink("target.go", filepath.Join(root, "link.go")); err != nil {
				t.Fatal(err)
			}
			descriptor := fixtureDescriptor(nil, ProfileGoV1)
			digest := treeDigestFromRejected(t, root, "tree", descriptor)
			descriptor.Origin = OriginEvidence{Locator: "fixture://tree", ImmutableID: "r1", LockRecord: "l1", ChecksumSHA256: digest, Verified: true}
			result, err := NewService().AdmitDependencyDirectory(t.Context(), DirectoryRequest{Descriptor: descriptor, Root: root, VirtualRoot: "tree"})
			requireCode(t, err, CodeArchiveUnsafeEntry)
			if got := requireNode(t, result, "tree/link.go").Class; got != ClassLink {
				t.Fatalf("symlink class = %q", got)
			}
		})

		t.Run("filesystem hardlink", func(t *testing.T) {
			root := t.TempDir()
			first := filepath.Join(root, "first.go")
			if err := os.WriteFile(first, []byte("package first\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.Link(first, filepath.Join(root, "second.go")); err != nil {
				t.Fatal(err)
			}
			descriptor := fixtureDescriptor(nil, ProfileGoV1)
			digest := treeDigestFromRejected(t, root, "tree", descriptor)
			descriptor.Origin = OriginEvidence{Locator: "fixture://tree", ImmutableID: "r1", LockRecord: "l1", ChecksumSHA256: digest, Verified: true}
			result, err := NewService().AdmitDependencyDirectory(t.Context(), DirectoryRequest{Descriptor: descriptor, Root: root, VirtualRoot: "tree"})
			requireCode(t, err, CodeArchiveUnsafeEntry)
			if requireNode(t, result, "tree/first.go").Class != ClassLink || requireNode(t, result, "tree/second.go").Class != ClassLink {
				t.Fatal("hardlink paths were not both classified as links")
			}
		})
	}
}

func makeSourceTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	directory := filepath.Join(root, "cmd", "tool")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module fixture.test/tool\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "LICENSE"), []byte("Fixture license.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}
