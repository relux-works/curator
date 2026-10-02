# Review note — TASK-261001-3bsyvh carrier identity review (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

This carrier re-applies the already ACCEPTED TASK-260728-rjxrgs rev1 unchanged:
- accepted tree: ac008990, on base 5432c85f;
- snapshot: refs/campaign/rjxrgs-rev1-20261001.

The new CR is rev1 on base bab2433b, tree 1e1c5d83, 17 paths, gate green.

Prove identity. The accepted change and the carrier change must be the same delta:
1. `git diff 5432c85f ac008990` vs `git diff bab2433b 1e1c5d83`, restricted to non-.task-board paths.
   - The path lists must be equal.
   - The +/- line multisets must be equal (two-way).
2. Any difference must be explained solely by base movement between 5432c85f and bab2433b. Prove that with the per-path blob or merge-tree comparison.
3. Check for stray files, CHANGELOG/LOGBOOK, and the employer name; never spell the employer name.

If identical: accept_cr, citing the substantive review in TASK-260728-rjxrgs_review-verdict (rev1). Otherwise: changes requested, listing the exact differing hunks.

The host has exec stalls: use patient retries, never conclude from a hung command.
