# TASK-261021-4bo0a7 review r1

Head: `9b657867a0c65f6e86caf85f6ff6fb7fa4235c0a`.
Base: `cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7`.

The orchestrator request and named reviewer were verified on the board. All three preconditions were read, including the producer and reviewer briefs. Origin was fetched; the detached review worktree matched the requested head, and the sole introduced commit verified `G`. The resumed validation clone also matched this head with a clean source tree. No product/source edits were made.

## Surface table

“Held” means that the research's bounded claim survived review. It does not mean that Muse is qualified for strict project exclusion.

| Surface | Held or broken | Attack run | Evidence and limit |
| --- | --- | --- | --- |
| Exact binary and opt-in gate | Held | Independently hashed the authorized copied binary; wrong binary and invalid run-root controls; explicit ungated package invocation with a nonexistent requested fixture root. | SHA-256 `6c04ce1719844685153ef6f04b98113cb61b942d97f0a8b0bfb3932c343984d9`. Wrong digest/root refused before vendor/fixture work. Ungated run skipped all 13 tests and created no root. |
| Isolation | Held | Real loopback positive outside sandbox, denied identical connection inside; harmless-file positive and unchanged synthetic Keychains denial; independent metadata-only probes of existing native directories. | Network mutant reached the server twice and failed the intended check. Keychains mutant failed its intended check. Native-directory metadata access denied; no directory contents or credentials read. |
| P1 root/ancestor instructions | Held at startup | All four instruction names, root/inside-checkout ancestor, trusted/untrusted pairs, and above-Git-root plant. Independent combined-source driver additionally used a normal `.git` directory. | Trusted project source count 1, untrusted 0, selected home count 1. Outside-Git plant absent in both modes with nearby positive controls. |
| P2 nested/lazy instructions | Held as bounded observation | Fresh nested launches with root and nested plants. | Trusted project count 2 versus untrusted 0; selected home retained. Actual lazy tool entry remains unmeasured, as stated. |
| P3 skills | Held at startup | Planted each of three project skill roots beside selected home skill, trusted and untrusted. Independent combined-source probe reproduced the counts. | Project count 1/0, home count 1. Catalog discovery, not invocation or later reload. |
| P4 hooks | Held | Actual UserPromptSubmit marker hooks in selected home and checkout, with and without trust. | Home hook executed in both modes; project hook only with trust. Untrusted catalog had 2 handlers, 1 runnable, 1 inactive. |
| P4 plugins | Held as bounded observation | Raw bundle, refused explicit untrusted project installation, trusted project installation, and selected user installation. | Raw bundle inert; untrusted install refused for workspace trust; successful installs admitted command capability. No claim about all cached/plugin capabilities. |
| P5 MCP | Held for configuration discovery | Empty/nonempty project configuration with selected home configuration; empty home plus nonempty trusted checkout. Independent combined-source probe repeated the latter. | Fresh untrusted project source count 0; trusted project config participated. Empty home MCP was not exclusive. Handshakes/tools remain unmeasured. |
| P6 settings/permissions | Held as bounded observation | Three project candidate files with poison bytes; identical user poison and valid home setting as controls. | Project candidates did not prevent startup; user poison refused; home setting supplied model source. No permission-enforcement proof claimed. |
| P7 resume/children | Held; negative runtime observation reproduced | Headless untrusted session followed by actual resume, trust prompt, and granting trust; native rules listing and hook marker inspection. | Root listing retained saved home-only snapshot, but project skill/hook/MCP revived and project hook executed. Delegation-unavailable diagnostic reproduced; actual child invocation remains unmeasured. |
| P8 late-created files | Held as bounded observation | Created root between two real TUI turns, required two distinct native listings, then fresh trusted launch. Independently started fresh untrusted TUI and selected Quit. | Existing listings retained home-only roots; fresh trusted process loaded the late root. Fresh untrusted TUI offered trust or quit only; Quit exited 0. Lazy reads and late other surfaces remain unmeasured. |
| Narrowing/checker | Held | Restored trust for excluded P1 case; removed network/Keychains denies; wrong binary and root; checker malformed/missing/duplicate/contradictory records. | All five wrapper controls caught their intended failure; each inner command exit 1, wrapper exit 0. P1 failed specifically with project_sources=1 instead of 0. |
| Free hunt: persisted trust | Held | Own production CLI driver combined rules, skills, hooks, and MCP across fresh untrusted, explicit trust, no flag afterward, and explicitly stored trust. | Fresh/no-persisted-trust cases excluded project sources. Explicit flag and stored trust admitted them; home hook remained functional. The flag itself did not persist trust. Verifies the fresh-trust-store requirement. |
| Hygiene/scope/docs | Held | Inspected all 18 added files, complete diff, whitespace, local links, and bounded personal-locator scan with known-violation detector control. | No product-code changes or personal paths/host/user/client/employer identifiers found. Qualification prose preserves observed limits and does not claim strict-adapter admission. |
| Default suite/build/vet | Held with recorded environment corrections | Full default suite in normal `.git` clone at exact head, short TMPDIR/GOTMPDIR; targeted corrected retries; research/product vet and product build. | See execution ledger below. |

## Execution ledger and evidence continuity

Go 1.25.5 on macOS/arm64; tmux 3.6a for private PTYs. All Go tests clear GIT_DIR and GIT_WORK_TREE. Native probes used env -i, disposable HOME/XDG, disabled updates, echo and the inspected outer sandbox, with no real credentials, authentication or remote inference.

Before the provider interruption, the reviewer independently completed:

| Command/run | Exit | Result |
| --- | ---: | --- |
| `go test ./.research/project-exclusion-muse -count=1 -v -timeout=8m`, gated | 0 | 54 passing top-level/subtest records, 72.608 seconds. |
| Explicit research package, ungated, corrected cache | 0 | All 13 tests skipped; no requested fixture root. |
| `sh .research/project-exclusion-muse/mutants.sh`, corrected cache | 0 | Five intended inner failures, exit 1 each. |
| `go vet ./.research/project-exclusion-muse ./...`, corrected cache | 0 | Research and product vet. |
| `go build ./...`, corrected cache | 0 | Product build. |
| Independent combined-source driver, final version | 0 | Fresh, flag-only, removed-flag and stored-trust observations above. |
| Independent native-directory metadata and fresh-TUI probes | 0 | Denial controls; native Quit exit 0. |
| Full `go test -json -count=1 -timeout=90m ./...` | Interrupted before final exit was observed | 85-package inventory: 77 package passes, 5 no-test skips, 2 package failures (CLI and rustsource), install still running. CLI completed all 434 top-level cases; envprofile completed all 393. |
| Complete `internal/rustsource` retry with installed rustc 1.91.0 directly on PATH | 0 | Both missing-rustc failures resolved; package completed. |
| Premature concurrent CLI retry | 1 | GOROOT package-lock acquisition timed out before tests; not passing evidence. |

The initial ungated/control/vet/build/default commands exited 1 due to missing shared Go-cache entries/import failures; corrected private-cache executions are distinguished above. The default-suite retry's rustsource failures were missing `rustc` on PATH; CLI failures came from user-installed providers discovered on inherited PATH, causing status checks to reject them. Product paths involved are unchanged by this research. The first independent driver exited 1 due to its incorrect assumption that a temporary trust flag persists; the corrected driver tested flag-only and stored-trust cases separately.

The user reports that the provider stopped this review and host disk cleanup deleted the old `.val`, `.temp`, and `.t` scratch. Those raw logs and the initial report draft are therefore unavailable. The observations above were already established in this review before interruption and were explicitly preserved by the user's continuation instruction. They were not rerun or represented as newly recovered files. Historical producer run claims are not treated as independent proof; this verdict uses the reviewer's observations.

On resumption, only unresolved validation was run in a recreated normal `.git` clone at the exact head, with a new private GOCACHE/GOMODCACHE, short temporary root, and PATH excluding personal provider directories. The CLI retry selects all 25 failed top-level cases, including the parents of failed subtests. The incomplete install package is rerun as a package because its prior raw completion inventory was deleted. Commands:

- `env -u GIT_DIR -u GIT_WORK_TREE go test -json -count=1 -timeout=35m -run '<exact 25-case anchored alternation>' ./cmd/curator`
- `env -u GIT_DIR -u GIT_WORK_TREE go test -json -count=1 -timeout=35m ./internal/install`

CLI repeat: exit 0, all 25 expected top-level tests passed, 50 passing test/subtest records, 286.705 seconds. Exact selected test names:

```text
TestDraftStatusReportsUpToDateThroughCLI
TestDraftStatusNeedsInstallAfterRefreshWithoutInstall
TestDraftStatusLiveMutationWithoutRefreshStaysCurrent
TestEnvStatusMatrix
TestEnvStatusRegistryBoundaryPostureAndCheck
TestEnvStatusCheckCurrentScopeOnly
TestEnvStatusMissingAndUnreadableKeepRecord
TestEnvStatusUnreadableApprovalStateSurfaced
TestEnvStatusReportsShellHookTrustPosture
TestLegacyPackageStatusBindingAndRepair
TestStatusCheckRefusesV1NULAppearingAfterInstall
TestStatusCheckAdmitsV2NULBearingInstall
TestCuratorStatusProviderPostureAndCheck
TestStatusReportsShellHookTrustPosture
TestStatusUnreadableApprovalStateSurfaced
TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands
TestStatusAcceptsAnUnchangedLegacyMarkerSchema
TestStatusRecordedButMissingStaysInInventory
TestCLIEndToEndInstallStatusAndTamperCheck
TestCLIStatusReportsLegacyPackageMarkerCurrent
TestGCRetainsAndReportsReferencedCompiledState
TestCompiledProjectStatusAndUntrustedRecovery
TestStatusReportsATransitivelyResolvedCompiledCommand
TestCompiledProjectRepairsCorruptCompiledState
TestCompiledProjectRestoresCacheWhenCommitFails
```

Install repeat: exit 0, 371 completed top-level cases (359 passed, 12 skipped); 758 passing and 12 skipped test/subtest records; 443.199 seconds. No failures. Together with completed pre-interruption package runs and the successful rustsource/CLI retries, the 85-package default inventory is accounted for: 80 package passes and 5 no-test package skips. This is combined evidence, not a claim of one uninterrupted all-green full-suite command. Native/conformance skips remain unqualified. Both resumed test commands exited before scratch cleanup; the private caches and disposable clones are not delivery artifacts.

This is a bounded research review, not a strict Muse adapter certificate. Lazy discovery, actual child invocation, permission decisions, successful MCP handshake and broader plugin/lifecycle coverage remain unmeasured. Resume revival is a faithfully documented limitation of the runtime recipe, not an undisclosed defect in this research change.

```json
{"verdict":"accepted","head":"9b657867a0c65f6e86caf85f6ff6fb7fa4235c0a","findings":[]}
```
