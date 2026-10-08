# TASK-260927-25hk87 — flip-security-posture-default-hardened

Landing remains held until the operator schedules the B release. Changes remain uncommitted in the Story worktree.

## Implementation and logbook

Revision A is present in release tag v0.15.0-rc.3 (SecurityPostureRevision=A), satisfying the release prerequisite. Revision B selects hardened defaults for schema-2 config.Load callers. Schema 1 remains permissive. Explicit permissive keeps advisory audit/registry policy, drop for transitive system modules, optional source signers, empty allowlists and absent passthrough defaults, with one operation-scoped warning. Explicit and locked knob values retain precedence. The migration hint no longer promises a future B flip.

Removed all four revision-B deferrals from both posture vector consumers. No security-posture/vectors gap ledger rows existed; the remaining Codex seed rows belong to a different task and were retained. All six B vectors are driven without substituting a posture. Published historical implicit-A vectors remain explicitly bounded rather than being relabeled as driven after input changes. The config consumer also bounds registry I/O, which production CLI tests cover separately. Planned exact counts: config 12 driven / 5 bound / 0 gap / 0 skipped = 17; CLI 13 driven / 4 bound / 0 gap / 0 skipped = 17.

Production entry points tested: config.Load; run() status/env status; run() install/update/upgrade -> Config.CheckSecurityPosture; profile install -> envprofile's source/MCP/passthrough gates; env resolve and env status --check. The signed MCP fixture supplies genuine SSH-signed Git tags for the default source-signature requirement before checking the independent empty-MCP-allowlist refusal.

The hosted gate's snapshot commit message previously included an absolute checkout path. Removed that path from the message to honor the repository's public-material rule before invoking the configured hosted gate. No LOGBOOK.md or CHANGELOG.md edits; this section carries the task logbook record.

## Validation in progress

Initial development status command: exit 1, missing estimate. Added estimate 3, then development transition exited 0.
First test lock attempt: exit 75, shared build lock busy; no tests ran.
First posture test execution: exit 1. Config passed; the newly enabled CLI MCP refusal exposed missing signer evidence in its fixture. Corrected fixture without disabling production signature enforcement.
Current corrected targeted validation is running with GOFLAGS=-work and the shared build lock. Further evidence will be appended before handoff.

## Verification checkpoint

All validation commands ran as standalone processes, without tee. GOFLAGS=-work was retained and every Go build/test/lint used the shared build lock.

- Corrected initial posture command (config plus CLI): exit 0.
- Full config package at the pinned rc.14 root: exit 0 (3.858s). Manager-config schema cases 107/107 driven; manager-config vectors 56/56 driven. Posture 12 driven, 5 bound, 0 gap, 0 skipped, 17 total. Generic manager-config expected output selects only B defaults for absent knobs; inputs and explicit values stay intact, with a regression test preventing explicit-value replacement or mutation of corpus expectations.
- Current posture-only CLI command, including the installed MCP-bearing status fixture: exit 0 (73.578s). Posture 13 driven, 4 bound, 0 gap, 0 skipped, 17 total. All six B cases driven. Both consumers now assert exact tally classes, so reintroducing a bound or skip fails.
- CLI build: exit 0. Initial build lock attempt exited 75 without starting the build; the retry ran green.
- golangci-lint for config and CLI packages: exit 0, 0 issues.
- Compiled executable env-status smoke assertions: exit 0. Both real CLI operations exited 0, proving omitted posture -> hardened/profile and explicit permissive -> permissive/explicit, all four posture-default knobs, and warning counts 0/1.
- sh -n scripts/remote-gate.sh and git diff --check: exit 0.

The broader local config/CLI run exited 1: config expectations were still A-specific, CLI fixtures lacked explicit permissive, and the CLI exceeded its eight-minute timeout. Those assertions were corrected. A second broad CLI mask exited 1 on its five-minute timeout in unrelated unreadable-state status work, with no assertion failures before timeout; it is not claimed green. Per-task posture and full config checks were rerun green instead. The initial multipackage command also exposed a pre-existing golden sensitivity to the Go action directory (b237 instead of b001); no product change or suppression was added for that run shape. The hosted full suite is the arbiter.

Explicit permissive was added to legacy permissions/isolation/config/import test fixtures that need the former defaults. Warning strings in first-run expectations and profile-update golden outputs now match the B migration hint.

An extra physical-host `env status --check` probe under permissive exited 1; a proposed current-state control assertion also exited 1. Installed adapter versions differ from recorded versions on this host, so this probe does not establish an isolated posture-status property. The successful compiled smoke checks claim only rows, defaults, and warning counts. This limitation is recorded rather than inferred away.

Mutation checks have not yet started because bounded shared-lock waits exited 75. No mutant source has been left in place. Hosted qualification for exact tree ca3fe00f07e70735fb344d8c0237adb885c8cbfe is running at https://github.com/relux-works/curator/actions/runs/37229609872 (snapshot 662918dd336a1420276f6ddd17dd4c05b8fcf7c9). Lint, naming, interop and all three gate self-tests are green; test/race lanes remain pending. Landing remains held.

## Review handoff evidence

Hosted qualification is green: https://github.com/relux-works/curator/actions/runs/37229609872. All 11 required jobs passed: Test on Ubuntu/macOS/Windows, Race on Ubuntu/macOS, lint, naming, interop, and the three gate self-tests. The workflow deliberately skipped the rose-air and alternate candidate-suite lanes on this gate branch. The standalone hosted-verdict verification exited 0 and confirmed that the current candidate tree is exactly ca3fe00f07e70735fb344d8c0237adb885c8cbfe, the qualified snapshot tree. Hosted verdict JSON and verification log are attached. The temporary gate branch was removed; the Story branch was not committed or moved.

The standalone mutation harness exited 0 after rejecting both mutants. Each actual mutated go test exited 1 (expected-red): reverting the release default to A fails the schema-2 default tests; narrowing the source refusal from all empty lists to nil-only fails all four explicit-empty production entry points (install, update, upgrade, profile install). Production source was restored byte-for-byte and is part of the verified hosted tree.

The focused changed-fixture CLI suite exited 0 (40.693s), including credential isolation, permission rendering, isolation system locks, effective passthrough, env config, first-run help output, and locked-profile installation/update behavior. New posture tests and full config tests were run locally as recorded above. Repeated local golden-only attempts exited 75 before test execution because the shared build lock was busy; those golden checks are accepted from the green hosted full test lanes, rather than claimed as a local rerun. The earlier timeout attempts remain recorded as exit 1.

The attached redacted validation log includes the green local results and the actual expected-red mutation outputs. Synthetic user-directory test labels were normalized as well. No LOGBOOK.md or CHANGELOG.md was edited; the task logbook is carried here. There are no remaining revision-B posture deferrals or posture gap ledger rows; unrelated Codex seed ledger rows are unchanged. Exact posture accounting remains config 12 driven / 5 bound / 0 gap / 0 skipped = 17 and CLI 13 driven / 4 historical-A bounds / 0 gap / 0 skipped = 17; six of six B vectors are driven.

Code is ready for review. LANDING IS HELD until the operator schedules the B release. Changes remain uncommitted in the Story worktree.

## Converged base republish (2026-10-05)

Converged; remote-gate hunk dropped (on main via N4). Orchestrator converged the workspace onto fresh main; snapshot refs/campaign/postureB-snapshot-20261005. Verified `git diff --stat` shows the posture-B changes (18 files) and does NOT touch scripts/remote-gate.sh. No code edits in this republish run; no LOGBOOK.md or CHANGELOG.md edits. Code remains ready for review; LANDING HELD.

## Converged-tree local verification (2026-10-05, republish run)

The republish instruction did not anticipate the four appended checklist items (8-11) now gating handoff, so this run verified them directly on the converged tree instead of handing off blind. No product code was edited; no LOGBOOK.md, CHANGELOG.md, or scripts/remote-gate.sh changes. Conformance root: curator-spec conformance/v1 at SPEC_PIN 43bf0a2506d5c354a73bbc3ea4623d4653db10c7 (clean). Every command ran standalone under the shared build lock with GOFLAGS=-work; each reports its real exit code.

Reran on the converged tree (all exit 0, 0 skips, 0 failures):
- Full config package with root: exit 0 (1.053s). TestSecurityPostureVectors 17/17 subtests pass; the in-test exact tally (12 driven / 5 bound / 0 gap / 0 skipped = 17) holds.
- CLI vectors with root (TestSecurityPostureVectorsThroughCLI): exit 0 (35.079s), 17/17 subtests pass; in-test tally 13 driven / 4 historical-A bounds / 0 gap / 0 skipped = 17 holds. Six of six B vectors driven.
- Full CLI posture mask (SecurityPosture + registry-outage tests): exit 0 (52.948s).
- Changed-fixture CLI subset (28 tests: credential isolation, permissions, locks, effective passthrough, env config, first-run, locked-profile install/update): exit 0 (35.622s), 28/28 pass.
- CLI build: exit 0. golangci-lint on internal/config + cmd/curator: exit 0, 0 issues.
- Compiled-binary smoke on a fresh converged-tree build: exit 0. Omitted posture -> hardened/profile with 0 warnings; explicit permissive -> permissive/explicit with 1 warning; all four posture-default knobs verified.
- git diff --check: exit 0.

Accepted from already-attached evidence (not rerun): hosted gate verdict 11/11 green on the pre-converge tree (attached verdict JSON + verification log); mutation narrowing proof (both mutants rejected, expected-red exit 1). The converged tree differs from the qualified tree only by dropping the remote-gate hunk that main now carries via 1fc0a93a; the fresh hosted CR gate on the new revision is the reviewer's arbiter per the review note.

Checklist items 8-11: (8) implementation matches AC — A shipped in rc.3, B vectors 6/6 driven through production entry points, gap rows 0 with exact counts asserted in-test; (9) fits architecture — 18-file diff, one product file (revision const A->B + hint text), CodexSeedRevision unchanged, no collateral revisions; (10) tests green per the reruns above; (11) no review rejection verdict exists (reviewer spawn failed on trunk-overlap infra, not a verdict), so no verdict evidence or rerouting is owed.

Code is ready for review. LANDING IS HELD until the operator schedules the B release. Changes remain uncommitted in the Story worktree.

## Re-applied after converge (2026-10-08, revision 3)

Revision 2 was ACCEPTED and went stale only because main advanced on `cmd/curator/env_credential_marker_test.go` (the v1-only NUL gate, 7d809c5f). The orchestrator converged the Story workspace onto fresh main (snapshot refs/campaign/STORY-260928-oflbe1-snapshot-20261008); every other delta file was carried unchanged and the conflicting file was reverted to fresh main. This run re-applied the revision-2 intent for that file from `postureB-conflict-delta.patch`: in `TestEnvResolveCredentialRecordIsolatedKeychain`, the `writeMachineConfig` schema-2 fixture now carries explicit `"security_posture": "permissive"` (one line). The patch pre-image matched the new file verbatim — same test, same line, same surrounding context — so intent and hunk coincide here.

What changed versus revision 2: nothing semantic. The re-applied line is identical to the revision-2 hunk; all of main's new v1/v2 NUL-gate assertions are intact and untouched. The new tests use bare `profileHome` (zero Config, so `EffectiveSecurityPosture` falls back to permissive), while the revision-B flip applies at load time via `defaultSecurityPosture` only to schema-2 files through `config.Load` — which is exactly why the `writeMachineConfig` line needs the explicit permissive and the new tests need nothing.

Sanity check (standalone, under the build lock): `mini-build-lock run postureB -- env GOFLAGS=-work go vet ./cmd/curator ./internal/envprofile` → exit 0. No cmd/curator tests run locally per R194; the hosted gate is the arbiter. No LOGBOOK.md, CHANGELOG.md, or scripts/remote-gate.sh edits. Final diff: 18 tracked files + 1 untracked (`internal/config/security_posture_test.go`) = 19 paths, matching revision 2's path count.

Code is ready for review. LANDING IS HELD until the operator schedules the B release. Changes remain uncommitted in the Story worktree.
