# Review-findings wave 4 (2026-10-12): owner decisions and orchestrator resolutions

The precondition `gw4-findings.md` holds the cycle-2 findings F-12 … F-28. F-26 was withdrawn and F-23 is pending, so neither is part of this wave. Each finding names its problem, evidence, proposed change, risks and affected sections. Apply the parts that fall in your repository.

## Owner decisions (operator, 2026-10-12): apply exactly

- **F-20 accepted.** It refines the earlier owner decision F-4:
  - the owner's admitted start can explicitly authorise successor executions for one logical session;
  - the host admission permit becomes session-scoped, while every physical start keeps its own `(execution_id, controller_epoch, op_id)` claim, fence and receipt interlock;
  - the dispatcher adds `host_stop`;
  - physical starts stay with the dispatcher only.
- **F-16 accepted.** No linked `git worktree` across isolated accounts. A child in another account gets a standalone repository with its own `.git`, configuration and hooks. Worktrees within one account, and today's board workflow, stay.
- **F-21 accepted.** The installer requires an authenticated or operator-controlled time source (for example NTS). Clock steps can only narrow authority. Expiry-driven destruction also needs real elapsed monotonic time.
- **F-12, F-13, F-14, F-15, F-17, F-18, F-19, F-22, F-24, F-25, F-27 and F-28 accepted** as contract gaps to fix.
- The cycle-1 decisions stay, including F-1, F-4 (as refined by F-20) and F-5, and the shared names of wave 3: R-OBS, R-VERDICT, R-START, R-DELEG, R-MAINT and R-FD3.

## Orchestrator resolutions: shared names, so that parallel edits agree

Use these names unchanged. Publish the rest at the level of detail of neighbouring operations. Anything your spec family treats as unpublished stays unpublished, listed as an open question ("would need …"). Do not invent refusal codes.

- **R-SESSION (F-20).**
  - The start member that carries the owner's lifecycle authorisation is named **session lifecycle authorisation**; its member format is unpublished.
  - The permit is the session-scoped **host admission permit** for one `logical_session_id`.
  - The new dispatcher operation is `host_stop` in `dispatch-op/1`. Only the configured session-host principal may call it, and only for an execution under its permit. It goes through the journaled E-stop path (`exec.stop`, confirmed absence), is recorded as host-requested and is distinct from owner `cancel`. It confers no start authority.
  - A successor `host_start` is admissible only under that authorisation. It is idempotent per `(logical_session_id, op_id)`, and the previous claim must be confirmed absent first.
- **R-REBIND (F-14).** A dispatcher effect **E-rebind** on an existing allocation:
  - the dispatcher re-signs the leaf under the current chains;
  - the broker performs a **binding replacement** for the same `(uid, generation)`, with a removal tombstone for the old binding;
  - every rebind gets a fresh `revocation.check` with context `prepare`;
  - it never widens; the renewing principal must already sit in the chain;
  - the account is not re-created.
- **R-FATE (F-18).** One process **fate table** per process class. The user manager's §7 owns it, next to R-FD3; the dispatcher, the session host and the broker reference it. It includes the fence against writers left over from a replaced daemon.
- **R-OUTCOME (F-22).** One outcome table per receiver operation, mapped onto the six R-OBS states, with no new state. The receivers own them: the user manager for create, launch, stop and retire; the broker for bind and unbind. The dispatcher (§§4.5–4.6, 7.1) references them. `exec.stop` continues under its own effect ID; authority-creating retries stay fresh.
- **R-TIME (F-21).**
  - The installer sections name an authenticated or operator-controlled time source.
  - The broker and the dispatcher each keep a durable time **watermark**. Expiry is checked at `max(now, watermark)` and `not_before` at `min(now, watermark)`. The watermark never advances on untrusted time, and resetting it is an explicit operator act.
  - Expiry-driven destruction (home deletion, binding removal) also requires boot-inclusive monotonic elapsed time of at least the lifetime.
- **R-LOCK (F-28).** Each platform service states the same rule: **never hold its state lock across a synchronous call to another platform service**; ownership is tracked by a per-record in-flight marker. Each component publishes its wait/lock graph.
- **R-WORKPLACE (F-19).**
  - A new account class named `workplace`, for the project-side home of a remote worker.
  - The project-side executor `swarma-rb-exec` runs under a service principal, never under the workplace UID.
  - The wire names `swarma-rb/1` and `swarma-workplace-exec/1` stay unpublished.
- **R-SVC-EGRESS (F-24).** Every non-root platform service account has an egress policy (a v1 capability beside the agent rules). The user manager's §8 holds the service-profile rows. Each service states its egress: the broker reaches the network only through the proxy; the dispatcher and the session host have none (deny-all).
- **R-NETBIND (F-27).**
  - The machine owns the policy that binds a child's network profile to its granted Curator profile; the user manager's §8 names its location.
  - The dispatcher pins it at admission.
  - `fw.apply` checks that the listener actually serves the policy whose digest was applied.
  - No new grant dimension is added.
- **R-GEN (F-17).** Account generations are fresh random 128-bit values. Each of the user manager and the dispatcher publishes its own state-loss and restore procedure, fail-closed like the broker's F-9 procedure.
- **Scope.** Docs only: spec text, README, CIP text and PlantUML diagrams. Update every figure that shows a changed behaviour. Add a revision note at the top of each changed spec in its existing style.
