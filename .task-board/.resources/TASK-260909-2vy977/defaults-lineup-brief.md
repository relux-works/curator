# Producer brief — TASK-260909-2vy977 complete-lineup-defaults-and-production-diagnostics (resume on host e11-1, 2026-09-15)

## Base and recovery (do first, verbatim)
- Story workspace: this run's managed worktree under curator-agent-launcher/.temp/STORY-260908-1wxjbs/worktree, forked from launcher main `cb232a120c9a04c56ae5037c82347921f020e688`.
- The accepted, never-landed defaults file family (TASK-260908-25z3wj, CR1, reviewer-accepted) survives ONLY as the outcome resource `TASK-260908-25z3wj_change-request_rev1.patch` (4 paths: .scripts/defaults-mutants.py, internal/defaults/defaults.go, internal/defaults/defaults_test.go, plus one). It applies cleanly at cb232a1 (`git apply --check` exit 0). Apply it first; do not rewrite those files except where this task's scope requires.
- The rejected rev2 candidate (`TASK-260909-2vy977_change-request_rev2.patch`, 14 paths, base 289ff42) is reference material only: it carries the lineup/Pi work. Port from it what the verdict did not reject (native-Pi registration via `agentic.NewRegistry`, per-member resolution, stderr origin line-group, tests). It does not apply at cb232a1 (go.mod already pins v0.5.11 — keep the pin, no replace).
- Read `TASK-260909-2vy977_review-verdict-rev2.md` and `TASK-260909-2vy977_vendor-policy-analysis.md` before coding; both are outcome resources on this task.

## F4 — Pi runtime preference (product decision, recorded)
Cross-vendor `Lineup` over the union of pi-anthropic/pi-openai/pi-google rows is FORBIDDEN (upstream v0.5.11 declares scores vendor-local). Implement analysis alternative A: an explicit ORDERED Pi runtime preference, then that vendor's own `Lineup`.
- Ordered preference, recorded as a pure convention by the orchestrator under the unified delivery goal (operator may override; the value is a single constant plus SPEC table row): `pi-anthropic`, then `pi-openai`, then `pi-google`.
- Level 3 for `pi` → `pi-native`: take the first preferred runtime whose declaration carries at least one driven row; run `Lineup` over THAT runtime's own models only; bind its contributor. If no preferred runtime has a driven row → `defaults_unresolvable` with the exact reason.
- Operator/machine `defaults.json` per-member overrides keep precedence over the preference (unchanged from CR1 semantics); a configured model still binds to the frozen runtime whose vendor carries it.
- SPEC §4.3: add the preference table row and the sentence that vendor scales are never compared; SPEC stays 0.3.0-draft (erratum found by implementation is allowed, recorded in the CHANGELOG/SPEC changelog section).

## F3 — validation and coverage accounting (repeat finding; hard rule)
- Do NOT run the full `make check` manually. The configured landing suite (`make check`) runs exactly once through the task-board runtime at Change Request publication; cite THAT log. Focused package tests are fine and must be reported with real exit codes from a shell with `set -o pipefail` (state the shell).
- Coverage table: rows are the owned AC requirements; a row is "driven" only when a committed test reaches the PRODUCTION call site `cmd/curator-run.run → defaults.Files.Complete → EmitGroup` (or BuildLaunch admission). Helper-direct tests, pin checks and dependency inspection are bounds, not driven rows. State the honest denominator.
- Mutants: keep the attached 34/34 result if still valid on the new tree; otherwise rerun and report; no ceremony reruns.

## Scope boundary
- This leaf owns SPEC §4.3: defaults file family (already accepted), real tagged Lineup fallback with the preference above, per-member origin/effort semantics (row recommendation, no-effort rows, no silent retry), stderr origin line-group, and its production ordering/refusals up to the plan step. Main wiring of the full pipeline (fragment → mapping → defaults → BuildLaunch/limits → prompt → composition → late checks → exec/handoff) belongs to TASK-260908-1o7i8y and must not be claimed here; the interim `not_implemented` line may remain after the group is printed.
- Three environments (claude_code, codex_cli, pi) must resolve defaults positively at the production entry with the real v0.5.11 module (GOWORK=off). No pseudo-version, no replace, no committed workspace.
- No hosted CI calls from the producer, no installs, no runtime-home writes, no tags. Signed commits are made by the integration path, not by you.
