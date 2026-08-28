# Reviewer verdict for TASK-260811-2gazym

Verdict: **changes requested -> to-dev**

## Goal and scope evidence

- Reviewer run: `RUN-260811-151b59`
- Authoritative goal immediately before verdict: `GOAL-260811-469ed9`
  revision 1
- Authoritative resolved scope: `TASK-260811-2gazym`
- Review policy: `required`
- Final directive checkpoint: no directives recorded
- Reviewed producer outcome:
  `TASK-260811-2gazym_implementation-evidence.md`
- Reviewed implementation: the 21-file, 7,563-line uncommitted
  `internal/artifactpolicy` package in the current worktree
- Normative inputs were rechecked against the task-scoped copies. The current
  source-closure specification and taxonomy copies are byte-identical; the
  taxonomy SHA-256 is
  `c5334433d6eddf37109e612a97866024a17d38c15cff7d7e5e36dac69fe0df15`.

This artifact records exactly one reviewer branch: `changes_requested`. It does
not record acceptance or a Stop-The-Line boundary. The findings are ordinary,
autonomously repairable implementation and test work.

## Acceptance-blocking findings

### R1 — Critical: trust-role authorization can be minted from self-asserted data

`ToolchainEvidence` and `LocalOutputEvidence` are exported structs containing
only caller-populated strings and booleans (`types.go:317-353`).
`validateToolchainEvidence` checks field presence, digest syntax, booleans, and
fingerprint equality (`policy.go:404-431`), but does not bind the fingerprint or
resolved root to the exact payload, does not require the payload path to equal
the selected executable path, and does not validate contained links or ordinary
nodes itself. `validateLocalOutputCausality` similarly checks only caller claims
and digest-shaped strings (`policy.go:434-455`); it does not authenticate or
decode a protected receipt or prove that the bytes were produced rather than
copied into staging.

Any internal adapter can therefore construct those structs, obtain an
`ALLOW_TOOLCHAIN` or `ALLOW_OUTPUT` token for arbitrary bytes, and pass
`AuthorizeAdapterExecution` or `AuthorizeCachePublication`. Sealing the resulting
`Admission` does not make the input assertions authoritative. This contradicts
the causal trust-role boundary, T03/T04/T05, and the task requirement to enforce
roles before adapter execution or cache publication.

Required rework: accept only manager-owned, non-forgeable checkpoint/receipt
evidence that is bound to the exact payload, virtual path, selected root,
complete fingerprint, action/input identity, and protected publication state.
If the later protected-executor task must mint that evidence, keep the seam
fail-closed until such evidence is supplied; do not authorize from freely
constructible booleans or digest strings. Add negative tests proving fabricated
evidence, path/payload replay, and copied or hard-linked pre-existing bytes
cannot produce authorization tokens.

### R2 — High: `max_entry_count` does not bound hostile enumeration

Container members are fully accumulated and sorted before accounting. In
`prepareMembers`, invalid paths, duplicates, and collisions are discarded from
the `prepared` map, and only `len(prepared)` is charged after both passes
(`containers.go:672-745`). Thus unsafe/duplicate logical entries never count,
and even 100,001 valid entries are allocated and processed before the limit is
noticed. ZIP, tar, and ar readers all build their raw member slices before this
late charge. The complete diagnostic slice can likewise grow without the
`max_recorded_findings` cap limiting working memory.

This violates the normative rule that `max_entry_count` counts files plus
explicit/synthetic directories across all containers and defeats the limit's
resource-exhaustion purpose. The named F08 test only calls `limitAccountant`
directly (`containers_test.go:145-167`); it never proves walker enforcement.

Required rework: charge every logical entry as it is enumerated, including
unsafe, duplicate, and colliding entries; charge each synthetic directory once;
stop or switch to bounded evidence immediately at the closed limit; and add
end-to-end ZIP/tar/ar/directory tests for 100,001 entries plus duplicate/invalid
entry floods. Declared leaf/container sizes and counts should be refused before
expensive materialization where the format exposes them.

### R3 — High: `artifact-manifest-v1` is incomplete and its decoder accepts semantically impossible evidence

The manifest model has the fixed limit vector but no accumulated traversal
counts, emitted-byte totals, or expansion observations (`types.go:458-478`),
although the accepted schema requires accumulated limits. Toolchain role facts
also omit `ResolvedRoot` and environment-search resolution even though the
checkpoint contract requires them; successful local-output facts omit the
independent expectation, so the canonical record does not show what was
compared with the observed node.

The codec validates closed enum membership and ordering, but not the semantic
relationship among trust role, node kind/class/decision, diagnostics, ancestor
decision, and final decision (`codec.go:236-336`, `codec.go:368-437`). A
dependency manifest can therefore be rewritten to contain a compiled node with
`ADMIT_INPUT`, zero findings, and final `ADMIT_INPUT`, then be rehashed and
accepted by `EncodeManifest`/`DecodeManifest`. Decode does not mint an
`Admission`, but the result is still invalid canonical audit evidence and is
unsafe for later checkpoint/cache consumption.

Required rework: include all required archive/accounting and role-checkpoint
evidence, and make codec validation reject every class/role/decision mismatch,
incomplete container, rejecting descendant with an admitted ancestor/final
decision, or diagnostic/decision inconsistency. Add adversarial codec tests that
rehash self-consistent but semantically impossible manifests.

### R4 — High: the accepted conformance corpus is not actually covered

The test names mention A01-A08, C01-C12, F01-F14, T01-T05, and V01, but several
normative cases are reduced or absent:

- A08 covers only one ELF object, not the required object, library, addon, and
  executable output set (`policy_test.go:131-142`).
- C03 has only one fat Mach-O executable; C06 omits PE/Mach-O Node addons; C09
  omits the LLVM wrapper and V8 snapshot branches.
- F03 omits sparse/external extents and other named special-node variants.
- F08 tests accountant methods instead of recursive walkers.
- F14 compares only decision and a reduced member projection
  (`containers_test.go:232-254`), not the required primary diagnostic and exact
  canonical manifest bytes/digest.
- T03 covers fingerprint drift only, and T05 covers digest drift only; escaping
  links/special nodes and path/size/complete-input drift are not exercised.

The accepted contract also requires each shared vector to assert stable class,
decision, primary code, canonical path, and manifest digest. Only one simple Go
source manifest has a fixed digest (`codec_test.go:9-38`), and the generated
fixture helpers are package-local test code rather than a reusable shared byte
corpus for adapter wrappers.

Required rework: publish the reusable accepted fixture corpus with exact
expected records/digests and execute every listed branch through the public
service APIs. Assertions must include the complete required result tuple and
prove no authorization token/later action/publication on every negative case.

## Validation evidence

The current implementation is mechanically healthy, but these green gates do
not cover the findings above:

- `go test -count=1 -race -cover ./internal/artifactpolicy` — exit 0,
  72.7% statement coverage
- `go test -count=10 ./internal/artifactpolicy` — exit 0
- repository-pinned `golangci-lint` v2.12.2 on
  `./internal/artifactpolicy/...` — exit 0, `0 issues.`
- `go vet ./internal/artifactpolicy` — exit 0
- `gofmt -l internal/artifactpolicy` — exit 0, no paths
- Linux amd64 and Windows amd64 package test binaries cross-compiled — exit 0
- `go test -count=1 ./...` — exit 0; all repository packages passed, including
  `cmd/curator` in 411.219s, `internal/artifactpolicy` in 1.684s,
  `internal/godriver` in 65.413s, and the existing Go admission regressions
- `go vet ./...` — exit 0
- `go build ./...` — exit 0
- `task-board validate` — `Board is valid. No issues found.`

No product code was modified by this reviewer. The task should return to
`to-dev`; after the findings are repaired and new task-scoped producer evidence
is attached, it requires another independent reviewer cycle.
