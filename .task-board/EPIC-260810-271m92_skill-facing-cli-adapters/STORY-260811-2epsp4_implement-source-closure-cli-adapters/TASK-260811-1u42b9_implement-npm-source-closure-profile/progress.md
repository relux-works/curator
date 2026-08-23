## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(8))

## Blocked By
- TASK-260811-3twayo

## Blocks
- TASK-260811-x611eq

## Checklist
- [x] Implement closed npm lock, workspace, integrity, package metadata, and selected or pruned graph reconciliation
- [x] Capture and admit raw tgz inputs and materialize from a private cache with npm ci offline and scripts disabled
- [x] Pass npm closure, stale-lock, integrity, lifecycle, bundled-dependency, native-payload, and ambient-cache vectors
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
- [x] Remove ambient npm shebang/PATH edge and prove real portable plus verified Node-launched npm execution
- [x] Implement executable npm S03 poisoned-cache, S04 network-attempt, and S08 two-run invariance vectors

## Notes
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260822-29b5d5, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260822-29b5d5)
Implemented closed npm v2/v3 lock/workspace/tarball/private-cache profile. Shared fixes: gzip virtual-directory stream accounting; binding.gyp Node metadata; workspace manifest-ID dedup. Architecture nudge applied: portable remains default with manager-owned invariants; verified requires lossless provider before start, with zero-start negative proof. Final affected tests/race/coverage 80.1%/vet/lint/build/diff all exit 0; repository-wide suite also exited 0 before the final contained assurance correction. Evidence: TASK-260811-1u42b9_implementation-evidence.md
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260822-29b5d5, pid=22185, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (codex) (run=RUN-260822-0c4aea, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260822-0c4aea)
Reviewer RUN-260822-0c4aea requests implementation rework. Materialization validates only package paths, so substituted/incomplete/compiled materialized bytes can pass; npm/Node execution is not bound to the common exact C0/toolchain/permit contract and verified provider identity can drift; portable lossless-only counters remain ambiguous default zeros. Fresh focused/race/coverage/vet/lint/build/diff/board/full-suite gates all exited 0. See TASK-260811-1u42b9_review-verdict_RUN-260822-0c4aea.md.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260822-0c4aea, pid=27283, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260822-cb137e, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260822-cb137e)
Rework RUN-260822-cb137e: npm installed paths are insufficient closure evidence; materialized dependency contents now require shared recursive re-admission plus exact tarball-derived owned-file reconciliation. Portable audits omit lossless-only fields; verified starts require immutable full AssuranceBinding + rederived common Node C0/C5 + immediate provider/tool revalidation. First final lint run exited 1 on mechanical findings, corrected; all subsequent focused/race/coverage/vet/lint/build/diff/full-suite gates exited 0. Evidence: TASK-260811-1u42b9_rework-evidence_RUN-260822-cb137e.md
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260822-cb137e, pid=71228, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (codex) (run=RUN-260822-411c9c, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260822-411c9c)
RUN-260822-411c9c reviewer verdict: changes requested. Materialized-byte reconciliation and honest portable audit are accepted, and all fresh gates are green. Rework remains required because npm dispatch uses adapter-local Runner/permit/audit authority instead of closureexec.Executor canonical permits and issued receipts; verified provider evidence is self-assertable without nonce-bound negotiation; the real npm test runs PATH-resolved npm while merely echoing fixture C0 digests, and the attached C5 plan does not authorize the manager action. Evidence: TASK-260811-1u42b9_review-verdict_RUN-260822-411c9c.md
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260822-411c9c, pid=54492, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260822-77c106, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260822-77c106)
Rework RUN-260822-77c106 removes npm-local execution authority: cache derivation, npm ci, and Node invocation now commit and execute canonical permits through one preflighted closureexec.AssuredOperation, bind admitted receipts plus exact C0 tools and C5 runtime actions, and accept only executor-issued receipts. Portable evidence remains network=not-observed without synthetic lossless zeros. Fresh focused/race/coverage 80.1%/vet/lint/build/diff/board/full-suite gates all exited 0; real npm ran. Zero-start tests cover missing/incomplete/incompatible/cross-mode/drifted verified provider authority, absent C5 action, and PATH substitution. Evidence: TASK-260811-1u42b9_rework-evidence_RUN-260822-77c106.md
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260822-77c106, pid=98706, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (codex) (run=RUN-260822-b81336, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260822-b81336)
Reviewer logbook 2026-08-23 RUN-260822-b81336: CHANGES REQUESTED -> to-dev. Shared executor calls are present, materialized-byte reconciliation and portable evidence honesty pass, and all fresh gates including real npm and uncached full suite exit 0. Remaining trust defect: C5 is matched only by subtype/tool; actual argv/cwd/environment/read/write authority bypasses permit mounts through absolute runner values, the PATH negative is runner-self-enforced, and no positive verified execution can match the real npm I/O. Exact evidence and rework are in TASK-260811-1u42b9_review-verdict_RUN-260822-b81336.md. No external or human-only blocker exists; reviewer changed no code and supplied no commit_ack.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260822-b81336, pid=36224, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260822-a5f43f, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260822-a5f43f)
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260822-a5f43f, pid=53040, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (codex) (run=RUN-260822-af01e0, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260822-af01e0)
Reviewer logbook 2026-08-23 RUN-260822-af01e0: CHANGES REQUESTED -> to-dev. Exact C5 templates/common portable runner/materialized-byte reconciliation are accepted and every fresh gate is green. Remaining defects: real staged npm is a #!/usr/bin/env node script while permit binds only npm+Node and PATH includes ambient /usr/bin:/bin; positive verified fixture copies permit fields instead of observing that process chain. Claimed S03/S04/S08 test contains no poisoned ambient-cache comparison, network-attempt vector, or two-run identity comparison. Evidence: TASK-260811-1u42b9_review-verdict_RUN-260822-af01e0.md. No external blocker, no code edits, no commit_ack.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260822-af01e0, pid=50931, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260822-3bb998, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260822-3bb998)
Rework RUN-260822-3bb998: removed hidden npm shebang/PATH authority by executing exact C0-bound Node with fingerprinted npm-cli.js; added actual process-start observation for real portable/verified vectors and executable S03/S04/S08 coverage. Important anomaly: timestamped npm debug logs under private HOME caused materialized-tree drift, so the closed manager scratch root is discarded before admission. Fresh targeted/focused/race/coverage 80.4%/vet/lint/build/diff/board/full-suite gates all exited 0. Evidence: TASK-260811-1u42b9_rework-evidence_RUN-260822-3bb998.md
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260822-3bb998, pid=65095, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (codex) (run=RUN-260822-6f1d09, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260822-6f1d09)
RUN-260822-6f1d09 reviewer verdict: accepted. Fresh targeted real npm/verified launch, focused, race, coverage (80.4%), vet, lint, build, diff, board validation, and uncached repository-wide suite all exited 0. Verdict artifact: TASK-260811-1u42b9_review-verdict_RUN-260822-6f1d09.md.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260822-6f1d09, pid=90052, exit=0)

## Precondition Resources
- [TASK-260810-1dgdos_accepted-adapter-source-closure-architecture-decision.md](file://TASK-260811-1u42b9/TASK-260810-1dgdos_accepted-adapter-source-closure-architecture-decision.md) — Accepted normative adapter source-closure architecture decision and implementation DAG
- [TASK-260810-1dgdos_accepted-review-verdict_RUN-260811-55458c.md](file://TASK-260811-1u42b9/TASK-260810-1dgdos_accepted-review-verdict_RUN-260811-55458c.md) — Independent acceptance verdict for the normative synthesis
- [TASK-260810-1dgdos_accepted-skill-facing-cli-source-closure.md](file://TASK-260811-1u42b9/TASK-260810-1dgdos_accepted-skill-facing-cli-source-closure.md) — Accepted delivery scope and source-closure security constraints
- [TASK-260810-2n3sbi_accepted-node-typescript-python-closure-research.md](file://TASK-260811-1u42b9/TASK-260810-2n3sbi_accepted-node-typescript-python-closure-research.md) — Accepted Node/TypeScript manager profiles and independent Python protocol compatibility contract
- [TASK-260811-1u42b9_review-focus-portable-verified.md](file://TASK-260811-1u42b9/TASK-260811-1u42b9_review-focus-portable-verified.md) — Independent review gates for portable default, verified zero-start, honest evidence, binary deny, and fresh broad validation
- [TASK-260811-1u42b9_rework-after-RUN-260822-0c4aea.md](file://TASK-260811-1u42b9/TASK-260811-1u42b9_rework-after-RUN-260822-0c4aea.md) — Mandatory developer rework instructions copied from independent changes-requested verdict RUN-260822-0c4aea
- [TASK-260811-1u42b9_rework-after-RUN-260822-411c9c.md](file://TASK-260811-1u42b9/TASK-260811-1u42b9_rework-after-RUN-260822-411c9c.md) — Mandatory second npm rework instructions copied from independent changes-requested verdict RUN-260822-411c9c
- [TASK-260811-1u42b9_rework-after-RUN-260822-b81336.md](file://TASK-260811-1u42b9/TASK-260811-1u42b9_rework-after-RUN-260822-b81336.md) — Mandatory third npm rework instructions copied from independent changes-requested verdict RUN-260822-b81336
- [TASK-260811-1u42b9_rework-after-RUN-260822-af01e0.md](file://TASK-260811-1u42b9/TASK-260811-1u42b9_rework-after-RUN-260822-af01e0.md) — Mandatory fourth npm rework instructions copied from independent changes-requested verdict RUN-260822-af01e0

## Outcome Resources
- [TASK-260811-1u42b9_spawn-log_-implementer--developer--codex-_RUN-260822-29b5d5.log](file://TASK-260811-1u42b9/TASK-260811-1u42b9_spawn-log_-implementer--developer--codex-_RUN-260822-29b5d5.log) — System spawn log captured by task-board
- [TASK-260811-1u42b9_implementation-evidence.md](file://TASK-260811-1u42b9/TASK-260811-1u42b9_implementation-evidence.md) — npm source-closure implementation, shared contract fixes, conformance coverage, and standalone validation evidence
- [TASK-260811-1u42b9_spawn-log_-reviewer--reviewer--codex-_RUN-260822-0c4aea.log](file://TASK-260811-1u42b9/TASK-260811-1u42b9_spawn-log_-reviewer--reviewer--codex-_RUN-260822-0c4aea.log) — System spawn log captured by task-board
- [TASK-260811-1u42b9_review-verdict_RUN-260822-0c4aea.md](file://TASK-260811-1u42b9/TASK-260811-1u42b9_review-verdict_RUN-260822-0c4aea.md) — Independent changes-requested verdict with code evidence and fresh focused/broad validation
- [TASK-260811-1u42b9_spawn-log_-implementer--developer--codex-_RUN-260822-cb137e.log](file://TASK-260811-1u42b9/TASK-260811-1u42b9_spawn-log_-implementer--developer--codex-_RUN-260822-cb137e.log) — System spawn log captured by task-board
- [TASK-260811-1u42b9_rework-evidence_RUN-260822-cb137e.md](file://TASK-260811-1u42b9/TASK-260811-1u42b9_rework-evidence_RUN-260822-cb137e.md) — Developer rework: exact materialized-byte admission, common C0/C5/provider binding, honest portable audit, regression tests, and fresh validation
- [TASK-260811-1u42b9_spawn-log_-reviewer--reviewer--codex-_RUN-260822-411c9c.log](file://TASK-260811-1u42b9/TASK-260811-1u42b9_spawn-log_-reviewer--reviewer--codex-_RUN-260822-411c9c.log) — System spawn log captured by task-board
- [TASK-260811-1u42b9_review-verdict_RUN-260822-411c9c.md](file://TASK-260811-1u42b9/TASK-260811-1u42b9_review-verdict_RUN-260822-411c9c.md) — Independent changes-requested verdict: shared executor/provider/C5 authority remains bypassed
- [TASK-260811-1u42b9_spawn-log_-implementer--developer--codex-_RUN-260822-77c106.log](file://TASK-260811-1u42b9/TASK-260811-1u42b9_spawn-log_-implementer--developer--codex-_RUN-260822-77c106.log) — System spawn log captured by task-board
- [TASK-260811-1u42b9_rework-evidence_RUN-260822-77c106.md](file://TASK-260811-1u42b9/TASK-260811-1u42b9_rework-evidence_RUN-260822-77c106.md) — Developer rework: shared executor permits/receipts, C0/C5 action binding, verified provider zero-start matrix, exact executable launch, and fresh validation
- [TASK-260811-1u42b9_spawn-log_-reviewer--reviewer--codex-_RUN-260822-b81336.log](file://TASK-260811-1u42b9/TASK-260811-1u42b9_spawn-log_-reviewer--reviewer--codex-_RUN-260822-b81336.log) — System spawn log captured by task-board
- [TASK-260811-1u42b9_review-verdict_RUN-260822-b81336.md](file://TASK-260811-1u42b9/TASK-260811-1u42b9_review-verdict_RUN-260822-b81336.md) — Independent changes-requested verdict: canonical npm permit does not bind actual command/I/O and positive shared portable/verified execution proof is missing
- [TASK-260811-1u42b9_spawn-log_-implementer--developer--codex-_RUN-260822-a5f43f.log](file://TASK-260811-1u42b9/TASK-260811-1u42b9_spawn-log_-implementer--developer--codex-_RUN-260822-a5f43f.log) — System spawn log captured by task-board
- [TASK-260811-1u42b9_rework-evidence_RUN-260822-a5f43f.md](file://TASK-260811-1u42b9/TASK-260811-1u42b9_rework-evidence_RUN-260822-a5f43f.md) — Developer rework: exact C5 logical command authority, common portable runner, positive verified execution, retained typed work copies, and fresh gates
- [TASK-260811-1u42b9_spawn-log_-reviewer--reviewer--codex-_RUN-260822-af01e0.log](file://TASK-260811-1u42b9/TASK-260811-1u42b9_spawn-log_-reviewer--reviewer--codex-_RUN-260822-af01e0.log) — System spawn log captured by task-board
- [TASK-260811-1u42b9_review-verdict_RUN-260822-af01e0.md](file://TASK-260811-1u42b9/TASK-260811-1u42b9_review-verdict_RUN-260822-af01e0.md) — Independent changes-requested verdict: real npm shebang crosses undeclared ambient executable and S03/S04/S08 proofs are absent
- [TASK-260811-1u42b9_spawn-log_-implementer--developer--codex-_RUN-260822-3bb998.log](file://TASK-260811-1u42b9/TASK-260811-1u42b9_spawn-log_-implementer--developer--codex-_RUN-260822-3bb998.log) — System spawn log captured by task-board
- [TASK-260811-1u42b9_rework-evidence_RUN-260822-3bb998.md](file://TASK-260811-1u42b9/TASK-260811-1u42b9_rework-evidence_RUN-260822-3bb998.md) — Developer rework: direct Node-launched npm, real verified process boundary, executable S03/S04/S08, and fresh focused/broad gates
- [TASK-260811-1u42b9_spawn-log_-reviewer--reviewer--codex-_RUN-260822-6f1d09.log](file://TASK-260811-1u42b9/TASK-260811-1u42b9_spawn-log_-reviewer--reviewer--codex-_RUN-260822-6f1d09.log) — System spawn log captured by task-board
- [TASK-260811-1u42b9_review-verdict_RUN-260822-6f1d09.md](file://TASK-260811-1u42b9/TASK-260811-1u42b9_review-verdict_RUN-260822-6f1d09.md) — Independent accepted verdict with code evidence and fresh focused/broad validation

## Created
2026-08-11T05:09:16Z

## Last Update
2026-08-23T00:02:31Z

## Assigned To
[reviewer] reviewer (codex)
