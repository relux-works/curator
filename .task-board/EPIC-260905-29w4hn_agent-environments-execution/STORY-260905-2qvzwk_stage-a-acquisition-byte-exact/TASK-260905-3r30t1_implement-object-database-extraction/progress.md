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
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] Gate, refusal, validation, authorization, and attestation behavior attacked, not read — positive-path-only evidence is not accepted
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

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

## Precondition Resources
- [producer-brief-acquisition-fix.md](file://TASK-260905-3r30t1/producer-brief-acquisition-fix.md) — Producer brief: byte-exact object-database extraction replacing git archive (review M3, environments §1.2)
- [review-brief-acq-1.md](file://TASK-260905-3r30t1/review-brief-acq-1.md) — Reviewer brief cycle 1: object-database extraction at 5beced46
- [producer-brief-acq-rework-1.md](file://TASK-260905-3r30t1/producer-brief-acq-rework-1.md) — Rework 1: satisfy the hosted platform-case gate (TestArchiveRejectsLinks required case, recognised skip class), reproduce the gate locally, push, watch CI

## Outcome Resources
- [TASK-260905-3r30t1_spawn-log_-implementer--developer--claude-_RUN-260905-269e50.log](file://TASK-260905-3r30t1/TASK-260905-3r30t1_spawn-log_-implementer--developer--claude-_RUN-260905-269e50.log) — System spawn log captured by task-board
- [TASK-260905-3r30t1_drafting-report.md](file://TASK-260905-3r30t1/TASK-260905-3r30t1_drafting-report.md) — Drafting report: object-database extraction, byte-exact tests, mutation evidence, gates
- [TASK-260905-3r30t1_logbook-entry.md](file://TASK-260905-3r30t1/TASK-260905-3r30t1_logbook-entry.md) — Logbook entry (board-held; brief forbids writing LOGBOOK.md)
- [TASK-260905-3r30t1_change-request_rev1.patch](file://TASK-260905-3r30t1/TASK-260905-3r30t1_change-request_rev1.patch) — Change Request CR-TASK-260905-3r30t1-1 revision 1 candidate patch (repository_delta=empty, 0 changed paths)
- [TASK-260905-3r30t1_spawn-log_-reviewer--reviewer--claude-_RUN-260905-1fb9eb.log](file://TASK-260905-3r30t1/TASK-260905-3r30t1_spawn-log_-reviewer--reviewer--claude-_RUN-260905-1fb9eb.log) — System spawn log captured by task-board
- [TASK-260905-3r30t1_spawn-log_-implementer--developer--claude-_RUN-260905-d3a1ca.log](file://TASK-260905-3r30t1/TASK-260905-3r30t1_spawn-log_-implementer--developer--claude-_RUN-260905-d3a1ca.log) — System spawn log captured by task-board
- [TASK-260905-3r30t1_rework-report-1.md](file://TASK-260905-3r30t1/TASK-260905-3r30t1_rework-report-1.md) — Rework 1 report: platform-case gate fix, local gate reproduction, CI watch (updated when CI finishes)

## Created
2026-09-05T08:20:33Z

## Last Update
2026-09-05T12:40:09Z

## Assigned To
[implementer] developer (claude)
