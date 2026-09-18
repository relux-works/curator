# Rework brief — TASK-260910-2qtiho, revision 2 (S1 + S3)

Revision 1 was rejected with two corrections
(`TASK-260910-2qtiho_review-verdict-rev1.md`). Everything else passed; keep it
byte-identical.

- **F1 — pin the posture in every required scenario (high; rule 7).** The
  `security-posture.json` gate pins the MCP refusal only to operation +
  declarations + absent environments, not to the effective posture: the
  case rewritten under `security_posture: permissive` (warning, `proceeds`)
  still passes under its negative name; the exhaustive substitution probe
  also accepted `revision-A-default-permissive-status` ←
  `locked-value-beats-explicit` and `revision-B-default-hardened-flip-install`
  ← `schema1-machine-is-permissive`. Pin every required scenario to its
  schema version, effective posture, precedence conditions (explicit value,
  lock, profile default, absence of an overriding lock) and the inputs that
  distinguish its branch; add same-name internally consistent replacement
  negatives for the three demonstrated holes (the reviewer's probe is
  attached as a task outcome — reuse its shape).
- **F2 — the full gate inventory on BOTH status commands (medium).**
  `profiles/manager.md` closes `curator status` to four posture rows ("No
  other `curator status` posture row exists") while environments §12
  prescribes eight rows, so neither command carries the twelve-gate
  inventory the brief settled. Require the complete applicable inventory on
  both `curator status` and `env status` when the environments capability
  exists (hook trust A/B, passthrough s4-warn/s4-enforce, transitive system
  modules, provider trust roots revision, signer verification, update
  confirmation revision, store boundary, write discipline, registry policy,
  audit mode, source allowlist size, MCP allowlist size, plus the profile in
  force) with one shared closed vocabulary and provenance values (`profile`,
  `explicit`, `lock`, `shipped`); keep the schema-1 / no-environments case
  explicitly bounded to the manager rows; pin both command outputs in the
  vectors.

## Validation and handoff
`make validate` and the regeneration proof (exit codes); evidence "Revision
2" section; `TASK-260910-2qtiho_spec-patch_rev2.patch` = `git diff HEAD` of
the worktree (base `e8b53a0`) with new files via `git add -N`; EMPTY curator
delta; `task-board handoff TASK-260910-2qtiho --role doc-writer`. Worktree
and rules unchanged.
