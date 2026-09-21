# Review note for BUG-260920-2d9gfv revision 2 (orchestrator, binding)

Revision 2 = rework 1 for your revision-1 verdict (BUG-260920-2d9gfv_review-verdict-rev1.md;
brief 2d9gfv-rework-1.md): function-entry `start := time.Now()` bound
(`now.Add(clockSkew).Add(time.Since(start))`), the inverse "slow sibling" unit rows (instant
registry with `created_at = now + 1 s` accepted; `now + skew + 3 s` still refused), the
two-registry production-entry row in `internal/install/registry_e2e_test.go` (fails on rev1,
passes on rev2), `crossSecondBoundary`/comments (F2/F3), CHANGELOG wording, results.md
"Revision 2" with the read-only-path bound and the hoist-vs-per-registry mutant. Gate green:
run 35555033097 — verify the gate commit resolves to the exact revision-2 tree. Rerun your
two-registry probe against rev2 (expect 3/3 pass), the R3 determinism command at a smaller
count, the mutants (drop elapsed term; hoist→per-registry now killed by the sibling row?),
and diff rev1→rev2 (expected: snapshot.go, registry_test.go, registry_e2e_test.go, harness
comment/margin, CHANGELOG, results.md — nothing else). Record exactly one verdict:
accept_cr(BUG-260920-2d9gfv, revision=2, evidence=<your outcome resource>) on ACCEPT, or
changes_requested with file:line and reproduction. Bounded local commands; retry once on host
stalls.
