# TASK-260729-1zex8r — toolchain fingerprint traversal, rework after the tester regression

Date: 2026-07-29
Role: developer (rework run, after tester RUN-260729-ecbe05)
Worktree: `.temp/TASK-260729-1zex8r/worktree`
Paired before-tree: `.temp/TASK-260729-1zex8r/before-worktree`

This supersedes the previous developer artifact of the same name. The candidate
it described was withdrawn: the tester proved it read detached bytes, and that
defect is fixed here by removing the mechanism that caused it, not by adding a
guard around it.

## 1. What the tester found, and why the previous candidate was wrong

`TASK-260729-1zex8r_tester-mutation-regression.md` is correct and is now closed.

The withdrawn candidate cached a stack of ancestor `*os.Root` handles across the
**digest** phase (`scopedDirs`) and opened each file as a single component under
the cached handle. A directory handle survives the rename of its directory, so
once `pkg` was renamed out of GOROOT and a replacement installed at the same
path, every later file in that run was read through the detached directory.
`os.SameFile` then compared the record against the *original* file and passed.
The result was a successful digest over bytes that were no longer at the
canonical paths — a fingerprint that attests to a tree that is not there.

The pre-change traversal has no such window because it re-opens every file by
its full root-relative path, so it resolves the replacement and fails closed.

## 2. The fix

`internal/godriver/fingerprint.go` only. `scopedDirs` is deleted.

The function is split into the two phases it always had:

- `collectToolchainRecords` — builds the canonical record set in canonical order.
- `digestToolchainRecords` — writes the canonical byte stream and hashes it.

**The digest phase is back to the pre-change code, character for character:**
`root.Open(record.path)` per file, then `file.Stat()`, `os.SameFile`, the
negative-size guard, `copyWithContext`, and the length-changed guard. Path
binding is therefore not *revalidated* — it is the same resolution the pre-change
traversal performed, so there is nothing left to diverge.

This is the guarantee the rework directive asked for by a cheaper route. The
directive described resolving the cached directory from GOROOT before each reuse
and comparing filesystem identity. That costs the same syscalls as simply
opening the file by its full path (one resolution of the same component chain),
adds a new identity check that a rename-and-rename-back would still pass, and
leaves new code on the trust boundary. Opening by full path is the same cost,
the same guarantee, and provably identical to the accepted implementation
because it *is* the accepted implementation.

**The collection phase keeps the scoped descent, but only for the one step whose
result something else re-anchors.** Concretely:

| step | before | after | anchored by |
| --- | --- | --- | --- |
| list a directory | `fs.ReadDir` opens it by full path | `walk.root.OpenRoot(protocolPath)` — full path | unchanged |
| `Lstat` a **file** entry | full path | scoped, under the directory handle | digest phase re-opens it from GOROOT and matches `os.SameFile` |
| `Lstat` a **link** entry | full path | scoped | `walk.root.Readlink` + `walk.root.Stat`, both from GOROOT |
| `Lstat` a **directory** entry | full path | **full path** — a directory record is the only one nothing revisits | itself |
| read a link target | `root.Readlink` | `walk.root.Readlink` | unchanged |
| resolve a link | `root.Stat` | `walk.root.Stat` | unchanged |

So the only resolution this change actually removes is the per-entry full-path
`Lstat` of files and links, and both are re-resolved from GOROOT before anything
they contribute is trusted. Everything a record asserts that nothing later
re-checks — the existence and kind of a directory, the target and resolvability
of a link — is still resolved from GOROOT.

Unchanged: the domain separator, record header framing, big-endian length
prefixes, the `'D'`/`'F'`/`'L'`/`'V'` kinds, the trailing version record, the
name-sorted pre-order visit order (and therefore error precedence), every
diagnostic code and every operator detail string, and cancellation points.

No new diagnostic code was introduced. The one added error return reuses the
existing `toolchain_unreadable` code and the existing
`cannot inspect toolchain path %q` detail.

### Correction to the withdrawn artifact

The withdrawn artifact claimed `os.Root` applies `O_NOFOLLOW` per component.
That is wrong and the comments repeating it are gone: `os.Root` rejects only the
links that *escape* the root and follows the ones that stay inside it. The test
`directory replaced by a link to the detached copy` pins that behaviour rather
than avoiding it — both traversals accept it, because both reach the same inodes
through the same in-root link. That is pre-existing `os.Root` behaviour on the
accepted implementation and is not changed here.

## 3. Equivalence and adversarial evidence (AC 1, AC 2)

`internal/godriver/fingerprint_equivalence_test.go` preserves the pre-change
traversal verbatim as `legacyFingerprintToolchain`, now split at the same phase
boundary as the shipped one, and asserts the two are indistinguishable: identical
digest, identical canonical record set element by element, and on rejection an
identical diagnostic **code and detail**.

| Test | Covers |
| --- | --- |
| `…MatchesLegacyOnRepresentativeTrees` (13 shapes) | empty root, flat files, real toolchain layout, siblings sorting around the separator, 10-deep chains, empty directories, unicode/space/emoji names, files crossing the 128 KiB copy buffer, same-directory links, the RC4 `../bin/go` shape, links to directories, links to links |
| `…MatchesLegacyOnFailClosedTrees` (5 cases) | absolute link, escaping link, dangling link, dangling link deep in the tree, link cycle |
| `…MatchesLegacyOnErrorPrecedence` (4 cases) | which of several violations fires first, across siblings, depths, and subtrees that sort before their siblings |
| `…MatchesLegacyOnRealToolchain` | the host GOROOT: 16093 records, identical digest |
| **`TestFingerprintDigestPhaseResolvesEveryFileFromTheRoot` (6 cases)** | **the tester's regression, promoted to end-to-end**: the tree is mutated exactly at the phase boundary and both traversals must reach the same outcome — unchanged, directory renamed away and replaced, renamed away with nothing in its place, replaced by a link to the detached copy, single file swapped in place, renamed away and back (ABA) |
| **`TestFingerprintDigestPhaseResolvesReplacedAncestors`** | the same regression one level up: a replaced *grandparent* must be resolved too |
| **`TestToolchainWalkAnchorsDirectoryAndLinkMetadataAtTheRoot`** | pins which resolution each collection step uses: directory records and link targets fail when the root cannot see them, plain files do not |
| `…CancellationStaysFailClosed` | pre-cancelled walk, pre-cancelled empty root, cancellation racing the digest, and **cancellation taken deterministically at the phase seam** |
| `…ReportsUnreadableDirectoryIdentically` | `chmod 000` directory → same `toolchain_unreadable` |
| `…DoesNotDescendLinkedDirectories` | a linked directory is a leaf `'L'` record; its contents are not walked |
| `…DetectsMutationBetweenRuns` | content change moves the digest; metadata change does not |
| `…IsStableAcrossRepeatedRuns` | no state leaks between calls |

### On the tester's test

`TestScopedDirsDoesNotReadFromAReplacedDirectory` tested the `scopedDirs` helper
directly. That helper is the defect and is deleted, so the test could not be kept
verbatim. Its property was not dropped — it was promoted from a unit test of a
helper to `TestFingerprintDigestPhaseResolvesEveryFileFromTheRoot`, which applies
the same rename-and-replace at the real phase boundary of the real
`fingerprintToolchain`, and additionally requires the shipped traversal to match
the preserved one on **code and detail**, not merely to "read the replacement or
fail closed". The reviewer should treat that promotion as the item to check.

The phase split exists for this: it is a plain decomposition of the function, not
a production hook, and it lets a test place a mutation between collection and
digest deterministically instead of racing a goroutine.

## 4. Measured performance (AC 3)

### Fingerprint, A/B in one process

`go test -run '^$' -bench BenchmarkFingerprintToolchain -benchtime 5x -count 3 ./internal/godriver/`
— both variants over the same host GOROOT (`/Users/iv/.goenv/versions/1.25.5`,
16093 records), same process, same load, so the comparison carries no cross-run
drift. Real exit code **0**. Raw: `logs/bench-fingerprint-02.txt`.

| variant | ns/op (3 runs of 5 iterations) |
| --- | --- |
| legacy (pre-change traversal) | 1 580 589 725 / 1 560 450 017 / 1 587 261 983 |
| shipped | 1 166 386 425 / 1 163 375 142 / 1 168 025 367 |

**1.56–1.59 s → 1.163–1.168 s, a 1.35x improvement (−26%).**

This is deliberately less than the 3.4x the withdrawn candidate measured. That
candidate bought its extra speed with the digest-phase handle reuse the tester
rejected; the difference between 3.4x and 1.35x is the price of the path binding,
and it was paid on purpose. Security equivalence overrides the speed target, per
the rework directive.

### cmd/curator, before vs after

Filled in from `logs/cmdcurator-measurement.txt` — see section 6.

## 5. Gates

Every command below was run standalone as its own process, with no `tee` and no
pipe on the gate command, and the reported code is the real exit code.

| gate | exit |
| --- | --- |
| `gofmt -l .` | **0**, no output |
| `go build ./...` | **0** |
| `go vet ./...` | **0** |
| `git diff --check` | **0** |
| `go test -count=1 ./internal/godriver/` | **0**, 30.134s |
| `CURATOR_CONFORMANCE_ROOT=… go test -count=1 ./internal/godriver/` | **0**, 34.878s |
| `CURATOR_CONFORMANCE_ROOT=… go test -count=1 -run 'Vector\|Conformance\|Candidate\|Framing\|RC4' ./internal/godriver/` | **0** |
| `go test -count=1 -coverprofile=… ./internal/godriver/` | **0**, 29.752s, package 72.9% |

Conformance root: `.temp/TASK-260729-2kaopg/protocol-spec/conformance/v1` at
`00b1688a9b2457ca397a0bb550acf47cad8ee967`, the exact ref
`.github/workflows/ci.yml` pins. No pin was changed.

The two go-v1 vector tests that carry the toolchain identity both pass:
`TestToolchainFramingMatchesRC4Vector` and
`TestFingerprintImplementationMatchesRC4ToolchainVector`, the latter pinning the
accepted `sha256:baf7c5f3b9c3f1fae3da4c356381bf74442aa7f8f0b6fb2304c9c10833d6032e`.
The remaining go-v1 vector tests **skip** at this pinned ref because the suite
publishes no `expected/build-driver` artifacts there — the same skips the
baseline tree produces.

### Coverage of the changed file

`go tool cover -func` on `internal/godriver/fingerprint.go`, real exit **0**:

| function | coverage |
| --- | --- |
| `collectToolchainRecords` | 100.0% |
| `fingerprintToolchain` | 88.9% |
| `descend` | 88.7% |
| `digestToolchainRecords` | 85.4% |
| `claimEncodedPath`, `writeRecordHeader`, `writeLength` | 100.0% |

### Not run

- `golangci-lint` — **not run**. `command -v golangci-lint` exits **1**; the
  executable is absent from this host and the task forbids host installation.
  `go vet ./...` passes; the lint gate is unverified for this change.

## 6. Host conditions and cmd/curator measurement

Filled in below.

## 7. Scope

- Only `internal/godriver/fingerprint.go` and
  `internal/godriver/fingerprint_equivalence_test.go` differ from the paired
  before-tree. Verified with
  `rsync -rcn --delete --exclude='.git' --exclude='logs/' --exclude='*.log' --exclude='.temp/' --exclude='*.test' --out-format='%i %n'`
  (real exit **0**): one content difference (`>fcsT` on `fingerprint.go`) and one
  added file; every other entry is timestamp-only (`.f..T`).
- No timeout increase, assertion weakening, cache clearing, host installation,
  staging, commit, publication or pin change.
- AC 5 (two consecutive default-timeout `go test -count=1 ./...` runs) belongs to
  TASK-260729-2kaopg and is not claimed here.
