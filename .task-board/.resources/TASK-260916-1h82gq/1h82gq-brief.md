# TASK-260916-1h82gq brief (orchestrator, binding) — R2 capabilities + mandatory portable controls

Story STORY-260822-2h0v9j; builds on R1 (TASK-260916-3dmjbc, checkpointed on the Story branch:
`internal/scriptworker`, `internal/config` script_interpreters, `internal/scriptpolicy` preflight
with the implemented-controls table, hidden worker mode in `cmd/curator`). Read R1's results.md
(all revisions) and the rev2/rev4 review verdicts first: residuals R-A…R-J are your input.
Spec at the CI pin 87a0d006: `profiles/manager.md` §3.1 (eight-step order; the eleven mandatory
controls; the reserved environment-name set incl. per-platform and per-interpreter names),
Protocol Core §4.1.1/§4.3; vector `conformance/v1/vectors/script-host-execution-policy.json`
keys `mandatory_controls` (11), `capability_derivation_cases` (4), `execution_policy`,
`interpreters`, `opt_in_cases` (6).

## Scope (R2 exactly)
Connect the declaration-derived containment profile to the R1 worker path and implement the
mandatory portable controls that R1 left "unavailable":
- `manager-built-environment`: empty bootstrap + manager-set values + exactly the host names in
  `env_read`; every reserved name (all platforms; macOS `DYLD_*`; Windows set,
  case-insensitive; `python3-v1` `PYTHON*`/`__PYVENV_LAUNCHER__`; `node-v1` `NODE_*`/`NPM_CONFIG_*`)
  never passes through even if named by `env_read`; an interpreter identifier without a
  reserved set is refused (spec).
- `manager-built-path`: exactly the resolved interpreter + the resolved declared `exec` names;
  inherited PATH discarded; `exec` names manager-resolved (derivation case
  `declared-exec-is-manager-resolved`), never from package data.
- `offline-network-configuration`: when derived `network` is `none`, offline configuration plus
  proxy and resolver scrubbing (the reserved proxy/resolver names); declared network hosts are
  reporting-only (`declared-network-hosts-are-reporting-only`) — do not promise host filtering.
- `operation-private-runtime-area` APPLIED (R-F): manager-selected working directory and
  operation-private tmp/config/cache roots bound via `TMPDIR/TEMP/TMP/XDG_*` (per platform).
- `declared-secrets-remain-identifiers`: secrets are identifiers, never injected values.
- `all-fields-absent-deny-by-default`: every absent capability field = deny.
Also: R-A — the production launcher: installed enforced shims route through the manager
(native launcher, "no shell, `.cmd`, or symlink shim" between manager and interpreter — see the
verdict R-A note); R-C — add the `permit` frame so the parent validates the worker's proof (and,
from R3, the evidence record) BEFORE the interpreter runs; R-G — replace the
`activeScriptCommands` `%w(<nil>)` guard shape. Stream model R-B (live stdin/pass-through
binding) is R3-or-later unless the launcher cannot work without it — say which.

## Rulings
R1 Admission honesty continues: after R2 the implemented-controls table marks the controls
above as implemented; `inventory-controls-applied` and
`closed-script-capability-evidence-record` stay "unavailable" until R3, so install/invoke STILL
refuse `script_execution_control_unavailable` on every production entry — no enforced launch
ships uncontained. The worker path with all R2 controls is proven at the process boundary and
through the launcher with the preflight table forced complete only in tests via the SAME
production predicate (a table injection seam that production never sets — document it and add
a row proving production cannot set it).
R2 Deny by default everywhere; package data can never widen the profile (row: manifest attempts
to add env/exec/network → refused or ignored per spec, say which per field).
R3 Evidence: all four derivation cases + all eleven mandatory controls driven at the production
entry (`install.Project`/CLI + worker process boundary) with a narrowing mutant per control
(e.g. pass one reserved name through; keep inherited PATH; skip proxy scrub when network=none;
skip runtime-area binding; …) — each killed by a named row; Windows rows run on windows-latest
(case-insensitive reserved names proven there); ledger rows registered; the six opt_in_cases
keep their classification.
R4 No behaviour change for declared-only/build/go-v1; CHANGELOG `### Added` extended; docs
(troubleshooting + a config reference entry for `script_interpreters`, R-I).
R5 Large leaf: publish when complete; if budget runs out, publish an honest partial with
results.md listing done/undone (the orchestrator continues you).
results.md: design (environment builder, PATH builder, offline config, runtime area binding,
launcher, permit frame), row table with production boundary per row, mutant table per control,
ratio line (vector keys consumed), Windows proof status, bounds.
