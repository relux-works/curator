# BUG-261017-8v470p — audit-cache-ignores-capabilities

Status: ready for review; full remote gate green (exit 0, 20 successful jobs and 2 configured skips). This is a developer Change Request handoff record, not an acceptance verdict.

## Candidate and spec mapping

- Base/checkpoint: `fcd785ba93c5d00e2e80615f8add2baa93ccd73d`, equal to freshly fetched `origin/main` before the gate; zero commits behind. The migration fix `51e8fcaf` is already present.
- Gate snapshot: `1f43a073eb958dd0f480ee96af2f871a76a44825`. Its signature verifies as `G`. This is a detached gate snapshot; the managed Story branch HEAD remains the checkpoint and its candidate is uncommitted.
- Normative source: curator-spec main at `1059b94b4135089d842bf017070b8ab47d79a536`.
- [Core section 4.3](https://github.com/relux-works/curator-spec/blob/1059b94b4135089d842bf017070b8ab47d79a536/protocol/core.md#43-capabilities): capability-aware schemas require capabilities; optional prompt_scope is a non-empty purpose statement. [Common schema capabilities definition](https://github.com/relux-works/curator-spec/blob/1059b94b4135089d842bf017070b8ab47d79a536/schemas/v1/common.schema.json) requires an object and a non-empty string for a present prompt_scope; null is neither. Schema-3's required capabilities references that definition. Load applies this object requirement consistently through supported schemas 3–9 and both manifest filenames.
- [Skillfile sources section 4](https://github.com/relux-works/curator-spec/blob/1059b94b4135089d842bf017070b8ab47d79a536/protocol/skillfile-sources.md): complete audit package equality remains independently enforced.
- [Environments sections 9.1 and 9.4](https://github.com/relux-works/curator-spec/blob/1059b94b4135089d842bf017070b8ab47d79a536/protocol/environments.md): skill audits consume their own manifest capabilities; a covering declaration satisfies that network check while other findings still apply. Non-skill subjects retain implicit-none. Caching findings from different effective inputs would violate those outcomes.
- Citation correction: the brief calls core section 4.4 an audit-cache equality section. At the pinned main it is Dependencies; capabilities are section 4.3. This is a stale section reference, not a conflicting product requirement.

## Use of the supplied patch

Verified the supplied patch SHA-256: `ad460fa257db376b87568cc147888c251b8892b80d1efef2680f0b1f0f24b035`.

Adopted its production correction: persist and compare the full effective `capabilities.Manifest`, increment the local ruleset namespace from 2 to 3, reject present null prompt_scope and null schema-3+ capability objects. No supplied production behavior was dropped.

Adapted tests to the current post-migration base. Expanded schemas through 9. Replaced fixture network addresses with reserved example domains. The supplied profile crossover controls changed package identity between skill and non-skill, which could mask a missing capability comparison. The revised controls hold every other cache dimension equal and poison each member's cache with the opposite capabilities through production Gate. Both orders run for skill, context and MCP members. Added v1/v2 network+exec crossover and immediate-repeat cache-hit controls through Gate.

The preceding migration mutant's unstated-package attack becomes redundant when differing capabilities independently cause a miss. Kept every prior test name, strengthened the same migration regression with a fresh-verdict package-binding assertion, and retargeted that existing mutant's expected named failure to this independent package guarantee. No prior gate was removed.

## Implementation

`internal/audit/audit.go`: `loadCachedFindings` compares every manifest field with `reflect.DeepEqual` after all existing identity checks; `storeCachedFindings` records the effective declaration. Ruleset 3 uses a separate filename. Ruleset-2 records are neither overwritten nor reinterpreted. No portable record schema or content-hash framing changes.

`internal/capabilities/capabilities.go`: a present prompt_scope must be a non-empty string, including when its raw value is null.

`internal/skillspec/parse.go`: required schema-3+ capabilities cannot be null; the existing Parse(nil) compatibility behavior for pre-capability callers remains.

## Coverage map and acceptance ratio

Brief gap: no attack-surface table was supplied. The following explicit AC coverage map records the requested scope; it does not claim coverage of the entire attack-surface catalog.

| AC row | Named tests / procedural gate | Production entry and bound |
| --- | --- | --- |
| 1. Equal effective capabilities only | TestVerdictCacheBindsEffectiveCapabilities; TestGateEffectiveCapabilityCacheCrossover | auditSubject -> auditSubjectWithOpaquePaths -> loadCachedFindings/storeCachedFindings. All seven independently mutable manifest components covered; unchanged inputs hit. |
| 2. Both crossover orders; skill/context/MCP; identical hit | TestGateEffectiveCapabilityCacheCrossover; TestStrictAuditMemberCapabilityCacheIsolation | Gate and strictAuditMember -> Gate. Both declared-first and implicit-none-first, same bytes/commit/directory/package; direct Gate tests explicitly assert CacheHit for v1 and v2. |
| 3. Previous records retain meaning | TestCapabilityCacheRulesetRetainsOldRecords | auditSubject on a seeded ruleset-2 allowance computes the current refusal, subsequently hits it, and leaves old bytes untouched. Old-namespace fixture is v1; namespace isolation is shared by v1/v2. |
| 4. Spec-backed null rejections and controls | TestCapabilitiesNullRejectedFromV3; TestMigratedGlobalSkillPromptScopeValidation; TestMigratedGlobalSkillNullCapabilityDeclarations; TestParseNilIsImplicitNone; TestParseDefaults; TestParseFull; TestParseRejections | skillspec.Load -> loadSkillManifest -> capabilities.Parse; EnsureDefaultWithPolicy -> migrateGlobalSkills -> strictAuditMember. Absent, valid and empty-object controls; null object/field failures; rejected migration leaves profile files unpublished. |
| 5. Narrowing mutants and full remote gate | TestCapabilityCacheNarrowingMutants; TestGlobalSkillCapabilitiesNarrowingMutants; sh scripts/remote-gate.sh | Overlay harness executes the named behavioral production-entry suites. Remote gate result recorded below. |

Measured ratio: **5 of 5 AC rows driven and green**. All four behavioral AC rows have named production-entry tests. The fifth is a procedural mutant/full-CI requirement.

Out-of-contract AC rows: none. Bounds: semantic capability canonicalization is not added; nil/empty slices, order, spelling and other representation differences may conservatively cause misses. Cache remains a single current verdict slot per existing filename tuple, so alternating declarations can replace each other's slots but cannot reuse differing findings. No new runtime containment or detector classes are introduced. No end-to-end external skill installation is newly added; the actual production Gate plus the profile member caller are exercised, with full existing installation suites delegated to the remote gate. Managed producer tests are intentionally uncommitted until the orchestrator snapshots/checkpoints the CR.

## Narrowing mutant evidence

Harness command: `go test ./internal/audit -run '^TestCapabilityCacheNarrowingMutants$' -count=1 -v` exited **0**. Each of its 14 nested mutant commands exited **1**, expected-red, with the intended named behavioral failure. A setup error, missing test or build failure does not count. No survivors.

| Mutant | What it narrows the gate to | Named failing test | Survivor bound |
| --- | --- | --- | --- |
| network | Compare all fields except Network | TestVerdictCacheBindsEffectiveCapabilities/network | Killed |
| exec | Compare all fields except Exec | TestVerdictCacheBindsEffectiveCapabilities/exec | Killed |
| filesystem-keyword | Ignore Filesystem.Keyword only | TestVerdictCacheBindsEffectiveCapabilities/filesystem-keyword | Killed |
| filesystem-paths | Ignore Filesystem.Paths only | TestVerdictCacheBindsEffectiveCapabilities/filesystem-paths | Killed |
| secrets | Ignore Secrets only | TestVerdictCacheBindsEffectiveCapabilities/secrets | Killed |
| env-read | Ignore EnvRead only | TestVerdictCacheBindsEffectiveCapabilities/env-read | Killed |
| prompt-scope | Ignore PromptScope only | TestVerdictCacheBindsEffectiveCapabilities/prompt-scope | Killed |
| admit-none-crossover/gate | Bypass unequal-capability rejection only for empty-network target | TestGateEffectiveCapabilityCacheCrossover/v1/declared-first=true | Killed |
| admit-none-crossover/members | Same one-direction capability bypass, preserving other cache dimensions | TestStrictAuditMemberCapabilityCacheIsolation/skill/skill-first; corresponding context and MCP cases also fail | Killed |
| admit-declared-crossover/gate | Bypass unequal-capability rejection only for declared-network target | TestGateEffectiveCapabilityCacheCrossover/v1/declared-first=true | Killed |
| admit-declared-crossover/members | Same opposite-direction bypass, preserving other cache dimensions | TestStrictAuditMemberCapabilityCacheIsolation/skill/skill-first; corresponding context and MCP cases also fail | Killed |
| reuse-ruleset-two | Read/write the prior ruleset namespace while keeping equality validation | TestCapabilityCacheRulesetRetainsOldRecords | Killed |
| accept-null-scope | Keep scope validation except for null | TestCapabilitiesNullRejectedFromV3/v3/agent-skill.json/null-scope | Killed |
| accept-null-schema-three | Keep null-object rejection except at schema 3 | TestCapabilitiesNullRejectedFromV3/v3/agent-skill.json/null | Killed |

No product gate inspects source tokens. The overlay harness uses unique source anchors only to plant mutants; the equality-token-preserving mutants execute behavioral tests, not a static checker. The retained prior six-mutant migration suite also ran green as part of the focused command below.

## Commands and evidence

All validation commands ran as standalone processes with output redirection, without tee or status-masking pipelines. Local environment: Go 1.25.5, darwin/arm64. The focused behavior and first mutant execution predate only the final README command-documentation addition. Their Go source, test, configuration and local environment inputs are unchanged; those suites do not read the repository README. Build, vet and full audit/parser package runs used the full gate candidate. The current candidate tree exactly matches signed gate tree `f273e0429b90c86e693756a876bf77fea7ba93e5`; ignored scratch logs and resource drafts are excluded.

| Command | Exit | Evidence |
| --- | ---: | --- |
| go test ./internal/audit ./internal/capabilities ./internal/skillspec ./internal/envprofile -run 'Test(VerdictCacheBindsEffectiveCapabilities\|GateEffectiveCapabilityCacheCrossover\|CapabilityCacheRulesetRetainsOldRecords\|StrictAuditMemberCapabilityCacheIsolation\|MigratedGlobalSkill\|GlobalSkillCapabilitiesNarrowingMutants\|CapabilitiesNullRejectedFromV3\|Parse)' -count=1 | 0 | focused-04.log |
| go test ./internal/audit -run '^TestCapabilityCacheNarrowingMutants$' -count=1 -v | 0 | mutants-01.log; all 14 nested commands expected-red exit 1 |
| go test ./internal/audit ./internal/capabilities ./internal/skillspec -count=1 | 0 | packages-01.log |
| go build ./internal/audit ./internal/capabilities ./internal/skillspec ./internal/envprofile | 0 | build-01.log |
| go vet ./internal/audit ./internal/capabilities ./internal/skillspec ./internal/envprofile | 0 | vet-01.log |
| git diff --check | 0 | diff-check-02.log |
| gofmt -l internal/audit internal/capabilities internal/skillspec internal/envprofile | 0 | gofmt-02.log; empty output |
| Existing named-test inventory versus HEAD | 0 | test-inventory-02.log: 511 of 511 names retained in affected modules |
| sh scripts/remote-gate.sh | 0 | remote-gate-01.log; https://github.com/relux-works/curator/actions/runs/39049991046 |

Configured workflow bounds: the self-hosted post-landing lane is skipped for gate branches (`.github/workflows/ci.yml`, test-self-hosted condition); alternate candidate-spec conformance runs only on an explicit workflow_dispatch with a candidate input. Neither was executed here. The requested gate uses the authoritative hosted test/race/toolchain/conformance matrix.

Logs above are task-local scratch under `.temp/BUG-261017-8v470p/`; task-scoped sanitized gate/mutant evidence is attached with this handoff. No pre-existing attached green evidence was substituted for these executions. The entire local envprofile suite was not separately rerun: its migration/member-focused tests ran locally; its full suite and all retained reviewer tests passed in the hosted gate. Full local audit/parser package execution additionally checked existing cache, source-audit policy fingerprint and parser invariants beyond the focused rows. Attached evidence: BUG-261017-8v470p_validation.log, BUG-261017-8v470p_mutants.log, and BUG-261017-8v470p_remote-gate.log.

Diagnostic history, not passing evidence: focused-01 ran before candidate tests were applied (exit 0, provides no new-test evidence). Initial broad patch application hit an outdated migration-test hunk and was not applied. focused-02 exited 1 for a fixture import rewritten accidentally during reserved-domain replacement; focused-03 exited 1 for a fixture assigning a bool-returning package builder to error. Both fixture setup issues were corrected before focused-04. Neither failed setup is mutant evidence.

## Diff scope

Production changes are confined to internal/audit, internal/capabilities and internal/skillspec, exactly as scoped. Tests under internal/envprofile verify its production caller and migration publication invariant. README documents exact new test/tool entry points and local artifact location, as repository instructions require. No control-root repository files, unrelated product modules, live data or installed applications were edited. No branch-tip commits, switch, rebase, merge or landing performed.

## Logbook

- Confirmed current base includes the migration fix and reads skill capabilities on the profile skill path.
- Verified the supplied patch against the pinned spec; corrected the brief's stale section reference in this evidence.
- Strengthened crossover tests to remove package-identity masking; the new one-direction mutants kill both strict member orders for all three kinds.
- Preserved the old package-identity mutant's independent purpose with an added stored-binding assertion after capability equality closed its former attack.
- Added schema-9 and both-filename controls without changing pre-v3 absence behavior.
- Full remote gate exited 0 on the signed candidate snapshot: 20 successful jobs, 2 configured skips, no failed jobs. Current candidate tree still matches the gate tree; managed HEAD remains the original checkpoint. Evidence is attached before developer handoff.
