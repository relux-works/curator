# Rework report 0013-1: Decision 0013 at 7cb24bd

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-decision-0013`,
branch `draft/decision-0013-execution-ownership`, base `b4f29cd`.
Edited file: `decisions/0013-execution-ownership-and-launch-plans.md` only.
Commit: `7cb24bd2bd6c99cefebd061afd9d5d105e12645b` (parent `71ac9d1`), signed.
`git log --show-signature -1`:
`Good "git" signature with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`
(the "No principal matched" line is the same local allowed-signers artifact the base
commit shows; not a signing defect). `git status --short` after commit: empty.
Delta: 99 insertions, 36 deletions, one file. Not pushed, no tag, no PR.

## Finding → disposition

| Finding | Severity | Disposition | Section(s) changed |
| --- | --- | --- | --- |
| F1 `ax.launch-plan-request` cannot fit a valid plan | major | Applied as decided: extension value is now `{form, base_argv_length, request_digest}`; suffix defined as `launch_plan.argv[base_argv_length:]`; 3.5 replays from there; residual bound stated (caller `extensions` + ax key + four Curator keys must fit ax §1.6, refused at 3.3 with `launch_plan_invalid`, `field: "extensions"`). | 3.3, 3.4, 3.5 |
| F2 bypass gate was MAY | major | Applied as decided: `caller_launch_plan` plugin MUST refuse any caller element (in `argv` or `argv_suffix`) equal to a flag of its own ax §7.7 `yolo` mapping (long form or documented alias) with `launch_plan_invalid`, `reason: "profile_flag"`, `details.argv_index`, before process creation; no spellings listed in 0013; "the composer never emits one" kept. Negative conformance case added to Decision 7 item 4 (suffix with provider yolo flag under `--profile standard` refused). Open question 2 scoped to non-profile collisions. | 3.5, 3.6, 7.4 |
| F3 secret code conflicted with ax §15.3 | minor | Applied: secret-rule violations refuse with existing `secret_policy_violation` (exit 16) with `details.field`; `launch_plan_invalid` covers shape, limits, exclusivity, unknown members, extension collisions only. Cites ax §15.3. | 3.3 |
| F4 `base64` vs ax §1.6 | minor | Applied: `encoding` is exactly `utf-8` or `base64url` (unpadded RFC 4648 URL-safe, ax §1.6); mapping paragraph updated. | 4 |
| F5 env_names/env_literals collision | minor | Applied: Decision 6.3 rule — a name in both composed `env_literals` and fragment `mcp.env_names` is dropped from `env_names` (composer-set literal wins over destination-local lookup) with a stderr warning naming the variable; document disjoint by construction; ax §5.1 disjointness never fires for a composed document. Referenced from 6.4. | 6.3, 6.4 |
| F6 default session name > 64 | nit | Applied: default is `<env-id>-<utc-stamp>`; profile name lives in `works.relux.curator.profile-name`; `--name` overrides; ax §2.1 grammar and 1–64 length cited; stated that the default always fits (longest launcher §4.2 env-id `claude_code` is 11 chars; any env-id up to 47 chars fits beside the 16-char stamp). Open question 1 reworded. | 6.4, Open q. 1 |
| F7 agents-management main moved | nit | Applied: pin moved to `91bf945` in the Status section-reference paragraph and in Decision 4's qwen statement. | Status, 4 |

No deviations from the orchestrator's author decisions.

## Verified in this cycle

- skill-agents-management `91bf945` ("Resolve a declared model alias to its identity
  before argv"): `git diff --stat 944c7b4 91bf945` touches `pkg/agentic/plan.go`
  (+60), `pkg/agentic/system.go` (+32, `AliasOf`/`LaunchIdentity`), vendorplugin alias
  files and tests; every `LaunchMode`/`Lineup`/`StdinPayload`/`EffortTransport` match in
  the diff is a test line. At `91bf945`: `ErrUnsupportedLaunchMode` (`plan.go:84`),
  `EffortTransportStdin` (`system.go:148`), `type StdinPayload` (`system.go:272`),
  qwen `EffortTransport: agentic.EffortTransportStdin` (`systems/qwen/qwen.go:109`),
  `func Lineup(models []Model) []RankedModel` (`pkg/vendorplugin/lineup.go:52`).
- ax `28bf96d` (v0.5.0): §15.3 exit 2 = `interactive_choice_required`,
  `invalid_arguments`; exit 16 = `confirmation_required`, `policy_refused`,
  `secret_policy_violation`. §7.7 yolo mappings: Codex
  `--dangerously-bypass-approvals-and-sandbox` (alias `--yolo` accepted), Claude
  `--dangerously-skip-permissions`, Gemini `--approval-mode=yolo`, Muse `--yolo`,
  Antigravity `--dangerously-skip-permissions`. §1.6 "bytes MUST be represented as a
  content-addressed blob or unpadded base64url" and RFC 4648 URL-safe alphabet note.
  §2.1 session name "1–64 characters matching `[A-Za-z0-9][A-Za-z0-9._-]{0,63}`".
- launcher `6de42d8` §3/§4.2: `<env-id>` is one of `claude_code`, `codex_cli`,
  `opencode`, `pi` (closed table); `claude_code` is the longest at 11 characters.
- curator-spec `protocol/core.md` §2 identifier grammar `^[A-Za-z0-9][A-Za-z0-9._-]*$`
  (inside ax §2.1 except length).

## Not verified (docs-confidence, unchanged from cycle 1)

- Idempotence of the current ax plugin `launch` operation (required by 3.4, not asserted).
- Per-tool "everything after the last flag is the user turn" behavior (left to launcher
  SPEC 0.2).

Validation commands: no build or test suite applies to a Markdown decision; checks run
were `git diff --stat` (1 file), `git status --short` (clean, exit 0), and
`git log --show-signature -1` (good signature, exit 0).

## Regression check (gate evidence for checklist items 10–11)

Script `.temp/TASK-260905-2ft7ts/check-0013-rework-1.sh <rev>` in the story worktree:
15 fixed-string assertions (7 "defect text absent", 8 "fix text present"), one per
finding, evaluated over `git show <rev>:decisions/0013-…md` of the decision worktree.

| Rev | Command | Exit | Result |
| --- | --- | ---: | --- |
| `71ac9d1` (pre-rework) | `check-0013-rework-1.sh 71ac9d1` | 1 | expected red: 15/15 assertions fail (every defect present, every fix missing) |
| `7cb24bd` (rework) | `check-0013-rework-1.sh 7cb24bd` | 0 | 15/15 pass |

Logs: `.temp/TASK-260905-2ft7ts/check-old-01.log`, `check-new-01.log`. Blind spot: this
is a text-level check of the decision document; it cannot exercise the ax, launcher, or
agents-management gates the decision specifies — those are the conformance cases named
in Decision 7 item 4 for the implementing PRs.
