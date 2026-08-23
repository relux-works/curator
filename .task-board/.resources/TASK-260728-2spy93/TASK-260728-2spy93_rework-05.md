# TASK-260728-2spy93 rework 05

Closes the single blocking finding in `TASK-260728-2spy93_review-verdict-cycle-5.md`
(P1: the shared artifact targets can still be widened without rejection). Nothing else
was reopened — the six driver identities, version allocation, local/external ownership,
toolchain placement, `manager-worker-v2` split, claim admission closure, downstream
traceability, and every cycle-1 through cycle-4 repair are untouched. Still exactly
three files different from the accepted `TASK-260728-2kp3tv` base:
`decisions/0008-additional-language-driver-boundary.md` (1082 lines, was 1049),
`tools/validate.py`, `tools/test_validate.py`.

## The finding, reproduced first

`check_exact_property_schemas` pinned `path`, `sha256`, and `size` to the three `$ref`
values, and `check_build_artifact_rejections` sampled invalid instances. Neither held the
*referenced definitions* to a shape, so a one-keyword widening of a target left the
`$ref` intact, left every sampled negative still rejected, and left every generated
positive case valid — while the compiled `build-receipt-v2` validator started accepting
an artifact the boundary does not admit.

Reproduced before touching the gate with `.temp/TASK-260728-2spy93/repro-cycle-5.py`
(exit 0 = all three reproduce against the submitted cycle-4 gate):

| Mutant | Cycle-4 gate | Shipped receipt | Mutated receipt | Positive case |
| --- | --- | --- | --- | --- |
| `portablePath.maxLength` 4096 -> 4097, path = 4097 `a` | accepted | rejects | **accepts** | still valid |
| `sha256.pattern` admits uppercase hex, digest = 64 `A` | accepted | rejects | **accepts** | still valid |
| `nonNegativeSafeInteger.maximum` + 1, size = 2^53 | accepted | rejects | **accepts** | still valid |

The reviewer's reading was exact: the widening is invisible to a `$ref` comparison and
survives every finite sample of bad values.

## Fix — option 1, structural pin

The review recommended pinning the exact current schemas; that is what was done, because
these are three small closed schemas and an equivalence/subsumption checker over
Draft 2020-12 would be a new, unaudited proof surface.

`tools/validate.py`:

- `ARTIFACT_REFERENCE_TARGETS` states the exact canonical schema of `portablePath`,
  `sha256`, and `nonNegativeSafeInteger`, keyword for keyword. The two regex literals are
  the schema text verbatim — Python and JSON agree on every escape used, so the table
  diffs against `schemas/v1/common.schema.json` character for character — and the size
  ceiling reuses the module's existing `SAFE_INTEGER`.
- `check_artifact_reference_targets` runs from `check_build_artifact_closure` and does
  two things. First it holds the gate's own tables together: every `$ref` in
  `ARTIFACT_PROPERTY_SCHEMAS` must be a local `#/$defs/` reference, and the set of names
  it references must equal the set of pinned names — so a member repointed at an unpinned
  definition, or a pin left behind for a definition the artifact no longer names, is a
  failure rather than a silent gap. Then it compares each pinned definition to the
  shipped one and reports the difference keyword by keyword.
- `schema_difference` produces that account (`maxLength changed from 4096 to 4097`,
  `pattern removed, was ...`, `expected a schema object, found None`). Printing both
  schemas whole would bury a one-digit bound change inside the path grammar next to it.
- Comments and docstrings that credited the behavioural proof with holding the three
  targets in place were corrected. `check_build_artifact_rejections` now states what it
  actually is: a deliberately finite behavioural layer that catches a target opened
  wholesale, complementary to the structural pin rather than a substitute for it.

Scope note: only the artifact's own reference targets are pinned. `toolchainRequirementV1`
is unminted and its internals belong to decision 0007 / `TASK-260728-1g0z69`; this
boundary continues to require that it exists, is an object schema, and is closed to
exactly `id` and `version`, and claims nothing more. That limit is stated in the code.

## Fix — decision 0008

The false claim was in the decision as well as in the gate, so both were corrected.

- Section 3 gained a normative paragraph: the three shared definitions MUST each remain
  exactly the schema the frozen rc.5 corpus ships; the path grammar, digest alphabet, and
  size ceiling are artifact identity rather than illustrative examples; a single keyword
  moves any of them while the `$ref`, the positive cases, and every sampled rejection are
  unaffected; enforcement MUST therefore be structural, and a proof by sampled rejections
  MUST NOT be presented as covering them.
- Section 11 item 11 no longer claims the compiled-validator proof keeps the three
  definitions from being widened underneath the pinned references. It now states the
  structural pin as the closure and describes the behavioural proof as a separate layer
  that is finite by construction.
- Security impact gained the consequence, stated concretely: an uppercase-admitting
  digest alphabet makes one artifact two identities under byte comparison, a widened path
  grammar stops describing where a published file may live, and a lifted size range
  records a size CCJ-1 cannot represent.

## The three mutants, before and after

`.temp/TASK-260728-2spy93/mutant-probe.py` now carries cycles 3, 4, and 5 — eight mutants.
It restores five checkers to their pre-fix bodies, runs each mutant against the restored
and the submitted gate, and re-proves each Draft 2020-12 escape. Exit 0 = every mutant
accepted before its fix and rejected after. It writes no project file.

```
[ok] pre-fix  | portablePath.maxLength 4096 -> 4097: ACCEPTED
[ok] post-fix | portablePath.maxLength 4096 -> 4097: REJECTED native-executable-v1 identity target
               common.schema.json $defs.portablePath is not its canonical schema, so the pinned
               reference no longer bounds what it names: maxLength changed from 4096 to 4097
[ok] draft2020-12 | shipped build-receipt-v2 rejects the artifact it admits (1 errors), mutated
               variant accepts it (0 errors), positive case still validates: True
[ok] pre-fix  | sha256.pattern admits uppercase hex: ACCEPTED
[ok] post-fix | sha256.pattern admits uppercase hex: REJECTED ... $defs.sha256 ... pattern changed
[ok] draft2020-12 | ... mutated variant accepts it (0 errors), positive case still validates: True
[ok] pre-fix  | nonNegativeSafeInteger.maximum + 1: ACCEPTED
[ok] post-fix | nonNegativeSafeInteger.maximum + 1: REJECTED ... maximum changed from
               9007199254740991 to 9007199254740992
[ok] draft2020-12 | ... mutated variant accepts it (0 errors), positive case still validates: True
PROBE PASS   (exit 0)
```

The cycle-3 and cycle-4 lines still pass unchanged in the same run.

## Regression tests (91 total, was 83)

New:

- `test_pinned_targets_are_the_shipped_shared_definitions` — the pin is the shipped text,
  and the pinned set is exactly the three identity targets.
- `test_artifact_identity_target_cannot_be_widened_by_one_keyword` — the three reviewer
  mutants, each rejected with the class, the target, and the moved keyword named.
- `test_artifact_identity_target_cannot_drop_a_bound` — `maxLength`, `not`, `pattern`, and
  `maximum` removed; removal widens the same way raising a bound does.
- `test_artifact_identity_target_cannot_disappear` — each of the three definitions deleted.
- `test_every_artifact_reference_must_be_pinned` — a member repointed at an unpinned
  definition, and a pin kept for a definition the artifact no longer references.
- `test_artifact_member_pinned_outside_the_shared_definitions` — a member pinned to an
  absolute URL the boundary cannot hold to a schema.
- `test_one_keyword_widening_survives_every_sampled_rejection` — the honest negative: each
  mutant passes `check_build_artifact_rejections` untouched. This is why the structural
  pin is required rather than more samples.
- `test_one_keyword_widening_lets_the_real_validator_accept_it` — and why it is not style:
  under each mutant the compiled `build-receipt-v2` validator accepts the over-long path,
  the uppercase digest, and the out-of-range size, with the positive case unaffected.

Updated: `test_artifact_reference_targets_cannot_be_widened_underneath` (the coarse
wholesale-opening cases) now asserts both layers — the structural pin rejects the mutant
through the full gate, and `check_build_artifact_rejections` still rejects it
independently when invoked directly.

Red-before-green was verified rather than asserted.
`.temp/TASK-260728-2spy93/red-check-cycle-5.py` stubs `check_artifact_reference_targets`
back to the cycle-4 no-op and runs the artifact tests (exit 0):

```
[ok] pre-fix red: test_artifact_identity_target_cannot_be_widened_by_one_keyword -> FAILED as required
[ok] pre-fix red: test_artifact_identity_target_cannot_disappear             -> FAILED as required
[ok] pre-fix red: test_artifact_identity_target_cannot_drop_a_bound          -> FAILED as required
[ok] pre-fix red: test_artifact_reference_targets_cannot_be_widened_underneath -> FAILED as required
[ok] pre-fix red: test_every_artifact_reference_must_be_pinned               -> FAILED as required
[ok] pre-fix red: test_artifact_member_pinned_outside_the_shared_definitions -> FAILED as required
[ok] pre-fix green: test_pinned_targets_are_the_shipped_shared_definitions   -> PASSED
[ok] pre-fix green: test_one_keyword_widening_survives_every_sampled_rejection -> PASSED
[ok] pre-fix green: test_one_keyword_widening_lets_the_real_validator_accept_it -> PASSED
RED CHECK PASS
```

The last three are green either way by design: they document the escape and the limit of
the behavioural layer, so they must not depend on the fix.

## Plant-and-revert CLI probes — with one honest limitation

`.temp/TASK-260728-2spy93/plant-probe-cycle-5.py` writes each widening into
`schemas/v1/common.schema.json`, runs the real CLI, then asks each earlier validation
layer in isolation about the same planted file, and restores the original bytes.
Pre-plant and post-revert SHA-256 of the file are identical
(`4c725692979f64104664f51391f7f99b63d84dedc350c34d92a9fa0e3a620190`), and
`python tools/validate.py` exits 0 after the last revert. Probe exit 0.

| Plant | CLI exit | First layer to reject |
| --- | --- | --- |
| `portablePath.maxLength` 4096 -> 4097 | 1 | boundary gate: `$defs.portablePath is not its canonical schema ... maxLength changed from 4096 to 4097` |
| `sha256.pattern` admits uppercase hex | 1 | boundary gate: `$defs.sha256 is not its canonical schema ... pattern changed` |
| `nonNegativeSafeInteger.maximum` + 1 | 1 | **`validate_schemas`**, not this fix — see below |
| `nonNegativeSafeInteger.maximum` removed | 1 | boundary gate: `$defs.nonNegativeSafeInteger is not its canonical schema ... maximum removed` |

**The limitation, stated plainly.** The reviewer's third mutant raises `maximum` to
`9007199254740992`, which is outside the CCJ-1 safe integer range, and `load_json` refuses
to parse an integer literal that large. On disk that mutant therefore never reaches the
boundary gate: it is rejected by a pre-existing parse guard, and claiming this fix as the
rejecting layer would be false. The mutant is still real in memory — which is exactly how
the reviewer demonstrated it, and how `repro-cycle-5.py`, `mutant-probe.py`, and
`test_artifact_identity_target_cannot_be_widened_by_one_keyword` exercise it — and the new
pin is what rejects it there. The fourth plant expresses the same widening in the form the
parse guard cannot see (`maximum` removed, so every integer is admitted), and there the
boundary gate is the only layer that rejects.

For the first, second, and fourth plants, `validate_schemas`, `validate_manifest`, and
`validate_vector_semantics` were each run in isolation on the planted file and all three
passed. `validate_schemas` runs before the boundary gate in `main()`, so this is not an
ordering artifact. (`validate_manifest` digests the conformance suite, not
`schemas/v1/`, which is why a common-schema edit is invisible to it.)

## Gates (standalone runs, real exit codes)

Task worktree `.temp/TASK-260728-2spy93/curator-spec-worktree`, venv `validation-venv`:

| Command | Exit |
| --- | --- |
| `python tools/validate.py` (42 schemas, 422 vector files) | 0 |
| `python -B -m unittest discover -s tools -p 'test_*.py'` (91 tests) | 0 |
| `go test ./...` | 0 |
| `go vet ./tools/...` | 0 |
| `gofmt -l .` (no output) | 0 |
| `python -m compileall -q tools` | 0 |
| `git diff --check` | 0 |
| `python .temp/TASK-260728-2spy93/mutant-probe.py` (8 mutants) | 0 |
| `python .temp/TASK-260728-2spy93/red-check-cycle-5.py` | 0 |
| `python .temp/TASK-260728-2spy93/plant-probe-cycle-5.py` | 0 |

Clean-checkout probe `.temp/TASK-260728-2spy93/clean-probe-05` at commit `5bd8e2c` (fresh
tree, fresh repo, no history from the task worktree, task validation environment on
`PATH`):

| Command | Exit |
| --- | --- |
| `make validate` | 0 |
| `make regenerate-check` (first) | 0 |
| `make regenerate-check` (second) | 0 |
| `make release-check VERSION=1.0.0-rc.5` | 0 |

`git status --porcelain` is empty after all four, so regeneration is deterministic and
moved nothing.

## Preservation

- Scope: `diff -rq` against `.temp/TASK-260728-2kp3tv/curator-spec-worktree` reports
  exactly the three owned files. Nothing under `conformance/`, `schemas/`, `release/`,
  `protocol/`, `profiles/` moved; `schemas/v1/common.schema.json` is byte-identical to the
  accepted base after the plant probes.
- rc.5 bytes: `conformance/v1/manifest.json` =
  `9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf`,
  `release/1.0.0-rc.5.json` =
  `b32ee9d35fc2fce0539d0b1c4b15f9c5239115c22fe54a72401f5da6d6646441` — both identical to
  the accepted base and unchanged by two regenerations.
- Schemas 1 through 7, `manager-worker-v1`, both Go driver identities, and Go semantics
  are unchanged. The pinned schemas are the shipped ones as written, so no shipped shape
  moved.
- File digests: decision 0008
  `5f04f333c53dd552eee334a1f548bcec2c627acac914233448bff86bfabb0056`; `tools/validate.py`
  `2e3a08dec4a66278af79d1690888c21b565dbc1c3c8b17eaa6531a345f3480ef`;
  `tools/test_validate.py`
  `4b9f1e407345bc9dd7cadc285cc75428945d2ea38cd829eb133971c0ba4ac7e8`.

Nothing staged, committed, published, pinned, or claimed in the project repository. No
platform or native validation claim is made. The clean-probe commit exists only inside
`.temp/` and is not a project commit.
