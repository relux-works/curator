## Status
done

## Review
light

## Task Class
research

## Estimate
estimated(fibonacci(3))

## Blocked By
- (none)

## Blocks
- TASK-260822-1f533i
- TASK-260822-3fkfmf
- TASK-260822-1mwy10

## Checklist
- [x] analysis.md outcome resource on this task answers all five questions with one recommendation each
- [x] Each recommendation carries rationale and at least one rejected option
- [x] Recommendations stay consistent with the decision 0006 doctrine: mandatory portable vs native inventory vs deferred hardened
- [x] Findings written to file
- [x] Key aspects highlighted
- [x] Fact-checking performed — claims verified, sources cited
- [x] Findings linked on the board as a new task-scoped outcome resource
- [x] All questions from task description answered
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
Working context: curator-spec checkout /Users/iv/Developer/ReluxWorks/curator-spec (main at a2d44eb, decision 0008 landed as decisions/0008-enforced-script-capabilities.md). Study there: decisions/0006-portable-manager-worker-execution.md, protocol/core.md section 4.2.1 (lines ~316-470) and 4.3, profiles/manager.md section 2.2.1, SECURITY.md. Story branch for downstream prose: spec/script-worker-v1-normative, worktree /Users/iv/Developer/ReluxWorks/curator-spec/.temp/STORY-260822-3k3hbs/normative-worktree. Deliverable stays on the board: add_resource analysis.md on this task.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [analyst] researcher (claude) (run=RUN-260822-9f1145, max_parallel=20)
spawn run started: [analyst] researcher (claude) (run=RUN-260822-9f1145)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [analyst] researcher (claude) (run=RUN-260822-178366, max_parallel=20)
spawn run started: [analyst] researcher (claude) (run=RUN-260822-178366)
Analysis complete and attached as TASK-260822-1l4r4f_analysis.md (outcome). All five open questions of decision 0008 answered with one recommendation each, rationale, and rejected options.

Answers in brief:
Q1 execution_policy placement -> per command on the script command object, one OPTIONAL field with the single closed value script-worker-v1, no manifest-level default and no override resolution.
Q2 interpreter identity -> neither posed option. Use the driver precedent: package names a closed interpreter identifier, manager resolves it under core.md 8.2 trust rules, per-invocation identity verification of the launched executable file only, library/site-packages tree declared TCB, shebang and file association inert.
Q3 network host globs -> reporting-only. No portable filtering, no inventory entry, no applied-control claim; enforced commands declaring host globs are admitted with a distinct audit warning class rather than rejected.
Q4 evidence cadence -> per invocation, probed once before worker launch, never re-probed mid-session, plus an install-time preflight of the mandatory set; record is result-only and never on the command stdout/stderr.
Q5 Windows scope -> full parity in the mandatory portable set; all asymmetry in the inventory. Five build inventory cells copied verbatim, one new cell descendant-exec-denial (Windows available only when the effective exec set is empty; macOS unavailable under a new reason no-unprivileged-per-process-exec-policy). macOS+Windows only, no Linux column.

Three conflicts found in decision 0008 that the prose tasks must resolve, not inherit:
1. capability-evidence-v1 is closed against a script policy identity (core.md:437) — recommended a new script-capability-evidence-v1 record over a new script-worker-v1-native-control-inventory-v1.
2. 0008 section 3 deny-by-default contradicts the schema default filesystem=repo.
3. 0008 Context cites a Linux mechanism; no execution-policy ledger in this spec has a Linux cell.
Plus: capability values have no normative semantics anywhere in the spec today, and the current script launcher (manager.md:503-509, runtimestore.go:119-145) structurally cannot host enforcement.

One fact-check caveat carried forward: the Windows child-process-policy and macOS deprecated-sandbox claims are corroborated only by decision 0006 in-repo; vendor-doc confirmation is owed before either becomes a normative mechanism string. Logbook entry added 2026-08-22 2020.
Revision 2 of TASK-260822-1l4r4f_analysis.md. A prior spawn (RUN-260822-178366/9f1145) had already written a complete analysis but left the task in backlog and never handed off; this run reconciled rather than overwrote, via resource update.

What revision 2 changed. Revision 1 explicitly recorded as an open maintainer item that the Windows child-process attribute had not been verified against vendor documentation. It has now been verified and the vendor text contradicts revision 1: Microsoft documents PROCESS_CREATION_CHILD_PROCESS_RESTRICTED as effective only inside a sandboxed (AppContainer) application and bypassable given PROCESS_CREATE_PROCESS/PROCESS_VM_WRITE handle access. Q5 was rewritten: descendant-exec-denial is unavailable on Windows (child-process-policy-requires-appcontainer). That cascaded into F5 (the script inventory now ships a Linux column, because otherwise all three new controls would be unavailable on every column at birth, which 0006 forbids) and replaced the applicable_when/not-applicable construct with a host-conditional availability state. Q1-Q4 are unchanged from revision 1 apart from added ledger evidence; Q3 gained the kernel-side argument that Landlock filters ports, never hosts.

Recommendations, one per question. Q1: per-command execution_policy only, closed const script-worker-v1, script commands only, absent = declared-only; no manifest default, no override, no opt-out spelling. Q2: neither 8.2 nor an exec entry - a closed interpreter identifier on the command (the driver precedent), manager-resolved under 8.2 resolution rules, per-invocation file identity only, library tree declared TCB. Q3: reporting-only; host globs configure nothing and get their own audit warning class; per-host allowlisting is a deferred guarantee, never an inventory entry. Q4: exactly one closed script-capability-evidence-v1 record per invocation, probed pre-launch, never re-probed mid-session, plus additive install-time preflight; new record version rather than widening the closed manager-worker-v1 constant. Q5: Windows at full parity in the mandatory portable set, nothing new in the inventory; three platform columns; host-conditional state; ledger rows specified.

Still unconfirmed and handed to the prose tasks: Windows Job Object nesting behaviour for active-process-count-limit, and cgroup-v2 delegation (mapped host-conditional so it cannot produce a false applied claim).

Logbook entry appended to LOGBOOK.md.
agent completed: [analyst] researcher (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-9f1145, pid=94736, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260822-378a01, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260822-378a01)
REVIEW VERDICT (RUN-260822-378a01, read-only): ACCEPTED. AC met — analysis.md revision 2 answers all five open questions of decision 0008 with exactly one recommendation each, 5-8 anchored rationale points per question, and 4-8 rejected options per question, plus a Prose hooks block per question naming the target file and sentence. Consistent with the 0006 doctrine: no new control enters the mandatory portable set, no deferred guarantee is named as an inventory entry, and network-isolation-domain is deliberately not spelled total-network-denial (core.md:443-450 freezes those six names against exactly that).

INDEPENDENT FACT-CHECK, not taken on trust: 20 of 24 in-repo ledger claims re-derived at first hand against curator-spec main=a2d44eb, all exact at the quoted lines — core.md 325-327, 369-370, 412-414, 421-422, 437, 439-442, 443-450, 461-468, 557-558, 942-968, 1288-1315; manager.md 211-219, 246-257, 268-273, 503-509; common.schema.json filesystem default=repo and host-glob pattern; agent-skill-v7 top-level properties; go-host-execution-policy.json platforms/reasons/host-firewall-profile case; provider receipt six established ids; zero linux hits in core.md and manager.md; runtimestore.go WriteBinShim; platform-cases.tsv rows 101-102. F2 confirmed load-bearing: core.md:437 does reject a script-worker-v1 record under capability-evidence-v1, so 0008 section 4 is genuinely not implementable as written and core.md:1306-1311 independently forbids fixing it by widening the closed constant. STORY-260822-2h0v9j AC verified verbatim as ubuntu/macos/windows, which makes the Linux column load-bearing rather than speculative.

EXTERNAL CLAIMS RE-FETCHED FROM PRIMARY SOURCES: Microsoft UpdateProcThreadAttribute quote is verbatim — PROCESS_CREATION_CHILD_PROCESS_RESTRICTED is only effective in sandboxed applications such as AppContainer and is bypassable given PROCESS_CREATE_PROCESS/PROCESS_VM_WRITE handle access. Revision 2 reversal of revision 1 is correct and vendor-sourced. kernel.org Landlock confirmed on every point including network = TCP/UDP ports only, never hosts, which is the sharpest argument in Q3 and it holds.

FOUR NON-BLOCKING FINDINGS for the prose tasks (none is rework of this deliverable): (1) ledger claim 21 overstates the version range — the cited blog says 23.10 shipped the userns restriction OPT-IN, not default-on; 24.04 is where the default lands, and Canonical 24.04 material confirms the substantive half exactly. No recommendation changes since ubuntu-latest is 24.04; narrow the range before 3fkfmf/f4qv7w cite it. (2) Q3 contradicts 0008 section 3 (kernel-grade host filtering enters the native inventory per platform) and is NOT listed among the three flagged 0008 conflicts — right on the merits, but the divergence belongs in that conflict list or 1f533i/f4qv7w inherit it silently. (3) Q5 routes host-conditional controls to platform-case class host-capability, but platform-cases.tsv defines that class as the runner FILESYSTEM cannot create the artefact — a missing Landlock/cgroup/netns kernel feature is not that; f4qv7w needs a widened class or a new one, decided in prose rather than at gate time. (4) Q3 rationale point 1 (the spec already rules on this, in the vectors) claims more than the evidence — the vector constrains the rc5 build inventory only, not what a new inventory may contain; points 2/5/6 carry it regardless.

CHECKED AND NOT FAULTED: Q1 vs core.md:325-327 is met head-on rather than routed around (one-value enum = opt-in to enforcement, not a policy choice; monotonically restrictive). The Linux column survives the strongest objection I could raise — host-conditional licenses applying and reporting on hosts that have the control, which unavailable forbids outright — and is honestly labelled as the largest call with a cheap undo path. The deliberate build-policy divergences (open stdin, streamed output, no wall-clock deadline) are correctly identified and restated as positive controls; inheriting the closed-stdin rule by reflex would have silently broken every interactive CLI shim.

NO COMMIT_ACK — reviewer archetype. Uncommitted deliverables: .research/260822_decision-0008-open-questions.md (untracked, byte-identical to the board artifact) and the LOGBOOK.md entry. The commit-owning mover commits that scope, then makes the final done transition with commit_ack=scope_committed. Evidence: TASK-260822-1l4r4f_review-verdict.md.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-378a01, pid=63968, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260822-1l4r4f_spawn-log_-analyst--researcher--claude-_RUN-260822-9f1145.log](file://TASK-260822-1l4r4f/TASK-260822-1l4r4f_spawn-log_-analyst--researcher--claude-_RUN-260822-9f1145.log) — System spawn log captured by task-board
- [TASK-260822-1l4r4f_spawn-log_-analyst--researcher--claude-_RUN-260822-178366.log](file://TASK-260822-1l4r4f/TASK-260822-1l4r4f_spawn-log_-analyst--researcher--claude-_RUN-260822-178366.log) — System spawn log captured by task-board
- [TASK-260822-1l4r4f_analysis.md](file://TASK-260822-1l4r4f/TASK-260822-1l4r4f_analysis.md) — Revision 2: five decision-0008 open questions resolved, one recommendation each with rationale and rejected options. Q5 corrected against Microsoft vendor documentation (Windows child-process policy is bypassable outside AppContainer), reversing the F5 Linux-column call and replacing applicable_when/not-applicable with a host-conditional availability state.
- [TASK-260822-1l4r4f_spawn-log_-reviewer--reviewer--claude-_RUN-260822-378a01.log](file://TASK-260822-1l4r4f/TASK-260822-1l4r4f_spawn-log_-reviewer--reviewer--claude-_RUN-260822-378a01.log) — System spawn log captured by task-board
- [TASK-260822-1l4r4f_review-verdict.md](file://TASK-260822-1l4r4f/TASK-260822-1l4r4f_review-verdict.md) — Reviewer verdict RUN-260822-378a01: ACCEPTED; independent re-verification of 20/24 in-repo citations and all 3 load-bearing external claims; 4 non-blocking findings for the prose tasks

## Created
2026-08-22T16:00:34Z

## Last Update
2026-08-22T16:34:47Z

## Assigned To
[reviewer] reviewer (claude)
