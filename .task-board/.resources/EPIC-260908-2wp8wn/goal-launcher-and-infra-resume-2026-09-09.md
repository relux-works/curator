# Continuation goal: finish curator run and the agents-infra migration

Checkpoint: 2026-09-09 local (Asia/Tbilisi), parked by the operator at 2026-09-08 23:18 UTC. This file describes a continuation, not permission to restart work during the parking turn. A new explicit request to execute this goal authorizes resumption.

## Mission and authoritative inputs

You are the orchestrator. Work from `/Users/iv/Developer/ReluxWorks/curator-agent-launcher` with the EXISTING shared product board `/Users/iv/Developer/ReluxWorks/curator/.task-board` and epic **EPIC-260908-2wp8wn**. Do not create a replacement epic or repeat completed implementation/research.

Read this file first, then the original plan and inventory beside it:
- `goal-launcher-and-infra-migration.md` — original scope, sources and DoD;
- `agents-infra-to-curator-mapping.md` — migration inventory at dee5403.
Both original files are already epic preconditions. This checkpoint and the current operator overrides supersede stale model/CI/version wording in the original files, not their unresolved product requirements.

Read current repository instructions and the project-management skill. Use task-board public CLI and tracked background runs. The original A0 installed-binary/API/argv/Codex -p investigation is COMPLETE and reviewed; use its board evidence. Reverify only facts whose tool/tag/state changed or which a specific remaining test needs.

The final objective is unchanged: ship real untracked `curator run claude_code|codex_cli|pi --profile relux-root-context-ivan`, then migrate relux-agents-infra onto Curator. Real ax integration and its PR are out of scope; tracked handoff is implemented and verified ONLY against fake ax.

## Parking facts — do not mistake these for completion

All runs in the four goal control roots were checked: none remain running/queued. The only active run was source recovery producer **RUN-260908-24e513**, cancelled through the public directive `RUN-260908-24e513:cancel:8756b9` on explicit operator request. Its runner PID 76875 is gone. The provider had already finished implementation; the runtime was executing command **10 of 15** in configured validation. Cancellation is not a test failure verdict or a passing suite. **No CR for BUG-260909-24mm7l was published.** Task and Story are now `to-dev`.

Do not automatically resume cancelled runs during parking. On explicit resumption, recover the existing code and finish normal validation/publication; do not redevelop it or fabricate a successful validation record. Do not manually execute the entire suite and then duplicate it through runtime handoff.

Local parking archive: `/Users/iv/Developer/ReluxWorks/curator-agent-launcher/.temp/parking-2026-09-09/`.
- source-recovery.json: exact base, tree and SHA-256 for all 9 preserved files;
- source-recovery.patch and source-recovery-files.tar.gz: tracked AND untracked code;
- execution-cr.json / execution.patch / execution-candidate.tar: accepted execution CR2;
- defaults-cr.json / defaults.patch / defaults-candidate.tar: accepted defaults checkpoint;
- board-scopes.tar.gz: owned board cards/resources/activity at parking (a backup, not a private-history import route);
- control-records.tar.gz and four control-root run inventories: read-only recovery evidence, NEVER permission to edit private records.
The actual worktrees, branches, CR stores and evidence remain in place.

Long chronological notes: launcher `.temp/launcher-migration/orchestration-state.md`. Its early sections contain superseded instructions; prefer this checkpoint and the newest notes. Detailed briefs, spawn handles and logs live in `.temp/model-switch-astra-medium/`.

## Current policies — mandatory

- ALL producers and reviewers: **Codex gpt-6-astra, reasoning medium**. This supersedes Claude/Fable/low in the original goal. Use scoped context for bounded work; recent narrow runs use `--context-profile lite`. Preserve explicit role, preflight, background mode and model/effort.
- **No hosted CI until the operator changes this policy.** Run necessary local checks. Never call cancelled/missing hosted statuses green. Use `[skip ci]` on new commits where workflows would otherwise trigger. Review, signatures and exact-head delivery remain mandatory.
- Option A: curator-run is the single composer. Interactive plans have no permission bypass; yolo is unavailable for the real delivery in this scope. Defaults belong to the launcher. MCP goes only into managed homes with source-identity allowlisting.
- Launcher SPEC is now **0.3.0-draft**, based on recorded errata; environments.md stays **1.1**. Do not reopen settled decisions or invent unsupported channels. Keep reserved `path_prepend`/managed skill-command-root limits honest.
- Worktrees under `/Users/iv/Developer/ReluxWorks/.worktrees/`; managed Story worktrees are beneath their frozen control root's `.temp/STORY-ID/worktree`.
- No LOGBOOK or ordinary source writes into control roots. Evidence via task resources. Explicit pathspec staging, never `git add -A`; preserve unrelated dirty/untracked work.
- Delivery: signed branch -> PR -> actual comment-review verdict for the exact head -> required LOCAL checks -> fresh signature/author/local=remote=PR-head/ancestor checks -> plain `${SHA}:refs/heads/main` push. No force to main. Signed rebase only with range-diff proof; never silently replace human identity/signatures. Delete branches only after the PR reads MERGED.
- Pull `--ff-only` before board-state pushes. Close each product Story with a signed Curator board-state commit. Source task-board and agents-infra use their owning boards. Do not bypass a typed refusal or fabricate missing history.
- No agent-created tags/GitHub Releases. New repos private until the operator flips visibility. No real ax calls/PR edits, no hand edits in ~/.agents, ~/.claude or ~/.codex. Install from source through its normal flow; never swap binaries/restart daemons from a hosted child. Never `--stop-hosted` implicitly.
- Prefix gh calls with `GODEBUG=netdns=go`; use `task-board --no-update-check`. Execute spawn/status/observe in the CORRECT control root; wrong-root run-not-found is not evidence that a run disappeared. Redirect observe output to task .temp logs.
- Use same producer role/archetype in a NEW tracked owner run for checkpoint/complete. Accepted is not delivered; code delivered is not necessarily board done. Do not attempt another reviewer spawn against an already accepted task: it is integration-owned.

## Frozen roots and configurations

| Code control root | Config (all paths below launcher .temp) | Board owner |
|---|---|---|
| `.worktrees/launcher-control` | `launcher-migration/task-board.config.json` | curator/.task-board; explicit curator repository owner |
| `.worktrees/taskboard-cross-repo-control` | `cross-repo-delivery/task-board.config.json` | skill-project-management/.task-board; explicit original skill-project-management checkout owner |
| `.worktrees/agents-management-control` | `launcher-migration/agents-management.config.json` | curator/.task-board |
| `.worktrees/curator-spec-control` | `launcher-migration/curator-spec.config.json` | curator/.task-board |

All `.worktrees` paths above expand under `/Users/iv/Developer/ReluxWorks/`.
Source config retains all 15 validation commands and workload recommendations: fresh preflight with task_class=code, then exact snapshot digest/rationale on spawn. Launcher suite is make check. Spec suite requires its correctly quoted PATH/venv recipe; do not regress the quotation around PATH (it contains a path with spaces).

Human identities differ: launcher Ivan Oparin <ivan@relux.works>; source task-board/Curator/agents-management generally Ivan Oparin <oparin@me.com>. Inspect actual configured identity/key, verify good signatures; do not copy signing configuration between owners.

## Delivered code and completed Stories

Launcher main at parking: **adf627607eb334e9839288cfffce63e1268ae688**. The binary's main still stops after mapping with not_implemented: complete execution is NOT wired or installed.

| Delivered item | PR / exact landed head | Board status |
|---|---|---|
| CLI grammar | launcher PR4, 25379f22245e3bd8d2b81b192287d04368bf42c7 | STORY-260908-xadoax done |
| Fragment resolve, strict JSON, CCJ-1, diagnostics evidence | launcher PR5, 84e659e1bda41c0b70fad72e9e29b3c7ad474a7d | STORY-260908-15st55 done |
| Environment/system mapping | launcher PR6, 13b28c9a8916464e7253551808ae9969d6aa0186 | STORY-260908-2s7idv done |
| BuildLaunch/env/argv normative errata | launcher PR7, 18aeaed9af7dc5ffbe6cc79a4731a852fbb716da | STORY-260908-33cxp5 done |
| Pi prompt precedence erratum | launcher PR8, d97e6cb8bb872996b3432394477d09f16316e0e1 | STORY-260908-17dcju done |
| Composition API | launcher PR9, 84747c326eee9863ddfd7e86ac65be1056718fbc | Code landed; STORY-260908-v16gn5 integrating, recovery below |
| System-prompt API | launcher PR10, adf627607eb334e9839288cfffce63e1268ae688 | STORY-260908-v0w76i done; board eb366939a4a3706d4a5084962b2494f056bba801 |
| Protocol errata | curator-spec PR48, d019f0e7179520b5c8dcde321c4fe51e04552f58 | STORY-260908-6gwnfc done |
| Native Pi plugin | agents-management PR23, a2a6e9f377f62a5872d99ecdfff0d1690e385f2a | STORY-260908-3lmnfs done; operator tag pending |
| Separate-owner base capability | skill-project-management PR187, 416693dd3c0da087374d2e8be011703eb15d2904 | Code delivered; original historical cards lost, never fabricate them |
| Strict changed-tree proof/export + setup/exit fixes | skill-project-management PR194, 5ec20de4b913fcbf30d54a035731b9e0aba1134a | Code installed; BUG-260908-1awn50 integrating, recovery incomplete |

A0 and accepted designs are done: STORY-260908-3mcpz5 (TASK-qblycn), STORY-260908-37tde1 (TASK-1c0fwn), STORY-260908-3p1kfp (TASK-ggxfte). E6 was reconciled without new code: STORY-260908-2utz8k / TASK-260909-1zcwvs done, Curator board commit **bfe603367241881888ee869a156092533eedbafe**. Curator epic snapshot was published in PR67 at05514d1fa04aa4b3c17a004734d83cdb1cd30a99. At parking **12/26 product Stories are done**; this is not a product completion percentage.

## Resume priority 1: finish task-board recovery at source

Task-board is NOT fully finished. Installed version is **0.24.3-330-g5ec20de4** from reviewed source PR194. Canonical setup verification passed; 4 existing daemon PIDs were unchanged and no restart occurred. Ordinary separate-owner completion works and has closed multiple real Curator Stories.

Remaining actual defects:
1. prepare-landed-review succeeds for launcher composition, but the documented ordinary set_status(to-dev) is refused by the real consistency guard.
2. Same-canonical code/board checkouts cannot export a current tree carrying independently upstream board changes since the old comparison base.

Fix already implemented, NOT published/accepted/installed:
- **STORY-260909-3ue5iq / BUG-260909-24mm7l**;
- workspace `taskboard-cross-repo-control/.temp/STORY-260909-3ue5iq/worktree`;
- base5ec20de4b913fcbf30d54a035731b9e0aba1134a;
- preserved 9-file tree **a9089702fd4fce9c6528b8bf17cbd5efafe4ad3a**;
- cancelled RUN-260908-24e513; task/Story to-dev, no published CR;
- outcomes `BUG-260909-24mm7l_recovery-evidence.md` and `BUG-260909-24mm7l_verification-logs.zip` describe code, 16/16 AC rows, actual real-CLI/datastore recovery, seven narrowing mutants and exact exits.

The proposed command is `worktree start-landed-rework ... --export FILE`: re-observe authority, compare export, persist intent, then use the privileged status transition. Original accepted CR/ledger remain intact; normal producer publication and NEW reviewer acceptance become reachable. Shared-canonical board bytes require independently proven ownership/subtree equality and remain visible in review. Normal snapshots and ordinary status guards remain strict. Review this implementation; do not assume tests alone accepted it.

Continue normal producer finalization from preserved bytes, completing configured validation exactly once through the runtime. The interrupted 10/15 suite is NOT green. Then canonical Astra-medium review, signed source PR/exact-head landing, safe source install/verification, and REAL recovery of:
- **TASK-260908-1wmb40 / STORY-260908-v16gn5**: old CR1 accepted tree0361a3dfe25405da63ed9e70f886ba25880701db; delivered PR9 tree16088377353639e64afda5bc4275aae8fc726a90. Successful original export is attached as TASK-260908-1wmb40_prepare-landed-review.json. Re-observe current authority rather than blindly using a now-stale export.
- **BUG-260908-1awn50 / STORY-260908-2f6ulf**: old CR2 accepted tree89facf609cdaa47ef025f1c5fa031b46a76462a4; delivered source PR194 tree d60bad868c47df55813b2df5abea29b64574ef4b. Current protected source head may additionally contain the new fix and board commits; exact fresh export/new review required.
Use the corrected documented public sequence, a new exact-tree CR, independent acceptance and bound-owner Complete. Never transfer old acceptance to another tree, fake an empty delta or edit private CR/status records.

Original source board checkout is dirty and behind: last observed HEAD64bbf49034bdb5040b2bc21d5ca80b720738b318. Pull --ff-only refused on FOREIGN STORY-260903-i05b8l and TASK-260908-12i8g0 activity/progress/resources (source-board-owner-pull.log). Preserve them; do not reset/stash/drop foreign board history or touch ax work. Complete's documented fresh-authority/temp-index publication and local checkout convergence are separate facts; use public safeguards and record any refusal. Do not assert that this checkout has been synchronized.

Earlier original source IDs STORY-260908-3r282p/TASK-260908-2lpcml and STORY-260908-ps7x2v/BUG-260908-2ovkto lost their uncommitted cards/activity during an external checkout advance. Cause was not attributed, no authentic backup was found, and no import/history reconstruction succeeded. Their code/evidence survived. New tracking above is truthful continuation, not a claim those missing cards closed.

## Resume priority 2: deliver accepted execution and finish the launcher

**Execution is accepted, not delivered.**
- TASK-260908-3ued5d / STORY-260908-zvrz83;
- **CR2 accepted**, tree **ff61be4a8bd43fa4ffb179d31aa38e41891d4313**, original base84747c326eee9863ddfd7e86ac65be1056718fbc;
- producer fa23f2 and reviewer d0259e completed; runtime make check passed at23:07:47Z;
- actual diff against current mainadf6276 is 11 execution/config/docs paths; 3 upstream systemprompt code/test/harness blobs are byte-identical and already carried in the candidate.
- Delivery worktree JUST CREATED and still CLEAN at adf6276: `/Users/iv/Developer/ReluxWorks/.worktrees/STORY-260908-zvrz83-delivery`, local branch `feat/execution-handoff`. **No execution delivery commit, push or PR exists.** Reuse it; inspect fresh main before materializing the frozen accepted tree.

CR1's real PTY defect (one Ctrl-C -> two SIGINT, 5/5 each route) was reproduced and fixed. CR2 uses child process-group/foreground ownership, parent-only relay, stop/resume and terminal restoration. Independent reviewer verified real PTY reads, single interrupt, parent-only interrupt, stop/continue, 12 exit/cleanup cases and F1/F2/E6 mutants. 48 process matrix cases and closed ax.json behavior passed. No real ax/model calls. Retain stated Linux/runtime, bg/disown, termios and pathname-race bounds. Do not rerun old development or claim the earlier defective revision was shipped.

Land exact accepted execution tree through signed PR canon, then new bound developer Complete with CR2 and exact landed SHA. This tree already preserves upstream §5, so avoid needless changed-tree recovery if main is unchanged.

**Defaults partial checkpoint:** STORY-260908-1wxjbs.
- TASK-260908-25z3wj CR1 is checkpointed, not in main: signed **e6827e350485c55c743c4cfcfc80ad6aeb4ed831**, tree5b154a523b71ec5daabb97a55e81d9af1d3a5c86, parent13b28c9. Files/strict schema/lock/per-member flags/operator/machine resolution are implemented and accepted.
- Remaining TASK-260909-2vy977 owns real tagged runtime-compatible Lineup fallback, per-member origin/effort behavior, stderr line-group and production wiring. Preserve the checkpoint and use the supported workspace lifecycle; do not restart this Story.

Remaining launcher tasks:
- Plan/limits: STORY-260908-3d3vza / TASK-260908-2so46q. Real tagged vendorplugin.BuildLaunch(LaunchModeInteractive), managed Home, empty Composition, explicit resolved pair, runtime/vendor admission; separate providerlimits verdict keyed by Runtime/Model/managed Home. No bypass or silent fallback.
- Diagnostics: STORY-260908-18kdnq / TASK-260908-1wr53w.
- Full main wiring and installed integration: STORY-260908-1gywcb / TASK-260908-1o7i8y. Load ax config BEFORE argument validation, resolve fragment, mapping/defaults/origin group, admitted plan/limits, prompt selection and composition, all THREE actual late checks in BOTH modes, direct execution or fake-ax-tested document. Wire systemprompt.PrepareLaunch and warnings; a success callback or isolated API tests do not discharge main integration.
- Release readiness: STORY-260908-1nbb5h / TASK-260908-3bxxzi: README/help/CHANGELOG, local goldens/race/build, no tag or hosted CI.
- Install curator-run to ~/.local/bin only when real pipeline is ready. Verify umbrella discovery, --repair/current homes and real launches of all3 adapters on this machine; record transcripts, model/effort group, MCP channel delivery. Do not claim current main's not_implemented binary is shipped.

## Human-only / product inputs still pending

1. Native Pi upstream is landed, but requested signed annotated **v0.5.11** at **a2a6e9f377f62a5872d99ecdfff0d1690e385f2a** was absent on the remote at last verification. Operator-only tag instructions are in launcher `.temp/launcher-migration/native-pi-operator-tag.md`. Never create it yourself without a new explicit operator override. No pseudo-version, local replace or committed workspace override. Current launcher composition imports real v0.5.10 for VALUE contracts only; that tag does not admit pi-native. Verify API/runtime behavior at the actual new tag before claiming native Pi integration.
2. Original DoD asks MCP in all3 environments, but environments1.1 §7.8 and the current Curator registry declare **no Pi MCP channel**, and installed Pi0.84.2 has no native one. Epic resource pi-mcp-scope-decision.md records the pending decision: MCP for Claude/Codex plus native Pi without MCP, OR explicitly authorize separate real Pi MCP scope. Neither has been selected. Do not silently waive the DoD or invent an unconsumed channel/workaround.

## Resume priority 3: agents-infra migration (not implemented yet)

Keep the original order: ship the launcher, then migration. Reuse existing B Stories; inspect source at `/Users/iv/Developer/IV/relux-agents-infra` @dee5403 and current standalone successors where the original plan permits them.

- B1 STORY-260908-l5nerr / TASK-260908-3jux68: private relux-root-context repo, core/workflow/style/claude/attachments packages; claude modules environments:[claude_code], module bytes and weights per0012; umbrella relux-root-context-ivan; strict operator-created v-tags.
- B2 STORY-260908-sd6xkr / TASK-260908-2kihaw: pdf, skill-creator/standalone successor and agents-attachments CLI skill manifests in their owning repos; umbrella requires.skills with ranges.
- B3 STORY-260908-2a4936 / TASK-260908-1bpra2: private relux-mcp, figma HTTPS, lldb bare lldb-mcp, safari bare safaridriver-mcp wrapper; source-identity allowlist and token env_names; umbrella requires.mcp.
- B4 STORY-260908-k88yk0 / TASK-260908-s1fdvr: launcher defaults.json preserving appropriate operator model/effort preferences. Worker Astra-medium policy is not automatically a new primary-session preference.
- B5 STORY-260908-2u6nly / TASK-260908-yl5x3k: agents-infra doctor baseline, recorded backup, `curator profile install <umbrella> --use --takeover`, env status current for3 adapters. Xcode only with explicit operator consent.
- B6 STORY-260908-3ry7gf / TASK-260908-2kqocx: through agents-infra's own board/PR canon, remove migrated instructions/skills/MCP composition and retired code; keep claude settings.json, Codex config merge/rules, Pi local-model runtime, lldb wrapper and required residual attachment contract; rewrite README. Retire-vs-deprecate launcher choice remains an operator product decision when concrete.
- B7 STORY-260908-2haegq / TASK-260908-1e55lp: file3 curator-spec issues/proposals for tool-config surfaces, per-project policy and skill command roots. No agents-infra workarounds.

## Completion standard

The original goal-file DoD remains binding except explicit Astra-medium and no-hosted-CI overrides, plus any FUTURE recorded Pi MCP decision. Completion means actual managed-home launches via curator run for3 environments with the selected profile and model/effort group; required MCP behavior proven under that decision; local goldens/build/race green; documented installation; umbrella installed/current; signed reviewed package/infra deliveries with operator-created tags; residual-only infra README/code;3 spec-gap proposals filed; ax untouched. Board evidence and signed Story completion must agree with reality. Do not mark the goal complete from percentages, accepted-only CRs, partial suites or API-only tests.

Parking receipt: primary board goal was updated by CAS to revision **11**, explicitly requiring a new operator resume before further workers/tests/publication. The parking goal/resource is a local handoff; no new feature commit, PR, tag or release was created during parking.
