# Producer brief: cli/curator.md rows for the explicit takeover and the onboarding import

## Where

- Repository `~/Developer/ReluxWorks/curator-spec`. The board hands you the managed Story workspace
  for `STORY-260905-2z9pw4`, branch `task-board/story/STORY-260905-2z9pw4`, base = curator-spec main
  `f39f4a9309f41a9208da817eba9129cf5a9f8dc0`. Work there, nowhere else.
- File in scope: `cli/curator.md` only. Plus `CHANGELOG.md` `## Unreleased` if the repository's
  convention requires an entry for a CLI-surface addition — check how the previous CLI batch did it
  and follow that exactly.

## The gap

`protocol/environments.md` §9.5 requires an operator-triggered takeover:

> Onboarding is triggered only by a **mutating** profile operation that meets unmanaged state —
> `profile install`, `profile use`, `profile sync`, `profile update`, `env resolve --repair`, and an
> explicit takeover.

> Takeover of a specific unmanaged file outside onboarding requires the explicit takeover flag and
> performs the same notice and backup; without the flag, section 8.3 applies and the operation fails
> rather than overwrite.

§9.5 step 4 and §9.6 require an operator-requested import:

> the import itself runs only on the operator's request and under the section 9.6 consent rules

> A lossy import stops with `environment_import_lossy` and the loss list; it proceeds only under an
> explicit per-operation consent flag, which re-reports the loss list as warnings under the same
> diagnostic. Machine configuration MUST NOT pre-record consent.

> `agent-context.json`, `schema_version` 1, `name` `imported` unless the operator supplies a name
> under the core §2 grammar

`cli/curator.md` at `f39f4a9` publishes no row for either. Its `curator profile` rows are install,
list, use, use --clear, update, remove, sync; its `curator env` rows are resolve, status, unmanage,
backups scrub; plus the informative compose and config rows. Nothing names a takeover, an import, a
consent flag, or the import's optional profile name.

## What to add

Two rows in the existing tables, in the existing one-line-per-command voice, plus one example line
each in the examples block.

1. **Takeover.** The operation §9.5 describes: take over specific unmanaged surfaces under the same
   notice and backup that onboarding performs. Decide from §9.5 and §7.6 alone whether it scopes by
   `--env`/`--target` like `env unmanage` does, or names a path; state the flag that makes the
   takeover explicit. It belongs with the `curator env` rows, next to `env unmanage`, because it is
   the inverse operation on the same surfaces.
2. **Import.** The §9.6 operation: classify the §9.5 inventory, reassemble a context-package-shaped
   directory, and install it through the ordinary `path` pipeline. The row must carry the optional
   operator-supplied profile name and the explicit per-operation consent flag for a lossy import.

## Rules

- `cli/curator.md` is a surface index, not a second normative source. Every clause of the two rows
  must be traceable to a sentence already in `environments.md` §9.5, §9.6, §8.3 or §7.6. Cite nothing
  new and invent no behaviour.
- Flag spellings must be consistent with the flags already published: `--env <env-id>`,
  `--target <target-id>`, `--as <name>`, `--purge`, `--restore-backups`, `--repair`. Pick spellings
  that read as part of that family.
- If a required detail genuinely is not in `environments.md` — for example whether the takeover names
  a path or a scope, and what the operand grammar is — do **not** decide it here. Record the exact
  missing sentence in your report as a spec follow-up and publish the row only for the part the spec
  does fix. A row you cannot source is a defect, not a deliverable.
- Do not edit `protocol/environments.md`, the schemas, the vectors, or the manager profile.

## Delivery

Exactly **one** signed commit past `f39f4a9309f41a9208da817eba9129cf5a9f8dc0`, human identity,
no `LOGBOOK.md` and no stray file. Gates: `make validate` and `make regenerate-check` green (venv
under `.temp/venv`) — they must stay green even though this batch touches only prose. Do not push,
tag, or open a PR.

Attach `TASK-260906-1hn93j_drafting-report.md`: a row → sourcing table quoting the environments.md
sentence behind every clause of both rows; the gate output tails; and any spec sentence you found
missing, quoted as the exact question it leaves open. Then `task-board handoff TASK-260906-1hn93j
--role developer`. Never write into the control root.
