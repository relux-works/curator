package registry

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/pathboundary"
)

func writeCheckpointFixture(t *testing.T, directory string, signingKey *signer, version int, conflicting bool, validSignature bool) string {
	t.Helper()
	body := snapshotBody(version, signerClock(t))
	body["head"] = repeatedHex("b")
	body["merkle_root"] = repeatedHex("a")
	body["log_size"] = version
	if conflicting {
		body["head"] = repeatedHex("c")
		body["merkle_root"] = repeatedHex("d")
	}
	signed := signingKey.sign(body)
	if !validSignature {
		signature := signed["sig"].(map[string]any)
		signature["signature"] = base64.StdEncoding.EncodeToString(make([]byte, 64))
	}
	payload, err := json.Marshal(signed)
	if err != nil {
		t.Fatal(err)
	}
	checkpointDir := filepath.Join(directory, "checkpoints")
	if err := os.Mkdir(checkpointDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := pathboundary.ProtectTree(checkpointDir); err != nil {
		t.Fatalf("protect checkpoint fixture directory: %v", err)
	}
	path := filepath.Join(checkpointDir, "checkpoint.json")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := pathboundary.ProtectTree(checkpointDir); err != nil {
		t.Fatalf("protect checkpoint fixture file: %v", err)
	}
	return path
}

func signerClock(t *testing.T) time.Time {
	t.Helper()
	return time.Now().UTC().Truncate(time.Second)
}

func repeatedHex(char string) string { return strings.Repeat(char, 64) }

func TestReconcileBootstrapCheckpointPreservesAndAdvancesHighWater(t *testing.T) {
	signer := newSigner(t)
	prior := snapshotState{
		HighestVersion:  8,
		Head:            repeatedHex("b"),
		MerkleRoot:      repeatedHex("a"),
		LogSize:         8,
		BootstrapSource: "first-use",
	}
	tests := []struct {
		name           string
		version        int
		conflicting    bool
		wantApplied    bool
		wantRegression bool
		wantVersion    int
		wantSource     string
	}{
		{name: "equal consistent no-op", version: 8, wantVersion: 8, wantSource: "first-use"},
		{name: "lower refused", version: 7, wantRegression: true, wantVersion: 8, wantSource: "first-use"},
		{name: "equal conflicting refused", version: 8, conflicting: true, wantRegression: true, wantVersion: 8, wantSource: "first-use"},
		{name: "higher advances", version: 10, wantApplied: true, wantVersion: 10, wantSource: "checkpoint"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			directory := t.TempDir()
			reg := Registry{Name: "primary", URL: "https://registry.example.test", PublicKeys: []string{signer.pinned},
				BootstrapCheckpoint: writeCheckpointFixture(t, directory, signer, testCase.version, testCase.conflicting, true)}
			next, applied, regression, err := reconcileBootstrapCheckpoint(reg, prior, true)
			if err != nil || applied != testCase.wantApplied || regression != testCase.wantRegression || next.HighestVersion != testCase.wantVersion || next.BootstrapSource != testCase.wantSource {
				t.Fatalf("reconcile = state %+v, applied=%v regression=%v error=%v", next, applied, regression, err)
			}
		})
	}
}

func TestReconcileBootstrapCheckpointRequiresValidSignatureOnlyOnFirstUse(t *testing.T) {
	signer := newSigner(t)
	directory := t.TempDir()
	path := writeCheckpointFixture(t, directory, signer, 8, false, false)
	reg := Registry{Name: "primary", URL: "https://registry.example.test", PublicKeys: []string{signer.pinned}, BootstrapCheckpoint: path}
	if _, applied, regression, err := reconcileBootstrapCheckpoint(reg, snapshotState{}, false); err == nil || applied || regression {
		t.Fatalf("invalid first-use checkpoint = applied %v regression %v error %v", applied, regression, err)
	}
	prior := snapshotState{HighestVersion: 8, Head: repeatedHex("b"), MerkleRoot: repeatedHex("a"), LogSize: 8, BootstrapSource: "first-use"}
	next, applied, regression, err := reconcileBootstrapCheckpoint(reg, prior, true)
	if err != nil || applied || regression || next != prior {
		t.Fatalf("invalid rebootstrap checkpoint changed state: state=%+v applied=%v regression=%v error=%v", next, applied, regression, err)
	}
}

func TestReadBootstrapCheckpointAcceptsCRLFJSON(t *testing.T) {
	signer := newSigner(t)
	path := writeCheckpointFixture(t, t.TempDir(), signer, 9, false, true)
	compact, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var formatted bytes.Buffer
	if err := json.Indent(&formatted, compact, "", "  "); err != nil {
		t.Fatal(err)
	}
	crlf := bytes.ReplaceAll(formatted.Bytes(), []byte("\n"), []byte("\r\n"))
	if err := os.WriteFile(path, crlf, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := pathboundary.ProtectTree(filepath.Dir(path)); err != nil {
		t.Fatalf("protect CRLF checkpoint fixture: %v", err)
	}

	got, err := readBootstrapCheckpoint(path, []string{signer.pinned})
	if err != nil {
		t.Fatalf("read valid CRLF checkpoint: %v", err)
	}
	if got.Version != 9 || got.LogSize != 9 || got.MerkleRoot != repeatedHex("a") {
		t.Fatalf("parsed checkpoint = %+v, want signed version 9 and its Merkle root", got)
	}
}

func TestReconcileBootstrapCheckpointFailsClosedForMalformedRebootstrap(t *testing.T) {
	signer := newSigner(t)
	path := writeCheckpointFixture(t, t.TempDir(), signer, 10, false, true)
	if err := os.WriteFile(path, []byte("{\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := pathboundary.ProtectTree(filepath.Dir(path)); err != nil {
		t.Fatalf("protect malformed checkpoint fixture: %v", err)
	}
	prior := snapshotState{
		HighestVersion:  8,
		Head:            repeatedHex("b"),
		MerkleRoot:      repeatedHex("a"),
		LogSize:         8,
		BootstrapSource: "first-use",
	}
	reg := Registry{Name: "primary", URL: "https://registry.example.test", PublicKeys: []string{signer.pinned}, BootstrapCheckpoint: path}

	next, applied, regression, err := reconcileBootstrapCheckpoint(reg, prior, true)
	if err == nil || applied || regression || next != prior {
		t.Fatalf("malformed rebootstrap = state %+v, applied=%v regression=%v error=%v; want a fail-closed error with unchanged state", next, applied, regression, err)
	}
}
