# TASK-260917-2vapkz developer handoff evidence

## Patch contents

The attached `TASK-260917-2vapkz_spec-patch_rev1.patch` contains the uncommitted Story-worktree changes. Core §8 normatively defines the v2 domain-separated SHA-256 framing and canonical file records; content identities compare `(hash_version, digest)`. Frozen v1 schema shapes remain unchanged, while current markers, context locks, environment state, verdict/pin state, Skillfile source state, and registry records use v2 shapes or required version members. Registry matching requires equal framing versions. The interim v1-reader rule treats a NUL in any regular skill/context snapshot file, at any depth, as a blocking opaque finding.

The conformance vectors include the constructed v1-colliding tree pair with distinct v2 hashes, empty-tree digest, an ordinary tree with exact v2 hex, a registry version-mismatch non-match, and nested NUL data. The changelog has an Unreleased entry. Schema and source corpus cases cover new shapes and frozen-v1 rejection.

## Verification results

Successful commands (real exit code 0):

- `git diff HEAD --check` (rerun after marking untracked source files intent-to-add).
- `PATH=/tmp/curator-spec-TASK-260917-2vapkz-venv/bin:$PATH python3 tools/validate.py` — validated 70 schemas and 1249 vector files.
- The Python conformance script in `conformance/skillfile-sources-v1/README.md` — 131/131 schema cases, 103/103 negatives, 11/11 wire schemas, and 3/3 snapshot byte vectors. Its output explicitly reports 0 manager semantic cases; those remain unverified by this specification command.
- `go test ./tools/...` — passed.
- `make regenerate-check` — passed after the generated outputs and source manifest were staged as the regeneration baseline; no commit was created.
- Focused Python unittest batches all exited 0: content-hash/context/environment/shared-fixture/manager-config/source-manifest selections (88 tests); additional source/release-policy selections (64); stable release gate (6); all 29 RC13 release-gate tests in six bounded groups; and validator selections of 96, 137, and 129 tests.

Non-green and interrupted outcomes:

- Initial `make validate` without the task venv exited 2 because `jsonschema` was unavailable in system Python. Installing requirements into system Python exited 1 under PEP 668; a task-local venv was created and used for the later successful direct validator run.
- Four iterative `python3 tools/validate.py` runs exited 2 while identifying, in sequence, an incomplete registry artifact/sort key, missing context-lock `hash_version`, a stale context-resolution vector, and a stale Skillfile source manifest. These issues were corrected; the final direct validator run above exited 0.
- The first `go test ./tools/...` exited 1 because the frozen marker schema indexed-example sets were inconsistent. The frozen-version negative case was added to both relevant corpora; the final Go run above exited 0.
- `make validate` with the task venv exited 130 when interrupted at the per-shell time bound during unittest discovery in `test_release_gate.py` setup. Its preceding `tools/validate.py` phase passed (70 schemas, 1248 vector files at that point). A separate bounded validator unittest group exited 130 at `DotfileManagersVectorTests.test_substituted_scenario_rejected_through_main` after 9:16. The exact isolated command `PYTHONPATH=tools PATH=/tmp/curator-spec-TASK-260917-2vapkz-venv/bin:$PATH python3 -B -m unittest test_validate.DotfileManagersVectorTests.test_substituted_scenario_rejected_through_main` then exited 0 (`Ran 1 test in 256.595s`). A broader `unittest discover ... -k test_substituted_scenario_rejected_through_main` retry also exited 130; the name matched three classes and was interrupted in the read-failure class. The aggregate `make validate` did not complete as one command, so its exit 130 remains explicit. The direct validator, bounded focused suites, Go suite, source conformance script, and regeneration check each have separate green results above.

## Handoff boundary

The work remains uncommitted in the Story worktree as required. Issue #59 is not closed by this developer handoff; closure depends on the landing commit during integration.
