# TASK-261013-vtmjnu — apply-wave5-transfer-untrusted-cips

Applied the current gw-CS6-brief.md against the wave-5 owner decisions of 2026-10-13. Scope is documentation and figure sources only. Changes remain uncommitted in the assigned Story worktree; integration belongs to the orchestrator. No wire fields, refusal codes, cap values, or qualified capabilities were invented.

## Current brief coverage

1. R-TRANSFER: CIP-0009 D5/D5a replaces reviewed-patch export and unresolved transfer ownership with the complete materialize/work/round-end/export/candidate/review/rework/integration round. RWP v2.1 §8/§10 gates adopted, including whole-export refusal, import refusals and explicit flagged-hunk review. Fresh rework uses the last candidate and reviewer findings via signed task letters; intermediate export uses the same turn-boundary fence/absence rules. Design labels registered as unpublished.
2. R-NET-TOOL: CIP-0009 D0/D4/D5b and CIP-0008 R-NET-TOOL state zero-egress workplaces with no proxy and denied loopback. Default donor loop has only provider endpoints and our host; no search/fetch or donor web server. The plan alone disables provider-side search inside provider API traffic. Rationale, signed task resources/read-only materialization/digests, worker requests for more resources, and both pending trusted-circle web options are explicit. swarma-web has its own B identity/C sandbox, never harness/bridge identity; search-provider API egress only, donor key custody/consent, no key transfer/project retrieval, results to donor model only. Other lockdown controls remain.
3. R-EXEC/R-VERIFY: CIP-0009 D3/D5c binds workplace exec to OS executable profiles, excluded categories and qualified light-check wrappers; dependency caches are read-only. Heavy work and all worker-code execution use secret-free isolated ephemeral runners on another machine, controlled dependency egress and board data in testing state. Agent tester orchestrates/reads only.
4. R-STANDING/R-UNTRUSTED: CIP-0009 D4/D5d records class at admission, default untrusted workplace, operator trusted-circle admission for donor opt-ins only, no host boundary relaxation, donor-side standing loop only, host matrix/class gates and no subscription bind/lease. CIP-0010 narrows C5/C7's broker boundary accordingly.
5. R-WATCH/R-ALLSTOP: CIP-0009 D7 applies one-account watcher freeze/journaled stop and manual node-wide hold/range deny/all-generation absence, broker freeze and session-host delivery closure/snapshots. Delivered tokens keep vendor lifetime. Each service journals; partial/unknown keeps admission closed; explicit recovery. Existing workplace suspended and donor channel-down paths are used, no new state.

## Changed sections and figures per shared name

Figure paths below are relative to cips/diagrams/. Protocol section titles identify the shared contract without inventing section numbers.

| Shared name | Changed sections | Changed figures |
|---|---|---|
| R-TRANSFER | CIP-0009 Revision 7, Current state, D5 backend/materialization/delivery, D5a, D7, Test plan, open question 14; remote-worker protocol Revision 4, Publication status, Tools/mailbox/lifecycle, workplace publication entry, R-TRANSFER section | remote-worker-transfer.puml (new), remote-worker-turn.puml, remote-worker-retire.puml, donor-lifecycle.puml, donor-layout.puml, donor-side.puml, donor-side.dsl, donor-variants.puml, credential-source-lifecycle.puml |
| R-UNTRUSTED | CIP-0009 D4 steps 5/7, D5d, Test plan; CIP-0010 Revision 3, broker boundary before C5, question 6; protocol Deployment/launch, Publication status, admission section | donor-first-deployment.puml, join-handshake.puml, donor-side.puml/.dsl, donor-variants.puml, credential-source-lifecycle.puml |
| R-NET-TOOL | CIP-0008 Revision 3, R3 hosted-tool row, R-NET-TOOL before R4, question 6; CIP-0009 D0 egress, D1 row, D4, D5 backend/worker view, D5b, Security, Test plan, questions 14/15/16; CIP-0010 C5/question 6; protocol Publication/Deployment, workplace publication entry, network section | donor-side.puml/.dsl, donor-variants.puml, donor-layout.puml, donor-first-deployment.puml, join-handshake.puml, remote-worker-turn.puml, remote-worker-transfer.puml, donor-lifecycle.puml, credential-source-lifecycle.puml |
| R-EXEC | CIP-0009 D3, D4 step 7, D5 backend, D5c/D5d, Test plan, question 16; protocol Publication/Deployment, exec section | donor-side.puml/.dsl, donor-variants.puml, donor-layout.puml, remote-worker-turn.puml, remote-worker-transfer.puml, donor-lifecycle.puml, donor-first-deployment.puml, join-handshake.puml |
| R-VERIFY | CIP-0009 D5a review, D5c, Test plan, question 16; protocol Publication status, transfer/network sections | remote-worker-transfer.puml, remote-worker-turn.puml, donor-side.puml/.dsl, donor-variants.puml, donor-layout.puml |
| R-CEILINGS | CIP-0009 D4 step 7, D5d, question 17, Test plan; protocol Publication/Deployment/admission | donor-first-deployment.puml, join-handshake.puml, donor-side.puml/.dsl, donor-variants.puml, donor-lifecycle.puml |
| R-WATCH | CIP-0009 D7, question 17, Test plan; protocol Publication/admission | remote-worker-retire.puml, donor-lifecycle.puml, remote-worker-turn.puml, donor-side.puml/.dsl, donor-layout.puml |
| R-ALLSTOP | CIP-0009 D7, question 17, Test plan; protocol Publication/admission | remote-worker-retire.puml, donor-lifecycle.puml, remote-worker-turn.puml, donor-side.puml/.dsl, donor-layout.puml |
| R-RANGE-DENY | CIP-0009 D5d, question 17, Test plan; protocol Publication/admission | donor-first-deployment.puml, join-handshake.puml, donor-side.puml/.dsl, remote-worker-retire.puml, donor-lifecycle.puml |
| R-STANDING | CIP-0009 D5d, question 17, Test plan; protocol admission | donor-side.puml/.dsl, donor-variants.puml, donor-lifecycle.puml |
| R-SOCKETS | CIP-0009 D5d, question 17, Test plan; protocol Deployment/Publication/admission | donor-first-deployment.puml, join-handshake.puml, donor-side.puml/.dsl |
| R-SHARED | CIP-0009 D5d/D9, question 17, Test plan; protocol Deployment/admission | donor-side.puml/.dsl (separate public entry host), donor-first-deployment.puml, join-handshake.puml, donor-variants.puml, donor-layout.puml |

All changed specs have top revision notes in their existing style. Changed figure sources have dated revision notes. cips/diagrams/README.md records the wave and the new transfer sequence.

## Earlier controls already satisfied and retained

- R-WORKPLACE service-principal executor and relay-only fixed SSH shell: CIP-0009 D4 step 7 and D5 connection/control/jobs. No workplace-owned control state or direct UID-switch path added.
- F-16 standalone private Git across accounts and same-account worktrees: CIP-0009 D5 materialization; strengthened by D5a. No linked worktree across accounts.
- R-OBS/F-7 authenticated durable claims/floors and unavailable-versus-absence distinction: CIP-0009 D5 remote effect claims/freshness/observation and protocol Remote effect claims/recovery. Retained without new states/codes.
- R-OPERATOR exact-act phone signing, terminal fallback and mandatory observer: existing operator notes in donor/retirement/join/credential figures retained; wave-5 all-stop's authentication/one-confirmation and agent-session refusal stated explicitly.
- CIP-0008 R3 default provider-search disablement (including Codex root web_search disabled) and R4 caller-tail refusal already existed. Retained; the new R-NET-TOOL text explains why network filtering cannot control provider search and requires a new tuple for opt-in.

## What remains open

- swarma-rb/1 and swarma-workplace-exec/1 wire ownership, closed transport/backend schemas, execution shape/concurrency/blast radius, peer authentication and private-home safe file access; existing join/renewal and donor provisioner schemas/qualification, with B still disabled.
- swarma-workspace-materialize, swarma-workspace-export and swarma-dispatcher workspace-result/1 are unpublished design labels. Would need archive/manifest encoding, exact plan/write-scope/resource fields, cap values, result schema/retention, cleanup-obligation/waiver records and immutable-object provenance/access/retention if used.
- Board deterministic-import command, candidate trailer and flagged-hunk verdict formats would need publication. Ownership/round path is decided; no longer left as an ownership question. Dispatcher question 22 and user-manager question 15 closed except expressly unpublished formats.
- Task-resource attachment formats/caps/digest binding; per-identity trusted-circle opt-in and donor consent; swarma-web format/identity/profile/key custody and qualified provider-search or MCP tuples. Both web options remain pending qualification; no current tuple opened.
- Per-task exec profiles and protected light-check wrappers; R-VERIFY runner pool/job/results contract and controlled dependency-egress qualification remain with board/infrastructure owners.
- Per-release capability matrices/class minimums; open owner decision on acceptance of soft rows on hosts lacking hard ceilings. Linux mechanisms/baseline and macOS gaps are stated, with no OS-name exclusion.
- Range-rule precedence, socket reachability, shared-service deployment and standing-surface qualification; watcher thresholds/sampler/alert and all-stop act/record/recovery formats. No runtime qualification asserted.

## Evidence and task-specific exceptions

- Documentation only; no tests, builds, candidate execution or renderer runs by this author, as required by gw-CS6-brief.md. Future negative qualification vectors were added to the spec Test plan; they are not claimed as executed tests.
- Standalone git diff --check ran after edits: exit 0. This is whitespace validation, not diagram rendering or runtime qualification. No already-attached test evidence was accepted as new validation.
- Manual text/figure sweep plus a scoped public-hygiene search found no first names, private host names, local user absolute paths, session links or board-resource links in the changed public surfaces. The rg no-match read exited 1 (normal no-match result), not a passing test gate. Product installation paths and existing localhost/provider examples are specification content, not private machine paths.
- The initial requested development transition exited 1 because an estimate was required. set_estimate(fibonacci=5) and the repeated development transition each exited 0. A Story base-preflight reminder was emitted; this worker used the explicitly assigned worktree without committing/rebasing or changing its base. Integration/base authority remains with the orchestrator.
- Generic code/test DoD is satisfied here by the expressly scoped spec/figure implementation and future qualification plan. Runtime code/tests are outside this docs-only assignment.
- Generic logbook DoD is inapplicable under the current brief's explicit “Never write LOGBOOK.md” rule. No logbook file was written. Findings/decisions and open publication work are persisted in this outcome and board notes instead.
- Configured CR validation is delegated only to the mandated task-board handoff. Its real result/evidence is owned by the board handoff record; it has not been pre-claimed green here.
