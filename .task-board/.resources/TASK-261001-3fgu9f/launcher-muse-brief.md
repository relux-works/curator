# TASK-261001-3fgu9f — curator-run: muse provider mapping + launch-env-fragment-v3 reader (THE ONLY CURRENT INSTRUCTION; priority, Muse root-session path)

curator-agent-launcher repository (curator-run). The curator side (TASK-261001-2yvag1) makes `curator env resolve muse --repair --format json`
emit launch-env-fragment-v3. The normative source is the curator-spec candidate branch `spec-muse-environment`, d373078a:
protocol/environments.md muse sections, Decision 0018 additions, Decision 0013 launch plans with the launcher §4.2 mapping, and
schemas/v1/launch-env-fragment-v3.schema.json. Today `curator run muse` exits 1 (`resolve_fragment_invalid`), because curator-run has
no muse mapping and no v3 reader.
1. Add `muse` to internal/mapping, using the env id → (launcher target, ax provider id) pair the spec gives. If the ax provider id for
   muse is not defined in the spec or ax, stop and report it in the results with evidence. Do not invent one.
2. Accept launch-env-fragment-v3 in the plan reader, keeping v1/v2 behaviour for the other environments. Apply the muse rules from the
   spec: XDG variables only, never HOME, and the permission interface per Decision 0018 (exec --yolo / serve flags) only when the
   operator asks.
3. Tests with a fake curator (a canned v3 fragment) and a fake muse or ax binary: the plan, argv and env for muse; v2 fragments
   unchanged; HOME never set. Mutants, with real exit codes: HOME set → a test fails; v3 rejected → a test fails.
4. Run `go test ./...` (bounded, -p 1 if the host is loaded) and `GOOS=windows go vet ./...`, with real exit codes. Update the
   CHANGELOG if this repo keeps one under Unreleased. Never spell any employer name.
Update the results, then run `task-board handoff TASK-261001-3fgu9f --role developer`, then END YOUR TURN.

## Decision (binding, 2026-10-01 03:4xZ): reduce scope, do not stop
The Muse plugin blocker (no interactive root-session plan in agents-management v0.5.33) is reported to tb-keeper as an upstream item. Do NOT change the plugin mode or rebuild argv. Implement now what is independent of it:
(a) the muse mapping row (env id → launcher target, ax provider id "muse");
(b) the launch-env-fragment-v3 reader in the plan path, with v1/v2 unchanged and HOME never set;
(c) unit tests with a canned v3 fragment, plus the two mutants (HOME set; v3 rejected).
Keep the interactive `curator run muse` plan.Build refusal as an explicit, tested bound: assert today's exact error, and make the row flip when a plugin declaring interactive is pinned. Set status development, update the results, run `task-board handoff TASK-261001-3fgu9f --role developer`, then END YOUR TURN.
