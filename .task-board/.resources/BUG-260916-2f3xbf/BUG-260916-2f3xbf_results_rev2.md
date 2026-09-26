# BUG-260916-2f3xbf results — Revision 2 (rework-1)

## Outcome

Revision-1's registry change is kept exactly (junction admission in
`reconcileWritableStoreOverlay` + undeferred ledger). Revision 2 adds the
missing post-registry Windows fix: `copyContainedNode` now resolves each
admitted link to its target before the recursive `filepath.EvalSymlinks`,
porting the npm sibling's proven pattern (`npmsource` passes on
windows-latest with it). All local evidence is green; Windows is verified by
the hosted gate.

## Diagnosis (evidence-driven, supersedes the rework's exec-seam theory)

Gate run 35481906193 (rev 1) failed both real-pnpm install cases on
windows-latest with a BARE `The system cannot find the path specified.` from
`Materialize` (no `code: detail`, no `fields=`), at the `t.Fatal` lines.

Evidence pulled from the run artifacts (`gh run download 35481906193`,
`test-evidence-windows-latest/test/go-test-served.json`):

- The `pnpm install` launch is well-formed in both runs (absolute node.exe,
  existing CWD, slash-relative `pnpm.cjs` argv, no PATH) — identical shape in
  pin-rev2 run 35098955988 and bug-rev1 run 35481906193.
- Pin-rev2 run 35098955988 PROVES the exec seam works on Windows: with the
  same harness, same fixtures, same pnpm 10.33.0, and the same production
  exec path, `session.run(install)` completed and execution reached
  `reconcileWritableStoreOverlay`, failing there with the TYPED
  `closure_input_undeclared ... entry:c8805e97...` / `entry:4990f83e...`.
  A broken exec seam (`.cmd` resolution, PATH lookup, missing CWD, POSIX
  store path) cannot reach the registry step.
- Therefore the rev-1 bare errno is POST-`session.run`, AFTER the (now
  admitted) registry step. Exhaustive audit of `Materialize`'s post-install
  path: every `fail()`/`failure()`/`fmt.Errorf`/os-wrapped site carries a
  prefix or code. The only bare-`Errno` producers reachable are
  `filepath.EvalSymlinks` (whose Windows `toNorm`/`normBase` returns raw
  `FindFirstFile` errno) in `copyContainedTree`/`copyContainedNode`, and
  `windows.MoveFile` in `renameTreeNoReplace`.
- The npm sibling (`internal/npmsource`, green on windows-latest) documents
  the mechanism in its `normalizeTreeLink` comment: EvalSymlinks does not
  evaluate junctions and refuses paths carrying one; its dereferencing
  copier normalizes each child with `os.Readlink` BEFORE recursing.
  pnpm's `copyContainedNode` lacked that normalization, so on Windows it
  descended through install-tree junctions textually and EvalSymlinks
  failed on the junction-carrying path — the bare errno. The sibling
  differential (npm flow also exercises intake `CaptureTree`/`MoveFile` on
  Windows green) rules out the rename site.

Rework-1 §1 harness sub-claims, verified (no change needed):

- pnpm resolution: CI installs real npm-global pnpm 10.33.0 into a
  lane-local prefix (no corepack); `exec.LookPath("pnpm")` finds
  `prefix/pnpm.cmd` and `resolveWindowsPNPMEntrypoint` maps it to the
  adjacent `node_modules/pnpm/bin/pnpm.cjs`, which is probed and staged
  via `node pnpm.cjs` — no PATHEXT/`cmd /c` involved. Matches CI layout.
- Native paths: harness builds all filesystem paths with `filepath.Join`
  (grep for hand-built separators is clean); slash forms exist only in
  argv/environment values Node accepts (evidenced working).
- Parent dirs: harness pre-creates `bin`/`work`/`output` via
  `privatedir.MakeAll`; the runner stages each CWD via work copies and
  Lstat-gates it before `Start` (typed refusal otherwise).

## Change (surgical, POSIX-identical)

`internal/pnpmsource/materialize.go`, `copyContainedNode` loop only: resolve
each child with the rev-1 `admittedLink`/`normalizeTreeLink` primitives
before recursing — the exact npm-sibling loop shape. On Unix
`normalizeTreeLink` is the identity, so paths, errors, and bytes are
unchanged (one cached `DirEntry.Info()` call added). On Windows, junctions
resolve to targets, restoring POSIX parity for the containment (`active` /
`Rel`) checks — which previously saw only the unresolved spelling — and
eliminating through-junction `EvalSymlinks`.

Side effect fixed as a consequence: an outside-tree junction previously
copied from outside undetected on Windows (containment-blind); it is now
refused with `closure_local_path_escape`, and junction cycles fail closed
with `closure_input_undeclared`, exactly as on POSIX.

No ledger, self-test, harness, or rev-1 registry changes in this revision.

## Tests

New:

- `internal/pnpmsource/copy_unix_test.go` (`//go:build unix`):
  `TestCopyContainedTreeDereferencesLinks` — through-link content copy with
  no link left behind; escape refused (`closure_local_path_escape`); cycle
  refused (`closure_input_undeclared`). Pins the POSIX contract.
- `internal/pnpmsource/linkentry_windows_test.go` (windows-only, `mklink /J`
  fixtures): `TestCopyContainedTreeDereferencesWindowsJunction` (content
  copied, destination link-free — reproduces the gate failure shape without
  the fix), `...RefusesWindowsJunctionEscape`,
  `...RefusesWindowsJunctionCycle`.

Retained: rev-1 negatives (`TestWritableStoreOverlay...` incl. the plain
directory member refusal and `assertUndeclaredMemberRefusal`) all green.

## Evidence (exit codes, `bash`, `set -o pipefail` semantics via PIPESTATUS)

Shell: `bash`. Worktree:
`.temp/STORY-260915-3w11un/worktree`.

| # | Command | Exit |
|---|---------|------|
| 1 | `gofmt -l internal/pnpmsource/` (empty) | 0 |
| 2 | `go vet ./internal/pnpmsource/` | 0 |
| 3 | `GOOS=windows go vet ./internal/pnpmsource/` | 0 |
| 4 | `GOOS=windows go test -c -o /tmp/pnpmsource-windows.test ./internal/pnpmsource/` | 0 |
| 5 | `go test -p 1 ./internal/pnpmsource -run 'TestCopyContainedTreeDereferencesLinks\|TestWritableStoreOverlay\|TestResolveWindowsPNPMEntrypoint' -count=1` | 0 (`ok`, 0.428s) |
| 6 | `go test -p 1 ./internal/pnpmsource -run 'TestRealPinnedPNPM' -count=1` with pinned pnpm 10.33.0 in `/tmp/pnpm-prefix` on PATH (darwin) | 0 (3/3 PASS, 37.1s) |
| 7 | `go test -p 1 ./internal/pnpmsource -count=1` (full package, same PATH) | 0 (`ok`, 53.6s) |
| 8 | POSIX no-op mutant: `materialize.go` fix stashed, new unix tests re-run | 0 (`ok` — identical) |

Windows is verified only by the hosted gate (no Windows runner here): the
new junction tests compile into the Windows test binary (#4) and run there;
the two real-pnpm cases stay undeferred in the ledger (rev 1, untouched).

## Checklist

- [x] Rev-1 registry change kept exactly; ledger deferral stays removed
- [x] Rework-1 §1 harness claims verified against gate evidence (no-op with cause)
- [x] Junction normalization ported into `copyContainedNode` (sibling pattern)
- [x] Portable parity tests + Windows junction tests added
- [x] `gofmt` clean, `go vet` clean (darwin + `GOOS=windows`)
- [x] Narrow + full-package tests green, incl. real pinned pnpm on darwin
- [x] POSIX no-op proven via stash mutant
- [x] Negative undeclared-member tests retained and green
- [ ] Hosted windows-latest gate green (adjudicated at handoff/integration)
