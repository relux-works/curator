# BUG-261018-dbeghc — managed-home-seeds-world-readable

## Publication after converge onto 5ecb68da

Base: `5ecb68da8f8ceab4dda118f2263abfb3da63ab8c`. No source, test or documentation file was edited in r7. `task-board worktree status STORY-261018-ogw2f1` exited 0 and reports the carried candidate and stale CR revision 3, with only the active run lease and expected uncommitted changes. It reports no refresh-specific refusal; `refresh-candidate` was not run. Story HEAD and index remain at the orchestrator checkpoint. README and CHANGELOG equal trunk. The five candidate paths are exactly the expected scope: `managed.go`, `nofollow.go`, both permission test files, and the existing docs mode table.

The carried `managed.go` contains the PR #144 Claude seed fix (`UseNumber`, lossless JSON merge, satisfied-state no-op) alongside the mode table and private writers. Both behaviours pass together: permission tests and the newly landed Claude regression tests are green. Native-skills integration also passes. Four other task blobs remain byte-identical to earlier signed candidate `9517bb1a1941d00c224be5d4eff5f9a93e926475`; the `managed.go` difference is incoming trunk work, not a producer edit. Exact current blobs, stream digests and named test verdicts are in `BUG-261018-dbeghc_publication-r7-evidence.json`.

| Command | Real exit | Evidence |
| --- | ---: | --- |
| `go test -count=1 -timeout=5m -json -run '^Test(Managed\|Resolve\|NativeSkills)' -skip 'NarrowingMutants' ./internal/envprofile ./cmd/curator` | 0 | 767 named passes, no skips; managed/native-skills/resolve integration |
| `go test -count=1 -timeout=4m -json -run '^TestClaude(Repair\|Absent\|Seed)' -skip 'NarrowingMutants' ./internal/envprofile` | 0 | 30 named passes, no skips; includes PR #144 numeric fidelity, no-op, golden and unsafe-state refusal tests |
| `go test -count=1 -timeout=4m -json -run '^TestManagedModesNarrowingMutants$' ./internal/envprofile` | 0 | Six controls killed; every child exits 1 with its named permission failure |
| `go build ./...` | 0 | Combined tree compiles |
| `go vet ./...` | 0 | Combined tree static validation |
| `git diff --exit-code HEAD -- README.md CHANGELOG.md`; same comparison against `main` | 0 each | rc.7 wave rule |
| `git diff --cached --exit-code`; `git diff --check`; changed-file `gofmt -l` | 0 each | Index unchanged, no whitespace errors, gofmt output empty |

**4 of 5 AC rows driven.** AC1: `TestManagedSeedPreservesNativePermissions` -> `Resolve/gatherSeeds/applyPlan`; AC2: `TestManagedHomeRepairTightensDirectories` -> `Resolve/verifyHomeObserved/managedDirectory`; AC3: `TestManagedAuthoredFilesArePrivate` and `TestManagedFallbackPermissions` -> `Resolve/claudeSeed/applyPlan/copyLinkFallback/copyTree`; AC4: `TestManagedSeedRefusalsPreserved` -> `Resolve/gatherSeeds`, plus the six controls below. AC5 is assigned to configured post-turn validation by r7; no current full local/hosted green result is claimed and no post-turn-validation checklist item was added. Prior full-suite evidence is historical, not substituted for the combined tree. No current blocker was found in these local checks.

| Mutant | What it narrows the gate to | Named failing test | Verdict / survivor bound |
| --- | --- | --- | --- |
| private-seed-0644 | Only 0600 seeds widen to 0644 | `TestManagedSeedPreservesNativePermissions/0600` | Killed: 0644 instead of 0600 |
| claude-trust-0644 | Only Claude trust output widens | `TestManagedAuthoredFilesArePrivate` | Killed: .claude.json 0644 instead of 0600 |
| single-fallback-0644 | Only single-file fallback widens | `TestManagedFallbackPermissions/tree-false` | Killed: 0644 instead of 0600 |
| executable-tree-fallback-0755 | Only 0755 sources bypass owner mask; searched token retained | `TestManagedFallbackPermissions/tree-true` | Killed: 0755 instead of 0700 |
| codex-directory-drift-admitted | Only Codex directory drift is admitted | `TestManagedHomeRepairTightensDirectories/codex_cli` | Killed: wide home admitted |
| codex-directory-not-tightened | Only Codex directory protection is skipped | `TestManagedHomeRepairTightensDirectories/codex_cli` | Killed: directory permission drift after repair |

No surviving controls. These are behavioral kills, not source-token checks. Current r7 brief carries no surface table (brief gap); no catalog-wide coverage is claimed. Out of contract: spec changes (leaf Scope); refreshing existing tool-owned seeds during repair (environment section 7.4); native passthrough/link targets and takeover backups (existing ownership contracts). Filesystem-race protection is not established. Current local platform is macOS arm64 / Go 1.25.5; Linux and Windows execution remain for configured validation. No reviewer test was removed or renamed. There is no repeat-of review verdict to address. No r7 source rework outside or inside the module.

Logbook: r7 verifies the combination after orchestrator converge. The orchestrator removed hosted-gate checklist rows and the previous handoff succeeded as CR revision 3; all earlier runner/refresh blockers below are historical and resolved. This round leaves the candidate uncommitted and hands off a fresh revision for review and post-turn validation. No runner, shared process, credential, service or control-root file was edited directly.

A diagnostic `git diff` against the old signed candidate exited 1 because it treats the carried untracked tests as index deletions. Direct byte comparison with the old commit then exited 0 for all four unchanged paths, including those tests. An initial board projection used an unknown operation/field and exited 1; documented projections recovered. Neither diagnostic is a product-test failure.

## Current blocker — handoff did not publish a new CR revision

The requested developer handoff exited **0** and changed the board task to `to-review`, but its output contains no new revision artifact. The immediately following `task-board worktree status STORY-261018-ogw2f1` exited **0** and still reports **revision 3 stale**, not a fresh revision on `5ecb68da`. Per r7, this run stops after persisting the exact outputs. A successful status transition is not proof of CR publication. The CLI/orchestrator owner must resolve the missing publication before review can be claimed against a fresh candidate; no bypass, repeated handoff, refresh, commit, or source edit was attempted. All local green evidence above remains valid for the carried blobs, but configured post-turn validation is not represented as passing.

Exact output of `task-board handoff BUG-261018-dbeghc --role developer` (exit 0):

```text
warning: could not compare board content for run RUN-261018-2132e2; no progress recorded for this command
id:BUG-261018-dbeghc role:developer status:to-review checklist:20/20 outcomes:[BUG-261018-dbeghc_validation.md,BUG-261018-dbeghc_local-evidence.md,BUG-261018-dbeghc_hosted-evidence.md,BUG-261018-dbeghc_change-request_rev1.patch,BUG-261018-dbeghc_change-request_rev1-validation.log,BUG-261018-dbeghc_refresh-permission-evidence.json,BUG-261018-dbeghc_refresh-commands.json,BUG-261018-dbeghc_refresh-402fdbc9.md,BUG-261018-dbeghc_change-request_rev2.patch,BUG-261018-dbeghc_change-request_rev2-validation.log,BUG-261018-dbeghc_refresh-513e0ee8-permissions.json,BUG-261018-dbeghc_refresh-513e0ee8-commands.json,BUG-261018-dbeghc_gaps.md,BUG-261018-dbeghc_publication-retry.md,BUG-261018-dbeghc_publication-refresh-refusal.log,BUG-261018-dbeghc_publication-after-converge.json,BUG-261018-dbeghc_publication-after-converge.md,BUG-261018-dbeghc_successor-checks.json,BUG-261018-dbeghc_successor-handoff.md,BUG-261018-dbeghc_publication-r6-evidence.json,BUG-261018-dbeghc_publication-r6.md,BUG-261018-dbeghc_change-request_rev3.patch,BUG-261018-dbeghc_change-request_rev3-validation.log,BUG-261018-dbeghc_publication-r7-evidence.json]
```

Exact output of `task-board worktree status STORY-261018-ogw2f1` (exit 0):

```text
STORY-261018-ogw2f1  active
  path:       .temp/STORY-261018-ogw2f1/worktree (present)
  branch:     task-board/story/STORY-261018-ogw2f1 (present)
  base:       main
  tip:        5ecb68da8f8ceab4dda118f2263abfb3da63ab8c
  tree:       dirty
  lease:      held by RUN-261018-2132e2
  blocked:    story lease is held by run RUN-261018-2132e2
  blocked:    managed worktree has uncommitted or untracked changes
  change-req: BUG-261018-dbeghc rev 3 stale (repository_delta=present, 5 changed path(s))
board debt:
  uncommitted: 0
  unpushed: 0
  unpublished closures: 0
```

## Resolved in an earlier round — historical evidence

The orchestrator repaired/converged the board projection, removed the hosted-gate checklist rows, and the subsequent developer handoff published CR revision 3. The runner cleanup issue was repaired by BUG-261018-bisu6n. Revision 3 later became stale solely because trunk advanced.

The following retained record describes prior rounds only. Its blocked statuses, pending jobs, recovery requests and relative words such as "current" refer to those earlier rounds; they are not current blockers or claims about the new candidate.

<details>
<summary>Earlier round evidence and original command results (historical)</summary>

# BUG-261018-dbeghc — managed-home-seeds-world-readable

## Refresh onto 513e0ee8

README inspection preserves trunk OWNPRES, CLIHELP and SNAPPUB text and the existing mode note once, in their existing sections. Byte comparison proves the README equals trunk plus the four-line mode note. No repository edits were needed. All five other task files are byte-identical to revision 2; no reviewer tests were removed or renamed. Previous reviewer launch was refused for freshness, not a code-review verdict; no repeat-of class was issued.

`task-board worktree refresh-candidate BUG-261018-dbeghc` exited 0: already current at `9c1605a74436d2448f759bd4b66ae6685237d83b`, the board-only successor of `513e0ee8`. Signed detached candidate `9517bb1a1941d00c224be5d4eff5f9a93e926475` has signature `G`, tree `088c305b5f9cfcd4c1bc913163cb10f6361e7551`. Story HEAD/index stay at the checkpoint; task changes remain uncommitted for handoff. Validation: all 71 standalone local test commands exited 0; collector, platform gate, build, vet, lint, ledger, formatting and integrity checks exited 0. The planted missing-batch control exited 1 for its intended reason. Hosted administrator cleanup exits 1 in six jobs; AC5 remains unmet.

## Commands and real exits

| Command | Exit | Evidence |
| --- | ---: | --- |
| 71 standalone `go test -json -count=1 -p 2 -timeout=8m` commands | 0 each | Exact vectors/exits in `BUG-261018-dbeghc_refresh-513e0ee8-commands.json`; 85/85 packages, CLI425/install371/envprofile375 top-level names; 12460 named pass/skip verdicts |
| `python3 <scratch>/collect-refresh.py` | 0 | Full manifest, command, package and heavy-test inventories match |
| `python3 <scratch>/collect-refresh.py --control-missing-batch` | 1 (expected red) | Specifically rejects `known-missing-batch`; never called a passing gate |
| `bash .github/ci/platform-case-gate.sh <scratch>/local-effective.json <scratch>/local-platform` | 0 | 36 classified skips; envprofile has 374 pass and one case-folding filesystem skip |
| `go build -o <scratch>/curator ./cmd/curator` | 0 | Build passes |
| `go vet ./...` | 0 | Static validation passes |
| `GOLANGCI_LINT_CACHE=<scratch>/lint-cache <installed-linter> run` | 0 | Lint retry passes; initial invocation did not start |
| `bash .github/ci/ledger-consistency.sh <scratch>/ledger` | 0 | Ledger matches actual builds |
| `bash .github/ci/suite-plan.sh <verified-pin-root> <scratch>/plan` | 0 | 85 served packages, no deferrals/exclusions |
| `git diff --check`; changed Go files' `gofmt -l` | 0 each | No diff or formatting errors |
| `python3 <scratch>/check-integrity.py` | 0, twice | Signed tree, real HEAD/index, all 1423 conformance blobs (1295 v1) unchanged |
| `python3 <scratch>/verify-local-permissions.py` | 0 | Six permission suites, six named narrowing-control failures and child exit 1 |
| `gh run download 39119749142 --name test-evidence-ubuntu-latest ...`; Linux permission verifier | 0 each | Exact-head Linux Test artifact confirms the same suites and kills |

Local environment: macOS 27.0 arm64, Go 1.25.5, Node 25.6.1, pnpm 10.33.0, Rust 1.91.0. Local Go uses `GOTOOLCHAIN=auto` and its configured GOENV; this is local-suite evidence, not a claim to reproduce CI's `local`/`off` environment. Effective PATH hash and resolved Go environment are recorded. Protocol pin `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`, manifest SHA256 `6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5`.

## Resolved in an earlier round — hosted runner

The orchestrator repaired/converged the board projection, removed the hosted-gate checklist rows, and the subsequent developer handoff published CR revision 3. The runner cleanup issue was repaired by BUG-261018-bisu6n. Revision 3 later became stale solely because trunk advanced.

Full [workflow 39119749142](https://github.com/relux-works/curator/actions/runs/39119749142) targets the exact signed candidate above. Latest persisted snapshot has 12 successful jobs, six failed jobs, two active Test jobs (macOS/Windows), and two intended conditional skips. The workflow is **not green**. `gh run view` exit 0 proves a read, never a passing gate. A terminal `gh run watch --exit-status` result is not yet available. Checklist row 5 remains unchecked; task status is blocked and no review handoff is admitted. The two active jobs are unknown, not counted as passes. No terminal watcher exit is invented.

Each of the six failed jobs has zero repository-step errors; its sole failure is the self-hosted Linux administrator `runner-job-cleanup.sh` hook, exit 1. Each logs 1041 Permission denied errors against the same old PNPM fixture. The affected jobs are Naming, Interop conformance, Linux Gate self-test, Linux Go 1.25.5 and 1.26.0 driver qualification, and Linux Race. Linux Test, Linux Go 1.27.1, macOS Race, and the other completed jobs passed. No repository source, gate, shared runner permission or service was changed.

A single-job retry exited 1 (`job ... cannot be rerun`) while the overall workflow was active. The failed fixture persists across six completed jobs; repairing shared runner state belongs to its owner. The task is blocked on runner-owner repair of stale temp cleanup permissions (recommended), or provision of a healthy equivalent Linux runner in the authorized lane, followed by a legitimate full gate. A later retry after workflow termination has not been run and supplies no current evidence. Disabling the hook, weakening gates or substituting local results would not satisfy AC5. The runner owner must make any shared-runner change.

## Logbook / diagnostics

- Initial `cli-18` exited 1 after 302.003s before any named tests: host-GOROOT lock acquisition timeout. Preserved as unrun diagnostic; later exact replay exited 0. No foreign process/lock was changed.
- Initial lint invocation did not start: saved test PATH lacked the linter (wrapper exit 1, no child exit). Explicit installed-binary retry exited 0.
- `gh run view --log-failed` and `--job --log` exited 1 while the workflow was active. A direct API log read without the escape-output flag also exited 1. Direct per-job API reads with that flag exited 0 and supply the classified failure evidence; no absence was inferred from failed reads.
- New gaps/command outcomes and permission-only evidence are task-scoped. No new implementation, spec, test or README edits in this refresh. All 71 current local test streams were run in this round; no revision-2 stream substitutes for current coverage.

## Production coverage and bounds

**4 of 5 AC rows driven through production; AC5 is external validation below.** The current refresh brief has no surface table, a brief gap rather than catalog-wide coverage. The unchanged implementation coverage is:

| AC | Named tests | Production call site |
| --- | --- | --- |
| 1 | `TestManagedSeedPreservesNativePermissions/{0600,0400,0640,0644}` | `Resolve` -> `gatherSeeds` / `applyPlan` |
| 2 | `TestManagedHomeRepairTightensDirectories` (five adapters; new directories, stale/read-only/dry-run refusal, repair retaining bytes) | `Resolve` -> `verifyHomeObserved` / `managedDirectory` |
| 3 | `TestManagedAuthoredFilesArePrivate`, `TestManagedFallbackPermissions` | `Resolve` -> `claudeSeed` / `applyPlan`; `copyLinkFallback` -> `copyTree` |
| 4 | `TestManagedSeedRefusalsPreserved/{symlink,unreadable,fifo}` and six narrowing controls | `Resolve` -> `gatherSeeds`, seed writer and directory verifier |
| 5 | Full local inventory, full hosted workflow | Validation entry points, not product behavior |

`TestEnvResolveCredentialRecordProvisionAndRepair` additionally exercises actual `env resolve ... --repair` through `cli.cmdEnvResolve` -> `envprofile.Resolve`; mode assertions are supplied by the direct Resolve tests.

Out-of-contract rows remain those justified in the original brief: spec changes excluded by Scope; existing tool-owned seeds do not refresh on repair (section 7.4); native passthrough/link targets and takeover backups retain their ownership contracts. Unix tests establish POSIX modes on Linux and macOS; the Windows job retains existing DACL behavior and makes no POSIX-bit equivalence claim. Deterministic tests do not establish protection against concurrent source replacement/chmod. Rework delta is zero; existing README/docs changes outside internal/envprofile are the expressly required docs note.

## Narrowing controls

The current local macOS stream and current hosted Linux stream exercise all six controls. Hosted macOS Test remains pending; its revision-2 stream is not counted as current evidence. Each wrapper requires the child command's real exit 1, the named failing behavioral test and the intended permission reason. Compile/setup failures do not qualify. No survivors.

| Mutant | What it narrows the gate to | Named failing test | Verdict / bound |
| --- | --- | --- | --- |
| private-seed-0644 | Only source 0600 widens to 0644 | `TestManagedSeedPreservesNativePermissions/0600` | Killed: 0644, expected 0600 |
| claude-trust-0644 | Only Claude trust writer widens to 0644 | `TestManagedAuthoredFilesArePrivate` | Killed: .claude.json 0644, expected 0600 |
| single-fallback-0644 | Only single-file fallback widens to 0644 | `TestManagedFallbackPermissions/tree-false` | Killed: copy 0644, expected 0600 |
| executable-tree-fallback-0755 | Only a 0755 source bypasses the owner mask | `TestManagedFallbackPermissions/tree-true` | Killed: 0755, expected 0700 |
| codex-directory-drift-admitted | Only Codex directory drift is admitted | `TestManagedHomeRepairTightensDirectories/codex_cli` | Killed: wide home admitted |
| codex-directory-not-tightened | Only Codex protection is skipped | `TestManagedHomeRepairTightensDirectories/codex_cli` | Killed: repair reports permission drift |

The tree mutant preserves the mask token in its other branch. Behavioral tests, not a source-text checker, establish these kills. Bounds are these named permission classes, not every possible mode or filesystem race.


## Retained state

All local validation processes have exited. The external workflow remains active on its exact signed candidate; its temporary gate branch is retained for diagnosis and any later authorized retry. No promise is made to attach later results after this headless run ends. The existing task changes, real Story HEAD and real index are preserved. Review acceptance and integration are not claimed.
# BUG-261018-dbeghc — Publication retry

## Candidate identity

The six task files still match the last signed candidate `9517bb1a1941d00c224be5d4eff5f9a93e926475`, tree `088c305b5f9cfcd4c1bc913163cb10f6361e7551`, signature `G`. The four tracked task files compare byte-identically with that candidate (`git diff --exit-code <candidate> -- README.md docs/environment-config.md internal/envprofile/managed.go internal/envprofile/nofollow.go`, exit 0). Both untracked tests match its blobs: `managed_modes_mutants_unix_test.go` = `f7dc7896879f3ab282b17098650a351fbcf3b15d`, `managed_modes_unix_test.go` = `b6ffb58f18fd8c905111412a134c3b7fed3bbada`.

No source, test, README, CHANGELOG or other repository documentation file was edited. Story HEAD remains `9c1605a74436d2448f759bd4b66ae6685237d83b`; task changes remain uncommitted. `git diff --cached --exit-code` and `git diff --check` each exit 0. Incoming README changes are disjoint from the existing permission note: audit cache ruleset 4 and network-contact test documentation. No combination was installed because refresh refused before installation.

## Resolved in an earlier round — refresh refusal

The orchestrator repaired/converged the board projection, removed the hosted-gate checklist rows, and the subsequent developer handoff published CR revision 3. The runner cleanup issue was repaired by BUG-261018-bisu6n. Revision 3 later became stale solely because trunk advanced.

`task-board worktree status STORY-261018-ogw2f1` exits 0 and reports revision 2 stale. The assignment explicitly requires refresh because protected trunk advanced on README. Two direct invocations of `task-board worktree refresh-candidate BUG-261018-dbeghc`, with the mandated agent Git environment, each exit **1**:

`change_request_candidate_drift: candidate refresh would leave tracked board paths outside its file transaction stale; synchronize the board with the prepared checkpoint before retrying`

Prepared checkpoint tree: `09b478fc86f54e1596c3fb07efb9857c50b6c186`. Workspace candidate tree: `088c305b5f9cfcd4c1bc913163cb10f6361e7551`. Refusal names eight tracked board paths belonging to the incoming reserved-TLD task: three activity streams and five element/progress files. The unchanged retry confirms this is persistent, not a successful refresh. Exact refusal log is attached separately.

## Resolved in an earlier round — recovery

The orchestrator repaired/converged the board projection, removed the hosted-gate checklist rows, and the subsequent developer handoff published CR revision 3. The runner cleanup issue was repaired by BUG-261018-bisu6n. Revision 3 later became stale solely because trunk advanced.

The failed assumption is that the producer refresh can install incoming tracked board changes. CLI help and project-management documentation expose no worker board synchronization route; `worktree converge` explicitly requires an orchestrator session. Direct board-file editing and clearing the tracked-run identity to impersonate an operator are prohibited. No such bypass was attempted.

Recommended external input: orchestrator/CLI owner restores the board/checkpoint projection through supported recovery, preserving this exact candidate, then resumes this producer for refresh and handoff. A CLI repair at its source is an alternative if supported operator recovery also refuses; that is outside this run's write boundary. Neither weakening the drift guard nor manually copying activity history is acceptable.

## Validation and bounds

No behavior changed in this retry. Previous full local evidence remains historical evidence for the old candidate; no validation of the combined new trunk tree is claimed. Full local suite, build, lint, permission mutants and hosted gate were not rerun in this publication-only retry: the required candidate refresh failed, and r4 delegates configured validation to handoff. No new signed combined candidate exists, and AC5 remains unchecked. No claim is made that runner repair alone establishes a green full gate. Prior AC coverage remains 4 of 5 rows driven; AC5 requires external validation. Prior test names, mutant evidence and contract bounds remain in the gaps file.

## Logbook

Publication retry found a new blocker before the formerly failing hosted cleanup could be exercised. Candidate, index and Story checkpoint preserved; only task-scoped scratch and board resources/notes changed. No new validation checklist item was added. Status is blocked pending supported board/checkpoint synchronization; review and integration are not claimed.


## Publication after converge

Current r5 publication handoff is based on `892dab098aef1497df9ccacdb173faaf4f8b8418`; uncommitted candidate tree `efc4e2371c2f7e21ec8b56d9ffc0273b5c98c722`. The orchestrator's converge resolves the previous board/checkpoint refresh blocker. `task-board worktree status STORY-261018-ogw2f1` exited 0: current lease and expected dirty candidate, no refresh-specific refusal. No `refresh-candidate` was run.

The implementation, both permission test files and `docs/environment-config.md` are byte-identical to signed candidate `9517bb1a1941d00c224be5d4eff5f9a93e926475`. The only producer delta removes the old four-line permission note from root README to honor the explicit rc.7 wave rule; README and CHANGELOG now equal trunk exactly. The full mode table remains in docs. No other source, test or documentation edit. HEAD and the real index remain at the checkpoint; no producer commit was created. The candidate above is a tree identity, not a claim of a new signed commit; the configured gate owns that snapshot.

| Command | Real exit | Result |
| --- | ---: | --- |
| `go test -count=1 -timeout=5m -json -run '^Test(Managed\|Resolve)' ./internal/envprofile` | 0 | 33 top-level tests / 63 named passes, no skips; all six narrowing controls killed |
| `CURATOR_CONFORMANCE_ROOT=<pinned-root> go test -count=1 -timeout=8m -json -run '^Test(CodexSeed\|EnvironmentsCodexSeedVectors$\|EnvironmentWriteNofollowVectors$\|DotfileManagerVectors$\|UnreadableDotfileStateKeepsTakeoverQuiet$)' ./internal/envprofile` | 0 | 8 top-level tests / 40 named passes, no skips |
| `go test -count=1 -timeout=8m -json -run '^Test(Env\|ResolveTakeover\|UmbrellaSymlinkIntoManaged)' ./cmd/curator` | 1 | 59 top-level passes; six status failures from ambient untrusted `curator-engines` providers, not called green |
| Isolated previous-validation PATH: `go test -count=1 -timeout=3m -json -run '^TestEnvStatus(Matrix\|RegistryBoundaryPostureAndCheck)$' ./cmd/curator` | 0 | First two failing tests pass; removes ambient provider discovery, no product/test edits |
| Same isolated PATH: `go test -count=1 -timeout=4m -json -run '^TestEnvStatus(CheckCurrentScopeOnly\|MissingAndUnreadableKeepRecord\|UnreadableApprovalStateSurfaced\|ReportsShellHookTrustPosture)$' ./cmd/curator` | 0 | Remaining four failing tests pass |
| `go build ./...` | 0 | Build passes |
| `go vet ./internal/envprofile ./cmd/curator` | 0 | Vet passes |
| `git diff --exit-code HEAD -- README.md CHANGELOG.md`; `git diff --check`; `git diff --cached --exit-code` | 0 each | Wave docs unchanged, no whitespace errors, real index unchanged |
| `gofmt -l` on four changed Go files; explicit empty-output assertion | 0 each | Formatting clean |
| Initial `go test -count=1 -timeout=8m -json ./internal/envprofile` | 1 | Deliberately stopped owned test PID before timeout: prior package inventory measured 1329.734s. Partial stream is not passing evidence |

Current evidence is in `BUG-261018-dbeghc_publication-after-converge.json`, including exact base/blob/tree identities, test verdicts, stream digests, mutant child outputs and exits. Local macOS arm64 / Go 1.25.5. Conformance tests use the unchanged `43bf0a2506d5c354a73bbc3ea4623d4653db10c7` fixture root. Prior full-suite results remain historical; none is represented as a current full run. Per r5, full configured validation runs after handoff; AC5/checklist row 5 remains unchecked. No post-turn-validation checklist item was added. A runner infrastructure failure remains a blocker, never a pass.

Current product coverage remains **4 of 5 AC rows driven**: AC1 `TestManagedSeedPreservesNativePermissions` -> `Resolve/gatherSeeds/applyPlan`; AC2 `TestManagedHomeRepairTightensDirectories` -> `Resolve/verifyHomeObserved/managedDirectory`; AC3 `TestManagedAuthoredFilesArePrivate` and `TestManagedFallbackPermissions` -> `Resolve/claudeSeed/applyPlan/copyLinkFallback/copyTree`; AC4 `TestManagedSeedRefusalsPreserved` -> `Resolve/gatherSeeds`, plus the controls below. `TestEnvResolveCredentialRecordProvisionAndRepair` additionally drives `cli.cmdEnvResolve` -> `envprofile.Resolve`. AC5 is deferred configured validation, not product behavior. Current r5 has no surface table (brief gap); prior product bounds and out-of-contract declarations above remain applicable. No reviewer test was removed/renamed; prior reviewer launches were freshness refusals, not a repeat-of verdict. Full platform matrix is unverified on this tree until the configured gate.

| Mutant | Gate narrowed to | Named failing test | Result / survivor bound |
| --- | --- | --- | --- |
| private-seed-0644 | Only 0600 seeds widen to 0644 | `TestManagedSeedPreservesNativePermissions/0600` | Killed; mode 0644, expected 0600 |
| claude-trust-0644 | Only Claude trust output widens | `TestManagedAuthoredFilesArePrivate` | Killed; .claude.json 0644, expected 0600 |
| single-fallback-0644 | Only single-file fallback widens | `TestManagedFallbackPermissions/tree-false` | Killed; 0644, expected 0600 |
| executable-tree-fallback-0755 | Only 0755 sources evade owner mask, mask token preserved | `TestManagedFallbackPermissions/tree-true` | Killed; 0755, expected 0700 |
| codex-directory-drift-admitted | Only Codex drift is admitted | `TestManagedHomeRepairTightensDirectories/codex_cli` | Killed; wide home admitted |
| codex-directory-not-tightened | Only Codex protection is skipped | `TestManagedHomeRepairTightensDirectories/codex_cli` | Killed; repair still reports permission drift |

All six behavioral wrappers pass and require child `go test` exit 1 with the named failure and intended permission reason. No survivors. These bounds cover the named permission classes, not filesystem races. Out-of-contract scope remains spec changes (leaf Scope), refreshed tool-owned seed bytes/modes (section 7.4), native passthrough/link targets and takeover backups (ownership contracts). Windows has no new POSIX-bit preservation claim.

Changelog text for rc.7: preserve native seed modes, make managed routes owner-only, and tighten directory modes through repair without discarding contents.

Logbook: converge recovered the prior refresh refusal; root README now obeys the wave rule. Ambient user providers explain six diagnostic status failures, all resolved by rerunning under the previously documented isolated PATH. No shared process, runner, service, login or credential was changed. Evidence precedes role handoff; review acceptance and integration are not claimed.


## Publication after converge — successor handoff recovery

The predecessor attached its r5 evidence but issued no developer handoff. This successor changes no source, test or documentation files. Base remains `892dab098aef1497df9ccacdb173faaf4f8b8418`; candidate tree remains `efc4e2371c2f7e21ec8b56d9ffc0273b5c98c722`. All five task-file SHA256 values match the predecessor evidence and those files remain byte-identical to signed candidate `9517bb1a1941d00c224be5d4eff5f9a93e926475`. The earlier README removal is already carried; README and CHANGELOG equal trunk. No refresh-specific refusal is reported, so no refresh was run. No commit or branch mutation was made.

This successor directly reran the six permission suites (including all six narrowing controls) with `go test -count=1 -timeout=4m -json -run '^TestManaged(SeedPreservesNativePermissions|HomeRepairTightensDirectories|AuthoredFilesArePrivate|SeedRefusalsPreserved|FallbackPermissions|ModesNarrowingMutants)$' ./internal/envprofile`, exit 0; `go build ./...`, exit 0; README/CHANGELOG equality, index equality, diff whitespace and changed-file gofmt checks, exit 0 each. Exact identities, command and test verdicts are attached as `BUG-261018-dbeghc_successor-checks.json`. Earlier conformance and CLI verification remains the predecessor's evidence, not a claim that this successor reran it; interrupted and failing diagnostics remain explicitly red above.

Product coverage remains **4 of 5 AC rows driven** with the production call sites, out-of-contract declarations and six-mutant table above. No reviewer tests removed or renamed. No new platform claim. Full local suite and full hosted gate on this combined tree are not claimed: r5 assigns configured validation after handoff. Existing checklist row 5 is left unchecked; no new post-turn-validation checklist item is added. This successor will invoke the requested developer handoff after attaching evidence, then end its turn. Any handoff refusal is an actual refusal, not review acceptance.

Logbook: successor confirms the carried tree is unchanged and closes the predecessor's missing-handoff step. No shared runner, service, process, credential or control-root file was changed directly.

## Resolved in an earlier round — r6 publication retry

The orchestrator repaired/converged the board projection, removed the hosted-gate checklist rows, and the subsequent developer handoff published CR revision 3. The runner cleanup issue was repaired by BUG-261018-bisu6n. Revision 3 later became stale solely because trunk advanced.

The candidate remains based on `892dab098aef1497df9ccacdb173faaf4f8b8418`, candidate tree `efc4e2371c2f7e21ec8b56d9ffc0273b5c98c722`. All five implementation/test/docs blobs match the previous publication evidence. Root README and CHANGELOG match current trunk `59746a4545db06607446dcfe5fc35f28118a9125` exactly. No source, test, documentation, index or producer commit was changed in this run.

The deferred-base instruction requires refresh because incoming trunk also changes `internal/envprofile/managed.go`. Inspection finds native-skills repair additions, separate from the mode changes; the tool did not install the combination. `task-board worktree refresh-candidate BUG-261018-dbeghc` exited **1** with the exact refusal:

```
change_request_candidate_drift: candidate refresh would leave tracked board paths outside its file transaction stale; synchronize the board with the prepared checkpoint before retrying (paths=.task-board/.activity/BUG-261018-uoisfe/events.ndjson, .task-board/.activity/EPIC-261016-stqc77/events.ndjson, .task-board/.activity/STORY-261018-k5yd2n/events.ndjson, .task-board/EPIC-261016-stqc77_curator-rc7-code-wave/STORY-261018-k5yd2n_native-skills-in-place/BUG-261018-uoisfe_in-place-profile-skills-miss-native-surface/README.md, .task-board/EPIC-261016-stqc77_curator-rc7-code-wave/STORY-261018-k5yd2n_native-skills-in-place/BUG-261018-uoisfe_in-place-profile-skills-miss-native-surface/progress.md, .task-board/EPIC-261016-stqc77_curator-rc7-code-wave/STORY-261018-k5yd2n_native-skills-in-place/README.md, .task-board/EPIC-261016-stqc77_curator-rc7-code-wave/STORY-261018-k5yd2n_native-skills-in-place/progress.md, .task-board/EPIC-261016-stqc77_curator-rc7-code-wave/progress.md, prepared_tree_oid=4b0d82cc6a7997ac15f1ce0d41ecedd3b52729d8, workspace_tree_oid=efc4e2371c2f7e21ec8b56d9ffc0273b5c98c722)
```

Prepared checkpoint tree: `4b0d82cc6a7997ac15f1ce0d41ecedd3b52729d8`; retained workspace tree: `efc4e2371c2f7e21ec8b56d9ffc0273b5c98c722`. Direct edits to board files and orchestrator impersonation are forbidden. Recovery requires the orchestrator/CLI owner to synchronize the tracked board/checkpoint projection through its supported flow, preserving this candidate, then resume refresh/handoff. Repeating the earlier identical drift refusal or weakening its guard is not a solution. No product decision or source workaround is needed.

Current product coverage: **4 of 5 AC rows driven**. AC1: `TestManagedSeedPreservesNativePermissions` -> `Resolve/gatherSeeds/applyPlan`; AC2: `TestManagedHomeRepairTightensDirectories` -> `Resolve/verifyHomeObserved/managedDirectory`; AC3: `TestManagedAuthoredFilesArePrivate` and `TestManagedFallbackPermissions` -> `Resolve/claudeSeed/applyPlan/copyLinkFallback/copyTree`; AC4: `TestManagedSeedRefusalsPreserved` -> `Resolve/gatherSeeds`, plus the narrowing controls. AC5 requires configured full validation after successful handoff; it is not claimed. The current retry brief has no surface table, a recorded brief gap; no catalog-wide coverage is claimed.

| Mutant | Gate narrowed to | Named failing test | Verdict / survivor bound |
| --- | --- | --- | --- |
| private-seed-0644 | Only 0600 seed copies widen | `TestManagedSeedPreservesNativePermissions/0600` | Killed; 0644 instead of 0600 |
| claude-trust-0644 | Only Claude trust output widens | `TestManagedAuthoredFilesArePrivate` | Killed; .claude.json 0644 instead of 0600 |
| single-fallback-0644 | Only single-file fallback widens | `TestManagedFallbackPermissions/tree-false` | Killed; 0644 instead of 0600 |
| executable-tree-fallback-0755 | Only 0755 source evades mask; mask token retained | `TestManagedFallbackPermissions/tree-true` | Killed; 0755 instead of 0700 |
| codex-directory-drift-admitted | Only Codex drift admitted | `TestManagedHomeRepairTightensDirectories/codex_cli` | Killed; wide home admitted |
| codex-directory-not-tightened | Only Codex directory protection skipped | `TestManagedHomeRepairTightensDirectories/codex_cli` | Killed; permission drift after repair |

Each control requires real child exit 1 and the named permission failure. No survivors. Bounds remain the named classes, excluding concurrent filesystem replacement/chmod. Out of contract: spec changes (leaf Scope); refreshing tool-owned seeds during repair (section 7.4); native passthrough/link targets and takeover backups (ownership contracts). Windows makes no new POSIX-mode claim. No reviewer test was removed or renamed; no repeat-of review verdict was issued. No rework source diff was added, inside or outside the named module. The existing docs note is explicitly in the leaf scope.

Logbook: the previously documented board/checkpoint drift refusal recurred for incoming native-skills board records. The candidate remains preserved; no shared runner, service, authentication or foreign process was changed. Full local/hosted validation was not rerun and is not represented as green on a combined tree. No post-turn-validation checklist item was added.

### Current commands and real exits

| Command | Exit | Evidence / limit |
| --- | ---: | --- |
| `go test -count=1 -timeout=5m -json -run '^Test(Managed\|Resolve)' ./internal/envprofile` | 0 | 63 named passes, including all six narrowing controls; child commands each fail with exit 1 for their intended permission reason |
| `go build ./...` | 0 | Build passes on retained base |
| `go vet ./internal/envprofile ./cmd/curator` | 0 | Static validation passes |
| `gofmt -l` on four changed Go files; empty-output assertion | 0 each | No formatting changes needed |
| `git diff --exit-code main -- README.md CHANGELOG.md` | 0 | Wave docs equal trunk |
| `git diff --cached --exit-code`; `git diff --check` | 0 each | Real index unchanged; no whitespace errors |
| `task-board worktree refresh-candidate BUG-261018-dbeghc` | 1 | Exact `change_request_candidate_drift` refusal above |

The new task-scoped JSON outcome records candidate hashes, all named passes, all six child failure outputs, command exits and log digests. Conformance and broad CLI suites were not rerun in this publication retry; their previous evidence is historical, not a current combined-tree result. Local environment remains macOS arm64 / Go 1.25.5. No new signed candidate or hosted gate result is claimed before the configured handoff snapshot.


</details>
