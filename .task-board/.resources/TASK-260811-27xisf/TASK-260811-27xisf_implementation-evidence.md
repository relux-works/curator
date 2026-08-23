# TASK-260811-27xisf implementation evidence

Board run: `RUN-260817-3f4e56`

Authoritative goal at the evidence checkpoint: `GOAL-260817-41e579`
revision 1, resolved scope `TASK-260811-27xisf`.

## Delivered code

Added `internal/closureexec`, a manager-neutral protected closure substrate:

- immutable content-addressed raw capture handles with owner-only storage,
  admission-receipt binding, and immediate content and boundary rechecks;
- canonical CCJ intake-admission, derivation-permit, derivation-receipt, and
  task-private derived-manager-cache records using the accepted domain labels;
- serialized pre-C5 execution whose committed permits bind C0 checkpoint,
  toolchain node and tree fingerprint, exact executable digest, argv, cwd,
  environment, host/target, process/read/write/network policy, expected
  evidence, admitted inputs, output limit, and immediate before/after rechecks;
- a sealed manager-owned execution-boundary interface, a Darwin
  `sandbox-exec` implementation with default deny and network denial, and
  fail-closed unsupported behavior on platforms without an implementation;
- absent-root task-private workspace creation with empty home/config/cache/
  output/temp roots and canonical derived-cache observations;
- generic sorted multi-output protected publication using immutable blobs and
  a no-replace canonical receipt as the atomic visibility point;
- independently derived expected-cache lookup, complete write-set/output-set
  reconciliation, read-only boundary checks, exact-hit reuse, and rejection of
  mutable, drifted, partial, undeclared, or poisoned entries.

The implementation consumes the existing `closuregraph` source closure,
expected-cache-input, produced-observation, execution-receipt, and
publication-receipt records. It does not alter capture, C4, C5, closure
identity, the Go cache, or ecosystem resolution.

## Test evidence

Focused tests cover immutable capture mutation, admitted-handle binding,
permit-before-start, input absence, exact executable/toolchain drift with zero
starts, observed network drift with no receipt, empty roots and poisoned
ambient cache, task-private derived cache receipts, undeclared outputs, sorted
multi-output publication, exact protected reuse, protected root/blob poison,
and the exact CGP10 `one`/`two` publication IDs.

Final standalone gates, all exit 0:

- `go test -race -count=1 ./internal/closureexec`
  (`ok`, 2.629s)
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run`
  (`0 issues.`)
- `go vet ./...`
- `go build ./...`
- `go test -count=1 ./...`
  (`cmd/curator` 395.112s; `artifactpolicy` 185.650s;
  `godriver` 110.665s; `install` 130.916s; all packages green)
- `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb
  .research/260811_cross-language-closure-graph-and-checkpoints.md`
  (`canonical_goldens=pass labeled_records=53`, references pass)
- gofmt cleanliness and `git diff --check` both exited 0.

The explicit compatibility run over `closureexec`, `closuregraph`,
`artifactpolicy`, `buildcache`, `godriver`, and `buildsource` also exited 0.

## Honest anomaly record

- The first `golangci-lint` invocation exited 127 because the binary was not
  installed. The repository-pinned v2.12.2 was then invoked with `go run`.
- Its first pinned run exited 1 with actionable lint findings; those findings
  were corrected. The final focused and repository-wide pinned runs exit 0.
- The first canonical-verifier invocation used the non-materialized attachment
  path and exited 1. The byte-identical repository verifier was located and
  the corrected standalone command exited 0.

No files were staged or committed. Existing unrelated worktree and board
changes were preserved.
