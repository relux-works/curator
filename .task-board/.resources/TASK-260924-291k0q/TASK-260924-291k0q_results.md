# TASK-260924-291k0q results

## Changes

- Added `tools/verify_skillfile_sources_independence.py`. It discovers every `$ref` in `schemas/skillfile-sources-v1` that resolves into `schemas/v1`, resolves the JSON Pointer in both candidate and `v1.0.0-rc.10`, and compares the raw definition bytes. Missing baseline files/pointers, changed definitions, and unreadable inputs fail closed.
- The gate checks numbered `core`, `registry`, `manager`, and `environments` citations in both requested protocol documents against clause headings present at rc.10. One adjacent, explicit `Informative-only` marker may qualify its next citation in the sentence. The current environments §9.4 pointer is labeled that way and described as an independent optional capability.
- Added five subprocess tests. Required mutants are a draft `$ref` to a candidate-only definition and a citation to manager §12.1; both assert the production CLI exits 1. A byte-change mutant also fails.
- Wired the gate into Specification CI, stated the rc.10 partial-client baseline in the schema README, and added an Unreleased changelog entry.
- Regenerated `conformance/skillfile-sources-v1/manifest.json` and the rc.13 candidate manifest pin because the suite manifest includes the protocol document bytes. The compatible-core tag and digest remain rc.10.

## Verification

- `go run ./tools/generate-vectors -root .` — exit 0.
- `/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/tmp.lK9d5kxXI6/venv/bin/python -B tools/verify_skillfile_sources_independence.py` — exit 0: **89/89** schema refs byte-identical; **28/28** unmarked clauses present at rc.10 across 21 citation groups; one clause reference explicitly marked informative-only.
- `/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/tmp.lK9d5kxXI6/venv/bin/python -B -m unittest tools.test_skillfile_sources_independence` — exit 0, 5 tests. Both required mutants were rejected by subprocess runs of the gate.
- `PYTHONPATH=tools /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/tmp.lK9d5kxXI6/venv/bin/python -B -m unittest test_skillfile_sources_independence test_validate.SkillfileSourcesSuiteManifestTests test_validate.WorkflowRegenerationScopeTests` — exit 0, 9 tests.
- `/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/tmp.lK9d5kxXI6/venv/bin/python -B tools/validate.py` — exit 0: 64 schemas and 1,169 vector files validated.
- `/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/tmp.lK9d5kxXI6/venv/bin/ruff check tools/verify_skillfile_sources_independence.py tools/test_skillfile_sources_independence.py` — exit 0.
- `git diff --check` — exit 0.
- `git rev-parse HEAD` and `git rev-parse main` both returned `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`. The direct gate ran on this candidate worktree based at that main commit; hosted PR/main CI was not run in this producer worktree.

## Verification limits and iteration notes

- The complete `/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/tmp.lK9d5kxXI6/venv/bin/python -B -m unittest discover -s tools -p 'test_*.py'` suite was interrupted with exit 130 after 4:32 while in the existing `test_validate.EnvironmentVectorTests` path. It was not a passing full-suite run. The focused 9-test group above passed.
- An initial validator/full-suite attempt without installed development requirements exited 1 because `jsonschema` was unavailable. The sole `requirements-dev.txt` dependency was installed in a temporary venv; final validator and focused tests passed there.
- An early validator call concurrent with a manifest-mutation test exited 1 on the test's temporary incomplete manifest. The sequential post-test validator rerun exited 0.
- The `python` alias was unavailable (the gate/test/compile invocations exited 127); Python 3 / the isolated venv interpreter was used. An initial Ruff run found style issues (exit 1); those were fixed and the final Ruff run exited 0.

The prose scanner measures explicit numbered citations; unnumbered allusions are outside its checked set. The informative-only label is an explicit author classification, not a semantic classifier. No `LOGBOOK.md` was edited; this outcome records the manifest-pin decision and the full-suite limitation per campaign instructions.

## Revision 2 — option B

This section supersedes the Revision 1 candidate and its verification where they conflict.

### Changes

- Restored `release/1.0.0-rc.13.json`, `protocol/skillfile-sources.md`, and `conformance/skillfile-sources-v1/manifest.json` from the rc.13 tag. Each `git diff --exit-code v1.0.0-rc.13 -- <path>` command exited 0.
- Replaced the generic informative-only bypass with one conditional citation allowlist entry keyed by the exact path `protocol/skillfile-sources.md` and exact rc.13 sentence about environments §9.4, with a reason string. The allowlist rejects repetitions, condition changes, citations in other files, and all other post-rc.10 clauses.
- Added eight gate subprocess tests covering the allowlisted sentence, its condition removed, an unrelated post-rc.10 citation even with an informative-only marker, moving the exception to another file, an rc.12-only `schemas/v1/agent-context-v1.schema.json` reference, an rc.12-only manager §12.1 citation, changed referenced bytes, and the positive baseline.
- Kept the Specification CI gate step. Updated the draft schema README and Unreleased changelog with the rc.10 baseline and the sole conditional exception.

### Revision 2 verification

- `python3.12 -B tools/verify_skillfile_sources_independence.py` — exit 0: 89/89 reused definitions byte-identical; 28/28 non-exempt cited clauses present at rc.10; exactly one conditional exception allowlisted.
- The same gate run with `--root /Users/administrator/Developer/ReluxWorks/curator/curator-spec` — exit 0 on the clean, unchanged main checkout.
- `uv run --no-project --python 3.12 --with-requirements requirements-dev.txt -- python -B -m unittest discover -s tools -p 'test_skillfile_sources_independence.py'` — exit 0, 8 tests.
- `uv tool run --from ruff ruff check tools/verify_skillfile_sources_independence.py tools/test_skillfile_sources_independence.py` — final run exit 0.
- `git diff --check` — exit 0.
- Three exact-tag diff commands for release metadata, protocol, and conformance manifest each exited 0.
- `uv run --no-project --python 3.12 --with-requirements requirements-dev.txt -- python -B tools/validate.py` — exit 1: `skillfile-sources-v1 manifest digest mismatch: schemas/skillfile-sources-v1/README.md`.
- Standard-library hash check confirmed the conflict: the tag manifest pins README digest `sha256:33f1192658d3f0e93a414de8c5c0efff0e66920a5cfc24bff4a5a3b20af4a601`; the required updated README hashes to `sha256:7bfaa3400130f0c211b75242e9ef4e18a4d7b49a8502916d0c54b2d68e0f44d6`.
- The all-tools unittest command was attempted under the pinned Python 3.12 environment but interrupted with exit 130 at the headless time boundary after it had reported failures. It is not a passing full-suite run. A first attempt without dev requirements exited 1 because `jsonschema` was unavailable. A first Ruff run exited 1 on ISC004; the string was parenthesized and the final Ruff run passed. The earlier `python` alias attempt exited 127 because this host provides `python3`, not `python`.

### Stop-the-line blocker

The task simultaneously requires (1) the schema README to state the rc.10 baseline and conditional exception, (2) the conformance manifest to remain byte-identical to rc.13, and (3) Specification validation to pass. The repository validator includes `schemas/skillfile-sources-v1/README.md` in the manifest inventory and verifies its SHA-256, so requirements 1 and 2 necessarily make requirement 3 fail. The manifest and release metadata cannot both remain rc.13-identical while recording the changed README digest.

Options:
1. Preserve rc.13 identity and move the baseline/exception statement to an unpinned repository README, with an explicit adjustment to the schema-directory README acceptance criterion.
2. Keep the statement in the schema README and authorize recalculating the suite manifest and rc.13 release pin, superseding the current option B restore requirement.
3. Change the validator to stop hashing this README; this weakens the suite integrity contract and is not recommended.

Recommendation: preserve rc.13 identity and explicitly allow an unpinned README for this statement. The exact human decision needed is whether that README-location adjustment is authorized, or whether option B should be superseded to permit the manifest and release-pin changes. No LOGBOOK.md was edited.
