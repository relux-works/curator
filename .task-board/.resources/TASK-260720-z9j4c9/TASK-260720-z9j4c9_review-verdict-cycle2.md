# TASK-260720-z9j4c9 independent review — cycle 2

## Verdict

CHANGES REQUESTED. Route to `to-dev` for focused schema-parser and
skill-check rework. These are ordinary, recoverable implementation gaps, not a
stop-the-line boundary.

## Acceptance-blocking findings

### P1 — schema 6 system commands remain too permissive

`src/csk/skillspec.py:235-249` applies the deployed schema 1–5 system-command
rules to schema 6 without the accepted schema-6 version gates. It requires only
a non-empty string for `command` and permits an explicitly empty `hint`.

The accepted candidate schema requires `systemCommand.command` to be an
identifier and a present `hint` to be non-empty
(`schemas/v1/common.schema.json:60-68`). The accepted Curator parity
implementation makes those checks only for schema 6, while explicitly proving
schema 5 retains `command: "bin/tool"` and `hint: ""`
(`internal/skillspec/build_test.go:203-225`).

An independent worktree-source probe reproduced the false accepts for both
manifest names:

```text
agent-skill.json non-identifier system command: ACCEPTED command='bin/tool' hint=None
agent-skill.json empty system hint: ACCEPTED command='tool' hint=''
csk-skill.json non-identifier system command: ACCEPTED command='bin/tool' hint=None
csk-skill.json empty system hint: ACCEPTED command='tool' hint=''
```

Required rework:

1. For schema 6 only, require system `command` to satisfy the existing
   identifier rule and reject a present empty `hint`.
2. Preserve the current schema 1–5 acceptance behavior byte-semantically.
3. Add focused cases for both manifest names and schema-5 compatibility
   controls mirroring the accepted Curator parity test.

### P2 — build-command validation omits the Windows `.cmd` resolver contract

`src/csk/skillcheck.py:144-165` includes build commands in the managed-command
set, but decides whether Windows `.cmd` guidance is required solely from
`command.win_path`. A build command has no `win_path`, so prompt instructions
that omit `.cmd` are accepted.

The accepted Curator parity implementation marks every build command as a
Windows managed command and emits
`skill.command_resolution_contract_missing` when `.cmd` guidance is absent
(`internal/skillcheck/skillcheck.go:152-165,195-208`). Its focused test requires
the warning and the exact missing-contract phrase
(`internal/skillcheck/skillcheck_test.go:155-172`).

The independent probe produced:

```text
build resolver without Windows .cmd guidance: issue_codes=[]
```

The new Python positive tests currently encode the false acceptance by
asserting no issues for build resolver text without `.cmd`
(`tests/test_skillcheck.py:320-332,345-353`).

Required rework:

1. Treat any build command as requiring Windows `.cmd` shim guidance.
2. Change the positive build resolver fixtures to include `.cmd`.
3. Add a focused case where every other resolver element is present but
   `.cmd` alone is missing, and assert the stable warning code/message.

## Verified evidence

- Reviewed detached worktree:
  `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z9j4c9/worktree`.
- Worktree `HEAD`, worktree `origin/main`, clean canonical main `HEAD`, and
  canonical `origin/main` are all
  `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`.
- Scope is limited to `src/csk/skillspec.py`,
  `src/csk/skillcheck.py`, new `src/csk/builds/__init__.py`,
  `tests/test_skillspec.py`, and `tests/test_skillcheck.py`.
- Accepted candidate manifest SHA-256 is
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`.
- Independent focused pytest against its 48 schema-6 cases:
  `125 passed`, exit 0.
- Independent strict `python -m mypy`:
  `Success: no issues found in 56 source files`, exit 0.
- `git diff --check`: exit 0.
- The prior schema-1 reserved `driver`/`source_dir` false-accept is closed by
  focused both-manifest tests; unrelated schema-1 extension controls remain
  green.
- Original and rework producer outcomes, the prior review verdict, the
  accepted protocol schema, and the accepted Curator parity implementation
  were inspected.

The green supplied tests do not exercise either false-accept branch above, so
they do not satisfy the parity and strict-schema acceptance requirements.

The reviewer changed no project code, staged nothing, committed nothing,
invoked no Go command, and supplied no `commit_ack`.
