## Status
backlog

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(5))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [ ] FU-1: profile_source_path_missing and profile_source_path_unreadable diagnostics exist and are tested so that neither fires on the other case (absence vs unreadable rule of environments 1.1)
- [ ] FU-2: loadMachinePolicy distinguishes absent config from unreadable config per the absence-vs-unreadable rule, or the bound is stated in code and docs with a test
- [ ] FU-3: machine-scope switch clears scope records equal to the effective machine default in the same journaled publish (section 9.3 invariant), covered by a narrowing mutant; FU-4: F16 fixtures scope two adapters

## Notes
Cycle-6 reviewer follow-ups (curator head 834b40f6), reproductions in TASK-260905-30zs8t_review-findings-stage-a-6.md:

FU-3 — envprofile/switch.go useLocked machine branch (~:206-214). Section 9.3 says "a scope record equal to the machine default is never kept", but a machine-scope switch onto a profile an existing scope record already names keeps that record. Since the F16 fix makes the machine pass honour scope records, the retained record then pins that adapter permanently: machine=alpha + env:codex_cli=beta, then profile use beta (record kept, equals the default), then profile use gamma -> codex stays beta forever. State stays self-consistent and profile list reports it, recovery is profile use --clear --env, and the spec sentence is ambiguous about whether it binds this third path — author call. If settled as an invariant, fold SetScoped(home, scope, "", true) for every record equal to effective into the same journaled publish, and cover it with a narrowing mutant that clears only the first such record. Repro: .temp/review-6/c4.sh in the curator worktree.

FU-4 — internal/envprofile/envprofile_f16_test.go and cmd/curator/profile_test.go: every F16 fixture scopes exactly one adapter (codex_cli), so a narrowing mutant that skips at most one scoped adapter survives the whole suite (envprofile ok 27.181s, cmd/curator TestProfile ok 8.492s). The mutant is live, not equivalent — built and driven, it reproduces the F16 defect. Shipped code is correct (four scoped adapters verified through the CLI). Fix: scope a second adapter in TestMachineUseSkipsScopedAdapter and assert both are absent from the results.

FU-5 — a machine-scope switch in which every registered adapter carries a scope record materializes nothing, prints no stdout and exits 0, so section 9.2 per-adapter reporting degenerates to silence exactly where the operator needs to be told the switch touched no home.

Carried from cycle 5: FU-1 (section 1.1 profile_source_path_missing / profile_source_path_unreadable diagnostic names do not exist; both shapes fail closed as profile_source_invalid) and FU-2 (a partial install whose activation fails before materializeScope prints no installed-profile line, cmd/curator/profile.go:78; the success path is unaffected).

## Precondition Resources
- [campaign-producer-rules.md](file://TASK-260906-1f2ng0/campaign-producer-rules.md) — Campaign producer/reviewer rules for host e11-1

## Outcome Resources
(none)

## Created
2026-09-06T03:40:31Z

## Last Update
2026-09-15T16:46:36Z
