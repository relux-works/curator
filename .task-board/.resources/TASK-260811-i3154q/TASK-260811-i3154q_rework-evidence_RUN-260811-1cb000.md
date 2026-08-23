# Developer rework evidence for TASK-260811-i3154q

Status: developer handoff evidence; repository-wide validation is source-stable

Run: `RUN-260811-1cb000`

Authoritative run goal at the latest directive checkpoint: `GOAL-260811-eab53e`
revision 1, resolved scope `TASK-260811-i3154q`.

Reviewed changes-requested input:
`TASK-260811-i3154q_review-verdict_RUN-260811-774fc7.md` and its two
independent overlay probes.

## Rework delivered

1. Every selected `targets` edge now has its binding role normalized against
   the current selection and checked against the source node's closed declared
   platform-role set. Product, target, action, external toolchain, expected
   output, and interop-boundary records reject undeclared extra target/host
   bindings. Counting uses the same normalization, preserving the documented
   host-to-target fallback when no distinct host platform is selected.
2. Duplicate semantic edge identity now excludes provenance-only
   `EvidenceOrigin` while retaining kind, endpoints, and every
   relationship-defining payload field. Duplicate diagnostics canonically
   include every conflicting table, edge key, distinct edge ID, origin field,
   and optional origin manifest digest.
3. Permanent regressions cover undeclared host bindings on all six
   platform-bearing node kinds; undeclared target bindings on the four kinds
   that can validly be host-only (target, action, toolchain, and host-extension
   boundary); both host fallback encodings; all six origin-bearing edge
   payloads (`declares`, `resolves_to`, `requires`, `targets`,
   `provides_interop`, and `consumes_interop`); relationship-field
   preservation; and reversed node/edge permutations.

The exact CGP05/CGP10 corpus was not edited. Kotlin, byte detectors, artifact
classification, and sandbox implementation remain outside this task.

## Validation evidence at this checkpoint

Each command ran directly and returned the stated process status.

| Command | Exit | Evidence |
| --- | ---: | --- |
| Reviewer overlay before fix | 1 | Expected red: both invalid graphs were accepted, exactly reproducing the verdict. |
| Initial permanent-selector compile attempt | 1 | Test-harness error: the test initially referenced the not-yet-added production origin extractor; it was replaced with a test-local helper before the behavior-level red run. |
| New permanent selector before fix | 1 | Expected red: all six undeclared-host cases, host fallback, and all cross-origin payload cases exposed the old behavior. |
| New permanent selector after fix | 0 | All platform-role and semantic-duplicate regressions, including the added host-only/extra-target cases, passed in 0.762s. |
| Original reviewer overlay after fix | 0 | Both independently supplied probes passed unchanged. |
| `go test -count=1 ./internal/closuregraph` | 0 | Full focused package passed in 10.086s. |
| Exact golden/Go/cycle/checkpoint selector | 0 | Accepted CGP05/CGP10, Go compatibility, checkpoint chain, SCC, plan permutation, and cycle checks passed. |
| `go test -race -count=1 ./internal/closuregraph` | 0 | Race suite passed in 109.459s. |
| `go test -count=1 -cover ./internal/closuregraph` | 0 | 82.1% statement coverage in 10.043s. |
| `go test -shuffle=on -count=10 ./internal/closuregraph` | 0 | Ten shuffled repetitions passed in 98.321s. |
| `go vet ./internal/closuregraph` | 0 | No findings. |
| `go build ./internal/closuregraph` | 0 | Package compiled. |
| `gofmt -l internal/closuregraph` | 0 | No files listed. |
| pinned `golangci-lint` v2.12.2 on `./internal/closuregraph/...` | 0 | `0 issues.` |
| accepted Ruby canonical verifier | 0 | 53 labeled records, two CGP05 target branches, two CGP10 observations, and every reference passed. |
| canonical corpus SHA-256 | 0 | `fed9657b33297650e64178d605c7ee7f47dde640fefd12a99dec3651bb9cadcb`. |
| `go test -count=1 ./...` | 0 | Every repository package passed uncached; `cmd/curator` took 382.331s, `internal/artifactpolicy` 154.109s, `internal/install/atomicity` 137.870s, `internal/install` 129.441s, and `internal/closuregraph` 14.625s. |
| `go vet ./...` | 0 | No findings. |
| pinned `golangci-lint` v2.12.2 on `./...` | 0 | Reported `0 issues.` It also emitted one non-failing generated-file-filter warning for a stale `/private/tmp/.../internal/buildmeta/buildmeta_test.go` path that no longer existed. |
| `go build ./...` | 0 | Every repository package compiled. |
| all-Go-source pre/post fingerprint | 0 | Both snapshots contained 355 files and were byte-identical at `sha256:152935e2a15928239815c36851b597fb37c4d284cb878900ab17777b4bc72423`. |
| `git diff --check` | 0 | No whitespace errors in tracked changes. |
| closuregraph UTF-8/newline/trailing-whitespace validator | 0 | `closuregraph_text=pass files=27`. |

The current 27-file `internal/closuregraph` sorted per-file SHA-256 manifest is
`7491158c1521b20abfe19464c306c1603b4ca26c90e748462afdb348cb7b0c88`.

Directive `nudge:410209` explicitly superseded the temporary withdrawal in
`nudge:daaedf` after sibling `TASK-260811-2gazym` reported a source-stable
checkpoint. Before the full lane, the process audit found no Go validation
process; after test, vet, lint, and build, the repository fingerprint was
unchanged. No implementation blocker or forced-fit condition exists. This
evidence is ready for the required independent review.
