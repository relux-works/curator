# THE ONLY CURRENT INSTRUCTION — N8 rework 1 (TASK-261008-1j34ro, developer)

Revision 1 was reviewed: **changes_requested**. Read `TASK-261008-1j34ro_review-verdict-rev1.md` first; fix both blocking findings and nothing else.

The orchestrator converged the Story workspace onto current trunk d46f2b1f (N7 landed in between). Your revision-1 delta is carried over unchanged; CHANGELOG.md now holds N7's entry and yours — keep both.

## Fix
- **F1 — explicit empty `--allow`.** Detect whether `--allow` was supplied (for example with `flag.Visit` or a presence-tracking flag value), validate the supplied value even when it is empty, and use the same presence decision for the pin branch. `audit --allow "" --reason x` and `audit --allow= --reason x` must return the usage exit before any configuration or filesystem access. Plain `audit` without `--allow` behaves as today.
- **F2 — ordering proof.** Add a production-entry regression through `run` with a configuration source that records `Load` calls: for every refused value (empty in both spellings, traversal, nested path, short digest, non-hex digest) assert the usage exit AND zero `Load` calls. Make `assertNoPinState` return `WalkDir` read errors instead of ignoring them.

## Evidence — hosted only
Never run `go test`, compiled test binaries or `go run` on this host: they are refused and they harm it. Compile-only checks (`go vet`, `go build`, `gofmt`) are fine.
From a disposable `git clone --shared` of the control root, push three scratch branches and wait for CI (push all of them at once so the runs overlap; this run has a 120-minute limit) (`gh run watch <id> --exit-status`):
1. **green:** the exact candidate → CI green;
2. **red:** the candidate with only the production fixes reverted → the new regressions fail (record the failing test names from the test-evidence artifact);
3. **ordering mutant:** the candidate with configuration loading moved ahead of the CLI parse → the no-`Load` assertion fails (record the test name).
Record the three run URLs and outcomes in your results resource, tick the checklist items with that evidence, then delete the scratch branches.

Then `task-board handoff TASK-261008-1j34ro --role developer` and END YOUR TURN. Do not edit LOGBOOK.md.
