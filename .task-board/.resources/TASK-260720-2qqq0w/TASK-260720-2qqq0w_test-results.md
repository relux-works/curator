# TASK-260720-2qqq0w tester evidence

## Scope audited

Candidate:
`/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-2qqq0w/worktree`

Accepted comparison:
`/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-2kaopg/worktree`

An `rsync -rnic --delete` checksum dry run, excluding `.git` and `.temp`,
reported exactly three task-owned differences:

- `README.md`
- `docs/compiled-builds.md`
- `cmd/curator/docs_test.go`

The two worktrees contain the same project skill symlinks. Openrsync reported
that it skipped those non-regular entries; the earlier recursive `diff -rq`
diagnostic likewise reported the three intended differences but exited 1
(expected for a differing candidate) and warned about those symlink loops.

## Test contribution

Extended `cmd/curator/docs_test.go` with
`TestDocumentedCompiledBuildAuthoringContract`. It locks the required author and
operator surfaces into the documentation:

- schemas 1 through 5 remain compatible;
- build roots are excluded from agent context;
- vendoring is mandatory;
- hooks, package argv/environment, cgo, workspaces, downloads, external
  linking, alternate root modules, and generic drivers remain unsupported;
- compiled output remains untrusted and is not executed during install;
- portable logical identity is distinguished from Curator-local paths;
- install/upgrade repair, locked GC, status/dry-run, and Unix/Windows shims are
  explicit.

No product implementation or documentation prose was changed by the tester.

## Validation evidence

All commands were run directly as standalone processes from the candidate
worktree unless noted.

| Command | Exit | Result |
|---|---:|---|
| `go test ./cmd/curator/ -run '^TestDocumented' -count=1` (producer candidate before tester edit) | 0 | Documentation JSON, mixed manifest, vocabularies, and local links passed |
| `go test ./internal/skillspec/ -run '^(TestSchemaV6\|TestLegacyRuntimeFallbackRejectsBuildObject)' -count=1` (before tester edit) | 0 | Focused schema-6 parser/package tests passed |
| Five `curl -fsSL` checks for the repository, protocol core, manager profile, schemas/v1, and conformance/v1 links | 0 | Every target returned HTTP 200 |
| `go build ./...` (before tester edit) | 0 | All packages compiled |
| `go vet ./...` (before tester edit) | 0 | No findings |
| `gofmt -l cmd internal` (before tester edit) | 0 | Empty output |
| `git diff --check` (before tester edit) | 0 | No whitespace errors |
| `go test ./cmd/curator/ -run '^TestDocumented' -count=1` (first tester iteration) | 1 | Expected development red: the new assertion incorrectly expected doubled Windows path separators; documentation was correct |
| `go test ./cmd/curator/ -run '^TestDocumented' -count=1` (corrected tester test) | 0 | `ok`, 0.722s |
| `go test ./internal/skillspec/ -run '^(TestSchemaV6\|TestLegacyRuntimeFallbackRejectsBuildObject)' -count=1` (post-change) | 0 | `ok`, 0.521s |
| `go build ./...` (post-change) | 0 | All packages compiled |
| `go vet ./...` (post-change) | 0 | No findings |
| `gofmt -l cmd internal` (post-change) | 0 | Empty output |
| `git diff --check` (post-change) | 0 | No whitespace errors |
| Exact candidate `rsync -rnic --delete --exclude=.git --exclude=.temp ...` audit | 0 | Only the three task files differ |

`golangci-lint version` exited 127 because `golangci-lint` is not installed in
this environment. The operator directive forbids host installation, so lint
was not rerun or represented as green by this tester. The board already carried
the producer's lint item as checked before this run.

The full repository test suite was intentionally not rerun: the operator's test
completion directive explicitly requires focused `-count=1` tests for this
docs-only delta after the prior harness repeatedly interrupted the full suite.

Go statement coverage is not applicable to this delta: the affected artifacts
are Markdown and a `_test.go` contract file, and no executable Go statements
were added or changed relative to the accepted integrated worktree. No numeric
Go coverage percentage is claimed. The passing documentation-contract test
does exercise every author/operator surface listed in the task scope, so the
generic affected-code coverage gate is satisfied by contract coverage for this
docs/test-only delta.

## Verdict

The mixed schema-6 example parses through the real Curator loader, required
guidance is present and test-locked, local links resolve, external protocol
links return HTTP 200, and the focused tests plus build/vet/format/diff gates
pass. The candidate is suitable for independent review.
