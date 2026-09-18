# TASK-260910-3u9t1e results — service-verify-backup-explicit-key (R7)

Story `STORY-260910-2xe3n2`, wave 4 leaf 3/4. Built on the Story branch
(checkpointed R5 `d7f424c` + R8 `bd27d32`); no commits made by this run
(work left uncommitted for handoff snapshot).

## Per-file changes (worktree, all paths relative to repo root)

- `src/csk_registry/cli.py:293` — new `_implicit_key_warning(home)` helper
  returning `keys resolved from the live home <path>; supply --public-key
  from an out-of-band copy for an independent check`.
- `src/csk_registry/cli.py:300` — `_cmd_verify_backup`: when `--public-key`
  is omitted, prints `WARNING: <text>` on stderr (line 304) and adds the
  same text as the `warning` member of the JSON result on both the
  `backup_valid: true` (lines 332-333) and `backup_valid: false`
  (lines 323-324) envelopes. Verdict logic, exit codes (`0`/`2`), and the
  explicit-key path are untouched.
- `src/csk_registry/cli.py:468` — `verify-backup --public-key` help now
  states the out-of-band expectation
  (`checkpoint signing key from an out-of-band copy (default: registry keys
  under --home, with a warning)`).
- `tests/test_registry.py:799` —
  `test_verify_backup_without_public_key_warns_on_stderr_and_in_output`:
  implicit path warns on stderr (`WARNING`, home path, `--public-key`) and
  in the JSON `warning` member (home path, `--public-key`, `out-of-band`)
  on success (`backup_valid is True`, exit `0`) and on refusal
  (`backup_valid is False` with `error`, exit `2`).
- `tests/test_registry.py:876` —
  `test_verify_backup_with_explicit_public_key_is_quiet`: explicit path
  keeps its exact previous envelopes (`{"backup_valid","log_size","head"}`
  on success, `{"backup_valid","error"}` on refusal), no `warning` member,
  no warning markers on stderr; verdicts match the implicit path
  (`True`/`False`, exits `0`/`2`).
- `README.md:89` — `verify-backup` example now passes `--public-key`;
  `README.md:101` — new paragraph: where the pinned key comes from
  (`genkey`/rotation output, `GET /v1/meta public_keys` while known good),
  how to keep it (with the checkpoint outside the primary store, encrypted,
  access controlled, refreshed at each rotation), and the warn-but-work
  fallback.
- `SECURITY.md:30` — restore paragraph requires `--public-key` from an
  out-of-band copy recorded at `genkey`/rotation time, kept with the
  checkpoint; documents the stderr + JSON warning on implicit resolution.
- `CHANGELOG.md:36` — Unreleased `Security` entry `R7: …` (warning text,
  `warning` member on both envelopes, verdict/exit codes unchanged,
  explicit path byte-identical, curator-spec `47c3c8c`, no wire-schema
  change).

## How each AC line is met

- "CLI behavior change with test": `_cmd_verify_backup`
  (`src/csk_registry/cli.py:300-304,323-324,332-333`) warns on the implicit
  path and is byte-identical on the explicit path; both paths are covered by
  the two tests above driving the real `main()` entry point end to end
  (backup → verify success → advance live log → verify refusal).
- "verify-backup without --public-key warns prominently on stderr and in the
  structured output": `WARNING: …` via `print(…, file=sys.stderr)` plus the
  `warning` JSON member on both envelopes — asserted in
  `tests/test_registry.py:799` (success lines 831-844, refusal lines 864-874).
- "explicit key path unchanged; exit code unchanged": exact-envelope
  assertions (`set(payload) == …`) and exit codes `0`/`2` in
  `tests/test_registry.py:876`; no other code path touched.
- "Tests for both paths; verdict unchanged": success verdict `True` and
  refusal verdict `False` asserted on both paths (implicit lines 836/866,
  explicit lines 916/951).
- "README/SECURITY document the out-of-band key expectation; CHANGELOG
  Unreleased entry R7": `README.md:89,101`, `SECURITY.md:30`,
  `CHANGELOG.md:36`.

## Validation transcripts (from the Story worktree)

Shell `bash`, interpreter `/tmp/csk-venv/bin/python` (CPython 3.14.6,
editable install of the worktree), `CURATOR_CONFORMANCE_ROOT` =
`/tmp/spec-47c3c8c/conformance/v1` (detached curator-spec worktree at the
CI-pinned `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe`):

- `export CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1; /tmp/csk-venv/bin/python -m pytest -q` → exit `0`:
  `179 passed, 2 warnings in 71.67s` (the 2 warnings are pre-existing
  third-party `starlette.testclient`/`anyio` deprecation warnings).
- `/tmp/csk-venv/bin/python -m mypy` (strict per `pyproject.toml`) → exit `0`:
  `Success: no issues found in 15 source files`.
- `/tmp/csk-venv/bin/python -m build` → exit `0`:
  `Successfully built curator_skill_registry-0.1.0.tar.gz and
  curator_skill_registry-0.1.0-py3-none-any.whl` (`dist/`/`build/` are
  gitignored; tree left with only the 5 intended files modified).
- Pre-change focused run (`pytest tests/test_registry.py -q -k verify_backup`,
  tests added before the fix): exit `1`, new implicit-path test failed with
  `KeyError: 'warning'` (authentic failure); post-fix focused run
  (`-k "verify_backup or backup_cli"`): exit `0`, `3 passed`.
- No linter is configured in this repo (no ruff/flake8 config, no lint CI
  job); the static gate is mypy strict, green as above.
- Hosted gate (`sh scripts/remote-gate.sh`) not run by the producer per
  campaign rules; it runs once at handoff.

## Deliberately out of scope

- R4 deployment/rate-limit docs (next leaf `TASK-260910-2c7s0u`), R6 key
  management (story `STORY-260910-9484i4`), spec edits, refusal/exit-code
  changes (brief fixes warning-not-refusal), manager client, tags/releases,
  Docker/deploy changes.

## Spec gaps

- None found. No wire-schema change; the `warning` member is a local CLI
  envelope addition on a non-spec'd admin output (consistent with the R5
  precedent of documenting local behavior against the pinned `47c3c8c`).

## Hygiene

`git status --short --untracked-files=all` lists only `CHANGELOG.md`,
`README.md`, `SECURITY.md`, `src/csk_registry/cli.py`,
`tests/test_registry.py`. No results/logbook/coverage/build files in the
candidate tree (this file was written to `/tmp/TASK-260910-3u9t1e/` and
attached from there).
