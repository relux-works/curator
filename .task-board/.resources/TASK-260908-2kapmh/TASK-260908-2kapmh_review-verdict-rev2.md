# TASK-260908-2kapmh — reviewer verdict, CR revision 2

**Verdict: accepted. R1 is closed.** Candidate tree `b9ab9127e6761067d82c655f75eafe8bece1abe7`, base `12f443d10bc217ca7a48e2edab19c739f441df9c`. Repository delta is present and required. Review scope is the 11-file R1 delta from accepted-in-principle rev1 tree `38639918`; unchanged areas retain the prior review evidence. All 310 regular candidate file blobs matched before review and after restoring mutants. No product edits or commits.

## Coverage and production behavior

**7 of 7 R1 behavioral AC rows driven**, using the candidate tests (left uncommitted as required by the managed Story contract):

| Row | Named driving test | Production call site |
| --- | --- | --- |
| sol ultra refusal | TestBuildLaunchRefusesAnEffortInstalledPiWouldDropOrClamp/gpt-5.6-sol/ultra | BuildLaunch → EffortAdmitter |
| terra ultra refusal | same test, gpt-5.6-terra/ultra | same |
| codex-model minimal refusal | same test, gpt-5.3-codex/minimal | same |
| gpt-5.2 minimal refusal | same test, gpt-5.2/minimal | same |
| sol max unchanged | TestTheMaxWordStillBuildsWhereInstalledPiRunsIt | BuildLaunch → pinative.Args |
| opus max unchanged | same test | same |
| all four pairs remain supported by Codex | TestBuildLaunchRefusesAnEffortInstalledPiWouldDropOrClamp, Codex assertions | BuildLaunch → Codex argv |

Errors preserve both error identities and name model, runtime, requested effort, native subset and row recommendation; refusals return no argv. The capability dispatch occurs after global row validation, uses binding-owned vendor and resolved launch identity, and does not rewrite effort. Only pinative implements the optional capability, so existing systems retain their behavior. Args independently invokes the same restriction with LaunchIdentity(), closing the direct-plugin argv path. Alias identity, home, vendor overwrite, no-preflight and provider-limit tests also passed in the narrow rerun.

`TestPiNativeThinkingRestrictionsMatchTheInstalledCatalog` passed **71 of 71 row/word pairs, 142 of 142 BuildLaunch calls** across interactive/dry-run. Expected native support comes from installed parser/catalog bytes; the test checks admission and refusal in both directions and exact argv for supported pairs. No installed-source test skipped in this review; all parser/catalog reads succeeded. Non-ENOENT parser errors and all catalog read/parse errors fail rather than imply absence. Optional missing-install skips do not constitute installed-Pi proof on another host.

The original 24/24 broader AC mapping and unchanged legacy/sourceport/smoke behavior are inherited from rev1 evidence, not claimed as newly re-proven in full. Rev2 changes no global model row or membership, wrapper/local-model implementation, frozen runtime or provider-limit implementation.

## Independent attacks and validation

All reviewer commands used Go 1.25.5 on darwin/arm64; broad suites were not duplicated.

| Command / attack | Exit and evidence |
| --- | --- |
| go test pinative/vendorplugin/providerlimits with Pi/Native/MaxWord/ArgsRefuses/OverwritesAVendor selectors, -count=1 -v | 0, narrow-01.log |
| Token-preserving nativeAccepts narrowing admitting only sol ultra | 1, mutant-ultra-sol.log; BuildLaunch refusal, installed matrix and direct Args tests fail |
| Null-map narrowing admitting only codex-model minimal | 1, mutant-minimal-codex.log; same behavioral tests fail |
| Capability dispatch retained but skipped for sol | 1, mutant-dispatch.log; BuildLaunch error-contract test fails despite the later Args guard |
| Restored focused suites | 0, restored-01.log and restored-02.log |
| Installed parseArgs/getSupportedThinkingLevels/clampThinkingLevel | 0, catalog-behavior.log |
| Seven isolated offline Pi invocations | probe driver 0; each Pi exits 1 at missing API key, pi-probes-02.log |
| git diff --check and final candidate blob comparison | 0 |
| Exact-rev2 runtime make vet/test/regress | all 0; reused board resource TASK-260908-2kapmh_change-request_rev2-validation.log |

The mutant attacks cover narrowed admission and check-present-but-uncalled-from-production shapes, preserve source tokens, and run behavioral tests rather than only a static checker. All mutants restored from captured original bytes. No survivors.

## Installed boundary and limitations

Independently read Pi 0.84.2 parser and models.js behavior. Actual parser drops ultra with a warning; actual clamp maps minimal to low for gpt-5.3-codex and gpt-5.2. max remains max for sol and opus; high parses unchanged. Catalog SHA-256 values match prior evidence: anthropic `e22c277e3a1ffddc3d2701b72787c9e0bd67b835de6b4cb806677b6b6a89a2f7`, openai `47746dfe79d92a7da58d7fae2c02a5d481b2eba2f5a914a42792552940b78b78`, google `c6d9822f7eda23cfeb7d29caee0baa7d2fa7bf3561252734a17428937fbcbfd1`. Full hashes in catalog-sha256.log.

Probes used fresh isolated homes/cwd and an explicit environment allowlist, offline with extensions/skills/templates disabled. No credentials, auth mutations or model calls. This proves parser/catalog and missing-auth boundaries, not an authenticated interactive session. Effort-none plans inject no thinking flag; Pi may use its own settings default, now stated in all four updated docs. Native restrictions are pinned to Pi 0.84.2. Direct plugin use does not replace vendor model admission. pi-google classifier and unmapped provider-limit bounds from rev1 remain unchanged.

Checklist: implementation and architecture accepted; relevant tests and local validation green; gate attacks evidenced; task-scoped evidence attached. No LOGBOOK per operator instruction; this artifact records the findings. Run goal queried: this run is not goal-bound. No hosted CI, tags, installs, daemon actions, private board writes or control-root writes.

Handoff: accept_cr revision 2 routes to integrating, not done. Parent owns signed PR 23 update and exact-head landing; operator alone creates signed annotated v0.5.11 after landing verification. No release performed by reviewer.
