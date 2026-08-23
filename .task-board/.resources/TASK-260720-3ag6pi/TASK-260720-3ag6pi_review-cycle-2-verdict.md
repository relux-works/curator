# TASK-260720-3ag6pi review cycle 2 verdict

## Verdict

**CHANGES REQUESTED → `to-dev`.**

This is ordinary, substantive implementation rework. It is not an external
blocker and does not require a human-only product or architecture decision.

## Candidate and supersession authority

The previous stop-line publication blocker is resolved:

- the board records human authorization for rc.5 landing/publication on
  `TASK-260730-1fsbqd`;
- remote `main` and annotated tag `v1.0.0-rc.5` resolve to exact commit
  `f5d7673039226ab81de2f4f87e2155ae995c4df3`;
- its tree `78210085727ec33b79a050a807f51da253ffb0c8` equals the independently
  accepted rc.5 candidate tree;
- its sole parent is legacy baseline
  `57c1f56846d221ecc55786bd3c2467ec32f11730`.

The review therefore applied the approved rc.5 supersession. Literal rc.4
correctly fails the current version gate; unwrapped
`make release-check VERSION=1.0.0-rc.5` passes at the landed commit.

## What independently passes

- `make validate`: 42 schemas, 447 manifest entries, 41 Python tests, and Go
  tests; no skip is reported.
- Two independent clean regenerations and `make regenerate-check`: no diff,
  empty Git status, and identical 448-file conformance aggregate SHA-256
  `e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae`.
- Unwrapped rc.5 release gate from a clean committed checkout using
  `/usr/bin/git`, with no alternate index or Git wrapper.
- All 447 manifest entries and hashes, all 376 schema-case index entries, both
  release manifest pins, all required v6 schema groups, 13 build fixtures, and
  11 expected build-driver files.
- Twelve frozen legacy schemas and all 24 baseline legacy cases are
  byte-identical to the parent baseline. Seventy newly added v7 guard cases for
  manifest schemas 1–5 are all invalid as intended. Install marker v1 and claim
  v1 remain frozen; claim v1 is still schema 1 / protocol rc.3 with its three
  exact historical SHA-256 values.
- The five v6 wire schemas are byte-identical to the accepted rc.4 composite.
  Every accepted build-driver case name remains; rc.5 adds one positive and two
  explicit non-alias rejection cases.
- No package fixture or generated artifact is executed, and no release
  evidence is fabricated.

## Acceptance-blocking regression

The landed rc.5 tree dropped the accepted protocol-v6 compiled manager
lifecycle conformance surface:

- accepted `TASK-260720-cw39jh`
  `manager-lifecycle.json` SHA-256:
  `676e617a0e0a6d575310f38e1de740eab583d709e2351be9eaa818c9882d78d4`;
- landed rc.5 `manager-lifecycle.json` SHA-256:
  `2ddbd2665a63f44dc0e03e060f4cd34bfde219a56b3192511fe1ef81047feedf`;
- the landed digest is exactly the old rc.3 baseline digest;
- accepted named lifecycle cases: 32; landed named cases: 10; missing: 22;
- `schema_version`, `compiled_build_fixture`, and ten compiled lifecycle groups
  are absent.

The missing cases cover:

- audit-before-build and provider/lexical order;
- compiled dry-run read-only behavior;
- private-build failure isolation;
- protected cache publication, races, corruption, and trust;
- deterministic locks, consumer-last commit, and reverse rollback;
- concurrent-project isolation;
- transaction-ID recovery and recovery-after-private-builds;
- compiled currentness, repair, and locked GC.

All 22 names are absent from every landed conformance file. The current
generator deterministically reproduces that incomplete vector. The current
validator no longer requires the accepted lifecycle groups or compiled fixture
identity, and the current release gate no longer contains the accepted v6
required-artifact and frozen claim-v1 guards. Consequently all green commands
validate an internally self-consistent but contract-incomplete suite.

## Required rework before another reviewer cycle

1. Port the accepted schema-6 compiled lifecycle surface from
   `TASK-260720-cw39jh` into the current rc.5 generator and vector. Preserve all
   22 accepted semantics, but bind `compiled_build_fixture` to the current
   rc.5 portable `manager-worker-v1` build identity rather than blindly copying
   stale rc.4 identity bytes.
2. Restore generator tests and fail-closed validator requirements for every
   lifecycle group/name, `schema_version`, and exact compiled fixture identity.
3. Restore rc.5-appropriate release negative coverage for missing/renamed v6
   artifacts and lifecycle cases, frozen claim v1, claim-v2 transition, and
   stale/duplicate suite evidence. A manifest hash alone is insufficient.
4. Regenerate the full inventory and rc.5 pins, then rerun validate, two
   independent regenerations, regenerate-check, and unwrapped release-check.
5. Attach a revised story-criterion/minimum-cluster matrix in which every
   lifecycle and fail-closed row has passing executable evidence.

Relevant ownership is `TASK-260720-cw39jh` for lifecycle vectors and
`TASK-260720-1u7hes` for validation/release guards. The integrated verifier
must remain in rework until both surfaces are restored together.

## Reviewer boundary

The reviewer made no product-code edit, stage, commit, push, tag, release, or
pin change and supplied no `commit_ack`.

