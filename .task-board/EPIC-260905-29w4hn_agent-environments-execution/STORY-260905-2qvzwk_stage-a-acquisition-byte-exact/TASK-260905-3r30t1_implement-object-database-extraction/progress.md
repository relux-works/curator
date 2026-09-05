## Status
development

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(8))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] internal/gitops extracts from the object database (ls-tree -r -z + cat-file --batch), refuses links/gitlinks/escapes/oversize/platform-duplicate paths, preserves 100755; both callers switched
- [x] Byte-exact testdata (git ls-files --eol proof) and a scratch-repo test reproduce sha256:500ea934...2bced0 under core.autocrlf=true and false with the export-subst placeholder intact; negative test fails if extraction reverts to git archive
- [x] CURATOR_CONFORMANCE_ROOT-driven test reads vectors/snapshot-acquisition.json and skips with a reason on a root without it
- [x] go build, go vet, gofmt, go test -count=1 -timeout 30m ./... green (exact commands and exit codes in the report); signed commits; drafting report attached; no push
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Gating, refusing, validating, authorizing, or attesting behavior covered by negative tests that fail when the gate admits what it must reject, with the production call site named
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] Gate, refusal, validation, authorization, and attestation behavior attacked, not read — positive-path-only evidence is not accepted
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
- [x] Every command, message, state, or refusal named in the AC is driven through the production entry point by a named committed test, or is declared a stated bound. Report coverage as a ratio — `n of m AC rows driven` — and name the production call site for each. Prose in place of the ratio is not evidence.
- [x] Every gate ships at least one NARROWING mutant — the gate stays present and is weakened to admit exactly one member of the class it must reject, and a named test must fail. A delete-only mutant proves only that the gate exists and is not accepted as evidence.
- [x] A gate that inspects source text is additionally attacked by a mutant that PRESERVES the searched-for token and changes behavior, and the mutant harness executes the behavioral suite, not only the static checker.

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260905-269e50, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260905-269e50)
Delivered in worktree /Users/iv/Developer/ReluxWorks/.worktrees/curator-acquisition-byte-exact, branch feat/byte-exact-acquisition, commits f855a34c + 5beced46 (signed, not pushed). gitops.Extract replaces Archive (ls-tree -r -z + cat-file --batch); both callers switched; byte-exact testdata + vector reproduction + negative git-archive test + refusal tests; interop conformance-root test skips on rc.9 root, passes on candidate root. All gates exit 0 (see drafting report). Deviation: fixture .gitattributes stored as gitattributes.fixture because nested attributes normalized siblings in the index. Lint: go vet + gofmt only, golangci-lint not installed.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260905-269e50, pid=58480, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260905-1fb9eb, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260905-1fb9eb)
agent completed: [reviewer] reviewer (claude) (exit=1)
spawn limit exhausted: the retry was refused before any subscription group was subtracted (reason selection_snapshot_unavailable, attempts 1, evidence RUN-260905-1fb9eb); provider reported: You've hit your session limit · resets 4:20pm (Asia/Tbilisi)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260905-d3a1ca, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260905-d3a1ca)
Rework 1 in progress: commit 5abec244 (signed) on feat/byte-exact-acquisition restores TestArchiveRejectsLinks (mktree 120000 entry, no ln, all GOOS), gives TestConformanceSnapshotAcquisition a root-content skip text, and registers ledger rows. ledger-consistency ok, gate-selftest 81/0, focused -race green, candidate-root interop test PASS. Full test-gate.sh on pinned root running locally; push + gh pr checks 58 --watch pending.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260905-d3a1ca, pid=28887, exit=0)
No Change Request revision was published for TASK-260905-3r30t1 (handoff_unsatisfied): the board is not at to-review
spawn autonomous recovery: run RUN-260905-d3a1ca queued successor RUN-260905-77013e (attempt 1/3, model=claude-fable-5-1): producer run RUN-260905-d3a1ca remains unsatisfied: producer run RUN-260905-d3a1ca published no Change Request and reached no handoff branch while TASK-260905-3r30t1 is development: the board is not at to-review
spawn run started: [implementer] developer (claude) (run=RUN-260905-77013e)
Rework 1 cont.: hosted run 33966680785 at 5abec244 passed every job except Test (windows-latest): TestExtractPreservesExecutableBit fails there (Windows synthesizes -rw-rw-rw-). Platform-case gate itself was ok on windows. Fix a46abc80: windows skip with the existing platform-control reason + ledger row (linux,darwin required; windows tolerated). Pushed; watching CI.
Rework 1 ready for review. Head a46abc80 (signed, on top of 5abec244 and 5beced46, no rewrite), pushed to PR #58; hosted run 33968713699 fully green (Test/Race on ubuntu, macos, windows; gates). 5abec244 fixed the platform-case gate findings; a46abc80 fixed TestExtractPreservesExecutableBit on Windows (no execute bit there; platform-control skip + ledger row). One macOS Race attempt flaked in internal/install TestStrictRegistryPolicyFailsUnknown (clock-skew check, unrelated); rerun passed. Report: TASK-260905-3r30t1_rework-report-1.md; logbook entry resource updated.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260905-77013e, pid=2907, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260905-4e318e, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260905-4e318e)
agent completed: [implementer] developer (claude) (exit=1)
spawn limit exhausted: the retry was refused before any subscription group was subtracted (reason selection_snapshot_unavailable, attempts 1, evidence RUN-260905-4e318e); provider reported: You've hit your session limit · resets 9:20pm (Asia/Tbilisi)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260905-6cdd0d, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260905-6cdd0d)
Publish-CR run: story workspace clean, no edits. Items 17-19 evidenced by TASK-260905-3r30t1_mutation-evidence.md (8 narrowing mutants in a scratch copy: 5 killed, 3 survivors with stated bounds; 9/9 AC rows driven via gitops.Extract). Curator worktree untouched.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260905-6cdd0d, pid=20168, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260905-f29486, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260905-f29486)
Review cycle 1 on a46abc80: CHANGES REQUESTED (to-dev), repeat-of: none. F1 blocking: gitops.writeBlobs deadlocks on any mid-stream refusal (oversize blob, duplicate platform path) once cat-file has >pipe-buffer output queued — Wait() before draining/killing. Reproduced 3 shapes with 20s watchdog. Everything else verified (vector, negative git-archive mutant, refusals, eol testdata, gates, signatures, CI). See TASK-260905-3r30t1_review-verdict.md.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260905-f29486, pid=26338, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260905-bea544, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260905-bea544)
Rework 2 in progress: F1 fixed (pre-pass refusals from ls-tree -l + kill/drain before Wait), O1 .git refusal, O2 closure staging rename, O3 nested exec test. Commits c9defdb8, a97265ad, bb14375a (lint fix) on feat/byte-exact-acquisition, pushed; local gates green (build/vet/gofmt/golangci-lint/race focused/all 60 packages incl cmd-curator 260s/gate-selftest/ledger). Watching PR #58 checks.

## Precondition Resources
- [producer-brief-acquisition-fix.md](file://TASK-260905-3r30t1/producer-brief-acquisition-fix.md) — Producer brief: byte-exact object-database extraction replacing git archive (review M3, environments §1.2)
- [review-brief-acq-1.md](file://TASK-260905-3r30t1/review-brief-acq-1.md) — Reviewer brief cycle 1: object-database extraction at 5beced46
- [producer-brief-acq-rework-1.md](file://TASK-260905-3r30t1/producer-brief-acq-rework-1.md) — Rework 1: satisfy the hosted platform-case gate (TestArchiveRejectsLinks required case, recognised skip class), reproduce the gate locally, push, watch CI
- [review-brief-acq-2.md](file://TASK-260905-3r30t1/review-brief-acq-2.md) — Reviewer brief: object-database extraction at a46abc80 with the platform-case gate changes; PR #58 green
- [producer-brief-acq-publish-cr.md](file://TASK-260905-3r30t1/producer-brief-acq-publish-cr.md) — No-edit run in a fresh workspace: hand off to publish a fresh empty-delta Change Request
- [producer-brief-acq-rework-2.md](file://TASK-260905-3r30t1/producer-brief-acq-rework-2.md) — Rework 2: fix the cat-file deadlock (pre-pass refusals, kill+drain before Wait), .git component refusal, closure scratch rename-into-place, nested exec test

## Outcome Resources
- [TASK-260905-3r30t1_spawn-log_-implementer--developer--claude-_RUN-260905-269e50.log](file://TASK-260905-3r30t1/TASK-260905-3r30t1_spawn-log_-implementer--developer--claude-_RUN-260905-269e50.log) — System spawn log captured by task-board
- [TASK-260905-3r30t1_drafting-report.md](file://TASK-260905-3r30t1/TASK-260905-3r30t1_drafting-report.md) — Drafting report: object-database extraction, byte-exact tests, mutation evidence, gates
- [TASK-260905-3r30t1_logbook-entry.md](file://TASK-260905-3r30t1/TASK-260905-3r30t1_logbook-entry.md) — Logbook entries kept on the board (LOGBOOK.md forbidden by brief); updated with rework 1
- [TASK-260905-3r30t1_change-request_rev1.patch](file://TASK-260905-3r30t1/TASK-260905-3r30t1_change-request_rev1.patch) — Change Request CR-TASK-260905-3r30t1-1 revision 1 candidate patch (repository_delta=empty, 0 changed paths)
- [TASK-260905-3r30t1_spawn-log_-reviewer--reviewer--claude-_RUN-260905-1fb9eb.log](file://TASK-260905-3r30t1/TASK-260905-3r30t1_spawn-log_-reviewer--reviewer--claude-_RUN-260905-1fb9eb.log) — System spawn log captured by task-board
- [TASK-260905-3r30t1_spawn-log_-implementer--developer--claude-_RUN-260905-d3a1ca.log](file://TASK-260905-3r30t1/TASK-260905-3r30t1_spawn-log_-implementer--developer--claude-_RUN-260905-d3a1ca.log) — System spawn log captured by task-board
- [TASK-260905-3r30t1_rework-report-1.md](file://TASK-260905-3r30t1/TASK-260905-3r30t1_rework-report-1.md) — Rework 1 report: platform-case gate fix (5abec244), Windows executable-bit fix (a46abc80), local gate reproduction, hosted CI green on PR #58
- [TASK-260905-3r30t1_spawn-log_-implementer--developer--claude-_RUN-260905-77013e.log](file://TASK-260905-3r30t1/TASK-260905-3r30t1_spawn-log_-implementer--developer--claude-_RUN-260905-77013e.log) — System spawn log captured by task-board
- [TASK-260905-3r30t1_change-request_rev2.patch](file://TASK-260905-3r30t1/TASK-260905-3r30t1_change-request_rev2.patch) — Change Request CR-TASK-260905-3r30t1-2 revision 2 candidate patch (repository_delta=empty, 0 changed paths)
- [TASK-260905-3r30t1_spawn-log_-implementer--developer--claude-_RUN-260905-4e318e.log](file://TASK-260905-3r30t1/TASK-260905-3r30t1_spawn-log_-implementer--developer--claude-_RUN-260905-4e318e.log) — System spawn log captured by task-board
- [TASK-260905-3r30t1_spawn-log_-implementer--developer--claude-_RUN-260905-6cdd0d.log](file://TASK-260905-3r30t1/TASK-260905-3r30t1_spawn-log_-implementer--developer--claude-_RUN-260905-6cdd0d.log) — System spawn log captured by task-board
- [TASK-260905-3r30t1_cr-publish-report.md](file://TASK-260905-3r30t1/TASK-260905-3r30t1_cr-publish-report.md)
- [TASK-260905-3r30t1_mutation-evidence.md](file://TASK-260905-3r30t1/TASK-260905-3r30t1_mutation-evidence.md) — Narrowing mutants, survivors with bounds, 9/9 AC coverage ratio
- [TASK-260905-3r30t1_change-request_rev3.patch](file://TASK-260905-3r30t1/TASK-260905-3r30t1_change-request_rev3.patch) — Change Request CR-TASK-260905-3r30t1-3 revision 3 candidate patch (repository_delta=empty, 0 changed paths)
- [TASK-260905-3r30t1_spawn-log_-reviewer--reviewer--claude-_RUN-260905-f29486.log](file://TASK-260905-3r30t1/TASK-260905-3r30t1_spawn-log_-reviewer--reviewer--claude-_RUN-260905-f29486.log) — System spawn log captured by task-board
- [TASK-260905-3r30t1_review-verdict.md](file://TASK-260905-3r30t1/TASK-260905-3r30t1_review-verdict.md) — Review verdict cycle 1 on a46abc80: changes requested (F1 cat-file deadlock on mid-stream refusal)
- [TASK-260905-3r30t1_logbook-entry-review.md](file://TASK-260905-3r30t1/TASK-260905-3r30t1_logbook-entry-review.md) — Logbook entry: cat-file StdoutPipe deadlock class
- [TASK-260905-3r30t1_spawn-log_-implementer--developer--claude-_RUN-260905-bea544.log](file://TASK-260905-3r30t1/TASK-260905-3r30t1_spawn-log_-implementer--developer--claude-_RUN-260905-bea544.log) — System spawn log captured by task-board

## Created
2026-09-05T08:20:33Z

## Last Update
2026-09-05T18:19:27Z

## Assigned To
[implementer] developer (claude)
