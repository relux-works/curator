# TASK-260916-yvxbs1 — review verdict rev6: ACCEPTED

Candidate tree 8ac4bae1 (worktree `git write-tree` = CR tree), base 97e85642 (trunk).
`git merge-tree --write-tree --merge-base 86552087 refs/campaign/wgt8vz-rev5-20260928 97e85642` = 246fffeb; diff to candidate = exactly 3 paths (confirms orchestrator's 30/33).

1. internal/envprofile/envprofile.go — merge-tree left one conflict block (comment + rootMember guard, ~L1191). Resolution keeps trunk's hoisted
   `rootMember, hasRootMember := oldLock.RootMember()` guard (L1173) and the rev5 comment. Line-level check: the +/- set of `diff 97e85642..8ac4bae1`
   equals the +/- set of `diff 86552087..rev5` for this file except one comment punctuation (". The" → "; the"). So E1's update/delta/confirmation
   code is untouched and E6's validateProfilePathSources preflight (non-default only, after the default `profile_update_blocked` branch) is present;
   nothing else added. Trunk's §9.6-import comment wording is replaced by rev5's wording, as in accepted rev5 — comment only.
2. .github/ci/platform-cases.tsv — one added row (TestResolveRejectsGroupWritableOnlyPathOverlay, linux,darwin must / windows skip platform-control);
   no duplicate (package,case) keys in the file; no existing row changed.
3. path_source_group_write_unix_test.go (new) — drives Resolve (production entry) after setting ONLY g+w on an installed path overlay; asserts no
   fragment + DiagPathSourceUntrusted naming "permissions". Correct per environments §4 ("no identity other than the operator may mutate").
   Not in rev5 so it breaches "add nothing", but it is a pure strengthening (narrowing test), not masking: mutant pathboundary/owner_unix.go:31
   `&0o022` → `&0o002` (other-write only) → test FAILS (killed); unmutated passes. Accepted as legitimate.

Run (zsh, pipefail): `go test ./internal/envprofile -run 'Path|Boundary|Update|Delta|Signer|Guarded' -count=1` → ok 132s, rc=0.
Mutant run in disposable archive of 8ac4bae1 in $TMPDIR: `go test ./internal/envprofile -run TestResolveRejectsGroupWritableOnlyPathOverlay` → FAIL.
Hosted gate: accepted from orchestrator (green); not rerun here.
