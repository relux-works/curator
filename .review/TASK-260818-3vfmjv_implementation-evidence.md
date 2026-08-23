# TASK-260818-3vfmjv implementation evidence

Board run: `RUN-260817-8b76d8`

Authoritative goal at the latest directive checkpoint: `GOAL-260817-d7105f`
revision 1, resolved scope `TASK-260818-3vfmjv`.

## Delivered rework

- The executor now consumes committed permits exactly once while holding the
  causal-head lock across the process-start seam. A competing same-head permit
  is rejected after the head advances, starts zero providers/processes, and
  issues no receipt.
- Protected intake supports immutable source snapshot trees. Canonical tree
  identity covers every path, size, and digest; replay exposes only permit-
  named admitted handles; roots and members are read-only; containment,
  symlink/special-node, permission, size, digest, and file-set identity are
  rechecked immediately before provider use.
- `EnforceObserveProvider` is a pluggable authoritative enforcement and
  observation boundary. Nil and enforcement-only providers fail before any
  start. The built-in Darwin constructor now fails closed because
  `sandbox-exec` is not a lossless observer. This work does not claim or
  synthesize Endpoint Security support.
- Publication requires immutable `PublicationEvidence` for exact C4, C5,
  source closure, active graph, build plan, action/output/produces records,
  paths/classes, targets, and toolchains. The plan is rederived from C4 before
  any blob becomes visible. Poisoned, wrong-kind, wrong-path/class, stale
  closure/checkpoint, target, and tool references create no protected entry.
- Protected entries retain each canonical observation. Every hit decodes and
  rehashes the entry, publication receipt, observations, paths, sizes, digests,
  blobs, expected output set, and execution references; receipt tampering,
  substitutions, duplicate/missing/extra outputs, and size/digest drift reject.
- Typed canonical derivation permits and receipts bind resource-limit values
  and identities, evidence schemas, artifact manifests, output paths/digests/
  sizes, explicit deterministic diagnostics, and the derived next causal head.
  Strict decoders reject unknown/noncanonical/drifted records. Manifest,
  vendor, mirror, and metadata variants round-trip canonically.
- Multi-output protected publication is exercised with a fully derived C4/C5
  graph, not a shape-only fixture. Existing CGP10 publication identities remain
  pinned and unchanged.

Manager neutrality is preserved: ecosystem adapters provide declarations, not
authority. No Kotlin path was added. The existing global compiled dependency
deny and accepted graph/canonical identities remain covered by repository and
canonical compatibility gates.

## Security-negative coverage

Focused tests prove stale-permit zero-start behavior; missing, mutated, linked,
writable, and substituted replay-tree rejection; ignored denied process/read/
write/network/evidence/output observations producing zero receipts; lossless-
provider absence failing closed; poisoned C4/C5/closure/action/output/produces/
path/class/target/tool references producing no publication; and tampered,
substituted, duplicate, missing, extra, size-drifted, digest-drifted, or wrong-
execution protected hits rejecting reuse.

## Final standalone gates

Every command below was run directly as a standalone process and exited 0:

- `go test -race -count=1 -cover ./internal/closureexec`
  (`coverage: 72.0% of statements`)
- `go test -count=1 ./internal/closureexec ./internal/closuregraph ./internal/artifactpolicy ./internal/buildcache ./internal/godriver ./internal/buildsource`
- `go test -count=1 ./...` (`cmd/curator` 382.002s;
  `artifactpolicy` 135.631s; `install/atomicity` 117.290s)
- `go vet ./...`
- `go build ./...`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run`
  (`0 issues.`)
- `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb .research/260811_cross-language-closure-graph-and-checkpoints.md`
  (`canonical_goldens=pass labeled_records=53`, all references pass)
- `test -z "$(gofmt -l internal/closureexec internal/closuregraph)"`
- `git diff --check`
- `task-board validate` (`Board is valid. No issues found.`)

## Development-loop anomalies

- The inherited partial worktree initially failed focused compilation with
  stale `Boundary`, `OutputLimit`, evidence, and fake-provider APIs (exit 1).
  The implementation and tests were reconciled to the lossless provider model.
- One focused run exited 1 because read-only snapshot permissions prevented
  `testing.TempDir` cleanup; test cleanup now restores permissions without
  weakening production replay permissions.
- Pinned lint truthfully exited 1 first with seven findings and once more with
  one conservative slice-index warning. All findings were corrected; the final
  pinned run exits 0 with zero issues.
- The full repository suite was rerun after the completion audit added explicit
  diagnostics and graph-backed multi-output coverage. The final run above is
  the authoritative result.

No files were staged or committed. Existing unrelated worktree and board
changes were preserved.
