# HOLD — conformance-pin lag blocks the hosted gate for knob-changing manager tasks (2026-09-17)

Orchestrator note for `TASK-260916-3oh0u8` (E4), `TASK-260916-55g9dg` (E2) and
`TASK-260910-gocke2` (S4).

The hosted gate runs the curator test suite against the committed
`SPEC_PIN` (curator-spec `87a0d00` = released tag `v1.0.0-rc.11`). That root
predates the wave-1 spec landings (`90d50c6`, `f544a01`, `0da4020`, `23dafa7`).
`internal/config` `TestManagerConfigV2Vectors` compares the normalized
manager config byte-for-byte with the root's `vectors/manager-config-v2.json`,
so any implementation that adds a §12.1 knob (`provider_directories`,
`transitive_system_modules`, `system_module_waivers`) or changes a default
(`passable_env_names` → `[]`) fails every default lane — exactly what happened
to E4 revision 1 (run 35174489288: eight `schema2-*` cases red on all three
platforms; the implementation is right against the new root and wrong
against the old one). Locally you tested against the `0da4020`/`23dafa7`
root, which is why it passed.

**Do not work around this in the test** (no key-subset comparison, no
root-version branching, no vendored vectors) and do not touch `SPEC_PIN`:
promoting the pin is a release decision (`TASK-260720-25d05o` qualifies the
release, `TASK-260720-38l1sy` promotes the pin) that the operator is deciding
now. Finish and keep your implementation complete against the landed spec;
if your run is cancelled or refused at the gate, your worktree state is
preserved and a successor will be routed once the pin question is settled.
The S6 task (`TASK-260910-1952mz`) is unaffected: its gate failure was the
non-POSIX hook, not the pin.

## Update 2026-09-17 06:30Z
The E2 revision-3 gate (run 35189087542) also failed the **gate self-test**:
`.github/ci/gate-selftest.sh` refuses `root-content` ledger rows for packages
the committed pin already serves (`internal/config TestSystemModuleSchemaSubset`:
"tolerates a skip that can never legitimately happen"). So neither the vector
comparison nor a root-content skip can absorb a pin that predates the spec
revision: the repository's gates assume `SPEC_PIN` serves the whole module.
The only sanctioned path is promoting `SPEC_PIN` to a curator-spec revision
that publishes the wave-1 families (release-qualification tasks
`TASK-260720-25d05o` → `TASK-260720-38l1sy`), which the operator decides.
Until then E2, E4 and S4 stay on HOLD with their candidates preserved.

## Update 2026-09-17 15:30Z — wave 2 manager/client tasks join the hold
The R1/P1 spec landed at curator-spec `dced9b8` (new `records-response-v2` /
`log-response-v2` schemas, `registry-client.json` `page_boundary_cases`) and the
E1 spec is in review (new `environments-source-signers.json` family, new
`manager-config-v2` knobs `source_signers` / `require_source_signers`). The
curator client task `TASK-260910-2n0233` and the E1 manager task
`TASK-260916-1zgucp` consume those vector families through packages the pin
already serves (`internal/registry`, `internal/config`), so the same
`root-content`/byte-exact rules would make their gates red at `87a0d00`. Both
stay unstarted until the pin decision; the service half (curator-skill-registry
PR #7, pinned at `dced9b8` in its own CI) is landed and does not wait.

## Update 2026-09-18 — pin plan after wave 3 specs landed
curator-spec `main` is `5146c7b` (R1/P1 `dced9b8` → E1 `684c9f1` → R3/P2
`47c3c8c` → E5 `9912db7` → S5 `e8b53a0` → E3 `4a2fa3e` → E6 `1ca4b3d` → S2
`1e73c03` → S1+S3 `5146c7b`). Curator tests consume NAMED vector families
(no dynamic enumeration of the root), so a newer pin breaks only the
byte-exact comparisons of families curator already implements — chiefly
`internal/config TestManagerConfigV2Vectors` (every manager-config-v2 case)
and the schema-case renderers. Knob-bearing revisions: E1 (`source_signers`,
`require_source_signers`), S2 (`bootstrap_checkpoint`, `mirror_group` on
registry entries), S1+S3 (`security_posture`). Plan:
- **rc.12 = `dced9b8`** (tag pending from the operator): the accepted union
  `TASK-260917-16l2md` (E2/E4/S4 + pin) lands after `TASK-260917-2ecpjv`
  qualifies the tag; unblocks R1-client `TASK-260910-2n0233` (no knob).
- **rc.13 = `684c9f1`** (E1): `TASK-260916-1zgucp` implements E1 and bumps
  the pin in one candidate (brief attached); a qualification sibling
  follows the rc.12 pattern. E5/S5/E3/E6 manager tasks (no knobs) can also
  run at rc.13 with their families read from a detached root at the
  landed revision, but their conformance tests would skip in CI until the
  pin reaches their revision — prefer landing them under rc.14.
- **rc.14 = `5146c7b`** (or the wave-3 tip at that time): S2 client
  `TASK-260910-2vnjej` (registry-entry members) and S1+S3 manager
  `TASK-260910-1sapuy` (`security_posture`) plus the E5/S5/E3/E6 manager
  tasks as one union with the pin bump, after the rc.13 story lands.
Each pin needs an operator-signed tag; the orchestrator asks for the next
tag as soon as its union candidate is accepted.
