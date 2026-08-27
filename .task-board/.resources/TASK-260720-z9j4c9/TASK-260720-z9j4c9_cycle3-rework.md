# TASK-260720-z9j4c9 cycle-3 rework evidence

## Outcome

The two acceptance-blocking findings from
`TASK-260720-z9j4c9_review-verdict-cycle2.md` are closed:

- Schema 6 now requires a top-level system command's `command` value to be a
  portable identifier and rejects an explicitly empty `hint`.
- Schemas 1 through 5 retain their deployed system-command behavior. Focused
  schema-5 controls still accept `command: "bin/tool"` and `hint: ""`.
- Every build command now makes Windows `.cmd` resolver guidance mandatory.
  Positive build resolver fixtures document `.cmd`, and a focused warning test
  supplies every other resolver element while omitting only `.cmd`.

The implementation remains limited to the task-owned Python manifest model,
validation-only skill check, build-domain initializer, and focused tests. It
does not execute Go or add hashing, toolchain probes, compilation, cache,
transaction, or install mutation.

## Provenance

- Blocking handoff `TASK-260720-1pvfj5` was independently accepted and `done`
  before the task worktree was created.
- Worktree:
  `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z9j4c9/worktree`.
- Worktree `HEAD`, worktree `origin/main`, canonical main `HEAD`, and canonical
  `origin/main`:
  `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`.
- Released schema 1-5 conformance root:
  `curator-spec@00b1688a9b2457ca397a0bb550acf47cad8ee967`;
  manifest SHA-256
  `7951cda1711d34d2a9dd9a873cf9d537c41ca4e9527e94f138f38743610a379e`.
- Accepted schema-6 candidate root:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`;
  manifest SHA-256
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`.
- The task worktree is detached, unstaged, and uncommitted. No remote was
  mutated.

## Focused rework

### Parser

`src/csk/skillspec.py` adds schema-6-only checks in the existing top-level
system-command branch:

- `commands.<name>.command` must satisfy the shared identifier rule.
- A present `commands.<name>.hint` must be non-empty.

The gates are explicitly versioned, so the schema-5 compatibility behavior
does not change.

### Skill validation

`src/csk/skillcheck.py` now treats either of these as a Windows managed
command:

- a script command with `win_path`;
- any build command.

If prompt-visible resolver instructions omit `.cmd`, validation emits the
existing stable warning code
`skill.command_resolution_contract_missing` with
`Windows .cmd shim suffix` in its message.

### Tests

`tests/test_skillspec.py` adds:

- four schema-6 rejection cases across both `agent-skill.json` and
  `csk-skill.json`;
- two schema-5 compatibility controls across both manifest names.

`tests/test_skillcheck.py` adds the isolated missing-`.cmd` warning case and
updates every positive build-resolver fixture to include `.cmd`.

## Validation ledger

Every gate below ran directly as a standalone process. The reported exit code
is the process exit code.

1. Test-first reviewer reproduction:

   ```text
   /Users/iv/Developer/ReluxWorks/cocoaskills-production/.venv/bin/python -m pytest -q tests/test_skillspec.py::test_schema_v6_rejects_invalid_system_command_fields tests/test_skillspec.py::test_schema_v5_keeps_legacy_system_command_shape_compatible tests/test_skillcheck.py::test_validate_build_command_requires_windows_cmd_resolver_guidance
   ```

   Exit 1, expected red: `5 failed, 2 passed`. The five failures were exactly
   the four schema-6 false accepts and the omitted build `.cmd` warning. This
   gate is intentionally reported as failing.

2. Same focused gate after the implementation:

   ```text
   /Users/iv/Developer/ReluxWorks/cocoaskills-production/.venv/bin/python -m pytest -q tests/test_skillspec.py::test_schema_v6_rejects_invalid_system_command_fields tests/test_skillspec.py::test_schema_v5_keeps_legacy_system_command_shape_compatible tests/test_skillcheck.py::test_validate_build_command_requires_windows_cmd_resolver_guidance
   ```

   Exit 0: `7 passed`.

3. Task-focused parser and skill-check suite with all 48 accepted schema-6
   candidate cases activated:

   ```text
   CURATOR_SCHEMA_V6_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 /Users/iv/Developer/ReluxWorks/cocoaskills-production/.venv/bin/python -m pytest -q tests/test_skillspec.py tests/test_skillcheck.py
   ```

   Exit 0: `132 passed`.

4. Strict type check:

   ```text
   /Users/iv/Developer/ReluxWorks/cocoaskills-production/.venv/bin/python -m mypy
   ```

   Exit 0: `Success: no issues found in 56 source files`.

5. Full test suite with the released schema 1-5 root and accepted schema-6
   candidate root:

   ```text
   CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-2kaopg/protocol-spec/conformance/v1 CURATOR_SCHEMA_V6_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 /Users/iv/Developer/ReluxWorks/cocoaskills-production/.venv/bin/python -m pytest -q
   ```

   Exit 0: `681 passed, 1 skipped in 91.39s`. The skip is the expected
   Windows-only PowerShell prompt integration case on macOS.

6. Distribution build:

   ```text
   /Users/iv/Developer/ReluxWorks/cocoaskills-production/.venv/bin/python -m build
   ```

   Exit 0. Both the sdist and wheel were built, and the wheel contains
   `csk/builds/__init__.py`.

7. Distribution metadata:

   ```text
   /Users/iv/Developer/ReluxWorks/cocoaskills-production/.venv/bin/python -m twine check dist/*
   ```

   Exit 0. The wheel and sdist both passed.

8. Diff hygiene:

   ```text
   git diff --check
   ```

   Exit 0.

## Environment and boundaries

- Project environment: Python 3.14.4, pytest 9.1.1, mypy 2.3.0, build 1.5.1,
  twine 6.2.0.
- The host still has no `python` shim; all validation used the existing project
  virtual environment explicitly.
- A readiness search included a nonexistent `Makefile` and therefore returned
  exit 2 after all tool versions printed successfully. The repository has no
  Makefile, and no Makefile-based workflow was assumed.
- No files were staged or committed. No branch, tag, release, protocol pin,
  `origin`, or `intranet` state was changed.
- No Go command was invoked.
