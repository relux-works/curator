# TASK-260910-5nrmtt results — rev3 gate repair (rev4 candidate)

Producer: developer. Shell for all commands below: bash, unpiped except
where noted; every exit code below is the gate's real status.

## Why this run exists

The rev3 candidate (rework 2: closed line grammar + Windows refusal, 12
paths, results in `TASK-260910-5nrmtt_results.md`) reached the remote gate
(run 35114028599, `TASK-260910-5nrmtt_change-request_rev3-validation.log`)
and FAILED on the platform-case gate only: `go test` itself was exit 0 in
every lane, but the gate was exit 1. Root cause, quoted from the log:

- `FAIL  skip with an unrecognised reason on linux: internal/buildrepo ::
  TestResolvedTransportWindowsRefusesBeforeAnyProcess`
- `reason: windows refusal probe runs on windows only`
- `add it to .github/ci/skip-classes.tsv with a class, or fix the case.`

Two skip reasons introduced by rework 2 matched nothing in
`.github/ci/skip-classes.tsv`, so every lane that executed (or skipped)
them failed Tier 2. The failure is in test wording only; production code
(`internal/buildrepo/transport.go` and siblings) is byte-identical to the
reviewed rev3 tree. This run fixes the case (rewording), per the gate's own
guidance — `skip-classes.tsv` itself is untouched (outside task scope).

## Fix (2 lines, `internal/buildrepo/transport_test.go` only)

Both reasons now match the existing `platform-control` row
`(is|are) exercised (on|by)` ("the behaviour is asserted on its own
platform's runner"):

- `requireResolvedLane` (skips on Windows; lane behaviour runs on unix):
  "bounded resolved lane refuses on windows; behavior tests run where the
  lane runs" -> "bounded resolved lane refuses on windows; behavior is
  exercised on unix runners"
- `TestResolvedTransportWindowsRefusesBeforeAnyProcess` (skips off
  Windows; probe runs on Windows): "windows refusal probe runs on windows
  only" -> "windows refusal probe is exercised on Windows"

Tree: `git status --short` shows exactly the 12 rev3 paths (4 modified:
`internal/buildrepo/admission.go`,
`internal/buildrepo/httpsbroker_test.go`, `internal/gitcred/gitcred.go`,
`internal/gitcred/provider_test.go`; 8 new:
`docs/draft-transport-resolution.md`,
`internal/buildrepo/{process_unix,process_windows,sshbroker,sshbroker_test,transport,transport_test}.go`,
`internal/gitcred/provider.go`) with the 2-line wording fix inside the new
`transport_test.go`. No other dirty files; sibling policy checkpoint,
parser/collections, `internal/gitops`, `internal/buildsource`, frozen v1
schemas, and `.github/ci/*` untouched. Worktree
`task-board/story/STORY-260910-1bhj0g` at `38e62cb`.

## Evidence (real exit codes)

- Skip-vocabulary probe (throwaway `/tmp/5nrmtt-ev/check-skips.sh`: every
  `t.Skip` reason in the touched test files against every regex in
  `skip-classes.tsv`): exit 0 — 7/7 reasons matched, both new reasons hit
  `platform-control :: (is|are) exercised (on|by)`. The printed skip line
  was confirmed verbatim via `go test -json`:
  `transport_test.go:1873: windows refusal probe is exercised on Windows`.
- Targeted entry set (`-run` over the platform gate unit test, the
  Windows-refusal probe, both `requireResolvedLane` users, the rev2
  mixed/unknown probes, the 62-case grammar unit): exit 0 — 5 PASS,
  1 SKIP (Windows-only probe, by design), 0 FAIL.
- Full `./internal/buildrepo` suite: exit 0 (`ok`, 60.676 s).
- Full `./internal/gitcred` suite: exit 0 (`ok`, 4.278 s).
- `go vet ./internal/buildrepo ./internal/gitcred`: exit 0.
- `GOOS=windows go build` + `GOOS=windows go vet ./internal/buildrepo`:
  exit 0.
- `gofmt -l internal/buildrepo internal/gitcred`: no output. `git diff
  --check`: exit 0.
- Full module suite deliberately NOT run (host rule; gate at handoff).

## Narrowing mutants (bytes restored, `cmp` clean)

- M1: reason 1873 reverted to "runs on windows only" -> probe exit 1,
  `UNMATCHED: windows refusal probe runs on windows only`. KILLED.
- M2: reason 1848 reverted to "behavior tests run where the lane runs" ->
  probe exit 1, `UNMATCHED: bounded resolved ...`. KILLED.

## Findings / bounds

- The rework-2 briefs attached to this spawn were already fully applied in
  the worktree (closed grammar + Windows refusal option b + rev2 reviewer
  probes verbatim); this run changed only the 2 skip strings. Production
  behaviour, diagnostics, and bounds are as stated in
  `TASK-260910-5nrmtt_results.md`, which remains the rev3 record — this
  file covers only the gate repair.
- Windows refusal runtime proof still rests with the hosted gate
  (windows-only test); macOS/Windows verdicts for THIS tree come from the
  rev4 gate run at handoff. Checklist on the task stays ticked.
