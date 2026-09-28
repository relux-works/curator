# TASK-260922-1s0zja — revision 2 review verdict

Verdict: ACCEPTED. Reviewed exact candidate tree `48207e294c24ba91ca780bb3eebbaf1ddb01b6d5`, base `11a76741a034e2936fbba5183f7cec8c5a42d082`, CR-TASK-260922-1s0zja-2.

Scope: revision-2 rework F1/F2 and evidence accuracy correction. Previously accepted revision-1 behavior remains accepted under the binding review note; not reopened. Read the producer's revised results and Decision 0018. No product files changed by this reviewer. Tests used a disposable `git archive` of the exact candidate inside the assigned worktree.

## Findings resolved

F1: `pkg/agentic/systems/claude/argvguard_test.go:379` now uses module-wide LiteralSites with a closed file/declaration allowlist: exactly one Claude `bypassPermissionsFlag` const and exactly one Agy `Args` occurrence. Unknown sites, repeated occurrences at an allowed declaration, and absent known sites fail. Codex's `argvguard_test.go:378` remains module-wide and requires exactly its one const site. The committed third-package and second-Agy-site counter tests pass.

F2: README and CHANGELOG now distinguish the entry points correctly. In `plan.go:295`, BuildPlan invokes refuseNonInteractiveParameters; its composition check at line 178 returns ErrCompositionNotInteractive before Argv. Direct Claude interactiveArgs (`args.go:153`) and Codex Args (`args.go:141`) scan Composition.Prefix and return ErrPermissionModeDuplicate for an exact flag element. README's interactive paragraph includes the optional typed permission member. The revised results correctly limit the early-validation claim to no Argv construction and acknowledge the earlier Capabilities call.

## Independent validation

Shell: zsh. Commands executed directly, no output pipeline masking exit codes. Mutant driver: Python subprocess argv with captured return codes, launched from zsh. Every Go test uses `-count=1`.

`go test ./internal/argvguard ./pkg/agentic/systems/claude ./pkg/agentic/systems/codex ./pkg/agentic/systems/pi ./pkg/agentic/systems/pinative ./pkg/agentic ./internal/regress -count=1`: exit 0, 7/7 packages passed.

`git diff --check BASE CANDIDATE`: exit 0.

The handoff's full `go build ./... && go test ./... -count=1` exit-0 evidence is reused as instructed by the binding review note; I did not rerun that full suite or independently rerun vet/build. No cross-platform claim.

## Adversarial checks: 4/4 killed; restoration 4/4 passed

Each mutation ran the committed `TestTheBypassFlagIsSpelledAtExactlyTwoKnownSites` via `go test ./pkg/agentic/systems/claude -run '^TestTheBypassFlagIsSpelledAtExactlyTwoKnownSites$' -count=1`.

| Mutation in disposable candidate | Mutant exit | Restored exit |
|---|---|---|
| Original rev1 reproduction: pkg/agentic/review_third_spelling.go, reviewThirdSpelling returning the literal | 1 | 0 |
| Same extra literal in pkg/agentic/systems/claude/review_second.go | 1 | 0 |
| Same extra literal in pkg/agentic/systems/agy/review_second.go | 1 | 0 |
| Change Claude const literal to --review-silent | 1 | 0 |

The last mutation changes the scan target too (the test references the const); it still fails because the required Agy site no longer matches. Full diagnostics attached as TASK-260922-1s0zja_review-rev2-mutations.log; executable driver attached separately.

Bound: source ownership is measured over gosources.Walk's non-test module sources, excluding nested modules, dot/underscore directories, vendor, node_modules and testdata. LiteralSites counts string literals containing the spelling, not computed runtime strings. These are static spelling proofs, not claims about all possible runtime construction techniques.

F-M1b owns capability-table versioning, drift and parsing rules; F-L1 owns launcher-side raw-argument integration. No release tag belongs in this leaf. No additional blocker or rework found.

Lifecycle: spawn goal queried; this run is not goal-bound. Acceptance routes revision 2 to integrating, not done; producer-side integration remains outstanding.
