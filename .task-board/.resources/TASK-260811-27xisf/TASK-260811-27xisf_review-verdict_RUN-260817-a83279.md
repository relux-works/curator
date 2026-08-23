# Reviewer verdict for TASK-260811-27xisf

Verdict: **changes requested -> to-dev**

## Goal and scope evidence

- Reviewer run: `RUN-260817-a83279`
- Authoritative goal at the final directive checkpoint: `GOAL-260817-b90615` revision 1
- Resolved scope: `TASK-260811-27xisf`
- Review policy: `required`
- Reviewed implementation: `internal/closureexec`
- Producer evidence: `TASK-260811-27xisf_implementation-evidence.md`

This artifact records only the changes-requested branch. The implementation
has useful foundations and its current tests pass, but the acceptance predicate
is not satisfied because the protected execution and reuse boundaries do not
yet prove the required observed behavior.

## Required rework

### R1 — process/read/write/network evidence is declared, not observed

`internal/closureexec/boundary_darwin.go:76` constructs the returned `Audit`
by copying `AllowedProcesses`, `ReadRoots`, `WriteRoots`, `ExpectedEvidence`,
and `network=none` directly from the permit after the child exits. It does not
collect the child's actual process, filesystem, evidence-output, or network
events. Consequently `auditDifference` in `executor.go:111-142` compares the
permit to a copy of itself. A sandbox-denied undeclared operation that the
child handles while still exiting zero is invisible and can receive a success
receipt. This contradicts the accepted requirement for actual audit and
CGN18's extra process/read/write/network/output evidence rejection, and it
does not prove that a network attempt cannot publish outputs.

Rework the Darwin boundary to collect authoritative OS-observed events (or use
another enforce-and-observe boundary), distinguish attempted/observed behavior
from the allowed policy, reconcile the actual sets, and add real-process
negative tests for ignored denied reads, writes, child execs, and network
attempts with zero receipt/publication.

### R2 — stale same-head permits can execute after the causal head advances

`Executor.Commit` checks `PreviousCausalHead` only at commit time
(`executor.go:48-55`). Multiple permits can therefore be committed against the
same head. `Execute` retrieves a committed permit but does not compare its
predecessor with the current head (`executor.go:63-86`). After one permit
succeeds and advances `e.head`, another stale permit can still start and issue
a receipt. This violates the serial canonical derivation journal and exact
predecessor rule.

Recheck the permit predecessor atomically at the process-start seam and define
single-use/failure invalidation for competing permits. Add a regression test
that commits two same-head permits, executes one, and proves the other starts
zero processes and cannot issue a receipt.

### R3 — admitted captures are not exposed as a protected read-only replay tree

`CaptureStore` stores one opaque byte stream and `Executor` only rehashes its
handle. The Darwin boundary resolves all declared non-toolchain reads below the
writable workspace (`boundary_darwin.go:97-109`); there is no capture/source-
snapshot tree mount or materialization API that maps admitted receipt IDs to
read-only paths. Thus the substrate does not implement the required immutable
source-snapshot storage and read-only admitted closure replay. An adapter would
have to copy inputs into ordinary workspace paths outside this receipt model.

Add immutable source-snapshot/tree handles, map only permit-named admitted
inputs into the sandbox as read-only inputs, recheck containment and identity
at time of use, and test missing, mutated, linked, writable, and ambient
substitutions.

### R4 — publication does not validate observations against the C4/C5 graph

`ProtectedStore.Publish` calls only `ProducedArtifactObservation.Validate`
(`store.go:90-123`), which checks shape. It never calls `ValidateAgainst` with
the immutable record tables, and its API receives no C4 selection binding,
active graph/build plan, target, or tool identity evidence. An observation with
the expected output-node ID but an unrelated producer action, produces edge,
path/class declaration, target, or tool can therefore be published if its
self-consistent execution receipt lists the resulting observation ID. The
happy-path CGP10 constants do not prove this negative boundary.

Bind publication to the exact C4/C5/closure evidence and validate every
observation against the unchanged action/output/produces declarations and
target/tool bindings. Add poisoned-reference and wrong-kind/path/class tests
that prove no entry is created.

### R5 — exact-hit inspection does not reconcile stored outputs with the receipt

`ProtectedStore.Inspect` validates the expected-input and execution IDs and
rehashes blobs, but it never compares `raw.Outputs[].ObservationID` with
`publication.PublishedObservationIDs`, never proves the output records are the
same set as the publication, and builds `Paths` as a map without rejecting
duplicate paths (`store.go:179-217`). A modified canonical entry can therefore
return blob paths not covered by the embedded protected publication receipt.
The existing poison test mutates only a blob and does not exercise receipt/
output-set drift.

Recompute and reconcile the complete stored entry, publication observation
set, output paths, sizes, digests, and execution references on every hit; reject
duplicates and any mismatch. Add receipt tampering, output substitution,
duplicate-path, missing-output, and extra-output reuse tests.

### R6 — derivation permit/receipt evidence is incomplete

The accepted record requires an expected evidence schema, deterministic
diagnostics, output paths/manifests/digests, and the next causal head.
`DerivationPermit` carries only an untyped `ExpectedEvidence` path list and one
output-byte limit (`models.go:80-122`); `DerivationReceipt` carries only permit,
fingerprints, copied audit, and decision (`models.go:154-171`). These records
cannot independently prove the permitted evidence type or the bytes admitted
into C1-C4.

Complete the canonical models and validation, bind resource-limit identity and
evidence schema/manifests/digests, and add canonical round-trip/identity/drift
tests for manifest, vendor, mirror, and metadata derivations.

## Independent validation

All commands below were run on the reviewed worktree and exited 0 unless noted:

- `go test -race -count=1 -cover ./internal/closureexec` — pass, **58.7%** statement coverage
- focused compatibility tests for `closureexec`, `closuregraph`,
  `artifactpolicy`, `buildcache`, `godriver`, and `buildsource` — pass
- `go vet ./...`
- `go build ./...`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run` — `0 issues.`
- `go test -count=1 ./...` — pass; slowest packages included
  `cmd/curator` 380.975s, `artifactpolicy` 150.021s,
  `install/atomicity` 121.935s, and `install` 119.099s
- canonical verifier — 53 labeled records and all references pass
- gofmt cleanliness and `git diff --check` — pass

Passing gates do not cover R1-R6. Focused statement coverage is 58.7%, and the
missing branches are the security-negative behaviors required by the task.

No product code was modified, staged, or committed by this reviewer. No
`commit_ack` is supplied.
