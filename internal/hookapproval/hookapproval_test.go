package hookapproval

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/stateread"
)

var (
	testDigest = Digest([]byte("manager env bytes"))
	testStamp  = time.Date(2026, 9, 16, 0, 0, 0, 0, time.UTC)
)

func TestDigestMatchesKnownSHA256(t *testing.T) {
	if got := Digest([]byte("abc")); got != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatalf("Digest(abc) = %q", got)
	}
}

func TestLookupOnAbsentStateIsEmpty(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")
	if _, found, err := Lookup(home, filepath.Join(home, ".agents", "env.sh")); err != nil || found {
		t.Fatalf("Lookup on absent state = %v, %v", found, err)
	}
	if records, err := List(home); err != nil || len(records) != 0 {
		t.Fatalf("List on absent state = %v, %v", records, err)
	}
}

func TestListDistinguishesAbsentAndUnreadableState(t *testing.T) {
	home := t.TempDir()
	if records, err := List(home); err != nil || len(records) != 0 {
		t.Fatalf("List on absent state = %v, %v; want an empty set", records, err)
	}
	if runtime.GOOS == "windows" {
		t.Skip("platform-control: POSIX mode-bit unreadability")
	}
	path := ApprovalsPath(home)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0); err != nil {
		t.Skipf("this environment cannot set unreadable mode: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })
	if _, err := os.ReadFile(path); err == nil {
		t.Skip("this environment can read a mode-000 file; unreadability is untestable here")
	}
	if records, err := List(home); err == nil || records != nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) || !strings.Contains(err.Error(), path) {
		t.Fatalf("List on unreadable state = (%v, %v); want typed unreadable error for %s", records, err, path)
	}
}

func TestUpsertLookupListRoundTrip(t *testing.T) {
	home := t.TempDir()
	var err error
	home, err = filepath.EvalSymlinks(home)
	if err != nil {
		t.Fatal(err)
	}
	first := filepath.Join(home, "project", ".agents", "env.sh")
	second := filepath.Join(home, "project", ".agents", "env.ps1")
	if err := Upsert(home, Record{Path: second, SHA256: testDigest, ApprovedBy: ApprovedByOperator, ApprovedAt: testStamp}); err != nil {
		t.Fatal(err)
	}
	if err := Upsert(home, Record{Path: first, SHA256: testDigest, ApprovedBy: ApprovedByManager, ApprovedAt: testStamp}); err != nil {
		t.Fatal(err)
	}
	records, err := List(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 || records[0].Path != second || records[1].Path != first {
		t.Fatalf("List is not sorted by path: %+v", records)
	}
	if records[1].ApprovedBy != ApprovedByManager || !records[1].ApprovedAt.Equal(testStamp) {
		t.Fatalf("record lost fields: %+v", records[1])
	}
	found, ok, err := Lookup(home, first)
	if err != nil || !ok || found.SHA256 != testDigest {
		t.Fatalf("Lookup = %+v, %v, %v", found, ok, err)
	}
	payload, err := os.ReadFile(ApprovalsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	want := second + "\t" + testDigest + "\toperator\t2026-09-16T00:00:00Z\n" +
		first + "\t" + testDigest + "\tmanager\t2026-09-16T00:00:00Z\n"
	if string(payload) != want {
		t.Fatalf("state file:\n%s\nwant:\n%s", payload, want)
	}
}

func TestTwoSpellingsResolveToOneRecord(t *testing.T) {
	home := t.TempDir()
	resolvedHome, err := filepath.EvalSymlinks(home)
	if err != nil {
		t.Fatal(err)
	}
	canonical := filepath.Join(resolvedHome, "project", ".agents", "env.sh")
	detour := home + "/project/nested/../.agents/env.sh"
	if detour == canonical {
		t.Fatal("fixture paths do not differ")
	}
	if err := Upsert(home, Record{Path: detour, SHA256: testDigest, ApprovedBy: ApprovedByManager, ApprovedAt: testStamp}); err != nil {
		t.Fatal(err)
	}
	found, ok, err := Lookup(home, canonical)
	if err != nil || !ok {
		t.Fatalf("Lookup(canonical) = %v, %v", ok, err)
	}
	if found.Path != canonical {
		t.Fatalf("stored path = %q, want %q", found.Path, canonical)
	}
	records, err := List(home)
	if err != nil || len(records) != 1 {
		t.Fatalf("List = %v, %v", records, err)
	}
	removed, err := Revoke(home, detour)
	if err != nil || !removed {
		t.Fatalf("Revoke(detour) = %v, %v", removed, err)
	}
	if _, ok, err := Lookup(home, canonical); err != nil || ok {
		t.Fatalf("Lookup after revoke = %v, %v", ok, err)
	}
}

func TestAliasAndRealPathShareOneRecord(t *testing.T) {
	base := t.TempDir()
	realDir := filepath.Join(base, "real-project", ".agents")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	realPath := filepath.Join(realDir, "env.sh")
	if err := os.WriteFile(realPath, []byte("env bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(base, "alias-project")
	if err := os.Symlink(filepath.Join(base, "real-project"), alias); err != nil {
		t.Skipf("this host cannot create the symlinked project directory the case needs: %v", err)
	}
	aliasPath := filepath.Join(alias, ".agents", "env.sh")
	home := filepath.Join(base, "home")
	// Record through the alias spelling; the real spelling must observe the
	// same single record, and revoking through either spelling removes it.
	record, err := ApproveFile(home, aliasPath, ApprovedByOperator, testStamp)
	if err != nil {
		t.Fatal(err)
	}
	resolvedReal, err := filepath.EvalSymlinks(realPath)
	if err != nil {
		t.Fatal(err)
	}
	if record.Path != resolvedReal {
		t.Fatalf("stored path = %q, want %q", record.Path, resolvedReal)
	}
	found, ok, err := Lookup(home, realPath)
	if err != nil || !ok || found.SHA256 != record.SHA256 {
		t.Fatalf("Lookup(real) = %+v, %v, %v", found, ok, err)
	}
	records, err := List(home)
	if err != nil || len(records) != 1 {
		t.Fatalf("List = %v, %v", records, err)
	}
	removed, err := Revoke(home, realPath)
	if err != nil || !removed {
		t.Fatalf("Revoke(real) = %v, %v", removed, err)
	}
	if _, ok, err := Lookup(home, aliasPath); err != nil || ok {
		t.Fatalf("Lookup(alias) after revoke = %v, %v", ok, err)
	}
}

func TestRevokeAbsentLeavesStateUnchanged(t *testing.T) {
	home := t.TempDir()
	if removed, err := Revoke(home, filepath.Join(home, "missing")); err != nil || removed {
		t.Fatalf("Revoke on absent state = %v, %v", removed, err)
	}
	path := filepath.Join(home, "project", ".agents", "env.sh")
	if err := Upsert(home, Record{Path: path, SHA256: testDigest, ApprovedBy: ApprovedByManager, ApprovedAt: testStamp}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(ApprovalsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	if removed, err := Revoke(home, filepath.Join(home, "other")); err != nil || removed {
		t.Fatalf("Revoke(other) = %v, %v", removed, err)
	}
	after, err := os.ReadFile(ApprovalsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatalf("revoke of an absent path rewrote state:\n%s\n%s", before, after)
	}
}

func TestApproveFileRefusesAbsentAndUnreadable(t *testing.T) {
	home := t.TempDir()
	missing := filepath.Join(home, "project", ".agents", "env.sh")
	if _, err := ApproveFile(home, missing, ApprovedByOperator, testStamp); err == nil {
		t.Fatal("ApproveFile of an absent file succeeded")
	}
	if _, ok, err := Lookup(home, missing); err != nil || ok {
		t.Fatalf("absent file left a record: %v, %v", ok, err)
	}
	file := filepath.Join(home, "env.sh")
	if err := os.WriteFile(file, []byte("bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	record, err := ApproveFile(home, file, ApprovedByOperator, testStamp)
	if err != nil {
		t.Fatal(err)
	}
	if record.SHA256 != Digest([]byte("bytes")) || record.ApprovedBy != ApprovedByOperator {
		t.Fatalf("ApproveFile record = %+v", record)
	}
	if _, err := ApproveFile(home, file, "self", testStamp); err == nil {
		t.Fatal("ApproveFile with a forged approved_by succeeded")
	}
}

func TestUpsertRejectsOpenShapes(t *testing.T) {
	home := t.TempDir()
	base := Record{Path: filepath.Join(home, "env.sh"), SHA256: testDigest, ApprovedBy: ApprovedByManager, ApprovedAt: testStamp}
	for _, mutate := range []struct {
		name   string
		change func(*Record)
	}{
		{"short digest", func(r *Record) { r.SHA256 = "abc" }},
		{"uppercase digest", func(r *Record) { r.SHA256 = strings.ToUpper(testDigest) }},
		{"non-hex digest", func(r *Record) { r.SHA256 = strings.Repeat("z", 64) }},
		{"open approver", func(r *Record) { r.ApprovedBy = "self" }},
		{"missing timestamp", func(r *Record) { r.ApprovedAt = time.Time{} }},
		{"tab in path", func(r *Record) { r.Path = filepath.Join(home, "a\tb") }},
		{"newline in path", func(r *Record) { r.Path = filepath.Join(home, "a\nb") }},
	} {
		t.Run(mutate.name, func(t *testing.T) {
			record := base
			mutate.change(&record)
			if err := Upsert(home, record); err == nil {
				t.Fatalf("Upsert(%+v) succeeded", record)
			}
		})
	}
	if records, err := List(home); err != nil || len(records) != 0 {
		t.Fatalf("rejected upserts left state: %v, %v", records, err)
	}
}

func TestFailedPublicationPreservesLastValidState(t *testing.T) {
	home := t.TempDir()
	first := filepath.Join(home, "project", ".agents", "env.sh")
	second := filepath.Join(home, "project", ".agents", "env.ps1")
	if err := Upsert(home, Record{Path: first, SHA256: testDigest, ApprovedBy: ApprovedByManager, ApprovedAt: testStamp}); err != nil {
		t.Fatal(err)
	}
	if err := Upsert(home, Record{Path: second, SHA256: testDigest, ApprovedBy: ApprovedByOperator, ApprovedAt: testStamp}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(ApprovalsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	renameFile = func(_, _ string) error { return errors.New("injected publication failure") }
	defer func() { renameFile = os.Rename }()
	if err := Upsert(home, Record{Path: filepath.Join(home, "third"), SHA256: testDigest, ApprovedBy: ApprovedByManager, ApprovedAt: testStamp}); err == nil {
		t.Fatal("Upsert over a failed publication succeeded")
	}
	if removed, err := Revoke(home, first); err == nil || removed {
		t.Fatalf("Revoke over a failed publication = %v, %v", removed, err)
	}
	after, err := os.ReadFile(ApprovalsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("failed publication rewrote state:\n%s\nwas:\n%s", after, before)
	}
	records, err := List(home)
	if err != nil || len(records) != 2 {
		t.Fatalf("List after failed publication = %v, %v", records, err)
	}
	staged, err := filepath.Glob(filepath.Join(home, ".hook-approvals.*.tmp"))
	if err != nil {
		t.Fatal(err)
	}
	if len(staged) != 0 {
		t.Fatalf("failed publication left staged files: %v", staged)
	}
}

// TestTimestampShapeMirrorsRFC3339 ties the hook-embedded TimestampShape to
// the Go reader's time.Parse(time.RFC3339): every timestamp Go accepts MUST
// match the shape (a hook must never warn on a record Go wrote), and
// structurally broken input MUST fail both. Well-formed but out-of-range
// input (month 13, February 30, second 60) matches the shape by design; the
// per-consumer range checks (awk, PowerShell TryParse) close that gap and are
// covered end to end by the hook malformed-record tests.
func TestTimestampShapeMirrorsRFC3339(t *testing.T) {
	shape := regexp.MustCompile(TimestampShape)
	goValid := []string{
		"2026-09-16T00:00:00Z",
		"2026-09-16T00:00:00.1Z",
		"2026-09-16T00:00:00.123456789+02:00",
		"2024-02-29T23:59:59-05:30",
		"2026-09-16T00:00:00+00:00",
	}
	for _, stamp := range goValid {
		if _, err := time.Parse(time.RFC3339, stamp); err != nil {
			t.Fatalf("corpus input %q is not Go-valid: %v", stamp, err)
		}
		if !shape.MatchString(stamp) {
			t.Fatalf("shape rejects Go-valid %q", stamp)
		}
	}
	broken := []string{
		"",
		"not-a-time",
		"2026-09-16 00:00:00",
		"2026-09-16T00:00:00",
		"2026-09-16T00:00:00+0000",
		"2026-09-16t00:00:00z",
		"2026-09-16T00:00:00.",
		"2026-09-16T00:00:00Z ",
	}
	for _, stamp := range broken {
		if _, err := time.Parse(time.RFC3339, stamp); err == nil {
			t.Fatalf("corpus input %q is Go-valid", stamp)
		}
		if shape.MatchString(stamp) {
			t.Fatalf("shape accepts broken %q", stamp)
		}
	}
	// Well-formed but out of range: shape matches, Go rejects, hook
	// range checks must reject (covered end to end by the hook tests).
	rangeBroken := []string{
		"2026-13-01T00:00:00Z",
		"2026-02-30T00:00:00Z",
		"2023-02-29T00:00:00Z",
		"2026-09-16T24:00:00Z",
		"2026-09-16T00:00:60Z",
	}
	for _, stamp := range rangeBroken {
		if _, err := time.Parse(time.RFC3339, stamp); err == nil {
			t.Fatalf("corpus input %q is Go-valid", stamp)
		}
		if !shape.MatchString(stamp) {
			t.Fatalf("shape rejects well-formed %q", stamp)
		}
	}
}

// TestWindowsDriveIdentity proves the Windows half of the Canonicalize
// identity rule: the stored spelling carries an uppercase drive letter and
// a differently-cased spelling of the same file resolves to the one
// record. It runs only on Windows; Unix filesystems keep exact identity.
func TestWindowsDriveIdentity(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows drive identity is exercised on Windows")
	}
	home := t.TempDir()
	// Mixed-case trailing components do not exist yet, so canonicalization
	// keeps their case lexically: the stored and lookup spellings differ
	// by case alone, and only the fold can match them.
	mixed := filepath.Join(home, "Project", ".agents", "Env.sh")
	lowerDrive := mixed
	if c := lowerDrive[0]; c >= 'A' && c <= 'Z' {
		lowerDrive = string(c-'A'+'a') + lowerDrive[1:]
	}
	canonical, err := Canonicalize(lowerDrive)
	if err != nil {
		t.Fatal(err)
	}
	if len(canonical) < 3 || canonical[1] != ':' || canonical[0] < 'A' || canonical[0] > 'Z' {
		t.Fatalf("canonical Windows path keeps a lowercase drive: %q", canonical)
	}
	if err := Upsert(home, Record{Path: lowerDrive, SHA256: testDigest, ApprovedBy: ApprovedByManager, ApprovedAt: testStamp}); err != nil {
		t.Fatal(err)
	}
	folded := strings.ToLower(canonical)
	if folded == canonical {
		t.Fatal("fixture path has no uppercase letter to fold")
	}
	found, ok, err := Lookup(home, folded)
	if err != nil || !ok || found.SHA256 != testDigest {
		t.Fatalf("Lookup(folded) = %+v, %v, %v", found, ok, err)
	}
	records, err := List(home)
	if err != nil || len(records) != 1 {
		t.Fatalf("List = %v, %v", records, err)
	}
	if removed, err := Revoke(home, folded); err != nil || !removed {
		t.Fatalf("Revoke(folded) = %v, %v", removed, err)
	}
}

func TestMalformedStateFailsClosed(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, "env.sh")
	good := path + "\t" + testDigest + "\tmanager\t2026-09-16T00:00:00Z\n"
	for _, body := range []string{
		"only-two-fields\t" + testDigest + "\n",
		path + "\tshort\tmanager\t2026-09-16T00:00:00Z\n",
		path + "\t" + testDigest + "\tself\t2026-09-16T00:00:00Z\n",
		path + "\t" + testDigest + "\tmanager\tnot-a-time\n",
		good + "trailing-garbage\n",
	} {
		if err := os.WriteFile(ApprovalsPath(home), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := List(home); err == nil {
			t.Fatalf("List(%q) succeeded", body)
		}
		if _, _, err := Lookup(home, path); err == nil {
			t.Fatalf("Lookup over %q succeeded", body)
		}
	}
}

func TestScanSkipsMalformedWithoutMutating(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(base, "home")
	good := filepath.Join(base, "project", ".agents", "env.sh")
	other := filepath.Join(base, "project", ".agents", "env.ps1")
	body := good + "\t" + testDigest + "\tmanager\t2026-09-16T00:00:00Z\n" +
		"only-two-fields\t" + testDigest + "\n" +
		other + "\t" + testDigest + "\toperator\t2026-09-16T00:00:00Z\n" +
		good + "\tshort\tmanager\t2026-09-16T00:00:00Z\n" +
		good + "\t" + testDigest + "\tself\t2026-09-16T00:00:00Z\n" +
		good + "\t" + testDigest + "\tmanager\tnot-a-time\n"
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ApprovalsPath(home), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	records, malformed, err := Scan(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 || records[0].Path != other || records[1].Path != good {
		t.Fatalf("Scan records = %+v, want the two valid lines sorted", records)
	}
	if records[0].ApprovedBy != ApprovedByOperator || records[1].ApprovedBy != ApprovedByManager {
		t.Fatalf("Scan lost approved_by: %+v", records)
	}
	wantMalformed := []int{2, 4, 5, 6}
	if len(malformed) != len(wantMalformed) {
		t.Fatalf("Scan malformed = %v, want %v", malformed, wantMalformed)
	}
	for i, line := range wantMalformed {
		if malformed[i] != line {
			t.Fatalf("Scan malformed = %v, want %v", malformed, wantMalformed)
		}
	}
	// The strict reader still fails closed over the same bytes: Scan's
	// tolerance is reporting-only and never authorizes.
	if _, err := List(home); err == nil {
		t.Fatal("List over malformed state succeeded")
	}
	after, err := os.ReadFile(ApprovalsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != body {
		t.Fatalf("Scan rewrote state:\n%s\nwas:\n%s", after, body)
	}
}

func TestScanOnAbsentStateIsEmpty(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")
	records, malformed, err := Scan(home)
	if err != nil || len(records) != 0 || len(malformed) != 0 {
		t.Fatalf("Scan on absent state = %v, %v, %v", records, malformed, err)
	}
}

func TestScanRejectsUnreadableState(t *testing.T) {
	// A directory at the state file's own path is unreadable-as-bytes on
	// every platform (EISDIR on POSIX, a handle/access error on Windows),
	// never absence: opening it fails without IsNotExist. (A regular file
	// at the home path does NOT work: Windows reports that open as
	// IsNotExist, so the code rightly sees absence there.)
	home := filepath.Join(t.TempDir(), "home")
	if err := os.MkdirAll(ApprovalsPath(home), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Scan(home); err == nil {
		t.Fatal("Scan over an unreadable state succeeded")
	}
}

func writePostureFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestClassifyReportsApprovedChangedAndUnapproved(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(base, "home")
	approved := filepath.Join(base, "a", ".agents", "env.sh")
	changed := filepath.Join(base, "b", ".agents", "env.sh")
	unapproved := filepath.Join(base, "c", ".agents", "env.sh")
	missing := filepath.Join(base, "d", ".agents", "env.sh")
	writePostureFile(t, approved, "approved bytes")
	writePostureFile(t, changed, "new bytes")
	writePostureFile(t, unapproved, "foreign bytes")
	if err := Upsert(home, Record{Path: approved, SHA256: Digest([]byte("approved bytes")), ApprovedBy: ApprovedByManager, ApprovedAt: testStamp}); err != nil {
		t.Fatal(err)
	}
	if err := Upsert(home, Record{Path: changed, SHA256: Digest([]byte("old bytes")), ApprovedBy: ApprovedByOperator, ApprovedAt: testStamp}); err != nil {
		t.Fatal(err)
	}
	records, err := List(home)
	if err != nil {
		t.Fatal(err)
	}
	rows := Classify([]string{missing, unapproved, changed, approved}, records)
	if len(rows) != 3 {
		t.Fatalf("Classify returned %d rows, want 3 (absent files yield no row): %+v", len(rows), rows)
	}
	byPath := map[string]Posture{}
	for _, row := range rows {
		byPath[row.Path] = row
	}
	got := byPath[approved]
	if got.State != PostureApproved || got.ApprovedBy != ApprovedByManager || got.Diagnostic != "" {
		t.Fatalf("approved row = %+v", got)
	}
	got = byPath[changed]
	if got.State != PostureChanged || got.Diagnostic != DiagnosticEnvChanged || got.ApprovedBy != ApprovedByOperator {
		t.Fatalf("changed row = %+v", got)
	}
	got = byPath[unapproved]
	if got.State != PostureUnapproved || got.Diagnostic != DiagnosticEnvUnapproved || got.ApprovedBy != "" {
		t.Fatalf("unapproved row = %+v", got)
	}
	if rows[0].Path != approved || rows[1].Path != changed || rows[2].Path != unapproved {
		t.Fatalf("Classify rows are not sorted by path: %+v", rows)
	}
}

func TestClassifyTreatsUnreadableAsUnapprovedNeverAbsent(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	// A directory is deterministically unreadable-as-bytes on every
	// platform and runner, including a privileged one.
	unreadableDir := filepath.Join(base, "project", ".agents")
	if err := os.MkdirAll(unreadableDir, 0o755); err != nil {
		t.Fatal(err)
	}
	rows := Classify([]string{unreadableDir}, nil)
	if len(rows) != 1 || rows[0].State != PostureUnapproved || rows[0].Diagnostic != DiagnosticEnvUnapproved ||
		rows[0].File != PostureFileUnreadable || rows[0].ApprovedBy != "" {
		t.Fatalf("directory candidate = %+v, want one unreadable-qualified unapproved row", rows)
	}
	if !rows[0].NonCurrent() {
		t.Fatal("an unreadable file is non-current under --check, never an ordinary warning row")
	}
	// A permission-stripped file is the realistic shape; it proves the
	// same verdict wherever the runner honors permission bits.
	stripped := filepath.Join(base, "stripped", "env.sh")
	writePostureFile(t, stripped, "bytes")
	if err := os.Chmod(stripped, 0o000); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(stripped, 0o600) }()
	if _, err := os.ReadFile(stripped); err == nil {
		t.Skip("this environment can read a mode-000 file; the directory case above carries the verdict")
	}
	rows = Classify([]string{stripped}, nil)
	if len(rows) != 1 || rows[0].State != PostureUnapproved || rows[0].File != PostureFileUnreadable {
		t.Fatalf("unreadable candidate = %+v, want one unreadable-qualified unapproved row", rows)
	}
	if !rows[0].NonCurrent() {
		t.Fatal("an unreadable file is non-current under --check, never an ordinary warning row")
	}
}

// TestClassifyKeepsRecordOnUnreadableCandidates proves the record is
// associated before the current bytes are inspected: a recorded file that
// becomes unreadable keeps its approved_by and reports the failed read
// explicitly, never via the ordinary no-record warning path.
func TestClassifyKeepsRecordOnUnreadableCandidates(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(base, "home")
	envPath := filepath.Join(base, "project", ".agents", "env.sh")
	writePostureFile(t, envPath, "approved bytes")
	if _, err := ApproveFile(home, envPath, ApprovedByOperator, testStamp); err != nil {
		t.Fatal(err)
	}
	// Replace the approved file with a directory: deterministically
	// unreadable-as-bytes on every platform and runner.
	if err := os.Remove(envPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(envPath, 0o700); err != nil {
		t.Fatal(err)
	}
	records, err := List(home)
	if err != nil {
		t.Fatal(err)
	}
	rows := Classify([]string{envPath}, records)
	if len(rows) != 1 {
		t.Fatalf("Classify = %+v, want the one recorded row", rows)
	}
	row := rows[0]
	if row.Path != envPath || row.State != PostureApproved || row.ApprovedBy != ApprovedByOperator ||
		row.File != PostureFileUnreadable || row.Diagnostic != "" {
		t.Fatalf("unreadable recorded row = %+v, want approved/operator/unreadable with no diagnostic", row)
	}
	if !row.NonCurrent() {
		t.Fatal("an unreadable recorded file is non-current under --check")
	}
}

func TestPostureNonCurrentCoversChangedMissingAndUnreadable(t *testing.T) {
	for _, row := range []Posture{
		{State: PostureApproved},
		{State: PostureUnapproved, Diagnostic: DiagnosticEnvUnapproved},
	} {
		if row.NonCurrent() {
			t.Fatalf("%+v is current", row)
		}
	}
	for _, row := range []Posture{
		{State: PostureChanged, Diagnostic: DiagnosticEnvChanged, ApprovedBy: ApprovedByOperator},
		{State: PostureApproved, ApprovedBy: ApprovedByOperator, File: PostureFileMissing},
		{State: PostureApproved, ApprovedBy: ApprovedByManager, File: PostureFileUnreadable},
		{State: PostureUnapproved, Diagnostic: DiagnosticEnvUnapproved, File: PostureFileUnreadable},
	} {
		if !row.NonCurrent() {
			t.Fatalf("%+v is non-current", row)
		}
	}
}

func TestAssessFailClosedDetailedDistinguishesUnreadableState(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	// A truly absent state is the normal unapproved-warning case, not an
	// unreadable one.
	absentHome := filepath.Join(base, "absent-home")
	current := filepath.Join(base, "current", ".agents", "env.sh")
	writePostureFile(t, current, "current bytes")
	rows, warnings, stateUnreadable := AssessFailClosedDetailed(absentHome, []string{current})
	if stateUnreadable {
		t.Fatal("an absent approval state is not an unreadable one")
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v, want none", warnings)
	}
	if len(rows) != 1 || rows[0].State != PostureUnapproved || rows[0].File != "" {
		t.Fatalf("rows = %+v, want the ordinary unapproved warning row", rows)
	}
	// A directory at the state file's own path is unreadable-as-bytes on
	// every platform (see TestScanRejectsUnreadableState).
	unreadableHome := filepath.Join(base, "unreadable-home")
	if err := os.MkdirAll(ApprovalsPath(unreadableHome), 0o755); err != nil {
		t.Fatal(err)
	}
	rows, warnings, stateUnreadable = AssessFailClosedDetailed(unreadableHome, []string{current})
	if !stateUnreadable {
		t.Fatal("an unreadable approval state must be reported as unreadable, never as an empty set")
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "cannot read") || !strings.Contains(warnings[0], Filename) {
		t.Fatalf("warnings = %v, want the state read failure", warnings)
	}
	if len(rows) != 1 || rows[0].State != PostureUnapproved {
		t.Fatalf("rows = %+v, want the candidate reported unapproved rather than trusted", rows)
	}
	// A readable state with a malformed line warns without reporting the
	// state unreadable.
	malformedHome := filepath.Join(base, "malformed-home")
	if err := os.MkdirAll(malformedHome, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ApprovalsPath(malformedHome), []byte("malformed-line\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, warnings, stateUnreadable = AssessFailClosedDetailed(malformedHome, nil)
	if stateUnreadable {
		t.Fatal("a malformed line is a warning row, not an unreadable state")
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "malformed") {
		t.Fatalf("warnings = %v, want the malformed line", warnings)
	}
}

func TestClassifyMergesTwoSpellingsIntoOneRow(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	realDir := filepath.Join(base, "real", ".agents")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	realPath := filepath.Join(realDir, "env.sh")
	writePostureFile(t, realPath, "bytes")
	alias := filepath.Join(base, "alias")
	if err := os.Symlink(filepath.Join(base, "real"), alias); err != nil {
		t.Skipf("this host cannot create the symlinked project directory the case needs: %v", err)
	}
	aliasPath := filepath.Join(alias, ".agents", "env.sh")
	home := filepath.Join(base, "home")
	record, err := ApproveFile(home, aliasPath, ApprovedByOperator, testStamp)
	if err != nil {
		t.Fatal(err)
	}
	records, err := List(home)
	if err != nil {
		t.Fatal(err)
	}
	rows := Classify([]string{aliasPath, realPath}, records)
	if len(rows) != 1 || rows[0].Path != record.Path || rows[0].State != PostureApproved {
		t.Fatalf("alias+real Classify = %+v, want one approved row", rows)
	}
}

func TestAssessFailClosedUnionsRecordedAndExtra(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(base, "home")
	recorded := filepath.Join(base, "recorded", ".agents", "env.sh")
	writePostureFile(t, recorded, "recorded bytes")
	if _, err := ApproveFile(home, recorded, ApprovedByManager, testStamp); err != nil {
		t.Fatal(err)
	}
	// A stale record (file deleted after approval) keeps its row: every
	// recorded path stays in the inventory, qualified as missing, while
	// the never-existing unrecorded extra candidate stays out.
	stale := filepath.Join(base, "stale", ".agents", "env.sh")
	writePostureFile(t, stale, "stale bytes")
	if _, err := ApproveFile(home, stale, ApprovedByOperator, testStamp); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(stale); err != nil {
		t.Fatal(err)
	}
	current := filepath.Join(base, "current", ".agents", "env.sh")
	writePostureFile(t, current, "current bytes")
	before, err := os.ReadFile(ApprovalsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	rows, warnings := AssessFailClosed(home, []string{current, filepath.Join(base, "absent", "env.sh")})
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v, want none", warnings)
	}
	if len(rows) != 3 {
		t.Fatalf("AssessFailClosed rows = %+v, want current, recorded and stale", rows)
	}
	if rows[0].Path != current || rows[0].State != PostureUnapproved || rows[0].File != "" {
		t.Fatalf("current row = %+v", rows[0])
	}
	if rows[1].Path != recorded || rows[1].State != PostureApproved || rows[1].ApprovedBy != ApprovedByManager || rows[1].File != "" {
		t.Fatalf("recorded row = %+v", rows[1])
	}
	staleRow := rows[2]
	if staleRow.Path != stale || staleRow.State != PostureApproved || staleRow.ApprovedBy != ApprovedByOperator ||
		staleRow.File != PostureFileMissing || staleRow.Diagnostic != "" {
		t.Fatalf("stale row = %+v, want approved/operator/missing with no diagnostic", staleRow)
	}
	if !staleRow.NonCurrent() {
		t.Fatal("a recorded-but-missing file is non-current under --check")
	}
	after, err := os.ReadFile(ApprovalsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("AssessFailClosed rewrote state:\n%s\nwas:\n%s", after, before)
	}
}

func TestAssessFailClosedWarnsOnMalformedAndKeepsValidRows(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(base, "home")
	valid := filepath.Join(base, "project", ".agents", "env.sh")
	writePostureFile(t, valid, "valid bytes")
	body := valid + "\t" + Digest([]byte("valid bytes")) + "\tmanager\t2026-09-16T00:00:00Z\n" +
		"malformed-line\n"
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ApprovalsPath(home), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	rows, warnings := AssessFailClosed(home, nil)
	if len(warnings) != 1 || !strings.Contains(warnings[0], "line 2") || !strings.Contains(warnings[0], Filename) {
		t.Fatalf("warnings = %v, want the malformed line 2", warnings)
	}
	if len(rows) != 1 || rows[0].Path != valid || rows[0].State != PostureApproved {
		t.Fatalf("rows = %+v, want the one valid approval", rows)
	}
	after, err := os.ReadFile(ApprovalsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != body {
		t.Fatalf("AssessFailClosed repaired state:\n%s\nwas:\n%s", after, body)
	}
}

func TestAssessFailClosedOnUnreadableStateReportsUnapproved(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(base, "home")
	// A directory at the state file's own path is unreadable-as-bytes
	// on every platform (see TestScanRejectsUnreadableState).
	if err := os.MkdirAll(ApprovalsPath(home), 0o755); err != nil {
		t.Fatal(err)
	}
	current := filepath.Join(base, "current", ".agents", "env.sh")
	writePostureFile(t, current, "current bytes")
	rows, warnings := AssessFailClosed(home, []string{current})
	if len(warnings) != 1 || !strings.Contains(warnings[0], Filename) {
		t.Fatalf("warnings = %v, want the unreadable state", warnings)
	}
	if len(rows) != 1 || rows[0].Path != current || rows[0].State != PostureUnapproved ||
		rows[0].Diagnostic != DiagnosticEnvUnapproved {
		t.Fatalf("rows = %+v, want the candidate unapproved", rows)
	}
}
