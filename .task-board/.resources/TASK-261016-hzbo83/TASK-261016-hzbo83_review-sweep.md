# Review sweep

Candidate tree: 43b7ed0e9efd36ffcc321f6604d72aea221e2af7.
Base and freshly fetched main: 782823d849841ae74efa160243cbac56721aa733 (zero commits behind).
Normative source: curator-spec main 15d970111277b1f68b3f7a4ddd63bc339b2c0a3f.
No catalog surface table or review time budget was supplied. Use the three
scoped AC rows from the attached coverage map as the sweep floor, ordered as
published; free hunt follows. The remaining original rows are explicitly
outside this authorized next-coherent-part candidate, not satisfied features.

## R1: Resolve and repair policy observation — held

Public entry: Resolve -> assembleHome/finalizeMarker/checkPassthrough ->
resolveCredentialPolicy. Ran candidate TestCredentialProjectionSingleObservation,
TestCredentialProjectionIsolationRefusals and TestCredentialStorePolicy* via
focused.log (exit 0). Independent TestReviewTransientPolicyRead via row1.log
(exit 0) injects a first permission failure followed by readable state. Repair
refuses with environment_seed_unreadable, emits no document and preserves the
snapshot, then a subsequent public operation succeeds. This attacks failure
laundering through retry and sticky cross-operation caching; neither reproduced.
The baseline valid/invalid selectors and unsupported mappings have positive
controls. Narrowing controls M1-M6 are recorded separately; no production
source was changed by the review.

## R2: Status observation and unreadable-state preservation — held

Public entry: StatusOf -> homeState -> verifyHome -> assembleHome ->
resolveCredentialPolicy. Candidate TestCredentialProjectionStatusObservation
passed in focused.log. Independent TestReviewStatusBackendSwitch passed in
row2.log: a previously file-backed home becomes non-current under valid keyring
and invalid native selectors, the invalid diagnostic is retained, restoring
file makes it current again, and every status call preserves protected state.
The permission-failure candidate test checks unreadable is not absence and
that later reads recover. These attacks did not reproduce a bypass.

## R3: Migration and identity-subrequest policy — held

Public entries: PlanMigration/ApplyMigration -> inventoryHome ->
resolveCredentialPolicy; Resolve -> repairUnderLock -> prepareIdentityMigration.
row3.log exits 0 for shared-to-isolated repair refusal followed by explicit
migration, plan drift rejection, rollback, interrupted apply/recovery,
unexpected-target rejection, no secret copies, and the RC14 identity migration
family. M7 exits 1 with TestCredentialProjectionIdentitySubrequest/isolated;
other adapters still clear their cache, so it is a narrowing control.
Priming a differing caller projection remains an explicitly internal test
seam, not claimed as a public-drive of that specific condition. The existing
public integration tests exercise migrated identity and recorded state.
No bypass or unintended refusal reproduced in these attacks.

## Free hunt — no blocking finding

Independent TestReviewDetachedBackendSwitch (free-hunt.log, exit 0) changes
native file to keyring after provisioning, once with the intact recorded link
and once with a detached synthetic credential. Resolve --repair refuses with
environment_credential_conflict and preserves native/managed bytes, links and
markers. Restoring file admits the intact-link positive control. This attacks
implicit migration and credential deletion during a backend change.

All seven narrowing controls were rerun by this reviewer using Go overlays,
not source edits. All exited 1 with their named behavioral failure. No static
checker exists in this delta; token-preserving source-checker controls do not
apply. No full mutation sweep or supported-release runtime qualification is
claimed. Full CI is recorded separately after terminal status.
