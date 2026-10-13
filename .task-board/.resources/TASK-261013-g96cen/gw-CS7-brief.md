# THE ONLY CURRENT INSTRUCTION — TASK-261013-g96cen: wave 6 for the curator-spec (CIPs, the platform contract disposition and the remote-worker register): accepted architecture-review findings (W-CS7, developer)

**Repository:** the curator-spec (CIPs, the platform contract disposition and the remote-worker register) module. Main carries waves 3–5b. The preconditions are:
- the CURRENT `gw6-decisions.md`, which holds the owner decision and the bounds of this wave;
- the attached finding records `f-<slug>.md`, one per finding below;
- `gw5-decisions.md`, `gw4b-decisions.md` and `gw4-decisions.md`, which hold the shared names;
- this brief.

**Do this.** For each finding, apply the record's *Proposed change* in this module's sections. Make the minimal change that meets the requirement and keeps it within the bound given:

- **F-65, record `f-r-name-namespace.md`.**
  - First consumer: Contract/docs; P0 and each later freeze.
  - Required: One typed owner-linked index, scoped names/domains/discriminators; no blanket rewrite.
  - Affected specs, as the record names them; apply only this module's part: `swarma-credential-broker/spec/broker.md` (§3.1, §18 preamble), `swarma-dispatcher/spec/dispatcher.md` (revision notes, §§5.3, 12, 16), `swarma-user-manager/spec/helper.md` (§7), `swarma-session-host/spec/session-host.md` (§5 traceability table), `curator-spec/cips/platform-contract-disposition.md`, `curator-spec/decisions/` (0005 duplication), and the new register home (`wiki/naming.md` or `swarm-platform-architecture` trust contract).
- **F-63, record `f-wire-register-drift.md`.**
  - First consumer: Owning modules/docs; P0 and each freeze.
  - Required: Repair owner/status references; distinguish draft result bytes from complete transport/lifecycle; removed tickets stay removed.
  - Affected specs, as the record names them; apply only this module's part: `wiki/naming.md`, `wiki/glossary.md`; `curator-spec` platform contract disposition and remote-worker publication register; dispatcher §16, user-manager §14, broker §18, session-host §7, launcher §10.5 (status rows become index references); session-host §3.3 and launcher §10.5 wording.
- **F-48, record `f-book-muse-c7.md`.**
  - First consumer: Provider/docs; P0 status and later C7 activation.
  - Required: Record decided priority, proposed channel and open qualification separately.
  - Affected specs, as the record names them; apply only this module's part: Wiki book chapters 10 and 14; curator-spec CIP-0010 open-questions register (one sentence). Broker `spec/broker.md` §§10.2/18.5 is the authority; no change proposed there; curator-agent-launcher `SPEC.md` open-questions table; swarma-session-host `spec/session-host.md` open questions.
- **F-31, record `f-exec-input-closure.md`.**
  - First consumer: Helper/launcher/qualification; first profile at S0a.
  - Required: Qualify full invocation and interpreted inputs; worker-code execution remains R-VERIFY.
  - Affected specs, as the record names them; apply only this module's part: User manager R-EXEC/R-VERIFY and qualification; dispatcher task profiles and qualification; session-host light checks; CIP-0009 D5c and remote-worker publication/qualification register; launcher §10.2 exec profiles; the future language/tool profiles and verification contract.
- **F-62, record `f-base-binding.md`.**
  - First consumer: Transfer/launcher/board; W then T import.
  - Required: Exact archive-to-B manifest before work and independent repository-B recomputation at import.
  - Affected specs, as the record names them; apply only this module's part: - dispatcher §§4.2, 5.2, 5.2.1, 15, 16 question 22; - user manager §§7, 12, 14 question 15; - agent launcher §10.1 workspace section; - spm PR #604 §§3–4 and the import leaf; - CIP-0009 D5a steps 1 and 5 (cross-reference).
- **F-43, record `f-export-refusal-pins.md`.**
  - First consumer: Dispatcher/transfer/ops; W.
  - Required: Four outcome classes, permission/failure distinction, immutable refusal, retained obligations and current-authority bounded repair/per-act waiver.
  - Affected specs, as the record names them; apply only this module's part: - dispatcher §§3.2, 5.2, 5.2.1, 5.4, 7.1, 15, 16 question 22 and `diagrams/workspace-transfer.puml`; - user manager §§4.4, 7, 12, 14 question 15; - CIP-0009 D5a steps 4 and 7; - session host §3.1 (round boundary), cross-reference; - spm PR #604 §3 flow and diagram; - book chapter 11 code-path section.
  - Bound for this wave: The standing discard of refused exports is an optional owner policy that has NOT been chosen. Default: the obligation is retained, with bounded repair under current authority or a per-act waiver.
- **F-42, record `f-export-cleanup-authority.md`.**
  - First consumer: Dispatcher/helper/transfer; W.
  - Required: Bounded original-admission cleanup consent, distinct execution evidence, all-blob custody outside deletion and authenticated retained-content release.
  - Affected specs, as the record names them; apply only this module's part: dispatcher §§3.4, 4.3–4.5, 5.2–5.4, 7–9, 14–16 and workspace/reconciliation figures; user manager §§4.4, 4.7, 5, 7 and launch/retirement figures; session-host transfer cross-reference; CIP-0009 D5a; board import contract's export/recovery handoff.
- **F-57, record `f-workspace-result-modes.md`.**
  - First consumer: Transfer/board; W representation and T import.
  - Required: Path/kind/mode/content semantics include mode-only edits, symlink transitions and deletion prior mode.
  - Affected specs, as the record names them; apply only this module's part: dispatcher §§5.2, 5.2.1, 16 question 22 and `schemas/workspace-result-1.schema.json` (no field change); user manager §7 and §14 question 15; launcher §10.1–10.2; session host §3.1; `curator-spec` `protocol/remote-worker.md` and CIP-0009 D5a; board import draft (spm #604) manifest baseline.
  - Bound for this wave: Keep the field set of `schemas/workspace-result-1.schema.json` unchanged; the change is semantics and text.

**Rules:**
- **Private sources, public repository.** The records are private review material. Never copy their seat names, event ids, local paths, host names or the review repository into this repository. Cite only module sections and shared rule names. Revision notes say "accepted architecture-review findings (2026-10-13)", with no F-numbers or record slugs.
- **Never edit mechanically.** Do not strip patterns with a regex: `F-<digits>` also matches inside `UTF-8`. The base text holds no review identifiers for you to remove, so simply add none. Every passage outside your findings' scope stays byte-identical.
- **Do not widen the scope.** Optional owner policies stay off and unchosen. No implementation or qualification is claimed. Do not invent wire fields, refusal codes or values; where the source leaves a format unpublished, say so and add an open question.
- **Keep the module coherent.** Remove superseded text, update every figure that shows changed behaviour, and add a revision note.
- **Docs only.** No tests or builds. Never write LOGBOOK.md.
- **Public hygiene.** No first names, host names, local absolute paths, session links or board-resource links.

**Hand-off:**
1. Attach `TASK-261013-g96cen_gaps.md`. It covers every finding: what changed, with section numbers, or why nothing was needed. It also lists what stays open.
2. Run `task-board handoff TASK-261013-g96cen --role developer`.
3. END YOUR TURN.
