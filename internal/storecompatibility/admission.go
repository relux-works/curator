package storecompatibility

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Selector names exactly the historical families delegated to the compatibility
// job. All other conformance families remain in the canonical suite.
const Selector = "^TestStoreBoundary(Resolve|DryRun|Repair|Status)VectorsDrive"

// HistoricalTest identifies a top-level family in the frozen archive roster.
func HistoricalTest(name string) bool {
	for _, row := range roster {
		if strings.SplitN(row.ID, "/", 2)[0] == name {
			return true
		}
	}
	return false
}

// Admit replays the complete same-head artifact, rather than trusting aggregate
// counts or a producer's success flag. The Actions artifact channel supplies
// provenance; these files are not a cryptographic attestation.
func Admit(dir, commit string) (Report, error) {
	if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(commit) {
		return Report{}, fmt.Errorf("compatibility evidence requires an exact head")
	}
	// #nosec G304 -- The caller supplies the explicit CI artifact directory.
	head, err := os.ReadFile(filepath.Join(dir, "head.txt"))
	if err != nil || string(head) != commit+"\n" {
		return Report{}, fmt.Errorf("missing or different-head compatibility evidence")
	}
	// #nosec G304 -- The caller supplies the explicit CI artifact directory.
	f, err := os.Open(filepath.Join(dir, "replay.jsonl"))
	if err != nil {
		return Report{}, err
	}
	defer func() { _ = f.Close() }()
	return Check(f)
}
