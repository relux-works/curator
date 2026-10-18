# TASK-261017-sxydvh — developer handoff

Ready for review after refresh onto `402fdbc9`. Story HEAD remains the checkpoint `402fdbc9b77803cb2d6b11f4f1cfd83961e6897b`; all repository changes remain uncommitted. The detached signed candidate is `5ceef5a6656555f87460ec4c05875d18c6e0b5fb` (signature G), tree `7ced0d2be8a93bfd6063b22e70f83c8e37a7e29a`; worktree equality is verified. Integration belongs to the Story orchestrator. The refresh section below is the current validation authority; revision-1 validation is retained as historical evidence.

## Contract and implementation

Normative source: curator-spec `c1a2bb256f10e583199f848bfb3a7a4d30e0d9bc`, environments §§9.1, 9.4. The exact 40-row corpus SHA-256 is `9b14ffe2f451de753154fbf55e9200be7e0f24978e232625152eed1974c8b46c`. This change does not edit the spec.

`internal/audit/audit.go` classifies URL authority independently of its path, applies IDNA ASCII conversion, ASCII lowercase and removal of one trailing dot. Loopback and multi-label hosts ending in test, example, invalid or localhost are non-network. Bare labels, suffix lookalikes, example.com/net/org and the exfiltration canary remain contacts. Invalid IDNA remains conservatively a contact. Reserved hosts skip only their own finding, preserving real hosts and other capability findings in the same file. Ruleset 4 invalidates ruleset-3 cached decisions.

Production entry points exercised: `audit.Gate`, `audit.GateReadOnly`, `envprofile.Install`, `envprofile.EnsureDefaultWithPolicy`. Member vectors exercise fresh and cached gates, exact host/file findings and CanaryPasses. Installation and migration vectors exercise the public pipeline and check refusal leaves no source/lock publication; migration preserves the original skill pin. Cache tests cover both hash versions and both stale-refusal and stale-allow directions, prove a new miss then hit, and preserve old cache bytes.

## Acceptance coverage

**4 of 5 AC rows driven through production entry points**; row 5 is validation-only and is verified by the local and hosted records below. All 5 rows have evidence.

| AC | Named tests / production call sites | Evidence |
| --- | --- | --- |
| 1 | TestNetworkContactVectors; TestProfileNetworkContactVectors — Gate/GateReadOnly, Install, EnsureDefaultWithPolicy | All reserved/loopback rows, cache variants and both pipeline paths pass |
| 2 | Same vector suites; TestNetworkContactNormalizationBoundaries — Gate | Lookalikes, bare labels, documentation domains, canary, both mixed orders and cross-file host/file pass |
| 3 | Same vector suites — all four public calls | 40 of 40 vectors; 120 named vector/path terminals pass |
| 4 | TestNetworkContactCacheInvalidatesRulesetThree; TestNetworkContactNarrowingMutants — Gate/GateReadOnly | Both stale-decision directions and six narrowing controls pass |
| 5 | Full local inventory and full hosted workflow | Exact candidate, green records below |

Brief gap: no attack-surface table was supplied. The conditional surface-table coverage-map precondition cannot be instantiated; the complete requested 40-vector mapping is recorded below. No supplied vector or AC row is treated as out of contract. Bound: the existing static detector examines HTTP(S) contacts on its supported script/manifest surfaces; these tests do not claim general network discovery. Pipeline fixture files use `scripts/`. A missing-capabilities projection uses supported legacy schema 2, since schema 3 requires that field; absence remains absence.

## Narrowing mutants

Each mutant keeps the gate present and admits one forbidden class. Each child `go test` exits **1**, with the named assertion required by the harness; each harness subtest exits successfully. No survivors. The full-URL mutant preserves the host-classification tokens and is attacked behaviorally.

| Mutant | Narrowed gate admits | Named failing test | Result |
| --- | --- | --- | --- |
| full-url-percent-path | Real host hidden by malformed percent escapes in a URL path | TestNetworkContactNormalizationBoundaries/percent-path-real | child 1; URL classification assertion |
| contains-test | x.test.evil.com | TestNetworkContactVectors/lookalike-nested-test | child 1; Gate pass assertion |
| exempt-example-com | example.com | TestNetworkContactVectors/documentation-com | child 1; Gate pass assertion |
| exempt-bare-test | bare test | TestNetworkContactVectors/bare-test | child 1; Gate pass assertion |
| reserved-clears-file | Real host after a reserved host in one file | TestNetworkContactVectors/mixed-reserved-first | child 1; Gate pass assertion |
| reuse-ruleset-three | Stale cached network decisions | TestNetworkContactCacheInvalidatesRulesetThree/v1/api.test (and neighbors) | child 1; stale ruleset reused assertion |

## Revision-1 verification (historical)

Revision-1 local run: **111 bounded standalone Go commands**, **85 packages**, **3475 independently inventoried Test names covered**, **3476 unique top-level terminals**, **3445 pass / 31 skip**, **51582 Go JSON events**. Every accepted command exited 0. Full package commands include the fuzz seed path. All 451 pre-existing audit/envprofile Test names remain present and are included in the run. Per-command exact arguments, exit status, duration, candidate identity, environment and complete log digest are in the attached plaintext bundles. Selection masks are checked against observed terminals and an independent complete inventory; the assembler also checks every package and rejects any fail event.

`go build ./...`, `go vet ./...`, `golangci-lint run`, gofmt with empty output, ledger consistency, platform-case gate, gate self-test and no-broad-suppression checks exit 0. Gate self-test reports 267 pass and 2 explicit Rustup-unavailable skips. The evidence assembler's isolated positive control exits 0; missing-result and self-consistent truncated-stream controls each exit 1 for the intended missing evidence, never accepted as passing suites.

Hosted workflow: [https://github.com/relux-works/curator/actions/runs/39077229686](https://github.com/relux-works/curator/actions/runs/39077229686), head `528618358696a3d498ac456667ad19b87f6fc8bd`, status `completed`, conclusion `success`. Every required job is green; conditional skipped jobs are listed in the attached native provider JSON. The remote-gate controller's real exit is 0.

Local toolchain: Go 1.25.5 darwin/arm64, Git 2.54, golangci-lint 2.12.2, Node 25.6.1, task-local pnpm 10.33.0, Rust 1.91, Swift 6.4. Task-local pnpm setup: `npm install --prefix .temp/TASK-261017-sxydvh/pnpm --no-save --ignore-scripts pnpm@10.33.0` (exit 0), required by existing real package-manager tests. Protocol root is raw-byte exact rc14 `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`. Child environment removes inherited Git override variables and excludes the operator's unrelated .local/bin provider; it does not alter host Git configuration, installed binaries or keys. See the hosted workflow for its platform/toolchain matrix. Local commands are serialized to avoid resource contention.

## Logbook and diagnostics

A production-path probe exposed a real-host false negative when full-URL parsing rejected a percent-escaped path; its actual expected-red exit 1 is attached separately. Authority-only extraction fixes it, paired real/reserved regression cases pass, and the new narrowing mutant kills the previous behavior. The entire local and hosted gates were rerun on the corrected candidate.

Three earlier hosted snapshots/runs were cancelled or superseded; their controllers were explicitly stopped (137). They are not full-green evidence. An early local grouped crossconformance command exited 1 at its aggregate timeout; all 54 names were rerun in bounded subsets. A second grouped wrapper timed out (wrapper 1, underlying Go exit unknown, truncated stream classified unrun); the partial log includes an artifactpolicy timeout and a buildrepo shared-clock assertion. All affected packages were rerun separately and passed with unchanged assertions and no -short escape. Original failures remain attached as diagnostics, excluded from accepted full-suite counts. Host load was recorded without changing other sessions' processes.

Initial setup diagnostics (raw git-archive export substitution, inherited Git signing overrides, an unrelated installed provider, and PATH lint lookup) are retained separately; exact raw-blob restoration and isolated child configuration resolved them before accepted runs. No secrets or host configuration were modified. A binary resource upload was refused by the board; evidence is attached as reconstructible sanitized UTF-8 JSONL bundles with per-file SHA-256, rather than a binary attachment. Private paths, emails, key material and URL userinfo are redacted; accepted event streams are retained in full.

## Scope and delivery

Product change is confined to `internal/audit`. Outside-module changes: `internal/envprofile/network_contacts_test.go` supplies required real install/migration coverage; Go module files add the IDNA dependency and required transitive versions; README documents the pinned corpus and reproducible test command. Existing cache-mutant anchor and reserved .example witnesses were adjusted to the new contract without renaming reviewer tests. No CI workflow or spec edits.

## Complete vector mapping

For each row, A = `TestNetworkContactVectors/<row>` (Gate and GateReadOnly, fresh/cache), I = `TestProfileNetworkContactVectors/<row>/install` (Install), M = `TestProfileNetworkContactVectors/<row>/migration` (EnsureDefaultWithPolicy). Every A/I/M terminal is pass; exact names and verdicts are also in coverage-evidence.json.

| # | Corpus row | Executable tests |
| ---: | --- | --- |
| 1 | loopback-localhost | A / I / M |
| 2 | loopback-ipv4 | A / I / M |
| 3 | loopback-ipv6 | A / I / M |
| 4 | loopback-case-dot | A / I / M |
| 5 | reserved-test | A / I / M |
| 6 | reserved-example | A / I / M |
| 7 | reserved-invalid | A / I / M |
| 8 | reserved-localhost | A / I / M |
| 9 | reserved-nested | A / I / M |
| 10 | reserved-case-dot | A / I / M |
| 11 | reserved-idn | A / I / M |
| 12 | reserved-punycode | A / I / M |
| 13 | lookalike-nested-test | A / I / M |
| 14 | lookalike-nested-example | A / I / M |
| 15 | lookalike-hyphen | A / I / M |
| 16 | bare-test | A / I / M |
| 17 | bare-example | A / I / M |
| 18 | bare-invalid | A / I / M |
| 19 | bare-test-dot | A / I / M |
| 20 | lookalike-idn-label | A / I / M |
| 21 | lookalike-punycode-label | A / I / M |
| 22 | lookalike-idn-real-tld | A / I / M |
| 23 | lookalike-case-dot | A / I / M |
| 24 | documentation-com | A / I / M |
| 25 | documentation-net | A / I / M |
| 26 | documentation-org | A / I / M |
| 27 | documentation-subdomain-com | A / I / M |
| 28 | documentation-subdomain-net | A / I / M |
| 29 | documentation-subdomain-org | A / I / M |
| 30 | documentation-case-dot | A / I / M |
| 31 | canary-undeclared | A / I / M |
| 32 | canary-declared | A / I / M |
| 33 | mixed-reserved-first | A / I / M |
| 34 | mixed-real-first | A / I / M |
| 35 | mixed-cross-file | A / I / M |
| 36 | declared-exact | A / I / M |
| 37 | declared-glob | A / I / M |
| 38 | declared-other-host | A / I / M |
| 39 | missing-manifest | A / I / M |
| 40 | missing-network-declaration | A / I / M |


## Refresh onto 402fdbc9

The converged README retains CLIHELP, SNAPPUB and this task's network-vector instructions exactly once, in a sensible order. The only round-2 repository edits move the network-vector paragraph beside the audit commands and correct the cache description to ruleset 4 (retaining ruleset-2/3 records). Implementation, tests and dependencies are byte-identical to revision 1; the upstream CLIHELP and SNAPPUB changes remain present. README is the sole out-of-module rework edit, explicitly required by the round-2 brief. No AC/vector row is out of contract. Revision 1 received no product-review verdict: reviewer preparation was refused with `revision_base_superseded`; its original refusal evidence remains attached, and this run followed the explicit development rework branch. The original developer brief still has no surface table; its previously declared coverage-map brief gap remains unchanged.

AC coverage remains **4 of 5 rows driven** through `audit.Gate`, `audit.GateReadOnly`, `envprofile.Install` and `envprofile.EnsureDefaultWithPolicy`; row 5 is validation-only. All 40 vectors / 120 production-path terminals and the same six named narrowing mutants pass on this candidate. The mutant table above remains applicable: every child exits 1 for its named behavioral assertion; no survivors. All 454 previous audit/envprofile Test names are retained and green.

| Command / verification | Real exit | Evidence |
| --- | ---: | --- |
| Full local suite, split into 112 bounded standalone `go test -json -count=1 -timeout=8m` commands | 0 each | 85 packages; 3499 independently inventoried Test names; 3469 pass / 31 skip unique top-level terminals; 54256 events |
| `go build ./...`; `go vet ./...`; `golangci-lint run` | 0 each | Exact candidate; isolated-v2 environment |
| `gofmt -l cmd internal`; `git diff --check` | 0 each | Empty formatting/whitespace output |
| `bash .github/ci/ledger-consistency.sh`; `bash .github/ci/platform-case-gate.sh` | 0 each | Fresh ledger and merged current-candidate stream |
| `bash .github/ci/gate-selftest.sh`; `bash .github/ci/no-broad-suppression.sh` | 0 each | Gate rejection controls and suppression guard |
| Evidence assembler positive / missing receipt / digest-consistent truncated stream | 0 / 1 / 1 | Both expected-red controls fail for missing evidence; not reported as passing suites |
| `sh scripts/remote-gate.sh` | 0 | [Full hosted run 39096450283](https://github.com/relux-works/curator/actions/runs/39096450283), exact signed head `5ceef5a6656555f87460ec4c05875d18c6e0b5fb`, completed/success |
| Candidate signature / worktree equality / implementation equality | 0 | Signature G; tree `7ced0d2be8a93bfd6063b22e70f83c8e37a7e29a`; source comparison to revision 1 is empty |

Diagnostic: the first `curator-19` command exited **1** before test execution because its package-host GOROOT lock wait expired (300 seconds). I had overlapped CLI groups. The original log/receipt are retained under diagnostics; the unchanged command was rerun after serializing remaining CLI work and exited 0. The failed attempt is not accepted suite evidence.

Every suite command above was rerun in this round; no revision-1 green log substitutes for the converged-tree run. Selection adds the five new CLIHELP tests omitted by the old masks; the 18 new snapshot tests and one CI-workflow test are included by full-package commands. The independent inventory and complete per-command receipts reject missing/truncated logs. Go 1.25.5 darwin/arm64, Node 25.6.1, pnpm 10.33.0, Rust 1.91 and Swift 6.4 match the earlier environment; the rc14 protocol root is unchanged. No filtered package contains Example/Fuzz entries outside the inventoried Test masks (search exit 1 means no matches, not a green test command).

Sanitized fresh logs, command arguments, real exit codes, digests, 40-row mapping and native hosted job verdicts are attached as `TASK-261017-sxydvh_refresh_{network,cli,install,godriver,envprofile,packages,validation}.jsonl`, with `TASK-261017-sxydvh_refresh_index.json`. The logbook resource records the README combination finding. The board-generated per-revision validation log is the authority for its managed handoff gate; this standalone hosted run is supplemental exact-tree evidence. The configured handoff command runs `sh scripts/remote-gate.sh` itself; its required managed validation binding cannot be supplied by the standalone receipt. No refresh-candidate operation was needed: convergence already placed HEAD and the recorded Story checkpoint at `402fdbc9`; no branch commit or index change was made.
