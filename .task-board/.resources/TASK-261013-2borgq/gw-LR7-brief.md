# THE ONLY CURRENT INSTRUCTION — TASK-261013-2borgq: wave 6 for the agent launcher: accepted architecture-review findings (W-LR7, developer)

**Repository:** the agent launcher module. Main carries waves 3–5b. The preconditions are:
- the CURRENT `gw6-decisions.md`, which holds the owner decision and the bounds of this wave;
- the attached finding records `f-<slug>.md`, one per finding below;
- `gw5-decisions.md`, `gw4b-decisions.md` and `gw4-decisions.md`, which hold the shared names;
- this brief.

**Do this.** For each finding, apply the record's *Proposed change* in this module's sections. Make the minimal change that meets the requirement and keeps it within the bound given:

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
- **F-44, record `f-exec-profile-authority.md`.**
  - First consumer: Protected policy/helper/dispatcher; S0a first profile.
  - Required: Machine-owned selection and immutable resolved body pin at prepare/start/retry.
  - Affected specs, as the record names them; apply only this module's part: Dispatcher §§3.4, 4.2, 5.1, 5.3, 16 q25; user manager §§7, 8, 11, 14 q19; launcher §10 composed-plan section; session host §3.2; book chapter 11.
- **F-62, record `f-base-binding.md`.**
  - First consumer: Transfer/launcher/board; W then T import.
  - Required: Exact archive-to-B manifest before work and independent repository-B recomputation at import.
  - Affected specs, as the record names them; apply only this module's part: - dispatcher §§4.2, 5.2, 5.2.1, 15, 16 question 22; - user manager §§7, 12, 14 question 15; - agent launcher §10.1 workspace section; - spm PR #604 §§3–4 and the import leaf; - CIP-0009 D5a steps 1 and 5 (cross-reference).
- **F-57, record `f-workspace-result-modes.md`.**
  - First consumer: Transfer/board; W representation and T import.
  - Required: Path/kind/mode/content semantics include mode-only edits, symlink transitions and deletion prior mode.
  - Affected specs, as the record names them; apply only this module's part: dispatcher §§5.2, 5.2.1, 16 question 22 and `schemas/workspace-result-1.schema.json` (no field change); user manager §7 and §14 question 15; launcher §10.1–10.2; session host §3.1; `curator-spec` `protocol/remote-worker.md` and CIP-0009 D5a; board import draft (spm #604) manifest baseline.
  - Bound for this wave: Keep the field set of `schemas/workspace-result-1.schema.json` unchanged; the change is semantics and text.

**Also, from the v0.3.0 landing (1438a28a, tag v0.3.0):**
- CHANGELOG: change `## 0.3.0 (pending tag)` to `## 0.3.0 — 2026-10-13`, and open an empty `## Unreleased` above it, in the v0.2.0 form.
- Reviewer B's P2 note N1: fix the stale in-repo version references after the 0.5.1 bump:
  - README.md line 5 link text "SPEC 0.5.0-draft";
  - the §10 preamble sentence that says the executable's specification version remains 0.5.0-draft;
  - §4.2 and §4.5, which cite agents-management v0.5.48 where v0.5.62 is now pinned.
- Reviewer B's P2 note N2: `diagrams/components.puml`, `launch-paths.puml`, `launch-states.puml`, `artifact-layout.puml` and `diagrams/README.md` still show the interactive hosted handoff as current. Add the v0.3.0 "hosted launch unavailable until the swarma dispatcher path ships" state.
- Reviewer B's P2 note N3 is a code string, not docs. The `session_host_missing` detail text says "task-board receiver is not available on PATH". Do not change that code string in this pass; record N3 in the gaps file as a code follow-up.

- **The release-pin test moves with the docs.** `cmd/curator-run/main_test.go` `TestReleaseVersionPinned` (about lines 400–418) pins README and CHANGELOG strings to the pre-tag state:
  - the `go install …/cmd/curator-run@v0.2.0` line;
  - the sentence "The v0.3.0 candidate (pending tag) reports …";
  - `## 0.3.0 (pending tag)`.

  Update README and CHANGELOG to the released v0.3.0 state, and update exactly those pinned entries in that test to match. That is the only code edit allowed in this pass. Keep `wantBuild = "0.3.0"` and the spec version pins unchanged. Revision 1 failed the hosted gate on this test alone.

**Rules:**
- **Private sources, public repository.** The records are private review material. Never copy their seat names, event ids, local paths, host names or the review repository into this repository. Cite only module sections and shared rule names. Revision notes say "accepted architecture-review findings (2026-10-13)", with no F-numbers or record slugs.
- **Never edit mechanically.** Do not strip patterns with a regex: `F-<digits>` also matches inside `UTF-8`. The base text holds no review identifiers for you to remove, so simply add none. Every passage outside your findings' scope stays byte-identical.
- **Do not widen the scope.** Optional owner policies stay off and unchosen. No implementation or qualification is claimed. Do not invent wire fields, refusal codes or values; where the source leaves a format unpublished, say so and add an open question.
- **Keep the module coherent.** Remove superseded text, update every figure that shows changed behaviour, and add a revision note.
- **Docs only.** No tests or builds. Never write LOGBOOK.md.
- **Public hygiene.** No first names, host names, local absolute paths, session links or board-resource links.

**Hand-off:**
1. Attach `TASK-261013-2borgq_gaps.md`. It covers every finding: what changed, with section numbers, or why nothing was needed. It also lists what stays open.
2. Run `task-board handoff TASK-261013-2borgq --role developer`.
3. END YOUR TURN.
