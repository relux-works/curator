# TASK-260917-2tx81l — manager-content-hash-v2: rev3 ACCEPTED

Reviewed base `30b3d6781924733ed7120c8adac5183b877f506a` to candidate tree `92a072c4e1b03678ea64a2801069ea471628e8e4`. The 45 changed paths and all 1,287 source/test/config blobs match the candidate. Review runs used an archive of this exact tree; no repository source, tests, index, branch, CHANGELOG or LOGBOOK were modified.

Blocking findings: **none**. Previous rev2 F1 and F2 are resolved. No repeated finding remains; `findings: []`. This acceptance covers the requested rev3 re-review, not an assertion that every test against the unreleased candidate spec is green; see N1 below.

## Swept review surfaces

| Surface | Disposition | Independent evidence |
|---|---|---|
| F1: default marker writer | Held; rev2 F1 resolved | `internal/marker/marker.go:1186` selects the legacy schema and clears HashVersion while the switch is off. Lines 1231–1238 document and guard recomputation to hash v2 only. Full marker package passes with base tests under rc.13. |
| F1: original marker assertions | Held | `marker_v2_test.go`, `marker_v3_test.go`, and `marker_v4_test.go` are byte-identical to base. This includes the authoritative compiled-marker byte comparison and legacy rewrite tests; they run in the default v1 mode. New v2 expectations are confined to new tests in `marker_hash_v2_test.go`. |
| F2: original credential/migration coverage | Held; rev2 F2 resolved | Each of credential_link_test.go, credential_record_test.go, and migrate_test.go retains its entire base file as a byte-identical prefix; v2 tests are appended with new names. All 61 original environment tests and all 15 original registry-install tests pass independently with the base overlay. |
| F2: four reader-version edits | Held | Results explain TestParseRejections/schema, TestUnknownSchemaVersionsRejected, TestUnsupportedVersionIsRejected, and TestReadKeepsUnsupportedVersionDiagnostic individually. Their 3→4 probes preserve unknown-version coverage now that v3 is supported. Base config has only the two expected rejection failures; base envmarker has only the expected diagnostic failure. Candidate config and envmarker pass. |
| New v2 counterparts | Held | All four new envprofile V2WriterMode tests pass through Resolve/publication/recovery and ApplyMigration. Marker package and new marker variants pass against both reviewed suite contexts. |
| Framing and version comparisons | Held | Hashing implementation is unchanged from rev2. The five published content-hashes-v2 vectors execute through hashing/registry, including exact collision-pair and empty-tree hex; explicit identity mismatch checks pass. All three required mutants are killed with exit 1, and the unmutated selectors pass with exit 0. |
| Ledger and count pins | Held | Both ledger files are unchanged from rev2. Recomputed suite digests, table sizes and totals match: rc.13 has 88 pins/1,636 cases; b1a2efb has 97 pins/1,722 cases. Independent schema-index counts cover 75 owned schema cases plus five content-hash vectors. All 80 owned core rows are removed. The seven deferred paths are absent from both indexes and correctly removed, not re-owned. |
| Additional rework and file hygiene | Held | The audit delta restores the base v1 Detect path and preserves stateread for new v2 reads. Existing managed-write nofollow paths remain in use. The base state-read scanner passes. All five added filenames are portable. No release/logbook or stray repository changes. |
| Free hunt beyond F1/F2 | No introduced blocker | N1 is independently reproduced on the exact base and documented below. |

## Base-test replay

Extracted all **21/21 modified existing test files** from base into an overlay over candidate code. The attached JSON contains their paths and SHA-256 digests. Canonical `/private/tmp` paths ensure Go actually applies the overlay; the known reader exceptions were used as a control. Existing marker writer files already equal base and need no overlay.

All commands use `-count=1`. Conformance root: `/tmp/curator-spec-rc13-task-260917/conformance/v1`, commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`.

| Replay | Observed result |
|---|---|
| Full conformancecoverage, contextaudit, contextlock, contextmaterialize, contextstore, marker, registry, interop/environments packages | Exit 0 |
| Base config and envmarker packages | Exit 1 with exactly the three explained failing tests above; no writer regression |
| Credential-link / credential-record / policy / state-read checks | 21/21 individual commands exit 0 |
| Managed-environment checks | 18/18 individual commands exit 0 |
| Migration checks | 22/22 individual commands exit 0 |
| Registry-install checks | 15/15 individual commands exit 0 |

Every one of the **76/76** individual commands contains an explicit named PASS record, not merely a package exit or a skipped/no-tests result. Their commands, durations, real process exits and log digests are attached.

## Mutants independently rerun

All mutations were applied only through canonical-path overlays to the disposable candidate archive.

| Mutant | Failing negative tests | Mutant exit | Legacy-only control exit | Unmutated exit |
|---|---|---:|---:|---:|
| M1: remove all four length-frame writes from both hashing paths | TestContentSHA256V2ConformanceVectors, TestContentSHA256V2SeparatesTheV1Collision, TestContentSHA256V2LengthFramingSeparatesAdjacentRecordBoundary | 1 | 0 | 0 |
| M2: remove version comparisons from both registry matching functions | TestMatchesRequiresEqualHashVersion; TestResolveVersionedDoesNotAdmitV1RecordForV2Artifact | 1 | 0 | 0 |
| M3: remove Current's recorded/expected version comparison | TestCurrentRejectsV1MarkerForV2Expectation | 1 | 0 | 0 |

Coverage is **3/3 specified mutants**. The legacy-only controls are tests against the mutated candidate, not a claim that these v2-only mutation sites existed on the base. Registry ResolveVersioned and marker Current are production entry points. The v2 envprofile tests also drive publication, crash recovery and migration rather than only validators.

## Candidate suite and evidence bounds

Candidate spec commit `b1a2efb6fa28d014968a2a8fd7641823b5f3cf28`; manifest `950ee74ad148615c273fe95bbb93f1bc0f9bdf2ea2bd9395f1dc8e3601419e60`. rc.13 manifest: `be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca`.

The candidate-root run passed hashing, conformancecoverage, config, contextaudit, contextlock, contextmaterialize, contextstore, envmarker, marker and registry. All required v2 drivers and all four new environment counterparts were separately rerun with exit 0. The broader interop package produced N1, so that broader command is recorded as exit 1.

**N1 — pre-existing candidate-suite snapshot driver mismatch (nonblocking for this re-review).** `internal/interop/environments/snapshot_acquisition_test.go:154` still calls frozen `hashing.ContentSHA256`, while b1a2efb's snapshot-acquisition vector changed its expected hash to v2 without an explicit hash-version field. Both autocrlf cases fail at line 159: actual `sha256:500ea934403d10a2a0b6b7e8874790e489ee002328d3dc0edbda2fe5be2bced0`, expected `sha256:ecca17aacc80360a390b2403016f71e04a29e5d46c6643b77a5ee77b47186a9e`. I independently archived and tested base 30b3d678 against the same b1a2efb root: the identical failure occurs, exit 1. This unchanged driver is outside the 80 newly driven core rows and passes under rc.13. Do not describe the entire b1a2efb suite as green. Address this driver/vector version selection before promoting that suite to SPEC_PIN; it is not an F1/F2 regression or a reason to change the frozen v1 hash.

Accepted existing evidence: hosted run **36801119718**, recorded in TASK-260917-2tx81l_change-request_rev3-validation.log, passes lint, native Linux/macOS/Windows tests and the recorded race/gate lanes. I verified its commit `33f1b9bc5d0c1879853068c837e7c69604c65237` has exactly the reviewed tree. Rose-air and hosted candidate-suite lanes are skipped, not verified. I did not rerun the full hosted gate, build or lint locally. Full count-pin exactness outside the spot-checked owned families is accepted from attached producer evidence; both count/ledger files remain unchanged from the prior review.

## Attempt accounting and attached outcomes

Initial `/tmp`-alias overlays did not affect compilation, as the expected reader failures revealed. Those runs are explicitly excluded. Incomplete broad environment/install groups were terminated and replaced by individually logged tests. The first bounded individual-test batch reached its wall limit during bootstrap (exit 124); the second batch reran bootstrap successfully and completed every outstanding selector. No incomplete attempt is counted as a pass. The initial overbroad archive extraction was stopped and replaced with a bounded source archive. All verification processes have returned before this verdict.

- TASK-260917-2tx81l_review-evidence-rev3.json — identities, test-file provenance, exact commands/exits, coverage counts, mutations and log digests.
- TASK-260917-2tx81l_review-logs-rev3.tar.gz — complete logs, including expected failures, excluded/interrupted attempts, base reproduction of N1, and the bounded replay harness.

The run is not goal-bound (`task-board spawn goal` returned no active goal). Acceptance uses `accept_cr(... revision=3 ...)`, routes to integrating, and leaves checkpoint/integration to the tracked producer. No commit acknowledgement is supplied.
