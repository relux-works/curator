// Package conformancecoverage classifies published case consumers at their
// production-entry tests and enforces the committed count pins and gap ledger.
package conformancecoverage

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/stateread"
)

// Gap is an owed implementation recorded in .github/ci/conformance-gaps.tsv.
type Gap struct {
	Family string
	CaseID string
	Owner  string
	Reason string
}

// defaultSuiteManifestSHA256 is the immutable manifest identity used by the
// checked-in draft corpus when no external conformance root is selected.
const defaultSuiteManifestSHA256 = "be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca"

// ContentHashV2CandidateManifestSHA256 is the candidate suite accepted by
// the conformance consumers in this change.
const ContentHashV2CandidateManifestSHA256 = "950ee74ad148615c273fe95bbb93f1bc0f9bdf2ea2bd9395f1dc8e3601419e60"

// MuseCandidateManifestSHA256 identifies spec-muse-environment at d373078a.
const MuseCandidateManifestSHA256 = "bd03456b92a7368d90ea74fe6953db10bc188588020683024a6a6b8735a40783"

// RC14CandidateManifestSHA256 identifies the released rc.14 default CI root
// at 43bf0a2506d5c354a73bbc3ea4623d4653db10c7 (unchanged candidate manifest).
const RC14CandidateManifestSHA256 = "6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5"

// IsImplementedContentHashV2Suite identifies suites whose hash-v2 cases are
// driven through the implemented readers rather than historical gap accounting.
func IsImplementedContentHashV2Suite(digest string) bool {
	return digest == ContentHashV2CandidateManifestSHA256 || digest == RC14CandidateManifestSHA256
}

// IsMuseCandidate identifies suites publishing the Muse and fragment-v3 cases.
func IsMuseCandidate(digest string) bool {
	return digest == MuseCandidateManifestSHA256 || digest == RC14CandidateManifestSHA256
}

// IsContentHashV2Candidate identifies suites that publish the separately owned hash-v2 families.
func IsContentHashV2Candidate(digest string) bool {
	return IsImplementedContentHashV2Suite(digest) || digest == MuseCandidateManifestSHA256
}

// Observation is the result of one published case. A failure may be accepted
// only when the ledger names this family and case. Bounds and skips are
// explicit classifications, never passes.
type Observation struct {
	FailureReason string
	BoundReason   string
	SkipReason    string
}

// Tally reports one classification per published case.
type Tally struct {
	Driven   int
	KnownGap int
	Bound    int
	Skipped  int
}

// Total returns the number of classified cases.
func (t Tally) Total() int { return t.Driven + t.KnownGap + t.Bound + t.Skipped }

// Result associates an observation with the published case id.
type Result struct {
	CaseID        string
	FailureReason string
	BoundReason   string
	SkipReason    string
}

// Check requires the case list to match its count pin and one result per case.
// It rejects stale ledger rows, unlisted failures, and rows whose case now
// passes. The count check intentionally happens after classifying supplied
// results, so TestPublishedCaseTallyRejectsMissingClassification proves the
// tally itself is the missing-case gate.
func Check(family string, caseIDs []string, results []Result, gaps []Gap, expectedCount int) (Tally, error) {
	var tally Tally
	if family == "" {
		return tally, fmt.Errorf("published-case family is empty")
	}
	if expectedCount < 0 {
		return tally, fmt.Errorf("published-case count pin for %q is negative: %d", family, expectedCount)
	}

	published := make(map[string]struct{}, len(caseIDs))
	for _, id := range caseIDs {
		if id == "" {
			return tally, fmt.Errorf("family %q contains an empty case id", family)
		}
		if _, exists := published[id]; exists {
			return tally, fmt.Errorf("family %q publishes duplicate case id %q", family, id)
		}
		published[id] = struct{}{}
	}
	if len(caseIDs) != expectedCount {
		return tally, fmt.Errorf("family %q publishes %d cases, want pinned count %d", family, len(caseIDs), expectedCount)
	}

	rows := make(map[string]Gap)
	for _, gap := range gaps {
		if gap.Family != family {
			continue
		}
		key := gap.Family + "\x00" + gap.CaseID
		if _, exists := rows[key]; exists {
			return tally, fmt.Errorf("duplicate gap ledger row for %s/%s", family, gap.CaseID)
		}
		rows[key] = gap
		if _, exists := published[gap.CaseID]; !exists {
			return tally, fmt.Errorf("gap ledger case %s/%s vanished from the published case list", family, gap.CaseID)
		}
	}

	seen := make(map[string]struct{}, len(results))
	for _, result := range results {
		if _, exists := published[result.CaseID]; !exists {
			return tally, fmt.Errorf("family %q produced a result for unpublished case %q", family, result.CaseID)
		}
		if _, exists := seen[result.CaseID]; exists {
			return tally, fmt.Errorf("family %q classified case %q more than once", family, result.CaseID)
		}
		seen[result.CaseID] = struct{}{}
		key := family + "\x00" + result.CaseID
		gap, listed := rows[key]

		classes := 0
		if result.FailureReason != "" {
			classes++
		}
		if result.BoundReason != "" {
			classes++
		}
		if result.SkipReason != "" {
			classes++
		}
		if classes > 1 {
			return tally, fmt.Errorf("family %q case %q has conflicting outcome classifications", family, result.CaseID)
		}

		switch {
		case result.SkipReason != "":
			if listed {
				return tally, fmt.Errorf("known-gap case %s/%s was skipped, so its failure is unverified", family, result.CaseID)
			}
			tally.Skipped++
		case result.BoundReason != "":
			if listed {
				return tally, fmt.Errorf("known-gap case %s/%s was classified as bound", family, result.CaseID)
			}
			tally.Bound++
		case result.FailureReason != "":
			if !listed {
				return tally, fmt.Errorf("failing published case %s/%s is not listed in the gap ledger: %s", family, result.CaseID, result.FailureReason)
			}
			_ = gap // owner and reason are validated while loading the ledger.
			tally.KnownGap++
		default:
			if listed {
				return tally, fmt.Errorf("known-gap case %s/%s now passes; remove its ledger row", family, result.CaseID)
			}
			tally.Driven++
		}
	}

	if err := ValidateTally(tally, expectedCount); err != nil {
		return tally, fmt.Errorf("family %q: %w", family, err)
	}
	return tally, nil
}

// ValidateTally is the explicit ratchet against omitted classifications.
func ValidateTally(tally Tally, publishedCount int) error {
	if total := tally.Total(); total != publishedCount {
		return fmt.Errorf("driven(%d)+known-gap(%d)+bound(%d)+skipped(%d)=%d, want published count %d",
			tally.Driven, tally.KnownGap, tally.Bound, tally.Skipped, total, publishedCount)
	}
	return nil
}

// Run executes and classifies every case in one published family.
func Run[T any](t *testing.T, family string, cases []T, caseID func(T) string, drive func(*testing.T, T)) Tally {
	t.Helper()
	return RunOutcomes(t, family, cases, caseID, func(caseT *testing.T, testCase T) Observation {
		drive(caseT, testCase)
		return Observation{}
	})
}

// RunOutcomes is Run with explicit bound, skip, or expected-failure
// observations for consumers that need those classes.
func RunOutcomes[T any](t *testing.T, family string, cases []T, caseID func(T) string, drive func(*testing.T, T) Observation) Tally {
	return runOutcomes(t, family, cases, caseID, drive, false)
}

// RunOutcomesParallel is RunOutcomes with parallel case subtests. The
// aggregate check runs in parent cleanup, after every parallel child finishes.
func RunOutcomesParallel[T any](t *testing.T, family string, cases []T, caseID func(T) string, drive func(*testing.T, T) Observation) {
	t.Helper()
	_ = runOutcomes(t, family, cases, caseID, drive, true)
}

func runOutcomes[T any](t *testing.T, family string, cases []T, caseID func(T) string, drive func(*testing.T, T) Observation, parallel bool) Tally {
	t.Helper()
	counts, gaps, err := Load()
	if err != nil {
		t.Fatalf("load published-case coverage policy: %v", err)
	}
	expectedCount, ok := counts[family]
	if !ok {
		t.Fatalf("published-case family %q has no committed count pin", family)
	}

	ids := make([]string, len(cases))
	results := make([]Result, len(cases))
	check := func() Tally {
		tally, err := Check(family, ids, results, gaps, expectedCount)
		if err != nil {
			t.Errorf("published-case coverage: %v", err)
		}
		t.Logf("published cases %s: %d driven, %d known-gap, %d bound, %d skipped, %d total",
			family, tally.Driven, tally.KnownGap, tally.Bound, tally.Skipped, tally.Total())
		return tally
	}
	if parallel {
		t.Cleanup(func() { check() })
	}
	for i, testCase := range cases {
		i, testCase := i, testCase
		ids[i] = caseID(testCase)
		var observation Observation
		t.Run(ids[i], func(caseT *testing.T) {
			if parallel {
				caseT.Parallel()
			}
			defer func() {
				if caseT.Skipped() {
					reason := observation.SkipReason
					if reason == "" {
						reason = "case test skipped"
					}
					observation = Observation{SkipReason: reason}
				} else if caseT.Failed() && observation.FailureReason == "" {
					observation.FailureReason = "case test failed"
				}
				results[i] = Result{
					CaseID:        ids[i],
					FailureReason: observation.FailureReason,
					BoundReason:   observation.BoundReason,
					SkipReason:    observation.SkipReason,
				}
			}()
			observation = drive(caseT, testCase)
			if observation.SkipReason != "" {
				caseT.Skip(observation.SkipReason)
			}
		})
	}
	if parallel {
		return Tally{}
	}
	return check()
}

// Load reads the count pins and gap rows for the selected conformance suite.
// An explicit root is identified by the SHA-256 of its manifest.json; without
// one, the checked-in draft corpus uses the immutable rc.13 identity.
func Load() (map[string]int, []Gap, error) {
	root, err := repositoryRoot()
	if err != nil {
		return nil, nil, err
	}
	suiteID, err := selectedSuiteIdentity()
	if err != nil {
		return nil, nil, err
	}
	allCounts, err := readCounts(filepath.Join(root, ".github", "ci", "conformance-case-counts.tsv"))
	if err != nil {
		return nil, nil, err
	}
	counts, ok := allCounts[suiteID]
	if !ok {
		return nil, nil, fmt.Errorf("no published-case count pins for conformance manifest sha256:%s", suiteID)
	}
	gaps, err := readGaps(filepath.Join(root, ".github", "ci", "conformance-gaps.tsv"), suiteID)
	if err != nil {
		return nil, nil, err
	}
	for _, gap := range gaps {
		if _, ok := counts[gap.Family]; !ok {
			return nil, nil, fmt.Errorf("gap ledger family %q has no count pin", gap.Family)
		}
	}
	return counts, gaps, nil
}

func selectedSuiteIdentity() (string, error) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		return defaultSuiteManifestSHA256, nil
	}
	manifestPath := filepath.Join(root, "manifest.json")
	manifest, err := readStateBytes(manifestPath)
	if err != nil {
		return "", fmt.Errorf("read conformance manifest identity %s: %w", manifestPath, err)
	}
	sum := sha256.Sum256(manifest)
	return hex.EncodeToString(sum[:]), nil
}

// RequirePublishedCount checks an independently loaded family count against
// the exact pin selected by the conformance manifest identity.
func RequirePublishedCount(t testing.TB, family string, actual int) {
	t.Helper()
	counts, _, err := Load()
	if err != nil {
		t.Fatalf("load published-case coverage policy: %v", err)
	}
	expected, ok := counts[family]
	if !ok {
		t.Fatalf("published-case family %q has no committed count pin", family)
	}
	if actual != expected {
		t.Fatalf("published-case family %q has %d cases, want pinned count %d", family, actual, expected)
	}
}

// SelectedSuiteManifestSHA256 returns the manifest digest that Load uses for
// count and gap selection. It is useful to gate cases that exist only in a
// named candidate suite.
func SelectedSuiteManifestSHA256() (string, error) { return selectedSuiteIdentity() }

func repositoryRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return repositoryRootFrom(current)
}

func repositoryRootFrom(current string) (string, error) {
	for dir := current; ; dir = filepath.Dir(dir) {
		marker := filepath.Join(dir, ".github", "ci", "conformance-case-counts.tsv")
		metadata, err := stateread.Stat(marker)
		if err != nil {
			return "", fmt.Errorf("inspect repository root marker %s: %w", marker, err)
		}
		switch metadata.Kind {
		case stateread.KindPresent:
			return dir, nil
		case stateread.KindAbsent:
			// Keep searching only after the seam proves this marker absent.
		case stateread.KindUnreadable:
			return "", stateread.UnusableError(marker, fmt.Errorf("stat returned unreadable state without an error"))
		default:
			return "", stateread.UnusableError(marker, fmt.Errorf("unknown metadata state %q", metadata.Kind))
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("cannot locate repository root above %s", current)
		}
	}
}

func readCounts(path string) (map[string]map[string]int, error) {
	data, err := readStateBytes(path)
	if err != nil {
		return nil, fmt.Errorf("read count pins %s: %w", path, err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	counts := map[string]map[string]int{}
	header := false
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if !header {
			if len(fields) != 3 || fields[0] != "suite_manifest_sha256" || fields[1] != "family" || fields[2] != "expected_cases" {
				return nil, fmt.Errorf("%s:%d: expected suite_manifest_sha256<TAB>family<TAB>expected_cases header", path, lineNo)
			}
			header = true
			continue
		}
		if len(fields) != 3 || fields[0] == "" || fields[1] == "" {
			return nil, fmt.Errorf("%s:%d: malformed count pin", path, lineNo)
		}
		if len(fields[0]) != 64 {
			return nil, fmt.Errorf("%s:%d: suite manifest identity must be a 64-character sha256", path, lineNo)
		}
		if _, err := hex.DecodeString(fields[0]); err != nil {
			return nil, fmt.Errorf("%s:%d: malformed suite manifest sha256 %q", path, lineNo, fields[0])
		}
		count, err := strconv.Atoi(fields[2])
		if err != nil || count <= 0 {
			return nil, fmt.Errorf("%s:%d: expected positive case count, got %q", path, lineNo, fields[2])
		}
		if counts[fields[0]] == nil {
			counts[fields[0]] = map[string]int{}
		}
		if _, duplicate := counts[fields[0]][fields[1]]; duplicate {
			return nil, fmt.Errorf("%s:%d: duplicate count pin for suite %s family %q", path, lineNo, fields[0], fields[1])
		}
		counts[fields[0]][fields[1]] = count
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read count pins %s: %w", path, err)
	}
	if !header {
		return nil, fmt.Errorf("%s: missing header", path)
	}
	if len(counts) == 0 {
		return nil, fmt.Errorf("%s: no case counts", path)
	}
	return counts, nil
}

func readGaps(path, suiteID string) ([]Gap, error) {
	data, err := readStateBytes(path)
	if err != nil {
		return nil, fmt.Errorf("read gap ledger %s: %w", path, err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var gaps []Gap
	seen := map[string]struct{}{}
	header := false
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if !header {
			want := []string{"suite_manifest_sha256", "family", "case_id", "owner", "reason"}
			if len(fields) != len(want) {
				return nil, fmt.Errorf("%s:%d: expected suite_manifest_sha256<TAB>family<TAB>case_id<TAB>owner<TAB>reason header", path, lineNo)
			}
			for i := range want {
				if fields[i] != want[i] {
					return nil, fmt.Errorf("%s:%d: invalid gap ledger header", path, lineNo)
				}
			}
			header = true
			continue
		}
		if len(fields) != 5 {
			return nil, fmt.Errorf("%s:%d: gap row must have exactly five tab-separated fields", path, lineNo)
		}
		for _, field := range fields {
			if strings.TrimSpace(field) == "" || strings.ContainsAny(field, "\r\n") {
				return nil, fmt.Errorf("%s:%d: gap row fields must be non-empty and one line", path, lineNo)
			}
		}
		if len(fields[0]) != 64 {
			return nil, fmt.Errorf("%s:%d: suite manifest identity must be a 64-character sha256", path, lineNo)
		}
		if _, err := hex.DecodeString(fields[0]); err != nil {
			return nil, fmt.Errorf("%s:%d: malformed suite manifest sha256 %q", path, lineNo, fields[0])
		}
		gap := Gap{Family: fields[1], CaseID: fields[2], Owner: fields[3], Reason: fields[4]}
		key := fields[0] + "\x00" + gap.Family + "\x00" + gap.CaseID
		if _, duplicate := seen[key]; duplicate {
			return nil, fmt.Errorf("%s:%d: duplicate gap row for %s/%s", path, lineNo, gap.Family, gap.CaseID)
		}
		seen[key] = struct{}{}
		if fields[0] == suiteID {
			gaps = append(gaps, gap)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read gap ledger %s: %w", path, err)
	}
	if !header {
		return nil, fmt.Errorf("%s: missing header", path)
	}
	return gaps, nil
}

func readStateBytes(path string) ([]byte, error) {
	file, err := stateread.ReadFile(path)
	if err != nil {
		return nil, err
	}
	switch file.Kind {
	case stateread.KindPresent:
		return file.Bytes, nil
	case stateread.KindAbsent:
		return nil, stateread.AbsentError(path)
	case stateread.KindUnreadable:
		return nil, stateread.UnusableError(path, fmt.Errorf("read returned unreadable state without an error"))
	default:
		return nil, stateread.UnusableError(path, fmt.Errorf("unknown file state %q", file.Kind))
	}
}
