## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(8))

## Blocked By
- TASK-260822-b0wg3a
- TASK-260822-4p3dcq

## Blocks
- (none)

## Checklist
- [x] Composite branch assembled from accepted patch artifacts; overlaps in admission.go and skip-classes.tsv reconciled
- [x] Full gate set green on the composite (fmt, build, vet, lint, gate-selftest, ledger, full go test)
- [x] PR opened, CI green incl. interop conformance gate, merged
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-3e579b, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-3e579b)
Composite assembled in .temp/TASK-260823-qwr5w9/composite on branch task/TASK-260823-qwr5w9-credential-scopes-composite off origin/main 6a9b201.

SOURCES. Three accepted patch artifacts applied clean with git apply (exit 0 each, no --3way needed): TASK-260822-4p3dcq_final.patch (the reconciled build_ssh chain 96m5pj+2505vo+3pkc80+b0wg3a+docs, so the admission.go and skip-classes.tsv overlaps named in the task description were already reconciled upstream and needed no hand merge here), TASK-260822-27bvo4_symlink-launcher.patch, TASK-260822-2v5e80_remedy.patch.

THREE TASKS SHIPPED NO PATCH ARTIFACT and had been left uncommitted in the main checkout, which sits on the stale handoff/cocoaskills-parity-20260731 branch, not on main. Ported by hand and re-verified on origin/main: 3j4bcb closure provenance (closure.go ports verbatim; closure_test.go needed a 3-way merge, additive conflict against main sibling tests, resolved by appending; install_test.go and skillcheck_test.go append clean), a6jpu9 vendor audit (internal/audit/vendor_test.go + the install-level fixture; the .research note is NOT added to the repo because .research/ is untracked on main and the note already lives as a board artifact), 5wfdfx manifest fallback.

5wfdfx NEEDED A REWRITE, NOT A PORT, exactly as its reviewer predicted. The branch patched a function that does not exist on main. I probed main directly and CONFIRMED THE DEFECT IS LIVE THERE: a dangling agent-skill.json symlink makes skillspec.Load return the agents/runtime.json fallback spec with err == nil, and ManifestSourcePath reports agents/runtime.json. Main pathExists already returns non-ENOENT errors, so only the Stat-vs-Lstat half was missing. Fixed by moving pathExists to os.Lstat with a wrapped error, and ported the test battery to main canonical/legacy/runtime names in a new internal/skillspec/fallback_test.go. Red-green captured: with Stat restored, the three dangling-symlink cases fail with the exact fallback spec printed; with Lstat, exit 0.

GATES SO FAR, real exit codes, each run standalone: gofmt -l cmd internal 0 (empty); go build ./... 0; go vet ./... 0; golangci-lint run 0 (one errcheck on the ported test fixed first); no-broad-suppression 0; gate-selftest 0 (75 passed); ledger-consistency 0 (72 rows across linux darwin windows); toolchain-identity 0 (needs PATH ahead of the goenv shim). Full test-gate.sh against the committed SPEC_PIN 00b1688 conformance root still running.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-3e579b, pid=43464, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-ccec5f, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-ccec5f)
RUN-260822-ccec5f resumed the composite. Prior run left 42 files staged and UNCOMMITTED on task/TASK-260823-qwr5w9-credential-scopes-composite at origin/main 6a9b201; nothing was ever committed.

VERIFIED THE ASSEMBLY BEFORE COMMITTING. All three accepted patch artifacts reverse-apply clean against the tree (git apply -R --check exit 0 each: 4p3dcq_final, 27bvo4_symlink-launcher, 2v5e80_remedy), so each is fully present, not partially applied. The three patches touch disjoint file sets, so the admission.go and skip-classes.tsv overlaps named in the task description were already reconciled inside the 4p3dcq_final chain and needed no hand merge. The three hand-ported groups (closure provenance, vendor audit, manifest fallback) are present as source diffs.

SPLIT THE FLAT BLOB INTO SIX LOGICAL COMMITS, one per concern, messages in the repo house style (imperative subject, no board IDs, Spec section references only). Only install_test.go was touched by two groups; split by truncating at the function boundary for the closure commit and restoring the full file for the vendor commit. Post-split diff against origin/main is byte-identical to the pre-split blob: 42 files, 6014 insertions, 30 deletions.

fdbd4c0 manifest symlink fallback / 7193651 vendor inert text / 46ed906 closure provenance / b28376f toolchain remedy / 831023d launcher symlink / 14db4d9 build_ssh credential scopes.

GATES, real exit codes, each run standalone, no pipes: gofmt -l cmd internal 0 (empty); go build ./... 0; go vet ./... 0; golangci-lint run 0 (0 issues); no-broad-suppression 0; gate-selftest 0 (75 passed, 0 failed); ledger-consistency 0 (72 rows across linux darwin windows); toolchain-identity 0 under the CI environment (GOTOOLCHAIN=local GOENV=off PATH ahead of the goenv shim -- it is exit 1 on the bare local shell for the host reasons the gate exists to catch, GOTOOLCHAIN=auto then a per-user go env file, neither a property of this branch). Full test-gate.sh against the SPEC_PIN 00b1688 conformance root is running now.
agent completed: [implementer] developer (claude) (exit=124)
spawn run completed: claude (run=RUN-260822-ccec5f, pid=80781, exit=124)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260823-00b261, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260823-00b261)
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260823-00b261, pid=51109, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260823-16c34e, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260823-16c34e)
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260823-16c34e, pid=78165, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260823-qwr5w9_spawn-log_-implementer--developer--claude-_RUN-260822-3e579b.log](file://TASK-260823-qwr5w9/TASK-260823-qwr5w9_spawn-log_-implementer--developer--claude-_RUN-260822-3e579b.log) — System spawn log captured by task-board
- [TASK-260823-qwr5w9_spawn-log_-implementer--developer--claude-_RUN-260822-ccec5f.log](file://TASK-260823-qwr5w9/TASK-260823-qwr5w9_spawn-log_-implementer--developer--claude-_RUN-260822-ccec5f.log) — System spawn log captured by task-board
- [TASK-260823-qwr5w9_composite-evidence.md](file://TASK-260823-qwr5w9/TASK-260823-qwr5w9_composite-evidence.md) — Composite assembly integrity proof, commit series with signature state, and local gate evidence with real exit codes
- [TASK-260823-qwr5w9_landing.md](file://TASK-260823-qwr5w9/TASK-260823-qwr5w9_landing.md)
- [TASK-260823-qwr5w9_spawn-log_-reviewer--reviewer--claude-_RUN-260823-00b261.log](file://TASK-260823-qwr5w9/TASK-260823-qwr5w9_spawn-log_-reviewer--reviewer--claude-_RUN-260823-00b261.log) — System spawn log captured by task-board
- [TASK-260823-qwr5w9_spawn-log_-reviewer--reviewer--claude-_RUN-260823-16c34e.log](file://TASK-260823-qwr5w9/TASK-260823-qwr5w9_spawn-log_-reviewer--reviewer--claude-_RUN-260823-16c34e.log) — System spawn log captured by task-board
- [TASK-260823-qwr5w9_review-verdict.md](file://TASK-260823-qwr5w9/TASK-260823-qwr5w9_review-verdict.md)

## Created
2026-08-22T20:59:57Z

## Last Update
2026-08-23T01:05:02Z

## Assigned To
[reviewer] reviewer (claude)
