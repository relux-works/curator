# TASK-260905-30a1bz — revision report: ax PR #1 per Decision 0013

Repository: `relux-works/agent-session-manager-spec`, branch
`draft/curator-environment-integration`, PR #1
(https://github.com/relux-works/agent-session-manager-spec/pull/1), base `main` = `28bf96d` (v0.5.0).
Authority: curator-spec Decision 0013 at `83de1a5` (Decisions 3, 4, 7 items 1–8, 8).
Worktree: `/Users/iv/Developer/ReluxWorks/.temp/ax-curator-integration/worktree`.

## Commits (new, on top of `d7075e1`; `d7075e1` unchanged)

| Commit | Subject | Signature |
| --- | --- | --- |
| `ef9198565f0221422e259438b6a68212aa821253` | spec: revise Curator integration per curator-spec Decision 0013 (ax start --launch-plan) | Good "git" signature for oparin@me.com with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM |
| `2c7f6421c82f149bba0d55d5d6a2b97fa729f506` | validation: expected-red mutations for the launch-plan gate | Good "git" signature for oparin@me.com with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM |

Author `Ivan Oparin <oparin@me.com>`, SSH-signed with `~/.ssh/ivanopcode`, no AI
attribution trailer (CONTRIBUTING "AI attribution policy"). Note: CONTRIBUTING's
release-publication gate says automation must hand commit/push commands to a
human; the producer brief for this task explicitly authorized signed commits and a
plain push of the PR branch, and that authorization was followed. No merge, no tag,
no touch of `VERSION`, `CHANGELOG.md`, `RELEASE_NOTES.md`, `README.md`.

## Decision 7 item → SPEC section

| Item | Content | SPEC.md sections touched |
| --- | --- | --- |
| 1 | §7.5 merge paragraph replaced by `--launch-plan` operation; "adds no member to SpawnPlan" withdrawn | §7.5 (caller-plan paragraphs, `SpawnPlan.stdin`) |
| 2 | Launch Plan `stdin` row + Launch Stdin object; 4th key `system-modules`; `profile-pin` = `sha256:` lock hash | §5.1 |
| 3 | `caller_launch_plan` (8th), `stdin_resume_replay` (9th); manifest/probe schema consequence stated; `SpawnPlan.stdin`; `resume.launch_plan`; verbatim-translation plugin contract; `capability_unavailable` pre-invocation | §7.3, §7.4, §7.5 (probe row bound 0..9), §14.2 `SessionSummary` bound 0..9 |
| 4 | Planning-role `launch` before persistence; final argv in `launch_plan.argv`; `ax.launch-plan-request` `{form, base_argv_length, request_digest}` (no suffix copy); persisted-extensions bound → `launch_plan_invalid field: extensions`; determinism → `provider_protocol_error`; `LAUNCH-PLAN-*` case table | §13.1, §19.4 `AC-LAUNCH-003` |
| 5 | Refuse-on-drift by default when `system-modules: true` (`policy_refused`, exit 16, `details.reason: "environment_drift"`); warn-and-continue otherwise; strict mode stays; failed resolution distinct; like-with-like | §13.10 |
| 6 | `fragment-digest` over CCJ-1 canonical bytes of the parsed fragment | §5.1 |
| 7 | Grammar row `ax start NAME --provider ID --launch-plan FILE|- …`; exclusivity with `--task-board` and `argv`+`--profile yolo` = `invalid_arguments`; document table + JSON example; `launch_plan_invalid` exit 2 `details.field`; secrets keep `secret_policy_violation`; mandatory profile-flag refusal (`reason: "profile_flag"`, `argv_index`); `curator session` note kept | §14.1, §15.3 ("Caller launch-plan code" subsection) |
| 8 | v0.6.0 proposed in PR body only; §1.5 row `urn:ax:schema:launch-plan-request` 1.0.0; Appendix D.2 row + D.4 bullet; fixture with 3 valid + 15 invalid cases; traceability A.1 row + new A.12 | §1.5, Appendix A.1, A.12, D.2, D.4, `fixtures/launch_plan_request_conformance.json` |

Tooling: `scripts/validate_launch_plan.py` (new gate, 6 classes, executes the fixture
through a reference implementation of §14.1 validation + §13.1 planning-role
resolution + §7.7 profile-flag refusal; ties fixture mappings to the §7.7 table;
18 SPEC semantic markers); wired into `scripts/validate_spec.py` (import, required
file, main, ledger line); `FROZEN_RELEASE_DOCUMENT_SHA256["SPEC.md"]` re-minted to
`ed08fe6bfcc4802c8055907546ba5b550018eef90379e248cee7473e376c7669` (only the SPEC
digest moved); 10 expected-red mutations added to `scripts/test_expected_red.sh`
(7 fixture mutations incl. narrowing — drop `--yolo` alias, relax extension bound,
relabel secret/argv-yolo codes, admit capability-less plugin, admit profile flag,
drift positive argv — and 3 SPEC-marker mutations: exit-class move, drift refusal
weakened to warn, registry reverted to seven names).

## Gates (standalone processes, real exit codes)

| Command | Exit | Evidence |
| --- | ---: | --- |
| `./scripts/validate_spec.py` | 0 | 286/286 semantic checks; launch-plan ledger gate_classes=6, positive=3, negative=15 |
| `./run_validation.sh` | 0 | structurizr-cli 2025.11.09 / plantuml present; "Validation successful"; log `run-validation-01.log` |
| `./scripts/test_expected_red.sh` | 0 | 314 passed, 0 failed out of 314 mutations (10 new launch-plan mutations all rejected); log `expected-red-01.log` |
| `git diff --check` | 0 | clean |

Logs live under curator-spec worktree `.temp/TASK-260905-30a1bz/` (gitignored).

## Publication

- Push: plain `git push origin draft/curator-environment-integration` (fast-forward from `d7075e1`; no `--force`).
- PR #1 new head: `2c7f6421c82f149bba0d55d5d6a2b97fa729f506` (origin branch = local head; PR state OPEN)
- PR description updated via `gh pr edit 1 --body-file` citing Decision 0013 and proposing v0.6.0. PR NOT merged.

## Unverified / left to the maintainer

- Per-contract minor version numbering (Session Record 1.1.0, Provider manifest/probe 1.1.0, Provider Protocol 2.1.0/3.1.0, Structured Error 1.4.0) is proposed in prose only; §1.5 rows other than the new schema are unchanged.
- The secret classifier in `validate_launch_plan.py` is a bounded fixture-gate rule (§16.2 name class + credential prefixes); the ax implementation owns its real scanner.
- No `ax` binary exists; `LAUNCH-PLAN-*` cases are executed against the spec-repo reference gate, not a product entry point (consistent with §20.2 / CONTRIBUTING).
- Decision 0013 open questions 2 and 4 remain open in the text by design.
