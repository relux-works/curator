## Status
development

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
- [x] snapshot-nonrecovering-capture: Acquire the existing manager-home mutation lock with its bounded wait through a new non-recovering path. Do not call `beginOperation` or `beginOperationValidate
- [x] snapshot-nonrecovering-capture: Read pending transaction state without running recovery or cleaning journals. A pending transaction returns `environment_materialize_recovery_required`; unreada
- [x] snapshot-nonrecovering-capture: Reselect profile and effective machine policy under the lock. Revalidate selected profile/policy, lock, marker, surfaces and every captured object's identity be
- [x] snapshot-nonrecovering-capture: Share a side-effect-free content verification core with metadata-only status work. Existing `verifyHome` invokes the vendor release probe (`managed.go:1957 →
- [x] snapshot-nonrecovering-capture: Bound capture; release the lock before publication transport or external service waits. Normal lock bookkeeping is distinct from forbidden source provisioning/r
- [x] Every row of the leaf's paired test table is implemented: the negative case asserts the intended refusal and unchanged forbidden state beside its positive control
- [x] The leaf's bounded checker or plant control is executed and shown to fail on the planted violation
- [x] Only the leaf's declared paths and seams change; anything outside is listed and justified in the handoff
- [ ] Focused tests and the repository's full gate pass; exact commands and results are in the handoff
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] In a managed Story worktree the candidate is left UNCOMMITTED in the worktree for the handoff to snapshot — never commit on the Story branch. A producer commit moves the branch tip off the recorded checkpoint and the handoff refuses with change_request_candidate_committed_past_checkpoint; repair with `git reset --soft <checkpoint_oid>` before completing again.
- [x] Every command, message, state, or refusal named in the AC is driven through the production entry point by a named committed test, or is declared a stated bound. Report coverage as a ratio — `n of m AC rows driven` — and name the production call site for each. Prose in place of the ratio is not evidence.
- [x] Gating, refusing, validating, authorizing, or attesting behavior covered by negative tests that fail when the gate admits what it must reject, with the production call site named
- [x] Every gate ships at least one NARROWING mutant — the gate stays present and is weakened to admit exactly one member of the class it must reject, and a named test must fail. A delete-only mutant proves only that the gate exists and is not accepted as evidence.
- [x] A gate that inspects source text is additionally attacked by a mutant that PRESERVES the searched-for token and changes behavior, and the mutant harness executes the behavioral suite, not only the static checker.
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Paired row 1: current verified source captures; prepared, committing, cleanup and rollback journals refuse without recovery or source/target mutation (TestMaterializeCapturePendingDoesNotRecover).
- [x] Paired row 2: lock release permits fresh selection and policy; timeout refuses before selection (TestMaterializeCaptureWaitReselectsAndTimeout).
- [x] Paired row 3: cooperating use, sync and repair serialize; external multi-file mixing refuses and preserves the planted generation (TestMaterializeCaptureSerializesProfileWriters, TestMaterializeCaptureConcurrentRepair, TestMaterializeCaptureExternalTornGeneration).
- [x] Paired row 4: stable source/store objects capture; equal-fragment ABA and byte, mode, identity or policy drift refuse without repair (TestMaterializeCaptureSameFragmentDrift, TestMaterializeCaptureStoreObjectIdentity, TestMaterializeCaptureAuthorizingDrift).
- [x] Paired row 5: successful capture and metadata status never execute the planted vendor executable with nil Detect (TestMaterializeCaptureAndStatusNeverProbeVendor); CLI and Windows POSIX-sentinel limits are declared.

## Notes
spawn selection rationale tuple: {"role":"developer","pair":"gpt-6.1-sol/high","text":"Wave 7 SNAPLOCK producer: docs-only, codex gpt-6.1-sol high (R80), owner go 2026-10-13"}
spawn selection rationale for gpt-6.1-sol/high: Wave 7 SNAPLOCK producer: docs-only, codex gpt-6.1-sol high (R80), owner go 2026-10-13
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[codex,claude], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (codex) (run=RUN-261016-624f8e, max_parallel=4)
spawn run started: [implementer] developer (codex) (run=RUN-261016-624f8e)
Implementing the non-recovering capture transaction and metadata-only verification seam. Normative snapshot sections 1/5 and CIP-0014 checked at curator-spec d2a942cf17e4eb017c45475442da1456d8dfbe6c; no conflict found. Story checkpoint and refreshed origin/main both equal 0d00a109ad1da77a5506747c89f24af9442eb341. Narrow required additions outside the named files: managerlock opens existing lock state without provisioning and revalidates lock identity; managed/status share the no-vendor verification core; transaction recovery-entry instrumentation proves non-entry. The existing remote gate explicitly disabled signing and substituted an identity; its snapshot creation is being corrected to use the configured signed identity so the required gate can run under repository policy. Candidate remains uncommitted. Missing catalog surface table is a brief gap; the supplied paired behavior table is mapped separately in the outcome.
Capture implementation and behavioral controls are frozen for review candidate 90685b712e500a752f23430a9b5f9385740cd8b5 (tree 3dc47dceb444751697ce4702fe85d63fd6d82f73, signature G). Current focused race suite, build, vet, lint and nine narrowing plants are green/expected-red as appropriate; every plant has a named failing behavioral test and exit 1. The permission-denied journal regression is retained. Outcome logbook, mutant results and reproduction harness are attached. Full hosted gate is still running at https://github.com/relux-works/curator/actions/runs/38835465876; no handoff or full-gate success is claimed while it is pending. CLI/closure/runtime integration bounds will be stated explicitly in the handoff.
The prior full gate returned exit 1: Windows test fixtures read a held byte-range lock, attempted replacement of read-only destinations, and emitted unclassified skip reasons. Test-only corrections preserve production source and use existing skip classes; full focused race, lint, build, vet and Windows cross-compilation now exit 0, and all nine narrowing plants were rerun and killed. The replacement signed candidate is 37f08891e6d0325cc6cd2d3c9a9d604de1860719; full gate is running at https://github.com/relux-works/curator/actions/runs/38840074132. Full-gate checklist remains open; no review handoff claimed.
Gate 38840074132 exited 1: 19 enabled jobs passed; Windows failed four held-object replacement plants. Readonly clearing was insufficient. Test fixtures now use Windows POSIX rename semantics through the existing dependency; production behavior is unchanged. Focused race, lint and Windows compilation exited 0. Third frozen full gate is running; full-gate checklist remains unchecked.
Current signed gate candidate: 1e6be490fee5768d0aa6808b825a20437058ef1c, tree b91a63aa36e9e266a7514d9a4d92356181bbfc6f, signature G. Focused race, lint, build, vet and Windows compilation exited 0; all nine narrowing plants were replayed and killed with named failures (Go exits 1, harness exit 0). Gate https://github.com/relux-works/curator/actions/runs/38845304296 has 17 enabled jobs green, including complete Linux and macOS Test suites; two Race suites and Windows Test remain running. The Linux artifact records all 54 SNAP-LOCK cases passing with no skips. Story HEAD/index remain unchanged; full-gate checklist is still open.

## Precondition Resources
- [rc7-leaf-snapshot-nonrecovering-capture.md](file://TASK-261016-4zofp0/rc7-leaf-snapshot-nonrecovering-capture.md)
- [gw-SNAPLOCK-brief.md](file://TASK-261016-4zofp0/gw-SNAPLOCK-brief.md)
- [rc7-code-common.md](file://TASK-261016-4zofp0/rc7-code-common.md)
- [rc7-decisions.md](file://TASK-261016-4zofp0/rc7-decisions.md)

## Outcome Resources
- [TASK-261016-4zofp0_spawn-log_-implementer--developer--codex-_RUN-261016-624f8e.log](file://TASK-261016-4zofp0/TASK-261016-4zofp0_spawn-log_-implementer--developer--codex-_RUN-261016-624f8e.log) — System spawn log captured by task-board
- [TASK-261016-4zofp0_logbook.md](file://TASK-261016-4zofp0/TASK-261016-4zofp0_logbook.md) — Development findings, failed Windows gates, corrected fixture research and integration bounds
- [TASK-261016-4zofp0_mutants.json](file://TASK-261016-4zofp0/TASK-261016-4zofp0_mutants.json) — Nine narrowing plants replayed against the current frozen candidate; actual expected-red exits and named failures
- [TASK-261016-4zofp0_mutant_harness.py](file://TASK-261016-4zofp0/TASK-261016-4zofp0_mutant_harness.py) — Reproduction harness for isolated behavioral mutation controls
- [TASK-261016-4zofp0_mutant_logs.md](file://TASK-261016-4zofp0/TASK-261016-4zofp0_mutant_logs.md) — Sanitized named failing Go test evidence for all nine narrowing plants
- [TASK-261016-4zofp0_coverage-map.md](file://TASK-261016-4zofp0/TASK-261016-4zofp0_coverage-map.md) — Five supplied paired behavior rows mapped to production entry tests and narrowing mutants; integration bounds explicit
- [TASK-261016-4zofp0_platform-cases.json](file://TASK-261016-4zofp0/TASK-261016-4zofp0_platform-cases.json) — Exact-candidate Linux and macOS leaf test/subtest verdicts from hosted artifacts; full gate and Windows still pending

## Created
2026-10-16T01:04:02Z

## Last Update
2026-10-16T03:52:27Z

## Assigned To
[implementer] developer (codex)
