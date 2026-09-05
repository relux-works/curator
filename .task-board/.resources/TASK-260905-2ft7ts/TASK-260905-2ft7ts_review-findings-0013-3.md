# Review findings 0013-3: Decision 0013 at 6cbe9ae (confirm F8/F9 edit)

Subject: `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-decision-0013`,
branch `draft/decision-0013-execution-ownership`, head `6cbe9ae` (parent `7cb24bd`,
base `b4f29cd`), PR relux-works/curator-spec#38 (head OID `6cbe9ae…`, base `main`,
OPEN — verified with `gh pr view 38`). Working tree clean. Commit signed:
`Good "git" signature with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`
(the "No principal matched" line is the same local allowed-signers artifact as before).

Verdict: **ACCEPT** (0 blocking, 0 major, 0 minor, 1 nit). Route: `accept_cr` → `to-review`.

## 1. Scope check

`git diff --stat 7cb24bd..6cbe9ae`: one file, 12+/9−, only
`decisions/0013-execution-ownership-and-launch-plans.md`. Three hunks: 3.3 residual-bound
paragraph, 3.4 planning-role step, Decision 7 item 4. Nothing else touched.

## 2. F8 (minor, cycle 2) — resolved

3.4 now reads "`ax start --launch-plan` adds one step before persistence for both forms:
`ax` calls provider `launch` with the candidate record in planning role, takes the returned
`SpawnPlan.argv` as the resolved final argv (for the `argv` form, the plugin's verbatim
translation plus its §3.6 profile-flag check), validates it under §3.3, and only then
persists the record — so a §3.6 refusal, like every §3.3 refusal, fires before any Session
Record exists." Decision 7 item 4 now names "for both forms" and "a document whose `argv`
or `argv_suffix` carries the provider's own ax §7.7 `yolo` flag". This closes the
`argv`-form gap: 3.3's blanket "before any Session Record" claim for `launch_plan_invalid`
is now true for the `profile_flag` reason too, and the negative conformance case covers
the form the cycle-2 attack found.

Consistency re-checked: ax §13.1 at the current ax main (`28bf96d`) still persists the
Session Record at step 2 and calls provider `launch` at step 4 (SPEC.md lines 8117–8124),
so the inserted pre-persistence planning-role call is the right place. 3.5's
`capability_unavailable` refusal ("before the plugin is invoked") and 3.6's MUST are
unchanged and compose with 3.4: capability check → planning-role launch (incl. §3.6) →
§3.3 validation → persist. No new contradiction with §3.3/§3.4/§3.6/Decision 7.

## 3. F9 (nit, cycle 2) — resolved where filed, residual copies elsewhere

3.3 now states the bound as caller `extensions` (with the four Curator keys among them for
a composed document — "`ax` is generic and knows no such class") ⊕ `ax.launch-plan-request`.
Correct.

### F10 — nit — the F9 double-counting phrasing survives in two places not in the diff

- 3.4, last sentence: "What remains bounded by ax §1.6 is the caller's own `extensions`
  plus this key plus the Curator keys".
- Decision 7 item 4: "a document whose `extensions` would not fit ax §1.6 together with the
  ax and Curator keys".

Both name the Curator keys as a third class again. Harmless — the check is over the same
persisted object either way, and 3.3 is the normative statement — but the wording is now
inconsistent with the corrected 3.3. Fix: "plus this key" / "together with the
`ax.launch-plan-request` key" when the file is next touched. Not worth another cycle.

## 4. Empty repository delta

`CR-TASK-260905-2ft7ts-2` rev 2 has `repository_delta=empty` (candidate tree == base
`b4f29cd`). Accepted as the correct outcome, as in cycles 1 and 2: the producer brief
requires the document to live only in the external decision worktree and to touch nothing
else; the reviewable artifact is signed commit `6cbe9ae` on
`draft/decision-0013-execution-ownership`, published as PR #38, which the orchestrator
lands. The story worktree correctly carries no change.

## 5. Not re-verified this cycle

Everything outside the 7cb24bd..6cbe9ae diff stands on the cycle-1/cycle-2 findings
(F1–F7 verified at 7cb24bd); no fact those cycles established is contradicted by this edit.
