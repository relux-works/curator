# Reviewer verdict — accepted

Task: BUG-260731-2rhy74 — marker-v2-writer-conformance-fixture
Reviewer run: RUN-260731-387024
Verdict: accepted
Board branch: done

## Scope reviewed

- curator-spec PR 15 exact head: `2629aecff19a33e8cd1b5ebcfd898894ff1eeae0`; marker-v2 implementation commit: `0c81c1f8d5321d822be2a2817b05aea03e656e15`.
- CocoaSkills PR 16 was rebased and merged to main at `c7dbd6daf6562a264275fca06b50a527bce236d4`; the marker change is present on merged main as `65dac15` (original PR commit `8a02e179fe35205490f081a7caa2e191b524e534`).
- No tags or GitHub Releases were created.

## Findings

No blocking or non-blocking correctness findings.

The change cleanly separates the two protocol roles. The legacy `conformance/v1/expected/marker.json` blob is unchanged at main, release/v1.0.0-rc.6, and PR 15 head (blob `b458e1153d7733a1291915fc9e88919d8e3f9310`, SHA-256 `80989f850887814ec09c724a7dd891ac7e2422d5fef7e31f330be3554aa9b28a`). The generated `expected/marker-v2.json` is SHA-256 `22117126c8932769636bd5ca1f1623f26a3c55d4bb9da3266acf3dc3a3b7fc2c`, represents the schema-5 golden skill, carries `build_roots: []` and `builds: {}`, omits `build_source`, and differs from marker-v1 only in those two fields plus `schema_version`.

Generator, validator, manifest inventory/hash, and rc.6 release metadata all cover the new writer fixture. Validator and release-gate tests freeze marker-v1, require marker-v2, schema-check both roles, reject invented build state, reject an illegal build_source, and reject any unrelated installation delta.

CocoaSkills reads `expected/marker-v2.json` directly as bytes, compares both marker payload semantics and canonical writer bytes, and keeps marker-v1 as an independent legacy-read assertion. The shared serializer returns UTF-8/LF bytes and both marker writer sites use `Path.write_bytes`, closing the prior Windows CRLF finding without platform special cases.

## Independent local verification

At curator-spec exact head `2629aecff19a33e8cd1b5ebcfd898894ff1eeae0`:

- `make release-check VERSION=1.0.0-rc.6`: exit 0.
- Validation: 42 schemas and 448 vector files.
- Python tests: 79 passed.
- `go test ./tools/...`: passed with repository-compatible Go 1.23.0.
- Regeneration produced no diff.
- rc.6 release gate passed at the exact head.
- `git diff --check`, `gofmt -l tools`, and post-regeneration status: clean.

At merged CocoaSkills main `c7dbd6daf6562a264275fca06b50a527bce236d4`, against the exact PR 15 conformance root:

- Focused marker/conformance selection: 35 passed, 190 deselected.
- Strict mypy: no issues in 67 source files.
- Full suite: 1,319 passed, 50 skipped.
- Negative fail-closed check against published rc.5 `f5d7673039226ab81de2f4f87e2155ae995c4df3`: expected failure, explicitly reporting that the root publishes no `expected/marker-v2.json`.

## Published CI evidence

- curator-spec PR 15 head `2629aecf`: 8/8 checks green. Specification and implementation conformance passed on Ubuntu, macOS, and Windows; formatting and links passed. Runs: 30616107890 and 30616107892.
- CocoaSkills PR 16 reviewed head `f8b90a5`: all 12 Python matrix cells across Ubuntu/macOS/Windows, strict mypy, and build artifacts passed (run 30641011440).
- Rebased merged main `c7dbd6da`: all 12 Python matrix cells across Ubuntu/macOS/Windows, strict mypy, and build artifacts passed again (post-merge run 30643627899).

## Architecture and acceptance

The implementation matches the acceptance criteria and fits the existing generated-fixture, manifest inventory, release-gate, typed marker-reader, and canonical serialization boundaries. Legacy-read compatibility remains byte-exact; writer conformance is explicit and fail-closed; the publication residual and literal cross-platform CI boundary are satisfied.

Accepted.