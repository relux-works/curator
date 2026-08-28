# TASK-260825-1yzubs refreshed-base developer outcome

## Result

The reviewed injectable-seams candidate was rebuilt on current `main` without
staging or committing. `HEAD`, `main`, and `origin/main` all resolve to
`de31754e854e385fca04de9cafeae06667a96123`.

The pre-refresh working delta was captured as a full-index binary patch before
the branch pointer moved. A detached `main` probe applied 10 paths cleanly and
reported one `LOGBOOK.md` conflict. The resolution retained both the current
documentation-campaign history and this task's existing implementation record.
The resulting main-based patch was then applied to the managed Story worktree.

The production and test behavior remains the reviewed design:

- `main` resolves the environment-backed config path once at the process
  boundary;
- `run(args, configSource, stdout, stderr)` receives invocation-local seams;
- command methods carry those dependencies through a `cli` value;
- the concurrent seam test uses distinct config/project/writer sets and rejects
  cross-output;
- formerly serial CLI cases use private configs and writers with `t.Parallel()`;
- the compiled status repair/rollback/recovery lifecycle is split across
  independent parallel fixtures;
- the process-level host-GOROOT lock preserves cross-package isolation while
  avoiding in-package self-deadlock.

No assertion was weakened.

## Exact candidate scope

`git diff --name-only main` reports exactly these 11 task-owned paths:

1. `LOGBOOK.md`
2. `cmd/curator/builds.go`
3. `cmd/curator/builds_test.go`
4. `cmd/curator/gc_test.go`
5. `cmd/curator/global_status_test.go`
6. `cmd/curator/lifecycle_conformance_test.go`
7. `cmd/curator/main.go`
8. `cmd/curator/main_test.go`
9. `cmd/curator/status_test.go`
10. `cmd/curator/toolchain_remedy_test.go`
11. `internal/testtoolchain/lock.go`

There are no non-ignored untracked files. The six documentation files named in
the refresh brief are present and absent from the candidate diff:
`docs/cli.md`, `docs/troubleshooting.md`, `docs/prose-style.md`,
`docs/ci-gates.md`, `docs/compiled-commands.md`, and
`docs/authoring-cli-commands.md`.

The final binary patch SHA-256 is
`244053d5ed91228619c3b7306e29df8753ad97ec9466ab9e3f66dc4835d8318e`.

## Timing

The three strict measurements were consecutive standalone
`/usr/bin/time -p go test ./cmd/curator -count=1` runs.

| Metric | Before | Refreshed run 1 | Refreshed run 2 | Refreshed run 3 |
| --- | ---: | ---: | ---: | ---: |
| Exit code | 0 | 0 | 0 | 0 |
| Wall clock | 408.01s | 215.77s | 215.48s | 215.09s |
| Delta vs baseline | — | -192.24s | -192.53s | -192.92s |
| Four-minute margin | -168.01s | +24.23s | +24.52s | +24.91s |

The worst refreshed run is 47.1% faster than the pre-refactor baseline and is
24.23s under the four-minute acceptance bound. An additional full run exited 0
with a package-reported 220.015s; its attempted shell wall timer returned zero,
so that invalid wall value is not used in the three-run claim.

## Coverage

| Metric | Pre-change | Refreshed | Delta |
| --- | ---: | ---: | ---: |
| Statement coverage | 62.2% | 63.3% | +1.1pp |
| Covered statements | 929 | 951 | +22 |
| Total statements | 1492 | 1502 | +10 |

The refreshed coverage command exited 0 in 214.13s wall clock. Coverage did not
regress.

## Validation

Every counted gate ran as a standalone process and its real exit code is in the
attached log.

| Command | Exit | Result |
| --- | ---: | --- |
| Focused injected config/writer test | 0 | `ok`, package 0.461s |
| Focused injected config/writer `-race` | 0 | `ok`, wall 5.69s |
| Full `cmd/curator` uncached runs x3 | 0 / 0 / 0 | 215.77s / 215.48s / 215.09s |
| Full `cmd/curator` coverage | 0 | 63.3%, 951/1502 statements |
| `go build -o .temp/... ./cmd/curator` | 0 | binary produced in task temp |
| `go vet ./cmd/curator ./internal/testtoolchain` | 0 | clean |
| golangci-lint 2.12.2 | 0 | `0 issues` |
| gofmt cleanliness | 0 | no files reported |
| `git diff --check` | 0 | clean after final logbook update |
| Scope and documentation-presence gate | 0 | exactly 11 paths, zero doc deletion |
| `task-board validate` | 0 | current board reports 598 issue(s) found |

The task brief required wording that `task-board validate` exits 0 while
reporting 1741 issue(s). That statement remains true for the prior attached
checkpoint evidence. The refreshed command run against the current authoritative
board exited 0 while reporting 598 issue(s), so this outcome reports both
snapshots rather than mislabeling the fresh evidence. Neither result is
semantically clean despite the CLI exit code.

## Base-state anomaly

The brief also stated that `internal/testtoolchain/swift.go` and
`internal/testtoolchain/swift_test.go` exist on current main. Git tree inspection
shows that neither PR #47 merge `2bb54a25` nor current main `de31754e` tracks
them. After the mixed base move, leftover copies matched checkpoint `defbc368`
byte-for-byte and appeared as untracked files. They were moved to task-scoped
temporary recovery storage and excluded from the candidate, as required by the
explicit 11-path scope. The tracked Windows and host-toolchain changes present
on main remain untouched.

## Non-counted diagnostics

- The first capture command exited 1 because its log redirect targeted a
  missing control-root temp directory. It changed no repository file; the
  capture was rerun successfully from the execution root.
- The detached three-way apply exited 1 on the expected `LOGBOOK.md` conflict.
  After retaining both sides, the resolved probe passed diff and scope checks.
- A first scope script was invalid because the zsh loop variable `path`
  overwrote the shell `PATH`; its apparent shell exit 0 is not counted. The
  corrected `doc_path` gate exited 0.
- A diagnostic after `go build ./cmd/curator` exited 1 because the command had
  produced an untracked root binary. The binary was moved into task temp and the
  build was rerun with explicit `-o`; the final build and scope gates exited 0.

No files are staged or committed.
