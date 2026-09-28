# TASK-260928-28epfn review verdict — CR rev1 ACCEPTED

Reviewer: claude-opus-5-5 (independent). Candidate: base 213a53e5, tree 60e96143 (1 path, test-only). Verified on a disposable `git clone --shared` + `read-tree -u --reset 60e96143`. Conformance root: curator-spec 23435129 (rc.13) `conformance/v1` extracted to $TMPDIR. Shell zsh, `pipestatus` captured. Test: `go test -count=1 -v -run TestEnvironmentWriteNofollowVectors ./internal/envprofile/` (NB: without CURATOR_CONFORMANCE_ROOT the test SKIPs — first attempt was discarded for that reason).

## Mutants (internal/envprofile/nofollow.go)
| mutant | candidate | base 213a53e5 |
|---|---|---|
| clean | ok, exit 0 | ok, exit 0 |
| M1 managedPath parent walk Lstat→Stat (~L43) | FAIL exit 1: materialize-symlinked-parent-refused (L329 "target was created outside the managed root"), takeover-symlinked-parent-authorized-still-refused (L326 link retargeted) | FAIL exit 1 (L312 error=nil) |
| M2a symlink refusal off in managedPath (~L60) | FAIL exit 1 (both parent cases, L337 diagnostic) | FAIL exit 1 |
| M2b refusal off at ~L60 + ~L111 (managedDirectory) | FAIL exit 1 (+ backup-symlinked-destination-refused, L426) | FAIL exit 1 |

## Findings
- Rows are production-entry: `Resolve(request)` (resolve/materialize, takeover) and `UseWithPolicy` (switch backup) with a planted symlink as PARENT component; assert environment_write_would_follow_link and that nothing was written/retargeted outside the tree. Environments §8.3.1 cited in the new comment (L316).
- The change reorders assertions so side-effect checks (outside target unchanged/not created, parent link intact, no backup) run before the diagnostic check; under M1 the row now reports the real escape (a write through the parent link), not just a missing diagnostic. Strict strengthening.
- Observation: with these mutant shapes, M1/M2 were already killed on base when the rc.13 root is set; the 3ed9m3 "survived" residual does not reproduce here (likely measured without the conformance root or with a different mutant shape). Not a defect of this CR.
- No production code change. Windows: symlink fixture failure → explicit `t.Skipf("symlink fixture unavailable")` (unverified on Windows, stated bound).
- No CHANGELOG/LOGBOOK edits.

Verdict: ACCEPTED (accept_cr revision=1).
