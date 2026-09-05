# Producer brief: byte-exact acquisition — rework 2 (cat-file deadlock on mid-stream refusal)

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-acquisition-byte-exact`, branch `feat/byte-exact-acquisition`,
head `a46abc80`, PR https://github.com/relux-works/curator/pull/58. Findings: `TASK-260905-3r30t1_review-verdict.md`
(reviewer RUN-260905-f29486). Work in the curator worktree; new signed commits on top of `a46abc80`, no rewrite.

## Author decisions
- **F1 (blocking)** — restructure `internal/gitops` `writeBlobs`: (a) run every refusal that needs no blob bytes as a
  pre-pass over the `ls-tree` entries before `cat-file` starts — `safeTarget` (escapes, `.`/`..`), duplicate platform
  paths, pre-existing destination entries, and the declared size from the `ls-tree -l` long form (add `-l` so oversize
  blobs are refused from the listing without streaming them); (b) for the residual mid-stream errors (framing, I/O,
  a size that disagrees with the header), terminate the child deterministically before `Wait`: close stdin, `Kill` the
  process, drain stdout to `io.Discard` with a bounded copy, then `Wait`, and return the typed error; no path may
  reach `Wait` with undrained stdout. (c) Named tests in `byteexact_test.go`, each under a 20 s watchdog, asserting a
  bounded return and an empty destination: a single ≥ 1 MiB blob with `maxSnapshotFileBytes` narrowed below it; a
  duplicate-path collision followed by an 8 MiB blob; an escape followed by an 8 MiB blob; and a mid-stream framing
  error injected through a fake `cat-file` on `PATH` (or a corrupt-object scratch repository) with a large queued blob.
  Register in `.github/ci/platform-cases.tsv` only what the ledger requires; the collision case stays darwin/windows
  as its host capability demands (say how you gated it).
- **O1** — reject a `.git` path component in `safeTarget` alongside `.`/`..` (git's own `verify_path` refuses it); add a
  refusal test.
- **O2** — closure scratch reuse of a partially written `target` (`internal/closure/closure.go` `snapshotFor`): extract
  into a sibling temp directory and rename into place only after `Extract` returns nil, so a refused extraction leaves
  no reusable partial tree; test it.
- **O3** — add the nested `100755` test the mutation evidence lacked; **O4** needs no change (record as verified).
- Gates: `go build ./...`, `go vet ./...`, `gofmt -l`, `go test -count=1 -race -timeout 30m ./internal/gitops
  ./internal/snapshot ./internal/closure ./internal/interop`, the platform-case gate locally per `ci.yml` for the three
  GOOS values, `bash .github/ci/gate-selftest.sh`; rerun `go test -count=1 -timeout 30m ./cmd/curator` once at the end.
  Push (plain, no force) and watch `gh pr checks 58 --watch` to green. Attach
  `TASK-260905-3r30t1_rework-report-2.md` (finding → disposition, test names, watchdog timings, gate outputs, check
  summary); `task-board handoff TASK-260905-3r30t1 --role developer`. Never write LOGBOOK.md or anything into the
  control root; the story workspace carries an empty delta by design — do not touch it.
