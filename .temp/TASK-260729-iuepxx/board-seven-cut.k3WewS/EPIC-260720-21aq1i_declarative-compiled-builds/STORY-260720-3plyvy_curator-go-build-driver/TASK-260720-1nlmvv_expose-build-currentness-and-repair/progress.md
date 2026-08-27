## Status
development

## Assigned To
[implementer] developer (claude)

## Created
2026-07-20T02:11:48Z

## Last Update
2026-07-28T23:59:33Z

## Blocked By
- TASK-260720-3itlly
- TASK-260720-2284br
- TASK-260720-1ljev5

## Blocks
- TASK-260720-2qqq0w
- TASK-260720-jrrgw9

## Checklist
- [ ] Human and JSON diagnostics expose every stable currentness code
- [ ] Install and upgrade repair invalid compiled state safely
- [x] Legacy CLI behavior and exit codes remain compatible
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
Implementation directive: start from the independently accepted TASK-260720-1ljev5 candidate at /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1ljev5/worktree, preserve it read-only, and work in this task own isolated worktree. Implement the complete CLI/currentness/repair surface across install, upgrade, dry-run, status, status --json, status --check and gc using existing protocol and code patterns. Enumerate every stable currentness/repair code from the accepted implementation/spec and keep human and JSON semantics mechanically aligned; preserve legacy output/exit behavior where compiled builds are not involved. Install/upgrade may repair invalid compiled state only through the accepted staged, atomic and GC-safe lifecycle, with deterministic diagnostics and manager-owned guidance; no auto-install or package-controlled executable/env/trust selection. Add focused unit/integration/race tests, macOS-native validation, Windows cross-build plus task-owned native checks over SSH alias win where applicable, and Linux cross-build/vet without claiming later native Linux validation. Do not stage, commit, publish, or overwrite predecessor/user changes. Attach task-scoped implementation and gate evidence, then hand off to independent review.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260728-5b614c, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260728-5b614c)
Compiled diagnostics and repair surface implemented on the accepted 1ljev5 tree in .temp/TASK-260720-1nlmvv/worktree (base origin/main 17804ce, nothing staged or committed).

New: cmd/curator/builds.go (stable currentness vocabulary, per-command classification, redaction-aware rendering, repair notices, toolchain guidance) and internal/install/diagnostics.go (RedactDiagnostic: one-line, non-printable-stripped, absolute-path-redacted, 240-rune-bounded rendering of untrusted detail, applied by both status and PlannedBuild.Describe).

status now reports one diagnostic row per active build command with driver, build_root, source_dir, build-source identity, native target, logical key, artifact path and cache outcome; --json adds an optional builds array and leaves the historical document byte-identical when the closure has no build commands; --check fails closed for every non-current code on both the declared-skill map and the build rows. install/upgrade print whether untrusted state was rebuilt or the previous installation was preserved. Result.BuildDiagnostic carries the stable go-v1 boundary code so guidance names CURATOR_GO, GOROOT and the tested families without suggesting PATH or a download.

Fixed a legacy regression: schema-1 markers were reported unsupported-marker since marker v2 landed.

Gates (real exit codes, standalone): gofmt 0, go build 0, go vet 0, go test ./... 0 (40 packages), golangci-lint v2.4.0 0 issues, go test -race ./internal/install -timeout 40m 0, go test -race ./cmd/curator ./internal/godriver 0, GOOS=windows/linux build 0, GOOS=linux vet 0. Expected red: GOOS=windows go vet exits 1 on internal/runtimestore (pre-existing and sibling-owned; identical on the untouched base; 0 when that package is excluded).

Carry-forward findings in the attached notes: compiled installs are not idempotent (stageNode passes no BuildCurrentness), global status is not build-aware, internal/install needs -timeout 40m under -race, and one non-reproducible registry snapshot-skew flake.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260728-5b614c, pid=53894, exit=0)
REVIEW CYCLE 1 DIRECTIVE: Independently review the CLI diagnostics/currentness/repair implementation; do not edit artifacts. Verify every task AC against the accepted composed base and challenge semantic claims, not just tests. Require one closed stable-code vocabulary with exact human/JSON agreement; status --check must be zero only for exact current and fail closed for every unknown/invalid/unsupported/missing/corrupt/drifted/racy state. Audit classification ordering and prove a logical-key mismatch can truthfully distinguish source, target, toolchain, command/build-root/policy causes rather than mislabeling several inputs. Verify cache-hit evidence binds receipt/path/hash, marker before/after race detection covers every relevant scope, currentness cannot promote a stale skill, transitive compiled commands are represented, referenced entries remain GC-safe, and repair through install/upgrade is gated, atomic, rollback-preserving, and never claims mutation during dry-run or failure. Fuzz/adversarially inspect RedactDiagnostic for Unix/Windows/UNC paths, URI/embedded path forms, invalid UTF-8, controls/zero-width characters, multiline and 240-rune bounds at both status and install surfaces. Check legacy schema acceptance cannot admit forged build state or regress old output/exit behavior; inspect hidden TestMain worker dispatch for production-equivalent identity checks and no visible/bypass command. Re-run focused unit/integration/race tests without treating increased timeout as correctness; validate macOS evidence, native Windows cross-compiled scope and honest missing git/Go limitations, and Linux crossbuild only. Adjudicate the reported repeated-install non-idempotence and missing global build-aware status against this task scope; reject if an AC-required workflow remains uncovered, otherwise require tracked follow-up. Publish TASK-260720-1nlmvv_review-verdict-cycle-1.md with ACCEPTED or exact CHANGES REQUESTED and route accordingly.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260728-13915e, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260728-13915e)
REVIEW CYCLE 1 VERDICT: CHANGES REQUESTED. Blocking evidence is attached as TASK-260720-1nlmvv_review-verdict-cycle-1.md. Real corrupt and unsupported outcomes short-circuit before stable human/JSON build rows; unknown marker schemas and non-go-v1 marker drivers are rejected before their documented codes can be emitted; corrupt cache state is refused rather than atomically repaired; opaque logical-key drift is over-labelled as target/toolchain; token-only redaction leaks embedded and URI absolute paths and raw blocking errors bypass it. Independent focused normal/race/vet/format/diff gates pass, so this is semantic rework rather than a mechanical test failure. Route to implementation rework.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260728-13915e, pid=24079, exit=0)
REWORK CYCLE 2 DIRECTIVE (review cycle 1): Fix the four semantic blockers in the preserved implementation worktree; keep the green mechanical work and do not substitute classifier-only tests for production reachability. (1) Add a read-only diagnostic planning path for status that never mutates and converts real corrupt/unsupported/cache/toolchain/marker failures into one build row per affected command for both human and JSON output; ordinary status may report them without aborting, while --check alone fails nonzero. Use a two-stage untrusted marker envelope/read path so unknown schema and unsupported driver are diagnosable without trusting or activating their payload, or remove any code that cannot be made real; every documented code needs an end-to-end persisted-state test through the actual CLI. Missing/malformed/incompatible Go must also produce the stable machine row plus bounded guidance, not stderr-only output. (2) Make corrupt receipt and artifact state repairable by install and upgrade only after all gates. Build a replacement privately, preserve the live install and referenced cache entry until atomic commit, quarantine/replace under the manager lock, and roll back both cache and install on every injected failure. Add real install+upgrade E2E for corrupt receipt/artifact success and failure, including concurrent GC/currentness. (3) Stop labeling every opaque key mismatch as target/toolchain. Compare only facets actually persisted and emit specific codes only with direct evidence; use a stable non-attributing build-input-drift code for unresolved differences, with adversarial cases for build root, command, source_dir, target, toolchain and policy. Do not change frozen wire schemas silently. (4) Replace whitespace-token redaction with bounded scanning that catches embedded key=value, file:// and other URI forms, Unix, Windows drive and UNC paths, quotes/punctuation, invalid UTF-8, controls/format runes and multiline data. Apply the same function before every status, plan, Result.Errors, repair notice and toolchain detail surface; add leak canaries asserting no raw absolute path survives. Add real transitive compiled-status coverage. Explicitly document global-status and repeated-install boundaries; if existing task AC can support global build currentness without inventing a new CLI contract, include it, otherwise leave no false claim and use the orchestrator-owned follow-up. Re-run focused normal/race/lint/build/macOS and scoped native Windows tests plus Windows/Linux cross gates; do not repeat full 12-minute race suites when targeted coverage suffices. Hand off revised evidence only after production-path tests prove each stable state, atomic repair and redaction. No commit/stage/publish.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260728-5fd800, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260728-5fd800)
Cycle 2 rework landed against the CHANGES REQUESTED verdict. (1) Reachability: cmdStatus no longer discards compiled diagnostics when the read-only plan fails but still derived commands (legacy path unchanged when it did not); planBuilds records a toolchain-unavailable inventory row per active command with the stable go-v1 boundary code; markerRefusal separates unsupported-marker and unsupported-build-driver from invalid-marker without weakening marker.Read. (2) Corrupt cache is now a buildable miss, so install and upgrade repair it through quarantine-and-replace; blocking is now only unsupported and toolchain-unavailable. (3) build-target-mismatch became build-input-drift plus stable cause subcodes (build-root, target, unattributed) with an adversarial pass over every logical input component. (4) RedactDiagnostic rewritten as a bounded embedded/URI/Windows/UNC scanner and Result.failBuild routes every build-phase failure through it. Follow-ups filed: TASK-260729-2kaopg (global status build awareness, explicit exclusion documented in README) and TASK-260729-3jku56 (compiled install idempotence).

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260720-1nlmvv_spawn-log_-implementer--developer--claude-_RUN-260728-5b614c.log](file://TASK-260720-1nlmvv/TASK-260720-1nlmvv_spawn-log_-implementer--developer--claude-_RUN-260728-5b614c.log) — System spawn log captured by task-board
- [TASK-260720-1nlmvv_implementation-notes.md](file://TASK-260720-1nlmvv/TASK-260720-1nlmvv_implementation-notes.md) — Design, task-only delta, gate evidence including native Windows, acceptance-criteria mapping, and carry-forward findings
- [TASK-260720-1nlmvv_gate-evidence.tar.gz](file://TASK-260720-1nlmvv/TASK-260720-1nlmvv_gate-evidence.tar.gz) — Raw gate logs: full go test, race runs, GOOS=windows vet on this tree and the base, and native Windows runs including the expected-red git-absent full-package run
- [TASK-260720-1nlmvv_spawn-log_-reviewer--reviewer--codex-_RUN-260728-13915e.log](file://TASK-260720-1nlmvv/TASK-260720-1nlmvv_spawn-log_-reviewer--reviewer--codex-_RUN-260728-13915e.log) — System spawn log captured by task-board
- [TASK-260720-1nlmvv_review-verdict-cycle-1.md](file://TASK-260720-1nlmvv/TASK-260720-1nlmvv_review-verdict-cycle-1.md) — Independent reviewer verdict cycle 1 with semantic findings and gate evidence
- [TASK-260720-1nlmvv_spawn-log_-implementer--developer--claude-_RUN-260728-5fd800.log](file://TASK-260720-1nlmvv/TASK-260720-1nlmvv_spawn-log_-implementer--developer--claude-_RUN-260728-5fd800.log) — System spawn log captured by task-board
