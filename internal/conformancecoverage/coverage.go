// Package conformancecoverage classifies published case consumers at their
// production-entry tests and enforces the committed count pins and gap ledger.
package conformancecoverage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Gap is an owed implementation recorded in .github/ci/conformance-gaps.tsv.
type Gap struct {
	Family string
	CaseID string
	Owner  string
	Reason string
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

// Load reads the only gap ledger and the committed per-family count pins.
func Load() (map[string]int, []Gap, error) {
	root, err := repositoryRoot()
	if err != nil {
		return nil, nil, err
	}
	counts, err := readCounts(filepath.Join(root, ".github", "ci", "conformance-case-counts.tsv"))
	if err != nil {
		return nil, nil, err
	}
	gaps, err := readGaps(filepath.Join(root, ".github", "ci", "conformance-gaps.tsv"))
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

func repositoryRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := current; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, ".github", "ci", "conformance-case-counts.tsv")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("cannot locate repository root above %s", current)
		}
	}
}

func readCounts(path string) (map[string]int, error) {
	file, err := os.Open(path) // #nosec G304 -- path is fixed beneath the discovered repository root
	if err != nil {
		return nil, fmt.Errorf("read count pins %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()
	scanner := bufio.NewScanner(file)
	counts := map[string]int{}
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
			if len(fields) != 2 || fields[0] != "family" || fields[1] != "expected_cases" {
				return nil, fmt.Errorf("%s:%d: expected family<TAB>expected_cases header", path, lineNo)
			}
			header = true
			continue
		}
		if len(fields) != 2 || fields[0] == "" {
			return nil, fmt.Errorf("%s:%d: malformed count pin", path, lineNo)
		}
		count, err := strconv.Atoi(fields[1])
		if err != nil || count <= 0 {
			return nil, fmt.Errorf("%s:%d: expected positive case count, got %q", path, lineNo, fields[1])
		}
		if _, duplicate := counts[fields[0]]; duplicate {
			return nil, fmt.Errorf("%s:%d: duplicate family %q", path, lineNo, fields[0])
		}
		counts[fields[0]] = count
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

func readGaps(path string) ([]Gap, error) {
	file, err := os.Open(path) // #nosec G304 -- path is fixed beneath the discovered repository root
	if err != nil {
		return nil, fmt.Errorf("read gap ledger %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()
	scanner := bufio.NewScanner(file)
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
			want := []string{"family", "case_id", "owner", "reason"}
			if len(fields) != len(want) {
				return nil, fmt.Errorf("%s:%d: expected family<TAB>case_id<TAB>owner<TAB>reason header", path, lineNo)
			}
			for i := range want {
				if fields[i] != want[i] {
					return nil, fmt.Errorf("%s:%d: invalid gap ledger header", path, lineNo)
				}
			}
			header = true
			continue
		}
		if len(fields) != 4 {
			return nil, fmt.Errorf("%s:%d: gap row must have exactly four tab-separated fields", path, lineNo)
		}
		for _, field := range fields {
			if strings.TrimSpace(field) == "" || strings.ContainsAny(field, "\r\n") {
				return nil, fmt.Errorf("%s:%d: gap row fields must be non-empty and one line", path, lineNo)
			}
		}
		gap := Gap{Family: fields[0], CaseID: fields[1], Owner: fields[2], Reason: fields[3]}
		key := gap.Family + "\x00" + gap.CaseID
		if _, duplicate := seen[key]; duplicate {
			return nil, fmt.Errorf("%s:%d: duplicate gap row for %s/%s", path, lineNo, gap.Family, gap.CaseID)
		}
		seen[key] = struct{}{}
		gaps = append(gaps, gap)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read gap ledger %s: %w", path, err)
	}
	if !header {
		return nil, fmt.Errorf("%s: missing header", path)
	}
	return gaps, nil
}
