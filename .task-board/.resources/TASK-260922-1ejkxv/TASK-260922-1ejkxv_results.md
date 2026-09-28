# TASK-260922-1ejkxv Results — fleet-isolated-policy

## Changes

- `system-config-v2` now permits exactly `shared` and `isolated` at the existing `environments.isolation` key. The key remains lockable; manager §1 whole-map replacement and precedence semantics remain unchanged. No new mode knob was added.
- Environments §12.2 states that a system-map entry forces its declared direction: `shared` forces shared credential-store mode and `isolated` forces isolated mode. Silence resolves to that locked direction. An explicit profile request for `shared` under an isolated lock is refused with `environment_isolation_lock_conflict`.
- An isolated lock that encounters an already-provisioned shared passthrough refuses use with `environment_credential_conflict`, leaves the link and credential bytes untouched, and points to the explicit inspect → plan → apply migration in manager §12.4 / F-C2 (`TASK-260922-1t2w1q — envprofile-explicit-credential-migration`).
- Added schema cases for valid `shared`, valid `isolated`, unknown direction `automatic`, and both directions at one key. Preserved the existing invalid-value case. Updated generator and validator guards/tests, regenerated the conformance manifest/index and rc.9 candidate pin, and added the unreleased CHANGELOG entry. No Curator implementation code or frozen v1 schema changed.

## Curator follow-up

Named manager-side Curator follow-up leaf: **Enforce the isolated environment lock in the environment manager**. It consumes the exact system key `environments.isolation`, refuses an explicit `shared` profile request with `environment_isolation_lock_conflict`, and handles an existing shared passthrough with `environment_credential_conflict` and the F-C2 migration path. Curator code is outside this task’s scope.

## Bounded validation gate

The Makefile `validate:` recipe is exactly:

```make
validate:
	python3 tools/validate.py
	python3 -B -m unittest discover -s tools -p 'test_*.py'
	go test ./tools/...
```

Following the continuation instructions, I ran those three recipe lines in order as bounded standalone processes. The unittest line was split by every `tools/test_*.py` file; `test_validate.py` exceeded one call and was rerun once across all 29 test classes. The whole-file attempt exited 130 and is recorded as interrupted, not passing evidence. No prior attached test evidence was used for these gate results.

The task-specific interpreter directory `.temp/task-260922-1ejkxv-venv` was absent from the worktree. The first validator attempt using that requested PATH exited 1 because the resolved Python lacked `jsonschema`. I reran with the existing repository virtualenv at `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv`; all successful Python commands below used that PATH.

### Regeneration and validator

| Command | Exit | Tail/result |
|---|---:|---|
| `make regenerate` | 0 | `go run ./tools/generate-vectors -root .` |
| `PATH=.temp/task-260922-1ejkxv-venv/bin:$PATH python3 tools/validate.py` | 1 | `ModuleNotFoundError: No module named 'jsonschema'` (the task-specific venv path was absent) |
| `PATH=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH python3 tools/validate.py` | 0 | `validated 64 schemas and 1169 vector files` |

### Python unit tests

The following commands were separate shell calls with `PATH=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH`.

| Command | Exit | Tail/result |
|---|---:|---|
| `python3 -B -m unittest discover -s tools -p 'test_implementation_coverage.py'` | 0 | Ran 36 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_release_gate.py'` | 0 | Ran 32 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py'` | 130 | `KeyboardInterrupt` in `ReadFailureVectorTests.test_substituted_scenario_rejected_through_main`; rerun by class below |
| `python3 -B -m unittest discover -s tools -p 'test_verify_release_commit.py'` | 0 | Ran 5 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_verify_release_merge_policy.py'` | 0 | Ran 5 tests; OK |

Every class-level rerun for `test_validate.py`:

| Command | Exit | Tail/result |
|---|---:|---|
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k SchemaRegistryCacheTests` | 0 | Ran 1 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k WireSemanticValidationTests` | 0 | Ran 18 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k AssuranceRelationalValidationTests` | 0 | Ran 3 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k RepositoryDescriptorIdentityTests` | 0 | Ran 6 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k ManagerLifecycleValidationTests` | 0 | Ran 4 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k BuildDriverGoldenSuiteTests` | 0 | Ran 13 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k SharedFixtureMarkerTests` | 0 | Ran 10 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k WorkflowRegenerationScopeTests` | 0 | Ran 2 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k EnvironmentVectorTests` | 0 | Ran 27 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k EnvPassthroughVectorTests` | 0 | Ran 27 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k StoreBoundaryVectorTests` | 0 | Ran 25 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k PathKindAdmissionVectorTests` | 0 | Ran 36 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k SourceSignersVectorTests` | 0 | Ran 47 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k CodexSeedVectorTests` | 0 | Ran 32 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k ContextVersionVectorTests` | 0 | Ran 15 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k ContextDetectorVectorTests` | 0 | Ran 9 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k SnapshotAcquisitionVectorTests` | 0 | Ran 5 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k ShellHookTrustVectorTests` | 0 | Ran 8 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k SecurityPostureVectorTests` | 0 | Ran 24 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k WriteNofollowVectorTests` | 0 | Ran 12 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k DotfileManagersVectorTests` | 0 | Ran 20 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k ReadFailureVectorTests` | 0 | Ran 18 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k UmbrellaProviderVectorTests` | 0 | Ran 18 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k ManagerConfigVectorTests` | 0 | Ran 27 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k SystemConfigV2SchemaTests` | 0 | Ran 27 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k RegistryPageBoundaryVectorTests` | 0 | Ran 13 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k RegistryCheckpointVectorTests` | 0 | Ran 17 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k RegistryBootstrapVectorTests` | 0 | Ran 39 tests; OK |
| `python3 -B -m unittest discover -s tools -p 'test_validate.py' -k TakeoverClosedSetTextTests` | 0 | Ran 21 tests; OK |

The class-level runs covered 524 tests; all returned exit 0 and ended in `OK`. Across the five Python test files, 602 tests passed.

### Go, regeneration check, formatting

| Command | Exit | Tail/result |
|---|---:|---|
| `go test ./tools/...` | 0 | `ok github.com/relux-works/curator-spec/tools/generate-vectors 1.731s` |
| `make regenerate-check` | 0 | generator ran; `git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json ... release/1.0.0-rc.9.json` returned clean |
| `gofmt -d tools/generate-vectors/main_test.go tools/generate-vectors/system_config.go` | 0 | no formatting diff |
| `git diff --check --cached` | 0 | no whitespace errors |

The three `validate:` recipe lines and `make regenerate-check` are green under the accepted ordered composition. `make validate` was not run as one aggregate shell process.