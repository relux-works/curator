# TASK-260922-1t2w1q brief (orchestrator, binding) — F-C2 explicit credential migration (0017)

Story STORY-260922-1cenbr; builds on F-C1 (TASK-260922-1t551d, checkpointed on the Story branch:
fix-first repairs, mis-targeted vs dangling-to-declared states, Pi agent root, TOML-parsed codex
admission). Read F-C1's results.md and review verdicts first. Normative source: curator-spec main
05053cd7 — environments §7.4/§10.1 ("never-silent migration"), manager §12.4/§12.5, decision 0017
choice 3 (Pi roots), option O3 (inspect → plan → apply under the manager lock, never inside
`resolve --repair`, no secret copies), C2 evidence (operator's dangling Pi link).

## Scope
An explicit `env` migration step (name it consistently with the existing CLI grammar, e.g.
`env migrate --inspect|--plan|--apply` or the shape the spec's §10 text implies — cite it) that:
- INSPECT: inventories, per profile/environment, the old marker, every recorded passthrough link and
  its current target, and BOTH Pi roots (`~/.pi/auth.json`, `~/.pi/agent/auth.json`), classifying
  each entry (current / mis-targeted / dangling-to-declared / regular-file-at-link / foreign);
  read-only, no lock needed beyond consistency, never reads credential bytes;
- PLAN: computes the exact operations (relink recorded symlink → declared native path; keep
  effective mode on upgrade; nothing for pending entries) and PRINTS them before any change;
  conflicts that need the operator (isolated→shared account choice, regular file at a link path,
  two live Pi credentials) are listed with the exact decision needed and block APPLY;
- APPLY: under the manager-home mutation lock, journaled and rollback-safe (temp + atomic rename
  as the marker publication already does), executes exactly the printed plan; refuses if the
  inventory changed since the plan (hash the plan); never copies, moves or rewrites credential
  bytes (a symlink relink preserves the native file untouched).
Also: `resolve --repair` must NOT perform migration silently — it may report "migration needed"
and point at the explicit step (row).

## Rulings
R1 No secret copies at any step (row asserts byte-identity of native files before/after and that
no new file with credential content appears anywhere). R2 Fail closed on ambiguity (two candidate
native files, unexpected target, regular file) — refuse naming the operator choice. R3 Preserve
the effective mode on upgrade; the plan is deterministic and printed; apply refuses on plan drift.
R4 Frozen v1 marker untouched. R5 Evidence: production-entry rows through the CLI on temporary
stores: the operator's Pi case (wrong-target home → `~/.pi/agent`, bytes preserved, mode intact,
plan printed before apply); conflict cases refuse with the named choice; `resolve --repair` does
not migrate; rollback on an injected apply failure restores the prior marker; mutants (skip the
drift check; copy instead of relink; migrate inside repair) each killed. Windows: symlink privilege
pattern of the sibling rows. CHANGELOG `### Added`, docs/cli.md + troubleshooting. results.md:
design, row table, mutant table, bounds (what F-C3 adds). Publish only on a green gate.
