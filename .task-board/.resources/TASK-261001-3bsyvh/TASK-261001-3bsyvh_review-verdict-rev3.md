# TASK-261001-3bsyvh rev3 — identity review verdict: ACCEPTED

Candidate: CR rev3, base e87d488b, tree 9cf8a21b (index write-tree equals it), 17 paths.
Normative: refs/campaign/rjxrgs-rev1-20261001 (fe2b5f61, base 5432c85f, tree ac008990). Substantive review: TASK-260728-rjxrgs_review-verdict (rev1, accepted).

## 1. Non-tsv paths
- Path lists equal (17 = 17, `diff` empty).
- Per-path two-way +/- line multisets (`git diff -U0`, sorted) equal for ALL 17 paths, tsv included.
- Blob identity vs rjxrgs rev1: 11/14 non-tsv paths byte-identical; protected.go, external.go, install.go, schema_coverage_test.go differ from fe2b5f61 only by trunk movement (these four are among the six paths trunk changed between 5432c85f and e87d488b). External.go/protected.go are byte-identical to rev2 (4a3d18c7); install.go and schema_coverage_test.go differ from rev2 solely by e87d488b trunk movement.
- Proof: `git merge-tree --write-tree --merge-base=5432c85f e87d488b fe2b5f61` yields blobs equal to the candidate for 16/17 paths (every non-tsv path and conformance-gaps.tsv, root-artifacts.tsv). The only mismatch is conformance-case-counts.tsv, where merge-tree conflicts (both sides add rows at the same spot) — i.e. the expected union case.

## 2. tsv union / counts
- case-counts: candidate vs trunk has 0 removed lines, exactly the 4 rjxrgs rows added (mixed_build_cases 6, path_shim_cases 3, signing_cases 4, transaction_cases 4), identical to the rjxrgs-added rows. No duplicate (digest,family) keys. Header/candidate-manifest rows untouched.
- Independent recompute: clone of curator-spec at rc.13 commit 23435129; manifest sha256 = be11bb1e… (matches the keyed digest); vectors/external-repository-lifecycle.json has 6/3/4/4 cases = the TSV rows.
- gaps: +5 marker/install-marker-v3 rows over trunk, equal to rjxrgs; root-artifacts: +1 internal/install row, internal/marker row extended with schema-cases/install-marker-v3, equal to rjxrgs.

## 3. Strays
- No path outside the 17; no .task-board, CHANGELOG, LOGBOOK, TASK-260728-rjxrgs_results.md in the tree. Added content is byte-/multiset-identical to the already-accepted delta, which carried no employer name (I did not spell or search for one beyond that identity).

## Executed by me (real exit codes)
- `go test ./internal/conformancecoverage -count=1` → ok, exit 0.
- `go test ./internal/install -run 'TestLegacyMixedBuildProjectInstallProducesMarkerV3|TestExternalCommandNameCollisionFailsBeforeMutation|TestGlobalMixedBuildStagesExternalBeforeLocal'` → PASS, exit 0.
- With CURATOR_CONFORMANCE_ROOT = rc.13 clone: TestAuthoritativeMixedBuild/PathShim/Signing/TransactionCasesUseProjectInstallEntry → all PASS (43.6/26.8/5.0/25.9 s), exit 0 (without the root they SKIP, so the root was required).
Not rerun by me: the full gate (accepted from attached green gate evidence) and the marker schema_coverage / buildrepo / scopes packages.

Bound: identity is shown by diff algebra and merge-tree; the substantive correctness rests on the rjxrgs rev1 verdict.