# Two-week critical review of the curator campaign (2026-09-16 .. 2026-10-01)

Read-only research for TASK-261001-1crd2k. No code and no board status were changed.
Reviewer stance: fresh eyes, critical-only. Sources: curator `origin/main` log since
2026-09-16, curator-spec and curator-skill-registry `main`, the board (read-only queries),
`.research/`, the security-audit epic resources, the orchestration journal, and hosted CI
run metadata via `gh`. Code was sampled where a risk looked real; nothing was read whole.

## 1. Honest assessment

The campaign is converging on the part of Apiary that curator owns, but the convergence is
bought at a high process price and the two things Apiary needs *next* are not on main yet.
Since 2026-09-16 curator main took 311 commits, of which 204 are board-state records; the
remaining ~107 landed the security remediation almost completely (EPIC-260910-2hw1xb:
37 of 42 stories done), the M0 release train (spec rc.13 on 09-26, curator v0.15.0-rc.2 on
09-26, launcher v0.1.0 on 09-24), the forward-pin gap ledger, and a lot of CI hygiene. What
is *not* there: the second-operator path (`docs/second-operator.md` is still in a Story
worktree, two landing refusals on 09-30), any leaf for the Gate A / M6 / M3 decisions
(0019, 0014, 0021 are all "proposed — not adopted"), and a clean pin story for content-hash
v2 (the curator half is being built dual-mode against a lagging pin). The process debt is
measurable: four disk-full incidents, one host-memory incident that killed seven runners,
roughly 45 re-apply or carry runs, 12 `integration_base_moved` refusals, and 115 task-board
issues filed in two weeks. The work is honest and well-evidenced; the throughput is being
eaten by the landing pipeline, not by the engineering.

Key aspects at a glance:

- Security remediation is functionally landed, but every user-visible flip is still two
  releases away and one posture row misreports the shipped revision (R2).
- Apiary priority (b), the second operator on 2026-10-01, is unlanded and only stub-verified (R1).
- Gate A / M6 / M3 have no curator or launcher leaves and no adopted decisions (R6).
- Content-hash v2 is being implemented behind an off switch because SPEC_PIN lags spec main (R3).
- The landing pipeline and the host, not the code, are the throughput ceiling (R4, R5).
- The board no longer tells the truth for three epics and about a dozen workspaces (R7).

## 2. Critical recommendations (ordered by impact on Apiary milestones and correctness)

### R1. Land the second-operator path today and prove it with one real login

- **Problem.** tb-keeper's priority (b) is dated 2026-10-01. The guide and the walk exist
  (`TASK-260930-o5uu7a`, ACCEPTED 20:16Z on 09-30) but are not on main: landing was refused
  at 20:41Z (`integration_runner_failed`, empty reason) and again at 21:05Z
  (`validation_suite_changed`, caused by an orchestration config edit). `git ls-tree
  origin/main` has no `docs/second-operator.md`. The first-run fixes (`TASK-260930-3b3oyi`,
  five findings, `to-review`) are behind it.
- **Evidence.** Journal 09-30 20:16Z, 20:41Z, 21:05Z; board: `STORY-260930-3ckyi9`
  integrating, `STORY-260930-9k3uil` to-review. The guide's own "Unverified" section: real
  `claude`/`codex` sessions were not run (stubs only), the Linux credential link, the Codex
  keyring branch and the release installer were not exercised
  (`.temp/STORY-260930-3ckyi9/worktree/docs/second-operator.md:145-154`).
- **Why now.** A second operator who follows a stub-verified guide hits the per-home
  Keychain login and the installer for the first time, with nobody having done it before.
- **Action (size S).** Land o5uu7a with the pre-edit config (already planned), then 3b3oyi.
  Then one 30-minute acceptance on a real second machine or second account: verified
  installer, `profile install --use`, `curator run claude` with a real login, `curator run
  codex` with `auth.json` passthrough. Record the transcript as the Story's acceptance
  evidence and delete the "Unverified" bullets it covers. Do not expand the guide.
- **Decision.** Ivan: is "stub-verified" acceptable for day one? Recommendation: no, one
  real login is the acceptance bar.

### R2. Fix the posture row that reports Codex seed revision B while A is shipped

- **Problem.** `curator status` prints the security-posture table with `codex-seed: B
  (shipped)` while the manager ships revision A. The two rows of the same tool disagree:
  `env status` prints `codex-seed: revision A (shipped)`.
- **Evidence.** `cmd/curator/security_posture.go:20` hard-codes
  `envregistry.CodexSeedRevisionB`; `internal/envregistry/envregistry.go:36` sets the
  shipped revision to A; the rows print at `cmd/curator/envstatus.go:164` (posture table)
  and `:53` (seed rule). The rc.13 vector `conformance/v1/vectors/security-posture.json`
  lines 107 and 313 expect `"codex_seed": "A"`. The production-entry test cannot catch this:
  `cmd/curator/security_posture_conformance_test.go:245-250` builds its expectation from the
  same `securityPostureRevisions()` value (self-referential). `envstatus_test.go:32` pins
  only the seed-rule row. Confirmed by running a binary built from this tree in a throwaway
  HOME with `schema_version: 2`: `curator status` line 13 prints `codex-seed: B (shipped)`;
  `env status` prints `codex-seed: B (shipped)` (line 12) and `codex-seed: revision A
  (shipped)` (line 19) in the same output. Bound: a schema-1 config prints no posture
  inventory, so schema-1 machines never see the wrong row.
- **Why now.** The audit's ground rule 3 is "posture is reported, not implied", and rc.13
  §7.4 makes the A-before-B order normative. curator v0.15.0-rc.3 is the release that ships
  revision A; its status output is the migration signal operators will read.
- **Action (size XS).** Derive the row from `envregistry.CodexSeedRevision`, add a negative
  test that pins the shipped constant independently of `securityPostureRevisions()`, and
  sweep the other four "shipped" rows for the same pattern (`UpdateConfirmation` is also a
  literal). Ship it inside rc.3.
- **Decision.** None.

### R3. Stop building content-hash v2 behind a switch: pin forward, then cut over once

- **Problem.** `SPEC_PIN` is rc.13 (`23435129`, `.github/workflows/ci.yml:43`); spec main is
  `add5023` = rc.13 + lock ordering (#113, curator side landed as `7bc05184`) + notes (#115)
  + content-hash v2 (#116, `b1a2efb`) + checklist (#118). The registry already refuses
  framing-version mismatches (`curator-skill-registry b2b03f7`). The curator half
  (`TASK-260917-2tx81l`) is being built dual-mode: writers stay rc.13, readers accept v2,
  the v2 write path sits behind an internal switch "defaulting off with a TODO for the
  future rc.14 SPEC_PIN bump" (results resource, "Implementation and write policy"). Rev2
  was CHANGES_REQUESTED because `marker.Write` recomputed the hash with the switch off
  (journal 09-30 19:25Z). A switch that exists only to satisfy a lagging pin is dead code on
  every release it ships in, and it is the third time pin lag has shaped code: E2/E4/S4 were
  held five days at rc.11 (`spec-pin-lag-hold.md`), and em42lw lost three gates to the rc.12
  pin alone (journal 09-23 20:06Z).
- **Evidence.** 2tx81l results: count pins 88 (rc.13) vs 97 (`b1a2efb`), both green in the
  lockstep leaf `TASK-260930-12i5zr` (landed `71360fcb`, spec PR #116 pins it).
  The forward-pin machinery with a gap ledger is done (`STORY-260922-2goxjs`, `e337d979`).
- **Why now.** Priority (a) is "spec rc.14 + curator rc.3". Every day the pin lags, the
  v2 code path is untested in production and the registry and manager disagree on what a
  record is.
- **Action (size S + M).** (1) Move `SPEC_PIN` to spec main `add5023` now through the
  forward-pin path (this is what 2goxjs and 18ex37 were built for); (2) rework 2tx81l as a
  single-mode cut-over on that pin, deleting the switch; (3) tag spec rc.14 from `add5023`
  when a release label is wanted (release prep like `TASK-260924-19n6g2`, Relux Bot is in
  `maintainers.allowed_signers` since `2343512`), leaving the csk skillfile-sources half to
  `TASK-260930-3ny11n` and rc.15; (4) curator v0.15.0-rc.3 after (2).
- **Decision.** Ivan: pin to spec main now versus wait for the rc.14 tag; and whether rc.3
  waits for 2tx81l. Recommendation: pin now, rc.3 waits at most two days for 2tx81l.

### R4. Pay down the landing-pipeline tax before the next multi-leaf wave

- **Problem.** The same four defects repeat on almost every landing: (1) a checkpointed
  leaf in a multi-leaf Story whose base moved cannot integrate, converge or withdraw, so a
  carrier task re-applies it (31gaka → 2iu83q, 4pv4au+1sapuy → 36r9k5, 2vapkz → 3vni9d);
  (2) `converge` leaves the old snapshot behind or does not refresh the protected authority
  (10d3l1 rev3 would have reverted 60 trunk files, caught by the reviewer, journal 09-25
  21:xxZ; 55g9dg 09-27 06:25Z; S5 09-29 11:49Z); (3) `integrate` refuses on board-only
  trunk advances and on config-digest drift (task-board #442, #474; o5uu7a 09-30 21:05Z);
  (4) the naming gate is fed by board records (three main reds on 09-29: 01:35Z, 05:55Z,
  08:25Z), fixed by `1kze6z` and `jup8re` but only after a landing queue stalled twice.
- **Evidence.** Journal since 09-23: 12 `base_moved`, 21 `invalidate`, ~45 re-apply or
  carry mentions, 40 carrier mentions; 115 task-board issues opened since 09-16, 57 of them
  between 09-28 and 09-30 (#422–#478). Each re-apply costs a producer run, an identity
  review and a hosted gate; CR gate runs were 30 on 09-29 and 26 on 09-30.
- **Why now.** The Gate A leaves (R6) are exactly the multi-leaf, shared-file Stories that
  trigger all four defects.
- **Action (size M, in skill-project-management).** One bounded bundle: #442, #474, #468 and
  a new issue for the converge snapshot bug, plus a campaign rule already learned but not
  written into the goal: one leaf per Story for anything touching
  `internal/envprofile/status.go`, `platform-cases.tsv`, `conformance-gaps.tsv`,
  `conformance-case-counts.tsv`, `gate-selftest.sh` or `CHANGELOG.md`, landed in a window
  with no other landing.
- **Decision.** Ivan with tb-keeper: allocate task-board time now, or accept the tax through
  the Gate A work. Recommendation: fix #442 and #474 first; they are the cheapest and the
  most frequent.

### R5. Stop using the host as the arbiter and stop paying full CI for board-only pushes

- **Problem.** The orchestration host filled its disk four times (09-23 at 301 MiB free;
  09-25 213 write-boundary snapshots = 27 GiB; 09-29 54 snapshots = 15 GiB; 09-30 97
  snapshots = 13.1 GiB) and went memory-critical once (09-24 13:10Z–13:25Z: seven carry
  runners and the codex app-server died; work stood ~12 h on 22→23.09 for lack of waiters).
  Separately, every board-state push runs the full 12-job matrix including the self-hosted
  rose-air lane: run 36756070796 (board-only commit `5ed5c4e1`) queued rose-air for 86 min;
  main runs took 47–186 min on 09-30. Main went red twice on 09-30 on board-only commits
  (`92c2d7af`, `30b3d678`: no non-board diff against their green neighbours; served-stage
  `go test exit=1`), and the Windows real-git broker row fails at random
  (`STORY-260930-tetmq6`, runs 36604374928 and 36726325163).
- **Evidence.** Journal 09-23 16:52Z, 09-25, 09-29 14:23Z, 09-30 16:00Z; task-board #363
  and #85 open (no snapshot retention); `gh run list` durations; `git diff --stat` of the
  two red commits excluding `.task-board/` is empty. Host `.temp` holds 30 GB in 177 entries.
- **Why now.** A red main from a flake on an unchanged tree turns into
  `revalidation_failed` for the next landing (S5 on 09-29 18:19Z), which turns into
  `base_moved` after the retry, which is R4 again.
- **Action (size S each).** (a) Test lanes skip pushes whose only paths are under
  `.task-board/` (keep the naming gate and gate self-test), or batch board-state commits;
  (b) a pruning timer for terminal write-boundary snapshots until #363 lands, and a written
  producer cap (four concurrent) in the goal file; (c) a `flake` ledger with a retry-once
  in `test-gate.sh` for the two known rows, so a flake is a recorded fact, not a landing
  refusal.
- **Decision.** Ivan: is rose-air a required lane for board-only pushes? Recommendation: no.

### R6. File the Gate A / M6 / M3 leaves now, spec-first, and adopt the three decisions

- **Problem.** Nothing on the board implements what Gate A and M6 need from curator.
  Decisions 0019 (one construction site), 0021 (sessions enter through `curator run`,
  `--untracked`) and 0014 (Claude settings/permissions and Codex rules as machine
  configuration) all read "Status: proposed — not adopted" (`curator-spec/decisions/0014,
  0019, 0021` line 5). No STORY or TASK under any epic names them (backlog and open lists
  queried 09-30). What already exists is real: `curator env resolve --repair --format json`
  is the per-child contract (`cmd/curator/env.go:126`) and is what curator-run calls
  (guide §6); the launcher depends on `skill-agents-management v0.5.22`
  (`curator-agent-launcher/go.mod:6`) and implements 0018 (`5d0d0a3`).
- **Why now.** The rule of the campaign is spec before code; until the decisions are
  adopted no producer can start, and 0014 option 1 is the last curator blocker before
  agents-infra can be archived (M6).
- **Action.** Three adoptions, then leaves in this order: 0019 → launcher leaf "channels
  through agents-management" + curator conformance rows (M); 0014 option 1 → spec closed
  machine surface + curator implementation + vectors (L); 0021 → `--untracked` in the
  launcher (S). File them under one new Story per decision so R4's one-leaf rule holds.
- **Decision.** Ivan: adopt 0019, 0021 and 0014 option 1 explicitly (three signatures).

### R7. Make the board true again: three epics and a dozen workspaces lie

- **Problem.** `EPIC-260822-3ar0tv` is `backlog` with 7 of 8 stories done; `EPIC-260908-2wp8wn`
  is `to-review`; `STORY-260906-1sgt9y` carries three leaves untouched since 09-06/09-15
  (`2x4s7i` to-review, `3x0w4y` development, `vlrjo1` integrating); `TASK-260918-bi6ouz` has
  been integrating since 09-26; `TASK-260924-2am4qa` has been integrating since 09-24 while
  spec PR #89 is CONFLICTING and Ivan ruled v9 outside the accepted revision (journal 09-25).
  Twelve dirty Story workspaces on 15–22.09 bases were listed on 09-27 01:40Z and are still
  the campaign's stated "no control root holding unlanded bytes" debt.
- **Why now.** The campaign's own done criterion and the Gate A epics both read this board.
- **Action (size S).** One read-only reconciliation leaf in the shape of
  `.research/260930_compiled-build-leaves-reconciliation.md` over EPIC-260905, EPIC-260908
  and EPIC-260910-ohqchs, then `close-landed --superseded-by` / `set_status` per verdict,
  close or re-parent #89, and discard every workspace whose bytes are already on main.
- **Decision.** Ivan: the fate of the v9 `directory` feature (#89 and `STORY-260924-1ckno7`).

## 3. What NOT to do now

- **Do not open more compiled-build leaves.** Finish the three in flight (`1uepyd`
  to-review, `rjxrgs`, `20ao7p`) and stop; `STORY-260930-22v3fu`, `2fsqtv`, `1eye8p` and
  `21bsr2` are not in tb-keeper's (a)–(f) and they take the producer and gate slots R1, R3
  and R6 need.
- **Do not start sandbox (EPIC-260728), verified providers (EPIC-260819), PC2 (#92–#95) or
  Muse (#100/#117)** before the R6 leaves exist; tb-keeper puts them after Gate A.
- **Do not add another dual-mode switch to absorb pin lag.** Pin forward (R3).
- **Do not flip revision B** (`STORY-260928-2n2ii7`, `STORY-260928-oflbe1`) in the same
  release that first ships A; the audit's rule 2 and rc.13 §7.4 forbid it, and the flips
  need R2 fixed first so the report is trustworthy.
- **Do not edit spawn or worker configuration while a landing is in flight**; it changed the
  validation digest and refused o5uu7a (09-30 21:05Z). Change configs between landings.
- **Do not rework the naming gate again or grow its exemptions.** Keep board notes free of
  the tokens; the 09-29 05:55Z red was a review note, not code.
- **Do not turn `docs/second-operator.md` into a manual.** Land the walked version, add the
  one real-login line, move on.
- **Do not reopen cross-manager parity or the schema-v6 interop chain.** Ivan dropped
  them on 09-30; the reconciliation already closed nine leaves.

## Appendix: what I ran and did not run

- Board reads only (`task-board q get/list`); no `m` mutations except this task's status.
- `gh run list` / `gh run view` for run metadata and the two failed macOS logs; the
  failing test name is not in `--log-failed` output (only `test-gate: go test exit=1`),
  so it is reported as "served-stage failure on an unchanged tree", not by name.
- `git diff --stat <green> <red> -- . ':(exclude).task-board'` for `39a2f2fd..92c2d7af`
  and `3a436352..30b3d678`: both empty.
- `go build -o /tmp/crit-curator ./cmd/curator` (exit 0, 170 s bound) from this worktree
  (`bab2433b`), then in a throwaway HOME: `curator bootstrap --if-missing --non-interactive
  --skills-root <dir> --default-agents claude_code,codex_cli` (exit 0), `curator status`
  (exit 0) and `curator env status` (exit 0), first with the schema-1 config bootstrap wrote
  (no posture inventory printed) and then with `schema_version` set to 2 (rows quoted in
  R2). No gate or test suite was run; nothing here is a validation claim.
- Not verified: whether any registry service is deployed at `b2b03f7` (unknown); the
  Apiary roadmap document itself (no local copy found; tb-keeper's summary in the journal
  was used as given).
