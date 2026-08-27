# Rework 02 — cycle-2 blockers closed

Scope: only the two blocking findings in `TASK-260728-2spy93_review-verdict-cycle-2.md`.
All cycle-1 repairs, frozen v1/v6/v7 shapes, `manager-worker-v1` semantics and rc.5
bytes are preserved. Still exactly three files differ from the accepted
`TASK-260728-2kp3tv` base: `decisions/0008-additional-language-driver-boundary.md`,
`tools/validate.py`, `tools/test_validate.py`.

## Blocker 1 — the reserved-shape gate false-accepted an unstructured `toolchain`

Accepted as reported. The member-set table answered *which* names a definition may
carry and said nothing about what they mean, so `{"type":"string"}` in the
`toolchain` slot of `buildCommandV8`, `repositoryBuildCommandV2` and
`skillBuildTargetV2` passed. The test helper minted exactly that shape and the
positive test accepted it.

Fix — placement is now structural, not nominal:

- Decision 0008 fixes the placement form. The requirement object is one shared
  definition, `toolchainRequirementV1`, allocated in section 1 alongside the three
  reserved wire shapes. In every slot that carries it the property schema MUST be
  exactly `{"$ref":"#/$defs/toolchainRequirementV1"}` with no sibling keyword, and
  the definition MUST be exactly `id` and `version`, both REQUIRED,
  `additionalProperties` false. Sections 4, 5, 10, 11 and the security-impact
  section were updated to match; the version grammar inside `version` stays decision
  0007's and is still not restated. Decision 0007 names the object but fixes no
  `$defs` identifier, so naming it here is boundary ownership ("where it lands"),
  not a redefinition — the two members enforced are the two members 0007 itself
  states.
- `tools/validate.py` gains `EXACT_PROPERTY_SCHEMAS`, `check_exact_property_schemas`
  and `check_toolchain_requirement_definition`. The property schema is compared for
  exact equality, and a reserved slot whose reference resolves to nothing is a
  failure. A reserved shape minted *without* a `driver` is now rejected too: it
  would not be driver-bearing, so it would previously have escaped both the
  member-set table and the property check.

Independent CLI probe (plant, run `python3 tools/validate.py`, revert):

```
PROBE A (buildCommandV8 toolchain as {"type":"string"}) exit=1
validation failed: common.schema.json $defs.buildCommandV8.toolchain is not exactly
{'$ref': '#/$defs/toolchainRequirementV1'}: found {'type': 'string'}
common.schema.json restored byte-identical: yes; post-revert validate.py exit=0
```

In-process probe over all three slots (`logs/probe-rework-02.log`):

| slot | string | open object | path-bearing | wrong ref | exact ref | missing definition |
|---|---|---|---|---|---|---|
| `buildCommandV8` | rejected | rejected | rejected | rejected | accepted | rejected |
| `repositoryBuildCommandV2` | rejected | rejected | rejected | rejected | accepted | rejected |
| `skillBuildTargetV2` | rejected | rejected | rejected | rejected | accepted | rejected |

Red tests added: `test_reserved_toolchain_slot_rejects_every_shape_but_that_reference`
covers six malformed shapes × three slots (string, open object, path-bearing object,
wrong `$defs` reference, the resolved `goToolchainIdentityV1` identity instead of the
gate, and a `$ref` with a sibling keyword);
`test_reserved_slot_cannot_reference_a_missing_requirement_definition`;
`test_requirement_definition_is_closed_to_id_and_version` (extra member, open object,
optional `version`); `test_reserved_shape_minted_without_a_driver_is_rejected`;
`test_reserved_toolchain_slot_carries_the_exact_requirement_reference` (positive).
The `_minted` helper now mints the exact reference and the requirement definition,
so the previously false-accepting positive test is genuinely positive.

## Blocker 2 — claim-v4 admission was unconditional despite a retirement path

Accepted as reported. Section 8 said claim schema 4 admits all eight identifiers
while section 2 and the Kotlin obligation allow a rejected contract to retire its
identifiers unused, so a rejected Kotlin contract forced `TASK-260728-251p01` either
to admit a retired identifier or to violate section 8.

Fix — membership is defined over the admitted set, not over a count:

- Section 8 now states claim schema 4 admits driver assertions for exactly the
  admitted wire driver set **as it stands when claim schema 4 is minted**, no fewer
  and never one more; eight identifiers if and only if all six reserved contracts are
  accepted; six if both Kotlin identifiers are retired. A retired identifier is
  structurally unassertable, not merely unclaimed. The per-driver `execution_policy`
  `const` is retained whatever the admitted set turns out to be
  (`manager-worker-v1` for the two Go drivers, `manager-worker-v2` for each admitted
  reserved driver). Claim schemas 1–3 stay valid because they assert a subset of a
  set that only grows by an accepted contract. Section 10 now tells integration to
  mint claim schema 4 over exactly the identifiers it moves in that same change.
- `tools/validate.py` gains `DRIVER_EXECUTION_POLICIES` (all eight identifiers, one
  policy each), `conformance_claim_schemas`, `closed_const` and
  `check_claim_driver_admission`. Every claim schema that asserts build drivers must
  assert only admitted identifiers, must pair each driver with the policy the closed
  table binds to it, must keep each assertion closed to its four members with
  `additionalProperties: false` and no duplicate driver, and the current (highest)
  claim schema must assert every admitted driver. This is reachable today against
  the shipped claim v3 and self-adjusts when integration admits a driver.

Probes:

```
reserved driver asserted   -> rejected: ... asserts 'kotlin-native-repository-v1',
                              which is not in the admitted wire driver set
mispaired policy           -> rejected: ... pairs 'go-v1' with execution policy
                              'manager-worker-v2' instead of 'manager-worker-v1'
admitted driver dropped    -> rejected: the current claim schema does not assert
                              every admitted wire driver: missing ['go-repository-v1']
unchanged (positive)       -> accepted
PROBE C (extra optional assertion member, real CLI) exit=1
validation failed: schemas/v1/conformance-claim-v3.schema.json: build_drivers assertion
is not the closed assertion member set ['driver','execution_policy','language','operating_systems']
conformance-claim-v3.schema.json restored byte-identical: yes; post-revert exit=0
```

Red tests added: `test_current_claim_schema_asserts_exactly_the_admitted_driver_set`,
`test_a_frozen_claim_may_assert_a_subset_but_the_current_one_may_not`,
`test_claim_schema_cannot_assert_a_reserved_or_retired_driver` (all six reserved
identifiers), `test_claim_assertion_must_pair_its_driver_with_the_closed_policy`
(mispaired const, free-form string, reference outside the shared definitions),
`test_claim_assertion_member_set_is_closed` (extra member, open object, duplicate
assertion), `test_driver_policy_table_binds_every_identifier_exactly_once`.

## Gate evidence — real exit codes, each command run standalone

Task worktree (`.temp/TASK-260728-2spy93/curator-spec-worktree`), task venv:

| command | exit |
|---|---|
| `python3 tools/validate.py` (42 schemas, 422 vector files) | 0 |
| `python3 -B -m unittest discover -s tools -p 'test_*.py'` (71 tests, was 60) | 0 |
| `go test ./tools/...` | 0 |
| `go vet ./tools/...` | 0 |
| `gofmt -l tools` (0 files listed) | 0 |
| `python3 -m compileall -q tools` | 0 |
| `git diff --check` | 0 |

Clean-checkout probe (`.temp/TASK-260728-2spy93/clean-probe-02`, commit `c12d0d8`, a
throwaway local repo used only because `release_gate.py` needs a HEAD — nothing in
the project repository was staged, committed, pushed or pinned):

| command | exit |
|---|---|
| `make validate` | 0 |
| `make regenerate-check` (run 1) | 0 |
| `make regenerate-check` (run 2) | 0 |
| `make release-check VERSION=1.0.0-rc.5` | 0 |

rc.5 preserved: `conformance/v1/manifest.json` digest
`sha256:9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf` unchanged,
`release/1.0.0-rc.5.json` unchanged (`b32ee9d3…`), two regenerations byte-identical,
release gate passed at the probe commit.

## Honest limitations

- The reserved wire shapes and the requirement definition still do not exist on disk;
  they are enforced the moment they are minted. The exact-reference and requirement
  checks are therefore exercised by unit tests and by plant-and-revert probes, not by
  a shipped schema. Probe A shows the CLI path is real, not test-only.
- The reserved-schema-slot guard remains unreachable through the CLI for claim v4
  specifically (the slot must not exist while the reservation stands), so the
  claim-admission gate is proved against the shipped claim v3 plus a synthetic newer
  sibling in `test_a_frozen_claim_may_assert_a_subset_but_the_current_one_may_not`.
- No platform, native-execution or signing claim is made or implied by this rework.
- Nothing was staged, committed, published or pinned in the project repository; the
  predecessor worktree was not touched.
