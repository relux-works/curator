# Rework cycle 3 — closes both blockers of review verdict cycle 3

Worktree: `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-1jafds/curator-spec-worktree`
Baseline: uncommitted `hardened-1.0.0-rc.1` candidate on top of `57c1f56`.
Nothing staged, committed, or published. No native validation claimed.

## Reproduced first

Both findings were reproduced against the shipped candidate before any edit
(`rework-cycle-3/probe-r3.py`, output in `rework-cycle-3/validation.log`):

| Probe | Before |
|---|---|
| backend version `0` against declared minimum `999999` | schema errors = 0, `check_claim_qualification` **ACCEPTED** |
| linux claim whose TCB observes a Windows kernel | schema errors = 0, `check_claim_qualification` **ACCEPTED** |
| `curator-hardened-component-{file,tree}-v1` | named at `protocol/hardened-execution.md:167,197,198`; **no construction anywhere in the repository** |

## R3-1 — the trusted-component digest algorithms are now constructions

**Normative.** `protocol/hardened-execution.md` gains sections 2.3.1 and 2.3.2.

- `curator-hardened-component-file-v1` — one regular file. Domain prefix, `0x00`,
  then `F` and the length-framed bytes. The path is not hashed; the record around
  it carries `kind` and `name`, and the whole record is inside
  `curator-hardened-tcb-v1`, so a rename still moves the TCB digest. A directory,
  symbolic link, device, FIFO, or socket is not a component file, and an
  implementation MUST NOT resolve one to reach a regular file.
- `curator-hardened-component-tree-v1` — one directory tree. Domain prefix,
  `0x00`, then one length-framed record per entry in unsigned bytewise order over
  the UTF-8 relative path, joined with `/` without normalization or case folding.
  Exactly three entry kinds: `D` (empty payload), `F` (exact bytes), `L` (exact
  `readlink` value). Links must be relative, non-dangling, and resolve within the
  root, and their referents get independent records; hard links are independent
  regular-file records; any other file type invalidates the whole tree rather
  than being skipped. Duplicate encoded paths and platform path collisions are
  invalid.
- Not hashed: mode, ownership, timestamps, ACLs, extended attributes. **Hashed:
  the entry kind** — which is exactly what makes a link substitution a different
  tree.
- Fail-closed: an unreadable, unwalkable, or wrong-type component is
  `hardened_tcb_identity_invalid` before domain entry with no partial digest; a
  component that changes between `tcb-identity-verification` and
  `identity-reverification` is invalid, and components stay byte-for-byte
  unchanged until the last domain member exits; no retry against another path, no
  cached digest, no degrading to naming a component without one.
- Section 2.3.2 closes which kind admits which algorithm:
  `installed-package-tree` is a tree; `helper-executable`, `interpreter`,
  `sandbox-policy-file`, `script`, `shared-library` are files; `capability-probe`,
  `enforcement-adapter`, `identity-verifier` may be either. Enforced by
  `hardenedComponentAlgorithmKindRelationV1` in the schema and by
  `check_tcb_record` in the conformance validator.

**Independently recomputable fixtures.** The identity-separation vector publishes
`component_digest_fixtures`: 10 fixtures carrying the exact bytes, per-field byte
lengths, and expected digest. `tools/generate-hardened` computes them in Go from
the construction; `tools/validate_hardened.py` recomputes every one of them in
Python from the published bytes. Two implementations agreeing is the reproducibility
argument. All 10 digests are distinct, and **no trusted-component digest anywhere in
the suite may come from anywhere but a published fixture** — the previous invented
constants `sha256:3333…`, `4444…`, `5555…`, `8888…` are deleted.

Two structural claims are checked rather than asserted: the empty file and the
empty tree do not share a digest, and the link-substitution fixture really does
hold the referent's exact bytes (a fixture that substitutes different bytes proves
nothing and is rejected).

**Per-facet rotation coverage.** The cycle-3 finding was that
`test_every_mutable_bound_field_has_a_rotation` treated any rotation of the
`trusted_components` array as coverage for kind, name, and algorithm alike. There
are now 8 component rotations, one per mutable facet, each declaring its own
`rotated_component_aspects`:

| Facet | Rotation |
|---|---|
| `kind` | `rotate-trusted-component-kind` (adapter reclassified as `identity-verifier`; both admit the file algorithm, so only the classification changes) |
| `name` | `rotate-trusted-component-name` |
| `algorithm` | `rotate-trusted-component-algorithm` (the probe suite shipped as one file instead of a tree) |
| `content` | `rotate-trusted-component-content` |
| `tree-membership` | `rotate-component-tree-membership` |
| `entry-type` | `rotate-component-entry-type` (a regular file becomes a directory at the same path) |
| `link-substitution` | `rotate-component-link-substitution` |
| `component-set` | `add-trusted-component` |

`validate_tcb_rotation` requires every facet to be covered, requires each named
rotation to actually declare that facet, and rejects a rotation that claims a
facet without rotating `trusted_components`. All 17 rotations produce distinct
TCB digests and distinct cache keys, and each carries the receipt, marker, and
claim rejection against the base.

## R3-2 — qualification relations are enforced

**`hardened-backend-version-v1`** (section 2.3.4). A version is
`series "-" number ( "." number ){0,3}`, where `series` is the token section 6.3
declares for the backend and each `number` is `0` or a nonzero digit followed by
at most eight more. Missing components are zero; components compare as **integers**,
so `cgroup2-6.9` is below `cgroup2-6.10` even though it sorts after it as a string.
Two series are not ordered against each other: that comparison is invalid, not
lower or higher.

- Schema: `hardenedBackendVersionV1` for `tcb.backend.version` and a claim's
  `minimum_version`; `hardenedBackendVersionSeriesRelationV1` and
  `hardenedMinimumVersionSeriesRelationV1` bind the series to the backend in both
  places.
- Conformance validator: `check_claim_qualification` now **compares**, with a
  distinct diagnostic for incomparable versus unsatisfied.
- 13 published comparison cases (above, equal, equal-after-padding, below,
  below-major, the lexical trap, the reviewer's zero-against-999999, incomparable
  series, and four malformed forms) are re-evaluated by the validator rather than
  trusted, and coverage of all five outcomes is required.

**Observed host** (section 2.3.3). `host.kind` is `operating-system` and
`host.identity` is the canonical kernel identity the platform declares — `linux`,
`darwin`, `windows-nt`, tabulated in section 6.3. The `hypervisor` value is
**removed rather than left unenforced**: every backend this revision declares is an
operating-system-kernel mechanism, and a hypervisor-supplied backend would change
which mechanism supplies which capability class, which section 2.2 already requires
to mint a new profile identity and execution-policy identity — a new record version
with its own host contract. `host.version` and `host.build` stay free-form: they
identify, they do not qualify, and nothing compares them.

Section 6.3 now declares the canonical host identity and version series per
platform alongside the backend.

## Self-review found two more holes, both fixed

Writing the mutation probes surfaced two defects in this rework that the shipped
checks would not have caught:

1. **A widened `hardenedHostIdentityValueV1` was undetected**, because the
   platform relation still covered the three declared platforms. That makes the
   value set a backstop nothing proves. New `validate_schema_closed_value_sets`
   pins every closed value set the shipped schemas declare and requires every
   relation to branch on exactly the declared members, so adding a platform or a
   backend without its relation branch now fails.
2. **The version pattern accepted a trailing newline.** `$` matches before a final
   newline in several engines, so `"sandbox-2.0\n"` passed the schema (the
   conformance validator caught it, but the schema alone did not). The pattern now
   ends with `(?![\s\S])`, the protocol states the value carries no surrounding
   whitespace and no trailing newline, and Python, Go, and schema-level tests lock
   it in.

The tree fixtures are also now checked for being trees a walk could have produced:
no absolute or `..` paths, no empty components, every entry's parent present as a
`D` record, and every symbolic link relative, non-dangling, and resolving inside
the root.

## Coverage

| | Before | After |
|---|---|---|
| Python tests | 113 | **151** |
| hardened suite files | 79 | **92** |
| schema cases | 75 | **88** (7 valid, 81 invalid) |
| TCB relations | 8 | **12** |
| TCB completeness cases | 20 | **29** |
| TCB rotations | 13 | **17** (8 per-component-facet) |
| component digest fixtures | 0 | **10** |
| backend version comparison cases | 0 | **13** |

Go tests add fixture distinctness and stability, domain separation of the two
algorithms, the link substitution, order-independence of the tree walk, every
component digest tracing to a fixture, algorithm-to-kind agreement, per-facet
rotation coverage, the version grammar and comparison table, per-platform series
and host agreement, and the example claim actually satisfying its own declared
minimum.

## Mechanical evidence — real exit codes

GREEN in the task worktree: `make validate` 0 (42 portable schemas, 422 portable
vector files, 6 hardened schemas, 92 hardened suite files, 151 Python tests),
`tools/validate.py` 0, `tools/validate_hardened.py` 0, `unittest discover` 0,
`go test -count=1 ./tools/...` 0, `go vet ./tools/...` 0, `gofmt -l tools` empty,
`git diff --check` 0. Hardened suite and release metadata byte-stable across
regeneration.

Adversarial probes: `probe-r3.py` shows both reviewer probes rejected plus two
well-formed below-minimum cases and all 10 fixtures reproduced.
`mutation-probe.py` neuters each of the 7 new schema rules in a sandbox and gets
**7/7 detections with 7 distinct messages**. `validator-mutation-probe.py` runs 6
adversarial instances against the conformance validator and gets **6/6 rejections
with 6 distinct messages**, with the shipped claim still qualifying as a positive
control.

EXPECTED-RED in the task worktree, reported truthfully and structural:
`make regenerate-check` 2, `make regenerate-hardened-check` 2,
`tools/release_gate.py --version 1.0.0-rc.5` 1. All three compare against
committed `57c1f56` or require a clean checkout, while the whole candidate is
uncommitted and the directive forbids staging.

RESOLVED in an ephemeral clean probe byte-identical to the worktree
(`diff -r` empty): `make validate` 0, `make regenerate-check` 0,
`make regenerate-hardened-check` 0, `release_gate.py --version 1.0.0-rc.5` 0 —
so the hardened CI gates do not alter rc.5 qualification. Probe deleted; nothing
staged or committed in curator-spec.

## Preserved

- Cycle-1 identity and ordering fixes, cycle-2 TCB completeness, the 17-phase
  authoritative list, package-influence exclusions, capability evidence as the
  only result-only identity.
- `conformance/v1`, `schemas/v1`, and `release/1.0.0-rc.5.json` are `diff -r`
  identical to the accepted predecessor after regenerating with
  `generate-vectors`; portable manifest sha256
  `9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf` equals the
  rc.5 pin.

## Not claimed

No platform is qualified. `native_evidence` is absent on all three declarations,
`claims_emitted` and `qualified_platforms` are empty, every adversarial case is
`pending-native-validation`, and no native test was run on any platform. Nothing
staged, committed, or published.

## Files touched

Normative: `protocol/hardened-execution.md`, `profiles/manager-hardened.md`.
Schemas: `schemas/hardened/v1/hardened-common.schema.json`,
`hardened-conformance-claim-v4.schema.json`.
Tools: `tools/generate-hardened/main.go` (+`main_test.go`),
`tools/validate_hardened.py` (+`test_validate_hardened.py`).
Regenerated: `conformance/hardened/v1/**`, `release/hardened-1.0.0-rc.1.json`.
Prose: `CHANGELOG.md`, `COMPATIBILITY.md`, `SECURITY.md`,
`decisions/0009-hardened-build-execution-profile.md`,
`docs/hardened-build-execution-profile.md`.
