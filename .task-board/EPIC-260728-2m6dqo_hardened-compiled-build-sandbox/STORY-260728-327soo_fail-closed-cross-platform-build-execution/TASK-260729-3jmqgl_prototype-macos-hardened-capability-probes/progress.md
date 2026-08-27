## Status
done

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
- [x] Implement executable probes for all six hardened guarantees with stable machine-readable output
- [x] Add adversarial escape and negative controls for every guarantee and prove fail-closed exit behavior
- [x] Run the harness on the macOS primary host and record exact commands, OS/tool versions, and exit codes
- [x] Document supported, unsupported, private, and deprecated mechanisms plus reuse boundaries for curator and csk
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Tests written and passing
- [x] Coverage target ~80%+ for affected code
- [x] New task-scoped outcome artifact attached on the board for reports, logs, screenshots, or other produced evidence
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260728-cb4bf5, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260728-cb4bf5)
Run RUN-260728-cb4bf5 was intentionally cancelled by the orchestrator after about five minutes to preempt this separate non-gating probe for the main toolchain critical-path rework. Preserve any task worktree/partial investigation unchanged; no completion is claimed. Resume with a fresh Claude Opus 5 producer after the critical-path barrier is accepted.
Resume directive: continue from the preserved partial worktree left by cancelled RUN-260728-cb4bf5; inspect and reuse valid partial probes before editing. Finish the original task AC and DoD. This remains a separate non-gating hardened story and must not alter or block the main compiled-skill epic.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260728-a1cd57, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260728-a1cd57)
RESUME DIRECTIVE 2026-07-28: Continue from the preserved task worktree /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3jmqgl/worktree and inspect prior partial work before changing it. Complete the macOS hardened capability probe and evidence packet under the existing AC. Do not claim guarantees that the host cannot prove, do not broaden this non-gating story into implementation, and do not stage, commit or publish. Preserve truthful unsupported/unqualified outcomes and hand off for independent review when all checklist/evidence gates are met.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260728-d7a61c, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260728-d7a61c)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [tester] tester (claude) (run=RUN-260728-1bc898, max_parallel=20)
spawn run started: [tester] tester (claude) (run=RUN-260728-1bc898)
ORCHESTRATOR PREEMPTION 2026-07-29: RUN-260728-1bc898 was operator-cancelled solely to reallocate the five-Opus ceiling to main-path Rust specification rework. This is not a defect, blocker or scope change. Preserve the existing isolated task worktree and resume from it later; do not restart or discard partial probe work. Hardened execution remains explicitly separate and non-gating.
RESUME DIRECTIVE 2026-07-29: A surplus fifth Opus slot reopened after hardened specification rework moved to independent review. Resume from the preserved task worktree and prior partial probes; inspect existing bytes before changing anything. Continue the original macOS capability evidence task only, remain separate and non-gating, make no guarantee beyond measured host evidence, and yield again if a main-path language or Curator rework becomes ready. No stage, commit, publish, pin or host install.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [tester] tester (claude) (run=RUN-260729-12189e, max_parallel=20)
spawn run started: [tester] tester (claude) (run=RUN-260729-12189e)
ORCHESTRATOR PAUSE 2026-07-29: RUN-260729-12189e was intentionally cancelled to release the fifth Opus slot for blocking Kotlin specification cycle-3 rework. This hardened macOS probe is surplus and non-gating; preserve its private worktree/evidence and resume from the existing state only after the main Rust/Swift/Kotlin specification path has a free Opus slot.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260729-32ca38, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260729-32ca38)
RUN-260729-32ca38 result — macOS hardened capability probes complete, ready for review.

RESULT: on macOS 26.5 / arm64 / SIP enabled / go1.25.5 / unprivileged uid 502, 6 of 11 capability classes are available and 1 of 6 hardened guarantees is established. The harness rejects fail-closed with exit 1, rejected_before=capability-probe, diagnostic=hardened_capability_unavailable, qualification_status=unqualified.

Available: network-syscall-denial, read-only-source-view, read-only-toolchain-view, write-path-confinement, exec-path-allowlist, active-capability-probe. Unavailable: domain-membership-enforcement, domain-atomic-termination, preexisting-endpoint-revocation, filesystem-view-restriction, aggregate-resource-bounds. Established: fail-closed-capability-preflight only.

THREE BLOCKING FINDINGS: (1) no unescapable domain - a descendant that calls setsid leaves the process group/session and survives the group SIGKILL; seatbelt policy IS inherited and cannot be renounced, but that binds access rules only, not accounting or termination, and this single class is required by 4 of 6 guarantees; (2) dyld requires file-read-data on the literal / or every dynamically linked program aborts at startup, so the root namespace stays enumerable and filesystem-view-restriction cannot be closed on this backend; (3) RLIMIT_* binds per process not per tree, and macOS has no per-directory byte cap, so no aggregate bound exists.

WHAT WAS DONE THIS RUN, on top of the preserved partial worktree: fixed all 46 golangci-lint findings to zero; added the missing negative controls (uncontained undeclared-path reads for filesystem-view-restriction, per-process-bound-binds for aggregate-resource-bounds, executed instrument self-tests for active-capability-probe, domain-teardown-was-issued for domain-atomic-termination) so every class now carries a positive test plus a negative control or adversarial escape; added end-to-end tests driving the real harness plus a cmd test suite, raising coverage from 37.1%/0.0% to 84.4%/81.2%; fixed a latent test fragility where a permanently-unset TMPDIR was masking an over-long unix socket path; added capture-evidence.sh and README.md.

VERIFICATION, each command run standalone with its real exit status: go build ./... =0, go vet ./... =0, gofmt -l . empty =0, go test ./... =0, golangci-lint run --config ../../.golangci.yml ./... = 0 issues =0. Coverage per package: cmd 81.2%, evidence 99.3%, inside 92.5%, probe 84.4%, seatbelt 97.7%, spec 100.0%. Evidence capture cases: list-classes exit 0, measure exit 1 (fail-closed, expected), fail-closed-sweep exit 1 with all 11 classes passing, assert-rejected exit 0, assert-established exit 2. capture-evidence.sh exit 0.

KNOWN PRE-EXISTING, NOT CAUSED HERE: repository-root golangci-lint run still cannot load export data for the skill-go-testing-tools/tuitestkit submodule dependency, so the prototype module is linted directly with the repository .golangci.yml.

SCOPE HELD: prototypes/macos-hardened-probes/ in the task worktree only, separate Go module, not in the Makefile, imported by nothing. No curator-spec change, no curator/csk production integration, no stage, no commit, no publish. Nothing claims enforcement or qualification.

ARTIFACTS: TASK-260729-3jmqgl_macos-hardened-capability-outcome.md (guarantee/class matrix, exact commands and exit codes, supported/deprecated/private/unavailable mechanism inventory, curator and csk reuse boundaries, explicit non-claims), TASK-260729-3jmqgl_evidence-packet.tar.gz, TASK-260729-3jmqgl_macos-hardened-probes-source.tar.gz. Findings recorded in LOGBOOK.md 2026-07-29 1527.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260729-32ca38, pid=80266, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260729-6e3c8f, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260729-6e3c8f)
Reviewer RUN-260729-6e3c8f: CHANGES REQUESTED. Independent task-local gate passed after the required clean-process barrier: go test -count=1 ./... exited 0 for all six prototype packages; artifact/source/invariant/cleanup checks also passed. Acceptance remains incomplete because aggregate-resource-bounds executes only RLIMIT_NOFILE and disk-byte probes, not CPU, memory/address-space, process-count, or wall-time descendant-tree limits required by the task; its supervisor-accounting failure is hard-coded rather than derived from same-run domain observations. E2E TestMain also skips all real probes if its binary build fails. Full evidence and required rework are in TASK-260729-3jmqgl_reviewer-verdict_RUN-260729-6e3c8f.md. Route to development for rework and another reviewer cycle.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260729-6e3c8f, pid=8222, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260729-10adc6, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260729-10adc6)
Rework cycle 1 complete after RUN-260729-6e3c8f; ready for review.

ALL FIVE REQUIRED REWORK ITEMS ADDRESSED BY EXECUTABLE MEASUREMENT:

(1) Executable bound probes + matched controls for CPU, memory/address space, process count and wall clock across descendants. New internal/inside/bounds.go (bound matrix: a bounded process installs the declared limit and tries to pass it, an unbounded control makes the identical attempt, a nested descendant inherits the limit and tries again), internal/probe/bounds.go (reduction), internal/probe/wallclock.go (real 3s deadline probe). Descriptor and disk-byte probes retained. Measured: RLIMIT_CPU binds at 1003ms of a declared 1000ms while its descendant burns a fresh 1002ms (aggregate 2005); RLIMIT_AS and RLIMIT_DATA refuse any build-sized value with EINVAL, measured floor 421707907058 bytes = 1570x the declared budget; RLIMIT_NPROC refuses the domains FIRST descendant under a declared budget of 4 because it is accounted per uid. Second escape measured for every installable bound: a member can raise its own soft limit back to the inherited hard limit (raised); matched control with the hard limit lowered is refused EPERM.

(2) supervisor-side-accounting-is-unescapable is now DERIVED in supervisorAccountingCheck from this runs measured session identifiers, teardown handle, teardown error and survivor observations. The hard-coded escapable/pass:false is gone. TestSupervisorAccountingIsDerivedFromThisRun feeds the reduction the numbers a host with an unescapable domain would report and requires the verdict to flip.

(3) Deadline cancellation coverage: probeWallClock measures the descendant tree after the deadline fires, after the process-group teardown, and after the harness pid-directed cleanup, reported as THREE separate checks. Measured: the root dies at 3.002s; BOTH attached and detached descendants survive cancellation; a group SIGKILL reaches the attached one; the detached one survives everything a plain supervisor can issue. Platform result and harness hygiene are deliberately not merged. An unbuildable probe binary now FAILS the e2e suite (TestProbeBinaryBuilds; requireAgent fatals instead of skipping; buildProbeBinary fatals).

(4) Regenerated on the macOS primary host (macOS 26.5 build 25F71, Darwin 25.5.0 arm64, go1.25.5, SIP enabled, uid 502). Gates each run standalone, real exit codes: go build 0, go vet 0, gofmt -l 0 (empty), golangci-lint 0 issues, go test -count=1 0, go test -count=1 -cover 0, capture-evidence.sh 0. Coverage: cmd 81.2, evidence 99.3, inside 87.4, probe 85.2, seatbelt 97.7, spec 100.0. Capture exit codes: list-classes 0, measure 1, fail-closed-sweep 1, assert-rejected 0, assert-established 2, leftover-processes count 0.

(5) Only task-local prototype tests and evidence capture were run. No repository-wide Curator suite, no race, no timeout changes.

ALSO: the mechanism inventory the reviewer flagged as concluding wider than its observations is now one entry per RLIMIT_ resource, and every entry carries exercised + observation naming the checks behind it. 9 of 20 mechanisms were exercised this run; the other 11 say in the artifact that their status is a reading of the published interface and not a measurement.

SELF-CAUGHT DEFECT: an earlier draft declared the CPU bound in milliseconds and passed the same number to setrlimit, which takes whole seconds, installing a limit 1000x the declared one. Every CPU measurement came back unrefused and read as a clean platform finding. Fixed with one named conversion plus an invariant test that every kinds stress ceiling sits above its declared budget.

Headline platform result UNCHANGED and still truthful: macOS establishes 1 of 6 hardened guarantees; the harness rejects fail-closed with exit 1; qualification_status is unqualified in every record it can emit. No production enforcement claim.

Source archive verified byte-identical to the task worktree. No production Curator code, specs, shared caches or unrelated files touched. Nothing committed, staged, published, installed or downloaded; no host configuration changed.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260729-10adc6, pid=25464, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260729-4a8bdb, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260729-4a8bdb)
Reviewer RUN-260729-4a8bdb ACCEPTED after rework cycle 1. Clean process barrier; task-local go test -count=1 ./... exit 0 across all six packages; 33/33 source files byte-identical; 11/11 fail-closed controls pass; no leftover probe processes. Full evidence: TASK-260729-3jmqgl_reviewer-verdict_RUN-260729-4a8bdb.md.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260729-4a8bdb, pid=98509, exit=0)

## Precondition Resources
- [TASK-260729-3jmqgl_review-instructions.md](file://TASK-260729-3jmqgl/TASK-260729-3jmqgl_review-instructions.md) — Independent review scope and process barrier for macOS hardened probes
- [TASK-260729-3jmqgl_rework-cycle-1.md](file://TASK-260729-3jmqgl/TASK-260729-3jmqgl_rework-cycle-1.md) — Focused rework for executable aggregate limit probes and fail-closed E2E coverage

## Outcome Resources
- [TASK-260729-3jmqgl_spawn-log_-implementer--developer--claude-_RUN-260728-cb4bf5.log](file://TASK-260729-3jmqgl/TASK-260729-3jmqgl_spawn-log_-implementer--developer--claude-_RUN-260728-cb4bf5.log) — System spawn log captured by task-board
- [TASK-260729-3jmqgl_spawn-log_-implementer--developer--claude-_RUN-260728-a1cd57.log](file://TASK-260729-3jmqgl/TASK-260729-3jmqgl_spawn-log_-implementer--developer--claude-_RUN-260728-a1cd57.log) — System spawn log captured by task-board
- [TASK-260729-3jmqgl_spawn-log_-implementer--developer--claude-_RUN-260728-d7a61c.log](file://TASK-260729-3jmqgl/TASK-260729-3jmqgl_spawn-log_-implementer--developer--claude-_RUN-260728-d7a61c.log) — System spawn log captured by task-board
- [TASK-260729-3jmqgl_spawn-log_-tester--tester--claude-_RUN-260728-1bc898.log](file://TASK-260729-3jmqgl/TASK-260729-3jmqgl_spawn-log_-tester--tester--claude-_RUN-260728-1bc898.log) — System spawn log captured by task-board
- [TASK-260729-3jmqgl_spawn-log_-tester--tester--claude-_RUN-260729-12189e.log](file://TASK-260729-3jmqgl/TASK-260729-3jmqgl_spawn-log_-tester--tester--claude-_RUN-260729-12189e.log) — System spawn log captured by task-board
- [TASK-260729-3jmqgl_spawn-log_-implementer--developer--claude-_RUN-260729-32ca38.log](file://TASK-260729-3jmqgl/TASK-260729-3jmqgl_spawn-log_-implementer--developer--claude-_RUN-260729-32ca38.log) — System spawn log captured by task-board
- [TASK-260729-3jmqgl_macos-hardened-capability-outcome.md](file://TASK-260729-3jmqgl/TASK-260729-3jmqgl_macos-hardened-capability-outcome.md) — Rework cycle 1: executable CPU/memory/process/wall-clock bound probes, derived supervisor verdict, per-resource mechanism inventory; numbers from evidence-run-06
- [TASK-260729-3jmqgl_evidence-packet.tar.gz](file://TASK-260729-3jmqgl/TASK-260729-3jmqgl_evidence-packet.tar.gz) — Evidence packet run-06: host facts, exit codes, leftover-process check, evidence.json, report.json, fail-closed report, per-case stdout/stderr
- [TASK-260729-3jmqgl_macos-hardened-probes-source.tar.gz](file://TASK-260729-3jmqgl/TASK-260729-3jmqgl_macos-hardened-probes-source.tar.gz) — Prototype source after rework cycle 1, byte-identical to the task worktree: bound matrix, wall-clock deadline probe, derived supervisor verdict, tests
- [TASK-260729-3jmqgl_spawn-log_-reviewer--reviewer--codex-_RUN-260729-6e3c8f.log](file://TASK-260729-3jmqgl/TASK-260729-3jmqgl_spawn-log_-reviewer--reviewer--codex-_RUN-260729-6e3c8f.log) — System spawn log captured by task-board
- [TASK-260729-3jmqgl_reviewer-verdict_RUN-260729-6e3c8f.md](file://TASK-260729-3jmqgl/TASK-260729-3jmqgl_reviewer-verdict_RUN-260729-6e3c8f.md) — Independent reviewer changes-requested verdict: resource/time probe gaps, test fail-closed gap, and verified passing task-local tests
- [TASK-260729-3jmqgl_spawn-log_-implementer--developer--claude-_RUN-260729-10adc6.log](file://TASK-260729-3jmqgl/TASK-260729-3jmqgl_spawn-log_-implementer--developer--claude-_RUN-260729-10adc6.log) — System spawn log captured by task-board
- [TASK-260729-3jmqgl_rework-cycle-1-verification.md](file://TASK-260729-3jmqgl/TASK-260729-3jmqgl_rework-cycle-1-verification.md) — Rework cycle 1 verification log: host, gate commands and real exit codes, test and coverage output, evidence capture exit codes, rework-item traceability
- [TASK-260729-3jmqgl_spawn-log_-reviewer--reviewer--codex-_RUN-260729-4a8bdb.log](file://TASK-260729-3jmqgl/TASK-260729-3jmqgl_spawn-log_-reviewer--reviewer--codex-_RUN-260729-4a8bdb.log) — System spawn log captured by task-board
- [TASK-260729-3jmqgl_reviewer-verdict_RUN-260729-4a8bdb.md](file://TASK-260729-3jmqgl/TASK-260729-3jmqgl_reviewer-verdict_RUN-260729-4a8bdb.md) — Independent accepted reviewer verdict after rework cycle 1

## Created
2026-07-28T20:07:18Z

## Last Update
2026-07-29T13:13:40Z

## Assigned To
[reviewer] reviewer (codex)
