# TASK-260905-3r30t1 review verdict — CR-TASK-260905-3r30t1-5 rev 5 (curator head bb14375a)

**Verdict: ACCEPTED.** `repeat-of: none` (cycle-1 F1 is fixed and independently reproduced as fixed; no
finding of this cycle repeats a prior one).

Reviewer run `RUN-260905-2d5b2b`. Subject: curator worktree
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-acquisition-byte-exact`, branch
`feat/byte-exact-acquisition` @ `bb14375a`, 7 commits over curator main `74c35b1c`,
PR https://github.com/relux-works/curator/pull/58 (`headRefOid` = `bb14375ad41ad57aec98e4966a430665c118ca5c`,
equal to the local head; every hosted check green). Authority: curator-spec `f39f4a9` (ec695ba+)
`protocol/environments.md` §1.2, `conformance/v1/vectors/snapshot-acquisition.json`, `core.md` §2/§6.2/§6.5/§8.
All my scratch work is under the worktree's `.temp/review-3/` (a full rsync copy of the tree with my own
test files added); the worktree itself is untouched — `git status --short` shows nothing outside `.temp`.

## Why an empty repository delta is the right outcome for this leaf

The Change Request candidate tree `c3e47989` is identical to its base `f39f4a9`;
`git diff --stat f39f4a9 c3e47989` prints zero paths and the patch resource hashes to the empty-file sha256.
That is correct here, not a failure:

- The leaf's deliverable is a **curator** change (object-database extraction in `internal/gitops`, both
  callers, tests, CI ledger rows). It lives on `feat/byte-exact-acquisition` / PR #58, verified above.
- The **curator-spec** story workspace is the repository this Change Request is cut from. Its side of the
  work — environments §1.2, `vectors/snapshot-acquisition.json`, `expected/byte-exact-snapshot_sha256.txt`,
  `fixtures/byte-exact/*` — was already landed before this leaf started; I confirmed all five fixture files
  and the expected hash are present at the base commit and that the curator testdata copies are byte-identical
  to them (table below). A spec-repo delta from this leaf would mean the producer had edited the authority
  it was implementing against, which is exactly what the brief forbade.
- The producer's `TASK-260905-3r30t1_cr-publish-report.md` run was a deliberate no-edit republish after a base
  refresh staled the earlier CR. Revision 5 is that same empty shape, not silent non-delivery.

So the emptiness is a workspace-topology fact. The work exists, I attacked it, and the evidence is below.

## What I attacked, and what happened

Everything below is my own execution against `bb14375a`, not a re-reading of the producer's report.
My adversarial tests are `internal/gitops/zz_rev3_adv_test.go`, `zz_rev3_stall_test.go`,
`zz_rev3_diverge_test.go` in `.temp/review-3/curator-copy/` (not committed).

### The cycle-1 blocking deadlock (F1) is gone — reproduced, and the fix is load-bearing

I rebuilt all three cycle-1 hang shapes from scratch under an independent 20 s watchdog, plus a fourth the
producer did not write (a POSIX `git` shim that answers `cat-file --batch` with a valid header naming the
**wrong** oid and then floods 32 MiB while staying alive):

| My test | Shape | Result |
|---|---|---|
| `TestRevHangSingleOversize` | 1 MiB blob, `maxSnapshotFileBytes` narrowed to 512 KiB | refused in **33 ms**, `file too large in git snapshot: "big.bin"`, destination empty |
| `TestRevHangDuplicateThenBig` | `A.txt`/`a.txt` on APFS with an 8 MiB blob queued behind | refused in **32 ms**, `duplicate platform path`, destination empty |
| `TestRevHangEscapeThenBig` | `..` entry with an 8 MiB blob queued behind | refused in **31 ms**, `unsafe path`, destination empty |
| `TestRevShimFloodsAfterGoodHeader` | live shim: wrong-oid header + 32 MiB flood + `sleep 60` | refused in **414 ms**, `unexpected git cat-file --batch response`, destination empty |

The fix is load-bearing, not incidental. Mutant **MU1** (my own): collapse `writeBlobs.abort` back to the
`a46abc80` shape — `stdin.Close(); cmd.Wait()`, no `Process.Kill`, no bounded drain — and the deadlock
returns, caught by two named committed tests **and** my independent one:

```
--- FAIL: TestExtractTerminatesCatFileOnMidStreamFramingError (20.07s)
--- FAIL: TestExtractTerminatesCatFileOnWriteFailure         (20.16s)
--- FAIL: TestRevShimFloodsAfterGoodHeader                   (20.28s)
```

I read `planWrites`/`writeBlobs` line by line for the property the fix claims: every `return` after
`cmd.Start()` goes through `abort`, which closes stdin, kills the process, drains stdout through
`io.LimitReader(reader, drainBound)` and only then calls `Wait`. The success path drains before `Wait` too.
**No path reaches `Wait` with undrained stdout.** The `ls-tree -l` size pre-pass refuses oversize entries from
the listing — the 1 MiB blob above is never streamed at all.

I separated one shape the watchdog would otherwise have mislabelled: a child that **stalls** mid-body (shim
writes a short body, then sleeps) is not the F1 class. `TestRevStallIsChildLifetimeBound` shows `Extract`
returns when the child exits (10.4 s for a 10 s sleep) with `reading blob …: EOF` and an empty destination.
There is no independent I/O timeout anywhere in `internal/gitops` — `run()` has none either — so this is a
pre-existing package-wide property, not something this change introduced or regressed. Recorded as an
observation, not a finding.

### Narrowing mutants — four of mine, all killed by named committed tests

I did not accept the producer's mutant table on trust; I applied my own and reverted each
(`diff -q` back to the committed file after every one).

| Mutant | The gate is NARROWED to | Named committed test that fails |
|---|---|---|
| MU1 `abort` → `Wait` only | a child that already exited | `TestExtractTerminatesCatFileOnMidStreamFramingError`, `TestExtractTerminatesCatFileOnWriteFailure` |
| MU2 `strings.EqualFold(c, ".git")` → `c == ".git"` | admits `.GIT`, `.Git`, `.GiT` | `TestExtractRefusesDotGitComponents` |
| MU3 `entry.size > maxSnapshotFileBytes` → `> 2*max` | admits blobs up to twice the bound | `TestExtractRefusesOversizeBlobWithoutStreaming` |
| MU7 mode allow-list widened to admit `120000` | symlink blobs written as regular files | `TestArchiveRejectsLinks` |
| MU4 whole `Extract` reverted to `git archive` + tar (the pre-change behaviour) | — (the brief's required negative) | `TestExtractReproducesByteExactVector` (both autocrlf sub-tests), `TestExtractIgnoresWorkingTreeConversion`, and 10 more |

MU4 is the one the brief asked for and it bites hard: reverting to `git archive` fails the byte-exact vector
under both `core.autocrlf` settings. The suite is not green around a fake.

I also reproduced the producer's self-declared **survivor** M5 (closure staging removed → the named test still
passes, because `Extract` now cleans up after itself either way) and confirmed it survives exactly as reported.
A producer that reports its own surviving mutant instead of hiding it is reporting honestly.

### Refusals, attacked directly

Each of these was driven through `Extract` — the exported entry point both production callers use
(`internal/snapshot/snapshot.go:48`, `internal/closure/closure.go:423`; I grepped: no `Archive(` symbol
remains anywhere outside tests) — with an 8 MiB blob queued behind the refusal where the shape allowed it,
and I asserted an **empty destination** after every one:

- `120000` symlink → `links in git snapshots are unsupported` (35 ms)
- `160000` gitlink → `unsupported entry type in git snapshot (submodule)` (37 ms)
- `.git`, `.GIT`, `.Git/config`, `sub/.git/HEAD`, `sub/.GiT` → `unsafe path` (33–38 ms each)
- `.gitattributes`, `.gitignore`, `git/keep.txt` → **admitted**, exactly the four expected entries. The
  `.git` refusal did not over-fire on legitimate names.
- Unicode NFD/NFC collision (`café.txt` vs `café.txt`) with 8 MiB queued behind — a shape the
  producer's suite does not cover — → `duplicate platform path`, 56 ms, destination empty. The pre-pass fold
  key (`strings.ToLower`) does not see this pair, so it lands on the mid-stream `Lstat` backstop; that
  backstop refuses (it does not overwrite: the open is `O_EXCL`) and aborts cleanly. **The bounded-refusal
  property holds on a path the producer never tested.**

One shape I expected to be a finding and is not: a raw tree entry with mode `100600`/`100664`/`100777`/`100000`
is **canonicalised by git itself** before we ever see it. I proved it on the installed git:

```
$ git ls-tree -r -l -z --full-tree e4a5a7e1   # raw tree literally contains 100600/100664/100777/100000
100644 blob c1b0730e… 1  aaa      # was 100600
100644 blob c1b0730e… 1  bbb      # was 100664
100755 blob c1b0730e… 1  ccc      # was 100777
100644 blob c1b0730e… 1  ddd      # was 100000
$ git cat-file tree e4a5a7e1 | cat -v   # confirms the stored modes really are the odd ones
```

So the mode allow-list cannot be bypassed by a crafted tree, and the written mode is deterministic. This
extends cycle-1's O4 (which only covered `100664`) and is why row 4 of the coverage table below is a stated
bound rather than a gap.

### The spec vector reproduces, and the conformance test's skip is honest

| Fixture (curator-spec `conformance/v1/fixtures/byte-exact/`) | sha256 | curator `internal/gitops/testdata/byte-exact/` |
|---|---|---|
| `crlf.txt` | `c8dba689…68b2` | identical |
| `lf.txt` | `4fdbc441…2996` | identical |
| `mixed.txt` | `c76a5bc3…a279` | identical |
| `subst.txt` | `ec9a6c8c…22bc` | identical |
| `.gitattributes` | `ced60fac…dc9c` | identical (stored as `gitattributes.fixture`) |

`git ls-files --eol internal/gitops/testdata/` → `crlf.txt i/crlf w/crlf attr/-text`,
`mixed.txt i/mixed w/mixed attr/-text`. The protection is the repository's **pre-existing** root
`.gitattributes` (`testdata/** -text`, `**/testdata/** -text`) — the change did not need to touch it, and the
full diff confirms it did not.

- `go test -run 'ByteExact|SnapshotAcquisition' ./internal/gitops ./internal/interop` with no conformance
  root: `TestExtractReproducesByteExactVector` PASS for `autocrlf=true` and `autocrlf=false`;
  `TestConformanceSnapshotAcquisition` SKIP (`CURATOR_CONFORMANCE_ROOT is not set`).
- With `CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1`:
  `byte-exact-snapshot/autocrlf=true` and `/autocrlf=false` both PASS against
  `sha256:500ea934403d10a2a0b6b7e8874790e489ee002328d3dc0edbda2fe5be2bced0`.

I attacked the skip specifically, because "absent" and "could not read" are different facts and a fallback
defined for absence must not fire on a read failure. Three roots, three outcomes:

| Root | Outcome |
|---|---|
| vector file removed | **SKIP** — `publishes no vectors/snapshot-acquisition.json (pre-environments suite; root-content)`, which matches the pre-existing `root-content` pattern `publishes no ` in `.github/ci/skip-classes.tsv:45`. No new skip class was invented; `git diff 74c35b1c..bb14375a -- .github/ci/skip-classes.tsv` is empty. |
| vector file `chmod 000` | **FAIL** — `reading …: permission denied`. Correct: a read failure is not an absence. |
| vector file truncated to `{"cases": [` | **FAIL** — `decoding …: unexpected end of JSON input`. |

The test's guard is `errors.Is(err, fs.ErrNotExist)` only; every other error is `t.Fatalf`. It also cross-checks
each fixture's on-disk bytes against the vector's per-file `sha256`/`bytes` and the expected file against the
vector's `expected_sha256`, so a normalizing checkout fails loudly instead of passing quietly.

### Byte-exactness itself

`TestRevIgnoresAutocrlfAndExportSubst` (mine): a scratch repo with `* text=auto eol=crlf` **and**
`*.txt export-subst` committed in `.gitattributes`, extracted under `core.autocrlf` = `true`, `false`, **and**
`input`. `lf.txt` keeps LF, `crlf.txt` keeps CRLF, `subst.txt` keeps the literal `$Format:%H$`. Under MU4
(git archive) this test fails. The extraction never invokes `--filters` or `--textconv`, never runs `archive`
or a checkout, and `EnsureRepo` runs first with fixed flags and `-C repo`.

### Gate changes

`.github/ci/platform-cases.tsv` +10 rows, and `TestArchiveRejectsLinks` widened from
`linux,darwin` (windows skipped, `host-capability`, "needs `ln`") to **all three GOOS**. The claim is
justified — the test now builds the `120000` entry through `commitRawTree` and needs neither `ln` nor host
symlink support — and it is **proved, not asserted**: `Test (windows-latest)` is green at `bb14375a`, so the
case actually ran there. The name is now a misnomer (nothing archives), but the ledger requires that exact
string and the rework brief explicitly allowed keeping it. `bash .github/ci/gate-selftest.sh` → exit 0,
`81 passed, 0 failed`.

### Closure staging (cycle-1 O2)

`snapshotFor` now extracts into `os.MkdirTemp(dir(target), commit+".extract-*")` and renames into place only
on success. I checked the new failure mode this shape can introduce — `rename` onto an existing non-empty
directory fails with `ENOTEMPTY` — and it is not reachable: `snapshotFor` has exactly one call site
(`closure.go:303`), there are no goroutines in the package, a repeated source+commit short-circuits on
`os.Stat(target)`, and `ScratchRoot` is a per-invocation private directory (`private.dir("closure-")` at
`internal/install/install.go:358` and `global.go:114`), so no two processes share it. No race introduced.

### Gates I ran myself

| Command | Result |
|---|---|
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `gofmt -l cmd internal` | exit 0, no files |
| `golangci-lint run ./...` | exit 0, `0 issues.` |
| `CURATOR_CONFORMANCE_ROOT=…/curator-spec/conformance/v1 go test -count=1 -race -timeout 25m ./internal/gitops ./internal/snapshot ./internal/closure ./internal/interop` | all `ok` — gitops 6.961s, snapshot 1.944s, closure 8.599s, interop 2.389s |
| `bash .github/ci/gate-selftest.sh` | exit 0, `81 passed, 0 failed` |
| `./cmd/curator` | **not rerun** — citing the producer's 260 s run (exit 0) at this head, plus green hosted `Test`/`Race` on ubuntu/macos/windows |

Hosted CI at `bb14375a` (run 33983692562): Lint, Test (ubuntu/macos/windows), Race (ubuntu/macos), Gate
self-test ×3, Interop conformance gate, Naming gate — **all pass**. No adapter-suite (Cargo/Swift/pnpm/yarn)
failure in this run, so the known x86 redness does not even need to be discounted here.

### Commits and scope

Seven commits over `74c35b1c`, every one `git log --format='%G?'` = **G** (good signature) for `oparin@me.com`,
author `Ivan Oparin <oparin@me.com>`, `Co-Authored-By: Claude Fable 5.1` trailer (acceptable for this
repository). The cycle-1 series `f855a34c…a46abc80` is byte-for-byte the series I reviewed before — **no
rewrite**. Diff surface is 16 files: `internal/gitops` (impl + 3 test files + 5 testdata), the two caller
one-liners, `internal/interop` conformance test, `internal/closure` staging + test, the CI ledger, CHANGELOG,
and one `docs/implementation-plan.md` line. Callers change nothing but the extraction call. README has no
snapshot-acquisition section to update.

## AC-row coverage

**19 of 20 AC rows driven by a named committed test through the production entry point; 1 declared a stated
bound with proof of unreachability.** Production call site for every row: `gitops.Extract`, invoked at
`internal/snapshot/snapshot.go:48` (skills snapshot cache) and `internal/closure/closure.go:423` (closure
scratch snapshots).

| # | AC row | Named driving test |
|---|---|---|
| 1 | Object-database extraction, no autocrlf/text/eol/filters/export-subst | `TestExtractReproducesByteExactVector`, `TestExtractIgnoresWorkingTreeConversion` |
| 2 | Symlink `120000` refused | `TestArchiveRejectsLinks` |
| 3 | Gitlink `160000` refused | `TestExtractRefusesSubmodules` |
| 4 | Any other entry mode refused | **stated bound** — unreachable through real git: `ls-tree` canonicalises every `100xxx` blob mode to `100644`/`100755` (probe above), and `-r` never emits a tree entry. The `kind != "blob"` half is driven by row 3. |
| 5 | Path escapes, `.`, `..`, absolute, empty | `TestExtractRefusesEscapingPaths`, `TestExtractRefusesEscapeBeforeLargeBlob` |
| 6 | `.git` component, any case | `TestExtractRefusesDotGitComponents` |
| 7 | Exact duplicate path | `TestExtractRefusesDuplicatePlatformPaths` |
| 8 | Duplicate platform path (case fold) refused before streaming | `TestExtractRefusesDuplicatePlatformPathBeforeLargeBlob`, `TestPlanWritesRefusesCaseCollisionBeforeStreaming` |
| 9 | Pre-existing destination entry refused | `TestExtractRefusesExistingDestinationEntries` |
| 10 | Oversize blob refused from the listing, never streamed | `TestExtractRefusesOversizeBlobWithoutStreaming`, `TestExtractRefusesOversizeBlob` |
| 11 | Missing/unreadable object (`BAD`) refused from the listing | `TestExtractRefusesMissingObjectsFromListing` |
| 12 | `100755` exec bit preserved | `TestExtractPreservesExecutableBit` |
| 13 | Nested `100755` exec bit preserved (cycle-1 O3) | `TestExtractPreservesNestedExecutableBit` |
| 14 | Suspicious commit operand refused | `TestExtractRefusesSuspiciousCommit` |
| 15 | Mid-stream framing error: child killed and drained before `Wait` | `TestExtractTerminatesCatFileOnMidStreamFramingError` |
| 16 | Write failure mid-stream: child killed and drained before `Wait` | `TestExtractTerminatesCatFileOnWriteFailure` |
| 17 | Files written before a mid-stream failure are removed | `TestExtractRemovesWrittenFilesOnMidStreamError` |
| 18 | Conformance vector via `CURATOR_CONFORMANCE_ROOT`; honest skip on a root without it | `TestConformanceSnapshotAcquisition` |
| 19 | Closure scratch: refused extraction leaves no reusable partial tree (cycle-1 O2) | `TestScratchSnapshotRefusalLeavesNoReusablePartialTree` |
| 20 | Both callers on the new path, behaviour otherwise unchanged | `TestExtractProducesExactTree` + `./internal/snapshot`, `./internal/closure` suites; no `Archive(` symbol remains |

## Findings

### N1 — non-blocking, follow-up: the platform-path fold gate does not cover DIRECTORY components

File: `internal/gitops/gitops.go`, `planWrites` — `key := strings.ToLower(target)`.

The new collision gate keys on the folded **full path**, so it catches `A.txt` vs `a.txt`, but two tree paths
whose *directory* components fold together while their basenames differ produce no duplicate key and are
admitted. On a case-folding destination both files land in one physical directory.

Reproduced on one host, same commit, two destinations (I attached a case-sensitive APFS volume via `hdiutil`
so both extractions ran on the same machine, same git, same code):

```
commit 0fb009b385c1324492e239c4bac78a49268f3f3c   # tree: Dir/x.txt (100644), dir/y.txt (100644)
case-folding   dest  [Dir/ Dir/x.txt Dir/y.txt]           -> sha256:1f97a4cb3b7f0d8c…1b5e57d2
case-sensitive dest  [Dir/ Dir/x.txt dir/ dir/y.txt]      -> sha256:5419409210d7e39d…4ac90ebc
```

The snapshot content hash for one commit therefore depends on the destination filesystem, and `Extract`
reports success either way. Under a plain reading of core §2 ("Filesystem extraction MUST detect two protocol
paths that map to one platform path and fail before writing") the directory pair `Dir` / `dir` is such a case.

**Why this does not block acceptance.** It is not a regression and not introduced here: the `git archive` path
this change replaces had *no* collision detection at all (it used `O_TRUNC` and silently let the later entry
win), and `git checkout` of the same commit on macOS merges the two directories identically — curator matches
git. Downstream the failure is closed, not silent: a lock hash produced on one filesystem class simply
mismatches on the other. The brief scoped this gate to "reuse whatever the current tar path does for duplicate
platform paths", i.e. nothing; the producer shipped considerably more than that and the hole is in the extra
mile, not in the delivery. I also checked the package doc for an overclaim and there isn't one — "the snapshot
is therefore a function of the commit alone" is scoped by its own `therefore` to the conversion axis, and the
Refused list says "two tree paths that map to one platform path", which `Dir/x.txt` and `dir/y.txt` do not.

**Suggested fix for the follow-up** (~10 lines in `planWrites`): fold every path *prefix*, not just the full
path — keep a `map[foldedPrefix]originalPrefix` and refuse when a folded prefix is already claimed by a
different original. That covers directory folds and the existing file case with one rule. Pair it with a
committed test asserting the two-hash divergence above cannot occur.

### O1 — observation: no independent I/O timeout on the git child (pre-existing, package-wide)

`TestRevStallIsChildLifetimeBound`: a `cat-file --batch` child that stalls mid-body blocks `Extract` for the
child's lifetime (10.4 s for a 10 s stall) before returning the typed error with an empty destination. This is
**not** the F1 class — no deadlock, and the abort path is never reached because no error has occurred yet.
`run()` (clone/fetch/rev-parse/ls-tree) has the same property and always did. Not a finding against this
change; recorded so a future reader does not mistake it for one.

### O2 — observation: git canonicalises tree modes (extends cycle-1 O4)

`100600`, `100664`, `100000` → `100644`; `100777` → `100755`, at `ls-tree` time. The mode allow-list cannot be
bypassed by a crafted tree and the written permission bits are deterministic. Probe pasted above.

### O3 — observation: refusal residue in a pre-existing destination

When `destination` already existed, `Extract`'s cleanup removes the planned target *files* but leaves the
intermediate directories `MkdirAll` created. `TestRevRefusalResidueInPreexistingDest`: after a refusal the
caller's `keep.txt` survives (correct) and no snapshot file remains (correct). Both production callers extract
into a directory they create and destroy themselves (`snapshot.Get` into `MkdirTemp` removed by `defer`;
`closure.snapshotFor` into a staging dir removed on failure), so no production path observes the residue.

## Not verified

- Windows behaviour of any of it beyond the green hosted `Test (windows-latest)` run at this head.
- `./cmd/curator` not rerun by me (cited, per the brief).
- The local `CI_GATE_GOOS=linux|windows` ledger evaluations — I relied on the green hosted runners for the
  non-darwin GOOS values rather than judging a darwin `go test -json` stream against a foreign ledger.
- N1's behaviour on Windows short-name (`8.3`) or trailing-dot folding; I only demonstrated case folding.
