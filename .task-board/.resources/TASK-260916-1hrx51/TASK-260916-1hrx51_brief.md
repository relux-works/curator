# Brief — TASK-260916-1hrx51: spec admission rule for `class: system` modules (E2)

Story `STORY-260916-2d9coh` (direct-only-system-modules), wave 1.
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-2d9coh/worktree`.
Rules: `remediation-spec-producer-rules.md` (attached). Role: doc-writer (technical writer) of
normative spec text; no implementation.

## Finding (read it first)
`docs/security-audit-2026-09.md` E2 and Appendix B (confirmed:
`internal/contextmaterialize/contextmaterialize.go:248-265` applies every
package's system modules; the only control is the always-warn finding
`context-system-module-present`, `contextaudit.go:22`). A `class: system` module
replaces or appends the tool's system prompt; any package in the closure may
carry one. Text to revise: `protocol/environments.md` §2 (context package shape)
/ §3 (context modules, the `class` field), §5.5 (system-prompt output), §12.1
knobs, §12.2 lockable set, §13 conformance surfaces; the story README carries
the operator's refined decision.

## Settled decisions (do not reopen)
- New closed §12.1 knob `transitive_system_modules` = `drop` | `error`,
  default `drop`; §12.2-lockable in the direction of `error` only (a strict
  machine locks it to `error`).
- `drop` (default, non-breaking): a `class: system` module carried by a package
  that the root (or an overlay) does NOT name directly is skipped at
  materialization with a warning naming package and module; root/direct
  modules still materialize; installs never break.
- `error` (strict machines): the same module is a resolution error
  `context_system_module_transitive` naming package and module.
- Admission: naming the package directly in the root's `requires` (or an
  overlay's) admits its system modules; additionally a per-package machine
  waiver admits a transitive package's system modules — a closed §12.1 knob
  (working name `system_module_waivers`, list of `{ package, reason }`,
  default empty, not lockable in the admitting direction).
- `context-system-module-present` stays the always-warn finding; the fragment
  flag `works.relux.curator.system-modules` (launch fragment §10.2) is kept so
  `ax` resume refuses on drift — state the interaction in one sentence.
- Impact row "E2 transitive system modules": direct, default non-breaking —
  no warn-first split is needed for `drop`; `error` is opt-in. Say so.

## Deliverable
1. §3 (or wherever `class: system` is defined): the admission rule with the
   direct / transitive definition ("direct" = the package is named by the root
   or an active overlay `requires` entry; everything reached only through
   another package's `requires` is transitive).
2. §5.5 / §5.7 diagnostics: the drop warning (working spelling
   `context_system_module_dropped`, naming package and module) and the
   `context_system_module_transitive` error; where they are reported
   (materialization, `env status` posture row listing dropped modules).
3. §12.1 rows `transitive_system_modules` and `system_module_waivers`; §12.2
   lockable entry for `transitive_system_modules` (direction rule).
4. Conformance vectors under `conformance/v1/` (+ manifest): direct package →
   materialized; transitive + `drop` → skipped with warning, bytes of the
   materialized output unchanged by the dropped module (byte-exact vector);
   transitive + `error` → refusal; transitive + waiver → materialized;
   schema cases for the two knobs in the machine-config schema if one exists
   under `schemas/v1/` (follow the existing pattern).
5. `CHANGELOG.md` Unreleased entry "E2: …".

## Out of scope
Implementation (`TASK-260916-55g9dg`), E1 signer rules (`STORY-260916-ioemse`),
the pi `SYSTEM.md` channel semantics beyond citing them.

## Checklist and handoff
Tick the checklist items you satisfy; attach
`TASK-260916-1hrx51_change-request_rev1.patch` and `TASK-260916-1hrx51_evidence.md`
(with the `make validate` transcript), then
`task-board handoff TASK-260916-1hrx51 --role doc-writer`.
