# Developer rework evidence for TASK-260811-i3154q

Status: developer handoff evidence; stable repository-wide gate passed

Run: `RUN-260811-07310a`

Authoritative run goal at the latest directive checkpoint:
`GOAL-260811-eb0974` revision 1, resolved scope `TASK-260811-i3154q`.

Reviewed changes-requested input:
`TASK-260811-i3154q_review-verdict_RUN-260811-cc0030.md` and its five
independent reviewer probes.

## Rework delivered

1. Owner-level `requires` edges to a produced output now resolve the exact
   `produces` lineage. Product, package, target, and interop-boundary consumers
   receive canonical provider-before-consumer evidence. Permutations preserve
   plan identity, while reciprocal owner/output dependencies return the same
   structured `closure_build_cycle` evidence.
2. Interop tool authority counts exactly one selected, binding-owned
   `requires(scope=toolchain)` edge whose endpoint is an authorized binding
   `toolchain_component`. Capture-table substitutes, wrong scopes, wrong kinds,
   missing records, duplicate records, and platform mismatches fail closed.
3. Build-plan derivation rejects duplicate selected output/write paths,
   including expected-output/expected-output and generated/output conflicts,
   before returning a `BuildPlan` or issuing C5. The structured diagnostic and
   its edge/output evidence are permutation-stable.
4. Every `resolves_to.artifact_manifest_id` must resolve in the capture
   manifest authority and match either the package manifest or the exact
   transformed source-set manifest. This admits explicitly separate raw and
   transformed manifests while rejecting absent or unrelated captured IDs.
5. Intrinsic validation and decoder failures now use stable field/key order.
   The fix covers node, edge, selection, active-graph, checkpoint, diagnostic,
   observation, exact-field, string-map, platform-role decoding, and
   condition-evaluator registry paths. Evaluator registrations are sorted
   before nil, invalid-ID, duplicate-ID, or unselected-ID rejection, so adapter
   emission order cannot change the primary failure.

The exact CGP05/CGP10 corpus was not edited. Kotlin, byte detectors, artifact
classification, and sandbox implementation remain outside this task.

## Permanent test evidence

`internal/closuregraph/reviewer_rework_test.go` permanently covers:

- all five reproduced reviewer failures;
- product/package/target/boundary output ordering and exact source evidence;
- output-requirement permutation and cycle behavior;
- wrong-table, wrong-kind, wrong-scope, missing, and duplicate interop tool
  bindings (with the pre-existing regression suite supplying the missing and
  wrong-scope cases);
- duplicate expected-output and generated/output write paths before C5;
- absent, unrelated-captured, package, and transformed-source manifest cases;
- repeated multiply-invalid validation across closed record families; and
- canonical decoder plus nil/invalid/duplicate/unselected evaluator-registry
  failures under input permutations.

The initial reviewer selector was run before the fix and truthfully exited 1:
all five reviewer probes failed exactly as reported. The same permanent
selector exits 0 on the current source.

## Current source-stable validation

Every command below ran directly and returned its reported process status.

| Command | Exit | Evidence |
| --- | ---: | --- |
| `go test -count=1 ./internal/closuregraph` | 0 | focused package passed in 12.450s |
| `go test -count=1 ./internal/closuregraph -run '^TestReviewerRework'` | 0 | expanded adversarial rework selector passed in 11.210s |
| evaluator-registry canonical-failure selector | 0 | targeted test passed in 6.916s |
| exact CGP05/CGP10/Go compatibility selector | 0 | exact golden set passed in 2.871s |
| `go test -race -count=1 ./internal/closuregraph` | 0 | race suite passed in 108.662s |
| `go test -count=1 -cover ./internal/closuregraph` | 0 | 81.9% statement coverage in 11.716s |
| `go test -shuffle=on -count=10 ./internal/closuregraph` | 0 | ten shuffled repetitions passed in 98.159s |
| `go vet ./internal/closuregraph` | 0 | no findings |
| `go build ./internal/closuregraph` | 0 | package compiled |
| `gofmt -l internal/closuregraph` | 0 | no files listed |
| pinned `golangci-lint` v2.12.2 on `./internal/closuregraph/...` | 0 | `0 issues.` |
| accepted Ruby canonical verifier | 0 | 53 labels, both CGP05 branches, both CGP10 observation branches, and all references passed |
| `shasum -a 256 internal/closuregraph/testdata/canonical-goldens.txt` | 0 | `fed9657b33297650e64178d605c7ee7f47dde640fefd12a99dec3651bb9cadcb` |
| `git diff --check` | 0 | no tracked-diff whitespace findings |
| closuregraph terminal-newline/trailing-whitespace validator | 0 | all 26 package files passed |
| repository Go-source fingerprint before/after full suite | 0 / 0 | identical `e2faf77a77a5969277e3d4b0c346556ce4bec08b4ab65695dc8f5911b3516a4e` over 339 files |
| `go test -count=1 ./...` | 0 | every repository package passed; `cmd/curator` 359.773s, `artifactpolicy` 132.329s, `closuregraph` 14.492s, `install` 116.268s, `install/atomicity` 116.457s, `transaction` 80.701s, `godriver` 67.128s |
| `task-board validate` | 0 | board valid with no issues after the full-suite evidence refresh |

The sorted per-file SHA-256 manifest for the current 26-file
`internal/closuregraph` package hashes to
`bba7611d282747ae0a9d6dc77e9eb26e67db2efa3f5ed6a66296ac3a4594b725`.

## Stable validation coordination

Directive `nudge:62dbe5` accepted this source-stable producer checkpoint and
held all further source and validation work while sibling `TASK-260811-2gazym`
used the exclusive repository runner. Directive `nudge:32f7b5` then cleared the
barrier after that task reached stable `to-review` and authorized this task's
second exclusive run. The direct uncached `go test -count=1 ./...` process
exited 0, and the independently computed 339-file Go-source fingerprint was
byte-identical before and after it. No implementation blocker or forced-fit
condition exists.
