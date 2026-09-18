# Review verdict — TASK-260918-3moznc revision 2

Verdict: **accepted** (`accept_cr(TASK-260918-3moznc, revision=2, …)`).
Reviewer run `RUN-260918-41c0f7` (claude-opus-5), 2026-09-18. Read-only review of
the curator-spec Story worktree
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-1ll22r/worktree`
(branch `task-board/story/STORY-260916-1ll22r`, HEAD = base `5146c7b9ed4b0c07b840ab58f9908f197d667478`,
uncommitted candidate). Nothing in the worktree was edited or created by this
review; every gate and probe ran on disposable clones under `/tmp` whose trees
were verified byte-identical to the worktree (excluding `.git`/`__pycache__`).

## Why an EMPTY curator repository delta is the right outcome

The Change Request `CR-TASK-260918-3moznc-2` carries zero changed paths in the
curator repository (`git diff 6401d3c…6a8cf0d` empty). That is correct for this
leaf: the deliverable is a curator-spec normative revision, and curator-spec is a
separate repository (read-only control root for producers) whose reviewed
artifact is the Story worktree plus the attached `_spec-patch_rev2.patch`
(campaign rules: "patch = git diff HEAD of the worktree … EMPTY curator delta").
The 46 "restored from base" `.task-board` paths listed in the CR diagnostic belong
to other tasks (TASK-260910-2g5v17 / TASK-260910-s9jz1g) and are control-plane
state, not candidate content; the curator Story worktree's `git status` is clean.
Accepting therefore means accepting the curator-spec patch, which is what the
producer role and integration scope of this task are bound to.

## Patch = worktree

- `git -C <worktree> diff HEAD` (new file already intent-to-added; `git ls-files --others` empty)
  sha256 `7fff0900a0d31164b82e695ad32c2466b4512bc7695917243b15a27743b3c2f7`
  = attached `TASK-260918-3moznc_spec-patch_rev2.patch` sha256 (identical); `cmp` BYTE-IDENTICAL;
  `git patch-id --stable` `16755b45e3c6453c43a898e39f99ae1029226265` on both.
- Shape: 8 files, +1817/−73 — `CHANGELOG.md`, `conformance/v1/manifest.json`,
  `conformance/v1/vectors/environments-read-failure.json` (new), `profiles/manager.md`,
  `protocol/environments.md`, `release/1.0.0-rc.9.json`, `tools/test_validate.py`, `tools/validate.py`.
- `git status --short` lists exactly those 8 paths (ignored `tools/__pycache__/` is the
  producer's bytecode, gitignored; not part of the candidate).
- Existing vectors: all 36 HEAD-tracked files under `conformance/v1/vectors/` byte-identical
  (`git diff --quiet HEAD -- <each>`); `conformance/v1/expected`, `conformance/v1/fixtures`
  and `schemas/` byte-identical; manifest diff = exactly one new entry; rc.9 diff = the two
  manifest-pin lines.

## rev1 → rev2 interdiff traces entirely to F1–F3 and the §9.4 note

Applied `_spec-patch_rev1.patch` to a clean copy of `5146c7b` and diffed against the
worktree (`diff -ru`, excluding `.git`; the only extra difference, `fixtures/byte-exact/subst.txt`,
is a `git archive` export-subst artifact of the copy, not a candidate change — the worktree's
copy is byte-identical to HEAD):

| File | Hunks | Traces to |
| --- | --- | --- |
| `protocol/environments.md` | §1.3 "never rebuilt from"; §4(b) exception + dry-run split + "recovers out of band"; §4 `path`-dir paragraph ("lock boundary-check failure"); §8.3 backup paragraph names the code; §8.4.1 lock row (precedence) and backup row (code); §8.5 row; §9.4 migration note; §9.7 two rows; §10.1 repair paragraph + resolve-reads "never rebuilt from"; §10.4 lock row (currency unknown, never rebuilt from) + dry-run rows split; §12 status backup mention; §13 five classes + repair/update + backup records | F1, F3, F2, §9.4 |
| `profiles/manager.md` | §12.2 backup prose + table row; §12.5 prose exception + table lock row; §12.7 status list | F1, F3 |
| `CHANGELOG.md` | never-rebuilt-from sentence; new backup diagnostic; vector description | F1, F3 |
| vectors | +3 lock cases (repair/update refused, repair-rebuilt negative), +5 backup-record cases; closed sets grow to 5 classes / 6 operations / 8 diagnostics; rule prose | F1, F3 |
| `tools/validate.py` | closed sets; `READ_FAILURE_APPLICABLE["backup-record"]`; 8 new scenario pins; 2 new negative pins; repair/update/backup dispositions; per-class operation constraints; entry-shape map | F1, F3 |
| `tools/test_validate.py` | 4 new narrowing tests (lock repair rebuild/write, lock update unknown/rebuild, backup reported-empty, backup restore write/silence); docstring | F1, F3 |
| `manifest.json`, `rc.9` | sha refresh for the changed family file | consequence |

No hunk outside F1–F3/§9.4. The 31 revision-1 cases are unchanged in revision 2.

## Round-2 findings, item by item

| Item | Evidence (file:line, quotes) | Verdict |
| --- | --- | --- |
| F1 precedence stated once | `environments.md:2025` lock row: "`environment_store_untrusted` (entry-class) — never `profile_unknown`; resolve emits no fragment, status is non-current with currency unknown, and no mutating operation (install, update, use, sync, repair, garbage collection) rebuilds, re-materializes, or replaces anything from the unreadable evidence — recovery is an explicit operator action (reinstall from the profile source, or the section 9.2 retained previous lock) that never reads the unreadable file as input" | Pass — one place |
| F1 referenced from §4 | `:742-755`: "…or, for the lock file, cannot be read or parsed (section 8.4.1, lock row) — … `environment_store_untrusted`; a real operation rebuilds it … except a lock file that cannot be read or parsed, which is never rebuilt from (section 8.4.1, lock row): no mutating operation rebuilds, re-materializes, or replaces anything from the unreadable evidence. Dry-run evaluation of a boundary-or-hash entry-class failure reports `would-rebuild-untrusted-store` …; dry-run evaluation of an unreadable or malformed lock reports `environment_store_untrusted` with no rebuild planned … Resolve fails closed … until rebuilt — or, for the unreadable lock, until the operator recovers out of band." `:800-801` `would-rebuild-untrusted-store` "names only a store entry or marker file failure, or a lock boundary-check failure, that a real operation would rebuild" | Pass — the rev1 wording that put the unreadable lock in the rebuild branch is gone |
| F1 referenced from §1.3, §10.1, §10.4, manager | `:216` "and is never rebuilt from (section 8.4.1, lock row)"; `:2669-2673` "an unreadable or malformed lock file is never rebuilt from (section 8.4.1, lock row) — install, update, use, sync, repair, and garbage collection refuse with `environment_store_untrusted` without rebuilding, re-materializing, or replacing anything"; `:2926` lock row "currency unknown; never `profile_unknown`; never rebuilt from — section 8.4.1, lock row"; `:2930-2931` dry-run rows split; `manager.md:2780-2783` "except an unreadable or malformed lock, which no mutating operation rebuilds, re-materializes, or replaces anything from (environments §8.4.1, lock row)"; `manager.md:2841` mirror row | Pass |
| F1 residual contradiction search | `grep 'rebuil'` × lock across both files: every hit is the exception or a reference; no sentence rebuilds from an unreadable lock; the landed `environments-store-boundary.json` has no lock-rebuild case (only `lock-file-wrong-owner-untrusted` at resolve and store-entry rebuilds) | Pass |
| F1 pinned vectors | `lock-unreadable-repair-refused-no-rebuild` (lock/repair/present/file/permission-denied → store_untrusted, no fragment, non-current, unknown, rebuilt false, written false); `lock-unreadable-update-refused-no-rebuild` (lock/update/present/file/unparseable-content → same without fragment); negative `lock-unreadable-repair-rebuilt` (rebuilt+written true → refused). Gate pins operation/class/failure class/outcome/no-mutation (probes below) | Pass |
| F2 §9.7 row | `environments.md:2600`: "recorded passthrough entry whose link state cannot be established (non-current, currency unknown; never "detached"; policy owned by §7.7) \| `environment_passthrough_unreadable`" — inside §9.7 (`:2577` heading, table `:2578-2601`) | Pass — cross-reference, no duplicated policy; no §12.1/§12.2 change (none needed) |
| F3 backup diagnostic | Evidence R2.3 surveys `environment_backup_exists`, `environment_write_would_follow_link`, `environment_import_lossy`, the other `*_unreadable` codes and `environment_store_untrusted` — each genuinely a different condition (checked against §8.3/§8.3.1/§9.6/§4 text), so ONE new code is admitted with the brief's working name `environment_backup_record_unreadable`, error class (no "(warning)" tag; stops restore/scrub/retention before mutating — reason stated). Spelled identically at `environments.md:1917` (§8.3), `:2026` (§8.4.1 row), `:2053` (§8.5), `:2601` (§9.7), `:3098` (§12 status), `:3520` (§13); `manager.md:2518` (§12.2 prose), `:2548` (§12.2 table), `:2911` (§12.7 status); CHANGELOG; 5 vector cases; `validate.py` ×3. Misspelling grep finds none; no schema names diagnostics | Pass |
| F3 vectors | `backup-record-unreadable-status-unknown`, `backup-record-unreadable-restore-stops`, absence-side `backup-record-absent-status-zero` / `backup-record-absent-restore-nothing`, negative `backup-record-unreadable-reported-empty` | Pass |
| §9.4 migration note | `:2399-2406`: "a record it cannot read is a loss with reason under the section 8.4.1 inventory-candidate row — never "not installed" — and migration stops before any write with `environment_import_lossy` naming the skill. Unlike import, the migration path admits no consent flag …; the skill is never dropped silently." Consistent with the §9.7 `environment_import_lossy` row "lossy classification without the consent flag (stops with the loss list)" | Pass |
| Coverage-claim tightening | Evidence R2.5 now separates direct / transitive / out-of-table sites (§6, §9.2, §2/§3/§5, §7.9 transitive; manager §1 out-of-table) — the claim matches the text | Pass |

## Standing items (round 1 re-verified on the rev2 tree)

| Requirement | Evidence | Verdict |
| --- | --- | --- |
| One general rule stated once | `environments.md:2000-2018` §8.4.1: closed 8-class list; "Absent" = ENOENT-class on `lstat`/`open` of the exact path; unreadable = permission/I-O on stat-open-read, non-directory component, symlink where regular required, directory where file expected, undecodable bytes, parse/schema failure; "MUST NOT be reported, persisted, or acted upon as absence, and MUST NOT trigger any absence-shaped action — provisioning, re-seeding, takeover, re-materialization, or an "unprovisioned" verdict"; non-current with currency unknown; "the operation that needs the file fails closed" | Pass |
| Closed per-class table with owners | `:2022-2031`: marker/8.2, lock/1.3, ledger/8.3, backup record/8.3, provisioning seed/7.4, passthrough entry/7.4, recorded surface/8.4, inventory candidate/9.5-9.6 — every row now names a diagnostic (or, for inventory candidates, the loss list + `environment_import_lossy`) | Pass |
| Boundary precedence | `:2033-2039` §4 verification precedes readability classification | Pass |
| Read sites reference the rule (brief-required) | §1.3 `:210-216`; §7.4 seeds `:1528-1537`, passthrough `:1402-1410`; §8.2 `:1855-1862`; §8.3 `:1910-1921`; §9.5 `:2436-2444`; §9.6 `:2499-2501`, `:2528-2531`; §10.1 `:2715-2728` — each cites "section 8.4.1" + its row and names the diagnostic | Pass |
| Other read sites | §1.1 `:110,:119-121`; §4 `:742-755`; §7.1 `:1276-1281`; §7.3 `:1344`; §7.5 `:1565-1568`; §7.6 `:1615-1619`; §9.4 sync `:2357-2359`, migration `:2399-2406`; §10.4 rows; §11 `:2974`; §12 list `:3079-3081`, currency `:3163`, GC `:3218-3221`; §13 `:3498-3529`; manager §12.2/§12.5/§12.7 | Pass |
| Restatement removed | §8.2 and §8.4 local restatements condensed to references (`:1855-1862`, `:1984-1989`) | Pass |
| New codes in every list | `environment_passthrough_unreadable`: §7.4, §7.7, §8.4.1, §9.7, §10.1, §10.4, §13 (7× environments), manager §12.2/§12.5 tables, CHANGELOG, vectors, validator; `environment_backup_record_unreadable` as above. No open-ended wording found ("and similar"/"etc." absent from the hunks) | Pass |
| Diagnostic reuse | lock → existing `environment_store_untrusted` (no `environment_lock_invalid` exists; `environment_lock_unavailable` is the mutation-lock timeout); seed → existing `environment_seed_unreadable`; marker/surface existing; inventory → loss list + `environment_import_lossy`; exactly two new codes, both authorised (passthrough by the producer brief, backup record by the rework brief) | Pass |
| §13 conformance surface | `:3498-3529` five classes, repair/update no-rebuild, backup records, absence contrasts, negatives incl. rebuild-shaped | Pass |
| CHANGELOG | `CHANGELOG.md:10-34` Unreleased/Added names `STORY-260916-1ll22r`, the rule, both codes, the vector family and gate, "Rollout is direct (correctness of an existing MUST)" | Pass; no warn-first two-step required (not a user-visible flip) |
| Scope discipline | Only spec, manager mirror, CHANGELOG, vectors, manifest, rc.9 pins, validator gate + tests; no schema change, no implementation, no proposal 0014–0018 content | Pass |
| Family choice | New hand-maintained `environments-read-failure.json` (generator-owned `environments.json` left untouched), registered in manifest; precedent: write-nofollow / store-boundary / codex-seed families | Pass |

## Independent validation (disposable clones, repo venv `…/curator-spec/.temp/venv`, Python 3.14.6, zsh, `set -o pipefail`)

Clone a (`/tmp/3moznc_rev2` = `5146c7b` + rev2 patch applied with `--index`; `diff -rq` vs worktree: identical):

```text
$ python3 tools/validate.py
validated 62 schemas and 1118 vector files
EXIT:0
$ go test -count=1 ./tools/...
ok  	github.com/relux-works/curator-spec/tools/generate-vectors	1.154s
EXIT:0
$ python3 -B -m unittest discover -s tools -p 'test_*.py'
----------------------------------------------------------------------
Ran 556 tests in 992.232s

OK
EXIT:0
```

(556 = 538 baseline + 18 `ReadFailureVectorTests`, including the producer's 39-name
`test_substituted_scenario_rejected_through_main` replay; the in-place write/restore of the
three touched files happened on the clone, never on the worktree.)

Clone c (`/tmp/3moznc_rev2c`, candidate STAGED so the check compares against the index per convention):

```text
$ make regenerate-check
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json … release/1.0.0-rc.9.json
EXIT:0
$ make regenerate   (second run)
EXIT:0
sha256 inventory of conformance/v1 + release (1124 files): FIXPOINT-IDENTICAL across runs;
regenerated tree `diff -rq` vs worktree: IDENTICAL
```

The three `make validate` gates therefore all exit 0 independently; `make regenerate-check` exits 0.

## Rule-7 probes replayed by the reviewer

In-memory against `validate_environments_read_failure_vectors` (clone c): **135/135 refused, 0 accepted** —
same-class internally consistent ABSENT case under each of the 25 unreadable names (P1);
absence-shaped outcome with unreadable inputs unchanged, all 25 (P2);
`rebuilt=True` / `written=True` / `fragment_emitted=True` under the two no-rebuild lock names (P3);
each of the 6 negatives rewritten as its passing twin, with and without `conforming:false` (P4/P4b);
each negative rewritten as a different violation (P5);
failure-class collapse onto permission-denied for every non-permission scenario (P6);
dropped / renamed / duplicated case (P7); every closed set widened or narrowed (P8);
currency relaxed to `known` on all 25 (P9); `repair_relinks=True` on the three passthrough cases (P10);
backup restore `written=True`, backup status `row_current=True` (P11);
operation swapped under a pinned name — repair→env-resolve, restore→env-status (P14);
absent positives rewritten as unreadable (P15); `rebuilt`/`written` keys dropped (P16).

Through the real entry point `python3 tools/validate.py` with manifest + rc.9 pins recomputed
(so only the gate can fail): rebuild-shaped under `lock-unreadable-repair-refused-no-rebuild`,
`profile_unknown` under `lock-unreadable-update-refused-no-rebuild`, reported-empty under
`backup-record-unreadable-status-unknown`, absent-substitution under
`backup-record-unreadable-restore-stops`, detached-shaped under
`passthrough-lstat-permission-denied-unreadable` → **exit 1 each**; the re-pinned published corpus
→ exit 0 (control). Chunked same-class absent substitution / negative-as-passing through the
entry point on clone b, all 31 names (25 unreadable positives → same-class internally consistent
absent body; 6 negatives → their passing twin): chunk 1 **11/11 exit 1**, chunk 2 **10/10 exit 1**,
chunk 3 **10/10 exit 1** — every refusal message is the scenario pin ("inputs do not match its
pinned scenario") or the negative pin ("a negative needs conforming=false and a reason"), never a
manifest/pin mismatch. All clones restored byte-identical after each probe; `main()` registers the
gate at `tools/validate.py:10192`.

Coverage bound: these are structural scenario-pinning gates over vector JSON, not filesystem
behaviour of the manager; the manager implementation is the follow-up task's scope.

## Observations (no correction required)

- Pre-existing §8.3 sentence "Backups are … never materialized, served, or read by any rule in
  this document" (`:1897-1898`) predates this task and already sat beside restore/scrub; the new
  paragraph speaks of the *inventory* being listed. Not introduced here; noted for the manager task.
- `tools/__pycache__/` is present in the worktree (gitignored producer bytecode). This review
  left no files in the worktree.
- Board: `task-board spawn goal RUN-260918-41c0f7` reports the run is not goal-bound.

## Decision

All three round-2 corrections and the §9.4 note are resolved as the rework brief settled them;
standing round-1 items hold on the rev2 tree; gates are independently green; the gate refuses every
name-preserving absence-shaped and rebuild-shaped replacement. Accepted — routed to `integrating`
via `accept_cr`, revision 2, this file as evidence.
