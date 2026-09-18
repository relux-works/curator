# Security-remediation campaign — rules for curator-skill-registry (service) producers and reviewers (2026-09-17)

These rules apply to every service task of `EPIC-260910-16qce1`. The task brief
carries the product scope; this file carries where and how.

## Where things are on THIS machine
- Authoritative board: `/Users/administrator/Developer/ReluxWorks/curator/curator/.task-board`
  (`TASK_BOARD_DIR` is set for you). Never run task-board against a copy of the
  board, never edit board bytes by hand; use `task-board m …` / `task-board resource add …`.
- Code: the curator-skill-registry control root is
  `/Users/administrator/Developer/ReluxWorks/curator/curator-skill-registry`
  (read-only for you; `main` = `c7ef32c`, the R3/P2 landing). Your managed **Story
  worktree** is `<control-root>/.temp/<STORY-ID>/worktree` on branch
  `task-board/story/<STORY-ID>`, provisioned by the runtime at spawn. Work ONLY
  there. No writes into the control root, no writes under `~/.agents`,
  `~/.claude`, `~/.codex`, `~/.curator`. Do not commit on the Story branch, do
  not push, no branches/PRs/tags/releases: the runtime publishes your Change
  Request at handoff and the orchestrator owns the signed landing.
- Specification: curator-spec `main` (`5146c7b` at the time of writing; the registry contract you implement is the revision the CI pin names — `47c3c8c`, R3/P2 — read it from a detached worktree at that commit, not from `main`) is checked out read-only at
  `/Users/administrator/Developer/ReluxWorks/curator/curator-spec`
  (`protocol/registry.md` §5/§9/§9.3, `profiles/registry-service.md` §2/§5/§11,
  `schemas/v1/records-response-v2.schema.json`, `log-response-v2.schema.json`,
  `conformance/v1/vectors/registry-service.json`, `registry-client.json`,
  `schema-cases/{records,log}-response-v2/`). The text is the contract, never a
  brief paraphrase.
- Conformance root for local runs = the commit the workflow pins:
  `.github/workflows/ci.yml` checks out the protocol suite at a pinned commit
  (`ref:` under "Checkout authoritative protocol suite"), currently
  `47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe` (R3/P2; the suite's
  `checkpoint_cases` test needs it — an older root fails that test, the
  `main` tip carries families the service does not implement). Create a
  detached worktree and export it:
  `git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec worktree add /tmp/spec-47c3c8c 47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe`
  (skip if present) and
  `export CURATOR_CONFORMANCE_ROOT=/tmp/spec-47c3c8c/conformance/v1`.
  A task that consumes newer vectors or schemas moves the pin forward in the
  same change (never backward) and keeps every pre-existing conformance test
  green at the new pin.
- Python: create a venv OUTSIDE the tree (e.g. `/tmp/csk-venv`:
  `python3 -m venv /tmp/csk-venv && /tmp/csk-venv/bin/pip install -e '.[dev]'`
  from the worktree), never `.venv` inside the worktree. State the interpreter
  version(s) you ran (CI runs 3.11 and 3.14 on three OSes).

## What an implementation must contain
1. Behaviour exactly as the landed spec text: closed error codes with the
   identical spelling, RFC 2119 obligations honoured, the response envelopes
   validating against the named schema version.
2. Conformance: the shared vectors of `registry-service.json` /
   `registry-behavior.json` the brief names are driven by tests in
   `tests/test_protocol_conformance.py` through the real endpoints or the
   store entry points the spec names (not helper-only tests); schema-cases
   are asserted against the served envelopes where the brief says so.
3. Posture/observability where the brief names it; `README.md`/`SECURITY.md`
   updated when operator-visible behaviour changes.
4. Release note: `CHANGELOG.md` Unreleased entry naming the finding id, the
   envelope/field/error changes and the spec revision it implements (the pinned curator-spec commit).

## Validation and evidence
- `python -m pytest -q` with `CURATOR_CONFORMANCE_ROOT` set (state the exact
  command, shell, interpreter and exit code) and `python -m mypy` (strict, per
  `pyproject.toml`), both from the worktree, `set -o pipefail`. Do NOT run the
  hosted gate yourself: the runtime runs `sh scripts/remote-gate.sh` (GitHub
  CI: three-OS × two-Python matrix, mypy, build, docker) once at handoff.
- Attach `<TASK-ID>_results.md` (task-scoped outcome) — WRITE IT OUTSIDE the
  Story worktree (e.g. `/tmp/<TASK-ID>/results.md`) and attach it from there:
  any file inside the worktree becomes part of the candidate tree, and a
  results/logbook/coverage file at the repository root is a changes_requested
  on its own (seen on TASK-260910-2g5v17 rev2). Contents: per-file changes, how
  each AC line is met with file:line, validation transcripts, what is
  deliberately out of scope, and any spec gap you found (report; never patch
  the spec). Tick the checklist items you satisfied with
  `task-board m 'check_item(<TASK-ID>, item=N)'`.
- Hand off with `task-board handoff <TASK-ID> --role developer`. Do not set
  `done`.

## Models and review
- Producers: Muse `muse-spark-1.3-contributor` at `max`; reviewers: Claude `claude-opus-5` at `max` (from 2026-09-18; codex `gpt-6-astra:low` before). Reviewers verify the exact worktree tree, re-run
  pytest + mypy independently, attack the change with narrowing mutants, and
  record exactly one verdict resource `<TASK-ID>_review-verdict-rev<N>.md`
  before routing (`accept_cr` or `set_status(to-dev)`), and leave no files in
  the worktree (use a disposable copy elsewhere for attacks). Hygiene is part
  of every verdict: `git status --short --untracked-files=all` of the
  candidate must list only source, tests, docs and packaging paths — a
  results/logbook/coverage/build file (repo root or anywhere) is a finding,
  never accepted.
- Out of scope for every service task: the manager client, curator-spec edits,
  tags/releases, Docker/deploy changes beyond what the brief names.
