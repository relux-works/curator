# Review note for BUG-260920-3vfwch revision 2 (orchestrator, binding)

Revision 2 = rework 1 for your revision-1 verdict (BUG-260920-3vfwch_review-verdict-rev1.md; brief
3vfwch-rework-1.md): F1 `logf: t.Logf` wired in both copies + a committed row driving the OUTER
entry through a real retry asserting the logged lines; F2 darwin rlimits via `unix.Getrlimit`
NOFILE+NPROC in one `linux || darwin` file, content pinned by regexp; N1 EPERM predicate row, N2
backoff literal pin, N3 count. Gate green: run 35531307725 — verify the gate commit resolves to
the exact revision-2 tree. Rerun your rev1 probe (`TestReviewProbeWiredEnvGitRetrySucceeds`
shape) against rev2 and expect the `retrying once`/`succeeded on retry` lines in the `-v`
stream; drive the diagnostic once and check the rendered rlimits line on darwin; rerun mutants
A, B-full, C (backoff), E (EPERM). Both package copies in sync (diff). Record exactly one
verdict: accept_cr(BUG-260920-3vfwch, revision=2, evidence=<your outcome resource>) on ACCEPT,
or changes_requested with file:line and reproduction. Bounded local commands; the host has
stall windows — retry a hung command once rather than blocking.
