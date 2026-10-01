# TASK-260917-2tx81l review verdict — rev2 (tree d9d1b2e8, base 30b3d678): CHANGES REQUESTED

The worktree index was written to a temp tree and it equals d9d1b2e8. Spec: disposable clone of curator-spec at b1a2efb. rc.13 is checked out as a worktree at v1.0.0-rc.13.

## Verified OK
- **Framing.** I recomputed every content-hashes-v2.json vector through the curator implementation: both colliding trees, the empty tree, nested_nul and ordinary. I used a throwaway main package in an archive of d9d1b2e8 and drove both ContentSHA256WithVersion (on-disk) and ContentSHA256Files (in-memory). All v2 hex values match, all v1 hex values match, and the v1 pair still collides. Identity{1,x}.Equal(Identity{2,x}) returns false.
- **Switch.** hashing.EnableV2Writers is the single switch. It defaults to false and has a TODO for rc.14 (internal/hashing/hashing.go).
- **Mutants.** I ran all three myself against the b1a2efb root, with real exit codes:
  - M1: v2 length framing removed in both v2 paths. `go test ./internal/hashing` exit 1. Failing tests: TestContentSHA256V2ConformanceVectors, …SeparatesTheV1Collision, …LengthFramingSeparatesAdjacentRecordBoundary.
  - M2: registry.go:352 and :373 version compare disabled. `go test ./internal/registry` exit 1. Failing tests: TestMatchesRequiresEqualHashVersion, TestResolveVersionedDoesNotAdmitV1RecordForV2Artifact.
  - M3: marker.go:1345 version compare disabled. `go test ./internal/marker` exit 1. Failing test: TestCurrentRejectsV1MarkerForV2Expectation.
  - After restoring the files, hashing, registry and marker all returned ok.
- **Ledger.** There are 0 rows owned by 2tx81l. The 7 skillfile-sources rows (install-marker-v6 ×2, skillfile-lock-v2 ×3, source-audit-v2 ×2) are correctly removed rather than re-owned: `git ls-files` at b1a2efb has no such paths, and index.json has 0 matches.

## Finding F1 (blocking): the rc.13 writer changed, and an existing test was rewritten to hide it
Gatefix-1 §1 says writers keep emitting exactly the rc.13 shapes. Review-note §2 says to flag weakened assertions.

- **The change.** internal/marker/marker.go:1231-1235 (`Write`) now always recomputes `m.ContentSHA256` from `dir`, even when the switch is OFF (v1). At base, Write persisted the caller's content_sha256 unchanged. This is an rc.13 writer behaviour change, and it also adds a full tree hash to every marker write.
- **Proof.** I restored the base (30b3d678) versions of the modified existing test files over the candidate code and ran them with CURATOR_CONFORMANCE_ROOT set to rc.13. `TestAuthoritativeCompiledMarkerRoundTripsThroughWriter` FAILS:
  - got `content_sha256: sha256:e3b0c442…` (the hash of the empty destination dir);
  - want the authoritative `sha256:e2bd2941…`;
  - no other field differs.
- **How the rev2 test hides it.** internal/marker/marker_v2_test.go:73-117 in rev2:
  - turns on `enableV2Writers(t)`;
  - injects extra fields;
  - drops the byte-exact comparison against the authoritative rc.13 compiled marker.

  This assertion was not among the 35 hosted failures. It was weakened, and it covers exactly the rc.13 shape.
- **Same pattern elsewhere.** TestWriteAlwaysProducesCanonicalMarkerV2 (renamed; :33-34) and TestReadLegacyV1AndRewrite* (:141, :208) switch to v2 mode. The rc.13-mode assertions they carried survive only partly, in the new TestCoreMarkerWriterKeepsRC13ShapesByDefault, which does not check the authoritative fixture bytes.
- **Required fix.**
  - Restore rc.13 Write semantics when the switch is off. Recompute only under v2, or not at all, and document the choice.
  - Restore the original byte-exact authoritative round-trip test unchanged, running with the switch OFF.
  - Add v2-mode variants as NEW tests instead of mutating the rc.13 ones. Do the same for the marker_v2_test.go rc.13 writer tests.

## Finding F2 (blocking): edited assertions in failing tests are not explained per test
Gatefix-1 §4 requires every edit to one of the 35 failing tests to be explained per test. The results claim these tests "pass unchanged", but four assertions were edited:

- internal/config/config_test.go:203 (TestParseRejections/schema), where 3 became 4;
- internal/config/environments_test.go:152 (TestUnknownSchemaVersionsRejected), where 3 became 4;
- internal/envmarker/envmarker_test.go:58 and :72, where 3 became 4. Run as base code, TestReadKeepsUnsupportedVersionDiagnostic fails because version 3 is now read as a known version: "version 3 requires hash_version 2".

These are legitimate reader changes, since manager-config v3 and agent-environment-marker v3 are now accepted versions. The results must say so per test instead of claiming "unchanged".

Also explain the envprofile tests that now opt into v2 writers and assert V3/hash 2:
- credential_link_test.go:411/448
- credential_record_test.go:21/69/83/122
- migrate_test.go:295/344

Each of these must keep an rc.13-mode counterpart, or state why none is needed. As it stands, the rc.13 writer path for credential-marker publication, recovery and migration upgrades lost its assertions.

## Not re-run by me
- The cmd/curator, scopes, install and envprofile full packages. I accept the hosted green gate for those.
- The count-pin exactness per digest. I accept the developer evidence; the b1a2efb index checks above are consistent with it.
