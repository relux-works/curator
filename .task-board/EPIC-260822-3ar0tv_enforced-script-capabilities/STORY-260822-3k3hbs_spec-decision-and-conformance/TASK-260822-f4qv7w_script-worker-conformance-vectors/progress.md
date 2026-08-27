## Status
done

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
- [x] Spec CI gates green on the merged story branch
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
GREEN CANDIDATE EVIDENCE for your blocked packet: candidate-conformance run 32651139699 — success on all three OSes for the superseding rc.9 identity edd0721 / manifest 803918bf (your bytes dd9c9fc are an ancestor input of that candidate; the rc.8 pin invariant is resolved by the rc.9 migration inside the candidate). This satisfies the exact external input your packet named. Ready to return to review with this evidence.
REFRESHED GREEN CANDIDATE EVIDENCE (supersedes the edd0721 / 32651139699 note above, and the e66cb72 / 32654422338 identity). Current qualified candidate is 6001dc33281b94a4ec7442ab15278550dd0f51d9 on candidate/schema-8-rc.9, protocol 1.0.0-rc.9, manifest sha256:803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403, 692 files; the suite digest is identical to edd0721 and e66cb72 because every commit after edd0721 is prose outside conformance/v1. Your bytes remain an ancestor input of the candidate. curator-spec Specification CI 32659168954 is success on exactly 6001dc3 across six jobs; curator candidate-conformance 32659157687 is success on all fourteen jobs first attempt, with Candidate suite green on ubuntu/macos/windows and the job logs showing CANDIDATE_REF 6001dc33..., revision accepted immutable, manifest digest matches the supplied expectation, and SPEC_PIN unchanged at 00b1688. rc.8 stays byte-frozen at 293f101d...e31ede, byte-identical to curator-spec origin/main. This satisfies the exact external input your packet named; ready to return to review. Full evidence: TASK-260822-c0rxj7_results.md and TASK-260822-c0rxj7_6001dc3-green-matrix.md.
Blocker resolved: TASK-260822-c0rxj7 is done. FINAL candidate identity: 6001dc3 (supersedes e66cb72 and edd0721; identity and green 3-OS matrix in the c0rxj7 artifacts _candidate-suite-identity-6001dc3.txt and _6001dc3-green-matrix.md). Your committed bytes dd9c9fc are an ancestor input. Fresh review cycle may accept on this evidence.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260823-da95fc, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260823-da95fc)
REVIEW VERDICT RUN-260823-da95fc: ACCEPTED. Independent verification, not a re-read of prior evidence. (1) Lineage: origin/spec/script-worker-v1-normative head == dd9c9fc; dd9c9fc is a verified ancestor of qualified candidate 6001dc3; the only diff to the script vector between them is protocol_version rc.8->rc.9 (c0rxj7 scope); rc.8 blob at 6001dc3 == origin/main. (2) Content: every vector identifier cross-checked against protocol/core.md, profiles/manager.md, SECURITY.md on the same branch; the native-control inventory matches the core.md table cell-for-cell; Linux active-process-count-limit is host-conditional / delegated-cgroup-v2-pids.max with both required Linux lane preflight cases, so the RLIMIT_NPROC rework regression is closed. All five AC behaviours carry positive and negative cases (opt-in 6, derivation 4, preflight 5, evidence closure 14, audit labels 4). Both task-note hooks closed: mixed_build_cases schema8-script-worker row and the install-marker-v4 golden; marker-v4 prose binding lands at e66cb72 in the candidate, so the fixture is not orphaned. (3) Gates re-run by me in a fresh isolated worktree at 6001dc3: validate.py 53 schemas / 691 vectors exit 0, 98 python tests OK, go test ./tools/... ok, gofmt clean. (4) Determinism: two regeneration passes at 6001dc3 byte-identical over 671 files with git diff --exit-code clean vs committed; two passes at dd9c9fc itself byte-identical over 633 files with conformance/v1 clean and only release/1.0.0-rc.8.json dirty, mechanically reproducing the stop-the-line. (5) Teeth check: mutating the Linux cell to available/RLIMIT_NPROC fails validate.py on digest mismatch and fails TestScriptWorkerConformanceContract; the generator restores byte-exactly. (6) Delivery gate: branch dd9c9fc CI is red and structurally un-greenable without rewriting immutable rc.8 - the rejected forced fit. AC satisfied via the candidate: curator-spec Specification CI 32659168954 success on exactly 6001dc3 across six jobs, curator CI 32659157687 success across all fourteen jobs with Candidate suite green on ubuntu/macos/windows. Both execute this task validator, Go contract test, and regenerate-check over the reviewed bytes on three platforms. Non-blocking follow-ups recorded in the verdict artifact: script_execution_policy_unsupported has no conformance coverage anywhere, and preflight_cases / capability_evidence_cases lack exact-set assertions in the validator. Reviewer archetype: no commit_ack supplied; reviewed scope is already committed and pushed. Evidence: TASK-260822-f4qv7w_review-verdict_RUN-260823-da95fc.md, _reviewer-gates_, _reviewer-regen-pass1/2_.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260823-da95fc, pid=56117, exit=0)

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
- [TASK-260822-f4qv7w_spawn-log_-reviewer--reviewer--claude-_RUN-260823-da95fc.log](file://TASK-260822-f4qv7w/TASK-260822-f4qv7w_spawn-log_-reviewer--reviewer--claude-_RUN-260823-da95fc.log) — System spawn log captured by task-board
- [TASK-260822-f4qv7w_review-verdict_RUN-260823-da95fc.md](file://TASK-260822-f4qv7w/TASK-260822-f4qv7w_review-verdict_RUN-260823-da95fc.md) — Reviewer accepted verdict: independent content/prose cross-check, isolated gate runs, double-regeneration determinism at both dd9c9fc and candidate 6001dc3, teeth check, and CI green evidence
- [TASK-260822-f4qv7w_reviewer-gates_RUN-260823-da95fc.log](file://TASK-260822-f4qv7w/TASK-260822-f4qv7w_reviewer-gates_RUN-260823-da95fc.log) — Independent isolated-worktree gate log at candidate 6001dc3: validate.py, 98 python tests, go test, gofmt
- [TASK-260822-f4qv7w_reviewer-regen-pass1_RUN-260823-da95fc.txt](file://TASK-260822-f4qv7w/TASK-260822-f4qv7w_reviewer-regen-pass1_RUN-260823-da95fc.txt) — Independent regeneration pass 1 SHA-256 inventory at 6001dc3 (671 generated JSON files)
- [TASK-260822-f4qv7w_reviewer-regen-pass2_RUN-260823-da95fc.txt](file://TASK-260822-f4qv7w/TASK-260822-f4qv7w_reviewer-regen-pass2_RUN-260823-da95fc.txt) — Independent regeneration pass 2 SHA-256 inventory at 6001dc3, byte-identical to pass 1

## Created
2026-08-22T16:00:35Z

## Last Update
2026-08-23T19:42:05Z

## Assigned To
[reviewer] reviewer (claude)
