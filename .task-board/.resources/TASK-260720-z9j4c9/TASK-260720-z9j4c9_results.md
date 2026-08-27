# TASK-260720-z9j4c9 implementation evidence

## Provenance

- Blocking handoff: `TASK-260720-1pvfj5` was `done` before repository work began.
- Canonical clone: `/Users/iv/Developer/intranet/cocoaskills`.
- `git fetch origin`: exit 0.
- `git merge --ff-only origin/main`: exit 0 (`Already up to date`).
- Recorded base and `origin/main`: `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`.
- Task worktree: `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z9j4c9/worktree`, detached at the recorded base.
- Released schema 1–5 conformance root: `curator-spec@00b1688a9b2457ca397a0bb550acf47cad8ee967`.
- Accepted schema-v6 candidate root: `curator-spec@57c1f56846d221ecc55786bd3c2467ec32f11730`; `conformance/v1/manifest.json` SHA-256 `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`.

## Delivered source and tests

- Added `src/csk/builds/__init__.py` with the import-safe closed `go-v1` schema identity.
- Extended `src/csk/skillspec.py` with schema 6, `build_roots`, build command fields, deterministic parsing, closed command shape, link-free path checks, root use/disjointness/runtime overlap checks, and nearest-module validation.
- Extended validation-only behavior in `src/csk/skillcheck.py` with build-root prompt warnings, build-root Markdown exclusion, and managed-command resolver guidance for build commands.
- Expanded `tests/test_skillspec.py` and `tests/test_skillcheck.py` across both manifest names, schema 1–5 compatibility gates, runtime fallback, all accepted schema-v6 case files, root/source/module failures, deterministic diagnostics, and prompt validation.
- Preserved existing positional dataclass fields by appending new schema fields after the deployed fields.
- Did not change closure, installer, hashing, toolchain, compiler, cache, shim, workflow, packaging metadata, or protocol pins.

## Validation ledger

- Baseline focused tests before edits: exit 0, `45 passed`.
- Test-first expected-red focused run: exit 1, `45 failed, 55 passed`; failures were the absent schema-v6 and skill-check behavior.
- Final focused pytest with accepted schema-v6 root: exit 0, `113 passed`.
- Task-owned accepted schema-v6 replay: exit 0; 48 cases across `agent-skill-v6` and `csk-skill-v6`.
- Released rc.3 manifest-resolution replay: exit 0, `8 passed`.
- Full pytest with released rc.3 and accepted schema-v6 roots: exit 0, `662 passed, 1 skipped`; the skip is the expected non-Windows PowerShell integration case.
- Focused strict mypy after normal setuptools-scm generation: exit 0, no issues in 3 source files.
- Full `python -m mypy`: exit 0, no issues in 56 source files.
- `git diff --check`: exit 0.
- Final `python -m build`: exit 0; sdist and wheel built with `csk/builds/__init__.py`.
- `python -m twine check dist/*`: exit 0; wheel and sdist both passed.

## Environment notes

- The host has no `python` shim (`command not found`, exit 127); validation used the existing project Python at `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.venv/bin/python`.
- System `python3` lacked pytest (`python3 -m pytest`, exit 1), while the existing project environment provided pytest 9.1.1 and mypy 2.3.0.
- The first focused mypy invocation before package generation exited 1 only because ignored `src/csk/_version.py` was absent. The repository’s normal `python -m build` generated that setuptools-scm file; focused and full strict mypy then exited 0.
- No files were staged or committed.

## Reserved schema-1 field rework

- Addressed the independent-review finding by rejecting only the schema-6
  command fields `driver` and `source_dir` when present in schema-1 command
  objects.
- Kept schema-1's deployed open-extension behavior for unrelated top-level and
  command fields.
- Added direct regressions for script/system commands, both reserved fields,
  and both `agent-skill.json` and `csk-skill.json`; added both-manifest,
  script/system compatibility controls for unrelated extensions.
- Test-first reserved-field gate: exit 1 as expected, `8 failed, 4 passed`;
  every failure was one of the pre-fix false accepts.
- First post-fix focused suite: exit 1, `124 passed, 1 failed`; parser behavior
  was correct, while the older schema-1 build-command test still required the
  previous `"unsupported type"` diagnostic. Its match was narrowed to the
  shared rejection term `"unsupported"`.
- Final focused pytest against the accepted 48-case schema-v6 root: exit 0,
  `125 passed`.
- Reviewer probe replay with `PYTHONPATH=src`: exit 0; all four former false
  accepts report `REJECTED`.
- Full strict mypy: exit 0, no issues in 56 source files.
- Full pytest with released rc.3 and accepted schema-v6 roots: exit 0,
  `674 passed, 1 skipped`; the skip is the expected non-Windows PowerShell
  integration case.
- `git diff --check`: exit 0.
- `python -m build`: exit 0.
- `python -m twine check dist/*`: exit 0; wheel and sdist passed.
- No staging, commit, push, tag, release, pin change, Go invocation, or `intranet`
  mutation occurred.
