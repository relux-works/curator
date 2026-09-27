# TASK-260916-55g9dg review verdict — CR rev5 — ACCEPTED

Reviewer: claude-opus-5-5. Base 97ca3370, candidate tree 445236ca (worktree == candidate: `git diff 445236ca` empty). 8 paths, no CHANGELOG/LOGBOOK, no SPEC_PIN change.
Spec root: own clone of curator-spec at SPEC_PIN 23435129 (v1.0.0-rc.13), CURATOR_CONFORMANCE_ROOT=$TMPDIR/rv55-spec/conformance/v1.

## Rev4 (accepted) → rev5 re-apply fidelity
Line-level diff of rev4 delta (bd3c0f43..refs/campaign/55g9dg-rev4-full-20260927) vs rev5 delta: identical except the gap-ledger / owner-test adaptation to trunk 97ca3370, where `environments.permissions` is now supported. First unsupported field is therefore `require_source_signers` (E1, STORY-260916-ioemse): 5 E2 manager schema rows re-attributed 1cenbr→ioemse, the rev4 overlay-row rewrites dropped (trunk already carries them). Test assertions updated consistently (environments_conformance_test.go:385-390, environments_test.go:694-704). No trunk content reverted.
Ledger: 73 lines before and after; 0 rows owned by STORY-260916-2d9coh (E2 rows removed/none remaining; remaining E2-case rows honestly attributed to the external first blocker E1).

## Spec (rc.13 environments §3/§5.5/§5.7/§12)
- Transitive class: system module refused under `error`: contextmaterialize/admission.go:160-190 (FirstTransitiveSystemModule, resolution/install/update) + contextmaterialize.go:272 (materialization); diag `context_system_module_transitive`. Lock unchanged on refusal: TestUpdateErrorLeavesLockUnchanged. Direct system module installs: vectors system-module-direct / overlay-direct.
- Default drop with `context_system_module_dropped`; posture: status.go:570-662 prints policy + every dropped module; unreadable transitive manifest → unknown (nil) + DiagManifestInvalid, non-current (§8.4). CLI entry: cmd/curator TestEnvStatusReportsDroppedSystemModuleThroughCLI.
- null waivers rejected at Load ("must be a list"); deterministic first-unsupported-field via sorted keys (environments.go:188-193).

## Validation (zsh, pipefail)
- go build ./... rc=0; go vet config/envprofile/cmd/curator rc=0; gofmt -l internal cmd empty.
- go test -count=1 ./internal/config/... ./internal/contextmaterialize/... ./internal/contextresolve/... ./internal/interop/environments/... → all ok, rc=0.
- TestSystemModuleAdmissionVectors: 5 admission cases (direct, transitive-drop, transitive-error, transitive-waived, overlay-direct) PASS, not skipped.
- envprofile -run 'Status|Admission|SystemModule|Policy|SystemPrompt' ok 109.8s; cmd/curator -run 'TestEnvStatusReportsDroppedSystemModuleThroughCLI|SystemModule|Transitive' ok 205.7s.

## Mutants (disposable git-archive copy)
- M1 materialization refusal softened (contextmaterialize.go:272 never errors): KILLED — TestSystemPromptErrorRefusesFirst, TestSystemModuleAdmissionVectors.
- M2 resolution refusal removed (admission.go:165 always returns nil): KILLED — TestInstallErrorRefusesTransitiveSystemModule, TestUpdateErrorLeavesLockUnchanged.
- M3 status unreadable manifest → `continue` (absence): KILLED — TestStatusUnreadableTransitiveManifestDoesNotBecomeNoDrops.
No survivors.

## Bounds / notes
- M1 alone is not caught by envprofile install/update tests (resolution-layer check fires first); the materialization layer is covered by contextmaterialize tests — two-layer defence, each killed by its own tests.
- stateread: the delta adds no new manager-state read; the one LoadManifest line is the pre-existing package-manifest read re-flowed to capture its path.
- E2 invalid manager/system schema cases do not reach the E2 diagnostic at rc.13 until E1 (ioemse) lands; ledger records it.
- Linux/Windows not rerun locally; hosted gate (green per CR) is the arbiter.
