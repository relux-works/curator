# Review verdict: ACCEPTED — CR-TASK-260909-2vy977-1 revision 1

Reviewer run RUN-260915-7b7aa6 (claude-fable-5-1, low). Independent of producer evidence; no source edits, no commits, no hosted CI, no installs, no runtime-home writes, no tags.

## Exact candidate
- Base OID cb232a120c9a04c56ae5037c82347921f020e688 (= worktree HEAD, verified: candidate is uncommitted).
- Candidate tree OID 680a1da6162e183e07d83e1e26cd012411cd5f46: recomputed from the working tree via a temporary index before AND after my mutant run — equal both times.
- Patch resource sha256 c69b2cd2b7e4c1236d620b941d16161b93170ff55b4b21c0f66f33db45bc8afc — matches the assigned CR.
- 13 changed paths as listed in the CR.

## Runtime validation (the log the DoD names)
- TASK-260909-2vy977_change-request_rev1-validation.log, sha256 080908f9dcd43cb6730a1d97f77a0fcecd1c772a2ae25c6af6541ffdd56aed8c: `make check` (build, vet, test, test -race across 11 packages) ends `[exit 0]`.
- Provenance verified: board payload mtime 2026-09-15 22:08:50, identical to the rev1 patch payload; the log lists internal/plan, which exists at cb232a1 and not at the September 9 base 289ff42, so it cannot be the historical CR1 log.
- Producer spawn log RUN-260915-042561 contains no manual `make check` or `go test ./...` invocation (only prompt text). Manual full-suite count for this revision: 0. Runtime count: 1. This closes repeat finding F3 for this revision.
- Source-owned attestcheck (skill-project-management tools/board-cli, run via `go run`, no install) over a structured attestation of the producer's claims (make check exactly 1, runtime provenance, 11 of 12 rows, row 12 kind dependency-inspection): verdict ok=true, exit 0. Attached as TASK-260909-2vy977_review-attestation-rev1.json and TASK-260909-2vy977_review-attestcheck-verdict-rev1.json. Bound: it validates declared claims against declared records, not authenticity.

## Independent reruns (zsh, set -o pipefail, GOWORK=off)
| Command | Exit |
|---|---|
| go test ./internal/defaults ./internal/diagnostics ./cmd/curator-run -count=1 | 0 |
| go vet (same three packages) | 0 |
| go build ./... | 0 |
| gofmt -l (three package dirs) | 0, no files |
| python3 .scripts/defaults-mutants.py with DEFAULTS_MUTANT_IDS = model-lock, exit-usage-downgrade, resolution-failure-admit, pi-union-leak, pi-empty-runtime, pi-preference-order, pi-configured-restrict, pi-unpreferred-admit, path-failure-as-absence, cross-driven-row | 0; 10 of 10 killed (each child go test exit 1 with the named --- FAIL) |
| git ls-remote refs/tags/v0.5.11 and ^{} | 0; 0ea486e46765ecf12fff7c2ac526e12da02e95ed / a2a6e9f377f62a5872d99ecdfff0d1690e385f2a |

Not rerun: the full 42-mutant harness (10 of 42 reran, all narrowing mutants; the remaining 32 are accepted from the attached mutants-current.tar.gz), the full suite (runtime log used, per DoD), tag signature verification (accepted from primary/producer records).

## Dependency
go.mod requires github.com/relux-works/skill-agents-management v0.5.11; no replace; no go.work; GOWORK empty; `go list -m -json` reports Version v0.5.11 with no Replace field from the module cache.

## Semantics audit (F4 closure)
- lineup.go: for pi-native with no configured model, it takes the first runtime in the constant order pi-anthropic, pi-openai, pi-google that has at least one row driven by pi-native, then calls vendorplugin.Lineup over THAT runtime's rows only. No cross-vendor union is ranked. Empty preference set is defaults_unresolvable with the ordered list in the detail. Configured models bind their exact contributor across all pi runtimes, independently of the preference; ambiguous contributors are refused, not guessed.
- The ordered preference matches the orchestrator-recorded convention (defaults-lineup-brief.md, campaign-producer-rules.md). SPEC §4.3 gains the table row plus the "vendor scales are never compared" sentence, and a dated changelog entry; SPEC stays 0.3.0-draft. README matches. No invented normalization, no hard-coded vendor, no refusal-only Pi.
- Effort semantics: effortForRow uses Effort.Recommended only when Effort.Support is Required; a no-effort row leaves effort unset; the display Recommended flag is never consulted. Own-row recommendation (claude-opus-4-8 -> xhigh) driven at run.
- No silent retry: Complete is called once from run; failures render one diagnostic line and return.

## Production wiring reviewed
cmd/curator-run.run: mapping -> configPaths (evaluated only for a mapped launch; discovery error is defaults_config_invalid) -> defaults.Load -> Files.Complete(env, flags, registry) -> defaults.EmitGroup -> interim not_implemented line -> exit 1. defaults.Error carries the code; usage (locked flag) exits 2, operational failures exit 1. main() builds the registry via defaults.NewRegistry, which registers claude-code, codex, pi-native, gemini-cli and antigravity systems, seeds frozen runtimes, and registers the anthropic/openai/google vendors; the legacy pi wrapper plugin is not registered. diagnostics: defaults_unresolvable now owned by defaults.Files.Complete, dropped from RemainingObligations with the owner pin test updated. main() itself and processDefaultsPaths are not driven by tests (stated bound; the tests inject launchDeps).

## Coverage (my own reading of the candidate tests)
11 of 12 AC rows driven at cmd/curator-run.run with the real v0.5.11 registry (testDeps calls defaults.NewRegistry); row 12 (tag pin) is inspection only. 3 of 3 environments resolve positively before the pending-plan refusal: claude_code -> claude-fable-5-1/high, codex_cli -> gpt-6-astra/max, pi -> claude-fable-5/high (TestRunLineupEnvsPrintGroupBeforeRefusal, TestRunPiResolvesNativeLineup). Negative production-entry cases: locked model/effort flags (exit 2), unknown member, broken symlink, path discovery failure, no preferred driven row, unpreferred-only runtime (all refuse before any group line). Helper-level tests (TestCompleteAmbiguousRuntime, TestCompleteRuntimeResolutionFailure, TestCompleteWrongVendorModels, TestCompletedPairsAdmitTaggedBuildLaunch) are bounds, not counted as driven rows.

## Bounds carried forward (not defects of this leaf)
- The plan request, BuildLaunch admission, provider limits, composition, prompt, late checks and exec/handoff are TASK-260908-1o7i8y / TASK-260908-2so46q scope; the interim not_implemented line stays by brief.
- An unknown configured model with no effort passes through with effort unset for the plan stage to refuse (documented design).
- No real provider execution, no ax, no cross-platform lanes.

## Checklist
All rows on the task checklist verified within the stated scope; planning-only rows (board size, traceability, research questions, dependencies, atomicity, planning artifacts) are inapplicable to this implementation leaf and were left as the producer recorded them. Logbook row: LOGBOOK writes are prohibited by the campaign rules; findings live on the board.

## Verdict
Accepted. Recorded with accept_cr(TASK-260909-2vy977, revision=1, evidence=TASK-260909-2vy977_review-verdict-e11-1-rev1.md). No commit_ack. Integration (signed branch, PR, hosted checks, fast-forward) belongs to the orchestrator-routed producer run.
