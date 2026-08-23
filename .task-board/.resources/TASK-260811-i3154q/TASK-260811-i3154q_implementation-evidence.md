# TASK-260811-i3154q implementation evidence

Status: developer rework evidence for independent review

- Developer run: `RUN-260811-7bf0c7`
- Authoritative goal at the latest checkpoint: `GOAL-260811-f88313`
  revision 1
- Resolved scope: `TASK-260811-i3154q`
- Rework authority:
  `TASK-260811-i3154q_review-verdict_RUN-260811-273b9c.md`

## Delivered model

`internal/closuregraph` implements the language-neutral capture, selection,
binding, active-graph, deterministic-plan, checkpoint, and receipt contract.
The package contains closed codecs for the ten node and eleven edge kinds,
selection-neutral `CaptureGraph`, `SelectionContext`, `SelectionBinding`,
`ActiveGraph`, stable ordering waves and cycle evidence, C0-C7 checkpoints,
immutable expected outputs, separate C6 observations, source closure, expected
cache input, execution/publication receipts, and adapter interfaces.

The exact accepted CGP05/CGP10 corpus remains at
`internal/closuregraph/testdata/canonical-goldens.txt`. Its SHA-256 is
`fed9657b33297650e64178d605c7ee7f47dde640fefd12a99dec3651bb9cadcb`.
The sorted per-file SHA-256 manifest for the current package hashes to
`f2ec25aa458d2a1993b136d09d9ab7cc84fc587c54fbe10d661cc7dc99abab29`.
No corpus record, selection-neutral capture rule, byte detector, sandbox
implementation, or Kotlin scope was added or changed by this rework.

## RUN-260811-273b9c reviewer finding closure

1. `ValidateCheckpointChain` now independently rederives the exact build plan
   from the supplied C4 `GraphBundle` and a separately supplied execution
   policy authority. A
   structurally self-consistent plan containing a foreign action, output,
   ordering edge, or wave fails with `closure_checkpoint_invalid` before any
   downstream checkpoint reference can legitimize it.
2. Every selected `generated_artifact` and immutable expected
   `output_artifact` must have exactly one selected `produces` edge. Zero
   producers fail with `artifact_generated_input_undeclared`; duplicate
   producers fail canonically with sorted action and edge evidence.
3. A target-level `reads` edge to a generated artifact now orders the unique
   producer before every selected action owned by that target. The derived arc
   retains the producer edge, target-to-action declaration path, and read edge;
   record permutations produce byte-identical plans and waves.
4. Explicit `platform_role_names:[]` no longer suppresses binding obligations.
   Command products, targets, actions, toolchains, interop boundaries, and
   expected outputs reject explicit-empty or domain-incomplete role lists,
   while an omitted field retains its schema-defined target/host default.
5. Each interop mode now requires its intrinsic ABI/runtime/protocol/interface/
   calling/link evidence. Every selected boundary additionally requires exactly
   one provider and consumer side, immutable provider evidence, compatible
   declared languages and ABI expectation, a shared exact platform, and a
   distinct provider-before-consumer action pair for compile/link modes.
   Non-subprocess boundaries also require an explicit toolchain-scoped binding
   on the same selected platform. Dynamic-load and host-extension providers
   must be produced outputs or selected external toolchains. Subprocess
   boundaries require distinct produced and published command outputs plus one
   matching invocation with protocol, argv, environment, working-directory,
   executable-resolution, and bundle evidence. Full selected-graph tests cover
   all seven modes, including the non-ordering dynamic-load and subprocess
   cases.
6. All node and edge optional-field decoders propagate type errors instead of
   silently normalizing them. Every public graph, plan, checkpoint, closure,
   observation, execution, and publication decoder now proves exact canonical
   decode/re-encode byte equality as a final invariant.

Permanent regressions cover the six findings in
`checkpoint_evidence_test.go`, `plan_test.go`, `codec_test.go`, and
`validation_regression_test.go`, including negative permutations and the
positive accepted interop/CGP05/CGP10/Go compatibility paths.

## Current validation checkpoint

Each command was run directly as a standalone process. Directive
`nudge:87ea1c` required the separately owned `TASK-260811-2gazym` source to
reach stable `to-review` before this task took the serialized full-suite slot.
That condition was satisfied before the current repository-wide gate began.
The sibling subsequently moved to independent `reviewing` while the gate was
running; no product source changed during the run, and the current package and
accepted-corpus digests remained unchanged afterward.

| Command | Exit | Evidence |
| --- | ---: | --- |
| `go test -count=1 ./internal/closuregraph` | 0 | focused package passed; latest run `1.188s` |
| focused twelve-test RUN-260811-273b9c regression selector | 0 | all six findings, every interop mode/tool binding/runtime contract, and positive/negative target-read permutation proofs passed; `0.550s` |
| focused CGP05/CGP10/Go golden selector | 0 | exact graph, plan, observation, and Go compatibility tests passed; `0.280s` |
| `ruby ...accepted-canonical-golden-verifier.rb ...accepted-cross-language-closure-graph-and-checkpoints.md` | 0 | all 53 labeled records and references passed; CGP05 capture reused |
| `go test -race -count=1 ./internal/closuregraph` | 0 | race suite passed; `11.875s` |
| `go test -count=1 -cover ./internal/closuregraph` | 0 | `80.6%` of statements |
| `go test -shuffle=on -count=10 ./internal/closuregraph` | 0 | ten shuffled repetitions passed; `9.973s` |
| `go vet ./internal/closuregraph` | 0 | no findings |
| `go test -c -o /tmp/TASK-260811-i3154q-closuregraph.test ./internal/closuregraph` | 0 | package and tests compile |
| `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run ./internal/closuregraph/...` | 0 | exact CI-pinned linter reported `0 issues.` |
| `gofmt -l internal/closuregraph` | 0 | no files listed |
| `git diff --check` | 0 | no tracked whitespace errors |
| `go test -count=1 ./...` | 0 | every repository package passed on the serialized stable-source snapshot; `cmd/curator` `378.779s`, `internal/artifactpolicy` `129.950s`, `internal/closuregraph` `4.086s`, `internal/install` `118.883s`, `internal/install/atomicity` `117.856s`, `internal/godriver` `75.296s`, and `internal/transaction` `81.381s` |
| `task-board validate` | 0 | `Board is valid. No issues found.` before the evidence update |

Expected-red/recoverable evidence is retained truthfully:

- The original six-probe reviewer overlay exited 1 before rework with all six
  gaps reproduced. After rework it still exits 1 because three probes use a
  helper that calls `t.Fatal` when malformed graphs are rejected during active
  projection, earlier than the probe's later assertion; the codec, foreign-plan,
  and target-read probes pass. The equivalent permanent negative tests invoke
  the error-returning constructor and all exit 0 without weakening the earlier
  fail-closed boundary.
- The first current focused run exited 1 because one new missing-sides test
  expected both collected diagnostics while the canonical collector returns
  one stable primary. It was split into independent missing-provider and
  missing-consumer cases; the rerun exits 0.
- One intermediate compile run exited 1 after decoder cleanup used an
  undeclared local `err`; the declaration was corrected and the full focused
  rerun exits 0.
- A later focused run exited 1 because the new every-mode fixture also changed
  the declared provider language in the deliberate mismatch case. The fixture
  now models actual and declared language independently; both the mismatch
  rejection and all seven positive mode projections pass.
- Direct `golangci-lint` exited 127 because the binary is not installed. The
  exact repository-pinned v2.12.2 invocation then exited 1 on one
  `ineffassign`; that assignment was removed and the exact rerun exits 0 with
  zero issues.

## Earlier reviewer finding closure retained for traceability

### 1. C0 and C0-C7 cross-record trust

- Added `CheckpointChainEvidence` and `CheckpointRecordEvidence` in
  `checkpoint_evidence.go`. The public chain validator now resolves the exact
  C0 selection/platform/tool authority; C1 declarations, lock, journal,
  candidate graph, conditions, and evaluator identities; C2 intake/origin/
  protected-handle/broker aggregates; main or Cargo C3a/C3b admission and
  derivation aggregates; C4 capture/selection/binding/active records; C5 plan
  and closure; C6 observations/execution/order/write set; and C7 publication
  and independently derived expected cache input.
- Added the domain-separated `ToolchainSelector` authority record. A C0-bound
  tool must resolve to the exact C0 checkpoint and its evidence-tool table. A
  C4 build-only tool must resolve to exactly one selector whose node ID,
  fingerprint, executable path, and policy match the complete binding node.
  Arbitrary evidence digests and wrong-kind platform/tool records reject.
- Added positive ordinary and Cargo chains plus negative drift cases for C0,
  every C1-C3 aggregate, C4, C5/closure/cache input, C6 observations/execution,
  and C7 publication in `checkpoint_evidence_test.go` and
  `validation_test.go`.

### 2. Closed action references and endpoint contracts

- Added a closed `$TOOL(name)`, `$READ(name)`, and `$WRITE(name)` parser in
  `action_template.go`. Slot names use a closed grammar; unsupported or
  unterminated placeholders reject; referenced and declared slot sets must be
  exactly equal before graph projection. `argv_template[0]` must itself be
  exactly one declared `$TOOL(slot)` placeholder, so a raw ambient executable
  cannot be hidden before a decoy tool reference.
- Structural graph validation rejects undeclared slot edges even when a
  capture branch is pruned. Selected actions still require exactly one edge
  for every declared read, write, and tool slot.
- Endpoint reconciliation now binds tool executable paths to the exact
  toolchain/local-tool node, source reads to the admitted projection, produced
  paths and optional classes to immutable generated/output declarations, and
  publication destinations to expected output paths. C6 observation
  validation also rejects nonempty produces-class drift.
- `action_contract_test.go` covers hidden/missing/unknown/unterminated
  placeholders and tool/read/produce/publish path or class drift.

### 3. Complete deterministic conditions

- Nonempty `TargetUnitPayload.ConditionExpressions` now fails closed with an
  instruction to use typed conditional capture edges. No hashed target
  condition can remain accepted but unevaluated.
- Conditional edge IDs are collected and sorted before registry lookup or the
  first evaluator call. Evaluation call order and primary failures are now
  independent of Go map order or adapter emission order.
- `condition_projection_test.go` covers target-condition rejection, exact
  canonical evaluator call order across permutations, and stable failing
  primaries.

### 4. Selected dependency ordering and causal cycle scope

- Only runtime and peer requirements are categorically non-ordering.
  Development, optional, workspace, build, tool, toolchain, and package
  relations derive provider-before-consumer arcs when their selected provider
  actually declares materialization actions.
- Action ownership is derived transitively through product/package/target
  declaration structure, preserving exact source-edge evidence and stable
  ordering identities.
- Cycle affected scope follows causal ownership, requirement, production,
  read/local-tool, interop, and publication flow. Target-platform and external
  toolchain hubs cannot widen the affected product/target set.
- `plan_test.go` covers development/optional/workspace target providers,
  optional package materialization, stable waves, and exclusion of an
  unrelated target sharing a platform with a build cycle.

## Prior validation evidence (superseded package manifest)

The commands below remain historical evidence for the earlier
`e54221d1...a6cfe2` package snapshot. They are not claimed as current-code
gates after the RUN-260811-273b9c rework; the current authoritative gates are
listed above.

Each gate below was run directly as a standalone process. Nonzero commands are
reported as failures, including expected-red development iterations.

| Command | Exit | Evidence |
| --- | ---: | --- |
| `go test -count=1 ./internal/closuregraph` | 0 | focused suite passed; `0.907s` package time on the latest run |
| `go test -count=1 -race ./internal/closuregraph` | 0 | race suite passed; `7.723s` |
| `go test -count=1 -cover ./internal/closuregraph` | 0 | `80.7%` statements; `0.947s` |
| `go test -count=10 ./internal/closuregraph` | 0 | ten repeated runs passed |
| `go test -shuffle=on -count=10 ./internal/closuregraph` | 0 | latest shuffled/repeated suite passed; `5.938s` |
| `/Users/iv/go/bin/golangci-lint run ./internal/closuregraph/...` | 0 | `0 issues.` |
| `go vet ./internal/closuregraph` | 0 | no findings |
| `go build ./internal/closuregraph` | 0 | package compiles |
| `gofmt -d internal/closuregraph` | 0 | no diff |
| `shasum -a 256 internal/closuregraph/testdata/canonical-goldens.txt` | 0 | exact accepted digest `fed9657b...9cadcb` |
| `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb .temp/resources/TASK-260811-i3154q/TASK-260810-1uu9lk_accepted-cross-language-closure-graph-and-checkpoints.md` | 0 | `canonical_goldens=pass labeled_records=53`; `canonical_references=pass`; capture reused and all references resolve |
| `git diff --check` | 0 | no tracked whitespace errors |
| `go test -short -count=1 ./...` | 0 | every repository package passed; `cmd/curator` `474.215s`, `internal/closuregraph` `1.276s`; the short mode skips the sibling 100,001-entry normative fixture |
| `go test -count=1 ./internal/artifactpolicy -run '^TestDirectoryEntryLimitStopsTheLiveWalker$'` | 0 | the separately owned repair stabilized; exact previously red vector passed in `14.289s` |
| `go test -count=1 ./...` | 0 | provisional unshortened repository snapshot passed every package; `cmd/curator` `373.757s`, `internal/artifactpolicy` `132.084s`, `internal/closuregraph` `1.310s`, install `116.430s`, atomicity `118.514s` |
| `go test -count=1 ./...` after `TASK-260811-2gazym` reached `to-review` | 0 | directive-mandated stable-snapshot gate passed every package; `cmd/curator` `359.361s`, `internal/artifactpolicy` `126.901s`, `internal/closuregraph` `4.004s`, install `118.884s`, atomicity `119.244s` |
| `task-board validate` | 0 | `Board is valid. No issues found.` at the latest evidence checkpoint |

At that historical checkpoint, the package manifest digest was rederived after
the green repository gate and remained `e54221d1...a6cfe2`.

The later `nudge:fa36b6` directive made the first green full-suite snapshot
provisional while the separately owned admission package was changing.
`TASK-260811-2gazym` subsequently reached `to-review`; a fresh exact
unshortened repository rerun then exited 0 as recorded above. No
`internal/artifactpolicy` file changed during that stable run. The sibling was
assigned to an independent reviewer only after its developer handoff; this
stable-snapshot rerun is the authoritative repository gate for this task.

### Expected-red and recoverable iterations

- A focused test run exited 1 after endpoint validation became structural:
  the old duplicate-tool-slot fixture intentionally changed the executable
  path and therefore hit the stronger path mismatch first. The fixture now
  varies invocation role while retaining the exact executable path, and the
  final focused suite exits 0.
- A later focused test run exited 1 after the command-position rule was added:
  the old absent-slot vector removed `argv_template[0]` and therefore correctly
  hit the new hidden-command diagnostic first. It now omits a declared read
  slot instead, independently exercising both failures; the rerun exits 0.
- An intermediate scoped lint run exited 1 on an ineffectual assignment and an
  obsolete helper left by the causal cycle-scope replacement. Both were
  removed; the exact lint rerun exits 0 with zero issues.
- The first Ruby command exited 1 because the attachment directory contains
  the accepted Markdown but not the verifier script. The authoritative
  `.research` verifier was then invoked against that read-only attachment and
  exited 0 with all 53 records passing.
- The first unshortened repository run exited 1 although every other package,
  including `internal/closuregraph`, passed. The independently owned
  `internal/artifactpolicy` live-walker returned
  `artifact_archive_unsafe_path` instead of its expected entry-limit
  diagnostic. The exact isolated test reproduced exit 1 three times. After
  that task's active implementer repaired its diagnostic precedence, the exact
  isolated command exited 0 and the full unshortened repository command above
  exited 0. The earlier failures remain recorded and are not presented as
  passes.

## Architecture and ownership boundaries

1. Capture remains selection-neutral. Requested platforms, concrete target
   edges, and selected external tools remain exclusively in distinct binding
   overlays.
2. Expected output declarations remain immutable. Observed C6 bytes can only
   change observation, execution, and publication identities.
3. This package validates record identities and causal joins but does not run
   pre-C5 derivations, execute sandboxed processes, inspect artifact bytes, or
   publish protected storage; those are explicit downstream ownership
   boundaries.
4. Kotlin, Dart, .NET, verified binaries, and ecosystem-specific adapter
   implementations remain excluded exactly as required by the accepted scope.
