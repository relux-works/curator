# TASK-260927-25hk87 — flip-security-posture-default-hardened: revision-2 review

Verdict: **accepted**. No findings requiring rework. LANDING HELD until the operator schedules the B release; this review authorizes acceptance only, not integration or release.

Reviewed CR-TASK-260927-25hk87-2 revision 2, base ff8f75a0209928bfeae94306f9171c8c8159cb87, candidate tree af82499fd08af44beeb05f50162fa107fafa2f4b. All 19 changed paths were inspected; all 19 working-file blob identities match the immutable candidate. Fresh origin/main was independently advertised and fetched at 77fabd45b88b7fd839a85f017d863e03c926953e: 0/19 candidate paths overlap its changes since the CR base. No code was modified by this reviewer.

## Swept surfaces

| Surface | Result and evidence |
| --- | --- |
| Release prerequisite | GitHub release v0.15.0-rc.3 is published, not draft (2026-10-04T03:03:40Z). Its internal/config/security_posture.go declares A. The warning release preceded this candidate. |
| Default and precedence | SecurityPostureRevision selects B through config.Load -> parseConfig -> defaultSecurityPosture/applyHardenedDefaults. Omitted schema-2 posture selects hardened; schema 1 remains permissive. Explicit permissive retains advisory audit/registry, drop, optional signers and original allowlist/passthrough defaults, with the operation warning. Explicit knobs retain precedence; locked-posture and schema-1 system-config controls pass. Matches pinned manager profile sections 1 and 7.1. |
| Production refusals and controls | run() -> cmdInstallMode/cmdUpdate -> CheckSecurityPosture refuses both absent and explicitly empty source lists across install, update, upgrade and profile install, before publication. Profile MCP refusal runs with genuine signed Git dependencies; empty MCP without declarations proceeds. Explicit permissive profile install succeeds and warns once. env status --check reproduces non-current and actual MCP declarations; env resolve refuses explicit unbounded passthrough. |
| Vector accounting | All 6/6 rollout-B vectors drive production entry points. Config: 12 driven + 5 bound + 0 known-gap + 0 skipped = 17. CLI: 13 driven + 4 bound + 0 known-gap + 0 skipped = 17. Exact class totals are asserted. Four B deferrals are removed from each consumer. No security-posture gap ledger rows existed at this base; unrelated Codex seed rows remain unchanged. |
| Collateral scope and architecture | Only one product-code file changes: existing revision selector and migration hint. The other changes are tests, fixtures and docs. CodexSeedRevision remains A and its registry file is unchanged by this CR; other gate selectors are unchanged. Manager-vector expectation adjustment changes only absent posture defaults, preserves explicit values and does not mutate corpus expectations. No new policy mechanism. |
| Documentation and protected files | Documentation states the B default and compatibility behavior and reports exact bounded coverage. CHANGELOG.md, LOGBOOK.md and scripts/remote-gate.sh have no CR delta. This artifact carries review findings/logbook evidence. |

## Hosted CR gate — current candidate

[Revision-2 hosted gate 37293008806](https://github.com/relux-works/curator/actions/runs/37293008806) completed successfully. Independently queried with gh run view. Its snapshot c8aae18349b5e26a32d5dafe5abcacbe40c4a760 has tree af82499fd08af44beeb05f50162fa107fafa2f4b, exactly the reviewed candidate, and parent ff8f75a0209928bfeae94306f9171c8c8159cb87. The attached CR resource TASK-260927-25hk87_change-request_rev2-validation.log records the hosted command exit 0.

20/20 executed jobs green: full Test on Ubuntu/macOS/Windows; Race on Ubuntu/macOS; lint; naming; interop; three gate self-tests; nine Go driver jobs (three versions across three platforms). Two workflow-declared skips: rose-air and alternate candidate-suite lane. This is the authoritative current-tree full-suite/build/lint evidence; the older pre-convergence run is not reused to qualify this CR. Independent condensed verification is attached as TASK-260927-25hk87_review-hosted-rev2.json.

## Reviewer reruns versus accepted evidence

Reran independently on the candidate, sequentially under the shared build lock, GOFLAGS=-work, -p 1, -count=1 and the clean SPEC_PIN 43bf0a2506d5c354a73bbc3ea4623d4653db10c7 conformance root:

- go test -timeout 2m -v ./internal/config -run 'SecurityPosture|ManagerVectorPostureDefaults': exit 0, 1.255s; exact posture tally 12/5/0/0, all B vectors driven.
- go test -timeout 3m -v ./cmd/curator -run 'SecurityPosture|HardenedRegistryOutage|PermissiveRegistryOutage': exit 0, 52.286s; exact posture tally 13/4/0/0, all B vectors driven; default source refusals, permissive success/warning, status and registry controls pass.
- git diff --check on the immutable CR delta: exit 0.

Neither test command skipped a test. Sanitized output and actual command exits are attached as TASK-260927-25hk87_review-validation-rev2.log. Bounds returned by the coverage harness are classified bounds, not executed historical-A behavior.

Accepted from attached producer evidence, not rerun by this read-only reviewer: TASK-260927-25hk87_validation.log contains two expected-red mutations (each actual go test exit 1): reverting B to A breaks default tests; narrowing the empty-source gate to nil-only breaks all four explicit-empty production command cases. The mutated production file and both affected test files are byte-identical between that evidence's candidate and revision 2. No mutation was performed here. Full suites, golden regressions, build and lint are accepted from the current hosted CR gate.

Bound: historical implicit-A vectors cannot describe the shipped B default and are explicitly accounted as bounds; config-level registry I/O is bounded and tested at the CLI. No whole-repository mutation-coverage claim is made.

## Lifecycle

Queried task-board spawn goal before verdict: this run is not goal-bound. All 11 live checklist items are checked. Record acceptance with accept_cr at revision 2; route to integrating. Do not mark done, supply commit_ack, checkpoint, integrate, or release. The operator's B-release scheduling hold remains in force.
