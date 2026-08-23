# TASK-260728-2spy93 rework 03

Closes the two blocking findings in `TASK-260728-2spy93_review-verdict-cycle-3.md`.
Nothing else in the decision was reopened. Still exactly three files different from
the accepted `TASK-260728-2kp3tv` base: `decisions/0008-additional-language-driver-boundary.md`
(1009 lines), `tools/validate.py`, `tools/test_validate.py`.

## Finding 1 — `toolchainRequirementV1` could be a non-object schema

`check_toolchain_requirement_definition` delegated to `check_closed_member_set`,
which asserted `properties`, `required` and `additionalProperties: false` and never
asserted the type. Those three keywords constrain objects only, so a definition
declared `type: string` kept the exact closed member set while accepting whatever
text the package put in the slot.

Fix: the object type is now part of the closure itself, in `check_closed_member_set`,
so every closed member-set definition — deployed, reserved and manager-authored —
must declare `"type": "object"` exactly. String types, an omitted type, and an
`["object","string"]` union all reject. All nine shipped closed definitions already
declared `type: object`, so no shipped shape moved.

Decision text made precise, not reopened:

- section 4 placement paragraph: the definition MUST be an object schema, `"type": "object"`
  exactly, never omitted and never unioned, with the reason stated;
- section 11 item 6: the object type is part of the exact member-set table;
- section 11 item 8: the referenced definition must exist, be an object schema, and
  be closed to `id` and `version`;
- Security impact: a definition keeping the two members while declaring itself a
  string, no type, or an object-or-string union is named as reopening the surface
  that paragraph claims does not exist.

## Finding 2 — claim admission was not closed at the container

Two separate holes, both closed.

**Currency was decided after filtering.** The loop skipped every claim schema without
`build_drivers` and then took the last survivor as current, so a newest claim that
simply dropped the member handed currency back to a frozen predecessor and asserted
nothing while the gate reported the older schema as covering the admitted set. The
newest schema is now chosen from `conformance_claim_schemas()` before anything is
inspected, and the current schema MUST declare the member.

**`items.oneOf` was not the only element path.** Under Draft 2020-12 `items` applies
only to elements `prefixItems` did not cover, so one `prefixItems` entry exempts
element zero from the closed `oneOf` entirely. New `check_claim_driver_container`
closes the container to keywords that cannot reach an element (`type`, `items`,
`minItems`, `maxItems`, `uniqueItems`, `description`, `title`); `prefixItems`,
`contains`, `unevaluatedItems`, `additionalItems` and any applicator added later
reject as unlisted rather than being interpreted. `items` must be exactly `{"oneOf": [...]}`.

Each assertion is now also closed by keyword set (`type`, `properties`, `required`,
`additionalProperties`) and must declare `type: object` — the same failure class as
finding 1, since without it a bare list element escapes every member check below.

Decision section 8 gained the assertion-list closure paragraph and the object-schema
requirement on an assertion; section 11 item 9 restates currency-before-inspection,
the container closure, and the rejected parallel applicators.

## The three mutants, before and after

`.temp/TASK-260728-2spy93/mutant-probe.py` restores both checkers to their cycle-2
bodies and runs the reviewer's exact mutants. Exit 0 = every mutant accepted before,
rejected after. It writes no project file.

```
[ok] pre-fix  | string toolchainRequirementV1: ACCEPTED
[ok] post-fix | string toolchainRequirementV1: REJECTED ... is not an object schema, so its closed member set constrains nothing: found type 'string'
[ok] pre-fix  | newest claim drops build_drivers: ACCEPTED
[ok] post-fix | newest claim drops build_drivers: REJECTED ... the current claim schema does not declare build_drivers, so it asserts no admitted wire driver
[ok] pre-fix  | kotlin-native-repository-v1 via prefixItems: ACCEPTED
[ok] post-fix | kotlin-native-repository-v1 via prefixItems: REJECTED ... build_drivers declares array keywords outside the closed container set: ['prefixItems']
[ok] draft2020-12 | shipped claim rejects kotlin-native-repository-v1 (1 errors), prefixItems variant accepts it (0 errors)
PROBE PASS   (exit 0)
```

The last line is the point of the second finding: the escape is not a reading of the
checker, it is what the compiled Draft 2020-12 validator does with the shipped claim
schema plus one `prefixItems` entry.

## Regression tests (76 total, was 71)

- `test_requirement_definition_must_be_an_object_schema` — string type, omitted type,
  `["object","string"]` union; also asserts against the real validator that the
  closed members say nothing about a string.
- `test_current_claim_schema_cannot_drop_the_driver_member` — synthetic newest claim
  v99 with the member removed.
- `test_reserved_driver_cannot_be_reached_by_a_parallel_array_applicator` — proves the
  `prefixItems` escape against the compiled validator, then rejects `prefixItems`,
  `contains`, `unevaluatedItems`, `additionalItems` on a claim version that exists on
  no disk, where no other check can reach it.
- `test_claim_driver_container_must_stay_an_items_only_array` — non-array container,
  open `items`, `items` with a sibling applicator, empty assertion list.
- `test_claim_assertion_without_an_object_type_is_rejected` — untyped assertion and an
  assertion carrying a keyword outside the closed set.

## Plant-and-revert CLI probes

Every planted file was restored and verified byte-identical by SHA-256, and
`validate.py` exits 0 after each revert.

| Plant | Command | Exit | Rejected by |
| --- | --- | --- | --- |
| `toolchainRequirementV1` minted with `type: string` | `python tools/validate.py` | 1 | boundary gate: `is not an object schema ... found type 'string'` |
| the same definition minted with `type: object` | `python tools/validate.py` | 0 | accepted — the check is not vacuously rejecting |
| claim v3 `prefixItems` naming `swift-v1` | `python tools/validate.py` | 1 | reserved-name absence scan, at `schemas/v1/conformance-claim-v3.schema.json:211` |
| claim v3 `prefixItems` with an open, policy-free assertion | `python tools/validate.py` | 1 | schema-case corpus (`invalid-hardened-execution-policy.json` started validating) |
| the same plant, boundary gate in isolation, cycle-2 checker | `validate_additional_driver_boundary()` | 0 | nothing — the pre-fix gate was blind |
| the same plant, boundary gate in isolation, shipped checker | `validate_additional_driver_boundary()` | 1 | `build_drivers declares array keywords outside the closed container set: ['prefixItems']` |

**Honest limitation, unchanged in kind from rework 02.** On-disk plants against the
shipped claim v3 are caught by earlier layers — the reserved-name absence scan when
the plant names a reserved driver, the frozen schema-case corpus when it widens an
assertion — so the container closure is not what rejects them at the CLI. Rows 5 and
6 isolate the boundary gate on the same planted file to show the difference the fix
makes. The situation the closure actually owns is a *newly minted* claim version whose
identifiers and negative cases are authored by the same change, which is reachable only
by unit test and by the in-memory probe until `TASK-260728-251p01` mints claim v4. The
reserved shapes and `toolchainRequirementV1` still do not exist on disk for the same
reason; the row-1/row-2 plants are the closest on-disk approximation.

## Gates (standalone runs, real exit codes)

Task worktree `.temp/TASK-260728-2spy93/curator-spec-worktree`, venv `validation-venv`:

| Command | Exit |
| --- | --- |
| `python tools/validate.py` (42 schemas, 422 vector files) | 0 |
| `python -m unittest discover -s tools -p 'test_*.py'` (76 tests) | 0 |
| `go test ./...` | 0 |
| `go vet ./tools/...` | 0 |
| `gofmt -l .` (no output) | 0 |
| `python -m compileall -q tools` | 0 |
| `git diff --check` | 0 |
| `python .temp/TASK-260728-2spy93/mutant-probe.py` | 0 |

Clean-checkout probe `.temp/TASK-260728-2spy93/clean-probe-03` at commit `86ec275`
(fresh tree, fresh repo, no history from the task worktree):

| Command | Exit |
| --- | --- |
| `make validate` | 0 |
| `make regenerate-check` (first) | 0 |
| `make regenerate-check` (second) | 0 |
| `make release-check VERSION=1.0.0-rc.5` | 0 |

## Preservation

- Scope: `diff -r` against `.temp/TASK-260728-2kp3tv/curator-spec-worktree` reports
  exactly the three owned files. Nothing under `conformance/`, `schemas/`, `release/`,
  `protocol/`, `profiles/` moved.
- rc.5 bytes: `conformance/v1/manifest.json` = `9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf`,
  `release/1.0.0-rc.5.json` = `b32ee9d35fc2fce0539d0b1c4b15f9c5239115c22fe54a72401f5da6d6646441`,
  both byte-identical to the accepted base and unchanged by two regenerations.
- Schemas 1 through 7, `manager-worker-v1`, and both Go driver identities are untouched;
  the object-type requirement matched all nine shipped closed definitions as written.
- Decision 0008 SHA-256: `36f98739f699580273aeea4768f64aa48925250e4e928db4cd1b3deded1be7de`.

Nothing staged, committed, published, pinned or claimed. No platform or native
validation claim is made.
