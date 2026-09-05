# Producer brief: environments 1.1 follow-ups (errata and reviewer minors from batches 1–3)

## Where
Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-follow-ups`, branch `draft/environments-1-1-follow-ups`,
base = curator-spec main `fd237ba`. Work HERE (the managed story worktree the spawn provisions stays untouched; the
story-side Change Request will be an empty delta). Exactly ONE signed commit; `make validate` and `make regenerate-check`
green (venv `.temp/venv`); no push/tag/PR; attach `TASK-260905-2tqh59_drafting-report.md` (item → file:section table, gate
tails); `task-board handoff TASK-260905-2tqh59 --role developer`. Never write LOGBOOK.md or anything into the control
root or the repository.

## Items (the task description is the list; apply every one except item 3)
1. Decision 0012 (`decisions/0012-context-packages-and-semver-locks.md`): add an `## Erratum (2026-09-05)` section after
   Status, in the style of Decision 0010's erratum, with: (a) the compatibility-impact row "manager §12.4 | unchanged" reads
   *bytes change* (isolation knob, liveness row, seeds — batch 3 extended §12.4); (b) the resolution diagnostic for
   disagreeing exact constraints is `context_range_conflict` (schemas-batch reading 4); (c) note that the §9 worked example
   omits `argument` on its system-prompt descriptors and is read as pre-revision (environments §13 says so). Leave the
   original passages in place and mark them `[Erratum 2026-09-05, item N]`.
2. environments §12.1: state the `secret_material_waivers.pin` spelling — bare lowercase hex of 40 (commit) or 64
   (`state_sha256`) characters, matching the marker's pin values and `manager-config-v2` — and mirror the wording in
   `profiles/manager.md` §12 if it repeats the row.
4. environments §7.7: add the `environment_form_unavailable` row (or a one-line cross-reference to §5.7) so every
   diagnostic the registry cites is tabled where a reader looks.
5. environments §8.1 `linked` bullet: add "(except the `claude_code` root-context surface, below)"; §10.1: one explicit
   sentence that a copied surface is verified by the marker's hash, not by link target (cycle-4 nits N-a/N-b).
6. `tools/validate.py` §12.1 cross-check: compare each knob's enum value set against the backticked values of the
   §12.1 `Values` column (a widened `precedence.winner` must fail `make validate`); add a `tools/test_validate.py` case
   that proves it with a scratch mutation.
7. `tools/generate-vectors/manager_config.go`: negative schema cases for `overlay.range` (versionRange pattern),
   `overlay.tag` (gitRefName), and an empty `overlay.source`; regenerate `manager-config-v2.json` and its cases
   (never `manager-config.json`, which must stay byte-identical — the pinned Go manager reads it).
8. `profiles/manager.md` §12.5 and `cli/curator.md`: cite environments §9.2/§10 (not Decision 0013 D6.4) for the
   "curator run always resolves with `--repair`" rule; D6.4 stays the citation for the provider column.
9. Schemas (batch-2 minors F1–F4 in `TASK-260905-1xkxe4_review-findings-schemas-1.md`): `context-lock-v1` rejects a
   member whose `required_by` names itself; `agent-environment-marker-v1` requires `copies ⊆ paths`, a per-surface
   `form` where the surface has one, and rejects unknown surface keys; `launch-env-fragment-v1` rejects path segments
   equal to `..`; the `mcp-pi-none` marker case carries the `env_names` note. Each rule gets a negative schema case
   (generated) and, where the schema language cannot express it (`copies ⊆ paths`, self-`required_by`), a validator
   check with a test.
Item 3 (`system-config-v2` with the environments lockable keys) is NOT this task — it is filed separately.

Every §/identifier you cite is verified against the checkout; the erratum quotes originals verbatim.
