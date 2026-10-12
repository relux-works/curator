# Spec-gap wave (2026-10-12): owner decisions and orchestrator resolutions

## Owner decisions (operator, 2026-10-12): apply exactly
- **Standing team allocations do not expire.** This covers allocations of the standing team: the orchestrator and fixed team roles. Only the operator retires or re-creates them, with a signed record. This matches D-IDENTITY: account retirement never revokes an identity, and standing identities are revoked only by the operator.
- **"Bring up team" operation.** The dispatcher gains an operation that creates the standing team's allocations from a description the operator signs. The exact wire is the dispatcher's to publish under D-WIRE-NAMES (`swarma-dispatcher <object>/<n>`).
- **Spec-gap wave approved.** Docs only. Highest priority: secrets in the user manager's home archive.

## Orchestrator resolutions: apply them; report a contradiction with the spec instead of overriding it
- **R1. Who redeems the launch ticket.** The ticket is redeemed by the user-manager module, by its launcher binary `swarma-user-launcher`, under the module's shared lock and journal. The session host presents the ticket. The dispatcher only issues it. Today dispatcher §12 says "the launcher accepts". Session-host §3.3 and `session-recovery.puml` say "the user manager redeems". Both must read the same way. The signature, schema and host role stay unpublished (gate I4).
- **R2. Who confirms that the native session lock is absent.** `exec.stop` (user manager §4.7) proves process absence only; it is harness-agnostic. Before the host reopens or resumes, a freshly launched runner under the agent account checks the harness's native session lock, under a closed fence, and reports the result to the host. An unknown result keeps the session parked.
- **R3. Execution runtime of standing sessions.** The 12-hour maximum execution runtime (dispatcher §3.5) does not fit long-lived standing sessions held by the session host. Record this as an open question with a proposal: executions of standing allocations held by the SH2+ backend have no runtime maximum and are stopped only by the operator or by an explicit stop. Do not decide it.
