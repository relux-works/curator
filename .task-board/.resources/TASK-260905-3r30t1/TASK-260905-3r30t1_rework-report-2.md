# TASK-260905-3r30t1 rework report 2 — cat-file deadlock on mid-stream refusal

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-acquisition-byte-exact`, branch
`feat/byte-exact-acquisition`, PR https://github.com/relux-works/curator/pull/58. Three signed commits on top of
`a46abc80`, no rewrite, plain pushes (`a46abc80..a97265ad`, `a97265ad..bb14375a`); PR head `bb14375a`:

| Commit | Subject | Signature |
|---|---|---|
| `c9defdb8` | Refuse snapshot entries from the listing and terminate cat-file on mid-stream errors | `Good "git" signature for oparin@me.com` (ECDSA `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`) |
| `a97265ad` | Stage closure scratch snapshots and rename them into place on success | same key, `Good "git" signature` |
| `bb14375a` | Return the watchdog duration before the error in extractBounded (revive `error-return`, caught by the hosted Lint job on `a97265ad`; `golangci-lint run ./...` local exit 0, `0 issues`) | same key, `Good "git" signature` |

Author `Ivan Oparin <oparin@me.com>`; `Co-Authored-By: Claude Fable 5.1` trailer as on the earlier commits.

## Finding → disposition

| Finding | Disposition | Where |
|---|---|---|
| F1 (blocking) deadlock on mid-stream refusal | Fixed. (a) `planWrites` pre-pass over the `ls-tree -r -l -z` listing decides every refusal that needs no blob bytes before `cat-file` starts: `safeTarget` (escapes, `.`/`..`, `.git`), size bound from the listed size, exact duplicate path, case-fold collision (destination probed once, lazily, only when two targets fold to one key; probe removed), pre-existing destination entry (`Lstat`). (b) `writeBlobs.abort`: close stdin → `Process.Kill` → drain stdout to `io.Discard` under `drainBound` (64 MiB) → `Wait` → remove the partial file. Success path drains any trailing bytes before `Wait` too. No path reaches `Wait` with undrained stdout. `Extract` removes what the call wrote on any failure (`RemoveAll` of a destination it created, else the planned targets). Header size must equal the listed size (a disagreement is a framing error). | `internal/gitops/gitops.go` `Extract`, `planWrites`, `destinationFoldsCase`, `writeBlobs` |
| O1 `.git` component admitted | Fixed: `safeTarget` refuses a component equal to `.git` case-insensitively (git `verify_path`); `.gitattributes`, `git/` stay admitted. | `TestExtractRefusesDotGitComponents` |
| O2 closure scratch reuses a partial target | Fixed: `snapshotFor` extracts into `os.MkdirTemp(parent, commit+".extract-*")` and renames into place after `Extract` returns nil; failure removes the staging dir. | `internal/closure/closure.go`, `TestScratchSnapshotRefusalLeavesNoReusablePartialTree` |
| O3 nested `100755` test gap | Added `TestExtractPreservesNestedExecutableBit` (`sub/run.sh` 100755, `sub/deep/exec.sh` 100755, `sub/deep/plain.sh` 100644; tree built with `update-index --cacheinfo` because `mktree` refuses slashes). | ledger row, unix runners |
| O4 `100664` canonicalised by ls-tree | Verified by the reviewer; no change. Recorded. | — |

Additional listing refusal discovered while building the injection: `ls-tree -l` prints `BAD` as the size of an object the repository cannot read, so a missing/unreadable blob is now refused from the listing (`missing or unreadable object in git snapshot`) and never reaches `cat-file` (`TestExtractRefusesMissingObjectsFromListing`).

## Git probes (git version 2.50.1, Apple Git-155)

```
$ git ls-tree -r -l -z b42ca8c… | cat -v
100644 blob 1f6907f8… 8388608	big.bin^@100644 blob 45b983be…      59	small.txt^@
$ git ls-tree -r -l -z <tree referencing an absent oid>
100644 blob bbbbbbbb…     BAD	gone.txt^@
$ printf '%s\n' <truncated-loose-oid> | git cat-file --batch      # loose object file cut in half
1f6907f8… blob 8388608
xxxxxx… (4177974 bytes)  fatal: unable to stream 1f6907f8… to stdout   rc=128
```
The truncated object kills git mid-body (an I/O error for the reader, not a blocked child); the blocked-child shape needs a failure on the Go side or a shim, which is why the suite carries both.

## Watchdog tests (`internal/gitops/deadlock_test.go`, 20 s watchdog each, timings from this host)

| Test | Injection | Refusal | Time | Destination |
|---|---|---|---|---|
| `TestExtractRefusesOversizeBlobWithoutStreaming` | 1 MiB blob, `maxSnapshotFileBytes=512 KiB` | `too large` | 33 ms | nothing written |
| `TestExtractRefusesDuplicatePlatformPathBeforeLargeBlob` | `A.txt`/`a.txt` then 8 MiB | `duplicate platform path` | 32 ms | nothing written (skips `test filesystem is case-sensitive…` on linux) |
| `TestPlanWritesRefusesCaseCollisionBeforeStreaming` | same pair through `planWrites` alone | `duplicate platform path`, probe removed | <1 ms | pins the pre-pass (see M4) |
| `TestExtractRefusesEscapeBeforeLargeBlob` | `..` then 8 MiB (literal tree) | `unsafe path` | 29 ms | nothing written |
| `TestExtractRefusesMissingObjectsFromListing` | absent oid then 8 MiB (`mktree --missing`) | `missing or unreadable object` | 30 ms | nothing written |
| `TestExtractRemovesWrittenFilesOnMidStreamError` | `a-first.txt` good, `zz-big.bin` truncated loose object | `reading blob …: EOF` | 62 ms | `a-first.txt` removed, caller's `keep.txt` intact |
| `TestExtractTerminatesCatFileOnMidStreamFramingError` | POSIX `git` shim on PATH: wrong-oid header then 8 MiB of `/dev/zero` (child alive and blocked) | `unexpected git cat-file --batch response` | 355 ms | nothing written; skips on Windows (`executes a POSIX launcher`, platform-control) |
| `TestExtractTerminatesCatFileOnWriteFailure` | real git; first target under a 0o555 directory, 8 MiB queued behind it | `permission denied` | 47 ms | `zz-big.bin` absent; skips when the process can write through the read-only dir (Windows/root, host-capability) |
| `TestExtractRefusesDotGitComponents` | `.git`, `.git/config`, `sub/.git/HEAD`, `.GIT/config`, `sub/.Git` | `unsafe path` | — | nothing written; `.gitattributes`, `git/.gitkeep` admitted and `EnsureRepo(dest)` fails |
| `TestExtractPreservesNestedExecutableBit` | nested 100755/100644 | — | — | exec bits as committed |
| `TestScratchSnapshotRefusalLeavesNoReusablePartialTree` (closure) | link entry then good commit through `snapshotFor` | link refusal | — | no target, no staging dir; good commit extracts; existing snapshot reused |

Collision gating: the two collision cases are registered `must_run_on=darwin,windows`, `skip_allowed_on=linux` with class
`host-capability` (skip text `test filesystem is case-sensitive…` matches the existing pattern). The case-sensitivity is
detected by probing (`PROBE` after writing `probe`, or `destinationFoldsCase`), not by GOOS.

## Mutant evidence

| Mutant | Narrows the gate to | Named failing test | Result |
|---|---|---|---|
| M1 `abort` = `Wait` only (no kill/drain; the a46abc80 shape) | a child that already exited | `TestExtractTerminatesCatFileOnMidStreamFramingError`, `TestExtractTerminatesCatFileOnWriteFailure` | both FAIL: `Extract did not return within 20s: cat-file deadlock` |
| M2 size bound `> 2*max` | admits blobs up to twice the bound | `TestExtractRefusesOversizeBlobWithoutStreaming` | FAIL (blob written) |
| M3 `.git` compared case-sensitively | admits `.GIT`, `.Git` | `TestExtractRefusesDotGitComponents` | FAIL |
| M4 pre-pass fold collision removed (mid-stream `Lstat` backstop kept) | collision refused only after `cat-file` has started and the first file was written | `TestPlanWritesRefusesCaseCollisionBeforeStreaming` | FAIL. Note: the end-to-end `…BeforeLargeBlob` test survives this mutant because `Extract`'s cleanup removes the first file and the backstop returns in time; the unit-level test is what pins the pre-pass. |
| M5 closure extracts straight into `target` | staging removed | `TestScratchSnapshotRefusalLeavesNoReusablePartialTree` | **SURVIVES**. Bound: with `Extract` now removing everything it wrote on failure, a refused extraction leaves no partial target either way; the staging rename only matters when the process dies between writes (SIGKILL, power loss), which the suite does not simulate. A pre-existing target directory (from a crash of the old code) is still reused by the `os.Stat` check, unchanged from before. |
| M6 `Extract` failure cleanup removed | files written before a mid-stream failure stay | `TestExtractRemovesWrittenFilesOnMidStreamError` | FAIL (after fixing the test: `mktree` sorts entries, so the good file had to sort before the corrupt one; the first version of the test was ineffective and this mutant caught it) |

All mutants applied as scratch edits and reverted (`git diff --stat` back to the committed tree).

## Gates (exact commands, exit codes; run in the worktree at `a97265ad` tree)

| Command | Exit |
|---|---|
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `gofmt -l cmd internal` | 0, no files |
| `go test -count=1 -race -timeout 30m ./internal/gitops ./internal/snapshot ./internal/closure ./internal/interop` | 0 (7.7 s / 2.5 s / 9.5 s / 2.6 s) |
| `CURATOR_CONFORMANCE_ROOT=…/curator-spec/conformance/v1 go test -json -count=1 -timeout 30m $(go list ./... \| grep -v cmd/curator)` | 0, 0 failing actions (59 packages, `.temp/rework2/all-but-cmd.json`) |
| `CURATOR_CONFORMANCE_ROOT=… go test -json -count=1 -timeout 30m ./cmd/curator` | 0, 260 s (`.temp/rework2/cmd-curator.json`) |
| `bash .github/ci/gate-selftest.sh` | 0 — `81 passed, 0 failed` |
| `CURATOR_CONFORMANCE_ROOT=… bash .github/ci/ledger-consistency.sh .temp/rework2/ledger` | 0 — `97 rows checked across linux darwin windows`, every new gitops row `ok` |
| `CI_GATE_GOOS=darwin bash .github/ci/platform-case-gate.sh .temp/rework2/full.json …` (merged stream of the two runs above) | 0 — `platform-case gate: ok`, 19 skips recorded, none in gitops/interop |
| `CI_GATE_GOOS=linux …` same stream | 1 — one FAIL, `internal/godriver :: TestBuildFailsClosedWhenTheGoChildCannotStart` skipped where linux does not tolerate it (darwin stream judged as linux); gitops/interop rows all `ok` |
| `CI_GATE_GOOS=windows …` same stream | 1 — 13 FAILs, all Windows-only cases that never ran on this darwin host (`buildcache`, `runtimestore`, `scopes`, `shell`, `godriver`); gitops/interop rows all `ok` |

The linux/windows local gate runs are a darwin stream evaluated against those ledgers; they prove the gitops/interop
rows, not those hosts. The hosted runners are the proof for the other packages (below). Adapter suites (Cargo/Swift/pnpm/yarn) ran locally inside the 59-package run and passed here; their hosted x86 status is a known separate class.

## Hosted CI (PR #58 at bb14375a, run 33983692562)

All checks green (`gh pr checks 58`):

| Check | Result | Duration |
|---|---|---|
| Lint | pass | 41s |
| Test (ubuntu-latest) | pass | 3m5s |
| Test (macos-latest) | pass | 9m59s |
| Test (windows-latest) | pass | 29m34s |
| Race (ubuntu-latest) | pass | 10m2s |
| Race (macos-latest) | pass | 16m17s |
| Gate self-test (ubuntu/macos/windows) | pass | 8s / 10s / 25s |
| Interop conformance gate | pass | 21s |
| Naming gate | pass | 10s |
| Candidate suite | skipping (not triggered for this PR) | — |

The first push (`a97265ad`, run 33983589928) failed only Lint (`internal/gitops/deadlock_test.go:21:1: error-return … (revive)`); fixed in `bb14375a`. No adapter-suite (Cargo/Swift/pnpm/yarn) failure appeared in either run.

Skip evidence from the run's `skips-observed.tsv` artifacts (every new skip landed in its declared class, `tolerated-by-ledger`):

| Runner | Case | Class | Reason |
|---|---|---|---|
| ubuntu (Test, Race) | `TestExtractRefusesDuplicatePlatformPathBeforeLargeBlob`, `TestPlanWritesRefusesCaseCollisionBeforeStreaming` | host-capability | test filesystem is case-sensitive |
| windows (Test) | `TestExtractPreservesExecutableBit`, `TestExtractPreservesNestedExecutableBit` | platform-control | Windows does not expose portable executable permission bits |
| windows (Test) | `TestExtractTerminatesCatFileOnMidStreamFramingError` | platform-control | the fake git shim executes a POSIX launcher |
| windows (Test) | `TestExtractTerminatesCatFileOnWriteFailure` | host-capability | this process can write through a read-only directory |
| all | `TestConformanceSnapshotAcquisition` | root-content | pinned root publishes no `vectors/snapshot-acquisition.json` |

Nothing in gitops skipped on macOS; the collision, oversize, escape, missing-object, corrupt-object, `.git`, and both kill-and-drain cases ran there and on ubuntu (collision cases excepted).

## Not verified / bounds

- Windows behaviour of the new tests only through hosted CI (see above).
- M5 survives (see table): process-death mid-extraction is not simulated.
- `drainBound` (64 MiB) is a cap on what the abort path discards from an already-killed child; a killed child cannot
  produce more than the pipe buffer, so the cap is never reached in practice and is not tested at its edge.
- Docs: CHANGELOG "Fixed" entry and `docs/implementation-plan.md` line updated; README has no acquisition section.
- The story workspace `.temp/STORY-260905-2qvzwk/worktree` carries an empty delta by design; nothing written to the control root or LOGBOOK.md.
