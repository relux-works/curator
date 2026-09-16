# TASK-260908-2kqocx — B6 closure evidence
Date: 2026-09-16. Evidence-only developer handoff; curator Story worktree untouched.

## Scope and landing
Epic goal-file goal-launcher-and-infra-migration.md B6 (lines 134–141) is implemented on agents-infra's own board: EPIC-260916-ej5yuv / STORY-260916-1vt3x2, inventory TASK-260916-38vqh4, implementation TASK-260916-3q6z2c. The later b6-closure-brief.md directs this task to capture evidence only.

Inspected /Users/administrator/Developer/ReluxWorks/relux-agents-infra-main on branch main:
- Signed implementation commit 0f4b7d00557aa688d95b29640af0a4b1caf33372, CR-TASK-260916-3q6z2c-6 rev 6.
- Board record commit 1c594f8613fa57d7a90b9745a9fb5ca1aee1e871.
- Implementation tree d91f1a1e3a7fd2420a586a127d774c332e975f18 equals CI snapshot 008bb38d9258c56e4c23bfae98978837f10a3b15 tree.
- git verify-commit 0f4b7d0: exit 0, good SSH signature for bot@relux.works, ED25519 SHA256:qbALzjdB9BRgYJjDkX/p9EAPLEofB2AbskJc6Ftwhds.
- git log -3 and git show --stat 0f4b7d0 inspected; 32 changed files.
This establishes the local main landing. The integration-results resource explicitly lists hosted branch/PR delivery and reconcile-trunk as next steps; this packet does NOT attest a merged hosted PR. That external delivery remains orchestrator-owned.

## Producer, independent review, and hosted gates
Sources under the agents-infra checkout .task-board/.resources/TASK-260916-3q6z2c/:
- TASK-260916-3q6z2c_results.md (producer, revisions 1–6).
- TASK-260916-3q6z2c_review-verdict-rev6.md: ACCEPTED, CR rev 6, exact tree above, base 459742ea67e3c6b84169520b92d74fe7f73e3002.
- TASK-260916-3q6z2c_change-request_rev6-validation.log: sh scripts/remote-gate.sh, exit 0, exact-command-shard 1/1 green.
- TASK-260916-3q6z2c_change-request_rev5-validation.log: earlier gate exit 0; historical only, not evidence for rev6.
- TASK-260916-3q6z2c_integration-results.md and b6-inventory-report.md.

Rev6 hosted run: https://github.com/relux-works/relux-agents-infra/actions/runs/35067793801
- macOS job 104701958982: success.
- Ubuntu job 104701959379: success.
Fresh gh run view in this closure run exited 0 and confirmed both jobs, including format and build/vet/test steps.
Earlier rev5 run: https://github.com/relux-works/relux-agents-infra/actions/runs/35064414410 (both lanes success according to its validation log; not freshly queried).
Workflow .github/workflows/ci.yml runs gofmt check, go build ./..., go vet ./..., go test ./... -count=1 -timeout 20m. scripts/remote-gate.sh snapshots the candidate and waits for that workflow.

Accepted existing independent reviewer evidence: build/vet exit 0; targeted main tests exit 0 (14.012s); targeted infra tests exit 0 (13.854s); full suite accepted from exact-tree hosted CI. Reviewer narrowing probes: 1/2 killed, 1/2 inconclusive due to process startup stalls; no demonstrated survivor, no claim of 2/2 independent kills. Producer rev6 records 2/2 killed and post-restoration green checks.

Production-entry deprecation tests read:
- TestDeprecatedProviderErrorMessagesAreExact
- TestRunCodexAndClaudeAreDeprecatedForEveryArgumentShape
- TestDeprecatedEntrypointsExitOneWithNoSideEffects (reviewed 58/58 executable cases)
- TestInstalledProviderAliasesReportDeprecationWithoutLaunching (producer 16/16)
- TestDeprecatedCanonicalAliasesRefuseWithoutSibling
- TestDeprecatedDirectProviderYoloAliasesRefuseWithoutSibling
- TestDeprecatedCanonicalWrappersRefuseWithoutDelegation
- TestCodexLocalLauncherRefusesWithoutDelegation
Executable cases check exit 1, exact stderr, empty stdout, no provider/Curator sentinel, unchanged fixture project/home, and refusal before build with absent cache and failing go.

## Operator decision and residual
DEPRECATE now; remove entrypoints next release. Do not describe the stubs as already removed. agents-infra codex|claude, target openai-infra|anthropic-infra, dange aliases, and .local/bin/codex-local print migration notices pointing to curator run codex_cli or curator run claude_code and exit 1, without launching/delegating/building. --print-config is also refused.

Removed from setup/refresh and provider launching: instruction sync and @ rendering, skill fan-out/validation, bundled MCP registry distribution and provider-launcher MCP composition.
Kept:
- claude-settings.json linking; Codex config.toml merge/modes and .rules.
- Pi local-model runtime, profiles, targets, harness, shared broker, runtime-launch, status/stop, quarantine/unquarantine, live Pi/qwen entrypoints.
- Attachment manifest/helper contract.
- task-board compose/prepare schema-v1 contracts, strict caller-supplied registry composition, and isolated prepare-only legacy instruction renderer.
- Source .instructions/ and .skills/ remain provenance/source material; pre-existing installed files are preserved, not newly distributed.
No LLDB wrapper exists in this repository revision; none was invented or claimed retained. This is a repository bound, not a claim about third-party availability.

README opening, deprecation notice, instruction/skills sections and LLDB section inspected. Some headings use “retired” for disabled launches; actual operator lifecycle is deprecation stubs this release, removal next release.

## Slice B follow-up
Retain v1 renderer/registry consumer compatibility until task-board changes its compose/prepare contract. task-board validators currently require real instruction artifacts. Coordinate the consumer contract change and evidence first, then delete the remaining v1 renderer/registry compatibility path in a later agents-infra slice. Do not fake rendered states or silently change frozen v1 schemas. Next-release stub removal is a separate release-owner action.

## Validation performed here and bounds
No product code, tests, docs or config changed; no Go test/build or full remote gate rerun in this evidence-only task. All test/build results above are explicitly accepted existing evidence.
Direct read-only verification: git verify-commit exit 0; gh run view exit 0; git rev-parse of both trees exit 0 with equal OIDs; git status --porcelain in curator worktree exit 0 and empty before/after inspection.
Windows native execution, live model/provider/authentication and external task-board decoder replay remain unverified. POSIX fixtures and hosted Ubuntu/macOS are the demonstrated platforms.
The initial required development transition exited 1 because an estimate was missing; set_estimate Fibonacci 1 exited 0 and retry exited 0. No test failure was hidden by this lifecycle repair.

## Task-scoped logbook
2026-09-16: Verified accepted rev6 exact-tree local main landing, bot signature and hosted CI. Recorded deprecate-not-remove decision and Slice B dependency. agents-infra checkout has existing board-only working changes and an untracked integration receipt; left untouched. Integration receipt does not establish merged hosted PR delivery. No repository LOGBOOK.md edits allowed by campaign rules; no standalone logbook CLI is installed, so this resource and board note persist the findings. Curator worktree remains clean.
Checklist “Code written per task description and AC” refers to the landed agents-infra implementation; this closure task authorizes no new code.
