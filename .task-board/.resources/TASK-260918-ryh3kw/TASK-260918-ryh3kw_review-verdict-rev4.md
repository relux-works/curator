# TASK-260918-ryh3kw review verdict — rev4: CHANGES REQUESTED

Candidate: base 0ffe2e1d, tree 93f1102d (the worktree matches it, apart from 2 untracked test files that are not in the index). Reviewer: claude-opus-5-5 low.

## Passes (checked by me, zsh with pipefail, real exit codes)
- `go test ./internal/stateread/` exit 0; `go test ./internal/envprofile/ -run TestManagerOwnedAbsenceReadsAreGuarded` exit 0.
- stateread.go missingPathWith: the ancestor walk uses os.Stat, stops at the first existing ancestor, and is bounded by the path depth. An existing non-directory ancestor makes the path unreadable. The hosted gate on rev4 is green, Windows included.

## Blocking finding
F1. The trunk carry was not done, and the ledger names a blocker that no longer exists.
- ryh3kw-gatefix-1 and gatefix-2 both require `git fetch origin main` and a combine. rev4 is still based on 0ffe2e1d. origin/main (eca2bf27) contains TASK-260927-1wc76r.
- `.github/ci/conformance-gaps.tsv:7-8` still attribute backup-record-absent-restore-nothing and backup-record-unreadable-restore-stops to TASK-260927-1wc76r with the blocker "no env unmanage --restore-backups production entry".
- That entry exists on trunk: origin/main cmd/curator/env.go:59 (`--restore-backups`) and cmd/curator/env_unmanage_test.go:125. The ledger claim is false after the landing, and review-note item 3 requires this case to be flagged.
- Overlapping trunk changes also touch internal/envprofile/status.go and conformance-gaps.tsv, so this candidate would need a combine at integration anyway.

## Required for rev5
1. Combine with origin/main, keeping both sides.
2. Drive the 2 restore vectors through `env unmanage --restore-backups`, or re-attribute them to the concrete blocker that actually stops them. Kill one mutant for each rule.
3. Check that `git diff --name-only origin/main -- . ':!.task-board'` lists only your paths.
4. Republish, and hand off only on a green hosted gate.
No CHANGELOG/LOGBOOK edit (confirmed absent in rev4).
