# TASK-260908-3jux68 producer evidence (B1: relux-root-context packages)

## Created

- Repo: https://github.com/relux-works/relux-root-context — visibility PRIVATE (confirmed via `gh repo view`).
- Bootstrap commit `9a6025d169a49b4cd692486bf8808f4dfc2d3044` on `main`
  (`feat: relux root-context packages from agents-infra dee5403 [skip ci]`),
  author `Ivan Oparin <ivan@relux.works>`, SSH signature verified:
  `Good "git" signature for ivan@relux.works with ED25519 key SHA256:Ng99XGF2pboYgFVfWJhYI2JRi0PyYsV9UwsJ70NBYd0`.
  Local `main` == `origin/main` (`git ls-remote` confirms).
- Layout: `packages/<name>/{agent-context.json,README.md,context/*.md}` for
  core, workflow, style, claude, attachments; `packages/relux-root-context-ivan/agent-context.json`
  (pure umbrella, no `context/`); repo `README.md`; `scripts/validate.sh`.
- Tags: NONE created. Per the brief, the six strict signed v-tags
  (`core/v1.0.0`, `workflow/v1.0.0`, `style/v1.0.0`, `claude/v1.0.0`,
  `attachments/v1.0.0`, `ivan/v1.0.0`) are cut on the reviewed head only
  after the independent (Astra) review verdict. Never retag/move.

## AC coverage: 11 of 13 rows driven pre-review

Production call sites: curator `internal/contextpkg.LoadManifest` +
`ValidateModules` (module presence/shape/bytes/selector), `internal/pkgversion.ParseRange` +
`Range.Satisfies` (umbrella ranges admit leaf versions); committed gate `scripts/validate.sh`.

| # | AC row | Evidence |
| - | --- | --- |
| 1 | New PRIVATE repo | `gh repo view` → `visibility: PRIVATE` |
| 2 | core: STRUCTURE, TOOLS, SKILLS, SKILL_TRIGGERS, DOCS, DIAGRAMS, PLATFORM | oracle `modules=9`, byte proof 9/9 EXACT |
| 3 | workflow: WORKFLOW, TESTING | oracle `modules=2`, byte proof 2/2 EXACT |
| 4 | style: STYLE, lowest leaf weight | oracle `modules=1`; validate.sh lowest-leaf check; mutant style-heavy killed |
| 5 | claude: EXTERNAL_RESOURCES, REMOTE_AGENTS, `environments: ["claude_code"]` | oracle, both modules selector-checked, zero unknown-env warnings; mutant unknown-env killed |
| 6 | attachments: ATTACHMENTS | oracle `modules=1`, byte proof EXACT |
| 7 | umbrella ivan: requires the five, weights, no own modules | oracle `hasContext=false`, 5 ranges each admitting `1.0.0`, weights agree |
| 8 | Valid `agent-context.json` per package | oracle `LoadManifest` × 6 PASS (curator's own strict schema-1 parser) |
| 9 | Module bytes exact from `dee5403` | sha256 15/15 EXACT vs `git rev-parse dee5403:<path>` blobs |
| 10 | Per-env modules (runtime text split, not rewritten) | claude selectors; no module reworded (see decision D1) |
| 11 | Weights per Decision 0012 | core 100 / workflow 70 / claude 50 / attachments 40 / style 10; umbrella `weights` map == leaf weights, no edge weights (avoids `context_weights_duplicate`); mutants weight-drift, edge-weight killed |
| 12 | Strict v-tag contract | BOUND: tags cut post-review on the reviewed head (names reserved above); manifests carry `1.0.0`, ranges `^1.0` admit it (proven both validators) |
| 13 | `requires.skills` (attachments CLI) + `requires.mcp` (figma, lldb, safari) in umbrella | BOUND: deferred to B2/B3 — target repos do not exist yet; umbrellla documents the gap; inventing git sources would be unverifiable |

## Validation log (exit codes)

- `scripts/validate.sh` → exit 0 (`manifests, modules, weights, ranges: OK`, `module bytes: OK`, `PASS`).
- Curator-parser oracle (`go run` inside curator module, removed afterwards) → exit 0, `ORACLE: PASS`
  (6 × LoadManifest, 15 × module byte/selector checks, 5 × range-admits, weights agreement).
- `bash -n scripts/validate.sh` → clean; `shellcheck -S warning` → clean.
- `git log --show-signature -1` → Good signature (above); `git status` clean, `main` == `origin/main`.

## Mutant table (harness `/tmp/b1mutants.sh`, throwaway; mutants in `/tmp/b1mut-*`, removed)

| Mutant | Narrows the gate to | validate.sh (named check, exit 1) | Oracle (exit) | Survivor bound |
| --- | --- | --- | --- | --- |
| trailing-lf (extra `\n` in DOCS module) | §3 exactly-one-trailing-LF | `module bytes invalid` | 1 `profile_module_bytes_invalid` | — killed both |
| unknown-env (`claude_code`→`claude_macro`) | selector must name a registered env | `selects unregistered environment 'claude_macro'` | 1 `profile_selector_unknown_environment` | — killed both |
| weight-drift (map core 100→90) | umbrella map == leaf weight | `disagrees with leaf weight (100)` | 1 same | — killed both |
| range-miss (`^1.0`→`^2.0`) | range must admit required version | `does not admit … (1, 0, 0)` | 1 same | — killed both |
| edge-weight (adds edge `weight`) | no edge weights beside `weights` map | `must not mix edge weights with the weights map` | 1 same | — killed both |
| unknown-field (`maintainer` in style) | strict schema, no unknown keys | `unknown fields ['maintainer']` | 1 `unknown field maintainer` | — killed both |
| style-heavy (style 10→45 + map) | style carries lowest leaf weight | `must carry the lowest leaf weight` | 0 PASS | BOUND: oracle checks the absolute contract only, not relative ordering; relative order is validate.sh's job |

Two validate.sh self-bugs found and fixed during development (umbrella weight 0 in lowest-weight scope; `ROOT` undefined in `check_bytes`) — the gate bites its own author.

## Producer decisions (for reviewer contest)

- D1 bit-exact over goal-rewrite: story title/description/DoD and the brief require exact bytes
  ("no rewording"); the goal permits rewriting retired `agents-infra codex|claude` launcher refs in
  TOOLS. Kept 15/15 bit-exact (proven above): the launcher (`curator run`) is unshipped and the
  agents-infra launchers unretired (B6), so a rewrite now would point at a nonexistent command.
  Revisit with B6.
- D2 BROWSER_AUTOMATION (unassigned by the split) carried in core: tool-capability module, core owns TOOLS.
- D3 source loader index (`.instructions/INSTRUCTIONS.md`, the fuller variant incl. EXTERNAL_RESOURCES)
  carried as core's entry module; its `@~/.agents/...` lines are inert post-migration (materialization
  supersedes them), noted in the core README. Root `AGENTS.md` render (Generated-by header) omitted:
  the mapping assigns that header to Curator materialization (`curator-root-context-v2`), not module bytes.
- D4 bootstrap commit directly on `main`, no PR: an empty repo admits no branch→PR flow (standard
  `gh repo create` bootstrap). Review happens on head `9a6025d`; any rework follows branch→PR→merge;
  tags follow the verdict. No review bypassed: nothing lands after review except the tags.
- D5 weights (0012 fixes machinery, not values): 100/70/50/40/10 with style lowest per mapping;
  claude-vs-attachments order (50/40) is arbitrary-but-stable, recorded here.
- D6 umbrella `requires` carry `{git, directory, range}` only; `git` is the canonical
  `https://github.com/relux-works/relux-root-context.git`.

## Known gaps / post-review steps

1. Astra (medium) independent review of head `9a6025d`; verdict recorded before tags.
2. Cut six signed v-tags on the reviewed head; never move them.
3. B2/B3 add `requires.skills` / `requires.mcp` edges + bump umbrella minor.
4. B5 installs `relux-root-context-ivan` via `profile install --use --takeover` (with backup).
5. Untouched by this task (per brief §6): `~/.agents`, `~/.claude`, `~/.codex`, real `ax`, agents-infra launchers, LOGBOOK.md, control root. Story worktree left uncommitted/clean of this work (deliverable lives in the new repo; this file is the board-attached outcome).
