# TASK-260910-32gki6 — review verdict rev8: ACCEPTED

Candidate: base 0a638288, tree ad18cefb (the worktree equals the candidate: `git diff --quiet ad18cefb` exit 0). Reference: accepted rev5, `refs/campaign/148pj1-rev5-20260929` (213a53e5 → 0fd98b72).

## 1. Path sets
The 31 paths are the same in rev5 and rev8 (the path-list `diff` is empty).

## 2. Line multisets for paths outside the conflict set
28 of the 31 paths have +/- line multisets identical to rev5. There is one exception outside the declared conflict set:
- `internal/envprofile/store_boundary_conformance_test.go`: adds `TestResolveRepairKeepsEnclosingStoreRootFailureClass`. This test is not in rev5; it was introduced in rev6 (`git diff 6a7deb11 refs/campaign/148pj1-rev6-20260929` contains it). It adds a test only. It pins the enclosing-boundary class under `Repair=true`: the ownership failure on the profile store root is refused, does not enter the rebuild path, emits no fragment and leaves the home unchanged. It is additive and strengthens S5 rule 2, and it weakens nothing. Accepted as an explained deviation.

## 3. Conflict paths
- `conformance-case-counts.tsv`: identical to rev5, so no bucket applies.
- `envprofile/status.go`: the only difference is gofmt realignment of 3 struct-field lines (trunk added a longer field name). (a) contains only whitespace re-alignment, and (b) and (c) are empty.
- `pathboundary/pathboundary.go`: every difference is the merge with uyak0e.
  - Trunk's closure `vanished` moved into `validateTree` with the guard `path == target || !IsAbsent(err)`. The inline `entry.Info()` lookup became `defaultEntryInfo`, and `entryInfo` is now threaded through `validateWithinWithOwnerAndEntryInfo`.
  - (a) contains only these both-sides adaptations plus comments. (b) No trunk rule was dropped: the skip for vanished children is intact. (c) Not applicable.

## 4. Substance and tests (zsh, `set -o pipefail`, real exit codes)
- `go test ./internal/pathboundary` → EXIT 0
- `go test ./internal/envprofile -run 'Store|Boundary|Pin|Resolve|Status|Legacy|PathInstall'` → EXIT 0
- The uyak0e rule is kept: in `validateTree` (pathboundary.go:244-253), ENOENT on a non-target child returns SkipDir or nil.
- The S5 rule is kept: every named component (root, lock, marker, store entry) is `os.Lstat`-ed in the route walk of `validateWithin…` (pathboundary.go:~203). Any error there, including ENOENT, returns `Failure{CheckRegular}`. The walked target is never skipped (guard `path == target`). Named store entries are additionally recomputed by `validateNamedStorePins` (store_boundary.go:159). A missing entry therefore has no hash and becomes an entry-class `storeEntryFailure`.

## 5. Mutants (run by the reviewer; the file was restored afterwards and the tree was verified equal to ad18cefb)
- **M-a:** vanishing skip applied to the walked target (`path == target ||` removed). **SURVIVED** (pathboundary and envprofile EXIT 0). The route walk had already lstat-ed the target, so this guard only closes the readdir→lstat race on the target itself. The mutant is equivalent under non-concurrent tests. Stated bound: this is race-only defense in depth.
- **M-b:** vanishing skip applied to named components in the route walk (`if IsAbsent(err) { return nil }` after the named `os.Lstat`). **SURVIVED** in `./internal/pathboundary` (EXIT 0) and in `./internal/envprofile -run 'Store|Boundary|Pin|Resolve'` / `'Missing|Absent|Untrusted|Env'` (EXIT 0). `cmd/curator -run Env` hit the 600 s go-test timeout, so no verdict comes from that package. Why it survives: a missing lock or marker is already an absence handled by `stateread.Lstat` before pathboundary is called (store_boundary.go:84-90, 103-109), so pathboundary never sees it. A missing lock-named store entry is still refused as `environment_store_untrusted` (entry class) by the pin-hash layer, so the S5 behaviour holds through a second layer.
  - Residual (not blocking, for a follow-up): no pathboundary unit row asserts that `ValidateWithin(root, missingTarget)` fails with CheckRegular. Suggested row: `pathboundary_test` `ValidateWithinWithOwner(root, root/absent)` → `*Failure{Check: CheckRegular}`.
- Earlier rev5 mutants (pin check skipped; missing object treated as pass; boundary check removed; class escalated/demoted) were killed in the rev5 review. Rev8 does not change that code (identical multisets).

## 6. Hygiene
No CHANGELOG or LOGBOOK change, and no stray files (31 product/test/.github/ci paths).

**Verdict: ACCEPTED rev8.** The only fidelity deviations are the rev6 test addition and the gofmt alignment, both explained. The uyak0e merge is correct. Missing named S5 state fails closed.
