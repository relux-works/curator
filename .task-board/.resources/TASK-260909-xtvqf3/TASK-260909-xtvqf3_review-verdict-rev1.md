# TASK-260909-xtvqf3 independent review — CR revision 1

Verdict: **changes_requested**. Route: **to-dev**. No external blocker.

Reviewed CR-TASK-260909-xtvqf3-1, base 3ff66a9421ff6ddf675a49fc0c2868309f6e3de3, candidate ff61be4a8bd43fa4ffb179d31aa38e41891d4313. Repository delta is empty (independent git diff exit 0). That is the correct delivery shape for this artifact-only task: production adoption belongs to the later diagnostics revision. Empty diff is neither the defect nor acceptance evidence.

## R1 — high: delivered runner reports success without running any gate

The attached run-gate.sh expects adjacent `gate_conformance_test.go` and `gate_framing_test.go`; attached resources use task-prefixed, hyphenated names. Adoption instructions say the attached runner can be used directly but do not stage those expected names. Independently ran the unmodified runner with all attachments copied under their published names into task .temp and a fresh disposable destination. Both cp operations failed. Every go test emitted `no tests to run`, all four mutant commands returned 0, and the runner itself returned 0. This is an executable false-success path, not a documentation-only concern.

The runner has only `set -u`, echoes statuses without checking them, masks setup failures, does not require named assertions, and ends with a successful echo. It therefore cannot enforce baseline success, mutant failure, or restoration. Archive extraction is not fail-closed and an existing destination may retain stale files. The first adoption example also omits destination creation, changes into an archive without Git metadata, and invokes a runner name not established by those commands.

Required: make published artifact names and adoption commands agree; require a fresh destination and exact source identity; fail closed on extraction/copy/mutation/test setup/restoration failures; explicitly require each baseline exit 0 and each mutant exit 1 with its named behavioral assertion. A compile error or missing test is not a killed mutant. Add negative runner checks for missing overlays, zero selected tests, baseline failure, and a surviving mutant. Expected negative Go exits must be handled explicitly, not by simply adding `set -e`.

## R2 — medium: positive joined coverage claimed but absent

`gateCheckRejects` covers foreign values in direct/wrapped/joined forms, but iterates owned-family positives only through fmt.Errorf wrapping. `TestGateOwnFamilyAndNilPreserved` has no joined positive case, despite its comment claiming wrapped/joined chains; nil values are only direct/wrapped. The runner selects TestGate, so unrelated pre-existing tests cannot discharge these gate claims.

Independent narrowly scoped mutant: CodeOf rejects a joined chain containing a direct ResolveError with the single owned code resolve_invocation_failed, retaining direct/wrapped acceptance and other code handling. All attached diagnostics TestGate tests passed (exit 0). Independent TestReviewJoinedResolveInvocationAccepted failed with the actual missing classification (exit 1). Diff, probe source, command and outputs are attached.

Required: use the full source-derived owned sets across direct/wrapped/joined positives, include joined typed-nil behavior, and prove a targeted joined-positive regression fails. State the coverage ratio and distinguish the three mutable-Code owners from fixed-code UsageError/axconfig.Error and other call-site-selected codes.

## R3 — small evidence correction: companion framing checks do not run under the mutant

The main framing test calls assertSingleDiagnostic before its companion loop. The exact mutant triggers t.Fatalf there, so the companion loop never executes. The runner then runs diagnostics TestGate, which does not exercise Line. Thus `framing-mutant-diagnostics-gate exit=0 (other framing intact)` does not prove that claim.

The mutant itself IS a valid single-complete-detail equality exception. Independent separate TestReviewFramingCompanions confirmed all 3 companion details remain byte-exact framed under it, while TestGateFramingSingleDetailAtRealResolver fails. Split companions into independently executed tests/subtests and correct the evidence label; retain this valid exact-detail mutant.

## Confirmed strengths and independent results

All seven substantive attachments inspected and SHA-256 hashed; the four hashes published in producer results match. Eight pinned source blobs and extracted bytes match exact diagnostics CR2 tree fbe90d5e60593a3a069721b2ad9e53cd071d8c02. SPEC SHA-256 matches 5a7ccf0ba95708cb573a586977eb0ba4bad1e46233a540dc99784c1d4d922d48. Independent SPEC table parsing confirms 18/18 normative codes, owned families 6/2/2, foreign sets 12/16/16: 44/44 normative rejection pairs and 168/168 cases with four extras and three forms. The current owner lists match normative resolve/mcp/system-prompt families and do not read the diagnostics allow maps.

After reviewer staging of the two expected filenames (test contents unchanged), independently replayed the attached runner:

| Check | Exit | Result |
| --- | ---: | --- |
| Baseline diagnostics TestGate | 0 | Pass |
| Baseline main TestGateFraming | 0 | Pass |
| Resolve admits mcp_layer_missing | 1 | TestGateResolveRejectsForeignCodes fails |
| Layer admits resolve_invocation_failed | 1 | TestGateLayerRejectsForeignCodes fails |
| Refusal admits resolve_invocation_failed | 1 | TestGateRefusalRejectsForeignCodes fails |
| Exact complete-detail framing exception | 1 | TestGateFramingSingleDetailAtRealResolver fails |
| Full baseline diagnostics + main suites, -count=1 | 0 | Pass |
| Independent joined-positive mutant, attached TestGate | 0 | Survives, R2 |
| Independent joined-positive assertion under same mutant | 1 | Actual regression confirmed |

Thus 3/3 supplied owner mutants and 1/1 supplied framing mutant are killed once files are staged correctly, including both prior R3 survivors. These are real behavioral failures, not compile failures. The additional joined-positive mutant is a separate missing acceptance dimension.

Framing reaches the actual production `run -> Resolver.Resolve -> ExecRunner -> diagnostics.Emit -> Line` path (main.go calls resolver at line 91, Emit at line 99), using a nonexistent fixed executable and injected profile. It tests run directly, not an installed executable or a real Curator/model invocation. CodeOf is API-only in this tree and has no non-test production caller; the owner matrix does not establish full-main launch integration.

## Provenance, preservation, and bounds

Reviewer environment: Darwin arm64, Go 1.25.5; tool versions attached. Scratch root: `.temp/TASK-260909-xtvqf3-review/` in assigned Story worktree. All mutations confined to git-archive disposable copies there. Each restored diagnostics.go byte-equal; eight source pins checked against the Git objects. Assigned worktree git status remains empty. Original diagnostics Story workspace was never opened for writing. No production edits, commits, branch changes, installs, tags, hosted CI, real ax/model calls, or private-history changes.

Read producer brief, results, adoption, vectors, Go tests, runner and complete producer gate evidence, plus prior diagnostics review-verdict-rev2. Independently reran all supplied probes and focused full suites; did not rerun make check, unrelated packages, or the producer's old-gate survivor supplement. Those historical results remain producer/prior-review evidence, not new reviewer runs.

Evidence resource: TASK-260909-xtvqf3_review-evidence-rev1.txt contains hashes, exact commands, real exits, mutation diff, runnable reviewer probe sources and logs. Rework remains artifact-only. This verdict does not accept diagnostics CR2, waive subsequent code adoption/review, or resolve main/defaults/plan/Pi obligations. No accept_cr or commit_ack supplied.

## Lifecycle receipt

Both outcome resources were attached before routing. Live checklist updated: executable-delivery/independent-acceptance items remain unchecked, AC and green-gates remain unchecked; architecture fit and actual gate attacks are checked. Review finding appended through set_notes. `set_status(TASK-260909-xtvqf3, status=to-dev)` exited 0; subsequent compact projection confirmed to-dev. Requested `task-board handoff TASK-260909-xtvqf3 --role reviewer` exited 1: `role "reviewer" has no end_status and cannot use handoff`. This extra handoff command is unsupported for the configured role; the explicit changes-requested route succeeded and remains authoritative. No acceptance or done transition attempted.
