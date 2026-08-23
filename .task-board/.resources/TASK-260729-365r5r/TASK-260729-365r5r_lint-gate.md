# `gate-lint-abs` — the one command this re-entry cycle was authorized to run

> **SUPERSEDED by `evidence/rework-lint.md` (RUN-260729-d36102).** The three
> `revive` findings below are fixed and `golangci-lint run ./internal/transaction/...`
> now exits 0. This document is kept as the historical record of how the blocker
> was found. Note that its three-finding list was **truncated** by
> golangci-lint's default `max-same-issues: 3` — a fourth instance existed at
> line 144 and only surfaced once the first three were fixed.

RUN-260729-3819ca operates under `TASK-260729-365r5r_reentry-constraint.md`: no
Go command, no baseline script, no product edit, no detached process. It MAY run
exactly one command — `/Users/iv/go/bin/golangci-lint run` in the prototype
worktree — after proving the shared process barrier is empty.

## Barrier proof, taken first

```
$ .temp/TASK-260729-365r5r/bin/barrier.sh > gates/gate-lint-abs.barrier
$ echo $?
0
```

`gates/gate-lint-abs.barrier`:

```
SCAN 1 empty
SCAN 2 empty
BARRIER_OK
```

Two scans, 2 s apart, no `go`, no `*.test`, no `go-build`, no `cmd/curator`
process on the host. Independently, `ps -ax` at 2026-07-29T20:04:17 showed no
`run-gates.sh`, no `wait-then-run-gates.sh`, and no
`run-gates-baseline-race.sh` alive.

## The command and its real exit code

```
cd .temp/TASK-260729-365r5r/worktree
env GOTMPDIR=.temp/TASK-260729-365r5r/gotmp/transaction \
    /Users/iv/go/bin/golangci-lint run
```

Standalone process. No `tee`, no pipe chain. `gate-lint-abs.exit` written last.

| | |
| --- | --- |
| exit | **1** |
| wall | 3 s |
| linter | golangci-lint 2.4.0, go1.25.5 |
| config | `.golangci.yml`, unmodified (not in the delta manifest) |

**This gate is RED.** It is reported as failing.

## The 4 issues

```
internal/godriver/builddriver_positive_conformance_test.go:178:4:
    ineffectual assignment to environment (ineffassign)
internal/transaction/namespace_pass_test.go:119:16:
    unused-parameter: parameter 't' seems to be unused (revive)
internal/transaction/namespace_pass_test.go:127:16:
    unused-parameter: parameter 't' seems to be unused (revive)
internal/transaction/namespace_pass_test.go:136:16:
    unused-parameter: parameter 't' seems to be unused (revive)
```

### 1 inherited, 3 introduced

The `ineffassign` finding is **pre-existing on the baseline**, proven statically
without running anything:

```
$ diff -q worktree/internal/godriver/builddriver_positive_conformance_test.go \
          worktree-baseline/internal/godriver/builddriver_positive_conformance_test.go
$ echo $?
0
```

The file is byte-identical to the never-edited twin and is not in
`manifest-pre-post.diff`. The rfrdfo baseline tree lints red on this issue too.
It is not this prototype's regression, and this prototype does not fix it.

The three `revive` findings **are** this prototype's. They are in
`internal/transaction/namespace_pass_test.go`, the one file this task added
(`worktree-baseline` has no such file). Three of the seven `build` closures in
`TestValidateIndependentTargetNamespacesRejectsOverlappingPaths` take a
`t *testing.T` they never use, because they share a table signature with the
four closures that do. The fix is renaming those three parameters to `_`.

## Why it was not fixed in this run

The re-entry constraint forbids product/test edits and permits exactly one
command. Editing `namespace_pass_test.go` would silently invalidate every gate
in `gates/` — `gate-transaction`, `gate-race-transaction`,
`gate-namespace-verbose` and `gate-equivalence` all compile that file — and
this run may not re-run them. A stale-green gate set is worse evidence than an
honestly red lint gate.

So: **recorded, not fixed.** This is a review-blocking finding for whoever runs
the integration cycle, and the "Lint clean" checklist item is left unchecked.

## Relationship to the driver's own `gate-lint`

`gates/gate-lint.exit` = **127**. The gate driver invoked bare `golangci-lint`,
which is not on the driver's `PATH`; the binary exists at
`/Users/iv/go/bin/golangci-lint`. That 127 is a missing-binary code, never a
pass. It is kept as the historical record and superseded by `gate-lint-abs`,
which is the same `make lint` command with an absolute path.
