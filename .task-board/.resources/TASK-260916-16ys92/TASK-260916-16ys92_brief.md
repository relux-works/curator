# Brief — TASK-260916-16ys92: print the resolved provider path in the launch line-group (E4, launcher)

Story `STORY-260916-2otjbn` (umbrella-provider-trust-roots), wave 1. Role:
developer. Repository: **curator-agent-launcher** (this control root,
`/Users/administrator/Developer/ReluxWorks/curator/curator-agent-launcher`);
your managed Story worktree is `<control-root>/.temp/STORY-260916-2otjbn/worktree`
on branch `task-board/story/STORY-260916-2otjbn`, provisioned at spawn. Work
ONLY there; no commits, pushes, branches, PRs, tags; no writes into the
control root or under `~/.agents`, `~/.claude`, `~/.codex`, `~/.curator`,
`~/.config/curator-run`. The board is the curator board (`TASK_BOARD_DIR` is
set); use `task-board m …` / `task-board resource add …` only. Models:
producers muse-spark max, reviewers gpt-6-astra low.

## Why (spec context)
curator-spec `0da4020` `protocol/environments.md` §11 (landed by PR #62,
E4): the manager resolves `curator-<name>` providers from trust roots (its
install directory, then the machine knob `provider_directories`), warns under
revision A when the `PATH`-selected provider lies outside the roots, refuses
under revision B, and reports the resolved absolute provider path in
`env status`. The launcher's own contribution: at every launch it prints the
resolved provider path — its own executable path as the umbrella resolved it —
in the §4.3 stderr line-group, so an operator sees which `curator-run` binary
is about to run and where it lives. Audit: curator-spec
`docs/security-audit-2026-09.md` E4; `verify-e-findings-rev4.md` on
`TASK-260916-dv7xv5`.

## Deliverable
1. **SPEC.md revision** (this repository owns the launcher SPEC): in §2
   (discovery and naming) state that the launcher is dispatched by the
   manager's umbrella under environments §11 trust-root resolution and that
   the launcher reports its resolved executable path; in §4.3 add the
   provider-path line to the line-group description: one line of the form
   `curator-run: provider: path=<absolute path> (<origin>)` where the path is
   the launcher's own executable resolved through symlinks
   (`os.Executable` + `filepath.EvalSymlinks`; on failure print the
   diagnostic-safe fallback the SPEC defines and never fail the launch), and
   `<origin>` is closed (e.g. `umbrella` when dispatched by the manager,
   `direct` when invoked by the operator — define the closed set and how it
   is detected, e.g. from the argv/env the umbrella already passes; if
   nothing distinguishes them today, print only the path and say so in the
   SPEC). Bump the SPEC revision history (§8.1) and the specification
   changelog; `CHANGELOG.md` Unreleased entry "E4: …".
2. **Implementation** (`cmd/curator-run`, `internal/defaults/lineup.go`
   `EmitGroup`/`Line`): emit the provider-path line as part of the same
   line-group at every launch, before the plan request; the value is folded
   with the existing `foldValue` framing rule so a hostile path never splits
   the line-group.
3. **Goldens/tests**: the pipeline goldens under `cmd/curator-run/testdata`
   show the provider path line (path normalized for the golden the way other
   host-specific values are); unit tests for the fold rule and the fallback
   path.
4. Keep every §4.3 model/effort semantics unchanged; no new flags; no
   behaviour change for `pi` beyond the line.

## Validation and handoff
`go build ./... && go vet ./... && gofmt -l . && go test ./...` from the
worktree (`set -o pipefail`, quote exit codes; regenerate goldens only
through the repository's documented mechanism). The runtime runs the
configured landing suite once at handoff; do not run it yourself. Attach
`TASK-260916-16ys92_results.md` (per-file changes, SPEC deltas, transcripts),
tick the checklist, then `task-board handoff TASK-260916-16ys92 --role developer`.
