# Review note — TASK-260924-20o9dk released skillfile-sources corpus conformance (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review the newest revision (rev4, 166 paths, gate green) against `20o9dk-brief.md`, `20o9dk-decision-1.md`, `20o9dk-gatefix-1.md` and the PR #96
review's curator section (`pr96-review-for-20o9dk.md`). Bounded, split runs only (host memory is tight).
1. Vendored corpus = curator-spec conformance/skillfile-sources-v1 at 5746367 (byte-identical; pin file + MANIFEST; counts from the corpus);
   no leftover draft-sources-v1 vendoring.
2. Every released case has a passing production-entry row (105/105 semantic, 3/3 snapshot) and the three local-snapshot-v1 schema documents
   are explicit bounds with the decision's reason — no gap rows.
3. Production rules: refresh uses the current resolved endpoint plan (stored remote.origin.url never selects it); SCP-like + alias port →
   repository_policy_invalid before I/O; SSH URI + alias port rendering; repository+commit revocation deny-wins under advisory; C1 replay
   verifies the locked object format; C2 declared-mirror driver asserts expectations. Kill at least two of the listed mutants yourself.
4. Gate fix: the two literal tests now use a current endpoint plan to their local fixture and still assert their original intent (+ the new
   "stored origin did not select the endpoint" assertion); draftsources_gitshim_test.go goes through the shared process seam.
5. No CHANGELOG/LOGBOOK edit, no stray files, no revert of trunk. accept_cr or changes requested with file:line. No LOGBOOK.md.
