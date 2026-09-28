# TASK-260922-1s0zja rework 1 → revision 2 (orchestrator brief, binding)

Verdict: CHANGES_REQUESTED on revision 1 (`TASK-260922-1s0zja_review-verdict-rev1.md`): one P1 and
one P2. Everything else the reviewer judged correct stays as committed — member, `Resolve`, named
sentinels, zero/native byte-identity, mapping order, pi/pinative refusal, the duplicate check's
scope, and the F-M1b/F-L1 boundaries. Do not re-open them.

## F1 (P1) — the ownership guard admits a THIRD spelling site
`pkg/agentic/systems/claude/argvguard_test.go:374` filters out every package but claude before
counting, and `:428` only requires Agy's occurrence to be non-empty. The agy sharing justifies
exactly TWO allowed sites, not an open exception. Reviewer's reproduction (survived, exit 0):

```sh
cat > pkg/agentic/review_third_spelling.go <<'GO'
package agentic
func reviewThirdSpelling() []string { return []string{"--dangerously-skip-permissions"} }
GO
go test ./pkg/agentic/systems/... ./internal/argvguard -run 'Argv|Spelling|Sharing|Bypass' -count=1
```

Fix: a MODULE-WIDE `LiteralSites` check over the non-test sources that admits exactly the two known
locations — the claude plugin's const and agy's construction site — and fails on any other
occurrence, including a second site inside agy (the current "agy is non-empty" assertion is not a
bound). Commit the third-package mutant as an executed row: with the extra literal planted the gate
must FAIL, without it pass. No agy argv change, no golden change. Do the same for
`--dangerously-bypass-approvals-and-sandbox` if any second site exists for it (the codex proof is
already module-wide — state which shape each flag ends with).

## F2 (P2) — documents contradict the production error contract
`CHANGELOG.md:12-15` and `README.md:92-96` say `BuildPlan` returns `ErrPermissionModeDuplicate` for
a yolo composition prefix. It does not: `BuildPlan` refuses composition earlier with
`ErrCompositionNotInteractive`; `ErrPermissionModeDuplicate` is the DIRECT plugin `Argv` path only
(your own committed test proves this). Correct both documents to name the two entry points
separately, and update the README interactive paragraph at :77-79, which still says "model/effort
only" and must acknowledge the optional typed permission-mode member.

## Accuracy note (no rework, do not overclaim again)
Results.md said a refused value "triggers no plugin surface, not even a file read"; the reviewer
found `sys.Capabilities()` is called before validation (`plan.go:249`). Keep the claim to what the
tests measure (no `Argv` construction) in the revised results.

## Rules
No product behaviour change beyond F1's guard and F2's documents. Re-run the narrow packages and the
new third-site mutant with real exit codes (state the shell, `-count=1`), the board validation
command runs once at handoff, then `task-board handoff TASK-260922-1s0zja --role developer`.
