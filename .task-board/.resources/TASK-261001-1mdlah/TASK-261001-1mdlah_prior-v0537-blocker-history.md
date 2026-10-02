# TASK-261001-1mdlah — Muse interactive v0.5.37 blocker

The dependency is re-pinned to v0.5.37 (8ee0150030de135946cb33b7389287d242417226), and WIP is uncommitted.

The binding brief requires Muse 1.4.1 to be refused as an unlisted release. Upstream pkg/agentic/systems/muse/policy.go lists BOTH 1.4.1 and 1.4.2. Real run-entry tests with fake binaries confirm that 1.4.1 is admitted for native and yolo. The requested refusal tests exit 1, while 9.9.9 and an absent version are refused in both modes (4/4 applicable refusal cases).

Decision needed: authorize using unlisted 1.4.0 instead (recommended), or supply an upstream tag that excludes 1.4.1. No launcher-owned exception was added. No answer to the clarification arrived in this run.

Independent evidence:
- Fake Muse prints exactly Muse Code 1.4.2 (1.4.2-R4684.1), from a sibling version fixture since upstream filters test-only environment variables.
- Native, yolo, and yolo-alias admission through run pass 3/3; exact argv/env, inherited HOME, four XDG overrides, and stdin bytes asserted.
- Both required mutants killed (2/2): HOME overlay makes all three root rows fail; extra --yolo at plan.Build makes both yolo rows fail and native pass. Each mutant command exits 1.
- Three new Muse goldens; only existing Claude direct/tracked golden change is CLAUDE_CODE_ENABLE_PROMPT_SUGGESTION=false, documented under CHANGELOG Unreleased for curator#102.
- go build ./..., make fmt-check, go vet ./..., git diff --check: each exit 0.
- GOOS=windows go vet ./... at baseline v0.5.22 and candidate v0.5.37: each exit 1, with identical diagnostic line multisets (concurrent package ordering differs). Existing POSIX syscall errors, not a passing Windows gate.
- go test -p 1 ./...: observed exit 1 after interruption when command execution stalled; no package summary or passing suite claim.
- Corrected focused duplicate-yolo rerun: observed exit 1 after interruption; plan-build native/yolo and composition rows passed before the stall. Initial duplicate tests expected exit 1; corrected to established usage exit 2.
- An attempted module-cache overlay exits 1 because Go prohibits it; not counted as a killed mutant.
- Basic process/echo probes and board attachment commands stalled too. Pending processes were interrupted; no command is intentionally left running.

Detailed evidence saved at:
/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/TASK-261001-1mdlah-2y_u7709
Files: TASK-261001-1mdlah_v0537-results.md and TASK-261001-1mdlah_v0537-gates.json, plus gate logs and overlay files.

No LOGBOOK per binding instruction. Review handoff withheld for the release-policy conflict and missing passing full-suite evidence. Board attachment/status attempts must be verified; do not assume they succeeded.

Board persistence outcome: the resource attachment command stalled and was interrupted (exit 130). A subsequent bounded mutation batch requesting resource attachment, blocker notes, and status=blocked also stalled and was interrupted (exit 130). No board persistence or blocked transition is claimed. The last verified board status was development. All pending commands from this run have been interrupted; there is no intentional background validation left. The owner must restore command execution, authorize the corrected refusal release, attach these saved artifacts through task-board, rerun the required checks, and then route the appropriate lifecycle transition.
