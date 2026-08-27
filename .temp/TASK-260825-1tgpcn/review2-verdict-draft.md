# TASK-260825-1tgpcn — Review Verdict (cycle 2)

**Verdict: ACCEPTED → done.** The rework delivers exactly what cycle 1 requested,
nothing else moved, and every gate is green — re-run independently by this
reviewer, including the regression proof.

## What cycle 1 requested, and what landed

1. **Bound the call, not just Git.** `internal/gitcred/gitcred.go:62` adds
   `drainDelay = 2 * time.Second`; `gitcred.go:275` sets
   `cmd.WaitDelay = drainDelay` in `Access.call`. One credential call now costs
   at most `Timeout + drainDelay` even when the helper — Git's child, holding
   Git's inherited stderr pipe — outlives the context kill.
2. **Pin the fix in the harness.** `gitcred_test.go` adds `modeHangGrandchild`
   (the stand-in Git spawns a helper and blocks on it, handing it its own
   stderr — the exact pipe the manager reads) and `modeSleeper` (the helper:
   the test binary re-executed, sleeping 30s, recording nothing so it is not
   counted as a credential call), plus
   `TestACallIsBoundedWhenTheHelperOutlivesGit` asserting the elapsed bound.

## Scope held

Environment construction, refusal guards, read-back verification and message
text are byte-for-byte what cycle 1 reviewed and accepted; the full-file
re-read against the cycle-1 verdict found no delta beyond the two constants,
the one `WaitDelay` line, the two harness modes, the helper-spawn function and
the new test. `TestACallIsBounded` untouched (0.22s). `go.mod`/`go.sum`
untouched — the package remains dependency-free stdlib.

## Regression proof, reproduced independently

Scratch copy of the package with only the `cmd.WaitDelay = drainDelay` line
removed (`.temp/TASK-260825-1tgpcn/review2-regress/`):

| Case | `cmd.WaitDelay` | Elapsed | Result |
| --- | --- | ---: | --- |
| grandchild helper outlives the kill | unset | 30.02s | FAIL — call outlived the orphan |
| grandchild helper outlives the kill | `drainDelay` (2s) | 2.23s | PASS |
| stand-in Git itself hangs (existing test) | `drainDelay` | 0.22s | PASS |

The test fails without the fix and passes with it, on this host
(darwin/arm64, go1.25.5) — the fix is pinned against a silent refactor back to
`exec.CommandContext` alone.

## Reviewer validation (all re-run by the reviewer, real exit codes)

| Command | Exit |
| --- | ---: |
| `go test -count=1 ./internal/gitcred/` | 0 |
| `go test -race -count=1 ./internal/gitcred/` | 0 (22.4s) |
| `go test -count=1 <all packages except cmd/curator>` | 0 |
| `go test -timeout 30m -count=1 ./...` | 1 — 41/42 ok incl. `cmd/curator` (882.7s) and `gitcred`; the one FAIL is an unrelated pre-existing flake, see below |
| `go test -race -count=1 ./internal/install/ ./internal/config/` (gitcred consumers) | 0 |
| `go test -race -count=5 -run TestACallIsBounded ./internal/gitcred/` (flake screen under heavy load) | 0 — 10/10 PASS |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `golangci-lint run ./internal/gitcred/...` | 0 (0 issues) |
| `golangci-lint run ./...` | 0 — the 5 sibling-owned issues the implementer reported have since been fixed by their owners |
| `gofmt -l internal/gitcred/` | clean |

Real-git integration tests pass against the operator's git
(`TestRealGitKeepsTheManagerEntrySeparate` 0.36s,
`TestRealGitWithoutAHelperIsCaughtByTheReadBack` 0.09s).

## Anomaly, not attributable to this delivery: flaky registry test under load

The reviewer's full-tree run exited 1 on `TestRegistryAttestationLandsInMarker`
(`internal/install/registry_e2e_test.go:104`): "registry test-reg snapshot
timestamp is too far in the future" → "every trusted audit registry served a
tampered snapshot". Attribution evidence:

- the test file and the tolerance check it trips
  (`internal/registry/snapshot.go:159`) are untouched in the working tree; the
  test was last changed in committed history (`cfffd7c`);
- registry attestation shares nothing with credential reads; `internal/gitcred`
  is not on its path;
- the same package passed three other times today in the same tree: the
  reviewer's non-curator sweep (exit 0), the reviewer's
  `-race` consumer run (`ok internal/install 60.8s`), and a targeted solo rerun
  after the failure (`PASS 1.04s`);
- the one failure happened while **three** full `curator.test` suites from
  concurrent sessions competed for the machine (`internal/install` took 371.6s
  against its usual ~60s), which points at a wall-clock tolerance that a loaded
  fixture-to-validation gap can exceed.

Verdict is unaffected — the failure is a pre-existing, load-sensitive test
outside this delivery. Flagged for the orchestrator: the timestamp-skew
tolerance in the registry snapshot validation deserves its own task before CI
runs suites in parallel.

## Logbook

Entry `2026-08-25 0223` present: names the superseded "(15s bound)" claim in
0158, records the mechanism, both fix sites, why `TestACallIsBounded` could
not see the defect, the measured 30.008s vs 2.21s, and the NOTE generalizing
the `WaitDelay` hole to other `exec.CommandContext` sites in the repo.

## AC (all verified across both cycles)

- Every read and write is `git credential fill|approve|reject` with prompting
  disabled four ways — pinned by argv/environment assertions.
- Operator home pinned (`HOME` + `USERPROFILE`, case-insensitive suppression);
  proven by the store landing in the pinned home's `.git-credentials`.
- A helper that persists nothing is caught by read-back, with guidance for all
  three platforms plus the env-var alternative, and no secret in the message.
- Namespaced entries (`curator-build-https:<scope>`) cannot collide with the
  operator's own — two refusal guards, pinned against real `store` helper
  front-insertion behaviour.
- go test green (package, race, full tree).

## Commit note for the mover

This reviewer records acceptance evidence only and supplies no `commit_ack`.
The delivery is `internal/gitcred/` (new package, currently untracked) plus
LOGBOOK entries 0158/0223; the commit-owning mover commits that scope.
