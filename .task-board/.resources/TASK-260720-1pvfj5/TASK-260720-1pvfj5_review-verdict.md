# TASK-260720-1pvfj5 independent final review verdict

Date: 2026-07-30
Role: reviewer
Verdict: CHANGES REQUESTED
Route: to-dev

## Material finding

P1 — candidate evidence can stamp a revision that does not identify the tested root. The workflow permits candidate_ref and candidate_root together at .github/workflows/ci.yml:303-305. It verifies and checks out candidate_ref at lines 316-329, exports both inputs at lines 349-355, then prefers candidate_root at lines 357-362. candidate-suite.sh does not bind the selected root to the revision: it copies CANDIDATE_REF into ref at line 98 and writes that value as candidate_revision at line 128 while the manifest/tree digests at lines 131-132 come from the independently selected root. Therefore a dispatch with revision A and root B tests B but publishes evidence naming A. This violates the acceptance requirement that candidate evidence record the exact suite revision and digest and undermines the anti-impersonation boundary. The current rc.5 root-only evidence is truthful; the defect is in the supported CI input surface.

Required rework: reject the both-nonempty input combination before checkout/recording, or cryptographically prove that the selected root belongs to candidate_ref before stamping that revision. Add a focused gate regression showing the ambiguous combination fails and that the single-ref and single-root paths remain valid. Preserve SPEC_PIN, candidate-only wording, the seven accepted product paths, and all unrelated overlay bytes. No heavy Go suite needs to be repeated solely for this input-validation change; rerun the affected gate self-test/static workflow checks and preserve/reuse the existing serialized full/race evidence when byte identity permits.

## Evidence that remains accepted

The final prepatch manifest is byte-identical to the accepted 372-entry rework composite; the live 374-entry composite matches every recorded manifest identity; the exact delta is five modified plus two added accepted blocker-owned product paths; all 16 prior CI/quality paths are unchanged; both accepted patch artifacts are byte-identical to their independently accepted board resources. SPEC_PIN appears once and remains 00b1688a9b2457ca397a0bb550acf47cad8ee967; Go remains 1.25.5; rc.4 wording is absent. Candidate manifest/tree identity remains b6f56aac...04c / e6a13215...2fae with 448 files and candidate-only/no-release wording. Attached default, rc.5 candidate, and single serialized race gates exited 0 in 371.993s, 354.229s, and 471.263s; lint v2.12.2 reports 0 issues; format, vet, build, no-broad-suppression, deterministic cancellation, selftest 70/70, and ledger 49 rows are green. These facts do not cure the ambiguous-input provenance defect.