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

## Revision 2 (rework: F1 scenario pinning, F2 private destinations, F3 LOGBOOK)

Revision 1 was rejected (`TASK-260916-1qfpu4_review-verdict-rev1.md`,
F1–F2) plus orchestrator rule item F3. Everything else passed and is
kept byte-identical (normative §§, CHANGELOG, profiles/manager.md,
existing 10 vector cases, manifest shape, rc.9 shape). Changes:

- **F1 — pin scenarios.** `tools/validate.py` gains
  `WRITE_NOFOLLOW_SCENARIOS`: every required name pinned to its
  discriminating inputs (operation, mode, takeover flag, marker
  ownership, target kind, link owner/destination, parent-link state,
  foreign-target fixture). `_validate_write_nofollow_case` refuses any
  named case whose inputs no longer exercise its branch
  (`inputs do not match its pinned scenario`);
  `WRITE_NOFOLLOW_CASES` is derived from the pin table (single source).
- **F2 — private destinations.** The gate models manager-private
  destinations distinctly: a `backup` write naming a pre-existing
  symlink at its target (no parent link) now refuses with
  `environment_write_would_follow_link` and the untouched-target
  expectation — never the §9.5 foreign-manager stop — matching the
  §8.3.1 sentence the verdict quoted. New vector case
  `backup-symlinked-target-refused` (foreign outside-pointing target
  link, no parent link, refused/unchanged/byte-identical former
  target); the existing `backup-symlinked-destination-refused`
  parent-traversal case is kept. §13 now names both backup refusals.
  Vectorized vs text-only: the backup private-destination shape is
  vectorized in both link positions (traversed parent, direct target);
  marker and ledger destinations share the identical §8.3.1 mechanics
  in normative text only — no dedicated marker/ledger vector operation
  exists (same code, same reason, as rev-1 evidence stated).
- **F3 — curator LOGBOOK delta removed.** `git checkout -- LOGBOOK.md`
  in the curator Story worktree; curator `git status --short` is empty.
  Findings live in this evidence and task notes only.
- `tools/test_validate.py`: `WriteNofollowVectorTests` grows from 9 to
  12 tests — `test_substituted_scenario_rejected_through_main`
  (substitutes an internally consistent foreign body under EACH of the
  11 retained names and requires rejection through literal
  `validate.main()` against a repinned on-disk corpus, asserting exit
  1 and the pinned-scenario message; files restored byte-identical),
  plus `test_backup_target_link_foreign_manager_code_fails` and
  `test_backup_target_link_claimed_replaced_fails`. All 8 rev-1
  semantic negatives are retained.

## Revision 2 validation transcript

Shell: `bash` in the spec worktree. Python gates with
`PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH"`.

- `make regenerate` → exit 0 (`go run ./tools/generate-vectors -root .`).
- `python3 tools/validate.py` → exit 0:
  `validated 62 schemas and 1094 vector files`
- `python3 -B -m unittest` over `WriteNofollowVectorTests` fast subset
  (11 tests: published-passes + 8 rev-1 negatives + 2 F2 negatives) →
  exit 0: `Ran 11 tests in 0.004s / OK`
- `test_substituted_scenario_rejected_through_main` → exit 0:
  `Ran 1 test in 236.280s / OK` (11/11 subTests: exit 1 + pinned-scenario
  message through `validate.main()`; tree restored byte-identical)
- Full suite in 4 contiguous discovery-order slices (whole file, unmodified):
  `Ran 100 tests in 241.656s / OK`, `Ran 100 tests in 166.095s / OK`,
  `Ran 100 tests in 199.481s / OK`, `Ran 65 tests in 197.635s / OK`
  → 365/365 green (362 rev-1 + 3 new).
- `go test ./tools/...` → exit 0:
  `ok github.com/relux-works/curator-spec/tools/generate-vectors 3.521s`
- `make regenerate-check` in the worktree → exit 2 (make wrapper;
  underlying `git diff --exit-code` exit 1): EXPECTED-RED — the
  candidate is uncommitted by design (orchestrator owns commits), so a
  diff-against-HEAD check is red by construction; the shown diff is
  exactly the intended deliverable (new vector file, manifest +1
  entry, rc.9 two pins), not regeneration drift.
- Regeneration proof (scratch copy `/tmp/e5-rev2-regen`, candidate
  staged as the scratch HEAD): `make regenerate` then `git diff
  --exit-code -- conformance/v1 release/1.0.0-rc.*.json` → exit 0
  (regeneration is a fixed point on the candidate content).
- F1 attack replication (scratch copy: all 11 case bodies replaced
  with the clean-path body, `make regenerate`, `python3 -B
  tools/validate.py`) → exit 1:
  `validation failed: environments-write-nofollow case
  takeover-symlinked-target-authorized-replaced inputs do not match
  its pinned scenario` (rev 1 accepted this same attack with exit 0).
- `git diff --check` → exit 0.
- Footprint: 8 changed paths only (CHANGELOG, manifest, new vector,
  profiles/manager.md, protocol/environments.md, rc.9, test_validate,
  validate); every pre-existing vector byte-identical
  (`git diff HEAD --name-only -- conformance/v1/vectors/` lists only
  the new file); manifest diff is +1 entry; rc.9 diff is the two
  manifest pins. No implementation code touched; no schema, proposal
  0014–0018, S5 or E7 content.

## Revision 2 deliberately out of scope (unchanged)

Implementation (`TASK-260916-19shmj`), E7 (`STORY-260916-33vuzm`), S5
parallel task (no ownership/permission validation defined here).
Marker/ledger destinations: normative text only (see F2 note above).
Authorized-takeover replace of an inside-pointing link: text mechanics
without a dedicated vector. Unrecorded-regular-file refusal:
pre-existing §8.3 behavior, not vectorized here. No §12.1 knob /
§12.2 lock entry: nothing to configure or lock.
