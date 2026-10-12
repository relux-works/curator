# THE ONLY CURRENT INSTRUCTION — TASK-261012-2gkqm3: rework after the review (W-CS3, revision 2, developer)

Reviewer B requested changes. Read `TASK-261012-2gkqm3_review-verdict-rev1.md`. Fix exactly its one P1 (D1), and nothing else:

1. In `cips/diagrams/join-handshake.puml`, around line 72, the note still lists connection-scoped ids and op_key caches as the replay defence. Rewrite that replay line to match CIP-0009 D5/D6 and `remote-worker-turn.puml`: later frame effects rely on durable binding-generation operation-identity/digest claims with authenticated admission floors (D5), not on connection IDs or the output cache. Keep the existing tombstone and pending-format wording. Do not invent fields or codes.
2. Add a one-line revision comment in the file's existing style.
3. Add `join-handshake.puml` to the F-7 figure list in `TASK-261012-2gkqm3_gaps.md`, update that resource, and correct the sentence that said no contradicting copy remains.
4. Sweep once more for figures that restate the retired claim (`op_key`, `unknown_result`, "cached result", "connection-scoped") and fix any you find in the same way.

Rules as in `gw-CS3-brief.md`: docs only, public hygiene, no tests or builds yourself, never LOGBOOK.md.

**Hand-off:** republish the Change Request as revision 2, then `task-board handoff TASK-261012-2gkqm3 --role developer`, and END YOUR TURN.
