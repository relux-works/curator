# Review note for BUG-260920-3vfwch revision 1 (orchestrator, binding)

Brief: 3vfwch-brief.md (rulings R1–R4). Gate green: run 35524579878 on gate commit d57ca591 —
verify the gate commit's tree equals the exact revision-1 candidate tree (CR patch attached).
Independent exact-head review; bounded local commands (the host has stall windows — retry a
hung command once rather than blocking).

Judge specifically:
1. R1 honoured: the CR is test-only (`*_test.go` only) — confirm against the patch; no product
   spawn path touched.
2. Retry rule is exactly as ruled: one retry, only on spawn `EACCES`/`EAGAIN` (never on
   `*exec.ExitError`, never on ENOENT/other errnos), 200 ms backoff, logged. Re-run the
   predicate table and the real-kernel EACCES row; run mutants A and B-full yourself.
3. Diagnostic block content at the moment of failure (argv0 resolution, binary lstat/stat +
   symlink target, dir stat, cwd, PATH, rlimits) and that it appears in the go-test stream on a
   persistent failure — drive it once.
4. Cause classification in results.md §1–2: check the run-1 claim (sequential test, product
   spawn in `gitops.writeBlobs` with bare `cmd.Start()` error, `cmd.Dir` unset) against the
   artifact and the source at head 208953b; check P2/P3 audits by your own grep (`Setenv`,
   `Chdir`, `Chmod`, `Umask` in the install test binary).
5. Wired call sites keep argv/env/dir/output byte-identical on the success path (install
   `env.git`, `draftsources_test.go testGit`, `drafttransport_test.go draftBareFixture.run`,
   `atomicity/fixture_test.go env.git`); the two package copies are in sync (diff them).
6. Bound B3 (product-shaped occurrence not mitigated) is stated honestly; note it as a
   residual for a separate follow-up leaf, not as a finding against this leaf, unless the
   AC as written requires product-side action — say which.
Record exactly one verdict: accept_cr(BUG-260920-3vfwch, revision=1, evidence=<your outcome
resource>) on ACCEPT, or changes_requested with file:line and reproduction.
