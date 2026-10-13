# THE ONLY CURRENT INSTRUCTION — reviewer B of two (R138, cross-provider), the deciding review of the Change Request of TASK-261012-13lcu6 (W-LRM1, docs only, read-only)

**Sources:**
- the preconditions `gw-LRM1-brief.md`, `gw5-decisions.md`, `gw4b-decisions.md` and `gw4-decisions.md`;
- this repository's current text;
- the cited sections of the neighbouring module specs.

**Note:** this is a code change in the launcher, not a docs pass. Also check:
- the pin bump and API-drift adaptations keep every refusal test meaningful;
- `--hosted` fails closed with `session_host_missing` before any handoff, and the docs say it is unavailable;
- the hosted gate result in the CR validation log;
- no local `go test` was run.

**Check:**
1. Every item of `gw-LRM1-brief.md` is done, or truthfully reported as already satisfied.
2. The owner decisions of 2026-10-13 are applied exactly, and superseded text is removed.
3. The shared names are used unchanged. Nothing beyond them is invented: no wire fields, refusal codes or values. Gaps are named, with open questions.
4. The security properties survive the edit:
   - fail-closed on unknown;
   - no automatic retry after an ambiguous effect;
   - an export is never partial and never bypasses review;
   - acceptance binds to the exact candidate;
   - workplaces have no egress and agent accounts never build or test;
   - narrowing acts never widen authority;
   - recovery after an emergency stop is explicit.
5. The figures match the text, and no contradicting copy remains.
6. The outcome `TASK-261012-13lcu6_gaps.md` is complete and true.
7. Public hygiene: no first names, host names, local absolute paths, session links or board-resource links.

No tests or builds on this host.

Do checks 1–7 yourself and write your verdict BEFORE you read `gw-LRM1-review-A.md`. Then reconcile, and check every P0/P1 that A raised.

Accept (`accept_cr`) if no P0/P1 remains; P2 issues are notes under tb-R226. Otherwise request changes, listing the exact defects. END YOUR TURN.
