# TASK-260906-2b3nar — revision 1 review

Verdict: ACCEPT. No blocking findings. Candidate tree 9ea731cba39f88072cdabc95a0de37ee7edb66a6, base 09b25ef6629b41455d91dcb252ab4e4034e12750. Acceptance is for integration, not a claim of landing.

All four changed working files matched candidate blobs byte-for-byte. Reviewed the exact tree delta. Candidate code was unchanged; attacks ran in an archived candidate under .temp/review-2b3nar/candidate and restored its original implementation after each experiment. No control-root or LOGBOOK edits.

## Independent execution

Shell zsh; pipelines used set -o pipefail. All checks below exit 0 unless explicitly stated.

- go test ./internal/gitops ./internal/snapshot -count=1 -timeout=120s: gitops 19.751s, snapshot 2.976s (also an earlier independent pass: 21.964s/2.398s).
- go vet ./internal/gitops ./internal/snapshot; go build ./internal/gitops ./internal/snapshot.
- golangci-lint run ./internal/gitops/...: 0 issues.
- bash .github/ci/ledger-consistency.sh .temp/review-2b3nar/ledger: 245 rows, linux/darwin/windows, OK. Initial invocation without its required evidence argument exited 2; corrected invocation passed.
- go test ./internal/envprofile -run '^TestImportCorruptMarkerIsLoss$' -count=1 -timeout=45s: 4.542s.

Coverage: 4/4 brief rows driven through production gitops.Extract on real filesystem classes. On host case-folding APFS, (a) directory folds, (c) file/directory in both orders, (d) nested folds pass; (b) explicitly skips. On scratch case-sensitive APFS, (b) passes with exact bytes and (a,c,d) explicitly skip. Created image with hdiutil -fs 'Case-sensitive APFS', attached under worktree, then detached. Initial APFSX spelling was rejected; no capability inferred from that failed attempt.

Production callers: internal/snapshot/snapshot.go:65,105 and internal/closure/closure.go:446. New refusals run in planWrites before writeBlobs. Both file-first and directory-first tests assert no residue. The shared closure caches the destination probe. Exact-path duplicate check still precedes folded checks. The new map admits repeated exact directory spelling and retains the first spelling for subsequent comparisons. Case-sensitive coexistence is preserved.

Additional reviewer production-entry attacks passed: three-way A/x+a/y+A/z; deeper-only a/B/x+a/b/y; dotted-I/i; K/Kelvin; empty tree; symlink/directory and gitlink/directory pairs; raw exact duplicate leaves. Refused shapes assert nothing written. Unsupported modes are rejected by listTree before destination creation. Empty path is rejected by safeTarget before component slicing (inspection); empty committed tree extracts successfully.

## Narrowing attacks

All 4/4 reviewer mutants killed (go test exit 1, behavioral assertion failures, not compilation failures):

| Mutant | Failure |
|---|---|
| m1 prefix keys reduced to folded basenames | committed file-first file/directory test receives raw mkdir error instead of typed refusal |
| m2 prefix collisions bypass the destination probe/refusal | committed directory and nested tests return nil; file-first returns mkdir error |
| m3 exact prefix keys instead of lowercase | same committed failures as m2 |
| reviewer m4 track only the first ancestor | reviewer deeper-only a/B/x+a/b/y returns nil |

Producer m0/full-path-only kill is accepted from attached results; m1-m3 independently executed above. The committed nested test already collides at the first component, so m4's deeper-only guarantee is covered by this review's executable attack, not by that committed row. Adding that shape to the permanent test would improve regression coverage, but implementation handles it correctly.

## Bounds and findings

strings.ToLower is not a complete filesystem equivalence predicate. NFC/NFD directory prefixes with different leaf basenames still merge and Extract returns nil. I executed the same production-entry case with the base gitops.go and reproduced it there too. Dotted-I and Kelvin pairs are refused by the candidate; this demonstrates the selected lowercase predicate, not exact equivalence with every APFS/HFS+/NTFS Unicode table. Accepted inherited bound under the binding review note's NEW-gap bar; no new normalization algorithm is required by this change. The comment claiming every other fold is caught by writeBlobs is too broad for directory-only normalization collisions, as it was before. Keep this as follow-up scope, not a cross-platform completeness claim.

The producer's envprofile full-suite timeout plausibly is an unrelated suite interaction: import.go has no direct path to changed gitops code, the isolated test passes independently, and exact-tree hosted suites passed. These facts support non-blocking treatment, but do not prove historical pre-existence: I did not reproduce the timeout on the base or rerun the full package. Root cause remains unknown; separate investigation is appropriate.

Ledger: four rows use existing host-capability class, required darwin/windows for refusal and linux for coexistence; opposite lanes tolerate capability skips. This matches ordinary hosted filesystems and actual probe behavior. No new skip class. CHANGELOG entry is under Unreleased / Fixed.

## Reused evidence and limits

Did not rerun the full landing suite. Attached revision-1 validation log reports hosted run https://github.com/relux-works/curator/actions/runs/35722367719 success, including Linux/macOS/Windows tests, macOS/Linux race, lint, naming, interop, and gate self-tests. Verified locally that gate commit e965b6eacc5566c4049e9a176c5c6c9265952f77 has exactly candidate tree 9ea731cba39f88072cdabc95a0de37ee7edb66a6. Rose-air and Candidate suite are explicitly skipped, not passing. Producer full build exit 0 accepted from its report; narrow build independently passed. No independent Windows/Linux runtime execution in this review.

Evidence bundle includes reviewer tests, mutation runner, execution logs, and ledger log. Goal query: run is not goal-bound. Changes-requested routing condition is inapplicable because verdict is acceptance; use accept_cr revision 1, never done or commit_ack.
