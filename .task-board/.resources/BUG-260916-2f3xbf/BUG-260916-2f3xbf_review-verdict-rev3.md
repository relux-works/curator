# BUG-260916-2f3xbf — review verdict, revision 3 (reviewer RUN-260920-46d84b, claude-opus-5)

**Verdict: ACCEPT** (`accept_cr(BUG-260916-2f3xbf, revision=3, evidence=BUG-260916-2f3xbf_review-verdict-rev3.md)`).

Reviewed: CR-BUG-260916-2f3xbf-3, base `7fa08e84`, candidate tree `d39576c7`, 9 changed paths.
Shell for every command below: `bash -c 'set -o pipefail; ...'` (host shell zsh); worktree left
untouched — all local runs on disposable `git clone --shared` checkouts of the gate commit under /tmp.
Windows was verified only through the hosted gate artifacts (no Windows runner here); real pnpm 10.33.0
was available locally at /tmp/pnpm-prefix (producer's prefix) for the darwin runs.

## 1. Exact candidate tree

| check | result |
|---|---|
| working tree of `.temp/STORY-260915-3w11un/worktree` hashed through a temp index (`read-tree HEAD; add -A; write-tree`) | `d39576c7590e2a6c304711fe299888ad60571536` = candidate tree OID |
| `BUG-260916-2f3xbf_change-request_rev3.patch` sha256 | `4bec436c…d40b` = board value; byte-identical to `…_rev2.patch` (`cmp` exit 0) |
| rev3 gate commit `61744f99` (validation log) | `tree d39576c7`, `parent 7fa08e84` |
| rev2 gate commit `29ac988c` | `tree d39576c7` — same tree, so run 35486280770 evidence is also for this candidate |
| `gh run view 35488745551` | headSha `61744f99…`, conclusion success; Test/Race/Gate self-test on ubuntu, macos, windows all `success`; rose-air + candidate suite skipped (not configured for gate runs) |

## 2. Hosted windows-latest: the two real-pnpm cases RAN and PASSED (not deferred, not skipped)

`gh run download 35488745551 -n test-evidence-windows-latest`, `test/go-test.json` + `test/go-test-served.json` + `test/observed-cases.tsv`:

| case | action | elapsed |
|---|---|---|
| internal/pnpmsource TestRealPinnedPNPMLockSupersetSnapshotDependencies | pass | 18.03 s |
| internal/pnpmsource TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization | pass | 18.2 s |
| internal/pnpmsource TestRealPinnedPNPMTargetPrunedUnreachableRejectsBeforeInstall | pass | 12.41 s |
| TestWritableStoreOverlayAdmitsWindowsJunctionRegistration / RefusesMisdirectedWindowsJunction / TestAdmittedLinkCoversWindowsJunctionShape / TestDirectNodeModulesAdmitWindowsJunctionLinks / TestCopyContainedTreeDereferencesWindowsJunction / …RefusesWindowsJunctionEscape / …RefusesWindowsJunctionCycle | pass (all 7, i.e. `mklink /J` worked on the runner, none skipped) | 0.05–0.07 s |
| TestWritableStoreOverlayAllowsOnlyExactProjectRegistration (4 subtests incl. `plain_directory_member_refused`) | pass | — |

`test/skips-observed.tsv` has no internal/pnpmsource row. ubuntu-latest and macos-latest evidence (same run) show the same cases `pass`.
Rev history corroborated from the attached logs: rev1 run 35481906193 — windows lane `FAIL required case failed` on exactly these two cases; rev2 run 35486280770 — `Test (windows-latest): success`, only `Race (ubuntu-latest)` failed (crossconformance row, filed BUG-260920-2d9gfv, status backlog); rev3 green.

## 3. The declared member: closed and platform-conditional

Member kind, verified against the pnpm 10.33.0 bundle at `/tmp/pnpm-prefix/lib/node_modules/pnpm/dist/pnpm.cjs`:
`registerProject(storeDir, projectDir)` writes `<store>/v10/projects/<createShortHash(projectDir)>` (sha256 hex prefix, 32 chars — the
`c8805e97…`/`4990f83e…` entries in the run 35098955988 failure) through `symlink-dir`, whose `symlinkType = IS_WINDOWS ? "junction" : "dir"`
(`fs.symlink(target, path, "junction")` → a directory junction, `IO_REPARSE_TAG_MOUNT_POINT`). With go.mod `go 1.25.5` semantics
(`os/types_windows.go` `Mode()`), a junction is a name-surrogate reparse point: no `ModeDir`, no `ModeSymlink`, so it reports `ModeIrregular`,
and `filepath.EvalSymlinks` does not evaluate it — exactly why the symlink-only check refused it.

Declaration (`internal/pnpmsource/linkentry_windows.go`, `//go:build windows`): `admittedLink` = symlink OR (`Mode().Type()==ModeIrregular` AND
`GetFileAttributes` has `REPARSE_POINT` AND `DIRECTORY`); `normalizeTreeLink` = `os.Readlink` for that shape (Go's readlink resolves only
SYMLINK/MOUNT_POINT tags, every other reparse tag errors → "link cannot be resolved" refusal), then the unchanged `EvalSymlinks(target) == EvalSymlinks(projectRoot)`
comparison. Effective admitted set = {symlink, mount-point junction} whose target IS the project — no wildcard; plain directories, regular files,
other reparse-point kinds and misdirected junctions all stay refused. `linkentry_unix.go` (`//go:build unix`) is the previous inline predicate
verbatim (`info.Mode()&fs.ModeSymlink != 0`) and an identity normalize. Same helper pair is used at the three link validators
(registry, `validateDirectNodeModules`, `validateSnapshotInstance`) and in `copyContainedNode`, mirroring the landed `internal/npmsource/linkentry_*.go`
(diff against the sibling: naming/comment differences only). Fits the architecture.

## 4. Local evidence (darwin, exit codes)

| # | command (clone of `61744f99`) | exit |
|---|---|---|
| 1 | `gofmt -l internal/pnpmsource/` (empty) | 0 |
| 2 | `go vet ./internal/pnpmsource/`; `GOOS=windows go vet …`; `GOOS=linux go vet …` | 0 / 0 / 0 |
| 3 | `GOOS=windows go test -c -o /tmp/… ./internal/pnpmsource/` (Windows test binary compiles, 10.6 MB) | 0 |
| 4 | `PATH=/tmp/pnpm-prefix/bin:$PATH go test -p 1 ./internal/pnpmsource -run 'TestRealPinnedPNPM\|TestWritableStoreOverlay\|TestCopyContainedTreeDereferencesLinks\|TestResolveWindowsPNPMEntrypoint' -count=1 -timeout=300s -v` — 3/3 real-pnpm cases PASS (7.55/11.15/12.03 s), 4/4 registry subtests, 3/3 copy subtests, 0 skips | 0 (`ok 31.220s`) |
| 5 | same PATH, `go test -p 1 ./internal/pnpmsource -count=1 -timeout=600s -v` — 29 top-level PASS, 0 FAIL, 0 SKIP | 0 (`ok 36.306s`) |
| 6 | `golangci-lint run ./internal/pnpmsource/...` (v2.12.2, 0 issues) | 0 |
| 7 | `bash .github/ci/ledger-consistency.sh /tmp/…` — 241 rows ok; both pnpm rows `must=linux,darwin,windows skip=-` | 0 |
| 8 | `bash .github/ci/no-broad-suppression.sh` | 0 |
| 9 | `bash .github/ci/gate-selftest.sh` — 185 passed, 0 failed, incl. the 8 rewritten pnpm assertions (`FATAL-not-tolerated` on windows and linux for a pnpm-absent skip) | 0 |

## 5. Gate attacks (mutants on separate disposable clones)

| mutant | expectation | observed |
|---|---|---|
| **A — narrowing: `admittedLink` (unix) returns `true` for every registry member** | the undeclared-member negative test must fail | `TestWritableStoreOverlayAllowsOnlyExactProjectRegistration/unclaimed_registry_member` and `/plain_directory_member_refused` FAIL: "rogue registry member refused by the wrong gate: … targets an undeclared project"; package `FAIL`, exit 1. `exact_registration` and `frozen_content_drift` still pass, so the kill is the kind gate, not collateral breakage. **Killed.** |
| **C — revert the `copyContainedNode` normalization hunk (rev2's only production delta)** | POSIX no-op | `TestCopyContainedTreeDereferencesLinks` + 3 real-pnpm cases pass identically on darwin, exit 0. Bound: this hunk is observable only on Windows; its before/after evidence is the hosted pair rev1 run 35481906193 (bare "The system cannot find the path specified." on both cases) → rev2/rev3 pass, plus `TestCopyContainedTreeDereferencesWindowsJunction` passing on windows-latest. |
| **L — ledger narrowing: `TestRealPinnedPNPMLockSupersetSnapshotDependencies` row back to `linux,darwin / windows / host-capability`** | gate-selftest must fail | `gate-selftest: 183 passed, 2 failed`, exit 1 — exactly `FAIL the ledger requires TestRealPinnedPNPMLockSupersetSnapshotDependencies on every runner and tolerates no skip (row: linux,darwin\|windows\|host-capability)` and `FAIL windows records TestRealPinnedPNPMLockSupersetSnapshotDependencies as ledger-refused, not tolerated`; the untouched sibling row stays `ok`. **Killed.** |

Interdiff rev1→rev2 (from the attached patches): production delta is exactly the `copyContainedNode` normalization; tests added:
`copy_unix_test.go` and the three `TestCopyContainedTree*WindowsJunction*` cases. rev3 = rev2 bytes.

## 6. Ledger deferral removed; results honest

`platform-cases.tsv`: both rows `linux,darwin,windows	-	-`; `skip-classes.tsv`: the `stage-deferred` BUG row deleted (comment placeholder only);
`conformance_test.go`: `skipOnWindowsForStoreRegistryGap` and both call sites deleted; `gate-selftest.sh`: deferral pins replaced by no-deferral pins and
skip-fatal behavioural cases; rust wrong-class probe re-pointed to a live `root-unset` reason. No `stage-deferred`/`skipOnWindows…` text survives in any repo
file. Results rev1–rev3 are consistent with the logs: rev1 honestly reported the host SIGKILL condition and static-only mutant review; rev2 replaced rev1's
"copyContainedNode ruled out" claim and the rework's exec-seam theory with the evidence-driven junction diagnosis; rev3 documents the byte-identical republish.

## 7. Observations (non-blocking)

- The seven Windows-only junction unit tests `t.Skipf("this host cannot create a directory junction …")` when `mklink /J` fails (class `host-capability`, allowed)
  and carry no platform-cases row, unlike the godriver junction rows (`must=windows skip=windows host-capability`). On this run they executed. The AC's Windows
  proof is carried by the two ledger-required real-pnpm cases (no tolerated skip), so this is a hardening follow-up (add ledger rows), not a defect.
- pnpmsource's `Materialize`/`DerivePrivateStore` have no production caller outside tests yet (crossconformance uses Parse/CaptureAndAdmit only); the real-pnpm
  conformance cases are the entry-point drivers named by the AC.
