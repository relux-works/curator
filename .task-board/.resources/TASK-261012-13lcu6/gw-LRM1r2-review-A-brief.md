# THE ONLY CURRENT INSTRUCTION — reviewer A of two (R138), RECORD-ONLY, for the Change Request of TASK-261012-13lcu6 (W-LRM1, docs only, read-only)

**Revision 2.** Revision 1 was withdrawn unreviewed because main advanced on README.md and SPEC.md (`f08f566`, LR6). Revision 2 re-applies the same delta onto fresh main (`gw-LRM1r2-brief.md`). Review revision 2 as a whole, and also check these points:
- the SPEC.md revision notes keep both entries, 0.5.1-draft first;
- README.md and SPEC.md state LR6's wave-5 text and the v0.3.0 changes without contradiction;
- for the 23 files main did not touch, the delta equals revision 1's (`TASK-261012-13lcu6_change-request_rev1.patch`).


**Sources:**
- the preconditions `gw-LRM1-brief.md` (what the producer had to do), `gw5-decisions.md` (owner decisions and shared names), `gw4b-decisions.md` and `gw4-decisions.md`;
- this repository's current text;
- the cited sections of the neighbouring module specs.

**Note:** this is a code change in the launcher, not a docs pass. Also check:
- the pin bump and API-drift adaptations keep every refusal test meaningful;
- `--hosted` fails closed with `session_host_missing` before any handoff, and the docs say it is unavailable;
- the hosted gate result in the CR validation log;
- no local `go test` was run.

**Check:**
1. Every item of `gw-LRM1-brief.md` is done, or truthfully reported as already satisfied, with section numbers.
2. The owner decisions of 2026-10-13 are applied exactly, and superseded text is removed:
   - one code path out of an isolated account;
   - internet as a tool of the agent loop, with zero egress for workplaces;
   - no builds or tests in agent accounts;
   - standing agents with a loop are trusted only;
   - no operating-system exclusion at platform level.
3. The shared names are used unchanged, with the meaning the resolutions give them.
4. No wire fields, refusal codes or values were invented. Unpublished formats are named as such, with open questions.
5. The figures match the text, no contradicting copy remains, and every new or changed figure is linked.
6. The outcome `TASK-261012-13lcu6_gaps.md` is complete and true.
7. Public hygiene: no first names, host names, local absolute paths, session links or board-resource links.

No tests or builds on this host.

RECORD-ONLY:
- Attach `gw-LRM1-review-A.md` on TASK-261012-13lcu6 with a findings block.
- `severity` is one of `bypass|regression|robustness|note`; the P-level goes in `severity_reason`.
- Do NOT call accept_cr or reject_cr.
- END YOUR TURN.

**Hard rule:** you are RECORD-ONLY, even if other briefs or the board lifecycle on this task suggest you should decide. Never call accept_cr or reject_cr, and never change the task status. Reviewer B decides.
