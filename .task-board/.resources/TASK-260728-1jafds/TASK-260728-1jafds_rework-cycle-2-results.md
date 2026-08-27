# TASK-260728-1jafds — rework cycle 2

Closes the single blocking finding of
`TASK-260728-1jafds_review-verdict-cycle-2.md`. Ready for review cycle 3.

Worktree: `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-1jafds/curator-spec-worktree`.
Nothing staged, committed, or published. No native adversarial test was run on
any platform, and no platform is qualified.

## R2-1: hardened TCB identity was structurally incomplete

The reviewer's finding was reproduced first, against the real schemas, before
any change:

| Adversarial instance | Schema errors, before |
|---|---|
| macOS TCB carrying the Linux enforcement backend, in a receipt | 0 |
| `darwin` build target with a `linux` TCB, in a receipt | 0 |
| trusted component named `"mutable-interpreter-with-no-cryptographic-identity"` | 0 |
| macOS TCB carrying the Windows enforcement backend, in a claim | 0 |
| manager parent identity present in the record | absent |
| observed OS/hypervisor identity present in the record | absent |
| enforcement-backend version present in the record | absent |

All four instances now produce errors, and all three identities are present.
The same probe is checked in as
`TcbCompletenessTests.test_review_cycle_2_probes_are_rejected`.

## What changed

### 1. `hardened-tcb-v1` is now the complete closed record

`schemas/hardened/v1/hardened-common.schema.json`. Twelve members, closed,
`additionalProperties: false`:

| Member | Identifies |
|---|---|
| `record_version`, `hardened_profile`, `execution_policy` | the contract; fixed by this revision |
| `platform` | the hardened platform the operation ran on |
| `enforcement_backend` | the concrete mechanism that supplied the capability classes |
| `backend` | **new** — observed backend `version` plus `configuration`, a set of closed `{setting, observed_value}` records |
| `host` | **new** — observed `{kind, identity, version, build}`, where `kind` is `operating-system` or `hypervisor` |
| `parent_sha256` | **new** — the installed manager parent bytes |
| `supervisor_sha256` | the hardened supervisor bytes |
| `worker_sha256` | the domain-root worker bytes |
| `toolchain` | the fingerprinted `go` launcher and `GOROOT` tools |
| `trusted_components` | **replaces** `additional_trusted_components`; closed `{kind, name, algorithm, content_sha256}` records, sorted and unique |

`additional_trusted_components` — an array of unconstrained strings — is gone,
and a schema case proves the field cannot be reintroduced. `kind` is one of nine
enumerated component kinds (`interpreter`, `installed-package-tree`, `script`,
`shared-library`, `sandbox-policy-file`, `helper-executable`,
`enforcement-adapter`, `capability-probe`, `identity-verifier`); `algorithm` is
`curator-hardened-component-file-v1` or `curator-hardened-component-tree-v1`.

The normative TCB claim was **not** narrowed. The observed host and backend
version/configuration on which qualification depends are bound, per the first
branch of the directive.

### 2. The three closed relations are enforced by the schemas

- **platform to backend** — `hardenedPlatformBackendRelationV1`, referenced from
  the TCB record and from the capability-evidence record. One platform admits
  exactly one backend.
- **operating system to backend** — `hardenedOperatingSystemBackendRelationV1`,
  referenced from each `enforcement_backends` entry of claim v4.
- **target to platform** — `hardenedTargetPlatformRelationV1`, referenced from
  receipt v3 and v4: `darwin`→`macos`, `linux`→`linux`, `windows`→`windows`. A
  hardened build input's `goos` is additionally narrowed to those three, so the
  relation is total rather than partial.

A claim's `required_configuration` changed from free-text strings to closed
`{setting, required_value}` records, so the qualification a claim declares can
be checked against the base it names.

### 3. Two further relations are enforced by the conformance validator

Not expressible in JSON Schema, so stated normatively and checked mechanically:

- a claim's `tcb.platform` must be an operating system the claim declares, with
  exactly the backend that claim declares for it;
- every `required_configuration` setting of that entry must appear in
  `tcb.backend.configuration` with exactly the required value.

`tcb_completeness.relations` in the identity vector labels each relation
`schema` or `conformance-validator`, and `validate_tcb_schema_relations` runs
every `schema`-labelled relation against the real schemas — so a schema that
silently stops enforcing one fails the validator instead of passing on prose.
This was verified by mutating the shipped schemas in a sandbox: dropping the
platform-to-backend relation, the target-to-platform relation, the claim
OS-to-backend relation, `required: parent_sha256`, or `required: host` each
produced a distinct validator failure.

### 4. Every bound identity is rotated, and every rotation moves the key

`conformance/hardened/v1/vectors/hardened-identity-separation.json` gains
`tcb_completeness` (twelve bound fields, eight relations, the
narrower-than-trusted rule) and `tcb_rotation_cases` — 13 cases covering every
mutable member:

`parent_sha256`, `supervisor_sha256`, `worker_sha256`, `host.kind`,
`host.identity`, `host.version`, `host.build`, `backend.version`,
`backend.configuration`, a rotated `trusted_components` digest, an added
trusted component, `toolchain`, and `platform`+`enforcement_backend` together.

Eleven of the thirteen change **nothing a package can see** — the build input
outside the closed `hardened` member is byte-identical to the base — so the key
movement is attributable to the trusted computing base and to nothing else. The
two that do change a visible value (`toolchain`, because the toolchain identity
is also a portable input member, and `platform`, because a receipt's target is
bound to its TCB platform) are flagged `package_visible_input_changed: true`
with a stated reason, and the validator rejects a case that misreports this in
either direction.

The validator requires: every mutable field has at least one rotation; constant
fields have none; every rotation reproduces its own digest and key; no rotation
aliases another or the base.

Cache keys now in play, all distinct: the five contract keys (pre-revision rc.4,
portable, rc.5 policy-slot reservation, hardened, hardened after a manager
update) plus the 13 rotations. The rotation `rotate-worker-identity` is the same
trusted base as the `hardened_rotated_tcb` contract case, so it reproduces that
key by construction — 18 entries, 17 distinct, one intentional coincidence, not
an alias.

### 5. Adversarial and schema coverage

- `hardened_adversarial_vectors.tcb_completeness_cases`: 21 cases — one omission
  case for each of the twelve members, plus relation mismatches, uncryptographic
  components, an undeclared mutable component, an unclaimed operating system, an
  unobserved required configuration, and a lying digest. Each names its
  enforcement site and asserts receipt, marker, and claim all reject it with
  `hardened_tcb_identity_invalid`, reusing no cache entry and publishing nothing.
- Hardened schema cases: **75** (was 33) — 7 valid, 68 invalid. New negatives
  cover every TCB omission on receipt v3/v4, marker v4 and claim v4; the
  platform/backend and target/platform mismatches; untyped, undigested and
  unknown-kind components; the revived string field; a non-hardened `goos`; a
  claim backend/OS mismatch; and prose `required_configuration`.
- Hardened suite files: **79** (was 48).

### 6. Documents

`protocol/hardened-execution.md` section 2.3 states the complete record, the
member-by-member table, the completeness requirement, and the three structural
relations; section 3.4 adds a normative mapping from each trusted item to the
record member that names it; section 8.5 states the two claim relations.
`profiles/manager-hardened.md` extends the `tcb-identity-verification`
obligation and section 3 — it still publishes no ordering of its own.
`decisions/0009` records the completeness decision, the cost, and two new
rejected alternatives. `docs/`, `SECURITY.md`, `COMPATIBILITY.md` and
`CHANGELOG.md` follow.

## What was preserved

- The cycle-1 fixes stand unchanged: `hardened-identity-binding-v1`, the
  17-phase list with one authority and one actor per phase, domain-entry before
  every in-domain actor, package exposure gated on self-test acceptance.
- Package-influence exclusions are unchanged; the profile still adds no
  package-visible field, and no new field is package-selectable.
- The capability-evidence record stays the one result-only identity, excluded
  from cache key, receipt, marker and claim.
- Every platform remains `unqualified`, `native_evidence: absent`,
  `claims_emitted: []`, `qualified_platforms: []`. Every adversarial case is
  `pending-native-validation`. No host can pass `platform-qualification`.
- `conformance/v1`, `schemas/v1` and `release/1.0.0-rc.5.json` are byte-identical
  to the accepted predecessor after regenerating with `generate-vectors`;
  `conformance/v1/manifest.json` sha256 is still
  `9ba9b8ec…` and still equals the rc.5 release pin.

## Verification

Real exit codes, each command run standalone. Full detail in
`TASK-260728-1jafds_rework-cycle-2-validation.log`.

GREEN: `make validate` 0 · `tools/validate.py` 0 · `tools/validate_hardened.py`
0 · `python3 -B -m unittest discover -s tools` 0 (**113 tests**, was 87) ·
`go test -count=1 ./tools/...` 0 · `gofmt -l tools` 0 (empty) ·
`go vet ./tools/...` 0 · `git diff --check` 0 · hardened suite byte stable
across regeneration · rc.5 parity `diff -r`/`cmp` all 0.

EXPECTED-RED in the task worktree, reported as failing: `make regenerate-check`
**2**, `make regenerate-hardened-check` **2**, `tools/release_gate.py --version
1.0.0-rc.5` **1**. All three compare against the committed `57c1f56` or require
a clean checkout, and the whole candidate is uncommitted; the directive forbids
staging or committing.

RESOLVED in an ephemeral clean probe (tar of the worktree into a fresh git repo,
signing disabled inside the probe only, `diff -r`-identical to the worktree,
deleted afterwards): `make validate` 0, `make regenerate-check` 0,
`make regenerate-hardened-check` 0, `release_gate.py --version 1.0.0-rc.5` 0 —
so the added hardened CI gates do not alter rc.5 qualification.

## Not claimed

No platform is qualified. No native adversarial test was executed on Linux,
macOS or Windows. Every guarantee is `established_in_this_revision: false`.
The conformance validator, not the published schemas, enforces the two claim
relations of item 3 — that split is stated explicitly in `tcb_completeness` so a
reader is not misled about where each check lives. Nothing was staged,
committed, or published in `curator-spec`.
