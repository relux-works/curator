# TASK-260811-1u42b9 rework evidence

Run: `RUN-260822-cb137e`

Role: developer

Review input: changes requested by `RUN-260822-0c4aea`

## Implemented corrections

1. Raw admitted npm tarballs now produce a canonical, sorted extraction inventory of every regular package member: path, SHA-256, size, and executable bit. The fake positive installer extracts those exact admitted bytes; a package-name-only synthetic tree no longer satisfies the positive vector.
2. After `npm ci`, each selected external package is recursively re-admitted through `artifactpolicy`, its embedded `package.json`, lifecycle/native/bundle metadata, and selected graph identity are reconciled again, and its complete owned file inventory must equal the admitted tarball extraction evidence. Selected nested dependencies are treated as separate owners; an unknown nested `node_modules` tree fails as a bundled dependency.
3. npm cache derivation, `npm ci`, and Node invocation use one preflighted `RunnerBinding` containing the complete `closureexec.AssuranceBinding`, common Node selection, exact `nodesource.RuntimeBinding` with C0 checkpoint, and C5 build plan. C0 and C5 are independently rederived from the admitted capture before any start. Each invocation carries C0/C5 IDs, exact tool identity, full assurance binding, and a domain-separated precommitted permit ID. Provider binding and tool identity are rechecked immediately before every start; tool identity is also checked after execution.
4. Portable audit evidence contains only assurance mode and exit code. Lossless observations are a nullable verified-only object with `omitempty`; portable serialization cannot synthesize resolver/cache/lifecycle/process/read/write zero counters. Verified mode requires the exact complete provider binding and lossless observation envelope.
5. README npm profile documentation now states exact installed-content reconciliation, verified C0/C5/permit binding, and honest portable evidence semantics.

## Regression coverage

- substituted materialized JavaScript bytes fail `closure_integrity_mismatch`;
- materialized direct, renamed, and nested compiled payloads fail the shared compiled diagnostic;
- materialized opaque payloads fail the shared opaque diagnostic;
- materialized `binding.gyp` fails `closure_native_build_unsupported`;
- an unadmitted nested `node_modules` tree fails `closure_bundled_dependency_unsupported`;
- missing provider, incomplete tool identity, incompatible provider contract, cross-mode plan, and drifted preflight all fail before any npm start (`starts == 0`);
- portable audit JSON contains none of the verified-only observation fields.

## Validation evidence

Every gate below was run as a standalone process with its real exit preserved.

| Gate | Exit | Evidence |
| --- | ---: | --- |
| `go test -count=1 ./internal/artifactpolicy ./internal/nodesource ./internal/npmsource` | 0 | `focused-03.log`; artifactpolicy `227.021s`, nodesource `3.808s`, npmsource `44.260s` |
| `go test -count=1 -race ./internal/npmsource` | 0 | `race-03.log`; `13.627s` |
| `go test -count=1 -cover ./internal/npmsource` | 0 | `coverage-03.log`; `80.1%` statements |
| `go vet ./internal/npmsource ./internal/nodesource ./internal/artifactpolicy` | 0 | `vet-03.log` |
| `golangci-lint run ./internal/npmsource ./internal/nodesource ./internal/artifactpolicy` | 1 | `lint-02.log`; first run truthfully failed on one unused parameter, one deprecated tar constant, and one unused test helper |
| same lint command after correction | 0 | `lint-03.log`; `0 issues` |
| `go build ./...` | 0 | `build-02.log` |
| `git diff --check` | 0 | `diff-check-03.log` |
| `task-board validate` | 0 | `board-validate-03.log`; board valid |
| `go test -count=1 ./...` | 0 | `repository-suite-02.log`; cmd/curator `535.107s`, artifactpolicy `161.674s`, rustsource `166.344s`, npmsource `20.404s` |

No gate result is inferred from an expected failure. The failed lint attempt is recorded as exit 1 and was followed by a fresh green lint run plus fresh focused, race, coverage, vet, build, diff, and uncached repository-wide validation.

## Files in scope

- `internal/npmsource/capture.go`
- `internal/npmsource/materialize.go`
- `internal/npmsource/errors.go`
- `internal/npmsource/conformance_test.go`
- `README.md`

No files were staged or committed.
