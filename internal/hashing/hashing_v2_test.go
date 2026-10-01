package hashing

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManagedWriterVersionDefaultsToV1(t *testing.T) {
	if EnableV2Writers {
		t.Fatal("v2 writer switch is on by default while SPEC_PIN is rc.13")
	}
	if got := WriteVersion(); got != VersionV1 {
		t.Fatalf("default write version = %d, want v1", got)
	}
}

func TestManagedWriterVersionSwitchSelectsV2(t *testing.T) {
	prior := EnableV2Writers
	EnableV2Writers = true
	t.Cleanup(func() { EnableV2Writers = prior })
	if got := WriteVersion(); got != VersionV2 {
		t.Fatalf("enabled write version = %d, want v2", got)
	}
}

func writeBytes(t *testing.T, root, rel string, content []byte) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestContentSHA256V2ConformanceVectors(t *testing.T) {
	tests := []struct {
		name  string
		files map[string][]byte
		want  string
	}{
		{
			name:  "empty tree",
			files: map[string][]byte{},
			want:  "sha256:f34a9d36a24a41168189f7ac17fc4fd0fb0a7b7c15ce1065fa4a2409d48c0c12",
		},
		{
			name:  "embedded nul differs from file boundary",
			files: map[string][]byte{"a": []byte("one\x00b\x00two")},
			want:  "sha256:576c3d373a5b48e95cbd87c736f55a3f6bcfa9f5c789186bb6ea775fcd23551c",
		},
		{
			name:  "nested nul byte",
			files: map[string][]byte{"deep/third/bytes.bin": []byte("before\x00after")},
			want:  "sha256:7a02fafc05a0e2eb24f219d73a38194512675f95b6b072089a5ede5c0da135f1",
		},
		{
			name:  "ordinary tree",
			files: map[string][]byte{"README.md": []byte("curator\n"), "src/main.go": []byte("package main\n")},
			want:  "sha256:55150544f675ac2ff1a88407e6231a7f93c5a78a75e1b3f183463ae78d9a95cf",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			for path, contents := range tc.files {
				writeBytes(t, root, path, contents)
			}
			got, err := ContentSHA256WithVersion(root, nil, VersionV2)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("ContentSHA256WithVersion() = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestContentSHA256V2SeparatesTheV1Collision(t *testing.T) {
	one := t.TempDir()
	writeBytes(t, one, "a", []byte("one\x00b\x00two"))
	two := t.TempDir()
	writeBytes(t, two, "a", []byte("one"))
	writeBytes(t, two, "b", []byte("two"))

	legacyOne, err := ContentSHA256(one, nil)
	if err != nil {
		t.Fatal(err)
	}
	legacyTwo, err := ContentSHA256(two, nil)
	if err != nil {
		t.Fatal(err)
	}
	if legacyOne != legacyTwo || legacyOne != "sha256:c518e668673099905d1dd12ce983ff9447008f4e87829dfcfe1316a52114970c" {
		t.Fatalf("v1 collision changed: one=%s two=%s", legacyOne, legacyTwo)
	}

	identityOne, err := ContentIdentity(one, nil, VersionV2)
	if err != nil {
		t.Fatal(err)
	}
	identityTwo, err := ContentIdentity(two, nil, VersionV2)
	if err != nil {
		t.Fatal(err)
	}
	if identityOne.Equal(identityTwo) {
		t.Fatalf("v2 identities for colliding v1 pair compare equal: %+v %+v", identityOne, identityTwo)
	}
	if identityOne.SHA256 != "sha256:576c3d373a5b48e95cbd87c736f55a3f6bcfa9f5c789186bb6ea775fcd23551c" ||
		identityTwo.SHA256 != "sha256:de107e6a5f86b44a79e035761d78ceba8fc125cde23520d4268ac11b214a28d4" {
		t.Fatalf("v2 vector mismatch: one=%s two=%s", identityOne.SHA256, identityTwo.SHA256)
	}
}

func TestContentSHA256V2LengthFramingSeparatesAdjacentRecordBoundary(t *testing.T) {
	// Without either length field, the first stream is byte-for-byte equal to
	// the second: F + "a" + (F + "b" + "two") ==
	// (F + "a" + empty) + (F + "b" + "two").
	one := t.TempDir()
	writeBytes(t, one, "a", []byte("\x46btwo"))
	two := t.TempDir()
	writeBytes(t, two, "a", nil)
	writeBytes(t, two, "b", []byte("two"))

	oneHash, err := ContentSHA256WithVersion(one, nil, VersionV2)
	if err != nil {
		t.Fatal(err)
	}
	twoHash, err := ContentSHA256WithVersion(two, nil, VersionV2)
	if err != nil {
		t.Fatal(err)
	}
	if oneHash == twoHash {
		t.Fatalf("v2 identities collided when one file's bytes contain the next record framing: %s", oneHash)
	}
}

func TestIdentityEqualityRequiresHashVersion(t *testing.T) {
	digest := "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	if (Identity{HashVersion: VersionV1, SHA256: digest}).Equal(Identity{HashVersion: VersionV2, SHA256: digest}) {
		t.Fatal("equal digest text with different hash versions compared equal")
	}
}

func TestContentSHA256V2RejectsUnknownVersion(t *testing.T) {
	if _, err := ContentSHA256WithVersion(t.TempDir(), nil, 3); err == nil {
		t.Fatal("unsupported hash version was accepted")
	}
}
