# Skillfile sources revision 1 conformance corpus (vendored, pinned)

The corpus and schemas below are copied without edits from curator-spec v1.0.0-rc.13 tag
commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065` (Decision 0022):
`conformance/skillfile-sources-v1` and `schemas/skillfile-sources-v1`.
`SKILLFILE_SOURCES_PIN` records the source commit. `MANIFEST.sha256` pins every
vendored byte; the cross-conformance pin test rejects drift, additions, and
removals.

The schema and semantic case counts are pinned in `.github/ci/conformance-case-counts.tsv`.
Every published semantic and schema case is exercised by the production-entry
matrix in `internal/crossconformance`.
