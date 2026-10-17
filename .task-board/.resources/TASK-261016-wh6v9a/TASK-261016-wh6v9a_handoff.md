# TASK-261016-wh6v9a — Muse policy and v4 producer handoff

## Scope and remaining work

This is the next coherent slice authorized by gw-L151C, not delivery of the
entire L15-1 remainder. Implemented: Muse effective file policy; shared and
isolated admission; captured native-file sharing and guidance; v4 composition,
including the verified root-context fact. Native/unset is reported as mapping
to file without changing the stored selector. Configuration, links, marker and
synthetic credential state are checked before/after refusals. Repair does not
migrate a link or erase a detached credential; explicit migration can unlink a
verified route through the existing transaction contract.

Remaining work is preserved in TASK-261017-0pawkw —
credential-store-policy-consumer-remainder (blocked on this producer slice):
readiness/login consumers, runtime-qualified application and Codex v5 single
payload delivery with strict consumers, the real hosted-override plant, and
adopted environment/node/helper/broker sources. Three unimplemented checklist
rows were explicitly transferred there using board mutations, not checked as
implemented here. No second credential-policy resolver was introduced.

Normative source read directly from curator-spec origin/main:
`1059b94b4135089d842bf017070b8ab47d79a536`, including login-readiness 2/2.1/5,
environments 7.4/8.2/10.1–10.3/12.1–12.2, profile-codex-settings 1/3–6 and
rc7-carriers 3–5. No contradiction requiring a product decision was found.

## Candidate identity

- Story checkpoint/unchanged branch HEAD: `c29663024abfe6dc0353749018fd65b2bf1833e1`.
- Detached signed gate snapshot: `e5b4f70b480ff25dfa84b11a0642c210ab61acc9`.
- Candidate tree: `ad5e8e7f45c56718d5889c9c675d4a33765b3b27`.
- Snapshot signature verified: `G`; Story branch and index remain unchanged.
- Local platform/toolchain: darwin/arm64, Go 1.25.5.
- Full gate result: success; script exit 0; Linux/macOS test and race, Windows tests, drivers, lint, naming, interop and gate self-tests green. Optional self-hosted and candidate-suite jobs were skipped by the committed workflow, not executed.
- Corrected gate run: https://github.com/relux-works/curator/actions/runs/38966014545.
- Exact CI conformance pin: `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`.

## Production coverage

**7 of 8 narrowed AC rows driven through production entrypoints.** The eighth
row is explicitly structural/boundary-only, because this repository emits a
fragment; it does not provide the external launcher's fragment ingestion API.
These eight rows define the accepted Muse producer slice, not eight claims of
full-leaf delivery.

| Row | Requirement | Named tests | Production call site / bound |
|---|---|---|---|
| 1 | Native/unset/file effective file; unchanged selector; mapping reported | TestMuseCredentialPolicyMatrix; TestMuseCLIStorePolicy | cmdEnvConfig → config store; cmdEnvResolve → Resolve → resolveCredentialPolicy → buildFragment |
| 2 | Shared uses captured native XDG auth route | TestMuseCLIFragmentAndFakeLaunch; TestMuseCredentialPolicyMatrix | cmdEnvResolve → Resolve → nativeHome → inspectMuseAuth; fake child only observes projection |
| 3 | Missing shared native file refuses with guidance and no placeholder | TestMuseMissingNativeFilePolicy/shared | Resolve → inspectMuseAuth; unchanged credential-scope snapshot |
| 4 | Isolated does not share/copy/synthesize credentials, even without native auth | TestMuseCredentialPolicyMatrix/file/isolated; TestMuseMissingNativeFilePolicy/isolated | Resolve → resolveCredentialPolicy → home plan / repair |
| 5 | Explicit Muse keychain refuses with no fragment or state change | TestMuseCredentialPolicyMatrix/keychain/shared; TestMuseCLIStorePolicy | cmdEnvResolve → Resolve → admitCredentialStore |
| 6 | Repair preserves recorded link or detached file; explicit migration works | TestMusePolicySwitchPreservesCredentials; TestMusePolicyExplicitMigration | Resolve → repair; PlanMigration / ApplyMigration → existing transaction |
| 7 | V4 five-value projection in JSON/env/shell, root fact iff verified surface, including header-only | TestMuseCLIStorePolicy; TestMuseRootContextVectors; TestMuseRootContextDrift | cmdEnvResolve → Resolve → buildFragment → renderFragment |
| 8 | Closed successor fields, agreement, old-token rejection and one-parent paths | TestMuseFragmentV4PublishedCases; TestMuseFragmentBoundary | Structural oracle + direct production CheckBoundary; malformed wire ingestion by an external launcher is not exercised |

Published corpus: **67 cases (12 valid, 55 invalid)**, copied verbatim and
selecting their exact frozen v1/v2/v3/v4 schemas. The one-parent/suffix relation
also runs through CheckBoundary. Windows uses the existing explicit POSIX-path
projection for schema structure; raw Windows path schema qualification is not
claimed. No prior test function was removed or renamed. Existing
TestMuseDirectoriesAndIsolation now checks stale-home refusal when switching
an already shared home; new isolated provisioning is separately tested.

## Leaf paired-table coverage and explicit bounds

The brief does not contain an attack-surface catalog table: this is a brief
coverage-map gap, not an implicit waiver. Its eight negative/positive rows are
accounted for below. A structural oracle is never called a launched consumer.

| Leaf paired row | Evidence in this slice | Remaining bound |
|---|---|---|
| Old config/system/package/overlay/malformed selectors | Retained TestCredentialStoreConfigGrammar, TestCredentialStoreConfigVersionBoundary, TestCredentialStoreConfigMachineOwnership and CLI config tests | No new source carrier adoption; reused established config contract |
| Codex shared explicit file resurrects stale auth under keyring/auto | TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery remains green | Explicit Codex application refuses before delivery; qualified mismatch/application row remains in successor task |
| Codex native auto/keyring isolated / per-home keyring confusion | Retained TestCredentialProjectionIsolationRefusals and activation tests | Isolated explicit-file positive application remains disabled pending qualification |
| Native config read/parse failure becomes absence | Retained TestCredentialStorePolicyNativeInspection and projection observation tests | No runtime qualification inferred from config parsing |
| Muse file literal omitted or Keychain readiness substituted | Matrix, CLI five-value rendering, isolated/shared controls | Login/readiness and real hosted composition remain deferred; fake executable is not qualification |
| Selector switch plus repair erases vendor file / ownership | TestMusePolicySwitchPreservesCredentials paired with TestMusePolicyExplicitMigration | Concurrent vendor refresh remains unqualified |
| Alias persisted or profile name normalized | Retained CLI round-trip/routing and reviewer dotted-profile regressions | No new source aliases/carriers |
| Foreign source admitted beside forced home store | Not implemented or claimed | Preserved in TASK-261017-0pawkw; native home sources only in this slice |

## Narrowing mutant evidence

Every listed subprocess actually exits **1** and prints the named failing
test. The harness itself exits **0** only after confirming those failures.
No compile/setup error is counted as a kill. **8 killed, 0 survivors.**

| Mutant | Gate narrowed to admit / omit | Named failing test | Survival bound |
|---|---|---|---|
| muse-keychain-admitted | additionally admit Muse keychain as file | TestMuseCredentialPolicyMatrix/keychain/shared | No survivor |
| missing-shared-source-admitted | treat only absent native shared auth as live | TestMuseMissingNativeFilePolicy/shared | No survivor |
| isolated-muse-native-link | let only Muse isolated use shared native-link construction | TestMuseCredentialPolicyMatrix/file/isolated | No survivor |
| renderer-drops-only-backend | omit only TBH_CREDENTIAL_BACKEND during rendering; all policy tokens remain in source | TestMuseCLIStorePolicy | No survivor; producer renderer plant, not hosted-consumer plant |
| header-only-root-fact-dropped | omit root fact only for header-only verified root | TestMuseRootContextVectors/zero-applicable | No survivor |
| boundary-admits-HOME | additionally admit Muse HOME while other variable checks remain | TestMuseFragmentBoundary/HOME | No survivor; direct boundary scope |
| boundary-admits-missing-literal | admit absent backend literal while other policy checks remain | TestMuseFragmentBoundary/missing-literal | No survivor; direct boundary scope |
| boundary-admits-false-root | admit false Muse root fact while non-Muse refusal remains | TestMuseFragmentBoundary/false-root | No survivor; direct boundary scope |

No new source-text acceptance checker is introduced. The renderer control
preserves searched policy literals and executes the real CLI behavioral test,
so static token agreement cannot substitute for its observed output.

## Validation commands and real exits

All gates were standalone processes with redirected logs, no tee or hidden
pipeline status. Logs are task-scoped; public attachments redact private local
paths and signing identity text without changing command results.

| Command | Exit | Evidence |
|---|---:|---|
| go test ./internal/envfragment ./internal/envprofile -run 'Test(Muse\|CredentialStorePolicy)' -count=1 -timeout=6m | 0 | focused-02.log |
| go test ./cmd/curator -run 'TestMuse(CredentialPolicyNarrowingMutants\|CLI\|Fragment\|RootContext)' -count=1 -timeout=9m | 0 | cli-03.log; retained root-context mutants included |
| go test ./cmd/curator -run TestMuseCredentialPolicyNarrowingMutants -count=1 -v -timeout=8m | 0 | mutants-01.log; eight child exits 1 |
| go test ./internal/envprofile -run TestMusePolicyExplicitMigration -count=1 -timeout=3m | 0 | migration-01.log |
| go test ./internal/config ./internal/envregistry ./internal/envprofile ./cmd/curator -run 'TestCredential(Store\|Projection)\|TestReview.*Credential' -count=1 -timeout=6m | 0 | retained-policy-01.log; registry reports no matching tests, not added coverage |
| go test ./internal/envfragment ./internal/envregistry ./internal/config -count=1 -timeout=5m | 0 | packages-01.log |
| TBH_CREDENTIAL_BACKEND=keychain go test ./cmd/curator -run '^TestMuseCLIFragmentAndFakeLaunch$' -count=1 -timeout=3m | 0 | ambient-backend-01.log; synthetic child observes file despite ambient keychain |
| go build ./... | 0 | build-02.log |
| go vet ./internal/envprofile ./internal/envregistry ./internal/envfragment ./cmd/curator | 0 | vet-02.log |
| GOLANGCI_LINT_CACHE=<task-local-cache> golangci-lint run | 0 | lint-04.log; 0 issues |
| git diff --check | 0 | whitespace-02.log |
| sh -n scripts/remote-gate.sh | 0 | shell-01.log |
| CURATOR_CONFORMANCE_ROOT=<exact-CI-pin>/conformance/v1 go test ./internal/envprofile ./cmd/curator -run '^TestMuse' -count=1 -timeout=8m | 0 | pinned-muse-01.log; includes preserved v3 corpus and registry layout |
| sh scripts/remote-gate.sh (first snapshot) | 137 | remote-gate-01.log; stopped obsolete waiter after test failures, CI cancellation requested |
| sh scripts/remote-gate.sh (corrected snapshot; agent Git environment loaded) | 0 | remote-gate-02.log |

Earlier diagnostic failures are not green evidence: focused-01 exited 1 due
the old isolated-refusal expectation; cli-01 exited 1 due a fake-program import
and incorrect corpus-oracle selection; lint-01 exited 1 due stale shared-cache
paths from unrelated worktrees. These were corrected and rerun. cli-02 exited
0 before the final schema-driver cleanup; cli-03 and the full gate supersede it.
The first hosted snapshot (`91ee318924058b3c43eb0e46864a7ab38fd4b374`, run 38958791689) found two retained conformance regressions: the historical registry test expected four env entries, and the v3 corpus lost the relational path check. Ubuntu/macOS test and race jobs exited 1. Corrected tests now assert four historical paths plus exactly one file literal, and project schema-checked v3 paths into the current boundary without claiming v3 wire ingestion. The obsolete local waiter was stopped (exit 137), its remaining Windows job was cancelled, and its result is not green evidence. The exact CI-pin local Muse replay above passed before the new full gate.

No pre-attached test outcome is accepted as this run's execution evidence.

## Path scope and review notes

Product changes stay in envprofile, envregistry and the necessary envfragment
producer seam. CLI files are tests only. Corresponding docs, README and the
worktree LOGBOOK record the new producer behavior and bounds. Frozen schema
and corpus fixtures are under envfragment/testdata. The supporting
scripts/remote-gate.sh edit removes unsigned snapshot creation and artificial
identity: it uses git commit-tree -S and verifies the detached object before
publication. Its actual signed snapshot was exercised; the Story branch,
index and checkpoint were never committed or moved. This supporting edit is
the explicitly declared extra-module delta.

No runtime vendor login, refresh, network inference, real hosted backend merge,
source enrolment, consumer release, v5 emission or source-account qualification
is claimed. Those are deferred requirements, not rejected product inputs or
out-of-contract exclusions. Explicit unqualified mappings refuse under the
leaf's "unsupported mappings remain explicit refusals" clause. No additional
input class is silently removed from the accepted Muse producer scope.

## Attached validation bundle

TASK-261016-wh6v9a_validation.txt contains sanitized command logs, both hosted results, exact candidate identity and named Muse pass/mutant events extracted from all five hosted test/race streams. The official run retains the complete gate artifacts, including platform-case and skip ledgers. The bundle summary contains 214 Muse pass events in each Linux/macOS lane and 213 in the Windows lane (including subtests), with eight named narrowing controls killed in each lane; these event counts are not additional AC coverage claims.
