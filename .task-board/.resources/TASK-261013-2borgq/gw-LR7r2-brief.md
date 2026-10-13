# THE ONLY CURRENT INSTRUCTION — TASK-261013-2borgq rework (revision 2): the scope decision you asked for (W-LR7r2, developer)

**Scope decision (orchestrator, 2026-10-13).** This task is expanded for exactly the repair your recovery note names. The amended `gw-LR7-brief.md` (current precondition) now says so in its item "The release-pin test moves with the docs".

**Do this:**
1. Keep your revision-1 documentation delta: six findings, README, SPEC §§10.1–10.5, the four figure sources, CHANGELOG `## 0.3.0 — 2026-10-13` with an empty `## Unreleased` above it, and the install line at `@v0.3.0`.
2. In `cmd/curator-run/main_test.go` `TestReleaseVersionPinned`, update exactly the pinned README and CHANGELOG entries to the released v0.3.0 state. Remove the pin on the "candidate (pending tag)" sentence, or replace it with its released wording. Keep `wantBuild = "0.3.0"` and the spec-version pins unchanged. Make no other code change.
3. Validation: no local go test or build. The CR validation (`scripts/remote-gate.sh`, the hosted gate) is the arbiter. It must go green on all four lanes: Test and Race on ubuntu-latest and macos-latest.
4. Update the gaps file: the test edit, with its exact lines, and the gate result.

**Rules:** as in `gw-LR7-brief.md` and `gw6-decisions.md`. Docs plus this one test edit; never edit mechanically; public hygiene.

**Hand-off:** run `task-board handoff TASK-261013-2borgq --role developer`, then END YOUR TURN.
