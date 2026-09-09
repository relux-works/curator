# TASK-260908-1wmb40 — integration verification (RUN-260909-f405fe)

Bound integration run, role developer (archetype implementer), task already
`integrating`. No status change, no generic handoff per integration assignment.
CR1 acceptance (rev1, exact tree 0361a3dfe25405da63ed9e70f886ba25880701db) is
preserved immutable; this run made zero source edits.

## Candidate intact

HEAD `18aeaed9af7dc5ffbe6cc79a4731a852fbb716da`; working tree holds exactly the
7 accepted paths, uncommitted, nothing else non-ignored:

- `README.md` (modified), `go.mod` (modified)
- `go.sum`, `.scripts/composition-mutants.py`, `internal/composition/composition.go`,
  `internal/composition/composition_test.go`, `internal/composition/probe.go` (untracked)

`.temp/` scratch (including this run's `composition-mutants-integration/`) is
gitignored and excluded from the candidate tree.

## Fresh verification by this run (all observed, not inherited)

Host: Darwin arm64, Go 1.25.5, uid 502 (non-root: unreadable-file probe
executed, not skipped). Module: `skill-agents-management v0.5.10`, no
`replace`, no `go.work`.

| Check | Exit | Note |
|---|---|---|
| `go test ./internal/composition -count=1 -v` | 0 | all suites pass incl. `TestLaunchBoundaryFilesystem/unreadable` |
| `make check` (build/fmt-check/vet/test/race) | 0 | full repo green, run once |
| `python3 .scripts/composition-mutants.py .temp/composition-mutants-integration` | 0 | 9 of 9 narrowing mutants killed, each exit 1 at named test |
| `git diff --check` | 0 | clean |
| `gofmt -l cmd internal` | 0 | empty (clean) |

Mutant kills (M1–M3 `TestComposeEnvironmentBoundary`; M4
`TestComposeOwnershipFailure`; M5–M8 matching
`TestLaunchBoundaryFilesystem/<missing|dangling|directory|unreadable>`; M9
`TestComposeStdin/empty`). Harness restores candidate bytes in `finally`;
post-run `git status --short` confirms the same 7 paths, `diff-check` clean.

## AC coverage — 10 of 12 expanded rows driven (10 of 10 API scope)

Through exported production APIs (`internal/composition.Compose`,
`Value.CheckLaunchBoundary`, `ChildEnvironment.ChildEnv`):

1. Plan/prompt/MCP/native exact order, native uninspected — `Compose`, `TestComposeOrderAndChannels`
2. Binary/WorkDir retained, input independence — `Compose`, `TestComposeOrderAndChannels`
3. Full Plan.Env base, no removed-name re-admission — `Compose`, `TestComposeEnvironmentBoundary`, `TestComposeReleasedChildEnvironment`
4. Nil-parent same-request ownership, no inherited-secret serialization — `Compose → ChildEnv`, same tests
5. Disjoint literal/lookup names, inherited lookups retained, names-only warnings — `Compose`, `TestComposeEnvironmentBoundary`, `TestComposeUnownedOverrideNoWarning`
6. Unattached/attached-empty/UTF8/binary D4 + raw bytes — `Compose`, `TestComposeStdin`
7. MCP path+companions/name/variable/absence, no plan rebuilding — `Compose`, `TestComposeOrderAndChannels`, `TestLaunchBoundaryUnengaged`
8. Fresh missing/unreadable/regular codex-layer refusal, flags retained — `Value.CheckLaunchBoundary`, `TestLaunchBoundaryFilesystem`, `TestLaunchBoundaryLateReplacement`
9. Ownership failure propagates — `Compose`, `TestComposeOwnershipFailure`
10. Empty direct Env stays non-nil — `Compose`, `TestComposeEmptyEnvironment`
11. BOUND (not this leaf): full executable pipeline, stderr in both modes, tracked schema/extensions, real main/process/ax — execution Story obligation
12. BOUND (not this leaf): native Pi BuildLaunch admission against real operator tag — plan/execution Story obligation; Pi fixtures here prove composition only

Negative/narrowing evidence: every gate above ships a narrowing mutant (M1–M9);
each weakens the gate to admit exactly one rejected member and the named test
fails (exit 1). No source-text-inspection gate exists, so the token-preserving
source-gate attack is N/A; the harness executes the behavioral suite.

## Explicitly retained obligations for later stories

- Execution Story must call `Value.CheckLaunchBoundary` immediately before BOTH
  direct process creation and ax handoff, with binary and §5 file-kind checks;
  composition-time success is stale. Pathname replacement window (TOCTOU) remains.
- Both launch modes must print `Warnings` (names only) to stderr.
- Full pipeline tests, BuildLaunch admission, model/default resolution, §5
  prompt policy/application, tracked schema/extensions, stderr delivery, actual
  exec/ax remain later stories' obligations. No native Pi admission claimed.
- Reserved `path_prepend` unchanged; no managed command-root claim.

Next: bound `worktree checkpoint TASK-260908-1wmb40` from the control root.

## Lifecycle transaction attempts (this run, exact outputs)

- `worktree checkpoint TASK-260908-1wmb40` (from control root): refused —
  `change_request_final_leaf_checkpoint`: task is the last open child, so a
  checkpoint would also close STORY-260908-v16gn5; directed to
  `worktree integrate STORY-260908-v16gn5` instead.
- `worktree integrate STORY-260908-v16gn5 --cr TASK-260908-1wmb40 --revision 1`
  (from control root): refused — `board_owner_separate`: board repo
  `/Users/iv/Developer/ReluxWorks/curator` differs from control root
  `/Users/iv/Developer/ReluxWorks/.worktrees/launcher-control`; this repo's
  delivery path is its own PR followed by `worktree complete`, not integrate.
- `worktree complete` not attempted: it requires a `--landed-commit` proving
  the exact candidate tree on the code repo's freshly observed protected
  default, which does not exist yet — landing is the parent's signed-delivery
  step. Task correctly remains `integrating`; no status write or generic
  handoff performed per integration assignment.
