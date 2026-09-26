# BUG-260922-306v4m review verdict — rev1: ACCEPTED

Candidate tree OID: ea959db51e32dc2f28ff31b104c5010ec99d238a
Base OID: 48da2690fe79ddb24eff078c5efb13d4869aa6a9
Change Request: CR-BUG-260922-306v4m-1 revision 1

Reviewed that exact tree. The worktree matches the candidate.

1. Success path: the only change on the success path is that `[ -n "$rustup_bin" ] || fail ...` became `if [ -z ]; then diag >&2; fail ...; fi`. Resolution, GITHUB_PATH writes, PATH export and the "using rustup at" line are unchanged. I checked this by running both scripts: base.sh and the candidate were run with a fake rustup under CARGO_HOME/bin, and they produced identical GITHUB_PATH files (cmp) and identical stdout/stderr.
2. Robustness: I ran it under `env -i PATH=/usr/bin:/bin`, with HOMEBREW_PREFIX, CARGO_HOME and RUNNER_* all unset and the cargo bin directory missing. The whole block printed, the remedy sentence was the last line, and it exited 1. Only the named variables are printed.
3. Self-test: the full `bash .github/ci/gate-selftest.sh` run passed: 224 passed, 0 failed, exit 0. All rows in the install-rust section are `ok` and none were skipped (this host has no /opt/homebrew or /usr/local rustup). That includes the absent fixture, the defaulted-CARGO_HOME fixture with its 3-candidate count, the 20-entry bound, the remedy-last check, and the narrowing mutant being killed.
4. The docs paragraph and the CHANGELOG entry are present. No probe paths were added, which is correct because there is no evidence yet.

Residuals (not blocking):
- The mutant only covers the /opt/homebrew candidate. The other candidates are pinned by named assert_contains rows, and the CARGO_HOME candidate by the count row.
- The `rust|cargo` filter also matches unrelated names, for example `trust`. This is harmless noise.
- On a host that has rustup at /opt/homebrew or /usr/local, the diagnostic rows are skipped by name.

## Board recording
`accept_cr(BUG-260922-306v4m, revision=1, ...)` was REFUSED by the runtime with `validation_not_bound_to_tree: ... the validation evidence carries no source tree identity ... revalidation is required (candidate_tree_oid=ea959db51e32dc2f28ff31b104c5010ec99d238a, evidence_tree_oid=ea959db51e32dc2f28ff31b104c5010ec99d238a)`.
On content the verdict is ACCEPT, but the board will not record it against this validation record. I routed the task to `to-dev`, following the TASK-260906-2b3nar precedent. **No code change is requested.** The only thing needed is to republish/revalidate: hand off the same tree again so the validation evidence is bound to it. The next reviewer can reuse these checks once they confirm the new candidate tree is still ea959db5.
