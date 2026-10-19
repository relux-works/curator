# BUG-261018-dbeghc developer evidence

Status: developer validation satisfied; ready for review handoff. Task acceptance and integration remain with the reviewer/orchestrator.

Signed detached candidate: `9e9d63976b5adfa61879183e913de8a1d7d61da8`, signature `G`; tree `df0863ad2abdd830f76d579ca9471498a63b6112`. Story checkpoint/HEAD remains `1724b4a52b4c2d163c0a3e6ab7c2ad2bbe54ac0c`. The real index and Story branch are unchanged; code and tests remain uncommitted for the handoff snapshot.

Production changes: `internal/envprofile/managed.go` and `nofollow.go`; two new Unix test files in that module. The requested mode note is in `docs/environment-config.md`, linked from README. No spec changes or existing tests removed/renamed.

## Acceptance coverage

Seeds retain source rwx bits, including transformed Codex TOML. Wider directory permissions are stale and ordinary locked repair tightens the root/profile/home route without losing contents. Read-only resolve and dry-run do not chmod it. Authored trust state is 0600; the code/docs table covers every managed-write class.

**4 of 5 AC rows driven through production entry points; AC5 is the stated validation bound below.** No attack-surface table was supplied: this is a brief gap, not a claim of catalog-wide coverage.

| AC | Named tests / evidence | Production call site |
| --- | --- | --- |
| 1 | `TestManagedSeedPreservesNativePermissions/{0600,0400,0640,0644}`: exact modes and transformed TOML bytes | `Resolve` -> `gatherSeeds` / `applyPlan` |
| 2 | `TestManagedHomeRepairTightensDirectories` for all five adapters: creation, read-only/dry-run refusal, repair, preserved tool bytes | `Resolve` -> `verifyHomeObserved` / `managedDirectory` |
| 3 | `TestManagedAuthoredFilesArePrivate`: trust/context/marker, repaired trust state; `TestManagedFallbackPermissions`: file/tree modes, owner execution, bytes; table inspection | `Resolve` -> `claudeSeed` / `applyPlan`; real `copyLinkFallback` -> `copyTree` for fallback modes |
| 4 | `TestManagedSeedRefusalsPreserved/{symlink,unreadable,fifo}`: unreadable diagnostic and no published home; narrowing controls | `Resolve` -> `gatherSeeds`, seed writer and directory verifier |
| 5 | Full local inventory and hosted Test/Race/toolchain/gate workflow | Validation commands rather than a product entry point |

The existing `TestEnvResolveCredentialRecordProvisionAndRepair` also drives actual `env resolve codex_cli --repair` through `cli.cmdEnvResolve` -> `envprofile.Resolve` for both provisioning and repair. Its marker assertions prove CLI routing; the new direct-Resolve tests supply the permission assertions.

## Narrowing mutants

`TestManagedModesNarrowingMutants` executes Go overlays against behavioral tests. It retains each writer/check and requires child `go test` **exit 1**, the named failing test, and its permission reason. A compile/setup failure is not a kill. Wrapper exit 0 proves all six expected-red controls were killed.

| Mutant | What it narrows the gate to | Named failing test | Verdict / bound |
| --- | --- | --- | --- |
| private-seed-0644 | Only source mode 0600 widens to 0644 | `TestManagedSeedPreservesNativePermissions/0600` | Killed: mode 0644, expected 0600 |
| claude-trust-0644 | Only Claude trust writer returns to 0644 | `TestManagedAuthoredFilesArePrivate` | Killed: .claude.json mode 0644, expected 0600 |
| single-fallback-0644 | Only single-file fallback returns to 0644 | `TestManagedFallbackPermissions/tree-false` | Killed: copy mode 0644, expected 0600 |
| executable-tree-fallback-0755 | Only a 0755 source escapes the owner mask | `TestManagedFallbackPermissions/tree-true` | Killed: mode 0755, expected 0700 |
| codex-directory-drift-admitted | Only Codex directory drift bypasses verification | `TestManagedHomeRepairTightensDirectories/codex_cli` | Killed: wide home admitted |
| codex-directory-not-tightened | Only Codex home skips protection | `TestManagedHomeRepairTightensDirectories/codex_cli` | Killed: repair reports permission drift |

No survivors. Bounds: these named permission classes, not arbitrary filesystem races or every possible mode. No source-text gate was introduced; the controls run behavior. The tree mutant retains the mask token in its unchanged branch.

## Validation identity

Local Darwin/arm64; Go 1.25.5, Node 25.6.1, pnpm 10.33.0, Rust 1.91.0. Tests use disposable homes. CLI PATH excludes an ambient installed provider; signer-fixture batches remove inherited numbered Git overrides. No user app/home data was reset.

Protocol pin `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`; manifest SHA256 `6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5`. All 1423 conformance blobs checked against committed Git blobs, including 1295 v1 blobs: 0 differences after restoring one export-subst fixture. `spec-root-identity-03.log` exited 0.

| Command / evidence | Real exit | Result |
| --- | ---: | --- |
| Early focused Resolve regressions, regression-after-02.log | 0 | Development-only check; the current test identity is covered by the full 375-test envprofile rerun |
| Focused fallback + six mutants, fallback-mutants-01.log | 0 | All six children exit 1 for named permission failures |
| `go build -o <scratch>/curator ./cmd/curator`, build-03.log | 0 | Candidate compiles |
| `go vet ./...`, vet-03.log | 0 | Candidate static validation |
| Private-cache `golangci-lint run`, lint-03.log | 0 | Candidate lint clean |
| `git diff --check`, diff-check-02.log | 0 | Whitespace clean |
| `bash .github/ci/ledger-consistency.sh <scratch>/ledger`, ledger-02.log | 0 | Existing platform ledger consistency |
| Suite plan, suite-plan-02.log | 0 | 85 served, 0 deferred/excluded packages |
| Full local package/test batches | 0 for every accepted command | 85/85 packages; CLI 420/420, install 363/363, envprofile 375/375 top-level names observed; 11746 named verdicts |
| `python3 <scratch>/collect-local.py`, local-inventory-04.log | 0 | Exact package and heavy-test inventories; original raw source hashes and exits in local-provenance.json |
| Existing platform-case gate over selected raw events, platform-gate-03.log | 0 | No synthetic pass events; 36 skips classified and recorded |
| `gh run watch 39088125072 --exit-status`, hosted-watch-03.log | 0 | Full [hosted workflow](https://github.com/relux-works/curator/actions/runs/39088125072) success on exact signed candidate: 20 successful jobs, two declared conditional skips |

Source/test/configuration/environment-identical package evidence is reused. Audit found early streams spanning protocol-root restoration; all those streams were superseded and replayed under the verified immutable root (80 packages and the first 120 install tests). After the final two test additions the full 375-test envprofile package was rerun. The prior signed candidate differs only in those two envprofile test files. Each selected raw stream must end with a real package verdict; the inventory checker requires all 85 packages and every heavy-package test name. No failed/truncated command is silently called passing.

The earlier local evidence index is superseded for exact environment identity; its archive remains retained for audit. The revised `BUG-261018-dbeghc_local-evidence.md` replaces it and binds the retained version-02 archive by digest and size. The old archive is retained. Board binary attachment refusal (CLI exit 1) was handled through the supported plain-text index; no binary was committed.

## Diagnostic / superseded runs, not green commands

| Run | Real exit | Reason / disposition |
| --- | ---: | --- |
| regression-before-01/02.log | 1 | Baseline permission failures; early fixture setup also needed correction |
| regression-after-01.log | 1 | Duplicate directory assertion after chmod and missing Muse fixture auth; corrected and reran green |
| local-packages-01.json | 1 | Export-subst changed a protocol fixture and artifactpolicy exceeded the 8m budget. All original package-group streams are superseded by root-bound replays; no partial original stream remains accepted |
| local-cli-01.json | 1 | Ambient provider affected three status tests; entire batch reran with isolated PATH, exit 0 |
| local-cli-03.json | 1 | Inherited numbered Git overrides defeated temporary signer config; focused test and entire batch reran without overrides, exit 0 |
| local-envprofile-01.json | 1 | Local 8m command budget exceeded; replaced by complete bounded reruns |
| envprofile-full-01.log | 143 | Superseded pre-freeze diagnostic using a different protocol root, deliberately stopped; unused |
| lint-01.log | 1 | Builtin-shadowing test identifier and stale shared-cache paths; renamed identifier and used private cache, then exit 0 |
| local-inventory-01.log | 1 | Inventory collector detected the previously uncounted artifactpolicy timeout; excluded that whole stream and reran the package, exit 0 |
| local-packages-02.json, local-extra-01.json, local-install-01/02.json | 0 | Command passes superseded for final environment identity because they began before fixture restoration; all affected coverage replayed |
| spec-root-identity-02.log | 1 | Hashes all matched, but the count assertion confused all conformance blobs (1423) with v1-only (1295); corrected scoped assertion passed in -03.log |
| root-identity-control-01.log | 1 (expected red) | Known old green stream rejected specifically as predating immutable fixture root; positive final inventory exited 0 |
| local-root-final-50.json / platform-gate-02.log | 0 / 1 | Replay omitted pinned pnpm PATH prefix: three cases skipped, two mandatory. Platform gate rejected these skips; whole pnpmsource rerun with pnpm 10.33.0 exited 0, final platform gate -03 exited 0 |
| Hosted 39085994376 / 39086022462 | cancelled | Superseded snapshots, no hosted evidence reused |

## Bounds and review scope

- Spec changes are excluded by the brief Scope clause.
- Seeds remain tool-owned after provisioning; repair does not refresh their bytes or modes (existing section 7.4). New native provisioning is the affected operation.
- POSIX assertions run on Unix; Windows retains its existing private DACL backend and hosted coverage. No Windows POSIX-bit equivalence is claimed.
- Native passthrough/link targets and takeover backups retain existing ownership contracts; AC3 documents them rather than recopying credentials.
- Fallback tests drive the real writer called from applyPlan, without forcing filesystem symlink rejection. `TestRC14IdentityMigrationPreservesFallbackCopies` remains present and green for preserved fallback state.
- Deterministic tests do not establish behavior under concurrent native-source replacement/chmod; no universal race-resistance claim.
- No reviewer tests removed/renamed. Initial implementation, not rework. README/docs changes outside the code module are the expressly requested mode note and link.
- At readback local main was 16 commits ahead of the Story checkpoint, with no changes to the two modified production files. Evidence identifies the exact Story candidate; integration belongs to the orchestrator. This producer did not rebase/commit on the Story branch.

## Logbook

The base already used private-directory creation helpers, but left existing wider directories unchanged and bypassed repair for byte-current homes. Seed modes were discarded at the shared 0644 writer. The change carries source modes and treats directory drift as stale, repairing through the ordinary mutation lock. Self-review added fallback permission tests and two further narrowing controls. Final source integrity check exited 0: signature G, remote head equal to CI head, worktree tree identical to the signed candidate, and Story checkpoint unchanged. Real index remained unstaged. `git diff --check` and formatting checks exited 0.

Validation setup anomalies were corrected in fixtures/environment without product workarounds or user-data changes.


## Refresh onto 402fdbc9

README combination inspected: trunk CLIHELP/SNAPPUB text and the existing mode note are retained once; no repository edits needed. All implementation/test files are byte-identical to revision 1. `task-board worktree refresh-candidate BUG-261018-dbeghc` exited 0 and advanced through the CHANGELOG-only release commit to `db65e39555b8dbac7e75f3b2d337daa33831f297`.

Refreshed signed candidate `04f9d8d2f06d6ca2df2e0e1b06f78078445fca35` (signature G), tree `8c29fe0bc3ad426e196062a7efa7c9d6a601658b`. This refresh supersedes the historical validation identity above. Story HEAD/index remain unchanged and task changes uncommitted. Integrity check exited 0; local main is not ahead.

71 standalone local `go test -json -count=1 -p 2 -timeout=8m` commands exited 0: 85/85 packages, CLI 425/425, install 363/363, envprofile 375/375; 12406 named verdicts. Exact commands, exits and hashes are in `BUG-261018-dbeghc_refresh-commands.json`. Collector and platform-case gate exited 0. Collector's planted missing-batch control exited 1 with the intended reason (expected red). Build replay, vet, lint, ledger, suite plan, formatting and diff checks exited 0.

Full hosted workflow [39102151566](https://github.com/relux-works/curator/actions/runs/39102151566) succeeded on the exact signed candidate: 20 successful jobs and two declared conditional skips; `gh run watch --exit-status` exited 0. Linux/macOS streams establish all six permission suites and six narrowing controls per platform; verifier exited 0. No mutant survives. Original AC ratio (4 of 5 production rows, AC5 external validation) and stated scope bounds remain unchanged. No reviewer tests removed/renamed; prior review launch was refused for freshness, not a code verdict.

Initial cli-13 and its first replay both exited 1 before tests on lock acquisition and are retained as diagnostics; later exact replay exited 0. Two initial green batches and the early build were superseded/replayed for checkpoint provenance. Initial unsourced refresh was stopped (143); sourced retry passed. Root-relative pin probe corrected the initial zero-row pathspec failure. All diagnostic exits are reported, never counted as green. Details, narrowing-mutant table, bounds and logbook are in `BUG-261018-dbeghc_refresh-402fdbc9.md`; no new implementation scope.


## Refresh onto 513e0ee8

This refresh supersedes the historical validation identities above. README combination inspected and retained once: OWNPRES, CLIHELP, SNAPPUB and the mode note. No repository edits; five other task files unchanged from revision 2. `task-board worktree refresh-candidate BUG-261018-dbeghc` exit 0: already current at board-only successor `9c1605a7`.

Signed detached candidate `9517bb1a1941d00c224be5d4eff5f9a93e926475`, signature G, tree `088c305b5f9cfcd4c1bc913163cb10f6361e7551`; Story HEAD/index unchanged and task delta uncommitted. 71 standalone `go test -json -count=1 -p 2 -timeout=8m` commands exit 0, accounting for 85/85 packages and CLI425/install371/envprofile375 top-level names (12460 pass/skip verdicts). Collector, platform gate (36 classified skips), build, vet, lint retry, ledger, format/diff and integrity checks exit 0. Missing-batch control exits 1, expected red for `known-missing-batch`. Initial cli-18 lock timeout exits 1 before tests; exact replay exits 0. Initial lint did not start (wrapper 1); installed-binary retry exits 0.

Hosted [workflow 39119749142](https://github.com/relux-works/curator/actions/runs/39119749142) is NOT GREEN on this exact signed candidate: latest snapshot has 12 successful jobs, six failures solely in administrator Linux cleanup hook (real exit 1 each), two active Test jobs and two intended skips. No terminal full-gate exit is available or claimed. One single-job retry exited 1 while the workflow was active. Linux Test artifact and current local macOS stream confirm six permission suites and six narrowing mutant kills; hosted macOS Test remains pending. AC5 is unmet; checklist row 5 is unchecked. Task is blocked on runner-owner cleanup-permission repair or a healthy authorized Linux lane, then a real green full gate. No local substitution or review handoff.

`BUG-261018-dbeghc_gaps.md` contains the short refresh section, exact command/exit table, 4-of-5 production AC map, named narrowing-mutant table, declared bounds and logbook. `BUG-261018-dbeghc_refresh-513e0ee8-commands.json` binds exact command vectors/exits, hashes and retained sanitized archive. All reviewer tests remain under original names. No forced-fit workaround, spec change, new scope or shared runner mutation.
