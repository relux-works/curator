# Producer brief: BUG-261017-47xdyt, install must not delete or replace content it does not own

Read the bug and the triage outcome `TASK-261017-yyg01d_triage.md` (attached to TASK-261017-yyg01d; its reproduction script and exact commands are there too). Items A and C2 are this task; B, D and E are separate.

**Normative source:** curator-spec main: skillfile-sources §2 (generated `.agents/skills` versus authored sources; unmanaged-path conflict protection; being inside a managed-output root is not evidence of ownership), core §10–§11 (install markers; adapter ledger removes only preceding-ledger entries and refuses unmanaged collisions), manager §2.5 (serialized publication, stale **managed** removals, rollback). Read them; where this brief and the spec disagree, the spec wins.

**Required:**
1. Stale cleanup (`internal/install/targets.go:stageStaleSkillRemovals`, called from `install.go:stageProjectTargets`) removes only entries with valid preceding Curator ownership (decide ownership from the preceding manager state and markers, not from basename, type or a SKILL.md being present). Any unowned entry the plan would delete makes the operation refuse before publication, naming every such path; their bytes and foreign native links stay as they were.
2. A selected node's existing canonical destination (`stageNode`) may be replaced only with valid preceding ownership; an unowned, unreadable or invalid one refuses by name with no partial change.
3. Unreadable or invalid ownership evidence refuses; it is never absence or permission.
4. Do not add a generic recursive takeover flag for `.agents/`. Keep native refusal and authorized-backup semantics unchanged.
5. Check the global and hybrid callers of the same helpers (`global.go:stageGlobalTargets`): keep their contracts, and say in the handoff what each caller now does.

**Tests** (names from the triage, keep existing names): `TestInstallPreservesUnownedStaleStoreEntries` (unmanaged skill folder, notes-only folder, loose file; refusal diagnostic, exact bytes, link resolution, unchanged ledger), `TestInstallRefusesUnownedSelectedCanonicalDestination` (manual SKILL.md plus note, no marker; invalid and unreadable marker and canonical link variants), neighbours for malformed or unreadable markers, link or special entries, and mixed owned and unowned candidates so a reject-everything change fails. Positive controls: owned stale node removed, clean reinstall, absent destination installs, owned node updates, unrelated `.agents/.instructions/` file survives. Narrowing mutants for the ownership check in both places. Keep `TestDraftUnmanagedAdapterDestinationRefuses` under its name.

**Validation:** full remote gate. Public rules: no hosts, people or paths; describe the source as a field report.

**Handoff:** Change Request with the control map, mutant results, caller analysis and gate run.
