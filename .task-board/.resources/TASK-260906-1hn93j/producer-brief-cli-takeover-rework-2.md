# Producer brief: rework 2 — close the self-referential enumeration, and nothing else

## Why this exists

Review cycle 2 **accepted** `e01de3f`: no blocking, no major, and the §9.5 amendment verified as
faithful and minimal by a reviewer that also mutation-tested the gates themselves. Three minors were
carried forward. One of them is not shippable as a carried-forward minor, because it makes the landed
normative text circular. That one clause is this rework. **Everything else in `e01de3f` stays exactly
as it is.**

## The defect

The amendment you wrote says:

> The flag is accepted on exactly the mutating operations named above as onboarding triggers —
> `profile install`, `profile use`, `profile sync`, `profile update`, and `env resolve --repair` —
> and on no other operation.

The list it points at, four paragraphs above at `protocol/environments.md:1772-1774`, still reads:

> Onboarding is triggered only by a **mutating** profile operation that meets unmanaged state —
> `profile install`, `profile use`, `profile sync`, `profile update`, `env resolve --repair`, **and
> an explicit takeover**.

So "the operations named above as onboarding triggers" is a six-member list whose sixth member is the
takeover itself, while the amendment enumerates five and says "no other operation". The set contains
itself and disagrees with its own restatement. `profiles/manager.md:2283-2285` carries the identical
list with the identical trailing clause.

This is the same class of defect the whole task exists to fix: a sentence that reads the takeover as an
operation of its own.

## The change

In **both** files, strike the trailing takeover clause from the onboarding-trigger list so the list
names exactly the five mutating operations. The amended paragraph already defines what the takeover is
and which operations carry it; the trigger list does not need to name it, and after the amendment it
must not.

Adjust only the punctuation the deletion requires. Do not reword the surviving list, do not touch the
amended paragraph, do not touch `cli/curator.md`, and do not act on the other two carried-forward
minors — they are filed as separate follow-ups and are explicitly out of scope here:

- the closed set excluding `profile import` activation and §9.4 global in-place materialization is an
  orchestrator decision, already filed;
- the editorial questions (the takeover clause repeated across six rows, the example's placement) are
  filed and deliberately deferred.

## Delivery

**Exactly one signed commit past curator-spec main `f39f4a9309f41a9208da817eba9129cf5a9f8dc0`.**
Amend `e01de3f` — do not stack a second commit; a Change Request candidate must be exactly one
single-parent commit past the Story base. Human identity, no `LOGBOOK.md`, no stray file.

Gates: `make validate` and `make regenerate-check`, each run as its own process, with the observed exit
code quoted. State whether the deletion touches any schema, vector, diagnostic table or conformance
case. Do not push, tag, or open a PR.

Attach `TASK-260906-1hn93j_rework-report-2.md`: the two sentences quoted before and after; a statement
that nothing else in the commit changed, supported by `git diff e01de3f..HEAD` limited to those two
hunks; the gate tails. Then `task-board handoff TASK-260906-1hn93j --role developer`. Never write into
the control root.
