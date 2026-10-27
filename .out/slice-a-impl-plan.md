# Slice A implementation plan

Task: TASK-261026-6b5ioc — role-profiles-and-skill-cli-dependencies.
Contract: curator-spec PR189, accepted carrier head 65574fc58b8a2d2aca3ac0335eac9feb7c967696; Curator design docs/design/profile-skill-runtime.md.

## PR1 — shared plan and lock preparation

Branch: feat/slice-a-plan-lock. Add explicit profile skill-plan preparation through the ordinary shared closure, lock-v5 strict parsing/canonical persistence, pinned transitive graph, activation union, overlapping direct provenance and selective reconciliation. Reuse source admission, strict audit, trusted build/receipt preparation before lock publication. Preserve unrelated declarations, lock members and installed artifacts; report their drift without converging them. No command publication or activation. Existing legacy context operations retain their behavior; an explicit preparation seam is handed to the CLI integration writer.

Coverage: PSP01 closed v5/legacy compatibility; PSP02 graph and context-origin agreement; PSP03 modes/ref/selection; PSP04 activation union; PSP05 plan and lock digests; PSP06 pinned validation and explicit upgrade; PSR01 selective preparation and retained members; PSR02 drift; PSR03 overlap refusal; PSR05 preparation-before-lock subset. Each implemented claim has a positive control and a paired negative with preserved preimages. Exercise applicable published schema/graph/digest vectors against Go code; oracle checks alone are supplemental. Real private builds exercise compile failure before publication. Full transaction/recovery, routing and runtime binding remain PR2 obligations.

## PR2 — isolated runtime and one journal

Depends on PR1. Add single-current-profile command generation/shims under supervisor-proven exclusive domain admission. Publish runtime, lock, owned shims and current through one journal, including rollback/recovery and adapter exceptions. No adoption claim while companions remain unqualified.

Coverage: remaining PSP04 surfaces, PSP05 runtime/receipt binding, PSP06 use/sync/repair pin reuse, PSR01 stale-owned cleanup, PSR04 bounded command preflight, PSR05 complete transaction and crash recovery, PSR06 exclusive routing. Pair install/update/use/sync/repair/recovery success with command conflicts, unmanaged paths, missing/unknown/other-profile admission, stale receipts, every journal interruption phase and unchanged preimage inventories.

## PR3 — companion and real-install CI gates

Depends on PR2 and accepted companion allocations TASK-261027-z89050 / TASK-261027-kb53ku for their implementations. Wire authoritative qualification checks; missing, accepted-only or stale evidence refuses completion/adoption with profile_skill_companion_required. Add the accepted design's real-install MUST coverage gate and bounded controls; route shared CI changes to its single integration writer.

Coverage: PSR04 protected-command companion and PSR06 managed-session companion; final-child PATH and reached artifact evidence for native and hosted launches; complete design MUST inventory. Positive qualified fixtures pair with missing/unqualified evidence, protected registry/platform names and aliases, omitted/inert/wrong-profile transport, and gate plants proving missing MUST evidence fails CI. Pure model checks never count as real-install qualification.

## Delivery and integration boundaries

Each PR is independently mergeable and reviewed on its exact signed head. PR1 is preparation only and cannot complete/adopt Slice A. Shared README, CHANGELOG, CI dispatch and cmd/curator/main.go remain integration-owned. credential_diagnostics*.go is excluded. Owner command-safety and final agent PATH outcomes stay open until both companions are accepted and qualified.

## PR1 implemented MUST coverage

| Obligation and boundary | Positive control | Paired refusal or invariant |
| --- | --- | --- |
| PSP01 closed portable carrier and historical readers | `TestProfileSkillPlanSchemaCorpus` (42 published shapes) | Unknown fields, missing fields, invalid shapes and successor fields on legacy schemas |
| PSP02 selected graph, package pins, retained members | `TestProfileClosureActivationAndRefusals`, `TestSkillPlanPreparationPublication` | Cycle, wrong revision, unknown/duplicate context edge, state/store/source/context mutation; unrelated acquisition is forbidden |
| PSP03 exact refs, modes and selections | `TestProfileSkillPlanPortableVectors` (24 published cases), `TestProfileSkillPlanExactTagGrammar` | Missing/system command, empty/unsorted selection, invalid mode/ref, bounded tag grammar violations |
| PSP04 incoming activation union (planning only) | `TestProfileClosureIncomingUnion`, `TestProfileClosureOverlapAndReuse` | Context-only dependencies remain context-only; context plus narrowed runtime does not widen commands; system commands never export |
| PSP05 portable plan and observed lock binding | `TestProfileSkillPlanLiteralBindings`, published literal CCJ digest | Forged plan digest, stale supplied binding, mutated selection; private artifact replacement fails receipt binding |
| PSP06 explicit preparation/reuse foundation | `TestProfileClosureOverlapAndReuse` | Legacy runtime requests require explicit upgrade; changed original declaration provenance refuses recomputation reuse; actual lifecycle reuse remains PR2 |
| PSR01 selected preparation and immutable-store persistence | `TestSkillPlanPreparationPublication` | Unreachable unrelated Git declaration receives zero acquisition/build calls; unrelated lock fields, declaration bytes, installed bytes and current remain intact |
| PSR02 read-only drift inventory | `TestSkillPlanUnrelatedCurrentness` | Changed ref/content reports drift; absent installation is drift with null observation; corrupt/non-directory observation is unknown; no repair |
| PSR03 exact overlapping machine constraint and full union | `TestProfileClosureOverlapAndReuse` | Source, commit and directory conflicts refuse before preparation; different exact refs resolving to one tuple are compatible |
| PSR05 admission/audit/compile-before-lock preparation subset | `TestProfileSkillPlanRealCompileBeforeLock` (real signed Git + ordinary Go worker and receipts) | Wrong signer, undeclared network contact, compiler syntax error, artifact replacement, stale observations; lock/current/declarations remain unchanged |
| Temporary companion barrier | `TestSkillPlanPreparationPublication` (readable v5), legacy context tests | `TestSkillPlanLegacyWritersRefusePreparedLock` refuses activation, materialization and old lock writers with `profile_skill_companion_required` |

PR1 exports `install.PrepareProfileSkillPlan(config, profile, observedLockHash, policy, options)` for an existing installed profile. It returns a prepared lock hash and a separate unrelated drift inventory. It has no CLI dispatch, runtime artifact publication, prompt-surface publication or qualification claim. The ordinary Go worker runs in the test manager image using the same fixed hidden worker dispatch; this is a real private compiler witness, not a final agent/PATH witness. Final install/update composition must retain PR2's single-journal boundary; sequentially calling the historical installer and this API is not the completed Slice A install contract.

Fixtures are copied unchanged from carrier head 65574fc5; portable-reader subset selection is explicit. Package/origin/provenance checks use admitted snapshots in the production planner. The complete published semantic corpus also contains publication, supervisor, recovery, companion and final PATH model cases outside PR1's claim; those remain PR2/PR3 obligations.
