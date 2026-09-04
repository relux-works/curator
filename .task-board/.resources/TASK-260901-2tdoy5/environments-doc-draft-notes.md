# Draft notes: protocol/environments.md draft 1 (TASK-260901-2tdoy5)

Deliverable: `protocol/environments.md` (947 lines), signed commit `eddd509e88194f914ca0473a7453d4568e649c7f`
on branch `draft/environments-protocol`, worktree
`~/Developer/ReluxWorks/.worktrees/curator-spec-environments-normative`, base `2a861e5` (exact
`origin/main` at authoring time). Signature verifies against `maintainers.allowed_signers`
(`Good "git" signature for oparin@me.com`, ECDSA `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`).
`make validate` on the worktree: exit 0 (53 schemas, 691 vectors, 134 tool tests, go tests) — docs-only
change, run as a tree-health check.

## Carry-forwards resolved (exact rules chosen)

1. **Zero-applicable-modules output** (§5.4): the generation header part alone; under composition,
   header plus the (empty) chapter parts. Zero-byte output never occurs; a zero-module
   materialization is valid, never an error. Deliberately distinguished from a profile with no
   `context/` directory at all, for which no root-context surface exists and no file is written (§2).
2. **Chapter separator bytes** (§5): a chapter part is exactly `---` LF LF `## Profile: ` name LF —
   thematic break, one empty line, heading. Emitted for every composed profile including the first
   and including one with an empty applicable set (uniform rule; empty chapters keep participation
   visible). Chapter carries the name only — pins live in the header and marker. Parts join with
   exactly one additional LF (one empty line), so every boundary is deterministic.
3. **Referenced-form layout naming** (§5.3): modules land at
   `<home>/.agent-context/modules/<profile-name>/<module-path>` (manifest path verbatim).
   The fixed literal `modules/` segment is the collision guard: profile names may contain dots
   (`[A-Za-z0-9][A-Za-z0-9._-]*`), so without it a profile named e.g. `system-prompt.md` could
   collide with the sibling system output file `.agent-context/system-prompt.md`. Per-profile
   grouping makes same-named modules across composed profiles collision-free by construction.
   claude_code reference part: single line `@.agent-context/modules/<profile>/<path>`; references
   stay inside the home, so the external-include approval is never needed. opencode referenced form:
   root file = header only; ordering lives in the managed `instructions` array of
   `<home>/opencode.json`, which becomes a marker-recorded surface; an unmanaged pre-existing
   `opencode.json` makes the form unavailable → monolithic fallback with
   `environment_form_unavailable`, never an edit of the unmanaged file (ledger discipline).

## Additional normative resolutions the decision left open

- **Generation header exact bytes** (§5.1): closed 7-line grammar with type line
  `curator-root-context-v1`; pins as `commit <hex>` / `state sha256:<hex>`; `compose:` lines and
  `precedence:` line present exactly when composition is active; fixed `generated:` and `notice:`
  byte strings. The `generated:` URL is the spec repo
  (`https://github.com/relux-works/curator-spec`) — fixed bytes so output is byte-identical across
  conforming managers, not only across platforms. Reviewer may prefer a different canonical URL.
- **System-prompt output** (§5.5): **no header and no chapters** — system bytes reach the model
  verbatim, so no generated text is injected; provenance/drift ride on the marker hash instead.
  This deliberately narrows the decision's "every materialized root-context file begins with the
  header" to root-context files only. Inert location `.agent-context/system-prompt.md`; pi's live
  files gated behind a per-profile×environment `system_prompt_files: off|append|replace` machine
  setting, default `off` (Decision 6's "materialized only under the machine setting" made concrete).
- **Composition is flat** (§6): only the activated profile's overlay list applies; overlays' own
  declarations are ignored — no recursion, no cycles by construction.
- **Marker as ledger of record** (§8.3): the core §11 fail-rather-than-overwrite rule extends to
  environment surfaces with the marker as the ownership record; skills keep the §11 adapter ledger;
  the two never merge. Backups in `.agent-environment-backup/` beside the marker; a backup is never
  overwritten (`environment_backup_exists` fails the operation).
- **Diagnostics families**: `profile_*` for source/index/manifest/module validation,
  `environment_*` for registry/materialization/marker/resolve, `subcommand_provider_missing` for
  umbrella dispatch, detector class `context-secret-material` for the manager §7 hook. New codes:
  `profile_source_kind_unsupported`, `profile_source_invalid`, `profile_name_invalid`,
  `profile_unknown`, `profile_index_invalid`, `profile_index_ambiguous`, `profile_root_invalid`,
  `profile_context_manifest_invalid`, `profile_module_missing`, `profile_module_bytes_invalid`,
  `profile_selector_unknown_environment` (warn), `environment_unknown`,
  `environment_target_unknown`, `environment_form_unavailable` (warn),
  `environment_shadowing_path_present` (warn), `environment_composition_invalid`,
  `environment_composition_skill_divergence` (warn), `environment_marker_invalid`,
  `environment_surface_drift`, `environment_surface_missing`,
  `environment_surface_unmanaged_conflict`, `environment_backup_exists`,
  `environment_foreign_manager_detected`, `environment_repair_failed`.
- **Unknown selector environments warn and select nothing** (§3) — mirrors manager §5 rather than
  rejecting, to keep profiles authored against later revisions installable.

## Deliberate deviations from Decision 0010, with rationale

1. **opencode skills target** (§7.1): kept at the manager §5 native surface (`~/.agents/skills`),
   not the decision table's `<home>/skills/`. This follows the decision's own open question 3
   recommendation ("keep the manager §5 native surface until a pinned release proves
   `<home>/skills/`"). Consequence the reviewer should weigh: opencode skills are not
   profile-isolated in managed homes in revision 1.
2. **System files carry no header** — see above; decision prose implies headers on every
   materialized root-context file, and system output is arguably not root context, but the decision
   did not state this explicitly.
3. **Onboarding renumbered** (§9.5): decision steps 1/3-notice/4 appear as inventory/notify/backup;
   the deferred steps (classification, lossy consent, import, `path` kind) are named with
   STORY-260901-2hkq49. Story IDs in normative text follow the repo precedent of decision docs
   citing task IDs; strip if reviewers want normative text ID-free.

## Open items for the reviewer

- **Header `generated:` URL** — spec repo vs product repo vs neutral identifier: pick before the
  determinism vectors freeze bytes.
- **`--append-system-prompt-file` / `--system-prompt-file` flag spellings** (§7.3, §10.2 example):
  Decision 0010 records `--system-prompt`/`--append-system-prompt`(-file); I declared the -file
  variants since the fragment hands over a file path. Verify exact spellings against pinned
  claude_code/pi releases before vectors freeze (open question 7 territory).
- **Xcode secondary-target paths** marked as recorded-from-docs pending pinned-Xcode verification
  (§7.6), mirroring open question 5.
- **Windows claude_code passthrough** declared empty/reserved for revision 1 (§7.4) — needs the
  open-question-5 platform evidence; decide whether `shared` on Windows should instead be an error.
- **`environment_composition_skill_divergence`**: warning today; consider whether strict mode
  should escalate it.
- **`profile use` warning wording** (§9.2) recommends "launching through managed homes" instead of
  naming the launcher, to avoid constraining the launcher spec from this document — confirm.
- The §12 currentness sentence treats a shadow-inert surface as non-current; decision §3 frames
  shadowing as a warning only. I made it non-current because the managed surface is objectively not
  what the tool reads; flip to warning-only if that reads too strict for `--check` exit codes.

## Out of scope, untouched

CHANGELOG.md, COMPATIBILITY.md, core §1.1 identifier list, manager profile sections
(STORY-260901-2rrbff), schemas/vectors (STORY-260901-2ywfl7), cli/curator.md, launcher spec
(STORY-260901-3kucw6), ax PR (STORY-260901-3dzrdw). No push, no PR, no tag — branch stays local
per the brief.
