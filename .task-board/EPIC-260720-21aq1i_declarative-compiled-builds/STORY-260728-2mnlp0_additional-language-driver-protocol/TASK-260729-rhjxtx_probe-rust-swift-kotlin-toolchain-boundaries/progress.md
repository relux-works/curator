## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(8))

## Blocked By
- (none)

## Blocks
- TASK-260728-1yhuqi

## Checklist
- [x] Implement reproducible Rust, Swift, and Kotlin Native version, target, and metadata probes with machine-readable fixtures
- [x] Add malformed, future-version, incompatible-target, and forbidden-selector negative controls that fail before compiler work
- [x] Run probes on macOS and the reachable Windows SSH host without installing or mutating toolchains
- [x] Attach exact command, version, output, exit-code, and primary-source guidance evidence consumable by all three language design tasks
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260728-bc90af, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260728-bc90af)
Probe built and run on macOS 26.5/arm64 and the reachable Windows host (SSH alias win, 10.0.19045/amd64). Evidence attached as 7 task-scoped outcome artifacts.

Results: macOS 19 cases, 15 matched, 0 divergences, 4 not run, exit 0; verdicts identical across two independent runs. Windows: no Rust/Swift/Kotlin/JDK/Go toolchain present at all, 19 not run, exit 0, nothing installed, remote workdir removed.

Load-bearing findings for the three design tasks:
1. Rust and Swift both accept a target they cannot build for (rustc --print target-libdir exit 0 with std absent; swiftc -print-target-info exit 0 for a linux triple). Admission is a manager-side stat inside the fingerprinted tree. Rust is the only case where upstream rejection arrives after compilation starts.
2. rust-toolchain.toml is a live selector: channel=nightly made the rustup shim attempt a download and install, path= redirected it; the directly resolved root is inert against both. Measured with the dist endpoint pointed at a closed loopback port; nothing installed.
3. swift package tools-version reads the header only (exit 0 against a syntax-error manifest and a fatalError manifest); dump-package compiles and executes the manifest program. A Swift graph phase is bounded at the header.
4. Three of four version probes need an explicit stream: swift --version splits one banner across stdout and stderr, kotlinc -version writes to stderr with empty stdout, rustc --version mixes in the commit hash.
5. Swift target.triple carries a deployment-version component (26.0 on a 26.5 host and SDK); unversionedTriple is the identity form.
6. Defect the probe found on itself: cargo resolves rustc from PATH, so a Rust driver must pass RUSTC explicitly or PATH order picks the second process-graph node.

Honest gaps: Kotlin/Native absent on both hosts, so 4 Kotlin cases are not_run, the konanc rule is named UNVERIFIED, and no Kotlin metadata file or field is proposed. No Windows platform claim is evidence-backed. Only Apple Swift was measured.

Gates: probe gofmt/vet/test/golangci-lint all exit 0; curator go build, make vet, make test all exit 0. make check exits 2 at its gofmt -l . stage on 7 files in other tasks' .temp scratch trees (0 from this task; all tracked .go files clean).
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260728-bc90af, pid=28514, exit=0)
Independent review directive: validate the probe as non-normative host evidence consumable by the Rust, Swift and Kotlin Native design tasks. Re-run the standalone Go probe tests/lint and inspect both machine-readable fixtures. Challenge every authoritative command, stream, normalization, target admission check, selector classification, and representability-versus-host-gate conclusion against primary local toolchain sources and measured outputs. Require no auto-install, no package-controlled executable/path/env/trust selection, Kotlin Native rather than JVM, explicit UNVERIFIED/not_run treatment, honest Windows absence, and no platform qualification inferred from missing toolchains. Verify macOS evidence and read-only SSH win execution, deterministic fixtures, scoped files, and attributed make-check red. Do not edit producer artifacts. Attach TASK-260729-rhjxtx_review-verdict-cycle-1.md with ACCEPTED or CHANGES REQUESTED and route status accordingly.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260728-42e1e8, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260728-42e1e8)
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260728-42e1e8, pid=53246, exit=0)
REWORK CYCLE 1 directive: preserve the existing task-owned probe module and close all four findings in TASK-260729-rhjxtx_review-verdict-cycle-1.md. (1) Unify .swift-version case expectations across executed and not_run fixtures from one canonical case table; add a regression comparing declared and executed expectations. (2) Stop inferring Swift compilation start from nonzero exit plus target-print success; use a measured output/process signal or set false, then regenerate fixture, rendered evidence and prose consistently and test it. (3) Implement the supplied-root Kotlin/Native branch as a genuinely runnable, no-install probe: authoritative stderr version/backend normalization, native target and -list-targets observations, positive version/target case, malformed/future requirement or explicit no-legal-metadata-boundary handling, and unknown-target control. Current macOS/Windows absence must remain not_run and UNVERIFIED until a real Kotlin/Native host exists; never fabricate qualification or auto-install. (4) Add exact Windows cross-compile, transfer, SSH argv, hash verification, execution, cleanup commands, outputs and per-command exit codes as task-scoped evidence. Re-run gofmt/vet/test/lint, deterministic macOS fixtures, read-only SSH win absence run and scoped project gates. Attach rework-cycle-1 evidence and hand off to a fresh independent reviewer. No normative spec edits, installs, staging, commit or publication. Queue for the first free Opus slot after the active main toolchain rework.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260728-ed8ee4, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260728-ed8ee4)
REWORK CYCLE 1 COMPLETE — all four review-verdict-cycle-1 findings closed with measured evidence.

(1) Case-expectation drift. Added cases.go: one canonical caseTable() declaring id/language/kind/title/expected/expect-early exactly once. notRunLanguage() and every executed builder derive from it, and finish() re-applies the declaration to every case it emits, so a builder that restates an expectation is overwritten rather than trusted. .swift-version is now GateNone in both places (the measured value: the file does not redirect a directly resolved toolchain). Regressions: TestEmittedCasesCarryTheDeclaredExpectation and TestNotRunAndExecutedCasesAgreeOnTheContract, both run against real local Rust/Swift/kotlinc roots (exit 0, in the gate log). Verified directly on the two produced fixtures: 19 shared case IDs, 0 cross-host contract divergences.

(2) Swift compilation inference. Removed. Observation gained build_compilation_evidence, the exact output line that proves a compiler process ran; BuildStartedCompilation is derived from it in finish() and cannot be set independently. Swift compiles now run with -v so the driver job trace is measurable. Measured result CORRECTS the cycle-0 prose: swiftc DOES spawn swift-frontend -frontend -c -primary-file for x86_64-unknown-linux-gnu (upstream_rejects_after_compilation=true) and spawns NO frontend for not-a-real-triple (false). Rust is therefore NOT the only case where rejection arrives after compilation starts; fixtures, rendered evidence and results.md all regenerated consistently. Regressions: TestCompilationBooleanIsDerivedFromMeasuredEvidenceOnly, TestCompilationMarkerReadsBothStreamsAndReturnsTheLine, TestCompilationMarkerBoundsTheRecordedLine, TestFinishOverwritesARestatedExpectation.

(3) Kotlin/Native supplied-root branch. Implemented and runnable: authoritative stderr version banner with backend token, UNVERIFIED konanc normalization rule, -list-targets parsing with a (default) native-target identity, positive version+target case, unknown-target representability control, and an unserved-target host gate whose target is SELECTED FROM the distribution own output rather than asserted. Every path degrades to not_run with a stated reason when real output does not match the declared shape. Exercised by 7 tests against a synthetic stub root, which verifies probe logic only and is explicitly not a Kotlin/Native claim; no fixture is ever generated from it. macOS and Windows Kotlin cases stay not_run and the rule stays kotlin.konanc.version.UNVERIFIED.

Metadata: took the explicit no-legal-metadata-boundary branch the reviewer offered. kotlinc-native defines no manifest and the ecosystem manifest is a Gradle script. NEW measured case kotlin.negative.metadata-is-a-program (needs no Kotlin/Native): answering the pure metadata query gradle properties compiled the build script as a program source unit (_BuildScript_) before answering. Contained via GRADLE_USER_HOME inside the work dir, --offline --no-daemon; verified the operator ~/.gradle was not written to. build.gradle.kts NOT measured (needs network). Consequence recorded for TASK-260728-168smo: mint a curator-owned metadata file for the build root, or decision 0008 section 10 retirement applies. This probe proposes no candidate file and no candidate field.

(4) Windows evidence. TASK-260729-rhjxtx_windows-commands.log: 12 steps, exact argv, per-command exit code, all 0. Probe sha256 e98ec602… matched by remote certutil; fixture sha256 af2653d1… matched remote-to-local after fetch; remote workdir removed and absence confirmed (REMOVED). Read-only throughout, nothing installed. Also attached TASK-260729-rhjxtx_macos-commands.log.

Results: macOS 19 cases, 16 matched, 0 divergences, 3 not run, exit 0; identical verdicts across two independent runs. Windows 19 not run, exit 0.

Gates, real exit codes: probe gofmt/build/vet/test/golangci-lint all 0; executed-vs-declared regression against real toolchains 0; curator go build 0, make vet 0, make test 0. make check exits 2 at its gofmt -l . stage on four VENDORED third-party files under .temp/TASK-260720-1zntv0/cycle2/curator/vendor/ — another task scratch tree, zero from this task. No tracked project file modified.

Honest gaps unchanged: no Kotlin/Native on either host so 3 Kotlin cases are not_run and konanc rules/parsers are UNVERIFIED; no Windows platform claim is evidence-backed; only Apple Swift was measured; Linux not probed.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260728-ed8ee4, pid=70661, exit=0)
REVIEW CYCLE 2 DIRECTIVE: Independently verify every cycle-1 blocker is actually closed. Require one deterministic .swift-version classification/expectation across macOS and Windows fixtures; require compilation-boundary claims to come only from measured signals and reject any inferred upstream_rejects_after_compilation value; exercise the supplied-root Kotlin/Native path and confirm it either runs the complete version/target/list-targets/positive/malformed/future matrix or emits a precise no-legal-metadata/unknown-target boundary while absent hosts remain honest not_run/UNVERIFIED; replay the exact Windows crosscompile, digest, mkdir, transfer, remote hash, inventory, run, fixture hash/fetch/verify, cleanup and absence-check commands with recorded exits. Re-run probe gofmt/vet/test/lint plus macOS and Windows fixtures, inspect manager obligations, and independently attribute the known repository-wide make check failure. Do not modify candidate artifacts. Attach TASK-260729-rhjxtx_review-verdict-cycle-2.md with decisive ACCEPTED or CHANGES REQUESTED and route accordingly.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260728-4f9c53, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260728-4f9c53)
REVIEW CYCLE 2 VERDICT: ACCEPTED. Independent replay confirmed the canonical 19-case contract across macOS/Windows, measured-only compilation-boundary flags, runnable but explicitly UNVERIFIED Kotlin/Native supplied-root branch with honest no-project-metadata boundary, byte-identical Windows fixture after hash/SSH/inventory/run/fetch/cleanup replay, and green probe format/build/vet/test/lint gates. Repository make check remains red only on four unrelated vendored files under .temp/TASK-260720-1zntv0. Full evidence: TASK-260729-rhjxtx_review-verdict-cycle-2.md. Reviewer supplies no commit_ack.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260728-4f9c53, pid=89787, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260729-rhjxtx_spawn-log_-implementer--developer--claude-_RUN-260728-bc90af.log](file://TASK-260729-rhjxtx/TASK-260729-rhjxtx_spawn-log_-implementer--developer--claude-_RUN-260728-bc90af.log) — System spawn log captured by task-board
- [TASK-260729-rhjxtx_results.md](file://TASK-260729-rhjxtx/TASK-260729-rhjxtx_results.md) — Rework cycle 1 developer handoff: all four review findings closed, corrected Swift compilation-boundary fact, Kotlin no-legal-metadata boundary, honest UNVERIFIED/not_run status, full gate attribution
- [TASK-260729-rhjxtx_probe.tar.gz](file://TASK-260729-rhjxtx/TASK-260729-rhjxtx_probe.tar.gz) — Reproducible probe, rework cycle 1: standalone Go module (9 sources, 4 test files, 49 test functions) plus the run/gate/render harness scripts; gofmt/vet/test/golangci-lint all exit 0
- [TASK-260729-rhjxtx_fixture-macos.json](file://TASK-260729-rhjxtx/TASK-260729-rhjxtx_fixture-macos.json) — toolchain-probe-fixture-v1, macOS 26.5 arm64, rework cycle 1: 19 cases, 16 matched, 0 divergences, 3 not run, exit 0; measured build_compilation_evidence per case
- [TASK-260729-rhjxtx_fixture-windows.json](file://TASK-260729-rhjxtx/TASK-260729-rhjxtx_fixture-windows.json) — toolchain-probe-fixture-v1, Windows 10.0.19045 amd64 via SSH alias win, rework cycle 1: all three toolchains absent, 19 not run, exit 0, nothing installed, remote workdir removed
- [TASK-260729-rhjxtx_command-evidence.log](file://TASK-260729-rhjxtx/TASK-260729-rhjxtx_command-evidence.log) — Both rework-cycle-1 runs rendered: every argv, env delta, exit code, output excerpt, fixture digest, manager obligation, and the measured compilation-evidence line
- [TASK-260729-rhjxtx_windows-inventory.log](file://TASK-260729-rhjxtx/TASK-260729-rhjxtx_windows-inventory.log) — Read-only Windows inventory: 12 PATH names and 12 candidate roots searched, all absent; no install performed
- [TASK-260729-rhjxtx_gate-log.txt](file://TASK-260729-rhjxtx/TASK-260729-rhjxtx_gate-log.txt) — Rework cycle 1 gate transcript with real exit codes: probe gates 0, curator build/vet/test 0, make check 2 attributed to four vendored files in another task's .temp tree
- [TASK-260729-rhjxtx_spawn-log_-reviewer--reviewer--codex-_RUN-260728-42e1e8.log](file://TASK-260729-rhjxtx/TASK-260729-rhjxtx_spawn-log_-reviewer--reviewer--codex-_RUN-260728-42e1e8.log) — System spawn log captured by task-board
- [TASK-260729-rhjxtx_review-verdict-cycle-1.md](file://TASK-260729-rhjxtx/TASK-260729-rhjxtx_review-verdict-cycle-1.md) — Reviewer cycle 1: CHANGES REQUESTED with reproducible evidence and exact rework requirements
- [TASK-260729-rhjxtx_spawn-log_-implementer--developer--claude-_RUN-260728-ed8ee4.log](file://TASK-260729-rhjxtx/TASK-260729-rhjxtx_spawn-log_-implementer--developer--claude-_RUN-260728-ed8ee4.log) — System spawn log captured by task-board
- [TASK-260729-rhjxtx_windows-commands.log](file://TASK-260729-rhjxtx/TASK-260729-rhjxtx_windows-commands.log) — Review finding 4: the 12 exact Windows steps (cross-compile, digest, mkdir, scp, certutil verify, inventory, probe run, fixture digest, fetch, verify, rmdir, absence check) with argv and per-command exit codes, all 0
- [TASK-260729-rhjxtx_macos-commands.log](file://TASK-260729-rhjxtx/TASK-260729-rhjxtx_macos-commands.log) — The exact macOS rework-cycle-1 steps with argv and per-command exit codes, including the check that the operator Gradle home was not written to
- [TASK-260729-rhjxtx_spawn-log_-reviewer--reviewer--codex-_RUN-260728-4f9c53.log](file://TASK-260729-rhjxtx/TASK-260729-rhjxtx_spawn-log_-reviewer--reviewer--codex-_RUN-260728-4f9c53.log) — System spawn log captured by task-board
- [TASK-260729-rhjxtx_review-verdict-cycle-2.md](file://TASK-260729-rhjxtx/TASK-260729-rhjxtx_review-verdict-cycle-2.md) — Reviewer cycle 2: ACCEPTED with independent macOS and Windows replay evidence

## Created
2026-07-28T20:08:13Z

## Last Update
2026-07-29T00:48:44Z

## Assigned To
[reviewer] reviewer (codex)
