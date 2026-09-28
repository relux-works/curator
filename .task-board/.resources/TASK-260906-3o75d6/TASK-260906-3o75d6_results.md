# TASK-260906-3o75d6 results — cli-takeover-clause-editorial

## Before and after

Before, six `cli/curator.md` command rows repeated takeover behavior: `profile install`, both `profile use` forms, `profile update`, `profile sync`, and `env resolve`. The repeated text named each operation's write set and unmanaged-conflict failure. The takeover example appeared after the `env unmanage` example.

After, those six rows retain their existing `--takeover` flag and point to one note immediately below the Commands table. The note copies the complete two-sentence environments §9.5 clause: takeover is not an operation or scope of its own, covers only unmanaged files written by the carrying operation, and is accepted on exactly the five named operations and no others. The example now sits with the `profile use` examples. No protocol or normative text changed. No new behavior is stated for `env resolve --takeover` without `--repair`, and import/global rows remain flag-free.

`CHANGELOG.md` has an unreleased editorial entry.

## Validator pins

- `TAKEOVER_CARRYING_OPERATIONS` is unchanged. The validator still parses §9.5 and compares the exact five-operation enumeration in order, and still requires the onboarding-trigger list to match it.
- The exact §9.5 exclusion sentence, §9.4 and §9.6 mirror sentences, manager §12.3 sentence, and their section-heading checks remain unchanged.
- The existing CLI pin still requires `profile import` and every `global` row to omit `--takeover`.
- The CLI pin now extracts the complete two-sentence clause from environments §9.5 and requires that exact clause once, as the note immediately below the Commands table. It also requires exactly six flagged rows (one install, two use, one update, one sync, one resolve row scoped to `--repair`); each must point to the note and may not repeat takeover wording. These checks keep the CLI carrier mapping tied to the unchanged §9.5 enumeration.
- New negative tests reject an incomplete note, a missing carrier flag, an `env resolve` row without `--repair`, and a repeated row-level clause. Existing negative tests for widened/dropped/reordered §9.5 carriers and import/global flags remain in place.

## Verification

- Focused validator tests: `uv run --no-project --with-requirements ../requirements-dev.txt -- python -B -m unittest test_validate.TakeoverClosedSetTextTests` (working directory `tools`) — **exit 0**, 26 tests.
- Full handoff validation: `uv run --no-project --with-requirements requirements-dev.txt -- make validate` — **exit 0**. It validated 64 schemas and 1,166 vector files; all 606 Python tests passed in 719.531 seconds; `go test ./tools/...` passed (`tools/generate-vectors`). This was the only `make validate` run.
- Final `git diff --check` — **exit 0**.
- The repository has no configured lint target or linter in its Makefile, CI workflow, README, or development requirements. `git diff --check` passed; the full validation suite passed.

## Corrected test invocations and intermediate results

These nonzero results are retained for evidence honesty; each issue was corrected before the green reruns above.

- `python3 -B -m unittest tools.test_validate.TakeoverClosedSetTextTests` — **exit 1**; direct module loading did not put `tools/` on `sys.path`, so `assurance` was not found.
- `PYTHONPATH=tools python3 -B -m unittest test_validate.TakeoverClosedSetTextTests` — **exit 1**; the ambient interpreter lacked `jsonschema`.
- `uv run --no-project --with-requirements requirements-dev.txt -- python -B -m unittest test_validate.TakeoverClosedSetTextTests` from the repository root — **exit 1**; `test_validate` was not importable from that working directory.
- The first dependency-correct focused run from `tools` — **exit 1**; 26 tests exposed pointer punctuation and test-mutant issues, which were corrected.
- The next focused run from `tools` — **exit 1**; one negative test showed the predicate also needed to reject “takes over”. The predicate was tightened and the focused suite then passed 26/26.
## Revision 2 — restore the write-coverage clause

### Before and after

Revision 1’s note stated the five-operation closed set and its closure but stopped before the §9.5 write-coverage rule. It omitted the onboarding notice/backup behavior and the `environment_surface_unmanaged_conflict` refusal when a carrying operation lacks `--takeover` outside onboarding.

Revision 2 extends the single note below the Commands table with the exact §9.5 sentence covering the notice, backup, and no-flag refusal, and states that `env resolve`’s `--takeover` applies only with `--repair`. All six command rows retain their existing flag and end with the note pointer. The takeover example remains grouped with the `profile use` examples. `env resolve --takeover` without `--repair` remains unstated as an operation. No protocol file or normative rule changed.

### Validator pins

- The fixed `TAKEOVER_CARRYING_OPERATIONS` sequence is unchanged. Validation still extracts the §9.5 carrier list, requires exactly the five members in order, and requires the onboarding-trigger list to match it.
- The exact §9.4, §9.5, §9.6, and manager §12.3 exclusion sentences and their section placement checks remain unchanged.
- The existing import/global row guard remains in `validate_cli_takeover_rows`: every `curator profile import` row and every `curator global` row must omit `--takeover`. `test_import_row_gaining_takeover_fails` and `test_global_row_gaining_takeover_fails` still exercise those refusals.
- The adapted CLI pin now extracts the exact §9.5 carrier/write-coverage text through the unmanaged-conflict failure sentence and requires that full text once in the immediate post-table note, followed by the explicit `env resolve`/`--repair` scope. It still requires the six row carriers, optional flags, `--repair` on the resolve row, note pointers, and no repeated takeover wording. This adds the missing coverage to the pin; it does not relax the prior fixed carrier or import/global checks.
- New negative tests reject a note with its write-coverage/refusal sentence removed, a note with a carrier removed, and a note missing the resolve/repair scope. The requested narrowing mutations are `test_cli_note_drop_of_carrier_fails` and `test_import_row_gaining_takeover_fails`; both were exercised in the focused class run and rejected by the validator.

### Revision 2 validation

All commands were run directly from zsh, without pipelines. The full Makefile test set took longer than the approximately ten-minute single-shell limit in the attached revision 1 run, so its three stages were rerun as bounded standalone commands and the Python suite was split by class/file. The revision 1 `make validate` exit 0 applies only to revision 1; no monolithic revision 2 `make validate` exit code is claimed.

- `PYTHONPATH=tools uv run --with-requirements requirements-dev.txt python3 -B -m unittest test_validate.TakeoverClosedSetTextTests` — **exit 0**, 29 tests. This includes the carrier, import/global, write-coverage, and repair-scope negative tests above.
- `uv run --no-project --with-requirements requirements-dev.txt -- python3 tools/validate.py` — **exit 0**; validated 64 schemas and 1,166 vector files.
- `go test ./tools/...` — **exit 0** (`tools/generate-vectors`).
- `git diff --check` — **exit 0**.

The remaining 502 tests in `test_validate.py` were run with the common command prefix `PYTHONPATH=tools uv run --no-project --with-requirements requirements-dev.txt -- python3 -B -m unittest`. Each invocation exited 0:

| Test classes | Tests | Exit |
|---|---:|---:|
| `AssuranceRelationalValidationTests BuildDriverGoldenSuiteTests CodexSeedVectorTests` | 48 | 0 |
| `ContextDetectorVectorTests ContextVersionVectorTests DotfileManagersVectorTests` | 44 | 0 |
| `EnvPassthroughVectorTests` | 27 | 0 |
| `EnvironmentVectorTests` | 27 | 0 |
| `ManagerConfigVectorTests` | 27 | 0 |
| `ManagerLifecycleValidationTests ReadFailureVectorTests` | 22 | 0 |
| `PathKindAdmissionVectorTests` | 36 | 0 |
| `RegistryBootstrapVectorTests` | 39 | 0 |
| `RegistryCheckpointVectorTests RegistryPageBoundaryVectorTests` | 30 | 0 |
| `RepositoryDescriptorIdentityTests SchemaRegistryCacheTests SecurityPostureVectorTests SharedFixtureMarkerTests` | 41 | 0 |
| `ShellHookTrustVectorTests SnapshotAcquisitionVectorTests SourceSignersVectorTests` | 60 | 0 |
| `StoreBoundaryVectorTests SystemConfigV2SchemaTests UmbrellaProviderVectorTests` | 69 | 0 |
| `WireSemanticValidationTests WorkflowRegenerationScopeTests WriteNofollowVectorTests` | 32 | 0 |

The other four test files were run separately with the common prefix `uv run --no-project --with-requirements requirements-dev.txt -- python3 -B -m unittest discover -s tools -p`. Each invocation exited 0:

| Test file | Tests | Exit |
|---|---:|---:|
| `test_release_gate.py` | 32 | 0 |
| `test_implementation_coverage.py` | 36 | 0 |
| `test_verify_release_commit.py` | 5 | 0 |
| `test_verify_release_merge_policy.py` | 5 | 0 |

Together with the 29-case takeover class run above, those bounded shards ran all 609 Python tests currently discovered by the Makefile (531 in `test_validate.py`, 78 in the other four files).

The first revision 2 focused invocation, `uv run --with-requirements requirements-dev.txt python3 -B -m unittest tools.test_validate.TakeoverClosedSetTextTests`, exited 1 because direct module loading did not put `tools/` on `sys.path`; the corrected `PYTHONPATH=tools` invocation above passed. The ambient `python3` also lacks `jsonschema` (the import probe exited 1), so Python validation used the development requirements through `uv`.

The revision 1 results above record its full `make validate` run as exit 0 in 719.531 seconds. Revision 2’s equivalent Makefile stages are green in bounded invocations; its monolithic `make validate` was not rerun because that measured duration exceeds this run’s single-shell time limit.

## Revision 3 — explicit environments reference

### Before and after

Before, the CLI note copied §9.5's source-relative phrase “named above as onboarding triggers.” In the CLI guide that phrase pointed at no preceding trigger list. It now says “named in environments section 9.5 as onboarding triggers.” The five carriers, closure, write coverage, notice and backup, no-flag refusal, and `env resolve`/`--repair` scope are otherwise unchanged. The stray blank line before the profile-use examples' closing fence is removed. No normative file or rule changed.

### Validator pin and regression mutant

`takeover_cli_clause` still extracts the §9.5 clause through the notice/backup and `environment_surface_unmanaged_conflict` refusal sentence. It now performs one exact, unique source-pointer substitution from “named above as onboarding triggers” to “named in environments section 9.5 as onboarding triggers.” `validate_cli_takeover_rows` still compares the normalized note for exact equality with that complete clause plus the `env resolve` repair-scope sentence, and still requires the clause exactly once. This changes the document reference while preserving exact equality over every other word and predicate.

The §9.5 five-carrier enumeration and onboarding-trigger equality pins, §9.4/§9.6/manager exclusion pins, carrier-row counts and optional flags, pointer-only row check, `env resolve --repair` row constraint, and import/global no-`--takeover` guards are unchanged. `test_cli_note_dangling_reference_mutant_fails` is the named regression test: it narrows the note by mutating the new explicit reference back to the dangling “named above” phrase and verifies that validation refuses it. The 30-test `TakeoverClosedSetTextTests` run also covers the carrier-drop, import-row gaining `--takeover`, and global-row gaining `--takeover` negative cases.

### Revision 3 verification

Commands were run directly as standalone zsh processes, without pipes:

- `python tools/validate.py` — **exit 127** because this host has no `python` executable (`zsh: command not found: python`).
- `uv run --no-project --with-requirements requirements-dev.txt -- python3 tools/validate.py` — **exit 0**; validated 64 schemas and 1,166 vector files.
- `PYTHONPATH=tools uv run --no-project --with-requirements requirements-dev.txt -- python3 -B -m unittest test_validate.TakeoverClosedSetTextTests` — first attempt **exit 1** because line wrapping split literals used by existing source-mutation fixtures; the note wrapping was adjusted, preserving those fixture targets. The rerun passed **30 tests, exit 0**.
- `make regenerate-check` — **exit 0**.
- `git diff --check` — **exit 0**.
- `make validate` — not run by the producer. The campaign handoff contract assigns the configured landing suite a single handoff run and forbids an additional manual full-suite run; its exit code is therefore owned by that handoff run and is not claimed here.
