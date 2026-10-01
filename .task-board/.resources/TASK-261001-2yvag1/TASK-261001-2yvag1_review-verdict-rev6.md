# TASK-261001-2yvag1 — curator-muse-environment: revision 6 review

Verdict: **changes_requested → to-dev**. Three findings require another producer/reviewer cycle. No external or human-only blocker.

Reviewed candidate tree `7b02cb7efa04470972c5af5018bb57ab4b993734`, base `e87d488b8fd892bc86c037ae5e325e596df8e745`. All 29 workspace delta files were compared byte-for-byte with that tree and matched. Normative source: local spec HEAD `d373078a27a437071c2416d447221e77cec65725`, environments sections 7.1/7.4/10.2, Decision 0018 and v3 schema; issue curator#100 read through gh. Candidate manifest SHA256 is `bd03456b92a7368d90ea74fe6953db10bc188588020683024a6a6b8735a40783`. Run goal queried: none.

## Findings

### R1 — P1: symlinked auth parent is accepted by both resolve paths

Location: `internal/envprofile/muse.go:58`, reached by `checkPassthrough`; `internal/envprofile/managed.go:2514` returns the current fragment before entering repair preflight.

Reproduction against otherwise unchanged candidate production sources: provision a synthetic Muse fixture, move `<managed-home>/config` to a temporary directory outside the managed home, and put a symlink at the original config path. Preserve its existing `muse/auth.json` link to the expected synthetic native regular file. Call production `Resolve` with Repair=false and Repair=true. **Both return success and a v3 document.** The added negative probe exits 1 with two assertion failures (0/2 required refusals). No credentials were read into the output.

`Lstat(full)` protects the final entry, but follows the intermediate config symlink. Because the terminal auth link still has the expected text, the home is classified current and even `--repair` bypasses write preflight. The emitted XDG_CONFIG_HOME therefore traverses an unowned outside directory. The binding review requirement that a symlinked auth path fail closed is not met. The existing config-parent-link test leaves the outside tree empty, so its refusal does not test this live-link bypass.

Fix: establish the managed parent boundary before accepting auth liveness, retain the intentional final credential file-link, and cover current bare resolve and current --repair with an outside parent containing a valid-looking auth link. Preserve outside/native bytes and emit no fragment on refusal. Reproducer source and output are in the evidence archive.

### R2 — P2: fragment corpus is not bound to production emission

Location: `cmd/curator/muse_test.go:267` and `:275` (static file read/schema.Validate), `:293` (direct CheckBoundary call).

The published 36 rows exercise static fixtures, the schema oracle and a boundary helper. **0/36 invoke Resolve.** The separate CLI smoke test exercises real resolution but only decodes selected fields and never validates the full emitted JSON against v3.

Independent proof: in the disposable copy, remove only Muse's required `permissions` member in `Fragment.Object`. Run `CURATOR_CONFORMANCE_ROOT=<d373078a>/conformance/v1 go test ./cmd/curator -run '^TestMuse' -count=1 -timeout=60s -v`. **Exit 0: the mutant survives**, including the CLI smoke test and all 36 schema rows, even though the emitted fragment violates the required schema. A preliminary mutation-script anchor assertion failed before any edit/test; it is not counted as mutation evidence.

The unmutated production Resolve document independently validates against the normative schema (exit 0). Thus this finding is a demonstrated regression-detection gap, not a claim that today's ordinary output is malformed. Bind permanent schema assertions to actual resolver output and satisfy the review's production-entry vector requirement; explicitly classify any reader-only rows instead of calling helper-only coverage production resolver coverage.

### R3 — P2: unrelated broker EPIPE changes remain in this CR

Location: `internal/testcli/cli.go:218` and `internal/testcli/cli_test.go:1`.

The candidate adds EPIPE suppression and dedicated broker tests. These are changes against the review base, even though the results describe them as unchanged carryover from revision 4. The binding briefs put the unrelated EPIPE flake in BUG-261001-2772iz and explicitly exclude broker/askpass work. Remove this delta from the resolver CR and route it through its owning bug. This is a scope finding, not a claim that the helper fix itself is incorrect.

## Review sweep

| Surface | Evidence and outcome |
| --- | --- |
| Adapter/layout | Exactly four ordered XDG parents; HOME is not emitted. Root-context/system-prompt/MCP remain unadmitted. Skills location and isolation limitations are explicit. |
| Profile seeds | Root snapshot sources, seed retention, auth-content refusal and secret detector scope reviewed; CLI refusal tests pass. |
| Auth observation/repair | All 16 published rows drive production StatusOf/Resolve and pass; additional errors, directories and fork cases pass. Reads use stateread with no new allowlist. Parent-link bypass is R1. |
| Credential confidentiality | inspectMuseAuth uses metadata/readlink and opens the native regular file only to establish readability; no credential bytes are read into output/logs. Reproducers use synthetic stores only. |
| Fragment | Ordinary Resolve output independently passes v3 schema; fake child preserves HOME/env/argv. Static 36/36 rows pass, but production binding is R2. |
| Status | Known absent optional Muse marker with known backups does not make current profiles non-current. Unreadable/malformed marker and unreadable backups remain non-current. |
| Conformance/refresh | Original base count rows remain unchanged; new rows are exact-digest scoped. Historical hash-v2 candidate gaps are explicitly owned/classified. Coverage package passes under candidate and rc.13. Other adapters retain the previous fragment revision. |
| Hygiene/scope | 29 paths match candidate; no LOGBOOK/CHANGELOG edits or stray candidate files; diff --check passes. Launcher remains separate. Broker carryover is R3. |

## Fresh verification

All listed test invocations use `-count=1`. Evidence archive includes logs and source identity.

| Reviewer command/scope | Real exit | Result |
| --- | --- | --- |
| Candidate envprofile: `TestMuse|TestManagerOwnedAbsenceReadsAreGuarded|TestEmptyAllowlistWarning`, timeout 90s | 0 | 68.883s; link states 16/16; no skips; guard 475/475 = 369 seam + unchanged 106 allowlisted |
| Candidate CLI: all TestMuse plus the four historical status failures and TestEnvStatusCheckCurrentScopeOnly, timeout 300s | 0 | 189.438s; schema fixtures 36/36; fake launch and secret/profile refusals pass |
| rc.13 envfragment/envregistry/stateread/conformancecoverage, timeout 60s | 0 | All four packages pass |
| Candidate conformancecoverage, timeout 60s | 0 | Pass |
| Supplemental production Resolve capture in disposable copy | 0 | Captures the actual emitted document |
| Python Draft202012Validator with all local spec schema resources | 0 | Captured unmutated v3 document validates |
| HOME emission mutant, production TestMuseDirectoriesAndIsolation | 1 | Killed: HOME is not registry-declared |
| Raw os.Lstat/absence seam-bypass mutant, source guard | 1 | Killed: 474/475; names inspectMuseAuth as the unguarded read |
| Supplemental symlinked config-parent negative probe | 1 | Genuine candidate defect: bare resolve and repair both emit |
| Missing-permissions emission mutant, all CLI TestMuse | 0 | Survived: evidence for R2 |

Mutants and supplemental test files exist only in `/tmp/TASK-261001-2yvag1-review-copy`. Each altered production source was restored from its captured original. No worktree code was changed.

The six previous failures now pass for these reasons:

- `TestEnvStatusMissingAndUnreadableKeepRecord`: synthetic registered environments are provisioned, leaving approval-record failures as the measured posture.
- `TestEnvStatusRegistryBoundaryPostureAndCheck`: the same complete control provisioning isolates the registry boundary/posture check.
- `TestEnvStatusReportsShellHookTrustPosture`: synthetic Muse provisioning keeps the baseline current while shell-hook trust changes.
- `TestEnvStatusUnreadableApprovalStateSurfaced`: the shared provisioned control isolates the unreadable approval state.
- `TestManagerOwnedAbsenceReadsAreGuarded`: Muse lstat/readlink are routed through stateread; independent seam mutant proves detection, no allowlist addition.
- `TestEmptyAllowlistWarningLeavesCurrentStatusCurrent`: all registered homes are provisioned, so the empty allowlist warning remains advisory. Separately, `TestMuseStatusOptionalProvisioning` and unchanged CLI `TestEnvStatusCheckCurrentScopeOnly` prove never-enabled Muse remains optional; the fixture changes do not substitute for that semantic coverage.

## Reused evidence and limits

The board's revision-6 validation log records hosted run 36862485192, exit 0: Linux/macOS/Windows tests, Linux/macOS race, lint, naming, interop and gate self-tests green. Its gate commit `ef9f25b2a21b3a17d0ebaf25d09ecdd73b141768` resolves to the exact reviewed tree. Those broad checks were accepted from attached evidence, not rerun by this reviewer. Hosted candidate-suite lanes were skipped, so candidate-vector claims above use the fresh local runs. Posture/help checks also have producer revision-6 evidence and unchanged trunk behavior; no fresh reviewer completion is claimed for the initially interrupted posture/help command.

Initial simultaneous test attempts stalled during an execution-host delay and were interrupted with exit 130. They are not passes or killed mutants and are superseded only for the explicitly rerun scopes above. No real Muse session, launcher integration run, or full reviewer race matrix was attempted. No claim of refresh/session race freedom or native surface discovery is made.

Logbook entry text (kept here per the task's no-LOGBOOK rule): revision-6 review found a live auth-link beneath a symlinked config parent is accepted by both resolve modes, and a missing-permissions emission mutant survives the Muse suite. Two required mutants were independently killed, all six prior regressions passed, and the task is routed to rework with reproducible evidence.
