# TASK-260908-1wmb40 publication — ordinary producer run after privileged recovery (RUN-260909-e08d63)

Role: developer (implementer), Muse Spark. Story STORY-260908-v16gn5.
This is a publication-only follow-up to successful landed recovery RUN-260909-49fa72. No source redevelopment, no export reapply, no start-landed-rework rerun, no broad test/mutant rerun this turn per recovery-assignment-correction + recovered-candidate-publication overrides.

## Preserved state (observed this run)

- Managed HEAD: `18aeaed9af7dc5ffbe6cc79a4731a852fbb716da` (unchanged, uncommitted candidate left for handoff snapshot).
- Branch: `task-board/story/STORY-260908-v16gn5`.
- Candidate tree (temp-index, explicit 21 pathspecs only, real index/branch untouched): `ff61be4a8bd43fa4ffb179d31aa38e41891d4313` — MATCHES recovery-run observed landing tree.
- Old acceptance immutable: CR-TASK-260908-1wmb40-1 rev1 (tree `0361a3dfe25405da63ed9e70f886ba25880701db`, base `18aeaed...`) stays accepted history; no acceptance transfer claimed.
- Current signed authority (accepted from RUN-260909-49fa72 fresh observation, not re-observed this run): protected default `3ff66a9421ff6ddf675a49fc0c2868309f6e3de3`, tree `ff61be4...`, `remote_url ssh://git@github.com/relux-works/curator-agent-launcher`, `protected_ref refs/heads/main`, author Ivan Oparin, `signature_status G`, installed CLI `0.24.3-332-gac72ed99`.
- `git status` (unchanged from recovery): `M README.md`, `M SPEC.md`, `M go.mod`, untracked `.scripts/composition-mutants.py`, `.scripts/execution-mutants.py`, `.scripts/systemprompt-mutants.py`, `go.sum`, `internal/axconfig/`, `internal/composition/`, `internal/execution/`, `internal/systemprompt/`.

## What this run verified itself (exact commands/exits)

1. `task-board m 'set_status(TASK-260908-1wmb40, status=development)'` — exit 0 (development; story/epic escalated).
2. `sha256sum .scripts/composition-mutants.py internal/composition/composition.go internal/composition/composition_test.go internal/composition/probe.go go.mod go.sum` — exit 0; all 6 match recovery-run pinned hashes (`d44ad219...`, `e53003ce...`, `ef28f326...`, `0f453629...`, `f5a91ac6...`, `61692218...`). Composition paths byte-identical, no redevelopment.
3. `git ls-files --others --exclude-standard` + `git diff --name-only` — 21 explicit paths (3 modified + 18 untracked), matches recovery `changed_paths` count.
4. Temp-index tree check with explicit pathspecs only (`GIT_INDEX_FILE=<mktemp> git read-tree HEAD`, `git add -- <21 explicit paths>`, `git write-tree`; temp index removed) — exit 0, tree `ff61be4a8bd43fa4ffb179d31aa38e41891d4313`, MATCH. Never `git add -A`. Real index/branch verified untouched (`git status`, `git rev-parse HEAD` unchanged).
5. `task-board --no-update-check worktree status STORY-260908-v16gn5` — active, tip `18aeaed...`, tree dirty, lease held by RUN-260909-e08d63 (this run), CR rev1 accepted.

## Accepted from already-attached evidence (not rerun this turn)

- `make check` green, focused tests, 9/9 narrowing mutants killed with named expected-red tests: accepted from CR1 outcome `TASK-260908-1wmb40_results.md`, `TASK-260908-1wmb40_logs.zip`, and review verdict `TASK-260908-1wmb40_review-verdict-rev1.md` + `TASK-260908-1wmb40_review-logs-rev1.zip`. No manual broad rerun per explicit override.
- Fresh export bytes/hashes and authorized start-landed-rework transition: accepted from `TASK-260908-1wmb40_recovery-RUN-260909-49fa72.md` + `TASK-260908-1wmb40_prepare-landed-review-3ff66a9.json`.
- New-tree `make check` validation is the runtime handoff's obligation for the NEW candidate publication; no outcome here asserts a new CR exists until runtime publishes it.

## AC / DoD stance for this publication turn

- AC ratio: 10/10 API scope already driven by named committed tests in accepted CR1 (evidence in prior outcome/verdict); this turn adds 0 new product-code rows by design — republication preserves the exact tree. Full pipeline wiring, execution call-site `CheckLaunchBoundary` obligation, names-only stderr warnings, TOCTOU window, and no-native-Pi-before-operator-tag bounds carry over unchanged from recovery evidence.
- Checklist: source, tests, mutants, and composition behavior items rest on CR1 + recovery evidence; this run's items are preservation (uncommitted candidate, exact tree, no commit on Story branch — satisfied: no commit made), outcome artifact (this file), and handoff.
- No installs/restarts, hosted CI, tags/releases, ax/model launches, LOGBOOK/control-root/private-record writes, PR edits, or board-history synthesis this run.

## Next

Normal developer handoff publishes the NEW revision through configured validation once. No Complete until the new revision is independently accepted.
