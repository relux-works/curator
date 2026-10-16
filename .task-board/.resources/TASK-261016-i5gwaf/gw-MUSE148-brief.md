# Brief MUSE148: rc.7 code wave, M2 — Qualify the Muse 1.4.8 loader without inventing model evidence (TASK-261016-i5gwaf)

Implement the attached plan leaf file(s) exactly: `rc7-leaf-muse-148-exclusion-replay.md` 
Read rc7-code-common.md and rc7-decisions.md (attached) first.

- The leaf is the contract: its normative clauses, acceptance criteria, mandatory paired tests, plant controls, dependencies and paths. Read the cited normative sections on curator-spec main yourself; where the leaf and the normative text disagree, the normative text wins: stop and report.
- Stay inside the leaf's attack surface and paths. If the leaf is too big for one reviewable change, implement its first coherent part, say exactly what remains, and do not fake the rest.
- No runtime qualification is claimed unless you executed it. Record skipped or unavailable rows honestly.
- Leave the change uncommitted for the Change Request flow.

Host safety for this probe (mandatory):
- Run Muse 1.4.8 only with `MUSE_NO_AUTO_UPDATE=1`, disposable `XDG_CONFIG_HOME`/`XDG_DATA_HOME`/`XDG_STATE_HOME`/`XDG_CACHE_HOME` under the task's `.temp/`, and the offline echo provider.
- Never `muse login`, `logout`, `auth` or any store-touching subcommand, and never point Muse at a real home. The real login lives in the Keychain item `ai.meta.dev.credentials`; do not read or touch it.
- Record the exact binary path, version and SHA-256 that you used.

## Rework 1 (convener, after the first run)

The first run produced a real result: on Muse 1.4.8-R6256.1 (macOS 27 arm64) the exec exclusion works, but the interactive `-c context.foreign_personal_skills=false` still loads the foreign skill. That is the deliverable of this leaf: a qualification record. It is not a reason to stop. Finish it as follows:
1. Write the qualification record as a file in the repository; follow existing qualification-record conventions if any, otherwise put it under `docs/qualification/`. It holds the exact build and SHA-256, platform, fixtures, commands, exit codes and per-row results, with exec and interactive reported separately:
   - exec form: loader rows observed and passing;
   - interactive form: P1–P4 skill-exclusion row FAILED on this build (vendor behaviour), the other rows as observed;
   - model-visible behaviour and refresh: not measured.
2. Tests: the default `go test` suite must stay green.
   - Keep the vendor probe tests behind the repository's opt-in probe mechanism (build tag or env gate), never in the default suite.
   - The probe test for the interactive row asserts the OBSERVED behaviour of this exact build as a regression witness. It must flip and fail loudly when a corrected build excludes the skill, and point to this record.
   - Keep the checker/parser/offline-boundary tests that are green.
3. Change no runtime constants or gates: Curator does not qualify anything itself. State plainly in the record that the §7.9 release gate stays pending for the interactive form on this build.
4. Hand off for review as usual (to-review).
