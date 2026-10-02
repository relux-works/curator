# TASK-261001-1mdlah — pin-agents-management-v0534-muse-interactive

Status: BLOCKED, no review handoff. Work remains uncommitted in the assigned Story worktree.

## Concrete upstream constraint

The pinned Muse System in skill-agents-management v0.5.34 does NOT implement
agentic.ToolReleaseProber. Its exported method set is attached in muse-method-set.txt.
agentic.ProbeToolRelease returns ErrToolReleaseUndetected for this System without
starting the fake muse --version command. The launcher production call site is
cmd/curator-run/main.go resolvePermission (ProbeToolRelease, then PermissionMapping).
Muse PermissionMapping in upstream policy.go requires a verified release row for
both native and yolo and wraps the missing release with ErrPermissionModeUnsupported.
The launcher therefore returns permission_mode_unsupported before plan.Build.

The actual run entry point is driven with a canned v3 fragment through a fake Curator
and a compiled fake Muse binary; PATH contains only the fixture directory. No real
Muse/Claude/Codex session or login was started. Native, yolo and the --yolo alias each
return exit 1 with builds=0 and verdicts=0. go-test.log contains the observed failures.
The fake Muse advertises 1.4.1, but the module never probes it. Supplying
ToolRelease=1.4.1 directly to plan.Build admits an interactive plan, which proves the
lower API supports that request; it does not prove the root launcher can do so.

## Stopped attempts and viable options

The dependency bump alone was attempted, followed by strict root-run admission tests.
Those fail at missing release detection. No fallback release, launcher-specific
probe, fake prober injection or exec-mode substitution was added. Such a workaround
would cross the module-owned release-probing boundary and hide the invalid assumption
that v0.5.34 already contains every capability the launcher needs.

Recommended: upstream adds Muse ToolReleaseProber and its fake-binary tests, publishes
an approved tag, and this task is revised to pin that tag. Then rerun the strict
root rows, generate Muse goldens and run root-driven mutants.
Alternative: explicitly authorize a launcher-owned Muse release probe. That changes
ownership and increases launcher platform/CLI knowledge; it needs an architecture
decision and is not authorized by this pin-only task.
Required external input: an approved upstream release with Muse probing, or explicit
authorization for an upstream follow-up and revision of the required v0.5.34 pin.

## Preserved candidate work

- go.mod/go.sum pin v0.5.34; go mod tidy removes old v0.5.22 checksums.
- README and SPEC name the current pin and accurately document the missing probe.
- CHANGELOG changes are under Unreleased; no LOGBOOK was written per the current brief.
- Muse conditional refusal/admission tests replaced with strict admission expectations,
  exact interactive argv/env and negative unlisted-release/duplicate-posture rows.
  These are WIP acceptance tests, intentionally red for the concrete blocker.
- Direct API native/yolo tests require exact argv, parent env and OwnedEnv. Composition
  verifies the four XDG overrides and absence of an owned HOME literal while retaining
  the inherited HOME.
- Claude direct/tracked goldens each add only CLAUDE_CODE_ENABLE_PROMPT_SUGGESTION=false,
  delivering the launcher half of curator#102. No other existing golden/argv changes.
- Interactive capability declaration test now includes Muse.

## Measured coverage and blind spots

| Surface | Measured result | Bound |
| --- | --- | --- |
| Root interactive native/yolo/alias admission | 0/3 admitted | Missing upstream prober |
| Explicit-release plan.Build native/yolo | 2/2 passed | Caller already supplies 1.4.1 |
| XDG composition with inherited HOME | 1/1 passed | Composition boundary, not admitted root launch |
| Boundary mutants HOME-set/double-yolo | 2/2 killed, each real exit 1 | Direct composition/plan API only; root mutants unproven |
| Root duplicate-posture policy rows | 0/3 reach target guard | Earlier release refusal; tests fail |
| Root unlisted/absent-release refusal rows | 6/6 refuse, 0/6 establish fake release | Per-release root policy coverage unknown; no admission inference |
| Muse root goldens | 0/3 generated | Blocked before child; golden calls retained in WIP tests |
| Existing Claude golden changes | 2/2 reviewed | Only prompt-suggestion env addition |

## Validation commands and real exits

All commands were run personally in this session; no previously attached validation
was accepted. Commands ran as standalone processes, without tee/pipeline masking.
The full test suite was run after mutant source bytes were restored.

| Command | Real exit | Evidence / result |
| --- | --- | --- |
| go get github.com/relux-works/skill-agents-management@v0.5.34 | 0 | Pin upgraded |
| go mod tidy | 0 | Clean checksums |
| UPDATE_PIPELINE_GOLDENS=1 go test ./cmd/curator-run -run '^TestMuseV3' -count=1, first attempt | 1 | Initial test compilation typo OwnEnv; corrected to OwnedEnv |
| Same Muse update command, second attempt | 1 | Missing upstream release probe; corrected direct-plan env order afterward |
| UPDATE_PIPELINE_GOLDENS=1 go test ./cmd/curator-run -run '^(TestProductionPipelineGoldens\|TestProductionAliasEquivalence)$' -count=1 | 0 | claude-golden-update.log |
| go test ./cmd/curator-run -run '^(TestMuseV3PlanBuildInteractive\|TestMuseV3CompositionPreservesHOME\|TestProductionPipelineGoldens\|TestProductionAliasEquivalence)$' -count=1 | 0 | scoped-green.log |
| python3 /tmp/TASK-261001-1mdlah-evidence/muse-boundary-mutants.py /tmp/TASK-261001-1mdlah-evidence | 0 | Both expected-red mutants killed; source restored |
| go test ./cmd/curator-run -count=1 -v -run '^TestMuseV3CompositionPreservesHOME$' (HOME-set mutant) | 1 | HOME-set.log; expected failure: managed HOME override detected |
| go test ./cmd/curator-run -count=1 -v -run '^TestMuseV3PlanBuildInteractive$/^yolo$' (double-yolo mutant) | 1 | double-yolo.log; expected failure: exact argv detects second flag |
| go test -p 1 ./... | 1 | go-test.log; expected blocker failure in 6 Muse root admission/duplicate subtests; 12 internal packages pass |
| GOOS=windows go vet ./... before bump | 1 | windows-vet-baseline.log; existing POSIX execution/FIFO portability failures |
| GOOS=windows go vet ./... after bump | 1 | windows-vet-after.log; same failures, not passing |
| Python exact-line Windows comparison after sorting | 0 | windows-comparison.txt; package block order differs, diagnostic lines identical |
| go build ./... | 0 | build.log |
| make fmt-check | 0 | fmt-check.log |
| go vet ./... | 0 | vet.log |
| git diff --check | 0 | diff-check.log |
| go doc github.com/relux-works/skill-agents-management/pkg/agentic/systems/muse.System | 0 | muse-method-set.txt |

Windows baseline stderr was saved from the original two tool-output chunks; baseline
exit 1 was observed directly before dependency changes. It matches the after output
line for line after sorting. The Windows gate remains failing and is not checked as green.
The important finding is attached on the board; LOGBOOK is intentionally omitted by
the task's overriding current instruction.
