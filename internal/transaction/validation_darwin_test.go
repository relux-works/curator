//go:build darwin

package transaction

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildJournalRejectsCaseAliasesOnCaseInsensitiveDarwinFilesystem(t *testing.T) {
	root := t.TempDir()
	caseInsensitive, err := namespaceCaseInsensitive(root)
	if err != nil {
		t.Fatal(err)
	}
	if !caseInsensitive {
		t.Skip("test filesystem is case-sensitive")
	}
	sourceA := filepath.Join(root, "stage", "a")
	sourceB := filepath.Join(root, "stage", "b")
	mustWrite(t, sourceA, "new-a")
	mustWrite(t, sourceB, "new-b")
	plan := Plan{TransactionID: "txn-darwin-case-alias", ProjectIdentity: "/project", Targets: []Target{
		{Class: "a", Identifier: "upper", LivePath: filepath.Join(root, "live", "Target"), StagedSource: sourceA, PreimageDigest: DigestAbsent},
		{Class: "b", Identifier: "lower", LivePath: filepath.Join(root, "live", "target"), StagedSource: sourceB, PreimageDigest: DigestAbsent},
	}}
	assertPlanRejectedBeforeJournal(t, filepath.Join(root, "home"), plan)
}

func TestBuildJournalRejectsNormalizationAliasesOnNormalizationInsensitiveDarwinFilesystem(t *testing.T) {
	root := t.TempDir()
	liveRoot := filepath.Join(root, "live")
	mustMkdirAll(t, liveRoot)
	probeNFC := filepath.Join(liveRoot, "\u00e9-probe")
	probeNFD := filepath.Join(liveRoot, "e\u0301-probe")
	mustWrite(t, probeNFC, "probe")
	if _, err := os.Stat(probeNFD); err != nil {
		if os.IsNotExist(err) {
			t.Skip("test filesystem distinguishes canonical Unicode spellings")
		}
		t.Fatal(err)
	}
	if err := os.Remove(probeNFC); err != nil {
		t.Fatal(err)
	}

	sourceA := filepath.Join(root, "stage", "a")
	sourceB := filepath.Join(root, "stage", "b")
	mustWrite(t, sourceA, "new-a")
	mustWrite(t, sourceB, "new-b")
	plan := Plan{TransactionID: "txn-darwin-normalization-alias", ProjectIdentity: "/project", Targets: []Target{
		{Class: "a", Identifier: "nfc", LivePath: filepath.Join(liveRoot, "\u00e9"), StagedSource: sourceA, PreimageDigest: DigestAbsent},
		{Class: "b", Identifier: "nfd", LivePath: filepath.Join(liveRoot, "e\u0301"), StagedSource: sourceB, PreimageDigest: DigestAbsent},
	}}
	assertPlanRejectedBeforeJournal(t, filepath.Join(root, "home"), plan)
}
