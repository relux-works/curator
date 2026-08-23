# Reviewer verdict for TASK-260811-2gazym

Verdict: **changes requested -> to-dev**

This artifact records exactly one reviewer branch for RUN-260811-068fd0. It does not accept delivery and supplies no `commit_ack`.

## Goal and reviewed scope

- Reviewer run: `RUN-260811-068fd0`
- Authoritative goal immediately before verdict: `GOAL-260811-60cec0` revision 1
- Resolved scope: `TASK-260811-2gazym`
- Reviewed producer evidence: `TASK-260811-2gazym_implementation-evidence.md`
- Reviewed producer-evidence SHA-256: `585604ad10cd41951709d49f485b0434f151a54dd756acef7152a75efe6eba1c`
- F14 option-1 amendment was respected: exact raw/full manifest identities differ while canonical logical evidence must remain order-independent.

## Acceptance blockers

### R1 — rejected-node manifest semantics remain forgeable

`DecodeManifest` does not canonically derive a rejected node classification from its detector evidence. In `internal/artifactpolicy/codec.go:1052-1073`, `validateNodeSemantics` returns success immediately for every `REJECT` node. The remaining diagnostic checks only copy-match node fields and require the selected detector ID to appear; they do not bind ELF `e_type`, detector result, variant, class, or rule to one closed semantic outcome (`codec.go:1164-1244`).

A self-consistently rehashed rejected ELF record can therefore change both node and diagnostic class from `native.executable` to `native.object`, retain contradictory `ET_EXEC` facts and the old variant/rule, recompute the findings and manifest digests, and pass the current decoder. This fails canonical `artifact-manifest-v1` evidence and the prior rework requirement to reject semantically impossible rejecting records.

Required rework: validate all decisions, including rejection, against a closed class/variant/rule/detector-result/fact contract. Add rehashed negative tests that alter class while retaining contradictory ELF, Mach-O, PE/COFF, VM/IR, container, and text observations.

### R2 — truncated findings summaries can be rehashed with invented totals/digests

`internal/artifactpolicy/codec.go:318-342` recomputes the findings digest only when `Total == Recorded`. It does not require `Recorded == max_recorded_findings` and `Total > max_recorded_findings` in the truncated branch. A manifest with two findings, one recorded finding under the v1 cap of 1,000, and an arbitrary well-formed SHA-256 can be rehashed and accepted even though runtime generation cannot produce that summary.

Required rework: enforce the only two canonical states: an exact fully recorded set, or a cap-saturated recorded prefix with a complete-set digest produced by a verifiable canonical representation. Add self-consistent rehash negatives for premature truncation, invented total, arbitrary complete-set digest, and rejecting nodes omitted behind forged truncation.

### R3 — ordinary native archives with symbol/name metadata cannot seal a manifest

`internal/artifactpolicy/containers.go:566` charges every `ar` member header, but `containers.go:609-615` silently omits GNU/BSD symbol-table and GNU string-table members (`/`, `/SYM64/`, and `//`) from the node set. `bindTraversalAccounting` then reports unmanifested entries. A valid compiled dependency has only a compiled diagnostic, and a valid protected local static-library output has no diagnostic; neither supplies the structural cause required by `traversalFailureEvidence` (`codec.go:80-100`). Manifest sealing therefore fails instead of returning the required compiled deny or `ALLOW_OUTPUT`.

The corpus helper constructs only metadata-free synthetic archives, so C04 and A08 do not cover the ordinary archive shape emitted by `ar`/ranlib.

Required rework: canonically manifest and validate archive metadata members, or account for them with an exact schema-bound structural representation that does not look like failed traversal. Add real GNU/BSD archive fixtures with symbol tables, long-name tables, and import-library metadata for dependency, toolchain, and output roles.

### R4 — duplicate members with different bytes are physical-order dependent

`prepareMembers` stable-sorts only by member name (`internal/artifactpolicy/containers.go:821-824`) and retains the first explicit duplicate while dropping later duplicates after recording the unsafe-path diagnostic (`containers.go:840-847`). For two archives containing the same duplicate path once as source and once as ELF, source-first and ELF-first order changes which bytes are hashed/classified and whether a compiled diagnostic exists. This changes canonical logical evidence and can change the primary diagnostic, contrary to the deterministic ordering policy and F14 logical-projection contract.

The current F02 duplicate fixture uses identical bytes, while F14 permutes distinct paths; neither exercises this case.

Required rework: inspect and bind every duplicate member without first-entry-wins semantics, then select diagnostics using canonical path/rule/class ordering. Add ZIP, tar, and native-archive permutations with conflicting duplicate bytes and assert identical logical classifications, primary code, finding identity, and no authorization.

### R5 — declared link/load use does not make benign bytes ambiguous

`Descriptor.ResolvedUses` reaches only the ELF detector (`internal/artifactpolicy/detect.go:118-129`). If no compiled detector matches, the service classifies text and applies ambiguity only from filename suffixes (`detect.go:183-193`). A neutral-path source/text file with a manager-resolved `link_or_load` edge is therefore admitted even though the accepted decision procedure requires a deny-indicating load/link reference with benign bytes to reject as `artifact_type_ambiguous`.

Required rework: apply declared-use ambiguity after all byte interpretations, while preserving valid interpreted-script handling. Add neutral-name and misleading-name link/load cases to shared conformance.

### R6 — gzip trailing bytes inflate the per-stream expansion budget

`walkGZIP` computes its per-stream budget from the entire payload size (`internal/artifactpolicy/containers.go:454`) before discovering trailing data or a second stream at `containers.go:495-501`. A tiny valid gzip stream followed by large invalid padding can therefore credit the padding as compressed input and cause substantially more decompression/spooling than the first applicable 200:1 limit permits. The existing first-limit test covers a clean envelope, not this padded invalid shape.

Required rework: meter actual compressed bytes consumed by the first stream and enforce the ratio incrementally; trailing bytes must not enlarge its budget. Add padded and concatenated high-ratio cases that prove bounded work before `artifact_archive_invalid`.

## Validation evidence

The existing suite is green but does not cover the findings above:

- `go test -count=1 ./internal/artifactpolicy/...`: exit 0; artifactpolicy 16.079s.
- `go test -race -count=1 ./internal/artifactpolicy/...`: exit 0; artifactpolicy 40.780s.
- `go test -count=1 ./...`: exit 0; `cmd/curator` 357.172s, artifactpolicy 116.569s, Go admission regressions green.
- `go vet ./...`: exit 0.
- `go build ./...`: exit 0.
- pinned `golangci-lint` 2.12.2 focused and full runs: exit 0, 0 issues. The full run repeated the known stale temporary-file processor warning only.
- `gofmt -l internal/artifactpolicy`: no output.
- `git diff --check`: exit 0.
- `task-board validate`: `Board is valid. No issues found.`

## Routing

These are ordinary, in-scope implementation and adversarial-test fixes. No external blocker, human-only decision, or Stop-The-Line condition exists. Route to `to-dev`, repair R1-R6, rerun focused/race/full/lint/vet/build/board gates, refresh task-scoped implementation evidence, and hand off to a new independent reviewer cycle.

No product code was modified by this reviewer. Kotlin remains excluded and `verified-binary-v1` remains unavailable.
