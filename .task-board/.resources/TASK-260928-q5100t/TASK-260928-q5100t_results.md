# TASK-260928-q5100t results

## Changes

- `protocol/environments.md` §9.4 now requires global add/install to publish the extended profile lock and referenced immutable store entries before in-place materialization. On `environment_surface_unmanaged_conflict`, it retains that lock, restores surfaces already written, leaves the conflicting unmanaged surface and other mutable state unchanged, and reports the surface. The existing `profile sync --takeover` / `profile use --takeover`, then retry recovery remains applicable.
- `profiles/manager.md` §§2.5, 2.6, and 12.3 specify publication order, the lock-retaining journal disposition, and crash recovery for that disposition. The five-operation takeover carrier set is unchanged.
- Added `conformance/v1/vectors/environments-global-lock-publication.json` for ordering and conflict cases for both global operations and `profile sync --takeover` recovery, with validator registration, negative tests, README wiring, manifest entry, and matching RC.13 candidate pin.
- Added the Unreleased changelog entry. No `LOGBOOK.md` was changed, per the brief.

## Validation run by this worker

All 641 Python unit tests passed in 18 bounded invocations. This covers all 542 `tools/test_validate.py` cases, including the 7 new global-lock vector tests and all 30 takeover closed-set text tests; all 99 cases in the other Python test modules also passed. Every listed invocation exited 0.

`tools/test_validate.py` shards:

| Test classes | Tests | Exit |
|---|---:|---:|
| SchemaRegistryCache, WireSemanticValidation, AssuranceRelationalValidation, RepositoryDescriptorIdentity, ManagerLifecycleValidation, BuildDriverGoldenSuite, SharedFixtureMarker, SkillfileSourcesSuiteManifest, WorkflowRegenerationScope | 59 | 0 |
| EnvironmentVector, EnvPassthroughVector, StoreBoundaryVector | 79 | 0 |
| SourceSignersVector, ContextVersionVector, ContextDetectorVector | 71 | 0 |
| CodexSeedVector, SnapshotAcquisitionVector, ShellHookTrustVector, SecurityPostureVector | 69 | 0 |
| WriteNofollowVector, DotfileManagersVector, GlobalLockPublicationVector, UmbrellaProviderVector | 57 | 0 |
| ManagerConfigVector, SystemConfigV2Schema, RegistryPageBoundaryVector | 67 | 0 |
| PathKindAdmissionVector, RegistryCheckpointVector | 53 | 0 |
| RegistryBootstrapVector, TakeoverClosedSetText | 69 | 0 |
| ReadFailureVector except `test_substituted_scenario_rejected_through_main` | 17 | 0 |
| `ReadFailureVectorTests.test_substituted_scenario_rejected_through_main` alone, retaining all 39 cases | 1 | 0 |

Other Python tests:

| Test selection | Tests | Exit |
|---|---:|---:|
| Allowed signers, implementation coverage, skillfile-source independence, release-commit verification, release-merge policy | 64 | 0 |
| Stable release gate | 6 | 0 |
| Protocol RC.13 release gate, six bounded groups (5, 5, 5, 5, 5, 4) | 29 | 0 |

Other checks:

| Command | Result |
|---|---|
| `make regenerate` | exit 0 |
| `PATH=.venv/bin:$PATH python3 tools/validate.py` (final run after tests) | exit 0; validated 64 schemas and 1,170 vector files |
| `go test ./tools/...` | exit 0 |
| `git diff --check` (final run) | exit 0 |

I did not run `make validate` as one monolithic process. Its validator, all 641 unit tests, and Go test stages were run directly and separately; the unit tests were sharded to stay within the headless per-command limit. The prior attached run's `make validate` exited 130 at the ten-minute limit, so that result is not treated as a pass.

## Failed or interrupted attempts and recovery

- The prior attached Change Request gate exited 1 because `test_section_95_sentence_moved_to_section_94_fails` used an anchor no longer adjacent after the new §9.4 paragraph. Updated the test to move the pinned §9.5 sentence after the exact §9.4 sentence; all 30 `TakeoverClosedSetTextTests` then passed in the 69-test shard.
- Two initial focused unittest commands from `tools/` exited 1 because the relative `.venv/bin` path pointed at `tools/.venv` and the fallback interpreter lacked `jsonschema`. The focused takeover and global-lock classes were rerun with the worktree interpreter and passed (30 and 7 tests); both classes also passed in the complete shards.
- An initial root unittest shard without `PYTHONPATH=tools` exited 1 on `ModuleNotFoundError: assurance`. The bounded shards were rerun with `PYTHONPATH=tools`; all passed.
- One combined vector shard was interrupted at 7:35 (exit 130) because the read-failure test repeatedly runs the full validator. Its other classes passed in a 57-test shard, its remaining 17 tests passed separately, and its 39-case main-entry test passed alone (197.893 seconds, exit 0). The temporary vector/manifest/release edits were restored; the validator subsequently exited 0.
- A temporary attempt to batch the 39 read-failure cases changed the validator's own closed-case set and exited 1 on six inventory checks. That harness was discarded; the unmodified test with the complete case set passed.

No validation result is accepted solely from an earlier attached artifact; the passing results above were rerun by this worker.
