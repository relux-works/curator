# TASK-260827-2gmk4c review verdict: accepted

Reviewer run `RUN-260827-083567` cycle 2, `CR-TASK-260827-2gmk4c-1` revision 1.
Delta reviewed: `git diff 9a35bc9e0e639f87e3e7f361b8da5032f4d1f80f edb5ef385b1d7fc537c7fc9c7b61f2f5bccb1e29`.
All five candidate blobs match the worktree byte for byte (`git show <tree>:<path> | diff - <path>` for each of
`LOGBOOK.md`, `README.md`, `docs/ci-gates.md`, `docs/compiled-commands.md`, `docs/prose-style.md`).

## Verdict

Accepted. Both blocking findings from cycle 1 are fixed and verified against the source tree, the
non-blocking nit is fixed as well, and the acceptance criteria hold on a fresh read of the whole file.

## Cycle 1 findings: resolution

**Finding 1 (blocking, `curator repair` does not exist) — fixed.** `docs/prose-style.md:19` now reads:

> "A compiled command links a skill binary to an executable shim in `.agents/bin/`.
> `curator install` rebuilds a missing or drifted shim."

Verified against the tree, not against the prior document. The shim location is real: shims land at
`<project>/.agents/bin/<command>` (`internal/install/install_test.go:171`,
`internal/install/dryrun_conformance_test.go:666`), and `internal/envfiles/envfiles.go:32` puts
`$CSK_PROJECT_ROOT/.agents/bin` on `PATH`. The repair claim is real:
`cmd/curator/builds.go:687-691` states "Curator has no separate repair command: install and upgrade are
the reconciliation path. They rebuild a missing, corrupt, drifted, or untrusted entry into new protected
state", and `docs/compiled-commands.md:59` says the same in the sibling document, so the two documents
now agree. `grep -rn "curator repair"` over all `*.md` and `*.go` in the worktree returns nothing: the
invented command is gone from the repository entirely, including from `LOGBOOK.md`.

**Finding 2 (cross-reference to a nonexistent `ARCHITECTURE.md`) — fixed.** `docs/prose-style.md:67` now
reads "see Reconciliation and repair in docs/compiled-commands.md". That target exists:
`grep -n '^#' docs/compiled-commands.md` puts `## Reconciliation and repair` at line 57. The rule now
demands a precise target and demonstrates one.

**Nit (em-dash Bad example printed a plain hyphen) — fixed.** Line 77 prints the real glyph
("`curator` — a skill manager"), matching how the guillemet entry at line 80 prints «Skillfile», and
line 102 restores the em-dash in the "Bad (slop)" block. The two rule statements now use one spelling:
line 45 and line 79 both say "Russian guillemets".

## Acceptance criteria

- **English rules ported**: Voice and sentences, Paragraphs, Terminology, Punctuation and typography,
  Lists vs prose, Tone, Blacklist, Worked contrast. The Russian engineering prose section is dropped and
  `grep -nP '[\x{0400}-\x{04FF}]'` finds no Cyrillic.
- **Blacklist with Bad/Good pairs**: all eight entries present with both halves (lines 73 to 96).
- **Comparative-overview allowance**: line 55. **Collapsible `<details>` allowance**: line 57.
- **Under 200 lines**: 106.
- **Self-consistent**: `grep -nP '[^\x00-\x7F]'` returns exactly three lines (77, 80, 102), and every one
  is an explicitly labelled Bad counter-example illustrating the glyph it bans. No em-dash, en-dash,
  guillemet, ellipsis, or exclamation point appears in the guide's own voice. A scan for the blacklisted
  constructions themselves ("not just", "it's not", "let's dive", "important to note", "powerful",
  "seamless", "robust", "blazingly") hits only rule statements and Bad examples.
- **Adaptation to Curator verified against the tree, not assumed**: `Skillfile.json` (`README.md:61`,
  `internal/manifest`), the `tag` ref kind (`internal/manifest/manifest.go:26,144`), `.agents/skills/`
  as the install root (`internal/install/install_test.go:150`,
  `internal/install/maintenance_test.go:72`), and the whitelist claim at line 61 (`tests`, `test`,
  `__tests__` are excluded at `internal/whitelist/whitelist.go:28`; the runtime-root escape hatch is
  exercised by `internal/whitelist/whitelist_test.go:104` "runtime root leaked into context").
- **DoD "No discrepancies between code and description"**: now satisfied. This was the item checked but
  unsatisfied in cycle 1.

## Gates

Not re-run, deliberately. `git diff --name-only <base> <candidate> | sed 's/.*\.//' | sort -u` yields
`md` alone: the delta touches zero Go files, so `gate-selftest`, `no-broad-suppression`, `go vet`, and
the test suites cannot distinguish this tree from its base and would be evidence of nothing. The checks
that do cover this change are the source-tree verifications above, each of which would have failed had
the rework been wrong: the `curator repair` grep, the heading grep in `docs/compiled-commands.md`, and
the non-ASCII inventory. Cycle 1 recorded `make gate-selftest` exit 0 (75/0) and
`make no-broad-suppression` exit 0 on the predecessor tree; nothing in this delta can have moved them.

## Observations (non-blocking, no rework required)

**The candidate tree bundles unreviewed sibling work.** This task's scope is `docs/prose-style.md` only,
but the candidate tree carries four more paths from `TASK-260827-2232c0` (README status-code expansion,
`docs/ci-gates.md` as a new file, the `cause`-subcode restructuring in `docs/compiled-commands.md`, and
`LOGBOOK.md` entry 0055). That sibling task is at `to-dev` and has not been reviewed. The bundling is an
artifact of the shared story worktree, and accepting this Change Request must not be read as accepting
those four paths. Flagged for the orchestrator at integration.

**The prose-style logbook entry did not survive the trunk rebase.** The DoD logbook item is checked, and
cycle 1 saw entry 0052 recording this deliverable, but `grep -n "prose-style\|2gmk4c" LOGBOOK.md` on the
candidate returns nothing; `git log -S` locates the entry only in the superseded WIP `81e56a13`, and
0052 was renumbered to a different topic upstream. Not raised as a defect: the lost text carried no
finding, only a deliverable line plus gate numbers, and that line was where `curator repair` was
recorded as a deliberate adaptation. The board outcome resource carries the deliverable record, and a
routine docs port with no anomaly does not meet the item's "when relevant" bar.

**`docs/compiled-commands.md` is named without code style at lines 39 and 67**, while line 37 requires
file names in code style. Both occurrences sit inside quoted illustrations and are inherited verbatim in
shape from the CocoaSkills source guide. Worth a pass if the file is touched again; not worth a cycle.

**Line 35, "In technical text, repetition is precision, not a defect", trails a negation.** It is not the
banned form: the blacklist bans leading with the negation and staging a strawman ("it's not X, it's Y"),
and this sentence states the claim first while correcting a real reader assumption. It is also verbatim
from the source guide. Recorded only so a future reviewer applying the blacklist mechanically does not
re-open it.
