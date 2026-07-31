## Status
development

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(5))

## Blocked By
- BUG-260731-33v6zz

## Blocks
- (none)

## Checklist
- [x] Parse the current combined Windows raw go-test evidence and name all six failing top-level cases.
- [ ] Fix the transaction and fingerprint platform defects without skips, tolerances, or ledger relaxation.
- [ ] Add focused Windows regression coverage and preserve Linux/macOS behavior.
- [ ] Publish a signed Curator PR and prove the six cases green on native windows-latest.
- [ ] Attach evidence, obtain independent Opus review, and land the accepted PR to main.
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [ ] Verify no other Windows lane case regressed on the PR 15 run

## Notes
Created from RUN-260731-1b588e round-8 evidence. Scheduling dependency is BUG-260731-33v6zz / PR 13 so the residual fix is developed and validated against the combined Windows base and does not collide in internal/godriver. PR 12 is already merged as b30773b8.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260731-997143, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260731-997143)
Six residuals named from the combined-base windows-latest go-test JSON. internal/transaction: TestCommitAndRollbackRestoreALinkExactly / TestRecoveryFinishesAPreparedLinkTransaction / TestEntryRemovalRestoresTheExactLink / TestEntryTargetsDoNotAliasTheirDestination / TestNamespaceIdentityIsReadOnceWithinOneValidationPass. internal/godriver: TestFingerprintReportsUnreadableDirectoryIdentically. Three causes -- link destinations are host syntax because os.Symlink applies filepath.FromSlash on Windows; the namespace pass snapshot left the Windows file identity deferred to os.SameFile so it was neither a snapshot nor fail-closed; os.Chmod 0o000 leaves a Windows directory listable. PR 15 opened on task/BUG-260731-38dz6m-windows-residuals.
agent completed: [implementer] developer (claude) (exit=1)
spawn run completed: claude (run=RUN-260731-997143, pid=96578, exit=1)

## Precondition Resources
(none)

## Outcome Resources
- [BUG-260731-38dz6m_spawn-log_-implementer--developer--claude-_RUN-260731-997143.log](file://BUG-260731-38dz6m/BUG-260731-38dz6m_spawn-log_-implementer--developer--claude-_RUN-260731-997143.log) — System spawn log captured by task-board

## Created
2026-07-31T14:13:39Z

## Last Update
2026-07-31T15:39:35Z

## Assigned To
[implementer] developer (claude)
