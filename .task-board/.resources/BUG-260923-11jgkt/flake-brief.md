# CI flake fix (curator) — read the task description first; it carries the exact failing run and error

Control root: curator; work only in your assigned Story worktree. Read `campaign-producer-rules.md`. Landing
gate = hosted CI, once, at handoff. This flake has been failing unrelated candidates' gates and costing full
republish cycles — fix the ROOT CAUSE, not the symptom.

1. Reproduce or evidence the failure: download the named gate run's `test-evidence-<os>` artifact
   (`gh api repos/relux-works/curator/actions/runs/<ID>/artifacts`, `test/go-test-served.json`) and quote the
   exact error. Search earlier gate runs for the same test to establish the frequency.
2. Decide: test race, runner/environment issue, or a REAL product defect. Cite the code. If it is a product
   defect, fix production minimally and name it.
3. The fix must be bounded and named: retry only the specific transient error (EACCES on the git spawn;
   Windows sharing violation on the concurrent open), never a blanket retry; a persistent/genuine error must
   still fail closed. Rows for BOTH the transient case (injected) and the genuine failure.
4. Do not weaken any assertion, do not skip the test, do not loosen to accept either outcome.
5. Prove determinism locally where the platform allows (state what you could not run), and let the hosted gate
   be the arbiter. CHANGELOG (Fixed). Attach `<ID>_results.md` and `task-board handoff <ID> --role developer`.
