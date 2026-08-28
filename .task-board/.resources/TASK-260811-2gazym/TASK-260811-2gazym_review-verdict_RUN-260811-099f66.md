# Reviewer verdict for TASK-260811-2gazym

Verdict: **changes requested -> to-dev**

## Goal and reviewed evidence

- Reviewer run: `RUN-260811-099f66`
- Authoritative reviewer goal immediately before verdict: `GOAL-260811-36e761` revision 1
- Resolved scope: `TASK-260811-2gazym`
- Required success predicate: record exactly one evidence-backed reviewer branch for this task
- Directive checkpoint: no directives recorded
- Reviewed producer evidence: `TASK-260811-2gazym_implementation-evidence.md`
- Producer-evidence SHA-256: `67a1c26f95cced79ece3f11ce8c808a5d9f1ae589037987445befdaecf61b1ae`
- Accepted F14 amendment SHA-256: `2f8a6eb4bbdc138745a865ffd804c3b47a6e688b2e6083b6a17ab31863e262a8`
- Reviewed implementation scope: `internal/artifactpolicy`, 25 Go files and 10,300 lines at the review snapshot

The prior rework materially improved causal role authorization, early entry counting, traversal evidence, manifest fields, and the reusable corpus. The task is not yet acceptable because the following remaining defects violate the accepted fail-closed policy and explicit rework requirements. They are ordinary implementation rework; no Stop-The-Line boundary or human-only decision exists.

## Required changes

### R1 — High: artifact-manifest-v1 accepts rehashed, semantically impossible evidence

The decoder verifies canonical syntax and the self-declared manifest digest, but its semantic checks do not bind several required audit facts:

- For a file payload, `codec.go:319-330` binds the root path to `RawPayload.Path` but never requires the root node size and SHA-256 to equal `RawPayload.Size` and `RawPayload.SHA256`. A record can therefore change the classified root hash, recompute the manifest digest, and still decode.
- `codec.go:415-423` requires a digest for byte-bearing nodes only when their decision is not `REJECT`. The accepted manifest requires size/hash evidence for every byte-bearing node, including the node that caused rejection.
- `codec.go:435-455` validates observation identifiers, result enums, and ordering only. It does not prevent a detector `ERROR` observation from appearing on an inspection-complete admitted node.
- `codec.go:319-340` proves only lexical parent containment and the prior existence of names in `ContainerChain`; it does not require an exact chain derived from the parent graph, a container-capable parent, or exactly one child for a compressed stream.
- `codec.go:503-548` accepts `Accounting.EntryCount` values larger than the manifested minimum and leaves other rejecting-manifest accounting only partially tied to recorded traversal evidence.
- `codec.go:352-375,784-807` checks diagnostic enums and a few role/class combinations, but does not bind a diagnostic path, size, SHA-256, container chain, detector evidence, collision key, or required code-specific detail fields to the node/failure it claims to describe.

Fix the semantic validator and add negative codec tests that mutate each relation, recompute canonical bytes/digest, and prove `DecodeManifest` rejects the forged record. Rejected traversals may legitimately count invalid entries that cannot become nodes; represent and validate that evidence explicitly rather than accepting unconstrained counters.

### R2 — High: gzip expansion limits are enforced after the expensive work

`containers.go:434-477` decompresses a gzip stream through `appendUnknown(..., MaxTotalEmittedBytes)`. `blob.go:104-124` can therefore spool up to 2 GiB plus one byte. Only after decompression does `limits.go:95-133` enforce remaining aggregate bytes and the 200:1 ratio, and the 256 MiB single-leaf limit is checked later when the decoded child is inspected.

A small gzip bomb, or a nested gzip encountered after most of the aggregate budget is consumed, can force far more I/O and storage than the first applicable closed limit permits before rejection. This defeats the resource-exhaustion boundary.

Bound streaming gzip output incrementally by the minimum of the single-leaf budget, remaining aggregate budget, and the exact ratio budget, with checked arithmetic. Add tests for a high-ratio gzip stream and for a nested stream with a nearly exhausted aggregate budget; assert the exact limit diagnostic and that bytes beyond the limit are never consumed or spooled.

### R3 — High: mixed-slice Mach-O classification depends on archive slice order

`native.go:563-624` selects the class of a fat/universal Mach-O using `compiledClassPriority`. `detect.go:277-290` assigns executable, object, static-library, and dynamic-library classes the same priority. A universal containing different valid slice classes therefore inherits whichever equal-priority slice appears first.

The accepted taxonomy requires a fat/universal artifact containing an executable slice to classify as `native.executable`, and classification must be deterministic rather than physical-order dependent. Existing C03 tests at `detectors_test.go:159-181` construct only one-slice fat fixtures and cannot expose the defect.

Define explicit deny-dominant mixed-slice resolution, retain every slice observation, and add reordered multi-slice fixtures proving an identical class/decision independent of slice-table order.

### R4 — High: printable serialized-compiler claims can bypass ambiguity rejection

`native.go:1040-1055` deliberately returns no compiler-serialized match when bytes claiming `.swiftdoc`, `.gch`, or `.ifc` are fully printable. `detect.go:309-334` omits those suffixes from the deny-indicating suffix table. Such content can consequently be accepted through a caller declaration, and names such as `README.swiftdoc` can be inferred as metadata.

The accepted taxonomy expressly places Swift `.swiftdoc`, Clang PCH forms, C++ BMI forms, and equivalent serialized compiler artifacts in the deny class. Benign text that claims one of those roles must fail as `artifact_type_ambiguous`, not become admitted text. Complete the deny-indicating role/suffix coverage and add positive binary detections plus printable-conflict negatives for every supported claimed form.

### R5 — High: reusable accepted vector corpus remains representative rather than complete

The published corpus states at `conformance/corpus.go:90-92` that it returns one reusable golden per label and leaves additional compound-label branches in package-local tests. That does not satisfy the rework requirement that the accepted A/C/F/T/V byte branches be reusable with exact pinned canonical results/digests.

Concrete omissions include:

- A08 publishes only the object branch, while object/library/addon/executable branches live only in `policy_test.go:124-150`.
- C03 publishes one single-slice fat executable; the other thin/fat branches are local, and no mixed-slice fixture exists.
- C04 publishes one archive/depth combination rather than `.a`, `.lib`, and `.rlib` at depths 1, 2, and 8.
- F01 and F08 publish one representative each rather than all named path and resource-limit branches.
- T03 and T05 publish one representative each rather than all escape/special/mutation and path/size/digest/full-input drift branches.
- C12 local coverage at `detectors_test.go:363-379` compares only class, SHA-256, size, and decision; it does not pin the required complete shared leaf evidence and exact canonical manifest outcome for all five adapter harnesses.

Expand the reusable corpus to every accepted branch, include exact expected canonical record bytes or an equivalent exact record oracle plus pinned manifest digests, and run each through the public pre-execution API. Every negative must also assert no admission authorization or cache-publication authority was emitted.

## Independent verification

All executed quality gates passed:

- `go test -count=1 ./internal/artifactpolicy/...` — pass; policy package 228.550s
- `go test -count=1 -run TestReusableArtifactManifestV1ConformanceCorpus -v ./internal/artifactpolicy` — pass; all 47 currently published cases
- `go test -short -count=1 -race -cover ./internal/artifactpolicy/...` — pass; policy coverage 74.8%
- `go test -count=1 ./...` — pass; `cmd/curator` 586.676s, `internal/artifactpolicy` 235.900s, existing Go admission packages green
- `go build ./...` — pass
- `go vet ./...` and `go vet ./internal/artifactpolicy/...` — pass
- pinned Go 1.25.5 `golangci-lint run ./internal/artifactpolicy/...` — 0 issues
- `gofmt -l internal/artifactpolicy` — clean
- `git diff --check` — pass
- `task-board validate` — Board is valid

Green existing tests do not override the untested fail-closed counterexamples above. No product code was modified by this reviewer, and no `commit_ack` is supplied.
