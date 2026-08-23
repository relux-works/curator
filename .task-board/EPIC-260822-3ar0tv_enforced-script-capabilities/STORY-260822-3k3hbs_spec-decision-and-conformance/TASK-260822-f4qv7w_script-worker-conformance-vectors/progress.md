## Status
blocked

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(5))

## Blocked By
- TASK-260822-1f533i
- TASK-260822-3fkfmf
- TASK-260822-1mwy10
- TASK-260822-c0rxj7

## Blocks
- (none)

## Checklist
- [x] script-host-execution-policy.json active-process-count-limit Linux entry is host-conditional (delegated cgroup v2 pids.max), not available/RLIMIT_NPROC — per core.md 78d544d — plus the probe-unavailable/evidence-unavailable/invocation-succeeds case on the Linux lane
- [x] Branches spec/sw-core-prose, spec/sw-manager-security, spec/sw-schema merged into spec/script-worker-v1-normative on top of current origin/main
- [x] Positive and negative vectors cover opt-in parsing, deny-by-default derivation, preflight rejection, evidence-record closure, and legacy declared-only labeling
- [x] Vector regeneration run twice with byte-identical second run
- [ ] Spec CI gates green on the merged story branch
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [ ] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
SCHEMA INPUT READY from TASK-260822-1mwy10: manifest schema 8 is on branch spec/sw-schema, commit ebfed81. Structural schema cases already generated: agent-skill-v8 and csk-skill-v8 get 51 cases each (19 script-worker branches — opt-in, mixed enforcement, node and python3 interpreters, unix-only and windows-only paths, and rejections for missing interpreter, interpreter without policy, unknown interpreter, successor policy script-worker-v2, hardened policy, compiled manager-worker-v1 policy, null and "none" opt-out spellings, policy on system and build commands, top-level placement, and a pathless enforced command), install-marker-v4 gets 27, and each of schemas 1 through 7 gets four invalid-v8-* rejection cases under both filenames.

TWO HOOKS LEFT FOR THIS TASK: (1) mixed_build_cases in tools/generate-vectors/main.go maps manifest_schema to marker_version and has no schema8 row yet — adding one records the schema-8 to marker-v4 binding behaviourally; (2) install-marker-v4 has structural schema cases but no expected marker fixture under conformance/v1/expected/. Behavioural vectors for deny-by-default capability derivation, mandatory-control preflight rejection, evidence-record closure, and legacy declared-only labeling remain entirely in this task scope. Double regeneration is already proven clean on the schema branch, so a rebase there should stay deterministic.
Work in the existing worktree /Users/iv/Developer/ReluxWorks/curator-spec/.temp/STORY-260822-3k3hbs/normative-worktree (branch spec/script-worker-v1-normative, based on a2d44eb). First: git fetch origin, merge origin/main (now be7861c), then merge the three reviewed branches spec/sw-core-prose (78d544d tip), spec/sw-manager-security, spec/sw-schema (ebfed81) — they touch disjoint files, conflicts should be trivial; push the merged branch. Then write the conformance vectors on top and prove double regeneration. Predecessor results live as outcome resources on TASK-260822-1f533i, TASK-260822-3fkfmf, TASK-260822-1mwy10. No PR — the landing task TASK-260822-c0rxj7 does that.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-0c1746, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-0c1746)
agent completed: [implementer] developer (claude) (exit=1)
spawn run completed: claude (run=RUN-260822-0c1746, pid=84031, exit=1)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260823-e8ac01, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260823-e8ac01)
Developer handoff: implemented script-worker-v1 conformance vectors, schema8 marker-v4 binding/golden, Go tests, and Python mutation validators in the normative worktree. Local CI-equivalent gates are green; regeneration runs 3 and 4 are byte-identical. Changes are intentionally unstaged/uncommitted under repository policy, so remote multi-OS CI awaits the authorized reviewer/landing commit. Evidence: TASK-260822-f4qv7w_results.md and TASK-260822-f4qv7w_make-validate.log.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260823-e8ac01, pid=3322, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (codex) (run=RUN-260823-f14b1a, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260823-f14b1a)
REVIEW VERDICT RUN-260823-f14b1a: changes requested. Content and architecture accepted; independent make validate green (52 schemas, 658 vectors, 95 Python tests, Go tests), two regenerations byte-identical, gofmt clean. Delivery AC is unmet: task bytes are uncommitted, origin/spec/script-worker-v1-normative is absent, and no remote branch CI ran on these bytes. Commit-owning mover must commit only reviewed scope excluding .temp and __pycache__, push, record green required CI, reconcile the historical rc.8 generated delta through TASK-260822-c0rxj7 rc.9 path, then return for review. Evidence: TASK-260822-f4qv7w_review-verdict_RUN-260823-f14b1a.md plus reviewer validation/checksum outcomes.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260823-f14b1a, pid=39350, exit=0)
DELIVERY REWORK ONLY (content accepted, reviewer regenerated checksums both passes): in worktree /Users/iv/Developer/ReluxWorks/curator-spec/.temp/STORY-260822-3k3hbs/normative-worktree the branch spec/script-worker-v1-normative sits at merge commit a690d63 with the whole vector delta unstaged. Do: (1) commit ONLY the reviewed scope — exclude .temp/ and tools/__pycache__/ and DO NOT commit the regenerated release/1.0.0-rc.8.json rewrite (rc.8 is immutable history; the rc.9 candidate task TASK-260822-c0rxj7 owns release metadata migration — leave that file out or restore it); (2) push -u origin spec/script-worker-v1-normative; (3) record the branch CI run and green required gates as evidence on this task; (4) handoff to to-review for a fresh reviewer cycle. No AI attribution in the commit.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260823-fddaeb, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260823-fddaeb)
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260823-fddaeb, pid=54631, exit=0)
Unblocked by immutable shared Schema 8 rc.9 candidate 859727b103ed175ff214cbb64641f4686d8c6a68; manifest sha256:782d6868f6d9725f7bf38d3fb1944f307f2d5d9c060b8816b0f55a5c2e97f11f. Specification CI 32633567102 is green across Linux/macOS/Windows. rc.8 remains byte-frozen; candidate branch is candidate/schema-8-rc.9.
Green-candidate evidence is still unavailable: exact rc.9 rerun 32638424105 passed candidate macOS but failed Ubuntu and Windows. Keep blocked/review release gate closed. Failure matrix and fix routing are recorded on TASK-260823-1l1p8q and TASK-260822-c0rxj7; same immutable SHA/digest must be rerun after Curator fixes land.

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260822-f4qv7w_spawn-log_-implementer--developer--claude-_RUN-260822-0c1746.log](file://TASK-260822-f4qv7w/TASK-260822-f4qv7w_spawn-log_-implementer--developer--claude-_RUN-260822-0c1746.log) — System spawn log captured by task-board
- [TASK-260822-f4qv7w_spawn-log_-implementer--developer--codex-_RUN-260823-e8ac01.log](file://TASK-260822-f4qv7w/TASK-260822-f4qv7w_spawn-log_-implementer--developer--codex-_RUN-260823-e8ac01.log) — System spawn log captured by task-board
- [TASK-260822-f4qv7w_results.md](file://TASK-260822-f4qv7w/TASK-260822-f4qv7w_results.md) — Developer handoff with scope, changed surfaces, deterministic regeneration proof, and gate exit codes
- [TASK-260822-f4qv7w_make-validate.log](file://TASK-260822-f4qv7w/TASK-260822-f4qv7w_make-validate.log) — Green integrated validation log: schemas, Python tests, and Go tests
- [TASK-260822-f4qv7w_spawn-log_-reviewer--reviewer--codex-_RUN-260823-f14b1a.log](file://TASK-260822-f4qv7w/TASK-260822-f4qv7w_spawn-log_-reviewer--reviewer--codex-_RUN-260823-f14b1a.log) — System spawn log captured by task-board
- [TASK-260822-f4qv7w_review-verdict_RUN-260823-f14b1a.md](file://TASK-260822-f4qv7w/TASK-260822-f4qv7w_review-verdict_RUN-260823-f14b1a.md) — Reviewer changes-requested verdict with independent content, ancestry, validation, determinism, delivery-gate, and rc.9 handoff evidence
- [TASK-260822-f4qv7w_reviewer-make-validate_RUN-260823-f14b1a.log](file://TASK-260822-f4qv7w/TASK-260822-f4qv7w_reviewer-make-validate_RUN-260823-f14b1a.log) — Independent isolated-copy make validate log
- [TASK-260822-f4qv7w_reviewer-generated-checksums-pass1_RUN-260823-f14b1a.txt](file://TASK-260822-f4qv7w/TASK-260822-f4qv7w_reviewer-generated-checksums-pass1_RUN-260823-f14b1a.txt) — Independent first regeneration generated-tree SHA-256 inventory
- [TASK-260822-f4qv7w_reviewer-generated-checksums-pass2_RUN-260823-f14b1a.txt](file://TASK-260822-f4qv7w/TASK-260822-f4qv7w_reviewer-generated-checksums-pass2_RUN-260823-f14b1a.txt) — Independent second regeneration generated-tree SHA-256 inventory, byte-identical to pass 1
- [TASK-260822-f4qv7w_spawn-log_-implementer--developer--codex-_RUN-260823-fddaeb.log](file://TASK-260822-f4qv7w/TASK-260822-f4qv7w_spawn-log_-implementer--developer--codex-_RUN-260823-fddaeb.log) — System spawn log captured by task-board
- [TASK-260822-f4qv7w_delivery-rework.md](file://TASK-260822-f4qv7w/TASK-260822-f4qv7w_delivery-rework.md) — Signed commit/push evidence, standalone gate exit codes, remote CI verdict, and rc.9 ownership blocker packet
- [TASK-260822-f4qv7w_github-spec-ci-32632173590.log](file://TASK-260822-f4qv7w/TASK-260822-f4qv7w_github-spec-ci-32632173590.log) — Specification CI workflow_dispatch watch log for exact pushed commit dd9c9fc

## Created
2026-08-22T16:00:35Z

## Last Update
2026-08-23T12:41:41Z

## Assigned To
[implementer] developer (codex)
