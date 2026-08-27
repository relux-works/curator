## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(3))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] Fixture: vendored third-party Makefile with curl-pipe line
- [x] Non-executable vendor text admitted; executable vendor script and first-party text still block — or evidence-backed gap note
- [x] go test green
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-205910, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-205910)
VERDICT: gap branch of the AC. The external-repository static audit does not block on non-executable vendor text because that audit subject is not implemented here — internal/skillspec/types.go:9 caps SupportedSchemaVersions at 5, and there is no go-repository-v1 driver, no skill-build.json, no go list/go build child anywhere in the Go tree (grep -rnE "go-repository-v1|build_repository|skill-build|BuildRoot" --include=*.go internal cmd returns nothing). Spec basis: curator-spec@a2d44eb protocol/core.md:531 and :844 place the whole-external-snapshot audit on the schema 6/7 subject; profiles/manager.md:616 makes the detector set implementation-defined while the decisions of :626-632 are fixed. No demotion-to-advisory code was written: there is nothing blocking to demote.

MEASURED: the audit surface that does exist scopes its detectors to the literal prefix scripts/ plus csk-skill.json (internal/audit/audit.go:307). The reproducing fixture (vendored third-party Makefile with a curl-pipe recipe, an executable script inside the same vendored module, the same line in first-party scripts/) yields exactly one finding, on scripts/setup.sh, and strict fail_on=low blocks on that file only. A vendor-only snapshot installs clean with zero warnings.

GAP 1 (asked-for half missed): an executable file below vendor/ is admitted too, for the same structural reason — scope, not policy. Not exploitable today: vendor/ is excluded from installed context by internal/whitelist and internal/skillspec/parse.go:546 rejects command paths outside runtime_roots.

GAP 2 (found while proving it, wider blast radius, not vendor-specific): the detector scope is the literal scripts/ prefix, not the declared runtime_roots. A snapshot with runtime_roots ["bin"] and bin/tool containing an undeclared-host curl yields 0 findings. Every skill naming its runtime root anything but scripts ships with zero audit coverage of its executable surface, at schemas 1-5, today. Not fixed here — widening the scope changes install outcomes for every existing skill and is a policy decision, not a bugfix.

DELIVERED: tests only, no product code. internal/audit/vendor_test.go (TestVendorTextIsOutsideTheDetectorScope, TestGateAdmitsVendorTextAndBlocksFirstPartyText, TestVendorOnlySnapshotIsAdmitted, TestRevocationStillBlocksVendoredSnapshot) and internal/install/install_test.go::TestAuditAdmitsVendoredThirdPartyText (end-to-end install under strict/fail_on=low, vendor/ absent from installed context, plus a negative control moving the identical curl-pipe line into scripts/ which fails the install with network-undeclared).

GATES (each run standalone, real exit codes): go build ./... 0; go vet ./... 0; gofmt -l internal cmd 0 with empty output; go test ./... 0, 31 ok / 0 FAIL; golangci-lint run 0, 0 issues.

EVIDENCE: TASK-260822-a6jpu9_vendor-inert-text-audit-policy.md and TASK-260822-a6jpu9_go-test.log (board outcome resources); .research/260823_vendor-inert-text-audit-policy.md; LOGBOOK.md 2026-08-23 0001.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-205910, pid=60226, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260822-2559c8, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260822-2559c8)
REVIEW VERDICT: ACCEPTED (run RUN-260822-2559c8, read-only, no worktree file modified). Evidence: TASK-260822-a6jpu9_review-verdict.md.

RE-DERIVED, NOT READ FROM THE REPORT. (1) The external-repository audit subject genuinely does not exist: internal/skillspec/types.go:9 caps SupportedSchemaVersions at 5; grep -rn -E "go-repository-v1|build_repository|skill-build|BuildRoot" --include=*.go internal cmd exits 1 with no matches; no go list / go build child anywhere. So the task descriptions second clause (no such audit exists -> record the gap) is the correct branch and there was nothing blocking to demote. (2) Spec citations verified verbatim at curator-spec a2d44eb: core.md:531, core.md:844, manager.md:616, manager.md:626-632. (3) Detector scope is the code as claimed, audit.go:307, literal scripts/ prefix plus csk-skill.json. (4) KEY CHECK the report did not make explicit: install.go:261-277 builds each audit.Subject from node.Snapshot, the full repo snapshot with vendor/ present, so vendored files are skipped by scope, not by absence. That is what makes the answer load-bearing. (5) Non-exploitability holds structurally: whitelist.go:19-22 IncludeRoots is an allowlist without vendor; parse.go:546 rejects command paths outside runtime_roots.

GATES re-run standalone with direct exit codes (not PIPESTATUS, empty under zsh): go build ./... 0; go vet ./... 0; gofmt -l internal cmd 0 empty; go test -count=1 ./... 0 with 31 ok / 0 FAIL; go test -count=1 -run Vendor ./internal/audit/... ./internal/install/... 5/5 PASS; golangci-lint run 0 issues.

WHY THE MISSED AC HALF IS NOT REWORK: making an executable vendor file block requires widening the detector scope past the literal scripts/ prefix, which changes install outcomes for every existing skill at schemas 1-5. That is the policy decision the task description routes to a follow-up, not a bugfix, and forcing it here would be the forced fit the role contract refuses.

NON-BLOCKING FINDINGS. F1: the adjacent gap needs a board item and has none. Detector scope is the literal scripts/ prefix, not the declared runtime_roots, which are arbitrary (parse.go:482, copied by runtimestore.InstallRuntimeRoots). A skill with runtime_roots [bin] gets zero audit coverage of its executable surface today, at schemas 1-5, independent of schema 6/7. Captured in LOGBOOK and the research note, but STORY-260822-2jdj9o has no child for it (children 27bvo4, 5wfdfx, a6jpu9). Recommend the orchestrator open one; a reviewer run should not widen the board unilaterally. F2: TestVendorTextIsOutsideTheDetectorScope is a deliberate change-detector, its len(findings)!=1 assertion fails the moment the scope widens; by design and documented in the test comment, but the follow-up owner should expect to edit it rather than read it as a regression. F3 cosmetic: TestGateAdmitsVendorTextAndBlocksFirstPartyText joins two positive assertions with &&, so it passes if either matches; both do today.

SCOPE: no commit_ack supplied. This tasks delta is internal/audit/vendor_test.go (new), the TestAuditAdmitsVendoredThirdPartyText block in internal/install/install_test.go, .research/260823_vendor-inert-text-audit-policy.md, and LOGBOOK.md 2026-08-23 0001. Sibling-task work is uncommitted in the same tree; the commit-owning mover lands this scope.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-2559c8, pid=76877, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260822-a6jpu9_spawn-log_-implementer--developer--claude-_RUN-260822-205910.log](file://TASK-260822-a6jpu9/TASK-260822-a6jpu9_spawn-log_-implementer--developer--claude-_RUN-260822-205910.log) — System spawn log captured by task-board
- [TASK-260822-a6jpu9_vendor-inert-text-audit-policy.md](file://TASK-260822-a6jpu9/TASK-260822-a6jpu9_vendor-inert-text-audit-policy.md) — Verdict, evidence, reproducing fixture and recorded gap for the vendor-inert-text audit policy question
- [TASK-260822-a6jpu9_go-test.log](file://TASK-260822-a6jpu9/TASK-260822-a6jpu9_go-test.log) — go test ./... full-suite output, exit 0, 31 ok / 0 FAIL
- [TASK-260822-a6jpu9_spawn-log_-reviewer--reviewer--claude-_RUN-260822-2559c8.log](file://TASK-260822-a6jpu9/TASK-260822-a6jpu9_spawn-log_-reviewer--reviewer--claude-_RUN-260822-2559c8.log) — System spawn log captured by task-board
- [TASK-260822-a6jpu9_review-verdict.md](file://TASK-260822-a6jpu9/TASK-260822-a6jpu9_review-verdict.md) — Reviewer verdict: ACCEPTED on the gap branch, with independently re-derived AC mapping, spec citations, gates, and three non-blocking findings

## Created
2026-08-22T16:12:28Z

## Last Update
2026-08-22T20:08:42Z

## Assigned To
[reviewer] reviewer (claude)
