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
