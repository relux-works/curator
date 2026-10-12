# TASK-261012-32vhet — apply-cycle2-findings-lr: gap report

Documentation handoff for the launcher repository, 2026-10-12. Changes are
uncommitted in the assigned Story worktree. The target mode remains specified,
unimplemented and unqualified; these edits are not runtime evidence.

## Findings applied

| Finding | Changed sections and companion text | Changed figures | Disposition and remaining questions |
|---|---|---|---|
| F-16, accepted owner decision | SPEC §10.2 workspace materialization and §10.5 transfer question; README target-mode summary; figure index | artifact-layout.puml adds the private standalone repository; components.puml, launch-paths.puml and launch-states.puml reference the same rule | Launcher contract applied: private `.git`, common directory, refs, index, configuration and hooks; freshly defined local config/hooks; no inherited backlink, executable local configuration or credential helpers; copy/reflink preserves the boundary. Linked worktrees stay within one account, including the board workflow. Optional sharing is immutable/read-only and limited to authorised source objects. Transfer ownership/authorisation across homes and object provenance/access/retention remain unpublished. |
| F-13 | SPEC §10.4 references user-manager §§4.4/4.7 and launch figure; README stop summary; figure index | All four figures reference account-wide termination; launch-paths.puml shows latest-receipt authority and recorded-stop replay; launch-states.puml labels account-wide quiet absence | Launcher reference applied. Stop and retirement share every process of the UID, TERM/grace/KILL, service-domain stop where present and quiet-period account-wide absence. Receipt identity grants authority and does not filter by start time. Old receipts cannot stop successors; recorded stops replay without signalling. Persistence denial and idle-zero-process qualification belong to the user manager. |
| F-18 / R-FATE | SPEC §10.4 lifetime text; README process-fate reference; figure index | All four figures reference user-manager §7 fate table and replaced-daemon writer fence | Launcher reference applied. User manager owns kill domains, stop/crash/restart fate, backend survival promises, journaled stop or lost observations, affected executions and writer fencing. No launcher-owned duplicate table, inferred setsid survival or live supervisor adoption. Unit publication and per-platform qualification remain upstream obligations. |
| F-20 / R-SESSION, owner decision required by AC | SPEC §10.2 replaces execution-scoped permit text and adds host_stop; §10.5 extends joint publication question; README lifecycle summary; figure index | All four figures remove the obsolete four-value permit binding; launch-paths.puml shows owner admission and host_stop | Applied exactly at launcher scope: session-scoped host admission permit for one logical_session_id; owner-admitted session lifecycle authorisation may explicitly permit successors; unpublished member format. Each physical start retains its (execution_id, controller_epoch, op_id) claim/fence/receipt interlock. Successor host_start requires previous-claim absence and is idempotent per (logical_session_id, op_id). Only configured session-host principal can call host_stop for an execution under its permit, through journaled E-stop/exec.stop/confirmed absence, recorded as host-requested and distinct from owner cancel. It confers no start authority. Joint transport, format and ordering publication/qualification still gate SH2. |
| F-21 / R-TIME, owner decision required by AC | SPEC §10.4 time/expiry reference; README time summary; figure index | All four figures reference watermark rules and elapsed-lifetime destruction gate | Applied at launcher reference scope: authenticated or operator-controlled installer time source; durable dispatcher and broker watermark, expiry at max(now, watermark), not_before at min(now, watermark), no untrusted watermark advance and explicit operator reset. Clock steps can only narrow authority. Expiry-driven home deletion/binding removal also require real boot-inclusive monotonic elapsed lifetime. Component implementations, time-policy publication and qualification belong upstream; the launcher has no installer. |

## Revision, figures and hygiene

- Added the dated 0.5.0-draft revision note at the top of SPEC, in its existing style.
- Updated the existing four PlantUML source figures and figure index. Used existing conservative syntax; added no rendered output.
- Removed the README's machine-specific optional CI runner name and setting spelling; kept a generic description and the workflow reference. No workflow/config/code changed.
- Reviewed changed text and figure diffs for stale execution-scoped permit claims, private names, workstation paths, session links and board-resource links. None remain in the changed public text. Platform-defined paths and public upstream references remain existing contracts.
- Shared names used unchanged: R-SESSION, R-FATE and R-TIME, with existing R-FD3/R-START references. Other wave-4 resolutions are outside this brief's launcher work. No new wire fields, refusal codes or values were assigned.
- No LOGBOOK.md was written: the task's current brief explicitly forbids it. Relevant decisions and limitations are captured here and in board notes; the inapplicable logbook checklist item is replaced with an explicit task-brief waiver and board/outcome evidence item, rather than falsely checked as a logbook write.

## Validation evidence and bounds

- No tests, builds, standalone validation gates or mutation tests were run locally. The current brief explicitly prohibits them and sanctions configured CR validation at handoff instead.
- Source review used file reads, scoped searches and git diff inspection. This is documentation review, not a passing automated test or platform qualification.
- No PlantUML rendering was run: the brief states this host has no renderer. Source/figure consistency was reviewed; rendering is unknown.
- The configured CR validation is delegated only to the synchronous developer handoff. Its real result and command exit codes are recorded by the handoff receipt; this report makes no advance claim that it passed.
- Handoff prerequisite: the first developer handoff returned exit 1 for unchecked generic logbook item 6, before CR validation. The task-specific prohibition remains authoritative: no LOGBOOK.md was written. The inapplicable generic item was replaced by explicit waiver/evidence documentation; the attached report and board notes preserve the decisions.
- Lifecycle anomaly: the first required set_status(development) returned exit 1 because an estimate was missing. set_estimate(estimated, fibonacci=2) returned exit 0; the retried status transition returned exit 0. This was a recoverable board prerequisite, not an implementation blocker.

## Open publication register retained

1. Cross-home source/result transfer owner and authorisation would need publication through the existing approved repository/artifact path; no new privileged workspace service or weaker home permissions.
2. Optional object sharing would need published authorised-object provenance, access and retention rules.
3. Host activation would need joint publication/qualification of session lifecycle authorisation member format, session permit, host_start/host_stop transport, successor absence/idempotency, per-start interlocks and cancellation/crash reconciliation.
4. Upstream fate-table unit configurations, writer fencing and platform qualification remain required; this repository only references their owner.
5. Upstream time-policy implementation and qualification remain required; this repository publishes the accepted reference rules only.
6. Existing FD3 mechanism/intake/digest/framing/deadline, helper signatures, shared launch library, consumed-plan/exit transport, post-exec renewal/release ownership, numeric exits/signal encoding and retirement export schema questions remain open. No formats were invented to close them.
