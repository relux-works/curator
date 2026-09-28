# TASK-260922-1s0zja revision 1 — changes requested

Reviewed exact candidate tree c394484a052796bc533620d9970e34eb5d598192 against base 11a76741a034e2936fbba5183f7cec8c5a42d082. Tests and mutations ran in a disposable git-archive copy under the assigned worktree. Product files were not modified. Verdict: changes_requested; route to to-dev.

## Findings

1. **P1 — module-wide ownership guard admits a third spelling.** `pkg/agentic/systems/claude/argvguard_test.go:374` filters out every other package before counting. `:428` merely requires at least one Agy occurrence. The existing Agy spelling justifies two explicitly allowed sites, not an unbounded exception outside Claude. A third production package can now spell the bypass flag without any argv guard failing. This fails the requested module ownership proof and the review note's explicit third-site attack.

Executable reproduction in a disposable candidate copy:

```sh
cat > pkg/agentic/review_third_spelling.go <<'GO'
package agentic
func reviewThirdSpelling() []string { return []string{"--dangerously-skip-permissions"} }
GO
go test ./pkg/agentic/systems/... ./internal/argvguard -run 'Argv|Spelling|Sharing|Bypass' -count=1
rm pkg/agentic/review_third_spelling.go
```

Observed exit 0, all nine selected packages pass. Add a module-wide LiteralSites check allowing exactly the existing Claude const and Agy construction location, with a committed third-package mutant that the actual gate rejects. This needs no Agy argv changes and preserves existing goldens. Also reject extra sites within Agy rather than merely asserting Agy is nonempty.

2. **P2 — public error contract contradicts production.** `CHANGELOG.md:12-15` says BuildPlan returns ErrPermissionModeDuplicate for a yolo composition prefix; `README.md:92-96` implies the same. BuildPlan rejects this earlier with ErrCompositionNotInteractive; ErrPermissionModeDuplicate is only the direct plugin Argv path. The committed test already proves this. Reproduce with `go test ./pkg/agentic/systems/claude ./pkg/agentic/systems/codex -run TestTheYoloRefusalsAreNamedAndTotal -count=1` (independently passing in baseline). Correct both documents to distinguish the entry points. The earlier README interactive paragraph at :77-79 also still says model/effort only and should acknowledge the optional typed mapping.

## Independent checks and attack results

Shell: zsh with pipefail for baseline; mutations invoked go directly with Python subprocess return codes. Baseline: `go test ./pkg/agentic ./pkg/agentic/systems/claude ./pkg/agentic/systems/codex ./pkg/agentic/systems/pi ./pkg/agentic/systems/pinative ./internal/argvguard ./internal/regress -count=1` — 7/7 pass. Full build/test suite exit 0 is accepted from the runtime evidence stated in the binding review note, not independently rerun. Candidate diff whitespace check passes. Existing golden/testdata paths have no changes in the exact tree delta; existing parity tests pass.

| Mutation | Test | Exit | Result |
|---|---|---|---|
| Producer unknown-value bound: add auto to native Resolve case | TestBuildPlanRefusesAnUnknownPermissionModeInEveryMode | 1 | killed, auto admitted with nil error |
| Producer duplicate bound: scan only Prefix[0] | Claude TestTheYoloRefusalsAreNamedAndTotal | 1 | killed, duplicate at index 1 admitted |
| Reviewer own attack: append bypass twice | Claude TestTheYoloArgvAppendsTheBypassFlagOnce | 1 | killed, exact argv mismatch |
| Reviewer own ownership attack: third package literal | all plugin argv/spelling/sharing/bypass tests + argvguard | 0 | survived |

Executed mutation coverage: 3/4 killed. Producer-required repeat coverage: 2/2 killed; the other four producer mutations were inspected in the submitted results, not independently rerun. Mutation logs and script are attached separately. All mutations restored in the disposable copy.

## Scope judgments

- Member, Resolve, named unknown/scope errors, zero/native behavior and supported mapping order are correct on the exercised BuildPlan paths. Validation precedes PrepareLaunchRequest and operational dispatch; strictly, static sys.Capabilities() is called before it (`plan.go:249`), so the producer's literal claim of no plugin surface whatsoever is overbroad. Tests only count Argv for this claim. No observed operational side effect arises from the static capability declaration.
- Explicit native outside interactive is defensible and required by the brief's any-nonzero rule.
- Composition.Prefix exact-element duplicate check and absence of NativeArgs are defensible for this leaf under the binding scope clarification: BuildPlan refuses composition wholesale; direct plugin Argv drives the duplicate gate. This is not evidence for launcher raw-argument parsing. Equals forms, parsing and capability/version-drift table belong to F-M1b; launcher after-double-dash coverage belongs to F-L1.
- Exact native/yolo argv tests drive BuildPlan; unsupported Pi rows drive BuildPlan. Duplicate sentinel tests drive the plugin Argv surface and independently prove BuildPlan's earlier composition refusal.
- Pi binary is available: independently ran `pi --version` (0.84.2) and `pi --help`, exit 0. Help exposes --approve/-a as trust for project-local files and --no-approve/-na, with no equivalent bypass option. Unsupported for pi and pinative is justified by local evidence plus Decision 0018.
- Claude 2.1.261 and Codex 0.153.2 match the existing args.go help-verification comments (baseline anchors), not an enforced runtime capability pin. No independent execution of those historical tool versions was performed. The submitted newer-help corroboration is producer evidence. This bound is acceptable for F-M1a; enforcement and re-verification belong to F-M1b.
- No release tag is needed or allowed in this leaf. No human decision or external blocker is required: both findings are ordinary rework.

## Review logbook

The third-site survivor is the important new finding: known shared spelling does not prevent enforcing a closed two-site allowlist. Recorded here and in board notes; no control-root LOGBOOK.md was edited, per campaign rules. Run goal query reported no bound goal. Return for another producer/reviewer cycle after these fixes.

Narrow `go vet` over the same seven packages also passed independently, exit 0, after mutation restoration.
