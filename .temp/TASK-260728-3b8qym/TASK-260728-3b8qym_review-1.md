# TASK-260728-3b8qym independent review cycle 1

## Verdict

CHANGES REQUESTED — route to `to-dev`.

## Verified evidence

- Assigned worktree HEAD is exactly `57c1f56846d221ecc55786bd3c2467ec32f11730`; index is unstaged and no task commit was added.
- Disposable clean candidate passed `make release-check VERSION=1.0.0-rc.5`, two additional consecutive `make regenerate-check` runs with zero diff, `go vet ./tools/...`, Go formatting, Python compileall, and `git diff --check`.
- Validation reported 42 schemas and 411 vector files; 17 Python tests and all Go tests passed.
- Manifest inventory is unique, complete, non-self-referential, and exactly 411 files. Manifest SHA-256 is `sha256:30a64ed0da6e4e68abb5f46e8807f7bc57a4545c7c582e644c9d09c9406c9324`; rc.5 metadata SHA-256 is `ddd8fc11060e164d8e86192263ca6abaf5ec5881edc264e6579a2951f92a0fc3`.
- Protected legacy schema bytes for manifest schemas 1-6, receipt v1, markers v1-v2, and claims v1-v2 are identical to the accepted baseline.
- Independently recomputed all raw-object IDs and parsed the three positive pack/index fixtures. Pack v2/v3 headers, zero counts, SHA-1/SHA-256 trailers and names, index-v2 magic/version, all-zero fanout, pack checksum, and index checksum are correct.
- Claim-v3 candidate claims are empty; macOS and Windows remain pending native evidence; Linux is excluded until `TASK-260728-1skseh`.

## Blocking rework findings

1. The exact receipt and mixed-marker oracle is internally invalid. For `conformance/v1/expected/external-repository/build-receipt-v2.json`, the declared `cache_key` is `sha256:4444444444444444444444444444444444444444444444444444444444444444`, but SHA-256 of `CCJ-1(input)` is `sha256:012564909df8f333004eb5aec867210d8973c74d9b71948e1f4fdb0d00c76559`. The mixed marker declares external `receipt_sha256` as `sha256:5555555555555555555555555555555555555555555555555555555555555555`, but SHA-256 of `CCJ-1(receipt)` is `sha256:0def2b595f6d086913bc65dcbeb40e937e5832d1e12c4aa8aa162953f7958969`. Protocol Core sections 9.1-9.2 require these exact relationships. The current validator and tests accept the mismatches, so a conforming implementation cannot use the published exact bytes as one integrated cache/currentness oracle.

2. The rc.5 external lifecycle corpus has no source-covering dry-run case and no audit-only case. The authoring guide says the full proof and independent audit order applies to a claimed cache hit and a source-covering dry run, but `external-repository-lifecycle.json` contains only cache, mixed-build, transaction, status/repair/GC, shim/PATH, and signing groups. The legacy `manager-lifecycle.json` dry-run cases forbid source fetch and do not prove this external-repository path. Add exact dry-run and coverage-claiming audit cases, and enforce audit-before-cache/compiler for hit, miss, dry-run, repair, and audit paths.

3. Two adversarial pack cases are descriptive rather than executable: `reject-index-checksum-mismatch` contains only `mutation: flip-final-index-byte` with no base fixture reference or concrete mutated bytes; `reject-pack-hash-family-mismatch` contains only checksum width and repository format. Both Python and Go validation merely require the case names, while positive pack tests check only magic. Publish concrete negative bytes or unambiguous base-plus-mutation references and make the harness verify fanout, trailers, checksums, hash-family mismatch, and exact expected errors.

## Required next producer actions

- Recompute the receipt cache key and marker receipt hash from CCJ-1, regenerate the exact files, manifest, release pin, docs, and tests, and add semantic assertions that fail on either mismatch.
- Add executable source-covering dry-run and audit-only ordering vectors and exact field assertions.
- Make the pack negative fixtures self-contained and cryptographically enforced.
- Re-run the same clean release qualification and attach revised hashes and evidence for review cycle 2.