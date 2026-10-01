# TASK-261001-3fgu9f — rework 1 results

Addresses TASK-261001-3fgu9f_review-verdict-rev1.md against base d0920353. No commits created; all repository changes remain uncommitted.

## Changes
- Removed the newly created LOGBOOK.md entirely.
- Restored go.mod, go.sum, and both Claude goldens from d0920353 byte for byte. The real agents-management pin is v0.5.22.
- README.md and SPEC.md now name v0.5.22. CHANGELOG.md retains the Muse behavior entry and removes the pin/prompt-suggestion sentences.
- All Muse production code, fixtures, tests, and refusal bounds remain unchanged by this rework.

## Validation run by this developer
| Command | Real exit | Result |
| --- | --- | --- |
| go test -p 1 ./... | 0 | All 12 packages ok at v0.5.22 |
| go build ./... | 0 | Native build passes |
| go vet ./... | 0 | Native vet passes |
| git diff --check | 0 | Whitespace check passes |
| git diff d0920353 -- go.mod go.sum '*.golden' | 0 | Empty output; base bytes restored |
| GOOS=windows go vet ./... (candidate) | 1 | Existing POSIX syscall/process-control compilation failures |
| GOOS=windows go vet ./... (d0920353 scratch baseline) | 1 | Same existing failures |

Windows output was independently reproduced on both trees in this run. Sorted diagnostics with line/column positions normalized compare equal (20 output lines on each). Windows vet remains a failing gate; it is not reported green.

## Named regression and narrowing mutant
Existing regression: TestProductionPipelineGoldens/claude_code/tracked=false, exercising production run -> plan.Build -> composition -> execution with fake binaries and the base golden.
Narrowing mutant: in a scratch candidate, restore only the v0.5.33 dependency pin and its go.sum, keeping base goldens and production code intact.
Command: go test -p 1 ./cmd/curator-run -run '^TestProductionPipelineGoldens$/^claude_code$/^tracked=false$' -count=1
Real exit: 1. The additional prompt-suggestion environment literal causes the expected golden mismatch. This proves the rejected dependency side effect is caught; no new test or product behavior was needed.

## Muse mutants independently rerun
- HOME overlay only for Muse in composition.Compose: go test -p 1 ./cmd/curator-run -run '^TestMuseV3CompositionPreservesHOME$' -count=1; exit 1, HOME replaced.
- Drop only v3 from the fragment revision enum while keeping v1/v2 admitted: go test -p 1 ./internal/fragment ./cmd/curator-run -run 'V3|Muse' -count=1; exit 1. Reader tests and production run tests fail on resolve_fragment_invalid.

All three applied mutants were killed: 3/3. Mutants ran in a scratch copy and were restored there; the task worktree was never mutated by them. The full green suite includes the named regression and all Muse tests. Earlier reviewer evidence was read for context; every validation and mutant listed above was rerun here.

## Findings preserved from the removed logbook
- Muse maps to launcher system muse and ax provider muse under the binding task decision.
- The v3 fragment carries four XDG parents under one managed parent; HOME remains inherited. Muse prompt and MCP members are refused; v1/v2 behavior remains unchanged.
- v0.5.22 already contains the Muse plugin and declaration-owned model authority needed for this scope; no pin bump is necessary. Defaults refuses missing system/model authority and does not assert a resolved vendor.
- The current run permission bound is: curator-run: permission_mode_unsupported: agentic: system maps no permission-mode bypass flag: muse has no release-pinned permission mapping.
- The separate real plan.Build interactive bound is: plan_refused: spawn plane refused the launch: agentic: system does not support launch mode: muse does not declare interactive.
- The interactive test switches to admission when a pinned plugin declares interactive support. No exec/serve mode substitution or argv reconstruction was introduced.
- Measured bound: today run refuses before composition. The HOME mutant is therefore killed by the composition test, not by the run path. The v3 mutant does reach the actual run resolver.
- Windows portability remains outside this rework, with identical baseline/candidate failures.

Ready for review. Raw outputs are attached as TASK-261001-3fgu9f_rework-1-validation.md.
