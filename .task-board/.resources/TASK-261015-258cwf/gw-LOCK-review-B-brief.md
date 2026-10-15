# Review B brief: LOCK (TASK-261015-258cwf), deciding review

You review the producer's Change Request for this task: the change described in the producer brief. Attached: the producer brief gw-LOCK-brief.md, rc7-common.md and rc7-decisions.md.

Decide ACCEPT or REJECT (changes requested) with a findings list, severity P1 (blocks) / P2 (should fix) / P3 (nit), each with file:line and a concrete fix. Sweep the whole surface; do not stop at the first hit. Check:

1. Brief coverage: every numbered item under "Required design content" is addressed with real substance.
2. Facts: spot-check at least eight "Current state" claims against the cited code or spec (open the files). Any wrong citation or invented field, refusal code or behaviour is P1.
3. Consistency with the owner decisions D1-D13: a recommendation that contradicts one is P1.
4. Correctness and minimality of the change; for code: tests at the production entry, the named mutant killed, lint and tests pass.
5. Hygiene of the public text: no host names, IPs, seat or lane names, event ids, local absolute paths, personal or employer names or emails, account names or key fingerprints (P1).
6. Scope: only the files the brief names or clearly implies; no unrelated edits (P1 if violated).

Attach the verdict as task resource TASK-261015-258cwf_review-verdict.md and complete the reviewer handoff (accept_cr on ACCEPT).
