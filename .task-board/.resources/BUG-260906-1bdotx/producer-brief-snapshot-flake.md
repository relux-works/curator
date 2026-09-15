# Producer brief: the Windows snapshot-timestamp flake

## Where and what

- Repository `~/Developer/ReluxWorks/curator`. Branch and worktree named at spawn; base is curator
  main after PR #63 lands — the orchestrator confirms the exact OID. First run
  `git submodule update --init --recursive`.
- The failing case: `internal/install` `TestRegistryAttestationLandsInMarker`, on
  `Test (windows-latest)` only.

## What is established

Run 34022458940, a `workflow_dispatch` of the candidate lane on `feat/agent-environments-stage-b` at
`1a936e77`. The Windows job failed after 39m44s with:

```
registry test-reg snapshot timestamp is too far in the future
every trusted audit registry served a tampered snapshot
```

Every other job in that run was green, including all three Candidate suite jobs. The **same test on
the same head passed** in the pull-request run for PR #60. Stage (b) does not touch
`internal/install` — `git diff origin/main..HEAD -- internal/install` was empty.

The three pieces of the mechanism, read from the tree:

- the fixture writes `"created_at": time.Now().UTC().Format(time.RFC3339)`
  (`internal/install/registry_e2e_test.go:44`) — RFC3339 truncates fractional seconds **downward**,
  so the value it writes is never ahead of the clock that wrote it;
- the gate is `parsed.CreatedAt.After(now.Add(clockSkew))` (`internal/registry/snapshot.go:158`);
- `now` is `time.Now()` at the call site (`internal/install/install.go:1192`), and `clockSkew` is
  `cfg.Audit.SnapshotClockSkewSeconds`, whose default is **300 seconds**
  (`internal/config/config.go:41`).

So for this to fire, the timestamp must be more than five minutes ahead of a `time.Now()` sampled
*later* than the one that produced it. A monotonic reading cannot do that. Something else did.

## Your task

**Establish the mechanism from evidence before changing anything.** Candidate explanations, none of
which you should assume:

- the hosted runner's wall clock stepped backwards mid-job — plausible on a 40-minute Windows job
  under NTP correction, and it would explain why only the longest job in the matrix ever shows it;
- the fixture's timestamp is produced somewhere other than where it appears to be — check whether the
  handler re-generates it per fetch, and when relative to the check;
- a timezone or parse asymmetry between what the fixture formats and what `parsed.CreatedAt` holds;
- the configured skew is not the default in this test's path.

Whatever you find, say how you established it. If you cannot establish it from evidence, say that
plainly and propose the smallest change that makes the test deterministic **without** weakening what
the gate detects — that is an acceptable outcome, an unproven guess dressed as a diagnosis is not.

## The line you must not cross

`environment "snapshot timestamp is too far in the future"` is a **tampering signal**. A snapshot
dated in the future is exactly what a rollback or equivocation attack looks like, and this gate exists
to catch it. Do not widen `clockSkew`, do not delete the assertion, and do not make the test ignore a
genuinely future timestamp. If the fix is a deterministic clock the checker and the fixture share,
that is right; if it is a tolerance, it must be justified from the mechanism you established and
stated as a bound.

Prove the gate still catches what it is for: a narrowing mutant that lets exactly one future-dated
snapshot through must fail a named test. If no such test exists today, that is itself part of the fix.

## Method

- Drive the production path, not the helper.
- Anchor each `-run` level separately and count `=== RUN` lines: `go test -run '^(Parent/child)$'`
  splits on the unbracketed slash, matches nothing and **exits 0**, which reads as a passing run.
- This defect appears on Windows only and no local lane can see it. `ledger-consistency.sh` proves
  compilation across GOOS but that is an argument, not a measurement — say so if you rely on it. The
  orchestrator will dispatch the hosted lanes; do not claim a Windows result you did not measure.

## Delivery

Small signed commits, human identity. **Do not write `LOGBOOK.md`.** Do not push, do not open a PR.
**Do not stage with `git add -A`.** Gates, each a standalone process with its observed exit code:
`go build ./...`, `go vet ./...`, `gofmt -l cmd internal`, `golangci-lint run ./...`,
`bash .github/ci/gate-selftest.sh`, `bash .github/ci/ledger-consistency.sh`, and
`go test -count=1 -race ./internal/install/ ./internal/registry/`. Materialize any conformance root as
a plain checkout verified against `manifest.json`, never with `git archive`; run the two `test-gate`
lanes sequentially.

Attach `BUG-260906-1bdotx_drafting-report.md`: the mechanism with the evidence that established it (or
the honest statement that it could not be established); the fix and why it does not weaken the gate;
the narrowing mutant and the named test it kills; and the gate table. Then
`task-board handoff BUG-260906-1bdotx --role developer`.
