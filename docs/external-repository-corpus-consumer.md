# External repository corpus consumer

This guide describes Curator's consumer for the accepted 60-case rc.5
external-repository corpus. The consumer authenticates externally supplied
bytes and binds them to Curator tests; it does not copy the normative corpus or
turn a local run into a release or cross-manager parity claim.

## Versioned corpus boundary

Call `conformanceconsumer.OpenCorpus(root, conformanceconsumer.RC5Boundary)`
with an externally supplied `interop/rc5/external-repository` root. The boundary requires
`curator-conformance-corpus/v1` and protocol `1.0.0-rc.5`. The consumer reads
only paths listed in `manifest.json` and verifies each selected file against its
declared SHA-256 digest. The root is supplied as
`CURATOR_EXTERNAL_REPOSITORY_CORPUS_ROOT`; no corpus bytes or repository-local
golden values are embedded here.

## Black-box adapter contract

Implement `conformanceconsumer.Runner`, or configure `ProcessRunner` with the
manager executable and arguments. `ProcessRunner` links no manager package. It
sends one `RunRequest` JSON object on standard input and captures exit status,
standard output, and standard error as observations. A non-zero manager exit is
data, while launch, transport, or context failures are runner errors.

The process environment is caller supplied. Harness authors should provide
only the environment required by the chosen case and should isolate the
fixture, manager home, cache, and temporary directories per run.

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
Case states are deliberately observational (`observed`, `mismatch`, `error`,
or `not-run`); the report format contains no qualification, parity, platform,
merge, or release claim.

## Corpus update checklist

- Replace the caller-supplied corpus root; do not copy corpus files into this
  repository.
- Record the accepted corpus commit and recompute the exact manifest SHA-256.
- Confirm the accepted manifest still declares protocol `1.0.0-rc.5` and that
  the `curator-conformance-corpus/v1` parser remains compatible.
- Diff every changed manifest entry and review all case additions, removals,
  expected outcomes, sources, threat coverage, and lifecycle coverage.
- Re-run `internal/conformanceconsumer` and `internal/rc5interop` with
  `CURATOR_EXTERNAL_REPOSITORY_CORPUS_ROOT` set and update only reviewed Curator
  bindings.
- Revalidate `report.schema.json` consumers and fixture provenance output.
- Keep qualification, native platform execution, conformance claims, release
  pins, tags, and promotion decisions in their separately authorized tasks.

## Accepted rc.5 input evidence

- Consumer code base: `74fe162415d800cd0a6975313827f9dc8594d299`.
- Protocol tag commit:
  `f5d7673039226ab81de2f4f87e2155ae995c4df3`.
- `conformance/v1` tree:
  `0ea6b7166482cfe951fdf62d72dbcbe3b5d8b8e4`.
- `conformance/v1/manifest.json` SHA-256:
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`.
- Accepted external-repository corpus SHA-256:
  `7652fa628812dbd9e72367b6aebc853a0db6178babbc11737fc05a515afbf771`.
- The accepted corpus has 60 cases, 18 architecture-v6 threat rows, and 12
  lifecycle boundaries. Candidate platform claims are empty; platform evidence
  is recorded by native qualification, not inferred from corpus presence.

These values identify inputs only. They do not claim Linux support.
