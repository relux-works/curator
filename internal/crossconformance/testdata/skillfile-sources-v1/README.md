# Skillfile sources revision 1 conformance corpus (vendored, pinned)

The corpus and schemas below are copied without edits from curator-spec main
commit `574636785c9da22757095ca279e8a9da801156ec` (Decision 0022):
`conformance/skillfile-sources-v1` and `schemas/skillfile-sources-v1`.
`SKILLFILE_SOURCES_PIN` records the source commit. `MANIFEST.sha256` pins every
vendored byte; the cross-conformance pin test rejects drift, additions, and
removals.

The schema and semantic case counts are pinned in `.github/ci/conformance-case-counts.tsv`.
Every published semantic and schema case is exercised by the production-entry
matrix in `internal/crossconformance`.
