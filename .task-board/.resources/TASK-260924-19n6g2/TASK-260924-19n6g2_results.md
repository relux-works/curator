# TASK-260924-19n6g2 — curator-spec v1.0.0-rc.13 release prep

## Changes

- Advanced the generated core suite and release references to `1.0.0-rc.13`; added `release/1.0.0-rc.13.json` with the exact core manifest pin.
- Restored `release/1.0.0-rc.9.json` to the bytes from tag `v1.0.0-rc.9` (SHA-256 `a0c58ff5e44bc93c013e4c7526573c2fc7e2467a67455b6898647db6a0879f82`). Rc.13 metadata preserves rc.9 as historical evidence; claim schema v5 remains pinned to protocol rc.9.
- Added a generated, independent `conformance/skillfile-sources-v1/manifest.json` pin covering the accepted source and repository-transport documents, schemas, and vectors. Rc.13 metadata ties that suite separately to core tag `v1.0.0-rc.10` and its manifest SHA-256 `803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403`.
- Updated release and validation gates, their tests, CI regeneration scopes, README, CHANGELOG, and COMPATIBILITY. The changelog keeps the existing entries and highlights the accepted source suite, no core manifest schema-v9 `directory` field, PR #88 hard-link erratum, and per-implementation claim checks.

## Validation

- Initial `make validate`: exit 2 because the system Python did not have `jsonschema`. Installed `requirements-dev.txt` into a temporary virtualenv and reran the command.
- `make validate` with that virtualenv: exit 0. Validator reported 64 schemas and 1,169 core vector files; unittest discovery ran 619 tests in 456.770 seconds; `go test ./tools/...` passed.
- `make regenerate-check`: exit 0 in a disposable clean checkout of the candidate snapshot (temporary local commit `56cd96e027ec10aead9a56db2a8e20f0c03c9731`); regeneration produced no diff. The Story worktree remains uncommitted.
- `python3 tools/release_gate.py --version 1.0.0-rc.13 --commit HEAD`: exit 0 in that clean candidate checkout.
- `python3 tools/verify_release_commit.py --commit origin/main`: exit 0 for base main commit `2c39c428508e69a623cf3c26bac5e90cf7f5bf16` (`trusted maintainer signature`). The release-prep merge commit still needs the hosted provenance check after PR merge.
- `gofmt -l tools`: exit 0 with no files listed. `git diff --check`: exit 0. Python byte-compilation: exit 0.

## Orchestrator handoff

After the release-prep PR merges and `origin/main` is confirmed at its merge commit, create and push the signed tag with:

```sh
git tag -s -m "Curator Protocol v1.0.0-rc.13" v1.0.0-rc.13 origin/main
git push origin refs/tags/v1.0.0-rc.13
```

Then watch the **Release specification** workflow (`.github/workflows/release.yml`), job **Publish signed specification artifacts**, triggered by the `v1.0.0-rc.13` tag. The main-branch **Release target provenance** check also verifies the merged commit before tagging.

No PR, signed tag, or release was created in this developer run. Findings are recorded in this task-scoped outcome resource per the campaign rule; `LOGBOOK.md` was not edited.
