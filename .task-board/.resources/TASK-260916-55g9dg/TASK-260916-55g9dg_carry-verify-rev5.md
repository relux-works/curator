# TASK-260916-55g9dg — Revision 5 carry-forward verification (STOP: stale worktree, do not publish)

Date: 2026-09-27. Run: bound developer carry-forward republish per `55g9dg-carry-5.md`.
Refresh: `task-board worktree refresh-candidate TASK-260916-55g9dg` → `refresh_advanced`,
TrunkOID `97ca33704e7f2bca27e457ca78323c6faae97034` (exit 0).
Previous base: `bd3c0f436226ec831c13b3a307c30acd035c73c2`.
Accepted candidate: `TASK-260916-55g9dg_change-request_rev4.patch` (8 paths, no CHANGELOG path).
Review verdict rev4 (`TASK-260916-55g9dg_review-verdict-rev4.md`): **accepted** on content.
Per the carry instruction, content is not in question and no production edits were made in this run
(`git diff` for the 8 candidate paths below is the preserved rev4 delta only; no new code, no new tests).

## 1. Lifecycle

`task-board m 'set_status(TASK-260916-55g9dg, status=development)'` → exit 0
(`status_changed` development, story promoted to development).

## 2. Per-path verification (rev4 patch vs worktree)

Trunk range `bd3c0f43..97ca3370` landed STORY-260923-1lu2o3 (`1a57c71c`,
fragment-permissions v2 + `environments.permissions`) which intersects 5 of the 8 rev4 paths.
The 3 paths trunk did NOT touch were proven byte-identical to rev4 by applying each path's
patch section onto its old-base bytes in /tmp and `cmp` against the worktree file (all exit 0):

| Rev4 path | Trunk touched? | Result |
|---|---|---|
| `internal/config/system_module_schema_test.go` | No (`git diff bd3c0f43 HEAD -- <path>` empty, exit 0) | IDENTICAL to rev4 post-image (`cmp` exit 0) |
| `internal/envprofile/admission_test.go` | No (same, empty) | IDENTICAL to rev4 post-image (`cmp` exit 0) |
| `internal/envprofile/status.go` | No (same, empty) | IDENTICAL to rev4 post-image (`cmp` exit 0) |
| `internal/config/environments.go` | YES (35+/8 trunk) | Worktree keeps rev4 E2 knobs (13 `transitive_system_modules\|system_module_waivers` hits) but has **0** `ermissions` lines vs 17 in `HEAD:internal/config/environments.go` — trunk `permissions` feature dropped, NOT both sides |
| `cmd/curator/env_test.go` | YES (v1→v2 1-liner) | Worktree line 61 asserts `launch-env-fragment-v1`; trunk asserts `v2` — trunk change reverted, NOT both sides |
| `internal/config/environments_test.go` | YES (43+/5 trunk) | Worktree has 3 `ermissions` hits; trunk's permissions grammar/lock tests absent — trunk additions dropped, NOT both sides |
| `internal/config/environments_conformance_test.go` | YES (17+/25 trunk) | Worktree 16 `ermissions` hits vs HEAD 7 — diverged ledgers, not a both-sides merge |
| `.github/ci/conformance-gaps.tsv` | YES (28+/28 trunk) | Worktree 36/36 vs trunk 28/28 — rev4 ledger preserved over trunk's re-histogrammed ledger, not combined |

## 3. CHANGELOG policy

`git diff HEAD --stat -- CHANGELOG.md` → empty (exit 0): the file equals trunk `97ca3370`.
No task hunk to revert; nothing was changed. Entry text preserved verbatim for release prep:

> E2: restrict system-class context modules to direct packages and explicitly waived packages. The default `transitive_system_modules=drop` is non-breaking and warns with `context_system_module_dropped`; `error` is opt-in and refuses resolution with `context_system_module_transitive`. Add `system_module_waivers` for reviewed package exceptions.

(Retyped from `TASK-260916-55g9dg_results.md:131-133`; `results.md` itself was not modified in this run.)

## 4. Stale-snapshot gate (CARRY STEP 4) — FAIL, STOP

`git diff --name-only HEAD -- . ':!.task-board'` (exit 0) lists, beyond the 8 rev4 paths,
trunk-only content the worktree reverts — so this tree MUST NOT be published:

- `internal/envfragment/envfragment.go` — worktree reverts v2→v1 (drops `Permissions`, `Version = "launch-env-fragment-v2"`)
- `internal/envfragment/envfragment_test.go`, `fragment_schema_test.go` — reverted
- 20 deleted `internal/envfragment/testdata/curator-spec-main/...launch-env-fragment-v2...` files (D entries)
- `internal/config/config.go`, `cmd/curator/env.go`, `cmd/curator/env_permissions_test.go` (D),
  `cmd/curator/profile_install_matrix_test.go` (D), `docs/environment-config.md`, `go.mod`, `go.sum`,
  `internal/envprofile/managed.go`, `internal/envregistry/envregistry.go`,
  `internal/install/draftsources_test.go`, `internal/registry/registry_test.go`,
  `.github/ci/platform-cases.tsv`, `.github/ci/skip-classes.tsv`

Cause: `refresh-candidate` advanced the branch OID to `97ca3370` but the working tree kept
old-base bytes on every trunk-touched path (before refresh the tree was dirty on exactly the 8
rev4 paths; after refresh the same bytes read as reverts against the new HEAD). Repair requires a
real converge carrying the rev4 delta onto trunk with the `permissions`/fragment-v2 hunks merged —
beyond this run's "change nothing else" mandate. Per step 4: STOP and report, do not publish.
No handoff was performed; no production file was edited.

## 5. Focused tests — NOT RUN

`go test ./internal/contextmaterialize ./internal/contextresolve -count=1` was deliberately not run:
the tree reverts trunk's landed permissions/fragment-v2 behavior, so any transcript would
misrepresent the publishable candidate. Rev4's accepted validation transcripts stand in
`TASK-260916-55g9dg_results.md:70-104`; the bounded run belongs to the re-converged successor.

## 6. Review-round note

The round brief asks for a named regression test plus narrowing mutant answering the rev4 verdict.
Revision 4's verdict is an **acceptance** (`review-verdict-rev4.md`), and the carry instruction
freezes content ("everything it marks as implemented/passing stays; change nothing else"), so no
new test or mutant was added in this run. The rev4 evidence (8/8 narrowing mutants killed,
`results.md:106-121`) is unchanged and unmodified.

## Recommendation

Orchestrator: re-converge properly (carrying the 8-path rev4 delta onto `97ca3370` with trunk's
`permissions`/fragment-v2 hunks merged on the 5 intersecting paths), then route a successor
developer run for the bounded tests + handoff. This tree must not be published as revision 5.
