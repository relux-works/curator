# Review verdict — TASK-260908-qblycn, CR-TASK-260908-qblycn-1 revision 1

Verdict: **accepted**. repeat-of: none.
Reviewer run: tracked reviewer run for STORY-260908-3mcpz5. Date 2026-09-08.
Reviewed subject: findings `TASK-260908-qblycn_a0-verification-findings.md`, bundle `TASK-260908-qblycn_a0-evidence.tar.gz`, worktree HEAD `484933b`, agents-management tag `v0.5.10` (`12f443d1`, good ECDSA signature), installed curator `v0.14.1-0.20260907213730-04550e282705`.

## Why an empty repository delta is the right outcome

Base `484933b` and candidate tree `d1f46b8d…` are byte-identical (`git diff --stat` empty). The brief forbids launcher production code and any edit to SPEC.md before the parent creates erratum stories; the deliverable is bounded evidence plus a contract matrix, delivered as three board outcome resources. No repository file was the correct output for this leaf.

## Independently reproduced (verified by the reviewer)

- Fragments: ran `curator env resolve <env> --profile default --format json` read-only on the installed binary for all three envs. Exit 0 each, absolute paths present, sha256 over stdout minus LF: claude `c7d17c397c0618cb715d7651b93df2756b1bce96a1a442c541a1bb013171545e`, codex `0445923769e5c8d3776ddfe2f142ec3e3d1dd563324b604594ecd64f686e2e32`, pi `c0512f558f8dc93780288db597a1a8250c0403a175c95fe988e0e8c0991701bf`. These match the producer's stated prefixes. stderr carries only `warning: environment_tool_version_unverified` for claude/codex.
- Read-only refusal transport (`03-resolve-readonly.log`): `curator: profile_unknown: …` and `curator: environment_home_stale: …` on stderr, exit 1. Confirms E6.
- agents-management at `v0.5.10` from the tag object: `pkg/agentic/systems/pi/binary.go` `executableName = "agents-infra"` (E1); `frozenRuntimes` declares claude, codex, qwen, gemini, agy, muse and no `System: "pi"` (E2); `vendorplugin.BuildLaunch(ctx, r, SpawnRequest{Runtime, Model, Effort, WorkDir, Home, Env…}, mode)` exists at `spawn.go:134` and `agentic.BuildPlan` refuses `ErrEffortMissing` only against `EffortSupportRequired` (E3); `refuseNonInteractiveParameters` fires in `LaunchModeInteractive` before grammar checks, yielding `ErrCompositionNotInteractive` / `ErrParameterNotInteractive`; `Lineup`/`LineupOf` at `lineup.go:52/96`, `Effort.Recommended` present.
- Provider limits (`providerlimits/verdict.go:114`): absent group record → `absentIsHealthy` (deliberate fail-open); indeterminate read → `AvailabilityUnknown` with `Failures`; blank runtime / no home rule → error, not a verdict. SPEC §4.4 distinguishes "checked and found nothing", "nobody looked", "read failed" and treats non-healthy as `plan_provider_limited`. **No conflation of unobserved health with absence in SPEC.** A1 note: with no state file at all the module answers Healthy by design, which is the "checked and found nothing" arm.
- Plan `Env` is a replacement environment, not a delta (claude clears `CLAUDECODE`; codex filters its family). E4 stands.
- pi 0.84.2 installed source `dist/core/resource-loader.js:380,386,808-829`: `systemPromptSource ?? discoverSystemPromptFile()`, `if (!appendSources) discoverAppend…`, project `.pi/SYSTEM.md` wins when trusted. E5 stands. `dist/cli/args.js:223` pushes every non-dash operand to `messages`; unknown `--flag` consumes the next non-dash arg.
- agents-infra: findings read source at `dee5403`; installed binary is `ab60e0d`. `pi_launch_posix.go` is identical between the two SHAs, so the strip-and-reset claim (lines 137/332) holds for the installed release. Provenance gap noted, resolved.
- codex 0.153.4 probes (`07-…log`): P1 missing layer exit 0, P3 layer applied exit 0, P6 duplicate `-p` exit 2, P8 directory exit 1. These prove layer consumption on the `codex mcp` path only; interactive runtime consumption is not reproducible non-interactively and the producer labels it so. `--strict-config` claim correctly downgraded to docs-confidence.
- claude 2.1.263 probes prove parser acceptance only, and the findings say "recognized", not "consumed". `--effort bogus` warns and exits 0 (verified from log).
- No secrets in the bundle (grep for key/token/bearer patterns hit only help text). No `/Users/` paths persisted.
- State: worktree tree identical to base; control root has no `LOGBOOK.md` and no untracked foreign files; `~/.claude`, `~/.codex`, `~/.agents` links shown unchanged in `06b`; installed-curator backup present at `.temp/TASK-260908-qblycn/backup/curator.v0.14.0-rc.3-113-g66e34a2` sha256 `725068f1fb2d4e7b…`; `ax` absent, no tag/release/PR touched. Installed-binary update was parent-authorized (findings §1.1) and hosted CI on `04550e28` is attached as `curator-main-ci-01.json`.

## Findings

F1 (erratum candidate, reviewer-added, SPEC §4.6 `argv_suffix`): the row says "the composed argv without its element 0" while §4.5 defines the composed argv as plan `Argv` ⊕ channel flags ⊕ native tail with `Binary` held separately. Applied literally it drops the first real argument (`--model` for claude/codex, `pi` for the pi plan). The row's own parenthetical shows the intent is the composed argv verbatim. Minimal correction: "`argv_suffix`: the §4.5 composed argv, verbatim; `Binary` is never a member." Not observed at runtime (no implementation yet); a text contradiction, evidence SPEC.md:544 vs SPEC.md:378-392. Parent should fold into the erratum story set (E1–E6 + this).

F2 (provenance, non-blocking): the bundle holds only `~`-sanitized fragment copies; no absolute-path copy and no digest values are in the bundle although the findings point to "06b". Mitigated by the reviewer's reproduction above. Future bundles should carry the raw stdout files and a digest file.

F3 (provenance, non-blocking): `07b-codex-layer-probe.log` P9–P13 have empty `exit=` fields. Those rows are informative or docs-confidence only, no verified claim rests on them.

F4 (bound, no action): the `cmd/curator` suite beyond the `Env|Profile|Umbrella|Resolve` mask was not run locally (exit 124 on the combined run); hosted CI success on the same SHA is the stated authority. Stated honestly by the producer.

## AC coverage

| AC | Evidence | Status |
|---|---|---|
| Real resolve output for claude_code, codex_cli, pi | §2.3/§7 + reviewer reproduction | met |
| Tagged API verified | §3 + reviewer tag-object reads | met |
| Native argv boundaries verified | §4 + logs 07/08 + pi source | met, bounds labeled |
| Discrepancies explicitly marked | E1–E6 | met; F1 added by review |

Errata E1–E6 are confirmed as evidence-backed SPEC errata, not implementation faults. E1/E2 are module-level (agents-management) design gaps requiring a human product decision on pi routing; the producer's options (a)/(b)/(c) are the right framing and belong to the parent's erratum stories, not to this leaf.
