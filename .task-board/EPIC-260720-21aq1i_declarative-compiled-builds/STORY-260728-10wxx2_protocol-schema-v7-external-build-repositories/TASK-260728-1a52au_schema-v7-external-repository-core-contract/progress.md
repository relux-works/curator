## Status
done

## Review
required

## Task Class
docs

## Estimate
estimated(fibonacci(5))

## Blocked By
- TASK-260720-1nvomm

## Blocks
- TASK-260728-17sclp
- TASK-260728-wy3dsw

## Checklist
- [x] Map every architecture-v6 normative section to owned spec text and record exclusions
- [x] Run protocol documentation and full curator-spec validation gates
- [x] Docs updated and consistent with current code
- [x] No discrepancies between code and description
- [x] Result linked as a new task-scoped outcome resource
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
Execution directive (2026-07-28): Target repository is /Users/iv/Developer/ReluxWorks/curator-spec, while the task board is in /Users/iv/Developer/ReluxWorks/curator. The curator-spec checkout contains accepted uncommitted rc.4 foundation changes in SECURITY.md, profiles/manager.md, protocol/core.md, and decisions/0004-compile-only-build-drivers.md; treat them as required input and never revert or overwrite them. Create an isolated task worktree under /Users/iv/Developer/ReluxWorks/.temp/TASK-260728-1a52au/curator-spec-worktree from the current curator-spec HEAD, seed only those accepted scoped working-tree files byte-for-byte, and implement this task there. Do not copy .DS_Store or .temp content. Do not commit or stage. Read the binding architecture-v6 and review-v6 resources from the curator board. Edit only normative decision/core/security documentation owned by this task; schemas, vectors, generated cases, release metadata, Curator, and cocoaskills are out of scope. Run scoped and full curator-spec validation in the pinned task-local environment, attach a task-scoped outcome and validation evidence to this board task, and route to to-review.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] doc-writer (codex) (run=RUN-260727-a527c3, max_parallel=20)
spawn run started: [implementer] doc-writer (codex) (run=RUN-260727-a527c3)
Implementation logbook 2026-07-28: Added Decision 0005 plus schema-7 external-repository normative core and security text in the isolated curator-spec worktree. The contract freezes schemas 1-6, go-v1, receipt v1, marker v1/v2, claim v1/v2, and rc.4; makes declared/effective source, full object-format commit locks, exact-tag-only assertions, whole-snapshot independent audit before cache/compiler, manager-derived output, operator credential/signing ownership, protected status/repair, typed failures, and closed future drivers normative. Deliberate exclusions: wire schemas/generated cases remain TASK-260728-17sclp; exact Git argv/SSH/config/ref/pack/cat-file and lifecycle diagnostics remain TASK-260728-wy3dsw; vectors/release gates remain TASK-260728-3b8qym. Accepted rc.4 profile/Decision 0004 seeds remain byte-identical. Post-edit gates: tools/validate.py exit 0 (30 schemas/93 vectors), make validate exit 0 (8 Python tests and Go tools), git diff --check exit 0. Outcome: TASK-260728-1a52au_core-contract-outcome.md.
agent completed: [implementer] doc-writer (codex) (exit=0)
spawn run completed: codex (run=RUN-260727-a527c3, pid=41955, exit=0)
Independent review directive (2026-07-28): Review the producer worktree /Users/iv/Developer/ReluxWorks/.temp/TASK-260728-1a52au/curator-spec-worktree without editing it or the source checkout. Treat /Users/iv/Developer/ReluxWorks/curator-spec current accepted rc.4 files as the seeded baseline; verify profiles/manager.md and decisions/0004 remain byte-identical and isolate the task delta in SECURITY.md, protocol/core.md, and decisions/0005-external-build-repositories.md. Compare every normative claim and the outcome ownership map to architecture-v6 SHA-256 2abae77d80eba6789f9911db7e9722595b4f21ba47391ca9eafd0064af03d67e and review-v6. Adversarially check schema 1-6/go-v1/receipt-v1/marker-v1-v2/rc.4 freeze; source lock/tag and declared/effective identity; monorepo descriptor versus manager-owned command/output; clean Git/raw snapshot and independent audit ordering; protected offline/currentness/repair/GC semantics; credential/signing ownership; stable failure classes; future closed-driver rule; and strict ownership boundaries that leave wire schemas, exact manager profile, vectors, implementations, and release metadata downstream. Run independent scoped and full curator-spec validation from clean task-local tooling, verify no staging/commits or unrelated files, attach TASK-260728-1a52au_core-contract-review.md with ACCEPTED or CHANGES REQUESTED and exact evidence, and set done only if accepted.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260727-39b7e2, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260727-39b7e2)
Independent review verdict 2026-07-28: CHANGES REQUESTED. Validation is green, seed integrity and scope isolation are confirmed, but normative rework is required: Core 6.3 ties network-substitution revision width to declared rather than effective object format; accepted section 9.1 snapshot-key/dedup rules and section 9.4 GC safety/root rules are absent despite the outcome map claiming Core ownership; and external allowlist/revocation/registry/tag-lock/audit policy gates are not explicitly required before cache/compiler work. Full evidence and exact resolution are in TASK-260728-1a52au_core-contract-review.md. Route to doc implementation rework, then a fresh reviewer cycle.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260727-39b7e2, pid=50884, exit=0)
Rework directive 1 (2026-07-28): Consume TASK-260728-1a52au_core-contract-review.md and modify only the existing task-owned files in /Users/iv/Developer/ReluxWorks/.temp/TASK-260728-1a52au/curator-spec-worktree. Close all three findings exactly: (1) protocol/core.md structured network-substitution revision MUST be a full lowercase object ID for the effective repository object format, while the declaration remains unchanged and effective state records the actual format; audit nearby wording for declared/effective collapse. (2) Add portable normative current-driver rules for the protected external snapshot key: complete effective identity kind/value, effective object format, full effective commit, and external build-source digest; deduplicate snapshots only on complete key equality; keep audit decisions subject-specific; add high-level GC roots and safety for valid marker-v3 external/local references, in-flight journal roots, unchanged marker-v1/v2 handling, receipt content not being a root, conservative retention of unreadable/unprovable state, and no execution/adoption. Leave physical paths and exact manager mechanics downstream, and correct the outcome ownership map/exclusions. (3) Require allowlist, revocation, registry, tag-lock, and applicable audit-policy gates independently for each external subject before artifact-cache lookup or compiler work. Preserve every accepted section, seeded rc.4 bytes, scope exclusions, and no commit/stage rule. Run independent-ready scoped/full validation and diff checks. Update the main outcome mapping if needed and attach a distinct new outcome TASK-260728-1a52au_core-contract-rework-1.md with exact changes and evidence, then route to to-review.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] doc-writer (codex) (run=RUN-260727-62e8e5, max_parallel=20)
spawn run started: [implementer] doc-writer (codex) (run=RUN-260727-62e8e5)
Rework 1 logbook 2026-07-28: Closed all three independent-review findings in isolated curator-spec worktree. Core 6.3 now sizes network-substitution revision by effective repository object format and keeps declared state unchanged; Core 6.5 independently requires applicable allowlist, revocation, registry, tag-lock, and audit-policy gates before cache/compiler; new Core 9.4 defines complete protected snapshot keys, complete-key-only deduplication, subject-specific audit decisions, marker/journal GC roots, unchanged marker-v1/v2 behavior, receipt non-root behavior, conservative retention, and no execution/adoption. Main ownership map corrected; exact manager mechanics remain downstream. Validation: tools/validate.py exit 0; first make validate attempt exit 2 because an over-restricted PATH hid Go (not counted green); rerun with task venv prepended to normal PATH exit 0; git diff --check exit 0; rc.4 seed cmp checks exit 0; index clean exit 0. New outcome: TASK-260728-1a52au_core-contract-rework-1.md.
agent completed: [implementer] doc-writer (codex) (exit=0)
spawn run completed: codex (run=RUN-260727-62e8e5, pid=57577, exit=0)
Fresh review cycle 2 directive (2026-07-28): Independently review the current producer worktree read-only, using architecture-v6 and the first verdict as binding. Verify all three requested changes, not just keyword presence: network substitution revision width follows the effective repository object format without mutating/collapsing declared state; current-driver snapshot key is the exact complete tuple, dedup requires complete equality and never reuses audit decisions; marker-v3 and in-flight journal roots, unchanged marker-v1/v2 behavior, receipt non-root, conservative uncertain retention, and no execute/adopt/permission-repair GC behavior are normative while physical mechanics remain downstream; every external subject independently passes applicable allowlist, revocation, registry, tag-lock, and audit-policy gates before cache/compiler. Re-run a regression sweep over all previously accepted core/security boundaries, baseline byte equality, scope isolation, task-owned hashes, clean index/no commit, tools/validate.py, make validate with the actual toolchain PATH, and git diff --check. Attach a distinct TASK-260728-1a52au_core-contract-review-2.md verdict. Mark done only on full acceptance; otherwise route to to-dev with exact remaining gaps.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260727-177dd8, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260727-177dd8)
Fresh independent review 2 (2026-07-28): ACCEPTED. All three prior findings are closed: effective-format revision width, independent external policy gates before cache/compiler, and complete snapshot-key/dedup/GC rules. Full architecture-v6 regression and ownership-map review found no remaining discrepancy. Independent tools/validate.py, make validate, git diff --check, seed byte-equality, clean-index, same-HEAD, and owned-hash checks passed. Verdict evidence: TASK-260728-1a52au_core-contract-review-2.md.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260727-177dd8, pid=63491, exit=0)

## Precondition Resources
- [TASK-260728-1a52au_architecture-v6-precondition.md](file://TASK-260728-1a52au/TASK-260728-1a52au_architecture-v6-precondition.md) — Accepted external-build-repository architecture and review pointer

## Outcome Resources
- [TASK-260728-1a52au_spawn-log_-implementer--doc-writer--codex-_RUN-260727-a527c3.log](file://TASK-260728-1a52au/TASK-260728-1a52au_spawn-log_-implementer--doc-writer--codex-_RUN-260727-a527c3.log) — System spawn log captured by task-board
- [TASK-260728-1a52au_core-contract-outcome.md](file://TASK-260728-1a52au/TASK-260728-1a52au_core-contract-outcome.md) — Architecture-v6 ownership map, exclusions, changed docs, rework closure, and validation evidence
- [TASK-260728-1a52au_spawn-log_-reviewer--reviewer--codex-_RUN-260727-39b7e2.log](file://TASK-260728-1a52au/TASK-260728-1a52au_spawn-log_-reviewer--reviewer--codex-_RUN-260727-39b7e2.log) — System spawn log captured by task-board
- [TASK-260728-1a52au_core-contract-review.md](file://TASK-260728-1a52au/TASK-260728-1a52au_core-contract-review.md) — Independent reviewer verdict and validation evidence
- [TASK-260728-1a52au_spawn-log_-implementer--doc-writer--codex-_RUN-260727-62e8e5.log](file://TASK-260728-1a52au/TASK-260728-1a52au_spawn-log_-implementer--doc-writer--codex-_RUN-260727-62e8e5.log) — System spawn log captured by task-board
- [TASK-260728-1a52au_core-contract-rework-1.md](file://TASK-260728-1a52au/TASK-260728-1a52au_core-contract-rework-1.md) — Focused normative rework closing reviewer findings with scoped and full validation evidence
- [TASK-260728-1a52au_spawn-log_-reviewer--reviewer--codex-_RUN-260727-177dd8.log](file://TASK-260728-1a52au/TASK-260728-1a52au_spawn-log_-reviewer--reviewer--codex-_RUN-260727-177dd8.log) — System spawn log captured by task-board
- [TASK-260728-1a52au_core-contract-review-2.md](file://TASK-260728-1a52au/TASK-260728-1a52au_core-contract-review-2.md) — Fresh independent accepted verdict after normative rework

## Created
2026-07-27T20:20:00Z

## Last Update
2026-07-27T21:03:17Z

## Assigned To
[reviewer] reviewer (codex)
