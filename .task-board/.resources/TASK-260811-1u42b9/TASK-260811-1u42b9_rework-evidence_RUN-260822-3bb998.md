# TASK-260811-1u42b9 rework evidence — RUN-260822-3bb998

Status: developer handoff evidence

## Implemented correction

- npm cache derivation and `npm ci` now execute the exact C0-bound Node binary
  directly and pass the exact fingerprinted staged `npm-cli.js` as argv[0].
  The npm shebang is never executed and the closed npm environment contains no
  `PATH` fallback.
- Common Node tool identity now binds the interpreted entry point and complete
  external-toolchain read roots into the toolchain node's invocation-facts
  digest. npm C5 templates bind the concrete manager entry point, argv, cwd,
  environment, read roots, write roots, and process set.
- The concrete shared process runner exposes a copied, instrumentation-only
  observation of the exact command at its real process-start seam. The verified
  integration fixture derives its process evidence from that observed launch,
  checks the physical Node executable, resolved npm CLI path, argv, cwd, and
  environment, and then submits the canonical logical audit.
- Unbound shebang/PATH substitutions fail before process start. Portable and
  verified real-operation tests both prove that Node launches npm and that no
  ambient PATH is present.
- Shared verified execution now preserves the stable
  `closure_network_attempted` diagnostic when a lossless provider observes a
  network attempt, before any causal receipt can be issued.
- npm-wrapped S03 and S08 run real offline materialization twice, seed a
  poisoned but unreferenced ambient npm cache between runs, and compare the
  complete materialized file inventory plus selected graph. Portable network
  evidence remains honestly `not-observed`.
- npm-wrapped S04 executes the declared cache action through the verified
  provider, reports an attempted TCP operation, asserts
  `closure_network_attempted`, one action start, no receipt, and no cache
  publication.
- Real replay exposed timestamped npm debug logs under the private HOME. They
  are manager scratch rather than source output and made two otherwise exact
  materializations differ. The closed private log root is now discarded before
  the retained project tree is reconciled and admitted.
- README npm behavior and exact integration-test entry points were updated.

## Development-loop evidence

The initial red iterations were retained honestly:

- focused compile attempt: exit 1 (`stringsAny` helper did not exist);
- focused attempt after compile repair: exit 1 (tool invocation-facts validation
  incorrectly required every tool with read roots to have an interpreter entry
  point);
- focused attempt after validation repair: exit 1 (fake Node observation and
  two-run identity expectation needed correction);
- focused attempt with complete inventory comparison: exit 1, proving real npm
  timestamped debug logs caused replay drift;
- subsequent targeted and focused runs exited 0 after the production fix.

## Fresh validation

Every command below ran directly as a standalone process; no gate was piped
through `tee` or a status-masking shell chain.

| Gate | Exit | Evidence |
| --- | ---: | --- |
| Real S03/S04/S08, direct Node npm, verified real launch, and zero-start vectors | 0 | `go test -count=1 ./internal/npmsource -run 'Test(S02S03S04S07S08N12N13OfflineMaterializationGates|N01RealNPMCIUsesOnlyDerivedPrivateCache|VerifiedProviderObservesRealNodeLaunchedNPMBoundary|C5AndExactExecutableGatesAreZeroStart)$'`; `7.383s` |
| Affected focused suite | 0 | `go test -count=1 ./internal/artifactpolicy ./internal/closureexec ./internal/nodesource ./internal/npmsource`; artifactpolicy `35.577s`, closureexec `2.628s`, nodesource `1.747s`, npmsource `10.160s` |
| npm race | 0 | Post-final-code rerun: `go test -count=1 -race ./internal/npmsource`; `35.006s` |
| npm coverage | 0 | Post-final-code rerun: `go test -count=1 -cover ./internal/npmsource`; `80.4%` statements, `9.789s` |
| Vet | 0 | `go vet ./internal/npmsource ./internal/nodesource ./internal/closureexec ./internal/artifactpolicy` |
| Lint | 0 | `golangci-lint run ./internal/npmsource ./internal/nodesource ./internal/closureexec ./internal/artifactpolicy`; `0 issues` |
| Build | 0 | `go build ./...` |
| Diff whitespace | 0 | `git diff --check` |
| Board validation before handoff | 0 | `task-board validate`; board valid |
| Uncached repository suite | 0 | `go test -count=1 ./...`; cmd/curator `390.840s`, artifactpolicy `141.560s`, closureexec `23.237s`, npmsource `87.357s`, rustsource `153.666s`, install `120.687s`, install/atomicity `121.213s` |

Tool readiness was recorded separately at
`.temp/TASK-260811-1u42b9/tool-readiness-01.log`; every required tool probe
exited 0.

## Scope and review handoff

No files were staged or committed. The worktree already contains broader
uncommitted implementation work from the parent delivery; this rework stayed
within the npm/common Node/shared executor behavior and its README contract.
The implementation is ready for independent review against the latest review
focus and all task acceptance criteria.
