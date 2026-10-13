# R-EXEC must constrain interpreted inputs as well as executable paths

- **Lens:** Abuse controls, adversarial pass (astra-19); trust and operations cross-checks.
- **Status:** proposed.

## Problem

R-EXEC permits interpreters when a task profile needs them and permits qualified light checks through protected wrappers with fixed subcommands. R-VERIFY simultaneously requires every execution of worker-produced code to happen on an off-host ephemeral runner. The specified executable-path restriction, non-executable writable areas and fixed subcommands do not establish that boundary for interpreted inputs.

A worker can write a readable script, executable project configuration or plugin that an admitted interpreter or light-check tool evaluates inside its existing process. The writable input need not have an executable bit, and the kernel need see no executable other than the approved tool. Read-only dependency caches do not make the candidate's configuration, imports or plugins inert. A wrapper that validates only a subcommand does not settle those inputs either.

This is a pre-activation specification and qualification gap, not a demonstrated exploit in an enabled implementation. The prohibition is already stated; the missing part is the qualification contract that makes it true. Some parsers and light checks can satisfy it. The finding does not claim that all interpreters or type checkers execute project code.

## Evidence

- `swarma-user-manager` at `ae540c596c9416f1592491fc6133c36538c41921`, `spec/helper.md:433`: OS executable allowlist, non-executable writable areas, profile-needed interpreters and fixed-subcommand wrappers. The same paragraph already requires direct access to a shared underlying executable to refuse. `:435` prohibits worker-produced-code execution locally and supplies dependencies only from read-only caches. Qualification at `:541-543` names executable paths, heavy forms, wrapper bypass and off-host execution, without specifying admission of interpreted code/configuration inputs.
- `swarma-dispatcher` at `19a75e63cd794f93c81c8c6fdb9d176eba1ed87c`, `spec/dispatcher.md:336,340,640`: corresponding exec, verification and qualification rules apply to actual task tooling.
- `swarma-session-host` at `70c9d97e686db293e1e9a998a35909f1c5315824`, `spec/session-host.md:172,244`: profile-needed interpreters and protected wrappers coexist with the requirement that all worker-produced-code execution occur off-host.
- `curator-spec` at `a01542857678108a77d5e1b1f051989b00384f1d`, `cips/CIP-0009-donor-side-deployment-and-bridge.md:175`: the workplace interface admits policy-controlled commands, stdin and file writes. `:209-211` repeats the interpreter/wrapper permission and R-VERIFY prohibition; `:286` lists future negative qualification without interpreted-input cases.
- `curator-agent-launcher` at `da54e24b7d3bf3a565e9d2daa47bc90cc1c0b412`, `SPEC.md:747-774`: native network-adapter selection already rejects scripts and mutable interpreter closures and executes the evaluated artifact. That boundary is specific to native adapter selection; it does not qualify all task-tool invocations.

**Final-corpus revalidation, 2026-10-13.** Dispatcher DP7 at `74687e600e04817fa6e7a2e8904a9f7a3c83a1d8` preserves the relevant rules at `spec/dispatcher.md:384,388,698`. Launcher LR6 at `f08f566dd378ea27ce23ca0351ec793e95a696b6` now repeats the task-tool boundary at `SPEC.md:2006-2021`: protected fixed-subcommand wrappers, OS executable allowlists and profile-needed interpreters, with all worker-code execution reserved for R-VERIFY. Its harness-surface digest checks at `:1980-1987` protect harness launch/restart/resume; they do not establish inertness of worker-controlled project inputs consumed by every admitted tool. Session-host SH7 at `798a321c82a6a3f86532300ad64a6ba04b52f114` retains these exec, harness-surface and verification rules at `spec/session-host.md:176,241,248`. User-manager UM7 at `7b59da740037865dd42936982e5a069ea05eab53` preserves R-EXEC/R-VERIFY at `spec/helper.md:437,439` and their qualification rows at `:553-555`; the future task-profile/light-check/runner contract remains unpublished at `:613`. None of these final landings adds the interpreted-input qualification proposed here.

## Proposed change

Publish one shared R-EXEC qualification requirement for the complete admitted invocation: executable, entrypoint and arguments, environment, and every input the tool can interpret as code or configuration.

1. Distinguish inert inspection of candidate bytes from executing candidate-supplied code, including inline evaluation, scripts, module imports, plugins, startup files and executable configuration. An executable allowlist or a non-executable mount is not evidence that these inputs are inert.
2. A locally admitted tool must prove that worker-controlled inputs cannot select or supply executable behavior. Use protected fixed entrypoints and configuration, constrained environment and disabled or pinned code-loading mechanisms where the tool supports them. Verify the transitive code inputs, not just a top-level command name.
3. Preserve the existing direct-bypass prohibition. Where light and heavy forms share a binary, qualification must show the actual invocation boundary prevents direct or indirect access to the heavy form. Root ownership of the wrapper file alone does not establish exclusive access to the underlying executable.
4. Exercise the real task-tool path with negative cases for readable worker scripts, inline evaluation, imported modules, project plugins/configuration, startup/environment injection and indirect shared-tool invocation. Keep positive controls for genuinely inert light checks.
5. If a tool cannot demonstrate that boundary, refuse that local profile and send the operation to R-VERIFY. Do not infer safety from the task name, a fixed subcommand or resource ceilings, and do not require a new privileged wrapper service as the default remedy.

This strengthens the accepted R-EXEC/R-VERIFY decisions without changing their ownership or moving ordinary inert source inspection off-host.

## Expected benefit

The promised no-worker-code-execution boundary becomes a qualification obligation that covers how admitted tools consume data, rather than only which executable paths the kernel permits. It also gives language-profile authors a clear refusal rule when a nominally light command loads project code.

## Risks and costs

Qualification is tool- and release-specific. Some useful light checks will need an off-host runner, increasing latency and runner load. General-purpose interpreters may be usable only through narrow fixed entrypoints. These costs should be explicit; resource ceilings still bound resource use but cannot substitute for the code-execution boundary.

## Affected specifications

User manager R-EXEC/R-VERIFY and qualification; dispatcher task profiles and qualification; session-host light checks; CIP-0009 D5c and remote-worker publication/qualification register; launcher §10.2 exec profiles; the future language/tool profiles and verification contract.

## Confirmed by

Room: `#swarma-review-3`, topic `f-exec-input-closure`.

- Candidate: astra-19, `$Sosw2qZP4EsrbUdcxEHEzqH16NlsCkCZBdcpLUTL2Ds`.
- Independent trust confirmation: astra-18 (`ivmbp_swarma-review-3_astra-18`), `$KLJsnYu2RYcwiXyz7-rnaWxy4nvbwEjNYnmpnpgGatk`.
- Independent operations confirmation and shared-binary qualification refinement: core astra (`macmini_swarma-review-3_astra`), `$dA03ElfFPfgcxPJjQkGzRbwieMup6UsppqVVffMUr7I`.
- Independent first-pass abuse-control confirmation against DP7/LR6/SH7 and the unchanged helper text: astra-09 (`macmini_swarma-review-3_astra-09`), `$-5ZWuyRsoySMrlwxGVeqISz16iq9HDhwnhEcIkCxOvY`.
- Scope clarification preserving the already-stated direct-bypass prohibition: astra-19, `$TwtVpGx-D-SHF9xGLkjD9VeSa6-v5k3rYlog8CQZkQs`.
