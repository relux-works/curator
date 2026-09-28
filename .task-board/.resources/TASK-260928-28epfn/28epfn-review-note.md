# Review note — TASK-260928-28epfn nofollow parent-walk rows (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev1 (base 213a53e5, tree 60e96143, 1 path, gate green) against `28epfn-sec-brief.md` and the residual found in the 3ed9m3 rev2
review: on trunk, M1 (internal/envprofile/nofollow.go managedPath parent walk Lstat→Stat, ~L43) and M2 (symlink refusal disabled at ~L60,
and at ~L60+~L111) SURVIVED because later layers held the rows. Verify on a disposable copy of the candidate, real exit codes:
1. M1 alone → the new row(s) FAIL; M2 alone (each variant) → the new row(s) FAIL; the unmutated candidate passes.
2. The rows go through a production entry (env resolve/repair/switch writes), plant a symlink as a PARENT directory component of a managed
   write path, and assert environment_write_would_follow_link with nothing written outside the tree; cite environments §8.3.1.
3. No production code change (or, if any, it is a justified real gap fix); Windows behaviour/skip reasons are explicit.
accept_cr or changes requested with file:line. No LOGBOOK.md.
