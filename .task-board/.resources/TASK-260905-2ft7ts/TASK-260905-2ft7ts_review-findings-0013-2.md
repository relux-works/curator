# Review findings 0013-2: Decision 0013 at 7cb24bd

Subject: `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-decision-0013`,
branch `draft/decision-0013-execution-ownership`, head `7cb24bd` (parent `71ac9d1`),
base `b4f29cd`. Change Request `CR-TASK-260905-2ft7ts-2` rev 2 on the story branch has
`repository_delta=empty` (candidate tree == base, zero-path patch). As in cycle 1 this is
expected by the producer brief's design: the deliverable lives only in the external
decision worktree, not in the story worktree. Judged there.

Verdict: **ACCEPT** (0 blocking, 0 major, 1 minor, 1 nit). Route: `accept_cr` → `to-review`.

## 1. Seven cycle-1 findings — verified against `git diff 71ac9d1..7cb24bd` (99+/36−, one file)

| Finding | Author decision | Resolved? | Where / evidence |
| --- | --- | --- | --- |
| F1 major | drop `argv_suffix` from `ax.launch-plan-request`; suffix = `launch_plan.argv[base_argv_length:]`; residual §1.6 bound refused at 3.3 with `field: "extensions"` | yes | 3.4 value is `{form, base_argv_length, request_digest}`; 3.5 replays `launch_plan.argv[base_argv_length:]`; 3.3 new paragraph states the bound (64 keys / 65,536 canonical bytes — matches ax §1.6 lines 344–350) |
| F2 major | MUST refuse §7.7 yolo flag (long or alias) with `launch_plan_invalid`, `reason: "profile_flag"`, `details.argv_index`; negative case in Decision 7 | yes | 3.6 rewritten as MUST, keyed on ax §7.7 (verified table: Codex long form + `--yolo` alias, Claude, Gemini, Muse, Antigravity); Decision 7 item 4 carries the negative case; "composer never emits one" kept; Open q. 2 narrowed to non-profile collisions |
| F3 minor | `secret_policy_violation` (exit 16) with `details.field`; `launch_plan_invalid` = shape/limits/exclusivity/collisions | yes | 3.3; ax §15.3 verified: exit 2 = `invalid_arguments`, exit 16 = `secret_policy_violation`, new codes admitted in a minor |
| F4 minor | `utf-8` or `base64url` unpadded per ax §1.6 | yes | Decision 4 both bullets + mapping paragraph; ax §1.6 line 232/247 verified; no residual `base64` spelling (grep) |
| F5 minor | literal-wins collision rule in 6.3, referenced from 6.4, warning on stderr | yes | 6.3 `env_names` bullet; 6.4 "with the Decision 6.3 collision rule already applied" |
| F6 nit | `<env-id>-<utc-stamp>`, `--name` overrides, ax §2.1 grammar, default always fits | yes | 6.4 + Open q. 1; ax §2.1 line 363 verified; launcher §3 env-id table closed (`claude_code` 11 chars is longest); the 6.2 example profile name is 44 chars as stated |
| F7 nit | pin to `91bf945` | yes | Status paragraph + Decision 4 qwen statement; `git log main` = `91bf945` "Resolve a declared model alias…"; all 21 identifier hits in `944c7b4..91bf945` are test lines; `ErrUnsupportedLaunchMode` plan.go:84, `EffortTransportStdin` system.go:148, qwen.go:109, `func Lineup` lineup.go:52 verified |

The rework report's disposition table matches the diff; "no deviations" is true.
The commit is signed (`Good "git" signature with ECDSA key SHA256:V6JiKG7J…`; the
"No principal matched" line is the same local allowed-signers artifact as `b4f29cd`).
`git status --short` clean. Only the decision file changed.

## 2. Attack pass on the amended text

Checked: 3.4/3.5 replay definition is single-sourced (`argv` form → `base_argv_length`
0, suffix = whole argv, nothing appended on resume — no second reading); 6.3/6.4
cross-references resolve; Decision 7 items 2–4 reflect F1–F4; no contradiction with
ax §1.6/§5.1/§7.7/§15.3 on re-read; no regression of cycle-1 facts (Option A, Option B
reasoning, `LaunchModeInteractive` constraints, launcher ordering, `ax_handoff_failed`
terminal, CCJ-1 digest, `system-modules` key all unchanged).

### F8 — minor — 3.6 vs 3.3/3.4: for the `argv` form the profile-flag refusal can fire after the Session Record exists

Quote (3.4): "`ax start --launch-plan` with `argv_suffix` adds one step before
persistence: `ax` calls provider `launch` … in planning role". Quote (3.6): "A
`caller_launch_plan` plugin MUST refuse, before process creation …". Quote (3.3):
"A violation refuses `ax start` before any Session Record … with … `launch_plan_invalid`".

What is wrong: the planning-role call is specified only for `argv_suffix`. For the
`argv` form the plugin is first invoked at ax §13.1 step 4, after step 2 persisted the
record, so a §7.7 flag in a complete `argv` is refused by the plugin with
`launch_plan_invalid` after a Session Record exists — contradicting 3.3's blanket
"before any Session Record" for that code and leaving an orphan record. Process
creation is still prevented (the gate holds); only the sequencing claim is
imprecise. Decision 7 item 4's conformance case is worded for `argv_suffix` only, so
it does not catch the `argv`-form gap.

Fix (one sentence in 3.4): make the planning-role `launch` step apply to both forms
(the `argv` form also needs the plugin's §7.7 check before persistence), and extend
the Decision 7 item 4 negative case to "in `argv` or `argv_suffix`". Can be folded into
the next edit of the document (e.g. the environments 1.1 batch or a cycle addressing
Open question 2); not worth a third producer cycle on its own.

### F9 — nit — 3.3 residual-bound wording double-counts the Curator keys

Quote: "the caller's `extensions` together with the `ax.launch-plan-request` key (§3.4)
and the four `works.relux.curator.*` keys the composer sets (Decision 6.4) MUST fit".
For a composed document the four Curator keys *are* members of the caller's
`extensions` (6.4); `ax` is generic and cannot know them as a separate class. The
bound is simply caller `extensions` ⊕ `ax.launch-plan-request`. Harmless (the check
is over the persisted object either way); fix wording when the file is next touched.

## 3. Empty repository delta

Accepted as the correct outcome: the producer brief instructs "the document lives
only in the worktree above" and "touch nothing else in the repository"; the story
worktree provisions are untouched by design, and the reviewable artifact is the
signed commit `7cb24bd` on `draft/decision-0013-execution-ownership`, which the
orchestrator lands.

## 4. Not verified (docs-confidence, unchanged)

Idempotence of the current ax plugin `launch`; per-tool "everything after the last
flag is the user turn" — both left to the implementing PRs, as the decision states.
