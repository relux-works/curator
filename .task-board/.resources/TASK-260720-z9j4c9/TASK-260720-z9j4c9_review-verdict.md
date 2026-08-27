# TASK-260720-z9j4c9 independent review verdict

## Verdict

CHANGES REQUESTED. Route to `to-dev` for focused parser and regression-test
rework. This is an ordinary, recoverable implementation gap, not a
stop-the-line boundary.

## Blocking finding

Schema 1 accepts build-only command fields on otherwise valid legacy script and
system commands.

The accepted schema-v6 contract requires schemas 1 through 5 to reject
`build_roots`, `type: "build"`, and every build-only field while preserving
schema-1's unrelated deployed extension behavior. In `src/csk/skillspec.py`,
script and system command unknown-field checks run only when
`schema >= 2`. The new tests cover `build_roots` and `type: "build"` across
schemas 1 through 5, but do not cover `driver` or `source_dir` mixed into
schema-1 script/system commands.

An independent worktree-source probe (`PYTHONPATH=src`) produced four false
accepts:

```text
script+driver: ACCEPTED keys=['tool']
script+source_dir: ACCEPTED keys=['tool']
system+driver: ACCEPTED keys=['tool']
system+source_dir: ACCEPTED keys=['tool']
```

Required rework:

1. Reject schema-1 command entries containing the reserved build-only fields
   `driver` or `source_dir`, regardless of whether the command otherwise has
   script or system shape.
2. Preserve schema-1's unrelated extension behavior; do not generally close
   all schema-1 command or top-level fields.
3. Add focused regressions for both `agent-skill.json` and `csk-skill.json`,
   covering script/system mixed with each reserved build-only field and a
   control proving unrelated schema-1 extension behavior remains compatible.
4. Rerun the focused schema/skill-check suite and strict `python -m mypy`.

## Scope and provenance audit

- Reviewed worktree:
  `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z9j4c9/worktree`
- Exact worktree base: `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`
- Worktree `origin/main`: `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`
- Clean canonical main clone and its `origin/main` are at the same SHA.
- Diff scope is limited to `src/csk/skillspec.py`,
  `src/csk/skillcheck.py`, new `src/csk/builds/__init__.py`,
  `tests/test_skillspec.py`, and `tests/test_skillcheck.py`.
- Nothing is staged or committed.
- The exact accepted schema-v6 suite was resolved by the producer-recorded
  manifest SHA-256
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`.

## Independent validation

- Focused pytest against that accepted 48-case schema-v6 root:
  `113 passed in 2.51s`.
- Full strict mypy: `Success: no issues found in 56 source files`.
- `git diff --check`: exit 0.
- Tool readiness: Git 2.50.1, Python 3.14.4, pytest 9.1.1, mypy 2.3.0.

The green shipped tests and type checks do not cover the false-accept branch
above, so they do not override the blocking acceptance finding.

The reviewer modified no project code, staged nothing, committed nothing, and
supplied no `commit_ack`.
