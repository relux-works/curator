# Developer rework evidence for TASK-260811-i3154q

Status: developer handoff evidence; repository-wide validation is source-stable

Run: `RUN-260811-3b6b12`

Authoritative run goal at the latest directive checkpoint:
`GOAL-260811-71061d` revision 1, resolved scope
`TASK-260811-i3154q`.

Reviewed changes-requested input:
`TASK-260811-i3154q_review-verdict_RUN-260811-552b3a.md` and unchanged probe
`TASK-260811-i3154q_reviewer-probes_RUN-260811-552b3a.go`, SHA-256
`7dc525a532027ac0908d3bbab01310955bff7d6f8af5a5709c7fe8c335c920cc`.

Directive `nudge:1d636c` limited this run to the verdict's two actionable
findings and reserved repository-wide validation. Directive `nudge:ecb252`
released the serialized lane only after all focused gates were green.

## Rework delivered

1. Platform binding validation now checks each raw
   `TargetsPayload.BindingRole` against the source node's raw closed declared
   role set and counts exact raw role slots. Host-to-target fallback is used
   only by `platformIDForRole` to resolve the destination platform ID for a
   genuinely declared host slot. A target-only declaration with a raw host
   edge therefore fails canonically and cannot create a second accepted
   binding identity, while host-declared actions still bind to the sole target
   platform when no distinct host platform exists.
2. `Node.Validate`, `Edge.Validate`, `Checkpoint.Validate`, and
   `NewCheckpoint` accept only the exact value payload representations emitted
   by their canonical codecs. Non-nil and typed-nil pointers are rejected
   before any interface method or downstream value assertion can panic, with
   stable `closure_graph_schema_unsupported` or
   `closure_checkpoint_invalid` diagnostics.
3. Permanent table-driven regressions cover non-nil and typed-nil forms for
   all ten node payloads, eleven edge payloads, and all checkpoint names C0-C7
   plus C3a/C3b. They exercise validation, canonical bytes, identity,
   constructor behavior, graph-level no-panic rejection, and permutation-
   independent issue ordering. Platform regressions cover the target-only host
   alias, distinct raw host/target slot cardinality on one fallback platform,
   and the existing positive host-declared action fallback.

Production changes are confined to `internal/closuregraph/node.go`,
`edge.go`, `checkpoint.go`, and `validation.go`. Test support preserves raw
host roles in `plan_test.go`; direct validation call sites were updated in
`reviewer_platform_semantic_rework_test.go`; permanent coverage is in
`payload_representation_rework_test.go` and `platform_role_rework_test.go`.

The exact 53-record CGP05/CGP10 corpus was not edited. Kotlin, byte detectors,
artifact classification, and sandbox execution remain outside this task.

## Reviewer-probe handling

The unchanged probe was first materialized byte-for-byte and verified at the
same SHA-256 as the board resource. Before production changes,
`go test -count=1 ./internal/closuregraph -run '^TestReviewerProbe'` exited 1
with exactly the target-role alias plus the node, edge, and checkpoint pointer
panics from the verdict.

After the fix, the same unchanged probe command exited 1 only because its
checkpoint helper calls `NewCheckpoint` and treats any constructor error as a
test failure. The constructor now rejects the noncanonical C0 pointer earlier
with `closure_checkpoint_invalid`; the role, node, and edge probes passed and
no panic remained. The temporary copy was then removed. Permanent exhaustive
tests assert this stronger earliest-boundary rejection and are green. The
authoritative probe resource itself remains byte-identical and is retained as
the required red evidence rather than misreported as a green gate.

## Validation evidence

Every command below ran directly as a standalone process and returned the
stated real exit code.

| Command | Exit | Evidence |
| --- | ---: | --- |
| Unchanged reviewer probe before fix | 1 | Expected red: one accepted host alias and three pointer panics, exactly reproducing the verdict. |
| Unchanged reviewer probe after fix | 1 | Expected historical-helper mismatch: only early `NewCheckpoint` canonical rejection remained; no panic and the other three probes passed. |
| Focused rework selector | 0 | Exhaustive payload, permutation, role-alias, exact-slot, and positive-fallback tests passed in 0.511s. |
| `go test -count=1 ./internal/closuregraph` | 0 | Full focused package passed after the lint correction in 10.217s. |
| `go test -count=1 -cover ./internal/closuregraph` | 0 | 81.9% statement coverage in 10.468s. |
| `go test -race -count=1 ./internal/closuregraph` | 0 | Race suite passed in 109.096s. |
| `go test -shuffle=on -count=10 ./internal/closuregraph` | 0 | Ten shuffled repetitions passed in 97.852s. |
| `go vet ./internal/closuregraph` | 0 | No findings. |
| `go build ./internal/closuregraph` | 0 | Package compiled. |
| `gofmt -l internal/closuregraph` | 0 | No files listed. |
| bare `golangci-lint` version/run attempts | 127 | Executable was absent from `PATH`; these attempts were not counted as lint success. |
| initial pinned v2.12.2 scoped lint | 1 | One local `revive` finding for the now-obsolete `selection` parameter. |
| pinned v2.12.2 lint on `./internal/closuregraph/...` after correction | 0 | `0 issues.` |
| accepted Ruby verifier on the authoritative contract | 0 | 53 labeled records, both CGP05 target branches, both CGP10 observation branches, and every reference passed. |
| accepted Ruby verifier on the implementation corpus | 0 | Same 53-record and reference result. |
| `go test -count=1 ./...` | 0 | Every repository package passed uncached; `cmd/curator` 358.052s, `artifactpolicy` 130.579s, `closuregraph` 15.010s, `install` 114.732s, and `install/atomicity` 115.673s. |
| `go vet ./...` | 0 | No findings. |
| `go build ./...` | 0 | Every repository package compiled. |
| pinned v2.12.2 lint on `./...` | 0 | `0 issues.` One non-failing generated-file-filter warning named a stale missing `/private/tmp/.../internal/transaction/engine_test.go` path. |
| `task-board validate` | 0 | `Board is valid. No issues found.` |

The 29-file sorted `internal/closuregraph` per-file SHA-256 manifest was
identical before and after all repository-wide gates at
`9b98e11915c03e66e0d685ef41d4311e0789ff1a2a8fdbed11dc0b78003f463f`.
The all-Go-source manifest was likewise identical at
`134cfa1ffdf5d29f339a88e7d8b0d5476577d12f1034d560ae184b92647502a7`.
The canonical corpus remains
`fed9657b33297650e64178d605c7ee7f47dde640fefd12a99dec3651bb9cadcb`,
and the accepted contract remains
`874c3a40b9ff9fcf130f37f22e9f5aa2bdd6d9f246403e059c98e84736e27bbc`.

No implementation blocker or forced-fit condition exists. This evidence is
ready for the required independent review.
