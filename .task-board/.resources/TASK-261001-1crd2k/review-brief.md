# TASK-261001-1crd2k — independent critical review of the last two weeks (THE ONLY CURRENT INSTRUCTION; read-only research)

You are a senior reviewer with fresh eyes. Ivan asked for an independent look at what the curator campaign did over roughly the last two
weeks (2026-09-16 .. 2026-10-01), and for **only the critically important** recommendations: coherent, actionable, and few (at most
~8). Skip nits, style and "nice to have". The output is recorded as research; nobody acts on it without Ivan.

## What we are doing (context)
- The campaign goal (curator/GOAL-2026-09-23.md): all our epics done; curator main green on every hosted lane (incl. the self-hosted
  rose-air lane); SPEC_PIN on a spec release; every non-executed conformance row an attributed gap with an owner; no control root
  holding unlanded bytes.
- Work is orchestrated through the task board. Producers publish Change Requests, reviewers accept, and landings go through
  `task-board worktree integrate` (curator) or GitHub PR squash-merges (curator-spec, curator-skill-registry).
- The last two weeks focused on:
  - **security remediation** (EPIC-260910-2hw1xb). It was driven by a 2026-09 security audit; the audit docs are in
    .task-board/.resources/EPIC-260910-2hw1xb/.
  - **staged security rollouts**: revision A shipped, revision B flips wait for a release.
  - **content-hash framing v2**: spec b1a2efb, registry b2b03f7; the curator implementation TASK-260917-2tx81l is in review/rework.
  - **CI hygiene**: the naming gate, runner labels, git-config isolation, Windows flakes.
  - **a compiled-builds board reconciliation**: .research/260930_compiled-build-leaves-reconciliation.md.
- The orchestration journal is at /Users/administrator/Developer/curator/.temp/orchestration/state.md. It is read-only for you, it is
  mostly Russian, and the last ~300 lines are most relevant.

## The Apiary plan and what it puts on curator (from tb-keeper, 2026-09-30; plan: relux-works/wiki roadmap/ecosystem-roadmap.md DRAFT v4 §2b)
Order:
1. SH session-host + CM1.
2. PC0/PC1 process configuration + demo playbooks.
3. The rest of M0, then M2 + M5a, then M6 = the agents-infra exit (Gate A EPIC-260927-137kpm, Gate B EPIC-260927-1anzwe).
4. CM2, M3, PC2 (roles as skills).
5. The rest of M4, M5b, PC3.

Later: M7 network, M8 routing, KP keeper, AR Apiary distribution.

What rests on curator:
- **M0:** the next rc tags; onboarding ivmbp/air-rose onto `curator run` + the relux-root-context profile.
- **M2 / Gate A:** Decision 0019 (one construction site) is still *proposed*. curator-run must build channels through the
  agents-management module, and `curator env resolve --repair --format json` is the per-child contract.
- **M3:** Decision 0021 (sessions enter through curator run; --untracked) is *proposed*.
- **M6:** Decision 0014 option 1 (Claude settings/permissions and Codex rules as Curator machine configuration) is *proposed*. It is
  the last curator blocker before agents-infra can be archived.
- **M5a:** Decision 0020 (inference plane) is not written yet.
- **PC2:** per-role profiles; curator-spec/curator #92–#95 (source shorthands, profile export/import, attestation fast path,
  command-only installs).
- **Muse:** curator-spec#117 / curator#100.
- **Remote workers on the Mac mini:** HOME and credentials via managed homes.

tb-keeper's priority for curator:
- (a) finish security → spec rc.14 + curator v0.15.0-rc.3;
- (b) **2026-10-01: Alexis becomes a second operator**. Managed-home profiles + curator run must work, with credential passthrough
  documented. TASK-260930-o5uu7a verified the path (no blockers) and wrote docs/second-operator.md; small first-run fixes are in
  flight;
- (c) Gate A curator blockers;
- (d) Decision 0014 option 1;
- (e) Muse;
- (f) PC2.

Sandbox and verified providers come later. Ivan dropped cross-manager parity with csk.

## What to review (sources)
- `git log --since=2026-09-16` on curator main (this worktree's origin/main).
- ../curator-spec and ../curator-skill-registry main (read-only, `git -C`).
- The board, read-only: `task-board q` over EPIC-260910-2hw1xb, EPIC-260908-2wp8wn, EPIC-260720-21aq1i and the open/blocked
  elements.
- Recent research docs in .research/.
- The journal above.

Sample the code where a risk looks real; don't read everything.

## What to produce
Write `.research/261001_two-week-critical-review.md` in this repo, with:
1. A one-paragraph honest assessment: are we converging toward what Apiary needs from curator?
2. **Critical recommendations**, at most ~8, ordered by impact on the Apiary milestones and on correctness/security. Each one must
   have:
   - the problem;
   - concrete evidence (commits, file:line, board ids, CI runs);
   - why it is critical now;
   - a coherent proposed action, with a rough size;
   - whether it needs a decision from Ivan.

   Consider in particular:
   - process debt that repeatedly costs time (landing refusals, re-applies, multi-leaf checkpoint traps, board staleness,
     disk/memory pressure);
   - risks in what shipped (security semantics, spec/manager drift, the rc.13 pin vs v2 writers);
   - gaps between curator and the Apiary gates (0019/0014/0021, curator run for a second operator).
3. "What NOT to do": work that looks tempting but is not critical now.

Rules:
- Change NO code and NO board status.
- Never spell any employer name.
- No LOGBOOK/CHANGELOG.
- Keep the document under ~250 lines.
- Attach it also as the results resource `TASK-261001-1crd2k_results.md`.

Then run `task-board handoff TASK-261001-1crd2k --role researcher` and END YOUR TURN.
