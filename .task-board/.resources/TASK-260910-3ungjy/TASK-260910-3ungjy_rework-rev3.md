# Rework brief — TASK-260910-3ungjy, revision 3 (S6 approval commands and status posture)

Revision 2 was rejected with three corrections
(`TASK-260910-3ungjy_review-verdict-rev2.md`, all P2, all in the posture path).
Everything else passed; keep it unchanged unless a correction touches it.
The rule behind all three is environments §8.4: absence and read failure are
different facts, and a read failure is never evidence that no approval exists.

## Corrections (all required)
- **R1 — every recorded path stays in the inventory.** `hookapproval`
  posture (`Classify`/`Posture`, ~`hookapproval.go:454`) drops absent
  candidates before consulting the records. A valid record whose file has
  disappeared MUST keep its posture row in `curator status` and `curator env
  status` (text and JSON): path, `approved_by`, and a state that says the
  bytes are missing — spelled with the existing closed vocabulary (no new
  shell-hook diagnostic code; §8.4 diagnostics are exactly
  `shell_hook_env_unapproved` / `shell_hook_env_changed`): report the row as
  recorded-with-missing-file (e.g. state `approved`, plus an explicit
  `file: missing` / `bytes: absent` field or text suffix that never claims a
  successful digest comparison), and make `--check` treat it as
  non-current (the recorded approval no longer matches anything on disk).
  Never-existing, unrecorded optional env files stay out of the inventory.
  Production-command tests for text and JSON on both status surfaces.
- **R2 — unreadable recorded candidates keep their record.** Associate the
  record BEFORE inspecting current bytes; on a stat/read failure keep
  `approved_by`, report the failed read explicitly (text and JSON) through
  the existing non-current/error machinery (an unreadable file is
  non-current under `--check`), not through the ordinary no-record warning
  path; `curator hook approvals`/`approve` error text likewise names the
  read failure, not "unapproved". Regression coverage through both real
  status entry points with an otherwise-current matrix so `env status
  --check`'s exit code is evidence of the trust posture alone.
- **R3 — JSON mode reports approval-state read failures.**
  `AssessFailClosed`'s read-error warning is dropped in JSON mode
  (`cmd/curator/main.go` ~:805 renders warnings only in text mode;
  `internal/envprofile/status.go` ~:129 excludes them). With an unreadable
  approval state (directory at `<manager-home>/hook-approvals.tsv`) both
  `status --check --json` and `env status --check --json` MUST surface the
  failure — in the report's established error representation (or stderr,
  if that is how other JSON-mode errors are surfaced) — and treat the
  unavailable record set as unreadable, propagating to currentness
  (`--check` non-current / error exit), never as an empty set. A truly
  absent approval state remains the normal unapproved-warning case. Paired
  absent/unreadable production-entry tests for both commands and both
  output modes.

## Validation and handoff
Narrow gates with `CURATOR_CONFORMANCE_ROOT` set (`go build ./... && go vet
./... && gofmt -l internal cmd && go test -count=1 ./cmd/curator/...
./internal/hookapproval/... ./internal/envprofile/... ./internal/shell/...`),
`set -o pipefail`, exit codes quoted; update `TASK-260910-3ungjy_results.md`
with a "Revision 3" section (per-correction file:line, transcripts); hand off
with `task-board handoff TASK-260910-3ungjy --role developer` (the runtime
runs the hosted gate). Worktree and rules unchanged
(`TASK-260910-3ungjy_brief.md`, `remediation-manager-producer-rules.md`);
do not touch the first leaf's hook text or trust decision logic.
