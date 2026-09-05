# TASK-260905-3r30t1 drafting report: byte-exact object-database extraction

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-acquisition-byte-exact`,
branch `feat/byte-exact-acquisition`, base curator main `74c35b1c`. Not pushed, no tag, no PR.
Authority: curator-spec `ec695ba` (`protocol/environments.md` §1.2, `vectors/snapshot-acquisition.json`,
`fixtures/byte-exact/*`, `expected/byte-exact-snapshot_sha256.txt`; core §2, §6.2, §6.5, §8).

## Commits (both signed, `git log --show-signature`)

```
commit 5beced46b811e005dabcc5f71deb736593e334bf  Test byte-exact snapshot acquisition against the spec vector
Good "git" signature for oparin@me.com with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM
Author: Ivan Oparin <oparin@me.com>
commit f855a34cabce16c9fce41237e78fccdce8f0f518  Extract git snapshots from the object database instead of git archive
Good "git" signature for oparin@me.com with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM
Author: Ivan Oparin <oparin@me.com>
```

Files: `internal/gitops/gitops.go` (Archive removed, `Extract` added; package doc updated),
`internal/snapshot/snapshot.go:48` and `internal/closure/closure.go:413` switched to `gitops.Extract`
(no other caller behavior changed), `internal/gitops/gitops_test.go` (renamed two tests to Extract),
`internal/gitops/byteexact_test.go` (new), `internal/gitops/testdata/byte-exact/*` (new),
`internal/interop/snapshot_acquisition_test.go` (new), `CHANGELOG.md` (Unreleased/Fixed),
`docs/implementation-plan.md` (gitops line). `README.md` has no snapshot-acquisition section; untouched.

## Extraction design (`gitops.Extract(repo, commit, destination)`)

1. `EnsureRepo(repo)`; refuse empty or dash-prefixed `commit`.
2. `git -C <repo> rev-parse --verify <commit>^{tree}`.
3. `git -C <repo> ls-tree -r -z --full-tree <tree>`; framing parsed is
   `<mode> <type> <oid>\t<path>\0` per record (paths unquoted under `-z`).
   Refusals at listing time: mode `120000` → `links in git snapshots are unsupported: <path>`;
   mode `160000` or type `commit` → `unsupported entry type in git snapshot (submodule): <path>`;
   any other non-blob or mode other than `100644`/`100755` → `unsupported entry type in git snapshot`.
4. One `git -C <repo> cat-file --batch` process; oids written on stdin, responses parsed as
   `<oid> blob <size>\n<bytes>\n`. Header oid/type must match the request; `size > maxSnapshotFileBytes`
   (512 MiB, now a `var` so tests can narrow it) → `file too large in git snapshot`. No `--filters`,
   no `--textconv`, working tree and attributes never consulted.
5. Per entry, before writing: path check (`safeTarget`: empty, `/`-prefixed, backslash, empty/`.`/`..`
   component, or `filepath.Rel` escape → `unsafe path in git snapshot`); duplicate detection: an
   `Lstat` hit at the target (either a prior write of this extraction on a case-insensitive
   filesystem, or a pre-existing file) → `duplicate platform path in git snapshot`; file opened with
   `O_CREATE|O_EXCL`. Mode `0o755` for `100755`, `0o644` otherwise. Trailing LF after the bytes
   verified as framing.

Note on "reuse whatever the tar path did for duplicate platform paths": the tar path did nothing
(it opened with `O_TRUNC`, so a second colliding entry silently overwrote the first). The new path
detects the collision empirically at write time (core §2 "detect two protocol paths that map to one
platform path and fail before writing"), which is exact on the acquiring filesystem: case-distinct
paths coexist on a case-sensitive filesystem and are refused on a case-insensitive one.

## Git probes (git version 2.50.1 (Apple Git-155))

- `ls-tree -r -z` output `od -c`: `1 0 0 6 4 4 ␠ b l o b ␠ <40 hex> \t . g i t a t t r i b u t e s \0 ...`;
  a symlink lists as `120000 blob`, executable as `100755 blob`.
- `printf '<oid>\n' | git -c core.autocrlf=true cat-file --batch | od -c` for a CRLF blob:
  `<oid> ␠ b l o b ␠ 6 \n a \r \n b \r \n \n` — raw bytes, no conversion, trailing LF.
- `git -c core.autocrlf=true archive --format=tar HEAD | tar -xO subst.txt` printed the commit hash
  (export-subst expanded); the negative test additionally proves `lf.txt` becomes `alpha\r\nbeta`
  under `git archive` with `* text=auto` + `core.autocrlf=true` in the scratch repository.

## Testdata proof (`git ls-files --eol internal/gitops/testdata/byte-exact`, at HEAD)

```
i/crlf  w/crlf  attr/-text   internal/gitops/testdata/byte-exact/crlf.txt
i/lf    w/lf    attr/-text   internal/gitops/testdata/byte-exact/gitattributes.fixture
i/lf    w/lf    attr/-text   internal/gitops/testdata/byte-exact/lf.txt
i/mixed w/mixed attr/-text   internal/gitops/testdata/byte-exact/mixed.txt
i/lf    w/lf    attr/-text   internal/gitops/testdata/byte-exact/subst.txt
```

Root `.gitattributes` already carries `**/testdata/** -text`, so no new rule was needed. Deviation:
the fixture's own `.gitattributes` (`* text=auto`) is stored as `gitattributes.fixture` — when it was
copied under its real name it applied to its siblings inside curator's index and normalized
`crlf.txt`/`mixed.txt` to `i/lf` despite the root rule (nested attributes win). The tests map it
back to `.gitattributes` when committing the scratch tree. Index blob sha256 of every fixture equals
the vector's `files[].sha256` (verified with `git cat-file blob :<path> | shasum -a 256`).

## Tests

`internal/gitops`: `TestExtractReproducesByteExactVector` (scratch repo via `hash-object -w --no-filters`
+ `update-index --cacheinfo` + `write-tree`/`commit-tree`; `cat-file -p` equality proven first; extract
under `core.autocrlf=true` and `false`; exactly five files; per-file bytes equal; `$Format:%H$` and
`$Format:%h$` intact; CRLF/mixed endings intact; `hashing.ContentSHA256 == sha256:500ea934…2bced0`),
`TestExtractIgnoresWorkingTreeConversion` (negative: asserts `git archive` in that repo does convert and
does expand, then asserts Extract does not), `TestExtractPreservesExecutableBit`,
`TestExtractRefusesSubmodules`, `TestExtractRefusesEscapingPaths` (raw tree via `hash-object -t tree
--literally`: `..`, `.`, `a/../../escape2`, `sub/./x`, empty name, `/abs`; nothing written outside dest),
`TestExtractRefusesOversizeBlob` (bound narrowed to 0 refuses, 1 admits a 1-byte blob),
`TestExtractRefusesDuplicatePlatformPaths` (`Readme.md`+`readme.md`; refused on case-insensitive FS,
coexist on case-sensitive), `TestExtractRefusesExistingDestinationEntries`, `TestExtractRefusesSuspiciousCommit`,
plus the renamed `TestExtractProducesExactTree` / `TestExtractRejectsLinks`.
Production call sites of the gate: `internal/snapshot/snapshot.go:48`, `internal/closure/closure.go:413`.

`internal/interop`: `TestConformanceSnapshotAcquisition` reads `vectors/snapshot-acquisition.json` and
`expected/byte-exact-snapshot_sha256.txt` from `CURATOR_CONFORMANCE_ROOT`, verifies each fixture's on-disk
sha256/bytes against the vector, commits with exact bytes, extracts under both autocrlf settings.
- root unset → `SKIP: CURATOR_CONFORMANCE_ROOT is not set`
- pinned CI root (`SPEC_PIN 0ed5c691…`, `git archive` of that commit) →
  `SKIP: conformance root … has no vectors/snapshot-acquisition.json (pre-environments suite)`
- candidate root (curator-spec `ec695ba` `conformance/v1`) → PASS (`byte-exact-snapshot/autocrlf=true`, `autocrlf=false`), exit 0.

## Mutation evidence (each mutant reverted afterwards; clean suite green)

| Mutant | Failing tests |
|---|---|
| `Extract` body replaced by the old git-archive tar path | ReproducesByteExactVector, IgnoresWorkingTreeConversion (`lf.txt acquired CRLF … "alpha\r\nbeta\r\ngamma\r\n"`), RefusesEscapingPaths, RefusesOversizeBlob, RefusesDuplicatePlatformPaths, RefusesExistingDestinationEntries |
| size gate `> max` → `> max+1` (narrowed by one) | RefusesOversizeBlob |
| symlink branch disabled | RejectsLinks (error class changed) |
| collision `Lstat` disabled + `O_EXCL`→`O_TRUNC` | RefusesDuplicatePlatformPaths, RefusesExistingDestinationEntries |
| `100755` branch disabled | PreservesExecutableBit |

## Gates (each command run directly; real exit codes)

| Command | Exit |
|---|---:|
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `gofmt -l $(git ls-files '*.go')` (0 files listed) | 0 |
| `git diff --check` | 0 |
| `go test -count=1 -timeout 30m ./cmd/curator` (463.2 s) | 0 |
| `go test -count=1 -timeout 30m <all 57 other packages from go list ./...>` | 0 |
| `go test -count=1 -race -timeout 30m ./internal/gitops ./internal/snapshot ./internal/closure ./internal/interop` | 0 |
| `CURATOR_CONFORMANCE_ROOT=<curator-spec ec695ba>/conformance/v1 go test -count=1 -run TestConformanceSnapshotAcquisition ./internal/interop` | 0 |

`./...` was split into `cmd/curator` and the remaining 57 packages because one shell call is bounded
at ten minutes; the union is the full `go list ./...` set. Adapter suites (rustsource, swiftpmsource,
npmsource, pnpmsource, yarnclassicsource, yarnmodernsource) passed locally on darwin/arm64; hosted x86
status not claimed. Logs: `.temp/TASK-260905-3r30t1/{gates.txt,test-cmd-curator.log,test-rest.log,test-race.log}`
in the worktree. `golangci-lint` is not installed here; lint evidence is `go vet` + `gofmt` (not verified beyond that).

## Not verified

- Behavior on Windows / Git-for-Windows (the motivating platform) — only darwin/arm64 (APFS, case-insensitive) ran.
- Hosted CI on the pinned rc.9 suite: reasoned from the skip path exercised locally with a `git archive` of `SPEC_PIN`.
