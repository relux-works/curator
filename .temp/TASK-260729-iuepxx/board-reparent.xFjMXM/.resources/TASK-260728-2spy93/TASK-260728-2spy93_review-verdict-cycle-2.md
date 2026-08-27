# Review verdict cycle 2 — CHANGES REQUESTED

## Verdict

Route to `to-dev`. The four cycle-1 findings are materially repaired, and all shipped/deterministic gates are green, but two contract defects remain.

## Blocking findings

1. **The reserved-shape gate false-accepts an unstructured toolchain field.** `tools/validate.py` lines 593-623 and 716-730 check only the property names, required set, and `additionalProperties`; they never check the schema of the `toolchain` property. The test helper at `tools/test_validate.py` lines 1094-1109 deliberately mints every non-`type` property as `{"type":"string"}`, and `test_reserved_schema_eight_shapes_are_enforced_the_moment_they_exist` accepts that object. Independent probe output was: `buildCommandV8 accepted toolchain schema {"type": "string"}`, `repositoryBuildCommandV2 accepted toolchain schema {"type": "string"}`, and `skillBuildTargetV2 accepted toolchain schema {"type": "string"}`. This contradicts decision 0008 lines 249-259 and 299-313, which require the closed `toolchain-requirement-v1` object, and leaves the trusted-preflight boundary unenforced. Rework must make the reserved property schemas structural and add red tests for a string, open object, path-bearing object, and wrong requirement reference in all three slots.

2. **Claim-v4 admission is unconditional despite an explicit retirement path.** Section 2 says a rejected reserved driver is retired unused and never admitted (lines 136-148); the Kotlin downstream obligation explicitly permits retiring both Kotlin identifiers when no backend satisfies the artifact boundary (lines 640-649); and integration is told to integrate only accepted contracts (lines 650-663). But section 8 unconditionally says claim schema 4 admits all eight identifiers (lines 578-587). If Kotlin or another contract is rejected, the integration task must either admit a retired identifier or violate section 8. Rework must define claim-v4 assertions over exactly the admitted set at minting, preserving the per-driver policy `const`; it may state eight only conditionally when all six contracts are accepted.

## Verification evidence

- `make validate` with the task validation environment: 42 schemas, 422 vector files, 60 Python tests, and Go tests passed.
- `go vet ./tools/...`, `gofmt` cleanliness, and `git diff --check` passed.
- Clean probe commit `92cf7a5`: `make regenerate-check` passed twice; `make release-check VERSION=1.0.0-rc.5` passed.
- rc.5 manifest digest remains `sha256:9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf`.
- Scope remains exactly decision 0008 plus `tools/validate.py` and `tools/test_validate.py` relative to the accepted predecessor. No reviewer code, staging, commit, publication, pin, or platform claim was made.

## Re-review gate

Require exact property-schema checks for the three reserved wire definitions, adversarial tests proving malformed toolchain shapes reject, conditional claim-v4 membership consistent with accepted/retired drivers, the full 60-test suite, two deterministic regenerations, and the unchanged rc.5 release/digest gate.