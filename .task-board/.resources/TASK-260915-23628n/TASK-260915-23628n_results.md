# TASK-260915-23628n — implementation and validation evidence

## Candidate and API choice

Uncommitted candidate in the assigned Story worktree, based on
`da59d8b1fe0070bfac803fa646da0779c7985511`. Fresh `git ls-remote --symref origin HEAD`
advertised main at that exact OID. No branch operations, commits or tags created.

Selected the brief's result API alternative: `vendorplugin.BuildLaunchWithEnvironment`
returns `BuildLaunchResult` (alias of `agentic.PlanWithEnvironment`), containing
`Plan` and `OwnedEnv`. `agentic.BuildPlanWithEnvironment` captures a copied,
lexically sorted `ChildEnv(nil, effectiveRequest)` after preparation and alias
projection, in the same planning pass. Snapshot errors return a zero result.
Legacy calls retain their dispatch behavior, Plan fields and JSON contract.
This avoids changing existing serialized plans or adding a second BuildPlan.
Package docs and a new CHANGELOG Unreleased v0.5.13 entry document the API.

## Driven coverage

Production entry `BuildLaunchWithEnvironment`: 10/10 supported system/mode
combinations (Claude 4, Codex 4, Pi-native 2) compare the snapshot with the real
system's ChildEnv on the prepared, projected vendor request and compare legacy
Plan JSON byte-for-byte. Pi-native's other 2/2 modes are tested as refusals.
The fake executable files are resolution fixtures; no provider or ax launches.

Production entry `BuildPlanWithEnvironment`: preparation-sensitive synthetic
system proves prompt preparation, alias projection, sorted copy independence,
snapshot failure refusal and preservation of the legacy API. These are contract
bounds, not claims that real Claude/Codex/Pi-native currently derive environment
values from prepared prompt bytes.

Narrowing mutant: capture the input request before preparation and use it for
owned ChildEnv, leaving all other planning and snapshot code present. The named
preparation/alias test failed with original prompt and unprojected alias (exit 1).
Restored from an exact saved copy; cmp exited 0; package tests reran green.

## Validation actually run

Shell: zsh, `set -o pipefail`, standalone commands, no tee. No older attached
validation accepted. All commands below were executed by this producer.

| Command | Exit | Evidence |
| --- | --- | --- |
| `go test ./pkg/agentic ./pkg/vendorplugin -run 'TestBuild(PlanWithEnvironment|LaunchWithEnvironment)' -count=1` | 0 | New API tests |
| `go test ./pkg/agentic -run '^TestBuildPlanWithEnvironmentPreparedAlias$' -count=1` on mutant | 1 | Expected failure, mutant killed |
| `go test ./pkg/agentic ./pkg/vendorplugin -count=1` after restoration | 0 | Both full package suites |
| `go build ./pkg/agentic/... ./pkg/vendorplugin/...` | 0 | Scoped build |
| `go vet ./pkg/agentic/... ./pkg/vendorplugin/...` | 0 | Scoped vet |
| `go test ./pkg/agentic/systems/claude ./pkg/agentic/systems/codex ./pkg/agentic/systems/pinative -count=1` | 1 | Claude and Pi-native passed; two Codex architecture-specific golden mismatches |
| `go test ./pkg/agentic/systems/codex -run '^TestPlansMatchTheCodexGoldens$' -count=1` in archived unchanged HEAD | 1 | Same two mismatches on baseline |
| `go test ./pkg/agentic/systems/codex -skip '^TestPlansMatchTheCodexGoldens$' -count=1` | 0 | Remaining Codex tests; excluded golden test is NOT passing |
| `git diff --check` | 0 | Whitespace validation |

## External validation blocker and orchestrator/logbook finding

Host reports x86_64; Go GOARCH and GOHOSTARCH are amd64. Existing Codex goldens
`exec-managed-npm-path` and `exec-native-shim` demand aarch64-apple-darwin paths;
both candidate and pristine base produce x86_64-apple-darwin paths. Thus the AC
requiring every existing golden green cannot be established on this host.
No existing test or golden was changed. Skipping those tests does not discharge
that AC. The configured full landing suite was not manually run, and handoff
was not invoked because its checklist cannot truthfully be marked all green.

Recommendation / exact external input needed: orchestrator route this unchanged
candidate and existing goldens to the authorized ARM64 validation lane, then
resume producer handoff with that evidence. Alternative: separately authorize a
portable golden-test design, with independent scope/review; do not overwrite
fixtures or spoof this host's architecture to conceal the mismatch. No access
to a remote lane is authorized in this worker assignment.

The implementation remains uncommitted for the managed integration path.
LOGBOOK.md edits are forbidden by the campaign, so this finding is persisted
here and in task notes for the orchestrator's logbook. Cross-platform validation
and independent review remain unverified. No tag or release was created.
