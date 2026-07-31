# TASK-260728-2spy93 rework 04

Closes the single blocking finding in `TASK-260728-2spy93_review-verdict-cycle-4.md`.
Nothing else in the decision was reopened: the six driver identities, version
allocation, local/external ownership, toolchain placement, `manager-worker-v2` split,
claim admission closure, and downstream traceability are untouched. Still exactly three
files different from the accepted `TASK-260728-2kp3tv` base:
`decisions/0008-additional-language-driver-boundary.md` (1049 lines), `tools/validate.py`,
`tools/test_validate.py`.

## The finding

`validate_additional_driver_boundary` compared only `buildArtifactV1`'s property-name
set and `additionalProperties`. `properties`, `required` and `additionalProperties`
constrain objects only, so the definition could keep all three names and still stop
being an object. The reviewer's two mutants both passed the gate unchanged:

1. `type` widened from `object` to `["object","string"]` — Draft 2020-12 then accepts a
   bare launcher string as a build artifact;
2. `required` emptied — Draft 2020-12 then accepts `{}` as a build artifact.

In both cases every shipped positive case still validates, which is exactly why a
member-name comparison cannot see it. This is the same object-keyword escape class that
cycle 3 closed for `toolchainRequirementV1` and the claim assertions, so the fix uses
the same machinery rather than a new special case.

## Fix — `tools/validate.py`

`buildArtifactV1` is now closed the way every driver-bearing shape already is.

- `ARTIFACT_DEFINITION`, `ARTIFACT_MEMBERS` and `ARTIFACT_PROPERTY_SCHEMAS` state the
  exact shape: required exactly `path` + `sha256` + `size`, no optional members, each
  pinned to `#/$defs/portablePath`, `#/$defs/sha256`, `#/$defs/nonNegativeSafeInteger`.
- New `check_build_artifact_closure` runs the shared `check_closed_member_set` (which
  already requires `"type": "object"` exactly, the closed property set, the closed
  `required` set and `additionalProperties: false`) and `check_exact_property_schemas`,
  re-raising with the `native-executable-v1 is not the closed single-file artifact`
  framing so the diagnostic names both the class and the precise cause. A missing
  definition is its own failure rather than a `KeyError`.
- `check_closed_member_set` gained an `isinstance(definition, dict)` guard. A boolean is
  a valid Draft 2020-12 schema and `true` accepts everything, so a definition need not
  be an object to be minted at all. This applies uniformly to the artifact, the
  requirement definition and the claim assertions.
- New `check_build_artifact_rejections` proves the closure against the compiled
  `build-receipt-v1` and `build-receipt-v2` validators, using the generated positive
  cases as the source of truth for what a valid artifact is. It rejects a launcher
  string, an array of bundle members, a boolean, an empty object, an extra `runtime`
  member, a traversing path, an absolute path, an unprefixed digest, a negative size, a
  non-integer size, and each of the three members omitted. This is what also holds
  `portablePath`, `sha256` and `nonNegativeSafeInteger` in place: pinning the three
  `$ref` values alone is satisfied by a reference to a definition that has itself been
  opened.
- `check_toolchain_requirement_definition` now derives its `referencing` list from the
  entries whose exact-schema table actually names `TOOLCHAIN_REQUIREMENT_REF`, because
  the table is no longer toolchain-only. Without this, adding the artifact entry would
  have made the unminted `toolchainRequirementV1` reservation look overdue.

## Fix — decision 0008

Made precise, not reopened.

- Section 3 gained a normative wire paragraph: `buildArtifactV1` MUST be an object
  schema, `"type": "object"` exactly, never omitted, never unioned with a scalar, never
  a boolean schema, requiring exactly the three members, closed, each bound to its
  canonical shared definition — with the reason stated, and with the compiled-validator
  proof named.
- Section 11 item 11 restates the exact object schema, the pinned property schemas, the
  behavioural proof against both receipt surfaces, and why a name-only comparison is not
  a closure here.
- Security impact: a definition that keeps the three names while declaring itself a
  string, a union, no type, or a boolean schema — or that requires none of them — is
  named as readmitting the rejected `runtime-bundle` class through the definition
  section 3 relies on to exclude it.

## The two mutants, before and after

`.temp/TASK-260728-2spy93/mutant-probe.py` now restores four checkers to their pre-fix
bodies and runs the cycle-3 and cycle-4 mutants together. Exit 0 = every mutant accepted
before its fix, rejected after. It writes no project file.

```
[ok] pre-fix  | buildArtifactV1 typed object-or-string: ACCEPTED
[ok] post-fix | buildArtifactV1 typed object-or-string: REJECTED native-executable-v1 is not the closed
               single-file artifact: ... is not an object schema, so its closed member set constrains
               nothing: found type ['object', 'string']
[ok] pre-fix  | buildArtifactV1 requires nothing: ACCEPTED
[ok] post-fix | buildArtifactV1 requires nothing: REJECTED native-executable-v1 is not the closed
               single-file artifact: ... required set is not closed: expected ['path','sha256','size'],
               found []
[ok] draft2020-12 | shipped build-receipt-v2 rejects a launcher string (1 errors), mutated variant
               accepts it (0 errors), positive case still validates: True
[ok] draft2020-12 | shipped build-receipt-v2 rejects an empty artifact object (3 errors), mutated
               variant accepts it (0 errors), positive case still validates: True
PROBE PASS   (exit 0)
```

The last two lines are the point: the escape is not a reading of the checker, it is what
the compiled Draft 2020-12 validator does with the shipped receipt schema plus one
widened `buildArtifactV1` — and the positive case is unaffected either way. The three
cycle-3 mutants and their Draft 2020-12 line still pass in the same run.

## In-memory mutant sweep (16 mutants, all rejected, baseline accepted)

Run against `validate_additional_driver_boundary` with the shipped checkers:

| Mutant | Rejected by |
| --- | --- |
| `type: ["object","string"]` (reviewer 1) | not an object schema |
| `required: []` (reviewer 2) | required set is not closed |
| `type` omitted | not an object schema |
| `type: "string"` | not an object schema |
| definition replaced by `true` | not a schema object |
| `required` without `path` / `sha256` / `size` (3 mutants) | required set is not closed |
| `path` / `sha256` / `size` widened to a bare type (3 mutants) | `.<member>` is not exactly `{...}` |
| `portablePath` opened to `{}` | compiled receipt accepts a traversing path |
| `sha256` opened to `{"type":"string"}` | compiled receipt accepts an unprefixed digest |
| `nonNegativeSafeInteger` opened to `{"type":"integer"}` | compiled receipt accepts a negative size |
| definition removed | no longer declares the closed single-file artifact |
| `additionalProperties: true` | does not close additionalProperties |

## Regression tests (83 total, was 76)

- `test_artifact_must_be_an_object_schema` — scalar type, `["object","string"]` union,
  omitted type, and the boolean schema `true`.
- `test_scalar_artifact_type_lets_a_launcher_pass_the_real_validator` — proves against
  the compiled `build-receipt-v2` validator that the union accepts a launcher string
  while the generated positive case is unaffected, so the type requirement is not style.
- `test_artifact_cannot_drop_a_required_member` — each of the three omitted, plus the
  empty `required` list.
- `test_artifact_properties_are_pinned_to_the_shared_definitions` — path, digest and size
  each widened to a bare type.
- `test_artifact_reference_targets_cannot_be_widened_underneath` — `portablePath`,
  `sha256` and `nonNegativeSafeInteger` opened, caught behaviourally rather than
  structurally.
- `test_artifact_definition_cannot_disappear` — the definition removed entirely.
- `test_artifact_rejections_cover_the_frozen_receipt_surfaces` — every schema file
  referencing `buildArtifactV1` is in `FROZEN_ARTIFACT_CASES`, so a later receipt version
  cannot publish an artifact the behavioural proof never sees.

`test_multi_file_artifact_is_rejected` is unchanged and still asserts the
`closed single-file artifact` framing, which the re-raise preserves.
`test_reserved_toolchain_slot_carries_the_exact_requirement_reference` was updated: its
invariant is now that the three reserved slots are exactly the shapes carrying the
requirement reference, not that they are the only entries in the exact-schema table.

## Plant-and-revert CLI probes

Both plants were restored and verified byte-identical by SHA-256
(`4c725692979f64104664f51391f7f99b63d84dedc350c34d92a9fa0e3a620190`), and `validate.py`
exits 0 after each revert.

| Plant | Command | Exit | Rejected by |
| --- | --- | --- | --- |
| `buildArtifactV1.type = ["object","string"]` | `python tools/validate.py` | 1 | boundary gate: `is not an object schema ... found type ['object', 'string']` |
| `buildArtifactV1.required = []` | `python tools/validate.py` | 1 | boundary gate: `required set is not closed: expected ['path','sha256','size'], found []` |

**No honest limitation this time, unlike rework 02 and 03.** The union plant was run
through each earlier validation layer in isolation on the same planted file:

```
validate_schemas:                     PASSED (blind to the widening)
validate_manifest:                    PASSED (blind to the widening)
validate_vector_semantics:            PASSED (blind to the widening)
validate_additional_driver_boundary:  rejected -- native-executable-v1 is not the closed
                                      single-file artifact: ...
```

`validate_schemas` runs *before* the boundary gate in `main()`, so this is not an
ordering artifact: no other layer sees the widening, and the submitted gate is what
rejects it at the CLI.

## Gates (standalone runs, real exit codes)

Task worktree `.temp/TASK-260728-2spy93/curator-spec-worktree`, venv `validation-venv`:

| Command | Exit |
| --- | --- |
| `python tools/validate.py` (42 schemas, 422 vector files) | 0 |
| `python -m unittest discover -s tools -p 'test_*.py'` (83 tests) | 0 |
| `go test ./...` | 0 |
| `go vet ./tools/...` | 0 |
| `gofmt -l .` (no output) | 0 |
| `python -m compileall -q tools` | 0 |
| `git diff --check` | 0 |
| `python .temp/TASK-260728-2spy93/mutant-probe.py` | 0 |

Clean-checkout probe `.temp/TASK-260728-2spy93/clean-probe-04` at commit `3c2ee38`
(fresh tree, fresh repo, no history from the task worktree, task validation environment
on `PATH`):

| Command | Exit |
| --- | --- |
| `make validate` | 0 |
| `make regenerate-check` (first) | 0 |
| `make regenerate-check` (second) | 0 |
| `make release-check VERSION=1.0.0-rc.5` | 0 |

`git status` is empty after all four, so regeneration is deterministic and moved nothing.

## Preservation

- Scope: `diff -r` against `.temp/TASK-260728-2kp3tv/curator-spec-worktree` reports
  exactly the three owned files. Nothing under `conformance/`, `schemas/`, `release/`,
  `protocol/`, `profiles/` moved.
- rc.5 bytes: `conformance/v1/manifest.json` =
  `9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf`,
  `release/1.0.0-rc.5.json` =
  `b32ee9d35fc2fce0539d0b1c4b15f9c5239115c22fe54a72401f5da6d6646441`, both byte-identical
  to the accepted base and unchanged by two regenerations.
- Schemas 1 through 7, `manager-worker-v1`, both Go driver identities, and every
  cycle-1 through cycle-3 fix are preserved. The object-type and pinned-property
  requirements matched the shipped `buildArtifactV1` as written, so no shipped shape
  moved.
- File digests: decision 0008
  `304152115559a004a2b09dfbbfd9cb85f7d1e9f5525e8108126526115e988725`; `tools/validate.py`
  `d1a8b090c27556b2e1aecc244ac7dfdf37cd841d6dad3b1022bffa2e3e4fefb8`;
  `tools/test_validate.py`
  `2eb03f35a74d1d37cecfd6845bc3c2885a81464052022c13dfa4a87c32732ba4`.

Nothing staged, committed, published, pinned or claimed in the project repository. No
platform or native validation claim is made. The clean-probe commit exists only inside
`.temp/` and is not a project commit.
