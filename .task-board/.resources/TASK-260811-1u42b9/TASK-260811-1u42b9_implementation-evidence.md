# TASK-260811-1u42b9 npm source-closure implementation evidence

## Outcome

Implemented `internal/npmsource` as the separate closed npm manager profile on
top of `artifactpolicy`, `closureexec`, `closuregraph`, and `nodesource`.

The adapter now:

- parses root `package-lock.json` and `npm-shrinkwrap.json` schemas 2 and 3;
- reconciles root/workspace manifests, v2 legacy dependency trees, workspace
  links, peer/optional/development edges, and exact selected/pruned target
  decisions;
- requires explicitly approved HTTPS registry origins, exact SHA-512 SRI, and
  separately records raw lock SHA-256 and canonical semantic lock identity;
- captures every exact raw `.tgz`, admits it through the shared recursive
  artifact classifier, reconciles embedded package metadata, and rejects
  lifecycle, bundle, implicit `binding.gyp`/node-gyp, native, Wasm, V8-cache,
  opaque, unsafe, or metadata-drifted inputs before npm execution;
- derives a deterministic private cacache content store from admitted tarballs,
  pruning temporary-local-locator/time-dependent index records;
- materializes with `npm ci --offline --ignore-scripts` in fresh private
  home/config/cache/output roots and reconciles the installed tree exactly;
- keeps portable assurance functional/default without inventing lossless host
  observations, while verified mode requires a compatible lossless provider
  before the first process start;
- composes npm output/invocation behavior with the common Node generated-output
  and runtime contracts.

## Shared contract fixes

- `artifactpolicy`: a gzip root below a portable virtual directory now emits
  its decoded member directly under the compressed-stream node, preserving
  compressed-size accounting; `binding.gyp` is recognized as inspectable Node
  metadata. Both have regression tests.
- `nodesource`: multiple workspace package instances may reference the same
  admitted workspace artifact manifest without duplicating the capture graph's
  unique manifest ID set. A regression test covers the shared snapshot case.

## Conformance coverage

The npm test corpus names and covers shared `S01`-`S08` and npm `N01`-`N13`
semantics, including a real Darwin `sandbox-exec` network-denied npm integration
run. That integration derives the private cache twice and requires identical
receipt IDs before running real offline `npm ci`. Separate tests prove portable
success and verified-missing-provider rejection with process-start count zero.

## Validation evidence

Every command below ran as a standalone process.

| Command | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 ./internal/artifactpolicy ./internal/nodesource ./internal/npmsource` | 0 | `artifactpolicy 42.498s`, `nodesource 1.690s`, `npmsource 3.490s` |
| `go test -count=1 -race ./internal/npmsource` | 0 | pass, `6.147s` |
| `go test -count=1 -cover ./internal/npmsource` | 0 | pass, `80.1%` statements |
| `go vet ./internal/npmsource ./internal/nodesource ./internal/artifactpolicy` | 0 | no findings |
| `golangci-lint run ./internal/npmsource ./internal/nodesource ./internal/artifactpolicy` | 0 | `0 issues` |
| `go build ./...` | 0 | repository build succeeds |
| `git diff --check` | 0 | clean |
| `go test -count=1 ./...` | 0 | repository suite passed before the final assurance-mode correction; slowest package `cmd/curator 526.851s` under concurrent full-suite CPU contention |

After the assurance-mode correction, the full affected package suite, race,
coverage, vet, lint, repository build, and diff gates above were rerun green.
The repository-wide suite was not redundantly rerun because the final delta was
contained to `internal/npmsource` plus README and `go build ./...` revalidated
all packages.

## Files

- `internal/npmsource/errors.go`
- `internal/npmsource/lock.go`
- `internal/npmsource/capture.go`
- `internal/npmsource/materialize.go`
- `internal/npmsource/conformance_test.go`
- shared regression updates in `internal/artifactpolicy` and `internal/nodesource`
- npm profile and tool documentation in `README.md`

No files were staged or committed.
