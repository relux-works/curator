# TASK-260720-31nl14 review cycle 8 rework validation

## Resolution

Closed the sync-before-journal crash window in internal/transaction without changing install consumers. Each next staging chunk now has a durable write-ahead byte boundary and SHA-256 source-prefix digest before the file write. A new PointAfterStagingChunkSync fault boundary executes after file sync and before promotion to acknowledged progress. Preparing recovery accepts only unchanged acknowledged bytes or a source-matching prefix within the exact durable write-ahead range; shorter acknowledged state, bytes beyond the authorized range, replacement, changed content, noncanonical range, and source drift remain implementation-corruption and are preserved.

## Tests

Added first- and second-chunk subprocess crash recovery, write-ahead journal validation, and explicit shorter/longer/replaced/changed concurrent-state preservation regressions. Focused tests pass; new preparation regressions pass 20 consecutive iterations; package race and vet pass; focused coverage is 75.0 percent; golangci-lint v2.4.0 reports 0 issues; overlay-backed make check and full repository race pass; native Darwin build and curator dev runtime smoke pass; Linux amd64 and Windows amd64 complete compile graphs pass. Formatting, diff, and staged-file checks are clean.

## Provenance and platform evidence

Worktree remains detached at exact origin/main 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8 with the imported TASK-260720-1zl1cj managerlock candidate provenance unchanged. No files were staged or committed. Native Windows runtime execution remains unavailable on Darwin; cross-compilation is not runtime evidence and the TASK-260720-1zl1cj qualification gate remains preserved. task-board validate reports 13 inherited unrelated issues: 12 legacy EPIC-260712 broken links and one orphan TASK-260713-7a9c1e/review.md resource.