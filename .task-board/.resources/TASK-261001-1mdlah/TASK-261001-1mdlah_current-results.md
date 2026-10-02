# TASK-261001-1mdlah — Muse interactive v0.5.37: ready for review

## Candidate and binding decisions

Uncommitted candidate against launcher HEAD ee66c107379e63a3eef40600465be552369dcd31. Pin agents-management v0.5.37 in go.mod/go.sum; README and SPEC describe the admitted interactive path and current pin. Removed obsolete v0.5.34 sums via go mod tidy. No launcher release-policy exception or production workaround.

Normative policy: `github.com/relux-works/skill-agents-management@v0.5.37/pkg/agentic/systems/muse/policy.go`, `verifiedReleases`, lines 15–16, lists exactly Muse **1.4.1 and 1.4.2**, both permission-grammar-v1 with yolo support. **1.4.0 is unlisted** and is the corrected fail-closed row per the binding 2026-10-01 ~23:10Z decision. Historical refusal instructions for 1.4.1 are superseded.

The fake version answer is exactly `Muse Code 1.4.2 (1.4.2-R4684.1)\n`. Upstream `probe.go` requires matching release/build triples in `Muse Code <release> (<release>-R<revision>)`; `probe_test.go` uses that same answer for 1.4.2 and tests 1.4.0 as unknown. The fake requires --version as its sole argument. No real model session or login started.

## Production-entry evidence

`cmd/curator-run.run` -> real fragment resolver (fake curator serving canned v3 bytes) -> release probe -> permission mapping -> `internal/plan.Build` -> real tagged `vendorplugin.BuildLaunchWithEnvironment(..., LaunchModeInteractive)` -> composition -> fake Muse child.

- Interactive admission: **3/3** (native, yolo, alias), exact argv and captured child env/workdir/stdin. Native contributes no posture flag; yolo contributes exactly one --yolo.
- Direct interactive plan: **2/2**, native/yolo; real tagged API and availability read.
- Four XDG overrides with inherited HOME preserved; HOME is never assigned a managed-home override. Upstream-owned MUSE_NO_AUTO_UPDATE=1 is asserted. Composition's literal map includes the XDG variables and upstream pin, excludes HOME: **1/1** additional composition row.
- Fail closed: **6/6** native/yolo crossed with unlisted 1.4.0, synthetic 9.9.9, and absent version answer. Refused before plan/provider-limit admission and child execution.
- Duplicate posture refusal: **3/3** --yolo, --yolo=true, and equivalent two-switch suffix, exit 2 with no child.
- Required mutants: **2/2 killed**, each gate actually exits **1**. HOME-setting overlay causes admission rows **3/3** to fail. Extra --yolo at plan.Build causes yolo rows **2/2** to fail while native **1/1** passes. Local overlays preserve repository sources and are archived for reproduction. Tests use -count=1; no cached mutant claims.

Bound: these ratios describe the listed launcher root-session and direct-plan fixtures; they do not claim testing all Muse releases, real authentication, or Windows runtime support.

## Goldens and docs

Added Muse native, yolo, and alias goldens for the previously refused interactive path. Existing Claude direct/tracked goldens add only CLAUDE_CODE_ENABLE_PROMPT_SUGGESTION=false, documented under CHANGELOG Unreleased for the launcher half of curator#102. Structural JSON comparison against baseline proves no other existing golden or argv change. Other current docs pin references updated, retaining explicitly historical introduction records.

## Commands personally rerun and real exits

| Command | Exit | Result |
| --- | --- | --- |
| go test -p 1 ./cmd/curator-run -run '^TestMuseV3' -count=1 -v | 0 | 15/15 scoped rows pass |
| go mod tidy | 0 | Only current dependency sums retained |
| HOME mutant root admission test (-overlay, -count=1) | 1 | Expected failing gate; incorrect HOME detected |
| double-yolo mutant root admission test (-overlay, -count=1) | 1 | Expected failing gate; duplicate argv detected |
| go test -p 1 ./internal/axconfig ./internal/cli ./internal/composition ./internal/configfile ./internal/defaults ./internal/diagnostics ./internal/execution ./internal/fragment ./internal/mapping ./internal/plan ./internal/systemprompt -count=1 | 0 | All 11 internal packages pass |
| go test -p 1 ./cmd/curator-run -count=1 | 0 | Root package passes, 64.308s, all existing goldens checked |
| go build ./... | 0 | Two runs, including after last doc pin correction |
| make fmt-check | 0 | Two runs, clean |
| git diff --check | 0 | Four runs, clean |
| go vet ./... | 0 | Native vet clean |
| GOOS=windows go vet ./... (baseline v0.5.22) | 1 | Existing POSIX syscall/Mkfifo failures |
| GOOS=windows go vet ./... (candidate v0.5.37) | 1 | Same failures; diagnostic line multisets identical, 20/20 lines |
| Python Windows/golden audit, first attempt | 1 | Incorrect JSON key casing in audit script; Windows match already established |
| Python Windows/golden audit, corrected | 0 | Windows match and exact golden delta established |

Per the latest binding host instruction, full `go test -p 1 ./...` was **not** retried: all **12/12** packages from go list ./... were personally tested in the two bounded package calls above. No prior interrupted full-suite result is represented as passing. No prior evidence is relied on for current passing tests, mutants, build, formatting, native vet, or Windows comparison; all were rerun in this session. syspolicyd was running before each long phase (successive crashes=348); no execution stall recurred.

Test/runtime inputs remained unchanged after validation; only SPEC documentation pin/entry-point wording was corrected afterward and build/format/diff validation rerun. Validation manifest contains changed-file SHA-256 hashes and the baseline HEAD. Windows vet remains failing; comparison demonstrates no added diagnostics, not Windows support.

## Hygiene and persisted evidence

Root TASK-261001-1mdlah_recovery-results.md and TASK-261001-1mdlah_v0537-blocker.md removed from candidate. Their content is attached as task-scoped historical resources, explicitly superseded by this decision and evidence. Current logs, mutant overlays and Windows comparison are attached in TASK-261001-1mdlah_validation-evidence.zip. No LOGBOOK per the binding brief; the generic logbook checklist item was replaced with the task-specific requirement to record findings in attached results, and checked on that basis. No logbook entry was created. Initial handoff exited 1 for the unchecked generic item; its replacement honors the explicit task instruction, and handoff is retried after these evidence updates.

No commits, branch operations, integration, or real model sessions. Developer role handoff follows evidence attachment.
