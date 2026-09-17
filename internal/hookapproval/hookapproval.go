// Package hookapproval owns the manager-home shell-hook approval state of
// Spec profiles/manager.md §8.2 (finding S6).
//
// The state lives below the manager home used by the rest of the manager
// (Config.Home, the directory holding the config file), in the single file
// <manager-home>/hook-approvals.tsv. It is outside every profile, package,
// and project surface: this package never reads a record from package,
// project, or profile data, and no such data influences a lookup.
//
// File format (read directly by the emitted POSIX and PowerShell hooks, so it
// stays parseable without JSON tooling): one record per line, four TAB
// separated fields, LF line endings, no header, sorted by path:
//
//		path \t sha256 \t approved_by \t approved_at \n
//
//	  - path: absolute, canonicalized env-file path (see Canonicalize;
//	    on Windows the native spelling with an uppercase drive letter).
//	  - sha256: lowercase hex SHA-256 over the exact bytes sourced (64 chars).
//	  - approved_by: exactly "manager" or "operator".
//	  - approved_at: RFC 3339 timestamp at which the digest was recorded.
//
// A path containing TAB, LF, or CR is refused at write time: it could never
// round-trip through this format. Canonicalization resolves symlinks
// (absolute + EvalSymlinks, see Canonicalize), so two spellings of one file
// resolve to one record. Writes are atomic (same-directory temporary file
// plus one rename onto the target) and never expose a partially written
// record set; a failed publication leaves the prior record set untouched.
package hookapproval

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// Closed trust sources of Spec §8.1.
const (
	ApprovedByManager  = "manager"
	ApprovedByOperator = "operator"
)

// Closed shell-hook trust diagnostics of Spec §8.4. The emitted hooks carry
// the same two spellings (internal/shell twins them); a cross-package test
// pins the agreement so the three surfaces cannot drift apart.
const (
	DiagnosticEnvUnapproved = "shell_hook_env_unapproved"
	DiagnosticEnvChanged    = "shell_hook_env_changed"
)

// Closed posture states of Spec §8.6: the trust state of one known project
// env file as reported by `curator status`, `curator env status` and the
// approval listing.
const (
	PostureApproved   = "approved"
	PostureUnapproved = "unapproved"
	PostureChanged    = "changed"
)

// Closed file qualifiers for a posture row whose current bytes could not be
// inspected. They extend the §8.6 state without adding a §8.4 diagnostic
// code (the two codes stay closed): the row keeps the state its record
// implies and the qualifier states plainly that no digest comparison ran.
const (
	// PostureFileMissing marks a recorded file whose bytes are absent: the
	// record exists but there is nothing on disk it could authorize.
	PostureFileMissing = "missing"
	// PostureFileUnreadable marks a file whose bytes failed to read
	// (stat or read error): absence and read failure stay different
	// facts, and the failure is never evidence about the record.
	PostureFileUnreadable = "unreadable"
)

// Closed record grammar of Spec §8.2, shared with the emitted hooks: the Go
// reader below is the single source of truth, and the POSIX and PowerShell
// hook templates are generated from these same constants, so the three
// consumers cannot drift apart.
const (
	// RecordFieldCount is the exact TAB-separated field count of one state line.
	RecordFieldCount = 4
	// SHA256HexLength is the exact length of the lowercase hex digest field.
	SHA256HexLength = 64
	// TimestampShape is the RFC 3339 shape check (POSIX ERE, also valid RE2
	// and .NET syntax) embedded in the emitted hooks. Consumers additionally
	// range-check the captured date and time fields; the Go reader parses
	// with time.Parse(time.RFC3339), which the shape mirrors.
	TimestampShape = `^[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]:[0-9][0-9](\.[0-9][0-9]*)?(Z|[+-][0-9][0-9]:[0-9][0-9])$`
)

// Filename is the manager-home file holding the approval record set.
const Filename = "hook-approvals.tsv"

// Record is one closed approval record of Spec §8.2.
type Record struct {
	Path       string
	SHA256     string
	ApprovedBy string
	ApprovedAt time.Time
}

// ApprovalsPath returns the state file below a manager home.
func ApprovalsPath(home string) string { return filepath.Join(home, Filename) }

// Canonicalize resolves a path to the single key its record is stored under:
// absolute, symlink-resolved, and clean, so two spellings of the same file
// (lexical detours, directory aliases, file symlinks) share one record. A
// relative path resolves against the process working directory. A path with
// missing trailing components resolves its longest existing ancestor and
// reattaches the remainder lexically, so records can be written before every
// component exists; a wholly unresolvable path falls back to its lexical
// absolute spelling rather than failing.
//
// Windows identity rule (the single definition the emitted POSIX hook cites):
// on Windows the canonical spelling is the native absolute path with
// backslash separators and an uppercase drive letter (UNC paths keep their
// leading double backslash), and record matching folds case, because the
// Windows filesystem is case-insensitive while Git Bash observes the same
// file through MSYS spelling (/c/...). The POSIX hook therefore maps its
// resolved candidate through `cygpath -w`, uppercases the drive letter, and
// compares case-insensitively in its record lookup; this package normalizes
// the drive letter here and compares with pathsEqual. Both sides resolve
// symlinks before this normalization, so native and MSYS spellings of one
// file share one record. When `cygpath` is absent the hook cannot prove the
// identity and warns plus refuses under both rollout profiles, never
// sourcing silently.
func Canonicalize(path string) (string, error) {
	canonical, err := canonicalizeResolve(path)
	if err != nil {
		return "", err
	}
	return normalizeWindowsDrive(canonical), nil
}

func canonicalizeResolve(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("hookapproval: path is empty")
	}
	if !filepath.IsAbs(path) {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("hookapproval: resolve %q: %w", path, err)
		}
		path = absolute
	}
	path = filepath.Clean(path)
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return filepath.Clean(resolved), nil
	}
	var rest []string
	current := path
	for {
		parent := filepath.Dir(current)
		if parent == current {
			root, err := filepath.EvalSymlinks(current)
			if err != nil {
				return path, nil
			}
			return filepath.Clean(reattach(root, rest)), nil
		}
		rest = append(rest, filepath.Base(current))
		if resolved, err := filepath.EvalSymlinks(parent); err == nil {
			return filepath.Clean(reattach(resolved, rest)), nil
		}
		current = parent
	}
}

// normalizeWindowsDrive uppercases the drive letter of a native Windows
// absolute path (C:\... or C:/...); every other spelling passes through.
// It runs after symlink resolution so both the stored record and the lookup
// key carry the same drive form before the case-insensitive comparison.
func normalizeWindowsDrive(path string) string {
	if runtime.GOOS != "windows" {
		return path
	}
	if len(path) >= 3 && path[1] == ':' && (path[2] == '\\' || path[2] == '/') {
		if c := path[0]; c >= 'a' && c <= 'z' {
			return string(c-'a'+'A') + path[1:]
		}
	}
	return path
}

// pathsEqual reports whether two canonical paths name one record: exact on
// Unix, ASCII/Unicode case-folded on Windows, mirroring the filesystem's
// own identity. The emitted POSIX hook folds the same way under MSYS/Git
// Bash/Cygwin (awk tolower on both sides of the lookup).
func pathsEqual(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// reattach joins ancestor-resolved rest components (innermost last) below a
// resolved directory.
func reattach(resolved string, rest []string) string {
	out := resolved
	for i := len(rest) - 1; i >= 0; i-- {
		out = filepath.Join(out, rest[i])
	}
	return out
}

// Digest returns the lowercase hex SHA-256 of data.
func Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// Validate reports whether r is a well-formed closed record.
func (r Record) Validate() error {
	if r.Path == "" || !filepath.IsAbs(r.Path) || filepath.Clean(r.Path) != r.Path {
		return fmt.Errorf("hookapproval: path %q is not an absolute canonical path", r.Path)
	}
	if strings.ContainsAny(r.Path, "\t\n\r") {
		return fmt.Errorf("hookapproval: path %q carries a field separator", r.Path)
	}
	if len(r.SHA256) != SHA256HexLength {
		return fmt.Errorf("hookapproval: sha256 %q is not %d hex characters", r.SHA256, SHA256HexLength)
	}
	for _, c := range r.SHA256 {
		digit := c >= '0' && c <= '9'
		lower := c >= 'a' && c <= 'f'
		if !digit && !lower {
			return fmt.Errorf("hookapproval: sha256 %q is not lowercase hex", r.SHA256)
		}
	}
	if r.ApprovedBy != ApprovedByManager && r.ApprovedBy != ApprovedByOperator {
		return fmt.Errorf("hookapproval: approved_by %q is not manager or operator", r.ApprovedBy)
	}
	if r.ApprovedAt.IsZero() {
		return fmt.Errorf("hookapproval: approved_at is missing")
	}
	return nil
}

// List reads every approval record. An absent state file is an empty set; a
// malformed line or an unreadable file is an error, never an empty set.
func List(home string) ([]Record, error) {
	payload, err := os.ReadFile(ApprovalsPath(home)) // #nosec G304 -- manager-home state path
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("hookapproval: read state: %w", err)
	}
	if len(payload) == 0 {
		return nil, nil
	}
	lines := strings.Split(string(payload), "\n")
	records := make([]Record, 0, len(lines))
	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			break // trailing newline
		}
		fields := strings.Split(line, "\t")
		if len(fields) != RecordFieldCount || fields[0] == "" {
			return nil, fmt.Errorf("hookapproval: line %d is malformed", i+1)
		}
		stamp, err := time.Parse(time.RFC3339, fields[3])
		if err != nil {
			return nil, fmt.Errorf("hookapproval: line %d carries a bad timestamp: %w", i+1, err)
		}
		record := Record{Path: fields[0], SHA256: fields[1], ApprovedBy: fields[2], ApprovedAt: stamp}
		if err := record.Validate(); err != nil {
			return nil, fmt.Errorf("hookapproval: line %d: %w", i+1, err)
		}
		records = append(records, record)
	}
	return records, nil
}

// Lookup returns the record authorizing path, canonicalizing two spellings of
// one file to one record before comparison.
func Lookup(home, path string) (Record, bool, error) {
	canonical, err := Canonicalize(path)
	if err != nil {
		return Record{}, false, err
	}
	records, err := List(home)
	if err != nil {
		return Record{}, false, err
	}
	for _, record := range records {
		if pathsEqual(record.Path, canonical) {
			return record, true, nil
		}
	}
	return Record{}, false, nil
}

// Upsert records (or re-records after a change) the digest for a path.
func Upsert(home string, record Record) error {
	canonical, err := Canonicalize(record.Path)
	if err != nil {
		return err
	}
	record.Path = canonical
	if err := record.Validate(); err != nil {
		return err
	}
	records, err := List(home)
	if err != nil {
		return err
	}
	replaced := false
	for i, existing := range records {
		if pathsEqual(existing.Path, canonical) {
			records[i] = record
			replaced = true
			break
		}
	}
	if !replaced {
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Path < records[j].Path })
	return writeRecords(home, records)
}

// ApproveFile hashes the live bytes of path and records the digest. It fails
// without recording when the file is absent or unreadable, and never trusts
// a path with no readable bytes.
func ApproveFile(home, path, approvedBy string, at time.Time) (Record, error) {
	if approvedBy != ApprovedByManager && approvedBy != ApprovedByOperator {
		return Record{}, fmt.Errorf("hookapproval: approved_by %q is not manager or operator", approvedBy)
	}
	canonical, err := Canonicalize(path)
	if err != nil {
		return Record{}, err
	}
	payload, err := os.ReadFile(canonical) // #nosec G304 -- caller-selected env file
	if err != nil {
		return Record{}, fmt.Errorf("hookapproval: read %s: %w", canonical, err)
	}
	record := Record{Path: canonical, SHA256: Digest(payload), ApprovedBy: approvedBy, ApprovedAt: at.UTC()}
	if err := Upsert(home, record); err != nil {
		return Record{}, err
	}
	return record, nil
}

// Revoke removes the record for path. A path with no record leaves state
// unchanged and reports false.
func Revoke(home, path string) (bool, error) {
	canonical, err := Canonicalize(path)
	if err != nil {
		return false, err
	}
	records, err := List(home)
	if err != nil {
		return false, err
	}
	kept := records[:0]
	removed := false
	for _, record := range records {
		if pathsEqual(record.Path, canonical) {
			removed = true
			continue
		}
		kept = append(kept, record)
	}
	if !removed {
		return false, nil
	}
	// A failed publication leaves the record in place, so the removal did
	// not happen: report it alongside the error.
	if err := writeRecords(home, kept); err != nil {
		return false, err
	}
	return true, nil
}

// Scan reads the record set tolerantly for reporting surfaces (`curator hook
// approvals`, status posture): well-formed lines become records sorted by
// path, malformed lines are reported by 1-based number and skipped — exactly
// how the emitted hooks treat them (an invalid line never authorizes). An
// absent state file is an empty set with no malformed lines; an unreadable
// file is an error, never an empty set. Scan never mutates state: it neither
// rewrites nor "repairs" a malformed line.
//
// List stays the strict reader for trust decisions (Lookup, Upsert, Revoke):
// a malformed line fails closed there rather than being skipped.
func Scan(home string) (records []Record, malformed []int, err error) {
	payload, err := os.ReadFile(ApprovalsPath(home)) // #nosec G304 -- manager-home state path
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("hookapproval: read state: %w", err)
	}
	if len(payload) == 0 {
		return nil, nil, nil
	}
	lines := strings.Split(string(payload), "\n")
	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			break // trailing newline
		}
		record, ok := parseRecord(line)
		if !ok {
			malformed = append(malformed, i+1)
			continue
		}
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Path < records[j].Path })
	return records, malformed, nil
}

// parseRecord decodes one state line into a validated closed record.
func parseRecord(line string) (Record, bool) {
	fields := strings.Split(line, "\t")
	if len(fields) != RecordFieldCount || fields[0] == "" {
		return Record{}, false
	}
	stamp, err := time.Parse(time.RFC3339, fields[3])
	if err != nil {
		return Record{}, false
	}
	record := Record{Path: fields[0], SHA256: fields[1], ApprovedBy: fields[2], ApprovedAt: stamp}
	if err := record.Validate(); err != nil {
		return Record{}, false
	}
	return record, true
}

// Posture is the trust state of one known project env file (Spec §8.6):
// approved (recorded digest matches), unapproved (no record) or changed
// (recorded digest differs; re-approval required). Diagnostic carries the
// §8.4 code for the two untrusted states; ApprovedBy names the recorder
// for recorded files. File qualifies a row whose current bytes could not
// be inspected: PostureFileMissing (recorded file absent — the record no
// longer matches anything on disk) or PostureFileUnreadable (stat/read
// failed — a read failure is never evidence that no approval exists).
// File is empty when the bytes were compared, or when no record exists
// and the file reads.
type Posture struct {
	Path       string `json:"path"`
	State      string `json:"state"`
	Diagnostic string `json:"diagnostic,omitempty"`
	ApprovedBy string `json:"approved_by,omitempty"`
	File       string `json:"file,omitempty"`
}

// NonCurrent reports whether the row fails `--check`: a changed file, or a
// row whose bytes are missing or unreadable (File is set). An unapproved
// file whose bytes read stays a warning row and never fails the check.
func (row Posture) NonCurrent() bool {
	return row.State == PostureChanged || row.File != ""
}

// Classify compares each candidate file against the approval records and
// returns one posture row per known file, sorted by path. Two spellings of
// one file resolve to one row. The record is associated BEFORE the current
// bytes are inspected, so a read failure can never drop the record it
// failed to compare against: absence and read failure stay different
// facts, and a read failure is never evidence that no approval exists.
//
//   - An absent file with no record yields no row: there are no bytes the
//     hook would source and no approval claiming them. An absent file WITH
//     a record keeps its row (approved, with its approved_by and the
//     missing qualifier): the approval no longer matches anything on disk,
//     which `--check` treats as non-current, and the row never claims a
//     digest comparison ran.
//   - A file whose bytes cannot be read (stat or read error) keeps its
//     record when one exists and reports the failed read explicitly via
//     the unreadable qualifier — never via the ordinary no-record warning
//     path. An unreadable file is non-current under `--check`.
func Classify(candidates []string, records []Record) []Posture {
	seen := map[string]bool{}
	var rows []Posture
	for _, candidate := range candidates {
		canonical, err := Canonicalize(candidate)
		if err != nil {
			continue
		}
		key := canonical
		if runtime.GOOS == "windows" {
			key = strings.ToLower(canonical)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		var (
			record Record
			found  bool
		)
		for _, candidate := range records {
			if pathsEqual(candidate.Path, canonical) {
				record, found = candidate, true
				break
			}
		}
		if _, err := os.Stat(canonical); err != nil {
			if os.IsNotExist(err) {
				if !found {
					continue
				}
				rows = append(rows, Posture{Path: canonical, State: PostureApproved, ApprovedBy: record.ApprovedBy, File: PostureFileMissing})
				continue
			}
			rows = append(rows, unreadablePosture(canonical, record, found))
			continue
		}
		payload, err := os.ReadFile(canonical) // #nosec G304 -- caller-selected env file
		if err != nil {
			rows = append(rows, unreadablePosture(canonical, record, found))
			continue
		}
		digest := Digest(payload)
		switch {
		case !found:
			rows = append(rows, Posture{Path: canonical, State: PostureUnapproved, Diagnostic: DiagnosticEnvUnapproved})
		case record.SHA256 == digest:
			rows = append(rows, Posture{Path: canonical, State: PostureApproved, ApprovedBy: record.ApprovedBy})
		default:
			rows = append(rows, Posture{Path: canonical, State: PostureChanged, Diagnostic: DiagnosticEnvChanged, ApprovedBy: record.ApprovedBy})
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Path < rows[j].Path })
	return rows
}

// unreadablePosture reports a candidate whose bytes failed to read. A
// recorded file keeps its state and approved_by with the unreadable
// qualifier and no diagnostic (neither "no record" nor "digest differs"
// was proven); an unrecorded file keeps the unapproved diagnostic the
// absent record earns, qualified as unreadable so it cannot be mistaken
// for an ordinary warning row. Both are non-current under `--check`.
func unreadablePosture(canonical string, record Record, found bool) Posture {
	if found {
		return Posture{Path: canonical, State: PostureApproved, ApprovedBy: record.ApprovedBy, File: PostureFileUnreadable}
	}
	return Posture{Path: canonical, State: PostureUnapproved, Diagnostic: DiagnosticEnvUnapproved, File: PostureFileUnreadable}
}

// AssessFailClosed reports the shell-hook trust posture (Spec §8.6) over the
// known project env files: every recorded path, plus the caller-supplied
// extra candidates (the current projects' `.agents/env.sh`/`.agents/env.ps1`).
// Rows are sorted by path; warnings name malformed state lines (skipped, never
// repaired) or an unreadable state file, in which case the candidates are
// reported unapproved rather than trusted. It never mutates state and never
// returns an error: reporting stays available even when the state cannot be
// proven.
//
// Callers that drive `--check` use AssessFailClosedDetailed instead: only it
// distinguishes an unreadable approval state (non-current) from a merely
// malformed line (warning row).
func AssessFailClosed(home string, extra []string) (rows []Posture, warnings []string) {
	rows, warnings, _ = AssessFailClosedDetailed(home, extra)
	return rows, warnings
}

// AssessFailClosedDetailed is AssessFailClosed with the state-read verdict:
// stateUnreadable is true when the approval state itself could not be read,
// in which case the unavailable record set MUST be treated as unreadable
// (propagating to currentness), never as an empty set. A truly absent state
// file is not unreadable: it is the normal unapproved-warning case.
func AssessFailClosedDetailed(home string, extra []string) (rows []Posture, warnings []string, stateUnreadable bool) {
	records, malformed, err := Scan(home)
	if err != nil {
		return Classify(extra, nil), []string{fmt.Sprintf("cannot read %s: %v; candidates reported unapproved", Filename, err)}, true
	}
	for _, line := range malformed {
		warnings = append(warnings, fmt.Sprintf("line %d of %s is malformed; skipped", line, Filename))
	}
	candidates := make([]string, 0, len(records)+len(extra))
	for _, record := range records {
		candidates = append(candidates, record.Path)
	}
	candidates = append(candidates, extra...)
	return Classify(candidates, records), warnings, false
}

func writeRecords(home string, records []Record) error {
	if err := os.MkdirAll(home, 0o755); err != nil {
		return fmt.Errorf("hookapproval: create state dir: %w", err)
	}
	var body strings.Builder
	for _, record := range records {
		body.WriteString(record.Path)
		body.WriteByte('\t')
		body.WriteString(record.SHA256)
		body.WriteByte('\t')
		body.WriteString(record.ApprovedBy)
		body.WriteByte('\t')
		body.WriteString(record.ApprovedAt.UTC().Format(time.RFC3339))
		body.WriteByte('\n')
	}
	temporary, err := os.CreateTemp(home, ".hook-approvals.*.tmp")
	if err != nil {
		return fmt.Errorf("hookapproval: stage state: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if _, err := temporary.WriteString(body.String()); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("hookapproval: stage state: %w", err)
	}
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("hookapproval: stage state: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("hookapproval: stage state: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("hookapproval: stage state: %w", err)
	}
	target := ApprovalsPath(home)
	// One atomic rename onto the target (POSIX rename; os.Rename on Windows
	// replaces the destination). A failed publication reports the error and
	// leaves the prior record set untouched: the target is never removed to
	// make room, so readers keep observing the last valid state.
	if err := renameFile(temporaryPath, target); err != nil {
		return fmt.Errorf("hookapproval: publish state: %w", err)
	}
	return nil
}

// renameFile publishes staged state. It is a variable so the failure-path
// test can prove a failed publication preserves the last valid state.
var renameFile = os.Rename
