# Evidence — TASK-260918-3moznc: spec-absence-vs-read-failure-discipline

Story `STORY-260916-1ll22r` (absence-vs-read-failure-collapse), wave 3 of the
2026-09 security-audit remediation, `EPIC-260910-2hw1xb`. Spec producer task
(doc-writer): make the environments §8.4 absence-vs-read-failure discipline
structural — one general rule stated once, referenced from every read site,
with a closed table of unreadable outcomes per file class and conformance
vectors exercising unreadable-but-present.

Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-1ll22r/worktree`
(branch `task-board/story/STORY-260916-1ll22r`).
Base commit (HEAD at start, recorded): `5146c7b` ("Add the hardened security
posture profile and name the advisory revocation residual (S1+S3)").
No commits made by the producer; patch = `git diff HEAD` with new files
intent-to-added (`git add -N`).

Finding read first: the brief's Finding section (environments §8.4 requires
absent-vs-failed-read to stay different facts; the launcher/migration
campaign found the two collapsed at five sites over three review cycles with
no structural barrier against the next). Neither
`curator-spec/docs/security-audit-2026-09.md` (S1–S6, E1–E6) nor
`curator/docs/security-audit-2026-09.md` names this collapse as an S/E
finding — it is a review-cycle finding owned by the story itself — so the
CHANGELOG entry names the story, not a finding id.

## 1. What changed per file, and why

### `protocol/environments.md` (normative)

- **New §8.4.1 "Absence versus read failure"**: the general rule stated once
  (settled wording): for the closed 8 file classes, "absent" (ENOENT-class on
  `lstat`/`open` of the exact path) and "present but unreadable or malformed"
  (permission, I/O, type, encoding, parse, schema errors; a non-directory
  path component; a symlink where a regular file is required) are distinct
  facts; a failed read/stat/parse MUST NOT be reported, persisted, or acted
  upon as absence and MUST NOT trigger any absence-shaped action
  (provisioning, re-seeding, takeover, re-materialization, "unprovisioned"
  verdicts); an unreadable file makes the affected row non-current with
  currency unknown and the operation that needs the file fails closed. Plus
  the closed per-file-class table (class, absence outcome, unreadable
  outcome, owning section) and the boundary-precedence note (§4 verification
  precedes readability classification: a §4-contract violation, or an
  unprovable boundary, on a lock/marker/store file is
  `environment_store_untrusted`, never an unreadable code).
- **§8.4 "Drift"**: the local marker/surface absence-vs-read restatement is
  replaced by a reference to §8.4.1 (drift compares against the record, so an
  unreadable surface is not drifted).
- **Reference + row edits**: §1.1 (lock row added; path-operand citation
  renumbered to §8.4.1), §1.3 (lock reads: absent → `profile_unknown`,
  unreadable/malformed → entry-class `environment_store_untrusted`, never
  `profile_unknown`), §4 (entry-class clause extended to lock read/parse
  failure), §7.1 (reconciliation presence checks), §7.3 (launcher path-check
  citation renumbered), §7.4 (seed sentence cites §8.4.1 + non-regular seed
  paths and unparseable revision-B TOML are `environment_seed_unreadable`;
  liveness row gains the unreadable case → new
  `environment_passthrough_unreadable`, repair leaves the entry untouched; a
  directory at the entry path stays detached-side), §7.5 (uncompletable
  shadowing probe reported present-but-unverifiable), §7.6 (unstatable probe
  path counts as present; status reports unverifiable), §7.7 (new
  `environment_passthrough_unreadable` row; seed row covers named XDG
  entries), §8.2 (marker condensed to a §8.4.1 reference), §8.3 (ledger +
  backup-record paragraph), §9.4 (sync: unreadable lock reported, never "not
  found"; migration: unreadable install record stops migration, never drops
  the skill), §9.5 (inventory reads: detected-surface candidates join the
  loss list; unreadable foreign-manager link stops with
  `environment_foreign_manager_detected` naming the path unverifiable;
  dotfile heuristic stays best-effort/warning-only), §9.6 (loss-list citation
  to §8.4.1; unestablishable ledger membership is a loss), §10.1 (resolve's
  reads: marker/surface/lock/passthrough classifications; `--repair` neither
  re-materializes an unreadable surface nor re-links an unreadable entry;
  provisioning-time seeds), §10.4 (new rows:
  `environment_surface_unreadable`, lock-unreadable
  `environment_store_untrusted`, `environment_passthrough_unreadable` at
  resolve), §11 citation renumbered, §12 `profile list` (unreadable lock
  listed as store-untrusted, never omitted), §12 currency citation
  renumbered, §12 GC (unreadable lock fails safe beside unreadable marker),
  §13 (new family paragraph).
- **§9.7 unchanged, deliberately**: no inventory-class code is needed — the
  loss list with reasons plus the `environment_import_lossy` gate already
  covers unreadable inventory candidates. The single new code belongs to its
  owning section's table (§7.7) per the campaign rules ("every new
  diagnostic goes into the section's diagnostics table"); §9.7 lists only
  §9.x lifecycle diagnostics and a passthrough row there would be a
  miscategorized duplicate. Reviewer note: the brief's "§9.7 table" phrase is
  read as covering inventory-class codes if any were needed (none are).

### `profiles/manager.md` (mirrors, kept consistent)

- §12.2 marker/ledger + seed + drift paragraphs cite environments §8.4.1
  instead of restating; mirror diagnostics table gains the
  `environment_passthrough_unreadable` row (header already cites §7.7).
- §12.5 resolve paragraph + mirror table: unreadable marker/surface/lock/
  passthrough never stale; three new resolve rows.
- §12.7 list sentence (unreadable lock listed, never omitted), status
  citation renumbered; GC mirror names the unreadable lock beside the
  unreadable marker.

### `conformance/v1/vectors/environments-read-failure.json` (new, hand-maintained)

31 cases: 20 unreadable-but-present positives (markers 4 at open/read/parse
stages, locks 7 across all failure classes, seeds 6, passthrough 4 across
status + resolve), 6 absence-side positives pinning the contrast per class,
4 negatives whose exact absence-shaped observation is non-conforming, and 1
present-but-absence-side case (directory at a passthrough entry path stays
detached). Closed sets: 4 file classes, 3 operations, 2 presence values, 4
entry kinds, 7 failure classes, 7 diagnostics, 2 currencies. New family file
chosen over extending `vectors/environments.json` because that file is
generated by `tools/generate-vectors/environments.go` (materialization
bytes) and read-failure scenarios do not belong to its generator; this
follows the hand-maintained precedent of the write-nofollow, store-boundary,
codex-seed, env-passthrough, source-signers, and path-kind families.

### `tools/validate.py` + `tools/test_validate.py`

- New `validate_environments_read_failure_vectors` gate (closed sets exact,
  case inventory exact, per-scenario typed input pinning incl. failure class,
  derived §8.4.1 disposition per positive, exact-refusal pins per negative),
  registered in `main()`.
- New `ReadFailureVectorTests`: published-passes, substitution-through-main
  for all 31 names, 4 absence-collapse narrowings, failure-class collapse
  across every non-permission unreadable scenario, currency + repair
  relaxations, absent-rewritten-as-unreadable, negative-rewritten-as-passing
  for all 4 negatives, negative-rewritten-as-other-violation, dropped case,
  relaxed closed set.

### `conformance/v1/manifest.json`, `release/1.0.0-rc.9.json`

- Manifest: one added entry (`vectors/environments-read-failure.json`,
  sorted position, sha256). rc.9: regenerated pins via `make regenerate`.
  All other vectors byte-identical (see §4).

### `CHANGELOG.md`

- Unreleased → Added entry naming the story (`STORY-260916-1ll22r`), the
  rule, the sections, the new diagnostic, the vectors + gate, and the direct
  rollout (correctness of an existing MUST).

## 2. Read-site inventory with verdicts

"Already conformant" = the text already distinguished absence from read
failure (edits, if any, only renumber the citation to §8.4.1).
"Restated → now referenced" = a local restatement condensed to a §8.4.1
reference (+ row). "Silent → now covered" = the site said nothing and now
references §8.4.1 with its outcome. Brief-required sections first.

| # | Read site | Before | Verdict | After |
| --- | --- | --- | --- | --- |
| 1 | §1.3 lock | contract failures → store-untrusted; read/parse failures unclassified | silent → now covered | lock row: absent → `profile_unknown`, unreadable/malformed → entry-class store-untrusted, never `profile_unknown` |
| 2 | §7.4 seeds | "absent not seeded; unreadable → seed_unreadable (section 8.4 discipline)" | restated → now referenced | cites §8.4.1 seed row; non-regular seed paths + unparseable revision-B TOML classified |
| 3 | §7.4 passthrough | liveness row only detached; unreadable silent | silent → now covered | new `environment_passthrough_unreadable`; repair leaves the entry untouched |
| 4 | §8.2 marker | full local restatement citing §8.4 | restated → now referenced | condensed to §8.4.1 marker row |
| 5 | §8.3 ledger/backups | silent | silent → now covered | ledger + backup-record paragraph (§8.4.1 rows) |
| 6 | §9.5 inventory | silent | silent → now covered | loss-list routing; unreadable link stops as foreign-manager/unverifiable |
| 7 | §9.6 import | loss-list rule citing §8.4 | restated → now referenced | cites §8.4.1 inventory row; ledger-membership probe covered |
| 8 | §10.1 resolve | marker sentences citing §8.4 | restated → now referenced | full resolve-reads paragraph (marker/surface/lock/passthrough + repair + seeds) |
| 9 | §1.1 path operand | already distinct codes + §8.4 citation | already conformant | citation → §8.4.1; lock row added to the table |
| 10 | §4 boundary | already fail-closed ("cannot prove → fail closed") | already conformant | one clause makes lock read/parse failure entry-class |
| 11 | §6 overlay source | unreadable source → §1.1 diagnostic | already conformant | no edit (transitive via §1.1) |
| 12 | §7.1 XDG reconciliation | silent on unstatable entries | silent → now covered | seed row: keep recorded link, seed nothing, report seed_unreadable |
| 13 | §7.3 launcher path check | already fail-closed + §8.4 citation | already conformant | citation → §8.4.1 |
| 14 | §7.5 shadowing probe | silent on lstat failure | silent → now covered | present-but-unverifiable, never absent |
| 15 | §7.6 target probe | silent on stat failure | silent → now covered | counts as present; status unverifiable |
| 16 | §7.9 detected release | already "unknown, never matching" | already conformant | no edit (informative row) |
| 17 | §9.4 sync ("locks it finds") | silent | silent → now covered | unreadable lock reported, never "not found" |
| 18 | §9.4 migration | silent on unreadable install records | silent → now covered | fail-closed sentence (core-§10 records; no table row — out-of-table class) |
| 19 | §9.2 use/update/remove | no absence claims; inherit lock row | silent → now covered | via §1.3 lock row (no per-operation sentence) |
| 20 | §10.4 table | silent on surface/lock/passthrough at resolve | silent → now covered | 3 new rows |
| 21 | §11 trust roots | already distinct + §8.4 citation | already conformant | citation → §8.4.1 |
| 22 | §12 profile list | silent (omission hazard) | silent → now covered | unreadable lock listed as store-untrusted |
| 23 | §12 status currency | cites §8.4 | already conformant | citation → §8.4.1 |
| 24 | §12 GC | fail-safe without citation; locks unnamed | restated → now referenced | cites §8.4.1; unreadable lock named |
| 25 | §2/§3/§5 store + snapshot reads | pin recompute/install validation fail closed (§4) | already conformant | no edit |
| 26 | manager §12.2/§12.5/§12.7 mirrors | restatements citing §8.4 or uncited | restated → now referenced | citations + 4 new mirror rows |
| 27 | manager §1 machine-config reads | already fail-closed | already conformant | no edit (out-of-table class) |

## 3. Diagnostic decisions (closed sets stay closed)

- Lock: no `environment_lock_invalid` exists — the only lock-named code is
  `environment_lock_unavailable` (mutation-lock wait timeout, unrelated).
  The existing lock-file-failure code is `environment_store_untrusted`
  (§1.3/§4 entry-class); lock read/parse failures join it. No new lock code.
- Seed: `environment_seed_unreadable` already exists (§7.7). No new code.
- Passthrough: no existing code covers unreadable (only
  `environment_passthrough_detached` for missing/replaced/retargeted links).
  ONE new code: `environment_passthrough_unreadable` (§7.7, §8.4.1 table,
  §7.4/§10.1 text, §10.4, manager mirrors, §13, vectors, gate). Exact
  spelling everywhere (verified by grep, §4).
- Ledger/backup: marker-as-ledger unreadable →
  `environment_marker_unreadable` (exists) with the §8.3 no-remove/replace
  consequence; adapter-ledger membership probes that fail are
  inventory-candidate failures (loss + `environment_import_lossy`); the
  backup-record row is behavioral (fail closed + unknown status, §8.3) with
  no dedicated code — none exists and the brief authorizes new codes only
  for seed/passthrough; link hazards at backup destinations keep
  `environment_write_would_follow_link`.
- Inventory: loss list + `environment_import_lossy` (exist). No new code.
- Marker/surface: both codes exist. No new code.
- New-code appearances: `environment_passthrough_unreadable` in §7.7, §7.4
  text, §8.4.1 table, §10.1 text, §10.4 table, §13 text, manager §12.2/§12.5
  mirrors, CHANGELOG, vectors, `validate.py`, `test_validate.py`. No schema
  enumerates diagnostics (verified: only CHANGELOG, environments.md,
  manager.md name them); no CLI rows name them.

## 4. Validation transcript

All commands run from the story worktree with the repo venv
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv`
on PATH. Exit codes are real, captured without pipe masking (`> file 2>&1;
echo "EXIT:$?"`).

Post-change (this revision):

- `make regenerate` → exit 0 (ran twice; second run byte-identical across
  all of `conformance/v1` + `release` — fixpoint proof, see below).
- `python3 tools/validate.py` → exit 0: `validated 62 schemas and 1118
  vector files` (was 1117; +1 new family file).
- `go test ./tools/...` → exit 0 (`ok .../tools/generate-vectors`).
- `python3 -B -m unittest test_validate.ReadFailureVectorTests` → exit 0:
  `Ran 14 tests ... OK` (14 methods incl. the 31-name
  substitution-through-main test).
- Full `python3 -B -m unittest discover -s tools -p 'test_*.py'` → exit 0:
  `Ran 552 tests ... OK` (538 baseline + 14 new `ReadFailureVectorTests`).
- Regeneration proof (in place of `make regenerate-check`, which diffs the
  worktree and therefore cannot pass on any dirty producer tree — the
  orchestrator/CI runs it post-commit): regenerate ran twice with exit 0
  both times and `diff` of per-file sha256 over `conformance/v1` + `release`
  empty (FIXPOINT-IDENTICAL); `git diff --name-only --
  conformance/v1/vectors/` empty and `git status` shows only the new
  untracked family file there — every pre-existing vector byte-identical;
  `release/1.0.0-rc.9.json` diff is exactly the two manifest pins;
  `manifest.json` diff is exactly the one new entry.
- New-code spelling: `environment_passthrough_unreadable` occurs 6× in
  `protocol/environments.md` (§7.4, §7.7, §8.4.1, §10.1, §10.4, §13), 2× in
  `profiles/manager.md` (§12.2, §12.5 tables), 1× in `CHANGELOG.md`, 3× in
  `tools/validate.py`, 6× in `vectors/environments-read-failure.json` (4
  case expectations, the diagnostics list, the rule prose); a misspelling grep
  (`passthough|passthrought|passthrough_unreadable` minus the exact code)
  finds nothing.

Baselines (pre-change, same venv, same worktree at `5146c7b`):

- `python3 tools/validate.py` → exit 0 (`62 schemas and 1117 vector files`).
- `go test ./tools/...` → exit 0.
- Full unittest discover → 538 tests, all pass-dots, no failures/errors
  (the OK summary line was truncated in background delivery and the wrapper
  exit reflected `tail`, not the suite — recorded green with this caveat;
  the post-change full run below carries a properly captured exit code).

## 5. Out of scope (deliberately untouched)

- Manager implementation (helper, analyzer/lint guard) — the follow-up
  manager task. Lock-rebuild mechanics for unreadable locks inherit the
  existing §4/§10.1 entry-class rebuild language unchanged (vague for locks
  before this change, still vague — not widened here).
- The launcher repository; any other finding.
- `profile sync`/`profile use` unconditional ledger-authorized writes: the
  rule governs classification at read sites, not §8.3-authorized writes
  that do not classify (stated boundary, not an edit).
- Snapshot-content reads (§2/§3 install validation), machine-configuration
  reads (manager §1, already fail-closed), `path`-operand probing (§1.1,
  already distinct codes): outside the 8-class table, already conformant.

## 6. Patch

`TASK-260918-3moznc_spec-patch_rev1.patch` = `git diff HEAD` (base
`5146c7b`) with new files intent-to-added. Curator delta: EMPTY (verified
via `git status` in the curator story worktree).

---

# Revision 2 (rework after `changes_requested`, 2026-09-18)

Reviewer verdict `TASK-260918-3moznc_review-verdict-rev1.md` (route
`to-dev`): three corrections F1–F3 plus two tightenings (the §9.4
migration note, the universal-coverage claim). Rework brief:
`TASK-260918-3moznc_rework-rev2.md`. This section records the revision-2
delta only; sections 1–6 above stand except where noted. Everything not
named below is byte-identical to revision 1: the 31 revision-1 vector
cases are untouched in revision 2 (proven: 0 missing, 0 differing against
the rev1-patch extraction, whose sha256 `d10e11f…` equals the rev1
manifest pin), and every other file carries only the hunks named here.

## R2.1. F1 — unreadable lock: no rebuild from unreadable evidence

Settled in the fail-closed direction (the settled decision). The
precedence is stated ONCE, in the §8.4.1 lock row; every other site
references it:

- `protocol/environments.md` §8.4.1 lock row: an unreadable or malformed
  lock is entry-class `environment_store_untrusted`, never
  `profile_unknown`; resolve emits no fragment, status is non-current with
  currency unknown, and no mutating operation (install, update, use, sync,
  repair, garbage collection) rebuilds, re-materializes, or replaces
  anything from the unreadable evidence — recovery is an explicit operator
  action (reinstall from the profile source, or the section 9.2 retained
  previous lock) that never reads the unreadable file as input.
- §4(b): the rebuild branch gains the exception ("except a lock file that
  cannot be read or parsed, which is never rebuilt from (section 8.4.1,
  lock row)"); dry-run splits (boundary-or-hash → `would-rebuild-
  untrusted-store`; unreadable lock → `environment_store_untrusted` with
  no rebuild planned); resolve fails closed "until rebuilt — or, for the
  unreadable lock, until the operator recovers out of band".
- §4 `path`-directory paragraph: `would-rebuild-untrusted-store` "names
  only a store entry or marker file failure, or a lock boundary-check
  failure, that a real operation would rebuild".
- §1.3 lock reads: "and is never rebuilt from (section 8.4.1, lock row)".
- §10.1 repair paragraph: the exception with the six-operation refusal
  list, and the dry-run split mirroring §4. §10.1 resolve-reads: "never
  `profile_unknown` and never rebuilt from".
- §10.4: lock row gains "currency unknown" (aligns the text with the
  vectors, which already pin currency unknown for locks) and "never
  rebuilt from"; the two dry-run rows split as in §4/§10.1.
- `profiles/manager.md` §12.5 resolve prose: entry-class-rebuild exception
  citing environments §8.4.1 lock row; §12.5 mirror-table lock row gains
  "currency unknown" and "never rebuilt from".

Boundary-or-hash lock failures keep the pre-existing §4(b)/§10.1 rebuild
branch unchanged (vague for locks before revision 1, still vague — only
the read/parse-failure case is settled here, per the brief).

Vectors (rule 7: the validator pins the operation, the file class, the
failure class, the outcome, and no-mutation):

- `lock-unreadable-repair-refused-no-rebuild` (lock / `repair` / present /
  file / permission-denied → store-untrusted, no fragment, non-current,
  currency unknown, rebuilt false, written false). `repair` is
  `env resolve --repair` (§10.1).
- `lock-unreadable-update-refused-no-rebuild` (lock / `update` / present /
  file / unparseable-content → same shape without the fragment key).
  `update` is `profile update` (§9.2, which reads the old lock for the
  audit gate and the delta, steps 2–3).
- `lock-unreadable-repair-rebuilt` (negative): a rebuild-shaped
  observation (right diagnostic, currency, and fragment; rebuilt and
  written true) is refused.

## R2.2. F2 — §9.7 passthrough admission row

`protocol/environments.md` §9.7 gains the
`environment_passthrough_unreadable` row ("policy owned by §7.7" —
cross-reference, no duplicated policy). No knob/lock changes (§12.1/§12.2
untouched).

## R2.3. F3 — backup-record diagnostic (one new code, error-class)

The closed vocabulary was surveyed for an applicable existing code:

- `environment_backup_exists`: fires only when the next generation's
  directory already exists (a half-finished predecessor) — wrong
  condition for an unlistable inventory.
- `environment_write_would_follow_link`: link hazard at destinations
  only — wrong condition.
- `environment_import_lossy`: the §9.6 import consent gate — wrong path.
- the marker / surface / seed / passthrough unreadable codes: wrong file
  class.
- `environment_store_untrusted`: the §4 protected boundary — the backup
  directory is not in the §4 verified set.

None applies, so exactly ONE new code is admitted (authorised by the
rework brief; the working-name spelling is adopted verbatim):
`environment_backup_record_unreadable`. Severity: error (refusal), not
warning — it stops restore, scrub, and retention pruning before mutating
and makes the status row non-current with currency unknown; a warning
would imply the operation proceeds with degraded reporting. This matches
every other `*_unreadable` code, none of which carries a warning tag.

Identical-spelling appearances (verified by grep, see R2.6):

- environments.md (6×): §8.4.1 closed-table backup row, §8.3 owning
  paragraph, §8.5 owning diagnostics table (campaign rule 1), §9.7
  admission row ("policy owned by §8.5"), §12 status-matrix backup row,
  §13 conformance paragraph.
- manager.md (3×): §12.2 prose, §12.2 mirror table (header cites §8.5),
  §12.7 status list.
- CHANGELOG (1×), vectors (5×: 2 case expectations, 1 negative reason,
  the diagnostics list, the rule prose), `validate.py` (3×: the
  diagnostics tuple, the two present-branch dispositions).

Backup vectors: `backup-record-unreadable-status-unknown` (env-status →
new code, non-current, unknown),
`backup-record-unreadable-restore-stops` (`restore` → new code, unknown,
written false; `restore` is the newest-generation restore,
`env unmanage --restore-backups`, §9.2),
`backup-record-absent-status-zero` (absent → null diagnostic, current,
known — a known-empty inventory is current),
`backup-record-absent-restore-nothing` (absent → null, known, written
false), and the negative `backup-record-unreadable-reported-empty`
(reported-empty observation refused). The family grows to five file
classes, six operations, and eight diagnostics; the backup-record
applicable failure classes are permission-denied, io-error, and
parent-not-directory (inventories are listed, never parsed, and the
inventory root is a directory).

## R2.4. §9.4 migration note

§9.4 migration now cites §8.4.1 explicitly (inventory-candidate row): an
install record that cannot be read is a loss with reason — never "not
installed" — and migration stops before any write with
`environment_import_lossy` naming the skill. Rationale: §9.6 classifies "a
skills entry with no recoverable exact declaration" as a loss, which is
exactly what an unreadable install record is; no new code is authorised
for this class, so the closed vocabulary is reused. The disposition
differs from import and is stated, not silent: the migration path admits
NO consent flag (migration always stops; the operator makes the record
readable — or removes the skill through the normal global-remove path —
and retries); the skill is never dropped.

## R2.5. Coverage-claim tightening (reviewer note)

The revision-1 claim "every read-site section references the rule" is
tightened to: every read site is EITHER directly cited (explicit §8.4.1
citation plus its named diagnostic at the site) OR transitively covered
(inherits via a directly cited section) OR declared out-of-table with its
own fail-closed contract:

- Direct: §1.1, §1.3, §4, §7.1, §7.3 (citation), §7.4 (seed and
  passthrough), §7.5, §7.6, §8.2, §8.3 (ledger and backup record), §9.4
  (sync and migration), §9.5, §9.6, §10.1, §10.4, §11 (citation), §12
  (list, status, GC), §13; manager §12.2/§12.5/§12.7.
- Transitive: §6 overlay source (via the §1.1 diagnostic), §9.2
  use/update/remove (via the §1.3 lock row), §2/§3/§5 store and snapshot
  reads (via §4 validation), §7.9 detected release (informative
  "unknown, never matching" row — no absence claim to correct).
- Out-of-table with its own fail-closed contract: manager §1
  machine-configuration reads.

Count correction: revision-1 evidence §1/§4 said "20
unreadable-but-present positives"; the reviewer correctly counted 21 (4
marker + 7 lock + 6 seed + 4 passthrough). The revision-2 corpus has 25
unreadable positives (4 marker + 9 lock including repair/update + 6 seed
+ 4 passthrough + 2 backup-record), 8 absence-side positives (5 absent +
1 present-detached + 2 backup-absent), and 6 negatives: 39 cases.
`ReadFailureVectorTests` grows 14 → 18 methods; the full suite 552 → 556
(see transcript).

## R2.6. Validation transcript (revision 2)

All commands run from the story worktree with the repo venv
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv`
on PATH. Exit codes are real, captured without pipe masking
(`> file 2>&1; echo "EXIT:$?"`).

Pre-revision-2 baseline (same worktree, revision-1 tree, before any
revision-2 edit):

- `python3 tools/validate.py` → exit 0 (`validated 62 schemas and 1118
  vector files`).
- `python3 -B -m unittest test_validate.ReadFailureVectorTests` (from
  `tools/`) → exit 0 (`Ran 14 tests ... OK`, 723.8 s — the
  substitution-through-main replay dominates). One earlier invocation
  from the worktree root failed with exit 1 (`ModuleNotFoundError: No
  module named 'test_validate'`) — a wrong-cwd invocation by the
  producer, not a code failure; rerun from `tools/` per the Makefile
  layout is green.

Post-change (this revision):

- `make regenerate` → exit 0 (first run; rc.9 pins refreshed for the new
  manifest bytes).
- `make regenerate` → exit 0 (second run); per-file sha256 over
  `conformance/v1` + `release` (1124 files) byte-identical between the
  two runs (FIXPOINT-IDENTICAL).
- `release/1.0.0-rc.9.json` diff vs HEAD is exactly the two manifest-pin
  lines; `manifest.json` diff vs HEAD is exactly the one new entry (rev1)
  plus the one-line sha refresh (rev2); `git status --short --
  conformance/v1/vectors/` lists only the new intent-to-added family
  file — all 36 pre-existing tracked vector files byte-identical
  (`git ls-files` counts 37 = 36 tracked + 1 intent-added; `git diff
  HEAD --name-only` under `vectors/` names only the new file).
- `python3 tools/validate.py` → exit 0 (`validated 62 schemas and 1118
  vector files`).
- `go test ./tools/...` → first attempt exit 1 (`FAIL ...
  660.001s`, `Test killed with quit: ran too long`, no test output —
  the binary was starved by a concurrent `go run` regenerate in this
  session plus unrelated `go test` load from other sessions on the
  shared machine; no test failed and no file under `tools/` changed in
  this revision). Rerun `go test -count=1 ./tools/...` (strictly
  stronger: cache disabled) → exit 0
  (`ok .../tools/generate-vectors 1.509s`).
- Focused `ReadFailureVectorTests` minus the substitution replay (17
  methods incl. the 4 new narrowing tests, named explicitly) → exit 0
  (`Ran 17 tests ... OK`, 0.310 s).
- Full `python3 -B -m unittest discover -s tools -p 'test_*.py'` (from the worktree root, per the Makefile) → exit 0 (`Ran 556 tests in 1320.260s ... OK` — 552 baseline + 4 new `ReadFailureVectorTests` methods; includes the 39-name substitution-through-main replay).
- New-code spelling: `environment_backup_record_unreadable` occurs 6× in
  `protocol/environments.md` (§8.3, §8.4.1, §8.5, §9.7, §12, §13), 3× in
  `profiles/manager.md` (§12.2 prose, §12.2 table, §12.7), 1× in
  `CHANGELOG.md`, 3× in `tools/validate.py`, 5× in
  `vectors/environments-read-failure.json`; a misspelling grep
  (`backup_record_unreadable` minus the exact code) finds nothing.
  `environment_passthrough_unreadable` is now 7× in `environments.md`
  (the 6 revision-1 sites plus the §9.7 row), 2× in `manager.md`.
- Regeneration proof (in place of literal `make regenerate-check`, which
  diffs the worktree and therefore cannot pass on any dirty producer
  tree — the orchestrator/CI runs it post-commit): regenerate ran twice
  with exit 0 both times and the 1124-file sha256 inventory is empty of
  differences (FIXPOINT-IDENTICAL, above).

## R2.7. Patch (revision 2)

`TASK-260918-3moznc_spec-patch_rev2.patch` = `git diff HEAD` (base
`5146c7b9ed4b0c07b840ab58f9908f197d667478`, recorded) with `git add -N`
applied (no-op: no new untracked files beyond the rev1 intent-added
vector file). Eight files, same set as revision 1; patch sha256
`7fff0900a0d31164b82e695ad32c2466b4512bc7695917243b15a27743b3c2f7`
(re-verified after the full suite: the suite restores its three
touched files byte-identical — confirmed: the `git diff HEAD` sha256 re-run after the full suite equals the snapshot sha above).
Curator delta: EMPTY (verified via `git status` in the curator story
worktree — no output).
