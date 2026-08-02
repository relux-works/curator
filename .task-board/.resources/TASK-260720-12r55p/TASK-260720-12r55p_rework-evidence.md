# TASK-260720-12r55p rework evidence

## Immutable candidate boundary

- Reviewed candidate root: curator-spec commit `432eb2ee1fe2d6b271e37269f867c8851c325539`.
- `conformance/v1/manifest.json`: `sha256:12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071`.
- `release/1.0.0-rc.6.json` records the same `downstream_consumption.required_manifest_sha256` and `committed_release_pin_advanced=false`.
- Candidate evidence only: the committed CocoaSkills curator-spec pin and `.github` tree are unchanged; no tag, GitHub Release, released-suite pin, or conformance claim was created.

## Reviewer findings and portability rework

The reworked harness has fail-closed bindings for all 77 build-driver rejection names and their CocoaSkills seams, executable adapters for all 10 build-source and 12 toolchain cases, real three-parent-probe and worker-plan capture for all 28 fixed environment values and all five argv forms, an exact 32-case lifecycle name/field map with rollback guards, and manifest membership/digest authentication for every consumed in-scope candidate artifact. Mutation tests cover expected errors, environment, argv, rollback fields, unknown lifecycle fields, and candidate-byte tampering.

Windows portability repair retains the real CocoaSkills validators. The Darwin fixture projects the vector execute bit at the filesystem-observation seam while regular-file, stable-identity, open, and native-header checks still run. The toolchain fixture materializes unsorted directories and files before links, constructs the relative target from native path components, verifies native `readlink` denotes the exact vector target, and exposes its POSIX spelling only at the protocol-byte boundary. Real resolution, escape/dangling validation, mutation revalidation, framing, and fingerprint digest calculation still run. There are no new skips, xfails, or `os.name` bypasses.

Signed CocoaSkills PR 19 head: `ba250bfc4dfe104a160eadd5b5f4e340693bf892` over signed `origin/main` `dacccaaf3ed18740a4d501fe8a3bfec64644c03e`; signature verified for `oparin@me.com`. PR: https://github.com/ivanopcode/cocoaskills/pull/19

## Exact local gates

- `python -m pytest ...`: exit 127 before test collection because this worktree has no `python` executable on PATH; rerun with the project `.venv/bin/python`.
- First eight-node focused selector: exit 4 because one parameterized test name was mistyped; no tests ran. The corrected exact selector is the evidence-bearing run below.
- Corrected exact eight-test Windows regression slice with `CURATOR_CONFORMANCE_ROOT`: exit 0, 8 passed.
- Four originally failing node IDs rerun on signed head `ba250bfc4`: exit 0, 4 passed.
- `tests/test_protocol_conformance.py` with the exact candidate root: exit 0, 452 passed.
- Full pytest suite with the exact candidate root: exit 0, 1,719 passed and 54 existing platform skips.
- Strict `python -m mypy`: exit 0, no issues in 68 source files.
- Post-commit `python -m build`: exit 0; built wheel and sdist for exact head `ba250bfc4`.
- `python -m twine check dist/*ba250bfc4*`: exit 0; wheel and sdist passed.
- `git diff --check HEAD^..HEAD`: exit 0; `git diff --quiet origin/main...HEAD -- .github`: exit 0; worktree and candidate checkout clean.
- Optional board-wide `task-board validate`: exit 0, but reported 1,227 pre-existing cross-board link, resource-payload, and parent-status issues outside this task. No unrelated board data was changed; the task-scoped query confirms all 13 checklist items and the updated outcome resource.

## Hosted exact-head evidence

- GitHub Actions run 30691018727 for exact head `ba250bfc4dfe104a160eadd5b5f4e340693bf892`: conclusion success, 14/14 jobs green. Run: https://github.com/ivanopcode/cocoaskills/actions/runs/30691018727
- `gh pr checks 19`: exit 0. Strict mypy, Ubuntu/macOS/Windows Python 3.11-3.14, and Build artifacts all report `pass`.
