# TASK-260906-2b3nar — fold directory components in the gitops platform-path gate

## Defect

`internal/gitops/gitops.go` `planWrites` keyed the case-folding collision gate on
`strings.ToLower(target)` of the FULL path only. Two tree entries whose directory
components fold together but whose basenames differ (`Dir/x.txt` + `dir/y.txt`)
produced distinct keys, were admitted, and landed in ONE physical directory on a
case-folding destination: the extracted tree no longer reproduced the committed tree
and the acquisition byte-exactness claim (environments.md 1.2) was violated without
a diagnostic. Pre-existing (the git archive path had it too), not a regression.

## Fix

`planWrites` now folds per path component. It tracks every ancestor prefix of every
planned target, lower-cased, alongside the folded full paths, and probes the
destination once via the existing `destinationFoldsCase` (result shared with the
full-path check through one closure):

- a directory prefix folding onto an already-planned prefix with a different exact
  spelling, in either order, is refused on a case-insensitive filesystem;
- a file folding onto a planned directory (`Dir/x.txt` + `dir`, both listing orders)
  and a prefix folding onto a planned file are refused the same way;
- refusal uses the EXISTING `duplicate platform path in git snapshot: %q` class and
  names the colliding entry; it happens in the pre-pass, before cat-file starts, so
  a refused extraction writes nothing;
- case-sensitive destinations admit the same trees exactly as before (both spellings
  extracted byte-exact); exact full-path duplicates keep the `duplicate path` message.

Production call site: `gitops.Extract` (used by `internal/snapshot`, `internal/closure`,
`internal/contextstore`, `internal/envprofile`); the gate lives in `planWrites`,
called from `Extract` before `writeBlobs`.

Files: `internal/gitops/gitops.go` (fix), `internal/gitops/dirfold_test.go` (new tests),
`CHANGELOG.md` (Unreleased/Fixed entry), `.github/ci/platform-cases.tsv` (4 rows).

## Tests (all at the exported `Extract` entry)

Helper `commitCaseTree` commits through `update-index --cacheinfo` + `write-tree`, so
folding names that cannot coexist on disk still commit (index/object DB only). Each
filesystem-dependent test probes with the production `destinationFoldsCase` and skips
with a named host-capability reason only where the platform cannot provide the needed
case semantics.

| ID | Test | Case-insensitive host FS (this host, APFS) | Case-sensitive scratch APFS vol |
|----|------|---------------------------------------------|----------------------------------|
| (a) | TestExtractRefusesDirectoryComponentFold (`Dir/x.txt`+`dir/y.txt` refused, nothing written) | PASS (refused, pre-pass) | SKIP, named: test filesystem is case-sensitive ... |
| (b) | TestExtractAdmitsDirectoryComponentFoldWhenCaseSensitive (same tree, both files exact bytes) | SKIP, named: test filesystem lacks case sensitivity ... | PASS (both files, exact bytes) |
| (c) | TestExtractRefusesFileDirectoryFold (`a/x.txt`+`A` and `Dir/x.txt`+`dir`, both orders refused) | PASS (refused, pre-pass) | SKIP, named |
| (d) | TestExtractRefusesNestedDirectoryComponentFold (`A/B/x`+`a/b/y` refused) | PASS (refused, pre-pass) | SKIP, named |

Note on (c): the file-first order (`a/x.txt` + `A`, sorts first) is the discriminating
one — without the pre-pass the file streams first and the directory then fails
mid-stream with a raw MkdirAll error instead of the duplicate-platform-path class.

Ledger: (a),(c),(d) require darwin,windows and tolerate linux (host-capability);
(b) requires linux and tolerates darwin,windows (host-capability).
`ledger-consistency.sh`: 245 rows checked, ok.

## Mutants (each applied, executed, reverted; revert verified by diff)

| Mutant | Shape | Result |
|--------|-------|--------|
| m0 (pre-fix code) | fold only the full path | KILLED: (a) err nil (silent merge), (d) err nil, (c) MkdirAll wrong-class error; exit 1 |
| m1 | fold only the basename (prefix tracking replaced by basename-fold check) | KILLED by (a),(c),(d); exit 1 |
| m2 | skip the prefix probe (prefix collisions detected, always admitted) | KILLED by (a),(c),(d); exit 1 |
| m3 | refuse only exact prefix duplicates (exact keys, no fold) | KILLED by (a),(c),(d); exit 1 |

No survivors.

## Validation (shell: bash, `set -o pipefail` semantics via PIPESTATUS)

- `go build ./...` — exit 0
- `go vet ./internal/gitops/ ./internal/snapshot/` — exit 0
- `go test ./internal/gitops/ ./internal/snapshot/ -count=1` — exit 0
  (gitops ok 38.9s; snapshot ok 4.4s)
- `go test ./internal/closure/ -count=1` — exit 0 (ok 226.6s)
- `go test ./internal/contextstore/ -count=1` — exit 0 (ok 3.8s)
- `go test ./internal/envprofile/ -run TestImport -count=1` — exit 0 (ok 87.9s)
- `golangci-lint run ./internal/gitops/...` — 0 issues, exit 0
- `gofmt -l` on touched Go files — clean

Windows/Linux hosted lanes: unverified from here (darwin host); ledger rows route
(a),(c),(d) to darwin,windows and (b) to linux.

## Finding (out of scope, no scope change)

`go test ./internal/envprofile/ -count=1` (full package) hit the go test timeout
(10m default; reproduced with `-timeout 9m`) in `TestImportCorruptMarkerIsLoss`.
Evidence it is unrelated to this change: `import.go` has no reference to
gitsource/snapshot/Extract (the hanging test cannot reach the changed code), and the
test passes alone on the fixed tree in 7s (exit 0), as does the whole `-run TestImport`
subset (exit 0). Pre-existing full-package interaction flake; left for its owner.

No LOGBOOK.md edit: campaign rules forbid worktree LOGBOOK writes.

---

## Revision 2 (base refresh only — no product change, no test weakened)

Revision 1 was ACCEPTED with no blocking findings, but integration refused on a
stale base: trunk advanced to `48da2690` (rc.12 pin promotion, PR #79) with a
`CHANGELOG.md` change revision 1 also makes. This revision replays the identical
candidate onto trunk `48da2690fe79ddb24eff078c5efb13d4869aa6a9` via
`task-board worktree refresh-candidate TASK-260906-2b3nar` (Outcome:
`refresh_advanced`, empty detail, no conflict path taken).

### Combine

Trunk `09b25ef6..48da2690` touches 42 files; the only path overlapping the
4-path candidate is `CHANGELOG.md` (union semantics: trunk adds E2/E4/S4/pin
entries in Added/Fixed; the candidate adds its 10-line Fixed entry after the
"Closure scratch ... place only on success." paragraph — disjoint hunks).
The refreshed `CHANGELOG.md` is trunk's file plus the candidate's entry
re-inserted at the same anchor.

### Byte-identity proof (refreshed tree vs revision 1)

`git status --short` on the refreshed tree:

```
M .github/ci/platform-cases.tsv
M CHANGELOG.md
M internal/gitops/gitops.go
?? internal/gitops/dirfold_test.go
```

- `cmp` clean (byte-identical to revision 1): `.github/ci/platform-cases.tsv`,
  `internal/gitops/gitops.go`, `internal/gitops/dirfold_test.go`. Trunk did not
  touch `internal/gitops` or `internal/snapshot` at all, so the narrow gate
  below runs identical code and tests.
- `CHANGELOG.md` diff vs new HEAD is exactly the revision-1 10-line insert
  (same hunk text; context shifted 89→156 by trunk's Added entries).
- `git diff --stat`: 3 files, 91 insertions(+), 9 deletions(-) — identical
  counts to revision 1; untracked `dirfold_test.go` unchanged at 134 lines.

### Re-run rows on the refreshed tree (shell: bash, real exit codes)

- `go test ./internal/gitops ./internal/snapshot -count=1` — exit 0
  (gitops ok 15.9s; snapshot ok 2.1s)
- Four fold rows (`-v -run`, case-folding host APFS): (a)
  TestExtractRefusesDirectoryComponentFold PASS; (b)
  TestExtractAdmitsDirectoryComponentFoldWhenCaseSensitive SKIP with the same
  named reason (`test filesystem lacks case sensitivity: Dir and dir fold into
  one directory, no coexistence to assert`); (c)
  TestExtractRefusesFileDirectoryFold PASS; (d)
  TestExtractRefusesNestedDirectoryComponentFold PASS — exit 0
- `sh .github/ci/ledger-consistency.sh <evidence-dir>` — exit 0
  (245 rows checked across linux darwin windows, ok — same count as rev1;
  the 4 new rows are present and routed as before)
- `go vet ./internal/gitops/ ./internal/snapshot/` — exit 0
- `gofmt -l` on touched Go files — clean
- `go build ./...` — exit 0

### Behaviour-change audit (rc.12 pin)

`SPEC_PIN=dced9b8` confirmed in `.github/workflows/ci.yml`. No row changed
behaviour: the narrow gate contains no `CURATOR_CONFORMANCE_ROOT`/SPEC-gated
skips (all skips are host-capability gates; the byte-exact/normative vectors
are hand-pinned local fixtures), and a full `-v` narrow run shows the single
expected SKIP (row (b), same named reason) with zero FAILs. The SPEC-consuming
conformance suites live outside the narrow gate and were not re-run here.
Mutants m0–m3 stand as killed on the byte-identical code (rev1 table above);
no fix was needed, so none was made.

---

## Revision 2 retry, attempt 2 (autonomous recovery RUN-260922-879f6c — no product change)

Attempt 1's Change Request validation (`remote-gate.sh`, run 35735392681)
failed on ONE lane: `Test (windows-latest)`, `go test overall exit=1`,
while the platform-case gate on that lane exited 0 and every other lane
(Lint, ubuntu, macos, both Race lanes, both Gate self-tests, Interop
conformance, Naming) passed.

### The single failure is outside this candidate and outside the trunk delta

From the run's `test-evidence-windows-latest` artifact
(`test/go-test-served.json`), the only failing test in the whole suite is:

- package `github.com/relux-works/curator/internal/managerlock`
- test `TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked`
- output: `managerlock_test.go:532: uncontended helper with tiny deadline =
  "acquired", want blocked`

The test spawns a helper subprocess with a **1-nanosecond** deadline
(`managerlock_test.go:531`) and asserts the result is `"blocked"`; on that
windows runner the uncontended helper acquired the lock inside the 1ns
window, so the assertion flipped to `"acquired"`. Scheduling-sensitive by
construction; the suite ran ~27 min on that runner.

Non-causation evidence (each checked on this tree, HEAD `48da2690`):

- `internal/managerlock` has no reference to `gitops` and
  `go list -deps ./internal/managerlock/` contains no `internal/gitops`
  edge — the changed code is unreachable from the failing test binary.
- The trunk delta `09b25ef6..48da2690` (single commit, rc.12 pin promotion)
  touches neither `internal/managerlock` nor `internal/gitops`
  (`git diff --stat` over both paths is empty).
- Revision 1 (byte-identical code and tests) passed `Test (windows-latest)`
  (validation run 35722367719, all lanes green).
- All `internal/gitops` rows passed on the red windows lane, including the
  three new refusal rows (`TestExtractRefusesDirectoryComponentFold`,
  `TestExtractRefusesFileDirectoryFold`,
  `TestExtractRefusesNestedDirectoryComponentFold`); the coexistence row
  skipped by design (host-capability tolerance).
- Local probe: the failing test passes 5/5 on this darwin host
  (`go test ./internal/managerlock/ -count=5 -run
  TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked`, exit 0).

Verdict: unrelated pre-existing timing flake on trunk, not a regression
from this candidate and not in this task's scope to fix (refresh brief
item 3 covers this task's own rows, which are green below). Left for its
owner; the rerun at handoff re-executes the full suite.

### Re-run rows on this tree (shell: bash, real exit codes)

Working tree verified as the rev2 candidate before the runs: HEAD
`48da2690fe79ddb24eff078c5efb13d4869aa6a9`, `git status --short` shows
exactly the 4 candidate paths (3 modified + `dirfold_test.go` untracked).

- `go build ./...` — exit 0
- `go vet ./internal/gitops/ ./internal/snapshot/` — exit 0
- `go test ./internal/gitops ./internal/snapshot -count=1` — exit 0
  (gitops ok 14.4s; snapshot ok 2.2s)
- Four fold rows (`-v -run`, case-folding host APFS): (a)
  TestExtractRefusesDirectoryComponentFold PASS; (b)
  TestExtractAdmitsDirectoryComponentFoldWhenCaseSensitive SKIP with the
  same named reason (`test filesystem lacks case sensitivity: Dir and dir
  fold into one directory, no coexistence to assert`); (c)
  TestExtractRefusesFileDirectoryFold PASS; (d)
  TestExtractRefusesNestedDirectoryComponentFold PASS — exit 0
- `sh .github/ci/ledger-consistency.sh <evidence-dir>` — exit 0
  (245 rows checked across linux darwin windows, ok)
- `gofmt -l` on touched Go files — clean
- Mutants m0–m3 stand as killed (code byte-identical to rev1; no fix made,
  none needed)

---

## Revision 3 retry, attempt 3 (autonomous recovery RUN-260922-0e1ca4 — no product change)

Attempt 2's Change Request validation (`remote-gate.sh`, run 35739653315)
failed on ONE lane: `Test (windows-latest)`, `go test overall exit=1`,
while the platform-case gate on that lane exited 0 and every other lane
(Lint, ubuntu, macos, both Race lanes, all three Gate self-tests, Interop
conformance, Naming) passed.

### The single failure is the same unrelated flake as rev2

Downloaded the run's `test-evidence-windows-latest` artifact and parsed
`test/go-test-served.json`: the only failing test in the whole suite is
again:

- package `github.com/relux-works/curator/internal/managerlock`
- test `TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked`
- output: `managerlock_test.go:532: uncontended helper with tiny deadline =
  "acquired", want blocked`

Byte-identical failure (same file, line, and message) to rev2's run
35735392681. The test spawns a helper subprocess with a 1-nanosecond
deadline (`managerlock_test.go:531`) and asserts `"blocked"`; on that
windows runner the uncontended helper acquired the lock inside the 1ns
window. Scheduling-sensitive by construction.

Non-causation evidence (each re-verified on this tree, HEAD `48da2690`):

- `grep -rn "gitops" internal/managerlock/` — no match (exit 1); the
  changed code is textually absent from the failing package.
- `go list -deps ./internal/managerlock/` contains no `internal/gitops`
  edge (count 0) — the changed code is unreachable from the failing
  test binary.
- The trunk delta `09b25ef6..48da2690` touches neither
  `internal/managerlock` nor `internal/gitops` (empty diff stat).
- Revision 1 (byte-identical code and tests) passed `Test
  (windows-latest)` (validation run 35722367719, all lanes green).
- All `internal/gitops` rows passed on the red rev3 windows lane:
  the three new refusal rows
  (`TestExtractRefusesDirectoryComponentFold`,
  `TestExtractRefusesFileDirectoryFold`,
  `TestExtractRefusesNestedDirectoryComponentFold`) PASS; the
  coexistence row SKIPs by design (host-capability tolerance).
- Local probe: the failing test passes 5/5 on this darwin host
  (`go test ./internal/managerlock/ -count=5 -run
  TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked`, exit 0).

Verdict: same unrelated pre-existing timing flake on trunk, not a
regression from this candidate and not in this task's scope to fix
(this task's own rows are green below). Left for its owner; the rerun
at handoff re-executes the full suite.

### Candidate identity (no refresh needed)

Working tree verified as the rev3 candidate before the runs: HEAD
`48da2690fe79ddb24eff078c5efb13d4869aa6a9` (`git fetch origin main`
confirms trunk is still there — no new base movement, so no
`refresh-candidate` replay), `git status --short` shows exactly the 4
candidate paths (3 modified + `dirfold_test.go` untracked),
`git diff --stat` 91 insertions(+)/9 deletions(-). `cmp` of the
rev2 and rev3 Change Request patches is clean (byte-identical); rev1
differs only in the CHANGELOG hunk header (base shift 89→156).

### Re-run rows on this tree (shell: bash, real exit codes via PIPESTATUS)

- `go build ./...` — exit 0
- `go vet ./internal/gitops/ ./internal/snapshot/` — exit 0
- `go test ./internal/gitops ./internal/snapshot -count=1` — exit 0
  (gitops ok 13.2s; snapshot ok 2.1s)
- Four fold rows (`-v -run`, case-folding host APFS): (a)
  TestExtractRefusesDirectoryComponentFold PASS; (b)
  TestExtractAdmitsDirectoryComponentFoldWhenCaseSensitive SKIP with the
  same named reason (`test filesystem lacks case sensitivity: Dir and dir
  fold into one directory, no coexistence to assert`); (c)
  TestExtractRefusesFileDirectoryFold PASS; (d)
  TestExtractRefusesNestedDirectoryComponentFold PASS — exit 0
- `sh .github/ci/ledger-consistency.sh <evidence-dir>` — exit 0
  (245 rows checked across linux darwin windows, ok)
- `gofmt -l` on touched Go files — clean (no output)
- Mutants m0–m3 stand as killed (code byte-identical to rev1; no fix made,
  none needed)
