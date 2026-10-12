# TASK-261012-2uuu3e — reviewer B verdict (deciding review), CR-TASK-261012-2uuu3e-1 revision 1

Verdict: **ACCEPTED**. No P0/P1 remains. Base 601cbb0, candidate tree 9c1b487 (15 files, docs only). Checks 1-7 were done on the exact candidate delta before `gw-CS4-review-A.md` was read; A was reconciled afterwards.

## Checks
1. **Brief items.** All four items are applied, and none is falsely reported as closed.
   - F-19 / R-WORKPLACE: CIP-0009 D4.7, D5, D12, implementation step 2, test plan and questions 13-14. `swarma-rb-exec` runs under a service principal. The `workplace` class has a root-owned relay-only fixed shell and no startup files. Claims, floors, audit, mailbox and CA custody sit outside the workplace UID. Both wire names are UNPUBLISHED in the `protocol/remote-worker.md` register.
   - F-15: CIP-0011 §4 splits recovery/control readiness from forward admission. Slice 0 and the test plan agree.
   - F-27 / R-NETBIND: CIP-0011 §6 cites N15 as owner, marks `[bindings.profiles]` as replaced and restates none of N15's details. The N15 link resolves (HTTP 200).
   - F-14, F-21, F-24: E-rebind, the time watermark and service egress follow the resolutions.
2. **Owner decisions.**
   - F-16: standalone repository across accounts; same-account worktrees stay.
   - F-20: session lifecycle authorisation, session-scoped host admission permit, per-physical-start claim/fence/receipt, `host_stop` (configured host only, journaled `exec.stop`, host-requested, no start authority), successor idempotent per `(logical_session_id, op_id)` after confirmed previous absence.
   - F-21: authenticated or operator-controlled time source; watermark as `max(now, watermark)` for expiry and `min(now, watermark)` for `not_before`; operator-only reset; boot-inclusive monotonic elapsed time before destructive expiry.
   - F-1/F-4/F-5 hold: confirmed absence plus key cleanup before home deletion, dispatcher-only physical starts, dispatcher Unix-socket service from slice 0. The old four-tuple-bound permit text and the same-UID executor text are gone.
3. **Names.** R-SESSION, R-REBIND, R-TIME, R-WORKPLACE, R-SVC-EGRESS, R-NETBIND, R-FATE (referenced only), R-OBS, R-VERDICT and R-DELEG are used with the assigned meanings. The only refusal code used is the existing `lease_network_enforcement_unavailable`. No wire field, enum value or new refusal code is invented. Unpublished formats are named as such in CIP-0009 questions 13-14 and CIP-0011 questions 4 and 6-8.
4. **Security properties.**
   - Fail-closed on unknown: an unknown permit refuses successor starts, unknown records retain claims, and time-untrusted admission refuses.
   - No automatic retry after an ambiguous effect: unknown claims are never replayed, an ambiguous `host_start` resolves the same effect/receipt, and the runner shape retains affected claims as unknown.
   - Ancestor revocation reaches descendants: the caller ancestry/dependency is kept, owner chain revocation withdraws the permit, and rebind re-signs under the current chains.
   - Point-in-time verdicts are not grants: every successor and every rebind needs a fresh `revocation.check` (`start` / `prepare`).
5. **Figures.** All ten changed sources were read, and all ten are linked in the diagrams README. They agree with the text: relay-only shell to the service principal, dispatcher/launcher/`exec.stop`, the helper-v1 gate for every variant, session authorisation and `host_stop`, and the time/network/egress notes. Source review only; no renderer on this host. Syntax checked by eye.
6. **Gaps file.** The eight touched findings, with sections and figures, are accurate against the diff. It claims no runtime qualification.
7. **Hygiene.** A scan of the added lines found no local paths, host names, session or board links, or personal names. The only URL added is the public N15 spec link.
- Validation log (board-managed): exit 0, 689 tests OK, `exact_command_shard` coverage 1/1. Test-case coverage is reported as unknown and not inferred. Not rerun here.

## Reconciliation with A
A reports one note (N1, publication and qualification stay pending). I agree with it, and it is not a defect. A raised no P0/P1.

## Notes (P2, tb-R226; not blocking)
- CIP-0011 rebind: the F-14 "renew ahead" sweep is not stated. Only the horizon reporting and explicit renewal are. The sentence that ancestor revocation still applies to rebound orchestrator descendants is implicit, via the re-signed current chains and the retained-ancestry text elsewhere. It could be made explicit at the next edit.
- CIP-0011 session lifecycle: the F-20 text says the owner authorisation "carries" a restart budget. The CIP says it "would need" an anchored or cumulative budget, with the format unpublished. This is a softening, and the test plan still tests it.
- CIP-0009 lines 45-47 (review summaries F14-F16) and 226-227 still use "pending" and "proven descendant cleanup" wording. The register now says UNPUBLISHED and the body says confirmed execution absence. This is a cosmetic mismatch, not a contradiction of the contract.
- All renderer and runtime qualification stays open, as the gaps file states.
