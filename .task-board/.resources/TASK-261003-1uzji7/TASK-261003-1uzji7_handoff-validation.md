# TASK-261003-1uzji7: developer handoff verification

This follow-up follows uzji7-handoff.md: finish scoped lint, record findings in
the task outcome, reconcile the checklist, and hand off the existing candidate.
The prior implementation and red-first/targeted evidence in
TASK-261003-1uzji7_validation.md are accepted as prior-run evidence, not claimed
as commands rerun here. No full suite or hosted gate ran in this follow-up.

## Lint findings and changes

The first command below exited **1**, reporting three gosec findings and two
staticcheck findings. The signed-int to uint64 conversions preserve the version
number for nonnegative ints; negative versions cannot alias supported v1/v2
values. Added narrowly named G115 explanations at those two casts. Added a
narrow G304 explanation for the fallback read of a store target recomputed by
the verified materialization plan, matching existing store reads. Removed the
unused pre-lock lock/hash parameters from repairUnderLock: it already reloads
both values after taking the mutation lock. No migration semantics changed.

## Commands run here and actual exits

Each validation ran directly, without tee or a pipe chain.

- **Lint:** `GOFLAGS=-work golangci-lint run ./cmd/curator
  ./internal/conformancecoverage ./internal/contextlock ./internal/envprofile
  ./internal/hashing ./internal/install ./internal/marker`. First exit **1**;
  after the fixes, exit **0**, output `0 issues.` This is repository-configured
  lint over all seven touched Go packages, not a full repository lint run.
- **Targeted tests:** `GOFLAGS=-work go test ./internal/contextlock
  ./internal/envprofile -run
  '^(TestUnversionedLockObjectFollowsManagedWriterSwitch|TestResolvedDeltaIncludesStatePinWhenHashVersionChanges|TestParseRejectsHashVersionOnFrozenContextLock|TestRC14MigrationRehashesLegacyIdentities|TestRC14ResolveRepairMigratesSiblingHomes|TestRC14IdentityMigrationRefusesVersionAndPinMismatch|TestRC14IdentityMigrationPreservesFallbackCopies)$'
  -count=1 -timeout=4m -v`. Exit **0**. Two contextlock cases and four
  production migration cases passed, including marker-version, lock-version,
  and store-pin negatives. No migration test skipped. The named conformance
  lock case **skipped** because CURATOR_CONFORMANCE_ROOT was unset: this
  follow-up supplies no evidence for that case. Package times: contextlock
  0.840s; envprofile 29.680s.
- **Build:** `GOFLAGS=-work go build -o .temp/TASK-261003-1uzji7/curator
  ./cmd/curator`. Exit **0**. The ignored local binary is not attached.
- **Suppression policy:** `bash .github/ci/no-broad-suppression.sh cmd/curator
  internal/conformancecoverage internal/contextlock internal/envprofile
  internal/hashing internal/install internal/marker`. Exit **0**;
  `no-broad-suppression: ok`.
- `gofmt -w internal/contextlock/contextlock.go
  internal/envprofile/identity_migration.go internal/envprofile/managed.go`:
  exit **0**. `git diff --check`: exit **0**.

## Findings, decisions, and handoff limits

Findings and decisions are recorded here and in the existing design/validation
outcomes instead of LOGBOOK.md, as expressly required by the binding brief and
handoff instruction. LOGBOOK.md and CHANGELOG.md remain untouched.

The live checklist retained the old hosted-green item alongside the operator's
new targeted-checks item. The binding operator decision assigns hosted
acceptance to the CR gate. Remove only that superseded checklist entry; do not
represent an unrun hosted gate as passing. Scoped lint now supports the lint
item. The outcome record supports the logbook item under the explicit handoff
instruction.

Full suite, full repository lint, hosted/platform/race validation, and landing
after curator v0.15.0-rc.3 remain pending. The exact snapshot count and production
case were proven in the prior attached validation, not rerun here. The existing
handoff-blocker outcome is historical and superseded by this follow-up's scoped
lint and the operator's handoff policy decision. No commit, branch switch,
rebase, merge, or tag creation was performed. The candidate remains uncommitted
for the developer handoff snapshot.
