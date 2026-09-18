# Brief — TASK-260917-16l2md: promote SPEC_PIN to v1.0.0-rc.12 with the wave-1 manager union

Story `STORY-260917-3w3lvj` (conformance-pin-rc12-promotion), `EPIC-260910-2hw1xb`.
Rules: `remediation-manager-producer-rules.md` (attached) — with ONE
sanctioned exception for this task only: you DO change `SPEC_PIN`. Role:
developer. Worktree: the managed Story worktree
`<control-root>/.temp/STORY-260917-3w3lvj/worktree` (forked from current
curator `main`).

## Why this task exists
`internal/config TestManagerConfigV2Vectors` renders every published
`manager-config-v2` case, so the hosted-gate pin can only advance to a root
whose knobs the code implements completely, and the three wave-1 manager
implementations (E2, E4, S4) each fail the gate at the old root because of
the others' knobs. The operator decided (2026-09-17): tag curator-spec
`v1.0.0-rc.12` at `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`, qualify it
(`TASK-260917-2ecpjv`, running in parallel), and land the pin move together
with the union of the three candidates. E1 and wave 3 follow in rc.13.

## Inputs (attached)
- `E2-candidate.patch` = `TASK-260916-55g9dg` revision-3 worktree state (base
  `1de6f8e`): transitive `class: system` modules → `transitive_system_modules`
  drop|error, waivers, `context_system_module_transitive`, posture. Its rev-2
  review (`TASK-260916-55g9dg_review-verdict-rev2.md`) demanded: no
  `prunePostRevisionKnobs` workaround, a null waiver list is refused, a
  root-content driver test; rev 3 (`TASK-260916-55g9dg_rework-rev3.md`,
  `_results.md`) claims closure — verify while merging.
- `E4-candidate.patch` = `TASK-260916-3oh0u8` revision-1 worktree state (base
  `1de6f8e`): provider resolution from the install dir + `provider_directories`,
  revision A warning / B refusal, resolved path in `env status`.
- `S4-candidate.patch` = `TASK-260910-gocke2` revision-1 worktree state (base
  `c64ceaf`): `passable_env_names` default `[]` (explicit `null` = unbounded),
  `mcp_package_allowlist_empty` warning, `s4-warn`/`s4-enforce` profiles,
  MCP surfacing rows at install/update/status.
- Each candidate's brief, results and gate-failure notes are attached on its
  own task; the landed spec text is curator-spec `dced9b8` (checkout at
  `/Users/administrator/Developer/ReluxWorks/curator/curator-spec` is at
  `684c9f1`, which includes it; for byte-exact vector expectations use the
  root of the pin: `git -C curator-spec worktree add /tmp/spec-rc12 dced9b8317e0e8af79edf2d0539b32bd22b6c85b`
  and `export CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc12/conformance/v1`).

## Deliverable
1. Apply the three candidate patches onto the Story worktree (`git apply
   --3way`), resolving overlaps as the UNION — they meet in
   `internal/config` (knob parsing/effective JSON/validation, defaults
   table), `cmd/curator` status/env status rows, `internal/envprofile`,
   `.github/ci/platform-cases.tsv`, `CHANGELOG.md`, docs. Keep every closed
   diagnostic/knob spelling exactly as the spec tables have it. Where a
   candidate registered a `root-content` skip row for a family the rc.12
   root now publishes, remove the row and the skip (the family is served).
2. `.github/workflows/ci.yml`: `SPEC_PIN: dced9b8317e0e8af79edf2d0539b32bd22b6c85b`
   (every job that checks out the suite uses `${{ env.SPEC_PIN }}` — confirm
   no second pin exists; the launcher-facing `release.yml` is out of scope).
3. Green at the new root: `go build ./... && go vet ./... && gofmt -l internal cmd`,
   then `go test -count=1 ./internal/config/... ./internal/contextresolve/...
   ./internal/contextmaterialize/... ./internal/contextaudit/... ./internal/shell/...
   ./internal/envprofile/... ./cmd/curator/...` with the rc.12 root; every
   vector-driven test (manager-config-v2, environments, umbrella-provider-
   resolution, environments-env-passthrough, shell-hook-trust) executes, no
   root-content skip for a published family, no weakened assertion. Do NOT
   run the full landing suite locally; the runtime runs
   `scripts/remote-gate.sh` at handoff and its gate self-test enforces the
   ledger rows.
4. `CHANGELOG.md` Unreleased: the three entries (E2/E4/S4, each naming the
   warn-first step shipped) plus a "Conformance pin → v1.0.0-rc.12" note;
   `docs/` where the candidates documented knobs.
5. `TASK-260917-16l2md_results.md`: per-candidate merge notes (what
   conflicted, how the union was chosen), the E2 rev-2 corrections'
   closure evidence, transcripts, and any behaviour you had to reconcile
   between candidates (report; do not redesign).

## Out of scope
E1 manager (`TASK-260916-1zgucp`), R1 client (`TASK-260910-2n0233`), the S6
story (`STORY-260910-2awkzu`, landing separately), spec edits, tags/releases,
`release.yml`.

## Checklist and handoff
Tick the checklist items you satisfy; hand off with
`task-board handoff TASK-260917-16l2md --role developer`.
