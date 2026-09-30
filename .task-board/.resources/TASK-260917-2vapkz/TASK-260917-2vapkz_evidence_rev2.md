# TASK-260917-2vapkz evidence — revision 2

## Rework applied

- Added `log-response-v3` referencing `registry-log-entry-v2` and `manager-config-v3` with `hash_version: 2` for v2 state-hash waivers. Restored `log-response-v2.schema.json` and `manager-config-v2.schema.json` byte-for-byte to `v1.0.0-rc.13`; the new schema cases are generated under v3, and the generator rebuilds `schema-cases` so obsolete locations are removed.
- Added the released-schema byte immutability guard to `tools/validate.py` and called it through `validate.main()`. It inventories schemas at the latest protocol release tag and compares the checked-in bytes against the tagged blobs. Added positive and negative regression tests.
- Updated schema references and normative text, documented all added schema versions in `CHANGELOG.md`, and retained generated conformance vectors. No `LOGBOOK.md` was created.
- `release/1.0.0-rc.13.json` changes because the vector generator writes its manifest pin (`tools/generate-vectors/main.go:212,2222`). It was not manually edited; `make regenerate-check` regenerates it and confirms it is stable.

## Verification and real exit codes

- `./.venv/bin/python -B tools/validate.py` — **0**, validated 72 schemas and 1,228 vector files.
- `make regenerate-check` — **0**; generator ran and the generated-file diff check was clean.
- `go test ./tools/...` — **0** after correcting the new schema test to compare `json.Number`. The first run was **1** because the test incorrectly expected `float64`; that assertion was fixed and the command rerun green.
- `make validate` — attempted twice; each invocation was interrupted at the single-call time bound while the full `test_validate.py` scenario matrix was still running, so each real exit code was **130**. These are not reported as passing aggregate runs. The constituents were then executed in bounded groups:
  - `test_validate.py`: the 548 non-matrix tests passed (**0**); the full ReadFailure, WriteNofollow, and DotfileManagers matrices passed in chunks, covering 39/39, 11/11, and 22/22 unique cases respectively. The last 1-case DotfileManagers slice failed in the chunk harness before executing a case (`IndexError` because that test requires two donor names; exit **1**); rerunning the overlapping 20:22 slice passed (**0**) and covered the remaining cases.
  - `ReleasedSchemaImmutabilityTests` — **0**, 2 tests passed, including unchanged-release positive and byte-drift negative through `validate.main()`.
  - `test_verify_release_merge_policy.py` — **0**, 5 tests; `test_verify_release_commit.py` — **0**, 5; `test_release_gate.py` — **0**, 35; `test_implementation_coverage.py` — **0**, 39; `test_skillfile_sources_independence.py` — **0**, 8; `test_allowed_signers.py` — **0**, 7.
- Narrowing mutant: changed the guard to inspect only `manager-config-v2.schema.json`, then ran the named byte-drift regression test against a modified `log-response-v2.schema.json`. The test failed as required (real exit **1**; `validate.main()` incorrectly returned 0 under the mutant), demonstrating that the guard covers the released-schema class rather than one selected file. Restored the production guard; the two immutability tests then passed again.
- `git diff --exit-code v1.0.0-rc.13 -- schemas/v1/log-response-v2.schema.json schemas/v1/manager-config-v2.schema.json` — **0**. `git diff --check` and `git diff --cached --check` — **0**. No `LOGBOOK.md` exists.

Two Python test launches before the local development environment was ready failed at setup (import path, then missing `jsonschema`; each exit **1**). Installed `requirements-dev.txt` into the worktree `.venv` and reran the tests listed above with that interpreter.

The worktree is staged and uncommitted for the Story handoff.
