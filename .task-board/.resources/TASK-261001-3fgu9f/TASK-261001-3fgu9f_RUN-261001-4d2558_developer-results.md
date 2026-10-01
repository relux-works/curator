# TASK-261001-3fgu9f — curator-run Muse mapping and fragment v3

Ready for review under the binding reduced-scope decision. Changes remain uncommitted in the assigned Story worktree; no branch operations or commits were performed.

## Implementation

- Mapping: env muse -> system muse, ax provider muse, as explicitly required by the binding 2026-10-01 decision.
- Reader: launch-env-fragment-v3 retains v2 permissions. Muse requires exactly four XDG parents ending in /config, /data, /state and /cache under the same absolute managed parent. HOME and unverified prompt/MCP members are rejected. Fragment.Home returns the parent. V1/v2 remain admitted for existing adapters and reject Muse; four legacy adapters are checked against unchanged single-variable rules.
- Production permission transport recognizes v2 and v3. No permission flag spelling, plugin mode substitution or provider argv reconstruction was added.
- Pin agents-management v0.5.33 and register its real Muse system. Defaults follows the module's documented system-only declaration-owned model authority, without asserting a vendor binding. Missing system/model authority remains a refusal.
- Update README, SPEC, Unreleased CHANGELOG and repository LOGBOOK. Two Claude goldens record the pin's newly owned CLAUDE_CODE_ENABLE_PROMPT_SUGGESTION=false literal.

## Tested bounds

The actual run -> fake Curator binary -> Resolver.Resolve -> defaults -> permission boundary accepts the canned v3 fragment and currently exits 1 with:

    curator-run: permission_mode_unsupported: agentic: system maps no permission-mode bypass flag: muse has no release-pinned permission mapping

Separately the real plan.Build -> vendorplugin.BuildLaunchWithEnvironment path currently refuses exactly:

    plan_refused: spawn plane refused the launch: agentic: system does not support launch mode: muse does not declare interactive

The plan test requires admission when a pinned plugin declares interactive. The run test likewise proceeds past the permission bound when the module supports it, then verifies fake-child argv and XDG/HOME environment on an admitted launch. No successful interactive Muse launch is claimed today. The current four-XDG overlay and unchanged inherited HOME are independently tested at the composition boundary. Exec/serve permission flags remain upstream-owned.

Coverage: 14/14 new invalid-fragment shapes tested, 4/4 legacy adapters compared, 2/2 requested mutants killed. HOME-set mutant failed the composition assertion; v3-rejection mutant failed both Parse and the real run/resolver entry. Saved-source copies were restored byte-for-byte. This measured coverage covers the reduced task scope, not every possible fragment or an unavailable Muse launch path.

## Verification (all commands run personally)

| Command / run | Real exit code | Evidence |
| --- | --- | --- |
| go test ./... -p 1, review candidate | 0 | All 12 packages green; attached log. Existing unchanged packages may use Go's cache. No previously attached board validation was substituted. |
| go build ./..., after changes | 0 | Native compilation. Earlier two build invocations also exited 0. |
| go vet ./..., after changes | 0 | Native lint. Earlier two vet invocations also exited 0. |
| make fmt-check, after changes | 0 | Earlier two invocations also exited 0. |
| git diff --check | 0 | Standalone final check; earlier checks also exited 0. |
| GOOS=windows go vet ./..., review candidate | 1 | Existing POSIX process-control and syscall.Mkfifo compilation errors; attached red log. Earlier two candidate invocations also exited 1. |
| GOOS=windows go vet ./..., unchanged HEAD archive | 1 | Same pre-existing failures reproduced independently on checkout base. Not presented as passing. |
| HOME-set mutant: go test -count=1 ./cmd/curator-run -run '^TestMuseV3CompositionPreservesHOME$' | 1 | Expected failure; HOME changed to managed home. Attached mutant log. |
| V3-rejected mutant: go test -count=1 ./internal/fragment ./cmd/curator-run -run '^TestMuseV3(Fragment\|ThroughRunPermissionBound)$' | 1 | Expected failure; the real resolver emitted resolve_fragment_invalid. Attached mutant log. |

Earlier diagnostics and corrections: the initial targeted fragment/mapping/run command exited 1 twice (first defaults' use of strict ResolveRuntime, then the newly exposed native permission bound), then 0 after aligning with the documented system-only API and testing the actual refusal. The targeted defaults Muse command exited 0. The first full suite exited 1 on the two pin-changed Claude goldens and a stale unregistered-Muse assertion; those were corrected. A targeted plan/defaults run exited 1 from an incorrectly named error symbol in the updated test, corrected to ErrUnknownModel and subsequently green in the full suite. An intermediate full suite was deliberately interrupted to replace the obsolete v1/v2 resolver diagnostic before review validation; it exited 1 and is not passing evidence. The first v3 mutant ran with fixture prevalidation and exited 1; its fixture was strengthened, and the mutant rerun exited 1 through the production resolver. All final code/test inputs are captured by the attached SHA256 candidate manifest.

Windows portability and Muse permission/interactive capabilities remain stated upstream/existing bounds. No Windows-green checklist or successful interactive Muse claim is made. The reduced scope is ready for review.
