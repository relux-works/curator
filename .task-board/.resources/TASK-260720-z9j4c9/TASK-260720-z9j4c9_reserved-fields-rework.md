# TASK-260720-z9j4c9 reserved-fields rework evidence

## Outcome

The schema-1 compatibility gap found in independent review is closed.
`driver` and `source_dir` are rejected when mixed into schema-1 script or
system commands, while unrelated schema-1 top-level and command extensions
remain accepted. The guard applies to both canonical `agent-skill.json` and
legacy `csk-skill.json` parsing.

## Changed scope

- `src/csk/skillspec.py`
  - Added the two-field schema-1 reserved command set.
  - Rejects those fields before dispatching the otherwise open legacy command
    shape.
- `tests/test_skillspec.py`
  - Added eight direct reserved-field regressions across both manifest names.
  - Added four unrelated-extension compatibility controls.
  - Kept the schema 1–5 build-command rejection assertion diagnostic-neutral
    across the earlier schema-1 reserved-field rejection.

No other source or test file was changed by this rework.

## Validation ledger

- Tool readiness using the project venv: exit 0; Python 3.14.4, pytest 9.1.1,
  mypy 2.3.0.
- Test-first focused regression: exit 1 as expected; `8 failed, 4 passed`.
- First post-fix focused suite: exit 1; `124 passed, 1 failed` because an older
  test required the previous schema-1 diagnostic wording.
- Final accepted-root focused suite: exit 0; `125 passed`.
- Strict `python -m mypy`: exit 0; 56 source files clean.
- Reviewer probe replay: exit 0; four `REJECTED` results.
- Full pytest: exit 0; `674 passed, 1 skipped`.
- `git diff --check`: exit 0.
- `python -m build`: exit 0.
- `python -m twine check dist/*`: exit 0.

Reviewer probe output:

```text
script+driver: REJECTED commands.tool has unsupported field(s): 'driver'
script+source_dir: REJECTED commands.tool has unsupported field(s): 'source_dir'
system+driver: REJECTED commands.tool has unsupported field(s): 'driver'
system+source_dir: REJECTED commands.tool has unsupported field(s): 'source_dir'
```

## Provenance and boundaries

- Worktree base and `origin/main`:
  `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`.
- Accepted schema-v6 conformance manifest:
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`
  with 48 candidate cases.
- Worktree remains unstaged and uncommitted.
- No push, tag, release, pin change, Go execution, or `wb` mutation occurred.
