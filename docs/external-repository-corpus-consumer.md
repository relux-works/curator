# External repository black-box corpus runner

This package consumes the accepted rc.5 external-repository corpus with an
implementation-neutral black-box runner. It emits qualification observations;
it does not establish parity, promote a candidate, create a tag, or make a
release claim.

## Versioned corpus boundary

Call `crossmanager.OpenCorpus(root, crossmanager.RC5Boundary)`
with the accepted `interop/rc5/external-repository` root. The boundary requires
`rc5-external-repository-interop-v1`, protocol `1.0.0-rc.5`, manifest schema 1,
and exactly 60 unique cases. Every local case source is authenticated through
the corpus manifest before execution. No corpus bytes or implementation-owned
expected values are embedded in Curator.

## Black-box adapter contract

Configure a `crossmanager.Adapter` for each candidate. Its `Prepare`
callback maps a neutral case to a complete `Invocation`; `Suite.RunAdapter`
then launches the executable directly and never imports a manager package. Its
`Normalize` callback projects protocol-required fields from CLI output while
leaving implementation-private physical paths unconstrained. A non-zero manager
exit is an observation; launch, transport, timeout, or context failures remain
runner errors.

The process environment is caller supplied. Each case receives distinct root,
home, repository, and temporary directories. The adapter must set all manager,
Git, cache, and temporary variables explicitly and must not inherit ambient
configuration. Exact candidate version, revision, binary SHA-256, toolchain,
OS, and architecture are required before execution.

`NativeBoundaryProbe` continuously samples processes and established TCP
connections on macOS and Windows while the manager runs. The suite also hashes
all watched filesystem roots before and after execution, rejects writes outside
the allowlist, and treats any mutation in a case declaring `mutation:false` as
a mismatch. Raw stdout, stderr, filesystem snapshots, and boundary observations
are preserved under the task-owned artifact root.

## Deterministic fixture materialization

Use `MaterializeCorpusFiles` with an empty, task-owned destination. Every source
must be manifest-listed; every target must be a clean relative slash path; file
mode is explicit; targets are sorted before creation; duplicate targets,
existing targets, path traversal, and symlink parents are rejected. The helper
returns source/target/SHA-256 provenance for the run report. Generated fixture
trees are transient inputs and must not be checked in or described as normative
release evidence.

## Machine-readable report

`ReportJSON` emits deterministic JSON identified by
`urn:relux-works:curator:cross-manager-report:v1`. The JSON Schema is
stored with the package as `report.schema.json` and exposed by `ReportSchema`.
Case states are deliberately observational (`observed`, `mismatch`, `error`, or
`not-run`). Reports include exact manager and corpus metadata, output digests,
normalized expected/observed objects, filesystem deltas, process/network
events, violations, and artifact locations. The format contains no parity,
merge, promotion, or release claim.

## Candidate qualification checklist

- Use the accepted task-owned corpus root; do not copy its normative bytes into
  this repository.
- Record corpus base commit, corpus manifest SHA-256, and deterministic corpus
  SHA-256.
- Build or consume exact signed/reviewed Curator and CocoaSkills candidate
  revisions as distributable binaries and record their SHA-256 values.
- Preserve the Curator 0.12.5 and csk 0.9.0 schema-7 rejection baseline.
- Run the same 60 cases with isolated state on native macOS and Windows.
- Preserve every mismatch or error directory before cleanup.
- Revalidate `report.schema.json` consumers and fixture provenance output.
- Keep release pins, tags, parity claims, and promotion decisions in their
  separately authorized tasks.

## Accepted corpus evidence

- Protocol base commit:
  `f5d7673039226ab81de2f4f87e2155ae995c4df3`.
- External corpus manifest SHA-256:
  `cc9e9c0f93b2497a060a533503a4d030d1a715fe1dd4eb8bf9820168a9257697`.
- Deterministic corpus SHA-256:
  `7652fa628812dbd9e72367b6aebc853a0db6178babbc11737fc05a515afbf771`.
- Accepted producer: `TASK-260728-2u5u14:
  shared-rc5-external-repository-cases`.

These values identify qualification inputs only. They are not promotion,
parity, or release claims.
