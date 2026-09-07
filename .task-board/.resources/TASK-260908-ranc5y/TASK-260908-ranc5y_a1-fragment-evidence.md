# TASK-260908-ranc5y — A1 fragment delivery (SPEC §4.1): producer evidence

Worktree `.worktrees/launcher-control/.temp/STORY-260908-15st55/worktree`, branch `task-board/story/STORY-260908-15st55`, base `25379f2` (parser PR4). Candidate left uncommitted.

## Changed files

| Path | Change |
|---|---|
| `internal/fragment/json.go` | new: CCJ-1 strict JSON reader (valid UTF-8, unique keys, integers only, no lone surrogates, no trailing content, depth bound 64) producing an ordered value tree |
| `internal/fragment/ccj1.go` | new: `Canonical` (registry.md §1 rules 1–8) and `Digest` (`sha256:<hex>` over the canonical bytes of the parsed object) |
| `internal/fragment/fragment.go` | new: closed `launch-env-fragment-v1` parser against the conformance schema plus the environments.md §7.3/§7.8 adapter channel registry; typed `Fragment`; `InvalidError` |
| `internal/fragment/resolve.go` | new: `Resolver.Resolve` — argv `env resolve <env> [--profile <name>] --repair --format json`, `Runner` injection at the process boundary, `ExecRunner` production runner (os/exec, stderr streamed verbatim), `curator: <code>: <detail>` mapping to `resolve_*`, no retry, no fallback |
| `internal/fragment/*_test.go` | new: reader/canonical rules, conformance corpus, A0 goldens, schema negatives, resolver mapping, real subprocess boundary with a shell fake curator |
| `internal/fragment/testdata/schema-cases/` | vendored: the 49 indexed `launch-env-fragment-v1` rows of curator-spec `conformance/v1/schema-cases` at `87a0d0060bad64ab883d007dcdf35df7485368bf`, verdicts in `index.json` |
| `internal/fragment/testdata/a0/` | the three A0 fragments (`verified-fragment-*.json`) and `digests.txt` from the task resources |
| `cmd/curator-run/main.go` | §4.1 wired after §3 parsing: resolve through `fragment.New()`, print `resolve_*` code line exit 1, else `not_implemented` refusal exit 1 with home and digest |
| `cmd/curator-run/main_test.go` | runner injection; launch resolves exactly once with exact argv; every resolve failure as code line; production resolver against a PATH fake; usage/info never resolve |
| `SPEC.md` §4.1 | E6 clarification only (see below) |
| `README.md` | status and tools table (fragment harness, corpus) |
| `.scripts/fragment-mutants.sh` | new narrowing-mutant harness (26 mutants) |
| `.scripts/cli-mutants.sh` | M14 re-anchored to the new refusal line and test name |

No new module dependencies; `go.mod` unchanged (standard library only). No ax, model/default, composition, or runtime-home code.

## Commands and real exit codes

| Command | Exit | Evidence |
|---|---|---|
| `make check` (build, fmt-check, vet, test, race) | 0 | `make-check-01.log`, `make-check-02.log` (second run after the mutant harnesses restored sources) |
| `go test -count=1 -cover ./...` | 0 | fragment 95.8 %, cmd 78.3 %, cli 98.4 % |
| `git diff --check` | 0 | clean |
| `.scripts/cli-mutants.sh` | 0 | 14/14 KILLED (`cli-mutants/summary.tsv`) |
| `.scripts/fragment-mutants.sh` | 0 | 26/26 KILLED (`fragment-mutants/summary.tsv`) |
| `curator env resolve <env> --profile default --repair --format json` (installed `v0.14.1-0.20260907213730-04550e282705`, read-only current homes) | 0 ×3 | digests recomputed: claude_code `c7d17c39…545e`, codex_cli `04459237…2e32`, pi `c0512f55…01bf` — equal to the A0 record; stderr `warning: environment_tool_version_unverified` for claude/codex (2.1.263 / 0.153.4), empty for pi |
| `curator env resolve nope --repair --format json` | 1 | `curator: environment_unknown: unregistered environment "nope"` |
| `curator env resolve pi --repair --format json` (no `--profile`, no current profile on this machine) | 1 | `curator: profile_unknown: no profile is current` |

Expected-red: none. No model was launched; no home was written by the launcher (the installed-curator probes were read-only on already-current homes; Curator itself repaired nothing, as its stderr shows).

## AC coverage — 9 of 9 rows driven

| AC row | Production call site | Named test(s) |
|---|---|---|
| subprocess always resolves with `--repair`, exact argv order | `Request.Argv` → `Resolver.Resolve` (`resolve.go`); `run` (`main.go`) | `TestArgvExactOrderAndRepair`, `TestExecRunnerEndToEnd`, `TestRunParsedLaunchResolvesThenRefuses`, `TestRunProductionResolverAgainstFakeCurator` |
| argv boundaries preserved (`--profile` as a separate token, operands intact) | same | `TestArgvExactOrderAndRepair` (space/`--repair`-containing profile, `-x`), mutant M03 |
| stderr forwarded verbatim incl. warnings on success | `Resolve` `io.MultiWriter(stderr, &captured)` | `TestResolveSuccessForwardsWarnings`, `TestNonZeroExitMapping`, `TestExecRunnerEndToEnd`, main tests (warning precedes the refusal/code line) |
| documented error codes mapped; home_stale/unknown → invocation_failed; start failure → invocation_failed | `curatorCodes`, `curatorDiagnostic`, `Resolve` | `TestNonZeroExitMapping`, `TestStartFailureIsInvocationFailed`, `TestExecRunnerNonZeroAndInvalidOutput`, `TestExecRunnerMissingBinaryAndCancellation`, `TestRunResolveFailuresExit1` |
| strict closed parsing; malformed/unknown/duplicate fail closed; withdrawn composition | `Parse` / `fromValue` / `channel` (`fragment.go`), `ParseJSON` (`json.go`) | `TestConformanceCorpus` (9 valid + 40 invalid), `TestRejections` (98 cases), `TestReaderRejections` (40), `TestInvalidOutputIsFragmentInvalidNeverAbsence` |
| CCJ-1 digest matches actual A0 fragments; canonical bytes equal Curator's printed line | `Canonical`, `Digest` | `TestA0FragmentsParseAndDigestMatch` (3/3), `TestDigestInvariantUnderPrintingAndSensitiveToBytes` |
| canonicalization adversarial: Unicode scalar key order, escaping incl. unescaped U+2028/U+2029, integers, surrogates | same | `TestCanonicalRules`, `TestReaderRejections`, `TestReaderAccepts`, `TestDigestIsSHA256OverCanonical` |
| no retry, no fragment-less fallback; invalid output ≠ absence; non-zero with valid stdout still fails | `Resolve` | `fakeRunner.calls == 1` in every resolver test, `TestNonZeroExitNeverYieldsFragmentEvenWithValidStdout`, `TestInvalidOutputIsFragmentInvalidNeverAbsence`, mutants M04/M05/M08 |
| E6 wording only in SPEC; no other stages/imports | SPEC §4.1 paragraph; `go.mod` unchanged | inspection: `git diff SPEC.md` is one paragraph; `git diff go.mod` empty |

## Mutant evidence (all narrowing; gate stays present)

| Mutant | Narrows the gate to admit | Failing test | Bound stated |
|---|---|---|---|
| M01 | `environment_home_stale` mapped to `resolve_repair_failed` | `TestNonZeroExitMapping` | — |
| M02 | argv without `--repair` | `TestArgvExactOrderAndRepair` | — |
| M03 | `--profile=<name>` single token | `TestArgvExactOrderAndRepair` | — |
| M04 | exit 1 with parseable stdout as success | `TestNonZeroExitNeverYieldsFragmentEvenWithValidStdout` | — |
| M05 | empty stdout treated as absence | `TestInvalidOutputIsFragmentInvalidNeverAbsence` | — |
| M06 | codex_cli fragment for any requested env | `TestInvalidOutputIsFragmentInvalidNeverAbsence` | — |
| M07 | token-preserving: `curator: ` matched mid-line | `TestNonZeroExitMapping` | source-text gate attacked with token preserved; behavioral suite ran |
| M08 | one retry on exit 1 | `TestNonZeroExitMapping` | — |
| M09 | stderr forwarded only on failure | `TestResolveSuccessForwardsWarnings` | — |
| M10 | duplicate `environment` key (last wins) | `TestRejections` | — |
| M11 | lone low surrogate repaired to U+FFFD | `TestReaderRejections` | — |
| M12 | lowercase-exponent numbers | `TestReaderRejections` | — |
| M13 | trailing second object | `TestReaderRejections` | — |
| M14 | `/` escaped in canonical strings | `TestA0FragmentsParseAndDigestMatch` | — |
| M15 | U+2028/U+2029 escaped | `TestCanonicalRules` | — |
| M16 | UTF-16 code-unit key order | `TestCanonicalRules` | — |
| M17 | withdrawn `composition` member | `TestConformanceCorpus` | — |
| M18 | registry grants pi an MCP descriptor | `TestConformanceCorpus` | — |
| M19 | `semantics: prepend` at enum and registry compare | `TestConformanceCorpus` | — |
| M20 | empty `with` list | `TestInvalidErrorShape` | registry equality still rejects; only the path-level assertion distinguishes |
| M21 | duplicate `env_names` | `TestRejections` | — |
| M22 | one-descriptor subset of the system-prompt list | `TestRejections` | — |
| M23 | `/usr/local/bin` as `path_prepend` | `TestConformanceCorpus` | — |
| M24 | two `env` variables | `TestConformanceCorpus` | — |
| M25 | `sha256:`-prefixed lock hash | `TestConformanceCorpus` | — |
| M26 | entry point drops the `resolve_fragment_invalid` code line | `TestRunResolveFailuresExit1` | — |

Survivors: none (0 of 26). CLI harness re-run after the wiring change: 14/14 killed.

## Source → evidence mapping

- SPEC §4.1 argv/`--repair`/failure modes → `resolve.go`; environments.md §10.1/§10.4 codes → `curatorCodes`.
- environments.md §10.2 + Decision 0012 D8 + `schemas/v1/launch-env-fragment-v1.schema.json` (common identifier/hex256/portablePath, marker precedence, agent-mcp envNames exclusions) → `fragment.go`; the `path_prepend` environments-root rule uses the `<root>/<profile>/<env>` layout (§8.1), root = two segments above the home.
- environments.md §7.3/§7.8 channel registry (claude_code, codex_cli, opencode, pi) → `systemPromptRegistry`/`mcpRegistry`; exact list equality in registry order.
- registry.md §1 CCJ-1 → `json.go` (rejections), `ccj1.go` (emission).
- A0 findings E6 (TASK-260908-qblycn_a0-verification-findings.md §6) + reviewer verdict rev1 + installed-curator reproduction today → SPEC §4.1 paragraph; STORY-260908-2utz8k referenced in the text.

## Stated bounds and notes for the reviewer

- Channel lists must equal the adapter registry exactly and in order (strictest reading of "reproduces the adapter's descriptors"; corpus `invalid-*-channel-not-registry` requires registry knowledge). A registry revision in curator-spec means a launcher change.
- `environment` values are the closed four; `opencode` parses (registry has its channels) though §4.2 will map it `env_unsupported`.
- `sig` is handled per registry §1 rule 1 in `Canonical` but is rejected earlier by `Parse` as an unknown member; the rule is exercised only through `ParseJSON`/`Canonical` directly.
- The conformance directory carries 13 stale pre-D8 files not present in `index.json` (e.g. `valid-empty-channels.json` with `profile.commit`); they were not vendored. Worth a curator-spec housekeeping note, not a launcher concern.
- Context cancellation maps to `resolve_invocation_failed` (process not started/completed); no distinct code exists in §6.
- `run` in `main.go` now takes a `context.Context` and a `*fragment.Resolver`; `main` passes `fragment.New()`.
