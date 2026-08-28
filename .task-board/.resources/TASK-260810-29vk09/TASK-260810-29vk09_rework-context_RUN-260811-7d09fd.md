# Reviewer verdict for TASK-260810-29vk09

Verdict: **changes requested -> analysis**

## Goal and scope evidence

- Reviewer run: `RUN-260811-7d09fd`
- Authoritative goal at verdict checkpoint: `GOAL-260811-4b4a66` revision 1
- Resolved scope: `TASK-260810-29vk09`
- Review policy: required
- No run directive was pending at the verdict checkpoint.

## Acceptance-blocking finding

### R1 — ELF PIE is not distinguished from a dynamic library

The proposed taxonomy maps `native.executable` to ELF `ET_EXEC` at
`.research/260811_compiled-artifact-taxonomy-and-deny-policy.md:165`, while it
maps all ELF `ET_DYN` to `native.library.dynamic` at line 168. Conformance
case `C01` at line 553 repeats that three-way `ET_EXEC` / `ET_REL` /
`ET_DYN` classification without a PIE case.

That is incomplete for the required native-executable versus dynamic-library
taxonomy. GNU toolchains produce position-independent executables with
`-pie`, and GNU binutils describes `ET_DYN` as either a
position-independent executable or a shared object:

- GCC link options: https://gcc.gnu.org/onlinedocs/gcc-13.3.0/gcc/Link-Options.html
- GNU binutils ET_DYN clarification:
  https://gcc.gnu.org/pipermail/binutils/2020-May/111252.html

The dependency admission outcome remains safely `REJECT`, so this is not a
binary-admission bypass. It is nevertheless material: a common native
executable can receive the wrong stable class and audit evidence, contrary to
the task scope and the deterministic taxonomy/audit acceptance criteria.

Required research rework:

1. Define deterministic ELF class resolution for `ET_DYN` PIE versus shared
   objects using validated structural/use evidence. If a sound distinction
   cannot be made, define an explicit conservative combined/ambiguous compiled
   class rather than reporting every `ET_DYN` artifact as a dynamic library.
2. Preserve `artifact_compiled_dependency_forbidden` as the primary
   dependency-input result in every branch.
3. Require audit observations to retain raw `e_type` plus the structural/use
   facts used for the selected class.
4. Add shared conformance fixtures for dynamic PIE, static PIE, ordinary shared
   objects, renamed/no-suffix forms, and structurally ambiguous variants. Assert
   both the stable class and unchanged rejection result.

## Test evidence

`make check` was run from the project root.

- `go vet ./...` completed successfully.
- The test output reported passing relevant packages including
  `internal/buildmeta`, `internal/buildcache`, and `internal/godriver`.
- The full gate was not green: `cmd/curator` timed out after 10 minutes in
  `TestCompiledProjectStatusRepairRollbackRecovery`, during toolchain
  fingerprinting. Another pre-existing full repository test run had consumed
  the host concurrently for part of this run. After the package timeout, the
  `go test` coordinator continued CPU-spinning with no test children; the
  reviewer terminated that already-failed command. This evidence does not prove
  a product regression, but it also cannot satisfy the task checklist item
  “Tests green.” Re-run the full gate without competing full-suite processes in
  the next producer/reviewer cycle.

## Positive review evidence

- The producer document and attached outcome are identical, SHA-256
  `174dfcb70d313a442a915a2cee884a225e783dbb94be4db48194dd6c15fb57de`.
- It covers source/authored and generated text, native executable/object/static
  and dynamic library classes, Apple bundles, Node/Python extensions, JVM and
  Python bytecode, WebAssembly, serialized IR, containers, external toolchains,
  local outputs, filesystem nodes, and an opaque catch-all.
- Recursive inspection, safe virtual paths, bounded accounting, deny-dominant
  ambiguity handling, stable diagnostics, manifest/checkpoint/receipt evidence,
  the future `verified-binary-v1` seam, and cross-adapter conformance vectors
  are otherwise well specified.
- The architecture recommendation is sound: one manager-owned
  `artifactpolicy` service with adapters only narrowing source profiles. It
  aligns with `.spec/skill-facing-cli-source-closure.md:39-78`.
- Local claims were checked against `internal/buildmeta/models.go`,
  `internal/buildmeta/codec.go`, `internal/buildcache/cache.go`,
  `internal/godriver/errors.go`, and
  `internal/godriver/fingerprint.go`.
- External claims were spot-checked against primary ELF, PE/COFF, JVM,
  WebAssembly, Node, Python packaging, LLVM, Swift, SLSA, Sigstore, and archive
  documentation cited in the producer artifact.
- Document quality gates passed: nonempty, no trailing whitespace, six balanced
  Markdown fences, required-section coverage, and valid local specification
  reference.

## Routing decision

This is ordinary research rework, not a Stop-The-Line boundary and not
implementation work. Route the task to `analysis`, revise the policy and
conformance cases, then submit a new producer handoff and reviewer cycle.
