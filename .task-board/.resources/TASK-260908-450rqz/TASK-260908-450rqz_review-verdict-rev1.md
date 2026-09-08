# TASK-260908-450rqz — A1 mapping CR1 review

Verdict: changes_requested
Route: to-dev
repeat-of: none
Candidate: CR-TASK-260908-450rqz-1 revision 1; tree 9ff5c7588e3ef9705be5bfa92d20d321f4619e07; base 84e659e1bda41c0b70fad72e9e29b3c7ad474a7d.
Repository delta: present, seven scoped paths. No previous review finding for this leaf. Reviewer run RUN-260908-c1ad07 reports no bound run goal; queried before verdict. No operator directives.

## F1 — P2: specification version contradicts the settled goal

repeat-of: none

The authoritative original goal at `/Users/iv/Developer/ReluxWorks/curator/.temp/goals/goal-launcher-and-infra-migration.md:61` says: `SPEC bumps to 0.3.0-draft only for recorded errata found by implementation.` The active parent goal explicitly preserves this file, and mapping-review.md explicitly names this destination. E1/E2 qualify as recorded errata, so choosing 0.2.2-draft is not an available alternative release policy.

CR1 instead pins 0.2.2-draft in SPEC.md:3, :844, :860; README.md:24; cmd/curator-run/main.go:29; and TestSpecVersionPinned at cmd/curator-run/main_test.go:231. Independently running `go run ./cmd/curator-run --version` on the exact frozen tree returns `curator-run 0.0.0-dev (specification 0.2.2-draft)` (exit 0), failing the authoritative 0.3.0-draft assertion. The existing pin test is green because its expectation copies the wrong destination.

Minimal rework: replace these six current-version occurrences with 0.3.0-draft, preserving the Pi erratum text and historical version rows. Correct the producer outcome's version statement. Keep the mapping implementation, refusal tests, native argument boundaries, dependency-free go.mod, and existing evidence. Rerun the focused version/info checks and configured publication validation; no further product stage or upstream release is needed. This is ordinary implementation rework, not a human decision or external blocker.

## Coverage reviewed before code

Producer reports **9 of 10 behavioral AC rows driven**, plus one explicit native-tail downstream bound. This is a truthful scoped report; it does not establish version-policy compliance (F1).

| Behavioral row | Driving test / production call site |
|---|---|
| claude_code -> claude-code/claude | TestRunMapping/claude_code; run -> fragment.Resolver.Resolve -> mapping.Resolve |
| codex_cli -> codex/codex | TestRunMapping/codex_cli; same |
| pi -> pi-native/pi | TestRunMapping/pi; same |
| known opencode refuses and stops | TestRunMapping/opencode; same; no later not_implemented diagnostic |
| unknown successfully resolved ID refuses | TestRunUnknownResolvedMapping; run -> injected successful resolver -> real mapping.Resolve |
| unknown actual fragment refuses resolution | TestRunUnknownFragmentStillRefusesResolution; run -> real Resolver.Resolve |
| failed resolution precedes mapping | TestRunResolveFailurePrecedesMapping; run -> real Resolver.Resolve |
| info skips resolution | TestRunInformationalFlags; run -> cli.Parse, forbidden resolver |
| usage skips resolution | TestRunUsageErrorsExit2; run -> cli.Parse, forbidden resolver |
| native argv preservation bound | TestNativeTailVerbatim verifies cli.Parse's native slice; TestRunMapping verifies run input unchanged. No downstream exec exists; delivered launch argv is not claimed. |

The unknown-resolver double models a future successful resolver, not a claim that the current closed fragment parser admits unknown IDs. Mapping's only production caller is main.go:93, after the successful resolve return; main calls run. TestKnownUnsupportedIsNotUnknown preserves registry membership for opencode separately from launch support. Supported pairs flow into the current later-stage stub. No retry/resume path or implemented defaults/plan/exec stage exists in this scope.

## Independent validation and attacks

Tests ran sequentially in a disposable archive of the exact candidate under .temp/TASK-260908-450rqz-review/candidate. Original source was never edited. All frozen-tree files match both the source worktree and isolated archive after tests/mutant restoration.

| Check | Result |
|---|---|
| go test ./cmd/curator-run ./internal/mapping ./internal/cli -count=1 -v | exit 0; includes production fake-Curator/executable boundaries and retained CLI coverage |
| python3 .scripts/mapping-mutants.py <review evidence>/mutants | exit 0; all named expected-red tests failed |
| M1: admit only opencode | TestRunMapping/opencode fails, Go exit 1 |
| M2: admit only future_env | TestRunUnknownResolvedMapping fails, Go exit 1 |
| M3: continue only after opencode mapping error | TestRunMapping/opencode fails, Go exit 1 |
| M4: Pi system becomes legacy pi | TestRunMapping/pi fails, Go exit 1 |
| go run ./cmd/curator-run --version | exit 0; independent goal-version assertion FAIL (F1) |
| candidate byte comparison before and after attacks | no differences |

M1/M2 narrow the refusal class without deleting the gate. M3 attacks the production bypass path; the behavioral suite catches it. M4 guards the exact Pi pair. These are runtime tests, not static token searches. No new source-text authorization gate exists; the source-token mutant rule is inapplicable to this mapping stage.

Accepted existing evidence rather than rerunning the broad suite: CR1 publication-validation.log records exact runtime `make check` exit 0 (build, formatting, vet, all tests, race). The producer's make-check and four mutant logs are also attached and were inspected. The independent focused run does not replace or claim to rerun that full check.

A0 findings resource was independently retrieved: E1 records the legacy wrapper overwriting the managed home and E2 the absent Pi runtime. A0 and accepted Pi design tasks report done; the producer upstream.json records PR23 MERGED at a2a6e9f377f62a5872d99ecdfff0d1690e385f2a. This review accepts that attached upstream evidence and the explicit task brief; it did not perform a fresh remote lookup or runtime/home launch. No dependency import or tag is needed for a closed string mapping. Other SPEC/API errata remain outside this leaf.

## Evidence and lifecycle

Evidence bundle: TASK-260908-450rqz_review-evidence-rev1.tar.gz contains focused log, four mutant logs and summary, independent version failure and goal-file hash, before/after candidate comparisons, readiness/recovered read failures, and accepted publication validation. No commits, hosted CI, installs, tags, real ax/models, runtime-home mutations, private board writes, or LOGBOOK/control-root writes were performed. Findings are persisted on the board under the explicit no-LOGBOOK constraint.

Review acceptance is withheld only for F1. Mapping architecture, focused tests, and refusal attacks pass; implementation-matches-AC remains unchecked. Route to producer for the minimal version correction and another independent review cycle.
