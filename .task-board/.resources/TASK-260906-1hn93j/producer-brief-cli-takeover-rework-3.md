# Producer brief: rework 3 — re-flow the two paragraphs the deletion left ragged

## Why this exists

Rework 2 made exactly the right deletion. It did not re-flow the paragraphs afterwards, so both files
now carry a short line in the middle of a filled paragraph:

`protocol/environments.md:1771-1777` — line lengths 75, 78, **42**, 79, 72, 48:

```
Onboarding is triggered only by a **mutating** profile operation that meets
unmanaged state — `profile install`, `profile use`, `profile sync`, `profile
update`, `env resolve --repair`. Read-only
commands — `profile list`, `env status`, `env resolve` without `--repair` —
```

`profiles/manager.md:2282-2287` — line lengths 73, 77, **47**, 74, 72:

```
Onboarding runs only on a mutating profile operation that meets unmanaged
state — `profile install`, `profile use`, `profile sync`, `profile update`,
`env resolve --repair` — never on a read-only
command, and follows environments §9.5 in order: inventory per registered
```

## The change

Re-flow those two paragraphs to the surrounding convention (hard-wrapped in the low-to-mid 70s, as
every neighbouring paragraph in both files is). **Whitespace only.** Not one word of prose changes:
`git diff --word-diff` restricted to these two hunks must show no added or removed word.

Touch nothing else. `cli/curator.md`, `CHANGELOG.md`, the amended takeover paragraph, and every other
paragraph in both files stay byte-identical.

## Delivery

**Exactly one signed commit past curator-spec main `f39f4a9309f41a9208da817eba9129cf5a9f8dc0`** —
amend `c25d78e`, do not stack. Human identity, no stray file.

Gates: `make validate` and `make regenerate-check`, each a standalone process, observed exit codes
quoted.

Attach `TASK-260906-1hn93j_rework-report-3.md`: both paragraphs quoted after the re-flow with their
line lengths; the `git diff --word-diff` output proving no word changed; the gate tails. Then
`task-board handoff TASK-260906-1hn93j --role developer`. Never write into the control root.
