## Status
integrating

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(5))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] SPEC_PIN names the v1.0.0-rc.11 released revision and its comment describes that pin, with no stale clause
- [x] The three schema drivers lose their root-unset tolerance and are required on every lane
- [x] Every other ledger row that loses a deferral loses its tolerance too, enumerated in the report
- [x] suite-plan reports deferred=0 for the four formerly deferred packages against the new root
- [x] test-gate against a materialized rc.11 root is green, run sequentially
- [x] The candidate lane still fails closed on a root missing an environments family
- [x] gate-selftest.sh and ledger-consistency.sh are green
- [x] Roots materialized as plain checkouts verified against manifest.json, never git archive
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Every command, message, state, or refusal named in the AC is driven through the production entry point by a named committed test, or is declared a stated bound. Report coverage as a ratio — `n of m AC rows driven` — and name the production call site for each. Prose in place of the ratio is not evidence.
- [x] Gating, refusing, validating, authorizing, or attesting behavior covered by negative tests that fail when the gate admits what it must reject, with the production call site named
- [x] Every gate ships at least one NARROWING mutant — the gate stays present and is weakened to admit exactly one member of the class it must reject, and a named test must fail. A delete-only mutant proves only that the gate exists and is not accepted as evidence.
- [x] A gate that inspects source text is additionally attacked by a mutant that PRESERVES the searched-for token and changes behavior, and the mutant harness executes the behavioral suite, not only the static checker.
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] Gate, refusal, validation, authorization, and attestation behavior attacked, not read — positive-path-only evidence is not accepted
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260907-f6262a, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260907-f6262a)
WORKSPACE MISMATCH (needs orchestrator action). The managed Story workspace for STORY-260907-2bddfc is a curator-spec worktree (.temp/STORY-260907-2bddfc/worktree, repository_id bound to https://github.com/relux-works/curator-spec, base 87a0d006), but this task changes the curator repository (.github/workflows/ci.yml, .github/ci/*.tsv, internal/interop/environments). Those paths do not exist in curator-spec, so the change cannot be made in the managed workspace. Work is therefore in the worktree the task Scope names: /Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-pin, branch chore/promote-spec-pin, base eca87fe38957ed08f4819836cc6fe12a3efd18dd. Consequence: the handoff will snapshot the Story workspace and see repository_delta=empty; the real delta is on chore/promote-spec-pin. Base check: eca87fe3 already contains curator main a66eec88 plus the PR63/PR64 stack (main..eca87fe3 = 13 commits, eca87fe3..main = 0).
Orchestrator answer to the WORKSPACE MISMATCH note: it is known, structural, and not a blocker. Delivery in this epic is cross-repository — the managed Story workspace is a curator-spec checkout while nearly every leaf delivers into curator — so every Change Request in this epic records an empty repository delta and integration happens through the orchestrator branch, PR, comment-review and fast-forward flow rather than through worktree integrate. Three reviewers verified that independently across stages (b), (c) and the consumption leaf. Working in the worktree the task Scope names (/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-pin) is correct; reporting the mismatch rather than working around it silently is exactly right. Continue.
RESULT — ready for review. SPEC_PIN 0ed5c691 -> 87a0d006 (signed tag v1.0.0-rc.11; manifest sha256:0e195ecd... = the digest release/1.0.0-rc.9.json requires; old was 803918bf...). suite-plan against the new root: served=73 deferred=0 (was 69/4). 13 ledger rows across internal/config, internal/envfragment, internal/envmarker and internal/interop/environments lost both tolerance columns and their deferral clause; all stay required on three runners. TestTheLedgerTolerationForThisPackageIsTheDeferredRootClass rewritten+renamed to TestNoLedgerRowForThisPackageToleratesASkip (checks BOTH columns); gate-selftest.sh gains 10 assertions incl. a behavioural drive of platform-case-gate.sh. Gates all exit 0: build, vet, gofmt, golangci-lint (0 issues), gate-selftest (140/0), no-broad-suppression, ledger-consistency (235 rows), and three sequential test-gate lanes (default, -race, CI_REQUIRE_FULL_ROOT=1) each served=73 deferred=0 with ZERO root-unset skips. 4 narrowing mutants, 4 killed, 0 survivors (M2 preserves the token: 0 root-unset occurrences in the mutated row, still caught). fail-closed against the real rc.11 root: 8/8 — each of the 7 declared environments artefacts is individually load-bearing. DELIVERY: two signed commits a1290fca and 25fc913a on chore/promote-spec-pin in /Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-pin; each bisects; nothing pushed, no PR. OPEN FOR ORCHESTRATOR: (1) Story workspace is a curator-spec worktree so its delta is empty — the change is on chore/promote-spec-pin; (2) hosted ubuntu/windows measurement is yours, everything here is darwin/arm64; (3) 7 root-unset rows in skillspec/marker/moduleroots/scriptpolicy remain unreachable, recommended follow-up; (4) after this promotion the candidate lane refuses candidate_ref=87a0d006 by anti-confusion, so re-qualifying that suite needs candidate_root. Report attached as TASK-260906-284db9_drafting-report.md. No LOGBOOK.md written, per the producer brief.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260907-f6262a, pid=3959, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260907-086d22, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260907-086d22)
Review cycle 1: ACCEPT. Premise re-derived in this run, not read from the PR body: old pin manifest 803918bf, new pin manifest 0e195ecd, release/1.0.0-rc.9.json requires 0e195ecd with committed_release_pin_advanced=false, v1.0.0-rc.11 is an annotated SSH-signed tag peeling to 87a0d006. Both roots materialized as plain checkouts and verified file-by-file both directions (691/1047 declared, 0 missing, 0 mismatch, 0 undeclared); subst.txt 40 bytes unsubstituted.

13 rows lost their root-unset tolerance, in exactly the four packages that stopped being deferred (environments 7, config 3, envfragment 2, envmarker 1). No other column changed in the 235-row ledger. suite-plan partitions compared by name against both roots on all three GOOS: exactly those four packages moved. The review brief says nineteen; the tree says 13 and the producer report is right.

The removal is load-bearing, not cosmetic: with SPEC_PIN reverted, the DEFAULT lane test-gate exits 1 with 26 named FAILs and all 13 cases FATAL-not-tolerated while go test exits 0; the same stream against the base ledger gives all 13 tolerated-by-ledger. Four narrowing mutants written and run in this review, all killed, including a token-preserving one (root-unset count 11 before and after) driven through the behavioural suite on a REAL go test -json stream, flipping the production gate verdict. Vacuity attack fires the guard instead of passing silently. PR #63 fail-closed proved per artefact: all seven declared environments artefacts individually fatal under CI_REQUIRE_FULL_ROOT=1, named by package and artefact.

Hosted run 34163836478 at 04550e28: 11 pass, 1 correctly skipping. All five uploaded evidence artifacts read: deferred=0, no deferred go test stage, 14/14 promoted cases observed PASSING, 0 root-unset skips, 0 tolerated rows in those packages, gate ok on Test and Race across all three runners. SERVED, not merely green.

Local gates all 0: build, vet, gofmt, golangci-lint (0 issues), ledger-consistency (235 rows), gate-selftest (140/0), no-broad-suppression, test-gate default lane (served=73, one invocation, 20 skips, 0 root-unset). Branch bisects re-verified at the post-rebase SHAs (494198f9: 130/0).

Non-blocking findings in the verdict artifact: F1 platform-exclusions.tsv:11 claims the committed pin publishes no qualification vector, which is false for both the old and the new pin -- pre-existing, file untouched by this branch, behaviourally inert, follow-up not rework. F2 contract_test.go:514 logs "none tolerating a skip" unconditionally so it prints beside its own failure. F3 the brief row count. F4 the seven remaining root-unset rows: leaving them is the right scoping call (they were already unreachable under the previous pin) and the file explains it adequately.

EMPTY repository_delta: a workspace-provisioning artifact, not a producer failure. The Story workspace is a curator-spec checkout; this leaf changes curator paths that do not exist there. The real delta is chore/promote-spec-pin at 04550e28, PR #66, two signed commits with human identity. Integrating THIS revision lands nothing -- PR #66 still needs its own landing.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260907-086d22, pid=94090, exit=0)

## Precondition Resources
- [producer-brief-spec-pin.md](file://TASK-260906-284db9/producer-brief-spec-pin.md) — Producer brief: move SPEC_PIN to the v1.0.0-rc.11 released revision and retire the deferrals it existed for
- [review-brief-spec-pin-1.md](file://TASK-260906-284db9/review-brief-spec-pin-1.md) — Review brief cycle 1: verify the premise, the removed tolerances, and that the families are served not deferred

## Outcome Resources
- [TASK-260906-284db9_spawn-log_-implementer--developer--claude-_RUN-260907-f6262a.log](file://TASK-260906-284db9/TASK-260906-284db9_spawn-log_-implementer--developer--claude-_RUN-260907-f6262a.log) — System spawn log captured by task-board
- [TASK-260906-284db9_drafting-report.md](file://TASK-260906-284db9/TASK-260906-284db9_drafting-report.md) — SPEC_PIN promotion to v1.0.0-rc.11 (87a0d006): pin+manifest evidence, suite-plan before/after, the 13 ledger rows that lost their root-unset tolerance, gate table, 4 killed mutants, candidate fail-closed against the real root, workspace mismatch
- [TASK-260906-284db9_change-request_rev1.patch](file://TASK-260906-284db9/TASK-260906-284db9_change-request_rev1.patch) — Change Request CR-TASK-260906-284db9-1 revision 1 candidate patch (repository_delta=empty, 0 changed paths)
- [TASK-260906-284db9_spawn-log_-reviewer--reviewer--claude-_RUN-260907-086d22.log](file://TASK-260906-284db9/TASK-260906-284db9_spawn-log_-reviewer--reviewer--claude-_RUN-260907-086d22.log) — System spawn log captured by task-board
- [TASK-260906-284db9_review-verdict-rev1.md](file://TASK-260906-284db9/TASK-260906-284db9_review-verdict-rev1.md) — Review cycle 1 verdict: ACCEPT. Premise re-derived independently (three manifest digests, signed tag, release record); 13 tolerances removed = exactly the four newly-served packages; four narrowing mutants incl. a token-preserving one killed through the behavioural suite on a real go test -json stream; PR #63 fail-closed proved per artefact; all five hosted lanes SERVED the families with 14/14 promoted cases observed passing; empty repository_delta explained as a curator-spec workspace artifact

## Created
2026-09-06T08:39:39Z

## Last Update
2026-09-07T22:40:35Z

## Assigned To
[reviewer] reviewer (claude)
