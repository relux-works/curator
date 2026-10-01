# TASK-261001-2yvag1 — curator-muse-environment

Resolver scope is ready for review under the binding 2026-10-01 scope split. Changes remain uncommitted in the assigned Story worktree. The curator-run provider mapping and v3 reader are the separate launcher leaf; no launcher repository was changed.

## Resolver behavior

Muse has four ordered XDG parents under its managed home: config, data, state, cache. HOME is preserved. Skills materialize at data/muse/skills with discovery explicitly unverified. No root-context target, system-prompt channel or MCP channel is invented. Status names the foreign-personal-context isolation gap and unverified surfaces; isolated credential operation refuses.

Optional config/muse/settings.json and trust.json come only from the locked root snapshot, pass profile/secret validation and become tool-owned after provisioning. Overlays and native homes do not supply these seeds. Profile auth.json is forbidden. Shared config/muse/auth.json links to the effective native XDG config auth path observed before overrides. State reads use internal/stateread; repairs use E5 nofollow writers. Every resolve checks link liveness. Only absent or empty-directory entries with a readable regular native target can be repaired. Forks, wrong targets, dangling/nonregular targets and unreadable observations refuse without changing credential bytes or emitting a fragment. Global reconciliation preserves explicit Muse provisioning and reconciles existing homes.

Muse env resolve emits launch-env-fragment-v3. Other adapters retain v2. Exact counts and owned gaps are keyed by manifest digest; the committed rc.13 pin is unchanged.

## Third gate recovery

The previous handoff ran `sh scripts/remote-gate.sh`, real exit 1, on snapshot 94f8cd8e00cb625930610f6fadd9d16829ba88f1. [Run 36827469593](https://github.com/relux-works/curator/actions/runs/36827469593) passed all three test lanes (macOS, Linux, Windows), Linux race, lint, naming, interop and gate self-tests. macOS race was the sole failing job. Raw downloaded go-test streams show four failing child rows (foreign user, bare prompt, no arguments, extra argument) in TestDraftSourcesBrokerAskpassDispatch, plus the aggregate test/package failures. The error is the test helper's HTTPS broker secret transport write returning broken pipe after a refusing child has already exited. No data-race report occurred in the failed stream. Raw failure rows and lane summaries are attached in the new evidence archive.

The original intermittent case passed ten local race-enabled repetitions before the correction (exit 0); this is not claimed as a reproduction. A new deterministic child-exit regression uses a synthetic payload larger than the Unix pipe buffer and reproduced the helper failure before the fix, real exit 1.

Only internal/testcli/cli.go and its new tests change in this recovery. After waiting for the actual child exit, the helper accepts only a delivery error whose complete cause chain consists of EPIPE. Other transport errors, close errors, and joined errors containing an additional cause remain failures. Child exit and stdout are still returned unchanged to the real compiled broker tests, which assert exact prompt authorization, silent refusal and absence of leaked credentials. Tests cover unused-handle children returning both 0 and 1, as well as permission, closed-file and mixed joined-error rejection. Production broker and transport behavior are unchanged. All 26 previous changed/new source-file hashes still match the attached r3 identity.

## Fresh verification

Each command ran as a standalone process without a pipe or tee. Exact commands, real exit codes and log SHA-256 values are recorded in TASK-261001-2yvag1-r4-exits.json inside the new archive.

| Validation | Real exit |
| --- | --- |
| New child-exit regression before correction | 1 (expected red; reproduces helper failure) |
| Original broker case before correction, race, count=10 | 0 (intermittent failure did not reproduce locally) |
| Corrected broker dispatch, child-exit and error-classification tests, race, count=10 | 0 |
| Candidate Muse tests, actual resolver CLI and fake child | 0 |
| Candidate affected package regressions | 0 |
| rc.13 affected package regressions | 0 |
| Seven status/audit regressions, race | 0 |
| go build ./cmd/curator | 0 |
| go vet ./... | 0 |
| Windows scoped go vet | 0 (static/compile check, not Windows execution) |
| golangci-lint run | 0 (0 issues) |
| git diff --check | 0 |
| gofmt -l cmd internal | 0 (no output) |

Published Muse link-state cases: 16/16 driven, 0 known gaps, 0 bounds, 0 skips. These exercise production StatusOf and Resolve. Published fragment-v3 cases: 36/36 driven, 0 known gaps, 0 bounds, 0 skips, through schema validation plus production CheckBoundary. Registry/layout rows, exact XDG set, preserved HOME, optional-seed preservation, synthetic native credentials, profile secret/auth refusals and additional link states pass. The reader audit remains 471/471 covered (365 through the seam and 106 existing reviewed exceptions, 437 production files); no audit rule was weakened.

All three required compiling mutants were rerun against the current tree and killed, 3/3. They use the existing source overlays; their Muse production inputs are byte-identical to the r3 source hashes. Each is a failing test command with real exit 1, never a passing gate:

- HOME added in production buildFragment: actual env resolve rejects the undeclared HOME override.
- Only XDG_DATA_HOME removed in production buildFragment: actual env resolve rejects the missing declared parent.
- Only replaced-file fork refusal narrowed to acceptance in inspectMuseAuth: production StatusOf incorrectly becomes current and fails TestMuseForkRefusedRegression. Resolve shares verifyHome.

## Identity and bounds

Candidate clone: /tmp/TASK-261001-2yvag1-spec, commit d373078a27a437071c2416d447221e77cec65725; conformance manifest bd03456b92a7368d90ea74fe6953db10bc188588020683024a6a6b8735a40783. Committed SPEC_PIN remains rc.13, 23435129ebc4c29e5b7f75ec72a0aa0cd3f16065, manifest be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca; fresh rc.13 tests use /tmp/TASK-261001-2yvag1-rc13/conformance/v1.

Current 28-file source identity: da1ff15bd1d01e64813a6c1efdcba5117517d73f5f2732f2e7f21c2efa8d1c73, base c803afd7e9ed9fe016f85be10fc366cd6dccc0e1. The attached identity defines its digest algorithm and records each changed/new file hash. No commit, branch switch, merge or rebase occurred.

curator run muse needs a Muse mapping and fragment-v3 reader in curator-run, the separately assigned launcher leaf. The earlier isolated launcher probe is accepted from previously attached evidence, not rerun here: install=0, resolve=0, run=1, resolve_fragment_invalid. That failing run row is an explicit bound. The fake-child tests apply the actual resolver fragment directly and do not establish curator-run integration. Exec/serve/session permission mappings are outside this leaf.

Native root-context loading, native skills discovery, refresh coordination, and internal state/cache layouts remain unverified. Content-hash-v2 and Muse marker-v3 emission remain explicit digest-scoped owned gaps assigned to TASK-260917-2tx81l; current markers stay v2. Previous green remote lane evidence is distinguished from fresh local checks above. The full repository test/race/remote matrix was not rerun by this developer; the task runner invokes the configured gate after handoff. No new full-gate pass is claimed.

## Entry text

Register Muse's managed XDG environment and shared native auth link; preserve HOME, audit root-profile settings/trust seeds, enforce credential detachment and read-failure handling, and emit fragment v3. Keep status fixtures current across registered adapters with synthetic auth and an inert Muse binary. Repair the compiled-broker test helper's unused secret-pipe lifecycle without changing production transport or relaxing child authorization/output assertions; add a deterministic regression and negative error-classification rows. Record launcher ownership and discovery/refresh bounds. Findings and entry text are in this outcome as required; no CHANGELOG or LOGBOOK edits.


## Revision 5 (refresh) — blocked before refresh

Run: RUN-261001-2759d9. Requested trunk: bd126a9a.

The required command `task-board worktree refresh-candidate TASK-261001-2yvag1` was executed directly and returned exit 1:

```text
candidate refresh requires a rework revision; TASK-261001-2yvag1 is ready
```

`task-board worktree status STORY-261001-1xuwlu --json` returned exit 0 and confirms revision 4 is still `ready`, base c803afd7e9ed9fe016f85be10fc366cd6dccc0e1, candidate tree 71308ba696e14ea5fba40fed5b65fa298f3effdb, with this run holding the Story lease. The failed reviewer preparation did not route this revision into rework. The earlier reviewer refusal names the supported remedy: `task-board worktree converge STORY-261001-1xuwlu --reason "trunk advanced on changed candidate paths"`. The project-management recovery reference requires that command from the control root outside a producer run; a tracked run is refused. `invalidate-acceptance` also requires an operator and an integrating element, so it does not apply to this ready revision.

Per-file resolution summary: no resolutions or repository changes were made. `.github/ci/conformance-case-counts.tsv`, `cmd/curator/env.go`, `internal/envprofile/managed.go`, and `internal/envprofile/status.go` retain revision 4 contents. Neither side's behavior has been combined yet. Counts were not recomputed against a refreshed candidate.

Not run in this attempt: `go test ./cmd/curator ./internal/envprofile -count=1`, the 16 Muse link-state rows, the 36 fragment-v3 rows, the six env-status/guard regression tests, TestPosture* and subcommand --help rows, build, lint, and mutants. The refresh failed before a revision 5 source tree existed; running these against the old base would not validate the required combined tree. Prior revision 4 hosted green evidence (run 36833591988) is historical evidence only, not revision 5 evidence.

Constraint: board lifecycle and operator-owned workspace convergence, not a product or source-code defect. Manual Git rebase/merge/commit, direct board edits, self-issued reviewer rejection, or withdrawal would bypass the managed lifecycle and are not viable repairs.

Recommended external action: after this run releases its Story lease, the orchestrator runs the named `converge` command from the control checkout, resolves any refusal through its supported instructions, and launches a developer rework run once revision 4 is stale/rework eligible. Then refresh, preserve both sides, run the requested checks with real exits, recompute exact digest counts, attach evidence, and hand off revision 5. No human product decision is needed.

No developer handoff was attempted because the required refresh and validation have not occurred. No launcher, askpass, CHANGELOG, or LOGBOOK changes were made. The separately tracked askpass EPIPE issue remains out of scope. Logbook entry text: revision 5 refresh was refused because revision 4 remained ready after reviewer preparation failed; operator convergence is required before developer rework.

## Revision 5 (refresh)

The preceding refresh-refusal section is historical. Operator convergence resolved that lifecycle prerequisite. This run worked at base bd126a9acdc51b6061917ba8c4d7d26a7abafd41; `task-board worktree refresh-candidate TASK-261001-2yvag1` returned real exit 0, `refresh_already_current`, with trunk, reviewed trunk and branch all equal to that base. No replay, commit, branch switch, rebase or merge was performed. The candidate stays uncommitted for the managed handoff.

### Per-file resolution summary

- `.github/ci/conformance-case-counts.tsv`: retained all 189 trunk count rows exactly, including the four new acquisition families. Recomputed all 103 Muse candidate families from d373078a fixtures instead of copying another candidate's totals. The Muse sibling skillfile-sources index contains 121 rows, not the carried 131. Added Muse-digest acquisition counts: cases 12, common-fetch-argv 55, clean-environment 17, forbidden-fetch-features 11. rc.13 and the other candidate's rows are unchanged.
- `cmd/curator/env.go`: the carried delta against refreshed trunk is only the two-line Muse native-home case. Trunk's parent-pid posture warning and FlagSet subcommand-help behavior remain intact.
- `internal/envprofile/managed.go`: retained the carried Muse XDG/profile/auth implementation together with trunk. Reviewed its additive Muse paths and the guarded credential-link reader. Added an internal marker-observation kind to the existing verifier: absent only after a successful marker read returns absence, present only after successful parsing. This carries absence evidence into status without re-reading state. Resolve refusals, posture, permission precedence and help handling remain intact.
- `internal/envprofile/status.go`: retained trunk plus the carried Muse inspection field, standing note and native-home resolution. The complete TestEnvStatus run exposed a collision with trunk's TestEnvStatusCheckCurrentScopeOnly: an unused Muse home made a healthy current profile non-current. Fixed aggregation to keep a Muse row informational only when marker absence is established by the verifier and backup inventory is known. The row stays visible and unprovisioned; unreadable/malformed markers, read failures before verification and unreadable backups still fail --check. The marker kind is internal, so the public JSON shape stays unchanged. Trunk's fixture is unchanged.
- `.github/ci/conformance-gaps.tsv`: removed seven Muse-digest sibling schema gaps for fixtures absent from this candidate. Other digest rows are unchanged.
- `internal/conformancecoverage/content_hash_v2_gaps_test.go`: updated only the Muse owned-gap assertion from 88 to the measured 81 after removing those absent rows; the hash-v2 candidate still expects 87. The Muse candidate retains rc.13's sibling corpus, unlike the hash-v2 candidate.
- `internal/buildrepo/acquisition_conformance_test.go`: combined the newly landed acquisition harness with Muse's candidate suite by admitting its exact manifest digest alongside rc.13. The candidate and rc.13 acquisition-vector bytes are identical, SHA-256 e8fb4420bb1d17322a50a4a87fa2dfc4ac87991afb59cab784549fd0bd35e0f1. Unknown digests still refuse, protocol-version and digest-keyed count checks still run, and production acquisition code is unchanged. The initially observed candidate-digest refusal is resolved by this test-only compatibility change.

`internal/envprofile/muse_test.go` adds four production StatusOf rows for this collision: known absence, unreadable marker, unreadable backups and malformed marker. The existing trunk CLI test is retained as an independent integration regression. All other carried files are unchanged by this refresh. There are 29 candidate paths, including that additional existing test file. No conflict markers, new repository artifacts, CHANGELOG or LOGBOOK edits were introduced. No new askpass, broker, crossconformance or launcher changes were made; the carried revision-4 test-helper changes remain unchanged. BUG-261001-2772iz, the unrelated askpass EPIPE flake, remains outside this task.

### Fresh evidence and limits

Every gate ran directly without tee or a pipe chain. TASK-261001-2yvag1-r5-exits.json in the new task-scoped evidence archive records exact commands, real exits and log hashes, including unsuccessful attempts. The count recomputation script, isolated mutant sources, source identity and logs are included.

Successful bounded checks, each real exit 0:

- Candidate Muse production Resolve/StatusOf link-state rows: 16/16 driven, 0 known gaps, 0 bounds, 0 skips. The CLI resolver and inert fake Muse child pass the XDG/HOME/seed/argv checks.
- Candidate fragment-v3 rows: 36/36 driven, 0 known gaps, 0 bounds, 0 skips. These are schema-oracle and production CheckBoundary consumers, not 36 separate CLI resolve invocations; the emitted CLI fragment is checked separately by the fake-child test.
- The four prior CLI status failures pass in a standalone bounded run; the reader guard and empty-allowlist test also pass. Registry posture, missing/unreadable approval records and hook trust are tested with an otherwise-current matrix that explicitly provisions the registered adapters using synthetic native auth and an inert Muse binary. This refresh preserves those fixture semantics and separately tests a current profile that never enabled Muse, including negative read-failure cases. The guard now reports 472/472 relevant reads covered, 366 through the seam and the same 106 reviewed exceptions; no allowlist entry was added.
- All 12 TestEnvStatus cases pass in the requested standalone command, real exit 0 (309.350 seconds), including the unchanged TestEnvStatusCheckCurrentScopeOnly trunk fixture. Candidate CLI Muse/posture/help checks and rc.13 scoped envprofile checks were rerun after the currency fix and returned 0.
- TestPosture* and all TestGlobalInit*Help rows pass in their own invocation, including env resolve/global add FlagSet output and parent-pid-bound warning behavior.
- Candidate nine-package core regressions pass after the exact gap-count correction. rc.13 nine-package core regressions and rc.13 coverage-policy rerun pass. The 121 sibling schema cases pass through their production readers and existing explicit bounds.
- Acquisition conformance passes under both candidate and rc.13 roots: 12/12 cases, 55/55 common argv rows, 17/17 environment rows, 11/11 forbidden-feature rows; all driven, no gaps, bounds or skips.
- Build, vet, lint (0 issues), formatting (no output) and whitespace validation pass. Build and lint were rerun after the test-only acquisition admission change.

Five valid compiling mutants were killed, 5/5, each a failing test command with real exit 1, never a green gate:

1. HOME added in production buildFragment: actual CLI resolve refuses the undeclared HOME override.
2. Only XDG_DATA_HOME omitted in production buildFragment: actual CLI resolve refuses the missing declared parent.
3. Only replaced-file fork refusal changed to acceptance in inspectMuseAuth: the production status detachment regression fails.
4. Auth metadata observation bypasses stateread with raw os.Lstat/absence handling: the source guard fails at inspectMuseAuth. An initial Go-overlay attempt returned 0 because the guard scans disk sources; that attempt is not counted as a killed mutant. The valid repeat ran in an isolated source tree containing the mutation on disk and returned 1 with the specific seam violation. A scratch-copy setup failure is also recorded separately, not as mutant evidence.

5. Optional participation narrowed incorrectly to skip every unprovisioned Muse row: the unreadable-marker and unreadable-backup negative cases fail.

The requested exact `go test ./cmd/curator ./internal/envprofile -count=1` was run and returned 1 after both packages hit Go's default ten-minute timeout. The CLI was executing TestProductionExternalDepsFalseDrivesMirrorFetchToSink and envprofile TestRemoveRefusesCurrent. No full-package green result is claimed. The first complete TestEnvStatus run also returned 1 on TestEnvStatusCheckCurrentScopeOnly; that genuine combination regression was fixed without changing the trunk fixture and the complete status suite was rerun. Several overlapping CLI attempts returned 1 at the shared GOROOT TestMain lock before tests could execute; the six-minute aggregate scoped attempt also returned 1 at its timeout. Successful independent bounded commands replace those attempts only for their explicitly named scope. The full repository/hosted gate is not rerun by this developer; prior hosted revision-4 run 36833591988 is historical green evidence, not proof for the refreshed tree. The runner owns the gate after handoff.

Source identity: c123b0bd44ae608236bfb40afb28216feb4d4ac13a7786bf6f19efdf0ddbe75e (29 paths). The source identity artifact defines its algorithm and covers every changed/new file at the refreshed base. The launcher mapping and v3 reader landed in the separately owned launcher leaf (ee66c107, per the binding instruction); no launcher repository was touched and no curator-run integration row was rerun here. Its run behavior remains outside this resolver-only handoff. Native discovery, foreign personal context, refresh coordination and internal state/cache layouts retain the previously stated bounds.

Entry text, retained in results instead of LOGBOOK/CHANGELOG: refresh the resolver candidate onto bd126a9a without changing trunk posture/help behavior; recompute Muse count and gap pins from the actual corpus; admit the exact Muse digest in the newly landed acquisition test after verifying byte-identical vectors; retain rc.13 behavior; preserve healthy-profile currency for known-absent optional Muse homes while failing closed on unknown state; attach bounded green checks, five killed mutants and honest timeout evidence for review.

Lifecycle preflight: the first developer handoff returned exit 1 because shared checklist items 12-15 were unchecked. Their attestations are now recorded: implementation/architecture evidence, green relevant scoped tests with the full-command timeout explicitly excluded, and a review-verdict condition not yet triggered. All checklist items were checked through the CLI; review acceptance is still pending. The developer handoff is retried after these evidence updates.


## Revision 6 (re-apply)

Re-applied the revision-5 delta onto trunk e87d488b8fd892bc86c037ae5e325e596df8e745. The prescribed git diff / git apply --3way command returned exit 1 because content_hash_v2_gaps_test.go conflicted; all other tracked files applied cleanly. The recovery ref 80fc6075 contains 24 tracked paths, while the registered revision-5 board patch contains 29. Restored the five omitted new files verbatim from that registered patch: cmd/curator/muse_test.go, internal/envfragment/muse_test.go, internal/envprofile/muse.go, internal/envprofile/muse_test.go, and internal/testcli/cli_test.go. No source was invented to replace the missing files. No commit, branch switch, rebase, merge or launcher change was made.

Per-file resolution:

- .github/ci/conformance-case-counts.tsv: automatic three-way union. Recomputed all 292 families against their actual corpus: rc.13 92, landed hash-v2 97, Muse 103. Every trunk pin remains exact and there are no duplicate keys. Muse has 16 environments-muse cases and 36 fragment-v3 schema cases. The count script and real command exits are attached. The first all-suite tooling attempt had an undefined root; the second used an older hash-v2 clone whose sibling index has 131 rows. The corrected recomputation uses landed spec b1a2efb6 with 121 sibling rows and passes. No pin was changed to accommodate that obsolete clone.
- .github/ci/conformance-gaps.tsv: automatic union retains trunk's removal of implemented hash-v2 gaps and appends only the immutable Muse-digest ledger from revision 5. Other digest rows remain unchanged.
- internal/envprofile/managed.go: automatic merge preserves trunk's hash-version import, v2 marker assembly/finalization, marker-versus-plan version comparison and credential handling for all non-v1 markers. All Muse changes remain additive, including the four XDG parents, auth observation and optional-marker state. No product decision or writer flip was introduced.
- internal/contextaudit/contextaudit.go: automatic merge preserves trunk's version-aware waivers and detectors; the Muse seed filename extension is retained in the v1 scope list.
- internal/contextlock/schema_conformance_test.go and internal/registry/schema_conformance_test.go: automatic merge preserves all landed v2 production-reader/schema consumers, plus revision 5's exact Muse-digest admission for frozen-schema negative rows.
- internal/conformancecoverage/content_hash_v2_gaps_test.go: the two conflict blocks overlapped trunk's removal of deferred carrier gap expectations and addition of actual v2 vector consumers with the prior static owned-gap assertions. Kept trunk's no-deferred-carriers/index test and five executable hash-v2 vectors unchanged. Retained the older candidate's schema/vector gap inventory under Muse-only named tests selected by its exact digest. Those historical gap tests classify inventory; they do not requalify the old Muse hash-v2 corpus through the newly landed readers, and no such coverage is claimed. Landed hash-v2 vectors remain 5/5 driven and never use the old static-gap classifier. This older-corpus qualification remains a stated bound outside the resolver refresh.
- cmd/curator/env.go and internal/envprofile/status.go are byte-identical to revision 5. The parent-process posture warning, FlagSet help behavior and known-absent optional Muse semantics remain intact. Unknown marker state and unreadable backups continue to make status non-current.

The attached byte-identity table proves 22/29 paths match revision 5 exactly. The remaining seven are the paths above with trunk context or the explicit conflict resolution. Source identity uses sorted path/file-SHA256 records and is recorded in the archive. No conflict markers, LOGBOOK, CHANGELOG, new broker/askpass/crossconformance edits or stray repository files were introduced. The carried test-helper changes remain identical to revision 5. The unrelated askpass EPIPE flake BUG-261001-2772iz stays outside this task.


Fresh verification and exact exits:

Every command below ran directly as a standalone subprocess, without tee or a pipe chain. The archive records commands, real exits, durations, log hashes, conformance roots, source comparison and the mutants themselves. All commands were awaited before attachment and handoff.

- The prescribed full `go test ./cmd/curator ./internal/envprofile ./internal/conformancecoverage -count=1` ran under rc.13 and exited 1 after 611.140 seconds. cmd/curator and envprofile reached Go's default ten-minute package timeout; they were then executing TestEnvStatusRegistryBoundaryPostureAndCheck and TestMuseAuthRepairAdditionalStates/target-directory respectively. Conformancecoverage passed. This is a failing full invocation, not an expected-green result. No full-package green claim is made.
- The focused candidate CLI command exited 0 after 280.690 seconds (275.606 seconds package runtime). It independently reran all TestEnvStatus rows, TestPosture*, all TestGlobalInit*Help rows, TestMuseCLIFragmentAndFakeLaunch, profile credential/secret refusals and fragment-v3 schema cases. Its short initial lock wait preserved the existing TestMain compiler-fixture lock.
- Candidate Muse link-state and other Muse envprofile tests exited 0: 16/16 published rows driven, zero gaps/bounds/skips. Production calls are StatusOf and Resolve in driveMuseLinkState. Additional unreadable, directory, wrong-link, optional-status and isolation rows pass.
- Fragment-v3: 36/36 driven, zero gaps/bounds/skips through the schema oracle and production envfragment.CheckBoundary. They are not 36 independent env resolve invocations. TestMuseCLIFragmentAndFakeLaunch separately drives the real resolver entry and applies its emitted fragment to an inert child, checking exact XDG paths, preserved HOME, argv, seed retention and fork refusal. No real Muse session or real credential fixture was used.
- The four earlier CLI regressions pass: TestEnvStatusMissingAndUnreadableKeepRecord, TestEnvStatusRegistryBoundaryPostureAndCheck, TestEnvStatusReportsShellHookTrustPosture, and TestEnvStatusUnreadableApprovalStateSurfaced. Their existing synthetic provisioning makes the control profiles current, so each measures its approval/posture/registry boundary. The unchanged TestEnvStatusCheckCurrentScopeOnly also passes and independently proves unused optional Muse does not make a profile non-current.
- The remaining two earlier regressions, TestManagerOwnedAbsenceReadsAreGuarded and TestEmptyAllowlistWarningLeavesCurrentStatusCurrent, pass under both roots. The guard measures 475/475 reads covered: 369 through stateread and the same 106 reviewed exceptions. The allowlist warning remains advisory for a known-current control. Optional status negative rows still refuse unknown marker/backups state.
- Nine supporting packages pass under both Muse and rc.13. The landed hash-v2 coverage package passes with 5/5 actual executable vectors, plus the existing schema-reader consumers pass in a separate scoped command. The earlier clone-based vector/schema run is recorded with its own root; the count rerun uses the actual landed b1a2efb6 corpus.
- Build, scoped vet, lint (0 issues), formatting (no output) and whitespace validation each exit 0.

Four compiling mutants were killed, 4/4, each with real exit 1 and an actual assertion failure:

1. HOME emitted in buildFragment: production Resolve rejects the registry-undeclared HOME variable.
2. XDG_DATA_HOME alone omitted: production Resolve rejects its missing declared parent.
3. Replaced-file auth.json fork accepted: production StatusOf loses the required detachment finding and the regression fails.
4. inspectMuseAuth bypasses stateread with raw os.Lstat/absence handling: the real source guard measures 474/475 reads and identifies that production call site. This mutant runs from an isolated physical copy of the current sources; only muse.go differs, so a Go overlay cannot mask the source scan.

Exact command ledger (the per-command JSON also binds its log hash):

| Check | Command | Real exit |
| --- | --- | --- |
| all-counts | `python3 /tmp/TASK-261001-2yvag1-r6-counts-all.py` | 1 |
| all-counts-corrected | `python3 /tmp/TASK-261001-2yvag1-r6-counts-all.py` | 1 |
| all-counts-current-corpora | `python3 /tmp/TASK-261001-2yvag1-r6-counts-all.py` | 0 |
| build | `go build -o /tmp/TASK-261001-2yvag1-r6-curator ./cmd/curator` | 0 |
| candidate-cli-focused | `go test ./cmd/curator -run ^(TestMuse.*\|TestEnvStatus.*\|TestPosture.*\|TestGlobalInit.*Help.*)$ -count=1 -v -timeout=8m` | 0 |
| candidate-coverage | `go test ./internal/conformancecoverage -count=1 -v` | 0 |
| candidate-support | `go test ./internal/envfragment ./internal/envregistry ./internal/stateread ./internal/contextpkg ./internal/contextaudit ./internal/contextlock ./internal/registry ./internal/marker ./internal/testcli -count=1 -timeout=5m` | 0 |
| counts | `python3 /tmp/TASK-261001-2yvag1-r5-counts.py` | 0 |
| formatting | `gofmt -l cmd internal` | 0 |
| guard-and-allowlist | `go test ./internal/envprofile -run TestManagerOwnedAbsenceReadsAreGuarded\|TestEmptyAllowlistWarning -count=1 -v -timeout=5m` | 0 |
| landed-hashv2-coverage | `go test ./internal/conformancecoverage -count=1 -v -timeout=3m` | 0 |
| lint | `golangci-lint run` | 0 |
| muse-link-states | `go test ./internal/envprofile -run ^TestMuse -count=1 -v -timeout=5m` | 0 |
| mutant-HOME | `go test -overlay=/tmp/TASK-261001-2yvag1-r6-evidence/mutants/HOME.json ./internal/envprofile -run ^TestMusePublishedRegistryLayout$ -count=1 -timeout=3m` | 1 |
| mutant-XDG | `go test -overlay=/tmp/TASK-261001-2yvag1-r6-evidence/mutants/XDG.json ./internal/envprofile -run ^TestMusePublishedRegistryLayout$ -count=1 -timeout=3m` | 1 |
| mutant-fork | `go test -overlay=/tmp/TASK-261001-2yvag1-r6-evidence/mutants/fork.json ./internal/envprofile -run ^TestMuseForkRefusedRegression$ -count=1 -timeout=3m` | 1 |
| mutant-seam | `go test ./internal/envprofile -run ^TestManagerOwnedAbsenceReadsAreGuarded$ -count=1 -timeout=3m` | 1 |
| rc13-envprofile-focused | `go test ./internal/envprofile -run ^(TestMuse.*\|TestManagerOwnedAbsenceReadsAreGuarded\|TestEmptyAllowlistWarning.*)$ -count=1 -v -timeout=5m` | 0 |
| rc13-support | `go test ./internal/envfragment ./internal/envregistry ./internal/stateread ./internal/contextpkg ./internal/contextaudit ./internal/contextlock ./internal/registry ./internal/marker ./internal/testcli -count=1 -timeout=5m` | 0 |
| requested-packages | `go test ./cmd/curator ./internal/envprofile ./internal/conformancecoverage -count=1` | 1 |
| trunk-hashv2-coverage | `go test ./internal/conformancecoverage ./internal/contextlock ./internal/registry ./internal/marker -run TestDeferred\|TestContentHashV2VectorsWhenPublished\|TestContextLockV2SchemaCases\|TestAuditRecordV2SchemaCases\|TestRegistryV2SchemaCases\|Test.*HashVersion\|Test.*FrozenRegistry\|Test.*MarkerV5 -count=1 -v -timeout=5m` | 0 |
| vet | `go vet ./cmd/curator ./internal/envprofile ./internal/conformancecoverage ./internal/envfragment ./internal/envregistry ./internal/stateread ./internal/testcli` | 0 |
| whitespace | `git diff HEAD --check` | 0 |

Only the focused and supporting checks named above are claimed green. The full repository hosted matrix is not rerun here; revision-4 hosted run 36833591988 and earlier mutant/launcher artifacts remain historical evidence, not proof for revision 6. curator run muse is outside this resolver handoff; the separately owned launcher mapping/v3-reader leaf ee66c107 is reported landed in the binding instructions but is not exercised here. Its integration row remains an explicit scope bound. The immutable Muse candidate's historical hash-v2 gap inventory was retained, not asserted as new reader coverage. Native context loading, skills discovery and credential refresh retain their existing unverified bounds.

Entry text (kept here instead of LOGBOOK/CHANGELOG): reapply the Muse resolver onto e87d488b; recover five snapshot-omitted source files from the registered revision-5 patch; preserve versioned hash and marker behavior from trunk; separate the landed hash-v2 vector consumers from the immutable Muse corpus inventory; recompute all digest-keyed pins from the matching spec roots; attach green focused regressions, four genuinely killed mutants and honest full-suite timeout evidence for review.

Source identity: `9d8477b158b20b45591004a1367075ad08a49ef9b6c7955d1c4536158dfaada3` (29 paths); base `e87d488b8fd892bc86c037ae5e325e596df8e745`. All work stays uncommitted and ready for review.


## Revision 7

Answers the full revision-6 verdict: R1, R2 and R3. Work remains the resolver-only leaf. No launcher repository was changed. No commit or branch operation was performed.

R1 (auth parent boundary): inspectMuseAuth now uses the existing E5 managed parent check before reading auth metadata or accepting a live credential link, including the current-home fast path. The final credential file-link is intentionally retained. The check permits absent provisioning parents without writing them, rejects every existing symlink below the managed root, and reports an unreadable passthrough on refusal. No allowlist entry or credential-byte read was added. TestMuseAuthParentBoundaryAtResolve drives production Resolve in 8/8 rows: config and config/muse parents, outside and inside symlink targets, bare resolve and repair. Its metadata spy proves auth inspection is not reached before refusal. Native bytes, parent sentinel bytes and the final link are preserved. TestMuseCLIAuthParentBoundary drives the real CLI in 2/2 outside-config rows: both return exit 1, stdout is empty, and native/outside bytes and auth link are preserved. These are test-command passes verifying intentionally non-zero CLI refusal exits.

R2 (production fragment validation): TestMuseCLIFragmentAndFakeLaunch now validates the actual emitted CLI document and a direct production Resolve document against the pinned d373078a v3 schema. The assertion is permanent and runs under rc.13, even though rc.13 publishes no v3 corpus. The schema fixture is copied byte-for-byte with SHA256 1e05c7f86d341873d4167348297f2339567c45a79e40572544d850c34e370449; its five shared schema dependencies were verified byte-identical to the existing fixtures. TestMuseFragmentV3PublishedCases binds all four producer-reachable Muse policy rows to actual Resolve and CLI emission in bare/repair modes, validates the unmodified documents, and compares the emitted permissions with the row. Production coverage is 4/36 corpus rows (16 emitted-document assertions); the other 32/36 rows are explicitly logged as static schema/reader oracles, including malformed wire documents and other adapters' v3 documents. All 36/36 oracle rows are driven, zero skips. This supersedes any earlier implication that all 36 static documents were production emissions. The other adapters continue to emit their existing fragment revision. On Windows the pinned schema admits only POSIX paths: the tests explicitly report that bound and project emitted env paths to POSIX form for structural schema validation, without synthesizing any members. Raw Windows path schema conformance is not claimed; Windows execution was not run locally.

R3 (scope): internal/testcli/cli.go is byte-identical to base e87d488b; internal/testcli/cli_test.go is absent, as at that base. No broker/askpass/crossconformance fix remains in this candidate. The EPIPE refusal-path flake belongs to BUG-261001-2772iz and remains out of scope. It was not observed in this revision's scoped checks. This corrects the older results' mistaken characterization of those files as harmless carryover.

The rev6 comparison artifact covers all 29 reviewed delta paths: 24/29 are byte-identical. The five differences are the three Muse source/test files above and the two broker reversions. The only added path is the pinned v3 schema fixture. The attached source identity covers all 28 candidate paths. Exact Muse manifest counts were independently measured and matched the unchanged digest-keyed pins: environments-muse/cases 16 and launch-env-fragment-v3/schema-cases 36 under bd03456b92a7368d90ea74fe6953db10bc188588020683024a6a6b8735a40783. SPEC_PIN remains rc.13.

Fresh validation (all commands are direct processes, no tee; exact command strings and real exits are in r7-exits.json):

| Check | Real exit | Evidence |
| --- | --- | --- |
| Requested two-package Muse/status/guard/empty-allowlist command, candidate root, count=1 | 0 | envprofile 90.353s; CLI 219.468s; all matching tests pass |
| New R1 production Resolve rows alone, count=1 | 0 | 8/8 pass, 15.286s |
| rc.13 core envfragment/envregistry/stateread/conformancecoverage packages, count=1 | 0 | All four pass |
| rc.13 production CLI fragment/fake-child and parent-boundary tests, count=1 | 0 | Actual emission schema assertions and both CLI refusal rows pass, 6.565s |
| Build after changes | 0 | go build -o /tmp/TASK-261001-2yvag1-r7-curator ./cmd/curator |
| Vet after changes | 0 | Three affected production packages |
| Lint | 0 | 0 issues |
| gofmt -l and git diff --check, separately | 0 each | No output |

Four valid compiling mutants killed, 4/4; each test command actually FAILS with exit 1, not a passing gate:

- R1 boundary check removed: all 8/8 parent rows fail because production Resolve accepts the live auth link and emits a document (16.882s).
- R1 narrowing mutant checks only the immediate auth parent, following its ancestors: all 4/4 config-link rows fail while the four terminal config/muse-link rows still pass (13.043s). This proves ancestor traversal is covered, rather than merely the existence of a check.
- R2 permissions removed only from Muse's Fragment.Object: the permanent production CLI schema assertion and all four producer policy rows fail with the required-member schema error (12.549s). The complete CLI TestMuse scope was run. This is the previously surviving mutant, now killed.
- R2 narrowing mutant removes permissions only for locked Muse fragments: the published locked/native production row fails, while the other policy rows pass (3.512s).

Mutants ran through Go source overlays; the archive includes replacement sources and overlay mappings. No mutant was left in the worktree. Earlier HOME/fork/XDG_DATA_HOME mutant evidence is historical attached evidence, not rerun in this revision.

All six historical failures were freshly rerun and pass: the four CLI status tests retain their previously reviewed synthetic registered-environment provisioning, so each measures its intended approval/registry/hook posture; TestManagerOwnedAbsenceReadsAreGuarded passes at 475/475 reads (369 seam, the unchanged 106 reviewed exceptions); TestEmptyAllowlistWarningLeavesCurrentStatusCurrent keeps the warning advisory. TestMuseStatusOptionalProvisioning and TestEnvStatusCheckCurrentScopeOnly also pass, independently proving that a never-enabled Muse home leaves the healthy profile current and unreadable state is not treated as absence. No status semantic or fixture change was made in revision 7.

Limits: the full repository suite, race/Windows matrix, hosted gate and launcher integration were not rerun by this developer; the runner owns the next gate. Historical rev6 hosted green evidence is not presented as validation for the changed rev7 tree. No real Muse session or real credentials were used. The separate curator-run mapping/v3 reader already landed in the launcher leaf; that integration remains outside this resolver handoff. Native discovery, personal-context isolation and refresh coordination retain their existing stated bounds.

Entry text retained here instead of LOGBOOK/CHANGELOG: reject symlinked Muse auth parents before accepting liveness; bind production v3 emission to the pinned schema and four reachable corpus policies; kill both full and narrowing regressions; remove the unrelated EPIPE delta. Revision 7 is ready for review with scoped green checks and honest coverage classification.
