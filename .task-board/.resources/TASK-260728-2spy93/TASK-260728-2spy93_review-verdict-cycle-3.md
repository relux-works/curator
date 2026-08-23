# TASK-260728-2spy93 review cycle 3 verdict

## Verdict

CHANGES REQUESTED. Route to `to-dev` for focused validator and regression-test rework. The decision text is architecturally coherent, but the gate does not yet enforce two claims it makes about the future reserved schemas.

## Blocking findings

1. `toolchainRequirementV1` can be a string schema while passing the boundary gate. `check_toolchain_requirement_definition` checks only properties, required, and additionalProperties; it never requires `type: object`. JSON Schema object keywords do not constrain non-objects. An in-memory `buildCommandV8` carrying the exact required `$ref`, with `toolchainRequirementV1` changed from `type: object` to `type: string` while retaining exactly id and version, was accepted by `validate_additional_driver_boundary`. This makes the referenced slot accept a package-controlled string even though Decision 0008 sections 4, 5, 11, and Security Impact require the closed object. Rework: require the referenced definition to be an object schema and add red tests for string type, omitted type, and object-or-string unions in addition to the existing slot-shape probes.

2. Claim admission is not closed at the `build_drivers` container. `check_claim_driver_admission` filters out claim schemas without `build_drivers` before choosing the current schema, so a synthetic newest claim v4 with that member removed passed and claim v3 was silently treated as current. The checker also ignores `prefixItems`; a synthetic current claim with a reserved-driver assertion in `prefixItems` passed the checker, and the real Draft 2020-12 validator accepted the shipped valid claim mutated to that reserved driver. This contradicts the requirement that current claim membership equal exactly the admitted set and that retired identifiers be structurally impossible. Rework: identify the newest claim schema before filtering and require its driver member; close or exhaustively inspect the array applicator shape so no parallel path such as `prefixItems` can admit an assertion outside `items.oneOf`. Add both demonstrated mutants as red regressions.

## Independent passing evidence

- Final worktree differs from the accepted TASK-260728-2kp3tv predecessor in exactly three files: Decision 0008, `tools/validate.py`, and `tools/test_validate.py`.
- Decision attachment SHA-256 matches the worktree: `811f6377e2427f73634dc82abec833c7184482f0e9f42e3f10aaf1a2d7c2a489`.
- `make validate`: 42 schemas, 422 vector files, 71 Python tests, and Go tool tests pass.
- `go vet ./tools/...`, gofmt, and `git diff --check` pass.
- `make regenerate-check` passes twice from the byte-identical clean probe.
- `make release-check VERSION=1.0.0-rc.5` passes at `c12d0d8`; frozen manifest digest remains `9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf`.

## Re-review gate

A new reviewer must reproduce rejection of all three mutants: non-object `toolchainRequirementV1`, newest claim schema missing `build_drivers`, and a reserved or retired driver reachable through `prefixItems` or any equivalent array-applicator route. Then rerun all unit, Go, deterministic-regeneration, and rc.5 release/digest gates. Preserve schemas 1 through 7, the rc.5 corpus, Go identities, and the task three-file scope.

The reviewer changed no project file, staging state, commit, publication, pin, or platform claim.