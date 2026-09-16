# Goal: implement the 2026-09 security audit findings across Curator

Precondition resource for `EPIC-260910-2hw1xb` (manager + spec) and
`EPIC-260910-16qce1` (registry service). Prerequisite: curator-spec PR #49 and
curator PR #68 landed (the epics live on `main` only after #68).

## Where the discussion is

| Document | Pull request | Findings |
|---|---|---|
| `curator-spec/docs/security-audit-2026-09.md` | https://github.com/relux-works/curator-spec/pull/49 (comments carry the E-series supplement and its implementation verification) | S1–S6, R1/P1, P2–P4, E1–E7, Appendix B verdicts |
| `curator/docs/security-audit-2026-09.md` | https://github.com/relux-works/curator/pull/68 | S6/I1, S4, R1 client, S1+S3, S5, S2, I2, I3, S7 |
| `curator-skill-registry/docs/security-audit-2026-09.md` | https://github.com/relux-works/curator-skill-registry/pull/5 | R1 service half, R2–R8, P4 |
| Board resources | `EPIC-260910-2hw1xb`: `security-audit-2026-09-spec.md`, `security-audit-2026-09-manager.md`, `security-audit-2026-09-spec-supplement.md`; `TASK-260916-dv7xv5`: `verify-e-findings.md` | full evidence |

Related, not part of this goal: proposals 0014/0015/0016 (own goal; 0016
depends on E4 landing first), 0017/0018 (curator-spec issues #54, #55).

## Ground rules that shape every fix

1. **Spec before code.** A finding with a `spec-*` task lands its
   curator-spec revision (normative text + schema + conformance vectors)
   before the manager or service task starts; implementation follows the
   landed text, never a draft.
2. **Warn first, then flip.** Any change that alters day-to-day behavior
   (the table below) ships in two steps: a release that reports the new
   posture as a warning with a migration hint, then the release that makes
   it the default. Never flip a default and a refusal in the same release.
3. **Posture is reported, not implied.** Every gate gained here shows its
   state in `env status` / `curator doctor`-class output.
4. **Closed sets stay closed.** New diagnostics, config keys and lock keys
   enter through the §12.3 admission rule with vectors.

## Priority waves

**Wave 1 — High, contained fixes (start immediately, parallel stories):**
- `STORY-260910-2awkzu` S6/I1 shell hook: source only digest-recorded or
  operator-approved `.agents/env.sh`; unknown files warn and are skipped;
  `hook-approval-command`. Tasks `TASK-260910-1wjst3`, `-1952mz`, `-3ungjy`.
- `STORY-260916-2otjbn` E4 provider trust roots: providers resolve from the
  manager install directory plus a machine-config list, never ambient
  `PATH`; resolved path printed. Tasks `TASK-260916-1x0ogh`, `-3oh0u8`,
  `-16ys92`. **Blocks 0016 / `path_prepend`.**
- `STORY-260916-2d9coh` E2 transitive `class: system` modules: machine policy
  `transitive_system_modules = drop | error`, default `drop` (a transitive
  system module is skipped with a warning naming package and module; root
  modules still materialize; installs never break), `error` for strict
  machines (`context_system_module_transitive`); naming the package directly
  in the root or a per-package waiver admits it. Tasks `TASK-260916-1hrx51`,
  `-55g9dg`.
- `STORY-260910-1lf0m5` S4 MCP exposure: `passable_env_names` default
  empty, loud warning on empty `mcp_package_allowlist`, `command`+`args`
  surfaced at install/update. Tasks `TASK-260910-2ohnjo`, `-gocke2`.

**Wave 2 — High, cross-repo (spec envelope first):**
- `STORY-260910-25yc0h` R1/P1 client + `STORY-260910-3rvvxh` R1 service:
  records/log pages carry the committed boundary; clients reject pages below
  their persisted high-water. Order: `TASK-260910-1b1ens` (spec) →
  `-14dnb7`/`-27yepb` (service) → `-2n0233` (client). Old registries without
  the field are reported, not silently accepted.
- `STORY-260916-ioemse` E1 signer allowlist + `profile update` delta with
  confirmation on system-module or MCP changes; `latest` residual named.
  Tasks `TASK-260916-y4sa6s`, `-1zgucp`.

**Wave 3 — Medium:**
- `STORY-260910-2qmrb8` S1+S3 hardened-defaults profile, unreachable
  trusted registries surfaced at install (`TASK-260910-2qtiho`, `-1sapuy`).
  Lands after S4 and E1 so the posture report covers all gates at once.
- `STORY-260910-148pj1` S5 store boundary (`TASK-260910-39fzpq`, `-32gki6`),
  with the E7 repair-as-persistence note.
- `STORY-260910-6bo7ej` S2 signed bootstrap checkpoint, cross-registry root
  check (`TASK-260910-1tvf2t`, `-2vnjej`).
- `STORY-260916-1i1gfo` E3 codex seed `mcp_servers` (`TASK-260916-2rnkei`,
  `-33abdk`). Operator decision embedded: strip at provisioning; documented
  in the manager README/CHANGELOG ("a managed codex home runs only the
  profile's MCP set; native `~/.codex/config.toml` servers are not
  inherited") and listed by `env status` as dropped entries.
- `STORY-260916-73a5zg` E5 nofollow rule + vector (`TASK-260916-1qfpu4`,
  `-19shmj`); code already mitigates, add atomicity review.
- `STORY-260916-wgt8vz` E6 path-kind admission (system modules + directory
  boundary; MCP half not applicable per Appendix B) (`TASK-260916-3l60rn`,
  `-yvxbs1`).
- `STORY-260910-234vmx` I3 installer attestation, I2 askpass via pipe, S7
  posture docs (`TASK-260910-2t0iun`, `-31ocjt`, `-3i6vod`).
- Service: `STORY-260910-1py4f3` R2 memoization + cached health verdict,
  `STORY-260910-35tbgb` R3+P2 serve-time checkpoint gate
  (`TASK-260910-33j1hu` spec first).

**Wave 4 — Low / Info:**
- `STORY-260916-33vuzm` E7 launcher config ownership + symlink refusal.
- Service `STORY-260910-2xe3n2` R4/R5/R7/R8, `STORY-260910-9484i4` R6 key
  passphrase, `STORY-260910-stz5f0` P4 import high-water.

## User-visible impact (drives the warn-first rollout)

| Fix | What changes for an operator | Rollout |
|---|---|---|
| S6 hook gate | a project's `.agents/env.sh` is no longer sourced on `cd` until approved once (`hook-approval-command`); re-approval after the file changes | warn-first |
| S4 passthrough default | MCP servers that read operator env vars (bearer tokens) stop receiving them until each name is allowed in machine config; empty allowlist warns loudly | warn-first, migration hint names the variables |
| S1 hardened defaults | strict registry policy blocks unknown artifacts; empty allowlists become warnings/refusals under the hardened profile | opt-in profile first, default flip one release later |
| E1 update confirmation | `profile update` stops for confirmation when system modules or MCP declarations change; unattended updates need an explicit flag | warn-first |
| E2 transitive system modules | a transitive `class: system` module is dropped with a warning (default) instead of applied; strict machines can make it an error; naming the package directly or a waiver admits it (relux-root-context today carries none) | direct: default is non-breaking |
| E3 codex seed | managed codex homes no longer inherit native `mcp_servers`; only the profile's set runs | warn-first with `env status` listing the dropped entries |
| E4 trust roots | `curator run` refuses a provider outside the trust roots; a `curator-run` installed outside the manager's directory (e.g. `/usr/local/bin`) must be listed in machine config | warn-first, `env status` names the resolved path |
| E7 config ownership | a symlinked or foreign-writable `defaults.json`/`ax.json` (dotfile managers) is refused | warn-first |
| R1, S2, S5, S3, E5, E6, I2, I3, R2–R8, P2–P4 | under the hood; visible only as new diagnostics or refusals on tampered/misconfigured state | direct |

## Operating rules

Those of `goal-launcher-and-infra-migration.md` (precondition on
`EPIC-260908-2wp8wn`): board first, stories with worktrees, producer/reviewer
subagents on `claude-fable-5-1` at LOW reasoning effort, briefs as
precondition resources and findings as outcome resources, board-state commit
per closed story pushed to `main`, runs never write into the control root,
PR canon with `${SHA}:refs/heads/main` landing and hard-fail checks, signed
commits never rewritten, no tags or GitHub Releases without the operator's
command, never merge the ax PR, verify facts on installed binaries, escalate
only for product decisions or human-only access. Registry-service work
follows the same canon in `curator-skill-registry`.

## Definition of done

Every story of both epics `done` with its spec revision (where any),
conformance vectors, manager/service implementation and `env status`
posture reporting landed on the respective `main`; the warn-first release
notes recorded per user-visible fix; `task-board validate` clean; both audit
documents amended with a "Remediation status" table pointing at the landed
commits.
