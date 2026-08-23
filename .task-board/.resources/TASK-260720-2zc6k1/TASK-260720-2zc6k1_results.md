# TASK-260720-2zc6k1 implementation results

## Baseline and scope

- Isolated worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-2zc6k1/worktree`
- Base: curator-spec `57c1f56846d221ecc55786bd3c2467ec32f11730`
- Accepted composite source: `/Users/iv/Developer/ReluxWorks/curator-spec/.temp/TASK-260720-12iigs/worktree`
- The tracked diff and accepted untracked product files were compared byte-for-byte after import.
- Relative to that composite, only the task-owned claim-v2 schema, generator/tests, generated v2 cases/index/manifest, and `tools/validate.py` protocol assertion differ.
- No commit or real-index staging was performed.

## Implementation

- Added strict `conformance-claim-v2.schema.json` with schema version 2, protocol `1.0.0-rc.4`, the existing required claim fields and closed allowed sets.
- Split `protocolVersion` rc.4 from immutable `conformanceClaimV1ProtocolVersion` rc.3.
- Generated the exact current-writer v2 fixture plus rejections for rc.3, schema version 1, duplicate classes, fail result, a missing required field, and an unknown field.
- Updated the generated suite manifest and validator assertion to rc.4.
- Added Go guards for the v2 schema/case inventory, exact valid fixture, manifest identity, and byte-frozen claim-v1 artifacts.

## Compatibility evidence

Claim-v1 artifacts remained byte-identical before and after repeated regeneration:

- schema: `c9f49460618ccc8b1d7d2dfaf760fc6ad3a53a870a6685a685ddc148d3c87b3f`
- valid case: `799682489be118331135d91798db90b8d020cbb703207331824ab113f037693c`
- invalid case: `de9568757a2bb89c87702e47f6d9c162df24f5ee964f1ef49b9e191ed94b7017`

Current outputs:

- claim-v2 schema: `4c05a97a1aa9f7dafe629a406a853239928413e79e95488ac2b20ebd0c52a38c`
- rc.4 manifest: `ce8a26cfe60c8ada0782328410c3e8e4cde78dd67c43dade066703a1d0703d36`
- manifest inventory: 164 vector files across 35 schemas

## Verification

- `gofmt -d tools/generate-vectors/main.go tools/generate-vectors/main_test.go` — pass
- `go vet ./tools/...` — pass
- `go test ./tools/generate-vectors` — pass
- `make regenerate` — pass
- `make validate` — pass: Python validation, 8 Python tests, and all Go tool tests
- `make regenerate-check` — pass twice with a task-local alternate Git index seeded from the intended uncommitted conformance baseline
- `git diff --check` — pass

The host Python lacked `jsonschema`; validation used a task-local virtual environment installed from pinned `requirements-dev.txt` (`jsonschema==4.25.1`).