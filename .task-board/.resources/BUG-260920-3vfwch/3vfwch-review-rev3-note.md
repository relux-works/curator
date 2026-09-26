# Review note for BUG-260920-3vfwch revision 3 (orchestrator, binding)

Revision 3 = rework 2 for your revision-2 verdict (BUG-260920-3vfwch_review-verdict-rev2.md; brief
3vfwch-rework-2.md): the only finding was F3 — the outer-entry row's timer-driven symlink flip.
Expected shape: event-driven flip performed by the recording TB on the `retrying once` line, no
goroutine/timer, swap errors fail the test, both package copies in sync. Gate green: run
35535957629 on gate commit 01ffa90a — verify the gate commit resolves to the exact revision-3
tree. Rerun your stress harness (`zz_stress_test.go`) next to the committed row on the rev3
tree (plain + `-race`), confirm mutants C and W still die, and diff the two copies. Everything
else was accepted at rev2 — do not re-open it unless rev3 changed those bytes (check with
`git diff` rev2→rev3: only the two `gitfixture_retry_test.go` files and results.md are
expected). Record exactly one verdict: accept_cr(BUG-260920-3vfwch, revision=3,
evidence=<your outcome resource>) on ACCEPT, or changes_requested with file:line and
reproduction. Bounded local commands; retry a hung command once (host stall windows).
