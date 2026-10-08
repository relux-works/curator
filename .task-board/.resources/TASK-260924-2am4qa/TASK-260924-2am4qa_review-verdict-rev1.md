# Review verdict — TASK-260924-2am4qa CR rev 1: CHANGES REQUESTED (to-dev)

Reviewed candidate tree 7c9fea4f (reproduced exactly in disposable clone /tmp/rv2am: base 3d2c611 + patch → write-tree 7c9fea4f).

## Blocking
1. `TASK-260924-2am4qa_results.md` (repo root) is part of the candidate. Task documents are board resources, never repository files.
   Fix: delete it from the worktree (keep it as the board outcome resource only). This is the ONLY required change.

## Checked and accepted (no rework needed)
- core §4.4 (protocol/core.md ~L1120-1148): `directory` optional, schema 9 only; grammar/containment referenced to skillfile-sources §1, not restated; absent = `.` root, earlier meaning preserved. Version-gate text (~L193-198): schemas 4-8 reject the member.
- Identity/closure: package identity = (canonical repo, commit, normalized directory); same-tuple unify, different directory for same name conflicts, distinct names from different folders = distinct nodes (diamond). Closure §(~L1457) updated. Lock/marker v5 `package.directory`, audit record + cache key bound to directory (core §10, manager §2.1/§7/status).
- Cross-refs: skillfile-sources.md intro, profiles/manager.md.
- Schemas: agent-skill-v9/csk-skill-v9 new revisions in schemas/draft-sources-v1, `directory` $ref's source-types-v1 `#/$defs/directory` (single shared definition). schemas/v1 only README touched; frozen v1 schemas untouched. Checked: valid-directory-subfolder manifest validates against v9 and is rejected by v8 when relabelled schema_version 8.
- Placement note (flag, not blocking): v9 lives in the draft-sources-v1 namespace; a namespace promotion to released will need to move these two files plus the core.md link (L165) together — mechanical, merges cleanly if the mover globs the directory.
- Vectors: valid absent/root/subfolder + invalid absolute/backslash/empty/escape/glob/parent-component for both manifests; manifest-dependency-directories.json covers grammar, missing folder, no SKILL.md, diamond.

## Rerun evidence (mine, python venv with jsonschema+referencing, pipefail)
- README draft-corpus command: `Schema cases: 138/138; negatives: 105/105; wire schemas: 10/10; Marker migration 25/25, mutants 18/18; Snapshot 3/3` rc=0.
- `python -B -m unittest test_validate.ManifestDependencyDirectoryDraftTests`: 4 tests OK rc=0 (includes parent-escape and diamond narrowing mutants, audit-identity-drop).
- Bound: full `make validate` unit suite (~20 min) and `make regenerate-check` were still running at verdict time and are NOT claimed by me; rework rerun should cite them.
