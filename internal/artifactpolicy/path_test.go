package artifactpolicy

import (
	"strings"
	"testing"
)

func TestPortableVirtualPathAndCollisionKey(t *testing.T) {
	valid, err := ValidateVirtualPath("Sources/Éclair/main.swift")
	if err != nil {
		t.Fatal(err)
	}
	if valid.Canonical != "Sources/Éclair/main.swift" || len(valid.Components) != 3 {
		t.Fatalf("validated path = %+v", valid)
	}
	left, err := ValidateVirtualPath("Readme.TXT")
	if err != nil {
		t.Fatal(err)
	}
	right, err := ValidateVirtualPath("README.txt")
	if err != nil {
		t.Fatal(err)
	}
	if left.CollisionKey != right.CollisionKey {
		t.Fatalf("portable case-fold keys differ: %q != %q", left.CollisionKey, right.CollisionKey)
	}
	if _, err := ValidateVirtualPath("Sources/e\u0301clair.swift"); err == nil {
		t.Fatal("non-NFC path admitted")
	}
	if _, err := ValidateVirtualPath("COM1.txt"); err == nil {
		t.Fatal("Windows device path admitted")
	}
}

func TestReservedContainerSeparatorAndComposedPathLimit(t *testing.T) {
	if _, err := ValidateVirtualPath("directory!/file.go"); err == nil {
		t.Fatal("reserved container separator was accepted in a physical path")
	}

	root := strings.Repeat("a/", 2_042) + "source.zip"
	payload := buildZIP(t, []zipFixtureEntry{{name: "main.go", content: []byte("package main\n")}})
	result, err := admitDependency(t, root, payload, ProfileGoV1)
	requireCode(t, err, CodeArchiveUnsafePath)
	requireDecision(t, result, DecisionReject)
	if result.Manifest.Diagnostics[0].LimitName != "max_path_bytes" {
		t.Fatalf("composed path diagnostic = %+v", result.Manifest.Diagnostics[0])
	}
}
