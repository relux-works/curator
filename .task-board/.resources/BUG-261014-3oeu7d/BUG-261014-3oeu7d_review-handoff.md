# BUG-261014-3oeu7d — migrated global skill capabilities

Developer candidate for review. Worktree HEAD remains the recorded checkpoint
`26d8ddbf1278ac025a3333aee1053adbdb37f1bf`; no Story commit, branch switch,
rebase or integration was performed. Signed detached gate snapshot:
`0a2cdb08929afa6b5130c90d9354818077299ff7`.
Normative source was fetched from curator-spec main and read directly:
`1059b94b4135089d842bf017070b8ab47d79a536`, environments §9.1 and §9.4.

## Implementation and candidate use

The supplied patch was reviewed and adapted, rather than applied verbatim.
`strictAuditMember` loads `skillspec.Load` only for skill members and forwards
the skill's declared capabilities. Parse/read errors block. Context and MCP
members retain their previous implicit-none behavior. Modern manifest schemas
are forwarded exactly. The existing profile schema floor 3 is retained for
missing/pre-capability manifests: §9.4 requires the network-undeclared finding
and remedy rather than the generic pipeline's require-pin decision. This is
the explicit orchestrator decision in the updated brief, documented beside
the code and in LOGBOOK. `internal/audit` and the generic skill pipeline are
unchanged.

Skill subjects additionally bind their actual frozen Git package and selected
directory using the existing `audit.SkillPackage` API. A production regression
proved that otherwise an old unstated-package cache could retain a false
implicit-none refusal or admit an undeclared host from an unstated allow. The
new binding makes those old records miss; no generic audit cache change,
namespace flag, or cache-file deletion is used in product code. The regression
fixture removes only its own temporary verdict before seeding the old subject
through the real `audit.Gate`.

Migration retains the strict finding and adds the required covering-host-glob
remedy naming both manifest filenames. No capabilities are synthesized on the
skill's behalf. The previously committed regression test
`TestMigratedGlobalSkillNetworkUndeclaredFollowUp` retains its original name
and now also asserts the remedy and absence of require-pin.

The candidate's two modified files became four envprofile files (product,
expanded controls, narrowing overlays, retained regression). The only change
outside internal/envprofile is LOGBOOK.md, required to persist the schema-floor
decision. No spec, generic audit, README, tool or board-config change is needed.

## AC and production-call coverage

Coverage: **5 of 5 AC rows driven**. The exact candidate full remote gate
returned exit 0 and its hosted run concluded success.

| AC | Production entry and call site | Named executable evidence |
|---|---|---|
| 1 | `EnsureDefaultWithPolicy -> ensureDefault -> migrateGlobalSkills -> strictAuditMember -> audit.Gate`; `Install -> auditAndStore -> auditMembers -> auditMember -> strictAuditMember` | `TestMigratedGlobalSkillNetworkCapabilities`; `TestInstalledMemberCapabilitiesStayKindScoped/declared-skill` |
| 2 | `EnsureDefaultWithPolicy -> migrateGlobalSkills`, before `op.publish` | NetworkCapabilities exact-host-declared and glob-host-declared publish source and lock containing the pinned skill. No-manifest, network-none, other-host-declared, pre-capabilities-schema refuse; ManifestErrorsBlock/malformed-network refuses; DeclarationKeepsOtherGates/neighbouring-host, revoked-source and canary-failure refuse. Every migration refusal asserts both profile files absent. |
| 3 | `migrateGlobalSkills` strict-audit error handling | `TestMigratedGlobalSkillNetworkUndeclaredFollowUp`; NetworkCapabilities refused cases; DeclarationKeepsOtherGates/neighbouring-host assert skill, detected host, script path, capabilities.network, canonical/legacy manifests and leave-unmigrated option. |
| 4 | `Install -> auditAndStore -> auditMembers -> auditMember -> strictAuditMember`; unchanged generic `audit.Gate` | `TestInstalledMemberCapabilitiesStayKindScoped/clean-context-ignores-skill-manifest` admits; `/context-network-not-authorized` refuses. Existing `TestRevokedMCPMemberIsRefused`, `TestRevokedSkillMemberIsRefused`, `TestStrictAuditCanaryPasses`, `TestStrictAuditCanaryFailureBlocksInstall` pass. Full internal/audit and internal/skillspec packages pass, including `TestRequirePinForOldSchemas` and `TestDetectorsAgainstCapabilities`. |
| 5 | Real migration and Install calls driven under overlays; signed candidate runs the configured remote gate | `TestGlobalSkillCapabilitiesNarrowingMutants` (six plants); `TestMigratedGlobalSkillIgnoresPreManifestVerdictCache`; remote run 39033322875, success; direct gate exit 0. |

Additional migration controls: NetworkCapabilities/legacy-alias, schema-nine,
network-absent; ManifestErrorsBlock/malformed-json and unreadable-manifest.
Existing `TestMigratedGlobalSkillBinaryWriteVersion` runs all 12 v1/v2 shapes.
No previously committed reviewer test was removed or renamed.

## Brief surface-map gap and stated bounds

The brief has no catalog surface table. This is an explicit brief gap, not an
implicit waiver: the AC/control map above enumerates all supplied controls, plus both cache directions;
no acceptance-criteria row or supplied control is excluded as out of contract.
There are no claimed out-of-contract rows or inputs.

The unreadable-manifest migration witness is a directory at the manifest path,
which reliably fails the real read on all platforms. It does not claim to
simulate permission-denied. Existing skillspec permission/symlink/fallback
controls remain in place. The canary-failure control uses the existing
`canaryPasses` seam through the real migration entry; it proves fail-closed
handling of a failed canary, not arbitrary corruption of every detector.
The schema-nine case proves admission under a valid modern schema; the exact
schema assignment is also verified by source review. Source identity and
revocation are exercised with synthetic example-domain fixtures only.
No new product gate inspects source text, so the additional source-text-gate
mutation obligation is inapplicable; every overlay runs behavioral tests.
The neighbour and cache overlays also preserve the original capability-load
and package-binding tokens while changing their behavior; the behavioral
controls fail rather than relying on a source-text checker.

## Mutation evidence

All child gate commands are expected-red `go test` executions, real exit 1;
the enclosing harness exits 0 only after a named intended failure. Unique
anchors and missing-test/compile-failure checks prevent false kills. The
candidate is not edited by the overlays. No mutant survived.

| Mutant | What it narrows the gate to | Named failing test | Result / survival bound |
|---|---|---|---|
| implicit-none | Skill capabilities are ignored while audit stays active; this is an admission-restricting regression control requested by the brief | `TestMigratedGlobalSkillNetworkCapabilities/exact-host-declared`, `/glob-host-declared`, `/legacy-alias`, `/schema-nine`; exactly those 4 fail and all 5 refused cases remain green | Killed, child exit 1; no survival |
| malformed-network-fallback | Only capabilities.network parse errors fall back to implicit-none; other manifest read/JSON errors still refuse | `TestMigratedGlobalSkillManifestErrorsBlock/malformed-network`; malformed-json and unreadable-manifest remain green | Killed, child exit 1; no survival |
| undeclared-neighbour-admitted | A declaration for downloads.example.com additionally authorizes only uploads.example.com; surrounding audit remains active | `TestMigratedGlobalSkillDeclarationKeepsOtherGates/neighbouring-host`; revoked-source and canary-failure remain green | Killed, child exit 1; no survival |
| context-authorized-by-skill-manifest | Only context members additionally read skill declarations; MCP and other gates stay present | `TestInstalledMemberCapabilitiesStayKindScoped/context-network-not-authorized` and `/clean-context-ignores-skill-manifest`; declared-skill stays green | Killed, child exit 1; no survival |
| unstated-verdict-reused | Only the new skill package binding is cleared; commit, content hash, selected directory and strict gate remain | `TestMigratedGlobalSkillIgnoresPreManifestVerdictCache/undeclared-skill-old-allow` and `/declared-skill-old-refusal` | Killed, child exit 1; no survival |
| legacy-schema-bypasses-network-finding | Only pre-capability/missing schemas lose the profile floor; modern capability checks remain | `TestMigratedGlobalSkillNetworkCapabilities/no-manifest` and `/pre-capabilities-schema` report pin instead of required network finding | Killed, child exit 1; no survival |

## Direct command ledger

All commands run unpiped, with output redirected to the named log. No incomplete
or inherited green log substitutes for these executions. Host: Go 1.25.5,
darwin/arm64. The product supports multiple OSes; the full gate qualifies its
configured platform matrix. No iOS build is applicable to this Go CLI.

| Command | Real exit | Evidence |
|---|---:|---|
| `go test ./internal/envprofile -run 'TestMigratedGlobalSkill(NetworkCapabilities\|DeclarationKeepsOtherGates\|ManifestErrorsBlock)\|TestInstalledMemberCapabilitiesStayKindScoped' -count=1` on unchanged product | 1 | regression-before-01.log; expected red: declarations ignored, malformed manifests admitted, remedies absent |
| `go test ./internal/envprofile -run 'TestMigratedGlobalSkill(NetworkCapabilities\|DeclarationKeepsOtherGates\|ManifestErrorsBlock\|NetworkUndeclaredFollowUp)\|TestInstalledMemberCapabilitiesStayKindScoped' -count=1 -v` after the initial manifest-only change | 0 | regression-after-01.log; historical source identity |
| `go test ./internal/envprofile -run '^TestGlobalSkillCapabilitiesNarrowingMutants$' -count=1 -v` on the initial snapshot | 0 | mutants-01.log; five historical child executions exit 1; superseded by regression-after-03.log with six plants |
| `go test ./internal/envprofile -run 'TestMigratedGlobalSkill\|TestInstalledMemberCapabilitiesStayKindScoped\|TestGlobalSkillCapabilitiesNarrowingMutants\|Test(Revoked(Skill\|MCP)MemberIsRefused\|StrictAuditCanary(Passes\|FailureBlocksInstall))$' -count=1 -v` | 0 | regression-after-03.log, revised candidate including cache controls and all six plants; 12 top-level tests |
| `go test ./internal/audit ./internal/skillspec -count=1` | 0 | generic-audit-01.log |
| `go vet ./internal/envprofile` | 0 | vet-02.log |
| `go build -o .temp/BUG-261014-3oeu7d/curator ./cmd/curator` | 0 | build-02.log |
| `git diff --check` | 0 | self-review, no whitespace errors |
| `sh scripts/remote-gate.sh` after sourcing the required agent git environment | 0 | remote-gate-02.log, https://github.com/relux-works/curator/actions/runs/39033322875 |

No test failures were accepted as unrelated. The current remote run has
20 successful jobs, with two optional lanes skipped by workflow configuration;
the skipped candidate suite and optional self-hosted macOS lane are not claimed
as passing evidence. Validation reuses only identity-bound evidence;
identity-02.json records exact candidate-file SHA256 values and toolchain.
The local generic audit/skillspec command is reused against unchanged source,
tests, configuration and local environment; the new full gate reruns its own
complete configured suite.

## Cache-control regression and identity change

`go test ./internal/envprofile -run '^TestMigratedGlobalSkillIgnoresPreManifestVerdictCache$' -count=1 -v`
ran twice before the package-binding fix. cache-regression-before-01.log exits 1
only because its seed setup hit an existing verdict; that is not a valid
negative-control verdict. The isolated fixture was corrected to remove its own
verdict before the production Gate seed. cache-regression-before-02.log exits 1
with both intended named behavioral failures: the declared skill reused the
stale refusal and the undeclared skill published after the unstated allow.
Both controls now pass in regression-after-03.log. The sixth overlay repeats
those failures by clearing only the new package binding.

Initial gate run 39030855485 tested snapshot
`dfa249c6c0cce3392b69a116957eee4bf00044db`; it does not qualify this revised
candidate. Its Linux and macOS artifacts each contain 42 passing records for
the original controls; those records exclude the subsequently added cache
controls. The source/test identity changed for a diagnosed regression, which
is the reason for rerunning the full gate on snapshot
`0a2cdb08929afa6b5130c90d9354818077299ff7`. No initial-snapshot green result
is substituted for the new full-gate verdict.

The historical `sh scripts/remote-gate.sh` invocation (remote-gate-01.log)
returned real exit 0 for run 39030855485. This is recorded as historical green,
not current-candidate qualification. The revised candidate obtained its own
real exit 0 from remote-gate-02.log and run 39033322875.

## Current hosted evidence and candidate identity

Linux, macOS and Windows hosted test artifacts each contain 46 passing named
records for the migration, cache, Install and mutation controls. The attached
hosted-controls-02.json preserves those extracted records and the real job
verdicts without runner paths. The remote-gate-02.log resource includes the
observed standalone process exit 0 and exact candidate proof. The full staged
candidate Git tree equals the signed gate snapshot tree
`88c4200db2dcfce899bde15444031a6e4ea28d13`; checkpoint HEAD is unchanged.
All five candidate file hashes were rechecked after CI. Candidate changes
remain uncommitted for the developer handoff snapshot.
