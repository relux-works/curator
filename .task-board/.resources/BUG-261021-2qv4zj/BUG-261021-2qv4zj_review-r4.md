# BUG-261021-2qv4zj review r4

Head: `a2244ec13e223e3d6f5b96cc11ba22829bc6a2a2`  
Original base: `2c89fd46443d461e25bd377fb94a0c6754282838`  
Round base: `d20c4363cf2dea88ce7418fa594664a466c116c8`

Verdict: **changes_requested** — P1 F4, repair cannot rebuild an existing projection with missing, extra or changed files. F1, F2 and F3 remain closed in the exercised cases.

The orchestrator request, board assignment and all nine precondition resources were checked. The exact head was fetched and checked out detached; all nine introduced commits verify `G`. Tests ran in an ordinary `.git` validation clone with private HOME, GOCACHE, GOPATH/module cache and short TMPDIR, always unsetting GIT_DIR/GIT_WORK_TREE. Repository source was unchanged; independent probes and bounded controls used Go overlays.

| Surface | Held or broken | Production exercise | Evidence |
| --- | --- | --- | --- |
| F3 native/managed publication | Held | Both original r3 probes build the exact-source CLI, install three excluded-contact fixtures and a declared visible control, then read Claude/Codex native and managed skill surfaces. | `r3-probes.log`, exit 0; excluded files absent, visible bytes readable. |
| Copy fallback | Held | Independent CLI/Resolve probes with only projected-skill symlink creation forced unavailable; assert the resulting skill entries are real directories before checking files. | `independent-copy.log`, exit 0. The overlay propagates into the child CLI build; no claim of an actual Windows run. |
| Complete core selection and repository-shaped package | Held | Install exercises repository research/board/vendor files, all 19 exclusions, visible assets, command-less scripts, exported scripts and runtime/build roots; links and both copy helpers compared with canonical context. | Focused projection publication/core-selection tests, plus independent r3/copy/runtime probes. |
| Projection identity and currentness | **Broken, F4** | Production Resolve on both adapters, starting from a current home, after removing a file, adding a file or changing SKILL.md. Nearby intact and whole-directory-missing controls. | `independent-store.log`: six repair failures; four positive controls pass. Status refuses stale state, but explicit repair never converges for existing altered directories. |
| Shared projection lifecycle | Held | Two real profile installs share the same member projection for both adapters; remove the first profile with purge and run production maintenance; resolve/read the second again. | `independent-lifecycle.log`, exit 0. Projection survives and the remaining profile stays current. |
| Owned raw-root migration | Held | Production native and managed repair from old markers, raw links and recorded copies; excluded old files removed, locks unchanged, no takeover. | Focused TestProfileSkillProjectionRepair; managed complete-copy transaction rollback/retry test. |
| Legacy identity migration | Held | Independent real v1 provisioning, replace skill surfaces with raw links/copies, then production Resolve migrates v1 to v2 for both adapters. Check excluded-file removal, visible bytes, copy mode and current resolve. | `independent-migration.log`, exit 0. Coordinated publisher stages the same projection before activation. |
| Foreign ownership | Held | Place an unrecorded native skill link to a fixture outside the store and call native materialization without takeover. | `independent-lifecycle.log`: intended foreign-manager refusal; original link unchanged. |
| Runtime and other publishers | Held | Independent project install exports a script that reads a declared runtime asset; assert runtime/research paths absent from context, mutate live source, launch installed command and require frozen runtime bytes. Trace project/global staging and profile-global publication. | `independent-runtime.log`, exit 0. Commands retain independent runtime/cache sources; profile native, managed and identity-migration skill links use projected roots. |
| F1/F2, cache and shared readers | Held | Real gates, exact-source context CLI installs, alias/Unicode/path neighbors, context module reader, old allowance records and kind separation; both prior mutation suites rerun. | `focused.log`, `independent-neighbors-corrected.log`, pinned conformance. Ruleset 8 unchanged; publication version is separate from admission cache identity. |
| Narrowing controls | Held | All thirteen new projection controls, including native/managed raw publication, missing store checks, old copy migration and changed-byte reuse. | Focused TestSkillProjectionNarrowingMutants; each requires exit 1 and its named behavioral assertion. |
| Documentation and records | Broken at repair acceptance boundary | Re-derived publisher inventory, revision identity, runtime separation and fixture corrections. Producer explicitly documents refusal of corrupt stores; review brief explicitly requires rebuilding missing/extra/changed-file projections. | The producer's narrower refusal claim is accurate but does not satisfy the r4 repair requirement. |
| Public repository hygiene | Held | Full production diff and fixture changes reviewed; added-line search for personal paths, host/user identifiers and organization-specific disclosures. | `hygiene.json`, clean diff whitespace, signed-object evidence. No finding. |

## F4: existing altered projections cannot be repaired

`internal/envprofile/skill_projection.go:75-76` immediately returns `verifySkillProjection` whenever the target directory exists. A digest mismatch therefore never reaches staging/republication. Managed repair calls this from `applyPlan`; native materialization uses the same publisher.

The independent probe provisions a valid managed home with Resolve(Repair:true), locates its projected `skills/myskill` target, and separately removes SKILL.md, adds extra.md, or changes SKILL.md. Read-only Resolve correctly reports environment_home_stale. Resolve(Repair:true) then returns `environment_repair_failed: skill context projection differs from its identity` in all six cases across Claude and Codex. Repeating repair cannot change that branch. An intact store resolves normally, and removing the entire store directory allows repair to recreate it; those controls distinguish the existing-directory defect from a broken fixture or inability to rebuild.

This blocks the explicit r4 projection-repair acceptance criterion and leaves every managed consumer of the shared damaged projection unavailable until manual store removal. The finding is about recovery, not acceptance of changed bytes: the current verifier correctly refuses them.

Required fix: during explicit repair, safely stage and verify the projection from the pinned source and recoverably replace a stale regular projection directory under the mutation lock. Keep read-only status non-mutating, reject linked/foreign routes without following or modifying their targets, preserve shared consumers, and retain interrupted-replacement/retry coverage. Do not simply weaken verification or accept the altered directory.

Reproduce with the companion `BUG-261021-2qv4zj_projection-repair-r4_test.go`, mapped through a Go overlay to `internal/envprofile/reviewer_store_test.go`, then run `env -u GIT_DIR -u GIT_WORK_TREE go test -overlay "$OVERLAY" -count=1 -v ./internal/envprofile -run '^TestReviewerProjectionRepair$'`. Expected at this head: exit 1, six named repair-convergence failures and four positive controls.

## Validation evidence and limits

The full suite uses `go test -json -p=2 -count=1 -timeout=90m ./...` without overlays. Pinned conformance uses spec `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`, matching the committed workflow pin. Platform: Go 1.25.5, Darwin arm64. Case-sensitive duplicate-root behavior and actual Windows symlink failure were not exercised; forced fallback is explicitly an instrumented control.

The first neighboring-path overlay mistakenly mapped a package-audit test into contextpkg and failed package setup. That run earns no behavioral evidence; the corrected overlay passed. The independent store probe's exit 1 is the intended F4 witness. Conditional/platform/corpus skips are reported separately from passes. Earlier producer development runs were not treated as reviewer validation of this head.

The first interop run exposed a reviewer fixture-materialization error: `git archive` expanded the pinned fixture's `export-subst` file. Restoring that file directly from its pinned Git blob (40 bytes, matching the vector digest) corrects the input without changing product source or expected values. The initial interop failure receives no passing credit; the exact-byte rerun is recorded separately. Other conformance runs do not consume this fixture.

Default suite: **85 packages** (80 passed, 5 without test files); 11759 test/subtest passes, 199 skips, 0 failures. The independent inventory reconciles all **3594 top-level tests/fuzz targets exactly once**. Full suite exit: 0.

| Reviewer run | Exit | Seconds |
| --- | ---: | ---: |
| `build` | 0 | 1.49 |
| `conformance-core` | 0 | 42.22 |
| `conformance-env-rest` | 0 | 42.50 |
| `conformance-env` | 0 | 261.07 |
| `conformance-interop-exact-bytes` | 0 | 2.35 |
| `conformance-interop` | 1 | 1.40 |
| `focused` | 0 | 102.62 |
| `full-suite` | 0 | 1475.86 |
| `independent-copy` | 0 | 12.67 |
| `independent-lifecycle` | 0 | 4.86 |
| `independent-migration` | 0 | 13.52 |
| `independent-neighbors-corrected` | 0 | 1.62 |
| `independent-neighbors` | 1 | 0.05 |
| `independent-runtime` | 0 | 4.37 |
| `independent-store` | 1 | 12.43 |
| `inventory` | 0 | 0.48 |
| `r3-probes` | 0 | 14.31 |
| `test-inventory` | 0 | 141.41 |
| `tools` | 0 | 0.01 |
| `vet` | 0 | 3.51 |

Exact commands, logs, overlay sources, inventory reconciliation and input-correction evidence are retained under `.out/BUG-261021-2qv4zj-review-r4-evidence/`. The report and F4 companion probe are prepared for board attachment; resource presence is verified separately before handover.

```json
{"verdict":"changes_requested","head":"a2244ec13e223e3d6f5b96cc11ba22829bc6a2a2","findings":[{"id":"F4","severity":"P1","file_line":"internal/envprofile/skill_projection.go:75-76","mechanism":"Any existing projection directory returns verification immediately. Missing, extra or changed files cause permanent repair refusal instead of rebuilding from the pinned source; all managed consumers of that shared projection remain stale.","reproduction":"Run the companion TestReviewerProjectionRepair overlay. For Claude and Codex, provision with Resolve(Repair:true), remove SKILL.md, add extra.md or change SKILL.md in the linked projection. Read-only Resolve reports stale; explicit repair returns environment_repair_failed: skill context projection differs from its identity. Intact and whole-directory-missing controls pass.","required_fix":"Safely stage, verify and recoverably replace a stale regular projection directory from the pinned source during explicit repair. Keep status read-only and linked/foreign routes refused, preserve shared consumers, and test missing/extra/changed files plus interrupted replacement and retry."}]}
```
