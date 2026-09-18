# Security-remediation campaign — rules for curator (manager) producers and reviewers (2026-09-17)

These rules apply to every manager/implementation task of `EPIC-260910-2hw1xb`
(and, mutatis mutandis, launcher tasks in curator-agent-launcher). The task
brief carries the product scope; this file carries where and how.

## Where things are on THIS machine
- Authoritative board: `/Users/administrator/Developer/ReluxWorks/curator/curator/.task-board`
  (`TASK_BOARD_DIR` is set for you). Never run task-board against a copy of
  the board, never edit board bytes by hand; use `task-board m …` /
  `task-board resource add …`.
- Code: the curator control root is `/Users/administrator/Developer/ReluxWorks/curator/curator`
  (read-only for you). Your managed **Story worktree** is
  `<control-root>/.temp/<STORY-ID>/worktree` on branch
  `task-board/story/<STORY-ID>`, provisioned by the runtime at spawn. Work ONLY
  there. No writes into the control root, no LOGBOOK.md edits, no writes under
  `~/.agents`, `~/.claude`, `~/.codex`, `~/.curator`, `~/.config/curator-run`.
  Do not commit on the Story branch, do not push, no branches/PRs/tags: the
  runtime publishes your Change Request at handoff and the orchestrator owns
  the signed landing.
- Specification: curator-spec main `23dafa7` is checked out read-only at
  `/Users/administrator/Developer/ReluxWorks/curator/curator-spec`. The
  landed normative text and the conformance vectors for this campaign live
  there (S6: `profiles/manager.md` §8; E2: `protocol/environments.md` §3/§5.5/
  §12; E4: §11/§12; S4: §2.2/§10.3/§12). Read the section your brief names
  before coding; the text is the contract, never a draft or a brief paraphrase.
- Conformance root for local runs:
  `export CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`
  (that checkout is at `23dafa7` and publishes the new vector families). The
  committed CI pin (`SPEC_PIN` in `.github/workflows/ci.yml`) is the released
  revision `87a0d00` (rc.11) and does NOT publish them: a test that consumes a
  new vector file MUST take the repository's established `root-content` path
  when the root lacks it (`t.Skipf("%s publishes no <family> vector", root)`)
  and register its ledger row in `.github/ci/platform-cases.tsv` with class
  `root-content`, exactly like `internal/godriver TestCandidateGoV1SourceAwareContract`.
  Never bump `SPEC_PIN` (owned by the release-qualification tasks), never
  vendor spec bytes into the repository, never weaken a gate.

## What an implementation must contain
1. Behaviour exactly as the landed spec text: closed diagnostics with the
   identical spelling, closed config knobs with the spec defaults, RFC 2119
   obligations honoured, "unreadable is never absence" (environments §8.4)
   at every new read site.
2. **Warn first, then flip**: where the spec defines rollout revisions
   (`A-warning`/`B-enforcing`, `s4-warn`/`s4-enforce`, E4 revisions A/B), the
   code implements BOTH profiles selectable by one internal constant/option
   (so vectors for both run), and the shipped default is the warning profile
   named by the spec. The flip is a later release, never this change.
3. **Posture**: every gate reports its state in `curator status` /
   `curator env status` as the spec's §-posture rows say.
4. Vector execution: a Go test consumes the spec vector family from
   `CURATOR_CONFORMANCE_ROOT` (execution recipe as the spec states), with the
   `root-content` skip when absent; a coverage row is "driven" only when a
   committed test reaches the production entry point named in the AC.
5. Release note: `CHANGELOG.md` Unreleased entry naming the finding id, the
   new commands/diagnostics/knobs and the warn-first step this release ships.

## Validation and evidence
- Run narrow package tests with real exit codes (`set -o pipefail`, state the
  shell): `go build ./... && go vet ./... && gofmt -l . && go test ./<pkgs>...`
  with `CURATOR_CONFORMANCE_ROOT` set, plus `golangci-lint run` if the
  repository's `.golangci.yml` is what CI uses. Do NOT run the full landing
  suite yourself: the runtime runs `scripts/remote-gate.sh` (hosted CI) exactly
  once at handoff.
- Attach `<TASK-ID>_results.md` (task-scoped outcome): per-file changes, how
  each AC line is met with file:line, validation transcripts, the profile
  shipped, what is deliberately out of scope, and any spec gap you found
  (report, do not patch the spec). Tick the checklist items you satisfied.
- Hand off with `task-board handoff <TASK-ID> --role developer`. Do not set
  `done`.

## Models and review
- Producers: Muse `muse-spark-1.3-contributor` at `max`; reviewers: Codex
  `gpt-6-astra` at `low`. Reviewers verify the exact candidate tree, re-run
  narrow tests independently, attack the gate with narrowing mutants, and
  record exactly one verdict through `accept_cr(<ID>, revision=<N>,
  evidence=<their own verdict resource>)` or a changes-requested verdict
  routed by `set_status`.
- Out of scope for every task: the ax integration, proposals 0014–0018,
  tags/releases, SPEC_PIN bumps, anything the brief does not name.

## Rule 8 — no build outputs in the candidate
Never build into the Story worktree root: `go build -o /tmp/… ./cmd/curator` or
the ignored `bin/` directory only. Before every handoff run
`git status --short --untracked-files=all` and confirm the candidate contains
only source, tests, docs and CI changes — no executables, `.review/`,
`LOGBOOK.md` or coverage files. A captured binary is a changes_requested on its
own (seen on TASK-260917-16l2md rev4).
