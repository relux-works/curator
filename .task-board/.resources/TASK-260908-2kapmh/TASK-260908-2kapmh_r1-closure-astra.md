# TASK-260908-2kapmh — R1 closure, Astra continuation

Ready for independent review. Preserved uncommitted Story WIP; no commit, push, tag, hosted CI, auth mutation, runtime-home edit or LOGBOOK. Parent owns PR 23 signed exact-head delivery; operator alone creates signed annotated v0.5.11 after landing and verification. No pseudo-version or committed replace.

## Candidate and changes

Resumed the interrupted producer's R1 implementation, compared actual file bytes to signed rev1 e8047d9 / tree 38639918f28d8f12ab447892c200445a2b283c37 (including untracked files; plain git diff alone omits their content). No leftover mutant found. Original accepted rev1 functionality and vendor memberships remain unchanged.

- `pkg/agentic/system.go`: optional `EffortAdmitter` capability.
- `pkg/vendorplugin/spawn.go`: `BuildLaunch` invokes it after global row effort validation; wraps `ErrEffortNotNativelySupported` with model, runtime, native vocabulary and row recommendation. Global rows and Codex are unchanged.
- `pkg/agentic/systems/pinative/catalog.go`: installed Pi parser vocabulary and per-model null-map restrictions. `args.go` applies the same refusal on the direct plugin path. No rewrite or default.
- `pinative_test.go`, `pkg/vendorplugin/pinative_runtime_test.go`: four refusal pairs, Codex parity, positive max, catalog/parser comparison. This continuation extended the catalog check through real BuildLaunch in BOTH modes for every row vocabulary word, pinned 71 pairs, distinguished missing parser from read failure, and corrected a stale comment.
- README and architecture/consumer/shipped-state docs now specify explicit refusal and effort-none default bound. Rework patch against rev1 is in the evidence archive.

## R1 coverage: 7 of 7 behavioral AC rows driven

| AC row | Production call site | Named test |
| --- | --- | --- |
| sol ultra refusal, full error, no plan | vendorplugin.BuildLaunch → EffortAdmitter | TestBuildLaunchRefusesAnEffortInstalledPiWouldDropOrClamp/gpt-5.6-sol/ultra |
| terra ultra refusal, full error, no plan | same | TestBuildLaunchRefusesAnEffortInstalledPiWouldDropOrClamp/gpt-5.6-terra/ultra |
| codex-model minimal refusal, full error, no plan | same | TestBuildLaunchRefusesAnEffortInstalledPiWouldDropOrClamp/gpt-5.3-codex/minimal |
| gpt-5.2 minimal refusal, full error, no plan | same | TestBuildLaunchRefusesAnEffortInstalledPiWouldDropOrClamp/gpt-5.2/minimal |
| sol max exact argv | vendorplugin.BuildLaunch → pinative.Args | TestTheMaxWordStillBuildsWhereInstalledPiRunsIt |
| opus max exact argv | same | TestTheMaxWordStillBuildsWhereInstalledPiRunsIt |
| all four same efforts remain supported by Codex | vendorplugin.BuildLaunch → Codex Argv | TestBuildLaunchRefusesAnEffortInstalledPiWouldDropOrClamp (each subtest) |

Additionally `TestPiNativeThinkingRestrictionsMatchTheInstalledCatalog` compares **71 of 71** declared row/effort pairs against actual installed parser/catalog bytes, then drives **142 of 142** interactive/dry-run BuildLaunch plans or refusals. Both over-refusal and under-refusal fail. `TestArgsRefusesAWordInstalledPiWouldDropOrClamp` exercises direct plugin refusal. Tests remain in the candidate for parent commit; this producer must not commit them. Original broader AC coverage and accepted limitations remain in rev1 outcome/reviewer verdict; they are not represented as independently re-proven by this focused matrix.

## Commands run in this continuation

All listed commands ran directly (no tee); logs in attached archive.

| Command / evidence | Exit | Result |
| --- | ---: | --- |
| focused go test pinative/vendorplugin, Pi/Native/Max/Args selectors (narrow-01/02) | 0 each | preserved and extended R1 tests |
| focused go test pinative/vendorplugin/providerlimits, same selectors (narrow-03) | 0 | restored gate; relevant runtime state regression |
| go test ./pkg/vendorplugin -run 'TestPiNativeThinkingRestrictionsMatchTheInstalledCatalog\|TestBuildLaunchRefusesAnEffortInstalledPiWouldDropOrClamp\|TestTheMaxWordStillBuildsWhereInstalledPiRunsIt' -mod=mod -count=1 -v | 0 | latest test-source state, r1-final.log |
| env -u TASK_BOARD_DIR make build BIN=.temp/r1-astra/agents-management | 0 | CLI compiles |
| git diff --check | 0 | whitespace clean |
| installed Node catalog-behavior.mjs, explicit empty environment plus PATH | 0 | real parser + getSupportedThinkingLevels/clampThinkingLevel |
| probes.py, seven isolated Pi invocations | script 0; each Pi 1 | expected missing-auth refusals; not passing model sessions |
| first no-prompt probe | Pi 0; harness 1 | inconclusive no-auth assertion: no prompt meant no request. Preserved pi-probes.log; corrected by explicit prompt in probes.py |

Full configured make vet/test/regress is reserved for the runtime's handoff validation ONCE, per focused brief. Prior rev1 suite was green on its own tree, not claimed as validation of this changed tree. The handoff's separately attached validation log is authoritative for the new candidate's broad suite; if it fails, handoff must refuse. This report does not pre-claim those pending commands green.

## Mutants (restored from exact in-memory original)

| Mutant | Gate narrowing | Named failing test | Exit / survivor bound |
| --- | --- | --- | --- |
| ultra-sol | nativeAccepts gate remains; admits only openai/gpt-5.6-sol ultra | TestBuildLaunchRefusesAnEffortInstalledPiWouldDropOrClamp/gpt-5.6-sol/ultra; TestPiNativeThinkingRestrictionsMatchTheInstalledCatalog | 1; killed |
| minimal-codex | null-map gate remains; admits only gpt-5.3-codex minimal among rejected pairs | TestBuildLaunchRefusesAnEffortInstalledPiWouldDropOrClamp/gpt-5.3-codex/minimal; TestPiNativeThinkingRestrictionsMatchTheInstalledCatalog | 1; killed |

No surviving mutants. The ultra mutant preserves all parser and null-map source tokens and executes the behavioral suite, not merely a textual checker. Mutant script and real output are archived. A final restored narrow run passed.

## Installed boundary and bounds

Pi 0.84.2 at `/opt/homebrew/lib/node_modules/@earendil-works/pi-coding-agent`. All three catalog SHA-256 values equal reviewer rev1: anthropic e22c277e3a1ffddc3d2701b72787c9e0bd67b835de6b4cb806677b6b6a89a2f7; openai 47746dfe79d92a7da58d7fae2c02a5d481b2eba2f5a914a42792552940b78b78; google c6d9822f7eda23cfeb7d29caee0baa7d2fa7bf3561252734a17428937fbcbfd1. Parser/models.js/package hashes are archived.

Fresh Pi homes, isolated cwd and explicit environment allowlist, offline and extension/skill/template discovery disabled: four refused pairs plus sol max and opus max/high all stop at missing API key (exit 1); only ultra emits invalid-thinking warning. Actual installed pure functions demonstrate minimal → low for both codex-model and gpt-5.2; max remains max for sol and opus. No provider credentials or model call used.

Bound: this is catalog/plan and no-auth boundary evidence, not authenticated interactive session execution. Effort-none rows inject no thinking flag; Pi may apply its own settings default. Native restriction data is pinned to 0.84.2 and must be reverified on a future Pi update. Direct plugin calls do not replace vendor model admission; BuildLaunch is the supported runtime boundary. Legacy Pi/local-model and provider-limit limitations from rev1 are preserved. No new MCP-in-Pi work.
