## Status
to-review

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(3))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] snapshot-identity-verifier: Implement closed manifest/descriptor codecs and full actual-tree verification. CCJ-1 plus exactly one LF frames both metadata files; source fragment hash exclud
- [x] snapshot-identity-verifier: Verify the complete inventory, bytes, sizes, semantic modes, path/component collisions, channel references, adapter/profile equality, unknown versions/members a
- [x] snapshot-identity-verifier: Keep exact plan identity and portable comparison distinct. Portable comparison excludes only `fragment_sha256`; producer version, profile pin, descriptor and ev
- [x] snapshot-identity-verifier: Reuse the fixed separate schema/corpus; do not extend released fragments, introduce archive encoding, or treat `manifest-only` as verification of an existing di
- [x] Every row of the leaf's paired test table is implemented: the negative case asserts the intended refusal and unchanged forbidden state beside its positive control
- [x] The leaf's bounded checker or plant control is executed and shown to fail on the planted violation
- [x] Only the leaf's declared paths and seams change; anything outside is listed and justified in the handoff
- [x] Focused tests and the repository's full gate pass; exact commands and results are in the handoff
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] In a managed Story worktree the candidate is left UNCOMMITTED in the worktree for the handoff to snapshot — never commit on the Story branch. A producer commit moves the branch tip off the recorded checkpoint and the handoff refuses with change_request_candidate_committed_past_checkpoint; repair with `git reset --soft <checkpoint_oid>` before completing again.
- [x] Every command, message, state, or refusal named in the AC is driven through the production entry point by a named committed test, or is declared a stated bound. Report coverage as a ratio — `n of m AC rows driven` — and name the production call site for each. Prose in place of the ratio is not evidence.
- [x] Gating, refusing, validating, authorizing, or attesting behavior covered by negative tests that fail when the gate admits what it must reject, with the production call site named
- [x] Every gate ships at least one NARROWING mutant — the gate stays present and is weakened to admit exactly one member of the class it must reject, and a named test must fail. A delete-only mutant proves only that the gate exists and is not accepted as evidence.
- [x] A gate that inspects source text is additionally attacked by a mutant that PRESERVES the searched-for token and changes behavior, and the mutant harness executes the behavioral suite, not only the static checker.
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant

## Notes
spawn selection rationale tuple: {"role":"developer","pair":"gpt-6.1-sol/high","text":"Wave 7 SNAPID producer: docs-only, codex gpt-6.1-sol high (R80), owner go 2026-10-13"}
spawn selection rationale for gpt-6.1-sol/high: Wave 7 SNAPID producer: docs-only, codex gpt-6.1-sol high (R80), owner go 2026-10-13
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[codex,claude], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (codex) (run=RUN-261016-03b791, max_parallel=4)
spawn run started: [implementer] developer (codex) (run=RUN-261016-03b791)
Development baseline: clean Story checkpoint 0d00a109ad1da77a5506747c89f24af9442eb341 equals freshly fetched origin/main. Normative spec main d2a942cf17e4eb017c45475442da1456d8dfbe6c read: managed-home-snapshots sections 2/4/6/7 and CIP-0014. No leaf/spec conflict found. Proposed reusable API: envsnapshot.DecodeManifest/DecodeDescriptor, EncodeManifest/EncodeDescriptor, Verify(root, expected manifest digest), PortableEqual(left,right). Metadata decoding alone never attests a tree. Capture/publish/CLI and hosted wiring belong to consuming leaves. Readiness: task-board, git, gh, Go 1.25.5 and rg available. Brief gap: no catalog surface table supplied; paired-test table will be mapped explicitly in handoff.
Reusable envsnapshot implementation and pinned separate corpus are in the uncommitted candidate. Focused tests, race tests, build and package lint pass. Ten narrowing controls were killed by named behavioral tests (M5 modes/extras and M6 saved-manifest trust included). Final identity reruns are underway. Full repository gate is running on detached signed snapshot 27b492711660252e78f97c2f69d5dcb67c78ada5; Story HEAD remains checkpoint 0d00a109ad1da77a5506747c89f24af9442eb341. The existing gate creates unsigned snapshots with synthetic identity; this invocation uses a temporary Git adapter solely to preserve configured identity, sign and verify the detached gate object. No tracked gate script change. Explicit bounds: no CLI/capture/publish/hosted integration in this helper leaf; Windows actual-tree backend refuses rather than claim ACL/mode equivalence; codecs are cross-platform. Final handoff will distinguish those bounds from executed evidence.
Current frozen candidate is signed snapshot 37a4424677a7e8253641fa48e9962a276f4da3e0. Hosted gate: https://github.com/relux-works/curator/actions/runs/38834659070 (pending; no green claim). Prior run was cancelled after typed UTF-8 self-review; its gate process exited 137 and is superseded. Current race suite passes with 86.9% statement coverage; 58 structural cases, 19 semantic vectors and 11/11 killed narrowing controls. Lint, vet, build and Windows cross-compilation pass. Handoff, candidate hashes and local verification logs are attached under the task-ID prefix. Eighteen evidenced checklist items checked; full-gate item 8 deliberately remains unchecked until its real verdict. Logbook findings and platform/integration bounds are recorded in the handoff outcome, with no control-root edits.
Repository gate on candidate 37a4424677a7e8253641fa48e9962a276f4da3e0 exposed a real test-integration failure: the platform-case gate rejected the default opt-in skip from TestNarrowingControls. Corrected within leaf paths by selecting controls with an explicit snapshotcontrols Go build tag, removing the default skipped test registration; no gate or skip whitelist was weakened. Fresh race tests and all 11 narrowing controls pass, and tagged-package lint reports 0 issues. A focused production platform-case gate run passes (exit 0); planting the former unrecognised TestNarrowingControls skip fails that same gate with exit 1 and the intended named diagnostic. Full repository gate has been relaunched on the corrected frozen candidate. Earlier gate failure is not green evidence.
Current candidate 5fd9bed1663a16a4914cc341949e69eb58a2b7b0 remains uncommitted against the Story checkpoint. Previous full gate 38836557736 failed with shell exit 1: Windows TestCorpusPins detected EOL conversion of pinned schema README bytes. Added leaf-local schemas/** -text attributes, a real autocrlf Git checkout regression with an unprotected CRLF witness, and narrowing M11 that enables conversion only for Markdown. Focused race, checkout regression, lint, vet, build and Windows compile pass; 12 of 12 narrowing controls killed. Updated all three task-scoped outcome resources. Full gate 38841469667 is pending; checklist item 8 remains unchecked. Windows actual-tree backend remains a declared fail-closed limitation.
Ready for review at candidate 5fd9bed1663a16a4914cc341949e69eb58a2b7b0. Full gate 38841469667 reports success on the exact head; directly executed remote-gate shell exit 0, 20 successful jobs and 2 skipped jobs (not claimed as executed). Current Linux/macOS package artifacts each contain 160 distinct passing test/subtest names; Windows contains 103, including TestCorpusPins, TestPinnedSchemaCheckoutBytes and TestUnqualifiedPlatformRefuses, with package-level pass and no package fail/skip. Focused race tests, 12 of 12 narrowing controls, lint, vet, build and Windows compile pass. Three task-scoped outcomes updated with exact commands, exits, coverage map, mutation table, historical red evidence and current named hosted results. Final identity check confirms 115 live files match the signed CI snapshot, only declared leaf directories change, Story checkpoint/index remain unchanged, and changes remain uncommitted. Windows actual-tree verification remains a declared fail-closed backend limitation; no CLI materialize, capture, hosted launch or runtime qualification is claimed by this helper leaf.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-261016-03b791, pid=37272, exit=0)

## Precondition Resources
- [rc7-leaf-snapshot-identity-verifier.md](file://TASK-261016-1v3gbc/rc7-leaf-snapshot-identity-verifier.md)
- [gw-SNAPID-brief.md](file://TASK-261016-1v3gbc/gw-SNAPID-brief.md)
- [rc7-code-common.md](file://TASK-261016-1v3gbc/rc7-code-common.md)
- [rc7-decisions.md](file://TASK-261016-1v3gbc/rc7-decisions.md)

## Outcome Resources
- [TASK-261016-1v3gbc_spawn-log_-implementer--developer--codex-_RUN-261016-03b791.log](file://TASK-261016-1v3gbc/TASK-261016-1v3gbc_spawn-log_-implementer--developer--codex-_RUN-261016-03b791.log) — System spawn log captured by task-board
- [TASK-261016-1v3gbc_candidate.json](file://TASK-261016-1v3gbc/TASK-261016-1v3gbc_candidate.json)
- [TASK-261016-1v3gbc_handoff.md](file://TASK-261016-1v3gbc/TASK-261016-1v3gbc_handoff.md)
- [TASK-261016-1v3gbc_verification.log](file://TASK-261016-1v3gbc/TASK-261016-1v3gbc_verification.log)

## Created
2026-10-16T01:03:48Z

## Last Update
2026-10-16T03:31:32Z

## Assigned To
[implementer] developer (codex)
