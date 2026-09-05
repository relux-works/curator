# Producer brief: byte-exact snapshot acquisition in curator (stage (a), item 1)

## Where and what

- Repository `~/Developer/ReluxWorks/curator` (Go 1.25.5). Worktree
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-acquisition-byte-exact`, branch
  `feat/byte-exact-acquisition`, base = curator main `74c35b1c`. First run
  `git submodule update --init --recursive` in the worktree (the build needs
  `agents/skills/skill-go-testing-tools`).
- Authority: curator-spec main `ec695ba`: `protocol/environments.md` §1.2 (snapshot
  byte-exactness), `conformance/v1/vectors/snapshot-acquisition.json`,
  `conformance/v1/fixtures/byte-exact/*`, `conformance/v1/expected/byte-exact-snapshot_sha256.txt`;
  `protocol/core.md` §6.2, §6.5 (raw-object discipline for external repositories), §8 (content
  hash). Review M3 in `pre-implementation-review-v3.md` (board resource on STORY-260901-zddtn8)
  states the defect.
- Deliverable: signed commits (`git commit -S`; paste `git log --show-signature` lines);
  gates green and recorded: `go build ./...`, `go vet ./...`, `gofmt -l` clean,
  `go test -count=1 -timeout 30m ./...` (the `cmd/curator` package alone takes ~10 minutes;
  never use a shorter timeout — it panics inside godriver and reads as a hang), and the
  focused packages with `-race`. Do not push, tag, or open a PR. Attach
  `TASK-260905-3r30t1_drafting-report.md` as an outcome resource, then
  `task-board handoff TASK-260905-3r30t1 --role developer`. Never write LOGBOOK.md or anything into the
  control root.

## The defect and the change

`internal/gitops/gitops.go` `Archive(repo, commit, destination)` runs `git archive --format=tar
<commit>` and extracts the tar. `git archive` honors `core.autocrlf`/`text`/`eol` attributes
and expands `export-subst`, so the extracted bytes depend on the acquiring machine's git
configuration and the repository's `.gitattributes`; on Git-for-Windows defaults every
git-sourced snapshot hashes differently. Callers: `internal/snapshot/snapshot.go:48` (the
skills snapshot cache) and `internal/closure/closure.go:413` (closure scratch snapshots).

Replace the extraction with object-database extraction that reproduces exact committed blob
bytes:

1. In `internal/gitops`, add `Extract(repo, commit, destination string) error` (or rename
   `Archive` in place — keep one exported entry point; if you keep the name `Archive`, fix its
   doc comment) that: resolves `<commit>^{tree}`; lists entries with
   `git -C <repo> ls-tree -r -z --full-tree <tree>` (mode, type, oid, path); refuses anything
   that is not a regular blob (`100644`/`100755`) — symlinks (`120000`), gitlinks (`160000`,
   submodules), and any other type are refused with the existing "links … unsupported" /
   "unsupported entry type" error classes; rejects path escapes, empty names, and duplicate
   platform paths (case-insensitive collision on case-insensitive filesystems — reuse whatever
   the current tar path does for "duplicate platform paths", core §6.2); writes each blob's
   bytes with `git cat-file --batch` (one process, NUL/LF-framed as `cat-file --batch` emits:
   `<oid> blob <size>\n<bytes>\n`), bounded by the existing `maxSnapshotFileBytes`; preserves
   the executable bit from mode `100755` (mode `0o755` vs `0o644`, as the tar path did); never
   consults the working tree, attributes, or filters (`git cat-file` reads raw objects; do not
   pass `--filters` or `--textconv`).
2. Keep every safety property of the tar path (escape check, size bound, no links) with tests;
   keep the `git` invocation discipline of the package (fixed binary and flags, `-C repo`,
   `--` before untrusted operands where supported, `EnsureRepo` first).
3. Both callers switch to the new path; nothing else in their behavior changes (cache-root
   mode, authentication of cache hits, scratch layout).
4. Conformance: add a test that reproduces the spec vector without needing the conformance
   root — copy the five `fixtures/byte-exact` files from curator-spec `ec695ba` into a testdata
   directory as byte-exact testdata (commit them through plumbing or with a `-text` attribute
   in curator's own `.gitattributes` for that testdata path — verify with
   `git ls-files --eol` that `crlf.txt` shows `i/crlf` and `mixed.txt` `i/mixed`; if curator's
   root `.gitattributes` would normalize them, add the narrowest rule that protects them and
   prove it), build a scratch repository in the test, commit the fixture tree with exact bytes
   (`git hash-object -w --no-filters` + `git update-index --add --cacheinfo`, then
   `write-tree`/`commit-tree`), extract it under `core.autocrlf=true` and `false`, and assert
   `hashing.ContentSHA256` of the snapshot equals
   `sha256:500ea934403d10a2a0b6b7e8874790e489ee002328d3dc0edbda2fe5be2bced0` and that
   `subst.txt` still contains the literal `$Format:%H$`. Add a second, conformance-root-driven
   test in `internal/interop/golden_test.go` style that reads
   `vectors/snapshot-acquisition.json` and `expected/byte-exact-snapshot_sha256.txt` from
   `CURATOR_CONFORMANCE_ROOT` and skips (with the reason) when the root has no
   `vectors/snapshot-acquisition.json` — the pinned rc.9 root does not; the candidate lane
   does.
5. A negative test proving the old behavior is gone: under `core.autocrlf=true` with a
   `* text=auto` attribute committed, the snapshot bytes are unchanged (the tar path would have
   produced CRLF) — the test must fail if extraction is switched back to `git archive`.
6. Docs: `README.md` tools/section that describes snapshot acquisition if any; the
   `internal/gitops` package doc; CHANGELOG if the repo keeps one (check).

## Constraints

- Preserve the working tree of the main checkout; work only in the worktree.
- Mind `curator-adapters-never-ci-green`: the adapter suites (Cargo/Swift/pnpm/yarn) are red
  on hosted x86 runners regardless — do not chase them; record which packages you ran locally.
- Verify facts on the installed `git` (`git --version`) — e.g. that `cat-file --batch` never
  applies filters and that `ls-tree -z` framing is what you parse; paste the probes.

## Report

`TASK-260905-3r30t1_drafting-report.md`: commits + signatures; the extraction design (commands, framing,
refusals); gate outputs with exact commands and exit codes; the `git ls-files --eol` proof for
the testdata; the conformance test's skip/run behavior; anything not verified labeled.
