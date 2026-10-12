# TASK-261012-2uhpzh — apply-review-findings-lr: gap report

Docs-only revision of SPEC.md, README.md, diagrams/README.md and all four
PlantUML sources. The instruction brief's three items are applied to the
current text; F-1 is also applied because the acceptance criteria explicitly
require that owner decision. No code, tests, dependency versions or shipped
capabilities change. Work is left uncommitted for review.

| Finding / brief item | Changed sections and figures | Disposition and remaining publication work |
| --- | --- | --- |
| F-6 / R-FD3 / item 1 | SPEC §10.1 (including bounds and publication questions), §10.2 reference; README target one-shot section; components.puml, launch-paths.puml, launch-states.puml, artifact-layout.puml; diagrams/README.md | Removed the close-all/new-pipe sequence. The user manager's spec §7 and diagrams/launch.puml own the single descriptor lifecycle, including trusted-handle retention and failure cleanup. Executor intake is only caller-owned 0–2 plus read-only FD3, with no service, broker or root handle. Completely staged, bounded, immutable bytes are preferred; any pipe needs published/qualified capacity, feeder, deadline and cleanup order. The 48 KiB plan and 80 KiB privileged-request bounds remain. Intake, digest, framing and read-deadline gates remain; joint mechanism publication and maximum-size/failure/concurrency qualification are added before activation. Concrete per-platform staging, intake schema/digest placement, exact refusal names and approved entry argv stay open. |
| F-4 / R-START / item 2 | SPEC §§10.2–10.3 (existing first-consumer reference retained), §10.5; README target one-shot section; all four figures; diagrams/README.md | All launch-ticket, redemption and helper-host-role references are removed. Only the dispatcher invokes the privileged launcher. host_start in dispatch-op/1 uses exactly (logical_session_id, execution_id, controller_epoch, op_id) and the durable host admission permit. Fence changes withdraw/reconcile; cancel before call prevents it; ambiguous call resolves the same effect/receipt, stops if required and never launches twice. Host failure retains the claim/closed fence; runner's fresh authenticated channel, current-epoch registration and readiness precede delivery. Controller-only re-adoption keeps execution_id; replacement uses a new one; fresh runner's native session-lock check stays. Joint authentication/transport, permit, fence/cancel/crash and receipt qualification remain activation work; SH2 stays disabled. |
| F-5 / item 3 | SPEC §§10.1–10.2, §10.4; README target one-shot section; all four figures; diagrams/README.md | Removed the stdin command-request row and deferred-daemon/detached-supervisor references. Dispatcher is an unprivileged Unix-socket service from its first protected-launch slice, using broker connection/framing rules. Recovery/readiness precede admission; unavailable/unready service refuses without command fallback. Caller signatures, peer/server identity checks and descriptor restrictions remain. Existing 128 KiB dispatcher service-frame and 64 KiB broker-frame bounds remain; obsolete 96 KiB command intake is removed. v0 still names the local-headless capability slice, not a command transport. Queueing/events/long waits/advanced integrations remain later gates on the same transport. |
| F-1 / acceptance criterion | SPEC §10.4 and its export publication question; README target cleanup description; all four figures; diagrams/README.md | Ordinary retirement deletes the home after quiesce, confirmed process absence and declared-key cleanup. Export is service-generated closed-schema metadata with no agent-controlled strings, names, links or bodies. Separate explicit operator home retention states confidentiality is not guaranteed. Code/results keep their separate git/acceptance policy without a secret-free assurance. Closed export schema/provenance checks remain publication work owned by the user manager. |

A dated revision note is added at the top of SPEC.md in its existing style.
D-EXECUTOR, D-IDENTITY and the existing receipt/effect roles are preserved.
R-FD3 and R-START names are used without changes; no new wire members,
signature grammar, refusal codes or numeric values are invented. Other
findings/resolutions are outside this launcher's assigned review wave.

Other inherited open items remain: helper v2 signature/version/canonical
bytes; consumed-plan/exit evidence and renewal ownership across exec;
numeric refusal exits/signal encoding; shared launch library publication;
Codex/Muse lane qualification; donor variant B; upstream amendment activation.
None is represented as implemented or qualified.

## Evidence and limits

- Inspected the documentation diff and all changed figure sources. The scope
  is seven documentation files. The current interactive FD3 terminal/FD4
  status transport is kept separate from target plan intake.
- Searched SPEC.md, README.md and diagrams for superseded ticket/redemption,
  host-role, command intake, detached-supervisor, deferred-daemon and close-all
  wording, and for personal local paths/session/board-resource links. No such
  target references remain; public installation paths are retained.
- No tests or builds were run by the developer, as gw-LR3-brief.md explicitly
  prohibits them. PlantUML was not rendered: this host has no renderer, and
  the brief requires conservative source-only edits. Syntax was inspected;
  rendered layout remains unverified.
- The configured CR validation is reserved for task-board handoff. Its real
  commands, exit codes and outcome belong to the handoff-generated validation
  evidence; this report does not preclaim a green gate or runtime qualification.
- LOGBOOK.md was not written, as explicitly required by the current brief.
  Relevant decisions and remaining gaps are recorded here and in board notes.

## Handoff preflight correction

The first `task-board handoff TASK-261012-2uhpzh --role developer` exited 1
before CR validation because checklist item 6 still required logbook evidence.
The current brief explicitly says never to write LOGBOOK.md. The generic
conditional logbook item is therefore replaced with a task-specific item:
findings/decisions/gaps are recorded in this outcome and board notes, and
LOGBOOK.md remains untouched. This corrects the checklist to the authorised
scope; it does not claim logbook evidence exists. Handoff will be retried.
