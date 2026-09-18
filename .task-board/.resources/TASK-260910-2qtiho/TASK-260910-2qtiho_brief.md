# Brief — TASK-260910-2qtiho: hardened-defaults profile and the advisory residual (S1 + S3)

Story `STORY-260910-2qmrb8` (secure-defaults-for-audit-gates), wave 3 of the
2026-09 security-audit remediation — deliberately last of the wave so the
posture report covers every gate landed so far (S6 hook trust, S4 passthrough
bounds, E2 transitive system modules, E4 provider trust roots, E1 signer
allowlist and update confirmation, S5 store boundary, E5 write discipline).
Manager task `TASK-260910-1sapuy` (install-time unreachable-registry notice)
follows.
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-2qmrb8/worktree`
(branch `task-board/story/STORY-260910-2qmrb8`, forked from curator-spec `main` `e8b53a0`).
Rules: `remediation-spec-producer-rules.md` (attached). Role: doc-writer of
normative spec text; no implementation.

## Findings (read them first)
`docs/security-audit-2026-09.md` S1 (Medium, systemic): registry §4 blocks
unknown artifacts only under strict policy; core §6.1 an empty
`allowed_sources` permits every network identity; environments §2.2/§12.1 an
empty MCP package allowlist permits all declaration sources; `audit.mode`
and `audit.registry_policy` default `advisory` (`manager-config-v1`
`$defs.audit`); a hostile repository plus `curator install` passes with zero
blocking gates under defaults. S3 (Medium): "Unreachable registries warn and
contribute no record" — under advisory policy a network attacker suppresses
a revocation; the 7-day offline grace widens the window; the composed
residual is never stated. Text to revise: `protocol/registry.md` §4 (and §8
cache/offline where the grace is defined), `SECURITY.md` ("Security model",
"Registry security state"), `profiles/manager.md` §1 (machine configuration,
locked keys) and §7 (source audit policy), §10 (status), environments §12
(posture) and §12.1/§12.2 only where the profile names environments knobs;
schemas: `manager-config-v2` / `system-config-v2` for the new knob.

## Settled decisions (do not reopen)
- **A named hardened-defaults profile** (impact row "S1 hardened defaults":
  opt-in profile first, default flip one release later): one closed machine
  knob `security_posture` with values `permissive` (today's defaults) and
  `hardened`, carried in `manager-config-v2` (top level, next to `audit`)
  and lockable in `system-config-v2` only in the direction of `hardened`.
  Under `hardened` the effective defaults are: `audit.mode: strict`,
  `audit.registry_policy: strict`, `allowed_sources` MUST be non-empty
  (`source_allowlist_empty` error at install/update), the environments
  `mcp_package_allowlist` MUST be non-empty for any profile that carries an
  MCP declaration (`mcp_package_allowlist_empty` becomes an error instead of
  the S4 warning), `passable_env_names` `null` (explicit unbounded) is
  refused (`passable_env_names_unbounded_refused`), `transitive_system_modules`
  is `error`, `require_source_signers` is `true`, and an unreachable trusted
  registry during install/update is a blocking gate notice (S3, below). An
  explicit per-knob value in the machine file still wins over the profile's
  default for keys that are not locked, EXCEPT the three refusals above,
  which are the profile's meaning. Rollout: revision A (this release) admits
  the knob with default `permissive` and requires every manager to report
  the posture it runs in; revision B (a later release) flips the default to
  `hardened` — state both explicitly with the migration hint printed under
  `permissive` in revision A (`security_posture_permissive`, warning, once
  per operation, naming the knob).
- **Posture report**: `curator status` and `env status` carry a
  `security_posture` row: the profile in force, and per gate the effective
  value and whether it comes from the profile default, an explicit machine
  value or a lock (hook trust A/B, passthrough s4-warn/s4-enforce,
  transitive system modules, provider trust roots revision, signer
  verification, update confirmation revision, store boundary, write
  discipline, registry policy, audit mode, source allowlist size, MCP
  allowlist size). Closed row vocabulary; `--check` treats a `hardened`
  machine whose effective values contradict the profile as non-current.
- **S3 residual named, not changed**: registry §4 and SECURITY.md state in
  one paragraph that under `advisory` registry policy revocation is
  network-dependent — an unreachable trusted registry can hide a revocation
  for up to the offline grace — and that `hardened` (strict policy) removes
  it at the cost of availability. Managers MUST surface an unreachable
  trusted registry during install/update as a prominent gate notice
  (`registry_unreachable_during_install`, warning under permissive, error
  under hardened) naming the artifacts resolved without registry evidence;
  the routine per-query warning stays as it is.
- Closed sets stay closed: exact spellings above (or the tables' preferred
  forms), spelled identically in prose, tables, schemas, vectors, CLI rows,
  CHANGELOG.

## Deliverable
1. `profiles/manager.md` §1 knob + lock rule, §7 effective policy under the
   profile, §10 posture row; `protocol/registry.md` §4 residual paragraph +
   gate notice, §8 offline grace cross-reference; `SECURITY.md` residual and
   hardened-profile paragraph; environments §12 posture row and the
   `mcp_package_allowlist_empty` / `passable_env_names` interactions in
   §2.2/§12.1 wording (reference the profile, do not re-specify S4).
2. Schemas: `manager-config-v2` `security_posture` (enum), `system-config-v2`
   lockable set (direction `hardened` only); generator + schema cases;
   vectors: a new `conformance/v1/vectors/security-posture.json` (effective
   defaults under each profile; explicit value vs profile default vs lock;
   the three refusals under hardened; the revision-A permissive warning;
   unreachable registry warning vs error; posture rows) registered in the
   manifest with a validator gate that pins every scenario (producer rule 7).
3. `CHANGELOG.md` Unreleased entry "S1/S3: …" naming the two rollout
   revisions and the residual.

## Out of scope
Implementation (`TASK-260910-1sapuy` and the manager posture work), S2
bootstrap checkpoint, changing any existing gate's own default (S4/E2/E4
defaults were settled by their own revisions).

## Checklist and handoff
Tick the checklist items you satisfy; attach
`TASK-260910-2qtiho_spec-patch_rev1.patch` (`git diff HEAD` of the worktree
with new files via `git add -N`; NOT `git diff origin/main`) and
`TASK-260910-2qtiho_evidence.md` (with the `make validate` and regeneration
transcripts), then `task-board handoff TASK-260910-2qtiho --role doc-writer`.
