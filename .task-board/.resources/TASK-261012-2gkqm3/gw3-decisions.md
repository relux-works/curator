# Review-findings wave (2026-10-12): owner decisions and orchestrator resolutions

The precondition `gw3-findings.md` holds the eleven findings F-1 … F-11 from the architecture review. Each finding names its problem, evidence, proposed change, risks and affected sections. Apply the parts of each finding that fall in your repository.

## Owner decision (operator, 2026-10-12): apply exactly

The operator accepted all eleven findings («все принимаем»). Three of them change earlier owner decisions:

- **F-1 supersedes the 2026-10-10 home-archive policy.** At retirement, after quiesce and declared-key cleanup, the agent's home is **deleted**. Only service-generated metadata in a closed schema is exported, with no agent-controlled strings, names, links or bodies. Keeping a home is a separate, explicit operator operation that states plainly that its confidentiality is not guaranteed. The morning wave's positive-classification rules for the archive are removed, not kept as an alternative.
- **F-4: only the dispatcher performs privileged starts.** Launch tickets are removed entirely: the ticket signature and schema, the helper's host role, host sudo authority, the redemption journal and the "launcher redeems the ticket" resolution (old R1). The session host asks the dispatcher for a start. D-EXECUTOR, D-IDENTITY and the standing-runtime decision are unchanged.
- **F-5: the dispatcher is a Unix-socket service from its first slice**, using the broker's connection and framing rules. The command mode, caller sudo rules for the dispatcher, the `SUDO_UID` branch of the dispatcher, the detached-command bootstrap, the external sweep timer and the command-to-daemon conversion are removed. Queueing, events, long waits and advanced integrations come later on the same transport, behind capability gates.

The other eight findings (F-2, F-3, F-6, F-7, F-8, F-9, F-10, F-11) are contract gaps to fix.

## Orchestrator resolutions: shared names, so that parallel edits agree

Apply these names exactly. Where a resolution fixes only a name and its meaning, publish the rest at the same level of detail as the neighbouring operations in your spec. Anything your spec family treats as unpublished (signature domains, canonical bytes, exact refusal codes) stays unpublished: list it under open questions as "would need". Do not invent refusal codes.

- **R-OBS (F-10), effect observation.** The user manager adds the operation `effect.observe` to `helper-op/1`. The broker adds the method `effect.observe` to `broker/1`. Both take the original effect's `request_id` (the effect ID), the operation and a digest of its canonical arguments. Only the original caller principal (and operators) may ask, and the read is serialized with the receiver's journal. Both answer with exactly one of six observation states:
  - `recorded`, with the recorded result and the current target state;
  - `pending`;
  - `rolled_back`;
  - `not_recorded`: complete authoritative evidence that the effect was never recorded;
  - `unavailable`: evidence missing, expired or unreadable; never treated as absence;
  - `contradictory`.

  The dispatcher (§§4.6, 7) maps them to its four observation classes:
  - `recorded` maps to present or absent, according to the current state;
  - `rolled_back` and `not_recorded` map to known absent for that effect only; this never authorizes an automatic new create;
  - `pending` and `unavailable` map to unknown;
  - `contradictory` maps to contradictory.

  The broker also publishes the matching unbind observation and its removal/tombstone retention.
- **R-VERDICT (F-8), revocation verdict.** The broker adds the method `revocation.check` to `broker/1`, accepted from configured dispatchers. It takes the exact chain or chains being checked and the check context (`prepare`, `dequeue` or `start`). It answers `not_revoked`, `revoked` or `unavailable`, together with the broker's revocation sequence at the time of the check. The answer is point-in-time evidence, never a grant or a permit for a future start. While a restore is fenced (F-9), the answer is `unavailable`. The dispatcher refuses on `revoked` and on `unavailable`. The final bind, lease and renewal checks in the broker stay mandatory.
- **R-START (F-4), host start request.** The dispatcher adds the operation `host_start` to `dispatch-op/1`, allowed only to the configured session-host service principal. It carries `(logical_session_id, execution_id, controller_epoch, op_id)` and refers to a **host admission permit**: a durable dispatcher record bound to those four values. The permit is created when a start with the session-host executor is admitted, and a fence change withdraws it or sends it to reconciliation. The dispatcher starts the approved agent-side runner through its existing claim, effect and root-receipt path. Cancellation before the call prevents the call. After an ambiguous call, the dispatcher resolves the same effect and receipt and stops the execution if required; it never launches twice. The runner then opens a fresh authenticated channel to the host. Registration in the current epoch and readiness come before delivery.
- **R-DELEG (F-3), orchestrator delegation.** The requester C signs the caller delegation for its new orchestrator O, narrowed from C's own grant, after O's key is registered under its generation. The dispatcher verifies and records it but never signs it. Registration (`register_key`) is separate from activating caller authority. Until a valid delegation is installed, O has no spawn authority. The pre-key registration exception, its replay and rotation rules, and the depth/chain-length preflight are published in the dispatcher spec. Recursive orchestrator provisioning stays refused until this bootstrap is published and qualified; direct worker launches and operator-enrolled orchestrators form the first slice.
- **R-MAINT (F-2), maintenance stop.** The dispatcher's operator activation transaction (§11) gains a bounded maintenance attempt:
  1. close new admission;
  2. inventory existing executions;
  3. stop the selected executions with explicit operator authority, through the dispatcher and `exec.stop`;
  4. require confirmed process absence and reconciliation before any change of mode, state migration or binary swap.

  The maintenance deadline bounds the attempt, never a standing execution's lifetime. An unknown or refused stop defers the maintenance. Standing allocations and identities are kept. A restart needs fresh authority. There is no live adoption and no automatic expiry.
- **R-FD3 (F-6), launcher descriptor lifecycle.** The user manager's spec (§7 and `diagrams/launch.puml`) owns the one descriptor-lifecycle contract. The agent launcher's spec (§10.1) references it and does not restate a different order. The required order:
  1. scrub untrusted inherited descriptors first;
  2. keep the generation lock until the root receipt and the completed UID/GID transition;
  3. keep the verified executable handle until exec;
  4. close internal handles at their last use, or by close-on-exec, and release them on every failure;
  5. the executor receives only 0–2 and a read-only FD3.

  The preferred plan staging is completely staged, bounded, immutable bytes on FD3. The concrete per-platform mechanism stays an open question with that requirement. A pipe is allowed only with a published capacity/feeder/deadline/cleanup order.
- **R-HOLD-OLD.** Old resolution R1 (the launcher redeems the ticket) is withdrawn by F-4. Old resolution R2 (a fresh runner checks the native session lock) stays.
- **Scope.** Docs only: spec text, README, CIP text and PlantUML diagrams. Update every figure that shows a changed behaviour. Add a revision note at the top of each changed spec in its existing style.
