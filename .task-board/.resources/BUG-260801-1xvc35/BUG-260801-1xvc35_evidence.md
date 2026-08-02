# BUG-260801-1xvc35 developer evidence

Date: 2026-08-01

## Isolated source identity

- CocoaSkills worktree: `/Users/iv/Developer/Wildberries/cocoaskills/.temp/BUG-260801-1xvc35/worktree`
- Branch: `task/BUG-260801-1xvc35-observed-rejections`
- Required base: `ba250bfc4dfe104a160eadd5b5f4e340693bf892`
- `git verify-commit ba250bfc4dfe104a160eadd5b5f4e340693bf892`: exit 0; good ECDSA signature for `oparin@me.com`.
- Signed task commit: `7b01638891646c3862b74be9be392d49e4b88521` (`test: bind rc6 rejections to observed outcomes`).
- `git verify-commit HEAD`: exit 0; good ECDSA signature for `oparin@me.com`.
- `git merge-base HEAD ba250bfc4dfe104a160eadd5b5f4e340693bf892`: exit 0 and returned the exact required base.
- Post-commit `git status --short`: exit 0 with no output.

## Scoped implementation

The signed commit changes only:

- `tests/protocol_conformance_adapters.py`
- `tests/test_protocol_conformance.py`
- `LOGBOOK.md`

The rejection adapter now stores only independently materialized boundary/condition bindings. It executes the relevant CocoaSkills validator, toolchain selector/session, worker policy, dependency validator, context copier, or cache backend and constructs the rejection trace from the observed product result. The exact comparator checks every expected vector leaf against that trace; it does not return a stored expected error/result/effect table.

Coverage includes all 77 rejection names and all 75 condition-bearing cases. Fail-closed regressions mutate all 75 conditions and all 321 expected-field leaves. Dedicated sabotage probes reject an unrelated `SkillSpecError`, reject the wrong toolchain error code, and require `artifact-hash-mismatch` to traverse a real cache `HIT` followed by `CORRUPT` inspection.

The accepted env/argv, manifest, build-source/toolchain, and Windows projection behavior remains in place. No product source, packaging metadata, workflow, protocol schema, pin, tag, release, or claim surface changed.

## Pre-fix expected-red evidence

These commands/probes were run on the exact signed base before the patch:

- Existing focused rejection test: exit 0, `77 passed`; this demonstrated the old false-positive coverage rather than proving observation.
- Standalone 75-condition mutation probe: exit 1 as expected, `condition_cases=75 accepted_mutations=75 rejected_mutations=0`.
- Standalone wrong-seam sabotage probe: exit 1 as expected; the old adapter accepted all three sabotages: `unknown-driver:unrelated-SkillSpecError`, `wrong-go-executable-path:wrong-code`, and `artifact-hash-mismatch:backend-omitted`.

## Post-fix validation

All commands below were run directly as standalone processes.

- Six focused rejection/mutation/sabotage nodes with authenticated exact root: exit 0, `232 passed in 4.08s`.
- `tests/test_protocol_conformance.py` with authenticated exact root: exit 0, `607 passed in 2.47s`.
- Full `pytest -qq` with `CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1`: exit 0 at 100%.
- An earlier verbose full run after the functional patch also exited 0 with `1874 passed, 54 skipped in 406.22s`.
- Strict `python -m mypy`: exit 0, `Success: no issues found in 68 source files`.
- Task-scoped Ruff (`--ignore UP033,RUF012,PYI034,PYI036` for inherited findings in the touched legacy files): exit 0, `All checks passed!`.
- `python -m tabnanny` on both changed Python files: exit 0.
- `python -m compileall -q` on both changed Python files: exit 0.
- `python -m build` from signed commit `7b016388...`: exit 0; built sdist and wheel for `0.12.6.dev38+g7b0163889`.
- `python -m twine check dist/*`: exit 0; both signed-commit distributions and both earlier base-version local distributions passed.
- `git diff --check ba250bfc4dfe104a160eadd5b5f4e340693bf892 HEAD`: exit 0.
- Protected workflow diff check (`.github`): exit 0 with no differences.
- Protected product/release diff check (`pyproject.toml src docs README.md README.ru.md CHANGELOG.md`): exit 0 with no differences.

## Diagnostic and corrected gate exits

For evidence honesty, these non-green invocations are retained:

- The first pre-fix test attempt used a nonexistent virtualenv in the fresh worktree and exited 127. The existing parent PR19 virtualenv was then reused without modifying the parent worktree.
- A diagnostic test invocation without the authenticated root exited 0 with one skipped test; it was not counted as an acceptance gate.
- The first strict-mypy invocation exited 1 because the fresh worktree had not yet generated ignored `src/csk/_version.py`. `python -m build` materialized the generated file; the exact mypy rerun exited 0.
- A quiet full-suite rerun pointed at the spec repository root instead of its manifest directory and exited 2 during collection (`invalid conformance root`). The corrected exact-root command used `curator-spec/conformance/v1` and exited 0.
- Broad Ruff discovery runs exited 1 because the touched legacy files contain inherited `UP033`, `RUF012`, `PYI034`, and `PYI036` findings. Newly introduced formatting/import/SIM findings were repaired; the task-scoped command with only those inherited codes excluded exited 0.

## External integrity

- Curator conformance checkout: `/Users/iv/Developer/ReluxWorks/curator-spec`
- Curator conformance commit: `432eb2ee1fe2d6b271e37269f867c8851c325539`
- Authenticated manifest SHA-256: `12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071`
- Curator conformance checkout status: clean (exit 0, no output).
- Parent PR19 worktree `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-12r55p/worktree`: untouched and clean (exit 0, no output).

