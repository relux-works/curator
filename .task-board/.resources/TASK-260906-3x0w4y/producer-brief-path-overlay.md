# Producer brief: make the section 6 `path` overlay declarable

## Where

- Repository `~/Developer/ReluxWorks/curator-spec`. The board hands you the managed Story workspace for
  `STORY-260905-2z9pw4`. Base = curator-spec main; read it with `git rev-parse origin/main`, do not
  assume a value. Note the story branch may still carry an older tip from an earlier task — start from
  current main.
- Files in scope: `protocol/environments.md` (§12.1 knob row only), `schemas/v1/manager-config-v2.schema.json`,
  `conformance/v1/schema-cases/manager-config-v2/`, the generated vectors, `CHANGELOG.md`. Do not touch
  §1 or §6 prose — they are correct and are the authority this fix serves.

## The contradiction

Verified at `550579d`:

- §6: "An overlay is an ordinary context package named by a `git` source with a range or exact form,
  **or by a `path` source under the section 1 rules**."
- §1, `path`: "A `path` declaration that carries `range`, `tag`, `branch`, `revision`, or `directory`
  is `profile_source_invalid`" and "A `path` package serves as a profile root **or as an overlay
  (section 6)**."
- §12.1 knob table: `overlays.<profile>` is an "ordered list of `{ source, range | tag | revision,
  directory?, weight? }`" — a form on every overlay.
- `manager-config-v2` `$defs/overlay`: `"oneOf": [{"required":["range"]},{"required":["tag"]},{"required":["revision"]}]`
  — the same.
- The published cases give the overlay source `/Users/operator/context` a `revision`, which §1 makes
  `profile_source_invalid`.

So §6 and §1 promise a `path` overlay and §12.1 plus the schema make it undeclarable. Stage (c)'s
review drove all three production surfaces in curator and confirmed no operator can declare one; the
only test that passes builds a state the config reader cannot produce.

## The change

Make the requirement form conditional on the source kind, not universal.

1. **§12.1 knob row.** Restate `overlays.<profile>` so the form is required for a `git` source and
   absent for a `path` source. Keep the row's one-line table shape and its existing spelling of
   `directory?` and `weight?`. Change nothing else in §12.1 and nothing in §12.2.
2. **`manager-config-v2` `$defs/overlay`.** Replace the unconditional `oneOf` with a conditional
   shape: a `git` source requires exactly one of `range`/`tag`/`revision`; a `path` source requires
   none and permits none — a `path` overlay carrying `range`, `tag`, `branch`, `revision` or
   `directory` must fail validation, matching §1. Decide how the schema distinguishes the two source
   kinds and justify that decision in the report from §1's own words; if §1 does not give the schema a
   way to tell them apart, say so plainly and publish the closest correct shape rather than inventing
   a discriminator.
3. **Cases.** Stop giving a `path` source a `revision` in the existing cases. Add a positive case for
   a form-free `path` overlay and a negative case for a `path` overlay carrying a form. Regenerate the
   vectors with the generator only.

## The constraint that has bitten this repository before

`.github/workflows/implementations.yml` checks out the **pinned** Go manager and runs its interop
suite against this PR's conformance root. Any case you add to a file that pin *consumes* must stay
passable by it. `vectors/manager-config.json` is such a file and schema-2 cases already live in the
separate `vectors/manager-config-v2.json` family for exactly this reason. Keep it that way, and say
in the report which files the pin consumes and how you checked.

## Delivery

Exactly one signed commit past current main, human identity, no `LOGBOOK.md`, no stray file — and note
that `make validate` generates `tools/__pycache__`, which is now ignored; do not stage it. Gates:
`make validate` and `make regenerate-check`, each a standalone process, observed exit codes quoted. Do
not push, tag, or open a PR.

Attach `TASK-260906-3x0w4y_drafting-report.md`: the knob row and the schema `$defs` quoted before and
after; the discriminator decision with the §1 sentence behind it; the case inventory added and changed;
which files the pinned manager consumes and how you verified none gained an unpassable case; the gate
tails. Then `task-board handoff TASK-260906-3x0w4y --role developer`. Never write into the control root.
