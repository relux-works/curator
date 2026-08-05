# External repository corpus consumer (provisional draft)

This draft describes scaffolding for consuming the rc.5 external-repository
corpus with an implementation-neutral black-box runner. It is provisional: it
does not qualify a manager, establish platform or parity results, promote the
candidate, or provide release evidence.

## Versioned corpus boundary

Call `conformanceconsumer.OpenCorpus(root, conformanceconsumer.RC5Boundary)`
with an externally supplied `conformance/v1` root. The boundary requires
`curator-conformance-corpus/v1` and protocol `1.0.0-rc.5`. The consumer reads
only paths listed in `manifest.json` and verifies each selected file against its
declared SHA-256 digest. Replacing the candidate with an accepted corpus is a
caller input change; no corpus bytes or repository-local golden values are
embedded here.

## Black-box adapter contract

Implement `conformanceconsumer.Runner`, or configure `ProcessRunner` with the
manager executable and arguments. `ProcessRunner` links no manager package. It
sends one `RunRequest` JSON object on standard input and captures exit status,
standard output, and standard error as observations. A non-zero manager exit is
data, while launch, transport, or context failures are runner errors.

The process environment is caller supplied. Harness authors should provide
only the environment required by the chosen case and should isolate the
fixture, manager home, cache, and temporary directories per run. This draft
does not run the unreleased manager qualification suite.

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
`urn:relux-works:curator:black-box-consumer-report:v1`. The JSON Schema is
stored with the package as `report.schema.json` and exposed by `ReportSchema`.
Case states are deliberately observational (`observed`, `mismatch`, `error`,
or `not-run`); the report format contains no qualification, parity, platform,
merge, or release claim.

## Rebase checklist after corpus acceptance

- Replace the caller-supplied corpus root; do not copy candidate files into this
  repository.
- Record the accepted corpus commit and recompute the exact manifest SHA-256.
- Confirm the accepted manifest still declares protocol `1.0.0-rc.5` and that
  the `curator-conformance-corpus/v1` parser remains compatible.
- Diff the accepted and candidate manifest entries under
  `fixtures/external-repository`, `vectors/external-repository-acquisition.json`,
  `vectors/external-repository-lifecycle.json`, and
  `expected/external-repository`; review every changed digest.
- Re-run the package tests with `CURATOR_CONFORMANCE_ROOT` set to the accepted
  root and update only adapter mappings affected by reviewed corpus changes.
- Revalidate `report.schema.json` consumers and fixture provenance output.
- Keep qualification, native platform execution, conformance claims, release
  pins, tags, and promotion decisions in their separately authorized tasks.

## Candidate/base evidence for this draft

- Consumer code base: `74fe162415d800cd0a6975313827f9dc8594d299`.
- Read-only candidate commit:
  `f5d7673039226ab81de2f4f87e2155ae995c4df3`.
- Candidate `conformance/v1` tree:
  `0ea6b7166482cfe951fdf62d72dbcbe3b5d8b8e4`.
- Candidate manifest SHA-256:
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`.
- Candidate path used for focused tests:
  `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-2u5u14/curator-spec-rc5-worktree/conformance/v1`.
- Acceptance dependency: `TASK-260728-2u5u14: shared-rc5-external-repository-cases`.

These values identify provisional inputs only. They are not promotion or
release claims.
