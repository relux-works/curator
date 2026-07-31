# TASK-260720-37ei85 review rework evidence

## Baseline and rationale

- Frozen curator-spec baseline: `57c1f56846d221ecc55786bd3c2467ec32f11730` (`origin/main`).
- The review gap was that exact declared-property and forbidden-name assertions alone could miss a later arbitrary-property expansion at a root or nested schema boundary.
- `TestFixedConfigurationAndEvidenceSchemasRemainByteFrozen` now pins every fixed manager, system, registry, and audit schema file to its baseline SHA-256. Any byte-level change, including an `additionalProperties` relaxation or nested expansion, fails before such a surface can enter generated evidence.
- The existing exact-property and recursive forbidden-name assertions remain in place to explain the intended manager/system build-policy and registry/audit provenance boundaries.
- `git diff --exit-code 57c1f568... -- <12 fixed schemas>` passed, independently confirming all pinned files are byte-identical to the frozen baseline.

## Pinned schema hashes

| Schema | SHA-256 |
| --- | --- |
| `manager-config-v1.schema.json` | `8c45cedfd962e27e23dbe8e28a49306d997aeef77f39b8bfc06de81c6ef7c657` |
| `system-config-v1.schema.json` | `6a89fb538621f78132a09208ee4cd5c57ea78d530d96596cef15a8814a9f38c3` |
| `audit-record-v1.schema.json` | `06c97bcc2562688ac399ba948be25258fa3d3954b29abaf5d972b6e142ed8cb4` |
| `registry-bundle-v1.schema.json` | `b02188a5dd17d02a0921d86ccf0ee6650f74482e95dd050829815f8d8adc49c5` |
| `registry-log-entry-v1.schema.json` | `62d21c6d0c87097e5fa32217bb2b22035f099007bf5801ec44635680e1fa00ed` |
| `registry-meta-response-v1.schema.json` | `5aa736ea0d0c3edc78fe99b130ad00def38694e3b94d2b6898a87bf97a028164` |
| `registry-snapshot-v1.schema.json` | `6d166071d904c13a93ddf9d84b21fa3377e8bba58df263459f9161f649b031ab` |
| `records-response-v1.schema.json` | `041f507e5d37b1e39297f66628af37fbcfa2dd587a2da08bd9a29dc704690087` |
| `log-response-v1.schema.json` | `0513f35b75f291c0295dc80cc2c97d390190c38aa7421edb03de94ec17932f86` |
| `submission-response-v1.schema.json` | `463e72cc9750340a6d89f14e7407e1b3d7b16f700e3cd913f2933a8d00dd28a5` |
| `health-response-v1.schema.json` | `6da2b960e6d8140ba18da2a4f87e5207c406d979b3907ae5bc5716374f9d4e7d` |
| `error-response-v1.schema.json` | `06ebd9618eb42ae61e7bcb34959a15e0eb3796bcf44d2b6741b74871d223c0bb` |

## Verification rerun

- `gofmt -d tools/generate-vectors/main.go tools/generate-vectors/main_test.go` — pass, no output.
- `go vet ./tools/...` — pass.
- `go test -count=1 ./tools/generate-vectors` — pass.
- `git diff --check` — pass.
- `PATH=<task-venv>/bin:$PATH make validate` — pass: 35 schemas, 164 vector files, 8 Python tests, and all Go tool tests.
- `make regenerate` — pass twice.
- `GIT_INDEX_FILE=<task-alternate-index> make regenerate-check` — pass twice with no conformance diff.
- Aggregate `conformance/v1` SHA-256 remained `8eb88a753b7a5cafb678dfb46dbf8fb4657fcb4a56a69cdbe1d39ebb3ae04f32` after both regeneration and regeneration-check passes.
- `git diff --cached --name-only` remained empty; the real Git index was not changed.

No product files were changed during this recovery run because the rework implementation and recorded evidence were accurate.
