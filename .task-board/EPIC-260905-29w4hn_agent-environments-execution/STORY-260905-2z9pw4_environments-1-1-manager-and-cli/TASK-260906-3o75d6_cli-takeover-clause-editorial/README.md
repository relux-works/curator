# TASK-260906-3o75d6: cli-takeover-clause-editorial

## Description
Editorial follow-up of review cycle 2 on TASK-260906-1hn93j: cli/curator.md repeats the ~25-word takeover clause across the carrying rows of a one-line-per-command table; state it once, completely, in a note under the table, move the takeover example into the profile use group, and keep every tools/validate.py pin added by TASK-260906-1xbrz6 green. No normative change.

## Scope
Editorial observations from review cycle 2 on TASK-260906-1hn93j, deliberately deferred out of the rework so the normative fix stayed minimal. The ~25-word takeover clause now repeats verbatim across six cli/curator.md rows in a table whose convention is one line per command, and each repetition carries the enumeration members without its closure; one note under the table would state it once and state it completely. env resolve --takeover without --repair is behaviourally determined by the amendment but syntactically unstated, and was correctly left uninvented rather than guessed. The takeover example sits outside the profile use example group it now illustrates.

## Acceptance Criteria
cli/curator.md states the takeover clause once, completely, and in the place the table conventions call for, without adding any rule environments.md does not state; the example sits with the operation group it illustrates. No normative text changes.
