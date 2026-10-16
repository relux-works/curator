# Deciding review B sweep
Candidate: 43b7ed0e9efd36ffcc321f6604d72aea221e2af7.
Coverage read before implementation: 2 of 3 scoped AC rows public-driven; identity cache priming explicitly internal. All three name driving tests. Original remainder: 0 of 5 workstreams complete; authorized next-coherent-part narrowing applies.
No catalog table or time budget supplied. Use scoped rows R1 Resolve/repair, R2 Status, R3 Migration/identity, in that order. Prior reviewer attacks are inspected and rerun, not counted as newly authored tests.

## R1 — held
Reran TestReviewTransientPolicyRead through Resolve/repair: first permission failure refuses with environment_seed_unreadable, no document and unchanged protected snapshot; next public operation recovers. Exit 0 in row1.log. Candidate projection tests also cover valid file versus isolated keyring/auto and fresh later reads.

## R2 — held
Reran TestReviewStatusBackendSwitch through StatusOf: file-backed home becomes non-current for keyring and invalid selector; invalid diagnostic retained; restoring file recovers current status. All observations preserve protected state. Exit 0 in row2.log.

## R3 — held
Reran production PlanMigration/ApplyMigration migration family, shared-to-isolated explicit migration and RC14 identity integration; exit 0 in row3.log. Includes plan/marker drift, refusal before mutation, rollback, interrupted apply recovery, unexpected recovery target, no secret copies and valid controls. IdentitySubrequest internal-seam test and M7 separately exercise recorded-isolation cache invalidation. No bypass reproduced.

## Free hunt — no blocking findings
Reran TestReviewDetachedBackendSwitch through Resolve --repair, both intact link and detached synthetic regular credential. Backend switch refuses environment_credential_conflict with unchanged protected snapshots; restored file backend admits the intact-link positive control. Exit 0 in free-hunt.log.

## Normative trace and scope
Fresh spec main: 1059b94b4135089d842bf017070b8ab47d79a536. Read login-readiness sections 2/2.1/5, environments 7.4/8.2/10.1-10.3/12.1-12.2 and profile-codex-settings 1-6. New changes since 15d9701 add separately gated local node source; that remains outside this native-only candidate. Native auto shared behavior is expressly retained by profile-codex-settings section 5 when no explicit store is selected.
R1 traces native inspection/isolation/link records to environments 7.4 and login-readiness 2/5. R2 retains environments 7.4 liveness and unreadable-state distinction, without claiming login readiness. R3 preserves identity-only migration (10.1), complete records (8.2), and explicit credential migration (7.4).
No catalog surface table or budget supplied. The three scoped acceptance rows are the sweep floor. The deferred adapter and paired-table rows are detailed in the producer coverage map; they are not represented as delivered. Hosted override plant remains deferred with successor delivery. No source-text gate is changed.
