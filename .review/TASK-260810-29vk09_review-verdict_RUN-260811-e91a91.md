# Reviewer verdict for TASK-260810-29vk09

Verdict: **accepted -> done**

## Goal and scope evidence

- Reviewer run: `RUN-260811-e91a91`
- Authoritative goal immediately before verdict: `GOAL-260811-0bef90`
  revision 1
- Resolved scope: `TASK-260810-29vk09`
- Review policy: `required`
- No operator directive was pending at the initial checkpoint.

The goal requires exactly one evidence-backed reviewer branch. This artifact
records the `accepted` branch; no `changes_requested` or `stop_the_line` branch
is recorded for this run.

## Acceptance result

The revised policy satisfies the task description, scope, acceptance criteria,
and checklist. No acceptance-blocking finding remains.

1. **Artifact taxonomy and deterministic decisions.** The shared taxonomy
   covers authored and generated text, metadata, inert/opaque data, native
   executables, objects, static and dynamic libraries, an explicit ambiguous
   ELF `ET_DYN` compiled class, Apple bundles, native extensions, JVM/Python/JS
   bytecode or cached executable state, WebAssembly and serialized IR,
   archives/compression, directories, links/special nodes, trusted external
   toolchains, and protected local outputs. A closed detector registry and
   `opaque.unknown` catch-all make unknown or incomplete interpretation reject.
2. **Trust boundaries.** The policy assigns trust from Curator-controlled
   provenance and pipeline state. Compiled `dependency_input` rejects across
   all adapters; independently selected/fingerprinted `external_toolchain`
   components and causally produced/receipted `local_build_output` artifacts
   remain outside that prohibition without becoming dependency source.
3. **Recursive detection and fail-closed behavior.** Recognized containers are
   inspected recursively through the same classifier. Safe virtual paths,
   entry-kind restrictions, canonical ordering, overflow-checked closed limits,
   deny-dominant ambiguity resolution, and explicit unsupported/encrypted/
   incomplete diagnostics prevent partial admission.
4. **Diagnostics and audit evidence.** The cross-adapter diagnostic table
   defines stable codes and required fields. `artifact-manifest-v1`, external
   toolchain checkpoints, and local-output receipts bind the policy, detector
   identities, origin, hashes, paths/container chains, raw format facts,
   decisions, targets, toolchains, and protected publication evidence.
5. **Future capability seam.** `verified-binary-v1` is a separate manager-owned
   capability and explicit graph edge. It is absent today; a signature alone
   cannot admit bytes. Cases `V01`-`V06` cover identity, provenance, signer/
   builder policy, target compatibility, nested subjects, and TOCTOU-safe use.
6. **Architecture fit.** One manager-owned language-neutral `artifactpolicy`
   service with adapters allowed only to narrow source profiles matches the
   repository source-closure invariant and prevents ecosystem-specific bypasses.

## Reviewer-requested ELF rework

Finding R1 from `RUN-260811-7d09fd` is resolved:

- Raw `e_type` is retained and `ET_DYN` is resolved separately.
- Structurally validated `DF_1_PIE` selects `native.executable`, with
  interpreter and no-interpreter/static-PIE variants.
- Non-conflicting `DT_SONAME` or manager-resolved link/load use can select
  `native.library.dynamic` under the ordered rules.
- Insufficient or conflicting non-decisive facts select
  `native.elf.et_dyn_ambiguous`, never an invented dynamic-library label.
- Every dependency branch keeps `REJECT` with
  `artifact_compiled_dependency_forbidden` as the primary diagnostic.
- Audit evidence retains raw ELF header, program-header, dynamic-tag, and
  manager-resolved use-edge observations plus the selected rule/reason.
- Shared fixtures `C01a`-`C01f` cover GNU dynamic PIE, GNU static PIE, ordinary
  shared objects, renamed/no-suffix forms, insufficient evidence, conflicting
  evidence, and validated-use fallback. They assert both stable class/variant
  and unchanged primary rejection.

Primary-source fact checking confirms the revised distinction:

- GCC defines `-pie`, `-static-pie`, and `-shared` as executable/shared-object
  link outputs: https://gcc.gnu.org/onlinedocs/gcc-13.3.0/gcc/Link-Options.html
- GNU binutils documents `ET_DYN` as either a position-independent executable
  or a shared object:
  https://gcc.gnu.org/pipermail/binutils/2020-May/111252.html
- GNU `readelf` identifies PIE by scanning `DT_FLAGS_1` for `DF_1_PIE`:
  https://sourceware.org/pipermail/binutils/2021-June/116921.html
- The generic ELF ABI says `PT_INTERP` is meaningful for executables but may
  occur in shared objects, and defines `DT_SONAME` as ignored for executables
  and optional for shared objects:
  https://gabi.xinuos.com/elf/07-pheader.html and
  https://gabi.xinuos.com/elf/08-dynamic.html

These sources support the document's use of `DF_1_PIE` as decisive PIE
evidence and its conservative treatment of other structural/use facts.

## Artifact and validation evidence

- Research source and board outcome are byte-identical, 796 lines, SHA-256
  `c5334433d6eddf37109e612a97866024a17d38c15cff7d7e5e36dac69fe0df15`.
- Document gates passed: nonempty; no trailing whitespace; six balanced
  Markdown fences; required-section coverage; valid local specification
  reference; source/outcome parity; `C01a`-`C01f` presence and primary-code
  invariance.
- Local claims were checked against `.spec/skill-facing-cli-source-closure.md`,
  `internal/buildmeta/models.go`, `internal/buildmeta/codec.go`,
  `internal/buildcache/cache.go`, `internal/godriver/errors.go`, and
  `internal/godriver/fingerprint.go`.
- `go test -count=1 ./...` completed uncontended with exit 0. All packages
  passed; notably `cmd/curator` passed in 368.404s, `internal/buildcache`,
  `internal/buildmeta`, and `internal/godriver` passed.
- `go vet ./...` completed with exit 0.
- `gofmt -l` over every tracked Go source emitted no paths and exited 0.
- Producer logbook evidence records an exact clean-snapshot `make check` exit
  0 after reversibly excluding an unrelated historical `.temp` tree.

Reviewer anomaly: an exact `make check` in the restored workspace completed
`go vet` and all product test binaries successfully, then `cmd/go` consumed
about 19 GB while computing test-cache input IDs through that historical
`.temp` tree. Sampling located the work in
`cmd/go/internal/test.computeTestInputsID`; the reviewer terminated only this
reviewer-launched coordinator to avoid host exhaustion. The uncached full suite
above avoids that cache traversal and proves the tests themselves. This is an
unrelated validation-environment contamination issue, not a policy artifact or
product regression, and it does not weaken the producer's clean exact-gate
evidence.

## Routing decision

Accept the research deliverable and route `TASK-260810-29vk09` to `done`. This
reviewer run supplies no `commit_ack`; the transition does not complete its
parent Story because other scoped discovery tasks remain active.
