package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/relux-works/curator/internal/hookapproval"
	"github.com/relux-works/curator/internal/manifest"
)

const hookUsage = `curator hook: approve project env files for the shell-hook trust gate (Manager profile §8)

Usage:
  curator hook approve <path>      record the current bytes as approved_by operator
  curator hook approvals           list every approval record read-only
  curator hook revoke <path>       remove the record for the path
`

// cmdHook is the closed approval command surface of Spec §8.3: exactly
// approve, approvals, and revoke. State lives below the manager home (the
// directory holding the config file), which needs no loadable config — like
// shell-init --install and like the hook itself, these commands work from the
// config path alone, before any bootstrap.
func (c cli) cmdHook(args []string) int {
	if len(args) == 0 {
		_, _ = fmt.Fprint(c.stderr, hookUsage)
		return exitUsage
	}
	home := filepath.Dir(c.config.Path())
	switch args[0] {
	case "approve":
		return c.cmdHookApprove(home, args[1:])
	case "approvals":
		return c.cmdHookApprovals(home, args[1:])
	case "revoke":
		return c.cmdHookRevoke(home, args[1:])
	}
	_, _ = fmt.Fprintf(c.stderr, "curator: unknown hook subcommand %q\n\n%s", args[0], hookUsage)
	return exitUsage
}

// cmdHookApprove records the digest of the operand's current bytes as
// approved_by operator, re-recording after a change. The operand resolves to
// the same canonical absolute identity the hook and the approval state use
// (absolute, symlinks resolved, Windows spelling folded). It fails without
// recording when the file is absent or unreadable, with distinct exit text
// for the two ("unreadable is never absence"), and never trusts a path with
// no readable bytes. Publication is atomic through hookapproval.Upsert; when
// the record already matches (same digest, recorded by the operator) the
// command is idempotent and leaves state untouched.
func (c cli) cmdHookApprove(home string, args []string) int {
	if len(args) != 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: hook approve needs exactly one path: curator hook approve <path>")
		return exitUsage
	}
	canonical, err := hookapproval.Canonicalize(args[0])
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	payload, err := os.ReadFile(canonical) // #nosec G304 -- operator-selected env file
	if err != nil {
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintf(c.stderr, "curator: %s: file is absent; nothing recorded\n", canonical)
		} else {
			_, _ = fmt.Fprintf(c.stderr, "curator: %s: file is unreadable (%v); nothing recorded\n", canonical, err)
		}
		return exitFail
	}
	if existing, found, lookupErr := hookapproval.Lookup(home, canonical); lookupErr != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator: cannot record approval:", lookupErr)
		return exitFail
	} else if found && existing.SHA256 == hookapproval.Digest(payload) && existing.ApprovedBy == hookapproval.ApprovedByOperator {
		_, _ = fmt.Fprintln(c.stdout, "approved", canonical)
		return exitOK
	}
	record := hookapproval.Record{
		Path:       canonical,
		SHA256:     hookapproval.Digest(payload),
		ApprovedBy: hookapproval.ApprovedByOperator,
		ApprovedAt: time.Now().UTC(),
	}
	if err := hookapproval.Upsert(home, record); err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator: cannot record approval:", err)
		return exitFail
	}
	_, _ = fmt.Fprintln(c.stdout, "approved", canonical)
	return exitOK
}

// cmdHookApprovals lists every approval record read-only: one TAB-separated
// path, sha256, approved_by, approved_at row per record on stdout, stable
// order (sorted by path). It never mutates state — a malformed record is
// reported on stderr and skipped, matching how the hooks treat it.
func (c cli) cmdHookApprovals(home string, args []string) int {
	if len(args) != 0 {
		_, _ = fmt.Fprintln(c.stderr, "curator: hook approvals takes no arguments")
		return exitUsage
	}
	records, malformed, err := hookapproval.Scan(home)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator: cannot list approvals:", err)
		return exitFail
	}
	for _, line := range malformed {
		_, _ = fmt.Fprintf(c.stderr, "warning: line %d of %s is malformed; skipped\n", line, hookapproval.Filename)
	}
	for _, record := range records {
		_, _ = fmt.Fprintf(c.stdout, "%s\t%s\t%s\t%s\n",
			record.Path, record.SHA256, record.ApprovedBy, record.ApprovedAt.UTC().Format(time.RFC3339))
	}
	return exitOK
}

// cmdHookRevoke removes the record for the operand's canonical path. On a
// path with no record it leaves state byte-identical and reports that there
// was nothing to revoke, exiting zero: the approval library returns a benign
// (false, nil) for a missing record — unlike declaration removes, where a
// missing entry is an error — and Spec §8.3 frames this case as "leave state
// unchanged and report", not as a failure.
func (c cli) cmdHookRevoke(home string, args []string) int {
	if len(args) != 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: hook revoke needs exactly one path: curator hook revoke <path>")
		return exitUsage
	}
	canonical, err := hookapproval.Canonicalize(args[0])
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	removed, err := hookapproval.Revoke(home, canonical)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator: cannot revoke approval:", err)
		return exitFail
	}
	if !removed {
		_, _ = fmt.Fprintf(c.stdout, "no approval for %s; nothing to revoke\n", canonical)
		return exitOK
	}
	_, _ = fmt.Fprintln(c.stdout, "revoked", canonical)
	return exitOK
}

// hookTrustPosture assesses the shell-hook trust posture (Spec §8.6) over
// the known project env files: every recorded path, plus the
// `.agents/env.sh`/`.agents/env.ps1` of each project root that is a project
// (carries Skillfile.json). It never mutates state. stateUnreadable reports
// whether the approval state itself could not be read, which `--check`
// treats as non-current (an unreadable record set is never an empty set).
func hookTrustPosture(home string, roots []string) (rows []hookapproval.Posture, warnings []string, stateUnreadable bool) {
	var extra []string
	for _, root := range roots {
		extra = append(extra, projectTrustCandidates(root)...)
	}
	return hookapproval.AssessFailClosedDetailed(home, extra)
}

// projectTrustCandidates returns the two hook-sourced env files of root when
// root is a project, or nothing when it is not.
func projectTrustCandidates(root string) []string {
	if _, err := os.Stat(filepath.Join(root, manifest.Name)); err != nil {
		return nil
	}
	return []string{
		filepath.Join(root, ".agents", "env.sh"),
		filepath.Join(root, ".agents", "env.ps1"),
	}
}

// hookTrustCheckFailed reports whether the posture fails `--check`: a
// changed file, a recorded file whose bytes are missing or unreadable, or
// an unreadable approval state. Unapproved files whose bytes read are
// warning rows and never fail the check.
func hookTrustCheckFailed(rows []hookapproval.Posture, stateUnreadable bool) bool {
	if stateUnreadable {
		return true
	}
	for _, row := range rows {
		if row.NonCurrent() {
			return true
		}
	}
	return false
}

// formatTrustRow renders one posture row: the absolute path, the trust
// state, and — for recorded files — the recorded approved_by value.
// Untrusted rows name the approval command that records the current bytes.
// Rows whose bytes are missing or unreadable say so explicitly and never
// claim a digest comparison ran.
func formatTrustRow(row hookapproval.Posture) string {
	switch {
	case row.File == hookapproval.PostureFileMissing:
		return fmt.Sprintf("shell-hook-trust: %s: approved (approved_by=%s; file is missing, so the recorded digest was not compared; restore the file or run curator hook revoke %s)",
			row.Path, row.ApprovedBy, row.Path)
	case row.File == hookapproval.PostureFileUnreadable && row.ApprovedBy != "":
		return fmt.Sprintf("shell-hook-trust: %s: approved (approved_by=%s; file is unreadable, so the recorded digest was not compared)",
			row.Path, row.ApprovedBy)
	case row.File == hookapproval.PostureFileUnreadable:
		return fmt.Sprintf("shell-hook-trust: %s: %s (file is unreadable; make it readable, then run curator hook approve %s)",
			row.Path, row.Diagnostic, row.Path)
	case row.State == hookapproval.PostureApproved:
		return fmt.Sprintf("shell-hook-trust: %s: approved (approved_by=%s)", row.Path, row.ApprovedBy)
	case row.State == hookapproval.PostureChanged:
		return fmt.Sprintf("shell-hook-trust: %s: %s (approved_by=%s; run curator hook approve %s to record the new bytes)",
			row.Path, row.Diagnostic, row.ApprovedBy, row.Path)
	default:
		return fmt.Sprintf("shell-hook-trust: %s: %s (warning; run curator hook approve %s)",
			row.Path, row.Diagnostic, row.Path)
	}
}
