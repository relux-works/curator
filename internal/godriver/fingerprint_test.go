package godriver

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToolchainFramingMatchesRC4Vector(t *testing.T) {
	records := []struct {
		kind    byte
		path    string
		payload []byte
	}{
		{kind: 'D', path: "bin"},
		{kind: 'F', path: "bin/go", payload: []byte("GO")},
		{kind: 'D', path: "pkg"},
		{kind: 'L', path: "pkg/tool-link", payload: []byte("../bin/go")},
	}
	var preimage bytes.Buffer
	preimage.WriteString(toolchainDomain)
	for _, record := range records {
		preimage.WriteByte(record.kind)
		writeVectorLength(&preimage, len(record.path))
		preimage.WriteString(record.path)
		writeVectorLength(&preimage, len(record.payload))
		preimage.Write(record.payload)
	}
	preimage.WriteByte('V')
	writeVectorLength(&preimage, 0)
	version := "go version go1.25.5 darwin/arm64"
	writeVectorLength(&preimage, len(version))
	preimage.WriteString(version)

	wantPreimageHex := "63757261746f722d676f2d746f6f6c636861696e2d76310044000000000000000362696e000000000000000046000000000000000662696e2f676f0000000000000002474f440000000000000003706b6700000000000000004c000000000000000d706b672f746f6f6c2d6c696e6b00000000000000092e2e2f62696e2f676f5600000000000000000000000000000020676f2076657273696f6e20676f312e32352e352064617277696e2f61726d3634"
	if got := hex.EncodeToString(preimage.Bytes()); got != wantPreimageHex {
		t.Fatalf("preimage = %s", got)
	}
	digest := sha256.Sum256(preimage.Bytes())
	if got := "sha256:" + hex.EncodeToString(digest[:]); got != "sha256:baf7c5f3b9c3f1fae3da4c356381bf74442aa7f8f0b6fb2304c9c10833d6032e" {
		t.Fatalf("digest = %s", got)
	}
}

func TestGoVersionLFAndCRLFNormalizeToSameIdentity(t *testing.T) {
	lf, _, _, _, err := parseGoVersion([]byte("go version go1.25.5 darwin/arm64\n"))
	if err != nil {
		t.Fatal(err)
	}
	crlf, _, _, _, err := parseGoVersion([]byte("go version go1.25.5 darwin/arm64\r\n"))
	if err != nil {
		t.Fatal(err)
	}
	if lf != crlf {
		t.Fatalf("LF = %q, CRLF = %q", lf, crlf)
	}

	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "bin", "go"), []byte("GO"), 0o755)
	lfDigest, _, err := fingerprintToolchain(context.Background(), root, lf)
	if err != nil {
		t.Fatal(err)
	}
	crlfDigest, _, err := fingerprintToolchain(context.Background(), root, crlf)
	if err != nil {
		t.Fatal(err)
	}
	if lfDigest != crlfDigest {
		t.Fatalf("LF digest %s != CRLF digest %s", lfDigest, crlfDigest)
	}
}

func TestCandidateRC4ToolchainArtifacts(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	expected := filepath.Join(root, "expected", "build-driver")
	preimage, err := os.ReadFile(filepath.Join(expected, "toolchain.preimage.bin"))
	if err != nil {
		t.Fatal(err)
	}
	digestFile, err := os.ReadFile(filepath.Join(expected, "toolchain-sha256.txt"))
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(preimage)
	got := "sha256:" + hex.EncodeToString(digest[:])
	if got != strings.TrimSpace(string(digestFile)) || got != "sha256:baf7c5f3b9c3f1fae3da4c356381bf74442aa7f8f0b6fb2304c9c10833d6032e" {
		t.Fatalf("candidate toolchain digest = %s, artifact = %q", got, digestFile)
	}
}

func TestFingerprintRejectsDuplicateEncodedClaim(t *testing.T) {
	seen := map[string]struct{}{"bin/go": {}}
	if err := claimEncodedPath(seen, "bin/go"); DiagnosticCode(err) != "duplicate_path" {
		t.Fatalf("duplicate error = %v", err)
	}
}

func TestFingerprintRejectsEscapingAndAbsoluteLinks(t *testing.T) {
	for name, target := range map[string]string{"escape": "../../outside", "absolute": filepath.Join(string(filepath.Separator), "outside")} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.MkdirAll(filepath.Join(root, "pkg"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, filepath.Join(root, "pkg", "link")); err != nil {
				t.Skipf("symlink unavailable: %v", err)
			}
			_, _, err := fingerprintToolchain(context.Background(), root, "go version go1.25.5 darwin/arm64")
			want := "toolchain_link_escape"
			if name == "absolute" {
				want = "toolchain_link_absolute"
			}
			if DiagnosticCode(err) != want {
				t.Fatalf("error = %v, want %s", err, want)
			}
		})
	}
}

func TestFingerprintRejectsDanglingLink(t *testing.T) {
	root := t.TempDir()
	if err := os.Symlink("missing", filepath.Join(root, "link")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	_, _, err := fingerprintToolchain(context.Background(), root, "go version go1.25.5 darwin/arm64")
	if DiagnosticCode(err) != "toolchain_link_dangling" {
		t.Fatalf("error = %v", err)
	}
}

func writeVectorLength(buffer *bytes.Buffer, length int) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(length))
	buffer.Write(encoded[:])
}
