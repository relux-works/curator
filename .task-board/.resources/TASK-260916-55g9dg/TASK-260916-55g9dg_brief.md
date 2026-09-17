# Brief — TASK-260916-55g9dg: direct-only `class: system` modules (E2, manager)

Story `STORY-260916-2d9coh` (direct-only-system-modules), wave 1.
Rules: `remediation-manager-producer-rules.md` (attached). Role: developer.
This is the story's only implementation task.

## Spec (read first; it is the contract)
curator-spec `0da4020` (landed by PR #61): `protocol/environments.md` §3
(admission rule: direct = root, active overlay, or a package named by their
`requires.contexts`; transitive = everything else), §5.5/§5.7 (diagnostics
`context_system_module_dropped` warning, `context_system_module_transitive`
resolution error), §12.1 (knobs `transitive_system_modules` = `drop` |
`error`, default `drop`; `system_module_waivers` = list of `{ package, reason }`,
default empty), §12.2 (`transitive_system_modules` lockable to `error` only;
waivers not lockable), §12 status posture, §13; `profiles/manager.md` §1 lock
set; vectors `conformance/v1/vectors/environments.json` (system-module
admission cases: direct / transitive-drop byte-identical / transitive-error /
waived / overlay-direct) and schema cases `manager-config-v2` /
`system-config-v2` (`invalid-transitive-system-modules-value`,
`valid-system-module-waiver`, waiver negatives, `invalid-transitive-system-
modules-drop-direction`). Audit context: `docs/security-audit-2026-09.md` and
`verify-e-findings-rev4.md` on `TASK-260916-dv7xv5` (E2: materialization
applies every closure package's system modules at
`internal/contextmaterialize/contextmaterialize.go:248-265`, called from
`internal/envprofile/managed.go:1785`; `contextaudit.go:109-115` blocks only
Findings; `envprofile.go:1138-1139` surfaces `SystemModules` as warnings).

## Deliverable
1. **Config knobs** (`internal/config/environments.go`, `config.go` lockable
   set, `write.go`): parse and validate `transitive_system_modules` (closed
   enum, default `drop`) and `system_module_waivers` (closed object list,
   package identifier grammar, non-empty reason, default empty); system-file
   locking of `transitive_system_modules` accepts `error` only and refuses
   `drop`; `system_module_waivers` is not lockable. Make the curator
   `manager-config-v2` / `system-config-v2` schema-case tests consume the new
   cases from `CURATOR_CONFORMANCE_ROOT` (root-content skip when absent).
2. **Admission in resolution/materialization**: compute the direct set
   (root, active overlays, packages named by their `requires.contexts`) and
   the waived set; under `drop` skip every transitive `class: system` module
   at materialization with the warning `context_system_module_dropped`
   naming package and module, keeping the materialized bytes exactly the
   admitted modules' bytes; under `error` fail resolution with
   `context_system_module_transitive` naming package and module, writing or
   changing no lock. `context-system-module-present` stays the always-warn
   finding over every member; the launch-fragment
   `works.relux.curator.system-modules` flag follows the ADMITTED set so `ax`
   resume refuses on drift.
3. **Posture**: `curator env status` reports the effective
   `transitive_system_modules` value and every dropped system module by
   package and path; a drop warning never makes a row non-current (§12).
4. **Vectors**: the environments materialization test consumes the five
   admission cases (byte-exact expected outputs) from the root; unit tests
   for direct/transitive/waived/overlay classification, the `error` refusal
   leaving the lock unchanged, and the lock-direction rule.
5. `CHANGELOG.md` Unreleased: "E2: … default `drop` (non-breaking), `error`
   opt-in" — no warn-first split applies (direct, non-breaking per the spec).

## Out of scope
Spec edits; E1 signer rules; pi `SYSTEM.md` channel changes; SPEC_PIN.

## Handoff
Narrow validation per the rules (`go build`, `go vet`, `gofmt -l`, `go test`
for `./internal/config/... ./internal/contextresolve/...
./internal/contextmaterialize/... ./internal/contextaudit/...
./internal/envfragment/... ./internal/envprofile/... ./cmd/curator/...` with
the conformance root set). Attach `TASK-260916-55g9dg_results.md`, tick the
checklist, then `task-board handoff TASK-260916-55g9dg --role developer`.
