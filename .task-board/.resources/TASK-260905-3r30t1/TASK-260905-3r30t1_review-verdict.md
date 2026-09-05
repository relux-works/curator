# TASK-260905-3r30t1 review verdict — CR-TASK-260905-3r30t1-3 rev 3 (curator head a46abc80)

**Verdict: CHANGES REQUESTED → `to-dev`.** `repeat-of: none` (no prior findings resource exists on this task).

Subject: curator worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-acquisition-byte-exact`,
`feat/byte-exact-acquisition` @ `a46abc80` (4 commits over main `74c35b1c`), PR #58 (head verified equal, all
hosted checks green). Change Request under review has `repository_delta=empty` in the story workspace
(`git diff f61ee9a..741e0387` empty, confirmed) — that is the correct shape for this leaf, since the deliverable
lives in the curator repository and this story workspace only carries spec; the emptiness is not the reason for
the verdict. The verdict is about the curator change itself.

## F1 — BLOCKING: `writeBlobs` deadlocks on every mid-stream refusal once git has more than a pipe buffer queued

File: `internal/gitops/gitops.go`, `writeBlobs` (the `defer { stdin.Close(); cmd.Wait() }` at the top of the
function, and every `return err` inside the loop after `cmd.Start()`).

What is wrong: `cat-file --batch` is started with `StdoutPipe`; every oid is queued on stdin by a goroutine. When
the loop returns early (size refusal, duplicate platform path, existing destination entry, framing error,
`safeTarget` refusal), nobody drains stdout. `cmd.Wait()` waits for process exit *before* closing the pipe, and
git blocks forever writing the undrained bytes. The gate does not "refuse" — the process hangs. Both production
callers (`internal/snapshot/snapshot.go:48`, `internal/closure/closure.go:413`) hang with it.

Evidence (my tests, scratch copy of the worktree at a46abc80, `.temp/review/curator-copy/internal/gitops/zz_review_adv*_test.go`,
each guarded by a 20 s watchdog):

```
=== RUN   TestAdvSingleOversizeBlobHangs      # one 1 MiB blob, maxSnapshotFileBytes=512 KiB — the exact production shape of the size gate
    single-oversize: Extract HUNG for 20s (cat-file deadlock)
=== RUN   TestAdvOversizeThenBigBlobDoesNotHang   # refusal at entry 1, 1 MiB blob queued at entry 2
    oversize: Extract HUNG for 20s (cat-file deadlock)
=== RUN   TestAdvDuplicateThen8MiB            # A.txt/a.txt collision on APFS, 8 MiB blob queued behind it
    dup-8mib: Extract HUNG for 20s (cat-file deadlock)
```

Production consequence: a source repository with a single file over 512 MiB makes `curator` hang instead of
reporting `file too large in git snapshot`; a case-colliding repository with any large file behind the collision
hangs on macOS/Windows instead of reporting `duplicate platform path`. The committed `TestExtractRefusesOversizeBlob`
uses a 1-byte blob, so it fits in the pipe buffer and never sees the deadlock — the suite is green around a
gate that is unreachable as an error in production. The escape refusal (`TestAdvEscapeThen8MiB`) happened to
return in time because the refusal precedes any read; that is timing-dependent, not a design property.

Fix: on any error after `cmd.Start()`, terminate or drain the child before waiting — e.g. `defer` that closes
`stdin`, closes/`io.Copy(io.Discard, stdout)` or `cmd.Process.Kill()`, then `Wait`; or run the refusals that do
not need the header (`safeTarget`, `written`, `Lstat`) in a pre-pass over `entries` before starting cat-file, and
kill on the remaining size/framing errors. Then add a NAMED negative test in `byteexact_test.go` that commits a
blob larger than any pipe buffer (≥ 1 MiB), narrows `maxSnapshotFileBytes` below it, and asserts `Extract` returns
`too large` within a bounded time (watchdog) and leaves nothing behind; and a second one for the duplicate-path
refusal with a large trailing blob. Register the new case(s) in `.github/ci/platform-cases.tsv` only if the
ledger requires it (rework report 1 says unlisted cases are merely unlisted).

## Non-blocking observations (fix with F1 or declare as bounds)

- O1 `internal/gitops/gitops.go` `safeTarget`: a tree entry named `.git` is admitted (`TestAdvDotGitEntry`: written,
  and `EnsureRepo(dest)` then succeeds on the snapshot). Git itself refuses this name in `verify_path`; the old tar
  path had the same gap, so this is pre-existing and out of the brief's scope. State it as a bound or reject `.git`
  components alongside `.`/`..`.
- O2 Partial writes on late refusal are pre-existing for the closure scratch caller (`snapshotFor` extracts into
  `target` directly and reuses it on the next call if `os.Stat` succeeds); the snapshot cache caller extracts into
  a temp dir it removes. Not a regression from this change; worth a stated bound in the report.
- O3 Mutation-evidence survivors M2 (submodule-specific branch) and M6 (nested exec bit) are correctly labelled
  bounds; I verified nested `100755` under `sub/run.sh` is written `-rwxr-xr-x` (`TestAdvNestedExec`), so M6 is a
  test gap, not a defect.
- O4 Mode `100664` in a literal tree is canonicalised by `ls-tree` to `100644` (verified), so the mode allow-list
  cannot be bypassed that way.

## What I verified and accept as-is

| Item | Evidence |
|---|---|
| Object-database only | `gitops.go`: `ls-tree -r -z --full-tree` + `cat-file --batch`, no `--filters/--textconv`, no `archive`/checkout; `EnsureRepo` first; fixed flags |
| Spec vector reproduces | `go test -count=1 -run 'ByteExact|SnapshotAcquisition|Refuses|Extract|RejectsLinks' ./internal/gitops ./internal/interop` PASS; with `CURATOR_CONFORMANCE_ROOT=…/curator-spec/conformance/v1` (main f61ee9a) `byte-exact-snapshot/autocrlf=true|false` PASS; unset root → SKIP |
| Negative test bites | git-archive mutant (old `Archive` body behind `Extract`) fails `TestExtractReproducesByteExactVector` (`.gitattributes`, `subst.txt` differ), `TestExtractIgnoresWorkingTreeConversion` (`lf.txt acquired CRLF`), plus Submodules/Escaping/Duplicate/Existing; reverted, `diff -q` clean |
| Refusals | symlink `120000`, gitlink `160000`, `..`/`.`/`/abs`/empty, `A.txt`/`a.txt` on APFS, pre-existing dest file, oversize (1-byte shape) — all refused with the typed classes; refusals at list time write nothing |
| Testdata | `git ls-files --eol`: `crlf.txt i/crlf`, `mixed.txt i/mixed`, all `attr/-text` via root `.gitattributes` line 3 `**/testdata/** -text`; scratch commit installs `gitattributes.fixture` as `.gitattributes` |
| Gates (my run, `.temp/review-2/gates.log` in the curator worktree) | `go build ./...` 0; `go vet ./...` 0; `gofmt -l` 0 files; `go test -count=1 -race -timeout 30m ./internal/gitops ./internal/snapshot ./internal/closure ./internal/interop` 0; `bash .github/ci/gate-selftest.sh` 81 passed/0 failed |
| `./cmd/curator` | not rerun; producer's 463 s run (exit 0) cited; hosted Test/Race on ubuntu/macos/windows green at a46abc80 |
| Platform-case ledger | hosted CI green on all three GOOS at a46abc80 is the proof; I did not rerun the local `CI_GATE_GOOS` overrides |
| Commits | 4 commits, `git verify-commit` Good signature for oparin@me.com on each; author Ivan Oparin; PR head == a46abc80 |
| Scope | callers changed one line each; CHANGELOG + implementation-plan + package doc updated; README has no acquisition section |

Not verified: Windows behavior beyond hosted CI; the closure scratch partial-reuse path (O2) on the new code.
