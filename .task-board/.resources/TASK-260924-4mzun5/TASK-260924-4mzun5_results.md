# TASK-260924-4mzun5 results — record the dependency directory in the legacy Skillfile lane (developer)

## Decision (R1)

The legacy (Skillfile schema-1) lane RECORDS the schema-9 dependency
directory. It does not refuse.

Deciding spec clause — skillfile-sources section 4 (curator-spec 7eaeb73f):

  "For schema-2 installations write marker schema 5, except that a core
  schema-9 installation writes draft marker schema 6, including when its
  root project still uses Skillfile schema 1; marker v5 cannot record
  manifest version 9."

Supporting clauses:

- skillfile-sources section 4: "Draft marker v6 is marker v5 with
  schema_version 6 and skill_schema_version through 9; every other
  field, requiredness rule, and currentness comparison is unchanged.
  Writers use draft marker v6 exactly for core schema-9 installations."
- core section 4.4: "Any dependency lock or install identity MUST record
  the normalized directory. ... A local source-audit record MUST bind
  that same package identity, so an audit result for one directory
  cannot be reused for another directory merely because repository and
  commit match."
- manager profile section 2.1: "For a schema-9 dependency, the normalized
  selected directory is part of the local audit subject and audit-cache
  key. ... The lock and marker package identity MUST carry the same
  directory."

Implementation of the decision:

- internal/marker/marker.go: new SchemaV6=6 carrier; Write() selects v6
  exactly for schema-9 installations with a package, v5 otherwise;
  validMarker enforces v5 skill schemas 1..8 and v6 skill schemas 1..9;
  SupportedSchema/BuildBearingSchema/Current cover v6; the v6 package
  shape reuses the closed v5 arms unchanged.
- internal/install/legacypackage.go (new): stages the frozen package
  identity plus a deterministic lock binding for legacy-lane nodes no
  legacy marker can record (every schema-9 installation, every
  subdirectory-selected installation). Network identity ->
  network-git arm with the normalized directory; local source ->
  configured-git arm rooted at "."; a subdirectory without a network
  identity fails closed with source_selection_invalid.
- internal/install/install.go + global.go: buildMarker now returns an
  error; schema-9 or subdirectory-selected nodes escalate to a package
  marker on the schema-1 lane too (v6 for schema 9, v5 for earlier
  schemas at a subfolder); package-selected nodes with compiled
  commands are refused with source_selection_invalid (receipt-3 build
  identity needs Skillfile schema 2). Audit subjects carry the
  normalized node directory.
- internal/audit/audit.go + sourceaudit.go: Subject gains Directory;
  verdict cache key and stored record bind the normalized directory
  (empty reads as "."); cmd/curator/main.go passes node.Directory
  through the audit entry.

## Production-entry rows (positive + negative)

- Positive: internal/install/legacy_directory_test.go
  TestLegacyLaneRecordsDependencyDirectory — real install entry over a
  schema-1 project whose schema-9 root requires subdirectory-selected
  schema-9 + schema-8 packages; asserts v6/v5/v6 markers with the
  normalized directories, audit verdicts binding the same directories,
  and a second install up-to-date. Plus cmd/curator/audit_directory_test.go
  TestCLIAuditBindsDependencyDirectory through `curator audit`.
- Negative: TestLegacyLaneRefusesGlobDependencyDirectory — real install
  entry with `?`, `[`, `*` dependency directories is refused, names
  dependencies.skills.backend.directory, and materializes nothing.
- Helper pins: TestLegacyMarkerPackageArms (arm selection, deterministic
  directory-sensitive binding, subdir-without-identity refusal),
  TestBuildMarkerRefusesPackageSelectedBuildsOnLegacyLane,
  internal/audit/directory_test.go (cache binds directory; legacy
  record without directory binds root),
  internal/marker/marker_v6_test.go (carrier selection, directory
  round trip, malformed-identity refusals incl. glob directory,
  currentness incl. v5-never-current-for-schema-9),
  TestMarkerV5RefusesSkillSchema9 (accepted v5 rejection pin).

## Glob class on the dependency path (R2)

- internal/skillspec/conformance_test.go:
  TestDraftManifestDependencyDirectoryGlobClass pins `?` and `[`
  (skills/devel?per, skills/[abc]per, skills/devel[, skills/devel?)
  refused on the dependency path exactly like the spec `*` vector;
  grammar + schema-case tests now read testdata/draft-sources-v2.
- internal/install/legacy_directory_test.go:
  TestLegacyLaneRefusesGlobDependencyDirectory drives `?` and `[`
  (plus `*`) through the real install entry.

## Files changed

Modified (14):
  CHANGELOG.md (one Unreleased line, operator-visible marker/audit rule)
  cmd/curator/main.go
  internal/audit/audit.go, audit_test.go, opaque_v1_rework_test.go,
    opaque_v1_test.go, sourceaudit.go
  internal/install/global.go, install.go
  internal/marker/marker.go, marker_v5_schema_test.go,
    marker_v5_test.go, schema_band_test.go
  internal/skillspec/conformance_test.go

Added (vendored spec suite, byte-identical to curator-spec 7eaeb73f):
  internal/skillspec/testdata/draft-sources-v2/
    manifest-dependency-directories.json (marker_schema_version 6)
    schema-cases/agent-skill-v9/* (11), schema-cases/csk-skill-v9/* (11)
  internal/marker/testdata/draft-sources-v2/
    install-marker-v6.schema.json

Added (new code/tests):
  internal/install/legacypackage.go
  internal/install/legacy_directory_test.go
  internal/audit/directory_test.go
  cmd/curator/audit_directory_test.go
  internal/marker/marker_v6_test.go, marker_v6_schema_test.go

Unmodified: every path under
  internal/skillspec/testdata/draft-sources-v1/ and
  internal/marker/testdata/draft-sources-v1/
  (`git diff HEAD --stat` over both directories is empty).

## Compile-only checks (R223: no local `go test`)

  go vet ./...                                  exit 0
  go build ./...                                exit 0
  gofmt -l internal cmd                         empty (clean)
  golangci-lint run <5 touched packages>        exit 0, 0 issues
  golangci-lint run ./...                       exit 1, sole finding is a
    stale-cache G304 in a sibling Story worktree file that does not
    exist (verified absent); no finding in any touched file.

No `go test` was run locally per R223. The hosted gate is the suite
arbiter; tests-passing and hosted-gate-green checklist items are left
for its evidence.

## What changed versus the previous attempt (rework)

The previous candidate moved the v1 suite to v2 (renames) and rewrote
the v1 marker schema maximum 9 -> 8, which the board refused as
reverting 24 trunk-changed paths. This revision, per the rework note:

1. Ran only the allowed scoped checkout:
     git checkout HEAD -- internal/skillspec/testdata/draft-sources-v1
       internal/marker/testdata/draft-sources-v1
   Both v1 directories now match trunk byte-for-byte (v1 manifest keeps
   marker_schema_version 5; v1 install-marker-v5.schema.json keeps
   maximum 9).
2. Kept the spec v2 suite as copies under .../testdata/draft-sources-v2/
   (verified byte-identical to curator-spec 7eaeb73f: manifest with
   marker 6, all 22 schema cases, install-marker-v6.schema.json) with
   the new tests pointed at v2.
3. Kept the rest of the work unchanged (legacy-lane decision,
   marker/audit/install changes, production-entry rows, glob vectors).

Known divergence left for a separate task (per the rework note, not
changed here): the trunk v1 vendored install-marker-v5.schema.json
carries skill_schema_version maximum 9, while the spec accepted
marker v5 (schemas/skillfile-sources-v1 at 7eaeb73f, and the
draft-sources-v2 README check) fixes it at 8 with v6 as the schema-9
carrier. No test asserts the vendored v1 maximum (validateV5AgainstSchema
pins 1..8 in code; the top-level test does not check the range), and
the implementation enforces v5 1..8 / v6 1..9, so this does not affect
the gates here.
