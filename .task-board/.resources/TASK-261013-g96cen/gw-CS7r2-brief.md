# THE ONLY CURRENT INSTRUCTION — TASK-261013-g96cen rework (revision 2): fix B1 and the notes from review (W-CS7r2, developer)

Reviewer B requested changes on revision 1: `TASK-261013-g96cen_review-verdict-rev1.md`, which reconciles reviewer A's A1 and A2. Fix exactly these items. Everything else in revision 1 stands.

**Do this:**
1. **B1 (P1).** Revert the "agents-management Decision 5" substitutions in `decisions/0010-…`, `0013-…` and `0019-…`.
   - These are in-document references:
     - in 0010 and 0013, "Decision 5" is that document's own §5;
     - in 0019, it is Decision 0013's §5.
   - Where you want to qualify a reference, write "this document's Decision 5" (0010, 0013) or "Decision 0013 Decision 5" (0019). Otherwise leave the bare citations unchanged and report them in the gaps file as unambiguous.
   - Keep the Decision 0005 retitle.
   - Correct or remove the three revision notes, and correct the false gaps statement.
2. **B2.** Replace the two normative uses of the audit label `R-LR1` with the launcher-intake owner link, as the disposition already does:
   - CIP-0010, "The intake release/grammar belongs to R-LR1 and Decision 0021's amendment";
   - CIP-0011, "Intake schema/version publication and qualification remain R-LR1 work".
3. **B3.** In Decision 0021, add a scoped cross-reference after the blanket naming-rule sentence, pointing to the disposition's D-WIRE-NAMES scope sentence.
4. **B4.** Give the remaining Muse sentences the same three-way wording: priority decided on 2026-10-13, channel (b) with fallback (c) proposed, qualification open. The sentences are in CIP-0003 ("Muse subscription support remain owner questions") and in the CIP-0010 summary line.
5. **B5.** The shared typed index now exists: wiki `naming.md`, section "Shared index: rules and wire names". Make the disposition's D-WIRE-NAMES paragraph say it quotes the dispatcher §3.3 scope sentence and that the index lives there. It is not a new owner decision.
6. **A1 and A2.**
   - Qualify `donor-layout.puml:84` ("fresh materialization each round") and `donor-lifecycle.puml:88` ("fresh B/C materialization") with "except an admitted repair round".
   - In CIP-0009 D5a's Export outcomes paragraph and in the transfer figure, state that the contradiction comparison applies to the same unchanged quiesced tree, not across an admitted repair.

**Rules:** as in `gw-CS7-brief.md` and `gw6-decisions.md`. Never edit mechanically; every passage outside these fixes stays byte-identical to revision 1. Docs only; no tests or builds; never write LOGBOOK.md. Public hygiene: no F-numbers, record slugs, host names or local paths.

**Hand-off:**
1. Update `TASK-261013-g96cen_gaps.md`.
2. Run `task-board handoff TASK-261013-g96cen --role developer`.
3. END YOUR TURN.
