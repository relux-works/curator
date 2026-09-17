# TASK-260910-3ungjy — revision 2 independent review

Verdict: **changes_requested**. Route: **to-dev**. No acceptance or integration attempted.

Reviewed candidate `ba52a45b6419a58e9b3c95a1fcf3a6a44c89b4e2`, cumulative base `aa46ecd80ad0b83853586454723ea99fb76977a9`. Published patch SHA-256 verified: `7fc8941a533cd1b78038d0851cbd6c61d9a1a28681b1da5f5fc77a0d27abe4e4`. All 873 candidate blobs outside `.task-board` match the live worktree. Validation used `git archive` of this exact tree in `/tmp/TASK-260910-3ungjy-review.N8ycG2`; probes/mutants used a separate disposable copy. No candidate edits, commits, or files left in the Story worktree.

Read campaign rules, producer/reviewer briefs, rev1 gate analysis, producer results, published patch and manager.md §8.2–8.7, plus sibling results and their revision addenda. The cumulative CR includes the sibling hook changes; comparison with current checkpoint HEAD `1e6a16578f32c6a1d078cda23c48ce54eea9c89d` isolates 11 leaf files. Hook text, existing trust decision logic, SPEC_PIN and platform ledger are unchanged by this leaf. One sibling symlink-test skip message changed wording only.

## Required corrections

### R1 — P2: recorded-but-deleted files disappear from both status commands

`internal/hookapproval/hookapproval.go:454–456` drops every absent candidate before consulting the records. This contradicts the brief's explicit inventory, **every recorded path**, and manager §8.6's each-known-file requirement. Approval is still present in the manager state, but deleting the env file makes both text/JSON status omit that record entirely.

Reproduction through production CLI `run()`: register an empty project, create and approve `.agents/env.sh`, delete it, then run `status app --check` and both status JSON surfaces. `status app --check` exits 0 with empty stdout; JSON has `shell_hook_trust: null`. `env status` also omits the recorded path. The latter fixture has unrelated unprovisioned homes, so its exit 1 is not evidence of a trust-posture check.

Correction: retain a posture row for every valid recorded identity even when its file disappears, preserve `approved_by`, and distinguish missing bytes explicitly without claiming a successful digest comparison. Keep never-existing, unrecorded optional env files out of the inventory. Add production command tests for text and JSON in both status surfaces. The producer's “known bound” is not an authorized exception to the brief.

### R2 — P2: unreadable recorded candidates lose metadata and look like ordinary unapproved files

`internal/hookapproval/hookapproval.go:454–464` emits an unapproved row for a stat/read failure before looking up the record at lines 467–475. It loses `approved_by` even though the record is valid, and discards the read error. `cmd/curator/hook.go:190–193` then prints the ordinary warning/approval hint with no unreadable explanation.

Reproduction: approve a readable file, replace the file with a directory (portable unreadable-as-bytes fixture), invoke both status commands. Both output an unapproved row lacking `approved_by`; `status app --check` exits 0. This violates §8.6's recorded-by requirement and the campaign's unreadable-evidence distinction. A read failure is not evidence that no approval exists.

Correction: associate the existing record before attempting to inspect current bytes, preserve its metadata on failures, report the failed read explicitly on text and machine-readable surfaces, and ensure unreadable evidence follows the existing non-current/error machinery rather than the ordinary no-record warning path. Do not invent an additional shell-hook diagnostic code. Add regression coverage through both real status entry points, with an otherwise-current matrix when asserting `env status --check`.

### R3 — P2: JSON status silently drops approval-state read failures

`AssessFailClosed` returns a read-error warning at `internal/hookapproval/hookapproval.go:499–501`, but `cmd/curator/main.go:805–813` only renders warnings in non-JSON mode; `internal/envprofile/status.go:129–130` excludes them from JSON entirely. With a directory at `<manager-home>/hook-approvals.tsv`, `status app --check --json` exits 0 and reports an ordinary unapproved row with empty stderr. `env status --check --json` similarly contains no indication of the state-read failure (in the probe, no current-project candidate means its trust value is null).

Correction: preserve/report read failures in JSON mode too, either in the report's established error representation or stderr, and propagate unreadable evidence to currentness rather than silently treating the unavailable record set as an empty set. Keep a truly absent approval state as the normal unapproved-warning case. Add paired absent/unreadable production-entry tests for both commands and output modes.

## Review matrix

| Requirement | Result and evidence |
|---|---|
| Approve current bytes, absent/unreadable refusal, re-record, canonical identity | PASS: `cmd/curator/hook.go:53–89`, existing `Canonicalize` at `hookapproval.go:117`; command tests pass, digest-reuse narrowing mutant killed. Native Windows behavior accepted from hosted lane; not executed locally. |
| Atomic publication and idempotence | PASS: command uses existing `Upsert`, `hookapproval.go:514` same-directory publication; matching operator record early return. Existing failure-publication tests pass. |
| Listing read-only, sorted, malformed tolerated | PASS: `hook.go:97–117`, `Scan` at `hookapproval.go:373`; committed negative/read-only tests pass. |
| Revoke present/absent | PASS: `hook.go:123–149`, existing `Revoke` at `hookapproval.go:337`; missing returns exit 0 with explicit no-op, byte-identity tests pass. |
| Status approved/unapproved/changed normal paths | PASS: `main.go:732–821`, `envprofile/status.go:206–216`, command checks and operator-only narrowing mutant. |
| Complete inventory / recorded metadata / unreadable posture | FAIL: R1–R3 above, independent production-entry probes. |
| JSON additive key | Top-level pin deliberately retains exactly four keys and no builds; tests pass (`status_test.go:1744`). Minor hardening: its per-row checks assert required values and no diagnostic but do not reject arbitrary extra row keys; add an exact row-key-count assertion to fully match the review brief's closed-shape pin. |
| Windows unreadable fixture repair | PASS by inspection: directory at state path in `TestScanRejectsUnreadableState` and `TestAssessFailClosedOnUnreadableStateReportsUnapproved`, no GOOS skip. Local tests and hosted Windows lane green. |
| Existing vectors, emitted hooks, warn-first profile | PASS: 14/14 vector cases, 0 vector skips, real sh/dash/bash/zsh and pwsh; A-warning unchanged. |
| CLI docs, release note and scope | PASS: `docs/cli.md:220,635`, `CHANGELOG.md:28–50` name all commands, posture and additive JSON key. No new spec bytes or pin changes. |
| Validation | All baseline checks below pass; new reviewer probes fail on R1–R3. |

## Independent validation

Shell: bash; `set -o pipefail`. All Go test commands set:
`CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`.

- `go build ./...`: exit 0.
- `go vet ./...`: exit 0.
- `gofmt -l internal cmd`: no output, exit 0.
- `go test -count=1 -timeout 5m ./internal/hookapproval/... ./internal/envfiles/...`: both PASS, exit 0.
- `go test -count=1 -timeout 5m ./cmd/curator/... -run 'TestHook|TestStatus.*ShellHook|TestEnvStatusReportsShellHook|TestStatusReportsRecordedPaths|TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands'`: PASS, 52.638s, exit 0.
- `go test -count=1 -timeout 5m ./internal/envprofile/... -run TestStatusShellHookTrust`: PASS, 7.068s, exit 0.
- `PATH=/tmp/pwsh/app:$PATH go test -count=1 -timeout 5m -v ./internal/shell/...`: full package PASS, 163.413s, exit 0. 14/14 vector families execute; Windows-native-only tests remain platform skips on macOS.
- `golangci-lint run`: 0 issues, exit 0.
- Separate binary E2E script: approve → silent sourcing; revoke → unapproved warning, still sources under A. `sh`, `bash`, real PowerShell: 6/6 activation checks PASS, script exit 0.
- Reviewer probes: `go test -count=1 -timeout 3m -v ./cmd/curator -run '^TestReview'`: exit 1, reproducible assertions for R1–R3; full probe source and transcript attached.

Full cmd/curator, envprofile, install, race and native-Windows suites were not independently replayed. Existing exact-revision hosted validation resource `TASK-260910-3ungjy_change-request_rev2-validation.log` was read: run 35237059691, all required lanes successful, `[exit 0]`, `coverage_unit=exact_command_shard required=1 green=1 failed=0 missing=0 test_case_coverage=unknown`. This is supplemental evidence, not a substitute for the independent checks above. No remote gate was rerun.

## Narrowing mutants

Only disposable copy edited; committed tests invoked with `-count=1`.

1. Narrow re-approval to first-time recording by treating any existing operator record as already current (remove digest equality from the early-return predicate). `TestHookApproveReRecordsAfterChange` fails, “re-approval kept the stale digest”, exit 1.
2. Narrow changed-file `--check` to manager-approved records only. `TestStatusReportsShellHookTrustPosture` fails, “status --check over a changed file = 0, want 1”, exit 1.

**2/2 mutants killed, 0 survivors**. This is coverage of these two selected narrowing shapes, not a claim of exhaustive mutation coverage. Original bytes restored and verified; both targeted tests rerun PASS, exit 0. Mutation harness exit 0.

## Lifecycle and evidence

`task-board spawn goal "$TASK_BOARD_RUN_ID"` returned “Active Goal: none (run is not goal-bound)”. No directives pending. Review findings are also recorded in task-scoped logbook outcome; repository LOGBOOK.md remains untouched per campaign rule. Verdict, raw transcripts, probe source and E2E/mutation scripts are attached before status routing. Rework is ordinary implementation work, not an external blocker; next step is producer fixes and another independent review.
