# TASK-260922-1hla8q — F-S2: fragment permission members + minimum transport token (curator-spec)

Control root: /Users/administrator/Developer/ReluxWorks/curator/curator-spec (work only in your assigned Story
worktree `.temp/STORY-260922-188t6n/worktree`). Board: shared curator board (TASK_BOARD_DIR is set). Read
`campaign-producer-rules.md` (precondition) first.

## What exists
- Decision 0018 is ADOPTED (`decisions/0018-curator-run-permission-interface.md`, adoption choice 7 and the
  Compatibility section): the fail-closed rule is normative now; "the minimum token value and the exact fragment
  member names are to be fixed by the implementing revision (follow-up — the token names the revision that
  defines it)". Non-interactive markers: the closed set {`CI`, `GITHUB_ACTIONS`}.
- `protocol/environments.md` §10.1 "Permission mode" paragraph and §10.2 last paragraph ("Curator delivers the
  section 12.1 `permissions` profile level and the section 12.2 lock engagement to the launcher through this
  fragment; the member names and the minimum transport version token are fixed by the implementing revision…
  until they are, every fragment predates the transport: the launcher MUST refuse would-be `yolo` with
  `permission_policy_unsupported`"). §10.2 defines `launch-env-fragment-v1` as a CLOSED object (readers reject
  unknown fields). §12.1 has the `permissions.<profile>` knob, §12.2 the lockable set with the `native`-only
  direction, §12.5 fragment shaping, §13 vector clause.
- Schemas/cases under `schemas/` and `conformance/` (find the fragment schema and its cases by grepping
  `launch-env-fragment-v1`); gate = `make validate` (the board's validation command is the orchestrator's
  `spec-gate.sh` wrapper of the same steps — do NOT run the full gate manually in addition to the runtime's).

## Deliverable (this revision fixes exactly what choice 7 left open)
1. Fragment members. Define the fragment revision that carries the permission policy: recommended
   `launch-env-fragment-v2` = v1 + one closed member `permissions` = `{ "mode": "native"|"yolo",
   "locked": true|false, "source": "profile"|"global"|"default" }` — pick the final names yourself but keep them
   closed, minimal, and readable by a launcher that never spells a flag. State: v1 readers MUST reject a v1
   fragment carrying the member (a v1 fragment with the member is invalid — the version token is the only
   transport signal); a v2 fragment MUST carry the member (absence is invalid — "a fragment that cannot carry
   the policy or the lock is never silence").
2. Minimum transport version token: the fragment `version` value (`launch-env-fragment-v2`) is the token; a
   launcher establishes transport support iff the fragment version is v2 or later. Name it in §10.1 and §10.2
   and in the 0018 decision's choice-7 row (append the fixed values; do not rewrite history).
3. Headless marker set: keep `{CI, GITHUB_ACTIONS}` stated in ONE normative place (§10.1) with the "additions by
   specification revision only" rule; the launcher SPEC mirrors it (name the mirror; no launcher edit here).
4. Schema + cases: the v2 fragment schema (closed object), cases: v2 valid with member (native unlocked, yolo
   unlocked, native locked); v2 invalid: member absent, unknown mode value, `yolo` with `locked=true` (a lock
   forces native — a fragment carrying both is contradictory and invalid); v1 fragment carrying the member
   invalid; v1 without member still valid (unchanged fixture).
5. `manager-config-v2` / §12.1 unchanged unless the member names force a rename (they should not).
6. CHANGELOG entry under the unreleased section naming F-L1 (launcher, STORY-260922-39hxog) and the curator
   emission follow-up as consumers; results.md must name the curator leaf that has to emit v2 (manager side)
   and the exact member names/values so the orchestrator can create it verbatim.

## Boundaries
No launcher/curator code. Frozen v1 protocol schemas untouched (the fragment is an environments-spec object,
not a v1 protocol schema — say so in results.md if the generator treats it otherwise). Do not touch the
decision's adoption date or resolved choices beyond appending the fixed token/member values to choice 7.
Hand off with `task-board handoff TASK-260922-1hla8q --role developer` after attaching
`TASK-260922-1hla8q_results.md` (what changed, the exact member grammar, cases added with their pass/fail
evidence, gate command + exit code, bounds).
