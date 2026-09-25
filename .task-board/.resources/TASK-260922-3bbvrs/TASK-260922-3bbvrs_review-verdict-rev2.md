# TASK-260922-3bbvrs — review verdict, CR revision 2: CHANGES REQUESTED

Reviewed in a disposable clone at the candidate tree add49e0e (base 48da2690), with the conformance root from curator-spec dced9b83 (conformance/v1).

## Verified (holds)
- The harness `internal/conformancecoverage` checks the tally, duplicates, unpublished results, vanished ledger rows, unlisted failures, passing gaps and conflicting classes. The ledger `.github/ci/conformance-gaps.tsv` has family/case_id/owner/reason columns and 5 rows owned by BUG-260923-2afgyq.
- Consumer runs I did myself: marker v2 14/0/0/0; marker v4 22 driven + 5 known-gap = 27; system-config-v2 28; manager-config-v2 48; buildsource and conformancecoverage packages pass.
- End-to-end ratchet on the real marker v4 consumer:
  - Deleting the sha256 row makes the case fail as "not listed in the gap ledger".
  - Adding a row for passing valid-empty-builds.json fails with "now passes; remove its ledger row".
- Code mutants I applied myself:
  - Removing the passing-gap refusal: TestPassingKnownGapIsRejected FAILs (killed).
  - Replacing the ValidateTally call with nil: TestPublishedCaseTallyRejectsMissingClassification FAILs (killed).
- Rework 1:
  - skip-classes.tsv and platform-cases.tsv are unchanged, so no class was broadened.
  - no-broad-suppression.sh exits 0.
  - gate-selftest.sh reports 185 passed, 0 failed.
- SPEC_PIN is unchanged (ci.yml:53 dced9b83). The docs/ci-gates.md:55 paragraph and CHANGELOG.md:13-14 carry the owed-implementation rule.

## Findings (blocking — AC1 / review-note item 1: consumer list incomplete)
Some consumers still render every case in a published family, but not through `conformancecoverage`, and have no count pin:
1. `internal/devsub/repository_test.go:43-57` iterates every file in `schema-cases/skillfile-dev-v2` (16 entries at rc.12). A failing case has no ledger path, and a vanished case is not detected.
2. `internal/buildrepo/admission_test.go:191` iterates every case in `fixtures/external-repository/raw-objects.json` (12 cases) without classification or a count pin.
3. `internal/buildrepo/admission_test.go:240` iterates every case in `fixtures/external-repository/lfs-pointers.json` (10 cases) the same way.
Neither the producer's results nor the docs list these as out of scope. Route each one through `conformancecoverage.Run`/`RunOutcomes`, add its pin to `conformance-case-counts.tsv` (and to `root-artifacts.tsv` if required), or record a narrow, reasoned exclusion.
Also check the `pack-index.json` and `local-config-and-refs.json` loops (`admission_test.go` ~269 and ~396). Tell whole-family loops apart from filtered single-case lookups (buildmeta, whitelist and skillcheck filter by name, which is fine).

## Non-blocking
- Tally and passing-gap mutants are killed by named tests. The unlisted-gap mutant is killed per the producer's table, and I reproduced it end to end through the ledger edit above.
