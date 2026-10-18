# BUG-261018-b2jnyg: producer evidence and bounds

Candidate: `7344e4d3b6fc042b2ac7f8b3c950e85eb9b7c79d` (detached signed snapshot, signature `G`), tree recorded in candidate-tree-01.txt. Story HEAD remains checkpoint `2e5ab6dab61fc78a525cbcb0de6c5fcc9f5ae882`; all task changes remain uncommitted. This checkpoint was two commits behind local main when inspected. Evidence applies to this candidate, not to newer trunk.

## Governing specification

Read curator-spec `protocol/environments.md` §7.4 at `fe2d1c6470a0cd8e69b3bbc9b0f9934fe1a1433b`, lines 1533–1546 and 1593–1605. Seeds become tool-owned after provisioning; Claude's per-launch project entry is the named repair exception. No requirement makes post-first-run `.claude.json` CCJ-1. CCJ-1 remains required for manager-authored locks/fragments/MCP surfaces; those do not change. No unresolved spec question blocks this change.

## Requirements and production coverage

**13 of 13 behavioral AC rows driven** through `Resolve`, using temporary homes only. `Resolve` → `repairUnderLock` → `finalizeMarker` → `claudeSeed` (`internal/envprofile/managed.go:1412`); verification calls `claudeProjects` at `managed.go:2204`. The direct CLI subprocess/parser is outside this measured ratio; the changed contract is the production Resolve repair/verification entry, not CLI parsing.

| Row | Requirement | Named test | Production implementation |
| --- | --- | --- | --- |
| 1 | Real-world fractions, exponent text, large integers, negatives, nested arrays and unknown values preserved | `TestClaudeRepairPreservesNumberText/*/new-launch` | `managed.go:868`, `managed.go:884` |
| 2 | Required flags repaired; unowned launch-entry members preserved | `TestClaudeRepairPreservesNumberText/*/{existing-launch,missing-trust,missing-approval}` | `managed.go:846` |
| 3 | Satisfied file bytes, inode and mtime unchanged during other-surface repair | `TestClaudeRepairSatisfiedIsNoOp` | `managed.go:862` |
| 4 | Float within launch entry retained | `TestClaudeRepairPreservesNumberText/*/existing-launch` | `managed.go:868` |
| 5 | Missing projects becomes an object | `TestClaudeRepairPreservesNumberText/*/missing-projects` | `managed.go:826` |
| 6 | Non-object projects retains legacy replacement behavior | `TestClaudeRepairPreservesNumberText/*/non-object-projects` | `managed.go:826` |
| 7 | Non-object/malformed root refuses without leaf replacement or byte changes | `TestClaudeRepairRefusesUnsafeState/{array,null,scalar,malformed}` | `managed.go:821` |
| 8 | Symlink refuses without changing link or target | `TestClaudeRepairRefusesUnsafeState/symlink` | `managed.go:2734`, `nofollow.go:294` |
| 9 | Non-regular leaf refuses without changing the leaf | `TestClaudeRepairRefusesUnsafeState/directory` | `managed.go:833` |
| 10 | Unreadable regular leaf refuses without changing bytes/mode/inode/mtime | `TestClaudeRepairRefusesUnsafeState/unreadable` | transaction backup via `readManagedRegular`, before seed merge |
| 11 | Absent-file seed retains pre-fix golden bytes for both forms | `TestClaudeAbsentSeedGolden` | `managed.go:871`, `managed.go:896` |
| 12 | Verification accepts floats and exponent outside float64 range anywhere | `TestClaudeRepairPreservesNumberText/*/{new-launch,overflow}` (second Resolve with Repair=false) | `managed.go:910` → `decodeJSONObject` |
| 13 | CCJ-1 regression plant rejected for intended encoding reason | `TestClaudeSeedNarrowingMutants/whole-object-ccj` | overlay of present-file encode, then real Resolve |

`*` covers referenced and monolithic forms. All previously committed reviewer tests remain under their original names; no existing test was edited. Whole-package and full-suite execution outcomes are recorded below.

## Mutant evidence

Every nested `go test` below exits **1 (expected failure)**; the parent harness exits **0** only after finding both the named test failure and its intended diagnostic. Setup/compilation failures are not accepted. All six plants were killed on the candidate. No survivor is claimed as a kill.

| Mutant | What it narrows the gate to | Named failing test | Result / survival bound |
| --- | --- | --- | --- |
| `whole-object-ccj` | Restores integer-only CCJ-1 encoding for present tool state | `TestClaudeRepairPreservesNumberText/referenced/new-launch` | Killed: shortest-form base-10 integer refusal |
| `normalize-numbers` | Preserves the `decoder.UseNumber()` source token but normalizes numbers through float64 | `TestClaudeRepairPreservesNumberText/referenced/new-launch` | Killed: unowned values/number text changed |
| `admit-untrusted-approved-entry` | Skips repair for an approved entry lacking trust | `TestClaudeRepairPreservesNumberText/referenced/missing-trust` | Killed: required trust flag unchanged |
| `admit-trusted-unapproved-entry` | Skips repair for a trusted entry lacking includes approval | `TestClaudeRepairPreservesNumberText/referenced/missing-approval` | Killed: production post-repair verification reports missing approval |
| `rewrite-satisfied-referenced-entry` | Applies the no-op only to monolithic homes | `TestClaudeRepairSatisfiedIsNoOp/referenced` | Killed: satisfied file rewritten |
| `admit-null-object` | Admits null as an empty object while retaining other invalid-object refusals | `TestClaudeRepairRefusesUnsafeState/null` | Killed: missing seed-unreadable refusal |

The normalization plant runs the behavioral suite, so retaining a searched-for token cannot count as correctness. No new source-inspection gate is introduced. Existing symlink/type/read guards are unchanged; their rejection and state-preservation tests are retained above.

## Sweep of tool-owned encodes

Searched all production `internal/` JSON/TOML encodes and all Claude/Codex/pi/Muse/opencode file references, then inspected ownership and callers. Scratch inventories: `encode-sweep-01.log`, `tool-files-sweep-01.log`.

| Harness / surface | File and line | Finding / action |
| --- | --- | --- |
| Claude `.claude.json` | `internal/envprofile/managed.go:810`, `:868`, `:884`, `:910` | Only existing tool-owned JSON read/merge/re-encode in envprofile; fixed numeric lexeme preservation and satisfied no-op; verification shares UseNumber reader |
| Codex native `config.toml` revision B | `internal/envprofile/managed.go:679`, `:689`, `:746` | Provision-only TOML parse/encode removes manager-forbidden MCP table per §7.4. TOML supports fractional values; no CCJ-1 conversion. May normalize TOML formatting/number spelling at initial seed. Existing homes never refreshed; no same float-refusal class. Left unchanged |
| pi `settings.json`, `models.json` | `internal/envregistry/envregistry.go:335`; `managed.go:701`, `:1294` | Raw provisioning copy only; repair preserves tool bytes. No re-encode |
| Muse settings/trust | `internal/envregistry/envregistry.go:239`; `managed.go:701`, `:1294` | Raw provisioning copy only; repair preserves tool bytes. No re-encode |
| opencode config-home seeds | `managed.go:777`, `:1419` | XDG symlink reconciliation; no JSON value rewriting |
| opencode managed `opencode.json` | `internal/contextmaterialize/referenced.go:164`, `:172` | Manager-authored instructions-only surface. CCJ-1 is required, not a tool-state merge |
| Claude/opencode MCP launch files; Codex MCP layer | `internal/contextmaterialize/mcp.go:265`, `:293`, `:299` | Manager-authored profile-derived objects/TOML only; no tool-owned input re-encoding |
| Native harness MCP/settings inspection | `internal/mcp/mcp.go:1`, `:24`, `:34`, `:44` | Read-only inspection; no write path |

No other harness tool-owned file re-encode through CCJ-1 was found. Manager-authored markers, migration journals, source manifests, profile deltas, snapshots, and build/package state were excluded by ownership, not silently treated as tool files. Hybrid Skillfiles in `internal/scopes/hybrid.go:191` and `:257` are manager manifests, not harness-owned state.

## Explicit bounds and brief gaps

- The brief contains **no surface table**. This is a brief gap; no catalog-derived coverage map is claimed or silently waived. The AC-derived map above describes actual production coverage.
- A necessary present-file merge uses standard JSON encoding: key order, whitespace and string-escape spelling may change. Decoded values/types and number lexemes are preserved. Exact text preservation is not attempted because that requires a JSON concrete-syntax editor rather than a semantic object merge. Already-satisfied bytes remain exact.
- Replacing missing/non-object `projects` and a non-object launch entry follows the brief's explicit instruction to retain current behavior. Such a value is not preserved; this is the named legacy exception, not an unreported coverage claim.
- The invalid leaf checks cover non-regular directories, not every OS special-file kind or a concurrent filesystem attacker. No filesystem safety gate was weakened.
- POSIX mode-000 unreadability is exercised on the non-root macOS host; Windows/root skip it with existing classified reasons. The symlink-dependent fixture uses the repository's established host-capability probe.
- No behavior row is otherwise treated as out of contract. CCJ-1 implementation/spec changes and other harness rewrites are excluded by the brief's Scope clause; the sweep still lists them above.
- Diff stays in `internal/envprofile` plus the explicitly authorized short note `docs/claude-seed-repair.md`. No root README, board files or control-root logbook are edited. This task-scoped outcome carries the logbook findings.

## Validation ledger

- Initial pre-fix targeted suite: exit 1, expected regression red; absent seed golden passed. Early overflow fixture additionally exposed float64 range rejection. Initial unreadability assertion expected seed_unreadable but production transaction backup uses `environment_repair_failed: permission denied`; corrected the test to pin the existing diagnostic.
- Targeted behavioral suite, regression-after-02.log: exit 0 (before adding adjacent flag cases; superseded by the complete candidate inventory below).
- Initial mutant harness: exit 1 because the missing-approval plant failed at post-repair verification rather than the expected comparison. Corrected only the harness diagnostic expectation. Candidate mutant harness, mutants-02.log: exit 0; all six nested tests exit 1 for the required reason.
- Build: `go build -o .temp/BUG-261018-b2jnyg/curator ./cmd/curator`, exit 0.
- Lint initial shared cache: `golangci-lint run`, exit 1, stale diagnostics naming deleted worktrees. `GOLANGCI_LINT_CACHE=$PWD/.temp/BUG-261018-b2jnyg/lint-cache golangci-lint run`, exit 0, 0 issues. No suppression added.
- `go vet ./...`: exit 0. `git diff --check`: exit 0.
- Environment: darwin/arm64, Go 1.25.5; GOTOOLCHAIN=local, GOENV=off for full tests; Rust 1.91.0, Node 25.6.1, npm 11.10.1, Swift 6.4; pnpm 10.33.0 installed task-locally with `npm install --prefix .temp/BUG-261018-b2jnyg/pnpm --no-save --ignore-scripts pnpm@10.33.0` (exit 0).
- Conformance: exact CI pin `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`; manifest SHA-256 `6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5`. Local suite-plan exit 0: 85 served, 0 deferred, 0 excluded.
- Full local coverage is green across the bounded commands and exact-name inventory below. Hosted run [39112682161](https://github.com/relux-works/curator/actions/runs/39112682161), attempt 2: **success**, exact signed candidate `7344e4d3b6fc042b2ac7f8b3c950e85eb9b7c79d`.

### Local verification diagnostics

- Initial CLI groups 1 and 2: exit 1. The host PATH exposed stale provider executables, so unrelated status tests reported `subcommand_provider_untrusted`. The named legacy status test also failed with an overlay restoring the exact checkpoint `managed.go` (exit 1). CLI groups are rerun with only the operator-local bin directory removed from the child PATH; no user executable or global setting is changed.
- Initial spec fixture setup used `git archive` of conformance/v1 only: sibling skillfile-sources corpus was missing. Adding the full archive exposed `export-subst` expansion in the byte-exact snapshot fixture. Replaced the fixture tree with `read-tree` + `checkout-index` of the exact pinned commit using a task-local index and destination, preserving raw tracked bytes. The conformance/v1 manifest digest remains exact. Crossconformance and interop/environments are rerun under the corrected fixture tree.
- Unsplit local envprofile run: exit 1 at its explicit 8-minute command budget (480.573s), no behavioral failure before timeout. 211 top-level tests had already passed/skipped legitimately on the candidate; preserve that complete log and reuse those exact test verdicts. The remaining 163 of 374 listed top-level tests run in five bounded groups. This timeout is not a green package result, and its interrupted test is included in the remaining masks.

### Continued bounded verification

- CLI clean-PATH groups 1–2: exit 0, all 100 top-level tests pass. Clean-PATH groups 3–4 reached the explicit 480s budget (exit 1), with 37 named top-level tests passed and no behavioral failure; the remaining CLI inventory is split into 20-test masks. All red invocations remain in the execution ledger.
- Install initial group 3 reached the 480s budget while a compiled-build case waited behind the CLI's process-wide host GOROOT lock (exit 1). Its 31 completed named tests are retained; the remaining 19 reran without a concurrent CLI lane and passed, exit 0. Subsequent install masks run separately from CLI.
- Envprofile remaining groups 1–5: each exit 0. Together with the 211 complete testcase verdicts from the explicit-timeout attempt, the exact source/test/configuration identity has **374 of 374 top-level tests accounted for: 373 pass, 1 existing filesystem-capability skip (`TestFoldedPathsAreSourceInvalid`)**, no missing test name. This is an inventory of actual completed test verdicts across bounded commands, not a claim that the original timed-out package command passed.
- All 82 other packages have successful full-package command results (including two no-test packages); the two initial fixture-related failing package invocations are superseded by complete green crossconformance and interop/environments reruns, exit 0. A byte comparison proves the sole archive-versus-checkout difference within conformance/v1 was `fixtures/byte-exact/subst.txt`; its restored raw digest is `ec9a6c8c260b3977dad2f1cba09414599879ac5c3e7353c3205a40d5fbe122bc`.

- Install inventory is now **363 of 363 named top-level tests accounted for**, with 361 passes and 2 existing platform skips. All bounded tail commands exit 0.
- CLI remainder group 4: exit 1 solely in `TestSignerRowsSurviveHostileGlobalGitConfig`: inherited `GIT_CONFIG_COUNT` overrides its generated fixture signing key with host configuration. Repeating that exact named test with `GIT_CONFIG_COUNT` absent in the child environment passes, exit 0 (`local-cli-signer-cleanenv.json`). No host key unlock, global config change, or source change was needed. Existing complete passes are retained; later CLI groups use that clean fixture environment.

### Hosted race diagnostic and retry

- Hosted attempt 1 macOS race job `117631946020`: go test exit **1**, platform-case gate exit **0**. Named failure: `TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted/artifact-link`; a production `git cat-file --batch` subprocess failed to spawn the runner's Git executable with permission denied. There was no reported data race. Raw served-test artifact is retained task-locally. Repository fixture comments already document this macOS transient spawn shape (`internal/install/gitfixture_test.go:4`), but this failing production command is outside that fixture retry. No unrelated source change or local-mirror substitution is made; request a real hosted retry of the same job at the exact signed candidate. Original failure remains evidence and is not called green.

- Hosted attempt 1 ended with 19 jobs successful, 2 intentional skips, and the one macOS race job failed as recorded above; Windows Test passed. Initial retry request exited 1 because GitHub rejects rerunning jobs while their workflow is still running (HTTP 403). After the workflow ended, `gh run rerun 39112682161 --failed` exited 0 and hosted attempt 2 started at the unchanged candidate SHA.
- Local install skips are `TestInstallHelperProcess` (helper-process sentinel) and `TestAcquireDraftNetworkWindowsRefusesBeforeAnyProcess` (Windows-only refusal probe). They are declared bounds, not executed Darwin behavior. The full Windows hosted lane passed in attempt 1.

- All five new test groups, including `TestClaudeSeedNarrowingMutants`, independently have named **pass** verdicts under macOS `-race` in the first hosted artifact. The overall attempt still failed as above; individual green evidence does not promote that job to green.

### Local coverage reconciliation

| Scope | Expected top-level tests | Accounted verdicts | Pass | Existing skip | Missing |
| --- | ---: | ---: | ---: | ---: | ---: |
| cmd/curator | 425 | 425 | 425 | 0 | 0 |
| internal/install | 363 | 363 | 361 | 2 | 0 |
| internal/envprofile | 374 | 374 | 373 | 1 | 0 |

All 85 packages in the exact pinned suite plan are covered: 82 have full-package green results; the three large packages above have complete named verdict inventories across bounded commands. No remaining test is inferred green from a timeout. The split is an execution-budget accommodation, not a reduced test selection. Existing completed verdicts were reused only at the unchanged candidate source/test identity. CLI remainder groups 1–15 each exit 0 except group 4 (exit 1, the isolated signing-environment failure recorded above); its sole failed test reruns green with the generated fixture key restored. Other group-4 named passes are retained. Final `git diff --check` exits 0.

The sanitized validation artifact records every retained top-level verdict, selector mask, command exit code, and raw-log SHA-256/size. Raw command output remains task-local to avoid publishing operator filesystem details. The absence of raw Output fields from that public artifact is intentional sanitization, not a claim that a missing/truncated log passed.

### Hosted gate and handoff evidence

Hosted attempt **2 is green** at the exact signed candidate, with **20 successful jobs and 2 intentional conditional skips**. The attempt-2 API lists all 22 jobs, so this is a full hosted attempt; no local mirror or manufactured status substitutes for it. macOS race job `117640220689` reports `go test overall exit=0` and `platform-case gate exit=0`. The original attempt-1 spawn failure remains recorded above and in the hosted evidence artifact.

`BUG-261018-b2jnyg_hosted-gate.json` preserves actual per-attempt job IDs, names, timestamps, conclusions, source identity, URLs and both race log hashes. `BUG-261018-b2jnyg_validation.json` preserves the local inventory and command ledger. Post-CI signature verification, remote gate-head equality, each changed file blob, Story checkpoint equality and `git diff --check` all exit 0. Story HEAD still trails local main by 2 commits; this evidence covers the stated candidate only. The four task files remain uncommitted for the managed Story handoff.

Non-object launch-entry replacement is unchanged code inspected during review, not a separately driven new regression case; the brief's required non-object `projects` case is driven. This is a stated neighboring-case bound. No remaining spec/product decision or validation blocker is asserted.


## Successor recovery: Change Request validation

The prior role handoff produced CR revision 1, whose configured hosted command
`sh scripts/remote-gate.sh` exited **1**. Hosted run
[39119290914](https://github.com/relux-works/curator/actions/runs/39119290914)
was bound to signed candidate `2c6e86caa53c1313a1caff0d1e0916b7497fa986`.
That candidate and the original tested candidate have the identical tree
`d1d16a8b497530c1c5ddca460fab693daf3bd069`; both signatures verify as `G`.
The four working-tree candidate blobs match exactly; the Story HEAD and
uncommitted delta are unchanged. No code or test rework was necessary.

Attempt 1 failed only at self-hosted runner `Complete runner` cleanup in Naming,
Lint, and Go 1.25.5 Linux driver jobs: permission denied removing a retained
read-only pnpm test store. Repository test/lint steps had passed. A real hosted
failed-job retry (command exit 0) recovered Lint and Go driver, but attempt 2
still failed Naming cleanup. A second real failed-job retry (command exit 0)
recovered Naming. **Attempt 3 is success: 20 successful jobs and 2 conditional
skips**, with the same exact signed candidate. The original gate command's exit
1 remains recorded; it is not rewritten as a successful command. No runner
service, cleanup policy, repository gate, or product source was modified.

Successor directly reran the full new Claude regression selection and all six
narrowing/plant controls (exit 0; each nested expected-red behavioral command
exits 1), the CLI build (exit 0), diff validation (exit 0), and signature/blob
identity checks (exit 0). All 57 retained raw local logs match their recorded
size and SHA-256, so the unchanged full local suite evidence above is reused,
not claimed as rerun. The existing 13 of 13 behavioral AC map, six-mutant
table, sweep, absent-surface-table brief gap and stated bounds remain applicable.
The new `BUG-261018-b2jnyg_recovery.json` outcome records actual hosted job
verdicts and successor validation. This entry is the task-scoped recovery
logbook. The managed producer handoff still owns subsequent CR construction;
this report does not claim that a future validation command has already run.

## Second successor recovery: CR revision 2

CR revision 2's configured `sh scripts/remote-gate.sh` exited **1** at
signed candidate `8bf861c356e69d3a11680445d78793b502e23dc9` (signature
`G`, unchanged tree `d1d16a8b497530c1c5ddca460fab693daf3bd069`).
Hosted [run 39123539414](https://github.com/relux-works/curator/actions/runs/39123539414)
attempt 1 had 19 successful jobs, 2 conditional skips, and one Linux Test
job failing only at external `Complete runner` cleanup: permission denied
removing a retained read-only pnpm fixture store. Its repository test and
platform-case gate steps had passed. No runner data, service or gate was
changed. A real `gh run rerun 39123539414 --failed` request exited **0**.

**Attempt 2 is success on that exact signed candidate: 20 successful jobs,
2 conditional skips.** Linux Test job `117675265348` reports both
`go test exit=0` and `platform-case gate exit=0`, and its runner cleanup
passed. The initial CR validation command remains exit 1 in the ledger;
this hosted retry does not rewrite its immutable validation log.

This successor directly reran all five new Claude test groups, including
all six narrowing/plant controls: outer command exit **0**, six nested
expected-red behavioral commands each exit **1** with their named intended
failure. CLI build and diff check exit **0**. All 57 original local-log
sizes and SHA-256 hashes match the attached full-suite inventory. The full
local suite is **reused at the unchanged source/test/configuration/environment
identity**, not claimed as rerun this turn. The four task blobs match all
three signed snapshots; Story HEAD remains the original checkpoint and
the entire four-file delta is uncommitted.

The 13 of 13 behavioral AC map, six-mutant table, ownership sweep,
no-surface-table brief gap, original reviewer tests and explicit bounds
above remain applicable. No new out-of-contract row or scope expansion
was introduced. `BUG-261018-b2jnyg_recovery2.json` preserves actual hosted
job verdicts, command exits and raw-log hashes. This section is the
task-scoped logbook entry. A subsequent manager-owned CR validation
command has not yet run and is not claimed green here.


## Third successor recovery: CR revision 3

CR revision 3's configured `sh scripts/remote-gate.sh` exited **1** at
signed candidate `954449d0b9e164ee99b634501c7d604145d6c9ef` (signature
`G`, unchanged tree `d1d16a8b497530c1c5ddca460fab693daf3bd069`).
Hosted [run 39130598333](https://github.com/relux-works/curator/actions/runs/39130598333)
attempt 1 failed only at Interop and Linux Test `Complete runner` cleanup:
permission denied removing retained read-only pnpm fixtures. The full job
logs were retrieved successfully; repository steps passed. The native CR
validation log is truncated to 65536 bytes and is used only as failure
evidence, never as passing suite evidence.

`gh run rerun 39130598333 --failed` exited **0**. **Attempt 2 is success:
20 successful jobs, 2 conditional skips**, at the same exact signed candidate.
Linux test and platform-case gate both report exit **0**; both retried jobs'
runner cleanup passed. The original CR command remains exit **1**; no
immutable log or hosted status was rewritten. The local `gh run watch`
observer was deliberately stopped by its recorded PID (exit **143**), and
bounded API reads continued until the real hosted success. No runner data,
service, policy, repository gate, or source code was changed.

This successor directly ran the new Claude regression selection and all six
narrowing controls: exit **0**, with six expected-red nested commands each
exit **1** for the named failure in the mutant table above. CLI build and
diff check exit **0**. All **57** saved full-suite raw-log sizes and hashes
match; source/test/configuration identity, tool versions and conformance
manifest remain unchanged. The full local suite is reused, not rerun this
turn. Four candidate blobs match; Story HEAD remains the checkpoint and
all task changes remain uncommitted. Local main was **4 commits ahead**
when checked; evidence applies to the named candidate.

The **13 of 13** behavioral AC map, six-mutant table, sweep, absent surface
table brief gap, retained reviewer tests and explicit bounds above remain
applicable. No new exclusion or scope expansion was introduced. The
`BUG-261018-b2jnyg_recovery3.json` outcome stores hosted verdicts, command
exits and raw-log hashes. This is the task-scoped logbook entry. Subsequent
manager-owned CR construction/validation is not claimed green in advance.


## Publication retry

The run follows the r2 publication-only brief. No source, test, tracked
 documentation, or configuration file changed. The full current candidate tree
matches the last signed CR revision 4 candidate
`2836a70138afcfb99f7f2e41a6f299725903aa5c` (signature `G`), tree
`d1d16a8b497530c1c5ddca460fab693daf3bd069`. Story HEAD remains the
recorded checkpoint `2e5ab6dab61fc78a525cbcb0de6c5fcc9f5ae882`;
the four task files remain uncommitted for the manager's snapshot.

A fresh `git fetch origin main` exited **0**. Remote main is
`b9caa20901ee859eec20e649940693033e6297d4`; its delta from the checkpoint
has no overlap with any of the four candidate paths. No refresh or
combination resolution was required. Full-tree blob comparison and signature
inspection exited **0**; `git diff --check` exited **0**.

Existing local suite and mutant evidence above is reused, not rerun in this
publication-only retry. The previous CR4 hosted failure remains a failure;
this note does not claim the forthcoming configured validation has passed.
The supplied retry brief records the healed runner and successful cleanup.
The manager will rerun its configured validation after this handoff; no
post-turn validation checklist item was added. The existing **13 of 13**
behavioral coverage map, named tests, six-mutant table, ownership sweep,
brief surface-table gap, and stated out-of-contract bounds remain applicable.

`BUG-261018-b2jnyg_publication-retry.json` records exact candidate and blob
identities. This section is the task-scoped logbook entry for the retry.
