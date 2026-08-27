## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(3))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] Status banding accepts marker v4 for schema-8 compiled commands; regression test added
- [x] Verified live: curator global status shows installed for project-management at 3958813
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260824-490d6e, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260824-490d6e)
Fix landed on PR #41 (https://github.com/relux-works/curator/pull/41), commit 564371c signed.

WORKSPACE: the STORY-260822-2lvw0e worktree was at 903af23, which predates the whole schema-8/marker-v4 subsystem (PolicySchemaVersion does not exist there) so the bug was not reproducible on it. origin/main was at 272b203. Work was done in a task-scoped worktree .temp/BUG-260825-1l1st9/worktree forked from origin/main; the story branch was left untouched.

ROOT CAUSE: three readers each carried their own inequality on the marker schema instead of asking the marker package. classifySkillBuilds (cmd/curator/builds.go:365) admitted schema 2 only -> the reported escape, and it refused marker v3 too. markerRefusal (builds.go:545) admitted schemas 1 and 2, so schemas 3 and 4 were reported as documents from a newer manager. marks.absorb (internal/scopes/gc.go:213) admitted schemas 2 and 3, so a marker v4 contributed NO live build reference and a GC pass could delete protected cache entries a live schema-8 installation was still running from - a second, unreported defect of the same class. marker.Current was already correct because it asks buildBearingSchema.

FIX: exported that predicate and the readable-schema predicate as marker.BuildBearingSchema / marker.SupportedSchema; all three readers now ask them. Added marker.NewestSchemaVersion, pinned to be the real maximum of the readable band. The remedy string no longer names a schema to record - that was the self-contradiction the report flagged.

TWO EXISTING TESTS ASSERTED THE DEFECT and were corrected, not deleted: TestMarkerRefusalSeparatesUnsupportedFromInvalid claimed schema 3 was unreadable, and the status matrix case named "marker schema cannot be read by this manager" pinned marker.SchemaVersion + 1, so it silently stopped testing the unsupported band the moment a newer schema became readable.

EVIDENCE: 7/7 mutants caught, both directions on every gate (see BUG-260825-1l1st9_mutation-evidence.txt). Local gates on darwin/arm64 against the materialized rc.9 pin (SPEC_PIN 0ed5c691, manifest sha256:803918bf..b44403): build 0, vet 0, gofmt clean, golangci-lint 0 issues, go test ./internal/... 0 (42 pkgs), go test ./cmd/curator/ 0, ledger-consistency 0, no-broad-suppression 0, gate-selftest 0. Windows/Linux/race/interop/naming lanes run on the PR.

LIVE AC MET without reinstalling: global status now reports project-management up-to-date with task-board, task-board-tui and tb-sessiond all state=current, on the same marker v4 at commit 3958813 that produced the escape.

CONFORMANCE GAP recorded for the suite owner in BUG-260825-1l1st9_conformance-gap.md: the suite proves the write side (expected/install-marker-v4.json) and the currentness side (manager-lifecycle.json status_cases) separately and never crosses them with the schema as the variable. compiled-installation-current lists marker-schema among validated steps but does not parameterise the schema; compiled-currentness-failure-matrix has 14 independent_conditions and none is a marker-schema condition; the go-build-skill fixture carries no schema_version field; gc_cases has no marker-schema-parameterised liveness vector.
agent completed: [implementer] developer (claude) (exit=124)
spawn run completed: claude (run=RUN-260824-490d6e, pid=77943, exit=124)
spawn run RUN-260824-490d6e failed; operator action required; failure: run exceeded --timeout 50m0s and was terminated by the launcher
Landed: PR 41 merged as 5cbd1b8 with every lane green pre-merge. Live verification on the reporting machine after rebuilding ~/.local/bin/curator (v0.14.0-rc.3-48-g5cbd1b8): curator global status now reports project-management up-to-date, all three compiled commands state=current on marker v4 at 3958813. Conformance-suite gap recorded in the conformance-gap resource for the suite owner. REVIEWER SCOPE: the PR 41 delta plus the regression test; targeted tests, cite lanes.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260825-723452, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260825-723452)
REVIEW VERDICT: ACCEPTED (run RUN-260825-723452, not goal-bound). Evidence: BUG-260825-1l1st9_review-verdict.md.

Scope: PR 41 head eb58395, merged 5cbd1b8 on main; merge is a clean no-op (tree(5cbd1b8)==tree(eb58395)==d298308, first parent 272b203). Reviewed in a detached worktree .temp/BUG-260825-1l1st9/review-wt at 5cbd1b8; the story worktree at 903af23 predates PolicySchemaVersion and cannot host this review. Story branch untouched.

AC verified independently, all three clauses. (1) Live on this machine without reinstalling: ~/.curator/global/skills/project-management/.csk-install.json is still schema_version 4 / skill_schema_version 8 / commit 3958813, installed_at 2026-08-24T23:20:46Z, mtime Aug 25 03:20:47 - it predates the 00:36:15Z merge and was not rewritten, so the flip came from the reader. curator v0.14.0-rc.3-48-g5cbd1b8 global status reports project-management up-to-date with task-board, task-board-tui and tb-sessiond all state=current. (2) The regression drives the real production entry point: statusReport (main.go:768) is the call site behind both curator status (main.go:677) and reportGlobalStatus (main.go:1219); all three fixed readers are production-reachable (classifySkillBuilds via main.go:807, markerRefusal via main.go:895, marks.absorb via markScopes<-scopes.Collect<-main.go:1713). (3) Every PR lane SUCCESS pre-merge; Candidate suite is SKIPPED by design (workflow_dispatch + candidate_ref/root, ci.yml:303-305). 564371c signed, G.

Mutation verification rerun by the reviewer, not accepted from the artifact: 5 mutants applied to the merged tree and reverted, all caught, both directions. Narrowing classifySkillBuilds back to the single written schema reproduces the reported string verbatim for schemas 3 and 4; reverting absorb to the {2,3} band, widening BuildBearingSchema, staling NewestSchemaVersion, and widening SupportedSchema each turn a distinct test red. Tree clean against 5cbd1b8 after every revert.

Completeness: grep over the merged tree finds exactly one install-marker-schema reference outside internal/marker - the operator message at builds.go:555. No hand-rolled inequality survives. The unreported GC defect is real and correctly rated: marked.builds feeds Referenced into cache.Sweep (gc.go:114), so an unmarked marker-v4 key would have been swept from under a live schema-8 install. No test function was deleted; the two that asserted the defect were corrected, and the matrix negative is now anchored to NewestSchemaVersion+1 (readable band edge) instead of SchemaVersion+1 (written+1), which is the only anchor that survives a widening read band.

Conformance-gap artifact spot-checked against curator-spec at the committed SPEC_PIN 0ed5c691: two status_cases with those exact names, compiled-installation-current lists marker-schema in validated without parameterising it, the failure matrix has exactly 14 independent_conditions and none is marker-schema, fixtures/go-build-skill/.csk-install.json has the single key package_marker, and neither gc_case is schema-parameterised. Every claim holds.

Reviewer gates rerun locally on darwin/arm64 at 5cbd1b8: build 0, vet 0, gofmt clean, golangci-lint 0 issues, go test ./internal/{marker,scopes,interop,registry,ui} ok, the three targeted cmd/curator regressions PASS, TestCompiledProjectStatusRepairRollbackRecovery PASS (193.9s, 15 currentness sub-cases including both new schema cases). NOT rerun locally and accepted from CI: internal/install and the rest of internal, plus the Windows/Linux/race/interop/naming lanes - a local go test ./internal/... exceeded the bounded-call budget and was stopped rather than backgrounded, so no result is claimed for it.

Two non-blocking follow-ups, neither grounds to withhold acceptance: (a) the conformance gap is recorded in the board artifact and LOGBOOK entry 0417 but curator-spec has no .task-board, so nothing tracks it on the suite owner side - the three requested vectors need carrying into that repo tracker; (b) pre-existing and out of scope, marker.Write maps SkillSchemaVersion>=8 to PolicySchemaVersion while validMarker requires ==8 for schema 4, so a future skill schema 9 would fail at write rather than being refused earlier with a reason.

HANDOFF: reviewer-archetype run, no commit_ack supplied. PR scope is already committed and merged as 5cbd1b8 on main. The commit-owning mover makes the final done transition with commit_ack=scope_committed, citing BUG-260825-1l1st9_review-verdict.md.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260825-723452, pid=90189, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [BUG-260825-1l1st9_spawn-log_-implementer--developer--claude-_RUN-260824-490d6e.log](file://BUG-260825-1l1st9/BUG-260825-1l1st9_spawn-log_-implementer--developer--claude-_RUN-260824-490d6e.log) — System spawn log captured by task-board
- [BUG-260825-1l1st9_results.md](file://BUG-260825-1l1st9/BUG-260825-1l1st9_results.md) — results
- [BUG-260825-1l1st9_conformance-gap.md](file://BUG-260825-1l1st9/BUG-260825-1l1st9_conformance-gap.md) — conformance-gap
- [BUG-260825-1l1st9_mutation-evidence.txt](file://BUG-260825-1l1st9/BUG-260825-1l1st9_mutation-evidence.txt) — Mutation evidence: 7/7 mutants caught
- [BUG-260825-1l1st9_spawn-log_-reviewer--reviewer--claude-_RUN-260825-723452.log](file://BUG-260825-1l1st9/BUG-260825-1l1st9_spawn-log_-reviewer--reviewer--claude-_RUN-260825-723452.log) — System spawn log captured by task-board
- [BUG-260825-1l1st9_review-verdict.md](file://BUG-260825-1l1st9/BUG-260825-1l1st9_review-verdict.md)

## Created
2026-08-24T23:21:56Z

## Last Update
2026-08-25T01:03:57Z

## Assigned To
[reviewer] reviewer (claude)
