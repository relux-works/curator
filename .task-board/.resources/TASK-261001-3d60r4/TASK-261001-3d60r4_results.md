# TASK-261001-3d60r4 — mark-0019-0021-adopted

Review handoff evidence for the 0019-0021 brief, 2026-10-01.

## Changes

- Decisions 0019 and 0021: status is `adopted 2026-10-01 by the operator`.
  All other proposal text is byte-for-byte preserved.
- Decision 0013: added status-section amendment backlinks to 0019
  (Decisions 1, 5, 6.3, 6.5) and 0021 (Decisions 1, 6.4).
  Its body beginning at Context is byte-for-byte preserved.
- UNRESOLVED_QUESTIONS.md: moved both decisions from Filed proposals to
  Adopted decisions, recording the date and operator; retained open questions
  and deferred normative work explicitly.
- CHANGELOG.md: recorded adoption under Unreleased; existing release history
  is byte-for-byte preserved.

Exactly these five repository paths changed. No protocol, schema, vector,
code, permanent test, or LOGBOOK.md changes. Work remains uncommitted in the
assigned Story worktree.

## Verification executed by this developer

No previously attached evidence was substituted for these checks.
Each validation process was invoked directly, without piping through tee.

| Command | Exit | Result |
| --- | --- | --- |
| `python3 tools/validate.py` | 1 | Initial environment failure: jsonschema absent. |
| `python3 -B -m unittest discover -s tools -p 'test_*.py'` | 1 | Initial environment failure: 66 tests, two import errors from missing jsonschema. |
| `/tmp/TASK-261001-3d60r4-venv/bin/python tools/validate.py` | 0 | Validated 72 schemas and 1,253 vector files, including document and local-link checks. Reran after interrupted test processes: exit 0 with the same counts. |
| `/tmp/TASK-261001-3d60r4-venv/bin/python -B -c 'import sys; sys.path.insert(0, "tools"); import validate; validate.validate_local_links(); print("PASS: repository-wide local Markdown links")'` | 0 | Separate repository-wide local Markdown link check. |
| `/tmp/TASK-261001-3d60r4-lychee/lychee-x86_64-apple-darwin/lychee --no-progress --max-retries 3 --retry-wait-time 2 --accept 200,206,429 '**/*.md'` | 0 | CI link-check arguments: 119 total, 54 unique, 118 OK, one excluded, zero errors, one redirect. Excluded links are outside the measured successful set. |
| `git diff --check` | 0 | Whitespace validation, rerun after all edits. |
| Task-scoped Python assertions via `python3` stdin | 0 | 5/5 allowed changed paths; 2/2 exact status-only changes; Decision 0013 body preserved; 2/2 index entries moved; prior release history preserved. |
| `/tmp/TASK-261001-3d60r4-venv/bin/python -B -m unittest discover -s tools -p 'test_*.py'` | 130 | Interrupted after several minutes in release-gate setUp repository copying (KeyboardInterrupt in shutil.copytree / fcopyfile); this is not a passing suite. |
| `/tmp/TASK-261001-3d60r4-venv/bin/python -B -m unittest discover -s tools -p 'test_validate.py'` | 130 | Interrupted after several minutes in an unrelated dotfile scenario mutation invoking validate.main; this is not a passing suite. |
| `/tmp/TASK-261001-3d60r4-venv/bin/python -B -m unittest discover -s tools -p 'test_validate.py' -k TakeoverClosedSetTextTests` | 0 | 30/30 documentation contract tests passed in 0.369 seconds, including negative prose mutations. |

No executable behavior changed, so no permanent behavior tests were added.
The task-scoped assertions verify this documentation change directly and fail
on proposal rewriting, out-of-scope edits, or stale index classification.
Go compilation/generator checks were not run: no Go, schema, or vector changes;
the relevant build/validation gate is tools/validate.py.
Neither the full Python suite nor the whole validator suite ran to a passing
result. The broader runs were stopped to keep verification bounded; the
documentation contract subset was rerun directly, rather than accepting prior
evidence. After interruption, tools/validate.py passed again (exit 0), and
git status --short again showed exactly the five expected documentation paths.

Environment setup: temporary venv creation exited 0; installation from
requirements-dev.txt exited 0 after waiting for venv creation. An early pip
attempt exited 1 while ensurepip was still creating the environment.
Lychee v0.24.2 was downloaded and unpacked into /tmp (both exit 0); an initial
invocation at the wrong extraction path exited 127 before the corrected
invocation above. These failures are not passing checks.

## Normative follow-ups (not performed here)

1. Revise protocol/environments.md §10.1's curator run paragraph for Decision
   0019: agents-management constructs the harness/channel argv; curator-run
   consumes the fragment and appends only native arguments. Record the
   consumer/host boundary under Decision 0021 in the subsequent specification
   revision as needed.
2. Revise Decision 0013 Decisions 1, 5, 6.3 and 6.5 for Decision 0019 and
   Decisions 1 and 6.4 for Decision 0021 in a separately scoped amendment.
3. Revise curator-agent-launcher/SPEC.md for channel construction delegation
   under 0019 and §3 / §4.6 under 0021: the untracked flag, machine-policy
   refusal, operator opt-out, and hosted entry/handoff rules.
4. Track the decisions' unresolved descriptor, environment-composition,
   system-prompt opt-in, bridge, host-upgrade, diagnostics and rollout questions
   before the respective implementation work. Adoption records no resolution
   of those open questions.

Per the current brief, no LOGBOOK.md entry was made. Findings and follow-ups
are recorded in this task-scoped outcome and board notes instead.
