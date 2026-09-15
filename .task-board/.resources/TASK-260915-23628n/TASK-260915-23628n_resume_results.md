# Producer review handoff — 2026-09-16

Choice: additive BuildPlanWithEnvironment / BuildLaunchWithEnvironment result APIs keep Plan and legacy JSON unchanged. OwnedEnv is cloned, sorted, nil when empty, and computed with ChildEnv(nil, effective request) after preparation and alias projection. Existing entry points do not request the additional snapshot. Docs and Unreleased v0.5.13 changelog included.

Codex anomaly: captured Darwin ARM64 binary paths were host-dependent. Expected Binary paths now project the running toolchain architecture/platform independently of the production resolver. All captured golden files remain byte-unchanged. A permanent negative test rejects the opposite architecture for both managed-npm and native-shim plans.

Validation personally executed in zsh with pipefail, standalone processes, no tee:
- go test ./pkg/agentic ./pkg/vendorplugin ./pkg/agentic/systems/codex -count=1: exit 0, rerun after restoring both mutants and adding negative test.
- go vet ./pkg/agentic ./pkg/vendorplugin ./pkg/agentic/systems/codex: exit 0.
- git diff --check: exit 0.
- Mutant: ChildEnv snapshot uses original unprepared request. go test ./pkg/agentic -run ^TestBuildPlanWithEnvironmentPreparedAlias$ -count=1: exit 1, expected failure showing original prompt and unresolved alias. Restored from preserved candidate copy.
- Mutant: golden expectation hard-codes wrong-architecture triple. go test ./pkg/agentic/systems/codex -run ^TestPlansMatchTheCodexGoldens$ -count=1: exit 1, expected Binary mismatches in both managed-npm and native-shim. Restored from preserved copy.

Coverage: BuildLaunchWithEnvironment driven for 3 runtimes x 4 modes = 12/12 combinations, with unsupported modes checked for refusal; supported results compared against prepared ChildEnv and legacy plan JSON. BuildPlanWithEnvironment drives preparation/alias, snapshot ownership and snapshot-error tests. Narrowing mutants killed 2/2. No real provider launches. ARM64/other-platform execution unverified; only this amd64 host ran tests.

Full go build/test landing suite reserved for runtime handoff exactly once; no prior evidence substituted for current narrow runs. No commits, tags, installs or LOGBOOK.md edits. Findings recorded here and in board notes under campaign exception. Integration and signed release belong to orchestrator.