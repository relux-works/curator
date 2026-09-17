# Evidence — TASK-260910-39fzpq: protected-boundary contract for the environments root and store (S5)

Story `STORY-260910-148pj1`, wave 3 of the 2026-09 security-audit remediation.
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-148pj1/worktree`
(branch `task-board/story/STORY-260910-148pj1`, base `684c9f1`).
Role: doc-writer (normative spec text + conformance vectors). No implementation code touched.

Finding: `docs/security-audit-2026-09.md` S5 (Medium) — core §9.3 gives the
build cache a normative ownership/permission/containment contract revalidated
on every lookup; environments §4/§8.2/§10.1 gave the store, locks, and markers
none ("link-target identity is sufficient currency"), so a same-user swap of
store bytes was undetected at resolve. Modeled the contract on core §9.3
"Shared protected-cache rules" (read first), per the brief.

## What changed per file

- `protocol/environments.md`
  - §4 "Profile store": the contract — environments root and every store
    entry are protected state in the core §9.3 sense (manager-created,
    manager-protected, resolved independently of package input).
    Verification on every `env resolve`, and again under the manager-home
    mutation lock for every mutating profile operation (install, update,
    use, sync, repair, GC): ownership, private mutation permissions or
    DACL, containment (entry paths resolve below the root), regular file
    types, link safety (`lstat`, no symlink at the root, entry root, or
    any component the manager did not create) — for the environments
    root, the store root, the lock file, the home markers, and every
    store entry the lock names. Fail-closed rule for implementations
    that cannot prove the boundary; rebuild / dry-run / status outcomes;
    "not configurable, rollout direct (impact row S5)".
  - §1.3 / §8.2: the same verification applies to the lock and marker
    files themselves; a file that fails the contract is
    `environment_store_untrusted`, distinct from `environment_marker_invalid`.
  - §8.4 / §8.5: store failure is not drift (home-vs-record vs
    store-vs-record-and-boundary); new diagnostics-table row.
  - §10.1: link-target identity is necessary but no longer sufficient
    currency; new "Store-trust verification" step (boundary checks +
    recompute of the recorded §5.6 system-prompt/root-context surface
    hashes from the store entries, compared against the marker);
    fail-closed outcome (no fragment, non-current, status row); repair
    re-applies only from entries that passed (E7: repair is not
    persistence); git entries rebuild from the revalidated snapshot,
    path/local entries cannot (no second copy → `environment_repair_failed`,
    operator reinstalls); dry-run outcome.
  - §10.4: rows for `environment_store_untrusted` and
    `would-rebuild-untrusted-store`.
  - §12: `env status` store-trust row per profile (names the failing
    check); store-untrusted joins the non-current list; GC revalidates
    the boundary, retains untrusted entries, never rebuilds.
  - §13: new vector family `vectors/environments-store-boundary.json`.
- `profiles/manager.md` (§12.2/§12.5/§12.7): minimal consistency edits
  only — the three places that restated the old "link-target identity as
  sufficient currency" rule and the mirrored diagnostics/currency/GC rows.
  No new normative rule beyond citing environments §4/§8.4/§10.1/§12.
- `CHANGELOG.md`: Unreleased → Added entry "S5: …" (top of list, newest first).
- `conformance/v1/vectors/environments-store-boundary.json` (new,
  hand-authored like the S4/E1 behavioral vectors): 13 resolve cases
  (intact → fragment; swapped system-prompt/root-context bytes,
  symlinked entry root, ownership, permissions, containment escape,
  non-regular component, lock-file owner, marker symlink → untrusted;
  3 negatives), 3 dry-run cases (would-rebuild / intact-plans-nothing /
  mutates-negative), 4 repair cases (git rebuilds / path+local refuse /
  re-apply-negative), 3 status cases (current / names-check / hides-negative).
- `conformance/v1/manifest.json`, `release/1.0.0-rc.9.json`: regenerated
  (`make regenerate`) — new vector registered (1094 files), rc.9 pins updated.
- `tools/validate.py`: new production gate
  `validate_environments_store_boundary_vectors` (registered in `main()`),
  recomputing the trust verdict from the six inputs and, from it, the
  diagnostic, fragment, currency, dry-run, repair, and status-row outcomes;
  exact case inventories; negatives must still violate. Follows the
  S4/E1 gate pattern. No schema change (brief: no knob, no marker field).
- `tools/test_validate.py`: new `StoreBoundaryVectorTests` (17 tests):
  published vectors pass; 16 narrowing mutants rejected (fragment for
  swapped bytes, untrusted silence, current-for-untrusted, intact
  refusal, multi-failure, swap/shape violations, wrong check name,
  healed negative, dry-run outcome/mutation, repair re-apply,
  path rebuild, unnamed check, inventory, capability).

## Producer decisions (brief-delegated or implied)

1. Hash rule: per-surface hashes of the applied modules (the brief's
   recommended minimum), not full-entry tree hash per resolve. A
   full-entry hash would re-hash the whole skills tree on every launch —
   the exact cost §10.1 exists to avoid. Cost bound stated in §10.1:
   O(applied system-prompt + root-context module bytes) per resolve.
2. Stated bound (not a gap): a byte swap confined to the skills tree or
   MCP file that preserves the boundary is detected at materialization
   and status time (§8.4 drift), not at resolve. Named in §10.1.
3. `would-rebuild-untrusted-store` binds to dry-run *evaluation* of a
   profile entry (§4/§10.1/§10.4 + vector): revision 1 defines no
   `--dry-run` flag on profile commands, so no new command was invented
   (out of scope); the outcome completes the closed set (real op →
   rebuild; dry-run → would-report; resolve → fail closed; status →
   non-current row) for any dry-run evaluation mode.
4. Unprovisioned homes have no marker to compare against: boundary checks
   still run (`environment_store_untrusted` on failure), else
   `environment_home_stale` as before; provisioning re-verifies under
   the repair lock before writing (§10.1).
5. manager.md touched minimally for consistency (the old rule was
   restated there; leaving it would contradict §10.1) — same convention
   as the S4 landing.

## Validation transcript

All commands run from the worktree root
(`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-148pj1/worktree`).
Shell: bash. Python: project `.venv` (`uv venv` + `uv pip install -r
requirements-dev.txt`, jsonschema 4.25.1) prepended to `PATH`, because the
system `python3` has no `jsonschema`; the venv was removed after validation
(it self-ignores via its own `.gitignore`, so it never entered git state).
Go 1.26.0.

Baseline before edits (`python3 tools/validate.py`): exit 0 —
`validated 62 schemas and 1093 vector files`.

1. `make regenerate` → exit 0 (`go run ./tools/generate-vectors -root .`).
2. Regeneration proof (sha256 over all 1220 files under
   `conformance/` + `release/`, before vs after): 1218 byte-identical;
   only `conformance/v1/manifest.json` (new vector registered, 1094
   files) and `release/1.0.0-rc.9.json` (pins
   `sha256:703b4de2…fbde1512`) changed. No other vector touched.
3. Idempotence: second `go run ./tools/generate-vectors -root .` leaves
   both files byte-identical (`shasum -c` OK). (`make regenerate-check`
   itself is a committed-tree gate — it runs `git diff --exit-code` —
   so it is not quoted in this uncommitted worktree; the direct rerun
   above is the equivalent proof.)
4. `PATH="$PWD/.venv/bin:$PATH" make validate` → exit 0:
   `validated 62 schemas and 1094 vector files`; `Ran 370 tests … OK`
   (includes the 17 new `StoreBoundaryVectorTests`); `go test
   ./tools/... ok`.
5. Focused gate run: `unittest … -k StoreBoundary` → `Ran 17 tests … OK`.

## Deliberately out of scope

Implementation (`TASK-260910-32gki6`); E6 path-overlay admission
(`TASK-260916-3l60rn`) — no path-admission rule defined here; E7 launcher
config ownership; build cache / core §9.3 (unchanged); schemas (no new
knob/field per the brief); CLI flags (none added); Decision 0012 (untouched).
No commits, pushes, branches, or PRs (orchestrator owns landing).

---

# Revision 2 (rework after changes_requested, F1–F4)

Rework brief `TASK-260910-39fzpq_rework-rev2.md` implements the orchestrator's
settled design exactly; everything the rev1 verdict table marks as passing
stays byte-identical unless a correction touches it. Base unchanged
(`684c9f1`); same 8 files as rev1, no new paths, no implementation code.

## F1 — home currency split from store integrity

- Store integrity is verified against the pin, not the marker (§4, §10.1):
  for `git` the entry tree MUST hash to the resolved commit's tree
  (recomputed from bytes — git tree object identity or recorded snapshot
  tree hash), for `path`/`local` the `state_sha256`. Runs at every
  `env resolve` and before provisioning/repair of any home; missing hashes
  never pass. Cost bound: O(store entry bytes named by the lock) per
  resolve; per-surface would need the marker and misclassifies stale as
  untrusted, so the pin baseline is chosen with justification.
- Home currency is separate (§10.1, §8.4): marker surface hashes vs CURRENT
  lock (§5.1/§5.6); old-marker-after-update (§9.2) is ordinary
  `environment_home_stale` repaired from the verified store, never
  `environment_store_untrusted`. Vectors: intact-updated-store + old marker
  → stale, repair succeeds; swapped-updated-store + old marker → untrusted,
  never adopted (pin precedence).

## F2 — unprovisioned verification and marker facts

- Unprovisioned homes verified the same way (pin, no marker): intact →
  `environment_home_stale`, provision after re-verification; swapped →
  `environment_store_untrusted`, never provisioned from swapped bytes.
  Added intact/swapped unprovisioned resolve + repair cases.
- Absent vs unreadable/malformed markers are distinct facts (§8.2, §8.4,
  §8.5, §10.4): absent is unprovisioned (`environment_home_stale`);
  unreadable/malformed is `environment_marker_unreadable` (non-current,
  currency unknown), never "absent"; `environment_marker_invalid` keeps
  meaning unsupported version. New closed spelling
  `environment_marker_unreadable` appears identically in §8.2/§8.4/§8.5/
  §10.1/§10.4/§12, manager mirror, vectors, validator pins, and CHANGELOG.

## F3 — enclosing vs entry failure classes with ordering

- Order stated in §4 and §10.1: enclosing boundary → entries → pin hashes →
  home currency.
- (a) Enclosing (environments root, store root: wrong owner,
  world/group-writable, symlinked, not a directory, not contained) refuses
  every resolve and every mutating operation before its first write; nothing
  rebuilt (no protected place); posture names the boundary; operator repairs
  out of band. Dry-run reports `environment_store_untrusted` with no rebuild
  planned; GC refuses collection before its first write with nothing rebuilt.
- (b) Entry (store entry, lock, marker) inside a proven enclosing boundary
  that fails its own checks or pin hash is `environment_store_untrusted`;
  real operation rebuilds from the revalidated snapshot into newly
  established protected state (operation-private staging, atomic publication
  under the mutation lock); dry-run reports
  `would-rebuild-untrusted-store`; resolve fails closed until rebuilt. GC
  retains untrusted entries, never collects, never rebuilds.
- Vectors: environments-root and store-root enclosing cases (resolve,
  dry-run no-rebuild, repair refuses-no-rebuild, status names-boundary) and
  the entry-rebuild case (boundary failure on a git entry rebuilds).

## F4 — validator pins scenarios

- `tools/validate.py`: each named case pinned to its discriminating inputs
  (object, check among ownership/permissions/containment/regular_types/
  link_safety/pin_hash/home_currency, home, surface; plus entry_kind for
  repair). `s5_check_pin` refuses a corpus where a named branch no longer
  exercises its check. Diagnostics pin extended with `home_stale` and
  `marker_unreadable`. Inputs renamed `surface_hash_match` →
  `pin_hash_match`; `home` required in every family.
- `tools/test_validate.py`: `StoreBoundaryVectorTests` grows 17 → 25. The
  reviewer's five-to-one narrowing (four boundary branches collapsed to
  ownership-only with `failing_check` updated) is added as
  `test_five_to_one_narrowing_is_rejected`, plus object-swap, stale-vs-
  untrusted (both directions), unprovisioned-swap, unreadable-vs-absent,
  enclosing dry-run, and enclosing-rebuild negatives.
- Vectors grow 13/3/4/3 → 20/4/10/4 (resolve/dry-run/repair/status); the
  S5 file is still the only new vector; all pre-existing conformance files
  byte-identical (see proof).

## What changed per file (rev2 deltas vs rev1)

- `protocol/environments.md`: §4 pin baseline + two failure classes + order;
  §8.2 absent/unreadable split + `environment_marker_unreadable`; §8.4
  pin-vs-record wording + currency sentence + absent/unreadable paragraph;
  §8.5 table split (unreadable row, invalid narrowed, untrusted → pin hash);
  §10.1 repair split (enclosing refuses / entry rebuilds with staging,
  dry-run split) + store-trust verification rewritten (order, pin
  recomputation, currency split, two reviewer cases, absent/unreadable,
  O(entry bytes) bound); §10.4 six rows (marker_unreadable, untrusted entry,
  untrusted enclosing, would-rebuild entry-class, untrusted enclosing
  dry-run); §12 posture names boundary + GC split; §13 vector inventory
  rewritten. §1.3 lock paragraph untouched (still passing).
- `profiles/manager.md`: mirrored tables (§12.2 marker split, pin hash;
  §12.5 resolve + diagnostics rows; §12.7 posture boundary + GC split).
- `CHANGELOG.md`: S5 entry rewritten for pin baseline, currency split,
  failure classes, `environment_marker_unreadable`, and new vector inventory.
- `conformance/v1/vectors/environments-store-boundary.json`: revised to
  pin-hash inputs, home states, 38 cases (20/4/10/4) as above.
- `conformance/v1/manifest.json`, `release/1.0.0-rc.9.json`: regenerated
  (same 1094-file registration, updated S5 digest + rc.9 pins).
- `tools/validate.py`: S5 gate rewritten with scenario pins, order, and
  split outcomes (see F4).
- `tools/test_validate.py`: S5 tests 17 → 25 (see F4).

## Validation transcript (rev2)

All commands run from the worktree root
(`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-148pj1/worktree`).
Shell: bash. Python: shared venv
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin`
prepended to `PATH` (the path the review brief names). Go toolchain as installed.
Headless run: no single shell call exceeds its time bound; long verification
is split into bounded sequential calls. What was rerun vs accepted is stated
explicitly below; every quoted exit code is the gate command's real status
(no `tee`, no pipe masking beyond `tail` for display).

Reran (this revision, exit codes quoted):

1. `PATH=".../.temp/venv/bin:$PATH" python3 tools/validate.py` → exit 0:
   `validated 62 schemas and 1094 vector files`.
2. `PATH=".../.temp/venv/bin:$PATH" go test ./tools/...` → exit 0:
   `ok github.com/relux-works/curator-spec/tools/generate-vectors 3.472s`.
3. S5 gate standalone
   `PATH=".../.temp/venv/bin:$PATH" PYTHONPATH=tools python3 -c "import validate as v; v.validate_environments_store_boundary_vectors(); print('S5 gate ok')"`
   → exit 0: `S5 gate ok`.
4. `PYTHONPATH=tools python3 -B -m unittest tools.test_validate.StoreBoundaryVectorTests -v`
   → exit 0: `Ran 25 tests in 0.048s OK` (17 rev1 tests updated for
   pin-hash inputs + 8 new F1–F4 narrowing tests).
5. Subsets (bounded, all exit 0):
   - `WireSemanticValidationTests AssuranceRelationalValidationTests` →
     `Ran 21 tests in 0.696s OK`
   - `RepositoryDescriptorIdentityTests ManagerLifecycleValidationTests SharedFixtureMarkerTests WorkflowRegenerationScopeTests` →
     `Ran 22 tests in 24.212s OK`
   - `EnvPassthroughVectorTests` → `Ran 27 tests in 49.382s OK`
   - `ContextVersionVectorTests` → `Ran 15 tests in 19.993s OK`
   - `discover -s tools -p 'test_implementation_coverage.py'` →
     `Ran 36 tests in 2.119s OK`
   - `discover -s tools -p 'test_verify_release_*.py'` →
     `Ran 10 tests in 0.003s OK`
   - `discover -s tools -p 'test_release_gate.py'` →
     `Ran 32 tests in 234.392s OK`
   Together: 188 unittest tests rerun green, plus the S5 gate and the two
   fast `make validate` gates. No unittest failure was observed in any
   subset that completed.
6. `go run ./tools/generate-vectors -root .` → exit 0 (twice).
   Idempotence: shasums of `conformance/v1/manifest.json`
   (`b42b5d7b…37b7a71`) and `release/1.0.0-rc.9.json`
   (`b963a511…ac07ac3b6`) identical before/after the second run
   (`diff` exit 0). (`make regenerate-check` is a committed-tree gate
   (`git diff --exit-code`), so it is not quoted in this uncommitted
   worktree; the direct rerun is the equivalent proof, as in rev1.)
7. `git diff HEAD --check` → exit 0.
8. Scope: `git status --short` lists only the same 8 files as rev1
   (spec ×2, CHANGELOG, S5 vector, manifest, rc.9, validator, validator
   tests); `git diff HEAD --name-only -- conformance/ ':!manifest'
   ':!S5-vector'` → exit 0 (all pre-existing conformance files
   byte-identical); `git diff HEAD --name-only -- release/
   ':!release/1.0.0-rc.9.json'` → exit 0 (only rc.9 pins move).
   No schema, implementation, E6, E7, or proposal 0014–0018 edits.

Accepted from already-attached evidence (not rerun here): the remaining
`test_validate.py` classes untouched by this revision (BuildDriver golden
suite, Environment, SourceSigners, ContextDetector, SnapshotAcquisition,
ShellHookTrust, UmbrellaProvider, ManagerConfig, SystemConfigV2,
RegistryPageBoundary). They passed in rev1 (`Ran 370 tests … OK`,
producer evidence) and independently in review
(`Ran 370 tests in 383.479s OK`, verdict rev1); their gates and vectors
are byte-identical here, and `tools/validate.py` (which runs every vector
gate, including theirs, on the published corpus) exits 0 above. The full
single-call `make validate` was therefore not re-attempted headless; the
reviewer re-runs it independently.

## Deliberately out of scope (unchanged)

Implementation (`TASK-260910-32gki6`); E6 path-overlay admission
(`TASK-260916-3l60rn`); E7 launcher config ownership; build cache / core
§9.3 (unchanged); schemas (no new knob/field; diagnostics only); CLI flags
(none added); Decision 0012 (untouched). No commits, pushes, branches, or
PRs (orchestrator owns landing). No LOGBOOK.md edits (campaign rule 7).
