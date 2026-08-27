# TASK-260720-jrrgw9 — production integration of accepted performance patches

Date: 2026-07-29
Role: developer (integration run)
Instruction: `TASK-260720-jrrgw9_production-integration-constraint.md`

Result: **both sanctioned patches applied cleanly to the candidate worktree; the
new delta is exactly the 15 allowlisted paths.** No heavy Go gate was run — the
serialized full-gate phase is owned by the Codex tester.

## 1. Precondition

| Gate | Evidence |
| --- | --- |
| `TASK-260729-365r5r` independently accepted/done | board status `done`; `TASK-260729-365r5r_review-verdict-cycle3.md` = **ACCEPTED** |
| Patch A accepted | `TASK-260729-rfrdfo_review-verdict-cycle-3.md` line 24: "Patch A is accepted: the three `internal/install` race runs exited 0 at 232.088s, 235.124s, and 226.191s" |
| `TASK-260729-2afulh` / fixture trim excluded | not applied; `internal/install/atomicity/fixture_test.go` byte-identical pre/post at `e0732e2e3df9adee95321ba28723a878699722747cb231a5309902a56f1f6120` |

Note on rfrdfo board status: the task itself sits in `analysis` because its
**Patch B** (atomicity 480s bar) was rejected in cycle 3. The stored patch file
is the accepted 13-file Patch A artifact, which is what this integration
consumes; the rejected 14th-file fixture trim was never in it.

## 2. Integration target

- Target: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree`
  (immutable candidate worktree, **not** the outer checkout)
- `HEAD` before and after: `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8` (detached, unchanged)
- Index staged paths after integration: **0**
- `git stash list`: **0 entries**
- No `git add`, `commit`, `stash`, `checkout`, `push`, or pin change was executed.

## 3. Patch provenance

| Patch | SHA-256 | Paths |
| --- | --- | --- |
| `TASK-260729-rfrdfo_install-race-timeout.patch` | `8e7a42ec65b937dc24dd7d6857c4fe2863f63f7172302b7b43a71b6b6c3f530f` | 13 test files (11 `internal/install` + 2 `internal/install/atomicity`) |
| `TASK-260729-365r5r_prototype.patch` | `3dbbbfbd06d586442bd08166142934763b217efae42f85506d9c6a258c4c50d2` | `internal/transaction/namespace.go` (modified) + `internal/transaction/namespace_pass_test.go` (new, from `/dev/null`) |

The 365r5r patch SHA matches the digest recorded in its accepted cycle-3 review
verdict (`3dbbbfbd…`). In Patch A every `--- a/` path equals its `+++ b/` path
(no renames, no deletions, no new files).

## 4. Patch applicability (dry run before mutation)

| Command | Real exit | Log |
| --- | --- | --- |
| `git apply --check --verbose <rfrdfo patch>` | **0** | `apply-check-A.log` — 13 "Checking patch …" lines, no fuzz/offset/reject |
| `git apply --check --verbose <365r5r patch>` | **0** | `apply-check-B.log` — 2 "Checking patch …" lines |

## 5. Application

| Command | Real exit | Log |
| --- | --- | --- |
| `git apply --verbose <rfrdfo patch>` | **0** | `apply-A.log` — 13 × "Applied patch … cleanly." |
| `git apply --verbose <365r5r patch>` | **0** | `apply-B.log` — 2 × "Applied patch … cleanly." |

## 6. Pre / post manifests

Manifests cover every tracked + untracked non-ignored path (`git ls-files
--cached --others --exclude-standard`), path-sorted, sha256 per regular file,
sha256 of the link target for symlinks, recursive content digest for the three
nested skill directories. Generator: `manifest.py` (attached in evidence dir).

| Manifest | Entries | SHA-256 of the manifest file |
| --- | --- | --- |
| `manifest-pre.txt` | 358 | `9df11025980fe102c861495f7e6e554810f09ebd3d1c203a624e626396eb6c36` |
| `manifest-post.txt` | 359 | `83b5df8b81ec21f472a90ad082d1ab3464b807968418a7d3d82c8672ff6a2819` |

`diff manifest-pre.txt manifest-post.txt` (`manifest-delta.diff`) reports
**14 modified + 1 added = 15 changed paths and nothing else**.

`git status --porcelain` delta: exactly one new line, ` M internal/install/registry_e2e_test.go`.
The other 14 paths were already inside untracked candidate regions
(`?? internal/install/…`, `?? internal/install/atomicity/`, `?? internal/transaction/`)
or already-modified tracked files, so their status lines did not change class.

## 7. Final new delta — exactly 15 paths

| # | Path | pre sha256 | post sha256 |
| --- | --- | --- | --- |
| 1 | `internal/install/atomicity/activation_test.go` | `438f7d1f…` | `4db13b99…` |
| 2 | `internal/install/atomicity/commit_atomicity_test.go` | `d67f438e…` | `b62e9aa5…` |
| 3 | `internal/install/cache_conformance_test.go` | `cf1e825a…` | `23847e7b…` |
| 4 | `internal/install/commit_test.go` | `e0567378…` | `5b3a678a…` |
| 5 | `internal/install/diagnostics_test.go` | `6a9cc230…` | `2531ccf7…` |
| 6 | `internal/install/dryrun_conformance_test.go` | `975c9504…` | `5aaf99f4…` |
| 7 | `internal/install/generation_test.go` | `97ea119b…` | `3c685f64…` |
| 8 | `internal/install/install_test.go` | `a2c4a780…` | `eb56b33b…` |
| 9 | `internal/install/maintenance_test.go` | `bee865df…` | `a7443962…` |
| 10 | `internal/install/private_test.go` | `812199e3…` | `1257d61d…` |
| 11 | `internal/install/registry_e2e_test.go` | `173a6d9d…` | `d1b0fe01…` |
| 12 | `internal/install/revalidation_test.go` | `95868902…` | `4d1baa8c…` |
| 13 | `internal/install/stage_test.go` | `849dee33…` | `bc5a867a…` |
| 14 | `internal/transaction/namespace.go` | `997d53df…` | `bb332038…` |
| 15 | `internal/transaction/namespace_pass_test.go` | *(absent)* | `3611f04f…` |

Full-length digests: `target15-pre.txt`, `target15-post.txt`.

Cross-check against the accepted 365r5r cycle-3 verdict: the pre-state
`namespace.go` = `997d53df…` is exactly the baseline the prototype was built
from, and the post-state `namespace.go` = `bb332038…` and
`namespace_pass_test.go` = `3611f04f…` are exactly the accepted prototype
digests. The integration reproduces the reviewed bytes, not a re-derivation.

## 8. Constraint conformance

| Constraint | Check | Result |
| --- | --- | --- |
| Rejected cross-save cache/state tokens | `namespaceGraphAccepted`, `acceptNamespaceGraph`, `forgetNamespaceGraph`, `namespaceChecked`, `namespaceGraph`, `namespaceMu` — scanned in both patch files **and** in the applied `internal/transaction/` tree | **0 hits each, all three scopes** |
| No `TASK-260729-2afulh` / `fixture_test.go` change | `internal/install/atomicity/fixture_test.go` absent from the manifest delta | unchanged |
| No spec / conformance vector / Makefile / workflow / pin / module change | none present in the manifest delta; `go.mod`, `go.sum`, `Makefile`, `.github/`, `conformance/` all untouched | clean |
| No journal schema or Engine field change | `internal/transaction/engine.go` and `journal.go` absent from the manifest delta | unchanged |
| No timeout inflation, no skip introduction | patch-wide scan for `Skip`/`SkipNow`/`timeout`/`Timeout`/`Deadline` on `+`/`-` lines returns a single **comment** line ("exceed its race-build timeout, so it is traded for the partition below") and no code change | clean |
| No production file outside `namespace.go` | the only non-`_test.go` path in the delta is `internal/transaction/namespace.go` | satisfied |
| Immutable conformance root preserved | `conformance/v1/vectors/build-drivers.json` = `f412c107091cf82f980523afe5361212a3b89a3425f5d885373191f8acb12aea` (matches the authoritative digest in the task notes); `conformance/v1/manifest.json` = `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c` | untouched |
| Outer working tree / user / board files preserved | outer `git status --porcelain` shows only the pre-existing untracked `LOGBOOK.md`, `diagrams/`, `task-board.config.json`, `.task-board/…`, `.temp/…`; `.temp/` is gitignored (`.gitignore:8`) | preserved |

## 9. Candidate delta vs the accepted integrated comparison

`diff -rq --exclude=.git .temp/TASK-260729-2kaopg/worktree .temp/TASK-260720-jrrgw9/worktree`
(`candidate-vs-2kaopg.txt`) lists only:

- the pre-existing jrrgw9 candidate delta (the new `*_conformance_test.go`
  files across `buildcache`, `buildmeta`, `buildsource`, `godriver`,
  `runtimestore`, `scopes`, `skillcheck`, `skillspec`, `whitelist`,
  `cmd/curator/lifecycle_conformance_test.go`, plus modified
  `cmd/curator/status_test.go`, `internal/closure/conformance_test.go`,
  `internal/buildcache/conformance_test.go`), and
- the 15 newly integrated paths from section 7.

No unexpected divergence. `diff` also reported "Directory loop detected" for the
three `skill-go-testing-tools` skill symlinks on **both** sides symmetrically;
the manifest generator hashes those entries explicitly and they are identical
pre/post.

## 10. Light validation actually run

Only non-compiling, non-test checks were run, per the instruction to leave the
serialized full-gate phase to the Codex tester.

| Command | Real exit | Result |
| --- | --- | --- |
| `gofmt -l` over all 15 delta paths | **0** | no path printed — every integrated file is gofmt-clean and parses |
| `git diff --check` (tracked modifications) | **0** | no whitespace error |
| duplicate top-level symbol scan in `internal/transaction/*.go` | n/a | the only duplicates are the pre-existing build-tag-separated `namespace_case_{darwin,windows,other}.go` and `durable*`/`sync*` platform variants; none of the 10 new `namespace_pass_test.go` top-level names collides |
| removed-symbol leftover scan in `internal/install/atomicity` | n/a | `sharedUserHome` = 0 hits (fully removed); `globalUserHomes` plural consistently used by the new partitioned sweep |

## 11. Gates deliberately NOT run (and why)

Truthfully **not executed** in this run, not claimed as passing:

- `go test ./...`, `go test -race ./...`
- focused `internal/install` and `internal/install/atomicity` race repetitions
- `go build`, `go vet`, `golangci-lint`, coverage
- any Windows / native-runner validation

Reason: the production-integration constraint states "Run no heavy Go gate in
the developer integration run because a Codex tester owns the serialized
full-gate phase." Whether the combined Patch A + namespace optimization brings
`internal/install/atomicity` under the authoritative 480s race bar is therefore
an **open, unmeasured question in this run**. The 365r5r cycle-3 verdict
recorded 84/76/75s race atomicity for the prototype in its own tree, and the
rfrdfo cycle-3 verdict recorded 591/560/564s for Patch B alone before that
optimization existed; the combination has not been measured here.

## 12. Environment

- Host disk free before: 21 GiB; after: 21 GiB (evidence dir 140 KiB)
- No shared Go cache was cleared; no host software was installed
- No process barrier violation: no Go build/test process was started

## 13. Evidence files (`.temp/TASK-260720-jrrgw9/integration/`)

`manifest.py`, `manifest-pre.txt`, `manifest-post.txt`, `manifest-delta.diff`,
`status-pre.txt`, `status-post.txt`, `target15-pre.txt`, `target15-post.txt`,
`patchA-paths.txt`, `apply-check-A.log`, `apply-check-B.log`, `apply-A.log`,
`apply-B.log`, `candidate-vs-2kaopg.txt`.
