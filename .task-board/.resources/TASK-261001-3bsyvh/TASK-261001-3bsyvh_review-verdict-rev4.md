# TASK-261001-3bsyvh rev4 — identity review verdict: ACCEPTED

Candidate: CR rev4, base 67d83539, tree 289196d5 (index write-tree equals it), 17 paths, +978/-43.
Normative: refs/campaign/rjxrgs-rev1-20261001 (fe2b5f61, base 5432c85f). Substantive review: TASK-260728-rjxrgs_review-verdict (rev1, accepted). Previous verdict (rev3) was ACCEPTED, no findings to answer; repeat-of sweep: none.

## 1. Identity (non-tsv and tsv)
- Path lists equal across rev4, rev3 (d540fbb1) and rjxrgs rev1: 17 = 17 = 17.
- Per-path two-way +/- line multisets (git diff -U0, header lines stripped, sorted) equal to BOTH rev3 and rjxrgs rev1 for all 17 paths.
- Blobs vs rev3: 14 of 17 byte-identical; the 3 that differ are install.go, conformance-case-counts.tsv, conformance-gaps.tsv. For each, the rev3->rev4 line delta equals exactly the trunk e87d488b->67d83539 delta of that path (sorted +/- lines compared, diff empty).
- `git merge-tree --write-tree --merge-base=e87d488b 67d83539 d540fbb1` exit 0 clean, tree 289196d5 = candidate tree: all base movement and trunk-only paths proven.

## 2. tsv union / counts
- case-counts: vs trunk 0 removed, 4 added rows (be11bb1e… mixed 6, path_shim 3, signing 4, transaction 4); no duplicate (digest,family) key.
- gaps: +5 marker/install-marker-v3 rows, 0 removed; root-artifacts: +1 internal/install row, internal/marker row extended with schema-cases/install-marker-v3. All equal to the rjxrgs additions. Trunk (Muse and other) rows untouched.
- Independent recompute: /tmp/curator-spec-rc13 at 23435129, manifest sha256 be11bb1e…, vectors/external-repository-lifecycle.json arrays mixed_build_cases 6, path_shim_cases 3, signing_cases 4, transaction_cases 4 = TSV rows.

## 3. Strays
- No path outside the 17; no .task-board, CHANGELOG, LOGBOOK or rjxrgs results file in the tree; `git diff --cached --check` clean. Added content is multiset-identical to the already accepted delta (no new name-bearing content).

## Executed by me (real exit codes)
- `go test ./internal/conformancecoverage -count=1` -> ok, exit 0.
- With CURATOR_CONFORMANCE_ROOT=/tmp/curator-spec-rc13/conformance/v1: `go test ./internal/install -run 'TestAuthoritativeMixedBuildCasesUseProjectInstallEntry|TestLegacyMixedBuildProjectInstallProducesMarkerV3|TestExternalCommandNameCollisionFailsBeforeMutation|TestGlobalMixedBuildStagesExternalBeforeLocal'` -> all PASS, exit 0.
Not rerun by me: full gate, marker/buildrepo/scopes/skillspec packages, path-shim/signing/transaction rows (accepted from attached developer evidence and green gate).

Bound: identity is shown by diff algebra and merge-tree; substantive correctness rests on the rjxrgs rev1 verdict.