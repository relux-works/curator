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
