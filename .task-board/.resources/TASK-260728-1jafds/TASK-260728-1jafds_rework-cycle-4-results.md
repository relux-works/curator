# Rework cycle 4 — closes both blockers of review verdict cycle 4

Worktree: `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-1jafds/curator-spec-worktree`
Baseline: uncommitted `hardened-1.0.0-rc.1` candidate on top of `57c1f56`.
Nothing staged, committed, published, pinned, or natively qualified.

## Reproduced first

Both findings were reproduced against the shipped candidate before any edit,
and the same probes now run green (`rework-cycle-4/probe-r4.py`, full output in
`validation.log`):

| Probe | Before | After |
|---|---|---|
| valid receipt-v3 with `tcb.host.build = null` | `schema_errors=0`, `check_tcb_record` **accepted** | `schema_errors=1`, semantic check **rejects**: "the observed host reports no kernel build identity" |
| two kernels sharing platform, release, and build identifier | one record, one digest, one cache key | distinct build digests, distinct TCB digests, distinct cache keys, receipt/marker/claim all rejected against the base |
| `identity-reverification` at phase 15, `domain-teardown` at 16 | re-verification precedes the proof that every domain member exited | `domain-teardown` 15 → `identity-reverification` 16 → `publication` 17, in all four places that state the order |
| manager obligation re-verifies 4 of 12 hashed members | supervisor, worker, snapshot, toolchain only | all 9 mutable members plus the frozen snapshot, as a byte-identical recomputation of the complete record |

## R4-1 — the observed host is an identity, not a description

**Normative.** `protocol/hardened-execution.md` section 2.3.3 gains
`curator-hardened-host-build-v1`; section 6.3 gains the per-platform kernel
build identity declaration.

`host.build` is a **required closed record** — `{algorithm, identifier,
content_sha256}` — and MUST NOT be null, absent, or a bare string. The digest is
a construction, not a name:

```text
SHA-256( "curator-hardened-host-build-v1" || 0x00
  || len‖identity || len‖version || len‖identifier || uint64be(source_count)
  || for every declared source, in declared order: len‖name || len‖value )
```

- `identifier` is the immutable build identifier the platform documents, in the
  grammar section 6.3 declares, and MUST be **exactly the value of that
  platform's declared identifier source** — it cannot drift from the bytes the
  digest covers.
- Section 6.3 declares, per platform, the ordered closed source list:
  linux `kernel.build-id`, `kernel.osrelease`, `kernel.version-string`;
  macos `kern.osversion`, `kern.osproductversion`, `kern.version`;
  windows `kernel.current-build-and-ubr`, `kernel.build-lab-ex`,
  `kernel.image-file-version`. Each source is named and defined in the document.
- `host.version` gets a bounded release grammar for the reason the backend
  version has one: an unconstrained string gives one kernel many spellings.
- **Fail-closed.** A declared source that cannot be read, is empty, or is
  unavailable is `hardened_tcb_identity_invalid` in `tcb-identity-verification`
  — before domain establishment and before any package byte. No partial digest,
  no cached value, no build-time constant, no identifier without a digest. A
  platform whose declared sources cannot be observed does not qualify.
- What it does **not** claim: a host that lies about every declared source lies
  inside its own TCB, which section 3.3 already excludes. What it does claim:
  two materially different kernels, observed truthfully, cannot produce one
  `hardened-tcb-v1` record. Both statements are in the document.

**Independently recomputable fixtures.** The identity-separation vector
publishes `host_build_fixtures`: 7 fixtures carrying the exact bytes, per-field
byte lengths, source count, and expected digest. `tools/generate-hardened`
computes them in Go from the construction; `tools/validate_hardened.py`
recomputes every one in Python from the published bytes. **No `host.build`
digest anywhere in the suite may come from anywhere but a published fixture**,
and the fixture must be the one computed over that record's own kernel identity,
release, and identifier — so a real digest cannot be carried to another host.

Two structural claims are checked rather than asserted:

- `macos-host-build-recompiled-kernel` shares the base's platform, release, and
  build identifier and differs only in `kern.version`; the digests must differ.
  That is the reviewer's aliasing case, as a fixture.
- `macos-host-build-source-boundary-shift` concatenates to exactly the base's
  observed values; an implementation that hashed them as one blob would alias
  the two. It is flagged `conforming_observed_host: false` and the validator
  rejects any record that adopts it as an identity.

**Per-facet rotation coverage**, the same rule review cycle 3 imposed on
components: rotating the `host` record is not coverage for the sources two
kernels differ in.

| Facet | Rotation |
|---|---|
| `identifier` | `rotate-host-build-identifier` (next build of the same release; the identifier source moves with it) |
| `source-value` | `rotate-host-build-source` (**the finding**: same platform, release, and identifier; only a declared source differs) |
| `release-binding` | `rotate-host-version` (same sources under another observed release) |

`rotate-host-build-source` is package-invisible: nothing a package can see
differs, so only the observed kernel moved the cache key.

**Schema.** `hardenedHostBuildV1` (closed, required),
`hardenedHostBuildAlgorithmV1` (const), `hardenedHostVersionV1` and
`hardenedHostBuildIdentifierV1` (bounded, `(?![\s\S])`-tailed for the trailing
newline reason cycle 3 recorded), and
`hardenedHostBuildIdentifierPlatformRelationV1` binding the identifier grammar
to the platform in both directions. `validate_schema_closed_value_sets` pins the
schema patterns against the profile's own table, so widening one fails.

## R4-2 — teardown first, then a complete re-verification

**Ordering.** `domain-teardown` is phase 15 and `identity-reverification` is
phase 16, in `protocol/hardened-execution.md` section 7.2, in
`profiles/manager-hardened.md`, in `hardenedPhaseV1`, and in the executable
`ordered_phases`. Teardown destroys the domain **and joins it**, so when
re-verification begins no domain member is running and none can start. Two new
ordering invariants are published and checked against the list itself:
`teardown-before-identity-reverification` and
`identity-reverification-before-publication`; `teardown-before-publication` is
kept and now follows from them.

**Completeness.** A new section — *End-of-operation re-verification* — states
that phase 16 recomputes the **complete** `hardened-tcb-v1` record from the same
canonical pinned identities phase 5 used, and requires byte identity plus digest
equality. Re-verifying a subset, restating the phase-5 record, or comparing the
phase-5 digest against itself explicitly do not discharge the obligation. The
manager row was rewritten from four identities to all of them.

The executable form derives its member set from the closed record rather than
restating it — `reverified_members` is every non-constant TCB member plus
`source-snapshot`, 10 in total — so a future member is covered the moment it is
added. Each one carries its own adversarial omission case, alongside the
phase-order, restated-record, and changed-member cases (13 reverification cases).

The alternative the verdict allowed — keeping the old order and proving
immutability with an immutable-handle or snapshot construction — is recorded in
decision 0009 as rejected: it would need a per-platform handle model this
revision has no qualified platform to validate, and it buys nothing over joining
a domain that is being torn down anyway.

## Self-review found three holes, all fixed

Writing the probes surfaced three defects in this rework that the shipped checks
would not have caught:

1. **The boundary-shift fixture did not prove length framing.** Because the
   declared source *names* sit between the values in the hashed stream, a
   value-boundary shift cannot alias even without framing. The fixture is now
   stated as what it actually proves — an implementation hashing the values as
   one blob would alias — and genuine framing tests (`a|bc` versus `ab|c`, and
   the adjacent leading fields) were added in both Go and Python. The
   construction probe carries both mutations and each is caught by the test that
   claims it.
2. **The hashed source count was overclaimed.** Per-field framing already
   separates a truncated list, so the count states cardinality explicitly rather
   than being load-bearing. The protocol wording, the Go test, and the Python
   test were corrected to say what is true; removing the count is still caught,
   by cross-implementation disagreement.
3. **The normative prose was not parsed.** Section 6.3's source table and the
   re-verification section could have drifted from the tools silently — the
   exact failure class this review keeps finding. `validate_host_build_declaration_document`
   now parses section 6.3 and requires it to declare exactly the sources and
   identifier source the tools hash; `validate_reverification_documents` requires
   the protocol subsection to name every re-verified member and the manager row
   to state the complete record, byte identity, the subset prohibition, and the
   joined domain. `hardenedPhaseV1` is pinned to the normative ordered list.

`conforming_observed_host` was also published-but-unchecked; it is now enforced
in both directions.

## Coverage

| | Cycle 3 | Cycle 4 |
|---|---|---|
| Python tests | 151 | **189** |
| hardened suite files | 92 | **108** |
| schema cases | 88 | **104** (7 valid, 97 invalid) |
| TCB relations | 12 | **14** |
| TCB completeness cases | 29 | **38** |
| TCB rotations | 17 | **18** (3 per-host-build-facet) |
| component digest fixtures | 10 | 10 |
| host build fixtures | 0 | **7** |
| ordering invariants | 8 | **10** |
| reverification cases | 0 | **13** |

Go tests add build-identity fixture distinctness and stability, the
same-tuple non-aliasing property, length framing, domain separation against the
component and tree algorithms, the observed tuple being inside the digest,
identifier-equals-declared-source, every observed host tracing to a fixture,
per-facet rotation coverage, the per-platform source declarations, the phase
order, the complete re-verified member set, and the omission-case coverage.

## Mechanical evidence — real exit codes

GREEN in the task worktree: `make validate` 0 (42 portable schemas, 422 portable
vector files, 6 hardened schemas, 108 hardened suite files, 189 Python tests,
both Go tool packages), `tools/validate.py` 0, `tools/validate_hardened.py` 0,
`unittest discover` 0, `go test -count=1 ./tools/...` 0, `go vet ./tools/...` 0,
`gofmt -l tools` empty, `git diff --check` 0.

Adversarial probes, all exit 0:

- `probe-r4.py` — both reviewer findings reproduced as passing checks, all 7
  build identities recomputed, all 25 conforming observed hosts traced to a
  published fixture, the phase order verified in all four documents.
- `mutation-probe-r4.py` — 16 shipped rules and 5 normative document statements
  neutered one at a time: **21/21 detected, 21 distinct messages**.
- `validator-mutation-probe-r4.py` — 13 adversarial instances against the
  conformance validator's semantics: **13/13 rejected, 13 distinct messages**,
  with the shipped records still validating as the positive control.
- `construction-probe-r4.py` — 7 mutations of the Go construction and the phase
  order, each regenerated and re-validated by both implementations:
  **7/7 detected**.

EXPECTED-RED in the task worktree, reported truthfully and structural:
`make regenerate-check` 2, `make regenerate-hardened-check` 2,
`tools/release_gate.py --version 1.0.0-rc.5` 1. All three compare against
committed `57c1f56` or require a clean checkout, while the whole candidate is
uncommitted and the directive forbids staging.

RESOLVED in an ephemeral clean probe byte-identical to the worktree
(`diff -r` empty): `make validate` 0, `make regenerate-check` 0,
`make regenerate-hardened-check` 0, `release_gate.py --version 1.0.0-rc.5` 0 —
so the hardened changes do not alter rc.5 qualification. Probe deleted; nothing
staged or committed in curator-spec.

## Preserved

- Cycle-1 identity and ordering fixes, cycle-2 TCB completeness, cycle-3
  component constructions and backend-version comparison, the 17-phase
  authoritative list, package-influence exclusions, capability evidence as the
  only result-only identity. All 10 component fixtures keep their published
  digests.
- `conformance/v1`, `schemas/v1`, and `release/1.0.0-rc.5.json` are `diff -r`
  identical after regenerating with `generate-vectors`; portable manifest
  sha256 `9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf`
  equals the rc.5 pin and the hardened metadata's recorded baseline.
- Hardened suite and release metadata are byte-stable across regeneration.

## Not claimed

No platform is qualified. `claims_emitted` and `qualified_platforms` are empty,
every platform declaration is `unqualified`, every adversarial case is
`pending-native-validation`, and no native test was run on any platform. Section
11 now additionally requires the owning qualification task to confirm natively
that every declared build-identity source is readable and that replacing the
kernel moves at least one of them.

## Files touched this cycle

Normative: `protocol/hardened-execution.md`, `profiles/manager-hardened.md`.
Schemas: `schemas/hardened/v1/hardened-common.schema.json`.
Tools: `tools/generate-hardened/main.go` (+`main_test.go`),
`tools/validate_hardened.py` (+`test_validate_hardened.py`).
Regenerated: `conformance/hardened/v1/**`, `release/hardened-1.0.0-rc.1.json`.
Prose: `CHANGELOG.md`, `COMPATIBILITY.md`, `SECURITY.md`,
`decisions/0009-hardened-build-execution-profile.md`,
`docs/hardened-build-execution-profile.md`.

No portable schema, vector, generator, or pin was edited.

## Reviewer-finding map

| Finding | Where it is closed |
|---|---|
| R4-1 nullable `build` permitted by schema | `hardenedHostBuildV1` required and closed; `invalid-tcb-host-build-null` cases on receipt v3, receipt v4, marker v4, claim v4 |
| R4-1 nullable `build` accepted by `check_tcb_record` | `check_tcb_record` rejects with a fail-closed message; `validator-mutation-probe-r4.py` case 1 |
| R4-1 `host.version`/`host.build` unconstrained strings | `hardenedHostVersionV1` bounded grammar; `build` is a digest over declared sources |
| R4-1 no platform-specific strong build identity | `curator-hardened-host-build-v1` + section 6.3 per-platform declarations, parsed by the validator |
| R4-1 absence must fail closed | section 2.3.3 fail-closed semantics, rejected in `tcb-identity-verification`, checked against the phase list |
| R4-1 schema/semantic/rotation/cache/receipt/marker/claim mutants | 16 new schema cases, 9 new completeness cases, 3 new rotations, 13 validator instances |
| R4-2 reverification before teardown | phases swapped in all four statements of the order; two new ordering invariants |
| R4-2 manager re-verifies a subset | manager row rewritten; `reverified_members` derived from the closed record; partial and restated forms forbidden |
| R4-2 vector locks the incomplete order in | `validate_identity_reverification`, `validate_reverification_documents`, phase-order and omitted-member adversarial cases, phase-enum pin |
