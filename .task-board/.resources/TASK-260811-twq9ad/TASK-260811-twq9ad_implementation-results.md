# TASK-260811-twq9ad developer rework outcome

## Review findings resolved

1. Yarn's default rc chain is disabled with `--no-default-rc`; the fixed C5
   action, concrete permit, and real launch all bind the flag. The child gets a
   minimal explicit environment without `PATH`, `YARN_RC_FILENAME`, or
   `npm_config_*` inheritance.
2. Capture discovers root/glob-expanded workspace manifests and `.yarnrc`,
   `.npmrc`, and `.yarnrc.yml` files from the immutable staged tree. The
   discovered sets must be bijective with parsed authority before Yarn starts.
3. The closed peer semver evaluator now applies correct zero-major caret upper
   bounds, rejects leading-zero releases, and rejects prerelease/compound range
   semantics outside the supported grammar. Exact prerelease equality remains
   deterministic.
4. The real Yarn 1.22.22 vector stages exact Node/Yarn inputs, materializes
   through the shared assured executor and verified-provider seam, and launches
   Yarn under macOS `sandbox-exec` with `(deny network*)`. It poisons ancestor,
   HOME, process-environment, and ambient cache/config inputs while retaining
   the same layout/config identity. Missing-artifact and omitted-authority
   branches assert zero Yarn starts and no publication path.

The shared Node artifact text grammar now admits `.yarnrc` as plain text
metadata. This was necessary because capture-time config bijection otherwise
made every declared Yarn Classic configuration impossible to admit. The Yarn
adapter still parses it with a narrow closed key/value grammar.

## Files in scope

- `internal/yarnclassicsource/{lock.go,capture.go,materialize.go,conformance_test.go}`
- `internal/artifactpolicy/{text.go,text_metadata_test.go}`
- `README.md`

The worktree already contained the original Yarn implementation and extensive
shared prerequisite changes. No files were staged, committed, reset, or
cleaned.

## Final verification

Every gate ran directly as a standalone process and the reported exit is the
real process exit.

| Gate | Exit | Result |
| --- | ---: | --- |
| `sandbox-exec -p '(version 1) (allow default) (deny network*)' /usr/bin/true` | 0 | OS network-denial harness readiness |
| `go test -count=1 ./internal/yarnclassicsource -run TestRealPinnedYarnClassicReadsPrivateMirrorWithEmptyOrdinaryCache -v` | 0 | real Yarn 1.22.22 OS-denied replay |
| `go test -count=1 -coverprofile=.temp/TASK-260811-twq9ad/yarn-cover-rework.out ./internal/yarnclassicsource` | 0 | 80.4% statement coverage |
| `go test -race -count=1 ./internal/yarnclassicsource` | 0 | race clean |
| `go vet ./internal/artifactpolicy ./internal/yarnclassicsource` | 0 | vet clean |
| `golangci-lint run ./internal/artifactpolicy/... ./internal/yarnclassicsource/...` | 0 | `0 issues.` |
| `go build ./internal/artifactpolicy ./internal/yarnclassicsource` | 0 | build green |
| `go test -count=1 ./...` | 0 | exact handoff-state repository suite green; `cmd/curator` 406.204s, Yarn 10.927s |
| `git diff --check` | 0 | whitespace clean |

## Handoff

The reviewer-requested rc-chain, workspace-closure, semver, and real offline
proof gaps are addressed. The implementation is ready for independent review.
