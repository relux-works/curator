# TASK-260729-3jku56 implementation evidence

## Outcome

`stageNode` now receives `marker.BuildCurrentness` for unchanged compiled
nodes. A repeated project or global install with an exact cache-hit plan reports
`up-to-date`, stages no build output, and leaves the installed context directory
byte-identical.

The proof is independently derived from:

- a newly validated raw snapshot token;
- a fresh protected-cache inspection;
- cloned planned `buildmeta.Input` values;
- the complete static prompt-context file set; and
- the complete active script-runtime source file set.

Only an all-cache-hit node is eligible. A miss, untrusted cache rebuild,
corrupt/unsupported outcome, absent proof, or derivation error passes no build
proof and preserves the existing fail-closed behavior.

## Task-only delta

The isolated worktree is:

`/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3jku56/worktree`

It was composed from the accepted read-only
`TASK-260720-1ljev5` worktree. The task-owned delta is limited to:

- `internal/install/targets.go`
- `internal/install/install.go`
- `internal/install/global.go`
- `internal/install/stage_test.go`

No predecessor or concurrent worktree was edited. Nothing was staged,
committed, published, or installed on the host.

## Regression and coverage

The focused regression was run before production changes:

```text
go test ./internal/install -run '^TestRepeatedUnchangedCompiledInstallIsUpToDateWithoutRestagingContext$' -count=1
exit 1 (expected red)
```

Observed failure: the second plan reported `outcome=cache-hit`, but the compiled
skill message still reported `installed`.

Green production-path coverage now proves:

- unchanged project compiled install: `up-to-date`, no build staging, context byte-identical;
- unchanged global compiled install: same;
- changed build source: build and context re-staged;
- changed native target tuning: build and context re-staged;
- changed toolchain identity: build and context re-staged;
- changed cache provenance: build and context re-staged;
- changed prompt-context boundary: build and context re-staged;
- complete context enumeration excludes build/runtime roots;
- complete runtime enumeration includes every copied runtime-root file,
  including prompt-whitelist-excluded names; and
- unavailable raw snapshot state never becomes `up-to-date` and leaves the
  prior installed context unchanged.

## Validation evidence

Every gate below was run as a standalone command with its real exit status:

```text
go test ./internal/install -run '^(TestRepeatedUnchangedCompiledInstallIsUpToDateWithoutRestagingContext|TestRepeatedUnchangedGlobalCompiledInstallIsUpToDateWithoutRestagingContext|TestCompiledInstallDriftStillRestagesContext|TestDerivedBuildCurrentnessUsesCompleteStaticFileSets|TestUnavailableCompiledCurrentnessFailsClosedAndPreservesInstalledContext)$' -count=1
exit 0

go test -race ./internal/install -run '^(TestRepeatedUnchangedCompiledInstallIsUpToDateWithoutRestagingContext|TestRepeatedUnchangedGlobalCompiledInstallIsUpToDateWithoutRestagingContext|TestCompiledInstallDriftStillRestagesContext|TestDerivedBuildCurrentnessUsesCompleteStaticFileSets|TestUnavailableCompiledCurrentnessFailsClosedAndPreservesInstalledContext)$' -count=1 -timeout=10m
exit 0

go test ./internal/install -count=1
exit 0

go test ./... -count=1
exit 0

go build ./...
exit 0

go vet ./...
exit 0

/Users/iv/go/bin/golangci-lint run
first run exit 1: two unused test callback parameter names
after correction exit 0: 0 issues

files=(internal/install/targets.go internal/install/install.go internal/install/global.go internal/install/stage_test.go); unformatted=$(gofmt -l "${files[@]}"); if [[ -n "$unformatted" ]]; then print -r -- "$unformatted"; exit 1; fi
exit 0

git diff --check
exit 0
```

After the lint-only callback parameter rename, the full focused normal suite
was rerun and exited 0.

## Concurrent-task reconciliation

At the final comparison checkpoint, concurrent `TASK-260720-1nlmvv` had no
delta in `internal/install/targets.go`. Its overlapping
`internal/install/install.go` and `internal/install/global.go` changes were
limited to build-diagnostic reporting before the staging functions. This
task's staging call-site hunks do not overlap those changes and can be composed
without priority rules or compatibility shims.

The root cause, eligibility decision, boundary behavior, test evidence, and
reconciliation finding were also recorded in `LOGBOOK.md`.
