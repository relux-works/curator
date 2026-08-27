# TASK-260720-jrrgw9 — independent verifier 3 results

Date: 2026-07-29  
Role: tester  
Verdict: development handback

## First actionable failure

The required full race gate failed:

```text
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 \
GOTMPDIR=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/verifier3-gotmp/race \
go test -count=1 -race ./...
```

- Real exit: **1**
- Wall time: **610 seconds**
- `internal/install`: default 10-minute package timeout, `FAIL ... 603.306s`
- `internal/install/atomicity`: default 10-minute package timeout, `FAIL ... 603.701s`
- No `-timeout` flag was supplied.
- No `WARNING: DATA RACE` or `DATA RACE` diagnostic appeared. This is not a
  claim that the race suite passed; the timeout makes the gate red.

The `internal/install` alarm happened while
`TestStrictRegistryPolicyFailsUnknown` had been active for only 3 seconds,
which is evidence of cumulative package duration rather than that named test
being stuck.

The atomicity alarm listed:

- `TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder/global-auto`
  active for 8m28s, with child `80-removal` active for 44s.
- `TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder/project-hybrid-auto`
  active for 8m28s, with child `50-env-file` active for 52s.

The task must not be accepted on this evidence. The next producer should reduce
the race-time cumulative duration of `internal/install` and
`internal/install/atomicity` under the unchanged default alarm, without hiding
the failure behind a timeout override or weakening the rollback matrix.

## Green gates before the failure

### Focused authoritative 12-package barrier

The exact accepted verifier-2 package list and test-name filter ran under the
immutable conformance root and a task-owned GOTMPDIR.

- Real exit: **0**
- Wall time: **34 seconds**
- All 12 packages printed `ok`.
- Slowest selected package: `internal/install 34.380s`.

### Exact full repository gate

```text
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 \
GOTMPDIR=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/verifier3-gotmp/full \
go test -count=1 ./...
```

- Real exit: **0**
- Wall time: **444 seconds**
- Every package printed `ok`.
- `cmd/curator`: **384.270s**, leaving 215.730s under the unchanged package
  alarm and establishing that the shared-fixture rework fixed the verifier-2
  full-suite timeout.
- `internal/install`: **341.415s**
- `internal/install/atomicity`: **441.122s**

The same shared-fixture surface also passed under race before the two other
packages timed out: `cmd/curator 557.779s`.

## Gate ledger

| Gate | Real exit | Result |
| --- | ---: | --- |
| Local tool-readiness inventory (`go version`) | 0 | Go 1.25.5 darwin/arm64 |
| Candidate HEAD | 0 | `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8` |
| Exact candidate source delta | 0 | 20 added tests, 3 modified tests |
| Producer reverse reconstruction | 0 | pre-patch SHA-256 reproduced |
| Producer regenerated diff comparison | 0 | byte-identical |
| Authoritative 448-file digest comparison | 0 | byte-identical to verifier 2 |
| Focused process/disk preflight | 0 | stable empty; 27,263,308 KB free |
| Focused authoritative barrier | 0 | all 12 packages green |
| Full process/disk preflight | 0 | stable empty; 27,275,368 KB free |
| Exact `go test -count=1 ./...` | 0 | all packages green in 444s |
| First race process/disk preflight | **1** | external Go run active; race not started |
| Race process/disk retry | 0 | stable empty after external PIDs ended; 26,851,432 KB free |
| Exact `go test -count=1 -race ./...` | **1** | two package timeouts in 610s |
| Final process barrier | 0 | stable empty |
| Candidate post-run delta comparison | 0 | same 23 files |
| Candidate post-run digest comparison | 0 | expected 22 prior digests plus intended status digest |
| Authoritative pre/post digest comparison | 0 | byte-identical |
| Task-owned GOTMPDIR size | 0 | 0 KB after descendants ended |
| Exact GOTMPDIR removal | 0 | removed and verified absent |

The first race preflight is reported as failing, not passing. It found external
PID 55173 (`go test -count=1 ./...`) and its test children in both stability
scans. No verifier race command started until those exact processes were
terminal and a fresh two-scan barrier exited 0.

One non-Go evidence command initially exited 127 because a zsh loop used the
special variable name `path`, replacing `PATH`. The corrected command used
`rel_file` plus absolute utility paths and exited 0; the failed attempt is not
presented as evidence.

## Provenance and immutable inputs

- Candidate:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree`
- Accepted comparison:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-2kaopg/worktree`
- Immutable root:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`
- Candidate HEAD:
  `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`
- Reworked `cmd/curator/status_test.go` SHA-256:
  `487b12bdf531e4714983eab83b804de7b4604513e435256e550f60391ee0d32e`
- Reconstructed pre-patch SHA-256:
  `4c2339dd36854cd51baf28c92c480b4341fa398c10118f41d44b8d29e845afbe`
- Producer diff SHA-256:
  `015d48fb809436a66becefb26b43cf8f19c02a6405e5176529bcac5e12781a3f`
- Accepted matrix SHA-256:
  `3e4e2ee020841a9f45ce11c788f7617b8dd7ec2a64dfcace9fc968c8dbe7e9f2`
- Manifest SHA-256:
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`
- `build-drivers.json` SHA-256:
  `f412c107091cf82f980523afe5361212a3b89a3425f5d885373191f8acb12aea`
- `manager-lifecycle.json` SHA-256:
  `2ddbd2665a63f44dc0e03e060f4cd34bfde219a56b3192511fe1ef81047feedf`
- `external-repository-lifecycle.json` SHA-256:
  `175d709f183775b22ed3db27bc923ca78d6394d93a13096b5d6890c509aab072`

The candidate delta remained exactly 23 test files: 20 candidate-only
conformance tests and three modified tests
(`cmd/curator/status_test.go`, `internal/buildcache/conformance_test.go`, and
`internal/closure/conformance_test.go`). There is no product, schema, golden,
registry, release-pin, or configuration delta.

## Logs and cleanup

- Focused log SHA-256:
  `82543dcdcb5ea353546b0232417649b9b168ac166db5316b66c1bba1ef3ec51c`
- Full log SHA-256:
  `fd412af8d3e07fa9542f6a31e54f1033dde900ae664440db8fd376f988d0b10e`
- Race log SHA-256:
  `e3270767049a88ea669fca9942d8f0262762d52e261f78556858e089319f2aea`

All three task-owned GOTMPDIR subtrees were 0 KB after their descendants
terminated. The exact `verifier3-gotmp/{focused,full,race}` directories and
their empty parent were removed with `rmdir`; absence was verified. Final
available disk was 26,578,276 KB and the two-scan process barrier was empty.

## Conditional gates not run

Native Windows inventory and execution were **not run** because both macOS
gates were required to be green and the race gate failed. `ssh win` was not
touched. Coverage was not run because the verifier instruction forbade it.

## Routing

Route to development. Do not add a timeout override, delete assertions, or
reinterpret this as a green race gate. The exact non-race full suite is now
green; the remaining required work is bounded race-time optimization or test
fixture reuse in the two timed-out install packages, followed by another exact
full/race verifier pass and only then conditional native Windows validation.
