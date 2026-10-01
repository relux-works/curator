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
