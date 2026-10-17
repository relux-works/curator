# BUG-261017-8v470p — audit-cache-ignores-capabilities: review revision 1

Verdict: accepted. Candidate tree `f273e0429b90c86e693756a876bf77fea7ba93e5`,
base `fcd785ba93c5d00e2e80615f8add2baa93ccd73d`. Repository delta is present.
The full manifest comparison and ruleset namespace separation address the
latent crossover while preserving independent package identity checks.
Null rejections match the normative schema. No code was modified by review.

```json
{
  "verdict": "accepted",
  "revision": 1,
  "findings": [],
  "notes": [
    "No surface table or review/free-hunt budget was supplied. The five numbered review-brief requirements were swept as rows; catalog mapping is identity for rows 1-3, text grammar/error surface for row 4, and evidence validation for row 5.",
    "Core section 4.4 is Dependencies on both the producer pin and current spec main. Core 4.3/common schema support null rejection; skillfile-sources section 4 supplies package cache equality.",
    "The producer attribution advisory remains advisory under the configured warn policy; this review does not claim to clear it or establish attribution for other runs.",
    "Acceptance is not landing. Current main has two subsequent board-only commits, with no overlap in the 12 reviewed paths. Landing must retain the accepted source and satisfy its own freshness checks."
  ],
  "surfaces": [
    {"row": "1-capability-equality", "result": "held", "attacks": "Seven independent field-narrowing mutants; identical-input control through auditSubject and Gate."},
    {"row": "2-crossover", "result": "held", "attacks": "Declared-first and implicit-none-first across skill/context/MCP; both directional narrowing mutants; Gate v1/v2 cache-hit controls."},
    {"row": "3-prior-records-and-package-binding", "result": "held", "attacks": "Ruleset-2 seeded allowance remains byte-identical; namespace mutant killed; independent package/commit/directory controls and retained migration package mutant pass."},
    {"row": "4-null-declarations", "result": "held", "attacks": "Load schemas 3-9 and both manifest names; absent/valid/empty controls; null object/scope narrowing mutants; migration refusal without publication."},
    {"row": "5-gate-and-public-hygiene", "result": "held", "attacks": "Fresh hosted result/head query; all 12 candidate blobs compared; exact gate tree checked; changed diff inspected for public hygiene."}
  ],
  "free_hunt": {"findings": [], "result": "held", "attack": "GateReadOnly with differing nonempty network/exec declarations, v1/v2, both orders; repeat decisions and cache-byte preservation; writable retry hit."}
}
```

Coverage: **5 of 5 AC rows driven**. Named tests and production call sites:

| AC | Named driver | Production path |
| --- | --- | --- |
| 1 | TestVerdictCacheBindsEffectiveCapabilities | auditSubject -> auditSubjectWithOpaquePaths -> loadCachedFindings |
| 2 | TestGateEffectiveCapabilityCacheCrossover; TestStrictAuditMemberCapabilityCacheIsolation | Gate; strictAuditMember -> Gate |
| 3 | TestCapabilityCacheRulesetRetainsOldRecords; TestVerdictCacheBindsFullPackageIdentity; TestVerdictCacheBindsSelectedDirectory; TestMigratedGlobalSkillIgnoresPreManifestVerdictCache | auditSubject; EnsureDefaultWithPolicy -> strictAuditMember |
| 4 | TestCapabilitiesNullRejectedFromV3; TestMigratedGlobalSkillPromptScopeValidation; TestMigratedGlobalSkillNullCapabilityDeclarations | Load -> loadSkillManifest -> capabilities.Parse; EnsureDefaultWithPolicy |
| 5 | TestCapabilityCacheNarrowingMutants; TestGlobalSkillCapabilitiesNarrowingMutants; remote-gate.sh | Behavioral overlay suites; hosted CI workflow |

Reviewer execution: focused four-package behavioral command, all 14 new
narrowing mutants, all six retained migration mutants, independent cache package
and directory controls, existing GateReadOnly controls, and the additional
four-case reviewer overlay passed (exit 0). Individual mutant commands failed
with the intended test/message; no setup failure is counted as a kill.
Commands and logs are in `BUG-261017-8v470p_review-evidence-rev1.log`.
The extra probe is `BUG-261017-8v470p_review-probe-rev1.go`; reproduce by overlaying
it as an additional audit package test, then run
`go test ./internal/audit -run '^TestReviewerNonemptyReadonlyCrossover$' -count=1 -v`.
Use Go `-overlay` to keep the candidate unchanged.

Full remote gate was reused, not rerun: [run 39049991046](https://github.com/relux-works/curator/actions/runs/39049991046),
completed/success, 20 successful jobs and two configured skips. Fresh GitHub
metadata confirms head `1f43a073eb958dd0f480ee96af2f871a76a44825`; its tree equals
the reviewed candidate exactly. Producer gate log is complete through the
result and matrix summary. The workflow and all validation inputs are unchanged.
Full local envprofile suite was not rerun; focused behavior and mutation suites
were. Hosted matrix supplies full-package/race/toolchain coverage.

Normative sources independently read at current curator-spec main
`9dfc557d247243e0e8ba60fb877afe6d1facfd06`: protocol/core.md 4.3,
protocol/skillfile-sources.md 4, protocol/environments.md 9.1/9.4, and
schemas/v1/common.schema.json capabilities definition. These retain the
producer's cited requirements. The equality check safely permits conservative
misses for representation differences; the single cache slot can be replaced
by alternating declarations. No canonicalization or new detector behavior is
claimed. Results establish the named attacks only, not proof of absence.

Logbook: independently reproduced both crossover mutants, confirmed migration
package identity remains independently checked, verified current spec null
semantics, and exercised dry-run/retry beyond the producer table. No blocking
finding or implementation change resulted. Run goal read: not goal-bound.
