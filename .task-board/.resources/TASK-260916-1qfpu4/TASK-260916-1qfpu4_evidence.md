# TASK-260916-1qfpu4 evidence — spec-nofollow-write-rule (E5)

Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-73a5zg/worktree`
(branch `task-board/story/STORY-260916-73a5zg`, base `684c9f1`).
Finding: `docs/security-audit-2026-09.md` E5 + Appendix B (takeover/repair
writes must replace the directory entry, never write through a symlink).

## What changed per file

- `protocol/environments.md`
  - New §8.3.1 "Write discipline": the one normative rule, stated once —
    every materialization, takeover, repair, or backup write to a managed
    surface (any mode), a backup, the marker, or the adapter ledger
    replaces the directory entry (operation-private temp + rename) and
    MUST NOT follow a symlink at the target or below the managed root
    (`O_NOFOLLOW`-class open, `lstat`-class inspection; components at or
    above the managed root resolve normally). Closed target-link
    disposition: manager-owned link → replace as entry; §9.5
    foreign-manager stop → exactly §9.5 (authorized takeover backs up
    the link itself, same link text, never dereferenced, then replaces);
    any other unowned link → ledger rule (`environment_surface_unmanaged_conflict`
    unless a takeover authorization covers the path). Non-manager link
    at a path component below the managed root, or at a backup/marker/
    ledger destination, refuses with the one new diagnostic
    `environment_write_would_follow_link` — no takeover flag authorizes
    traversal. Backup reads MUST NOT dereference either.
  - §8.5 diagnostics table: new `environment_write_would_follow_link` row.
  - §9.5: takeover writes are §8.3.1 writes (authorization covers entry
    replace after backup, never opening the link target); step 3 backs
    up a replaced symlink as a symlink.
  - §8.4: drift inspection is `lstat`-class (`readlink`, never opened
    through). §10.1: repair writes are §8.3.1 writes; link-target
    currency read is `lstat`-class.
  - Pointers where surfaces are written: §5 head, §5.8 MCP file, §7.5
    shadowing existence check (`lstat`, symlink counts as present, never
    dereferenced), §8.1 first-provisioning transaction.
  - §12: `env status` reports link-blocked managed-surface/backup/marker
    paths; the row is non-current.
  - §13: conformance surface naming `vectors/environments-write-nofollow.json`
    and its case inventory.
- `profiles/manager.md` (§12.2): one write-discipline pointer sentence +
  the `environment_write_would_follow_link` mirror-table row (same
  spelling as §8.5).
- `conformance/v1/vectors/environments-write-nofollow.json` (new,
  hand-written per the shell-hook-trust precedent): 10 cases —
  authorized-takeover replace with backup-as-link, unauthorized stop
  (`environment_foreign_manager_detected`), symlinked-parent refusals
  under both authorization states, planted-link repair, manager-owned-link
  repair, backup-destination refusal, inside-link ledger refusal
  (`environment_surface_unmanaged_conflict`), clean-path and recorded-file
  positives. Every foreign-link case asserts the link's former target is
  byte-identical afterwards (fixture digest recomputed by the gate).
- `conformance/v1/manifest.json`, `release/1.0.0-rc.9.json`: regenerated
  (`make regenerate`); manifest +1 entry, rc.9 the two manifest pins.
  All pre-existing vectors byte-identical (regeneration footprint is
  exactly the S6 shape: manifest +4, rc.9 4 lines).
- `tools/validate.py`: `validate_environments_write_nofollow_vectors`
  (registered in `main()`), deriving every case's outcome/diagnostic/
  backup/entry/untouched-target values from the §8.3.1 disposition
  model; fixture digests recomputed from bytes. Spec-harness gate, not
  implementation code (same standing as the S6/E1 spec-task gates).
- `tools/test_validate.py`: `WriteNofollowVectorTests` (9 tests:
  published vector passes + 8 narrowing negatives).
- `CHANGELOG.md`: Unreleased entry "E5: …" (direct rollout, under the hood).

## Why

E5: §9.5 detects a foreign-manager symlink but no sentence forbids
writing *through* it; a conforming implementation written from the text
reproduces the defect the reference implementation already fixed
locally. The rule is stated once in §8.3.1 and referenced from every
write site, so takeover = backup + entry replace, never a write through.

## New-code justification (closed sets stay closed)

Only one new diagnostic was needed. Reused where the situation is
already named: `environment_foreign_manager_detected` (unauthorized
takeover of an outside-pointing link), `environment_surface_unmanaged_conflict`
(unowned target link outside takeover). The uncovered case is traversal:
a symlinked path component below the managed root (or a link at a
manager-private backup/marker/ledger destination) — the target file may
be perfectly recorded, so no existing stop names the refusal, and the
takeover flag must not authorize it. Hence
`environment_write_would_follow_link` (error), spelled identically in
§8.3.1, the §8.5 table, the manager.md mirror row, §10.1/§12/§13,
CHANGELOG, vectors, and the gate. No new knob (the rule is
unconditional), no schema change, no CLI change.

## Validation transcript

Shell: `bash` in the worktree. Python gates run with the worktree
`.venv` first on `PATH` (created per repo `requirements-dev.txt`:
`uv venv && uv pip install -r requirements-dev.txt`; `.venv` is in the
validator's `NON_SURFACE_DIRECTORIES` and leaves no tree diff).

- `make regenerate` → exit 0 (`go run ./tools/generate-vectors -root .`).
- `python3 tools/validate.py` → exit 0:
  `validated 62 schemas and 1094 vector files`
- `python3 -B -m unittest discover -s tools -p 'test_*.py'` → exit 0:
  `Ran 362 tests in 618.639s / OK` (includes the 9 new
  `WriteNofollowVectorTests`)
- `go test ./tools/...` → exit 0:
  `ok github.com/relux-works/curator-spec/tools/generate-vectors 2.323s`

## Deliberately out of scope

- Implementation (`TASK-260916-19shmj`), E7 launcher ownership
  (`STORY-260916-33vuzm`), S5 store boundary (parallel task; no
  ownership/permission validation defined here).
- Marker/ledger destination links share the §8.3.1 mechanics in text;
  only the backup-destination shape is vectorized (same code, same
  reason). Authorized-takeover replace of an inside-pointing link uses
  the case-1 mechanics by text (§8.3.1 bullet 3) without a dedicated
  vector. Unrecorded-regular-file refusal is pre-existing §8.3
  behavior, not vectorized here.
- No `§12.1` knob / `§12.2` lock entry: nothing to configure or lock.
