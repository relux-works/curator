# Review cycle 4 verdict — CHANGES REQUESTED

## Scope reviewed

Reviewed the final three-file delta against the accepted TASK-260728-2kp3tv candidate: decisions/0008-additional-language-driver-boundary.md, tools/validate.py, and tools/test_validate.py. Cross-checked decision 0007 and the downstream obligations for TASK-260728-2jaw7h, the Rust/Swift/Kotlin contract tasks, TASK-260728-251p01, TASK-260728-2bu2q6, and STORY-260728-327soo. No project file was edited, staged, committed, published, pinned, or used to claim a native platform.

## Passing evidence

- Exact delta is three owned files; manifest and rc.5 release digests remain sha256:9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf and sha256:b32ee9d35fc2fce0539d0b1c4b15f9c5239115c22fe54a72401f5da6d6646441.
- validate.py: 42 schemas and 422 vector files.
- Full Python discovery: 76 tests pass in the task worktree and the exact clean probe.
- Go tests, go vet, gofmt, and git diff --check pass.
- The cycle-3 mutant probe proves all three historical escapes are accepted by the pre-fix bodies and rejected by the final bodies: scalar toolchainRequirementV1, newest claim without build_drivers, and reserved driver through prefixItems. The Draft 2020-12 probe reproduces the prefixItems bypass and the final gate closes it.
- Frozen manifest schema probes for agent-skill and csk-skill versions 1 through 5 also reject all six reserved identifiers; versions 6 and 7 plus descriptor, receipt, marker, and claim rejections are covered by the submitted gate.
- In an exact clean copy, make validate passes, make regenerate-check passes twice, and make release-check VERSION=1.0.0-rc.5 passes at 86ec275. The first ambient make attempt failed before validation because ambient python3 lacked jsonschema; rerunning with the task validation environment on PATH passed and is not a product failure.

## Blocking finding

The artifact closure claimed by decision 0008 section 11 item 11 is incomplete in tools/validate.py. The buildArtifactV1 check compares only the property-name set and additionalProperties. It does not require type to be exactly object, does not require path, sha256, and size to remain required, and does not hold those three property schemas to their canonical definitions.

Two in-memory mutants pass validate_additional_driver_boundary unchanged:

1. Changing buildArtifactV1 type from object to the union [object, string] is ACCEPTED by the boundary gate; Draft 2020-12 then accepts the string launcher plus runtime as a build artifact while all existing positive objects still validate.
2. Changing required to an empty list is ACCEPTED by the boundary gate; Draft 2020-12 then accepts an empty object as a build artifact.

This contradicts the normative native-executable-v1 model: exactly one bounded regular file, with a path, digest, and size. It also repeats the same object-keyword escape that cycle 3 correctly closed for toolchainRequirementV1 and claim assertions. Green submitted tests do not cover it; test_multi_file_artifact_is_rejected checks only an extra property.

## Required rework and re-review gate

Route to implementation rework. Close buildArtifactV1 as an exact object schema: type object exactly, required exactly path plus sha256 plus size, properties exactly those three, additionalProperties false, and each property schema fixed to its canonical shared definition or an equivalently strict compiled-schema contract. Add red regressions for omitted type, scalar type, object-or-scalar union, each missing required member, and widened path/digest/size schemas. Preserve all cycle-1 through cycle-3 fixes and the three-file scope, then rerun 76 tests, the mutant probe, Go/vet/format/diff checks, deterministic regeneration twice, and the unchanged rc.5 release/digest gates.

Everything else reviewed in decision 0008 is coherent with the task acceptance criteria; this verdict does not reopen the six driver identities, version allocation, local/external ownership, toolchain placement, manager-worker-v2 split, claim admission closure, or downstream traceability.